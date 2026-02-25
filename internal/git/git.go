package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func StagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func LastCommitDiff() (string, error) {
	cmd := exec.Command("git", "show", "HEAD", "--format=", "--no-color")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git show failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func AddAll() error {
	cmd := exec.Command("git", "add", "-u")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add -u failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func AddFiles(files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files provided to add")
	}
	args := append([]string{"add", "--"}, files...)
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func WriteIndexTree() (string, error) {
	cmd := exec.Command("git", "write-tree")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git write-tree failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func ReadIndexTree(tree string) error {
	if strings.TrimSpace(tree) == "" {
		return fmt.Errorf("empty tree id")
	}
	cmd := exec.Command("git", "read-tree", tree)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git read-tree failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func HasHeadCommit() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if strings.Contains(msg, "Needed a single revision") || strings.Contains(msg, "unknown revision") {
			return false, nil
		}
		return false, fmt.Errorf("git rev-parse failed: %v: %s", err, msg)
	}
	return true, nil
}

func CommitWithMessage(message string, amend, edit bool) error {
	if edit {
		return commitWithEdit(message, amend)
	}
	args := []string{"commit", "-F", "-"}
	if amend {
		args = append(args, "--amend")
	}
	cmd := exec.Command("git", args...)
	cmd.Stdin = strings.NewReader(message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %v", err)
	}
	return nil
}

func commitWithEdit(message string, amend bool) error {
	f, err := os.CreateTemp("", "git-ai-commit-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(message); err != nil {
		f.Close()
		return fmt.Errorf("failed to write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	args := []string{"commit", "--edit", "-F", f.Name()}
	if amend {
		args = append(args, "--amend")
	}
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %v", err)
	}
	return nil
}
