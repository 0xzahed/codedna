# Contributing to CodeDNA

Thank you for contributing to CodeDNA! We welcome contributions to make codebase intelligence faster, more accurate, and accessible to developers worldwide.

---

## Development Setup

### Prerequisites

- **Go 1.23+**
- **Node.js 18+** and npm
- **Git**
- **Make** (standard on Linux/macOS)

### Build Instructions

1. **Clone the repository:**
   ```bash
   git clone https://github.com/0xzahed/codedna.git
   cd codedna
   ```

2. **Run tests & static analysis:**
   ```bash
   make test
   make test-race
   make lint
   ```

3. **Build the CLI binary:**
   ```bash
   make build-cli
   ```

4. **Build the Next.js Web Dashboard:**
   ```bash
   make build-web
   ```

5. **Start the local development daemon:**
   ```bash
   ./bin/codedna serve --port 8080
   ```

---

## Pull Request Guidelines

1. **Create a feature branch:**
   ```bash
   git checkout -b feat/my-feature
   ```
2. **Follow Conventional Commits:**
   Use clear semantic commit prefixes:
   - `feat(scope): add new feature`
   - `fix(scope): resolve bug`
   - `perf(scope): optimize algorithm or throughput`
   - `docs(scope): update documentation`
   - `test(scope): add or improve test coverage`
3. **Verify tests:**
   Ensure all tests pass cleanly before opening a PR:
   ```bash
   make test-race
   ```
4. **Submit Pull Request:**
   Push your branch and open a PR against `main` describing the changes and problem solved.
