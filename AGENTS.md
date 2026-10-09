# AGENTS.md

Start here for a fresh session:

1. `git pull` and `git log -1` — compare to [HANDOFF.md](HANDOFF.md).
2. Read [backlog.md](backlog.md) (honest open items) and README “What a new session should trust”.
3. Skill routing: load **`skill-library-map`** at `C:\Users\oimir\.cursor\skills\skill-library-map\SKILL.md` — then only the leaves it selects. Do not invent skills; compose via that map / `compose-build-pipeline`.
4. Merge gate: `CGO_ENABLED=1 go test -race -count=1 -timeout 120s ./internal/common/ ./internal/model/`.
5. After dependency work: re-run `govulncheck ./...` and require a positive “0 affecting” (or document leftovers in backlog).

Derived from nektos/act; see NOTICE / PARITY.md. Product gaps vs Rehearse are backlog B8–B10 — need specs before coding.
