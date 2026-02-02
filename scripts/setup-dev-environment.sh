#!/bin/bash
#
# Moltbunker Development Environment Setup
# ========================================
# This script sets up all dependencies needed for production-like E2E testing
#
# Usage:
#   ./scripts/setup-dev-environment.sh
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  Moltbunker Development Environment Setup"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Check if running on macOS
if [[ "$(uname)" != "Darwin" ]]; then
    log_error "This script is designed for macOS. For Linux, please install dependencies manually."
    exit 1
fi

# Check for Homebrew
if ! command -v brew &> /dev/null; then
    log_info "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi

log_success "Homebrew is available"

# ============================================================================
# Core Dependencies
# ============================================================================

echo ""
log_info "Installing core dependencies..."

# Go
if ! command -v go &> /dev/null; then
    log_info "Installing Go..."
    brew install go
else
    GO_VERSION=$(go version | awk '{print $3}')
    log_success "Go is installed ($GO_VERSION)"
fi

# jq (for JSON parsing in tests)
if ! command -v jq &> /dev/null; then
    log_info "Installing jq..."
    brew install jq
else
    log_success "jq is installed"
fi

# ============================================================================
# Container Runtime
# ============================================================================

echo ""
log_info "Setting up container runtime..."

# Check for Docker Desktop or Colima
CONTAINER_RUNTIME_OK=false

if command -v docker &> /dev/null && docker info &> /dev/null 2>&1; then
    log_success "Docker Desktop is running"
    CONTAINER_RUNTIME_OK=true
elif command -v colima &> /dev/null; then
    if colima status 2>/dev/null | grep -q "Running"; then
        log_success "Colima is running"
        CONTAINER_RUNTIME_OK=true
    else
        log_info "Starting Colima..."
        colima start
        CONTAINER_RUNTIME_OK=true
    fi
fi

if [[ "$CONTAINER_RUNTIME_OK" != "true" ]]; then
    echo ""
    echo "No container runtime detected. Choose an option:"
    echo ""
    echo "  1) Install Docker Desktop (GUI, easiest)"
    echo "  2) Install Colima (CLI, lightweight)"
    echo "  3) Skip (tests will use mocks only)"
    echo ""
    read -p "Enter choice [1-3]: " choice

    case $choice in
        1)
            log_info "Installing Docker Desktop..."
            brew install --cask docker
            echo ""
            log_warning "Please launch Docker Desktop from Applications and wait for it to start."
            log_info "Then run this script again or run: ./scripts/e2e-production-test.sh"
            ;;
        2)
            log_info "Installing Colima and Docker CLI..."
            brew install colima docker
            log_info "Starting Colima..."
            colima start
            log_success "Colima is running"
            ;;
        3)
            log_warning "Skipping container runtime installation"
            log_info "E2E tests will use mock implementations"
            ;;
        *)
            log_warning "Invalid choice, skipping"
            ;;
    esac
fi

# ============================================================================
# Tor (Optional)
# ============================================================================

echo ""
log_info "Setting up Tor (optional)..."

if ! command -v tor &> /dev/null; then
    read -p "Install Tor for anonymous networking tests? [y/N]: " install_tor
    if [[ "$install_tor" =~ ^[Yy]$ ]]; then
        log_info "Installing Tor..."
        brew install tor
        log_success "Tor installed"
    else
        log_info "Skipping Tor installation"
    fi
else
    log_success "Tor is installed"
fi

# ============================================================================
# IPFS (Optional)
# ============================================================================

echo ""
log_info "Setting up IPFS (optional)..."

if ! command -v ipfs &> /dev/null; then
    read -p "Install IPFS for distributed storage tests? [y/N]: " install_ipfs
    if [[ "$install_ipfs" =~ ^[Yy]$ ]]; then
        log_info "Installing IPFS..."
        brew install ipfs

        # Initialize IPFS if needed
        if [[ ! -d ~/.ipfs ]]; then
            log_info "Initializing IPFS..."
            ipfs init
        fi
        log_success "IPFS installed and initialized"
    else
        log_info "Skipping IPFS installation"
    fi
else
    log_success "IPFS is installed"
    # Initialize if needed
    if [[ ! -d ~/.ipfs ]]; then
        log_info "Initializing IPFS..."
        ipfs init
    fi
fi

# ============================================================================
# Development Tools (Optional)
# ============================================================================

echo ""
log_info "Setting up development tools (optional)..."

# golangci-lint
if ! command -v golangci-lint &> /dev/null; then
    read -p "Install golangci-lint for code linting? [y/N]: " install_lint
    if [[ "$install_lint" =~ ^[Yy]$ ]]; then
        log_info "Installing golangci-lint..."
        brew install golangci-lint
        log_success "golangci-lint installed"
    fi
else
    log_success "golangci-lint is installed"
fi

# ============================================================================
# Build Project
# ============================================================================

echo ""
log_info "Building Moltbunker..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

make build

if [[ -f "bin/moltbunker" ]] && [[ -f "bin/moltbunker-daemon" ]]; then
    log_success "Build successful"
else
    log_error "Build failed"
    exit 1
fi

# ============================================================================
# Summary
# ============================================================================

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  Setup Complete!"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "  Installed components:"

command -v go &> /dev/null && echo "    ✓ Go $(go version | awk '{print $3}')"
command -v docker &> /dev/null && docker info &> /dev/null 2>&1 && echo "    ✓ Docker"
command -v colima &> /dev/null && colima status 2>/dev/null | grep -q "Running" && echo "    ✓ Colima"
command -v tor &> /dev/null && echo "    ✓ Tor"
command -v ipfs &> /dev/null && echo "    ✓ IPFS"
command -v golangci-lint &> /dev/null && echo "    ✓ golangci-lint"
echo "    ✓ Moltbunker (bin/moltbunker, bin/moltbunker-daemon)"

echo ""
echo "  Next steps:"
echo ""
echo "    # Run unit tests"
echo "    make test-all"
echo ""
echo "    # Run production E2E tests"
echo "    ./scripts/e2e-production-test.sh"
echo ""
echo "    # Run production E2E tests with verbose output"
echo "    ./scripts/e2e-production-test.sh --verbose"
echo ""
echo "    # Run doctor to verify system"
echo "    ./bin/moltbunker doctor"
echo ""
