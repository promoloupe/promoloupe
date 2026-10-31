# 0003. Explicit promotable paths

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Environment folders differ on purpose. Replica counts, resource limits,
hostnames and secrets references should stay different between `dev` and
`prod`. Copying a whole folder forward would overwrite those differences.

Only some values are meant to travel: usually image tags and chart versions.

## Decision

A `PromotionPolicy` lists the YAML paths that are promotable, for example the
image tag in a Kustomize `images` entry or `image.tag` in a Helm values file.
When no paths are given, the defaults are image tags and chart versions.

Only the values at those paths are copied from the source environment to the
target environment. Everything else in the target folder is left untouched.

## Consequences

Promotions are predictable and environment-specific settings are never lost.
The merge request diff shows only the promoted values, which keeps reviews
focused.

Users must declare paths for anything beyond the defaults. A change outside the
promotable paths in the source environment is not promoted, and the agent
reports it in the review so it is not silently missed.
