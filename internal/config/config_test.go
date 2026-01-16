package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingConfig(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.SystemPrompt == "" {
		t.Fatalf("expected default system prompt")
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDir := filepath.Join(configHome, "git-ai-commit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	data := []byte("engine = 'codex'\nsystem_prompt = 'hi'\nprompt_strategy = 'replace'\n")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.DefaultEngine != "codex" {
		t.Fatalf("DefaultEngine = %q", cfg.DefaultEngine)
	}
	if cfg.SystemPrompt != "hi" {
		t.Fatalf("SystemPrompt = %q", cfg.SystemPrompt)
	}
	if cfg.PromptStrategy != "replace" {
		t.Fatalf("PromptStrategy = %q", cfg.PromptStrategy)
	}
}

func TestLoadPromptPreset(t *testing.T) {
	prompt, err := LoadPromptPreset("default")
	if err != nil {
		t.Fatalf("LoadPromptPreset error: %v", err)
	}
	if prompt == "" {
		t.Fatalf("expected prompt content")
	}
}
