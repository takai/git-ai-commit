package prompt

import "strings"

func Build(systemPrompt, context, diff string) string {
	var b strings.Builder
	if systemPrompt != "" {
		b.WriteString("System Prompt:\n")
		b.WriteString(systemPrompt)
		b.WriteString("\n\n")
	}
	if context != "" {
		b.WriteString("Context:\n")
		b.WriteString(context)
		b.WriteString("\n\n")
	}
	b.WriteString("Git Diff:\n")
	b.WriteString(diff)
	return b.String()
}
