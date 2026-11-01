// CodeDNA 2.0 CLI
// A Living Intelligence Map of Software
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/codedna/codedna/internal/ai"
	"github.com/codedna/codedna/internal/api"
	"github.com/codedna/codedna/internal/engine"
	"github.com/codedna/codedna/internal/impact"
	"github.com/codedna/codedna/internal/model"
	"github.com/codedna/codedna/internal/report"
	"github.com/codedna/codedna/internal/runtime"
	"github.com/codedna/codedna/internal/scanner"
	"github.com/codedna/codedna/internal/store"
	"github.com/codedna/codedna/internal/watcher"
)

const version = "2.2.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "scan":
		runScan(args, false)
	case "analyze":
		runScan(args, true)
	case "graph":
		runGraph(args)
	case "impact":
		runImpact(args)
	case "report":
		runReport(args)
	case "serve":
		runServe(args)
	case "watch":
		runWatch(args)
	case "runtime":
		runRuntime(args)
	case "version", "--version", "-v":
		fmt.Printf("CodeDNA version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'codedna help' for usage.\n", command)
		os.Exit(1)
	}
}

func printUsage() {
	banner := `
  ____          _      ____  _   _    _      ____     ___  
 / ___|___   __| | ___|  _ \| \ | |  / \    |___ \   / _ \ 
| |   / _ \ / _| |/ _ \ | | |  \| | / _ \     __) | | | | |
| |__| (_) | (_| |  __/ |_| | |\  |/ ___ \   / __/ _| |_| |
 \____\___/ \__,_|\___|____/|_| \_/_/   \_\ |_____(_)\___/ 
`
	fmt.Print(banner)
	fmt.Println("CodeDNA: Understand Your Code. See Its DNA.")
	fmt.Println("\nUsage: codedna <command> [arguments] [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  scan <path>           Walks and parses a codebase into the intelligence store")
	fmt.Println("  analyze <path>        Full scan plus deep architectural & dependency analysis")
	fmt.Println("  graph <path>          Outputs nodes and dependency relationships")
	fmt.Println("  impact [path]         Analyzes blast radius and consequences of git changes")
	fmt.Println("  report [path]         Generates comprehensive Markdown or JSON reports")
	fmt.Println("  serve                 Starts REST API server and interactive web dashboard")
	fmt.Println("  watch <path>          Watches directory for changes and live-updates index")
	fmt.Println("  runtime enable        Configures opt-in runtime instrumentation adapters")
	fmt.Println("  version               Prints version information")
	fmt.Println("\nRun 'codedna <command> --help' for details on a specific command.")
}

func getStore() (*store.Store, error) {
	dbPath := store.DefaultPath()
	return store.Open(dbPath)
}

func runScan(args []string, deep bool) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "Output results in JSON format")
	fs.Parse(args)

	target := "."
	if fs.NArg() > 0 {
		target = fs.Arg(0)
	}

	st, err := getStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening store: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	if !*jsonOut {
		fmt.Printf("\n[CodeDNA] Scanning %s ...\n", target)
	}

	progress := func(done, total int, path string) {
		if !*jsonOut {
			fmt.Printf("\r  -> Parsed %d files [%s]                  ", done, filepath.Base(path))
		}
	}

	res, err := engine.Run(target, st, progress)
	if err != nil {
		if !*jsonOut {
			fmt.Printf("\n")
		}
		fmt.Fprintf(os.Stderr, "\nScan error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOut {
		b, _ := json.MarshalIndent(res.Stats, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Printf("\n\n✓ Scan Complete [%s]\n", res.Stats.Duration)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Files Parsed:       %d\n", res.Stats.FilesParsed)
	fmt.Printf("Total LOC:          %d\n", res.Stats.LOC)
	fmt.Printf("Discovered Symbols: %d\n", res.Stats.SymbolsFound)
	fmt.Printf("Relationships:      %d (resolved: %d)\n", res.Stats.Relationships, res.Stats.ResolvedCalls)

	if len(res.Stats.Frameworks) > 0 {
		fmt.Printf("Frameworks:         %s\n", strings.Join(res.Stats.Frameworks, ", "))
	}

	if len(res.Stats.Languages) > 0 {
		fmt.Println("\nLanguages:")
		for lang, count := range res.Stats.Languages {
			fmt.Printf("  • %-12s: %d files (%d LOC)\n", lang, count, res.Stats.LOCByLanguage[lang])
		}
	}

	if deep {
		fmt.Println("\nDeep Analysis Findings:")
		cycles := 0
		hotspots := 0
		dead := 0
		for _, iss := range res.Issues {
			switch iss.Kind {
			case model.IssueCircularDep:
				cycles++
			case model.IssueComplexity:
				hotspots++
			case model.IssueDeadCode:
				dead++
			}
		}
		fmt.Printf("  • Circular Dependencies: %d\n", cycles)
		fmt.Printf("  • Complexity Hotspots:   %d\n", hotspots)
		fmt.Printf("  • Dead Code Candidates:  %d\n", dead)
		fmt.Printf("  • Layer Health Score:    %.1f / 100\n", res.Architecture.LayerHealthScore)
	}
	fmt.Println("--------------------------------------------------")
	fmt.Println("Run 'codedna serve' to explore the interactive DNA graph in your browser.")
}

func runGraph(args []string) {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)
	format := fs.String("format", "summary", "Output format: summary, json")
	fs.Parse(args)

	target := "."
	if fs.NArg() > 0 {
		target = fs.Arg(0)
	}

	st, err := getStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening store: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	res, err := engine.Run(target, st, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Engine run failed: %v\n", err)
		os.Exit(1)
	}

	if *format == "json" {
		data := model.GraphData{}
		for _, s := range res.Symbols {
			data.Nodes = append(data.Nodes, model.GraphNode{
				ID:         s.ID,
				Type:       string(s.Kind),
				Label:      s.Name,
				File:       s.File,
				Line:       s.Line,
				Complexity: s.Complexity,
			})
		}
		for i, r := range res.Rels {
			data.Edges = append(data.Edges, model.GraphEdge{
				ID:         fmt.Sprintf("e-%d", i),
				Source:     r.FromID,
				Target:     r.ToID,
				Kind:       string(r.Kind),
				Confidence: r.Confidence,
				Method:     string(r.Method),
			})
		}
		b, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Printf("\n[CodeDNA Graph: %s]\n", res.Repo.Name)
	fmt.Printf("Nodes: %d symbols across %d files\n", len(res.Symbols), len(res.Files))
	fmt.Printf("Edges: %d call & dependency relationships\n\n", len(res.Rels))

	fmt.Println("Top Central Connected Symbols:")
	edgeCounts := make(map[string]int)
	for _, r := range res.Rels {
		edgeCounts[r.FromID]++
		edgeCounts[r.ToID]++
	}
	count := 0
	for _, s := range res.Symbols {
		if ec := edgeCounts[s.ID]; ec > 0 {
			fmt.Printf("  • %-30s in %s:%d (degree: %d, complexity: %d)\n", s.QualifiedName, s.File, s.Line, ec, s.Complexity)
			count++
			if count >= 10 {
				break
			}
		}
	}
}

func runImpact(args []string) {
	fs := flag.NewFlagSet("impact", flag.ExitOnError)
	base := fs.String("base", "main", "Base branch or git commit to diff against")
	jsonOut := fs.Bool("json", false, "Output impact report in JSON")
	fs.Parse(args)

	target := "."
	if fs.NArg() > 0 {
		target = fs.Arg(0)
	}

	st, err := getStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Store error: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	res, err := engine.Run(target, st, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Engine run failed: %v\n", err)
		os.Exit(1)
	}

	diffs, err := impact.ParseGitDiff(target, *base)
	if err != nil {
		diffs, _ = impact.ParseGitDiff(target, "")
	}

	rep := impact.AnalyzeImpact(target, *base, diffs, res.Index, res.Files)

	if *jsonOut {
		b, _ := json.MarshalIndent(rep, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Printf("\n[CodeDNA Change Impact Analysis vs '%s']\n", *base)
	fmt.Println("==================================================")
	fmt.Printf("Risk Score:         %.1f / 100 (%s)\n", rep.RiskScore, rep.RiskLevel)
	fmt.Printf("Changed Files:      %d\n", len(rep.ChangedFiles))
	fmt.Printf("Modified Symbols:   %d\n", len(rep.ChangedSymbols))
	fmt.Printf("Upstream Blast:     %d affected caller(s)\n", len(rep.UpstreamAffected))
	fmt.Printf("Downstream Calls:   %d callee(s)\n", len(rep.DownstreamCalls))
	fmt.Printf("Affected APIs:      %d\n", len(rep.AffectedEndpoints))
	fmt.Printf("Relevant Tests:     %d\n", len(rep.AffectedTests))
	fmt.Println("--------------------------------------------------")

	if len(rep.ChangedSymbols) > 0 {
		fmt.Println("Modified Code Entities:")
		for _, cs := range rep.ChangedSymbols {
			fmt.Printf("  • [%s] %s in %s:%d\n", cs.ChangeType, cs.Symbol.QualifiedName, cs.Symbol.File, cs.Symbol.Line)
		}
	}

	if len(rep.AffectedEndpoints) > 0 {
		fmt.Println("\nPotentially Affected API Endpoints:")
		for _, ep := range rep.AffectedEndpoints {
			fmt.Printf("  ⚠ %s\n", ep)
		}
	}

	if len(rep.AffectedTests) > 0 {
		fmt.Println("\nRecommended Tests to Run:")
		for _, t := range rep.AffectedTests {
			fmt.Printf("  ✓ %s\n", t)
		}
	}
	fmt.Println("==================================================")
}

func runReport(args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	format := fs.String("format", "markdown", "Report format: markdown, json")
	outFile := fs.String("out", "", "Output file path (default stdout)")
	fs.Parse(args)

	target := "."
	if fs.NArg() > 0 {
		target = fs.Arg(0)
	}

	st, err := getStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Store error: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	res, err := engine.Run(target, st, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Analysis error: %v\n", err)
		os.Exit(1)
	}

	data := report.CompleteReportData{
		Repo:         res.Repo,
		Scan:         res.Scan,
		Stats:        res.Stats,
		Files:        res.Files,
		TopSymbols:   res.Symbols,
		Issues:       res.Issues,
		Architecture: &res.Architecture,
		GeneratedAt:  time.Now(),
	}
	data.HealthScore = report.CalculateHealthScore(res.Stats, res.Issues, &res.Architecture)

	var outputText string
	if *format == "json" {
		b, _ := report.GenerateJSON(data)
		outputText = string(b)
	} else {
		outputText = report.GenerateMarkdown(data)
	}

	if *outFile != "" {
		if err := os.WriteFile(*outFile, []byte(outputText), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write report to %s: %v\n", *outFile, err)
			os.Exit(1)
		}
		fmt.Printf("✓ Report saved to %s\n", *outFile)
	} else {
		fmt.Println(outputText)
	}
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "Port for the HTTP server")
	repo := fs.String("repo", ".", "Repository root path to serve")
	fs.Parse(args)

	st, err := getStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Store error: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	resolvedRepo, err := scanner.ResolveTarget(*repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Target resolution error: %v\n", err)
		os.Exit(1)
	}

	// Initial scan of target repository
	fmt.Printf("[CodeDNA] Initializing intelligence store for %s ...\n", resolvedRepo)
	res, err := engine.Run(resolvedRepo, st, nil)
	if err != nil {
		fmt.Printf("Warning: initial scan had notices: %v\n", err)
	} else {
		fmt.Printf("✓ Initial scan ready: %d symbols, %d files parsed\n", res.Stats.SymbolsFound, res.Stats.FilesParsed)
	}

	// Determine web dashboard assets
	var webFS http.FileSystem
	for _, dir := range []string{"web/out", "../web/out", "/codedna-app/web/out"} {
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			webFS = http.Dir(dir)
			break
		}
	}

	server := api.NewServer(api.Config{
		Store:    st,
		AIConfig: ai.DefaultConfig(),
		RepoRoot: resolvedRepo,
		WebFS:    webFS,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server.StartWatcher(ctx)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Println("==================================================")
	fmt.Printf("  CodeDNA 2.0 Server Active\n")
	fmt.Printf("  Web Dashboard:   http://localhost:%d\n", *port)
	fmt.Printf("  REST API Base:   http://localhost:%d/api\n", *port)
	fmt.Printf("  Live SSE Stream: http://localhost:%d/api/events\n", *port)
	fmt.Println("==================================================")
	fmt.Println("Press Ctrl+C to terminate.")

	httpServer := &http.Server{
		Addr:    addr,
		Handler: server.Handler(),
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nShutting down CodeDNA server...")
		shutdownCtx, sCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer sCancel()
		httpServer.Shutdown(shutdownCtx)
		cancel()
	}()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func runWatch(args []string) {
	target := "."
	if len(args) > 0 {
		target = args[0]
	}

	st, err := getStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Store error: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	fmt.Printf("[CodeDNA Watch] Monitoring %s for code changes...\n", target)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := watcher.New(target, 800*time.Millisecond, func(changed []string) {
		fmt.Printf("[%s] Change detected in %d file(s):\n", time.Now().Format("15:04:05"), len(changed))
		for _, f := range changed {
			fmt.Printf("   • %s\n", f)
		}
		fmt.Print("   Re-indexing ... ")
		res, err := engine.Run(target, st, nil)
		if err != nil {
			fmt.Printf("failed: %v\n", err)
		} else {
			fmt.Printf("done (%d symbols, %d edges)\n", res.Stats.SymbolsFound, res.Stats.Relationships)
		}
	})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go w.Start(ctx)

	<-sigCh
	fmt.Println("\nStopped watching.")
}

func runRuntime(args []string) {
	if len(args) == 0 || args[0] != "enable" {
		fmt.Println("Usage: codedna runtime enable [--lang python|node|go]")
		return
	}

	fs := flag.NewFlagSet("runtime", flag.ExitOnError)
	lang := fs.String("lang", "python", "Programming language: python, node, go")
	fs.Parse(args[1:])

	snippet := runtime.TracerSnippet(*lang)
	outDir := filepath.Join(".codedna", "tracer")
	os.MkdirAll(outDir, 0o755)

	ext := "py"
	if *lang == "node" || *lang == "javascript" || *lang == "typescript" {
		ext = "js"
	} else if *lang == "go" {
		ext = "go"
	}
	outFile := filepath.Join(outDir, "codedna_tracer."+ext)

	if err := os.WriteFile(outFile, []byte(snippet), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write tracer file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ CodeDNA Runtime Tracer Adapter created at %s\n", outFile)
	fmt.Println("\nInstructions:")
	fmt.Println("1. Include or import this adapter in your local test runner or dev entrypoint.")
	fmt.Println("2. Run 'codedna serve' in another terminal to listen on http://localhost:8080.")
	fmt.Println("3. Execute your tests or application to record real-time function paths and durations.")
	fmt.Println("4. Open the Web Dashboard > Runtime Timeline to compare static versus runtime behavior.")
}
