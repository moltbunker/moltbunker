<#
.SYNOPSIS
  moltbunker installer for Windows.

.DESCRIPTION
  Supply-chain integrity: this script never trusts a raw binary blindly. It
  downloads the release archive AND the signed checksums.txt, then:

    1. (always, hard-required) verifies the archive's SHA-256 against
       checksums.txt;
    2. (best-effort) if cosign is on PATH, verifies that checksums.txt was
       keyless-signed (Sigstore OIDC) by the moltbunker GitHub Actions release
       workflow.

  The SHA-256 check is mandatory and aborts on mismatch. The cosign check is
  advisory. See docs/release-integrity.md.

  NOTE: the release pipeline does not build a windows/arm64 archive
  (.goreleaser.yml ignores goos=windows goarch=arm64), so only x86_64 (amd64)
  is installable on Windows.

.PARAMETER Version
  Release tag to install, e.g. v1.2.3. Defaults to the latest release.

.EXAMPLE
  iwr -useb https://moltbunker.dev/install.ps1 | iex
  .\install.ps1 -Version v1.2.3
#>
param(
  [string]$Version = "",
  [string]$InstallDir = "$env:LOCALAPPDATA\moltbunker\bin"
)

$ErrorActionPreference = 'Stop'

$Repo = if ($env:MOLTBUNKER_REPO) { $env:MOLTBUNKER_REPO } else { "moltbunker/moltbunker" }
$GitHubApi = "https://api.github.com/repos/$Repo"
$GitHubDl  = "https://github.com/$Repo/releases/download"

# Public OIDC identity that signed checksums.txt — the trust anchor for keyless
# verification. Do NOT change without coordinating with .github/workflows.
$CertIdentityRegexp = "^https://github.com/$Repo/\.github/workflows/release\.yml@refs/tags/"
$CertOidcIssuer     = "https://token.actions.githubusercontent.com"

function Write-Info($msg) { Write-Host "==> $msg" -ForegroundColor Blue }
function Write-Ok($msg)   { Write-Host "  ok $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Warning $msg }
function Die($msg)        { Write-Error $msg; exit 1 }

function Get-Arch {
  # Only x86_64 archives are published for Windows (see note above).
  $arch = (Get-CimInstance Win32_Processor | Select-Object -First 1).Architecture
  switch ($arch) {
    9 { return "x86_64" }   # x64
    default {
      Die "unsupported Windows architecture (only x86_64 release archives are published)"
    }
  }
}

function Resolve-Version {
  if ($Version) { return $Version }
  Write-Info "Resolving latest release"
  $rel = Invoke-RestMethod -Uri "$GitHubApi/releases/latest" -Headers @{ "User-Agent" = "moltbunker-install" }
  if (-not $rel.tag_name) { Die "could not resolve latest release tag" }
  return $rel.tag_name
}

# Winget fast-path: if a moltbunker winget package is available, prefer it.
function Try-Winget {
  if ($env:MOLTBUNKER_NO_WINGET) { return $false }
  if (-not (Get-Command winget -ErrorAction SilentlyContinue)) { return $false }
  Write-Info "winget detected — attempting winget install (set MOLTBUNKER_NO_WINGET=1 to force direct download)"
  try {
    winget install --id moltbunker.moltbunker --accept-source-agreements --accept-package-agreements -e
    if ($LASTEXITCODE -eq 0) {
      Write-Ok "installed via winget"
      return $true
    }
  } catch {
    Write-Warn "winget install failed; falling back to verified direct download"
  }
  return $false
}

function DownloadAndVerify {
  $arch = Get-Arch
  $tag = Resolve-Version
  $ver = $tag.TrimStart('v')

  $archive   = "moltbunker_${ver}_windows_${arch}.zip"
  $archiveUrl = "$GitHubDl/$tag/$archive"
  $checksumsUrl = "$GitHubDl/$tag/checksums.txt"
  $bundleUrl    = "$GitHubDl/$tag/checksums.txt.sigstore.json"

  $work = Join-Path $env:TEMP ("moltbunker-install-" + [System.Guid]::NewGuid().ToString("N"))
  New-Item -ItemType Directory -Path $work -Force | Out-Null

  try {
    Write-Info "Installing moltbunker $tag (windows/$arch)"

    $archivePath   = Join-Path $work $archive
    $checksumsPath = Join-Path $work "checksums.txt"

    Write-Info "Downloading $archive"
    Invoke-WebRequest -Uri $archiveUrl -OutFile $archivePath -UseBasicParsing

    Write-Info "Downloading checksums.txt"
    Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath -UseBasicParsing

    # ── Step 1: SHA-256 (hard-required) ──
    Write-Info "Verifying SHA-256 checksum"
    $expectedLine = (Get-Content $checksumsPath | Where-Object { $_ -match [regex]::Escape($archive) + '$' } | Select-Object -First 1)
    if (-not $expectedLine) {
      Die "archive $archive not listed in checksums.txt (wrong version/platform?)"
    }
    $expected = ($expectedLine -split '\s+')[0].ToLower()
    $actual = (Get-FileHash -Algorithm SHA256 -Path $archivePath).Hash.ToLower()
    if ($actual -ne $expected) {
      Die "SHA-256 checksum verification FAILED for $archive (expected $expected, got $actual) — refusing to install"
    }
    Write-Ok "checksum verified"

    # ── Step 2: cosign keyless ──
    # When cosign IS present, the signature is MANDATORY: a missing/404'd
    # bundle is a hard failure, not a silent downgrade to checksum-only.
    # Otherwise an attacker who can serve a tampered checksums.txt could also
    # 404 the bundle to strip the cosign check, and the SHA-256 would then
    # validate against the ATTACKER's checksums.txt. Mirrors install.sh /
    # verify-release.sh. Set MOLTBUNKER_SKIP_COSIGN=1 to explicitly opt out.
    if ((Get-Command cosign -ErrorAction SilentlyContinue) -and (-not $env:MOLTBUNKER_SKIP_COSIGN)) {
      Write-Info "Verifying cosign signature on checksums.txt"
      $bundlePath = Join-Path $work "checksums.txt.sigstore.json"
      try {
        Invoke-WebRequest -Uri $bundleUrl -OutFile $bundlePath -UseBasicParsing
      } catch {
        Die "cosign present but cosign bundle download failed ($bundleUrl) — refusing to install (set MOLTBUNKER_SKIP_COSIGN=1 to override)"
      }
      & cosign verify-blob $checksumsPath `
        --bundle $bundlePath `
        --certificate-identity-regexp $CertIdentityRegexp `
        --certificate-oidc-issuer $CertOidcIssuer | Out-Null
      if ($LASTEXITCODE -ne 0) {
        Die "cosign signature verification FAILED — refusing to install"
      }
      Write-Ok "cosign signature verified (keyless / Sigstore)"
    } elseif (Get-Command cosign -ErrorAction SilentlyContinue) {
      Write-Warn "cosign found but MOLTBUNKER_SKIP_COSIGN is set — skipping signature verification."
      Write-Warn "SHA-256 integrity is verified, but signature provenance is NOT."
    } else {
      Write-Warn "cosign not found on PATH — skipping signature verification."
      Write-Warn "SHA-256 integrity is verified, but signature provenance is not."
    }

    # ── Step 3: extract + install ──
    Write-Info "Extracting archive"
    $extract = Join-Path $work "extract"
    Expand-Archive -Path $archivePath -DestinationPath $extract -Force

    if (-not (Test-Path $InstallDir)) {
      New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # Install whatever .exe binaries the archive actually contains, rather than
    # a hardcoded list — mirrors install.sh's loop. Today the Windows archive
    # ships moltbunker.exe + moltbunker-daemon.exe only (api is linux/darwin,
    # exec-agent is linux), but this stays correct if that set ever changes.
    $installed = 0
    foreach ($srcBin in (Get-ChildItem -Path $extract -Filter "*.exe" -File)) {
      Write-Info "Installing $($srcBin.Name) -> $InstallDir\$($srcBin.Name)"
      Copy-Item -Path $srcBin.FullName -Destination (Join-Path $InstallDir $srcBin.Name) -Force
      $installed++
    }
    if ($installed -eq 0) { Die "no moltbunker binaries found in the archive" }
    Write-Ok "installed $installed binaries to $InstallDir"

    # Add InstallDir to the user PATH if not already present.
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$InstallDir*") {
      [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
      Write-Info "Added $InstallDir to your user PATH (restart your shell to pick it up)"
    }
  }
  finally {
    if (Test-Path $work) { Remove-Item -Path $work -Recurse -Force -ErrorAction SilentlyContinue }
  }
}

if (-not (Try-Winget)) {
  DownloadAndVerify
}

Write-Host ""
Write-Host "moltbunker installed."
Write-Host "  moltbunker --help            # CLI"
Write-Host "  moltbunker-daemon --help     # node daemon"
Write-Host ""
Write-Host "To re-verify a release, see:"
Write-Host "  https://github.com/moltbunker/moltbunker/blob/main/docs/release-integrity.md"
