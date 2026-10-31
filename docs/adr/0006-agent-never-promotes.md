# 0006. The agent recommends and never promotes

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

An AI agent with access to the cluster and the repository could, in
principle, merge promotions itself. That would make promoloupe faster, but it
would also mean that a wrong judgement, a hallucination or a successful prompt
injection could change production directly.

Teams adopting AI in their delivery process need to trust it gradually, and
auditors need a clear record of who decided what.

## Decision

The reviewer agent only produces a recommendation. It has read-only access to
the cluster and to observability data, and no write access to the GitOps
repository. Its output is a structured verdict in the `PromotionReview`
status and a comment on the merge request.

The promotion itself happens only when the merge request is merged, either by
a human or by an automerge rule defined explicitly in the `PromotionPolicy`.
If the agent fails, times out or returns an invalid verdict, the review is
marked as failed and nothing is promoted.

## Consequences

The worst outcome of a faulty review is a bad recommendation that a human or
a policy still has to act on. Every promotion has a merge request as its
audit trail.

Fully automatic promotion is still possible, but it is a deliberate choice
written in the policy, never a side effect of the agent's access.