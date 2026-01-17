package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func gitCmd(repo string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	return cmd
}

func writeFile(t *testing.T, repo, name, content string) {
	t.Helper()
	path := filepath.Join(repo, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
