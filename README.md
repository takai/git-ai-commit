# git-ai-commit

AI-assisted Git commit messages from staged diffs.

## Install

Build the binary into `bin/`:

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

Settings are loaded in order (later overrides earlier):

1. User config: `~/.config/git-ai-commit/config.toml`
2. Repo config: `.git-ai-commit.toml` at the repository root
3. Command-line flags

This lets you set your preferred engine in user config and override just the prompt preset per repository.

Supported settings:

- `engine` Default engine name (string)
- `prompt` Bundled prompt preset: `default`, `conventional`, `gitmoji`, `karma`
- `prompt_file` Path to a custom prompt file (relative to the config file)
- `engines.<name>.args` Argument list for the engine command (array of strings)

Note: `prompt` and `prompt_file` are mutually exclusive within the same config file. If both are set, an error is returned. When settings come from different layers (user config vs repo config), the later layer wins.

Supported engine names (by convention):

- `claude`
- `codex`
- `cursor-agent`
- `gemini`

If no engine is configured, auto-detection tries commands in this order: `claude` → `gemini` → `codex`. The first available command is used.

Any engine name not listed above is treated as a direct command. For example, `engine = "my-llm-cli"` will execute `my-llm-cli` with the prompt on stdin.

Example:

```toml
engine = "codex"
prompt = "conventional"

[engines.codex]
args = ["exec", "--model", "gpt-5-mini"]
```

Using a custom prompt file:

```toml
engine = "claude"
prompt_file = "prompts/commit.md"
```

### Engine examples

By default, the prompt is sent via stdin. For CLIs that require the prompt as an argument, use `{{prompt}}` as a placeholder in the args list.

Claude:

```toml
engine = "claude"

[engines.claude]
args = ["-p", "--model", "haiku"]
```

Codex:

```toml
engine = "codex"

[engines.codex]
args = ["exec", "--model", "gpt-5-mini"]
```

Cursor agent:

```toml
engine = "cursor-agent"

[engines.cursor-agent]
args = ["-p"]
```

Gemini:

```toml
engine = "gemini"

[engines.gemini]
args = ["-m", "gemini-2.5-flash", "-p", "{{prompt}}"]
```

## Prompt presets

Bundled presets live in `internal/config/assets/`:

- `default`
- `conventional`
- `gitmoji`
- `karma`

## Safety

- No commit if there is no staged diff
- No commit if `--amend` is set and there is no previous commit
- No commit on engine failure or empty output
- Git state is not modified on errors

## Release acceptance check

For manual, pre-release verification across engines and prompt presets:

```sh
./scripts/release-check.sh
```

The script writes a report under `tmp/acceptance-<timestamp>/` with a
`summary.md` checklist for human review.
