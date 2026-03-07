package config

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// gitConfigScope holds ai-commit settings parsed from one git config scope.
type gitConfigScope struct {
	engine                 string
	prompt                 string
	promptFile             string
	maxFileLines           int
	maxFileLinesSet        bool
	excludePatterns        []string
	defaultExcludePatterns []string
}

// gitConfigScopes holds settings parsed from all git config scopes.
type gitConfigScopes struct {
	system   gitConfigScope
	global   gitConfigScope
	local    gitConfigScope
	worktree gitConfigScope
}

// readGitConfigScopes runs "git config --list --show-scope" and returns
// ai-commit settings organised by scope. If git is unavailable or we are not
// in a repository the returned struct is zero-valued and err is nil.
func readGitConfigScopes() (gitConfigScopes, error) {
	cmd := exec.Command("git", "config", "--list", "--show-scope")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		// Not in a git repo or git not available; no git config to read.
		return gitConfigScopes{}, nil
	}

	scopes := gitConfigScopes{}
	layerMap := map[string]*gitConfigScope{
		"system":   &scopes.system,
		"global":   &scopes.global,
		"local":    &scopes.local,
		"worktree": &scopes.worktree,
	}

	for line := range strings.SplitSeq(stdout.String(), "\n") {
		if line == "" {
			continue
		}
		before, after, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		scope := before
		rest := after

		lyr, ok := layerMap[scope]
		if !ok {
			// "command" scope and any unknown scopes are ignored.
			continue
		}

		before, after, ok0 := strings.Cut(rest, "=")
		if !ok0 {
			continue
		}
		// git config keys are case-insensitive; the list output is lowercase.
		key := strings.ToLower(before)
		value := after

		switch key {
		case "ai-commit.engine":
			lyr.engine = value
		case "ai-commit.prompt":
			lyr.prompt = value
		case "ai-commit.promptfile":
			lyr.promptFile = value
		case "ai-commit.maxfilelines":
			n, err := strconv.Atoi(value)
			if err != nil {
				lyr.maxFileLinesSet = true // mark that the key was present but invalid
				lyr.maxFileLines = -1      // sentinel for "invalid"
			} else {
				lyr.maxFileLines = n
				lyr.maxFileLinesSet = true
			}
		case "ai-commit.excludepatterns":
			lyr.excludePatterns = append(lyr.excludePatterns, value)
		case "ai-commit.defaultexcludepatterns":
			lyr.defaultExcludePatterns = append(lyr.defaultExcludePatterns, value)
		}
	}

	return scopes, nil
}

// applyGitConfigScope merges one git config scope into cfg.
//
// scopeLabel is the human-readable name used in error messages, e.g. "repo
// git config".
//
// promptFileBase is the directory used to resolve relative promptFile values.
// Pass "" to disallow relative paths (require absolute).
func applyGitConfigScope(cfg *Config, scope gitConfigScope, scopeLabel, promptFileBase string) error {
	// Validate maxFileLines
	if scope.maxFileLinesSet && scope.maxFileLines == -1 {
		return fmt.Errorf("invalid ai-commit.maxFileLines value in %s: not an integer", scopeLabel)
	}

	// Validate prompt exclusivity at this scope level.
	if strings.TrimSpace(scope.prompt) != "" && strings.TrimSpace(scope.promptFile) != "" {
		return fmt.Errorf("%s: cannot set both 'prompt' and 'promptFile'", scopeLabel)
	}

	if scope.engine != "" {
		cfg.DefaultEngine = scope.engine
	}
	if scope.prompt != "" {
		cfg.Prompt = scope.prompt
		cfg.PromptFile = ""
	}
	if scope.promptFile != "" {
		cfg.PromptFile = resolvePromptFilePath(scope.promptFile, promptFileBase)
		cfg.Prompt = ""
	}
	if scope.maxFileLinesSet && scope.maxFileLines >= 0 {
		cfg.Filter.MaxFileLines = scope.maxFileLines
	}
	if len(scope.defaultExcludePatterns) > 0 {
		cfg.Filter.DefaultExcludePatterns = scope.defaultExcludePatterns
	}
	if len(scope.excludePatterns) > 0 {
		cfg.Filter.ExcludePatterns = append(cfg.Filter.ExcludePatterns, scope.excludePatterns...)
	}

	return nil
}

// resolvePromptFilePath resolves a promptFile value relative to baseDir.
// Absolute paths are returned unchanged. If baseDir is empty, the path is
// returned as-is.
func resolvePromptFilePath(promptFile, baseDir string) string {
	if promptFile == "" || baseDir == "" || filepath.IsAbs(promptFile) {
		return promptFile
	}
	return filepath.Join(baseDir, promptFile)
}

// gitScopePromptPaths returns updated (promptFilePath, promptFileRepoRoot)
// values after a git config scope has been applied. If the scope set a
// promptFile, the resolved value is taken from cfg.PromptFile (already set by
// applyGitConfigScope). If the scope set a prompt preset, both values are
// cleared. Otherwise the current values are returned unchanged.
func gitScopePromptPaths(scope gitConfigScope, curPath, curRoot string, cfg Config) (string, string) {
	if scope.promptFile != "" {
		return cfg.PromptFile, ""
	}
	if scope.prompt != "" {
		return "", ""
	}
	return curPath, curRoot
}
