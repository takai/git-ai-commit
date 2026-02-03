---
allowed-tools: Bash(git status:*), Bash(git diff:*), Bash(git add:*), Bash(git reset:*), Bash(git ai-commit:*), Bash(git log:*)
context: fork
description: Organize changes into atomic commits using git-ai-commit
argument-hint: ""
---

## Context

- Current git status: !`git status`
- Staged diff: !`git diff --staged`
- Unstaged diff: !`git diff`
- Current branch: !`git branch --show-current`
- Recent commits: !`git log --oneline -10 2>/dev/null || true`

## Your task

Your goal is to organize the current set of unstaged/staged changes into atomic, logical commits.

You must create multiple commits if needed. Prefer small, coherent commits over a single large commit.

### Process

1) Analyze
- Use the context above (and additional `git status` / `git diff` tool calls if needed) to understand all pending changes.
- Mentally group changes into logical units (e.g., "refactor", "feature", "bugfix", "docs", "tests").

2) Execute loop (repeat for each logical unit)
- Step A: Identify the exact files or hunks that belong to one unit.
- Step B: Stage only that unit.
  - Use `git add <path>` to stage entire files.
  - When a file contains changes belonging to different logical units, stage the whole file in the unit where the majority of changes belong.
  - If you staged something by mistake, fix it with `git reset <path>`.
- Step C: Create the commit by invoking `git ai-commit`.
  - Pass a short, human-readable summary via `--context` that describes the intent and the logical unit of this commit.
  - Do not manually write a commit message and run `git commit` directly.
- Step D: Repeat until `git status` is clean (no staged or unstaged changes).

### Safety rules

- If `git ai-commit` fails, stop immediately and report the error. Do not retry.
- If the repository is in a merge, rebase, or cherry-pick state, stop immediately and report the state. Do not attempt commits.
- Do not loop more than 20 iterations. If changes remain after 20 commits, stop and report.

### Output rules

- Respond with tool invocations only (no explanations, no confirmations).
- Do not use any tools other than those listed in `allowed-tools`.
