// Package architecture enforces architectural boundaries and detects
// layer violations, forbidden dependencies, and architecture drift.
package architecture

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codedna/codedna/internal/analysis"
	"github.com/codedna/codedna/internal/model"
)

// DefaultRules returns standard clean architecture boundaries.
func DefaultRules() []model.ArchRule {
	return []model.ArchRule{
		{
			ID:          "rule-layer-domain-clean",
			Name:        "Domain Isolation",
			FromPattern: "**/domain/**",
			ToPattern:   "**/controllers/**",
			Forbidden:   true,
			Description: "Domain models and business logic must not depend on presentation or HTTP controllers",
			Severity:    model.SevHigh,
		},
		{
			ID:          "rule-layer-domain-db",
			Name:        "Domain Storage Isolation",
			FromPattern: "**/domain/**",
			ToPattern:   "**/db/**",
			Forbidden:   true,
			Description: "Domain core must not depend directly on database implementations",
			Severity:    model.SevMedium,
		},
		{
			ID:          "rule-controllers-repo",
			Name:        "Controller to DB Bypass",
			FromPattern: "**/controllers/**",
			ToPattern:   "**/store/**",
			Forbidden:   false, // Example of allowed or customized
			Description: "Controllers should go through services rather than direct raw storage",
			Severity:    model.SevLow,
		},
	}
}

// LoadRules loads custom rules from .codedna/architecture.json or codedna.arch.json,
// or returns DefaultRules if none exists.
func LoadRules(repoRoot string) ([]model.ArchRule, error) {
	paths := []string{
		filepath.Join(repoRoot, ".codedna", "architecture.json"),
		filepath.Join(repoRoot, "codedna.arch.json"),
	}

	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			var rules []model.ArchRule
			if err := json.Unmarshal(data, &rules); err == nil && len(rules) > 0 {
				return rules, nil
			}
		}
	}

	return DefaultRules(), nil
}

// SaveRules persists architecture rules into .codedna/architecture.json in repoRoot.
func SaveRules(repoRoot string, rules []model.ArchRule) error {
	dir := filepath.Join(repoRoot, ".codedna")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "architecture.json"), data, 0644)
}

// AddRule appends or updates an architecture rule and persists the rule set.
func AddRule(repoRoot string, rule model.ArchRule) ([]model.ArchRule, error) {
	rules, err := LoadRules(repoRoot)
	if err != nil {
		rules = DefaultRules()
	}

	if rule.ID == "" {
		slug := strings.ToLower(rule.Name)
		slug = strings.ReplaceAll(slug, " ", "-")
		if slug == "" {
			slug = "rule"
		}
		rule.ID = fmt.Sprintf("%s-%d", slug, time.Now().UnixNano()%10000)
	}

	found := false
	for i, r := range rules {
		if r.ID == rule.ID {
			rules[i] = rule
			found = true
			break
		}
	}
	if !found {
		rules = append(rules, rule)
	}

	if err := SaveRules(repoRoot, rules); err != nil {
		return rules, err
	}
	return rules, nil
}

// DeleteRule removes a rule by ID and saves the updated rule set.
func DeleteRule(repoRoot string, ruleID string) ([]model.ArchRule, error) {
	rules, err := LoadRules(repoRoot)
	if err != nil {
		rules = DefaultRules()
	}

	var updated []model.ArchRule
	for _, r := range rules {
		if r.ID != ruleID {
			updated = append(updated, r)
		}
	}

	if err := SaveRules(repoRoot, updated); err != nil {
		return updated, err
	}
	return updated, nil
}

// ResetRules restores default architectural rules.
func ResetRules(repoRoot string) ([]model.ArchRule, error) {
	defaults := DefaultRules()
	if err := SaveRules(repoRoot, defaults); err != nil {
		return defaults, err
	}
	return defaults, nil
}

// CheckArchitecture evaluates relationships against architecture rules and cycle detection.
func CheckArchitecture(rules []model.ArchRule, rels []model.Relationship, syms []model.Symbol, cycles []model.Issue) model.ArchHealthReport {
	report := model.ArchHealthReport{
		Violations: []model.ArchViolation{},
		Cycles:     []model.CycleDetail{},
		RuleCount:  len(rules),
		Rules:      rules,
	}

	symMap := make(map[string]model.Symbol)
	for _, s := range syms {
		symMap[s.ID] = s
	}

	// 1. Check relationship violations against forbidden rules
	for _, rel := range rels {
		fromSym := symMap[rel.FromID]
		toSym := symMap[rel.ToID]

		fromFile := fromSym.File
		if fromFile == "" {
			continue
		}

		toFile := toSym.File
		if toFile == "" && rel.ImportPath != "" {
			toFile = rel.ImportPath
		}
		if toFile == "" {
			continue
		}

		for _, rule := range rules {
			if !rule.Forbidden {
				continue
			}

			if matchPattern(rule.FromPattern, fromFile) && matchPattern(rule.ToPattern, toFile) {
				locLine := 0
				if len(rel.Locations) > 0 {
					locLine = rel.Locations[0].Line
				}

				report.Violations = append(report.Violations, model.ArchViolation{
					RuleID:     rule.ID,
					RuleName:   rule.Name,
					Severity:   rule.Severity,
					FromFile:   fromFile,
					ToFile:     toFile,
					ImportPath: rel.ImportPath,
					Reason:     rule.Description,
					Line:       locLine,
				})
			}
		}
	}

	// 2. Extract cycle details
	for _, c := range cycles {
		if c.Kind == model.IssueCircularDep {
			report.CircularCount++
			if dt, ok := c.Detail.(map[string]interface{}); ok {
				var files []string
				if fList, ok := dt["files"].([]string); ok {
					files = fList
				}
				chain, _ := dt["chain"].(string)
				report.Cycles = append(report.Cycles, model.CycleDetail{
					Files: files,
					Chain: chain,
				})
			}
		}
	}

	// 3. Calculate Layer Health Score (0-100)
	penalty := 0.0
	for _, v := range report.Violations {
		switch v.Severity {
		case model.SevHigh:
			penalty += 15.0
		case model.SevMedium:
			penalty += 8.0
		case model.SevLow:
			penalty += 3.0
		default:
			penalty += 1.0
		}
	}
	penalty += float64(report.CircularCount) * 12.0

	score := math.Max(0.0, 100.0-penalty)
	report.LayerHealthScore = math.Round(score*10) / 10

	return report
}

// EvaluateRepo runs complete architecture audit on an index and module graph.
func EvaluateRepo(repoRoot string, idx *analysis.Index, graph *analysis.FileGraph) (model.ArchHealthReport, error) {
	rules, err := LoadRules(repoRoot)
	if err != nil {
		rules = DefaultRules()
	}

	var cycles []model.Issue
	if graph != nil {
		cycles = graph.Cycles()
	}

	return CheckArchitecture(rules, idx.Rel, idx.Symbols, cycles), nil
}

// matchPattern evaluates simple glob patterns (** and *).
func matchPattern(pattern, target string) bool {
	pattern = filepath.ToSlash(pattern)
	target = filepath.ToSlash(target)

	// Direct equality
	if pattern == target || pattern == "*" || pattern == "**" {
		return true
	}

	// Standard prefix / suffix matches
	if strings.HasPrefix(pattern, "**/") && strings.HasSuffix(pattern, "/**") {
		mid := pattern[3 : len(pattern)-3]
		return strings.Contains(target, "/"+mid+"/") || strings.HasPrefix(target, mid+"/") || strings.Contains(target, mid)
	}

	if strings.HasPrefix(pattern, "**/") {
		sub := pattern[3:]
		return strings.HasSuffix(target, sub) || strings.Contains(target, "/"+sub)
	}

	if strings.HasSuffix(pattern, "/**") {
		prefix := pattern[:len(pattern)-3]
		return strings.HasPrefix(target, prefix) || strings.Contains(target, prefix+"/")
	}

	matched, err := filepath.Match(pattern, target)
	return err == nil && matched
}
