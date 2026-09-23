package packager

import (
	"archive/tar"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

func TestParseReplayTimelineAcceptsLargeRRWebEvent(t *testing.T) {
	const largeEventBytes = 2 << 20
	trace := `{"type":2,"timestamp":1788256800000,"data":"` + strings.Repeat("x", largeEventBytes) + `"}` + "\n" +
		`{"type":3,"timestamp":1788256802500}` + "\n"

	timeline, err := parseReplayTimeline(strings.NewReader(trace))
	if err != nil {
		t.Fatalf("parse large rrweb event: %v", err)
	}
	if timeline == nil {
		t.Fatal("expected timeline for valid rrweb events")
	}
	if timeline.FirstTimestamp != 1788256800000 || timeline.LastTimestamp != 1788256802500 || timeline.DurationMs != 2500 {
		t.Fatalf("unexpected timeline: %#v", timeline)
	}
}

func TestReadDiagnosticTarAllowsLargeRecordStreams(t *testing.T) {
	line := `{"id":"network-1","payload":"` + strings.Repeat("x", 300<<10) + `"}`
	var archive bytes.Buffer
	writer := tar.NewWriter(&archive)
	if err := writer.WriteHeader(&tar.Header{Name: "diagnostics/network.jsonl", Mode: 0o600, Size: int64(len(line))}); err != nil {
		t.Fatalf("write diagnostic header: %v", err)
	}
	if _, err := writer.Write([]byte(line)); err != nil {
		t.Fatalf("write diagnostic contents: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("finish diagnostic archive: %v", err)
	}

	result := DiagnosticEvidence{Network: []map[string]any{}}
	if err := readDiagnosticTar(archive.Bytes(), dawgtypes.MediaTypeDiagnostics, &result); err != nil {
		t.Fatalf("read aggregate diagnostic stream: %v", err)
	}
	if len(result.Network) != 1 {
		t.Fatalf("expected one network record, got %#v", result.Network)
	}
}

func TestDecompressZstdLimitedRejectsOversizedLayer(t *testing.T) {
	compressed, err := compressZstd([]byte(strings.Repeat("x", 1024)))
	if err != nil {
		t.Fatalf("compress fixture: %v", err)
	}
	if _, err := decompressZstdLimited(compressed, 512); !errors.Is(err, errDecompressedLayerTooLarge) {
		t.Fatalf("expected decompression size limit, got %v", err)
	}
}
