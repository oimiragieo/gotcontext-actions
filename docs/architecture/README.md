# Architecture

gotcontext-actions is a modular monolith derived from [nektos/act](https://github.com/nektos/act).

## Dependency direction

```
cmd mains → internal/cli → internal/runner → adapters (container, artifacts, oidc)
internal/expr is the expression language used by runner
adapters and runner use domain packages (model, common)
```

- `cmd/gotcontext-actions` and `cmd/act` are thin mains. Cobra lives in `internal/cli`.
- `internal/expr` evaluates `${{ }}` expressions.
- Vendored `node_modules` under `internal/runner/testdata/actions` are fixture payloads for action tests, not product dependencies. Dependabot alerts on those files are not the runtime module set in `go.mod`.
- `internal/*` is not importable by external modules.
- Adapters never import `runner`.

## Binaries

| Binary | Path |
|--------|------|
| `gotcontext-actions` | `cmd/gotcontext-actions` |
| `act` (compat) | `cmd/act` |
| root `main.go` | same wiring for Makefile/legacy builds |

## Fitness

Import bans are enforced via `.golangci.yml` depguard (see ADR-0001). That is the only architecture fitness gate in CI. A standalone `.go-arch-lint.yml` was removed because it was incomplete relative to the real package graph and was never a merge requirement.
