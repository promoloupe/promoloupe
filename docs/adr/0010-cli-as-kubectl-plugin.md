# 0010. Go CLI usable as a kubectl plugin

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Users need a quick way to see pending promotions, inspect a review and follow
the agent's reasoning while it runs. Raw `kubectl get` output shows the
resources, but not in a form that is easy to read, and it cannot stream the
agent's progress.

kubectl discovers plugins automatically: any executable named `kubectl-<name>`
on the `PATH` becomes `kubectl <name>`. Many cloud native tools follow this
model and are distributed through Krew, the kubectl plugin manager.

## Decision

promoloupe ships a single Go binary, `promoloupe`. The same binary works as a
kubectl plugin when installed or linked as `kubectl-promoloupe`, so both
`promoloupe get reviews` and `kubectl promoloupe get reviews` are valid.

The CLI is built with Cobra and uses the standard kubeconfig loading rules,
including `--context` and `--namespace` flags. It can stream the review events
the reviewer emits while a review is in progress.

The binary is published in GitHub Releases for Linux, macOS and Windows, and
submitted to the Krew index once the API reaches beta.

## Consequences

Users get one tool that fits both workflows, and it shares the API types with
the controller because both are in the same Go module.

The CLI is a convenience. Everything it shows must also be available through
the Kubernetes API, so users who prefer plain `kubectl` or other tools lose
nothing.