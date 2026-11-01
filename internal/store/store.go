// Package store persists scans in SQLite (local-first storage).
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/codedna/codedna/internal/model"
)

// Store wraps the SQLite database.
type Store struct {
	db *sql.DB
}

// Open opens (creating when necessary) the database at path.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	db, err := sql.Open("sqlite3", path+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS repos (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	root TEXT NOT NULL UNIQUE,
	created_at INTEGER NOT NULL,
	git_json TEXT
);
CREATE TABLE IF NOT EXISTS scans (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	repo_id INTEGER NOT NULL REFERENCES repos(id),
	started_at INTEGER NOT NULL,
	finished_at INTEGER,
	status TEXT NOT NULL,
	stats_json TEXT
);
CREATE TABLE IF NOT EXISTS files (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	scan_id INTEGER NOT NULL REFERENCES scans(id),
	path TEXT NOT NULL,
	language TEXT NOT NULL,
	size INTEGER, loc INTEGER, sha256 TEXT, tags_json TEXT
);
CREATE TABLE IF NOT EXISTS symbols (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	scan_id INTEGER NOT NULL REFERENCES scans(id),
	sid TEXT NOT NULL,
	file_path TEXT NOT NULL,
	kind TEXT NOT NULL,
	name TEXT NOT NULL,
	qualified_name TEXT NOT NULL,
	line INTEGER, col INTEGER, end_line INTEGER,
	exported INTEGER, complexity INTEGER, line_count INTEGER,
	language TEXT, parent_class TEXT
);
CREATE TABLE IF NOT EXISTS relationships (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	scan_id INTEGER NOT NULL REFERENCES scans(id),
	kind TEXT NOT NULL,
	from_id TEXT NOT NULL,
	to_id TEXT,
	to_name TEXT,
	import_path TEXT,
	confidence REAL NOT NULL,
	method TEXT NOT NULL,
	locations_json TEXT
);
CREATE TABLE IF NOT EXISTS issues (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	scan_id INTEGER NOT NULL REFERENCES scans(id),
	kind TEXT NOT NULL,
	severity TEXT NOT NULL,
	title TEXT NOT NULL,
	file_path TEXT,
	symbol_sid TEXT,
	detail_json TEXT
);
CREATE TABLE IF NOT EXISTS runtime_traces (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	scan_id INTEGER NOT NULL REFERENCES scans(id),
	run_id TEXT NOT NULL,
	function_name TEXT NOT NULL,
	file_path TEXT NOT NULL,
	line INTEGER,
	duration_us INTEGER,
	timestamp_ns INTEGER,
	caller_function TEXT,
	caller_file TEXT,
	http_path TEXT,
	http_method TEXT,
	error TEXT
);
CREATE TABLE IF NOT EXISTS arch_rules (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	repo_id INTEGER NOT NULL REFERENCES repos(id),
	rule_id TEXT NOT NULL,
	name TEXT NOT NULL,
	from_pattern TEXT NOT NULL,
	to_pattern TEXT NOT NULL,
	forbidden INTEGER NOT NULL,
	description TEXT,
	severity TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS impact_reports (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	scan_id INTEGER NOT NULL REFERENCES scans(id),
	base_ref TEXT NOT NULL,
	target_ref TEXT NOT NULL,
	risk_score REAL NOT NULL,
	risk_level TEXT NOT NULL,
	report_json TEXT NOT NULL,
	created_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS workspaces (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	repo_ids_json TEXT NOT NULL,
	created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_symbols_sid ON symbols(scan_id, sid);
CREATE INDEX IF NOT EXISTS idx_symbols_file ON symbols(scan_id, file_path);
CREATE INDEX IF NOT EXISTS idx_rel_from ON relationships(scan_id, from_id);
CREATE INDEX IF NOT EXISTS idx_rel_to ON relationships(scan_id, to_id);
CREATE INDEX IF NOT EXISTS idx_traces_scan ON runtime_traces(scan_id);
`)
	return err
}

func defaultDBPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codedna", "codedna.db")
}

// DefaultPath resolves the DB path (CODEDNA_HOME overrides ~/.codedna).
func DefaultPath() string {
	if h := os.Getenv("CODEDNA_HOME"); h != "" {
		return filepath.Join(h, "codedna.db")
	}
	return defaultDBPath()
}

// GraphQL — no; aliasing convention only.
type txnFunc = func(*sql.Tx) error

// UpsertRepo registers a repository by root path.
func (s *Store) UpsertRepo(name, root string, git model.GitMeta) (int64, error) {
	gj, _ := json.Marshal(git)
	res, err := s.db.Exec(`INSERT INTO repos (name, root, created_at, git_json)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(root) DO UPDATE SET name=excluded.name, git_json=excluded.git_json`,
		name, root, time.Now().Unix(), string(gj))
	if err != nil {
		return 0, err
	}
	if id, err := res.LastInsertId(); err == nil && id > 0 {
		return id, nil
	}
	var id int64
	row := s.db.QueryRow(`SELECT id FROM repos WHERE root = ?`, root)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// BeginScan starts a scan row with status running.
func (s *Store) BeginScan(repoID int64) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO scans (repo_id, started_at, status) VALUES (?, ?, 'running')`,
		repoID, time.Now().Unix())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// FinishScan finalizes a scan with stats.
func (s *Store) FinishScan(scanID int64, status string, stats *model.RepoScanStats) error {
	now := time.Now().Unix()
	if stats != nil {
		gj, _ := json.Marshal(stats)
		_, err := s.db.Exec(`UPDATE scans SET finished_at=?, status=?, stats_json=? WHERE id=?`,
			now, status, string(gj), scanID)
		return err
	}
	_, err := s.db.Exec(`UPDATE scans SET finished_at=?, status=? WHERE id=?`, now, status, scanID)
	return err
}

// SaveFiles persists file nodes.
func (s *Store) SaveFiles(scanID int64, files []model.FileNode) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmtInsert, err := tx.Prepare(`INSERT INTO files (scan_id, path, language, size, loc, tags_json) VALUES (?,?,?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, f := range files {
		tj, _ := json.Marshal(f.Tags)
		if _, err := stmtInsert.Exec(scanID, f.Path, f.Language, f.Size, f.LOC, string(tj)); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// SaveSymbols persists symbols.
func (s *Store) SaveSymbols(scanID int64, syms []model.Symbol) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	st, err := tx.Prepare(`INSERT INTO symbols (scan_id, sid, file_path, kind, name, qualified_name, line, col, end_line,
		exported, complexity, line_count, language, parent_class)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, sm := range syms {
		exp := 0
		if sm.Exported {
			exp = 1
		}
		if _, err := st.Exec(scanID, sm.ID, sm.File, string(sm.Kind), sm.Name, sm.QualifiedName,
			sm.Line, sm.Column, sm.EndLine, exp, sm.Complexity, sm.LineCount, sm.Language, sm.ParentClassName); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// SaveRelationships persists graph edges.
func (s *Store) SaveRelationships(scanID int64, rels []model.Relationship) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	st, err := tx.Prepare(`INSERT INTO relationships
		(scan_id, kind, from_id, to_id, to_name, import_path, confidence, method, locations_json)
		VALUES (?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, r := range rels {
		lj, _ := json.Marshal(r.Locations)
		if _, err := st.Exec(scanID, string(r.Kind), r.FromID, r.ToID, r.ToName, r.ImportPath,
			r.Confidence, string(r.Method), string(lj)); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// SaveIssues persists analysis findings.
func (s *Store) SaveIssues(scanID int64, issues []model.Issue) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	st, err := tx.Prepare(`INSERT INTO issues (scan_id, kind, severity, title, file_path, symbol_sid, detail_json)
		VALUES (?,?,?,?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, is := range issues {
		dj, _ := json.Marshal(is.Detail)
		if _, err := st.Exec(scanID, string(is.Kind), string(is.Severity), is.Title, is.File, is.SymbolID, string(dj)); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// Repo lookup helpers -----------------------------------------------------

// repoRow identifies a stored repository.
type repoRow struct {
	model.RepoInfo
}

// RepoByRoot returns the repo id for a root path (0 when missing).
func (s *Store) RepoByRoot(root string) (int64, error) {
	var id int64
	var k int64
	row := s.db.QueryRow(`SELECT id FROM repos WHERE root = ?`, root)
	if err := row.Scan(&k); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	id = k
	return id, nil
}

// LastScanID returns the newest completed scan for a repo, 0 when none.
func (s *Store) LastScanID(repoID int64) (int64, error) {
	row := s.db.QueryRow(`SELECT MAX(id) FROM scans WHERE repo_id = ? AND status = 'complete'`, repoID)
	var id sql.NullInt64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	if !id.Valid {
		return 0, nil
	}
	return id.Int64, nil
}

// LoadSymbols returns all symbols for a scan.
func (s *Store) LoadSymbols(scanID int64) ([]model.Symbol, error) {
	rows, err := s.db.Query(`SELECT sid, file_path, kind, name, qualified_name, line, col, end_line,
		exported, complexity, line_count, language, parent_class FROM symbols WHERE scan_id = ?`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Symbol
	for rows.Next() {
		var sm model.Symbol
		var kind, lang string
		var exp int
		var pc sql.NullString
		if err := rows.Scan(&sm.ID, &sm.File, &kind, &sm.Name, &sm.QualifiedName, &sm.Line, &sm.Column,
			&sm.EndLine, &exp, &sm.Complexity, &sm.LineCount, &lang, &pc); err != nil {
			return nil, err
		}
		sm.Kind = model.SymbolKind(kind)
		sm.Language = lang
		sm.Exported = exp == 1
		sm.ParentClassName = pc.String
		out = append(out, sm)
	}
	return out, rows.Err()
}

// LoadRelationships returns all relationships for a scan.
func (s *Store) LoadRelationships(scanID int64) ([]model.Relationship, error) {
	rows, err := s.db.Query(`SELECT kind, from_id, to_id, to_name, import_path, confidence, method, locations_json
		FROM relationships WHERE scan_id = ?`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Relationship
	for rows.Next() {
		var r model.Relationship
		var kind, method string
		var toID, toName, importPath sql.NullString
		var lj sql.NullString
		if err := rows.Scan(&kind, &r.FromID, &toID, &toName, &importPath, &r.Confidence, &method, &lj); err != nil {
			return nil, err
		}
		r.Kind = model.RelationshipKind(kind)
		r.Method = model.ResolutionMethod(method)
		r.ToID, r.ToName, r.ImportPath = toID.String, toName.String, importPath.String
		if lj.Valid && lj.String != "" {
			json.Unmarshal([]byte(lj.String), &r.Locations)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// LoadIssues returns issues for a scan.
func (s *Store) LoadIssues(scanID int64) ([]model.Issue, error) {
	rows, err := s.db.Query(`SELECT kind, severity, title, file_path, symbol_sid, detail_json
		FROM issues WHERE scan_id = ?`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Issue
	for rows.Next() {
		var is model.Issue
		var kind, sev string
		var fp, sid, dj sql.NullString
		if err := rows.Scan(&kind, &sev, &is.Title, &fp, &sid, &dj); err != nil {
			return nil, err
		}
		is.Kind = model.IssueKind(kind)
		is.Severity = model.Severity(sev)
		is.File, is.SymbolID = fp.String, sid.String
		if dj.Valid && dj.String != "" {
			json.Unmarshal([]byte(dj.String), &is.Detail)
		}
		out = append(out, is)
	}
	return out, rows.Err()
}

// LoadFiles returns file nodes for a scan.
func (s *Store) LoadFiles(scanID int64) ([]model.FileNode, error) {
	rows, err := s.db.Query(`SELECT path, language, size, loc, tags_json FROM files WHERE scan_id = ?`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.FileNode
	for rows.Next() {
		f := model.FileNode{}
		var tj sql.NullString
		if err := rows.Scan(&f.Path, &f.Language, &f.Size, &f.LOC, &tj); err != nil {
			return nil, err
		}
		if tj.Valid && tj.String != "" {
			json.Unmarshal([]byte(tj.String), &f.Tags)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// RepoRootByID returns root path of repo.
func (s *Store) RepoRootByID(id int64) (string, error) {
	row := s.db.QueryRow(`SELECT root FROM repos WHERE id = ?`, id)
	var root string
	if err := row.Scan(&root); err != nil {
		return "", err
	}
	return root, nil
}

// LastStats returns stats of the most recent scan.
func (s *Store) LastStats(repoID int64) (*model.RepoScanStats, int64, error) {
	row := s.db.QueryRow(`SELECT id, stats_json FROM scans WHERE repo_id = ? AND status='complete' ORDER BY id DESC LIMIT 1`, repoID)
	var id int64
	var gj sql.NullString
	if err := row.Scan(&id, &gj); err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	if !gj.Valid {
		return nil, id, nil
	}
	var st model.RepoScanStats
	if err := json.Unmarshal([]byte(gj.String), &st); err != nil {
		return nil, id, err
	}
	return &st, id, nil
}

// SaveRuntimeTrace records a single execution event.
func (s *Store) SaveRuntimeTrace(scanID int64, event model.RuntimeTraceEvent) error {
	return s.SaveRuntimeTraces(scanID, []model.RuntimeTraceEvent{event})
}

// SaveRuntimeTraces batch saves runtime trace events.
func (s *Store) SaveRuntimeTraces(scanID int64, events []model.RuntimeTraceEvent) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO runtime_traces (
		scan_id, run_id, function_name, file_path, line,
		duration_us, timestamp_ns, caller_function, caller_file,
		http_path, http_method, error
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		_, err := stmt.Exec(
			scanID, e.RunID, e.FunctionName, e.File, e.Line,
			e.DurationUs, e.TimestampNs, e.CallerFunction, e.CallerFile,
			e.HTTPPath, e.HTTPMethod, e.Error,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// LoadRuntimeTraces returns all runtime trace events for a scan.
func (s *Store) LoadRuntimeTraces(scanID int64) ([]model.RuntimeTraceEvent, error) {
	rows, err := s.db.Query(`SELECT id, run_id, function_name, file_path, line,
		duration_us, timestamp_ns, caller_function, caller_file,
		http_path, http_method, error
		FROM runtime_traces WHERE scan_id = ? ORDER BY timestamp_ns ASC`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.RuntimeTraceEvent
	for rows.Next() {
		var e model.RuntimeTraceEvent
		var callerFn, callerFile, httpPath, httpMethod, errStr sql.NullString
		if err := rows.Scan(
			&e.ID, &e.RunID, &e.FunctionName, &e.File, &e.Line,
			&e.DurationUs, &e.TimestampNs, &callerFn, &callerFile,
			&httpPath, &httpMethod, &errStr,
		); err != nil {
			return nil, err
		}
		if callerFn.Valid {
			e.CallerFunction = callerFn.String
		}
		if callerFile.Valid {
			e.CallerFile = callerFile.String
		}
		if httpPath.Valid {
			e.HTTPPath = httpPath.String
		}
		if httpMethod.Valid {
			e.HTTPMethod = httpMethod.String
		}
		if errStr.Valid {
			e.Error = errStr.String
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// SaveArchRules saves architecture rules for a repo.
func (s *Store) SaveArchRules(repoID int64, rules []model.ArchRule) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM arch_rules WHERE repo_id = ?`, repoID); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO arch_rules (
		repo_id, rule_id, name, from_pattern, to_pattern,
		forbidden, description, severity
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range rules {
		forb := 0
		if r.Forbidden {
			forb = 1
		}
		if _, err := stmt.Exec(
			repoID, r.ID, r.Name, r.FromPattern, r.ToPattern,
			forb, r.Description, string(r.Severity),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// LoadArchRules loads architecture rules for a repo.
func (s *Store) LoadArchRules(repoID int64) ([]model.ArchRule, error) {
	rows, err := s.db.Query(`SELECT rule_id, name, from_pattern, to_pattern,
		forbidden, description, severity FROM arch_rules WHERE repo_id = ?`, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ArchRule
	for rows.Next() {
		var r model.ArchRule
		var forb int
		var desc sql.NullString
		var sev string
		if err := rows.Scan(&r.ID, &r.Name, &r.FromPattern, &r.ToPattern, &forb, &desc, &sev); err != nil {
			return nil, err
		}
		r.Forbidden = forb == 1
		if desc.Valid {
			r.Description = desc.String
		}
		r.Severity = model.Severity(sev)
		out = append(out, r)
	}
	return out, rows.Err()
}

// SaveImpactReport records a change impact analysis report.
func (s *Store) SaveImpactReport(scanID int64, report *model.ImpactReport) error {
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO impact_reports (
		scan_id, base_ref, target_ref, risk_score, risk_level, report_json, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		scanID, report.BaseRef, report.TargetRef, report.RiskScore, string(report.RiskLevel), string(data), report.CreatedAt.Unix(),
	)
	return err
}

// LoadImpactReport loads the latest impact report for a scan.
func (s *Store) LoadImpactReport(scanID int64) (*model.ImpactReport, error) {
	row := s.db.QueryRow(`SELECT report_json FROM impact_reports WHERE scan_id = ? ORDER BY id DESC LIMIT 1`, scanID)
	var data string
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var r model.ImpactReport
	if err := json.Unmarshal([]byte(data), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// LoadAllRepos lists all repositories discovered or scanned.
func (s *Store) LoadAllRepos() ([]model.RepoInfo, error) {
	rows, err := s.db.Query(`SELECT id, name, root, created_at, git_json,
		COALESCE((SELECT MAX(s.id) FROM scans s WHERE s.repo_id = repos.id), 0)
		FROM repos ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.RepoInfo
	for rows.Next() {
		var r model.RepoInfo
		var createdSec int64
		var gitJ sql.NullString
		if err := rows.Scan(&r.ID, &r.Name, &r.Root, &createdSec, &gitJ, &r.LatestScanID); err != nil {
			return nil, err
		}
		r.CreatedAt = time.Unix(createdSec, 0)
		if gitJ.Valid && gitJ.String != "" {
			json.Unmarshal([]byte(gitJ.String), &r.Git)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// LoadRepo gets a repo by ID.
func (s *Store) LoadRepo(id int64) (*model.RepoInfo, error) {
	row := s.db.QueryRow(`SELECT id, name, root, created_at, git_json FROM repos WHERE id = ?`, id)
	var r model.RepoInfo
	var createdSec int64
	var gitJ sql.NullString
	if err := row.Scan(&r.ID, &r.Name, &r.Root, &createdSec, &gitJ); err != nil {
		return nil, err
	}
	r.CreatedAt = time.Unix(createdSec, 0)
	if gitJ.Valid && gitJ.String != "" {
		json.Unmarshal([]byte(gitJ.String), &r.Git)
	}
	return &r, nil
}

// LoadScans returns all scans for a repo.
func (s *Store) LoadScans(repoID int64) ([]model.Scan, error) {
	rows, err := s.db.Query(`SELECT id, repo_id, started_at, finished_at, status, stats_json FROM scans WHERE repo_id = ? ORDER BY id DESC`, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Scan
	for rows.Next() {
		var sc model.Scan
		var startSec, finSec sql.NullInt64
		var statsJ sql.NullString
		if err := rows.Scan(&sc.ID, &sc.RepoID, &startSec, &finSec, &sc.Status, &statsJ); err != nil {
			return nil, err
		}
		if startSec.Valid {
			sc.StartedAt = time.Unix(startSec.Int64, 0)
		}
		if finSec.Valid {
			t := time.Unix(finSec.Int64, 0)
			sc.FinishedAt = &t
		}
		if statsJ.Valid && statsJ.String != "" {
			var st model.RepoScanStats
			if err := json.Unmarshal([]byte(statsJ.String), &st); err == nil {
				sc.Stats = &st
			}
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

// SaveWorkspace creates a workspace grouping multiple repos.
func (s *Store) SaveWorkspace(name string, repoIDs []int64) (*model.Workspace, error) {
	now := time.Now().Unix()
	rj, err := json.Marshal(repoIDs)
	if err != nil {
		return nil, err
	}
	res, err := s.db.Exec(`INSERT INTO workspaces (name, repo_ids_json, created_at) VALUES (?, ?, ?)`, name, string(rj), now)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.LoadWorkspace(id)
}

// LoadWorkspaces returns all workspaces.
func (s *Store) LoadWorkspaces() ([]model.Workspace, error) {
	rows, err := s.db.Query(`SELECT id, name, repo_ids_json, created_at FROM workspaces ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Workspace
	for rows.Next() {
		var ws model.Workspace
		var rj string
		var createdSec int64
		if err := rows.Scan(&ws.ID, &ws.Name, &rj, &createdSec); err != nil {
			return nil, err
		}
		ws.CreatedAt = time.Unix(createdSec, 0)
		json.Unmarshal([]byte(rj), &ws.RepoIDs)
		out = append(out, ws)
	}
	return out, rows.Err()
}

// LoadWorkspace loads a single workspace by ID.
func (s *Store) LoadWorkspace(id int64) (*model.Workspace, error) {
	row := s.db.QueryRow(`SELECT id, name, repo_ids_json, created_at FROM workspaces WHERE id = ?`, id)
	var ws model.Workspace
	var rj string
	var createdSec int64
	if err := row.Scan(&ws.ID, &ws.Name, &rj, &createdSec); err != nil {
		return nil, err
	}
	ws.CreatedAt = time.Unix(createdSec, 0)
	json.Unmarshal([]byte(rj), &ws.RepoIDs)
	return &ws, nil
}

// LoadRepoRootForScan retrieves the absolute filesystem root path of the repo for a given scan ID or repo ID.
func (s *Store) LoadRepoRootForScan(scanID int64) (string, error) {
	var root string
	err := s.db.QueryRow(`SELECT r.root FROM scans s JOIN repos r ON s.repo_id = r.id WHERE s.id = ?`, scanID).Scan(&root)
	if err == nil && root != "" {
		return root, nil
	}
	// Fallback: check if scanID was actually a repo_id
	err = s.db.QueryRow(`SELECT root FROM repos WHERE id = ?`, scanID).Scan(&root)
	if err == nil && root != "" {
		return root, nil
	}
	return "", err
}

// FindRepoRootForFile searches all registered repos to see which one contains the specified relative file.
func (s *Store) FindRepoRootForFile(cleanedPath string) string {
	rows, err := s.db.Query(`SELECT root FROM repos`)
	if err != nil {
		return ""
	}
	defer rows.Close()

	for rows.Next() {
		var root string
		if err := rows.Scan(&root); err == nil && root != "" {
			candidate := filepath.Join(root, cleanedPath)
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
				return root
			}
		}
	}
	return ""
}

// Close releases the DB handle.
func (s *Store) Close() error { return s.db.Close() }
