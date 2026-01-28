package app

import (
	"fmt"
	"os"
	"strings"

	"git-ai-commit/internal/config"
	"git-ai-commit/internal/engine"
	"git-ai-commit/internal/git"
	"git-ai-commit/internal/prompt"
)

func Run(context, contextFile, promptName, promptFile, engineName string, amend, addAll bool, includeFiles []string, debugPrompt, debugCommand bool) (err error) {
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

	diff, err := commitDiff(amend, cfg)
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
		return err
	}
	message := sanitizeMessage(output)
	if message == "" {
		return fmt.Errorf("empty commit message from engine")
	}

	if err := git.CommitWithMessage(message, amend); err != nil {
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

func commitDiff(amend bool, cfg config.Config) (string, error) {
	var diff string
	var err error
	if amend {
		diff, err = git.LastCommitDiff()
	} else {
		diff, err = git.StagedDiff()
	}
	if err != nil {
		return "", err
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
	}
	result := git.Filter(diff, opts)

	if result.Truncated || len(result.ExcludedFiles) > 0 {
		return result.Diff + formatFilterNotice(result), nil
	}
	return result.Diff, nil
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
