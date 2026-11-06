# promoloupe

[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/promoloupe/promoloupe/badge)](https://scorecard.dev/viewer/?uri=github.com/promoloupe/promoloupe)

**Agentic risk review for GitOps promotions.**

promoloupe watches promotions between environments in a GitOps repository
(for example `dev` → `staging` → `prod`), opens the merge request for each one
and attaches an explained risk review written by an AI agent. The agent
inspects what changes, what the target environment looks like right now and
what went wrong last time. It recommends; it never promotes. A human (or a
policy you define) still merges.

> [!WARNING]
> promoloupe is in early development. Nothing here is usable yet, APIs will
> change without notice, and there are no releases. Follow the
> [roadmap](#roadmap) to see where things stand.

## Why

Promotion tools answer *how* a change moves to the next environment. They
rarely answer *should it move now*. That judgement usually lives in someone's
head: the diff looks small, but the target cluster is degraded, a dependency
was bumped in the same release, or the last rollout of this service failed on
the same config key.

promoloupe makes that judgement explicit, reviewable and repeatable:

- **Explained, not opaque.** Every review states what was checked, what was
  found and why it matters, with links to the evidence.
- **Recommend, never act.** The agent has read-only access to your cluster and
  observability stack. Promotion happens only through a merge request.
- **GitOps native.** Environments are folders in Git, promotable values are
  explicit YAML paths, and the merge request is the review interface.
- **Kubernetes native.** An operator reconciles `PromotionPolicy` and
  `PromotionReview` resources, so promotions are visible with the tools you
  already use.

## How it works

1. A change lands in an environment folder (for example, a new image tag in
   `envs/dev`).
2. The controller evaluates the `PromotionPolicy` for the next environment:
   gates such as Argo CD health, soak time, approvals and freeze windows.
3. When the gates pass, it creates a `PromotionReview` and opens a merge
   request with the promoted values.
4. The reviewer agent gathers context through read-only MCP tools
   (Kubernetes, Prometheus, Argo CD), writes a structured verdict into the
   `PromotionReview` status and comments on the merge request.
5. A human or an automerge rule decides. If the agent fails, nothing is
   promoted.

## Components

| Component | Language | Role |
| --- | --- | --- |
| `promoloupe-controller` | Go | Kubernetes operator: policies, gates, merge requests |
| `promoloupe-reviewer` | Python | AI agent built on the Claude Agent SDK |
| `promoloupe` CLI | Go | Inspect promotions and stream reviews; also a `kubectl` plugin |

## Roadmap

1. **Foundations.** Monorepo, tooling, CI, community files. *(in progress)*
2. **Operator.** CRDs, reconciliation, gates, GitLab adapter.
3. **Distribution.** Signed images, SBOM, Helm chart, CLI releases.
4. **First AI layer.** Change summaries and log triage.
5. **Reviewer agent.** Full agentic review with MCP tools, guardrails and evals.
6. **Maturity and 1.0.** API stabilisation, docs, observability, hardening.
7. **UI.** OpenShift console dynamic plugin.

Architecture decisions are recorded in [`docs/adr`](docs/adr).

## Contributing

Contributions are welcome once the foundations are in place. Read
[CONTRIBUTING.md](CONTRIBUTING.md) first. Everyone taking part is expected to
follow the [Code of Conduct](CODE_OF_CONDUCT.md).

To report a security issue, see [SECURITY.md](SECURITY.md). Please do not open
a public issue.

## License

promoloupe is licensed under the [Apache License 2.0](LICENSE).
