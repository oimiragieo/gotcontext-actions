# Backlog — gotcontext-actions

Honest state + receipts. Newest closeout: **2026-10-09**. Tip verified at write time: see `.orchestrator/state.json` / `git log -1`.

## P0 — security / CI signal (next session)

| ID | Item | State | Receipt |
| --- | --- | --- | --- |
| B1 | Bump Go toolchain past `1.26.8` → **`1.26.9`** (stdlib `net/http`, `crypto/tls`, `net/textproto` findings) | **OPEN** | `govulncheck ./...` on `284a2db` / tip at closeout: *Your code is affected by 10 vulnerabilities from the Go standard library*; Fixed in `go1.26.9`. Do **not** treat earlier “0 affecting” claim as still true. |
| B2 | Bump `golang.org/x/net` → **`v0.60.0`** (co-required with several stdlib/http findings) | **OPEN** | Same `govulncheck` run lists Fixed in `golang.org/x/net@v0.60.0`. Land on tip after B1; do not fold into Dependabot mega-PR #18. |
| B3 | Remote `checks` lint debt (gocyclo `execAsDocker` / `newJobExecutor`, gosec G114 OIDC Serve, depguard `artifacts`→`runner` in `server_test.go`, unused `loadJobResult`, nolintlint) | **OPEN** | Observed on Dependabot/session PRs; host tests often green while lint red. Not caused by crypto/go-git bumps. |
| B4 | Snapshot/Chocolatey: missing `dist/act_windows_*` → choco `cp` fails | **OPEN** | Known on PR #8/#10 era runs; GoReleaser can succeed while Windows artifact path missing. |

## P1 — Dependabot leftovers (one PR at a time)

| ID | Item | State | Receipt |
| --- | --- | --- | --- |
| B5 | [#7](https://github.com/oimiragieo/gotcontext-actions/pull/7) `otel/sdk` 1.43→1.45 | **OPEN** | Leave alone until B1/B2 if it churns `go.mod`. Prefer tip `go get` if stale. |
| B6 | [#18](https://github.com/oimiragieo/gotcontext-actions/pull/18) gomod group (18 updates) | **OPEN — do not merge with single security bumps** | Replaces older #1/#14 mega-bumps. High blast radius (docker/moby/etc.). |
| B7 | [#5](https://github.com/oimiragieo/gotcontext-actions/pull/5) GitHub Actions group (11 updates) | **OPEN** | Touches merge gate (setup-go, megalinter majors). Separate from Go module security. |

## P2 — product gaps vs Rehearse (not started)

| ID | Item | State | Receipt |
| --- | --- | --- | --- |
| B8 | Host-mode speed path (less Docker for local feedback) | **NOT STARTED** | README “What to do next”; competitor gap only. |
| B9 | Service-container health checks | **NOT STARTED** | Same. |
| B10 | Watch / pre-push local runner UX | **NOT STARTED** | Same. |

## P3 — test / tooling hygiene

| ID | Item | State | Receipt |
| --- | --- | --- | --- |
| B11 | `TestGetSocketAndHostNoHostNoSocketDefaultLocation` Windows slash flake (`C:\\` vs `C:/`) | **OPEN** | Failed during go-archive verification; unrelated to archive API. |
| B12 | `TestGitCloneExecutor` live-network timeout under 120s race budget | **OPEN** | Timed out on go-git bump; not merge gate. Consider skip-network or longer timeout / httptest. |
| B13 | Race CI step proven green after CGO fix | **DONE** | [PR #10](https://github.com/oimiragieo/gotcontext-actions/pull/10) run `37774733432` Race unit tests `conclusion: success`. |
| B14 | `govulncheck` cleared go-git / crypto / go-archive module CVEs at land time | **DONE (module)** / **STALE for stdlib** | Module findings cleared after #11/#13/#15; stdlib re-opened as B1. |
| B15 | Delete `.go-arch-lint.yml`; depguard-only fitness | **DONE** | Tip has no file; `docs/architecture/README.md` states depguard primary. |
| B16 | Fixture `node_modules` Dependabot noise | **OPEN (document only)** | Not the Go module; do not “fix” as product deps. |

## Ideas to improve next (research-backed, not scoped this session)

| Idea | Why | Note |
| --- | --- | --- |
| Add `golang/govulncheck-action` (SARIF → Code Scanning) | Continuous reachable-vuln signal; Jamie Tanna / official action pattern | Wire after B1 so baseline isn’t all red |
| Keep merge gate = short `-race` + `govulncheck`; never require full Docker `./...` | Full suite timeouts; false reds hide security merges | Already README policy — enforce in branch protection if enabled |
| Fix Windows path normalize in host socket tests | Stops false product-bug chases | Small, high leverage |
| One-at-a-time Dependabot; close stale PRs after tip `go get` | Learned from #3/#4 (6 behind `internal/` move) | Keep |
| Optional: host backend / watch (Rehearse) | Product differentiation | Needs design; use skill-library-map → brainstorming / writing-plans first |

## Closed this wave (receipts)

- #10 CGO race step · #11 go-git 5.19.2 · #13 crypto 0.57 + Go 1.26.8 · #15 go-archive 0.3.0 + `compression.None` · #16/#17 README queue · closed Dependabot #2/#3/#4/#6/#9 (and superseding #12).
