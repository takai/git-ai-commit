package git

import (
	"bytes"
	"fmt"
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

func CommitWithMessage(message string, amend bool) error {
	args := []string{"commit", "-F", "-"}
	if amend {
		args = append(args, "--amend")
	}
	cmd := exec.Command("git", args...)
	cmd.Stdin = strings.NewReader(message)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}
