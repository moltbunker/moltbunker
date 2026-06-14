# Release integrity

moltbunker releases are built and published by a GitHub Actions pipeline
([`.github/workflows/release.yml`](../.github/workflows/release.yml)) driven by
GoReleaser ([`.goreleaser.yml`](../.goreleaser.yml)). Every release carries the
supply-chain artifacts you need to verify that the binary you downloaded is
exactly the one we built — and that *we* built it.

Each release attaches:

| Artifact | What it is |
|---|---|
| `moltbunker_<version>_<os>_<arch>.tar.gz` / `.zip` | the release archive (binaries + README + LICENSE) |
| `checksums.txt` | SHA-256 of every archive |
| `checksums.txt.sigstore.json` | cosign **keyless** (Sigstore) signature bundle over `checksums.txt` |
| `<archive>.cdx.json` | CycloneDX SBOM (software bill of materials) for that archive, produced by syft |

The trust model is layered: the **SBOM** tells you what is inside; the
**SHA-256 checksums** pin every archive; the **cosign bundle** proves the
checksums file was produced by the moltbunker release workflow and nobody else.

> The installers do steps 1 and 2 automatically. `scripts/install.sh` (Linux /
> macOS) and `scripts/install.ps1` (Windows) verify the SHA-256 before moving
> any binary into place, and additionally run `cosign verify-blob` when `cosign`
> is on your PATH. The SHA-256 check is mandatory; the cosign check is best-effort
> and only skipped (with a warning) if cosign is not installed.

---

## 1. SHA-256 checksum verification

Download the archive and `checksums.txt` for your platform, then:

```bash
# Linux (GNU coreutils)
sha256sum --ignore-missing -c checksums.txt

# macOS (BSD)
shasum -a 256 --ignore-missing -c checksums.txt
```

`--ignore-missing` lets you verify only the archive(s) you actually downloaded
without failing on the others listed in `checksums.txt`. A passing line reads
`moltbunker_<version>_<os>_<arch>.tar.gz: OK`.

This defends against accidental corruption and CDN/cache poisoning. It does
**not**, on its own, prove who produced the checksums file — that is what
cosign adds.

---

## 2. cosign keyless signature verification

moltbunker signs `checksums.txt` with [cosign](https://docs.sigstore.dev/) using
**keyless** (Sigstore OIDC) signing. There is **no private key** anywhere: the
GitHub Actions runner mints a short-lived OIDC token, exchanges it for an
ephemeral Fulcio certificate bound to the release-workflow identity, signs, and
records the event in the [Rekor](https://docs.sigstore.dev/logging/overview/)
transparency log. The ephemeral key is discarded; only the public certificate
and Rekor inclusion proof survive, inside `checksums.txt.sigstore.json`.

Requires **cosign v2.0 or newer** (the `--bundle` flag and `.sigstore.json`
bundle format are v2). Install:

```bash
# Linux / macOS (Homebrew)
brew install cosign
# or download from https://github.com/sigstore/cosign/releases
```

Verify:

```bash
cosign verify-blob checksums.txt \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp '^https://github.com/moltbunker/moltbunker/.github/workflows/release.yml@refs/tags/' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'
```

The two `--certificate-*` flags are the trust anchor. They assert that the
signing certificate was issued to the moltbunker release workflow
(`release.yml`, on a tag ref) and that the identity was authenticated by
GitHub's OIDC issuer. A successful run prints `Verified OK`.

If verification succeeds, you have cryptographic proof that `checksums.txt` —
and therefore every archive it pins — was produced by the moltbunker release
pipeline and recorded in a public transparency log.

---

## 3. Inspect the SBOM

Each archive has a companion CycloneDX JSON SBOM
(`<archive>.cdx.json`). To list components:

```bash
# Pretty-print component names + versions with jq
jq -r '.components[] | "\(.name) \(.version)"' moltbunker_<version>_linux_x86_64.tar.gz.cdx.json

# Or scan/diff it with any CycloneDX-aware tool, e.g. grype:
grype sbom:moltbunker_<version>_linux_x86_64.tar.gz.cdx.json
```

CycloneDX is the format mandated by most enterprise / government supply-chain
policies (CISA, US EO 14028). It includes component hashes and `purl`
identifiers for every Go dependency compiled into the binary.

---

## 4. One-shot re-verification

The repo ships a standalone helper that does steps 1 + 2 for you against an
already-downloaded archive:

```bash
# from the directory containing the downloaded archive(s)
./scripts/verify-release.sh v1.2.3
```

It downloads only `checksums.txt` (and the cosign bundle) for the given tag,
verifies the SHA-256 of every matching local archive, then runs
`cosign verify-blob` if cosign is installed. It exits non-zero on any failure.

---

## Why keyless?

A long-lived GPG/minisign key would require the maintainers to secure, rotate,
and revoke a private key forever — and a leaked key silently forges every future
release. Keyless signing eliminates the private key entirely: the signing
identity *is* the GitHub Actions workflow, which is public and version-controlled,
and every signature is independently logged in Rekor. To trust a release you only
need to know the org + workflow path — both visible in this repository.

---

## Known release-toolchain debt (tracked follow-ups)

These do not affect the integrity guarantees above, but they gate an *actual*
`goreleaser release` run and the long-term maintainability of the pipeline.
They are deliberately deferred (the integrity/signing work shipped without them)
and tracked here so the deprecation horizon is visible.

1. **GoReleaser pinned to `~> v2.4`.** Both `release.yml` and the `release-lint`
   CI job pin the same `~> v2.4` so the config validates today, but this freezes
   the toolchain on an aging v2 minor. Several stanzas in `.goreleaser.yml` are
   already deprecated and slated for hard removal upstream (deprecated in
   v2.16+): `archives.format` → `archives.formats`, `brews` →
   `homebrew_casks`, `dockers` → `dockers_v2`. When upstream hard-removes them a
   future bump breaks the release until they are migrated, and because the lint
   job is pinned to the same `~> v2.4` it will stay green and will NOT surface
   the break in advance.
   - **TODO:** migrate the three stanzas, then unpin / bump the GoReleaser
     version. Optionally add a *non-blocking* CI step that runs
     `goreleaser check` against `latest` so the deprecation horizon is at least
     visible before it becomes a hard break.

2. **Windows archive cannot currently be produced.** The daemon and CLI depend
   on Linux-only syscalls and do not compile for `GOOS=windows`, and the mixed
   per-platform binary counts trip GoReleaser's "different count of binaries for
   each platform" archive error. `scripts/install.ps1` therefore documents a
   path that no real release can yet feed (both pre-existing; out of REL-01
   scope).
   - **TODO:** either drop `windows` from the build matrix in `.goreleaser.yml`,
     or split archives per build-id so a real `goreleaser release` can run. Once
     Windows archives actually ship, `install.ps1` should enumerate `*.exe` in
     the extract dir rather than hardcoding the binary list (see the comment at
     the install loop in `scripts/install.ps1`).
