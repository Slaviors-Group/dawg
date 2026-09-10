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

	// scanLine carries one scanned line or a terminal scan error from the
	// background goroutine to the main select loop.
	type scanLine struct {
		text string
		err  error // non-nil signals end of input (io.EOF → nil text, or scan error)
		done bool  // true when the scanner has finished (successfully or not)
	}
	lines := make(chan scanLine, 16)

	// Run the scanner in a goroutine so that scanner.Scan(), which blocks in a
	// syscall, does not prevent ctx cancellation from being observed promptly.
	// Previously, ctx.Err() was only checked between completed lines, meaning a
	// Stop() request would be ignored while Scan() was blocked waiting for the
	// next line from a slow or live pipe.
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(source)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			select {
			case lines <- scanLine{text: scanner.Text()}:
			case <-ctx.Done():
				return
			}
		}
		// Signal end: send scanner.Err() (nil on clean EOF).
		lines <- scanLine{done: true, err: scanner.Err()}
	}()

	result := DBDiffCaptureResult{}
	lineNumber := 0
	for {
		select {
		case <-ctx.Done():
			return DBDiffCaptureResult{}, fmt.Errorf("capture: read DB diff: %w", ctx.Err())
		case sl, ok := <-lines:
			if !ok {
				// Channel closed early (ctx cancelled inside goroutine).
				return DBDiffCaptureResult{}, fmt.Errorf("capture: read DB diff: %w", ctx.Err())
			}
			if sl.done {
				if sl.err != nil {
					return DBDiffCaptureResult{}, fmt.Errorf("capture: scan DB diffs: %w", sl.err)
				}
				captureLogger(logger).Info("DB capture completed", "component", "capture", "entries", result.Entries)
				return result, nil
			}
			lineNumber++
			line := strings.TrimSpace(sl.text)
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
	}
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
