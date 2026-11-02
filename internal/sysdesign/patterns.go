package sysdesign

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/codedna/codedna/internal/model"
)

// AnalyzeSystemDesign performs end-to-end architectural analysis: tier categorization,
// data flow hops, software pattern recognition, and bottleneck detection.
func AnalyzeSystemDesign(files []model.FileNode, syms []model.Symbol, rels []model.Relationship) model.SystemDesignReport {
	tiers := ClassifyTiers(files, syms)
	dataFlow := ComputeDataFlow(rels, syms)
	patterns := DetectPatterns(files, syms, rels)
	entrypoints := detectKeyEntrypoints(files, syms)
	bottlenecks := detectBottlenecks(tiers, dataFlow, rels)

	archStyle := inferArchitectureStyle(tiers, patterns, dataFlow)
	summary := generateDesignSummary(archStyle, tiers, patterns, dataFlow)
	reliability := calculateReliabilityScore(tiers, patterns, bottlenecks, dataFlow)
	recommendations := generateRecommendations(tiers, dataFlow, bottlenecks)

	return model.SystemDesignReport{
		ArchitectureStyle:  archStyle,
		Summary:            summary,
		Tiers:              tiers,
		DataFlow:           dataFlow,
		DetectedPatterns:   patterns,
		KeyEntrypoints:     entrypoints,
		BottleneckHotspots: bottlenecks,
		Recommendations:    recommendations,
		ReliabilityScore:   reliability,
	}
}

// DetectPatterns recognizes classic Gang of Four and modern distributed/architectural patterns.
func DetectPatterns(files []model.FileNode, syms []model.Symbol, rels []model.Relationship) []model.DetectedPattern {
	patterns := make([]model.DetectedPattern, 0)

	if p, ok := detectRepositoryPattern(files, syms); ok {
		patterns = append(patterns, p)
	}
	if p, ok := detectFactoryBuilderPattern(syms); ok {
		patterns = append(patterns, p)
	}
	if p, ok := detectSingletonPattern(files, syms); ok {
		patterns = append(patterns, p)
	}
	if p, ok := detectEventDrivenPattern(files, syms); ok {
		patterns = append(patterns, p)
	}
	if p, ok := detectMiddlewareChainPattern(files, syms); ok {
		patterns = append(patterns, p)
	}
	if p, ok := detectCleanArchitecturePattern(files, syms, rels); ok {
		patterns = append(patterns, p)
	}
	if p, ok := detectCQRSPattern(files, syms); ok {
		patterns = append(patterns, p)
	}
	if p, ok := detectResiliencePattern(files, syms); ok {
		patterns = append(patterns, p)
	}
	if p, ok := detectAdapterPattern(files, syms); ok {
		patterns = append(patterns, p)
	}
	if p, ok := detectStrategyPattern(files, syms); ok {
		patterns = append(patterns, p)
	}
	if p, ok := detectObserverPattern(files, syms); ok {
		patterns = append(patterns, p)
	}

	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Confidence > patterns[j].Confidence
	})

	return patterns
}

func detectRepositoryPattern(files []model.FileNode, syms []model.Symbol) (model.DetectedPattern, bool) {
	var repoFiles []string
	var repoSymbols []string

	for _, f := range files {
		p := strings.ToLower(f.Path)
		if strings.Contains(p, "repo") || strings.Contains(p, "store") || strings.Contains(p, "dao") {
			repoFiles = append(repoFiles, f.Path)
		}
	}

	for _, s := range syms {
		name := strings.ToLower(s.Name)
		if strings.HasSuffix(name, "repo") || strings.HasSuffix(name, "repository") ||
			strings.HasSuffix(name, "store") || strings.HasSuffix(name, "dao") {
			repoSymbols = append(repoSymbols, s.QualifiedName)
		}
	}

	if len(repoFiles) == 0 && len(repoSymbols) == 0 {
		return model.DetectedPattern{}, false
	}

	conf := 0.65
	if len(repoFiles) >= 2 || len(repoSymbols) >= 3 {
		conf = 0.92
	}

	evidence := []string{
		fmt.Sprintf("Found %d storage/repository file(s)", len(repoFiles)),
		fmt.Sprintf("Discovered %d data access symbol(s) encapsulating query operations", len(repoSymbols)),
	}
	if len(repoSymbols) > 0 {
		sample := repoSymbols[:min(len(repoSymbols), 4)]
		evidence = append(evidence, fmt.Sprintf("Key components: %s", strings.Join(sample, ", ")))
	}

	return model.DetectedPattern{
		ID:          "pattern-repository",
		Name:        "Repository / Data Mapper Pattern",
		Category:    model.PatternStructural,
		Confidence:  conf,
		Description: "Mediates between domain logic and persistence layers using collection-like interfaces, decoupling SQL/storage specifics from business rules.",
		Evidence:    evidence,
		Files:       repoFiles[:min(len(repoFiles), 10)],
		Tradeoffs:   "Provides clear boundaries and simple unit testing via mocks; introduces minor boilerplate for simple CRUD queries.",
	}, true
}

func detectFactoryBuilderPattern(syms []model.Symbol) (model.DetectedPattern, bool) {
	var factorySyms []string
	var files []string
	fileSeen := make(map[string]bool)

	for _, s := range syms {
		name := s.Name
		if (strings.HasPrefix(name, "New") || strings.HasPrefix(name, "Build") || strings.HasPrefix(name, "Create")) &&
			(s.Kind == model.SymbolFunction || s.Kind == model.SymbolMethod) {
			factorySyms = append(factorySyms, s.QualifiedName)
			if !fileSeen[s.File] && s.File != "" {
				fileSeen[s.File] = true
				files = append(files, s.File)
			}
		}
	}

	if len(factorySyms) < 2 {
		return model.DetectedPattern{}, false
	}

	conf := 0.75
	if len(factorySyms) >= 5 {
		conf = 0.95
	}

	return model.DetectedPattern{
		ID:          "pattern-factory",
		Name:        "Factory / Creational Constructor Pattern",
		Category:    model.PatternCreational,
		Confidence:  conf,
		Description: "Centralizes object construction, initialization invariants, dependency injection, and configuration instantiation.",
		Evidence: []string{
			fmt.Sprintf("Discovered %d creational constructors (e.g. New*, Build*) across %d files", len(factorySyms), len(files)),
			fmt.Sprintf("Representative constructors: %s", strings.Join(factorySyms[:min(len(factorySyms), 4)], ", ")),
		},
		Files:     files[:min(len(files), 10)],
		Tradeoffs: "Enforces invariant validation upon creation and shields clients from internal struct configurations.",
	}, true
}

func detectSingletonPattern(files []model.FileNode, syms []model.Symbol) (model.DetectedPattern, bool) {
	var singletonEvidence []string
	var matchedFiles []string
	seenFiles := make(map[string]bool)

	for _, s := range syms {
		name := strings.ToLower(s.Name)
		if strings.Contains(name, "instance") || strings.Contains(name, "singleton") ||
			strings.HasPrefix(name, "default") || strings.HasPrefix(name, "global") {
			singletonEvidence = append(singletonEvidence, s.QualifiedName)
			if !seenFiles[s.File] && s.File != "" {
				seenFiles[s.File] = true
				matchedFiles = append(matchedFiles, s.File)
			}
		}
	}

	if len(singletonEvidence) == 0 {
		return model.DetectedPattern{}, false
	}

	conf := 0.60
	if len(singletonEvidence) >= 3 {
		conf = 0.85
	}

	return model.DetectedPattern{
		ID:          "pattern-singleton",
		Name:        "Singleton / Shared Registry Pattern",
		Category:    model.PatternCreational,
		Confidence:  conf,
		Description: "Ensures a component or service coordinates through a globally accessible shared instance or centralized coordinator.",
		Evidence: []string{
			fmt.Sprintf("Found %d global shared references or instance coordinators", len(singletonEvidence)),
			fmt.Sprintf("Sample instances: %s", strings.Join(singletonEvidence[:min(len(singletonEvidence), 3)], ", ")),
		},
		Files:     matchedFiles[:min(len(matchedFiles), 10)],
		Tradeoffs: "Ensures single point of coordination for shared resources; can introduce hidden global state and require careful test teardown.",
	}, true
}

func detectEventDrivenPattern(files []model.FileNode, syms []model.Symbol) (model.DetectedPattern, bool) {
	var eventFiles []string
	var eventSyms []string

	for _, f := range files {
		p := strings.ToLower(f.Path)
		if strings.Contains(p, "event") || strings.Contains(p, "queue") || strings.Contains(p, "worker") ||
			strings.Contains(p, "pubsub") || strings.Contains(p, "watcher") {
			eventFiles = append(eventFiles, f.Path)
		}
	}

	for _, s := range syms {
		name := strings.ToLower(s.Name)
		if strings.Contains(name, "event") || strings.Contains(name, "publish") ||
			strings.Contains(name, "subscribe") || strings.Contains(name, "emit") ||
			strings.Contains(name, "listener") || strings.Contains(name, "watcher") ||
			strings.Contains(name, "notify") {
			eventSyms = append(eventSyms, s.QualifiedName)
		}
	}

	if len(eventFiles) == 0 && len(eventSyms) < 2 {
		return model.DetectedPattern{}, false
	}

	conf := 0.70
	if len(eventFiles) >= 2 || len(eventSyms) >= 4 {
		conf = 0.90
	}

	return model.DetectedPattern{
		ID:          "pattern-event-driven",
		Name:        "Event-Driven / Publisher-Subscriber Pattern",
		Category:    model.PatternBehavioral,
		Confidence:  conf,
		Description: "Decouples producers and consumers through asynchronous message channels, event dispatchers, or observer notifications.",
		Evidence: []string{
			fmt.Sprintf("Identified %d event-handling/messaging file(s)", len(eventFiles)),
			fmt.Sprintf("Found %d pub/sub/notification symbols", len(eventSyms)),
		},
		Files:     eventFiles[:min(len(eventFiles), 10)],
		Tradeoffs: "Highly extensible and decoupled execution; tracing message flows across async queues requires correlation IDs.",
	}, true
}

func detectMiddlewareChainPattern(files []model.FileNode, syms []model.Symbol) (model.DetectedPattern, bool) {
	var mwFiles []string
	var mwSyms []string

	for _, f := range files {
		p := strings.ToLower(f.Path)
		if strings.Contains(p, "middleware") || strings.Contains(p, "interceptor") || strings.Contains(p, "filter") {
			mwFiles = append(mwFiles, f.Path)
		}
	}

	for _, s := range syms {
		name := strings.ToLower(s.Name)
		if strings.Contains(name, "middleware") || strings.Contains(name, "cors") ||
			strings.Contains(name, "recovery") || strings.Contains(name, "chain") ||
			strings.Contains(name, "interceptor") {
			mwSyms = append(mwSyms, s.QualifiedName)
		}
	}

	if len(mwFiles) == 0 && len(mwSyms) == 0 {
		return model.DetectedPattern{}, false
	}

	return model.DetectedPattern{
		ID:          "pattern-middleware",
		Name:        "Chain of Responsibility / Middleware Pattern",
		Category:    model.PatternBehavioral,
		Confidence:  0.88,
		Description: "Passes requests along an extensible chain of handlers (e.g. auth, logging, rate-limiting, error handling) before reaching business logic.",
		Evidence: []string{
			fmt.Sprintf("Detected %d middleware pipeline file(s) and %d interceptor functions", len(mwFiles), len(mwSyms)),
		},
		Files:     mwFiles[:min(len(mwFiles), 10)],
		Tradeoffs: "Clean separation of cross-cutting concerns; execution order must be kept strictly coordinated.",
	}, true
}

func detectCleanArchitecturePattern(files []model.FileNode, syms []model.Symbol, rels []model.Relationship) (model.DetectedPattern, bool) {
	hasDomain := false
	hasAPI := false
	hasStorage := false

	var domainFiles, apiFiles, storageFiles []string

	for _, f := range files {
		tier := detectFileTier(f.Path)
		switch tier {
		case model.TierService:
			hasDomain = true
			domainFiles = append(domainFiles, f.Path)
		case model.TierAPI:
			hasAPI = true
			apiFiles = append(apiFiles, f.Path)
		case model.TierStorage:
			hasStorage = true
			storageFiles = append(storageFiles, f.Path)
		}
	}

	if !(hasDomain && (hasAPI || hasStorage)) {
		return model.DetectedPattern{}, false
	}

	conf := 0.85
	if hasDomain && hasAPI && hasStorage {
		conf = 0.94
	}

	var allSampleFiles []string
	allSampleFiles = append(allSampleFiles, apiFiles[:min(len(apiFiles), 3)]...)
	allSampleFiles = append(allSampleFiles, domainFiles[:min(len(domainFiles), 3)]...)
	allSampleFiles = append(allSampleFiles, storageFiles[:min(len(storageFiles), 3)]...)

	return model.DetectedPattern{
		ID:          "pattern-clean-architecture",
		Name:        "Clean Architecture / Ports & Adapters (Hexagonal)",
		Category:    model.PatternArchitectural,
		Confidence:  conf,
		Description: "Enforces layered boundary separation where domain business rules remain isolated from external drivers, presentation controllers, and databases.",
		Evidence: []string{
			fmt.Sprintf("Discovered %d domain service files, %d presentation API files, and %d persistence storage files",
				len(domainFiles), len(apiFiles), len(storageFiles)),
			"Strict hierarchical layering observed across application domains",
		},
		Files:     allSampleFiles,
		Tradeoffs: "Exceptional testability and modularity; requires strict discipline to prevent layer leaking across package boundaries.",
	}, true
}

func detectCQRSPattern(files []model.FileNode, syms []model.Symbol) (model.DetectedPattern, bool) {
	var cmdSyms, querySyms []string

	for _, s := range syms {
		name := strings.ToLower(s.Name)
		if strings.Contains(name, "command") || strings.Contains(name, "mutate") || strings.Contains(name, "execute") {
			cmdSyms = append(cmdSyms, s.QualifiedName)
		}
		if strings.Contains(name, "query") || strings.Contains(name, "read") || strings.Contains(name, "fetch") || strings.Contains(name, "find") {
			querySyms = append(querySyms, s.QualifiedName)
		}
	}

	if len(cmdSyms) == 0 || len(querySyms) == 0 {
		return model.DetectedPattern{}, false
	}

	conf := 0.70
	if len(cmdSyms) >= 2 && len(querySyms) >= 2 {
		conf = 0.88
	}

	return model.DetectedPattern{
		ID:          "pattern-cqrs",
		Name:        "CQRS (Command Query Responsibility Segregation)",
		Category:    model.PatternArchitectural,
		Confidence:  conf,
		Description: "Segregates write operations (commands/mutations) from read operations (queries/lookups) into specialized execution paths.",
		Evidence: []string{
			fmt.Sprintf("Found %d write command symbols and %d read query symbols", len(cmdSyms), len(querySyms)),
		},
		Files:     nil,
		Tradeoffs: "Allows fine-grained optimization for high-throughput reads vs writes; introduces duplicate data structures.",
	}, true
}

func detectResiliencePattern(files []model.FileNode, syms []model.Symbol) (model.DetectedPattern, bool) {
	var resilienceSyms []string
	var matchedFiles []string
	seenFiles := make(map[string]bool)

	for _, s := range syms {
		name := strings.ToLower(s.Name)
		if strings.Contains(name, "retry") || strings.Contains(name, "backoff") ||
			strings.Contains(name, "timeout") || strings.Contains(name, "circuit") ||
			strings.Contains(name, "ratelimit") || strings.Contains(name, "throttle") {
			resilienceSyms = append(resilienceSyms, s.QualifiedName)
			if !seenFiles[s.File] && s.File != "" {
				seenFiles[s.File] = true
				matchedFiles = append(matchedFiles, s.File)
			}
		}
	}

	if len(resilienceSyms) == 0 {
		return model.DetectedPattern{}, false
	}

	conf := 0.75
	if len(resilienceSyms) >= 3 {
		conf = 0.90
	}

	return model.DetectedPattern{
		ID:          "pattern-resilience",
		Name:        "Resilience & Fault Tolerance (Circuit Breaker / Retry)",
		Category:    model.PatternStructural,
		Confidence:  conf,
		Description: "Safeguards system boundaries with retry backoffs, circuit breakers, rate limiters, or context timeout deadlines to prevent cascading service degradation.",
		Evidence: []string{
			fmt.Sprintf("Discovered %d fault tolerance and resilience symbols", len(resilienceSyms)),
			fmt.Sprintf("Key components: %s", strings.Join(resilienceSyms[:min(len(resilienceSyms), 4)], ", ")),
		},
		Files:     matchedFiles[:min(len(matchedFiles), 10)],
		Tradeoffs: "Prevents resource exhaustion under degraded dependencies; increases response latency when retrying failing downstream backends.",
	}, true
}

func detectAdapterPattern(files []model.FileNode, syms []model.Symbol) (model.DetectedPattern, bool) {
	var adapterFiles []string
	var adapterSyms []string

	for _, f := range files {
		p := strings.ToLower(f.Path)
		if strings.Contains(p, "adapter") || strings.Contains(p, "facade") ||
			strings.Contains(p, "wrapper") || strings.Contains(p, "proxy") {
			adapterFiles = append(adapterFiles, f.Path)
		}
	}

	for _, s := range syms {
		name := strings.ToLower(s.Name)
		if strings.HasSuffix(name, "adapter") || strings.HasSuffix(name, "facade") ||
			strings.HasSuffix(name, "wrapper") || strings.HasSuffix(name, "client") {
			adapterSyms = append(adapterSyms, s.QualifiedName)
		}
	}

	if len(adapterFiles) == 0 && len(adapterSyms) < 2 {
		return model.DetectedPattern{}, false
	}

	conf := 0.70
	if len(adapterFiles) >= 2 || len(adapterSyms) >= 4 {
		conf = 0.88
	}

	return model.DetectedPattern{
		ID:          "pattern-adapter",
		Name:        "Adapter / Facade Pattern",
		Category:    model.PatternStructural,
		Confidence:  conf,
		Description: "Converts interface contracts between incompatible third-party services or subsystems into unified internal domain APIs.",
		Evidence: []string{
			fmt.Sprintf("Found %d adapter/wrapper files and %d client adapter symbols", len(adapterFiles), len(adapterSyms)),
		},
		Files:     adapterFiles[:min(len(adapterFiles), 10)],
		Tradeoffs: "Shields domain core from vendor changes; introduces a lightweight translation layer.",
	}, true
}

func detectStrategyPattern(files []model.FileNode, syms []model.Symbol) (model.DetectedPattern, bool) {
	var strategySyms []string

	for _, s := range syms {
		name := strings.ToLower(s.Name)
		if strings.HasSuffix(name, "strategy") || strings.HasSuffix(name, "policy") ||
			strings.HasSuffix(name, "matcher") || strings.HasSuffix(name, "evaluator") ||
			strings.HasSuffix(name, "rule") {
			strategySyms = append(strategySyms, s.QualifiedName)
		}
	}

	if len(strategySyms) < 2 {
		return model.DetectedPattern{}, false
	}

	return model.DetectedPattern{
		ID:          "pattern-strategy",
		Name:        "Strategy / Policy Pattern",
		Category:    model.PatternBehavioral,
		Confidence:  0.86,
		Description: "Encapsulates swappable business algorithms and validation rules behind unified strategy contracts, enabling dynamic runtime switching.",
		Evidence: []string{
			fmt.Sprintf("Identified %d strategy/policy implementations", len(strategySyms)),
			fmt.Sprintf("Sample policies: %s", strings.Join(strategySyms[:min(len(strategySyms), 4)], ", ")),
		},
		Files:     nil,
		Tradeoffs: "Complies with Open-Closed Principle (OCP); requires client awareness of strategy selection parameters.",
	}, true
}

func detectObserverPattern(files []model.FileNode, syms []model.Symbol) (model.DetectedPattern, bool) {
	var observerSyms []string

	for _, s := range syms {
		name := strings.ToLower(s.Name)
		if strings.Contains(name, "observer") || strings.Contains(name, "listener") ||
			strings.Contains(name, "callback") || strings.Contains(name, "hook") {
			observerSyms = append(observerSyms, s.QualifiedName)
		}
	}

	if len(observerSyms) < 2 {
		return model.DetectedPattern{}, false
	}

	return model.DetectedPattern{
		ID:          "pattern-observer",
		Name:        "Observer / Listener Callback Pattern",
		Category:    model.PatternBehavioral,
		Confidence:  0.84,
		Description: "Defines one-to-many subscription mechanics where state changes in subject components automatically notify and update registered observers.",
		Evidence: []string{
			fmt.Sprintf("Discovered %d observer/listener callback hooks", len(observerSyms)),
		},
		Files:     nil,
		Tradeoffs: "Decouples event publishers from consumer reactions; requires correlation IDs for asynchronous sequence debugging.",
	}, true
}

func generateRecommendations(tiers []model.SystemTier, dataFlow []model.DataFlowHop, bottlenecks []string) []model.ArchRecommendation {
	recs := make([]model.ArchRecommendation, 0)

	// 1. Check for layer bypass (Client direct to Storage)
	for _, hop := range dataFlow {
		if hop.FromTier == model.TierClient && hop.ToTier == model.TierStorage {
			recs = append(recs, model.ArchRecommendation{
				ID:          "rec-decouple-client-storage",
				Title:       "Eliminate Direct Client-to-Storage Layer Bypass",
				Category:    "Decoupling",
				Severity:    model.SevHigh,
				Description: fmt.Sprintf("Observed %d direct queries from Client Tier to Storage Tier bypassing API and domain service boundaries.", hop.CallCount),
				ActionSteps: []string{
					"Route all client mutations and lookups through API presentation controllers",
					"Enforce domain authorization and input validation in application services",
					"Never expose raw SQL queries or database driver bindings to client views",
				},
				Impact: "Hardens application security, eliminates unauthorized DB access risks, and restores strict Clean Architecture boundary isolation.",
			})
			break
		}
	}

	// 2. Check for high storage query density (Caching recommendation)
	for _, hop := range dataFlow {
		if hop.ToTier == model.TierStorage && hop.CallCount >= 20 {
			recs = append(recs, model.ArchRecommendation{
				ID:          "rec-storage-cache",
				Title:       "Introduce In-Memory Cache for Heavy Storage Reads",
				Category:    "Caching",
				Severity:    model.SevMedium,
				Description: fmt.Sprintf("High query density observed between %s and Storage (%d cross-boundary calls).", hop.FromTier, hop.CallCount),
				ActionSteps: []string{
					"Deploy an in-memory Redis or LRU cache ahead of repetitive database read operations",
					"Adopt Cache-Aside pattern with TTL expiration for high-read, low-write entities",
					"Implement cache key invalidation hooks on storage write operations",
				},
				Impact: "Reduces database I/O pressure by up to 60-80% and cuts median request latency to single-digit milliseconds.",
			})
			break
		}
	}

	// 3. Check for high synchronous orchestration (Async offloading)
	for _, hop := range dataFlow {
		if hop.FromTier == model.TierAPI && hop.ToTier == model.TierService && hop.CallCount >= 30 {
			recs = append(recs, model.ArchRecommendation{
				ID:          "rec-async-offloading",
				Title:       "Offload Intensive Orchestration to Asynchronous Queue Workers",
				Category:    "Scalability",
				Severity:    model.SevMedium,
				Description: fmt.Sprintf("API layer performs %d synchronous orchestration calls in the request path.", hop.CallCount),
				ActionSteps: []string{
					"Identify non-critical operations (email alerts, metrics aggregation, PDF generation, re-indexing)",
					"Publish background events to the Async Messaging tier via lightweight channel workers",
					"Return HTTP 202 Accepted with status polling for long-running workflows",
				},
				Impact: "Keeps HTTP request handlers non-blocking and prevents API gateway timeouts under spike traffic.",
			})
			break
		}
	}

	// 4. Check for monolithic tier concentration
	for _, t := range tiers {
		if t.ID == model.TierService && t.FileCount >= 25 {
			recs = append(recs, model.ArchRecommendation{
				ID:          "rec-service-modularity",
				Title:       "Decompose Application Services into Bounded Contexts",
				Category:    "Modularity",
				Severity:    model.SevLow,
				Description: fmt.Sprintf("Application & Domain Services tier currently aggregates %d files into a single tier domain.", t.FileCount),
				ActionSteps: []string{
					"Partition large service modules into explicit DDD sub-domains (e.g. Identity, Billing, Engine)",
					"Define clean Go interface contracts between distinct bounded domains",
					"Use package-private visibility to prevent cross-domain leaking",
				},
				Impact: "Improves code readability, accelerates unit test suite velocity, and facilitates parallel engineering workflows.",
			})
			break
		}
	}

	// 5. If no issues found, provide general resilience guideline
	if len(recs) == 0 {
		recs = append(recs, model.ArchRecommendation{
			ID:          "rec-resilience-monitoring",
			Title:       "Maintain Boundary Discipline & Instrument Distributed Traces",
			Category:    "Resilience",
			Severity:    model.SevLow,
			Description: "Codebase demonstrates clean architectural layering with no critical layer bypasses detected.",
			ActionSteps: []string{
				"Maintain CI/CD quality gate checks (codedna check) on every pull request",
				"Propagate correlation trace IDs across all inter-tier service hops",
			},
			Impact: "Sustains long-term maintainability and guarantees zero architectural regressions over time.",
		})
	}

	return recs
}

func detectKeyEntrypoints(files []model.FileNode, syms []model.Symbol) []string {
	var entrypoints []string

	for _, f := range files {
		p := filepath.ToSlash(f.Path)
		if strings.HasSuffix(p, "main.go") || strings.HasSuffix(p, "index.ts") ||
			strings.HasSuffix(p, "server.go") || strings.HasSuffix(p, "app.ts") ||
			strings.HasSuffix(p, "main.py") {
			entrypoints = append(entrypoints, f.Path)
		}
	}

	for _, s := range syms {
		if s.Name == "main" || s.Name == "StartServer" || s.Name == "NewServer" || s.Name == "Execute" {
			entrypoints = append(entrypoints, fmt.Sprintf("%s (%s)", s.QualifiedName, s.File))
		}
	}

	if len(entrypoints) > 8 {
		entrypoints = entrypoints[:8]
	}
	return entrypoints
}

func detectBottlenecks(tiers []model.SystemTier, dataFlow []model.DataFlowHop, rels []model.Relationship) []string {
	var bottlenecks []string

	// 1. High call density hops
	for _, hop := range dataFlow {
		if hop.CallCount > 15 {
			bottlenecks = append(bottlenecks,
				fmt.Sprintf("High Traffic Flow: %s ➔ %s (%d cross-boundary calls)", hop.FromTier, hop.ToTier, hop.CallCount))
		}
	}

	// 2. Storage layer direct access check (layer bypassing)
	for _, hop := range dataFlow {
		if hop.FromTier == model.TierClient && hop.ToTier == model.TierStorage {
			bottlenecks = append(bottlenecks,
				"Architectural Risk: Client tier directly references Persistence tier, bypassing API and Service validation boundaries")
		}
	}

	// 3. Heavy tier concentration
	for _, t := range tiers {
		if t.FileCount > 50 {
			bottlenecks = append(bottlenecks,
				fmt.Sprintf("Monolithic Tier Warning: %s contains %d files; consider decomposing into sub-domains", t.Name, t.FileCount))
		}
	}

	if len(bottlenecks) == 0 {
		bottlenecks = append(bottlenecks, "No critical architectural bottlenecks or layer violations detected.")
	}

	return bottlenecks
}

func inferArchitectureStyle(tiers []model.SystemTier, patterns []model.DetectedPattern, dataFlow []model.DataFlowHop) string {
	hasMessaging := false
	hasClean := false

	for _, t := range tiers {
		if t.ID == model.TierMessaging && t.FileCount > 0 {
			hasMessaging = true
		}
	}

	for _, p := range patterns {
		if p.ID == "pattern-clean-architecture" {
			hasClean = true
		}
	}

	if hasMessaging && len(dataFlow) > 3 {
		return "Event-Driven & Multi-Tier Layered Architecture"
	}
	if hasClean {
		return "Clean Modular Monolith (Layered Hexagonal Architecture)"
	}
	return "Multi-Tier Application Architecture"
}

func generateDesignSummary(style string, tiers []model.SystemTier, patterns []model.DetectedPattern, dataFlow []model.DataFlowHop) string {
	activeTiers := 0
	for _, t := range tiers {
		if t.FileCount > 0 {
			activeTiers++
		}
	}

	return fmt.Sprintf("Classified codebase into %s across %d active structural tiers with %d cross-tier communication flows and %d recognized software design patterns.",
		style, activeTiers, len(dataFlow), len(patterns))
}

func calculateReliabilityScore(tiers []model.SystemTier, patterns []model.DetectedPattern, bottlenecks []string, dataFlow []model.DataFlowHop) float64 {
	score := 85.0

	// Bonus for recognized clean patterns
	for _, p := range patterns {
		if p.ID == "pattern-clean-architecture" {
			score += 5.0
		}
		if p.ID == "pattern-repository" {
			score += 3.0
		}
	}

	// Penalty for bypasses or warnings
	for _, b := range bottlenecks {
		if strings.HasPrefix(b, "Architectural Risk") {
			score -= 15.0
		}
		if strings.HasPrefix(b, "Monolithic Tier") {
			score -= 5.0
		}
	}

	if score > 100.0 {
		score = 100.0
	}
	if score < 10.0 {
		score = 10.0
	}

	return score
}
