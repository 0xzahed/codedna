// Package sysdesign reconstructs high-level system design architecture, multi-tier data flow,
// and design patterns from static AST facts and dependency graphs.
package sysdesign

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/codedna/codedna/internal/model"
)

// ClassifyTiers categorizes files and symbols into high-level system design tiers.
func ClassifyTiers(files []model.FileNode, syms []model.Symbol) []model.SystemTier {
	tierDefs := []struct {
		id    model.TierKind
		name  string
		desc  string
		order int
	}{
		{
			id:    model.TierClient,
			name:  "Client & Gateway Tier",
			desc:  "Web views, frontend pages, reverse proxy configs, and API gateway routing entrypoints",
			order: 1,
		},
		{
			id:    model.TierAPI,
			name:  "API & Presentation Layer",
			desc:  "HTTP handlers, controllers, REST endpoints, GraphQL resolvers, and RPC service interfaces",
			order: 2,
		},
		{
			id:    model.TierService,
			name:  "Application & Domain Services",
			desc:  "Core business domain logic, orchestrators, use cases, domain entities, and policy rules",
			order: 3,
		},
		{
			id:    model.TierStorage,
			name:  "Persistence & Data Storage",
			desc:  "Repositories, database entities, ORM models, SQL queries, cache access, and migrations",
			order: 4,
		},
		{
			id:    model.TierMessaging,
			name:  "Asynchronous Workers & Queues",
			desc:  "Background task workers, message queues, event emitters, consumer loops, and webhooks",
			order: 5,
		},
		{
			id:    model.TierIntegration,
			name:  "External Third-Party Services",
			desc:  "Payment gateways, external cloud storage, third-party API clients, and mailers",
			order: 6,
		},
	}

	tierFiles := make(map[model.TierKind][]string)
	tierSyms := make(map[model.TierKind][]string)

	for _, f := range files {
		tier := detectFileTier(f.Path)
		tierFiles[tier] = append(tierFiles[tier], f.Path)
	}

	for _, s := range syms {
		tier := detectFileTier(s.File)
		tierSyms[tier] = append(tierSyms[tier], s.QualifiedName)
	}

	var tiers []model.SystemTier
	for _, def := range tierDefs {
		fList := tierFiles[def.id]
		sList := tierSyms[def.id]

		tiers = append(tiers, model.SystemTier{
			ID:          def.id,
			Name:        def.name,
			Description: def.desc,
			Order:       def.order,
			FileCount:   len(fList),
			SymbolCount: len(sList),
			Files:       fList,
			KeySymbols:  sList[:min(len(sList), 10)],
		})
	}

	return tiers
}

// ComputeDataFlow calculates end-to-end data flow hops across tiers from relationship graph edges.
func ComputeDataFlow(rels []model.Relationship, syms []model.Symbol) []model.DataFlowHop {
	symFileMap := make(map[string]string)
	for _, s := range syms {
		symFileMap[s.ID] = s.File
	}

	type hopKey struct {
		from model.TierKind
		to   model.TierKind
	}

	hopCounts := make(map[hopKey]int)
	hopSamples := make(map[hopKey][]string)

	for _, r := range rels {
		fromFile := symFileMap[r.FromID]
		if fromFile == "" && strings.HasPrefix(r.FromID, "file:") {
			fromFile = strings.TrimPrefix(r.FromID, "file:")
		}
		toFile := symFileMap[r.ToID]
		if toFile == "" && strings.HasPrefix(r.ToID, "file:") {
			toFile = strings.TrimPrefix(r.ToID, "file:")
		}
		if toFile == "" && r.ImportPath != "" {
			toFile = r.ImportPath
		}

		if fromFile == "" || toFile == "" {
			continue
		}

		fromTier := detectFileTier(fromFile)
		toTier := detectFileTier(toFile)

		if fromTier == toTier {
			continue
		}

		key := hopKey{from: fromTier, to: toTier}
		hopCounts[key]++
		if len(hopSamples[key]) < 3 {
			sample := fmt.Sprintf("%s ➔ %s (%s)", filepath.Base(fromFile), filepath.Base(toFile), r.Method)
			hopSamples[key] = append(hopSamples[key], sample)
		}
	}

	var hops []model.DataFlowHop
	for key, count := range hopCounts {
		protocol := determineProtocol(key.from, key.to)
		hops = append(hops, model.DataFlowHop{
			FromTier:    key.from,
			ToTier:      key.to,
			Protocol:    protocol,
			CallCount:   count,
			SampleCalls: hopSamples[key],
		})
	}

	sort.Slice(hops, func(i, j int) bool {
		return hops[i].CallCount > hops[j].CallCount
	})

	return hops
}

func detectFileTier(path string) model.TierKind {
	p := strings.ToLower(filepath.ToSlash(path))

	// 1. Messaging & Background Jobs
	if strings.Contains(p, "/worker") || strings.Contains(p, "/queue") || strings.Contains(p, "/task") ||
		strings.Contains(p, "/consumer") || strings.Contains(p, "/jobs") || strings.Contains(p, "/events") ||
		strings.Contains(p, "/subscriber") || strings.Contains(p, "/watcher") {
		return model.TierMessaging
	}

	// 2. External Third-Party Integration
	if strings.Contains(p, "/stripe") || strings.Contains(p, "/payment") || strings.Contains(p, "/paypal") ||
		strings.Contains(p, "/client") || strings.Contains(p, "/external") || strings.Contains(p, "/third_party") ||
		strings.Contains(p, "/webhook") || strings.Contains(p, "/mailer") || strings.Contains(p, "/s3") {
		return model.TierIntegration
	}

	// 3. Persistence & Data Storage
	if strings.Contains(p, "/store") || strings.Contains(p, "/db") || strings.Contains(p, "/model") ||
		strings.Contains(p, "/repo") || strings.Contains(p, "/repository") || strings.Contains(p, "/dao") ||
		strings.Contains(p, "/migration") || strings.Contains(p, "/entity") || strings.Contains(p, "/schema") ||
		strings.Contains(p, "/database") || strings.Contains(p, "/cache") || strings.Contains(p, "/redis") {
		return model.TierStorage
	}

	// 4. API & Presentation
	if strings.Contains(p, "/controller") || strings.Contains(p, "/handler") || strings.Contains(p, "/api") ||
		strings.Contains(p, "/router") || strings.Contains(p, "/endpoint") || strings.Contains(p, "/routes") ||
		strings.Contains(p, "/views") || strings.Contains(p, "/graphql") || strings.Contains(p, "/grpc") {
		return model.TierAPI
	}

	// 5. Client & Gateway
	if strings.HasPrefix(p, "web/") || strings.Contains(p, "/pages") || strings.Contains(p, "/components") ||
		strings.Contains(p, "/frontend") || strings.Contains(p, "/ui") || strings.Contains(p, "/app") ||
		strings.Contains(p, "nginx") || strings.Contains(p, "gateway") || strings.HasSuffix(p, ".tsx") ||
		strings.HasSuffix(p, ".jsx") || strings.HasSuffix(p, ".vue") || strings.HasSuffix(p, ".html") {
		return model.TierClient
	}

	// 6. Application & Domain Service (Default core)
	return model.TierService
}

func determineProtocol(from, to model.TierKind) string {
	if from == model.TierClient && to == model.TierAPI {
		return "HTTPS / REST API"
	}
	if from == model.TierAPI && to == model.TierService {
		return "In-Memory Orchestration"
	}
	if to == model.TierStorage {
		return "SQL / Storage Interface"
	}
	if to == model.TierMessaging || from == model.TierMessaging {
		return "Async Event / Channel Pipeline"
	}
	if to == model.TierIntegration {
		return "External HTTPS API"
	}
	return "Internal Contract Call"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
