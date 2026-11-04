# Branch protection and repository settings

This document records how `main` is protected and how the repository is
configured on GitHub. GitHub settings are not versioned with the code, so this
file is the reference for both contributors and maintainers. Whenever a setting
changes, this file changes in a pull request too.

## What this means for contributors

- Every change reaches `main` through a pull request. Nobody pushes to `main`
  directly, maintainers included.
- Every commit is signed off under the
  [Developer Certificate of Origin](https://developercertificate.org/)
  (`git commit -s`).
- Every commit is cryptographically signed (GPG, SSH or S/MIME) and shows as
  **Verified** on GitHub. GitHub signs the final squash commit, but unsigned
  commits on your branch can still block the merge. See
  [Signing commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits).
- The pull request title is a valid
  [Conventional Commit](https://www.conventionalcommits.org/). It becomes the
  commit message on `main`.
- The four required checks pass and the branch is up to date with `main`. Use
  the **Update branch** button or rebase onto `main`.
- All review conversations are resolved.

## The `main` ruleset

| Rule | Value | Why |
| --- | --- | --- |
| Enforcement status | Active | |
| Bypass list | Empty | The rules apply to everyone, administrators included. |
| Target | Default branch (`~DEFAULT_BRANCH`) | Follows the default branch even if it is renamed. |
| Restrict deletions | On | `main` cannot be deleted. |
| Block force pushes | On | History on `main` is never rewritten. |
| Require linear history | On | `main` is a straight line of squash commits, easy to read, bisect and revert. |
| Require signed commits | On | Every commit on `main` is verified. |
| Require a pull request before merging | On, 0 approvals | The project has one maintainer, and GitHub does not let authors approve their own pull requests. Raise this to 1 when a second maintainer joins. |
| Require conversation resolution | On | Review comments cannot be left unanswered. |
| Additional approval for unattributed Copilot pull requests | On | GitHub default. |
| Allowed merge methods | Squash | One commit per pull request on `main`. |
| Require status checks to pass | On, branches up to date | Checks run against the code that will actually land on `main`. |

### Required status checks

| Check | Source | Defined in |
| --- | --- | --- |
| DCO | DCO app | [github.com/apps/dco](https://github.com/apps/dco) |
| Lint | GitHub Actions | `.github/workflows/lint.yml` |
| Unit tests | GitHub Actions | `.github/workflows/test.yml` |
| E2E tests on kind | GitHub Actions | `.github/workflows/test-e2e.yml` |

Each check is bound to its source, so a different app or workflow reporting a
status with the same name cannot satisfy it.

The reviewer workflow (`.github/workflows/reviewer.yml`) is not required. It
only runs when files under `reviewer/` change, and a required check that never
runs stays in "Expected" and blocks the merge. It will become required once the
workflows report through a single aggregate job.

## Repository settings

Under **Settings > General > Pull Requests**:

| Setting | Value | Why |
| --- | --- | --- |
| Allow merge commits | Off | Matches the linear history rule. |
| Allow squash merging | On | The only merge method. |
| Default squash commit message | Pull request title and commit details | Keeps the `Signed-off-by` lines of the branch commits in the squash commit. |
| Allow rebase merging | Off | |
| Always suggest updating pull request branches | On | Needed because checks must run on an up-to-date branch. |
| Automatically delete head branches | On | Merged branches are removed from GitHub. |

## DCO

The [DCO app](https://github.com/apps/dco) checks that every commit in a pull
request has a `Signed-off-by` line matching its author. To fix a failing check,
sign off the commits and push again:

```
git commit --amend -s --no-edit && git push --force-with-lease
```

For more than one commit:

```
git fetch origin && git rebase --signoff origin/main && git push --force-with-lease
```

Only force-push your own pull request branch, never `main` or someone else's
branch.

## Dependabot

Dependabot is configured in `.github/dependabot.yml` and opens grouped pull
requests every Monday. They go through the same ruleset as any other pull
request. Do not push commits to Dependabot branches. If a branch falls behind
`main`, comment `@dependabot rebase` on the pull request.

## Auditing the configuration

Compare GitHub with this document using the
[GitHub CLI](https://cli.github.com/). All the commands are read-only.

Pull request settings:

```
gh api repos/promoloupe/promoloupe --jq '{merge_commit: .allow_merge_commit, squash: .allow_squash_merge, rebase: .allow_rebase_merge, squash_title: .squash_merge_commit_title, squash_message: .squash_merge_commit_message, suggest_update: .allow_update_branch, delete_branch: .delete_branch_on_merge}'
```

The ruleset itself, with its enforcement, bypass list and target:

```
gh api repos/promoloupe/promoloupe/rulesets --jq '.[] | {id, name, enforcement, target}'
```

Then, with the `id` from that output:

```
gh api repos/promoloupe/promoloupe/rulesets/26372471 --jq '{name, enforcement, bypass_actors, conditions}'
```

The rules in force on `main`:

```
gh api repos/promoloupe/promoloupe/rules/branches/main --jq '.[] | {type, parameters}'
```

In that output, `deletion` is "Restrict deletions", `non_fast_forward` is
"Block force pushes", and `strict_required_status_checks_policy: true` means
branches must be up to date. In the required checks, `integration_id` 1861 is
the DCO app and 15368 is GitHub Actions. `PR_TITLE` with `COMMIT_MESSAGES` is
the "Pull request title and commit details" squash message.

If GitHub and this document differ, either a setting was changed by mistake
and must be restored, or the change was intended and this document needs a
pull request.
