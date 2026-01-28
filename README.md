# git-ai-commit

Generate Git commit messages from staged diffs using your preferred LLM CLI.

git-ai-commit does not talk to LLM APIs directly.
Instead, it delegates generation to existing LLM CLIs such as Claude Code, Gemini, or Codex, so no API keys, SDKs, or vendor-specific integrations are required.

## Install

Package managers:

Homebrew:

```sh
brew tap takai/tap
brew install git-ai-commit
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
- `-a`, `--all` Stage modified and deleted files before generating the message
- `-i`, `--include VALUE` Stage specific files before generating the message
- `--debug-prompt` Print the prompt before executing the engine
- `--debug-command` Print the engine command before execution
- `-h`, `--help` Show help

## Configuration

Configuration is layered, allowing global defaults with per-repository overrides:

1. User config: `~/.config/git-ai-commit/config.toml`
2. Repo config: `.git-ai-commit.toml` at the repository root
3. Command-line flags

This makes it easy to keep personal preferences (engine, style) while enforcing repository-specific commit rules without relying on hosted services. Repository config is applied only after an initial trust prompt.

Example: Use Codex with Conventional Commits by default

```toml
engine = "codex"
prompt = "conventional"
```

Supported settings:

- `engine` Default engine name (string)
- `prompt` Bundled prompt preset: `default`, `conventional`, `gitmoji`, `karma`
- `prompt_file` Path to a custom prompt file (relative to the config file)
- `engines.<name>.args` Argument list for the engine command (array of strings)
- `filter.max_file_lines` Maximum lines per file in diff (default: 200)
- `filter.exclude_patterns` Additional glob patterns to exclude from diff
- `filter.default_exclude_patterns` Override built-in exclude patterns

### Engines

git-ai-commit treats LLMs as external commands, not as APIs. This design avoids direct network calls and API key management, and lets you reuse your existing LLM CLI setup.

Supported engines:

- `claude`
- `gemini`
- `codex`

If no engine is configured, auto-detection tries commands in this order: `claude` → `gemini` → `codex`. The first available command is used.

Any other engine name is treated as a direct command and executed with the prompt on stdin.

Example: Use ollama with `gemma3:4b`

```toml
engine = "ollama"

[engines.ollama]
args = ["run", "gemma3:4b"]
```

### Prompt presets

Bundled presets live in `internal/config/assets/`:

- `default` – Commit messages aligned with the recommendations in [Pro Git](https://git-scm.com/book/ms/v2/Distributed-Git-Contributing-to-a-Project)
- `conventional` – [Conventional Commits](https://www.conventionalcommits.org/) format
- `gitmoji` – [gitmoji](https://gitmoji.dev/)-based commit messages
- `karma` – [Karma-style](https://karma-runner.github.io/6.4/dev/git-commit-msg.html) commit messages

### Custom Prompts

Example: Use a custom prompt file

```toml
engine = "claude"
prompt_file = "prompts/commit.md"
```

Note: `prompt` and `prompt_file` are mutually exclusive within the same config file. If both are set, an error is returned. When settings come from different layers (user config vs repo config), the later layer wins.

### Diff Filtering

When the staged diff is large, it can exceed LLM context limits or degrade commit message quality. git-ai-commit automatically filters the diff to help LLMs focus on meaningful changes.

**Default behavior:**

- Each file is limited to 200 lines (configurable via `filter.max_file_lines`)
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

- `/ai-commit:organize-commits` - Analyzes all changes and organizes them into atomic commits
- `/ai-commit:commit-staged` - Commits only the currently staged changes

## Acknowledgements

This tool was inspired by how [@negipo](https://github.com/negipo) used a tool like this in his workflow to make his work much more efficient.
