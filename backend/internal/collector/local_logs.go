package collector

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultLimit    = 500
	defaultMaxBytes = 250000
	maxLimit        = 5000
	maxBytesLimit   = 1000000
)

type LogTailResult struct {
	File      string   `json:"file"`
	Cursor    int64    `json:"cursor"`
	Size      int64    `json:"size"`
	Lines     []string `json:"lines"`
	Truncated bool     `json:"truncated"`
	Reset     bool     `json:"reset"`
}

func TailLocalLogs(logDir string, cursor int64, limit int, maxBytes int) (LogTailResult, error) {
	file, err := resolveLogFile(logDir)
	if err != nil {
		return LogTailResult{}, err
	}
	return readLogSlice(file, cursor, limit, maxBytes)
}

func resolveLogFile(logDir string) (string, error) {
	info, err := os.Stat(logDir)
	if err == nil && !info.IsDir() {
		return logDir, nil
	}

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return "", err
	}
	type candidate struct {
		path  string
		mtime time.Time
	}
	candidates := make([]candidate, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "openclaw-") || !strings.HasSuffix(name, ".log") {
			if name != "openclaw.log" {
				continue
			}
		}
		full := filepath.Join(logDir, name)
		stat, err := os.Stat(full)
		if err != nil {
			continue
		}
		candidates = append(candidates, candidate{path: full, mtime: stat.ModTime()})
	}
	if len(candidates) == 0 {
		return "", errors.New("no log files found")
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].mtime.After(candidates[j].mtime) })
	return candidates[0].path, nil
}

func readLogSlice(path string, cursor int64, limit int, maxBytes int) (LogTailResult, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return LogTailResult{}, err
	}
	size := stat.Size()
	limit = clamp(limit, 1, maxLimit, defaultLimit)
	maxBytes = clamp(maxBytes, 1, maxBytesLimit, defaultMaxBytes)
	reset := false
	truncated := false
	start := int64(0)

	if cursor > 0 {
		if cursor > size {
			reset = true
			start = maxInt64(0, size-int64(maxBytes))
			truncated = start > 0
		} else {
			start = cursor
			if size-start > int64(maxBytes) {
				reset = true
				truncated = true
				start = maxInt64(0, size-int64(maxBytes))
			}
		}
	} else {
		start = maxInt64(0, size-int64(maxBytes))
		truncated = start > 0
	}

	if size == 0 || size <= start {
		return LogTailResult{File: path, Cursor: size, Size: size, Lines: []string{}, Truncated: truncated, Reset: reset}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return LogTailResult{}, err
	}
	defer file.Close()

	var prefix byte
	if start > 0 {
		buf := make([]byte, 1)
		if _, err := file.ReadAt(buf, start-1); err == nil {
			prefix = buf[0]
		}
	}
	length := size - start
	buf := make([]byte, length)
	if _, err := file.ReadAt(buf, start); err != nil && !errors.Is(err, io.EOF) {
		return LogTailResult{}, err
	}
	text := string(buf)
	lines := strings.Split(text, "\n")
	if start > 0 && prefix != '\n' {
		if len(lines) > 0 {
			lines = lines[1:]
		}
	}
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}

	return LogTailResult{
		File:      path,
		Cursor:    size,
		Size:      size,
		Lines:     lines,
		Truncated: truncated,
		Reset:     reset,
	}, nil
}

func clamp(value, min, max, def int) int {
	if value == 0 {
		value = def
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
