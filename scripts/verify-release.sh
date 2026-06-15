#!/usr/bin/env bash
#
# verify-release.sh — standalone verification of a downloaded moltbunker release.
#
# Use this when you have already downloaded one or more release archives (or
# want to re-verify after installing cosign). It downloads ONLY the release's
# checksums.txt (and the cosign bundle) for the given version, then:
#
#   1. (hard-required) verifies the SHA-256 of every local archive file in the
#      current directory that matches the given version;
#   2. (if cosign is on PATH) verifies the keyless Sigstore signature on
#      checksums.txt against the moltbunker release-workflow OIDC identity.
#
# Exits 0 only if all attempted checks pass. See docs/release-integrity.md.
#
# Usage:
#   ./verify-release.sh v1.2.3          # verify archives in CWD against v1.2.3
#   ./verify-release.sh v1.2.3 /tmp/dl  # verify archives in /tmp/dl
#
set -euo pipefail

REPO="${MOLTBUNKER_REPO:-moltbunker/moltbunker}"
GITHUB_DL="https://github.com/${REPO}/releases/download"
CERT_IDENTITY_REGEXP="^https://github.com/${REPO}/.github/workflows/release.yml@refs/tags/"
CERT_OIDC_ISSUER="https://token.actions.githubusercontent.com"

info()  { printf '\033[0;34m==>\033[0m %s\n' "$*"; }
warn()  { printf '\033[0;33mwarning:\033[0m %s\n' "$*" >&2; }
ok()    { printf '\033[0;32m  ok\033[0m %s\n' "$*"; }
die()   { printf '\033[0;31merror:\033[0m %s\n' "$*" >&2; exit 1; }

usage() {
  cat >&2 <<EOF
usage: $0 <version> [directory]

  <version>    release tag, e.g. v1.2.3
  [directory]  directory holding the downloaded archives (default: .)
EOF
  exit 2
}

download() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$1" -o "$2"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$2" "$1"
  else
    die "no downloader found (need curl or wget)"
  fi
}

sha256_check() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c "$1"
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 -c "$1"
  else
    die "no SHA-256 tool found (need sha256sum or shasum)"
  fi
}

main() {
  [ "$#" -ge 1 ] || usage
  local tag="$1"
  local dir="${2:-.}"
  local version="${tag#v}"

  [ -d "${dir}" ] || die "directory not found: ${dir}"

  local workdir
  workdir="$(mktemp -d)"
  # shellcheck disable=SC2064  # expand workdir now, intentionally.
  trap "rm -rf '${workdir}'" EXIT

  info "Fetching checksums.txt for ${tag}"
  download "${GITHUB_DL}/${tag}/checksums.txt" "${workdir}/checksums.txt" \
    || die "failed to download checksums.txt for ${tag}"

  # ── SHA-256 of every matching local archive ──
  local matched=0 line fname
  : > "${workdir}/to_check.txt"
  while IFS= read -r line; do
    # checksums.txt lines look like: "<hex>  moltbunker_<version>_<os>_<arch>.tar.gz"
    fname="${line##* }"
    case "${fname}" in
      moltbunker_"${version}"_*) ;;
      *) continue ;;
    esac
    if [ -f "${dir}/${fname}" ]; then
      echo "${line}" >> "${workdir}/to_check.txt"
      matched=$((matched + 1))
    fi
  done < "${workdir}/checksums.txt"

  [ "${matched}" -gt 0 ] \
    || die "no local archives for ${tag} found in ${dir} (looked for moltbunker_${version}_*)"

  info "Verifying SHA-256 for ${matched} local archive(s)"
  (
    cd "${dir}"
    sha256_check "${workdir}/to_check.txt"
  ) || die "SHA-256 verification FAILED"
  ok "all ${matched} archive(s) match checksums.txt"

  # ── cosign keyless signature on checksums.txt ──
  if command -v cosign >/dev/null 2>&1; then
    info "Verifying cosign signature on checksums.txt"
    download "${GITHUB_DL}/${tag}/checksums.txt.sigstore.json" \
      "${workdir}/checksums.txt.sigstore.json" \
      || die "failed to download cosign bundle for ${tag}"
    cosign verify-blob "${workdir}/checksums.txt" \
      --bundle "${workdir}/checksums.txt.sigstore.json" \
      --certificate-identity-regexp "${CERT_IDENTITY_REGEXP}" \
      --certificate-oidc-issuer "${CERT_OIDC_ISSUER}" \
      || die "cosign signature verification FAILED"
    ok "cosign signature verified (keyless / Sigstore)"
  else
    warn "cosign not found on PATH — skipped signature verification."
    warn "SHA-256 integrity is confirmed; install cosign to also verify provenance."
  fi

  ok "release ${tag} verified"
}

main "$@"
