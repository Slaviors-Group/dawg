package packager

import (
	"archive/tar"
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
)

// DiagnosticEvidence is the bounded, already-sanitized evidence available to
// local inspectors and exporters. It never reconstructs missing body content.
type DiagnosticEvidence struct {
	Summary  *dawgtypes.DiagnosticsSummary `json:"summary,omitempty"`
	Timeline *ReplayTimeline               `json:"timeline,omitempty"`
	Console  []map[string]any              `json:"console"`
	Network  []map[string]any              `json:"network"`
	Errors   []map[string]any              `json:"errors"`
	Bodies   map[string]string             `json:"bodies"`
}

// ReplayTimeline describes the rrweb time range used for approximate
// diagnostic-to-replay correlation. It is omitted when no valid trace exists.
type ReplayTimeline struct {
	FirstTimestamp int64 `json:"firstTimestamp"`
	LastTimestamp  int64 `json:"lastTimestamp"`
	DurationMs     int64 `json:"durationMs"`
}

var errTraceCorrelationLimit = errors.New("diagnostics: trace layer exceeds correlation size limit")

// ReadDiagnostics verifies declared diagnostic blobs before reading their
// bounded tar entries. Legacy artifacts return empty evidence and no error.
func ReadDiagnostics(layoutDirectory string) (DiagnosticEvidence, error) {
	value, err := manifest.ReadValidated(filepath.Join(layoutDirectory, "dawg-manifest.json"))
	if err != nil {
		return DiagnosticEvidence{}, fmt.Errorf("diagnostics: read manifest: %w", err)
	}
	result := DiagnosticEvidence{Summary: value.Diagnostics, Console: []map[string]any{}, Network: []map[string]any{}, Errors: []map[string]any{}, Bodies: map[string]string{}}
	for _, layer := range value.Layers {
		if layer.MediaType == dawgtypes.MediaTypeTrace {
			timeline, timelineErr := readReplayTimeline(layoutDirectory, layer)
			if timelineErr != nil && !errors.Is(timelineErr, errTraceCorrelationLimit) {
				return DiagnosticEvidence{}, timelineErr
			}
			result.Timeline = timeline
			continue
		}
		if layer.MediaType != dawgtypes.MediaTypeDiagnostics && layer.MediaType != dawgtypes.MediaTypeDiagnosticBodies {
			continue
		}
		limit := int64(dawgtypes.DefaultDiagnosticLimits().MaxLayerBytes)
		if layer.MediaType == dawgtypes.MediaTypeDiagnosticBodies {
			limit = dawgtypes.DefaultDiagnosticLimits().MaxBodiesBytes
		}
		if layer.Size < 0 || layer.Size > limit {
			return DiagnosticEvidence{}, fmt.Errorf("diagnostics: layer exceeds compressed size limit")
		}
		digestPath, err := layerDigestPath(layer.Digest)
		if err != nil {
			return DiagnosticEvidence{}, err
		}
		contents, err := os.ReadFile(filepath.Join(layoutDirectory, "blobs", "sha256", digestPath))
		if err != nil {
			return DiagnosticEvidence{}, fmt.Errorf("diagnostics: read layer: %w", err)
		}
		digest := sha256.Sum256(contents)
		if int64(len(contents)) != layer.Size || fmt.Sprintf("sha256:%x", digest) != layer.Digest {
			return DiagnosticEvidence{}, fmt.Errorf("diagnostics: layer descriptor does not match blob")
		}
		decoded, err := decompressZstdLimited(contents, limit)
		if err != nil {
			return DiagnosticEvidence{}, err
		}
		if err := readDiagnosticTar(decoded, layer.MediaType, &result); err != nil {
			return DiagnosticEvidence{}, err
		}
	}
	return result, nil
}

func readReplayTimeline(layoutDirectory string, layer dawgtypes.LayerSpec) (*ReplayTimeline, error) {
	const maxCompressedTraceBytes = 64 << 20
	const maxDecodedTraceBytes = 128 << 20
	if layer.Size < 0 || layer.Size > maxCompressedTraceBytes {
		return nil, errTraceCorrelationLimit
	}
	digestPath, err := layerDigestPath(layer.Digest)
	if err != nil {
		return nil, err
	}
	contents, err := os.ReadFile(filepath.Join(layoutDirectory, "blobs", "sha256", digestPath))
	if err != nil {
		return nil, fmt.Errorf("diagnostics: read trace layer: %w", err)
	}
	digest := sha256.Sum256(contents)
	if int64(len(contents)) != layer.Size || fmt.Sprintf("sha256:%x", digest) != layer.Digest {
		return nil, fmt.Errorf("diagnostics: trace layer descriptor does not match blob")
	}
	decoded, err := decompressZstdLimited(contents, maxDecodedTraceBytes)
	if errors.Is(err, errDecompressedLayerTooLarge) {
		return nil, errTraceCorrelationLimit
	}
	if err != nil {
		return nil, err
	}
	reader := tar.NewReader(bytes.NewReader(decoded))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("diagnostics: read trace tar: %w", err)
		}
		if header.Typeflag != tar.TypeReg || header.Size < 0 || header.Size > maxDecodedTraceBytes || strings.Contains(header.Name, "..") || filepath.IsAbs(header.Name) {
			return nil, fmt.Errorf("diagnostics: unsafe trace tar entry %q", header.Name)
		}
		if filepath.ToSlash(header.Name) != "traces/rrweb.jsonl" {
			continue
		}
		return parseReplayTimeline(io.LimitReader(reader, header.Size+1))
	}
}

func parseReplayTimeline(reader io.Reader) (*ReplayTimeline, error) {
	lineReader := bufio.NewReader(reader)
	var timeline ReplayTimeline
	found := false
	for {
		line, err := lineReader.ReadBytes('\n')
		if len(line) > 0 {
			var event struct {
				Timestamp int64 `json:"timestamp"`
			}
			if json.Unmarshal(line, &event) == nil && event.Timestamp > 0 {
				if !found {
					timeline.FirstTimestamp = event.Timestamp
					found = true
				}
				timeline.LastTimestamp = event.Timestamp
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("diagnostics: scan replay trace: %w", err)
		}
	}
	if !found || timeline.LastTimestamp < timeline.FirstTimestamp {
		return nil, nil
	}
	timeline.DurationMs = timeline.LastTimestamp - timeline.FirstTimestamp
	return &timeline, nil
}

func readDiagnosticTar(contents []byte, mediaType dawgtypes.MediaType, result *DiagnosticEvidence) error {
	limits := dawgtypes.DefaultDiagnosticLimits()
	maxEntries := 3
	maxEntryBytes := limits.MaxLayerBytes
	if mediaType == dawgtypes.MediaTypeDiagnosticBodies {
		maxEntries = int(limits.MaxNetworkRecords * 2)
		maxEntryBytes = limits.MaxBodyBytes
	}

	reader := tar.NewReader(bytes.NewReader(contents))
	entries := 0
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("diagnostics: read tar: %w", err)
		}
		entries++
		if entries > maxEntries || header.Typeflag != tar.TypeReg || header.Size < 0 || header.Size > maxEntryBytes || strings.Contains(header.Name, "..") || filepath.IsAbs(header.Name) {
			return fmt.Errorf("diagnostics: unsafe diagnostic tar entry %q", header.Name)
		}
		contents, err := io.ReadAll(io.LimitReader(reader, header.Size+1))
		if err != nil || int64(len(contents)) != header.Size {
			return fmt.Errorf("diagnostics: read bounded entry %q", header.Name)
		}
		name := filepath.ToSlash(header.Name)
		if mediaType == dawgtypes.MediaTypeDiagnosticBodies {
			result.Bodies[name] = string(contents)
			continue
		}
		var destination *[]map[string]any
		switch name {
		case "diagnostics/console.jsonl":
			destination = &result.Console
		case "diagnostics/network.jsonl":
			destination = &result.Network
		case "diagnostics/errors.jsonl":
			destination = &result.Errors
		default:
			return fmt.Errorf("diagnostics: unexpected records entry %q", name)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(contents)), "\n") {
			if line == "" {
				continue
			}
			var record map[string]any
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				return fmt.Errorf("diagnostics: decode %s: %w", name, err)
			}
			*destination = append(*destination, record)
		}
	}
}
