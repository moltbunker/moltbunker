.PHONY: build build-all daemon api cli exec-agent test test-quick test-smoke test-e2e test-colima \
       test-containerd-linux test-integration test-localnet test-fuzz test-contracts test-all test-production \
       test-production-verbose clean install lint vet coverage doctor setup setup-linux \
       dev localnet localnet-stop localnet-status localnet-logs localnet-clean \
       docker docker-dev docker-up docker-down release release-snapshot tidy check help \
       gen-addresses gen-addresses-check gen-bindings bindings-check

# ─── Configuration ────────────────────────────────────────────────────────────

VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS_DAEMON := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"
LDFLAGS_CLI    := -ldflags "-X github.com/moltbunker/moltbunker/cmd/cli/commands.Version=$(VERSION) -X github.com/moltbunker/moltbunker/cmd/cli/commands.Commit=$(COMMIT) -X github.com/moltbunker/moltbunker/cmd/cli/commands.BuildDate=$(BUILD_DATE)"
FOUNDRY_BIN := $(HOME)/.foundry/bin

# ─── Build ────────────────────────────────────────────────────────────────────

build: daemon api cli

build-all: build exec-agent
	@echo "All binaries built in bin/"

daemon:
	@echo "Building daemon..."
	@go build $(LDFLAGS_DAEMON) -o bin/moltbunkerd ./cmd/daemon

api:
	@echo "Building API server..."
	@go build $(LDFLAGS_DAEMON) -o bin/moltbunker-api ./cmd/api

cli:
	@echo "Building CLI..."
	@go build $(LDFLAGS_CLI) -o bin/moltbunker ./cmd/cli

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

test-containerd-linux:
	@echo "Running real-containerd E2E tests against the system containerd (Linux/CI parity)..."
	@echo "Requires containerd at /run/containerd/containerd.sock (or set COLIMA_CONTAINERD_SOCKET)."
	@go test -v -tags=colima -timeout 12m ./tests/e2e/colima/...

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

# ─── Contract Bindings (codegen) ────────────────────────────────────────────────
# The generated Go bindings in internal/payment/bindings/ are COMMITTED, so a
# normal `make build` needs only the Go toolchain. Run gen-bindings only when the
# Solidity contracts change; it requires forge, jq, and abigen on PATH.

gen-bindings:
	@command -v jq >/dev/null 2>&1 || { echo "error: 'jq' not on PATH (macOS: brew install jq)"; exit 1; }
	@command -v abigen >/dev/null 2>&1 || { echo "error: 'abigen' not on PATH (go install github.com/ethereum/go-ethereum/cmd/abigen@latest)"; exit 1; }
	@command -v forge >/dev/null 2>&1 || [ -x "$(FOUNDRY_BIN)/forge" ] || { echo "error: 'forge' not on PATH (https://getfoundry.sh)"; exit 1; }
	@echo "Generating contract bindings (forge build -> abigen -> internal/payment/bindings/)..."
	@PATH="$(FOUNDRY_BIN):$$PATH" go generate ./internal/payment/...
	@echo "Bindings generated. Review the diff before committing."

bindings-check: gen-bindings
	@echo "Checking committed bindings match a fresh generation..."
	@git diff --exit-code -- internal/payment/bindings/ \
		|| { echo "error: committed bindings differ from generated output; run 'make gen-bindings' and commit."; exit 1; }
	@echo "Bindings are up to date."

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

# ─── Codegen ──────────────────────────────────────────────────────────────────

# Canonical contract-address manifest. Single source of truth; mainnet cutover
# = edit this file then `make gen-addresses` and commit the diff.
ADDR_MANIFEST    ?= deployments/addresses.json
ADDR_OUT_YAML    ?= configs/addresses-fragment.yaml
# Cross-repo TS / env emitters are OPT-IN (scope: this repo by default). Point
# these at the sibling web/, web-admin/ checkouts to regenerate those consumers,
# e.g. `make gen-addresses ADDR_OUT_WEB=../web/src/lib/generated-addresses.ts`.
ADDR_OUT_WEB     ?=
ADDR_OUT_ADMIN   ?=
ADDR_OUT_ADMIN_ENV ?=

gen-addresses:
	@echo "Generating contract-address artifacts from $(ADDR_MANIFEST)..."
	@go run ./tools/gen-addresses \
		--manifest $(ADDR_MANIFEST) \
		--out-yaml $(ADDR_OUT_YAML) \
		--out-web "$(ADDR_OUT_WEB)" \
		--out-admin "$(ADDR_OUT_ADMIN)" \
		--out-admin-env "$(ADDR_OUT_ADMIN_ENV)"
	@echo "Wrote $(ADDR_OUT_YAML)"

# gen-addresses-check regenerates the in-repo artifacts and fails if they are
# stale relative to deployments/addresses.json. Intended for CI once stable.
gen-addresses-check: gen-addresses
	@git diff --quiet -- $(ADDR_OUT_YAML) $(ADDR_MANIFEST) || \
		{ echo "ERROR: generated address artifacts are stale. Run 'make gen-addresses' and commit."; exit 1; }
	@echo "Generated address artifacts are up to date"

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
	@echo "Codegen:"
	@echo "  gen-addresses        Regenerate contract-address artifacts from deployments/addresses.json"
	@echo "                       (in-repo YAML by default; set ADDR_OUT_WEB/ADDR_OUT_ADMIN/ADDR_OUT_ADMIN_ENV"
	@echo "                        to also emit the web/web-admin TS + .env.example)"
	@echo "  gen-addresses-check  Fail if generated address artifacts are stale (CI lint)"
	@echo "Generate:"
	@echo "  gen-bindings         Regenerate Go contract bindings (needs forge, jq, abigen)"
	@echo "  bindings-check       Verify committed bindings match a fresh generation (drift)"
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
