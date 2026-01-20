# git-ai-commit-plugin

A Claude Code plugin that organizes git changes into atomic, logical commits using [git-ai-commit](https://github.com/takai/git-ai-commit) as the execution backend.

## Prerequisites

- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) CLI installed
- [git-ai-commit](https://github.com/takai/git-ai-commit) installed and configured

## Installation

Clone this repository into your Claude Code plugins directory:

```bash
git clone https://github.com/takai/git-ai-commit-plugin ~/.claude/plugins/git-ai-commit-plugin
```

## Usage

In any git repository with Claude Code, run:

```
/ai-commit:organize-commits
```

The plugin will:

1. Analyze all staged and unstaged changes
2. Group changes into logical units (refactor, feature, bugfix, docs, tests, etc.)
3. Create atomic commits for each unit using `git ai-commit`

## Commands

### `/ai-commit:organize-commits`

Organizes pending changes into multiple atomic commits. The command:

- Examines current git status and diffs
- Groups related changes together
- Stages files/hunks for each logical unit
- Invokes `git ai-commit` to generate commit messages automatically

### `/ai-commit:commit-staged`

Commits only the currently staged changes. The command:

- Works exclusively with already staged changes
- Does not stage, unstage, or modify any files
- Invokes `git ai-commit` to generate the commit message automatically

## License

MIT
