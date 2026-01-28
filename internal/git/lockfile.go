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

// LockFileSummaryThreshold is the minimum number of changed lines
// before a lock file diff is summarized instead of shown in full.
const LockFileSummaryThreshold = 200

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

	// Identify large lock files that should be summarized
	largeLockFiles := make(map[string]FileStat)
	for _, stat := range stats {
		if IsLockFile(stat.Filename) {
			totalChanges := stat.Added + stat.Deleted
			if totalChanges >= LockFileSummaryThreshold {
				largeLockFiles[stat.Filename] = stat
			}
		}
	}

	// If no large lock files, return standard diff
	if len(largeLockFiles) == 0 {
		return StagedDiff()
	}

	var result strings.Builder

	for _, stat := range stats {
		if lockStat, isLargeLock := largeLockFiles[stat.Filename]; isLargeLock {
			result.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", stat.Filename, stat.Filename))
			result.WriteString(fmt.Sprintf("[Lock file: +%d -%d lines, content omitted]\n\n", lockStat.Added, lockStat.Deleted))
		} else {
			fileDiff, err := stagedDiffForFile(stat.Filename)
			if err != nil {
				return "", err
			}
			result.WriteString(fileDiff)
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
