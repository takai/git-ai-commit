# Repository Guidelines

## Project Structure & Module Organization
- `cmd/git-ai-commit/`: CLI entrypoint for the `git-ai-commit` binary.
- `internal/app/`: Orchestration logic, end-to-end flow, and acceptance tests.
- `internal/config/`: TOML config loading plus embedded prompt presets.
- `internal/git/`: Git interactions (diff, commit, staging, index safety).
- `internal/engine/`: LLM engine interface and CLI-backed implementation.
- `internal/prompt/`: Prompt assembly.
- `internal/config/assets/`: Prompt preset Markdown files (`default`, `conventional`, `gitmoji`, `karma`).
- `docs/`: Design documents.
- `tmp/`: Local scratch; content is ignored except `tmp/.keep`.

## Build, Test, and Development Commands
- `make build`: Builds `bin/git-ai-commit`.
- `make test`: Runs all Go tests.
- `go test ./...`: Direct test runner (same as `make test`).

## Coding Style & Naming Conventions
- Language: Go. Use standard `gofmt` formatting.
- Indentation: tabs (Go default).
- Filenames: lowercase with underscores only when needed (e.g., `config_test.go`).
- Package layout: keep CLI logic in `cmd/` and implementation in `internal/`.

## Testing Guidelines
- Framework: Go `testing` package.
- Unit tests live next to source (e.g., `internal/git/git_test.go`).
- Acceptance tests are in `internal/app/acceptance_test.go` and use temp git repos.
- Run all tests before commits: `go test ./...`.

## Commit & Pull Request Guidelines
- Commit messages follow Conventional Commits: `type: short summary`.
  Examples: `feat: add debug output options`, `docs: add README`.
- Keep summaries under ~72 characters, no trailing period.
- If contributing via PR, include a clear description and test results.

## Configuration Tips
- Config file: `~/.config/git-ai-commit/config.toml`.
- Example engine config:
  ```toml
  engine = "codex"
  prompt_preset = "default"

  [engines.codex]
  args = ["exec", "--model", "gpt-5-mini"]
  ```
