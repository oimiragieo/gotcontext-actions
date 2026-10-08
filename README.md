# gotcontext-actions

Local GitHub Actions runner for development and CI dogfooding.

**Derived from [nektos/act](https://github.com/nektos/act)** (MIT). See [NOTICE](NOTICE) and [LICENSE](LICENSE). Upstream project: [nektos/act](https://github.com/nektos/act). This is not an official nektos product.

Compatibility with GitHub Actions is intentionally incomplete; see [PARITY.md](PARITY.md). Layout notes are in [docs/architecture/README.md](docs/architecture/README.md).

## What a new session should trust

Local merge gate: `go test -race -count=1 -timeout 120s ./internal/common/ ./internal/model/` (with `CGO_ENABLED=1` for race). Full `go test ./...` (Docker plus live clones) is not the merge gate.

| Result | Evidence |
| --- | --- |
| Worked, after a seen failure | `06f6b2c` ThenError, scalar `concurrency`, job-result lock, action-cache lock, Node 22 defaults. `TestGetGitHubContext` failed first because the git remote is `oimiragieo/gotcontext-actions`. |
| Worked | `3f5bcb2` `queue: single` waits; `queue: max` overlaps; `cancel-in-progress` still cancels. `go test -run Concurrency ./internal/common` |
| Worked, after a seen failure | `00f7451` moves Cobra to `internal/cli` and expressions to `internal/expr`. `hashFiles` returned empty hashes on Windows until patterns were slash-normalized. `go test -race ./internal/common/ ./internal/model/` failed on unsynchronized counters in `TestNewParallelExecutor`, then passed after the test mutex. |
| Worked | Hermetic step summary at `d297ae6`. OIDC mock, `--strict-platforms`, `--env-secret-file`, `--env-var-file`, `--step-summary-file`. |
| Worked | Branch `deps/x-crypto-0.52-x-net-0.55`: `go get` crypto `v0.52.0` then net `v0.55.0` + `go mod tidy`. Race short tests passed. Stale Dependabot #3/#4 were 6 commits behind `main` (pre-`internal/` layout); closed after this tip landing. |
| Worked | Deleted `.go-arch-lint.yml`. Architecture fitness remains golangci **depguard** only (adapters must not import `runner`/`cli`). |
| Worked | Race CI: workflow keeps global `CGO_ENABLED=0`, but the “Race unit tests” step sets `CGO_ENABLED=1` so `-race` works on Linux (`3cb9ce2` / #10). |
| Worked | `go-git/v5` → `v5.19.2` (path traversal / symlink advisories). Race on `./internal/common/` + `./internal/model/` passed. Live `TestGitCloneExecutor` can timeout on network clones — not the merge gate. |
| Worked | `golang.org/x/crypto` → `v0.57.0`; module `go` → `1.26.8`. Race short tests passed. `govulncheck` crypto SSH findings cleared. |
| Worked, after a seen failure | `moby/go-archive` → `v0.3.0`. `archive.Uncompressed` removed upstream — use `compression.None` in `docker_build.go`. `govulncheck ./...` reports 0 affecting vulns. |
| Do not treat as proven | Remote CI fully green (lint/snapshot debt may remain). Full `go test ./...`. Dependabot #14 (gomod group), #5 (Actions group), #7 (otel) still open. |
| Do not "fix" as product bugs | `TestGetSocketAndHostNoHostNoSocketDefaultLocation` can fail on Windows path slash style (`C:\\` vs `C:/`); unrelated to go-archive. |
| Do not "fix" as product bugs | Empty tensor-grep blast-radius means gopls was missing, not that a symbol has zero callers. Vendored `internal/runner/testdata/actions/**/node_modules` is fixture payload, not the Go module. Remote lint/snapshot failures on Dependabot PRs were pre-existing (gocyclo, Windows GoReleaser/Chocolatey), not caused by crypto/net bumps. |

Binaries: `cmd/gotcontext-actions` and `cmd/act` call `internal/cli`. Package `internal/model` was left in place on purpose.

## What to do next

Review Dependabot [#7](https://github.com/oimiragieo/gotcontext-actions/pull/7) (otel/sdk) alone if needed; do not merge gomod group [#14](https://github.com/oimiragieo/gotcontext-actions/pull/14) or Actions mega [#5](https://github.com/oimiragieo/gotcontext-actions/pull/5) with single security bumps. Snapshot/Chocolatey and remaining lint debt are separate. Host-mode speed, service health checks, and watch/pre-push versus Rehearse are not started.

## Overview

> "Think globally, run locally"

Run your [GitHub Actions](https://developer.github.com/actions/) locally! Why would you want to do this? Two reasons:

- **Fast Feedback** - Rather than having to commit/push every time you want to test out the changes you are making to your `.github/workflows/` files (or for any changes to embedded GitHub actions), you can use `act` to run the actions locally. The [environment variables](https://help.github.com/en/actions/configuring-and-managing-workflows/using-environment-variables#default-environment-variables) and [filesystem](https://help.github.com/en/actions/reference/virtual-environments-for-github-hosted-runners#filesystems-on-github-hosted-runners) are all configured to match what GitHub provides.
- **Local Task Runner** - I love [make](<https://en.wikipedia.org/wiki/Make_(software)>). However, I also hate repeating myself. With `act`, you can use the GitHub Actions defined in your `.github/workflows/` to replace your `Makefile`!

> [!TIP]
> **Now Manage and Run Act Directly From VS Code!**<br/>
> Check out the [GitHub Local Actions](https://sanjulaganepola.github.io/github-local-actions-docs/) Visual Studio Code extension which allows you to leverage the power of `act` to run and test workflows locally without leaving your editor.

# How Does It Work?

When you run `act` it reads in your GitHub Actions from `.github/workflows/` and determines the set of actions that need to be run. It uses the Docker API to either pull or build the necessary images, as defined in your workflow files and finally determines the execution path based on the dependencies that were defined. Once it has the execution path, it then uses the Docker API to run containers for each action based on the images prepared earlier. The [environment variables](https://help.github.com/en/actions/configuring-and-managing-workflows/using-environment-variables#default-environment-variables) and [filesystem](https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners#file-systems) are all configured to match what GitHub provides.

Let's see it in action with a [sample repo](https://github.com/cplee/github-actions-demo)!

![Demo](https://raw.githubusercontent.com/wiki/nektos/act/quickstart/act-quickstart-2.gif)

# Act User Guide

Please look at the [act user guide](https://nektosact.com) for more documentation.

Compatibility with GitHub Actions is intentionally incomplete; see [PARITY.md](PARITY.md) for what works, what is mocked, and what cannot be cloned locally.

# Support

Need help? Ask in [discussions](https://github.com/oimiragieo/gotcontext-actions/discussions)!

# Contributing

Want to contribute to act? Awesome! Check out the [contributing guidelines](CONTRIBUTING.md) to get involved.

## Manually building from source

- Install Go tools 1.25+ - (<https://golang.org/doc/install>)
- Clone this repo `git clone https://github.com/oimiragieo/gotcontext-actions.git`
- Run unit tests with `make test`
- Build and install: `make install`
