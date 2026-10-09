# Agent handoff — gotcontext-actions

Written **2026-10-09** at session closeout. If `git log -1` ≠ the SHA below after pull, re-verify before trusting version pins.

## Objective

Local GitHub Actions runner derived from nektos/act (`github.com/oimiragieo/gotcontext-actions`). No new product features in this handoff — security/deps/CI hygiene only until backlog P0 clears.

## Repository and branch

- Path: `C:\dev\projects\gotcontext-actions`
- Remote: `https://github.com/oimiragieo/gotcontext-actions.git`
- Branch: `main` (keep ahead=0 with origin)

## Relevant commit

Closeout documents were written against tip **`284a2db`** (`docs: fix What to do next Dependabot links (#17)`). After this closeout commit lands, use the new HEAD SHA from `git log -1`.

## Completed (verified)

| Claim | How verified |
| --- | --- |
| Race CI `CGO_ENABLED=1` on race step | `checks.yml`; PR #10 race step success run `37774733432` |
| `go-git/v5 v5.19.2` | `go.mod`; PR #11 |
| `x/crypto v0.57.0`, `go 1.26.8` | `go.mod`; PR #13 |
| `moby/go-archive v0.3.0` + `compression.None` | `go.mod` + `docker_build.go`; PR #15 |
| `.go-arch-lint.yml` gone; depguard primary | file absent; architecture README |
| Short race gate green locally | `CGO_ENABLED=1 go test -race … common/model` → ok (closeout) |
| Session branches harvested | local deleted; remotes pruned |

## Commands run and results (closeout)

```
CGO_ENABLED=1 go test -race -count=1 -timeout 90s ./internal/common/ ./internal/model/
→ ok common; ok model; RACE_EXIT=0

go run golang.org/x/vuln/cmd/govulncheck@latest ./...
→ affected by 10 stdlib vulns; Fixed in go1.26.9 (+ x/net v0.60.0)
→ VULN_EXIT=1  ← do not claim clean
```

## Unresolved problems (do not “fix” as product bugs)

1. **govulncheck red for stdlib** — needs Go **1.26.9** + **x/net v0.60.0** (backlog B1/B2).
2. **lint / snapshot** remote jobs still fail on known debt (B3/B4).
3. Windows socket path slash test flake (B11).
4. Live git clone test can timeout under 120s (B12).
5. Open Dependabot: **#18** (gomod mega), **#5** (Actions mega), **#7** (otel) — never merge #18 with single security bumps.

## Decisions and constraints

- Merge gate = short race + documented govulncheck; **not** full `go test ./...`.
- Prefer tip `go get` over rebasing stale Dependabot after layout moves.
- Architecture fitness = golangci **depguard** only.
- Skill routing: load `skill-library-map` first — do not dump skill lists into CLAUDE.md.
- No force-push; do not touch other agents’ in-flight worktrees.

## Next steps (ordered)

1. Read `backlog.md` + `README.md` “What a new session should trust”.
2. Land **B1** (`go get go@1.26.9`) then **B2** (`golang.org/x/net@v0.60.0`), tidy, race, govulncheck → expect 0 affecting (or document leftovers).
3. Optionally #7 otel alone; leave #18/#5 until later.
4. Lint debt (B3) only if branch protection blocks merges.
5. Product gaps B8–B10 only after explicit spec (feature-batch + wayfinder).

## Hard rules

- Credit nektos/act; no fake 100% GHA parity (`PARITY.md`).
- Do not invent skills; compose via skill-library-map / compose-build-pipeline.
