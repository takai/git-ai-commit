package app

import (
	"testing"

	"git-ai-commit/internal/config"
)

func TestApplyPromptStrategy(t *testing.T) {
	base := "base"
	override := "override"

	got, err := applyPromptStrategy(base, override, "append")
	if err != nil {
		t.Fatalf("append error: %v", err)
	}
	if got != "base\noverride" {
		t.Fatalf("append = %q", got)
	}

	got, err = applyPromptStrategy(base, override, "prepend")
	if err != nil {
		t.Fatalf("prepend error: %v", err)
	}
	if got != "override\nbase" {
		t.Fatalf("prepend = %q", got)
	}

	got, err = applyPromptStrategy(base, override, "replace")
	if err != nil {
		t.Fatalf("replace error: %v", err)
	}
	if got != "override" {
		t.Fatalf("replace = %q", got)
	}

	_, err = applyPromptStrategy(base, override, "unknown")
	if err == nil {
		t.Fatalf("expected error for unknown strategy")
	}
}

func TestSelectEngineCodexDefault(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultEngine = "codex"
	cfg.Engines = map[string]config.EngineConfig{}

	_, command, err := selectEngine(cfg)
	if err != nil {
		t.Fatalf("selectEngine error: %v", err)
	}
	if command != "codex exec" {
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
