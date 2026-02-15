.PHONY: build build-all daemon api cli exec-agent test test-quick test-smoke test-e2e test-colima \
       test-integration test-localnet test-fuzz test-contracts test-all test-production \
       test-production-verbose clean install lint vet coverage doctor setup setup-linux \
       dev localnet localnet-stop localnet-status localnet-logs localnet-clean \
       docker docker-dev docker-up docker-down release release-snapshot tidy check help

# ─── Configuration ────────────────────────────────────────────────────────────

VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"
FOUNDRY_BIN := $(HOME)/.foundry/bin

# ─── Build ────────────────────────────────────────────────────────────────────

build: daemon api cli

build-all: build exec-agent
	@echo "All binaries built in bin/"

daemon:
	@echo "Building daemon..."
	@go build $(LDFLAGS) -o bin/moltbunkerd ./cmd/daemon

api:
	@echo "Building API server..."
	@go build $(LDFLAGS) -o bin/moltbunker-api ./cmd/api

cli:
	@echo "Building CLI..."
	@go build $(LDFLAGS) -o bin/moltbunker ./cmd/cli

exec-agent:
	@echo "Building exec-agent (linux/amd64)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/exec-agent-amd64 ./cmd/exec-agent
	@echo "Building exec-agent (linux/arm64)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/exec-agent-arm64 ./cmd/exec-agent

# ─── Test ─────────────────────────────────────────────────────────────────────

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

test-colima:
	@echo "Running Colima E2E tests (requires Colima running)..."
	@go test -v -tags=colima -timeout 5m ./tests/e2e/colima/...

test-integration:
	@echo "Running integration tests..."
	@go test -v -tags=integration -timeout 10m ./tests/integration/...

test-localnet:
	@echo "Running localnet tests (requires mDNS)..."
	@go test -v -tags=localnet -timeout 5m ./tests/localnet/...

test-fuzz:
	@echo "Running fuzz tests (30s each)..."
	@go test -fuzz=FuzzEncrypt -fuzztime=30s ./internal/security/
	@go test -fuzz=FuzzDecrypt -fuzztime=30s ./internal/security/
	@go test -fuzz=FuzzMessageParsing -fuzztime=30s ./internal/p2p/

test-contracts:
	@echo "Running Foundry contract tests..."
	@cd contracts && $(FOUNDRY_BIN)/forge test -v

test-contracts-gas:
	@echo "Running contract tests with gas report..."
	@cd contracts && $(FOUNDRY_BIN)/forge test --gas-report

test-all: test test-e2e test-contracts
	@echo "All tests completed"

test-production: build
	@echo "Running production E2E tests..."
	@./scripts/e2e-production-test.sh

test-production-verbose: build
	@echo "Running production E2E tests (verbose)..."
	@./scripts/e2e-production-test.sh --verbose

# ─── Coverage ─────────────────────────────────────────────────────────────────

coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./internal/... ./pkg/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

coverage-text:
	@echo "Coverage summary:"
	@go test -coverprofile=coverage.out ./internal/... ./pkg/... > /dev/null 2>&1
	@go tool cover -func=coverage.out | tail -1

# ─── Code Quality ─────────────────────────────────────────────────────────────

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

tidy:
	@echo "Tidying modules..."
	@go mod tidy
	@go mod verify

check: tidy vet lint test
	@echo "All checks passed"

# ─── Localnet ─────────────────────────────────────────────────────────────────

localnet: build
	@./scripts/localnet.sh start

localnet-stop:
	@./scripts/localnet.sh stop

localnet-status:
	@./scripts/localnet.sh status

localnet-logs:
	@./scripts/localnet.sh logs

localnet-clean:
	@./scripts/localnet.sh clean

# ─── Docker ───────────────────────────────────────────────────────────────────

docker:
	@echo "Building Docker image..."
	@docker build -t moltbunker/moltbunker:$(VERSION) -t moltbunker/moltbunker:latest .

docker-dev:
	@echo "Building dev Docker image..."
	@docker build -f Dockerfile.dev -t moltbunker/moltbunker:dev .

docker-up:
	@echo "Starting Docker Compose stack..."
	@docker compose up -d

docker-down:
	@echo "Stopping Docker Compose stack..."
	@docker compose down

# ─── Release ──────────────────────────────────────────────────────────────────

release:
	@echo "Creating release..."
	@goreleaser release --clean

release-snapshot:
	@echo "Creating snapshot release (no publish)..."
	@goreleaser release --snapshot --clean

# ─── Setup ────────────────────────────────────────────────────────────────────

setup:
	@echo "Setting up development environment..."
	@./scripts/setup-dev-environment.sh

setup-linux:
	@echo "Setting up Linux production environment..."
	@sudo ./scripts/setup-linux.sh

# ─── Utilities ────────────────────────────────────────────────────────────────

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

install: build
	@echo "Installing binaries..."
	@cp bin/moltbunkerd /usr/local/bin/
	@cp bin/moltbunker-api /usr/local/bin/
	@cp bin/moltbunker /usr/local/bin/

doctor: cli
	@echo "Running doctor..."
	@./bin/moltbunker doctor

dev:
	@echo "Starting daemon in development mode..."
	@go run ./cmd/daemon --data-dir ~/.moltbunker-dev

# ─── Help ─────────────────────────────────────────────────────────────────────

help:
	@echo "Moltbunker Build System"
	@echo ""
	@echo "Build:"
	@echo "  build                Build daemon + API + CLI"
	@echo "  daemon               Build daemon only (bin/moltbunkerd)"
	@echo "  api                  Build API server only (bin/moltbunker-api)"
	@echo "  cli                  Build CLI only (bin/moltbunker)"
	@echo "  exec-agent           Build exec-agent for linux/amd64+arm64"
	@echo ""
	@echo "Test:"
	@echo "  test                 Run unit tests"
	@echo "  test-quick           Unit tests + smoke tests"
	@echo "  test-smoke           Smoke tests only"
	@echo "  test-e2e             Full E2E tests (mock services)"
	@echo "  test-colima          Colima E2E tests (real containers, macOS)"
	@echo "  test-integration     Integration tests"
	@echo "  test-localnet        Local network tests (requires mDNS)"
	@echo "  test-fuzz            Fuzz tests (30s per target)"
	@echo "  test-contracts       Foundry smart contract tests"
	@echo "  test-contracts-gas   Contract tests with gas report"
	@echo "  test-all             Unit + E2E + contract tests"
	@echo "  test-production      Production E2E (real dependencies)"
	@echo "  coverage             HTML coverage report"
	@echo "  coverage-text        Coverage percentage summary"
	@echo ""
	@echo "Quality:"
	@echo "  lint                 Run golangci-lint"
	@echo "  vet                  Run go vet"
	@echo "  tidy                 go mod tidy + verify"
	@echo "  check                tidy + vet + lint + test (pre-commit)"
	@echo ""
	@echo "Localnet:"
	@echo "  localnet             Build and start local network (Anvil + contracts + daemon + API)"
	@echo "  localnet-stop        Stop local network"
	@echo "  localnet-status      Show local network status"
	@echo "  localnet-logs        Tail local network logs"
	@echo "  localnet-clean       Stop and remove all localnet data"
	@echo ""
	@echo "Docker:"
	@echo "  docker               Build production Docker image"
	@echo "  docker-dev           Build development Docker image"
	@echo "  docker-up            Start Docker Compose stack"
	@echo "  docker-down          Stop Docker Compose stack"
	@echo ""
	@echo "Release:"
	@echo "  release              Create release with goreleaser"
	@echo "  release-snapshot     Dry-run release (no publish)"
	@echo ""
	@echo "Setup:"
	@echo "  setup                Setup macOS dev environment"
	@echo "  setup-linux          Setup Linux production environment"
	@echo "  install              Install binaries to /usr/local/bin"
	@echo "  doctor               Run system health check"
	@echo "  dev                  Start daemon in dev mode"
	@echo "  clean                Remove build artifacts"
