package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingConfig(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("PATH", "")
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

func TestAutodetectEngineOrder(t *testing.T) {
	configHome := t.TempDir()
	binDir := filepath.Join(configHome, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	makeExecutable(t, binDir, "codex")
	makeExecutable(t, binDir, "gemini")
	makeExecutable(t, binDir, "claude")

	t.Setenv("PATH", binDir)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.DefaultEngine != "claude" {
		t.Fatalf("DefaultEngine = %q", cfg.DefaultEngine)
	}
}

func makeExecutable(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}
}
