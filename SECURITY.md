# Security policy

promoloupe runs inside Kubernetes clusters and gives an AI agent read access
to cluster state and observability data. We take security reports seriously.

## Supported versions

promoloupe has no releases yet. Once releases begin, security fixes will be
provided for the latest minor version.

## Reporting a vulnerability

Please **do not** report security issues through public GitHub issues,
discussions or pull requests.

Report them privately through
[GitHub private vulnerability reporting](https://github.com/promoloupe/promoloupe/security/advisories/new).

Include as much of the following as you can:

- the affected component (controller, reviewer, CLI, Helm chart)
- the version or commit
- steps to reproduce, or a proof of concept
- the impact you expect

You can expect an acknowledgement within 5 working days. We will keep you
informed while we work on a fix and credit you in the advisory unless you
prefer otherwise.

## Scope

In addition to classic vulnerabilities, we consider these in scope:

- prompt injection that makes the reviewer agent take actions, call tools or
  reach endpoints outside its read-only design
- ways to make a promotion happen without a merge request, or with a forged
  or tampered review
- leaks of secrets, credentials or cluster data through reviews, logs or
  merge request comments
- RBAC or NetworkPolicy defaults in the Helm chart that grant more than
  documented

A review verdict you disagree with is not a vulnerability. Please open a
regular issue for that.
