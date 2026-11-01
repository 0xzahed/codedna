// Package blame analyzes git commit history, numstat, and line authorship to compute
// developer knowledge distribution, file ownership, and bus factor risks.
package blame

import (
	"bufio"
	"bytes"
	"math"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/codedna/codedna/internal/model"
)

// AnalyzeKnowledge computes knowledge distribution, file primary ownership, and bus factor risks.
func AnalyzeKnowledge(repoRoot string) (*model.RepoKnowledgeMap, error) {
	// Run git log with author and numstat
	cmd := exec.Command("git", "-C", repoRoot, "log", "--numstat", "--format=AUTHOR|%an|%ae")
	out, err := cmd.Output()
	if err != nil {
		// Non-git repository fallback
		return &model.RepoKnowledgeMap{
			IsGitRepo:       false,
			TotalAuthors:    0,
			TotalCommits:    0,
			HighRiskFiles:   0,
			BusFactorScore:  100.0,
			TopContributors: []model.AuthorStats{},
			Files:           []model.FileOwnership{},
		}, nil
	}

	return parseGitLogOutput(out)
}

func parseGitLogOutput(data []byte) (*model.RepoKnowledgeMap, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	type authorKey struct {
		name  string
		email string
	}

	authorCommits := make(map[authorKey]int)
	authorAdded := make(map[authorKey]int)
	authorDeleted := make(map[authorKey]int)
	authorFiles := make(map[authorKey]map[string]bool)

	fileAuthors := make(map[string]map[authorKey]int) // file -> author -> linesAdded
	totalCommits := 0
	totalLinesAdded := 0

	var currentAuthor authorKey
	hasCurrentAuthor := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "AUTHOR|") {
			parts := strings.Split(line, "|")
			name := "Unknown"
			email := ""
			if len(parts) > 1 && parts[1] != "" {
				name = parts[1]
			}
			if len(parts) > 2 {
				email = parts[2]
			}
			currentAuthor = authorKey{name: name, email: email}
			hasCurrentAuthor = true
			authorCommits[currentAuthor]++
			if authorFiles[currentAuthor] == nil {
				authorFiles[currentAuthor] = make(map[string]bool)
			}
			totalCommits++
			continue
		}

		if !hasCurrentAuthor {
			continue
		}

		// numstat line format: <added>\t<deleted>\t<filepath>
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			added, _ := strconv.Atoi(fields[0])
			deleted, _ := strconv.Atoi(fields[1])
			filePath := fields[2]

			// Handle renames e.g. "old => new" or "prefix/{old => new}"
			if strings.Contains(filePath, "=>") {
				parts := strings.Split(filePath, "=>")
				filePath = strings.TrimSpace(parts[len(parts)-1])
				filePath = strings.Trim(filePath, "}")
			}

			authorAdded[currentAuthor] += added
			authorDeleted[currentAuthor] += deleted
			authorFiles[currentAuthor][filePath] = true
			totalLinesAdded += added

			if fileAuthors[filePath] == nil {
				fileAuthors[filePath] = make(map[authorKey]int)
			}
			fileAuthors[filePath][currentAuthor] += added
		}
	}

	// Calculate File Ownership and Bus Factor Risks
	var fileList []model.FileOwnership
	highRiskCount := 0
	primaryFileCountPerAuthor := make(map[authorKey]int)

	for file, contribs := range fileAuthors {
		if len(contribs) == 0 {
			continue
		}

		fileTotalAdded := 0
		var shares []model.ContributorShare
		for auth, added := range contribs {
			fileTotalAdded += added
			shares = append(shares, model.ContributorShare{
				Name:       auth.name,
				Email:      auth.email,
				LinesAdded: added,
			})
		}

		sort.Slice(shares, func(i, j int) bool {
			return shares[i].LinesAdded > shares[j].LinesAdded
		})

		for i := range shares {
			if fileTotalAdded > 0 {
				shares[i].Percentage = math.Round((float64(shares[i].LinesAdded)/float64(fileTotalAdded))*1000) / 10
			} else {
				shares[i].Percentage = 100.0
			}
		}

		primary := shares[0]
		primaryKey := authorKey{name: primary.Name, email: primary.Email}
		primaryFileCountPerAuthor[primaryKey]++

		// Bus factor risk determination
		risk := model.SevLow
		if len(shares) == 1 || primary.Percentage >= 90.0 {
			risk = model.SevHigh
			highRiskCount++
		} else if primary.Percentage >= 70.0 {
			risk = model.SevMedium
		}

		fileList = append(fileList, model.FileOwnership{
			FilePath:          file,
			PrimaryAuthor:     primary.Name,
			ContributorsCount: len(shares),
			TopContributors:   shares,
			BusFactorRisk:     risk,
		})
	}

	sort.Slice(fileList, func(i, j int) bool {
		return fileList[i].FilePath < fileList[j].FilePath
	})

	// Calculate Top Contributors
	var topAuthors []model.AuthorStats
	for auth, commits := range authorCommits {
		added := authorAdded[auth]
		deleted := authorDeleted[auth]
		filesTouched := len(authorFiles[auth])
		primaryCount := primaryFileCountPerAuthor[auth]

		share := 0.0
		if totalLinesAdded > 0 {
			share = math.Round((float64(added)/float64(totalLinesAdded))*1000) / 10
		}

		topAuthors = append(topAuthors, model.AuthorStats{
			Name:             auth.name,
			Email:            auth.email,
			Commits:          commits,
			LinesAdded:       added,
			LinesDeleted:     deleted,
			FilesTouched:     filesTouched,
			PrimaryFileCount: primaryCount,
			KnowledgeShare:   share,
		})
	}

	sort.Slice(topAuthors, func(i, j int) bool {
		return topAuthors[i].LinesAdded > topAuthors[j].LinesAdded
	})

	// Bus Factor Score: 0 to 100
	// Higher score = more distributed knowledge, healthy team collaboration.
	// Penalized if majority of files are single-contributor high risk.
	busScore := 100.0
	if len(fileList) > 0 {
		highRiskRatio := float64(highRiskCount) / float64(len(fileList))
		busScore = math.Max(10.0, math.Round((1.0-highRiskRatio*0.6)*1000)/10)
		if len(topAuthors) == 1 {
			busScore = 50.0
		}
	}

	return &model.RepoKnowledgeMap{
		IsGitRepo:       true,
		TotalAuthors:    len(topAuthors),
		TotalCommits:    totalCommits,
		HighRiskFiles:   highRiskCount,
		BusFactorScore:  busScore,
		TopContributors: topAuthors,
		Files:           fileList,
	}, nil
}
