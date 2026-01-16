package git

import "os/exec"

func gitCmd(repo string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	return cmd
}
