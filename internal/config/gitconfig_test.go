package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// setGitConfig sets a git config value in the given repo using git config --local.
func setGitConfig(t *testing.T, repo, key, value string) {
	t.Helper()
	runGit(t, repo, "config", "--local", key, value)
}

// addGitConfig appends a git config multi-value in the given repo.
func addGitConfig(t *testing.T, repo, key, value string) {
	t.Helper()
	runGit(t, repo, "config", "--add", key, value)
}

// initTestRepo creates a temp git repo and returns its path.
func initTestRepo(t *testing.T) string {
	t.Helper()
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	runGit(t, repo, "init")
	return repo
}

// isolateGitConfig points global/system config to empty temp files so they
// don't interfere with local-scope tests.
func isolateGitConfig(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	// Empty global config
	globalCfg := filepath.Join(tmp, "global.gitconfig")
	if err := os.WriteFile(globalCfg, []byte{}, 0o644); err != nil {
		t.Fatalf("write global config: %v", err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", globalCfg)
	// Disable system config
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

// TestGitConfigEngine verifies that ai-commit.engine in local git config sets
// the engine and overrides the user TOML.
func TestGitConfigEngine(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	setGitConfig(t, repo, "ai-commit.engine", "codex")

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if cfg.DefaultEngine != "codex" {
			t.Fatalf("DefaultEngine = %q, want 'codex'", cfg.DefaultEngine)
		}
	})
}

// TestGitConfigPrompt verifies that ai-commit.prompt in local git config sets
// the prompt preset.
func TestGitConfigPrompt(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	setGitConfig(t, repo, "ai-commit.prompt", "conventional")

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if cfg.Prompt != "conventional" {
			t.Fatalf("Prompt = %q, want 'conventional'", cfg.Prompt)
		}
		if cfg.ResolvedPrompt == "" {
			t.Fatal("ResolvedPrompt is empty")
		}
	})
}

// TestGitConfigPromptFile verifies that ai-commit.promptFile in local git
// config loads the file relative to the repo root.
func TestGitConfigPromptFile(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	promptContent := "My custom prompt from git config"
	promptFile := filepath.Join(repo, "my-prompt.md")
	if err := os.WriteFile(promptFile, []byte(promptContent), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}

	setGitConfig(t, repo, "ai-commit.promptFile", "my-prompt.md")

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if cfg.ResolvedPrompt != promptContent {
			t.Fatalf("ResolvedPrompt = %q, want %q", cfg.ResolvedPrompt, promptContent)
		}
		if cfg.Prompt != "" {
			t.Fatalf("Prompt = %q, want empty", cfg.Prompt)
		}
	})
}

// TestGitConfigPromptFileAbsolutePath verifies that an absolute promptFile
// path in local git config is accepted (no containment restriction).
func TestGitConfigPromptFileAbsolutePath(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Prompt file is outside the repo – this is allowed for git config sources.
	promptDir := t.TempDir()
	promptFile := filepath.Join(promptDir, "external-prompt.md")
	if err := os.WriteFile(promptFile, []byte("External prompt"), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}

	setGitConfig(t, repo, "ai-commit.promptFile", promptFile)

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if cfg.ResolvedPrompt != "External prompt" {
			t.Fatalf("ResolvedPrompt = %q, want 'External prompt'", cfg.ResolvedPrompt)
		}
	})
}

// TestGitConfigMaxFileLines verifies that ai-commit.maxFileLines is applied.
func TestGitConfigMaxFileLines(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	setGitConfig(t, repo, "ai-commit.maxFileLines", "200")

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if cfg.Filter.MaxFileLines != 200 {
			t.Fatalf("MaxFileLines = %d, want 200", cfg.Filter.MaxFileLines)
		}
	})
}

// TestGitConfigMaxFileLinesInvalid verifies that a non-integer maxFileLines
// returns a descriptive error.
func TestGitConfigMaxFileLinesInvalid(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	setGitConfig(t, repo, "ai-commit.maxFileLines", "notanumber")

	withDir(t, repo, func() {
		_, err := Load()
		if err == nil {
			t.Fatal("expected error for invalid maxFileLines")
		}
		if !contains(err.Error(), "maxFileLines") {
			t.Fatalf("error should mention maxFileLines: %v", err)
		}
		if !contains(err.Error(), "not an integer") {
			t.Fatalf("error should mention 'not an integer': %v", err)
		}
	})
}

// TestGitConfigExcludePatterns verifies that multi-value excludePatterns are
// collected and appended.
func TestGitConfigExcludePatterns(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	addGitConfig(t, repo, "ai-commit.excludePatterns", "*.lock")
	addGitConfig(t, repo, "ai-commit.excludePatterns", "vendor/**")

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if !slices.Contains(cfg.Filter.ExcludePatterns, "*.lock") {
			t.Fatalf("ExcludePatterns missing '*.lock': %v", cfg.Filter.ExcludePatterns)
		}
		if !slices.Contains(cfg.Filter.ExcludePatterns, "vendor/**") {
			t.Fatalf("ExcludePatterns missing 'vendor/**': %v", cfg.Filter.ExcludePatterns)
		}
	})
}

// TestGitConfigDefaultExcludePatterns verifies that defaultExcludePatterns in
// local git config overrides the default list.
func TestGitConfigDefaultExcludePatterns(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	addGitConfig(t, repo, "ai-commit.defaultExcludePatterns", "*.lock")

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if len(cfg.Filter.DefaultExcludePatterns) != 1 || cfg.Filter.DefaultExcludePatterns[0] != "*.lock" {
			t.Fatalf("DefaultExcludePatterns = %v, want ['*.lock']", cfg.Filter.DefaultExcludePatterns)
		}
	})
}

// TestGitConfigPromptExclusivity verifies that setting both prompt and
// promptFile at the same scope returns an error.
func TestGitConfigPromptExclusivity(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	setGitConfig(t, repo, "ai-commit.prompt", "conventional")
	setGitConfig(t, repo, "ai-commit.promptFile", "my-prompt.md")

	withDir(t, repo, func() {
		_, err := Load()
		if err == nil {
			t.Fatal("expected error for setting both prompt and promptFile")
		}
		if !contains(err.Error(), "cannot set both") {
			t.Fatalf("error should mention 'cannot set both': %v", err)
		}
	})
}

// TestGitConfigLocalOverridesRepoToml verifies that local git config takes
// priority over the repo TOML.
func TestGitConfigLocalOverridesRepoToml(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Repo TOML sets engine to "gemini"
	repoConfig := filepath.Join(repo, ".git-ai-commit.toml")
	if err := os.WriteFile(repoConfig, []byte("engine = 'gemini'\n"), 0o644); err != nil {
		t.Fatalf("write repo config: %v", err)
	}
	trustRepoConfig(t, repo, repoConfig)

	// Local git config overrides to "codex"
	setGitConfig(t, repo, "ai-commit.engine", "codex")

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if cfg.DefaultEngine != "codex" {
			t.Fatalf("DefaultEngine = %q, want 'codex' (local git config should override repo TOML)", cfg.DefaultEngine)
		}
	})
}

// TestGitConfigLocalOverridesUserToml verifies that local git config takes
// priority over the user TOML.
func TestGitConfigLocalOverridesUserToml(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)

	// User TOML sets engine to "gemini"
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDir := filepath.Join(configHome, "git-ai-commit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte("engine = 'gemini'\n"), 0o644); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	// Local git config overrides to "codex"
	setGitConfig(t, repo, "ai-commit.engine", "codex")

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if cfg.DefaultEngine != "codex" {
			t.Fatalf("DefaultEngine = %q, want 'codex' (local git config should override user TOML)", cfg.DefaultEngine)
		}
	})
}

// TestGitConfigGlobalOverriddenByUserToml verifies that the user TOML takes
// priority over global git config.
func TestGitConfigGlobalOverriddenByUserToml(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	// Global git config sets engine to "gemini"
	globalCfg := filepath.Join(tmp, "global.gitconfig")
	if err := os.WriteFile(globalCfg, []byte("[ai-commit]\n\tengine = gemini\n"), 0o644); err != nil {
		t.Fatalf("write global config: %v", err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", globalCfg)

	// User TOML sets engine to "codex"
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDir := filepath.Join(configHome, "git-ai-commit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte("engine = 'codex'\n"), 0o644); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	repo := initTestRepo(t)
	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if cfg.DefaultEngine != "codex" {
			t.Fatalf("DefaultEngine = %q, want 'codex' (user TOML should override global git config)", cfg.DefaultEngine)
		}
	})
}

// TestGitConfigGlobalPromptFile verifies that a promptFile in global git
// config is resolved relative to $HOME.
func TestGitConfigGlobalPromptFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Create prompt file in a "home" directory
	homeDir := filepath.Join(tmp, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	promptFile := filepath.Join(homeDir, "my-prompt.md")
	if err := os.WriteFile(promptFile, []byte("Global home prompt"), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}
	t.Setenv("HOME", homeDir)

	// Global git config points to relative promptFile (resolved from $HOME)
	globalCfg := filepath.Join(tmp, "global.gitconfig")
	if err := os.WriteFile(globalCfg, []byte("[ai-commit]\n\tpromptFile = my-prompt.md\n"), 0o644); err != nil {
		t.Fatalf("write global config: %v", err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", globalCfg)

	repo := initTestRepo(t)
	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if cfg.ResolvedPrompt != "Global home prompt" {
			t.Fatalf("ResolvedPrompt = %q, want 'Global home prompt'", cfg.ResolvedPrompt)
		}
	})
}

// TestGitConfigExcludePatternsAppendedAcrossLayers verifies that
// excludePatterns from multiple layers are all collected.
func TestGitConfigExcludePatternsAppendedAcrossLayers(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	// Global adds one pattern
	globalCfg := filepath.Join(tmp, "global.gitconfig")
	if err := os.WriteFile(globalCfg, []byte("[ai-commit]\n\texcludePatterns = global-pattern\n"), 0o644); err != nil {
		t.Fatalf("write global config: %v", err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", globalCfg)

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	repo := initTestRepo(t)
	// Local adds another pattern
	addGitConfig(t, repo, "ai-commit.excludePatterns", "local-pattern")

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if !slices.Contains(cfg.Filter.ExcludePatterns, "global-pattern") {
			t.Fatalf("ExcludePatterns missing 'global-pattern': %v", cfg.Filter.ExcludePatterns)
		}
		if !slices.Contains(cfg.Filter.ExcludePatterns, "local-pattern") {
			t.Fatalf("ExcludePatterns missing 'local-pattern': %v", cfg.Filter.ExcludePatterns)
		}
	})
}

// TestGitConfigDefaultExcludePatternsHigherScopeWins verifies that a
// higher-priority scope's defaultExcludePatterns overrides a lower one.
func TestGitConfigDefaultExcludePatternsHigherScopeWins(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	// Global sets default patterns
	globalCfg := filepath.Join(tmp, "global.gitconfig")
	if err := os.WriteFile(globalCfg, []byte("[ai-commit]\n\tdefaultExcludePatterns = global-default\n"), 0o644); err != nil {
		t.Fatalf("write global config: %v", err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", globalCfg)

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	repo := initTestRepo(t)
	// Local overrides default patterns
	addGitConfig(t, repo, "ai-commit.defaultExcludePatterns", "local-default")

	withDir(t, repo, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		// Only local default should be present (overrides global)
		if len(cfg.Filter.DefaultExcludePatterns) != 1 || cfg.Filter.DefaultExcludePatterns[0] != "local-default" {
			t.Fatalf("DefaultExcludePatterns = %v, want ['local-default']", cfg.Filter.DefaultExcludePatterns)
		}
	})
}

// TestGitConfigPromptFileNoRepoToml verifies that a relative promptFile in
// local git config resolves from the repo root even when .git-ai-commit.toml
// is absent (repoConfigPath previously returned an empty root in that case).
func TestGitConfigPromptFileNoRepoToml(t *testing.T) {
	repo := initTestRepo(t)
	isolateGitConfig(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// No .git-ai-commit.toml in this repo.
	promptContent := "Prompt loaded without repo TOML"
	if err := os.WriteFile(filepath.Join(repo, "prompt.md"), []byte(promptContent), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}

	setGitConfig(t, repo, "ai-commit.engine", "fake")
	setGitConfig(t, repo, "ai-commit.promptFile", "prompt.md")

	// Run from a subdirectory to confirm CWD is not used as the base.
	subdir := filepath.Join(repo, "sub")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}

	withDir(t, subdir, func() {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if cfg.ResolvedPrompt != promptContent {
			t.Fatalf("ResolvedPrompt = %q, want %q", cfg.ResolvedPrompt, promptContent)
		}
	})
}
