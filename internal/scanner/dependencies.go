package scanner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/codedna/codedna/internal/model"
)

// detectGit collects git metadata without executing any repository code.
func detectGit(root string) (model.GitMeta, error) {
	meta := model.GitMeta{}
	gitDir := filepath.Join(root, ".git")
	isGit := false
	if _, err := os.Stat(gitDir); err == nil {
		isGit = true
	} else {
		// Check if inside a worktree or git repo
		cmd := exec.Command("git", "-C", root, "rev-parse", "--git-dir")
		if err := cmd.Run(); err == nil {
			isGit = true
		}
	}

	if !isGit {
		return meta, nil // not a git repo — never fatal
	}
	meta.IsGitRepo = true

	git := func(args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"-C", root, "--no-pager"}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		out, err := cmd.Output()
		return strings.TrimSpace(string(out)), err
	}

	// Determine branch or tag
	if b, err := git("branch", "--show-current"); err == nil && b != "" {
		meta.Branch = b
	} else if b, err := git("symbolic-ref", "--short", "HEAD"); err == nil && b != "" {
		meta.Branch = b
	} else if b, err := git("describe", "--tags", "--exact-match"); err == nil && b != "" {
		meta.Branch = b
	} else if b, err := git("rev-parse", "--abbrev-ref", "HEAD"); err == nil && b != "HEAD" {
		meta.Branch = b
	}

	// Commit info (may fail on unborn branch / empty repo)
	if c, err := git("rev-parse", "HEAD"); err == nil {
		meta.Commit = c
	}
	if line, err := git("log", "-1", "--format=%an%n%ae%n%aI", "HEAD"); err == nil {
		if parts := strings.SplitN(line, "\n", 3); len(parts) == 3 {
			meta.AuthorName, meta.AuthorEmail = parts[0], parts[1]
			if t, terr := time.Parse(time.RFC3339, strings.TrimSpace(parts[2])); terr == nil {
				meta.LastCommitAt = &t
			}
		}
	}
	if n, err := git("rev-list", "--count", "HEAD"); err == nil {
		fmt.Sscanf(n, "%d", &meta.CommitCount)
	}
	if out, err := git("ls-files"); err == nil && out != "" {
		meta.TrackedFiles = int64(len(strings.Split(out, "\n")))
	}
	if r, err := git("remote", "get-url", "origin"); err == nil {
		// do not leak embedded credentials in the URL
		meta.RemoteURL = redactURL(r)
	}
	if st, err := git("status", "--porcelain"); err == nil {
		meta.IsDirty = len(strings.TrimSpace(st)) > 0
	}
	return meta, nil
}

var credRe = regexp.MustCompile(`(https?://)([^@/\s]+@)`)

func redactURL(u string) string {
	return credRe.ReplaceAllString(u, "$1")
}

// extractDependencies reads dependency manifests. Read-only.
func extractDependencies(root string, cfg map[string]string) map[string][]string {
	deps := map[string][]string{}
	if p, ok := cfg["package.json"]; ok {
		d := readPackageJSONDeps(filepath.Join(root, p))
		deps["npm"] = d
		deps["javascript"] = d
	}
	if p, ok := cfg["go.mod"]; ok {
		d := readGoMod(filepath.Join(root, p))
		deps["go"] = d
	}
	if p, ok := cfg["requirements.txt"]; ok {
		d := readRequirements(filepath.Join(root, p))
		deps["pip"] = d
		deps["python"] = d
	}
	if p, ok := cfg["pyproject.toml"]; ok {
		d := append(unique(deps["pip"]), readPyProject(filepath.Join(root, p))...)
		deps["pip"] = d
		deps["python"] = d
	}
	if p, ok := cfg["composer.json"]; ok {
		d := readComposer(filepath.Join(root, p))
		deps["composer"] = d
		deps["php"] = d
	}
	if p, ok := cfg["pom.xml"]; ok {
		d := readPom(filepath.Join(root, p))
		deps["maven"] = d
		deps["java"] = d
	}
	for eco, list := range deps {
		sortStrings(list)
		deps[eco] = unique(list)
	}
	return deps
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func unique(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	seen := map[string]bool{}
	out := []string{}
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func readPackageJSONDeps(path string) []string {
	var out []string
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var pkg struct {
		Deps         map[string]string `json:"dependencies"`
		DevDeps      map[string]string `json:"devDependencies"`
		PeerDeps     map[string]string `json:"peerDependencies"`
	}
	if err := json.NewDecoder(f).Decode(&pkg); err != nil {
		return nil
	}
	for k := range pkg.Deps {
		out = append(out, k)
	}
	for k := range pkg.DevDeps {
		out = append(out, k+" (dev)")
	}
	for k := range pkg.PeerDeps {
		out = append(out, k+" (peer)")
	}
	return out
}

func readGoMod(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "require ") {
			fields := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(fields) >= 2 {
				out = append(out, fields[0]+" "+fields[1])
			}
			continue
		}
		if strings.HasPrefix(line, "\t") {
			l := strings.TrimSpace(line)
			if l == "" || l == ")" || strings.HasPrefix(l, "//") {
				continue
			}
			fields := strings.Fields(l)
			if len(fields) >= 2 && !strings.HasPrefix(fields[0], "/") {
				out = append(out, fields[0]+" "+fields[1])
			}
		}
	}
	return out
}

func readRequirements(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		name := line
		for i := 0; i < len(name); i++ {
			c := rune(name[i])
			if c == '=' || c == '>' || c == '<' || c == '!' || c == '~' || c == ';' || c == ' ' || c == '[' {
				name = strings.TrimSpace(name[:i])
				break
			}
		}
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func readPyProject(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	inDeps := false
	inArray := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.Contains(line, "dependencies") && strings.Contains(line, "[") && strings.HasSuffix(line, "[") {
			inDeps = true
			inArray = true
			continue
		}
		if inDeps && inArray {
			if strings.HasPrefix(line, "]") {
				inDeps = false
				inArray = false
				continue
			}
			q := strings.Trim(strings.Trim(line, ","), `"' `)
			if q != "" && strings.Contains(line, `"`) {
				name := q
				for i := 0; i < len(name); i++ {
					if c := name[i]; c == '=' || c == '>' || c == '<' || c == '!' || c == '~' || c == '[' || c == ';' {
						name = strings.TrimSpace(name[:i])
						break
					}
				}
				out = append(out, name)
			}
			continue
		}
		if inDeps && !inArray {
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			q := strings.Trim(strings.Trim(line, ","), `"' `)
			out = append(out, q)
		}
	}
	return out
}

func readComposer(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var c struct {
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil
	}
	var out []string
	for k := range c.Require {
		if k == "php" || k == "ext-*" || strings.HasPrefix(k, "ext-") {
			continue
		}
		out = append(out, k)
	}
	for k := range c.RequireDev {
		out = append(out, k+" (dev)")
	}
	return out
}

var reG = regexp.MustCompile(`<groupId>([^<]+)</groupId>`)
var reA = regexp.MustCompile(`<artifactId>([^<]+)</artifactId>`)

func readPom(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	inDep := false
	g := ""
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.Contains(line, "<dependency>") {
			inDep = true
			g = ""
			continue
		}
		if strings.Contains(line, "</dependency>") {
			if g != "" {
				out = append(out, g)
			}
			inDep = false
			continue
		}
		if inDep {
			if m := reA.FindStringSubmatch(line); m != nil && g != "" {
				out = append(out, g+":"+m[1])
			} else if m := regroup(line); m != "" {
				g = m
			}
		}
	}
	return out
}

func regroup(line string) string {
	if m := reG.FindStringSubmatch(line); m != nil {
		// skip parent groupId (project-level); readPom only enters during <dependency>
		return m[1]
	}
	return ""
}
