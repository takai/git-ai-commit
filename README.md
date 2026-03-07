# git-ai-commit

Generate Git commit messages from staged diffs using your preferred LLM CLI.

git-ai-commit does not talk to LLM APIs directly.
Instead, it delegates generation to existing LLM CLIs such as Claude Code, Gemini, or Codex, so no API keys, SDKs, or vendor-specific integrations are required.

## Install

Package managers:

Homebrew:

```sh
brew install takai/tap/git-ai-commit
```

mise:

```sh
mise use -g github:takai/git-ai-commit@latest
```

Build from source (outputs to `bin/`):

```sh
make build
```

Put `bin/git-ai-commit` on your `PATH` to enable `git ai-commit`.

## Usage

```sh
git ai-commit [options]
```

Common options:

- `--context VALUE` Additional context for the commit message
- `--context-file VALUE` File containing additional context
- `--prompt VALUE` Bundled prompt preset: `default`, `conventional`, `gitmoji`, `karma`
- `--prompt-file VALUE` Path to a custom prompt file
- `--engine VALUE` Override engine name
- `--amend` Amend the previous commit
- `-e`, `--edit` Open the generated commit message in an editor before committing
- `-a`, `--all` Stage modified and deleted files before generating the message
- `-i`, `--include VALUE` Stage specific files before generating the message
- `-x`, `--exclude VALUE` Hide specific files from the diff for message generation
- `--debug-prompt` Print the prompt before executing the engine
- `--debug-command` Print the engine command before execution
- `-h`, `--help` Show help

## Configuration

Configuration is layered. Later layers override earlier ones:

| Priority | Source |
|----------|--------|
| 1 (lowest) | System git config (`/etc/gitconfig`) |
| 2 | Global git config (`~/.gitconfig`) |
| 3 | User TOML (`~/.config/git-ai-commit/config.toml`) |
| 4 | Repo TOML (`.git-ai-commit.toml` at repo root) |
| 5 | Local git config (`.git/config`) |
| 6 | Worktree git config |
| 7 (highest) | Command-line flags |

Repository TOML config is applied only after an initial trust prompt, since it is a tracked file that could be set by a repo maintainer.

### TOML config files

`~/.config/git-ai-commit/config.toml` for user-wide defaults, `.git-ai-commit.toml` at the repo root for project defaults.

Example: Use Codex with Conventional Commits by default

```toml
engine = "codex"
prompt = "conventional"
```

Supported settings:

- `engine` Default engine name (string)
- `prompt` Bundled prompt preset: `default`, `conventional`, `gitmoji`, `karma`
- `prompt_file` Path to a custom prompt file (relative to the config file; must be within the repo root for repo TOML)
- `engines.<name>.args` Argument list for the engine command (array of strings)
- `filter.max_file_lines` Maximum lines per file in diff (default: 100)
- `filter.exclude_patterns` Additional glob patterns to exclude from diff
- `filter.default_exclude_patterns` Override built-in exclude patterns

### git config

All settings except `engines.<name>.args` can also be set via `git config` using the `ai-commit` section. This is useful for per-repository preferences in repositories you do not own, since `.git/config` is never committed or pushed.

```sh
# Set for the current repository only
git config --local ai-commit.engine claude
git config --local ai-commit.prompt conventional

# Or apply user-wide defaults
git config --global ai-commit.engine claude
git config --global ai-commit.prompt conventional
```

Supported keys:

| git config key | Equivalent TOML setting |
|----------------|------------------------|
| `ai-commit.engine` | `engine` |
| `ai-commit.prompt` | `prompt` |
| `ai-commit.promptFile` | `prompt_file` |
| `ai-commit.maxFileLines` | `filter.max_file_lines` |
| `ai-commit.excludePatterns` | `filter.exclude_patterns` |
| `ai-commit.defaultExcludePatterns` | `filter.default_exclude_patterns` |

`excludePatterns` and `defaultExcludePatterns` support multiple values via `git config --add`:

```sh
git config --add ai-commit.excludePatterns '*.pb.go'
git config --add ai-commit.excludePatterns 'vendor/**'
```

Relative `promptFile` paths are resolved from the repo root for `--local`/`--worktree` scope, and from `$HOME` for `--global` scope. No path containment restriction applies — unlike repo TOML, git config cannot be set by a repository maintainer via push.

`prompt` and `promptFile` cannot both be set within the same scope. Setting both returns an error.

### Engines

git-ai-commit treats LLMs as external commands, not as APIs. This design avoids direct network calls and API key management, and lets you reuse your existing LLM CLI setup.

Supported engines:

- `claude`
- `gemini`
- `codex`

Built-in defaults are applied when `engines.<name>.args` is not set.
For `claude`, defaults include:

- `-p --model haiku`
- `--settings "{\"attribution\":{\"commit\":\"\",\"pr\":\"\"}}"` (prevents automatic `Co-authored-by` metadata)

If no engine is configured, auto-detection tries commands in this order: `claude` → `gemini` → `codex`. The first available command is used.

Any other engine name is treated as a direct command and executed with the prompt on stdin.

Example: Use ollama with `gemma3:4b`

```toml
engine = "ollama"

[engines.ollama]
args = ["run", "gemma3:4b"]
```

### Prompt presets

Bundled presets:

- `default` – Commit messages aligned with the recommendations in [Pro Git](https://git-scm.com/book/ms/v2/Distributed-Git-Contributing-to-a-Project)
- `conventional` – [Conventional Commits](https://www.conventionalcommits.org/) format
- `gitmoji` – [gitmoji](https://gitmoji.dev/)-based commit messages
- `karma` – [Karma-style](https://karma-runner.github.io/6.4/dev/git-commit-msg.html) commit messages

### Custom Prompts

Point to a custom prompt file in TOML config:

```toml
engine = "claude"
prompt_file = "prompts/commit.md"
```

Or via git config (useful for repos you do not own):

```sh
git config --local ai-commit.promptFile /path/to/prompt.md
```

`prompt` and `prompt_file` (or `promptFile` in git config) are mutually exclusive within the same config layer. When they come from different layers, the higher-priority layer wins.

### Diff Filtering

When the staged diff is large, it can exceed LLM context limits or degrade commit message quality. git-ai-commit automatically filters the diff to help LLMs focus on meaningful changes.

**Default behavior:**

- Each file is limited to 100 lines (configurable via `filter.max_file_lines`)
- Lock files and generated files are excluded by default

**Default exclude patterns:**

- `**/*.lock`, `**/*-lock.json`, `**/*.lock.yaml`, `**/*-lock.yaml`, `**/*.lockfile`
- `**/*.min.js`, `**/*.min.css`, `**/*.map`
- `**/go.sum`

Example: Customize filtering

```toml
[filter]
max_file_lines = 300

# Add patterns to exclude (merged with defaults)
exclude_patterns = [
    "**/vendor/**",
    "**/*.generated.go",
    "**/dist/**"
]

# Or replace defaults entirely
# default_exclude_patterns = ["**/my-lock.json"]
```

## Claude Code Plugin

If you use [Claude Code](https://docs.anthropic.com/en/docs/claude-code), you can integrate git-ai-commit as a plugin for a more convenient workflow.

### Installation

```
/plugin marketplace add takai/git-ai-commit
/plugin install ai-commit@git-ai-commit-plugins
```

### Commands

- `/ai-commit:staged` - Commits only the currently staged changes
- `/ai-commit:all` - Stages all pending changes and commits as a single commit
- `/ai-commit:organize` - Organizes pending changes into multiple atomic commits

## Acknowledgements

This tool was inspired by how [@negipo](https://github.com/negipo) used a tool like this in his workflow to make his work much more efficient.
