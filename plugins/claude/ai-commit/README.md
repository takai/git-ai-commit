# git-ai-commit-plugin

A Claude Code plugin that organizes git changes into atomic, logical commits using [git-ai-commit](https://github.com/takai/git-ai-commit) as the execution backend.

## Prerequisites

- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) CLI installed
- [git-ai-commit](https://github.com/takai/git-ai-commit) installed and configured

## Installation

If you installed `git-ai-commit` via `go install`, the plugin is already available at `plugins/claude/ai-commit/` inside the source tree.

To register it with Claude Code, add the plugin directory to your project or global settings:

```bash
# If you have the git-ai-commit source cloned
claude --plugin-dir /path/to/git-ai-commit/plugins/claude/ai-commit

# Or symlink into your plugins directory
ln -s /path/to/git-ai-commit/plugins/claude/ai-commit ~/.claude/plugins/ai-commit
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
- Stages files for each logical unit
- Invokes `git ai-commit` to generate commit messages automatically

### `/ai-commit:commit-staged`

Commits only the currently staged changes. The command:

- Works exclusively with already staged changes
- Does not stage, unstage, or modify any files
- Invokes `git ai-commit` to generate the commit message automatically

## License

MIT
