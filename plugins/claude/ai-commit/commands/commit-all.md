---
allowed-tools: Bash(git status:*), Bash(git diff:*), Bash(git add:*), Bash(git ai-commit:*)
context: fork
description: Stage all changes and commit using git-ai-commit
argument-hint: ""
---

## Context

- Current git status: !`git status`
- Staged diff: !`git diff --staged`
- Unstaged diff: !`git diff`
- Current branch: !`git branch --show-current`
- Recent commits: !`git log --oneline -10 2>/dev/null || true`

## Your task

Stage **all** pending changes (both staged and unstaged) and create a single Git commit.

Follow these rules strictly:

- If there are no staged or unstaged changes, do nothing.
- If the repository is in a merge, rebase, or cherry-pick state, stop immediately and report the state. Do not attempt a commit.
- Stage all changes with `git add -A`.
- Generate and apply the commit by invoking `git ai-commit`.
- Pass a short, human-readable summary via `--context` that describes the intent and the logical unit of this commit.
- Do not manually write a commit message and run `git commit` directly.
- If `git ai-commit` fails, stop immediately and report the error. Do not retry.
- Do not output explanations or confirmations.

Respond with tool invocations only.
