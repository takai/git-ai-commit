package engine

import (
	"errors"
	"strings"
	"testing"
)

func TestEngineErrorMessage(t *testing.T) {
	inner := errors.New("exit status 1")
	e := &EngineError{Err: inner, Stderr: "some error output"}
	got := e.Error()
	if got != "engine command failed: exit status 1" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestEngineErrorMessageEmptyStderr(t *testing.T) {
	inner := errors.New("exit status 1")
	e := &EngineError{Err: inner, Stderr: ""}
	got := e.Error()
	if got != "engine command failed: exit status 1" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestEngineErrorUnwrap(t *testing.T) {
	inner := errors.New("exit status 1")
	e := &EngineError{Err: inner, Stderr: ""}
	if !errors.Is(e, inner) {
		t.Fatal("errors.Is should match the wrapped error")
	}
}

func TestEngineErrorDoesNotIncludeStderr(t *testing.T) {
	inner := errors.New("exit status 1")
	e := &EngineError{Err: inner, Stderr: "secret output that should not appear"}
	if strings.Contains(e.Error(), "secret output") {
		t.Fatal("Error() must not include stderr content")
	}
}

func TestCLIGenerateReturnsEngineErrorOnFailure(t *testing.T) {
	cli := CLI{Command: "/bin/sh", Args: []string{"-c", "echo 'err output' >&2; exit 1"}}
	_, err := cli.Generate("ignored")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var engineErr *EngineError
	if !errors.As(err, &engineErr) {
		t.Fatalf("expected *EngineError, got %T: %v", err, err)
	}
	if !strings.Contains(engineErr.Stderr, "err output") {
		t.Fatalf("Stderr = %q, want it to contain 'err output'", engineErr.Stderr)
	}
}

func TestCLIGenerateUsesStdin(t *testing.T) {
	cli := CLI{Command: "/bin/cat", Args: nil}
	out, err := cli.Generate("hello")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if out != "hello" {
		t.Fatalf("output = %q", out)
	}
}

func TestCLIGenerateUsesPromptArg(t *testing.T) {
	cli := CLI{Command: "/bin/echo", Args: []string{"{{prompt}}"}}
	out, err := cli.Generate("hello")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if out != "hello\n" {
		t.Fatalf("output = %q", out)
	}
}

func TestCLIGenerateUnsetsClaudeCodeEnv(t *testing.T) {
	t.Setenv("CLAUDECODE", "1")
	cli := CLI{
		Command: "/bin/sh",
		Args:    []string{"-c", "printf %s \"$CLAUDECODE\""},
	}
	out, err := cli.Generate("ignored")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if out != "" {
		t.Fatalf("output = %q", out)
	}
}
