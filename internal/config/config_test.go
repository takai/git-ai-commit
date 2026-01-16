package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTOML(t *testing.T) {
	data := `engine = "codex"
	system_prompt = "hello"
	prompt_strategy = "prepend"

	[engines.codex]
	command = "codex"
	args = ["exec", "--model", "gpt"]
	`
	values, err := parseTOML(data)
	if err != nil {
		t.Fatalf("parseTOML error: %v", err)
	}
	if got := values["engine"].str; got != "codex" {
		t.Fatalf("engine = %q", got)
	}
	if got := values["system_prompt"].str; got != "hello" {
		t.Fatalf("system_prompt = %q", got)
	}
	if got := values["prompt_strategy"].str; got != "prepend" {
		t.Fatalf("prompt_strategy = %q", got)
	}
	if got := values["engines.codex.command"].str; got != "codex" {
		t.Fatalf("command = %q", got)
	}
	args := values["engines.codex.args"].list
	if len(args) != 3 || args[0] != "exec" || args[2] != "gpt" {
		t.Fatalf("args = %#v", args)
	}
}

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
