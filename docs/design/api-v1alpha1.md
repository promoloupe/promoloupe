# API design: promoloupe.com/v1alpha1

- **Status:** Implemented
- **API version:** `promoloupe.com/v1alpha1`

This document specifies the custom resources that promoloupe exposes. The Go
types in `api/v1alpha1/` implement it, and any change to the types changes
this document in the same pull request. `v1alpha1` gives no compatibility
guarantees: fields may change or disappear between releases.

## Overview

promoloupe defines two namespaced kinds:

- **PromotionPolicy** is written by users. It describes a chain of
  environments in a GitOps repository and the rules a change must meet to be
  promoted from one environment to the next.
- **PromotionReview** is created by the controller, never by users. It
  records one promotion between two adjacent environments. It is the contract
  between the controller and the reviewer (ADR 0008) and the audit record of
  the promotion.

## Conventions

The types follow the Kubernetes API conventions and are checked by
kube-api-linter.

- `spec` holds the desired state and `status` the observed state. Status is
  written only through the status subresource.
- Every status has `observedGeneration` and a list of standard
  `metav1.Condition`. There are no `phase` fields.
- Durations are integers with the unit in the field name, such as
  `minSoakTimeMinutes`, so that clients in any language can read them.
- Enums are preferred over booleans, so that a field can gain new values
  later. Enum values are PascalCase.
- Every string and list has a maximum size, which also bounds the cost of the
  CEL validation rules.
- Credentials are referenced through Secrets, never stored in a resource.
- The controller and the reviewer write different status fields using
  server-side apply, with the field managers `promoloupe-controller` and
  `promoloupe-reviewer`.

## PromotionPolicy

```yaml
apiVersion: promoloupe.com/v1alpha1
kind: PromotionPolicy
metadata:
  name: checkout
  namespace: team-payments
spec:
  repository:
    provider: GitLab
    url: https://gitlab.com/acme/gitops
    branch: main
    credentialsSecretRef:
      name: gitlab-token
  environments:
    - name: dev
      path: apps/checkout/overlays/dev
    - name: staging
      path: apps/checkout/overlays/staging
      merge: Automatic
      gates:
        minSoakTimeMinutes: 60
        argoCD:
          application: checkout-dev
    - name: prod
      path: apps/checkout/overlays/prod
      gates:
        minSoakTimeMinutes: 1440
        requiredApprovals: 1
        freezeWindows:
          - schedule: "0 18 * * 5"
            durationMinutes: 3840
        argoCD:
          application: checkout-staging
  promotable:
    - Images
    - HelmChartVersions
status:
  observedGeneration: 3
  conditions:
    - type: Ready
      status: "True"
      reason: Reconciled
      message: All environments are tracked.
  environments:
    - name: staging
      revision: 9f2c1ab
      pendingReview: checkout-staging-to-prod-9f2c1ab
```

### Spec

| Field | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| `repository.provider` | enum | Yes | | Git provider. `GitLab` is the only value in v1alpha1. |
| `repository.url` | string | Yes | | HTTPS URL of the GitOps repository. |
| `repository.branch` | string | No | `main` | Branch that holds every environment. |
| `repository.credentialsSecretRef.name` | string | Yes | | Secret in the same namespace with the access token under the key `token`. |
| `environments` | list | Yes | | Ordered chain of 2 to 10 environments. Each one is promoted to the next. |
| `environments[].name` | string | Yes | | DNS label, unique within the policy. |
| `environments[].path` | string | Yes | | Directory of the environment in the repository (ADR 0002). |
| `environments[].merge` | enum | No | `Manual` | How promotions into this environment are merged: `Manual` or `Automatic` when every gate passes. |
| `environments[].gates` | object | No | | Conditions a change must meet to be promoted into this environment. At least one when present. |
| `gates.minSoakTimeMinutes` | integer | No | | Minimum time, in minutes, the change must have run in the previous environment. 1 to 43200. |
| `gates.requiredApprovals` | integer | No | | Approvals required on the merge request. 1 to 10. |
| `gates.freezeWindows[].schedule` | string | Yes | | Cron expression, in UTC, for the start of a window in which promotions are blocked. |
| `gates.freezeWindows[].durationMinutes` | integer | Yes | | Length of the window in minutes. 1 to 43200. |
| `gates.argoCD.application` | string | Yes | | Argo CD Application of the previous environment that must be Healthy and Synced. |
| `gates.argoCD.namespace` | string | No | `argocd` | Namespace of the Argo CD Application. |
| `promotable` | list of enum | No | `[Images, HelmChartVersions]` | What travels between environments (ADR 0003). |

### Validation

- `environments` has between 2 and 10 items, and their names are unique.
- The first environment has no `gates`, because nothing is promoted into it.
- Promotions into the last environment cannot use `merge: Automatic`. The
  final promotion is always merged by a person.
- `repository.url` starts with `https://`.

### Status

| Field | Description |
| --- | --- |
| `observedGeneration` | Generation of the spec that the status reflects. |
| `conditions` | `Ready` is `True` when the repository is reachable and every environment is tracked. |
| `environments[].name` | Environment name. |
| `environments[].revision` | Commit of the repository last observed for the environment. |
| `environments[].pendingReview` | Name of the open PromotionReview from this environment to the next, if any. |

`kubectl get promotionpolicies` shows the `Ready` condition and the age.

## PromotionReview

```yaml
apiVersion: promoloupe.com/v1alpha1
kind: PromotionReview
metadata:
  name: checkout-staging-to-prod-9f2c1ab
  namespace: team-payments
  ownerReferences:
    - apiVersion: promoloupe.com/v1alpha1
      kind: PromotionPolicy
      name: checkout
spec:
  policyRef:
    name: checkout
  from: staging
  to: prod
  sourceRevision: 9f2c1ab
  changes:
    - path: apps/checkout/overlays/prod/kustomization.yaml
      field: images[name=checkout].newTag
      from: "1.4.2"
      to: "1.5.0"
status:
  observedGeneration: 1
  mergeRequest:
    url: https://gitlab.com/acme/gitops/-/merge_requests/42
  gates:
    - name: MinSoakTime
      result: Passed
    - name: RequiredApprovals
      result: Failed
      message: 0 of 1 approvals.
  verdict:
    recommendation: Hold
    risk: Medium
    summary: Version 1.5.0 changes the payment timeout and staging shows a rise in 5xx errors.
    evidence:
      - source: Prometheus
        detail: 5xx rate in staging rose from 0.1% to 0.8% after the deploy.
  conditions:
    - type: GatesPassed
      status: "False"
      reason: ApprovalsMissing
    - type: Reviewed
      status: "True"
      reason: VerdictPublished
    - type: Merged
      status: "False"
      reason: AwaitingApproval
```

### Spec

The spec is immutable after creation. If the source environment changes, the
controller creates a new PromotionReview instead of editing the existing one.

| Field | Type | Description |
| --- | --- | --- |
| `policyRef.name` | string | PromotionPolicy in the same namespace. |
| `from` | string | Source environment. |
| `to` | string | Target environment, the next one in the chain. Different from `from`. |
| `sourceRevision` | string | Commit of the source environment being promoted, as 7 to 64 lowercase hexadecimal characters. |
| `changes[].path` | string | File changed in the target environment. |
| `changes[].field` | string | Promotable field within the file. |
| `changes[].from` | string | Current value in the target environment. Omitted when the field does not exist yet. |
| `changes[].to` | string | Value promoted from the source environment. |

`changes` has between 1 and 100 items.

### Status

Each field has a single owner.

| Field | Owner | Description |
| --- | --- | --- |
| `observedGeneration` | controller | Generation of the spec that the status reflects. |
| `mergeRequest.url` | controller | Merge request opened for the promotion. |
| `gates[].name` | controller | `MinSoakTime`, `RequiredApprovals`, `FreezeWindows` or `ArgoCD`. |
| `gates[].result` | controller | `Passed`, `Failed`, or `Unknown` when the gate could not be evaluated. |
| `gates[].message` | controller | Explanation of the result. |
| `verdict.recommendation` | reviewer | `Approve`, `Hold` or `Block`. A recommendation only (ADR 0006). |
| `verdict.risk` | reviewer | `Low`, `Medium` or `High`. |
| `verdict.summary` | reviewer | Plain-language explanation of the change and its risk. |
| `verdict.evidence[]` | reviewer | Source and detail of each fact behind the verdict. |
| `conditions` | both | `GatesPassed` and `Merged` are owned by the controller, `Reviewed` by the reviewer. |

`kubectl get promotionreviews` shows the source and target environments, the
recommendation, the `Merged` condition and the age.

### Lifecycle

1. The controller detects a new revision in an environment and creates a
   PromotionReview for the next one, owned by the policy.
2. The controller opens the merge request and records it in the status.
3. The reviewer analyses the promotion and publishes its verdict in the
   status. The controller copies it to the merge request.
4. The controller evaluates the gates on every reconciliation.
5. When the gates pass, the controller merges if the target environment uses
   `merge: Automatic`. Otherwise a person merges.
6. If the reviewer fails, `Reviewed` stays `False` and nothing is promoted
   automatically (ADR 0008).

Reviews are named `<policy>-<from>-to-<to>-<short revision>`. Deleting a
policy deletes its reviews through the owner reference.

## Permissions

| Actor | PromotionPolicy | PromotionReview |
| --- | --- | --- |
| Users | Create, update, delete | Read |
| Controller | Read, write status | Create, read, delete, write status |
| Reviewer | None | Read, write status |

The reviewer cannot create, change or delete reviews, and cannot touch
policies. A wrong verdict can only delay a promotion, never cause one.

## Out of scope for v1alpha1

- Configuring how the verdict is used. In v1alpha1 it is always advisory: it
  is published on the merge request and never blocks it. A future `review`
  field may add a mode where a `Block` verdict prevents automatic merges.
- Git providers other than GitLab.
- Flux as a source of health checks (ADR 0004).
- Time zones for freeze windows, which use UTC.
