package app

import (
	"testing"

	"git-ai-commit/internal/config"
)

func TestSelectEngineCodexDefault(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultEngine = "codex"
	cfg.Engines = map[string]config.EngineConfig{}

	_, command, err := selectEngine(cfg)
	if err != nil {
		t.Fatalf("selectEngine error: %v", err)
	}
	if command != "codex exec --model gpt-5.1-codex-mini" {
		t.Fatalf("command = %q", command)
	}
}

func TestSelectEngineClaudeDefault(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultEngine = "claude"
	cfg.Engines = map[string]config.EngineConfig{}

	_, command, err := selectEngine(cfg)
	if err != nil {
		t.Fatalf("selectEngine error: %v", err)
	}
	if command != "claude -p --model haiku" {
		t.Fatalf("command = %q", command)
	}
}

func TestSelectEngineCursorAgentDefault(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultEngine = "cursor-agent"
	cfg.Engines = map[string]config.EngineConfig{}

	_, command, err := selectEngine(cfg)
	if err != nil {
		t.Fatalf("selectEngine error: %v", err)
	}
	if command != "cursor-agent -p" {
		t.Fatalf("command = %q", command)
	}
}

func TestSelectEngineGeminiDefault(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultEngine = "gemini"
	cfg.Engines = map[string]config.EngineConfig{}

	_, command, err := selectEngine(cfg)
	if err != nil {
		t.Fatalf("selectEngine error: %v", err)
	}
	if command != "gemini -m gemini-2.5-flash -p {{prompt}}" {
		t.Fatalf("command = %q", command)
	}
}

func TestSelectEngineCustomArgs(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultEngine = "claude"
	cfg.Engines = map[string]config.EngineConfig{
		"claude": {Args: []string{"-p", "--model", "sonnet"}},
	}

	_, command, err := selectEngine(cfg)
	if err != nil {
		t.Fatalf("selectEngine error: %v", err)
	}
	if command != "claude -p --model sonnet" {
		t.Fatalf("command = %q", command)
	}
}

func TestSanitizeMessageCodeFence(t *testing.T) {
	input := "```\nfeat: add thing\n```"
	got := sanitizeMessage(input)
	if got != "feat: add thing" {
		t.Fatalf("sanitizeMessage = %q", got)
	}
}

func TestSanitizeMessageInlineBackticks(t *testing.T) {
	input := "`feat: add thing`"
	got := sanitizeMessage(input)
	if got != "feat: add thing" {
		t.Fatalf("sanitizeMessage = %q", got)
	}
}
