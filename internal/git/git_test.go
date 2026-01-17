package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStagedDiffEmpty(t *testing.T) {
	repo := setupRepo(t)
	withRepo(t, repo, func() {
		diff, err := StagedDiff()
		if err != nil {
			t.Fatalf("StagedDiff error: %v", err)
		}
		if diff != "" {
			t.Fatalf("expected empty diff, got %q", diff)
		}
	})
}

func TestLastCommitDiff(t *testing.T) {
	repo := setupRepo(t)
	withRepo(t, repo, func() {
		writeFile(t, repo, "file.txt", "hello")
		runGit(t, repo, "add", "file.txt")
		runGit(t, repo, "commit", "-m", "initial")

		diff, err := LastCommitDiff()
		if err != nil {
			t.Fatalf("LastCommitDiff error: %v", err)
		}
		if diff == "" {
			t.Fatalf("expected last commit diff")
		}
	})
}
func TestHasHeadCommitFalse(t *testing.T) {
	repo := setupRepo(t)
	withRepo(t, repo, func() {
		has, err := HasHeadCommit()
		if err != nil {
			t.Fatalf("HasHeadCommit error: %v", err)
		}
		if has {
			t.Fatalf("expected no HEAD commit")
		}
	})
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

func runGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := gitCmd(repo, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v (%s)", args, err, output)
	}
}
