# git-ai-commit

Generate Git commit messages from staged diffs using your preferred LLM CLI.

## Directory Structure

- Follow the standard Go project layout.

## Documentation

- @README.md – User guide

## Build Commands

- `mise run build` – build the binary.
- `mise run test` – run the full test suit.
- `mise run lint` – run lint.

## Development Guidelines

- Use the red/green/refactor TDD cycle.
- Use `git ai-commit` to create commits.
  - Always include a short, clear summary in English using the `--context` option.
