package engine

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type CLI struct {
	Command string
	Args    []string
}

func (c CLI) Generate(prompt string) (string, error) {
	cmd := exec.Command(c.Command, c.Args...)
	cmd.Stdin = strings.NewReader(prompt)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("engine command failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
