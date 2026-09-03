package verify

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

// CompareHTTPResponses reads two jsonl files containing dawgtypes.HTTPPair
// and compares their status codes. It returns the number of mismatches.
func CompareHTTPResponses(expectedPath, actualPath string) (int, error) {
	expectedPairs, err := loadHTTPPairs(expectedPath)
	if err != nil {
		// If baseline doesn't exist, we can't diff, but maybe that's fine.
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("verify: load expected http pairs: %w", err)
	}

	actualPairs, err := loadHTTPPairs(actualPath)
	if err != nil {
		if os.IsNotExist(err) {
			return len(expectedPairs), nil // all missing
		}
		return 0, fmt.Errorf("verify: load actual http pairs: %w", err)
	}

	mismatches := 0
	// For MVP, simply compare the sequences.
	// A more robust approach would correlate by URL/method.
	for i := 0; i < len(expectedPairs); i++ {
		if i >= len(actualPairs) {
			mismatches++
			continue
		}

		// Simple status code check for MVP
		if expectedPairs[i].Response.Status != actualPairs[i].Response.Status {
			mismatches++
		}
	}

	// Any extra responses are also mismatches
	if len(actualPairs) > len(expectedPairs) {
		mismatches += len(actualPairs) - len(expectedPairs)
	}

	return mismatches, nil
}

func loadHTTPPairs(path string) ([]dawgtypes.HTTPPair, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pairs []dawgtypes.HTTPPair
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var pair dawgtypes.HTTPPair
		if err := json.Unmarshal(scanner.Bytes(), &pair); err != nil {
			return nil, fmt.Errorf("parse http pair: %w", err)
		}
		pairs = append(pairs, pair)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return pairs, nil
}
