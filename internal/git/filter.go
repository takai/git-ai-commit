package git

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// DefaultMaxFileLines is the default maximum number of lines per file.
const DefaultMaxFileLines = 100

// DefaultExcludePatterns returns the built-in default patterns for files to exclude from diff.
func DefaultExcludePatterns() []string {
	return []string{
		"**/*.lock",
		"**/*-lock.json",
		"**/*.lock.yaml",
		"**/*-lock.yaml",
		"**/*.lockfile",
		"**/*.min.js",
		"**/*.min.css",
		"**/*.map",
		"**/go.sum",
	}
}

// Options holds filtering configuration.
type Options struct {
	MaxFileLines    int      // Maximum lines per file (0 = no limit)
	ExcludePatterns []string // Glob patterns for files to exclude
}

// Result holds the filtering outcome.
type Result struct {
	Diff           string   // The filtered diff output
	Truncated      bool     // True if any file was truncated
	TruncatedFiles []string // List of truncated file paths
	ExcludedFiles  []string // List of excluded file paths
}

// Filter filters a unified diff according to the given options.
func Filter(diff string, opts Options) Result {
	if strings.TrimSpace(diff) == "" {
		return Result{Diff: diff}
	}

	files := splitDiffByFile(diff)
	if len(files) == 0 {
		return Result{Diff: diff}
	}

	var result Result
	var filteredParts []string

	// Sort file names for deterministic output
	fileNames := make([]string, 0, len(files))
	for name := range files {
		fileNames = append(fileNames, name)
	}
	sort.Strings(fileNames)

	for _, fileName := range fileNames {
		content := files[fileName]

		// Check exclusion patterns
		if matchesAnyPattern(fileName, opts.ExcludePatterns) {
			result.ExcludedFiles = append(result.ExcludedFiles, fileName)
			continue
		}

		// Apply line limit if configured
		if opts.MaxFileLines > 0 {
			truncated, newContent := truncateFileDiff(content, opts.MaxFileLines, fileName)
			if truncated {
				result.Truncated = true
				result.TruncatedFiles = append(result.TruncatedFiles, fileName)
			}
			content = newContent
		}

		filteredParts = append(filteredParts, content)
	}

	result.Diff = strings.Join(filteredParts, "")
	return result
}

// splitDiffByFile splits a unified diff into per-file sections.
// Returns a map of file path to diff content (including header).
func splitDiffByFile(diff string) map[string]string {
	files := make(map[string]string)
	lines := strings.Split(diff, "\n")

	var currentFile string
	var currentContent strings.Builder
	var inFile bool

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Detect start of a new file diff
		if strings.HasPrefix(line, "diff --git ") {
			// Save previous file if any
			if inFile && currentFile != "" {
				files[currentFile] = currentContent.String()
			}

			// Extract file path from "diff --git a/path b/path"
			currentFile = extractFilePath(line)
			currentContent.Reset()
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
			inFile = true
			continue
		}

		if inFile {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	// Save last file
	if inFile && currentFile != "" {
		files[currentFile] = currentContent.String()
	}

	return files
}

// extractFilePath extracts the file path from a "diff --git a/path b/path" line.
func extractFilePath(line string) string {
	// Format: "diff --git a/path/to/file b/path/to/file"
	parts := strings.SplitN(line, " ", 4)
	if len(parts) < 4 {
		return ""
	}
	// Use the b/ path (destination)
	bPath := parts[3]
	if strings.HasPrefix(bPath, "b/") {
		return bPath[2:]
	}
	return bPath
}

// truncateFileDiff truncates a file diff to the specified number of lines.
// Returns (wasTruncated, newContent).
func truncateFileDiff(content string, maxLines int, fileName string) (bool, string) {
	lines := strings.Split(content, "\n")

	// Count only the actual diff lines (not headers)
	headerEnd := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "@@") {
			headerEnd = i
			break
		}
	}

	// Count diff content lines (after first @@ marker)
	diffLineCount := 0
	for i := headerEnd; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, " ") {
			diffLineCount++
		}
	}

	if diffLineCount <= maxLines {
		return false, content
	}

	// Truncate: keep header + maxLines of diff content
	var result strings.Builder
	contentLinesSeen := 0

	for _, line := range lines {
		// Always include header lines (before first @@)
		if !strings.HasPrefix(line, "@@") && contentLinesSeen == 0 {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Include @@ hunk headers
		if strings.HasPrefix(line, "@@") {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Count and potentially truncate content lines
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, " ") {
			contentLinesSeen++
			if contentLinesSeen <= maxLines {
				result.WriteString(line)
				result.WriteString("\n")
			}
		} else {
			// Other lines (like "\ No newline at end of file")
			if contentLinesSeen <= maxLines {
				result.WriteString(line)
				result.WriteString("\n")
			}
		}
	}

	// Add truncation marker
	result.WriteString(fmt.Sprintf("\n... [%s truncated: showing %d of %d lines]\n", fileName, maxLines, diffLineCount))

	return true, result.String()
}

// matchesAnyPattern checks if the file path matches any of the glob patterns.
func matchesAnyPattern(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchPattern(filePath, pattern) {
			return true
		}
	}
	return false
}

// matchPattern checks if a file path matches a glob pattern.
// Supports ** for recursive directory matching.
func matchPattern(filePath, pattern string) bool {
	// Handle ** patterns specially
	if strings.Contains(pattern, "**") {
		return matchDoubleStarPattern(filePath, pattern)
	}

	// Use standard filepath.Match for simple patterns
	matched, err := filepath.Match(pattern, filePath)
	if err != nil {
		return false
	}
	if matched {
		return true
	}

	// Also try matching against just the filename
	matched, err = filepath.Match(pattern, filepath.Base(filePath))
	if err != nil {
		return false
	}
	return matched
}

// matchDoubleStarPattern handles patterns containing **.
func matchDoubleStarPattern(filePath, pattern string) bool {
	// Split pattern by **
	parts := strings.Split(pattern, "**")
	if len(parts) == 1 {
		// No ** found, use standard matching
		matched, _ := filepath.Match(pattern, filePath)
		return matched
	}

	// Handle common case: **/suffix (e.g., **/*.lock)
	if parts[0] == "" && len(parts) == 2 {
		suffix := strings.TrimPrefix(parts[1], "/")
		if suffix == "" {
			return true // ** matches everything
		}
		// Match suffix against file name or any suffix of the path
		if matchPathSuffix(filePath, suffix) {
			return true
		}
	}

	// Handle prefix/**/suffix patterns
	if len(parts) == 2 {
		prefix := strings.TrimSuffix(parts[0], "/")
		suffix := strings.TrimPrefix(parts[1], "/")

		// If prefix is empty, we've handled this above
		if prefix == "" {
			return false
		}

		// Check if path starts with prefix
		if !strings.HasPrefix(filePath, prefix) && !matchPathPrefix(filePath, prefix) {
			return false
		}

		// Check if remaining path matches suffix
		remaining := strings.TrimPrefix(filePath, prefix)
		remaining = strings.TrimPrefix(remaining, "/")
		if suffix == "" {
			return true
		}
		return matchPathSuffix(remaining, suffix)
	}

	return false
}

// matchPathSuffix checks if the path or any of its suffixes match the pattern.
func matchPathSuffix(path, pattern string) bool {
	// Try direct match
	matched, _ := filepath.Match(pattern, path)
	if matched {
		return true
	}

	// Try matching just the filename
	matched, _ = filepath.Match(pattern, filepath.Base(path))
	if matched {
		return true
	}

	// Try matching path suffixes
	parts := strings.Split(path, "/")
	for i := range parts {
		suffix := strings.Join(parts[i:], "/")
		matched, _ := filepath.Match(pattern, suffix)
		if matched {
			return true
		}
	}

	return false
}

// matchPathPrefix checks if the path starts with the given prefix pattern.
func matchPathPrefix(path, prefix string) bool {
	pathParts := strings.Split(path, "/")
	prefixParts := strings.Split(prefix, "/")

	if len(pathParts) < len(prefixParts) {
		return false
	}

	for i, pp := range prefixParts {
		matched, _ := filepath.Match(pp, pathParts[i])
		if !matched {
			return false
		}
	}

	return true
}
