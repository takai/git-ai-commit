package git

import (
	"strings"
	"testing"
)

func TestFilter_Empty(t *testing.T) {
	result := Filter("", Options{MaxFileLines: 200})
	if result.Diff != "" {
		t.Errorf("expected empty diff, got %q", result.Diff)
	}
	if result.Truncated {
		t.Error("expected Truncated to be false")
	}
	if len(result.TruncatedFiles) != 0 {
		t.Errorf("expected no truncated files, got %v", result.TruncatedFiles)
	}
	if len(result.ExcludedFiles) != 0 {
		t.Errorf("expected no excluded files, got %v", result.ExcludedFiles)
	}
}

func TestFilter_NoTruncation(t *testing.T) {
	diff := `diff --git a/file.go b/file.go
index abc123..def456 100644
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"
 func main() {}
`
	result := Filter(diff, Options{MaxFileLines: 200})
	if result.Truncated {
		t.Error("expected Truncated to be false")
	}
	if len(result.TruncatedFiles) != 0 {
		t.Errorf("expected no truncated files, got %v", result.TruncatedFiles)
	}
	// Diff content should be preserved
	if !strings.Contains(result.Diff, "import \"fmt\"") {
		t.Error("expected diff content to be preserved")
	}
}

func TestFilter_TruncateFile(t *testing.T) {
	// Create a diff with many lines
	var sb strings.Builder
	sb.WriteString(`diff --git a/large.go b/large.go
index abc123..def456 100644
--- a/large.go
+++ b/large.go
@@ -1,100 +1,110 @@
`)
	for i := 0; i < 50; i++ {
		sb.WriteString("+new line\n")
	}
	diff := sb.String()

	result := Filter(diff, Options{MaxFileLines: 10})
	if !result.Truncated {
		t.Error("expected Truncated to be true")
	}
	if len(result.TruncatedFiles) != 1 || result.TruncatedFiles[0] != "large.go" {
		t.Errorf("expected truncated file 'large.go', got %v", result.TruncatedFiles)
	}
	if !strings.Contains(result.Diff, "truncated") {
		t.Error("expected truncation marker in diff")
	}
	if !strings.Contains(result.Diff, "showing 10 of 50 lines") {
		t.Errorf("expected truncation info, got %s", result.Diff)
	}
}

func TestFilter_ExcludeFiles(t *testing.T) {
	diff := `diff --git a/package-lock.json b/package-lock.json
index abc123..def456 100644
--- a/package-lock.json
+++ b/package-lock.json
@@ -1,3 +1,4 @@
+lock content
diff --git a/main.go b/main.go
index abc123..def456 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
+code content
`
	result := Filter(diff, Options{
		ExcludePatterns: []string{"**/*-lock.json"},
	})

	if len(result.ExcludedFiles) != 1 || result.ExcludedFiles[0] != "package-lock.json" {
		t.Errorf("expected excluded file 'package-lock.json', got %v", result.ExcludedFiles)
	}
	if strings.Contains(result.Diff, "package-lock.json") {
		t.Error("expected package-lock.json to be excluded from diff")
	}
	if !strings.Contains(result.Diff, "main.go") {
		t.Error("expected main.go to be in diff")
	}
}

func TestFilter_ExcludeGoSum(t *testing.T) {
	diff := `diff --git a/go.sum b/go.sum
index abc123..def456 100644
--- a/go.sum
+++ b/go.sum
@@ -1,3 +1,4 @@
+module v1.0.0 h1:abc=
diff --git a/main.go b/main.go
index abc123..def456 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
+code
`
	result := Filter(diff, Options{
		ExcludePatterns: DefaultExcludePatterns(),
	})

	if len(result.ExcludedFiles) != 1 || result.ExcludedFiles[0] != "go.sum" {
		t.Errorf("expected excluded file 'go.sum', got %v", result.ExcludedFiles)
	}
	if strings.Contains(result.Diff, "go.sum") {
		t.Error("expected go.sum to be excluded from diff")
	}
}

func TestFilter_MultipleFiles(t *testing.T) {
	diff := `diff --git a/a.go b/a.go
index abc123..def456 100644
--- a/a.go
+++ b/a.go
@@ -1,1 +1,2 @@
+line a
diff --git a/b.go b/b.go
index abc123..def456 100644
--- a/b.go
+++ b/b.go
@@ -1,1 +1,2 @@
+line b
diff --git a/c.go b/c.go
index abc123..def456 100644
--- a/c.go
+++ b/c.go
@@ -1,1 +1,2 @@
+line c
`
	result := Filter(diff, Options{MaxFileLines: 200})
	if result.Truncated {
		t.Error("expected no truncation")
	}
	if !strings.Contains(result.Diff, "a.go") {
		t.Error("expected a.go in diff")
	}
	if !strings.Contains(result.Diff, "b.go") {
		t.Error("expected b.go in diff")
	}
	if !strings.Contains(result.Diff, "c.go") {
		t.Error("expected c.go in diff")
	}
}

func TestFilter_Deterministic(t *testing.T) {
	diff := `diff --git a/z.go b/z.go
index abc123..def456 100644
--- a/z.go
+++ b/z.go
@@ -1,1 +1,2 @@
+z
diff --git a/a.go b/a.go
index abc123..def456 100644
--- a/a.go
+++ b/a.go
@@ -1,1 +1,2 @@
+a
diff --git a/m.go b/m.go
index abc123..def456 100644
--- a/m.go
+++ b/m.go
@@ -1,1 +1,2 @@
+m
`
	result1 := Filter(diff, Options{MaxFileLines: 200})
	result2 := Filter(diff, Options{MaxFileLines: 200})

	if result1.Diff != result2.Diff {
		t.Error("expected deterministic output")
	}

	// Files should be sorted alphabetically
	aIdx := strings.Index(result1.Diff, "a.go")
	mIdx := strings.Index(result1.Diff, "m.go")
	zIdx := strings.Index(result1.Diff, "z.go")

	if !(aIdx < mIdx && mIdx < zIdx) {
		t.Errorf("expected files in alphabetical order, got a=%d, m=%d, z=%d", aIdx, mIdx, zIdx)
	}
}

func TestFilter_ExcludeMinifiedFiles(t *testing.T) {
	diff := `diff --git a/app.min.js b/app.min.js
index abc123..def456 100644
--- a/app.min.js
+++ b/app.min.js
@@ -1,1 +1,2 @@
+minified
diff --git a/style.min.css b/style.min.css
index abc123..def456 100644
--- a/style.min.css
+++ b/style.min.css
@@ -1,1 +1,2 @@
+minified css
diff --git a/app.js b/app.js
index abc123..def456 100644
--- a/app.js
+++ b/app.js
@@ -1,1 +1,2 @@
+normal
`
	result := Filter(diff, Options{
		ExcludePatterns: DefaultExcludePatterns(),
	})

	if len(result.ExcludedFiles) != 2 {
		t.Errorf("expected 2 excluded files, got %v", result.ExcludedFiles)
	}
	if !strings.Contains(result.Diff, "app.js") {
		t.Error("expected app.js in diff")
	}
	if strings.Contains(result.Diff, "app.min.js") {
		t.Error("expected app.min.js to be excluded")
	}
	if strings.Contains(result.Diff, "style.min.css") {
		t.Error("expected style.min.css to be excluded")
	}
}

func TestFilter_NoLimitWhenZero(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`diff --git a/large.go b/large.go
index abc123..def456 100644
--- a/large.go
+++ b/large.go
@@ -1,100 +1,150 @@
`)
	for i := 0; i < 100; i++ {
		sb.WriteString("+new line\n")
	}
	diff := sb.String()

	result := Filter(diff, Options{MaxFileLines: 0}) // 0 means no limit
	if result.Truncated {
		t.Error("expected no truncation when MaxFileLines is 0")
	}
}

func TestDefaultExcludePatterns(t *testing.T) {
	patterns := DefaultExcludePatterns()
	if len(patterns) == 0 {
		t.Error("expected default patterns")
	}

	// Check that expected patterns are present
	expected := []string{
		"**/*.lock",
		"**/*-lock.json",
		"**/go.sum",
	}
	for _, exp := range expected {
		found := false
		for _, p := range patterns {
			if p == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected pattern %q in defaults", exp)
		}
	}
}

func TestMatchPattern_DoubleStarGlob(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"go.sum", "**/go.sum", true},
		{"vendor/go.sum", "**/go.sum", true},
		{"package-lock.json", "**/*-lock.json", true},
		{"node_modules/package-lock.json", "**/*-lock.json", true},
		{"deep/nested/package-lock.json", "**/*-lock.json", true},
		{"app.min.js", "**/*.min.js", true},
		{"dist/app.min.js", "**/*.min.js", true},
		{"src/styles/main.min.css", "**/*.min.css", true},
		{"app.js", "**/*.min.js", false},
		{"main.go", "**/go.sum", false},
		{"vendor/module/main.go", "vendor/**", true},
		{"vendor/main.go", "vendor/**", true},
	}

	for _, tt := range tests {
		got := matchPattern(tt.path, tt.pattern)
		if got != tt.want {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
		}
	}
}

func TestExtractFilePath(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"diff --git a/file.go b/file.go", "file.go"},
		{"diff --git a/path/to/file.go b/path/to/file.go", "path/to/file.go"},
		{"diff --git a/old/name.go b/new/name.go", "new/name.go"},
	}

	for _, tt := range tests {
		got := extractFilePath(tt.line)
		if got != tt.want {
			t.Errorf("extractFilePath(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}
