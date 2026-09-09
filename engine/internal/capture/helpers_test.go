package capture

import (
	"os"
	"strings"
	"testing"
	"time"
)

func waitForNonEmptyFile(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		contents, err := os.ReadFile(path)
		if err == nil && len(strings.TrimSpace(string(contents))) > 0 {
			return
		}
		select {
		case <-timer.C:
			t.Fatalf("timed out waiting for %s", path)
		case <-ticker.C:
		}
	}
}

