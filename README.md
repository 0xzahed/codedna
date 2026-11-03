<div align="center">

# 🧬 CodeDNA

### Living Software Intelligence, Dynamic Call Graphs & Architectural Topology Platform in Go
#### The Open-Source Codebase Cartography Engine for Modern Engineering Teams

[![CI](https://github.com/0xzahed/codedna/actions/workflows/ci.yml/badge.svg)](https://github.com/0xzahed/codedna/actions/workflows/ci.yml)
[![Release Build](https://github.com/0xzahed/codedna/actions/workflows/release.yml/badge.svg)](https://github.com/0xzahed/codedna/actions/workflows/release.yml)
[![Docker Image](https://github.com/0xzahed/codedna/actions/workflows/docker.yml/badge.svg)](https://github.com/0xzahed/codedna/actions/workflows/docker.yml)
[![Latest Release](https://img.shields.io/github/v/release/0xzahed/codedna?color=brightgreen)](https://github.com/0xzahed/codedna/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/0xzahed/codedna)](https://goreportcard.com/report/github.com/0xzahed/codedna)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/0xzahed/codedna)](https://golang.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.5-3178C6?logo=typescript)](https://www.typescriptlang.org/)
[![React Flow](https://img.shields.io/badge/React_Flow-11.11-FF0072?logo=react)](https://reactflow.dev/)
[![SQLite](https://img.shields.io/badge/SQLite-3-003B57?logo=sqlite)](https://www.sqlite.org/)
[![Website](https://img.shields.io/badge/Website-0xzahed.github.io%2Fcodedna-38bdf8?logo=googlechrome&logoColor=white)](https://0xzahed.github.io/codedna/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

<p align="center">
  <b>CodeDNA</b> is a lightning-fast, privacy-first software intelligence platform written in Go that constructs an interactive, living topological map of complex codebases. It combines <b>Tree-sitter AST parsing</b>, <b>cross-file call resolution</b>, <b>Tarjan's cycle detection</b>, <b>Git diff blast-radius propagation</b>, <b>architectural layer drift enforcement</b>, and a standalone <b>React Flow Web Intelligence Suite</b> with zero cloud dependencies.
</p>

[🌐 Live Website](https://0xzahed.github.io/codedna/) •
[📦 Installation](#-instant-installation-zero-dependencies--no-git-clone) •
[⚡ Quick Start](#-30-second-quick-start) •
[🎯 The Problem](#-the-problem) •
[⚡ The CodeDNA Engine](#-the-codedna-engine-architecture) •
[📊 Feature Comparison](#-feature-comparison) •
[🚀 Terminal Preview](#-terminal-preview) •
[✨ Features](#-features) •
[🖥️ Web Dashboard](#-web-dashboard-intelligence-suite) •
[💻 CLI Reference](#-cli-commands--flag-matrix) •
[🏗️ Architecture](#-system-architecture) •
[🚀 Benchmarks](#-performance-benchmarks) •
[🛠️ Project Structure](#-project-structure) •
[🔒 Security & Privacy](#-security-local-first-guarantee--privacy) •
[🤝 Contributing](#-contributing) •
[📄 License](#-license)

</div>

---

## ⚡ Instant Installation (Zero Dependencies & No Git Clone)

Install CodeDNA natively in one command without Docker or cloning the repository:

### 🐧 Linux & 🍎 macOS (Terminal)
```bash
curl -fsSL https://raw.githubusercontent.com/0xzahed/codedna/main/install.sh | bash
```

### 🪟 Windows (PowerShell 10 & 11)
```powershell
irm https://raw.githubusercontent.com/0xzahed/codedna/main/install.ps1 | iex
```

### 🐳 Optional Docker Run (Container Fallback)
```bash
docker run -d -p 8080:8080 -v $(pwd):/workspace ghcr.io/0xzahed/codedna:latest
```

> [!TIP]
> **No sudo or admin privileges required!** CodeDNA installs into your user-level binary path (`~/.local/bin` on Linux/macOS, `$HOME\.codedna\bin` on Windows), verifies SHA256 checksums, and bundles the complete web intelligence studio offline.

---

## 🚀 30-Second Quick Start

```bash
# 1. Navigate to ANY local code repository (Go, TS, JS, Python, Rust, Java, C++, PHP, etc.)
cd /path/to/your-project

# 2. Launch CodeDNA Architecture Studio
codedna serve
```

```text
CodeDNA Architecture Studio

Project: /home/user/your-project
Runtime: Native (v1.0.0)
Scanner: Initializing

[OK] Project scan completed (142 files, 1,280 symbols parsed)
[OK] Architecture graph generated
[OK] Web server started

Dashboard: http://127.0.0.1:8080
Press Ctrl+C to stop.
```

**3. Open [http://localhost:8080](http://localhost:8080)** in your browser to explore:
* 🗺️ **Living React Flow Topology** with automatic tier clustering & blast-radius heatmaps
* 🏛️ **System Design Engine** with GoF design pattern gallery & cross-tier dataflow ribbons
* 🛡️ **Architectural Guardrails** enforcing Clean Architecture & acyclic boundaries
* ⚡ **Live SSE Hot-Reloading** syncing code modifications in real time without refreshing

---

## 📦 Detailed Installation Options & Platform Support

### Supported Platforms & Environments

| Operating System | Architectures | Installer | Prebuilt Binaries | Sudo Needed? |
|:---|:---|:---:|:---:|:---:|
| **Ubuntu / Debian / Mint** | AMD64, ARM64 | `install.sh` | [v1.0.0](https://github.com/0xzahed/codedna/releases) | ❌ No (`~/.local/bin`) |
| **Fedora / RHEL / CentOS** | AMD64, ARM64 | `install.sh` | [v1.0.0](https://github.com/0xzahed/codedna/releases) | ❌ No (`~/.local/bin`) |
| **Arch Linux / Manjaro** | AMD64, ARM64 | `install.sh` | [v1.0.0](https://github.com/0xzahed/codedna/releases) | ❌ No (`~/.local/bin`) |
| **macOS Apple Silicon (M1/M2/M3/M4)** | ARM64 | `install.sh` | [v1.0.0](https://github.com/0xzahed/codedna/releases) | ❌ No (`~/.local/bin`) |
| **macOS Intel** | AMD64 | `install.sh` | [v1.0.0](https://github.com/0xzahed/codedna/releases) | ❌ No (`~/.local/bin`) |
| **Windows 10 & 11** | AMD64 | `install.ps1` | [v1.0.0](https://github.com/0xzahed/codedna/releases) | ❌ No (`~/.codedna/bin`) |
| **WSL2 / Git Bash** | AMD64, ARM64 | `install.sh` | [v1.0.0](https://github.com/0xzahed/codedna/releases) | ❌ No (`~/.local/bin`) |

<details>
<summary><b>Option A: Pre-Compiled Binary Archives (GitHub Releases)</b></summary>

Download and unpack official standalone release archives directly from [GitHub Releases](https://github.com/0xzahed/codedna/releases/tag/v1.0.0):

```bash
# Linux x86_64
mkdir -p ~/.codedna ~/.local/bin
curl -sSL https://github.com/0xzahed/codedna/releases/latest/download/codedna-linux-amd64.tar.gz | tar -xz -C ~/.codedna
ln -sf ~/.codedna/bin/codedna ~/.local/bin/codedna

# macOS Apple Silicon (M1/M2/M3/M4)
mkdir -p ~/.codedna ~/.local/bin
curl -sSL https://github.com/0xzahed/codedna/releases/latest/download/codedna-darwin-arm64.tar.gz | tar -xz -C ~/.codedna
ln -sf ~/.codedna/bin/codedna ~/.local/bin/codedna

# macOS Intel
mkdir -p ~/.codedna ~/.local/bin
curl -sSL https://github.com/0xzahed/codedna/releases/latest/download/codedna-darwin-amd64.tar.gz | tar -xz -C ~/.codedna
ln -sf ~/.codedna/bin/codedna ~/.local/bin/codedna

# Checksum Verification
curl -sSLO https://github.com/0xzahed/codedna/releases/latest/download/SHA256SUMS
sha256sum --ignore-missing -c SHA256SUMS
```
</details>

<details>
<summary><b>Option B: Install via Go Toolchain</b></summary>

If you have Go 1.21+ installed on your system:

```bash
go install github.com/0xzahed/codedna/cmd/codedna@latest
```
</details>

<details>
<summary><b>Option C: Build from Source (Developer Setup)</b></summary>

```bash
git clone https://github.com/0xzahed/codedna.git
cd codedna

# Build Go CLI and export static web assets
make all

# Or build CLI binary only
make build-cli

# Run complete test suite with race detector
make test-race

# Run architecture quality gate
make check
```
</details>

<details>
<summary><b>Option D: Optional Docker Container Mode</b></summary>

If you prefer containerized isolation without installing local binaries:

```bash
# Native CLI auto-delegation (uses local Docker daemon)
codedna docker --port 8080

# Pure Docker run command
docker run -d -p 8080:8080 -v $(pwd):/workspace ghcr.io/0xzahed/codedna:latest
```
</details>

<details>
<summary><b>Option E: Environment Diagnostics, Updates & Uninstallation</b></summary>

```bash
# Verify your installation and toolchains
codedna doctor

# Check for new releases and upgrade
codedna update

# Safely uninstall CodeDNA and remove cached assets
codedna uninstall -y

# Or uninstall via one-liner
curl -fsSL https://raw.githubusercontent.com/0xzahed/codedna/main/install.sh | bash -s -- --uninstall

# Or on Windows PowerShell
powershell -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/0xzahed/codedna/main/install.ps1 | iex -Uninstall"
```
</details>

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
Powered by native **Tree-sitter C bindings** and resilient static parsers, CodeDNA deploys an adaptive 16-worker pool with streaming channel pipelines across **Go, Python, TypeScript, JavaScript, Java, PHP, Rust, C, C++, C#, Ruby, Kotlin, Swift, Dart, SQL, and GraphQL**. It extracts AST symbols (functions, classes, interfaces, structs, methods, imports, and calls) with exact line/column byte offsets in milliseconds while respecting `.gitignore` rules and automatically ignoring non-source build artifacts (`dist`, `node_modules`, `vendor`, `.venv`, binary files).

### 2. Cross-File Call & Import Resolution
CodeDNA resolves symbolic references across disparate files with explicit confidence scoring (`exact-local`, `exact-module`, `exact-unique`, `module-qualified`, `self-method`). It tracks both direct AST imports and semantic invocation calls, enabling precise directory cluster grouping and directional dependency graphs.

### 3. Tarjan Strongly Connected Components (SCC) Cycle Resolver
Circular dependencies degrade modularity, prevent clean code decoupling, and induce race conditions. CodeDNA implements **Tarjan’s linear-time SCC algorithm** to discover cyclical dependency clusters and calculate BFS reachability trees for any given component.

### 4. Git Diff Blast-Radius Propagation Engine
Integrated directly with your Git workflow, CodeDNA compares your working tree or branch against any base commit (`main`, `HEAD~1`, `origin/main`). It maps modified lines to exact AST nodes and propagates upstream/downstream impact paths, calculating a quantified **Risk Score (0–100)** and pinpointing affected API endpoints and candidate tests.

### 5. Architectural Guardrail & Layer Drift Detection
Define clean architectural contracts in `.codedna/architecture.json` (e.g. `domain` must never depend on `infra` or `controller`). CodeDNA evaluates all resolved dependencies against these rules, pinpoints boundary violations with line-accurate call chains, and computes a system-wide **Layer Health Score (0–100%)**.

### 6. System Design Topology & Design Pattern Recognition Gallery
CodeDNA automatically classifies repository components into **6 architectural tiers** (Client & Gateway, API Presentation, Domain Services, Persistence Storage, Async Messaging, External Integrations) and discovers classic **Gang of Four (GoF)** and distributed patterns (Clean Architecture, CQRS, Repository, Factory, Singleton, Pub-Sub, Middleware Pipelines) with confidence scoring and concrete code evidence.

### 7. Multi-Repository Workspace & Cross-Service API Linker
Microservice and full-stack systems span multiple repositories. CodeDNA scans frontend API invocations (`fetch`, `axios`) and backend route definitions (FastAPI, Django, Express, Next.js, Gin, Spring Boot), using route normalization and prefix matching to weave individual repositories into a unified full-stack architecture diagram.

### 8. Native Mermaid & C4 Architecture Export
Transform living codebase topology into standard **Mermaid.js Flowcharts** (`flowchart TD`) and **C4 Component Diagrams** with a single click or CLI flag, perfect for markdown documentation, PR descriptions, and architectural RFCs.

### 9. Production-Grade Safety & Port Conflict Management
CodeDNA protects your system by blocking accidental scans of user home directories (`~` or `$HOME`) unless explicitly permitted via `--allow-home`. During `codedna serve`, synchronous TCP socket binding verifies port readiness before printing startup URLs, detects port collisions gracefully, binds to `127.0.0.1` by default for local privacy, and handles Ctrl+C termination smoothly.

---

## 📊 Feature Comparison

| Capability | Sourcegraph | SonarQube | CodeSee | **CodeDNA** |
| :--- | :---: | :---: | :---: | :---: |
| **100% Local-First & Zero Cloud Upload** | ❌ (Cloud/Server) | ❌ (Server setup) | ❌ (SaaS cloud) | **✅ (100% Offline / Local)** |
| **Interactive Living Topology Canvas** | ❌ (Code search only)| ❌ (Static tables) | ✅ (Web app) | **✅ (React Flow + Hierarchical)** |
| **Polyglot AST & Symbol Parsing** | Partial | Partial (Proprietary AST) | Proprietary | **✅ (16+ Languages: Go, TS, JS, Py, Java, PHP, Rust, C/C++, C#, Ruby, Kotlin, Swift, Dart, SQL, GraphQL)** |
| **Tarjan Circular Dependency Detection** | ❌ | ⚠️ (Limited metrics) | ❌ | **✅ (Tarjan SCC + Cycle Chains)** |
| **Git Diff Blast-Radius Risk Scoring** | ❌ | ❌ | Partial | **✅ (AST-mapped Risk 0–100)** |
| **Architectural Layer Drift Guardrails** | ❌ | ⚠️ (Config-heavy) | ❌ | **✅ (Declarative Layer Rules)** |
| **Multi-Repo Cross-Service API Linking** | ❌ | ❌ | ❌ | **✅ (Frontend/Backend Stitching)** |
| **Mermaid & C4 Architectural Export** | ❌ | ❌ | ❌ | **✅ (One-click MMD & C4)** |
| **Database Requirement** | Postgres / Heavy DB | PostgreSQL | Cloud Hosted | **Zero (Built-in Embedded SQLite)** |
| **Resource Footprint** | Heavy (>2 GB RAM) | Heavy (>1 GB RAM) | Browser-heavy | **Ultra-Lightweight (<45 MB RAM)** |

---

## 🚀 Terminal Preview

```text
╔═══════════════════════════════════════════════════════════════╗
║                      🧬 C O D E D N A 🧬                      ║
║     Living Software Intelligence & Topological Architecture   ║
╚═══════════════════════════════════════════════════════════════╝

[1/5] Ingesting Codebase Filesystem
✔ Target: /home/zahed/codedna (Branch: main @ 429afcc)
✔ Multi-Core Worker Pool: 16 parallel threads active
✔ Discovered: 104 source files (Go, TypeScript, JavaScript, CSS)

[2/5] Tree-sitter AST Parallel Parsing
Parsing: [██════════════════════════════] 100.0% (104/104 files) | 582 files/sec
✔ Extracted: 489 functions, 42 classes/structs, 612 import statements, 1,280 calls

[3/5] Cross-File Dependency & Symbol Resolution
✔ Resolving call targets with multi-tier confidence heuristics...
✔ Resolved: 729 file-to-file relationships (exact-module: 482, exact-local: 247)

[4/5] Graph Theory & Architectural Guard Evaluation
✔ Tarjan SCC: 0 circular dependency cycles detected
✔ Layer Drift Engine: Evaluated 6 architectural boundary rules (Health Score: 98.4%)
✔ BFS Reachability: Indexing upstream/downstream blast radius chains

[5/5] Generating Intelligence Snapshot & Reports
✔ Local SQLite index stored: ~/.codedna/codedna.db
✔ Server Daemon listening on: http://localhost:8080

═══════════════════════════════════════════════════════════════
  🎉 CODEBASE TOPOLOGY READY! DASHBOARD ACTIVE! 🎉
═══════════════════════════════════════════════════════════════
 • System Health:    98 / 100 (Optimal)
 • Circular Cycles:  0 detected
 • Complexity Alert: 3 high-cyclomatic functions highlighted
 • Interactive Web:  http://localhost:8080/graph/
 • Export Diagrams:  http://localhost:8080/api/scans/latest/export?format=mermaid
═══════════════════════════════════════════════════════════════
```

---

## ✨ Features

- ⚡ **22x Multi-Core Parallel Engine:** High-throughput AST ingestion using 16 concurrent worker routines to scan 15,000+ lines of code in under 1.5 seconds.
- 🌳 **Polyglot Code Coverage (16+ Languages):** Native Tree-sitter parsers (Go, Python, TypeScript, JavaScript, Java, PHP) and resilient static AST parsers (Rust, C, C++, C#, Ruby, Kotlin, Swift, Dart, SQL, GraphQL).
- 🏛️ **Automated System Design Topology:** Categorizes code across 6 tiers (Client, API Presentation, Domain, Storage, Async Workers, Integrations) and discovers GoF design patterns (Clean Architecture, CQRS, Repository, Factory, Singleton, Pub-Sub, Middleware).
- 🔄 **Tarjan Circular Dependency Analysis:** Pinpoints circular dependency chains with linear-time Strongly Connected Components (SCC) calculations.
- 🎯 **Git Diff Blast Radius:** Quantifies the upstream caller and downstream callee blast radius of pending pull requests or working tree changes with a 0–100 risk score.
- 🛡️ **Layer Drift & Architectural Guardrails:** Enforces clean separation of concerns (Domain, Application, Infrastructure, Presentation) with automated CI quality gate (`codedna check`).
- 👥 **Team Knowledge Map & Bus Factor:** Parses git blame and numstat histories to identify line ownership distribution and single-contributor high-risk files.
- 🌐 **Cross-Service Full-Stack Linking:** Discovers REST endpoints in backends and client-side HTTP calls in frontends, linking independent repositories into a composite graph.
- 📐 **Instant Mermaid & C4 Export:** Render or export live architecture diagrams into GitHub-flavored Markdown or C4 component definitions.
- 🔒 **Zero-Knowledge Privacy & Safety:** 100% offline, local SQLite embedded storage with zero remote telemetry, automatic secret scrubbing, and home-directory guardrails.
- 🚀 **Zero-Clone Native Installers:** Single-command curl and PowerShell installers with automated SHA256 checksum verification and binary self-updates.
- 🖥️ **Full-Featured Web Intelligence Suite:** Built-in React Flow canvas with dark-mode Monaco code drawer, dependency manifest inspector, and runtime execution profiler.

---

## 🎮 CLI Cheatsheet & Core Workflows

### CLI Command Overview

| Command | Description | Needs Server? | Output Formats |
| :--- | :--- | :--- | :---: |
| `codedna serve [flags]` | Launch REST API daemon and interactive React Flow Web Studio | ❌ No | HTTP (`127.0.0.1:8080`) |
| `codedna scan [path]` | Ingest directory and index AST symbols into SQLite | ❌ No | Console / JSON |
| `codedna doctor` | Diagnose environment, installation health, ports & toolchains | ❌ No | Console Diagnostics |
| `codedna install` | Self-installs standalone executable into user PATH (~/.local/bin) | ❌ No | System Config |
| `codedna update` | Queries GitHub API for latest releases and self-updates | ❌ No | Network / Binary |
| `codedna uninstall [-y]` | Safely removes CodeDNA binary, web assets, and local cache | ❌ No | Console |
| `codedna docker [flags]` | Runs CodeDNA inside an isolated container with volume mount | ❌ No | Container / Web |
| `codedna check [flags]` | Automated CI/CD architecture quality gate & rule enforcement | ❌ No | Console / JSON |
| `codedna design [--repo <path>]` | Reconstruct multi-tier system design, protocol flows & design patterns | ❌ No | Console / JSON |
| `codedna blame, knowledge` | Analyzes git line ownership, team knowledge map & bus factor risk | ❌ No | Console / JSON |
| `codedna analyze <path>` | Deep scan with Tarjan SCC cycles & layer drift verification | ❌ No | Console / JSON |
| `codedna graph <path>` | Print connected symbol graph and dependency edges | ❌ No | Text / JSON |
| `codedna impact [path]` | Calculate Git diff blast radius against target branch | ❌ No | Table / JSON |
| `codedna report [path]` | Generate comprehensive markdown or JSON intelligence report | ❌ No | Markdown / JSON |
| `codedna watch <path>` | Monitor directory for live file modifications and hot-reindex | ❌ No | Live Stream / SSE |
| `codedna runtime enable` | Configure opt-in runtime instrumentation profiler | ❌ No | Middleware |
| `codedna version` | Display version, build, runtime, and license information | ❌ No | Text |
| `codedna help` | Display comprehensive command reference manual | ❌ No | Text |

---

### 1. Ingesting & Analyzing a Repository

```bash
# Fast AST scan and symbol indexing
codedna scan .

# Deep architectural analysis (detects circular cycles, layer violations, dead code)
codedna analyze .

# Automated CI/CD architecture quality gate
codedna check --min-health 80 --max-cycles 0

# Inspect System Design tiers and GoF design pattern gallery
codedna design

# Machine-readable JSON output for CI/CD integration
codedna analyze . --json
```

---

### 2. Inspecting Change Blast Radius (Git Diff Impact)

Before submitting a Pull Request, evaluate which upstream callers and downstream consumers are affected:

```bash
# Compare working tree or feature branch against main
codedna impact --base main

# Compare against specific Git commit or release tag
codedna impact --base v1.0.0 --json
```

---

### 3. Launching the Interactive Web Intelligence Dashboard

```bash
# Start server daemon on default port 8080 (binds to 127.0.0.1)
codedna serve

# Start on custom port and host without auto-opening browser
codedna serve --port 8081 --host 0.0.0.0 --no-browser

# Target a specific external repository
codedna serve --repo /path/to/project
```
Then navigate to **`http://localhost:8080`** in your browser.

---

### 4. Exporting Architecture Diagrams (Mermaid & C4)

Generate instant documentation diagrams without manual drawing:

```bash
# Export as Mermaid Flowchart for GitHub markdown
curl -s http://localhost:8080/api/scans/latest/export?format=mermaid > architecture.mmd

# Export as C4 Component Model
curl -s http://localhost:8080/api/scans/latest/export?format=c4 > c4_components.txt
```

---

### 5. Advanced CLI Flags Matrix

| Flag | Applicable Commands | Description | Default |
| :--- | :--- | :--- | :---: |
| `--port <num>` | `serve`, `docker` | HTTP port for the local REST API and Web Dashboard | `8080` |
| `--host <ip>` | `serve` | Network bind address (`127.0.0.1` for local, `0.0.0.0` for LAN) | `127.0.0.1` |
| `--no-browser` | `serve` | Do not attempt to automatically open browser on startup | `false` |
| `--repo <path>` | `serve`, `watch`, `docker` | Repository root directory to bind and analyze | `.` |
| `--allow-home` | `serve`, `scan`, `analyze` | Bypass safety guard to permit scanning user home directory | `false` |
| `--min-health <num>`| `check` | Minimum layer health score required to pass CI gate | `80.0` |
| `--max-cycles <num>`| `check` | Maximum circular dependency cycles allowed before failure | `0` |
| `--max-violations` | `check` | Maximum boundary violations allowed before failure | `0` |
| `--json` | `scan`, `analyze`, `graph`, `impact`, `check`, `design`, `blame` | Output results in machine-readable structured JSON format | `false` |
| `--base <ref>` | `impact` | Target Git branch, commit SHA, or tag for diff calculation | `main` |
| `--format <fmt>` | `report`, `export` | Output formatting profile (`markdown`, `json`, `mermaid`, `c4`) | `markdown` |
| `--out <file>` | `report` | Destination file path for generated report | `stdout` |
| `--lang <type>` | `runtime` | Target runtime instrumentation adapter (`python`, `node`, `go`) | `python` |

---

## 🖥️ Web Dashboard Intelligence Suite

When running `codedna serve`, access the interactive UI at **`http://localhost:8080`**. CodeDNA provides 10 specialized intelligence views:

### 1. 🌐 System Overview & Health Dashboard (`/`)
* **Aggregate Health Score (0–100):** Real-time algorithmically weighted health rating calculated from circular cycles, cyclomatic hotspots, dead code volume, and architectural layer violations.
* **Telemetry Metrics:** Instant inventory of total source files, parsed AST symbols, functions, classes/structs, and resolved dependency edges.
* **System Design & Topology Card:** Snapshot of inferred architecture style, reliability score, active tier count, and detected pattern totals with 1-click drilldown.
* **Team Knowledge & Bus Factor Card:** Line additions leaderboard, primary maintainer distribution, and single-contributor high-risk module alerts.
* **Top Hotspot Indicators:** Surface complex, high-churn functions that pose architectural maintenance risk.

### 2. 🏛️ System Design & Architecture Topology Flow (`/design/`)
* **Multi-Tier Architectural Classification:** Automatically organizes codebase components into 6 industry-standard tiers:
  * 🌐 **Tier 1 — Client & Gateway:** Frontend pages, reverse proxies, and edge routing entrypoints.
  * 🖥️ **Tier 2 — API & Presentation:** HTTP controllers, REST endpoints, GraphQL resolvers, and RPC interfaces.
  * ⚙️ **Tier 3 — Application & Domain Services:** Core business domain logic, orchestrators, use cases, and policy rules.
  * 🗄️ **Tier 4 — Persistence & Data Storage:** Repositories, database entities, ORM models, SQL queries, and cache layers.
  * 📻 **Tier 5 — Asynchronous Workers & Queues:** Background workers, event emitters, consumer loops, and webhooks.
  * 🔗 **Tier 6 — External Integrations:** Third-party payment gateways, external cloud storage, and mailers.
* **Cross-Tier Protocol Data Flow Pipeline:** Reconstructs inter-tier communication hops from AST call edges with protocol tagging (`HTTPS / REST API`, `In-Memory Orchestration`, `SQL / Storage Interface`, `Async Event / Channel Pipeline`) and call counts.
* **Software Design Pattern Recognition Gallery:** Automatically detects classic Gang of Four and modern distributed patterns with confidence scores, code evidence, and engineering trade-offs:
  * **Architectural:** Clean Architecture / Ports & Adapters (Hexagonal), CQRS.
  * **Structural:** Repository / Data Mapper Pattern.
  * **Creational:** Factory / Constructor Pattern, Singleton / Shared Registry.
  * **Behavioral:** Event-Driven / Pub-Sub Pattern, Chain of Responsibility / Middleware Pipeline.
* **Interactive Tier Inspector:** Click any tier to drill down into its constituent files and key exported symbols.
* **Architectural Bottleneck & Hotspot Diagnostics:** Detects layer-bypassing violations (e.g. Client directly querying Storage) and high-traffic inter-tier hops.

### 2. 🧬 Interactive Living DNA Graph (`/graph/`)
* **React Flow Topology Engine:** Hardware-accelerated canvas with smooth zooming, panning, auto-centering, minimap navigation, and fuzzy symbol search.
* **Architectural Cluster Grouping:** Intelligent spatial layout organizing nodes by physical directory hierarchy and package domains.
* **Multi-Color Edge Classification:**
  * 🟪 **Violet Edges:** Direct AST module imports (`import "..."` / `require`).
  * 🟩 **Green Edges:** Semantic function/method invocation calls across files.
  * 🟦 **Sky Blue Edges:** Cross-service REST API calls linking frontend actions to backend route handlers.
* **Slide-over Source Drawer:** Click any node to instantly view the actual source file with syntax highlighting, exact line offset, and symbol caller details.
* **Live Mermaid & C4 Export:** One-click modal to copy or download production Mermaid flowcharts and C4 component diagrams.

### 3. 📂 Repository Explorer & Source Code Inspector (`/repo/`)
* **Two-Pane Explorer:** Fast hierarchical directory file tree with file sizes and language detection badges.
* **Deep Code Inspector:** Syntax-highlighted code viewer with line numbers, Git commit author annotations, and automatic credential scrubbing.

### 4. 📦 Dependency Manifest Matrix (`/dependencies/`)
* **Ecosystem Manifest Coverage:** Automatically indexes `go.mod`, `package.json`, `requirements.txt`, `pyproject.toml`, `composer.json`, `pom.xml`, and `build.gradle`.
* **Third-Party Dependency Audit:** Categorizes direct vs transitive packages and visualizes library adoption footprint.

### 5. ⏱️ Runtime DNA & Execution Timeline (`/runtime/`)
* **Microsecond Profiling:** Non-invasive execution path telemetry capturing function invocation duration and stack traces.
* **Static vs Runtime Reconciliation:** Juxtaposes static AST call trees against live runtime executions to reveal dead code and undocumented dynamic dispatches.

### 6. 🎯 Change Impact & Blast Radius Analyzer (`/impact/`)
* **Git Diff Symbol Mapping:** Parses incoming Git modifications line-by-line and maps changed lines to discrete AST functions and methods.
* **Quantified Risk Score (0–100):** Evaluates cyclomatic complexity, incoming callers, and public API exposure of modified symbols.
* **Blast Radius Tree & Test Selection:** Visualizes cascading upstream caller impacts and recommends the minimal set of unit/integration tests to execute.

### 7. 🛡️ Architectural Health & Layer Drift Enforcer (`/architecture/`)
* **Layer Contracts:** Configurable boundaries defining relationships between `presentation`, `application`, `domain`, and `infrastructure`.
* **Zero-Tolerance Boundary Violation Audit:** Lists illicit cross-layer imports with exact source line and column numbers.
* **Layer Health Gauge:** Visual 0–100% compliance gauge to prevent architectural debt from slipping into releases.

### 8. 🤖 AI Codebase Intelligence & Context Exporter (`/ai/`)
* **Business Logic Domain Mapping:** Automatically classifies codebase subsystems (Authentication, Billing, Data Ingestion, API Controllers, Database).
* **Targeted AI Context Generation:** One-click generation of optimized context files for **Claude Code** (`CLAUDE.md`), **Cursor** (`.cursorrules`), and **Codex**.
* **Zero-Cloud Privacy Guarantee:** Local-first heuristics with optional local Ollama integration. Zero source code transmitted externally without explicit user confirmation.

### 9. ⚙️ Multi-Repo Workspace Switcher & Configuration (`/settings/`)
* **Multi-Repository Workspace Linker:** Switch seamlessly between open codebases or link frontend and backend repositories together.
* **Scan Parameters & Filters:** Configure ignore globs (`.codednaignore`), AST parse budgets, and local SQLite database maintenance.

---

## 🏗️ System Architecture & Data Flow

CodeDNA operates as a self-contained, layered architecture combining high-speed Go systems programming with modern reactive web topology:

```mermaid
flowchart TD
    subgraph INGEST["1. Ingestion & File Discovery"]
        FS["Source Filesystem (Go, Py, TS, JS, Java, PHP)"]
        GIT["Git Metadata & Diff History"]
        MAN["Dependency Manifests (go.mod, package.json, etc.)"]
    end

    subgraph WORKERS["2. Parallel Multi-Core AST Engine"]
        POOL["16-Worker Parallel Processing Pool"]
        TS["Tree-sitter C Parser Bindings"]
        FILTER["Asset & Minified Bundle Fast Filter"]
    end

    subgraph RESOLUTION["3. Intelligence & Graph Theory Engine"]
        CALLS["Cross-File Call & Import Resolver"]
        TARJAN["Tarjan's SCC Cycle Detector"]
        IMPACT["Git Diff Blast-Radius Analyzer"]
        DRIFT["Architectural Layer Guard"]
        API_MAP["Cross-Service REST Route Matcher"]
    end

    subgraph STORAGE["4. Embedded Persistence Tier"]
        SQLITE[("Embedded SQLite Engine (~/.codedna/codedna.db)")]
    end

    subgraph SERVICE["5. API & Live Streaming Tier"]
        DAEMON["Go HTTP Server (REST API)"]
        SSE["Server-Sent Events (SSE) Real-Time Streamer"]
        EXPORT["Mermaid.js & C4 Architecture Exporter"]
    end

    subgraph UI["6. Web Intelligence Suite (React Flow)"]
        CANVAS["React Flow Interactive Graph Canvas"]
        DRAWER["Monaco-Style Code Source Drawer"]
        DASH["9 Intelligence Specialized Dashboards"]
    end

    FS --> POOL
    GIT --> POOL
    MAN --> POOL

    POOL --> TS
    TS --> FILTER
    FILTER --> CALLS

    CALLS --> TARJAN
    CALLS --> IMPACT
    CALLS --> DRIFT
    CALLS --> API_MAP

    TARJAN --> SQLITE
    IMPACT --> SQLITE
    DRIFT --> SQLITE
    API_MAP --> SQLITE

    SQLITE --> DAEMON
    DAEMON --> SSE
    DAEMON --> EXPORT

    DAEMON --> CANVAS
    SSE --> CANVAS
    CANVAS --> DRAWER
    CANVAS --> DASH
```

---

## 🚀 Performance Benchmarks

CodeDNA's multi-core parallel pipeline delivers industry-leading throughput by streaming AST parse jobs across all available CPU cores:

### Parallel Parsing Throughput (v2.1.0 Multi-Core vs Single-Threaded)

| Codebase Scale | Source Files | Total Lines of Code | Single-Threaded AST | 16-Core Parallel AST | Speedup Multiplier | Peak Memory |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Small Project** | 45 files | ~6,200 LOC | 0.85s | **0.14s** | **6.1x faster** | 28 MB |
| **Medium Service** | 185 files | ~24,500 LOC | 14.20s | **0.92s** | **15.4x faster** | 54 MB |
| **Large Polyglot Repo** | 639 files | ~89,000 LOC | 124.95s | **5.44s** | **22.9x faster** | 118 MB |

### System Hardware Footprint
* **Process Memory (Idle):** ~38 MB RSS
* **Process Memory (Active Scan):** ~95–125 MB RSS
* **Binary Size (Compiled):** ~19.4 MB (completely statically linked, zero dynamic library dependencies)
* **Cold Daemon Startup Time:** `< 12 ms`
* **Local Database Storage:** `< 3 MB` SQLite storage footprint per 15,000 indexed symbols

---

## 🛠️ Project Structure

```text
codedna/
├── cmd/
│   └── codedna/
│       ├── main.go                 # CLI entry point, argument parser, and command dispatcher
│       ├── doctor.go               # Environment diagnostics, port availability & PATH checks
│       ├── lifecycle.go            # Native installer, updater, and uninstaller routines
│       └── docker.go               # Container runner wrapper and volume mount manager
├── internal/
│   ├── ai/                         # Privacy-first local AI synthesis & prompt exporter
│   ├── analysis/                   # Graph algorithms, Tarjan's SCC cycles & reachability
│   ├── api/                        # Embedded HTTP REST & Server-Sent Events daemon
│   ├── architecture/               # Architectural boundary & layer drift enforcement
│   ├── blame/                      # Team knowledge distribution, line ownership & bus factor
│   ├── engine/                     # Multi-core concurrent ingestion coordinator (16 workers)
│   ├── impact/                     # Git diff blast-radius propagation engine (0–100 risk score)
│   ├── model/                      # Strongly-typed domain models & data schemas
│   ├── parser/                     # Polyglot Tree-sitter & heuristic static AST extractors
│   │   ├── parse.go                # Tree-sitter parsers (Go, Py, TS, JS, Java, PHP)
│   │   ├── heuristic.go            # Static parsers (Rust, C/C++, C#, Ruby, Kotlin, Swift, SQL)
│   │   └── spec.go                 # AST node extraction queries and capture rules
│   ├── report/                     # Executive intelligence reports & diagram exporters
│   ├── runtime/                    # Non-invasive runtime execution profiler
│   ├── scanner/                    # Filesystem crawler, Git inspector & manifest auditor
│   │   ├── dependencies.go         # Manifest parser & Git metadata extractor
│   │   ├── frameworks.go           # Framework identifier (Next.js, FastAPI, Express, Django)
│   │   └── scanner.go              # Path resolver, gitignore filter & home directory guard
│   ├── store/                      # Embedded SQLite persistent storage tier
│   ├── sysdesign/                  # Multi-tier topology classifier & GoF pattern recognizer
│   ├── watcher/                    # Live filesystem hot-reloader & incremental indexer
│   └── workspace/                  # Multi-repo & cross-service architectural linker
├── web/                            # Next.js 14 + React Flow 11 Interactive Web Dashboard
│   ├── src/
│   │   ├── app/                    # 10 specialized intelligence dashboard views
│   │   │   ├── page.tsx            # Overview & System Health dashboard
│   │   │   ├── design/page.tsx     # System Design & Multi-Tier Topology flow
│   │   │   ├── graph/page.tsx      # React Flow interactive topological canvas
│   │   │   ├── repo/page.tsx       # Monaco-style repository file tree & code inspector
│   │   │   ├── dependencies/       # Dependency manifest matrix
│   │   │   ├── runtime/page.tsx    # Runtime execution timeline & tracer
│   │   │   ├── impact/page.tsx     # Change blast radius & Git diff risk analyzer
│   │   │   ├── architecture/       # Architectural drift & layer compliance
│   │   │   ├── ai/page.tsx         # AI assistant & context export interface
│   │   │   └── settings/page.tsx   # Multi-repo workspace switcher & engine config
│   │   └── components/             # Reusable UI widgets, modals, and SourceDrawer
├── install.sh                      # Zero-clone native installer for Linux & macOS
├── install.ps1                     # Zero-clone native installer for Windows PowerShell
├── CHANGELOG.md                    # Chronological release log and feature history
├── CONTRIBUTING.md                 # Contribution standards and development guidelines
├── CODE_OF_CONDUCT.md              # Contributor Covenant v2.1 code of conduct
├── Dockerfile                      # Multi-stage production container build
├── docker-compose.yml              # Local containerized orchestration
├── LICENSE                         # MIT Open Source License
├── Makefile                        # Multi-platform build, test, and release targets
├── README.md                       # Comprehensive project documentation
├── SECURITY.md                     # Security vulnerability disclosure & safety policy
└── TASKS.md                        # Feature-wise roadmap and task execution tracker
```

---

## 🔒 Security, Local-First Guarantee & Privacy

CodeDNA was architected from day one to operate in high-security, privacy-critical enterprise environments:

* 🛡️ **Zero Remote Data Transmission:** All AST parsing, graph calculations, and database indexing execute 100% locally on your machine. No telemetry or source code ever leaves your environment.
* ⚡ **Safe Static Analysis:** CodeDNA never executes untrusted repository code. It relies exclusively on static Tree-sitter AST queries and declarative manifest inspection.
* 🚫 **Strict Path Traversal Protection:** All filesystem and source drawer APIs enforce strict root directory boundary checks to prevent arbitrary directory traversal.
* 🔑 **Automatic Token Sanitization:** Git remote URLs containing embedded authentication tokens or credentials (e.g. `https://x-token@github.com/...`) are scrubbed before indexing and storage.
* 🤖 **Opt-in AI Privacy Guard:** AI domain classification runs locally. Remote LLM queries require explicit user confirmation (`userConsent: true`), ensuring zero proprietary code is dispatched to external APIs.

---

## 🤝 Contributing

Contributions are what make the open-source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**!

Please check out our [Contributing Guide](CONTRIBUTING.md) to learn how to:
1. Fork the Project and create your Feature Branch (`git checkout -b feat/AmazingFeature`)
2. Run test suites and static analysis (`make test`, `make lint`)
3. Commit your changes following [Conventional Commits](https://www.conventionalcommits.org/)
4. Push to the Branch (`git push origin feat/AmazingFeature`)
5. Open a Pull Request

---

## 📄 License

Distributed under the **MIT License**. See [`LICENSE`](LICENSE) for more information.

---

<div align="center">
  <sub>Crafted with ❤️ by <a href="https://github.com/0xzahed">Abu Zahed</a> and open-source contributors.</sub>
</div>
