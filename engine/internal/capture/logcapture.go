package capture

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CaptureStructuredLogs copies JSON-object log lines from a stream to logs/structured.jsonl.
func CaptureStructuredLogs(ctx context.Context, source io.Reader, sessionDirectory string, logger *slog.Logger) (int, error) {
	if source == nil {
		return 0, fmt.Errorf("capture: structured log source is required")
	}
	path, output, err := openStructuredLogOutput(sessionDirectory)
	if err != nil {
		return 0, err
	}
	defer output.Close()
	entries, err := copyStructuredLogLines(ctx, source, output)
	if err != nil {
		return 0, err
	}
	captureLogger(logger).Info("structured log capture completed", "component", "capture", "entries", entries, "path", path)
	return entries, nil
}

// TailStructuredLogFile copies complete JSON log lines appended to a file until the context ends.
func TailStructuredLogFile(ctx context.Context, sourcePath, sessionDirectory string, pollInterval time.Duration, logger *slog.Logger) (int, error) {
	if sourcePath == "" {
		return 0, fmt.Errorf("capture: structured log file path is required")
	}
	if pollInterval <= 0 {
		pollInterval = 100 * time.Millisecond
	}
	path, output, err := openStructuredLogOutput(sessionDirectory)
	if err != nil {
		return 0, err
	}
	defer output.Close()

	var offset int64
	entries := 0
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		count, nextOffset, err := copyAppendedLogLines(sourcePath, offset, output)
		if err != nil {
			return 0, err
		}
		entries += count
		offset = nextOffset
		select {
		case <-ctx.Done():
			captureLogger(logger).Info("structured log file tail stopped", "component", "capture", "entries", entries, "path", path)
			return entries, nil
		case <-ticker.C:
		}
	}
}

func openStructuredLogOutput(sessionDirectory string) (string, *os.File, error) {
	if sessionDirectory == "" {
		return "", nil, fmt.Errorf("capture: session directory is required")
	}
	path := filepath.Join(sessionDirectory, "logs", "structured.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", nil, fmt.Errorf("capture: create log output directory: %w", err)
	}
	output, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return "", nil, fmt.Errorf("capture: open structured log output: %w", err)
	}
	return path, output, nil
}

func copyStructuredLogLines(ctx context.Context, source io.Reader, output io.Writer) (int, error) {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	entries := 0
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		if err := ctx.Err(); err != nil {
			return 0, fmt.Errorf("capture: read structured logs: %w", err)
		}
		if err := writeStructuredLogLine(output, scanner.Text(), lineNumber); err != nil {
			return 0, err
		}
		if strings.TrimSpace(scanner.Text()) != "" {
			entries++
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("capture: scan structured logs: %w", err)
	}
	return entries, nil
}

func copyAppendedLogLines(sourcePath string, offset int64, output io.Writer) (int, int64, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, offset, nil
		}
		return 0, offset, fmt.Errorf("capture: open structured log file %s: %w", sourcePath, err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return 0, offset, fmt.Errorf("capture: inspect structured log file %s: %w", sourcePath, err)
	}
	if info.Size() < offset {
		offset = 0
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return 0, offset, fmt.Errorf("capture: seek structured log file %s: %w", sourcePath, err)
	}
	count, err := copyStructuredLogLines(context.Background(), file, output)
	if err != nil {
		return 0, offset, err
	}
	nextOffset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, offset, fmt.Errorf("capture: determine structured log offset: %w", err)
	}
	return count, nextOffset, nil
}

func writeStructuredLogLine(output io.Writer, line string, lineNumber int) error {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(line), &value); err != nil {
		return fmt.Errorf("capture: decode structured log line %d: %w", lineNumber, err)
	}
	contents, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("capture: encode structured log line %d: %w", lineNumber, err)
	}
	if _, err := output.Write(append(contents, '\n')); err != nil {
		return fmt.Errorf("capture: write structured log line %d: %w", lineNumber, err)
	}
	return nil
}
