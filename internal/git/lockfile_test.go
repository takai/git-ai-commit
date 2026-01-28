package git

import (
	"strconv"
	"strings"
	"testing"
)

func TestIsLockFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		// Lock files that should be detected
		{"uv.lock", "uv.lock", true},
		{"poetry.lock", "poetry.lock", true},
		{"package-lock.json", "package-lock.json", true},
		{"yarn.lock", "yarn.lock", true},
		{"pnpm-lock.yaml", "pnpm-lock.yaml", true},
		{"Gemfile.lock", "Gemfile.lock", true},
		{"Cargo.lock", "Cargo.lock", true},
		{"go.sum", "go.sum", true},
		{"composer.lock", "composer.lock", true},

		// Lock files in subdirectories
		{"nested uv.lock", "packages/uv.lock", true},
		{"nested package-lock.json", "frontend/package-lock.json", true},

		// Non-lock files
		{"regular go file", "main.go", false},
		{"regular json", "config.json", false},
		{"go.mod", "go.mod", false},
		{"package.json", "package.json", false},
		{"Gemfile", "Gemfile", false},
		{"Cargo.toml", "Cargo.toml", false},
		{"pyproject.toml", "pyproject.toml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLockFile(tt.filename)
			if got != tt.want {
				t.Errorf("IsLockFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestParseNumstat(t *testing.T) {
	// git diff --numstat output format: added<TAB>deleted<TAB>filename
	input := strings.Join([]string{
		"10\t5\tmain.go",
		"500\t200\tuv.lock",
		"3\t1\tREADME.md",
	}, "\n")

	stats := ParseNumstat(input)

	if len(stats) != 3 {
		t.Fatalf("expected 3 stats, got %d", len(stats))
	}

	tests := []struct {
		filename string
		added    int
		deleted  int
	}{
		{"main.go", 10, 5},
		{"uv.lock", 500, 200},
		{"README.md", 3, 1},
	}

	for i, tt := range tests {
		if stats[i].Filename != tt.filename {
			t.Errorf("stats[%d].Filename = %q, want %q", i, stats[i].Filename, tt.filename)
		}
		if stats[i].Added != tt.added {
			t.Errorf("stats[%d].Added = %d, want %d", i, stats[i].Added, tt.added)
		}
		if stats[i].Deleted != tt.deleted {
			t.Errorf("stats[%d].Deleted = %d, want %d", i, stats[i].Deleted, tt.deleted)
		}
	}
}

func TestParseNumstatBinaryFile(t *testing.T) {
	// Binary files show as "-" in numstat
	input := "-\t-\timage.png"

	stats := ParseNumstat(input)

	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}

	if stats[0].Filename != "image.png" {
		t.Errorf("Filename = %q, want %q", stats[0].Filename, "image.png")
	}
	if !stats[0].Binary {
		t.Error("expected Binary to be true")
	}
}

func TestStagedDiffWithSummary(t *testing.T) {
	repo := setupRepo(t)
	withRepo(t, repo, func() {
		// Create a regular file and a large lock file (above 150KB threshold)
		writeFile(t, repo, "main.go", "package main\n\nfunc main() {}\n")
		writeLockFileWithSize(t, repo, "uv.lock", 160*1024) // 160KB, above threshold

		runGit(t, repo, "add", "main.go", "uv.lock")

		diff, err := StagedDiffWithSummary()
		if err != nil {
			t.Fatalf("StagedDiffWithSummary error: %v", err)
		}

		// Should contain full diff for main.go
		if !strings.Contains(diff, "package main") {
			t.Error("expected diff to contain main.go content")
		}

		// Should contain summary for uv.lock, not full content
		if !strings.Contains(diff, "uv.lock") {
			t.Error("expected diff to mention uv.lock")
		}
		if strings.Contains(diff, "package-version-1") {
			t.Error("expected diff NOT to contain uv.lock content details")
		}
		// Should contain "content omitted" marker
		if !strings.Contains(diff, "content omitted") {
			t.Error("expected diff to contain summary marker")
		}
	})
}

// writeLockFileWithSize creates a lock file with approximately the specified size in bytes
func writeLockFileWithSize(t *testing.T, repo, name string, targetBytes int) {
	t.Helper()
	var content strings.Builder
	lineSize := len("package-version-0 = \"1.0.0\"\n")
	lines := targetBytes / lineSize
	for i := 0; i < lines; i++ {
		content.WriteString("package-version-")
		content.WriteString(strconv.Itoa(i))
		content.WriteString(" = \"1.0.0\"\n")
	}
	writeFile(t, repo, name, content.String())
}

// writeLargeLockFile creates a lock file with the specified number of lines (for backward compatibility)
func writeLargeLockFile(t *testing.T, repo, name string, lines int) {
	t.Helper()
	var content strings.Builder
	for i := 0; i < lines; i++ {
		content.WriteString("package-version-")
		content.WriteString(strconv.Itoa(i))
		content.WriteString(" = \"1.0.0\"\n")
	}
	writeFile(t, repo, name, content.String())
}

func TestStagedDiffWithSummaryOnlyLargeLockFile(t *testing.T) {
	repo := setupRepo(t)
	withRepo(t, repo, func() {
		// Only stage a large lock file (above 150KB threshold)
		writeLockFileWithSize(t, repo, "package-lock.json", 160*1024)
		runGit(t, repo, "add", "package-lock.json")

		diff, err := StagedDiffWithSummary()
		if err != nil {
			t.Fatalf("StagedDiffWithSummary error: %v", err)
		}

		// Should contain summary
		if !strings.Contains(diff, "package-lock.json") {
			t.Error("expected diff to mention package-lock.json")
		}
		if !strings.Contains(diff, "Lock file") {
			t.Error("expected diff to contain lock file summary marker")
		}
		// Should NOT contain actual content
		if strings.Contains(diff, "package-version-") {
			t.Error("expected diff NOT to contain lock file content")
		}
	})
}

func TestStagedDiffWithSummaryNoLockFiles(t *testing.T) {
	repo := setupRepo(t)
	withRepo(t, repo, func() {
		// Only regular files
		writeFile(t, repo, "main.go", "package main\n")
		writeFile(t, repo, "util.go", "package util\n")
		runGit(t, repo, "add", "main.go", "util.go")

		diff, err := StagedDiffWithSummary()
		if err != nil {
			t.Fatalf("StagedDiffWithSummary error: %v", err)
		}

		// Should contain full diff
		if !strings.Contains(diff, "package main") {
			t.Error("expected diff to contain main.go content")
		}
		if !strings.Contains(diff, "package util") {
			t.Error("expected diff to contain util.go content")
		}
		// Should NOT contain lock file marker
		if strings.Contains(diff, "Lock file") {
			t.Error("expected diff NOT to contain lock file marker")
		}
	})
}

func TestStagedDiffWithSummarySmallLockFile(t *testing.T) {
	repo := setupRepo(t)
	withRepo(t, repo, func() {
		// Small lock file (below 150KB threshold) should show full diff
		writeLockFileWithSize(t, repo, "uv.lock", 100*1024) // 100KB, below threshold
		runGit(t, repo, "add", "uv.lock")

		diff, err := StagedDiffWithSummary()
		if err != nil {
			t.Fatalf("StagedDiffWithSummary error: %v", err)
		}

		// Should contain full diff content (not summarized)
		if !strings.Contains(diff, "package-version-") {
			t.Error("expected small lock file to show full diff content")
		}
		// Should NOT contain summary marker
		if strings.Contains(diff, "content omitted") {
			t.Error("expected small lock file NOT to be summarized")
		}
	})
}

func TestStagedDiffWithSummaryLargeLockFile(t *testing.T) {
	repo := setupRepo(t)
	withRepo(t, repo, func() {
		// Large lock file (above 150KB threshold) should be summarized
		writeLockFileWithSize(t, repo, "uv.lock", 160*1024) // 160KB, above threshold
		runGit(t, repo, "add", "uv.lock")

		diff, err := StagedDiffWithSummary()
		if err != nil {
			t.Fatalf("StagedDiffWithSummary error: %v", err)
		}

		// Should contain summary marker
		if !strings.Contains(diff, "content omitted") {
			t.Error("expected large lock file to be summarized")
		}
		// Should NOT contain full content
		if strings.Contains(diff, "package-version-") {
			t.Error("expected large lock file NOT to show full content")
		}
	})
}
