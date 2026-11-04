# Contributing to promoloupe

Thanks for your interest in promoloupe. This guide explains how to propose
changes and what we expect from a contribution.

## Before you start

- For bugs and small fixes, open a pull request directly or file an issue
  using the bug report template.
- For new features, API changes or anything that affects the CRDs, open an
  issue first so the design can be discussed before code is written.
  Significant decisions are recorded as ADRs in [`docs/adr`](docs/adr).
- Security issues follow a separate, private process described in
  [SECURITY.md](SECURITY.md).

## Repository layout

| Path | Contents |
| --- | --- |
| `api/` | CRD types (`promoloupe.com/v1alpha1`) |
| `cmd/` | Entry points for the controller and the CLI |
| `internal/` | Controller logic, gates and SCM adapters |
| `config/` | Generated manifests (CRDs, RBAC) |
| `reviewer/` | Python reviewer agent |
| `charts/` | Helm chart |
| `docs/` | Documentation and ADRs |

## Development setup

You will need:

- Go (version pinned in `go.mod`)
- Python 3.12 or later and [uv](https://docs.astral.sh/uv/)
- Docker or Podman, and [kind](https://kind.sigs.k8s.io/) for end-to-end tests
- [pre-commit](https://pre-commit.com/)

Setup instructions and `make` targets will be documented here as the
corresponding parts of the repository land.

## Making a change

1. Fork the repository and create a branch from `main`.
2. Make your change, with tests. Code without tests will not be merged.
3. Run the linters and tests locally (`make lint test` for Go,
   `uv run ruff check`, `uv run mypy` and `uv run pytest` in `reviewer/`).
4. Update documentation when behaviour changes.
5. Open a pull request and fill in the template.

### Commit messages

We use [Conventional Commits](https://www.conventionalcommits.org/). The
changelog and release versions are generated from them.

```
feat(controller): add soak time gate
fix(reviewer): handle empty Prometheus responses
docs: explain promotable paths
```

Common types: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `build`,
`ci`, `chore`. Mark breaking changes with `!` after the type
(`feat(api)!: ...`) and describe them in the commit body.

### Developer Certificate of Origin

All commits must be signed off to certify that you wrote the change or have
the right to submit it under the project's license, as described in the
[Developer Certificate of Origin](https://developercertificate.org/):

```
git commit -s -m "feat(controller): add soak time gate"
```

### AI-assisted contributions

You may use AI coding assistants to help with a contribution. The repository
ships instructions for them in [`AGENTS.md`](AGENTS.md), and an assisted
change is held to the same standards as any other.

- **You are the author.** You must have read, understood and tested every
  line you submit, and be able to explain it during review. "The assistant
  wrote it" is not an answer to a review comment.
- **Sign-off is yours alone.** Only a person can certify the DCO. Assistants
  must not add `Signed-off-by` lines. You add your own with `git commit -s`
  once you have reviewed the change, and by doing so you confirm you have the
  right to submit it under the project's license.
- **Verify before you report.** Do not file issues or security reports you
  have not reproduced yourself. Unverified, AI-generated reports will be
  closed.

### Code standards

- **Go:** `gofmt`, `golangci-lint` and controller-runtime conventions.
- **Python:** `ruff` for linting and formatting, `mypy --strict`, Pydantic
  models for every external input and output.
- **APIs:** CRD changes follow Kubernetes API conventions. Fields are never
  removed or repurposed within a version; deprecations are announced at least
  one minor release in advance.

## Pull request review

A maintainer will review your pull request. CI must pass before merge. We may
ask for changes; please push them as new commits so the review is easy to
follow. Pull requests are squash-merged, so the title must itself be a valid
Conventional Commit.

## License

By contributing, you agree that your contributions are licensed under the
[Apache License 2.0](LICENSE).
