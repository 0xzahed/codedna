# CodeDNA Task & Feature Tracker 🚀

This document tracks all features, bug fixes, performance optimizations, and architectural tasks for CodeDNA. Every task follows the **Feature-Wise / Task-Wise Commit Protocol**:
1. Single logical unit of work per commit.
2. Conventional commit format (`feat:`, `fix:`, `perf:`, `docs:`, `test:`).
3. Full verification (`go test`, static build, UI check) before commit.
4. Update task status here (`[x]`) and push to GitHub `main`.

---

## 🏁 Completed Milestones & Tasks

### Milestone 1: Platform Foundation & Core Intelligence (v2.0.0)
- [x] **AST Parsing Engine:** Tree-sitter parsers for Go, Python, TypeScript, JavaScript, Java, PHP.
- [x] **Symbol & Call Resolution:** Cross-file call resolution with confidence metrics.
- [x] **Graph Analysis Algorithms:** Tarjan's SCC for cycle detection, BFS reachability.
- [x] **Git Diff Impact Analyzer:** Blast-radius propagation and modified symbol mapping.
- [x] **Architecture Guard:** Layer boundary rules and drift detection.
- [x] **Local Storage Engine:** SQLite persistence with migrations and zero external database dependencies.
- [x] **Next.js Web Dashboard:** React Flow dependency canvas, dark-first UI, and 9 core views.

### Milestone 2: Dynamic Repositories & Engine Optimization (v2.1.0)
- [x] **Task 2.1 — Multi-Core Parallel Parsing Engine:**
  - Implemented 16-worker pool with channel pipeline (`internal/engine/engine.go`).
  - Added fast skip lists for build/minified assets (`dist`, `static_cdn`, `staticfiles`, `*.min.js`).
  - *Benchmark:* Reduced 639-file scan time from **124.95s to 5.44s** (**22x speedup**).
  - *Commit:* `6aa8867` (`perf(engine): 22x parallel AST parsing, real-time SSE scan progress`)

- [x] **Task 2.2 — Any Local Path or Remote GitHub URL Importer:**
  - Added `scanner.ResolveTarget()` to clone GitHub repos or resolve local filesystem paths.
  - Added "Open Codebase" modal with path validation in UI header.
  - *Commit:* `892d073` (`feat(scanner): add ResolveTarget for any local path or remote GitHub URL`)

- [x] **Task 2.3 — Dynamic Multi-Repo Source Code Inspector:**
  - Added `LoadRepoRootForScan(scanID)` and `FindRepoRootForFile(path)` fallback to resolve external codebase root paths.
  - Fixed 404 error when clicking files in external repositories (`GET /api/scans/:id/file`).
  - Added smooth auto-scroll to highlighted symbol lines in `SourceDrawer`.
  - *Commit:* `a8fadb3` (`feat: multi-repo workspace linking and dynamic per-scan source code resolution`)

- [x] **Task 2.4 — Multi-Repository Workspace & Cross-Service API Architecture:**
  - Created `internal/workspace/` engine to scan API endpoints (Django, FastAPI, Express, Next.js, Go).
  - Fuzzy and prefix route matching between frontend API calls (`fetch`, `axios`) and backend endpoints.
  - "Connect Repos" modal to compose full-stack system architecture maps.
  - *Commit:* `a8fadb3` (`feat: multi-repo workspace linking`)

- [x] **Task 2.5 — File-to-File Dependency Resolution & Architectural Cluster Graph:**
  - Connected file-to-file call aggregations and direct import links (`imports`, `calls`).
  - Directory-based hierarchical architectural layout in React Flow.
  - Directional arrows and colored edge labels (Violet: Imports, Green: Calls, Sky Blue: Cross-Service API).
  - Rendered 729+ active relationship edges on real codebases.
  - *Commit:* `73ef2a0` (`feat: complete file-to-file dependency resolution, import links, and architectural cluster graph rendering`)

- [x] **Task 2.6 — GitHub Remote Repository & CI Integration:**
  - Created public GitHub repository [0xzahed/codedna](https://github.com/0xzahed/codedna).
  - Pushed all 9 feature-wise commits with conventional messages.

---

## 📌 Active & Upcoming Roadmap Tasks

### Milestone 3: Real-Time Visual Analytics & Documentation Excellence (v2.2.0)
- [x] **Task 3.1 — Interactive Graph Export & Shareable Snapshots:**
  - Export system architecture map as Mermaid.js flowchart (`flowchart TD`) and C4 Component diagrams.
  - Added REST endpoint `GET /api/scans/:id/export?format=mermaid|c4`.
  - Added Export Modal in Graph View with instant one-click copy and `.mmd`/`.c4` download.
  - *Commit:* `3c1b28b` (`feat(graph): add Mermaid diagram and C4 architecture export engine with UI snapshot modal`)

- [x] **Task 3.2 — Production Build Automation & Developer Experience:**
  - Unified `Makefile` supporting `build-cli`, `build-web`, `test`, `test-race`, `lint`, `clean`, and `run`.
  - Clean binary outputs and automated Go compiler flag optimization (`-ldflags="-s -w"`).
  - *Commit:* `23dd560` (`build: add production Makefile for unified Go CLI and Next.js dashboard lifecycle`)

- [x] **Task 3.3 — World-Class Ventox-Styled Documentation Overhaul:**
  - Full redesign of `README.md` following [0xzahed/ventox](https://github.com/0xzahed/ventox) aesthetic and structure.
  - Glowing shields.io badges (Go, TS, React Flow, SQLite, License, Release, Privacy-first).
  - Visual Mermaid System Architecture dataflow diagram.
  - Comprehensive feature comparison matrix (CodeDNA vs Sourcegraph, SonarQube, CodeSee).
  - Terminal ASCII execution preview with step-by-step progress ribbons.
  - Complete CLI command & flags matrix for all 8 CLI workflows.
  - Deep walkthrough of the 9 specialized Web Intelligence Dashboard views.
  - Real-world 22x parallel throughput benchmarks and hardware memory metrics.
  - Synchronized `CONTRIBUTING.md`, `SECURITY.md`, and version bumped to `2.2.0`.
  - *Commits:* `4130bf9`, `429afcc`, `45131fc`, `7a2e512`, `8a505c1`, `28418ef`, `60389a5`, `be6b9bd`

---

## 📌 Active & Upcoming Roadmap Tasks

### Milestone 4: Real-Time Live Watcher & AI Architecture Refactoring (v2.3.0)
- [x] **Task 4.1 — Real-Time Live Hot-Reloading Watcher & Incremental AST Re-Indexing:**
  - Automated file modification listener with code-file filtering (`.go`, `.ts`, `.py`, `.java`, etc.).
  - Automatic incremental background re-indexing with synchronization locks and store updates.
  - Rich SSE event stream broadcasting `files_changed` with file list, duration, symbol count, and scan ID.
  - Real-time frontend standing listener with pulsing live notification pill and reactive tab re-rendering.
- [x] **Task 4.2 — Custom Architectural Rules Editor in UI & Dynamic Layer Enforcement:**
  - Added rule persistence engine (`SaveRules`, `AddRule`, `DeleteRule`, `ResetRules`) targeting `.codedna/architecture.json`.
  - Added REST endpoints: `GET /api/scans/:id/architecture`, `POST /api/scans/:id/architecture/rules`, `DELETE /api/scans/:id/architecture/rules/:ruleId`, `POST /api/scans/:id/architecture/rules/reset`.
  - Interactive Rule Builder in `ArchitectureView.tsx` with presets (Clean Architecture, Domain Isolation, UI Purity), glob patterns, and live re-scoring.
  - Visualized Tarjan SCC circular dependency clusters and layer health penalty impact.
- [x] **Task 4.3 — AI-Powered Architecture Refactoring Suggestions:**
  - Automated refactoring synthesis engine (`internal/ai/SuggestRefactoring`) proposing concrete decoupling steps for circular dependencies and layer violations.
  - Added REST endpoint: `GET /api/scans/:id/refactor` returning structured decoupling plans with architectural pattern blueprints.
  - Integrated "AI Decoupling Advisor" in `ArchitectureView.tsx` with pattern blueprints, step-by-step checklists, and one-click code copying.

---

## 📌 Active & Upcoming Roadmap Tasks

### Milestone 5: Deep Intelligence & Multi-Language Evolution (v2.4.0)
- [x] **Task 5.2 — Git Blame & Author Knowledge Map / Bus Factor Analysis:**
  - High-performance git history & numstat parser in `internal/blame/blame.go`.
  - Calculates line ownership percentage, primary author per file, and Bus Factor risk levels (`high`, `medium`, `low`).
  - Added REST endpoint: `GET /api/scans/:id/knowledge` (and `/api/scans/:id/blame`).
  - Added CLI command: `codedna blame [--repo <path>] [--json]`.
  - Added Team Knowledge & Bus Factor Analysis dashboard card in `OverviewView.tsx`.
- [x] **Task 5.3 — CI/CD Architecture Guard & Automated Quality Gate (`codedna check`):**
  - Standalone quality gate CLI command `codedna check [--min-health=80] [--max-cycles=0] [--max-violations=0] [--json]`.
  - Non-zero exit code on threshold failure with clear terminal diagnostics for failing builds.
  - Added `make check` target in unified Makefile for automated CI/CD pipeline gating.
- [x] **Task 5.4 — System Design & Multi-Tier Architecture Topology Engine:**
  - Automated tier classification into 6 structural tiers (`TierClient`, `TierAPI`, `TierService`, `TierStorage`, `TierMessaging`, `TierIntegration`).
  - Cross-tier data flow hop reconstruction from AST call graph edges with protocol detection (`HTTPS / REST API`, `In-Memory Orchestration`, `SQL / Storage Interface`, `Async Event / Channel Pipeline`).
  - Gang of Four & modern distributed pattern detection: Clean Architecture/Ports & Adapters, Repository/Data Mapper, Factory/Builder, Singleton, Event-Driven/Pub-Sub, Chain of Responsibility/Middleware, CQRS.
  - Architectural bottleneck hotspot detection and reliability score calculation.
  - Dedicated CLI command: `codedna design [--repo <path>] [--json]` with ASCII tiered topology.
  - REST endpoints: `GET /api/scans/:id/system-design` and `GET /api/scans/:id/patterns`.
  - Dedicated 10th Core Intelligence View: `SystemDesignView.tsx` (`/design`) with interactive tier inspection, hop flow pipeline, and filterable pattern recognition gallery.
  - Overview dashboard summary card with direct navigation link.
- [x] **Task 5.5 — System Design Deep Optimization Engine & Interactive Architecture Graph:**
  - Extended pattern recognition engine with Strategy/Policy, Adapter/Facade, Observer/Listener, and Resilience/Circuit Breaker patterns.
  - Concrete Actionable Architectural Recommendations engine (`ArchRecommendation`: Decoupling, Caching, Scalability, Modularity) with specific implementation steps and architectural payoff metrics.
  - Terminal CLI (`codedna design`) printing structured optimization recommendations, severity ratings, and impact summaries.
  - Enhanced React Flow Architecture Graph with in-graph component search, match glow effects, and multi-protocol edge filtering (HTTPS, In-Memory, SQL/Storage, Async Event).
  - Interactive Architectural Optimization Engine in web dashboard with category/severity filters, tracking checklists, and impact analysis.
- [x] **Task 5.6 — Function Call Graph Overhaul & Interactive Call Hierarchy Inspector:**
  - Solved single-column 20,000px vertical stacking issue by introducing **Call Flow DAG layout** with topological depth leveling (Left ➔ Right: Callers ➔ Orchestrators ➔ Services ➔ Leaf Callees).
  - Included struct methods alongside standalone functions in the function filter and tracked active `inDegree` / `outDegree` in API.
  - Added **Interactive Call Hierarchy Focus & Trace Mode**: clicking any function highlights its direct callers in Emerald Green and direct callees in Sky Blue, dimming unrelated nodes.
  - Built **Call Hierarchy Inspector Drawer** detailing callers, callees, complexity, and line locations with one-click focal hopping.
  - Integrated full **Embedded Code Editor Viewer modal** with syntax highlighting and direct line jump from any graph node.
  - Added **Package / Module isolation dropdown filter** and **Connected Only toggle** to eliminate visual clutter.
- [x] **Task 5.7 — Neon Lighting Wires & Sub-Column Call Graph Layout Optimization:**
  - Engineered **NeonLightingEdge** custom React Flow edge component rendering luminous ambient glow, core electric wire, and live animated electrical light pulse particles traveling from Caller to Callee at 60fps.
  - Implemented **Sub-Column DAG Wrapping** (max 7 rows per sub-column) preventing 30,000px vertical stretching and keeping caller-callee pairs tightly grouped on screen.
  - Added **Prod Only / Hide Tests toggle** and **Language filter (Go, TypeScript)** to separate backend execution flow from test suites and frontend components.
  - Added **Isolate Call Subgraph mode**: one-click isolated 3-tier view (Callers ➔ Target ➔ Callees) for distraction-free function analysis.
  - Added custom SVG glowing neon arrow markers (Emerald, Cyan, Purple, White).
- [x] **Task 5.8 — Widescreen Canvas Optimization & Collapsible Sidebar Mode:**
  - Expanded main view content containers across all pages (`SystemDesignView`, `ArchitectureView`, `OverviewView`, `DependencyView`, `ImpactView`, `RuntimeView`, `AIView`) from 1280px (`max-w-7xl`) to 1920px (`max-w-[1920px] w-full`).
  - Added collapsible sidebar toggle with icon-only mode (`w-16` vs `w-64`), yielding +200px of additional horizontal space for diagrams.
  - Increased `ArchitectureDiagramGraph` canvas height to 850px and integrated one-click Fullscreen expansion mode.
- [ ] **Task 5.1 — Rust & C++ AST Tree-Sitter Grammars:**
  - Expand parser coverage to Rust (`.rs`) and C/C++ (`.c`, `.cpp`, `.h`) with trait and macro resolution.

---

## 🛠️ Feature Commit Protocol
Every subsequent change must follow this cycle:
```bash
# 1. Implement feature / task
# 2. Run unit tests & static build
go test -v -race ./...
cd web && npm run build
# 3. Commit task-wise with conventional format
git add <files>
git commit -m "<type>(<scope>): <clear descriptive summary>"
# 4. Push to remote
git push origin main
```
