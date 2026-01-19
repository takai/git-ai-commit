package prompt

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed prompt.tmpl
var promptTemplateText string

var promptTemplate = template.Must(template.New("prompt").Parse(promptTemplateText))

type PromptData struct {
	SystemPrompt string
	Context      string
	Diff         string
}

func Build(systemPrompt, context, diff string) string {
	data := PromptData{
		SystemPrompt: strings.TrimSpace(systemPrompt),
		Context:      strings.TrimSpace(context),
		Diff:         diff,
	}
	var buf bytes.Buffer
	if err := promptTemplate.Execute(&buf, data); err != nil {
		// Fallback to simple concatenation on template error
		return systemPrompt + "\n\n" + context + "\n\n" + diff
	}
	return buf.String()
}
