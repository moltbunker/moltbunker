#!/usr/bin/env bash
#
# moltbunker installer (Linux / macOS).
#
# Supply-chain integrity: this script NEVER trusts a raw binary blindly. Every
# install path downloads the release archive AND the signed checksums.txt, then:
#
#   1. (always, hard-required) verifies the archive's SHA-256 against the
#      checksums.txt that ships with the release;
#   2. (best-effort) if `cosign` is on PATH, verifies that checksums.txt itself
#      was keyless-signed (Sigstore OIDC) by the moltbunker GitHub Actions
#      release workflow, anchoring the whole release to a public identity.
#
# The SHA-256 check is mandatory and aborts on mismatch. The cosign check is
# conditionally mandatory: if `cosign` is on PATH it MUST succeed (a missing or
# unverifiable signature bundle aborts the install — set MOLTBUNKER_SKIP_COSIGN=1
# to opt out), so an attacker cannot 404 the bundle to strip signature checking
# and downgrade you to attacker-controlled checksums. A host WITHOUT cosign still
# gets integrity against corruption / CDN cache poisoning, the most common attack
# vector. Install cosign to also defend against a compromised release page. See
# docs/release-integrity.md.
#
# Usage:
#   curl -fsSL https://moltbunker.dev/install.sh | bash
#   ./install.sh                 # latest release
#   VERSION=v1.2.3 ./install.sh  # pin a version
#
set -euo pipefail

# ─── Configuration ──────────────────────────────────────────────────────────

REPO="${MOLTBUNKER_REPO:-moltbunker/moltbunker}"
INSTALL_DIR="${MOLTBUNKER_INSTALL_DIR:-/usr/local/bin}"
GITHUB_API="https://api.github.com/repos/${REPO}"
GITHUB_DL="https://github.com/${REPO}/releases/download"

# The cosign identity that signed checksums.txt. This is the OIDC claim path of
# the release workflow — public, version-controlled, and the trust anchor for
# keyless verification. Do NOT change without coordinating with .github.
CERT_IDENTITY_REGEXP="^https://github.com/${REPO}/.github/workflows/release.yml@refs/tags/"
CERT_OIDC_ISSUER="https://token.actions.githubusercontent.com"

# ─── Output helpers ───────────────────────────────────────────────────────────

info()  { printf '\033[0;34m==>\033[0m %s\n' "$*"; }
warn()  { printf '\033[0;33mwarning:\033[0m %s\n' "$*" >&2; }
ok()    { printf '\033[0;32m  ok\033[0m %s\n' "$*"; }
die()   { printf '\033[0;31merror:\033[0m %s\n' "$*" >&2; exit 1; }

# ─── Platform detection ───────────────────────────────────────────────────────

detect_os() {
  local os
  os="$(uname -s)"
  case "${os}" in
    Linux)  echo "linux" ;;
    Darwin) echo "darwin" ;;
    *)      die "unsupported OS: ${os} (moltbunker supports Linux and macOS)" ;;
  esac
}

# Maps `uname -m` to the GoReleaser archive arch token. The .goreleaser.yml
# name_template rewrites amd64 -> x86_64; everything else passes through.
detect_arch() {
  local arch
  arch="$(uname -m)"
  case "${arch}" in
    x86_64|amd64) echo "x86_64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) die "unsupported architecture: ${arch} (moltbunker supports x86_64 and arm64)" ;;
  esac
}

# ─── Dependency checks ────────────────────────────────────────────────────────

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

# Selects a SHA-256 checker. GNU coreutils ships `sha256sum`; macOS ships
# `shasum`. Both support the `-c` (check-against-list) mode we rely on.
sha256_check() {
  # $1 = checksum list file (already filtered to the target archive line)
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c "$1"
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 -c "$1"
  else
    die "no SHA-256 tool found (need sha256sum or shasum)"
  fi
}

# Picks a downloader. curl preferred; wget fallback.
download() {
  # $1 = url, $2 = output path
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$1" -o "$2"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$2" "$1"
  else
    die "no downloader found (need curl or wget)"
  fi
}

# ─── Version resolution ───────────────────────────────────────────────────────

resolve_version() {
  if [ -n "${VERSION:-}" ]; then
    echo "${VERSION}"
    return
  fi
  local tmp tag
  tmp="$(mktemp)"
  download "${GITHUB_API}/releases/latest" "${tmp}" \
    || die "could not reach GitHub API to resolve the latest release"
  # Extract "tag_name": "vX.Y.Z" without requiring jq.
  tag="$(grep -m1 '"tag_name"' "${tmp}" | sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/')"
  rm -f "${tmp}"
  [ -n "${tag}" ] || die "could not parse latest release tag from GitHub API"
  echo "${tag}"
}

# ─── Core: download + verify + install ────────────────────────────────────────

download_and_verify() {
  local os arch tag version
  os="$(detect_os)"
  arch="$(detect_arch)"
  tag="$(resolve_version)"
  # GoReleaser archive names use the version WITHOUT the leading "v".
  version="${tag#v}"

  local archive="moltbunker_${version}_${os}_${arch}.tar.gz"
  local archive_url="${GITHUB_DL}/${tag}/${archive}"
  local checksums_url="${GITHUB_DL}/${tag}/checksums.txt"
  local bundle_url="${GITHUB_DL}/${tag}/checksums.txt.sigstore.json"

  local workdir
  workdir="$(mktemp -d)"
  # shellcheck disable=SC2064  # expand workdir now, intentionally.
  trap "rm -rf '${workdir}'" EXIT

  info "Installing moltbunker ${tag} (${os}/${arch})"

  info "Downloading ${archive}"
  download "${archive_url}" "${workdir}/${archive}" \
    || die "failed to download archive: ${archive_url}"

  info "Downloading checksums.txt"
  download "${checksums_url}" "${workdir}/checksums.txt" \
    || die "failed to download checksums.txt"

  # ── Step 1: SHA-256 (hard-required) ──
  info "Verifying SHA-256 checksum"
  # Filter checksums.txt to ONLY the line for our archive, so the -c mode
  # neither fails on (nor verifies) the other archives we did not download.
  grep " ${archive}\$" "${workdir}/checksums.txt" > "${workdir}/checksums.filtered" \
    || die "archive ${archive} not listed in checksums.txt (wrong version/platform?)"
  (
    cd "${workdir}"
    sha256_check "checksums.filtered"
  ) || die "SHA-256 checksum verification FAILED for ${archive} — refusing to install"
  ok "checksum verified"

  # ── Step 2: cosign keyless ──
  # When cosign IS present, the signature is MANDATORY: a missing/404'd bundle
  # is a hard failure, not a silent downgrade to checksum-only. Otherwise an
  # attacker who can serve a tampered checksums.txt (compromised release page /
  # CDN MITM) could also 404 the bundle to strip the cosign check entirely, and
  # the SHA-256 would then validate against the ATTACKER's checksums.txt — the
  # exact threat this script's header claims to defend against. This mirrors
  # verify-release.sh's hard-fail posture. Set MOLTBUNKER_SKIP_COSIGN=1 to
  # explicitly opt out (e.g. air-gapped re-install of an already-trusted tag).
  if command -v cosign >/dev/null 2>&1 && [ -z "${MOLTBUNKER_SKIP_COSIGN:-}" ]; then
    info "Verifying cosign signature on checksums.txt"
    download "${bundle_url}" "${workdir}/checksums.txt.sigstore.json" \
      || die "cosign present but cosign bundle download failed (${bundle_url}) — refusing to install (set MOLTBUNKER_SKIP_COSIGN=1 to override)"
    if cosign verify-blob "${workdir}/checksums.txt" \
      --bundle "${workdir}/checksums.txt.sigstore.json" \
      --certificate-identity-regexp "${CERT_IDENTITY_REGEXP}" \
      --certificate-oidc-issuer "${CERT_OIDC_ISSUER}" >/dev/null 2>&1; then
      ok "cosign signature verified (keyless / Sigstore)"
    else
      die "cosign signature verification FAILED — refusing to install"
    fi
  elif command -v cosign >/dev/null 2>&1; then
    warn "cosign found but MOLTBUNKER_SKIP_COSIGN is set — skipping signature verification."
    warn "SHA-256 integrity is verified, but signature provenance is NOT."
  else
    warn "cosign not found on PATH — skipping signature verification."
    warn "SHA-256 integrity is verified, but signature provenance is not."
    warn "Install cosign and run scripts/verify-release.sh to fully verify."
  fi

  # ── Step 3: extract + install ──
  info "Extracting archive"
  tar -xzf "${workdir}/${archive}" -C "${workdir}" \
    || die "failed to extract ${archive}"

  install_binaries "${workdir}"
}

install_binaries() {
  local src="$1"
  local sudo=""
  # If we cannot write to INSTALL_DIR, escalate with sudo (if present).
  if [ ! -w "${INSTALL_DIR}" ]; then
    if command -v sudo >/dev/null 2>&1; then
      sudo=sudo
    else
      die "cannot write to ${INSTALL_DIR} and sudo is not available; set MOLTBUNKER_INSTALL_DIR to a writable path"
    fi
  fi

  local installed=0 bin
  for bin in moltbunker moltbunker-daemon moltbunker-api exec-agent; do
    if [ -f "${src}/${bin}" ]; then
      info "Installing ${bin} -> ${INSTALL_DIR}/${bin}"
      ${sudo} install -m 0755 "${src}/${bin}" "${INSTALL_DIR}/${bin}" \
        || die "failed to install ${bin}"
      installed=$((installed + 1))
    fi
  done
  [ "${installed}" -gt 0 ] || die "no moltbunker binaries found in the archive"
  ok "installed ${installed} binary/binaries to ${INSTALL_DIR}"
}

# ─── macOS Homebrew fast-path ─────────────────────────────────────────────────

try_homebrew() {
  # Homebrew is the preferred macOS UX. The tap formula is verified by
  # Homebrew's own checksum mechanism, so it satisfies our integrity bar.
  if [ "$(detect_os)" = "darwin" ] && command -v brew >/dev/null 2>&1 \
     && [ -z "${MOLTBUNKER_NO_BREW:-}" ]; then
    info "Homebrew detected — installing via tap (set MOLTBUNKER_NO_BREW=1 to force direct download)"
    brew tap moltbunker/homebrew-tap 2>/dev/null || true
    if brew install moltbunker; then
      ok "installed via Homebrew"
      return 0
    fi
    warn "brew install failed; falling back to verified direct download"
  fi
  return 1
}

# ─── Main ─────────────────────────────────────────────────────────────────────

main() {
  need_cmd tar
  if try_homebrew; then
    exit 0
  fi
  download_and_verify
  cat <<'EOF'

moltbunker installed.

  moltbunker --help            # CLI
  moltbunker-daemon --help     # node daemon

To re-verify a release at any time (incl. cosign signature provenance), see:
  https://github.com/moltbunker/moltbunker/blob/main/docs/release-integrity.md
EOF
}

main "$@"
