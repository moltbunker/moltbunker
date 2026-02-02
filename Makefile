.PHONY: build daemon cli test test-quick test-smoke test-e2e test-all test-production clean install lint vet coverage doctor setup

# Build targets
build: daemon cli

daemon:
	@echo "Building daemon..."
	@go build -o bin/moltbunker-daemon ./cmd/daemon

cli:
	@echo "Building CLI..."
	@go build -o bin/moltbunker ./cmd/cli

# Test targets
test:
	@echo "Running unit tests..."
	@go test ./internal/... ./pkg/...

test-quick: test test-smoke
	@echo "Quick tests completed"

test-smoke:
	@echo "Running smoke tests..."
	@go test -v -tags=e2e -run TestSmoke ./tests/e2e/smoke/...

test-e2e:
	@echo "Running E2E tests..."
	@go test -v -tags=e2e -timeout 10m ./tests/e2e/...

test-all: test test-e2e
	@echo "All tests completed"

# Coverage
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./internal/... ./pkg/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Code quality
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

# Utilities
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

install: build
	@echo "Installing binaries..."
	@cp bin/moltbunker-daemon /usr/local/bin/
	@cp bin/moltbunker /usr/local/bin/

# Doctor command (run after install)
doctor: cli
	@echo "Running doctor..."
	@./bin/moltbunker doctor

# Production E2E tests (requires real dependencies)
test-production: build
	@echo "Running production E2E tests..."
	@./scripts/e2e-production-test.sh

test-production-verbose: build
	@echo "Running production E2E tests (verbose)..."
	@./scripts/e2e-production-test.sh --verbose

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@./scripts/setup-dev-environment.sh

# Development helpers
.PHONY: dev
dev:
	@echo "Starting daemon in development mode..."
	@go run ./cmd/daemon --data-dir ~/.moltbunker-dev

# Help
.PHONY: help
help:
	@echo "Moltbunker Build System"
	@echo ""
	@echo "Build targets:"
	@echo "  build        - Build both daemon and CLI"
	@echo "  daemon       - Build the daemon only"
	@echo "  cli          - Build the CLI only"
	@echo ""
	@echo "Test targets:"
	@echo "  test              - Run unit tests"
	@echo "  test-quick        - Run unit tests + smoke tests"
	@echo "  test-smoke        - Run smoke tests only"
	@echo "  test-e2e          - Run full E2E tests (with mocks)"
	@echo "  test-all          - Run all tests"
	@echo "  test-production   - Run production E2E tests (real deps)"
	@echo "  coverage          - Generate coverage report"
	@echo ""
	@echo "Quality targets:"
	@echo "  lint         - Run golangci-lint"
	@echo "  vet          - Run go vet"
	@echo ""
	@echo "Other targets:"
	@echo "  setup        - Setup development environment"
	@echo "  clean        - Remove build artifacts"
	@echo "  install      - Install binaries to /usr/local/bin"
	@echo "  doctor       - Run system health check"
	@echo "  dev          - Start daemon in development mode"
