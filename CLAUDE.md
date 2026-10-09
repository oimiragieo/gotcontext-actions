# CLAUDE.md

Guidance for agents working in this repository.

## Skill routing

Before inventing process or loading many skills: read **`C:\Users\oimir\.cursor\skills\skill-library-map\SKILL.md`** (and its `manifest.json`). Load 1–3 leaves from that map. Do not paste the full skill list here — it drifts.

Session state for the next agent: [HANDOFF.md](HANDOFF.md), [backlog.md](backlog.md), [README.md](README.md) “What a new session should trust”, `.orchestrator/state.json`.

## Project Overview

**gotcontext-actions** — local GitHub Actions runner for development/CI dogfooding. Derived from [nektos/act](https://github.com/nektos/act) (MIT; see NOTICE). Module: `github.com/oimiragieo/gotcontext-actions`. Go **1.26+** (see `go.mod`). Compatibility is intentionally incomplete — [PARITY.md](PARITY.md).

## Common Commands

- `make build` — binary to `dist/local/act`
- `make test` / `make lint-go` / `make format` / `make tidy` / `make pr`
- **Merge gate (preferred):** `CGO_ENABLED=1 go test -race -count=1 -timeout 120s ./internal/common/ ./internal/model/`
- `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` — require a real “0 affecting” line; exit 0 alone is not enough
- Full `go test ./...` is heavy (Docker/live clones); not the merge gate

## Architecture

Layout: thin `cmd/*` → `internal/cli` → `internal/runner` → adapters (`container`, `artifacts`, `oidc`, …). Expressions: `internal/expr`. Domain: `internal/model`, `internal/common`.

Fitness: golangci **depguard** (adapters must not import `runner`/`cli`). No `go-arch-lint` gate.

Executor pattern: `internal/common` `Executor` = `func(ctx context.Context) error` with `.Then` / pipelines / parallel / `.If`.

## Linting

`.golangci.yml`: stdlib `errors`, `logrus` as `log`, testify (not gotest.tools), complexity ≤20, importas aliases.

## Testing

testify; fixtures under `internal/runner/testdata/` (vendored action `node_modules` are fixtures, not Go deps).

## Session lessons (2026-10)

- Stale Dependabot after layout moves → tip `go get`, then close the bot PR.
- Race needs per-step `CGO_ENABLED=1` when workflow sets global `0`.
- `moby/go-archive` v0.3+: use `compression.None`, not `archive.Uncompressed`.
- “govulncheck clean” expires when the Go patch line moves — re-run on tip before claiming.
