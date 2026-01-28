package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// lockFileNames contains lock files that don't match common suffixes
var lockFileNames = map[string]bool{
	"go.sum": true,
}

var lockFileSuffixes = []string{
	".lock",
	"-lock.json",
	"-lock.yaml",
}

// LockFileSummaryThreshold is the minimum diff size in bytes
// before a lock file diff is summarized instead of shown in full.
// 150KB is chosen based on LLM context limits (Claude Haiku: 200k tokens).
const LockFileSummaryThreshold = 150 * 1024 // 150KB

func IsLockFile(filename string) bool {
	base := filepath.Base(filename)

	if lockFileNames[base] {
		return true
	}

	lower := strings.ToLower(base)
	for _, suffix := range lockFileSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}

	return false
}

type FileStat struct {
	Filename string
	Added    int
	Deleted  int
	Binary   bool
}

func ParseNumstat(output string) []FileStat {
	var stats []FileStat
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}

		stat := FileStat{Filename: parts[2]}

		if parts[0] == "-" && parts[1] == "-" {
			stat.Binary = true
		} else {
			added, _ := strconv.Atoi(parts[0])
			deleted, _ := strconv.Atoi(parts[1])
			stat.Added = added
			stat.Deleted = deleted
		}

		stats = append(stats, stat)
	}

	return stats
}

func StagedDiffWithSummary() (string, error) {
	numstatCmd := exec.Command("git", "diff", "--staged", "--numstat")
	var numstatOut bytes.Buffer
	var numstatErr bytes.Buffer
	numstatCmd.Stdout = &numstatOut
	numstatCmd.Stderr = &numstatErr
	if err := numstatCmd.Run(); err != nil {
		return "", fmt.Errorf("git diff --numstat failed: %v: %s", err, strings.TrimSpace(numstatErr.String()))
	}

	stats := ParseNumstat(numstatOut.String())

	// Get diff for each file and check if lock files exceed threshold
	type fileDiffInfo struct {
		stat    FileStat
		diff    string
		isLarge bool
	}
	fileDiffs := make([]fileDiffInfo, 0, len(stats))

	for _, stat := range stats {
		diff, err := stagedDiffForFile(stat.Filename)
		if err != nil {
			return "", err
		}

		isLarge := IsLockFile(stat.Filename) && len(diff) >= LockFileSummaryThreshold
		fileDiffs = append(fileDiffs, fileDiffInfo{
			stat:    stat,
			diff:    diff,
			isLarge: isLarge,
		})
	}

	// Check if any lock file exceeds threshold
	hasLargeLockFile := false
	for _, fd := range fileDiffs {
		if fd.isLarge {
			hasLargeLockFile = true
			break
		}
	}

	// If no large lock files, return standard diff
	if !hasLargeLockFile {
		return StagedDiff()
	}

	var result strings.Builder

	for _, fd := range fileDiffs {
		if fd.isLarge {
			result.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", fd.stat.Filename, fd.stat.Filename))
			result.WriteString(fmt.Sprintf("[Lock file: +%d -%d lines, %d bytes, content omitted]\n\n",
				fd.stat.Added, fd.stat.Deleted, len(fd.diff)))
		} else {
			result.WriteString(fd.diff)
		}
	}

	return result.String(), nil
}

func stagedDiffForFile(filename string) (string, error) {
	cmd := exec.Command("git", "diff", "--staged", "--", filename)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff for %s failed: %v: %s", filename, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
