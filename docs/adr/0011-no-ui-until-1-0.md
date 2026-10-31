# 0011. No UI until 1.0, then an OpenShift console plugin

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

A web UI makes a tool easier to adopt, but building and maintaining a
standalone one is a large effort: a frontend stack, authentication,
authorisation and a separate release cycle. Before 1.0 the API will still
change, and every change would also have to be carried into the UI.

Most users already work in a console they trust. OpenShift supports dynamic
console plugins that add pages for custom resources, and Argo CD supports UI
extensions with a similar purpose.

## Decision

promoloupe has no user interface of its own before 1.0. Until then, the merge
request is the main review interface, complemented by the CLI and `kubectl`.

After 1.0, the first UI will be an OpenShift console dynamic plugin that lists
promotions and displays reviews next to the resources they refer to. An Argo
CD UI extension is the alternative if there is more demand from users outside
OpenShift.

## Consequences

Effort before 1.0 goes into the controller, the reviewer and a stable API.
Users who rely on a graphical console will have to wait for 1.0.

Because the UI will be a plugin of an existing console, it reuses that
console's authentication and RBAC instead of implementing its own.