// Package model defines the core data structures shared across CodeDNA engines.
package model

import (
	"time"
)

// SymbolKind classifies code entities discovered by the parser.
type SymbolKind string

const (
	SymbolFunction  SymbolKind = "function"
	SymbolMethod    SymbolKind = "method"
	SymbolClass     SymbolKind = "class"
	SymbolInterface SymbolKind = "interface"
	SymbolType      SymbolKind = "type"
	SymbolStruct    SymbolKind = "struct"
	SymbolVariable  SymbolKind = "variable"
)

// RelationshipKind describes how two entities relate.
type RelationshipKind string

const (
	RelCalls      RelationshipKind = "calls"
	RelImports    RelationshipKind = "imports"
	RelDefines    RelationshipKind = "defines"
	RelInstantiates RelationshipKind = "instantiates"
)

// ResolutionMethod explains how a relationship was inferred. Confidence is
// never faked: dynamic or unresolved calls stay low-confidence.
type ResolutionMethod string

const (
	MethodExactLocal      ResolutionMethod = "exact-local"      // callee defined in the same file
	MethodExactModule     ResolutionMethod = "exact-module"     // callee defined elsewhere in the same package/module
	MethodExactUnique     ResolutionMethod = "exact-unique"     // single global match for the callee name
	MethodQualified       ResolutionMethod = "module-qualified" // pkg.Function / module.Class resolved via imports
	MethodSelfMethod      ResolutionMethod = "self-method"      // method call resolved on the enclosing class/receiver
	MethodConstructor     ResolutionMethod = "constructor"      // new Obj() resolved to a discovered constructor
)

// Symbol is a code entity discovered by static parsing.
type Symbol struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	QualifiedName   string     `json:"qualifiedName"`
	Kind            SymbolKind `json:"kind"`
	File            string     `json:"file"`
	Language        string     `json:"language"`
	Line            int        `json:"line"`            // 1-based
	Column          int        `json:"column"`          // 0-based byte offset
	EndLine         int        `json:"endLine"`
	EndColumn       int        `json:"endColumn"`
	Exported        bool       `json:"exported"`
	Complexity      int        `json:"complexity"`
	ParameterCount  int        `json:"parameterCount"`
	LineCount       int        `json:"lineCount"`
	ParentClassName string     `json:"parentClassName,omitempty"` // for methods

	// Internal byte ranges used during parsing (never serialized).
	ParentStartByte int `json:"-"`
	ParentEndByte   int `json:"-"`
}

// Ref points at a source location that produced a relationship.
type Ref struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
}

// Relationship is a directed edge in the dependency graph.
type Relationship struct {
	Kind        RelationshipKind   `json:"kind"`
	FromID      string             `json:"fromId"`
	ToID        string             `json:"toId"`   // resolved point
	ToName      string             `json:"toName"` // raw reference (always populated, even when unresolved)
	ImportPath  string             `json:"importPath,omitempty"`
	Confidence  float64            `json:"confidence"` // 0.0-1.0
	Method      ResolutionMethod   `json:"method"`
	Locations   []Ref              `json:"locations"`
}

// FileNode is a file participating in the graph.
type FileNode struct {
	Path     string   `json:"path"`
	Language string   `json:"language"`
	Size     int64    `json:"size"`
	LOC      int      `json:"loc"`
	Hash     string   `json:"hash"`
	Tags     []string `json:"tags,omitempty"` // e.g. "entrypoint", "test", "config"
}

// ParseError records a file that could not be fully parsed.
type ParseError struct {
	File      string `json:"file"`
	Language  string `json:"language"`
	Message   string `json:"message"`
	Recovered bool   `json:"recovered"`
}

// Import is a single import statement found in a file.
type Import struct {
	Path      string `json:"path"`     // module/package path as written
	Alias     string `json:"alias,omitempty"`
	Names     []string `json:"names,omitempty"` // from-import entries / named imports
	IsStdLib  bool   `json:"isStdLib"`
	Line      int    `json:"line"`
	Col       int    `json:"col"`
}

// Call is a call site found in a file.
type Call struct {
	Callee     string `json:"callee"` // raw callee expression text, e.g. "fs.readFile" or "process"
	Receiver   string `json:"receiver,omitempty"` // object/operand part, e.g. "fs"
	IsNew      bool   `json:"isNew,omitempty"`    // constructor invocation
	Line       int    `json:"line"`
	Col        int    `json:"col"`
	InSymbolID string `json:"inSymbolId,omitempty"` // enclosing symbol when known
}

// FileParseResult is the parser output for a single file.
type FileParseResult struct {
	File      string       `json:"file"`
	Language  string       `json:"language"`
	Symbols   []Symbol     `json:"symbols"`
	Imports   []Import     `json:"imports"`
	Calls     []Call       `json:"calls"`
	BranchCount int        `json:"branchCount"` // loop/branch occurrences, complexity input
	ParseErrors []ParseError `json:"parseErrors,omitempty"`
}

// GitMeta holds repository metadata extracted via git.
type GitMeta struct {
	IsGitRepo     bool      `json:"isGitRepo"`
	RemoteURL     string    `json:"remoteUrl,omitempty"`
	Branch        string    `json:"branch,omitempty"`
	Commit        string    `json:"commit,omitempty"`
	AuthorName    string    `json:"authorName,omitempty"`
	AuthorEmail   string    `json:"authorEmail,omitempty"`
	LastCommitAt  *time.Time `json:"lastCommitAt,omitempty"`
	CommitCount   int64     `json:"commitCount"`
	TrackedFiles  int64     `json:"trackedFiles"`
	IsDirty       bool      `json:"isDirty"`
}

// RepoScanStats is an aggregate summary of a scan.
type RepoScanStats struct {
	FilesScanned     int               `json:"filesScanned"`
	FilesParsed      int               `json:"filesParsed"`
	FilesSkipped     int               `json:"filesSkipped"`
	ParseErrors      int               `json:"parseErrors"`
	SymbolsFound     int               `json:"symbolsFound"`
	Relationships    int               `json:"relationships"`
	ResolvedCalls    int               `json:"resolvedCalls"`
	UnresolvedCalls  int               `json:"unresolvedCalls"`
	LOC              int               `json:"loc"`
	Languages        map[string]int    `json:"languages"`           // language -> file count
	LOCByLanguage    map[string]int    `json:"locByLanguage,omitempty"`
	Frameworks       []string          `json:"frameworks,omitempty"`
	DirectDeps       map[string][]string `json:"directDeps,omitempty"` // ecosystem -> dependency list
	Duration         string            `json:"duration"`
}

// Issue kinds discovered by the analysis engines.
type IssueKind string

const (
	IssueCircularDep IssueKind = "circular-dependency"
	IssueDeadCode         IssueKind = "dead-code"
	IssueComplexity       IssueKind = "complexity-hotspot"
	IssueLayerViolation   IssueKind = "layer-violation"
	IssueEntryPoint       IssueKind = "entry-point"
)

type Severity string

const (
	SevInfo     Severity = "info"
	SevLow      Severity = "low"
	SevMedium   Severity = "medium"
	SevHigh     Severity = "high"
)

// Issue is a finding (cycle, dead code candidate, hotspot, ...).
type Issue struct {
	Kind     IssueKind    `json:"kind"`
	Severity Severity     `json:"severity"`
	Title    string       `json:"title"`
	File     string       `json:"file,omitempty"`
	SymbolID string       `json:"symbolId,omitempty"`
	Detail   interface{}  `json:"detail"` // typed payload per kind
}

// CycleDetail describes a circular dependency between modules.
type CycleDetail struct {
	Files []string `json:"files"`
	Chain string   `json:"chain"`
}

// RepoInfo identifies a scanned repository.
type RepoInfo struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Root         string     `json:"root"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastScan     *time.Time `json:"lastScan,omitempty"`
	LatestScanID int64      `json:"latestScanId,omitempty"`
	Git          GitMeta    `json:"git"`
}

// Scan is one execution of the scan+analyze pipeline.
type Scan struct {
	ID        int64      `json:"id"`
	RepoID    int64      `json:"repoId"`
	StartedAt time.Time  `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Status    string     `json:"status"` // running | complete | failed
	Stats     *RepoScanStats `json:"stats,omitempty"`
}

// GraphNode is a node formatted for UI graph renderers like React Flow.
type GraphNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "file", "function", "class", "module", "endpoint", "model"
	Label      string                 `json:"label"`
	File       string                 `json:"file,omitempty"`
	Line       int                    `json:"line,omitempty"`
	Complexity int                    `json:"complexity,omitempty"`
	Exported   bool                   `json:"exported,omitempty"`
	Language   string                 `json:"language,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// GraphEdge is an edge formatted for UI graph renderers.
type GraphEdge struct {
	ID                string  `json:"id"`
	Source            string  `json:"source"`
	Target            string  `json:"target"`
	Kind              string  `json:"kind"` // "calls", "imports", "defines", "instantiates", "runtime_called"
	Confidence        float64 `json:"confidence"`
	Method            string  `json:"method,omitempty"`
	Locations         []Ref   `json:"locations,omitempty"`
	IsRuntimeObserved bool    `json:"isRuntimeObserved,omitempty"`
}

// GraphData is the complete graph payload for UI rendering.
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// RuntimeTraceEvent is an execution event captured at runtime.
type RuntimeTraceEvent struct {
	ID             int64     `json:"id,omitempty"`
	RunID          string    `json:"runId"`
	FunctionName   string    `json:"functionName"`
	File           string    `json:"file"`
	Line           int       `json:"line"`
	DurationUs     int64     `json:"durationUs"`
	TimestampNs    int64     `json:"timestampNs"`
	CallerFunction string    `json:"callerFunction,omitempty"`
	CallerFile     string    `json:"callerFile,omitempty"`
	HTTPPath       string    `json:"httpPath,omitempty"`
	HTTPMethod     string    `json:"httpMethod,omitempty"`
	Error          string    `json:"error,omitempty"`
}

// RuntimeSummary aggregates observed runtime behavior.
type RuntimeSummary struct {
	TotalEvents       int                 `json:"totalEvents"`
	UniqueFunctions   int                 `json:"uniqueFunctions"`
	FrequentFunctions []FunctionFrequency `json:"frequentFunctions"`
	ExecutionPaths    []ExecutionPath     `json:"executionPaths"`
	AvgDurationUs     float64             `json:"avgDurationUs"`
	RuntimeEdges      []GraphEdge         `json:"runtimeEdges"`
	Comparison        RuntimeComparison   `json:"comparison"`
}

type FunctionFrequency struct {
	Function    string  `json:"function"`
	File        string  `json:"file"`
	CallCount   int     `json:"callCount"`
	TotalTimeUs int64   `json:"totalTimeUs"`
	AvgTimeUs   float64 `json:"avgTimeUs"`
}

type ExecutionPath struct {
	TraceID   string   `json:"traceId"`
	Root      string   `json:"root"`
	Steps     []string `json:"steps"`
	TotalTime int64    `json:"totalTimeUs"`
}

type RuntimeComparison struct {
	StaticEdgeCount       int     `json:"staticEdgeCount"`
	RuntimeEdgeCount      int     `json:"runtimeEdgeCount"`
	ObservedStaticEdges   int     `json:"observedStaticEdges"`
	UnexecutedStaticEdges int     `json:"unexecutedStaticEdges"`
	DynamicOnlyEdges      int     `json:"dynamicOnlyEdges"`
	CoveragePercentage    float64 `json:"coveragePercentage"`
}

// Change Impact models
type ChangedSymbol struct {
	Symbol       Symbol `json:"symbol"`
	ChangeType   string `json:"changeType"` // "modified", "added", "deleted"
	ChangedLines []int  `json:"changedLines"`
}

type ImpactReport struct {
	BaseRef           string          `json:"baseRef"`
	TargetRef         string          `json:"targetRef"`
	ChangedFiles      []string        `json:"changedFiles"`
	ChangedSymbols    []ChangedSymbol `json:"changedSymbols"`
	UpstreamAffected  []string        `json:"upstreamAffected"`  // Files or symbol IDs that depend on changes
	DownstreamCalls   []string        `json:"downstreamCalls"`   // Files or symbol IDs called by changes
	AffectedEndpoints []string        `json:"affectedEndpoints"` // API routes impacted
	AffectedTests     []string        `json:"affectedTests"`     // Tests that exercise changed code
	RiskScore         float64         `json:"riskScore"`         // 0 - 100
	RiskLevel         Severity        `json:"riskLevel"`         // Low, Medium, High
	Summary           string          `json:"summary"`
	CreatedAt         time.Time       `json:"createdAt"`
}

// Architecture Rule models
type ArchRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	FromPattern string   `json:"fromPattern"` // glob pattern e.g. "controllers/**"
	ToPattern   string   `json:"toPattern"`   // glob pattern e.g. "repositories/**"
	Forbidden   bool     `json:"forbidden"`   // true if from is NOT allowed to import/call to
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
}

type ArchViolation struct {
	RuleID     string   `json:"ruleId"`
	RuleName   string   `json:"ruleName"`
	Severity   Severity `json:"severity"`
	FromFile   string   `json:"fromFile"`
	ToFile     string   `json:"toFile"`
	ImportPath string   `json:"importPath,omitempty"`
	Reason     string   `json:"reason"`
	Line       int      `json:"line,omitempty"`
}

type ArchHealthReport struct {
	Violations       []ArchViolation `json:"violations"`
	CircularCount    int             `json:"circularCount"`
	Cycles           []CycleDetail   `json:"cycles"`
	LayerHealthScore float64         `json:"layerHealthScore"` // 0-100
	RuleCount        int             `json:"ruleCount"`
	Rules            []ArchRule      `json:"rules,omitempty"`
}

// AI Context & Assistance models
type AIQueryRequest struct {
	Query       string `json:"query"`
	Provider    string `json:"provider,omitempty"` // "ollama", "openai", "anthropic"
	Model       string `json:"model,omitempty"`
	UserConsent bool   `json:"userConsent"` // Explicit privacy confirmation
}

type AIQueryResponse struct {
	Answer     string  `json:"answer"`
	SourceRefs []Ref   `json:"sourceRefs,omitempty"`
	Confidence float64 `json:"confidence"`
	Provider   string  `json:"provider"`
}

type BusinessLogicModule struct {
	Domain      string   `json:"domain"`
	Description string   `json:"description"`
	Files       []string `json:"files"`
	KeySymbols  []string `json:"keySymbols"`
	Confidence  float64  `json:"confidence"`
}

// Workspace models for multi-repository / cross-service intelligence
type Workspace struct {
	ID        int64              `json:"id"`
	Name      string             `json:"name"`
	RepoIDs   []int64            `json:"repoIds"`
	Repos     []RepoInfo         `json:"repos,omitempty"`
	CreatedAt time.Time          `json:"createdAt"`
	Links     []CrossServiceLink `json:"links,omitempty"`
}

type CrossServiceLink struct {
	ID             string  `json:"id"`
	SourceRepoID   int64   `json:"sourceRepoId"`
	SourceRepoName string  `json:"sourceRepoName"`
	SourceFile     string  `json:"sourceFile"`
	SourceSymbol   string  `json:"sourceSymbol,omitempty"`
	TargetRepoID   int64   `json:"targetRepoId"`
	TargetRepoName string  `json:"targetRepoName"`
	TargetFile     string  `json:"targetFile"`
	TargetSymbol   string  `json:"targetSymbol,omitempty"`
	HTTPMethod     string  `json:"httpMethod"`     // "GET", "POST", "PUT", "DELETE"
	RoutePath      string  `json:"routePath"`      // e.g. "/api/v1/courses"
	Confidence     float64 `json:"confidence"`
	Description    string  `json:"description"`
}

type APIEndpoint struct {
	RepoID   int64  `json:"repoId"`
	RepoName string `json:"repoName"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Handler  string `json:"handler"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type APICallSite struct {
	RepoID   int64  `json:"repoId"`
	RepoName string `json:"repoName"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Caller   string `json:"caller"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// RefactoringProposal outlines concrete decoupling steps for architectural violations and cycles.
type RefactoringProposal struct {
	ID              string   `json:"id"`
	IssueType       string   `json:"issueType"` // "circular-dependency", "layer-violation"
	Target          string   `json:"target"`
	Severity        Severity `json:"severity"`
	DecouplingSteps []string `json:"decouplingSteps"`
	Pattern         string   `json:"pattern"` // e.g. "Dependency Inversion", "Shared Abstraction", "Service Orchestration"
	ProposedFix     string   `json:"proposedFix"`
	Explanation     string   `json:"explanation"`
}

// Git Knowledge & Bus Factor models
type AuthorStats struct {
	Name             string  `json:"name"`
	Email            string  `json:"email"`
	Commits          int     `json:"commits"`
	LinesAdded       int     `json:"linesAdded"`
	LinesDeleted     int     `json:"linesDeleted"`
	FilesTouched     int     `json:"filesTouched"`
	PrimaryFileCount int     `json:"primaryFileCount"`
	KnowledgeShare   float64 `json:"knowledgeShare"` // 0 - 100%
}

type ContributorShare struct {
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	LinesAdded int     `json:"linesAdded"`
	Percentage float64 `json:"percentage"`
}

type FileOwnership struct {
	FilePath          string             `json:"filePath"`
	PrimaryAuthor     string             `json:"primaryAuthor"`
	ContributorsCount int                `json:"contributorsCount"`
	TopContributors   []ContributorShare `json:"topContributors"`
	BusFactorRisk     Severity           `json:"busFactorRisk"` // high, medium, low
}

type RepoKnowledgeMap struct {
	IsGitRepo       bool            `json:"isGitRepo"`
	TotalAuthors    int             `json:"totalAuthors"`
	TotalCommits    int             `json:"totalCommits"`
	HighRiskFiles   int             `json:"highRiskFiles"` // Bus factor = 1
	BusFactorScore  float64         `json:"busFactorScore"` // 0 - 100
	TopContributors []AuthorStats   `json:"topContributors"`
	Files           []FileOwnership `json:"files"`
}

// System Design & Architecture Flow models
type TierKind string

const (
	TierClient      TierKind = "client_gateway"
	TierAPI         TierKind = "api_presentation"
	TierService     TierKind = "domain_service"
	TierStorage     TierKind = "persistence_storage"
	TierMessaging   TierKind = "async_messaging"
	TierIntegration TierKind = "external_integration"
)

type SystemTier struct {
	ID          TierKind `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Order       int      `json:"order"`
	FileCount   int      `json:"fileCount"`
	SymbolCount int      `json:"symbolCount"`
	Files       []string `json:"files,omitempty"`
	KeySymbols  []string `json:"keySymbols,omitempty"`
}

type DataFlowHop struct {
	FromTier    TierKind `json:"fromTier"`
	ToTier      TierKind `json:"toTier"`
	Protocol    string   `json:"protocol"` // "HTTP", "In-Memory Call", "SQL / Query", "Async Event"
	CallCount   int      `json:"callCount"`
	SampleCalls []string `json:"sampleCalls,omitempty"`
}

type PatternCategory string

const (
	PatternArchitectural PatternCategory = "Architectural"
	PatternStructural    PatternCategory = "Structural"
	PatternCreational    PatternCategory = "Creational"
	PatternBehavioral    PatternCategory = "Behavioral"
)

type DetectedPattern struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Category    PatternCategory `json:"category"`
	Confidence  float64         `json:"confidence"` // 0.0 - 1.0
	Description string          `json:"description"`
	Evidence    []string        `json:"evidence"`
	Files       []string        `json:"files"`
	Tradeoffs   string          `json:"tradeoffs"`
}

type ArchRecommendation struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Category    string   `json:"category"` // "Decoupling", "Caching", "Resilience", "Scalability", "Modularity"
	Severity    Severity `json:"severity"` // "high", "medium", "low"
	Description string   `json:"description"`
	ActionSteps []string `json:"actionSteps"`
	Impact      string   `json:"impact"`
}

type SystemDesignReport struct {
	ArchitectureStyle  string               `json:"architectureStyle"`
	Summary            string               `json:"summary"`
	Tiers              []SystemTier         `json:"tiers"`
	DataFlow           []DataFlowHop        `json:"dataFlow"`
	DetectedPatterns   []DetectedPattern    `json:"detectedPatterns"`
	KeyEntrypoints     []string             `json:"keyEntrypoints"`
	BottleneckHotspots []string             `json:"bottleneckHotspots"`
	Recommendations    []ArchRecommendation `json:"recommendations"`
	ReliabilityScore   float64              `json:"reliabilityScore"` // 0 - 100
}

