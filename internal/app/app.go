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

func Run(context, contextFile, systemPrompt, promptStrategy, promptPreset, engineName string, amend, addAll bool, includeFiles []string, debugPrompt, debugCommand bool) (err error) {
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

	diff, err := commitDiff(amend)
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

	if promptPreset != "" {
		prompt, err := config.LoadPromptPreset(promptPreset)
		if err != nil {
			return err
		}
		cfg.SystemPrompt = prompt
	}

	if systemPrompt != "" {
		strategy := cfg.PromptStrategy
		if promptStrategy != "" {
			strategy = promptStrategy
		}
		combined, err := applyPromptStrategy(cfg.SystemPrompt, systemPrompt, strategy)
		if err != nil {
			return err
		}
		cfg.SystemPrompt = combined
	} else if promptStrategy != "" {
		return fmt.Errorf("prompt strategy set without a system prompt override")
	}

	promptText := prompt.Build(cfg.SystemPrompt, contextText, diff)
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

func applyPromptStrategy(base, override, strategy string) (string, error) {
	switch strings.ToLower(strategy) {
	case "", "append":
		if base == "" {
			return override, nil
		}
		return base + "\n" + override, nil
	case "prepend":
		if base == "" {
			return override, nil
		}
		return override + "\n" + base, nil
	case "replace":
		return override, nil
	default:
		return "", fmt.Errorf("unknown prompt strategy: %s", strategy)
	}
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

func commitDiff(amend bool) (string, error) {
	if amend {
		return git.LastCommitDiff()
	}
	return git.StagedDiff()
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
