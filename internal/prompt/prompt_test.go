package prompt

import "testing"

func TestBuild(t *testing.T) {
	got := Build("sys", "ctx", "diff")
	expected := "System Prompt:\nsys\n\nContext:\nctx\n\nGit Diff:\ndiff"
	if got != expected {
		t.Fatalf("Build = %q", got)
	}
}
