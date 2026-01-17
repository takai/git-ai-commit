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
