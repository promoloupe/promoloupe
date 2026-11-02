# 0012. Versioned configuration for AI coding assistants

- **Status:** Accepted
- **Date:** 2026-11-02

## Context

Contributors increasingly work with AI coding assistants. Each session starts
without knowledge of the project: its two languages, its commands, which files
are generated and which conventions apply. Left to each contributor, this
context is inconsistent, and the limits on what an assistant may run or read
are never reviewed.

Instruction files are context for the assistant, not enforcement: an assistant
usually follows them but is not guaranteed to. Hard limits need tool
permissions and hooks, and the project's existing gates (pre-commit, CI and the
`main` ruleset) apply to every change regardless of how it was written.

## Decision

The repository versions its assistant configuration in two layers.

Instructions:

- `AGENTS.md` at the root is the single source of truth, in the open AGENTS.md
  format read by most coding assistants.
- Each component with its own rules has its own `AGENTS.md`, starting with
  `reviewer/AGENTS.md`.
- `CLAUDE.md` files import the `AGENTS.md` next to them and add only what is
  specific to Claude Code.

Enforcement:

- `.claude/settings.json` defines shared permissions and hooks for Claude Code:
  commands that run without asking, files that are never read and checks that
  run after every edit.
- Personal files (`CLAUDE.local.md`, `.claude/settings.local.json`) are ignored
  by git.

Contributions made with AI assistance follow the policy in CONTRIBUTING.md: the
person who opens the pull request is responsible for every line in it.

## Consequences

Every contributor's assistant starts with the same context and limits, and
changes to them are reviewed in pull requests like any other code.

The instruction files must be kept in sync with the Makefile, the tooling and
the conventions they describe; an outdated instruction is worse than none.
Contributors who do not use AI assistants are not affected.
