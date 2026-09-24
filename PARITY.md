# GitHub Actions parity notes for act

Act aims for local fidelity with GitHub Actions, but **is not 100% compatible**.

Official upstream gap list (also see https://nektosact.com/not_supported.html):

## Locally implemented / improved in this tree

| Feature | Status |
| --- | --- |
| Matrix `fail-fast` | Enforced (cancels sibling matrix cells) |
| Job `timeout-minutes` | Enforced |
| Job `continue-on-error` | Enforced |
| Local action `pre:` | Runs (parity with remote actions) |
| `GITHUB_TOKEN` env | Injected when a token is configured |
| Workflow `run-name` | Interpolated and logged |
| `concurrency` (single act process) | Cancel-in-progress within one invocation |
| `permissions` | Parsed and logged (advisory only; PAT scopes ≠ GITHUB_TOKEN) |
| Step summaries | Collected and printed at job end |
| Artifact server | Starts by default under cache home (`--no-artifact-server` to disable) |
| `--strict-platforms` | Fail instead of skipping unmapped `runs-on` |
| `--oidc-mock` | Local JWT endpoint for `ACTIONS_ID_TOKEN_REQUEST_URL` |
| `--env-secret-file name=path` | Optional deployment-environment secrets |

## Cannot fully clone (GitHub cloud-only)

| Feature | Why |
| --- | --- |
| Real OIDC / cloud role assumption | Only GitHub can mint tokens trusted by AWS/GCP/Azure |
| Environment protection rules / reviewers | Requires GitHub API and human approvals |
| True `GITHUB_TOKEN` permission scopes | Hosted runners mint a job-scoped App token |
| Windows / macOS hosted VM fidelity | act uses Linux containers or host execution |
| Cross-run artifact download by run id | Local server is scoped to the current act process |

## Guidance

- Prefer documenting gaps over claiming full parity.
- Use `--oidc-mock` only for local smoke tests of OIDC *clients*; use long-lived cloud credentials for real deploys under act.
- Prefer `catthehacker/ubuntu:act-*` images for broader action compatibility than micro Node slim tags.
