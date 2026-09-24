# ADR-0002: Honest local GHA parity

## Status

Accepted

## Context

Users expect GitHub Actions fidelity. Cloud-only features (signed OIDC, environment protection rules, hosted Windows/mac VMs) cannot be cloned honestly.

## Decision

- Implement local-enforceable parity (fail-fast, timeouts, concurrency cancel groups, advisory permissions).
- Provide optional mocks (`--oidc-mock`, `--env-secret-file`, `--env-var-file`) with explicit documentation.
- Never claim 100% GitHub Actions compatibility; keep [PARITY.md](../../PARITY.md) current.

## Consequences

Workflows that require real GitHub trust boundaries must use alternate local credentials or run on GitHub.
