# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
