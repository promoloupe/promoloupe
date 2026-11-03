#!/usr/bin/env bash
# PostToolUse hook for Claude Code: formats and lints each file Claude edits.
# Exit code 2 sends the problems back to Claude so it can fix them.
set -euo pipefail

file="$(jq -r '.tool_input.file_path // empty')"
if [[ -z "$file" || ! -f "$file" ]]; then
  exit 0
fi

cd "$CLAUDE_PROJECT_DIR"

case "$file" in
  *.go)
    if ! output="$(gofmt -l -w "$file" 2>&1)"; then
      echo "gofmt failed on $file:" >&2
      echo "$output" >&2
      exit 2
    fi
    ;;
  */reviewer/*.py)
    uv run --project reviewer ruff format --quiet "$file" >/dev/null 2>&1 || true
    if ! output="$(uv run --project reviewer ruff check --fix --quiet "$file" 2>&1)"; then
      echo "ruff found problems in $file:" >&2
      echo "$output" >&2
      exit 2
    fi
    ;;
esac

exit 0
