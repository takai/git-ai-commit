package engine

import "testing"

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
