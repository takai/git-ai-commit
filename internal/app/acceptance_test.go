package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCreatesCommit(t *testing.T) {
	repo := setupRepo(t)
	configHome := t.TempDir()
	writeConfig(t, configHome)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	withRepo(t, repo, func() {
		writeFile(t, repo, "file.txt", "hello")
		runGit(t, repo, "add", "file.txt")

		if err := Run("", "", "", "", "", "", false, false, nil); err != nil {
			t.Fatalf("Run error: %v", err)
		}

		out := runGitOutput(t, repo, "log", "-1", "--pretty=%s")
		if out != "test commit" {
			t.Fatalf("commit message = %q", out)
		}
	})
}

func writeConfig(t *testing.T, configHome string) {
	t.Helper()
	configDir := filepath.Join(configHome, "git-ai-commit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	data := []byte("engine = 'codex'\n[engines.codex]\ncommand = '/bin/sh'\nargs = ['-c', 'cat >/dev/null; echo test commit']\n")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func setupRepo(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.name", "Test User")
	runGit(t, repo, "config", "user.email", "test@example.com")
	return repo
}

func withRepo(t *testing.T, repo string, fn func()) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	fn()
}

func writeFile(t *testing.T, repo, name, content string) {
	t.Helper()
	path := filepath.Join(repo, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func runGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v (%s)", args, err, output)
	}
}

func runGitOutput(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v (%s)", args, err, output)
	}
	return strings.TrimSpace(string(output))
}
