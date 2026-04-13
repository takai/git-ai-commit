package app

import (
	"errors"
	"strings"
	"testing"

	"git-ai-commit/internal/config"
	"git-ai-commit/internal/engine"
	"git-ai-commit/internal/git"
)

func TestBuildEngineFailureErrorNonEngineError(t *testing.T) {
	plain := errors.New("some other error")
	result := buildEngineFailureError(plain, git.Result{}, nil)
	if result != plain {
		t.Fatalf("expected original error to be returned unchanged, got %v", result)
	}
}

func TestBuildEngineFailureErrorNoHint(t *testing.T) {
	inner := errors.New("exit status 1")
	engErr := &engine.EngineError{Err: inner, Stderr: "some stderr output"}
	result := buildEngineFailureError(engErr, git.Result{}, nil)
	msg := result.Error()

	if !strings.HasPrefix(msg, "engine command failed: exit status 1") {
		t.Fatalf("message should start with engine error, got: %q", msg)
	}
	if !strings.Contains(msg, "Full engine output saved to:") {
		t.Fatalf("message should contain log path, got: %q", msg)
	}
	if strings.Contains(msg, "Hint:") {
		t.Fatalf("message should not contain hint when no truncated/excluded files, got: %q", msg)
	}
}

func TestBuildEngineFailureErrorWithTruncatedFiles(t *testing.T) {
	inner := errors.New("exit status 1")
	engErr := &engine.EngineError{Err: inner, Stderr: ""}
	filterResult := git.Result{
		TruncatedFiles: []string{"large.txt"},
	}
	result := buildEngineFailureError(engErr, filterResult, nil)
	msg := result.Error()

	if !strings.Contains(msg, "Hint:") {
		t.Fatalf("message should contain hint for truncated files, got: %q", msg)
	}
	if !strings.Contains(msg, "--exclude large.txt") {
		t.Fatalf("message should contain --exclude large.txt, got: %q", msg)
	}
}

func TestBuildEngineFailureErrorWithExcludedFiles(t *testing.T) {
	inner := errors.New("exit status 1")
	engErr := &engine.EngineError{Err: inner, Stderr: ""}
	filterResult := git.Result{
		ExcludedFiles: []string{"go.sum", "package-lock.json"},
	}
	result := buildEngineFailureError(engErr, filterResult, nil)
	msg := result.Error()

	if !strings.Contains(msg, "--exclude go.sum") {
		t.Fatalf("message should contain --exclude go.sum, got: %q", msg)
	}
	if !strings.Contains(msg, "--exclude package-lock.json") {
		t.Fatalf("message should contain --exclude package-lock.json, got: %q", msg)
	}
}

func TestBuildEngineFailureErrorSkipsUserExcluded(t *testing.T) {
	inner := errors.New("exit status 1")
	engErr := &engine.EngineError{Err: inner, Stderr: ""}
	filterResult := git.Result{
		TruncatedFiles: []string{"large.txt"},
		ExcludedFiles:  []string{"already-excluded.txt"},
	}
	result := buildEngineFailureError(engErr, filterResult, []string{"already-excluded.txt"})
	msg := result.Error()

	if strings.Contains(msg, "--exclude already-excluded.txt") {
		t.Fatalf("message should not re-list user-excluded file, got: %q", msg)
	}
	if !strings.Contains(msg, "--exclude large.txt") {
		t.Fatalf("message should still contain truncated file hint, got: %q", msg)
	}
}

func TestBuildEngineFailureErrorEmptyStderr(t *testing.T) {
	inner := errors.New("exit status 1")
	engErr := &engine.EngineError{Err: inner, Stderr: ""}
	result := buildEngineFailureError(engErr, git.Result{}, nil)
	msg := result.Error()

	// Even with empty stderr, a log file should be created
	if !strings.Contains(msg, "Full engine output saved to:") {
		t.Fatalf("message should contain log path even for empty stderr, got: %q", msg)
	}
}

func TestBuildExcludeCandidates(t *testing.T) {
	filterResult := git.Result{
		TruncatedFiles: []string{"big.go"},
		ExcludedFiles:  []string{"go.sum", "user-excluded.txt"},
	}
	candidates := buildExcludeCandidates(filterResult, []string{"user-excluded.txt"})
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d: %v", len(candidates), candidates)
	}
	if candidates[0] != "big.go" {
		t.Fatalf("first candidate should be big.go, got %q", candidates[0])
	}
	if candidates[1] != "go.sum" {
		t.Fatalf("second candidate should be go.sum, got %q", candidates[1])
	}
}

func TestSelectEngineCodexDefault(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultEngine = "codex"
	cfg.Engines = map[string]config.EngineConfig{}

	_, command, err := selectEngine(cfg)
	if err != nil {
		t.Fatalf("selectEngine error: %v", err)
	}
	if command != "codex exec --model gpt-5.1-codex-mini" {
		t.Fatalf("command = %q", command)
	}
}

func TestSelectEngineClaudeDefault(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultEngine = "claude"
	cfg.Engines = map[string]config.EngineConfig{}

	_, command, err := selectEngine(cfg)
	if err != nil {
		t.Fatalf("selectEngine error: %v", err)
	}
	if command != "claude -p --model haiku --settings {\"attribution\":{\"commit\":\"\",\"pr\":\"\"}} --no-session-persistence" {
		t.Fatalf("command = %q", command)
	}
}

func TestSelectEngineCursorAgentDefault(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultEngine = "cursor-agent"
	cfg.Engines = map[string]config.EngineConfig{}

	_, command, err := selectEngine(cfg)
	if err != nil {
		t.Fatalf("selectEngine error: %v", err)
	}
	if command != "cursor-agent -p" {
		t.Fatalf("command = %q", command)
	}
}

func TestSelectEngineGeminiDefault(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultEngine = "gemini"
	cfg.Engines = map[string]config.EngineConfig{}

	_, command, err := selectEngine(cfg)
	if err != nil {
		t.Fatalf("selectEngine error: %v", err)
	}
	if command != "gemini -m gemini-2.5-flash -p {{prompt}}" {
		t.Fatalf("command = %q", command)
	}
}

func TestSelectEngineCustomArgs(t *testing.T) {
	cfg := config.Default()
	cfg.DefaultEngine = "claude"
	cfg.Engines = map[string]config.EngineConfig{
		"claude": {Args: []string{"-p", "--model", "sonnet"}},
	}

	_, command, err := selectEngine(cfg)
	if err != nil {
		t.Fatalf("selectEngine error: %v", err)
	}
	if command != "claude -p --model sonnet" {
		t.Fatalf("command = %q", command)
	}
}

func TestSanitizeMessageStripANSIEscapeSequences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "cursor back and erase",
			input: "Add uni\x1b[3D\x1b[Kntended behavior fix",
			want:  "Add unintended behavior fix",
		},
		{
			name:  "multiple ANSI sequences",
			input: "Fix \x1b[8D\x1b[K\x1b[4D\x1b[Kbug in parser",
			want:  "Fix bug in parser",
		},
		{
			name:  "SGR color codes",
			input: "\x1b[32mAdd feature\x1b[0m",
			want:  "Add feature",
		},
		{
			name:  "no ANSI sequences",
			input: "Clean commit message",
			want:  "Clean commit message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeMessage(tt.input)
			if got != tt.want {
				t.Fatalf("sanitizeMessage(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeMessageCodeFence(t *testing.T) {
	input := "```\nfeat: add thing\n```"
	got := sanitizeMessage(input)
	if got != "feat: add thing" {
		t.Fatalf("sanitizeMessage = %q", got)
	}
}

func TestSanitizeMessageInlineBackticks(t *testing.T) {
	input := "`feat: add thing`"
	got := sanitizeMessage(input)
	if got != "feat: add thing" {
		t.Fatalf("sanitizeMessage = %q", got)
	}
}
