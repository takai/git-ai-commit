package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// EngineError is returned by Generate when the engine subprocess exits with a
// non-zero status. Stderr holds the full standard error output of the engine
// process, which may be empty.
type EngineError struct {
	Err    error  // the underlying exec error (e.g. *exec.ExitError)
	Stderr string // full stderr content, possibly empty
}

func (e *EngineError) Error() string {
	return fmt.Sprintf("engine command failed: %v", e.Err)
}

func (e *EngineError) Unwrap() error { return e.Err }

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
	cmd.Env = filteredEnv(os.Environ(), "CLAUDECODE")
	if !usePromptArg {
		cmd.Stdin = strings.NewReader(prompt)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", &EngineError{Err: err, Stderr: stderr.String()}
	}
	return stdout.String(), nil
}

func filteredEnv(env []string, names ...string) []string {
	if len(names) == 0 {
		return env
	}
	blocked := make(map[string]struct{}, len(names))
	for _, name := range names {
		blocked[name] = struct{}{}
	}
	out := make([]string, 0, len(env))
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if _, skip := blocked[key]; skip {
			continue
		}
		out = append(out, entry)
	}
	return out
}
