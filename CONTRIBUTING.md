# Contributing to CodeDNA

Thank you for contributing to CodeDNA! We welcome contributions to make codebase intelligence faster, more accurate, and accessible to developers worldwide.

## Development Setup

### Prerequisites

- Go 1.23+
- Node.js 18+ and npm
- Git

### Build Instructions

1. Clone the repository:
   ```bash
   git clone https://github.com/codedna/codedna.git
   cd codedna
   ```

2. Run Go tests:
   ```bash
   go test -v ./...
   ```

3. Build the CLI binary:
   ```bash
   go build -o bin/codedna ./cmd/codedna
   ```

4. Build the Web Dashboard:
   ```bash
   cd web
   npm install
   npm run build
   ```

## Pull Request Guidelines

1. Create a feature branch (`git checkout -b feature/my-feature`).
2. Write unit tests for new algorithms and features.
3. Verify all tests pass (`go test ./...` and `npm test`).
4. Ensure code adheres to standard Go and TypeScript idioms.
5. Submit a pull request describing the changes and problem solved.
