# Design Doc: AI-assisted Git Commit CLI

## Overview

This tool is a CLI that generates Git commit messages using an LLM based on Git diffs and optional user-provided context, then safely executes `git commit` with the generated message.
The command name is `git-ai-commit`, and it will also be invokable as `git ai-commit`.

The primary responsibility of the tool is orchestration between Git and LLMs, not generation logic itself. The goal is to reduce friction at commit time while keeping commit history high-quality and safe.

---

## Goals

- Reduce cognitive load when writing commit messages
- Keep commit message quality consistent
- Never corrupt or pollute Git history
- Be usable from CLI tools and development agents

---

## Non-Goals

- Reimplementing Git behavior
- Automating PRs or releases
- Learning from commit history
- Providing GUI or IDE integrations

---

## Core Design Principles

1. **Thin Wrapper**  
   Git handles Git. LLMs handle generation. This tool coordinates.

2. **Safe by Default**  
   If preconditions fail or generation fails, no commit is performed.

3. **Configurable, Not Opinionated**  
   The tool provides good defaults but allows users to override rules and formats.

4. **Engine Agnostic**  
   LLM backends are interchangeable via a common interface.

---

## Input Model

The following inputs are passed to the LLM:

- Git diff (required)
- System prompt (rules, conventions, output format)
- Additional context (optional)
  - Background
  - Intent
  - Constraints

Git diffs and user context are treated as separate logical inputs.

---

## Output Model

- LLM output is treated as a commit message
- Output format is defined entirely by the system prompt
  - Single-line or multi-line is user-configurable
- Output is normalized (e.g. trimming whitespace)
- Empty or invalid output results in no commit

---

## LLM Engine Abstraction

The tool supports multiple LLM execution backends:

- Development agent CLIs  
  - claude code  
  - codex  
  - cursor-agent  

- Local LLMs  
  - ollama  

- APIs  
  - OpenAI  
  - Claude  
  - Gemini  

All backends conform to a shared interface:

- Input: text (diff + prompt + context)
- Output: text (commit message)

Backend-specific I/O differences are handled internally.

---

## Configuration

- Configuration format: TOML
- Configuration follows the XDG Base Directory Specification
- The tool works with zero configuration

Configurable items include:

- Default LLM engine
- Engine-specific settings (CLI command, model name, API references)
- Default system prompt
- Prompt override strategy (replace / prepend / append)

---

## Safety Rules

- Do not commit if no staged diff exists
- Do not commit if `--amend` is requested and no previous commit exists
- Do not commit if the diff is empty
- Do not commit if LLM execution fails
- Do not modify Git state on errors

Safety rules always override user-provided prompts.

---

## Success Criteria

- Feels as fast and natural as `git commit`
- Users stop noticing the tool itself
- Commit history quality improves organically
- Identical behavior across CLI, agents, and automation
