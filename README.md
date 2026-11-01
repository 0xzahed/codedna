# CodeDNA 2.0 🧬

### *Understand Your Code. See Its DNA.*

[![CI](https://github.com/codedna/codedna/actions/workflows/ci.yml/badge.svg)](https://github.com/codedna/codedna/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/codedna/codedna)](https://goreportcard.com/report/github.com/codedna/codedna)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.5-3178C6?logo=typescript)](https://www.typescriptlang.org/)

**CodeDNA** is an advanced open-source software intelligence platform that creates a living, interactive topological map of your codebase. It discovers how software components are structured, how they interact, how they behave at runtime, and how code changes affect the entire system.

This is **not** a basic visualization tool, and **not** an AI chatbot wrapper. It is a production-grade developer tool combining Tree-sitter AST parsing, Tarjan cycle detection, Git diff blast-radius propagation, runtime tracing, and privacy-first local AI assistance.

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
