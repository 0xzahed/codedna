# CodeDNA Makefile
# Author: Abu Zahed <zahed04x@gmail.com>

BINARY_NAME=codedna
BIN_DIR=bin
GO_FILES=$(shell find . -name "*.go" -type f -not -path "./web/*")

.PHONY: all build build-cli build-web test test-race lint clean run help

all: build-cli build-web

## build-cli: Build the Go CLI binary
build-cli:
	@echo "==> Building CodeDNA CLI binary..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/codedna
	@echo "✔ Binary built: $(BIN_DIR)/$(BINARY_NAME)"

## build: Alias for build-cli
build: build-cli

## build-web: Build the Next.js interactive web dashboard
build-web:
	@echo "==> Building Next.js Web Dashboard..."
	@cd web && npm run build
	@echo "✔ Web dashboard built successfully"

## test: Run all Go unit and integration tests
test:
	@echo "==> Running Go test suite..."
	go test -v ./...

## test-race: Run Go test suite with data race detector enabled
test-race:
	@echo "==> Running Go test suite with race detector..."
	go test -v -race ./...

## lint: Run code vetting and static checks
lint:
	@echo "==> Running go vet..."
	go vet ./...

## check: Run automated CI/CD architecture quality gate
check: build-cli
	@echo "==> Running CodeDNA Architecture Quality Gate..."
	@./$(BIN_DIR)/$(BINARY_NAME) check --min-health=80 --max-cycles=0

## doctor: Run environment diagnostics and health checks
doctor: build-cli
	@./$(BIN_DIR)/$(BINARY_NAME) doctor

## install: Install CodeDNA into user PATH (~/.local/bin)
install: build-cli
	@./$(BIN_DIR)/$(BINARY_NAME) install

## run: Start the local CodeDNA server daemon on port 8080
run: build-cli
	@echo "==> Launching CodeDNA daemon on http://localhost:8080..."
	./$(BIN_DIR)/$(BINARY_NAME) serve --port 8080

## clean: Remove compiled binaries and build artifacts
clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -rf web/.next
	@echo "✔ Clean completed"

## help: Display available Makefile targets
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':'
