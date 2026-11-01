<div align="center">

# 🧬 CodeDNA

### Living Software Intelligence, Dynamic Call Graphs & Architectural Topology Platform in Go
#### The Open-Source Codebase Cartography Engine for Modern Engineering Teams

[![CI](https://github.com/0xzahed/codedna/actions/workflows/ci.yml/badge.svg)](https://github.com/0xzahed/codedna/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/0xzahed/codedna)](https://goreportcard.com/report/github.com/0xzahed/codedna)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/0xzahed/codedna)](https://golang.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.5-3178C6?logo=typescript)](https://www.typescriptlang.org/)
[![React Flow](https://img.shields.io/badge/React_Flow-11.11-FF0072?logo=react)](https://reactflow.dev/)
[![SQLite](https://img.shields.io/badge/SQLite-3-003B57?logo=sqlite)](https://www.sqlite.org/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Release](https://img.shields.io/github/v/release/0xzahed/codedna?include_prereleases)](https://github.com/0xzahed/codedna/releases)

<p align="center">
  <b>CodeDNA</b> is a lightning-fast, privacy-first software intelligence platform written in Go that constructs an interactive, living topological map of complex codebases. It combines <b>Tree-sitter AST parsing</b>, <b>cross-file call resolution</b>, <b>Tarjan's cycle detection</b>, <b>Git diff blast-radius propagation</b>, <b>architectural layer drift enforcement</b>, and a standalone <b>React Flow Web Intelligence Suite</b> with zero cloud dependencies.
</p>

[⚡ Quick Start](#-quick-start) •
[🎯 The Problem](#-the-problem) •
[⚡ The CodeDNA Engine](#-the-codedna-engine-architecture) •
[📊 Feature Comparison](#-feature-comparison) •
[🚀 Terminal Preview](#-terminal-preview) •
[✨ Features](#-features) •
[🖥️ Web Dashboard](#-web-dashboard-intelligence-suite) •
[💻 CLI Reference](#-cli-commands--flag-matrix) •
[🏗️ Architecture](#-system-architecture) •
[🚀 Benchmarks](#-performance-benchmarks) •
[📦 Installation](#-installation) •
[🛠️ Project Structure](#-project-structure) •
[🔒 Security & Privacy](#-security-local-first-guarantee--privacy) •
[🤝 Contributing](#-contributing) •
[📄 License](#-license)

</div>

---

## 🎯 The Problem

Modern software codebases grow rapidly in scale and intricacy, making architectural comprehension, refactoring, and code review perilous:
1. **Monolithic Coupling & Invisible Spaghetti:** In large monorepos or multi-service architectures, components become deeply tangled through cross-file calls, transitive imports, and circular dependencies that traditional linters and grep cannot visualize.
2. **Blind Refactoring & Unpredicted Blast Radius:** Changing a signature, refactoring an internal service, or modifying database schemas often triggers cascading regressions across distant upstream callers or downstream consumers with zero warning during code review.
3. **Architectural Erosion & Layer Drift:** Even when engineering teams establish Clean Architecture, Domain-Driven Design (DDD), or Hexagonal boundaries on paper, codebases steadily erode as developers inadvertently import presentation/database models into pure domain layers.
4. **Cloud Privacy & Security Exposure:** Existing codebase visualizers and AI analysis platforms mandate uploading proprietary source code to external SaaS servers, creating severe compliance, regulatory, and intellectual property liabilities.

---

## ⚡ The CodeDNA Engine Architecture

CodeDNA provides an autonomous, multi-dimensional codebase intelligence engine engineered from first principles for extreme speed, precision, and zero-trust privacy:

### 1. High-Throughput Multi-Core AST Ingestion (22x Parallel Pipeline)
Powered by native **Tree-sitter C bindings**, CodeDNA deploys an adaptive 16-worker pool with streaming channel pipelines across **Go, Python, TypeScript, JavaScript, Java, and PHP**. It extracts AST symbols (functions, classes, interfaces, methods, imports, and calls) with exact line/column byte offsets in milliseconds while automatically filtering non-source build artifacts (`dist`, `node_modules`, `staticfiles`, minified bundles).

### 2. Cross-File Call & Import Resolution
CodeDNA resolves symbolic references across disparate files with explicit confidence scoring (`exact-local`, `exact-module`, `exact-unique`, `module-qualified`, `self-method`). It tracks both direct AST imports and semantic invocation calls, enabling precise directory cluster grouping and directional dependency graphs.

### 3. Tarjan Strongly Connected Components (SCC) Cycle Resolver
Circular dependencies degrade modularity, prevent clean code decoupling, and induce race conditions. CodeDNA implements **Tarjan’s linear-time SCC algorithm** to discover cyclical dependency clusters and calculate BFS reachability trees for any given component.

### 4. Git Diff Blast-Radius Propagation Engine
Integrated directly with your Git workflow, CodeDNA compares your working tree or branch against any base commit (`main`, `HEAD~1`, `origin/main`). It maps modified lines to exact AST nodes and propagates upstream/downstream impact paths, calculating a quantified **Risk Score (0–100)** and pinpointing affected API endpoints and candidate tests.

### 5. Architectural Guardrail & Layer Drift Detection
Define clean architectural contracts in `.codedna/architecture.json` (e.g. `domain` must never depend on `infra` or `controller`). CodeDNA evaluates all resolved dependencies against these rules, pinpoints boundary violations with line-accurate call chains, and computes a system-wide **Layer Health Score (0–100%)**.

### 6. Multi-Repository Workspace & Cross-Service API Linker
Microservice and full-stack systems span multiple repositories. CodeDNA scans frontend API invocations (`fetch`, `axios`) and backend route definitions (FastAPI, Django, Express, Next.js, Gin, Spring Boot), using route normalization and prefix matching to weave individual repositories into a unified full-stack architecture diagram.

### 7. Native Mermaid & C4 Architecture Export
Transform living codebase topology into standard **Mermaid.js Flowcharts** (`flowchart TD`) and **C4 Component Diagrams** with a single click or CLI flag, perfect for markdown documentation, PR descriptions, and architectural RFCs.

---

## 📊 Feature Comparison

| Capability | Sourcegraph | SonarQube | CodeSee | **CodeDNA** |
| :--- | :---: | :---: | :---: | :---: |
| **100% Local-First & Zero Cloud Upload** | ❌ (Cloud/Server) | ❌ (Server setup) | ❌ (SaaS cloud) | **✅ (100% Offline / Local)** |
| **Interactive Living Topology Canvas** | ❌ (Code search only)| ❌ (Static tables) | ✅ (Web app) | **✅ (React Flow + Hierarchical)** |
| **Tree-sitter AST Multi-Language Parsing**| Partial | Partial (AST proprietary)| Proprietary | **✅ (Go, Py, TS, JS, Java, PHP)** |
| **Tarjan Circular Dependency Detection** | ❌ | ⚠️ (Limited metrics) | ❌ | **✅ (Tarjan SCC + Cycle Chains)** |
| **Git Diff Blast-Radius Risk Scoring** | ❌ | ❌ | Partial | **✅ (AST-mapped Risk 0–100)** |
| **Architectural Layer Drift Guardrails** | ❌ | ⚠️ (Config-heavy) | ❌ | **✅ (Declarative Layer Rules)** |
| **Multi-Repo Cross-Service API Linking** | ❌ | ❌ | ❌ | **✅ (Frontend/Backend Stitching)** |
| **Mermaid & C4 Architectural Export** | ❌ | ❌ | ❌ | **✅ (One-click MMD & C4)** |
| **Database Requirement** | Postgres / Heavy DB | PostgreSQL | Cloud Hosted | **Zero (Built-in Embedded SQLite)** |
| **Resource Footprint** | Heavy (>2 GB RAM) | Heavy (>1 GB RAM) | Browser-heavy | **Ultra-Lightweight (<45 MB RAM)** |

---

## 🌟 Key Features

### 1. Multi-Language Repository Scanner
- Parses **Go, Python, TypeScript, JavaScript, Java, and PHP** using **Tree-sitter** C bindings.
- Automatic Git metadata detection (commit hashes, authors, branches, uncommitted status with credential redaction).
- Deep dependency discovery across manifests (`package.json`, `requirements.txt`, `pyproject.toml`, `go.mod`, `composer.json`, `pom.xml`, `build.gradle`).
- Framework stack identification (Next.js, FastAPI, Express, Django, Gin, Spring Boot, etc.).

### 2. Static Code Intelligence & Dependency Graph
- AST extraction of functions, methods, classes, interfaces, imports, and calls.
- Cross-file call resolution with explicit confidence scoring (`exact-local`, `exact-module`, `exact-unique`, `module-qualified`, `self-method`).
- Source location preservation: exact file, line, and column offsets.
- **Circular Dependency Detection**: Implements Tarjan's Strongly Connected Components algorithm.
- **Reachability Analysis**: Breadth-First Search (BFS) upstream and downstream dependency tracing.
- Cyclomatic and cognitive complexity hotspot indicators.
- Dead code candidate identification.

### 3. Interactive Living DNA Graph
- Powered by React Flow with zoom, pan, search, filter, and minimap controls.
- Multiple layout engines (Hierarchical DAG layout and Grid layout).
- Click any node to open the **Source Drawer** and inspect the actual code file and highlighted line.

### 4. Runtime DNA & Execution Timeline
- Opt-in non-invasive runtime instrumentation (Python `sys.setprofile`, Node.js wrapper, Go middleware).
- Execution path tracking and function duration profiling (microseconds).
- **Static vs Runtime Reconciliation**: Compares inferred static relationships against actual observed execution paths.

### 5. Change Impact Analyzer
- Integrated with `git diff` against any base branch or commit (e.g. `main`, `HEAD~1`).
- Maps modified diff lines to AST symbols.
- Calculates blast radius: upstream callers, downstream callees, affected API routes, and relevant tests to run.
- Quantified Risk Scoring (0 - 100) based on complexity and exposure.

### 6. Architecture Drift & Layer Boundary Rules
- Enforces clean architectural boundaries (e.g., Domain must not depend on Presentation / Controllers).
- Configurable rules in `.codedna/architecture.json` or `codedna.arch.json`.
- Calculates overall Layer Health Score (0 - 100) and pinpoints boundary violations.

### 7. AI Codebase Intelligence & Context Export
- Context export formatted for **Claude Code** (`CLAUDE.md`), **Cursor** (`.cursorrules`), and **Codex**.
- Automatic Business Logic Domain mapping (Auth, Billing, Users, Orders, APIs, Database).
- Codebase Q&A powered by local Ollama or OpenAI-compatible backends.
- **Strict Privacy Safeguard**: Zero source code or AST facts sent to external AI providers without explicit user consent.

### 8. Intelligence Reports
- Generates comprehensive Markdown and JSON software intelligence reports.

---

## 🚀 Quick Start

### Installation

#### Pre-built Binary or From Source:
```bash
# Clone repository
git clone https://github.com/codedna/codedna.git
cd codedna

# Build Go CLI binary
go build -o bin/codedna ./cmd/codedna

# Optional: add to PATH
export PATH=$PATH:$(pwd)/bin
```

---

## 💻 CLI Commands

```bash
# 1. Scan a repository and index symbols
codedna scan .

# 2. Deep architectural and dependency analysis
codedna analyze .

# 3. View connected symbol graph in terminal or JSON
codedna graph .
codedna graph . --format json

# 4. Analyze change impact against main branch
codedna impact --base main

# 5. Generate comprehensive executive report
codedna report --format markdown
codedna report --format json --out report.json

# 6. Start API server and interactive Web Dashboard
codedna serve --port 8080

# 7. Watch repository for file modifications and live-index
codedna watch .

# 8. Enable runtime instrumentation adapter
codedna runtime enable --lang python
```

---

## 🖥️ Web Dashboard Views

When running `codedna serve`, access `http://localhost:8080` (or `http://localhost:3000` in dev mode):

1. **Overview Dashboard**: System health score, circular cycles, complexity hotspots, and layer health.
2. **Repository Explorer**: Two-pane file tree and source code inspector with syntax lines.
3. **DNA Graph**: Interactive React Flow node-link canvas with search and layout controls.
4. **Dependencies**: External package manifests categorized by ecosystem.
5. **Runtime Timeline**: Observed execution traces, hot functions, and static vs runtime coverage.
6. **Change Impact Analysis**: Diff risk scoring, modified functions, and test recommendations.
7. **Architecture Health**: Enforced rules and boundary violation reports.
8. **AI Assistant**: Q&A, business logic mapping, and Claude Code / Cursor context export.
9. **Settings**: Ignore patterns, Ollama configuration, and database storage info.

---

## 🔒 Security & Privacy

- **No Remote Transmission**: All parsing, graph construction, and analytics happen locally.
- **Safe Scanning**: Never automatically executes untrusted repository code.
- **Path Traversal Protection**: All filesystem endpoints validate boundaries against the repository root.
- **Credential Redaction**: Sanitizes embedded tokens from remote Git URLs.

---

## 📄 License

CodeDNA is licensed under the [MIT License](LICENSE).
