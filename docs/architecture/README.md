# Architecture

gotcontext-actions is a modular monolith derived from [nektos/act](https://github.com/nektos/act).

## Dependency direction

```
cmd (CLI) → internal/runner → adapters (container, artifacts, oidc)
                           → domain (model, exprparser, common)
```

- `cmd` is the thin Cobra surface and wiring.
- `internal/*` is not importable by external modules.
- Adapters never import `runner`.

## Binaries

| Binary | Path |
|--------|------|
| `gotcontext-actions` | `cmd/gotcontext-actions` |
| `act` (compat) | `cmd/act` |
| root `main.go` | same wiring for Makefile/legacy builds |

## Fitness

Import bans are enforced via `.golangci.yml` depguard (see ADR-0001).
