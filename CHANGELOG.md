# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.3.0] - 2026-11-02

### Added
- **AI Architecture Refactoring Advisor**: Automated architectural decoupling proposal engine (`GET /api/scans/:id/refactor`) generating concrete decoupling plans, step-by-step checklists, and code blueprints for cyclic dependencies and layer boundary violations.
- **Interactive Refactoring UI**: Integrated Advisor panel in Architecture View with pattern blueprints (Dependency Inversion, Ports & Adapters, Shared Contracts) and one-click blueprint copying.
- **Real-Time Live Watcher & Incremental Re-Indexing**: Background directory file watcher that filters code-file modifications, triggers synchronized background re-indexing, and streams rich SSE events (`files_changed`).
- **Reactive Web Live-Reload**: Persistent SSE connection in the web dashboard with instant notification pill (`⚡ Live: X file(s) re-indexed`) and reactive canvas/overview re-render without page refresh.
- **Custom Architectural Rule Builder**: Interactive UI and REST API to create, customize, and enforce architectural boundary rules (e.g. `Domain ↛ Controllers`, `Controllers ↛ Store`) directly from the Web Dashboard.
- **Dynamic Rule Persistence**: Full persistence engine saving rules to `.codedna/architecture.json` with fallback to clean architecture defaults (`GET/POST/DELETE /api/scans/:id/architecture/rules`).
- **Live Layer Health Re-scoring**: Instant re-evaluation of layer health scores, boundary violations, and rule penalties upon saving or deleting constraints.
- **Circular Dependency Diagnostics**: Visual breakdown of Tarjan SCC cyclic graph clusters and offending dependency chains in Architecture View.

## [2.2.0] - 2026-11-01

### Added
- **Mermaid.js Flowchart & C4 Component Exporter**: Instant one-click export and download of system architecture topology as Mermaid (`flowchart TD`) and C4 Component diagrams (`GET /api/scans/:id/export`).
- **Graph Snapshot Modal**: Visual UI modal in React Flow canvas to preview, copy to clipboard, or save `.mmd` and `.c4` diagram specifications.
- **Production Build Automation (`Makefile`)**: Unified lifecycle targets (`build-cli`, `build-web`, `test-race`, `lint`, `clean`, `run`).
- **Ventox-Quality Documentation**: Complete architectural overhaul of `README.md` inspired by the world-class presentation of [0xzahed/ventox](https://github.com/0xzahed/ventox).

### Documentation
- **Mermaid System Architecture Flowchart**: End-to-end dataflow visualization from multi-language files to 16-worker AST pool, SQLite store, and React Flow UI.
- **Competitive Feature Comparison Matrix**: Detailed side-by-side comparison of CodeDNA vs Sourcegraph, SonarQube, and CodeSee.
- **Comprehensive CLI Reference & Flags Matrix**: Complete command manual covering `scan`, `analyze`, `graph`, `impact`, `report`, `serve`, `watch`, and `runtime`.
- **9-View Web Dashboard Walkthrough**: Exhaustive guides for Overview, Living Graph, Repo Explorer, Dependencies, Runtime Timeline, Impact Blast Radius, Architectural Guardrails, AI Assistant, and Multi-Repo Settings.
- **Parallel Parsing Performance Benchmarks**: Real-world benchmarks demonstrating 22x speedup and sub-45MB memory footprint.

## [2.1.0] - 2026-11-01

### Added
- **Multi-Repository Workspace Linking**: Connect separate repositories (e.g., Next.js frontend + Django/FastAPI/Express backend) into a unified cross-service system architecture map.
- **Cross-Service HTTP API Discovery**: Automatic discovery of backend endpoints (`urls.py`, FastAPI routers, Express routes) and frontend API consumer calls (`fetch`, `axios`).
- **File-to-File Dependency Flow**: File import and call relationships with directional arrows and color-coded edge badges (Violet: Imports, Green: Calls, Sky Blue: Cross-Service API).
- **Architectural Directory Clustering**: Hierarchical React Flow canvas layout grouping files by module/folder.
- **Dynamic Codebase Scanner**: "Open Codebase" modal supporting arbitrary local filesystem paths and remote GitHub URLs (`https://github.com/...`).
- **Dynamic Source Code Viewer**: Per-scan repository root path resolution with smooth auto-scroll to highlighted symbol lines in `SourceDrawer`.
- **Live SSE Scan Progress Stream**: Real-time progress bar displaying file count, percentage, and current filename during indexing.

### Performance
- **Parallel AST Parsing Engine**: 16-worker pipeline reducing large repository scan times from 124s to 5.4s (22x speedup).
- **High-Density Node Limiting**: Intelligent connectivity-based node prioritization keeping browser canvas fluid at 60 FPS.

## [2.0.0] - 2026-11-01

### Added
- **Multi-language Tree-sitter Parser Engine**: High-fidelity AST parsing for Go, Python, TypeScript, JavaScript, Java, and PHP.
- **Static Code Intelligence**: Cross-file call resolution with confidence metrics and exact resolution methods (`exact-local`, `exact-module`, `exact-unique`, `module-qualified`, `self-method`).
- **Circular Dependency Detection**: Tarjan's Strongly Connected Components algorithm for identifying module and file cycles.
- **Reachability Analysis**: Upstream and downstream breadth-first search dependency tracing.
- **Change Impact Analyzer**: Git diff integration mapping modified line ranges to AST symbols, blast-radius propagation, affected API identification, affected test discovery, and risk scoring.
- **Architecture Boundaries & Drift Detection**: Configurable architectural layer rules and violation reporting.
- **Runtime DNA**: Opt-in runtime tracing, execution path reconstruction, frequent function profiling, and static vs observed runtime graph comparison.
- **AI Codebase Intelligence**: Context generation for Claude Code (`CLAUDE.md`), Cursor (`.cursorrules`), and Codex, business logic mapping, and privacy-first Q&A.
- **Executive Intelligence Reporting**: Full Markdown and JSON software intelligence reports.
- **CodeDNA CLI (`codedna`)**: `scan`, `analyze`, `graph`, `impact`, `report`, `serve`, `watch`, and `runtime enable`.
- **Next.js & React Flow Interactive Dashboard**: Dark-first UI featuring 9 complete operational pages with zero mock data.
- **Local-first SQLite Storage & Migration Engine**.
