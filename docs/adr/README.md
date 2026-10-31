# Architecture Decision Records

This directory records the significant architecture decisions of promoloupe,
following the format described by Michael Nygard in
[Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).

Each record is immutable once accepted. To change a decision, add a new record
that supersedes the old one and update the status of both.

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-go-operator-python-agent.md) | Go operator with a Python agent service | Accepted |
| [0002](0002-folder-per-environment.md) | One folder per environment | Accepted |
| [0003](0003-explicit-promotable-paths.md) | Explicit promotable paths | Accepted |
| [0004](0004-gates-as-plugins.md) | Gates as plugins, Argo CD first | Accepted |
| [0005](0005-in-cluster-operator.md) | Run as an in-cluster operator | Accepted |
| [0006](0006-agent-never-promotes.md) | The agent recommends and never promotes | Accepted |
| [0007](0007-github-main-repository.md) | GitHub as the main repository | Accepted |
| [0008](0008-promotionreview-contract.md) | PromotionReview CRD as the contract | Accepted |
| [0009](0009-ports-and-adapters.md) | Ports and adapters | Accepted |
| [0010](0010-cli-as-kubectl-plugin.md) | Go CLI usable as a kubectl plugin | Accepted |
| [0011](0011-no-ui-until-1-0.md) | No UI until 1.0, then an OpenShift console plugin | Accepted |

New records use [template.md](template.md).
