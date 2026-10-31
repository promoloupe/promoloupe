# 0004. Gates as plugins, Argo CD first

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Before a promotion is proposed, some conditions must hold: the source
environment is healthy, the change has been running long enough, the right
people have approved, and no freeze window is active. Every organisation
checks a different mix of these, and the health signal comes from different
places depending on the delivery tool in use.

Hard-coding a fixed set of checks would make promoloupe either too rigid for
most teams or bloated with options nobody uses.

## Decision

Each check is a gate behind a common Go interface. A gate receives the
promotion context and returns one of three results: passed, failed or
pending, with a human-readable reason. A `PromotionPolicy` lists the gates
that apply to each step of the environment chain.

The first release ships four gates:

- Argo CD application health and sync status
- soak time since the change reached the source environment
- required approvals on the merge request
- freeze windows defined in the policy

Argo CD is the first health provider because it is the most widely used GitOps
delivery tool. Other providers, such as Flux, can be added as new gates
without changing the controller.

## Consequences

New checks are added by implementing one interface and registering it, with
no changes to the reconciliation logic. Each gate can be tested in isolation.

Gates are compiled into the controller, so adding one means a new release.
Loading gates at runtime (for example over gRPC or WebAssembly) is out of
scope until there is clear demand for it.
