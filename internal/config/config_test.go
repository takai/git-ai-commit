package config

import "testing"

func TestParseTOML(t *testing.T) {
	data := `engine = "codex"
	system_prompt = "hello"
	prompt_strategy = "prepend"

	[engines.codex]
	command = "codex"
	args = ["exec", "--model", "gpt"]
	`
	values, err := parseTOML(data)
	if err != nil {
		t.Fatalf("parseTOML error: %v", err)
	}
	if got := values["engine"].str; got != "codex" {
		t.Fatalf("engine = %q", got)
	}
	if got := values["system_prompt"].str; got != "hello" {
		t.Fatalf("system_prompt = %q", got)
	}
	if got := values["prompt_strategy"].str; got != "prepend" {
		t.Fatalf("prompt_strategy = %q", got)
	}
	if got := values["engines.codex.command"].str; got != "codex" {
		t.Fatalf("command = %q", got)
	}
	args := values["engines.codex.args"].list
	if len(args) != 3 || args[0] != "exec" || args[2] != "gpt" {
		t.Fatalf("args = %#v", args)
	}
}
