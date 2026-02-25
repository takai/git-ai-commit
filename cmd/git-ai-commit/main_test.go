package main

import (
	"testing"
)

func TestParseArgs_ExcludeLong(t *testing.T) {
	opts, err := parseArgs([]string{"--exclude", "go.sum"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(opts.excludeFiles) != 1 || opts.excludeFiles[0] != "go.sum" {
		t.Errorf("expected excludeFiles=[go.sum], got %v", opts.excludeFiles)
	}
}

func TestParseArgs_ExcludeLongEquals(t *testing.T) {
	opts, err := parseArgs([]string{"--exclude=go.sum"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(opts.excludeFiles) != 1 || opts.excludeFiles[0] != "go.sum" {
		t.Errorf("expected excludeFiles=[go.sum], got %v", opts.excludeFiles)
	}
}

func TestParseArgs_ExcludeShort(t *testing.T) {
	opts, err := parseArgs([]string{"-x", "go.sum"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(opts.excludeFiles) != 1 || opts.excludeFiles[0] != "go.sum" {
		t.Errorf("expected excludeFiles=[go.sum], got %v", opts.excludeFiles)
	}
}

func TestParseArgs_ExcludeMultiple(t *testing.T) {
	opts, err := parseArgs([]string{"--exclude", "go.sum", "-x", "vendor/deps.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(opts.excludeFiles) != 2 {
		t.Fatalf("expected 2 excludeFiles, got %v", opts.excludeFiles)
	}
	if opts.excludeFiles[0] != "go.sum" || opts.excludeFiles[1] != "vendor/deps.go" {
		t.Errorf("expected [go.sum, vendor/deps.go], got %v", opts.excludeFiles)
	}
}

func TestParseArgs_ExcludeMissingValue(t *testing.T) {
	_, err := parseArgs([]string{"--exclude"})
	if err == nil {
		t.Fatal("expected error for missing value")
	}
}

func TestParseArgs_ExcludeShortMissingValue(t *testing.T) {
	_, err := parseArgs([]string{"-x"})
	if err == nil {
		t.Fatal("expected error for missing value")
	}
}

func TestParseArgs_ExcludeShortCluster(t *testing.T) {
	opts, err := parseArgs([]string{"-ax", "go.sum"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.addAll {
		t.Error("expected addAll to be true")
	}
	if len(opts.excludeFiles) != 1 || opts.excludeFiles[0] != "go.sum" {
		t.Errorf("expected excludeFiles=[go.sum], got %v", opts.excludeFiles)
	}
}

func TestParseArgs_EditLong(t *testing.T) {
	opts, err := parseArgs([]string{"--edit"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.edit {
		t.Error("expected edit to be true")
	}
}

func TestParseArgs_EditShort(t *testing.T) {
	opts, err := parseArgs([]string{"-e"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.edit {
		t.Error("expected edit to be true")
	}
}

func TestParseArgs_EditShortCluster(t *testing.T) {
	opts, err := parseArgs([]string{"-ae"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.addAll {
		t.Error("expected addAll to be true")
	}
	if !opts.edit {
		t.Error("expected edit to be true")
	}
}
