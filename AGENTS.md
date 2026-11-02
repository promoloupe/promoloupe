# promoloupe

Agentic risk reviewer for GitOps promotions. A Go Kubernetes operator
(`promoloupe-controller`) manages promotions between environments, and a
Python service (`promoloupe-reviewer`) uses an AI agent to review the risk of
each one. The agent recommends; it never promotes (ADR 0006).

## Layout

| Path | Contents |
| --- | --- |
| `cmd/`, `internal/` | Go controller and CLI |
| `api/v1alpha1/` | CRD types: the contract between controller and reviewer (ADR 0008) |
| `config/` | Kustomize manifests, partly generated |
| `reviewer/` | Python reviewer service; read `reviewer/AGENTS.md` before working there |
| `docs/adr/` | Architecture decision records |

## Commands

Run Go commands from the repository root:

- Build: `make build`
- Unit tests: `make test`
- Lint: `make lint` (golangci-lint; slow on first run)
- Regenerate code and manifests after changing API types or markers:
  `make manifests generate`
- End-to-end tests on kind: `make test-e2e` (needs Docker; several minutes)

Python commands are in `reviewer/AGENTS.md`.

## Generated files

Never edit these by hand; change the source and regenerate:

- `zz_generated.*.go` and `config/crd/bases/*.yaml`: from API types and
  kubebuilder markers, with `make generate` and `make manifests`.
- `config/rbac/role.yaml`: from `// +kubebuilder:rbac` markers, with
  `make manifests`.
- `go.sum`: with `go mod tidy`. `reviewer/uv.lock`: with `uv add` or `uv lock`.

## Rules

- API types follow Kubernetes API conventions. Never remove or repurpose a
  field in an existing version; add new optional fields instead.
- Every Go file starts with the license header in `hack/boilerplate.go.txt`.
- Do not change the Go toolchain in `go.mod`, the uv version in
  `reviewer/pyproject.toml` or tool versions in the Makefile unless asked.
- Never disable, skip or weaken a linter, test or CI check to make it pass.
  Fix the cause, or explain why it cannot be fixed.
- Never read, print or commit secrets: `.env` files, kubeconfigs, API keys.
- Read the relevant ADR in `docs/adr/` before changing a design decision. A new
  significant decision needs a new ADR.

## Contributing

Commits follow Conventional Commits and are signed off (`git commit -s`).
Changes reach `main` only through pull requests. See `CONTRIBUTING.md` and
`docs/maintainers/branch-protection.md`.
