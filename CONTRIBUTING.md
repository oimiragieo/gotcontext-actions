# Contributing to Act

Help wanted! We'd love your contributions to Act. Please review the following guidelines before contributing. Also, feel free to propose changes to these guidelines by updating this file and submitting a pull request.

- [I have a question...](#questions)
- [I found a bug...](#bugs)
- [I have a feature request...](#features)
- [I have a contribution to share...](#process)

## <a id="questions"></a> Have a Question?

Please don't open a GitHub issue for questions about how to use `act`, as the goal is to use issues for managing bugs and feature requests. Issues that are related to general support will be closed and redirected to [discussions](https://github.com/nektos/act/discussions).

For all support related questions, please [open a discussion post](https://github.com/nektos/act/discussions/new/choose).

## <a id="bugs"></a> Found a Bug?

If you've identified a bug in `act`, please [submit an issue](#issue) to our GitHub repo: [nektos/act](https://github.com/nektos/act/issues/new). Please also feel free to submit a [Pull Request](#pr) with a fix for the bug!

## <a id="features"></a> Have a Feature Request?

All feature requests should start with [submitting an issue](#issue) documenting the user story and acceptance criteria. Again, feel free to submit a [Pull Request](#pr) with a proposed implementation of the feature.

## <a id="process"></a> Ready to Contribute

### <a id="issue"></a> Create an issue

Before submitting a new issue, please search the issues to make sure there isn't a similar issue doesn't already exist.

Assuming no existing issues exist, please ensure you include required information when submitting the issue to ensure we can quickly reproduce your issue.

We may have additional questions and will communicate through the GitHub issue, so please respond back to our questions to help reproduce and resolve the issue as quickly as possible.

New issues can be created with in our [GitHub repo](https://github.com/nektos/act/issues/new).

### <a id="pr"></a>Pull Requests

Pull requests should target the `master` branch. Please also reference the issue from the description of the pull request using [special keyword syntax](https://help.github.com/articles/closing-issues-via-commit-messages/) to auto close the issue when the PR is merged. For example, include the phrase `fixes #14` in the PR description to have issue #14 auto close. Please send documentation updates for the [act user guide](https://nektosact.com) to [nektos/act-docs](https://github.com/nektos/act-docs).

### <a id="style"></a> Styleguide

When submitting code, please make every effort to follow existing conventions and style in order to keep the code as readable as possible. Here are a few points to keep in mind:

- Please run `go fmt ./...` before committing to ensure code aligns with go standards.
- We use [`golangci-lint`](https://golangci-lint.run/) for linting Go code, run `golangci-lint run --fix` before submitting PR. Editors such as Visual Studio Code or JetBrains IntelliJ; with Go support plugin will offer `golangci-lint` automatically.
- There are additional linters and formatters for files such as Markdown documents or YAML/JSON:
  - Please refer to the [Makefile](Makefile) or [`lint` job in our workflow](.github/workflows/checks.yml) to see how to those linters/formatters work.
  - You can lint codebase by running `go run main.go -j lint --env RUN_LOCAL=true` or `act -j lint --env RUN_LOCAL=true`
  - In `Makefile`, there are tools that require `npx` which is shipped with `nodejs`.
  - Our `Makefile` exports `GITHUB_TOKEN` from `~/.config/github/token`, you have been warned.
  - You can run `make pr` to cleanup dependencies, format/lint code and run tests.
- All dependencies must be defined in the `go.mod` file.
  - Advanced IDEs and code editors (like VSCode) will take care of that, but to be sure, run `go mod tidy` to validate dependencies.
- For details on the approved style, check out [Effective Go](https://golang.org/doc/effective_go.html).
- Before running tests, please be aware that they are multi-architecture so for them to not fail, you need to run `docker run --privileged --rm tonistiigi/binfmt --install amd64,arm64` before ([more info available in #765](https://github.com/nektos/act/issues/765)).

Also, consider act's design principles:

- **Local fidelity** - Prefer behavior that matches GitHub Actions runners where it is practical to emulate locally (expressions, steps, containers, artifacts/cache).
- **Composable executors** - Prefer small `Executor` chains over large imperative control flow.
- **Honest gaps** - Document features that cannot be fully cloned locally (OIDC signing, deployment protection rules, hosted Windows/macOS VMs) instead of pretending full parity.
- **Docker-first** - Default execution uses containers; host/`-self-hosted` is an escape hatch, not a full runner replacement.
- **Upstream-friendly** - Prefer small, tested changes that can merge to `nektos/act`.

### License

By contributing your code, you agree to license your contribution under the terms of the [MIT License](LICENSE).

All files are released with the MIT license.
