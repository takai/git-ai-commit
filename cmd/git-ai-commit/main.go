package main

import (
	"flag"
	"fmt"
	"os"

	"git-ai-commit/internal/app"
)

type options struct {
	context        string
	contextFile    string
	systemPrompt   string
	promptStrategy string
	engine         string
	amend          bool
}

func main() {
	opts := parseFlags()
	if err := app.Run(opts.context, opts.contextFile, opts.systemPrompt, opts.promptStrategy, opts.engine, opts.amend); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseFlags() options {
	var opts options
	flag.StringVar(&opts.context, "context", "", "Additional context for the commit message")
	flag.StringVar(&opts.contextFile, "context-file", "", "Path to a file containing additional context")
	flag.StringVar(&opts.systemPrompt, "system-prompt", "", "Override system prompt text")
	flag.StringVar(&opts.promptStrategy, "prompt-strategy", "", "Prompt override strategy: replace, prepend, append")
	flag.StringVar(&opts.engine, "engine", "", "LLM engine name override")
	flag.BoolVar(&opts.amend, "amend", false, "Amend the previous commit")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: git-ai-commit [options]")
		fmt.Fprintln(os.Stderr, "Generates a commit message from staged diff and commits safely.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()
	return opts
}
