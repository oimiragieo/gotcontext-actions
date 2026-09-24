# ADR-0001: Modular monolith under internal/

## Status

Accepted

## Context

nektos/act used a flat `pkg/` tree. gotcontext-actions needs clearer boundaries for enterprise maintenance without splitting into multiple Go modules.

## Decision

1. Single module: `github.com/oimiragieo/gotcontext-actions`
2. Move former `pkg/*` packages to `internal/*`
3. Keep Cobra in `cmd/` with dual binaries `gotcontext-actions` and `act`
4. Reject `go.work` multi-module split in wave 1

## Consequences

- External consumers cannot import runner internals (intentional).
- Import rewrites are mechanical; behavior changes land in separate commits.
- Upstream cherry-picks from nektos/act require path mapping `pkg/` → `internal/`.
