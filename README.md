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
- `--system-prompt VALUE` Override system prompt text
- `--prompt-strategy VALUE` `replace`, `prepend`, or `append`
- `--prompt-preset VALUE` `default`, `conventional`, `gitmoji`, `karma`
- `--engine VALUE` Override engine name
- `--amend` Amend the previous commit
- `-a`, `--all` Stage modified and deleted files before generating the message
- `-i`, `--include VALUE` Stage specific files before generating the message
- `--debug-prompt` Print the prompt before executing the engine
- `--debug-command` Print the engine command before execution
- `-h`, `--help` Show help

## Configuration

Config file (repo-local): `.git-ai-commit.toml` at the repository root

Fallback config: `~/.config/git-ai-commit/config.toml`

If the repo-local file exists, it takes precedence over the fallback config.

Supported settings:

- `engine` Default engine name (string)
- `system_prompt` Override the system prompt text (string)
- `prompt_strategy` How to merge `system_prompt` with default: `replace`, `prepend`, `append` (requires `system_prompt` to be set)
- `prompt_preset` Use a bundled prompt preset: `default`, `conventional`, `gitmoji`, `karma`
- `engines.<name>.args` Argument list for the engine command (array of strings)

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
prompt_preset = "default"

[engines.codex]
args = ["exec", "--model", "gpt-5-mini"]
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
