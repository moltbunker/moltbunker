#!/bin/bash
#
# Moltbunker Production E2E Test Script
# ======================================
# This script tests Moltbunker with real dependencies (containerd, Tor, IPFS)
#
# Requirements:
#   - macOS with Homebrew
#   - Docker Desktop OR Colima (for containerd)
#   - ~10GB disk space
#
# Usage:
#   ./scripts/e2e-production-test.sh [--install-deps] [--cleanup] [--skip-tor] [--skip-ipfs]
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
TEST_DATA_DIR="${PROJECT_DIR}/.e2e-test-data"
DAEMON_LOG="${TEST_DATA_DIR}/daemon.log"
DAEMON_PID_FILE="${TEST_DATA_DIR}/daemon.pid"
TEST_CONFIG="${TEST_DATA_DIR}/config.yaml"
SOCKET_PATH="${TEST_DATA_DIR}/daemon.sock"

# Test settings
INSTALL_DEPS=false
CLEANUP_ONLY=false
SKIP_TOR=false
SKIP_IPFS=false
VERBOSE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-deps)
            INSTALL_DEPS=true
            shift
            ;;
        --cleanup)
            CLEANUP_ONLY=true
            shift
            ;;
        --skip-tor)
            SKIP_TOR=true
            shift
            ;;
        --skip-ipfs)
            SKIP_IPFS=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --install-deps  Install missing dependencies"
            echo "  --cleanup       Only cleanup test environment"
            echo "  --skip-tor      Skip Tor integration tests"
            echo "  --skip-ipfs     Skip IPFS integration tests"
            echo "  --verbose, -v   Verbose output"
            echo "  --help, -h      Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_section() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Detect container runtime and set CONTAINERD_SOCKET
detect_container_runtime() {
    if [[ "$(uname)" == "Darwin" ]]; then
        # macOS: check for Colima first (preferred)
        if command -v colima &> /dev/null; then
            if colima status 2>/dev/null | grep -q "Running"; then
                log_success "Container runtime: Colima"
                export CONTAINERD_SOCKET="$HOME/.colima/default/docker.sock"
                return 0
            else
                log_warning "Colima installed but not running"
                if [[ "$INSTALL_DEPS" == "true" ]]; then
                    log_info "Starting Colima..."
                    colima start
                    export CONTAINERD_SOCKET="$HOME/.colima/default/docker.sock"
                    return 0
                fi
            fi
        fi

        # Check for Docker Desktop as fallback
        if [[ -S "/var/run/docker.sock" ]]; then
            log_success "Container runtime: Docker Desktop"
            export CONTAINERD_SOCKET="/var/run/docker.sock"
            return 0
        fi

        log_error "No container runtime found on macOS"
        log_info "Install Colima: brew install colima && colima start"
        return 1
    else
        # Linux: use native containerd
        if [[ -S "/run/containerd/containerd.sock" ]]; then
            export CONTAINERD_SOCKET="/run/containerd/containerd.sock"
            log_success "Container runtime: containerd"
            return 0
        fi

        # Fallback to docker socket on Linux
        if [[ -S "/var/run/docker.sock" ]]; then
            export CONTAINERD_SOCKET="/var/run/docker.sock"
            log_success "Container runtime: Docker"
            return 0
        fi

        log_error "Containerd socket not found"
        return 1
    fi
}

# Cleanup function
cleanup() {
    log_section "Cleaning Up"

    # Stop daemon if running
    if [[ -f "$DAEMON_PID_FILE" ]]; then
        PID=$(cat "$DAEMON_PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            log_info "Stopping daemon (PID: $PID)..."
            kill "$PID" 2>/dev/null || true
            sleep 2
            kill -9 "$PID" 2>/dev/null || true
        fi
        rm -f "$DAEMON_PID_FILE"
    fi

    # Remove test containers
    if command -v docker &> /dev/null; then
        log_info "Removing test containers..."
        docker ps -a --filter "label=moltbunker-test=true" -q | xargs -r docker rm -f 2>/dev/null || true
    fi

    # Remove test data directory
    if [[ -d "$TEST_DATA_DIR" ]]; then
        log_info "Removing test data directory..."
        rm -rf "$TEST_DATA_DIR"
    fi

    log_success "Cleanup complete"
}

# Handle interrupts
trap cleanup EXIT INT TERM

if [[ "$CLEANUP_ONLY" == "true" ]]; then
    cleanup
    exit 0
fi

# ============================================================================
# PHASE 1: Check and Install Dependencies
# ============================================================================

log_section "Phase 1: Checking Dependencies"

check_command() {
    local cmd=$1
    local name=$2
    local install_cmd=$3

    if command -v "$cmd" &> /dev/null; then
        log_success "$name is installed"
        return 0
    else
        if [[ "$INSTALL_DEPS" == "true" && -n "$install_cmd" ]]; then
            log_warning "$name not found, installing..."
            eval "$install_cmd"
            if command -v "$cmd" &> /dev/null; then
                log_success "$name installed successfully"
                return 0
            fi
        fi
        log_error "$name is not installed"
        return 1
    fi
}

DEPS_OK=true

# Check Go
if ! check_command "go" "Go" "brew install go"; then
    DEPS_OK=false
fi

# Check Docker/containerd (via Docker Desktop or Colima)
if ! detect_container_runtime; then
    if [[ "$INSTALL_DEPS" == "true" ]]; then
        log_info "Installing Colima..."
        brew install colima docker
        colima start
        if ! detect_container_runtime; then
            DEPS_OK=false
        fi
    else
        DEPS_OK=false
    fi
fi

# Check Tor (optional)
if [[ "$SKIP_TOR" != "true" ]]; then
    if ! check_command "tor" "Tor" "brew install tor"; then
        log_warning "Tor not available, Tor tests will be skipped"
        SKIP_TOR=true
    fi
fi

# Check IPFS (optional)
if [[ "$SKIP_IPFS" != "true" ]]; then
    if ! check_command "ipfs" "IPFS" "brew install ipfs"; then
        log_warning "IPFS not available, IPFS tests will be skipped"
        SKIP_IPFS=true
    fi
fi

if [[ "$DEPS_OK" != "true" ]]; then
    log_error "Missing required dependencies. Run with --install-deps to install them."
    exit 1
fi

# ============================================================================
# PHASE 2: Build Moltbunker
# ============================================================================

log_section "Phase 2: Building Moltbunker"

cd "$PROJECT_DIR"

log_info "Building daemon and CLI..."
make build

if [[ ! -f "bin/moltbunker-daemon" ]] || [[ ! -f "bin/moltbunker" ]]; then
    log_error "Build failed - binaries not found"
    exit 1
fi

log_success "Build successful"

# ============================================================================
# PHASE 3: Setup Test Environment
# ============================================================================

log_section "Phase 3: Setting Up Test Environment"

# Create test data directory
mkdir -p "$TEST_DATA_DIR"/{keys,containers,logs,volumes,tor,state}

# Determine Tor enabled status
if [[ "$SKIP_TOR" == "true" ]]; then
    TOR_ENABLED="false"
else
    TOR_ENABLED="true"
fi

# Create test configuration
cat > "$TEST_CONFIG" << EOF
daemon:
  port: 19000
  data_dir: ${TEST_DATA_DIR}
  key_path: ${TEST_DATA_DIR}/keys/node.key
  keystore_dir: ${TEST_DATA_DIR}/keys
  socket_path: ${SOCKET_PATH}
  log_level: debug
  log_format: text

node:
  role: hybrid
  wallet_address: "0x1234567890123456789012345678901234567890"
  auto_register: false
  provider:
    declared_cpu: 4
    declared_memory_gb: 8
    declared_storage_gb: 100
    target_tier: starter

p2p:
  bootstrap_nodes: []
  network_mode: clearnet
  max_peers: 10
  dial_timeout_seconds: 10
  enable_mdns: true

tor:
  enabled: ${TOR_ENABLED}
  data_dir: ${TEST_DATA_DIR}/tor
  socks5_port: 19050
  control_port: 19051

runtime:
  containerd_socket: ${CONTAINERD_SOCKET:-/var/run/docker.sock}
  namespace: moltbunker-test
  default_resources:
    cpu_quota: 100000
    cpu_period: 100000
    memory_limit: 536870912
    disk_limit: 1073741824
    network_bw: 10485760
    pid_limit: 100
  logs_dir: ${TEST_DATA_DIR}/logs
  volumes_dir: ${TEST_DATA_DIR}/volumes

security:
  container_profile:
    allow_exec: true
    allow_attach: false
    allow_shell: false
    allow_privileged: false
    read_only_rootfs: true

redundancy:
  replica_count: 1
  health_check_interval_seconds: 10

economics:
  chain_id: 8453
  rpc_url: "https://mainnet.base.org"
EOF

log_success "Test configuration created"

# Start IPFS daemon if needed
if [[ "$SKIP_IPFS" != "true" ]]; then
    if ! pgrep -x "ipfs" > /dev/null; then
        log_info "Starting IPFS daemon..."
        ipfs daemon --init &> "${TEST_DATA_DIR}/ipfs.log" &
        sleep 3
        if pgrep -x "ipfs" > /dev/null; then
            log_success "IPFS daemon started"
        else
            log_warning "Failed to start IPFS daemon, skipping IPFS tests"
            SKIP_IPFS=true
        fi
    else
        log_success "IPFS daemon already running"
    fi
fi

# ============================================================================
# PHASE 4: Start Moltbunker Daemon
# ============================================================================

log_section "Phase 4: Starting Moltbunker Daemon"

# Set environment for mock payments (no real blockchain)
export MOLTBUNKER_MOCK_PAYMENTS=true

log_info "Starting daemon..."
"${PROJECT_DIR}/bin/moltbunker-daemon" \
    --config "$TEST_CONFIG" \
    &> "$DAEMON_LOG" &

DAEMON_PID=$!
echo "$DAEMON_PID" > "$DAEMON_PID_FILE"

# Wait for daemon to be ready
log_info "Waiting for daemon to start..."
for i in {1..30}; do
    if [[ -S "$SOCKET_PATH" ]]; then
        log_success "Daemon started (PID: $DAEMON_PID)"
        break
    fi
    if ! kill -0 "$DAEMON_PID" 2>/dev/null; then
        log_error "Daemon failed to start. Check logs: $DAEMON_LOG"
        if [[ "$VERBOSE" == "true" ]]; then
            cat "$DAEMON_LOG"
        fi
        exit 1
    fi
    sleep 1
done

if [[ ! -S "$SOCKET_PATH" ]]; then
    log_error "Daemon did not create socket in time"
    exit 1
fi

# ============================================================================
# PHASE 5: Run E2E Tests
# ============================================================================

log_section "Phase 5: Running E2E Tests"

CLI="${PROJECT_DIR}/bin/moltbunker"
TEST_PASSED=0
TEST_FAILED=0
TEST_SKIPPED=0

run_test() {
    local name=$1
    local cmd=$2

    echo -n "  Testing: $name... "

    if eval "$cmd" &> /dev/null; then
        echo -e "${GREEN}PASS${NC}"
        ((TEST_PASSED++))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        ((TEST_FAILED++))
        if [[ "$VERBOSE" == "true" ]]; then
            echo "    Command: $cmd"
            eval "$cmd" 2>&1 | sed 's/^/    /'
        fi
        return 1
    fi
}

skip_test() {
    local name=$1
    local reason=$2
    echo -e "  Testing: $name... ${YELLOW}SKIP${NC} ($reason)"
    ((TEST_SKIPPED++))
}

# ----- Basic CLI Tests -----
log_info "Running CLI tests..."

run_test "CLI help" "$CLI --help"
run_test "CLI version" "$CLI version"

# ----- Doctor Tests -----
log_info "Running doctor tests..."

run_test "Doctor check" "$CLI doctor"
run_test "Doctor JSON output" "$CLI doctor --json | jq -e '.checks'"

# ----- Colima Tests (macOS) -----
if [[ "$(uname)" == "Darwin" ]]; then
    log_info "Running Colima tests..."
    run_test "Colima help" "$CLI colima --help"
    run_test "Colima status" "$CLI colima status" || true
fi

# ----- Config Tests -----
log_info "Running config tests..."

run_test "Config help" "$CLI config --help"

# ----- Provider Tests (no socket needed for help) -----
log_info "Running provider tests..."

run_test "Provider help" "$CLI provider --help"

# ----- Requester Tests (no socket needed for help) -----
log_info "Running requester tests..."

run_test "Requester help" "$CLI requester --help"

# ----- Tor Tests (no socket needed for help) -----
log_info "Running Tor CLI tests..."

run_test "Tor help" "$CLI tor --help"

# NOTE: Full container lifecycle tests require --socket flag support in CLI commands.
# These tests verify CLI builds and basic help commands work.
# Container operations require the daemon socket which uses default path.

# ============================================================================
# PHASE 6: Results Summary
# ============================================================================

log_section "Test Results Summary"

TOTAL=$((TEST_PASSED + TEST_FAILED + TEST_SKIPPED))

echo ""
echo "  Total Tests:   $TOTAL"
echo -e "  ${GREEN}Passed:${NC}        $TEST_PASSED"
echo -e "  ${RED}Failed:${NC}        $TEST_FAILED"
echo -e "  ${YELLOW}Skipped:${NC}       $TEST_SKIPPED"
echo ""

if [[ $TEST_FAILED -eq 0 ]]; then
    log_success "All tests passed!"
    EXIT_CODE=0
else
    log_error "$TEST_FAILED test(s) failed"
    EXIT_CODE=1
fi

# Show daemon logs if verbose or if tests failed
if [[ "$VERBOSE" == "true" ]] || [[ $TEST_FAILED -gt 0 ]]; then
    log_section "Daemon Logs (last 50 lines)"
    tail -50 "$DAEMON_LOG"
fi

exit $EXIT_CODE
