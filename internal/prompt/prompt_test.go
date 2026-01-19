package prompt

import (
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	got := Build("sys", "ctx", "diff")

	// Check that output contains expected sections
	if !strings.Contains(got, "=== INSTRUCTIONS ===") {
		t.Fatal("Build output should contain INSTRUCTIONS section")
	}
	if !strings.Contains(got, "sys") {
		t.Fatal("Build output should contain system prompt")
	}
	if !strings.Contains(got, "OUTPUT RULES:") {
		t.Fatal("Build output should contain OUTPUT RULES section")
	}
	if !strings.Contains(got, "=== CONTEXT ===") {
		t.Fatal("Build output should contain CONTEXT section when context provided")
	}
	if !strings.Contains(got, "ctx") {
		t.Fatal("Build output should contain context content")
	}
	if !strings.Contains(got, "=== GIT DIFF ===") {
		t.Fatal("Build output should contain GIT DIFF section")
	}
	if !strings.Contains(got, "diff") {
		t.Fatal("Build output should contain diff content")
	}
	if !strings.Contains(got, "=== OUTPUT ===") {
		t.Fatal("Build output should contain OUTPUT section")
	}
}

func TestBuildWithoutContext(t *testing.T) {
	got := Build("sys", "", "diff")

	// Check that output omits CONTEXT section when empty
	if strings.Contains(got, "=== CONTEXT ===") {
		t.Fatal("Build output should not contain CONTEXT section when context is empty")
	}
	if !strings.Contains(got, "=== INSTRUCTIONS ===") {
		t.Fatal("Build output should contain INSTRUCTIONS section")
	}
	if !strings.Contains(got, "=== GIT DIFF ===") {
		t.Fatal("Build output should contain GIT DIFF section")
	}
}

func TestBuildOutputRules(t *testing.T) {
	got := Build("sys", "ctx", "diff")

	// Verify output rules are included
	if !strings.Contains(got, "Exclude explanations, preambles") {
		t.Fatal("Build output should contain preamble suppression rule")
	}
	if !strings.Contains(got, "Begin with the commit subject line") {
		t.Fatal("Build output should contain direct start instruction")
	}
}
