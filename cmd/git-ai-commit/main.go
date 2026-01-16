package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"git-ai-commit/internal/app"
)

type options struct {
	context        string
	contextFile    string
	systemPrompt   string
	promptStrategy string
	promptPreset   string
	engine         string
	amend          bool
	addAll         bool
	includeFiles   []string
	debugPrompt    bool
	debugCommand   bool
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		if errors.Is(err, errHelp) {
			printUsage(os.Stdout)
			return
		}
		fmt.Fprintln(os.Stderr, err)
		printUsage(os.Stderr)
		os.Exit(2)
	}
	if err := app.Run(
		opts.context,
		opts.contextFile,
		opts.systemPrompt,
		opts.promptStrategy,
		opts.promptPreset,
		opts.engine,
		opts.amend,
		opts.addAll,
		opts.includeFiles,
		opts.debugPrompt,
		opts.debugCommand,
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var errHelp = errors.New("help requested")

func parseArgs(args []string) (options, error) {
	var opts options
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			return opts, fmt.Errorf("unexpected arguments after --")
		}
		if strings.HasPrefix(arg, "--") {
			name, value, hasValue := strings.Cut(arg[2:], "=")
			switch name {
			case "help":
				return opts, errHelp
			case "context", "context-file", "system-prompt", "prompt-strategy", "prompt-preset", "engine", "include":
				if !hasValue {
					if i+1 >= len(args) {
						return opts, fmt.Errorf("missing value for --%s", name)
					}
					i++
					value = args[i]
				}
				if err := applyLongOption(&opts, name, value); err != nil {
					return opts, err
				}
			case "amend":
				opts.amend = true
			case "all":
				opts.addAll = true
			case "debug-prompt":
				opts.debugPrompt = true
			case "debug-command":
				opts.debugCommand = true
			default:
				return opts, fmt.Errorf("unknown option --%s", name)
			}
			continue
		}
		if strings.HasPrefix(arg, "-") && arg != "-" {
			if err := parseShortOptions(&opts, arg, args, &i); err != nil {
				return opts, err
			}
			continue
		}
		return opts, fmt.Errorf("unexpected argument %q", arg)
	}
	return opts, nil
}

func applyLongOption(opts *options, name, value string) error {
	switch name {
	case "context":
		opts.context = value
	case "context-file":
		opts.contextFile = value
	case "system-prompt":
		opts.systemPrompt = value
	case "prompt-strategy":
		opts.promptStrategy = value
	case "prompt-preset":
		opts.promptPreset = value
	case "engine":
		opts.engine = value
	case "include":
		if value == "" {
			return fmt.Errorf("missing value for --include")
		}
		opts.includeFiles = append(opts.includeFiles, value)
	default:
		return fmt.Errorf("unknown option --%s", name)
	}
	return nil
}

func parseShortOptions(opts *options, arg string, args []string, index *int) error {
	cluster := arg[1:]
	if cluster == "h" {
		return errHelp
	}
	for i := 0; i < len(cluster); i++ {
		switch cluster[i] {
		case 'a':
			opts.addAll = true
		case 'i':
			value := ""
			if i+1 < len(cluster) {
				if cluster[i+1] == '=' {
					value = cluster[i+2:]
				} else {
					value = cluster[i+1:]
				}
				i = len(cluster)
			} else {
				if *index+1 >= len(args) {
					return fmt.Errorf("missing value for -i")
				}
				*index++
				value = args[*index]
			}
			if value == "" {
				return fmt.Errorf("missing value for -i")
			}
			opts.includeFiles = append(opts.includeFiles, value)
		case 'h':
			return errHelp
		default:
			return fmt.Errorf("unknown option -%c", cluster[i])
		}
	}
	return nil
}

func printUsage(out *os.File) {
	fmt.Fprintln(out, "Usage: git-ai-commit [options]")
	fmt.Fprintln(out, "Generates a commit message from staged diff and commits safely.")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Options:")
	fmt.Fprintln(out, "  --context VALUE           Additional context for the commit message")
	fmt.Fprintln(out, "  --context-file VALUE      Path to a file containing additional context")
	fmt.Fprintln(out, "  --system-prompt VALUE     Override system prompt text")
	fmt.Fprintln(out, "  --prompt-strategy VALUE   Prompt override strategy: replace, prepend, append")
	fmt.Fprintln(out, "  --prompt-preset VALUE     Bundled prompt preset: default, conventional, gitmoji, karma")
	fmt.Fprintln(out, "  --engine VALUE            LLM engine name override")
	fmt.Fprintln(out, "  --amend                   Amend the previous commit")
	fmt.Fprintln(out, "  -a, --all                 Stage modified and deleted files before generating the message")
	fmt.Fprintln(out, "  -i, --include VALUE       Stage specific files before generating the message")
	fmt.Fprintln(out, "  --debug-prompt            Print the prompt before executing the engine")
	fmt.Fprintln(out, "  --debug-command           Print the engine command before execution")
	fmt.Fprintln(out, "  -h, --help                Show help")
}
