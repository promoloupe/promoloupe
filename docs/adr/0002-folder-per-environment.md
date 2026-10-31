# 0002. One folder per environment

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

GitOps repositories model environments in several ways: a branch per
environment, a repository per environment, or a folder per environment inside
one repository. Branch-per-environment is widely considered an anti-pattern
because branches drift and merges between them carry unrelated changes.

The most common layouts today are Kustomize overlays (`overlays/dev`,
`overlays/prod`) and per-environment Helm values files, both inside a single
branch.

## Decision

promoloupe supports one folder per environment on a single branch. Each
`PromotionPolicy` names the folder of every environment in its chain, for
example `envs/dev`, `envs/staging` and `envs/prod`.

Branch-per-environment and repository-per-environment layouts are out of scope.

## Consequences

A promotion is always a small, readable diff between two folders on the same
branch, which is exactly what a human and the agent need to review.

Teams using branch-per-environment must restructure their repository before
adopting promoloupe. We accept this because supporting every layout would make
the diff and the review far harder to reason about.
