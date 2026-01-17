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
	args := make([]string, len(c.Args))
	copy(args, c.Args)
	usePromptArg := false
	for i, arg := range args {
		if strings.Contains(arg, "{{prompt}}") {
			args[i] = strings.ReplaceAll(arg, "{{prompt}}", prompt)
			usePromptArg = true
		}
	}
	cmd := exec.Command(c.Command, args...)
	if !usePromptArg {
		cmd.Stdin = strings.NewReader(prompt)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("engine command failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
