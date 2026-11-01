package report

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/codedna/codedna/internal/model"
)

var safeIDRe = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func sanitizeID(id string) string {
	cleaned := safeIDRe.ReplaceAllString(id, "_")
	cleaned = strings.Trim(cleaned, "_")
	if len(cleaned) > 0 && cleaned[0] >= '0' && cleaned[0] <= '9' {
		cleaned = "n_" + cleaned
	}
	if cleaned == "" {
		cleaned = "node"
	}
	return cleaned
}

// GenerateMermaid converts graph nodes and edges into a formatted, subgraphed Mermaid flowchart.
func GenerateMermaid(repoName string, graph model.GraphData) string {
	var sb strings.Builder
	sb.WriteString("```mermaid\n")
	sb.WriteString("flowchart TD\n")
	sb.WriteString(fmt.Sprintf("    %% Architecture Map for %s\n\n", repoName))

	// Group nodes by directory or service
	moduleMap := make(map[string][]model.GraphNode)
	idMap := make(map[string]string)

	for _, n := range graph.Nodes {
		mod := "root"
		if n.Data != nil && n.Data["service"] != nil {
			mod = fmt.Sprintf("%v", n.Data["service"])
		} else if n.File != "" && strings.Contains(n.File, "/") {
			mod = strings.Split(n.File, "/")[0]
		}
		moduleMap[mod] = append(moduleMap[mod], n)
		idMap[n.ID] = sanitizeID(n.ID)
	}

	// Sort modules alphabetically
	var modules []string
	for m := range moduleMap {
		modules = append(modules, m)
	}
	sort.Strings(modules)

	// Emit subgraphs
	for _, m := range modules {
		nodes := moduleMap[m]
		subID := sanitizeID("sub_" + m)
		sb.WriteString(fmt.Sprintf("    subgraph %s [\"📁 %s\"]\n", subID, m))
		for _, n := range nodes {
			safeNodeID := idMap[n.ID]
			label := n.Label
			if label == "" {
				label = filepath.Base(n.File)
			}
			label = strings.ReplaceAll(label, `"`, `'`)
			sb.WriteString(fmt.Sprintf("        %s[\"%s\"]\n", safeNodeID, label))
		}
		sb.WriteString("    end\n\n")
	}

	// Emit edges
	edgeCount := 0
	for _, e := range graph.Edges {
		src := idMap[e.Source]
		tgt := idMap[e.Target]
		if src == "" || tgt == "" || src == tgt {
			continue
		}
		edgeCount++
		if edgeCount > 400 { // Keep Mermaid renderable without hitting browser node limit
			break
		}

		relLabel := e.Kind
		if e.Method != "" {
			relLabel = e.Method
		}
		relLabel = strings.ReplaceAll(relLabel, `"`, `'`)

		if e.Kind == "cross_service_api" {
			sb.WriteString(fmt.Sprintf("    %s -.->|\"%s\"| %s\n", src, relLabel, tgt))
		} else if e.Kind == "calls" {
			sb.WriteString(fmt.Sprintf("    %s ==>|\"%s\"| %s\n", src, relLabel, tgt))
		} else {
			sb.WriteString(fmt.Sprintf("    %s -->|\"%s\"| %s\n", src, relLabel, tgt))
		}
	}

	// Add styles
	sb.WriteString("\n    classDef default fill:#18181b,stroke:#3f3f46,color:#f4f4f5,stroke-width:1px;\n")
	sb.WriteString("```\n")

	return sb.String()
}

// GenerateC4 converts graph components into a C4 PlantUML / C4 Mermaid container specification.
func GenerateC4(repoName string, graph model.GraphData) string {
	var sb strings.Builder
	sb.WriteString("```mermaid\n")
	sb.WriteString("C4Component\n")
	sb.WriteString(fmt.Sprintf("    title Component Architecture Diagram for %s\n\n", repoName))

	// Group by boundary/container
	containers := make(map[string][]model.GraphNode)
	idMap := make(map[string]string)

	for _, n := range graph.Nodes {
		cont := "Core Application"
		if n.Data != nil && n.Data["service"] != nil {
			cont = fmt.Sprintf("%v", n.Data["service"])
		} else if n.File != "" && strings.Contains(n.File, "/") {
			cont = strings.Split(n.File, "/")[0]
		}
		containers[cont] = append(containers[cont], n)
		idMap[n.ID] = sanitizeID(n.ID)
	}

	for cName, cNodes := range containers {
		cID := sanitizeID("cont_" + cName)
		sb.WriteString(fmt.Sprintf("    Container_Boundary(%s, \"%s\") {\n", cID, cName))
		for _, n := range cNodes {
			cNodeID := idMap[n.ID]
			label := n.Label
			if label == "" {
				label = filepath.Base(n.File)
			}
			desc := fmt.Sprintf("%s (%s)", n.Type, n.Language)
			if n.Complexity > 0 {
				desc = fmt.Sprintf("%s, LOC/Complexity: %d", desc, n.Complexity)
			}
			sb.WriteString(fmt.Sprintf("        Component(%s, \"%s\", \"%s\", \"%s\")\n", cNodeID, label, n.Type, desc))
		}
		sb.WriteString("    }\n\n")
	}

	// Relationships
	seen := make(map[string]bool)
	for _, e := range graph.Edges {
		src := idMap[e.Source]
		tgt := idMap[e.Target]
		if src == "" || tgt == "" || src == tgt {
			continue
		}
		pair := src + "->" + tgt
		if seen[pair] {
			continue
		}
		seen[pair] = true

		rel := e.Kind
		if e.Method != "" {
			rel = e.Method
		}
		sb.WriteString(fmt.Sprintf("    Rel(%s, %s, \"%s\")\n", src, tgt, rel))
	}

	sb.WriteString("```\n")
	return sb.String()
}
