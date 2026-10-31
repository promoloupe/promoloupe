# 0005. Run as an in-cluster operator

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

promoloupe could run in two ways. As a CI job, it would be triggered by
pipeline events and exit when done. As a Kubernetes operator, it would run
continuously inside the cluster and reconcile resources.

A CI job is simpler to install, but it only acts when a pipeline runs. It
cannot notice that the target environment became unhealthy after the review
was written, that a soak time has elapsed, or that a freeze window has ended.
It also needs cluster credentials stored in the CI system.

## Decision

promoloupe runs as an operator inside the cluster. It reconciles
`PromotionPolicy` and `PromotionReview` resources, watches the GitOps
repository and re-evaluates gates over time.

A CI mode may be added later as a thin client of the same logic, but it is not
part of the core design.

## Consequences

Gates such as soak time and freeze windows work naturally, because the
controller re-evaluates them on its own schedule. Promotions are visible with
standard Kubernetes tooling, and access to the cluster is governed by RBAC
instead of credentials held in CI.

Installation requires cluster-admin rights to install CRDs, and the
controller must be operated like any other workload: upgrades, monitoring
and resource limits.
