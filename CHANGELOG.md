# Changelog

All notable changes to this project are documented here.
This project adheres to [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- `SECURITY.md` describing the vulnerability disclosure process.
- `.github/PULL_REQUEST_TEMPLATE.md` with summary / linked tickets / change list / test plan / risk sections.
- `.githooks/commit-msg` enforcing the project's commit-subject format; activated via `core.hooksPath` on a per-clone basis.
- DBus session setup in the CI test job so keyring-dependent tests pass on Ubuntu runners.

### Changed

- Test code across 37 `_test.go` files now checks error returns from setup helpers (state, networking, snapshot, security, p2p, runtime, storage, payment, agent, tunnel, ingress, redundancy, molt, crawl, agent, daemon, identity, proxy, client, upgrade).
- Production code: cleanup paths (Delete / Stop / Close / Disconnect / network deadlines) now log on failure; init paths propagate errors to the caller.
- Crypto-randomness call sites (`crypto/rand.Read`) propagate errors. ID generators in `snapshot`, `storage/multipart`, `agent`, `crawl`, `proxy`, `cloning` now return `error`. The reverse-tunnel HMAC secret bootstrap panics rather than producing zero bytes if `rand.Read` fails.
- `Checkpointer.SetInterval` ticker is now stored via `atomic.Pointer[time.Ticker]` so the read in `checkpointLoop`'s select case does not race with the swap.
- Inline wallet auth (`VerifyInlineAuth`) now requires the `moltbunker-auth:` prefix and a well-formed body; messages with other prefixes are rejected before signature verification.
- Exec ownership check (`/v1/exec/challenge`) now rejects an empty `Owner` field on the deployment rather than treating it as a wildcard.
- Inbound P2P messages with an empty `Signature` are now rejected with a logged warning and a peer-score penalty when a key manager is configured; there is no signature-skipping fast path.

### Fixed

- `golangci-lint` clean across the repo (`errcheck`, `gosimple`, `staticcheck`, `ineffassign`, `unused` all at 0).
- Data race in `HealthChecker.Reset` vs `Start`'s deferred `close(doneCh)` (close moved inside the mutex).
- Data race between `Checkpointer.SetInterval` and `checkpointLoop` (atomic-pointer ticker swap, as above).
- `internal/networking/TestFallbackNetwork_ContainerIP` now skips on Linux where `linuxContainerNetwork` is in use.
- `internal/api` SARIF upload step in the Security Scan job has the `security-events: write` permission required by `codeql-action/upload-sarif`.

### Security

- `logging.EnableRedaction()` is now called at daemon startup. Structured-log attributes that look like API keys (`mb_live_*` / `mb_test_*`), wallet keystore JSON, session tokens (`wt_*`), private keys, or EIP-191 signatures are scrubbed by the redacting handler before reaching the underlying log handler.
- `NewAPIKeyManagerInMemory` no longer emits the plaintext development API key as a structured field; the plaintext is now printed once to stdout and never enters the log pipeline.

### Removed

- 17 unused symbols across production and test code (and their now-dead imports).
- Windows entries from the build matrix. The daemon depends on Linux-only subsystems (containerd, namespaces, hypervisor integration) and Unix syscalls (`syscall.Statfs`); Windows was never a supported target. The supported build matrix is now `{linux, darwin} × {amd64, arm64}`.

[Unreleased]: https://github.com/moltbunker/moltbunker/compare/HEAD~1...HEAD
