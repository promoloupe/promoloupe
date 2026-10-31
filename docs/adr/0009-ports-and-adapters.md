# 0009. Ports and adapters

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

promoloupe talks to several external systems: a source code host to read the
repository and open merge requests, delivery tools to check health, and
optionally a database to store review history. The first release supports
GitLab and Argo CD, but users will ask for GitHub, Bitbucket, Flux and others.

If the reconciliation logic calls the GitLab and Argo CD clients directly,
every new integration means changing and retesting the core.

## Decision

The core logic depends only on interfaces (ports) defined in the domain
packages. Each external system is reached through an adapter that implements
one of those ports.

The initial ports are:

- `SCMProvider`: read files, create branches, open and update merge requests,
  post comments
- `Gate`: evaluate one promotion condition (see ADR 0004)
- `ReviewStore`: persist review history, with an in-cluster default and an
  optional PostgreSQL adapter

The first adapters are GitLab for `SCMProvider` and Argo CD for the health
gate. Adapters are selected in the `PromotionPolicy` or in the controller
configuration.

## Consequences

The reconciliation logic is tested with in-memory fakes of each port, so unit
tests do not need a network or a real GitLab instance. A new provider is a new
adapter with its own integration tests.

Interfaces add some indirection, and each port must be kept small so that
adapters for other providers remain realistic to write.
