package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"git-ai-commit/internal/config"
	"git-ai-commit/internal/engine"
	"git-ai-commit/internal/git"
	"git-ai-commit/internal/prompt"
)

func Run(context, contextFile, promptName, promptFile, engineName string, amend, addAll, edit, showDiff bool, includeFiles, excludeFiles []string, debugPrompt, debugCommand bool) (err error) {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	contextText, err := loadContext(context, contextFile)
	if err != nil {
		return err
	}

	var restoreIndex func()
	if addAll || len(includeFiles) > 0 {
		restoreIndex, err = stageChanges(addAll, includeFiles)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil && restoreIndex != nil {
				restoreIndex()
			}
		}()
	}

	if amend {
		hasHead, err := git.HasHeadCommit()
		if err != nil {
			return err
		}
		if !hasHead {
			return fmt.Errorf("cannot amend without an existing commit")
		}
	}

	diff, filterResult, err := commitDiff(amend, cfg, excludeFiles)
	if err != nil {
		return err
	}
	if strings.TrimSpace(diff) == "" {
		if amend {
			return fmt.Errorf("no changes found in the last commit")
		}
		return fmt.Errorf("no staged changes to commit")
	}

	if engineName != "" {
		cfg.DefaultEngine = engineName
	}

	// Apply CLI prompt overrides
	if err := config.ApplyCLIPrompt(&cfg, promptName, promptFile); err != nil {
		return err
	}

	promptText := prompt.Build(cfg.ResolvedPrompt, contextText, diff)
	eng, commandLine, err := selectEngine(cfg)
	if err != nil {
		return err
	}
	if debugCommand {
		fmt.Fprintf(os.Stderr, "engine command: %s\n", commandLine)
	}
	if debugPrompt {
		fmt.Fprintln(os.Stderr, "prompt:")
		fmt.Fprintln(os.Stderr, promptText)
	}

	output, err := eng.Generate(promptText)
	if err != nil {
		return buildEngineFailureError(err, filterResult, excludeFiles)
	}
	message := sanitizeMessage(output)
	if message == "" {
		return fmt.Errorf("empty commit message from engine")
	}

	if showDiff {
		edit = true
	}
	if err := git.CommitWithMessage(message, amend, edit, showDiff); err != nil {
		return err
	}
	return nil
}

func loadContext(context, contextFile string) (string, error) {
	if contextFile == "" {
		return strings.TrimSpace(context), nil
	}
	data, err := os.ReadFile(contextFile)
	if err != nil {
		return "", fmt.Errorf("read context file: %w", err)
	}
	parts := []string{strings.TrimSpace(string(data)), strings.TrimSpace(context)}
	combined := strings.TrimSpace(strings.Join(parts, "\n"))
	return combined, nil
}

func selectEngine(cfg config.Config) (engine.Engine, string, error) {
	name := strings.TrimSpace(cfg.DefaultEngine)
	if name == "" {
		return nil, "", fmt.Errorf("no engine configured")
	}
	if spec, ok := cfg.Engines[name]; ok {
		return engine.CLI{Command: name, Args: spec.Args}, strings.Join(append([]string{name}, spec.Args...), " "), nil
	}
	if args, ok := config.DefaultEngineArgs[name]; ok {
		return engine.CLI{Command: name, Args: args}, strings.Join(append([]string{name}, args...), " "), nil
	}
	return engine.CLI{Command: name, Args: nil}, name, nil
}

func sanitizeMessage(message string) string {
	clean := strings.TrimSpace(message)
	if clean == "" {
		return ""
	}
	lines := strings.Split(clean, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			continue
		}
		filtered = append(filtered, line)
	}
	clean = strings.TrimSpace(strings.Join(filtered, "\n"))
	if len(clean) >= 2 && strings.HasPrefix(clean, "`") && strings.HasSuffix(clean, "`") {
		clean = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(clean, "`"), "`"))
	}
	return clean
}

func commitDiff(amend bool, cfg config.Config, excludeFiles []string) (string, git.Result, error) {
	var diff string
	var err error
	if amend {
		diff, err = git.LastCommitDiff()
	} else {
		diff, err = git.StagedDiff()
	}
	if err != nil {
		return "", git.Result{}, err
	}

	// Determine exclude patterns
	patterns := git.DefaultExcludePatterns()
	if len(cfg.Filter.DefaultExcludePatterns) > 0 {
		patterns = cfg.Filter.DefaultExcludePatterns
	}
	patterns = append(patterns, cfg.Filter.ExcludePatterns...)

	// Determine max file lines
	maxLines := cfg.Filter.MaxFileLines
	if maxLines == 0 {
		maxLines = git.DefaultMaxFileLines
	}

	opts := git.Options{
		MaxFileLines:    maxLines,
		ExcludePatterns: patterns,
		ExcludeFiles:    excludeFiles,
	}
	result := git.Filter(diff, opts)

	if result.Truncated || len(result.ExcludedFiles) > 0 {
		return result.Diff + formatFilterNotice(result), result, nil
	}
	return result.Diff, result, nil
}

func formatFilterNotice(result git.Result) string {
	var parts []string
	if len(result.ExcludedFiles) > 0 {
		parts = append(parts, fmt.Sprintf("Excluded files: %s", strings.Join(result.ExcludedFiles, ", ")))
	}
	if len(result.TruncatedFiles) > 0 {
		parts = append(parts, fmt.Sprintf("Truncated files: %s", strings.Join(result.TruncatedFiles, ", ")))
	}
	if len(parts) == 0 {
		return ""
	}
	return "\n\n[Filter notice: " + strings.Join(parts, "; ") + "]"
}

// buildEngineFailureError converts an engine error into an actionable user
// message. If err is an *engine.EngineError, it saves the full stderr to a
// temp log file and appends an --exclude hint when the filter result contains
// truncated or pattern-excluded files. Non-EngineError values are returned
// unchanged.
func buildEngineFailureError(err error, filterResult git.Result, userExcluded []string) error {
	var engineErr *engine.EngineError
	if !errors.As(err, &engineErr) {
		return err
	}

	var msg strings.Builder
	msg.WriteString(engineErr.Error())

	logPath := writeTempLog(engineErr.Stderr)
	if logPath != "" {
		msg.WriteString("\nFull engine output saved to: ")
		msg.WriteString(logPath)
	}

	candidates := buildExcludeCandidates(filterResult, userExcluded)
	if len(candidates) > 0 {
		msg.WriteString("\nHint: the following files were truncated or excluded and may have caused\na context window overflow. Re-run with --exclude to skip them:")
		const binary = "git ai-commit"
		padding := strings.Repeat(" ", 2+len(binary)+1)
		for i, f := range candidates {
			if i == 0 {
				msg.WriteString("\n  ")
				msg.WriteString(binary)
				msg.WriteString(" --exclude ")
				msg.WriteString(f)
			} else {
				msg.WriteString(" \\\n")
				msg.WriteString(padding)
				msg.WriteString("--exclude ")
				msg.WriteString(f)
			}
		}
	}

	return errors.New(msg.String())
}

// writeTempLog writes content to a new temp file and returns its path.
// Returns empty string if the file cannot be created or written.
func writeTempLog(stderr string) string {
	content := stderr
	if strings.TrimSpace(content) == "" {
		content = "(empty)\n"
	}
	f, err := os.CreateTemp("", "git-ai-commit-stderr-*.log")
	if err != nil {
		return ""
	}
	defer f.Close()
	_, _ = f.WriteString(content)
	return f.Name()
}

// buildExcludeCandidates returns the list of files to suggest in the
// --exclude hint: truncated files plus pattern-excluded files that the user
// has not already explicitly excluded on the current invocation.
func buildExcludeCandidates(filterResult git.Result, userExcluded []string) []string {
	excluded := make(map[string]bool, len(userExcluded))
	for _, f := range userExcluded {
		excluded[f] = true
	}
	candidates := make([]string, 0, len(filterResult.TruncatedFiles)+len(filterResult.ExcludedFiles))
	candidates = append(candidates, filterResult.TruncatedFiles...)
	for _, f := range filterResult.ExcludedFiles {
		if !excluded[f] {
			candidates = append(candidates, f)
		}
	}
	return candidates
}

func stageChanges(addAll bool, includeFiles []string) (func(), error) {
	tree, err := git.WriteIndexTree()
	if err != nil {
		return nil, err
	}
	restore := func() {
		_ = git.ReadIndexTree(tree)
	}
	if addAll {
		if err := git.AddAll(); err != nil {
			restore()
			return nil, err
		}
		return restore, nil
	}
	if len(includeFiles) == 0 {
		return restore, nil
	}
	if err := git.AddFiles(includeFiles); err != nil {
		restore()
		return nil, err
	}
	return restore, nil
}
