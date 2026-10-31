# 0007. GitHub as the main repository

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

promoloupe targets GitOps repositories hosted on GitLab, so hosting the
project itself on GitLab could seem natural. However, most open source
projects in the cloud native ecosystem live on GitHub, and that is where
contributors, users and tooling expect to find them.

GitHub also provides the supply chain pieces this project relies on at no
cost for public repositories: GitHub Actions, GitHub Releases, the GitHub
Container Registry with OCI support for images and Helm charts, artifact
attestations, Dependabot, private vulnerability reporting and the OpenSSF
Scorecard action.

## Decision

The source code, issues, releases, container images, Helm chart and website
are hosted on GitHub, under the `promoloupe` organisation. The website is
served from GitHub Pages.

GitLab is used only as a test target. Integration tests run against a sample
GitOps repository hosted on GitLab, so the GitLab adapter is always tested
against the real API.

## Consequences

Contributors find the project where they expect it, and the whole release
pipeline uses one platform with one set of permissions.

The project depends on GitHub's free tier for public repositories. A GitLab
CI component, if one is added later, will be published separately in the
GitLab CI/CD Catalog and built from this repository.
