package packager

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

// validateReplayTrace prevents publishing artifacts that have timestamps but no
// rrweb document snapshot, which cannot produce a usable replay.
func validateReplayTrace(sessionDirectory string) error {
	path := filepath.Join(sessionDirectory, "traces", "rrweb.jsonl")
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("packager: open replay trace: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	events := 0
	fullSnapshots := 0
	validTimestamps := 0
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > dawgtypes.MaxJSONLRecordBytes {
			return fmt.Errorf("packager: replay trace record exceeds %d bytes", dawgtypes.MaxJSONLRecordBytes)
		}
		if len(line) > 0 {
			var event struct {
				Type      int   `json:"type"`
				Timestamp int64 `json:"timestamp"`
			}
			if err := json.Unmarshal(line, &event); err != nil {
				return fmt.Errorf("packager: decode replay trace record: %w", err)
			}
			events++
			if event.Timestamp > 0 {
				validTimestamps++
			}
			if event.Type == 2 {
				fullSnapshots++
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("packager: read replay trace: %w", readErr)
		}
	}
	if events == 0 {
		return fmt.Errorf("packager: replay trace contains no rrweb events")
	}
	if validTimestamps == 0 {
		return fmt.Errorf("packager: replay trace contains no valid timestamps")
	}
	if fullSnapshots == 0 {
		return fmt.Errorf("packager: replay trace contains no FullSnapshot event")
	}
	return nil
}
