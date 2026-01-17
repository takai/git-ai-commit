package config

import (
	"os"
	"os/exec"
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
	if cfg.ResolvedPrompt == "" {
		t.Fatalf("expected default resolved prompt")
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
	data := []byte("engine = 'codex'\nprompt = 'conventional'\n")
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
	if cfg.Prompt != "conventional" {
		t.Fatalf("Prompt = %q", cfg.Prompt)
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

func TestLoadRepoConfigOverrides(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	runGit(t, repo, "init")

	repoConfig := filepath.Join(repo, ".git-ai-commit.toml")
	if err := os.WriteFile(repoConfig, []byte("engine = 'codex'\n"), 0o644); err != nil {
		t.Fatalf("write repo config: %v", err)
	}

	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDir := filepath.Join(configHome, "git-ai-commit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("engine = 'gemini'\n"), 0o644); err != nil {
		t.Fatalf("write xdg config: %v", err)
	}

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if cfg.DefaultEngine != "codex" {
			t.Fatalf("DefaultEngine = %q", cfg.DefaultEngine)
		}
	})
}

func withDir(t *testing.T, dir string, fn func()) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	fn()
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v (%s)", args, err, output)
	}
}

func TestConfigMerging(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	runGit(t, repo, "init")

	// Repo config: only prompt
	repoConfig := filepath.Join(repo, ".git-ai-commit.toml")
	if err := os.WriteFile(repoConfig, []byte("prompt = 'conventional'\n"), 0o644); err != nil {
		t.Fatalf("write repo config: %v", err)
	}

	// User config: engine
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDir := filepath.Join(configHome, "git-ai-commit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("engine = 'claude'\n"), 0o644); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		// Engine from user config should be preserved
		if cfg.DefaultEngine != "claude" {
			t.Fatalf("DefaultEngine = %q, want 'claude'", cfg.DefaultEngine)
		}
		// Prompt from repo config should be applied
		if cfg.Prompt != "conventional" {
			t.Fatalf("Prompt = %q, want 'conventional'", cfg.Prompt)
		}
	})
}

func TestPromptExclusivityError(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDir := filepath.Join(configHome, "git-ai-commit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	// Setting both prompt and prompt_file in the same file should error
	data := []byte("prompt = 'conventional'\nprompt_file = 'custom.md'\n")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for setting both prompt and prompt_file")
	}
	if !contains(err.Error(), "cannot set both") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPromptFileLoading(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDir := filepath.Join(configHome, "git-ai-commit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	// Create a custom prompt file
	promptFile := filepath.Join(configDir, "custom.md")
	if err := os.WriteFile(promptFile, []byte("Custom prompt content"), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}

	// Create config pointing to the prompt file
	configPath := filepath.Join(configDir, "config.toml")
	data := []byte("prompt_file = 'custom.md'\n")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.ResolvedPrompt != "Custom prompt content" {
		t.Fatalf("ResolvedPrompt = %q, want 'Custom prompt content'", cfg.ResolvedPrompt)
	}
}

func TestPromptFileMerging(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	runGit(t, repo, "init")

	// Create a custom prompt file in the repo
	promptFile := filepath.Join(repo, "prompts", "commit.md")
	if err := os.MkdirAll(filepath.Dir(promptFile), 0o755); err != nil {
		t.Fatalf("mkdir prompts dir: %v", err)
	}
	if err := os.WriteFile(promptFile, []byte("Repo custom prompt"), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}

	// Repo config: prompt_file (should override user's prompt)
	repoConfig := filepath.Join(repo, ".git-ai-commit.toml")
	if err := os.WriteFile(repoConfig, []byte("prompt_file = 'prompts/commit.md'\n"), 0o644); err != nil {
		t.Fatalf("write repo config: %v", err)
	}

	// User config: prompt
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDir := filepath.Join(configHome, "git-ai-commit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("prompt = 'conventional'\n"), 0o644); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		// prompt_file from repo config should override prompt from user config
		if cfg.PromptFile != "prompts/commit.md" {
			t.Fatalf("PromptFile = %q, want 'prompts/commit.md'", cfg.PromptFile)
		}
		if cfg.Prompt != "" {
			t.Fatalf("Prompt = %q, want empty", cfg.Prompt)
		}
		if cfg.ResolvedPrompt != "Repo custom prompt" {
			t.Fatalf("ResolvedPrompt = %q, want 'Repo custom prompt'", cfg.ResolvedPrompt)
		}
	})
}

func TestValidateCLIPromptExclusivity(t *testing.T) {
	err := ValidateCLIPromptExclusivity("conventional", "custom.md")
	if err == nil {
		t.Fatal("expected error for setting both prompt and prompt_file")
	}
	if !contains(err.Error(), "CLI") {
		t.Fatalf("error should mention CLI: %v", err)
	}

	// Single values should be OK
	if err := ValidateCLIPromptExclusivity("conventional", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ValidateCLIPromptExclusivity("", "custom.md"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
