# 0008. PromotionReview CRD as the contract

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

The controller (Go) and the reviewer (Python) must exchange two things. The
controller has to ask for a review of a specific promotion, and the reviewer
has to return a verdict. This could be done over HTTP or gRPC, through a
message queue, or through Kubernetes resources.

Direct calls would couple the two services: the controller would need to
handle retries, timeouts and the reviewer being unavailable, and the exchange
would leave no record in the cluster.

## Decision

The `PromotionReview` custom resource is the only contract between the two
services.

- The controller creates a `PromotionReview` whose `spec` describes the
  promotion: policy, source and target environments, promoted values and the
  merge request.
- The reviewer watches `PromotionReview` resources, performs the review and
  writes the result to `status`: a verdict, a risk level, the findings with
  their evidence, the model used and the token cost.
- Only the reviewer's ServiceAccount may update the `status` subresource. Only
  the controller may create the resource.

The verdict schema is defined once in the CRD's OpenAPI schema. The reviewer
validates its output against Pydantic models generated from that schema
before writing it.

## Consequences

Each service can be restarted, upgraded or scaled on its own. Reviews are
visible with `kubectl`, survive restarts and can be audited later. Retries are
handled by normal reconciliation.

Every change to the verdict format is an API change to the CRD and follows
Kubernetes versioning rules. Large evidence (full logs, long diffs) cannot be
stored in the resource because of etcd size limits, so the status keeps
summaries and links instead.