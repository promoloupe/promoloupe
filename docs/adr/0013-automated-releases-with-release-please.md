# 0013. Automated releases with release-please

- **Status:** Accepted
- **Date:** 2026-11-05

## Context

promoloupe has no versions yet: no tags, no changelog and no GitHub releases.
Go modules are versioned through `vX.Y.Z` tags, and later phases will publish
images, a Helm chart and a CLI that need a common version. Commits on `main`
already follow Conventional Commits, and pull requests are squash-merged, so
each pull request title becomes one commit that a tool can read to decide the
next version.

Several tools were considered:

- **semantic-release** releases on every merge with no review step, and its
  plugin ecosystem is centred on Node.js.
- **GoReleaser** builds and publishes artifacts but does not decide versions
  or create tags.
- **git-cliff** with manual tags produces good changelogs, but versioning
  stays manual.
- **Changesets** requires a version intent file per pull request, which
  duplicates what Conventional Commits already express.

## Decision

We use [release-please](https://github.com/googleapis/release-please) through
its GitHub Action.

- A release pull request collects the changes since the last release and
  proposes the next version and changelog. Merging it creates the tag and the
  GitHub release, so publishing a version is a reviewed, explicit step.
- The whole project shares one version. The controller, the reviewer, the CLI
  and the chart depend on the same CRD contract, so they are released
  together. The single package lives at the repository root with the `go`
  release type, and tags are `vX.Y.Z` without a component prefix.
- While the version is below 1.0.0, breaking changes bump the minor version.
- The version in `reviewer/pyproject.toml` is not managed by release-please,
  because it cannot update `reviewer/uv.lock` and the lockfile check in CI
  would fail. The reviewer's released version is the image version, injected
  at build time.
- Release pull requests are opened with a token from a dedicated GitHub App.
  Resources created with `GITHUB_TOKEN` do not trigger workflows, so the
  required checks would never run on the release pull request.
- A required check validates that every pull request title is a Conventional
  Commit, since that title is what release-please reads.

## Consequences

- Versions and changelogs follow from commit history, so pull request titles
  must be accurate. The title check enforces the format, not the meaning.
- Releasing is a merge, with the same checks and review as any other change.
- The project depends on a GitHub App whose private key is stored as a
  repository secret and must be rotated if exposed.
- Building and publishing artifacts is out of scope here. GoReleaser will
  build on the tags that release-please creates.
