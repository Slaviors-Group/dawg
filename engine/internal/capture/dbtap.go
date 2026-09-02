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

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

// DBDiffCaptureResult records whether an ORM adapter supplied any database changes.
type DBDiffCaptureResult struct {
	Entries int
	Skipped bool
	Reason  string
}

// CaptureDBDiffs validates normalized ORM change entries and writes db/diff.jsonl.
// A nil source means no supported DB adapter is configured and is a non-fatal P0 skip.
func CaptureDBDiffs(ctx context.Context, source io.Reader, sessionDirectory string, logger *slog.Logger) (DBDiffCaptureResult, error) {
	if source == nil {
		captureLogger(logger).Warn("DB capture skipped", "component", "capture", "reason", "no supported ORM adapter configured")
		return DBDiffCaptureResult{Skipped: true, Reason: "no supported ORM adapter configured"}, nil
	}
	if sessionDirectory == "" {
		return DBDiffCaptureResult{}, fmt.Errorf("capture: session directory is required")
	}
	path := filepath.Join(sessionDirectory, "db", "diff.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return DBDiffCaptureResult{}, fmt.Errorf("capture: create DB output directory: %w", err)
	}
	output, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return DBDiffCaptureResult{}, fmt.Errorf("capture: open DB diff output: %w", err)
	}
	defer output.Close()

	result := DBDiffCaptureResult{}
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		if err := ctx.Err(); err != nil {
			return DBDiffCaptureResult{}, fmt.Errorf("capture: read DB diff: %w", err)
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var diff dawgtypes.DBDiff
		if err := json.Unmarshal([]byte(line), &diff); err != nil {
			return DBDiffCaptureResult{}, fmt.Errorf("capture: decode DB diff line %d: %w", lineNumber, err)
		}
		if err := validateDBDiff(diff); err != nil {
			return DBDiffCaptureResult{}, fmt.Errorf("capture: validate DB diff line %d: %w", lineNumber, err)
		}
		contents, err := json.Marshal(diff)
		if err != nil {
			return DBDiffCaptureResult{}, fmt.Errorf("capture: encode DB diff line %d: %w", lineNumber, err)
		}
		if _, err := output.Write(append(contents, '\n')); err != nil {
			return DBDiffCaptureResult{}, fmt.Errorf("capture: write DB diff line %d: %w", lineNumber, err)
		}
		result.Entries++
	}
	if err := scanner.Err(); err != nil {
		return DBDiffCaptureResult{}, fmt.Errorf("capture: scan DB diffs: %w", err)
	}
	captureLogger(logger).Info("DB capture completed", "component", "capture", "entries", result.Entries)
	return result, nil
}

func validateDBDiff(diff dawgtypes.DBDiff) error {
	if diff.Timestamp.IsZero() || diff.Table == "" || diff.Source == "" || len(diff.PrimaryKey) == 0 {
		return fmt.Errorf("timestamp, table, primary key, and source are required")
	}
	switch diff.Operation {
	case "INSERT", "UPDATE", "DELETE", "FIXTURE":
		return nil
	default:
		return fmt.Errorf("unsupported operation %q", diff.Operation)
	}
}

func captureLogger(logger *slog.Logger) *slog.Logger {
	if logger != nil {
		return logger
	}
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}
