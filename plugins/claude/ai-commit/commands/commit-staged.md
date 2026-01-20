---
allowed-tools: Bash(git status:*), Bash(git diff:*), Bash(git ai-commit:*)
description: Commit staged changes only using git-ai-commit
---

## Context

- Current git status: !`git status`
- Staged diff only: !`git diff --staged`
- Current branch: !`git branch --show-current`
- Recent commits: !`git log --oneline -10 2>/dev/null || true`

## Your task

Create a Git commit **only from already staged changes**.

Follow these rules strictly:

- Do NOT stage, unstage, or modify any files.
- If there are no staged changes, do nothing.
- Generate and apply the commit by invoking `git ai-commit`.
- Pass a short, human-readable summary via `--context` that describes the intent and the logical unit of this commit.
- Do not manually write a commit message and run `git commit` directly.
- Do not output explanations or confirmations.

Respond with tool invocations only.
