# Security Policy

## Reporting a Vulnerability

Please report security vulnerabilities by email to **security@moltbunker.com**.

You can expect:

- An acknowledgement within **48 hours**.
- An initial assessment within **7 days**.
- A coordinated disclosure timeline agreed with you before any public details are shared.

Please do **not** open a public GitHub issue, social media post, or other public channel for security issues. Use the email above.

## Scope

This policy covers the Moltbunker daemon, HTTP API, CLI, P2P protocol, container runtime integration, and the Solidity smart contracts under `contracts/`. Public marketing pages and third-party services we link to are out of scope.

In-scope examples:

- Authentication or authorization bypass on the HTTP API.
- P2P message validation or transport security issues.
- Container escape from the runtime sandbox.
- Smart contract issues: reentrancy, access-control bypass, oracle manipulation, economic logic errors.
- Cryptographic weaknesses: weak primitives, nonce reuse, signature verification flaws.
- Secrets exposed in logs or API responses.

Out of scope:

- Denial-of-service against the public API (mitigated by upstream rate limiting / Cloudflare).
- Issues that require physical access to a provider host.
- Vulnerabilities in third-party dependencies that have no demonstrable impact on Moltbunker (please report those upstream).
- Marketing/copy errors on `moltbunker.com`.

## Supported Versions

Until 1.0.0, only the latest released version is supported with security fixes. Re-base on the latest tag before reporting issues against older versions.

## Safe Harbor

We will not pursue legal action against good-faith security research that follows this policy and does not:

- Access, modify, or exfiltrate data that is not yours.
- Cause service disruption beyond what is strictly necessary to demonstrate the vulnerability.
- Use social engineering, phishing, or physical attacks.

## Recognition

Researchers who report valid in-scope issues will be credited in the relevant CHANGELOG.md entry and on the security page, with their permission.
