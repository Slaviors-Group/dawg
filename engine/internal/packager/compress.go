// Package packager implements PRD §6.3 OCI artifact construction for sanitized capture output.
// It does not sanitize data, push artifacts, or execute replay.
package packager

import (
	"bytes"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

func compressZstd(contents []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer, err := zstd.NewWriter(&buffer)
	if err != nil {
		return nil, fmt.Errorf("packager: create zstd writer: %w", err)
	}
	if _, err := writer.Write(contents); err != nil {
		writer.Close()
		return nil, fmt.Errorf("packager: compress layer: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("packager: finish layer compression: %w", err)
	}
	return buffer.Bytes(), nil
}

func decompressZstd(contents []byte) ([]byte, error) {
	reader, err := zstd.NewReader(bytes.NewReader(contents))
	if err != nil {
		return nil, fmt.Errorf("packager: open zstd layer: %w", err)
	}
	defer reader.Close()
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("packager: decompress layer: %w", err)
	}
	return decoded, nil
}
