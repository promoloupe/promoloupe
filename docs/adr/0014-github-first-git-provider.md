# 0014. GitHub as the first Git provider

- **Status:** Accepted
- **Date:** 2026-11-15

## Context

ADR 0007 hosted the project on GitHub and planned GitLab as the host of the
GitOps repositories that promoloupe promotes, with integration tests against
a sample repository on GitLab. ADR 0009 made GitLab the first adapter of the
`SCMProvider` port.

GitLab was chosen to set promoloupe apart from Kargo. Kargo, however, already
opens pull and merge requests on GitHub, GitLab, Bitbucket, Gitea and Azure
DevOps. What sets promoloupe apart is the agentic review, not the provider.

Two more facts weigh on the choice:

- The maintainer has far more experience with GitHub than with GitLab.
- A GitHub App gives a tool its own identity, with installation tokens that
  expire after an hour and only reach the repositories where the app is
  installed, on the free plan. On GitLab.com, project access tokens require
  a Premium or Ultimate subscription, and a separate identity requires a
  separate user account.

## Decision

GitHub is the first Git provider. In `v1alpha1`, `repository.provider` only
accepts `GitHub`, and the sample GitOps repository for integration tests
lives in the `promoloupe` organisation on GitHub.

The core stays independent of the provider. Every provider implements the
same `SCMProvider` port and passes the same contract tests. GitLab is the
next adapter, and its value returns to the enum in the same change.

The API uses the general term pull request, so the status field of a
PromotionReview is `pullRequest`.

## Consequences

The project, its tests and its sample repository use one platform and one
set of credentials, and promotions are authored by the app instead of a
person.

Teams on GitLab cannot use promoloupe until its adapter exists. Removing
`GitLab` from the enum and renaming `mergeRequest` are breaking changes,
accepted while the API is `v1alpha1` and before any user depends on them.
