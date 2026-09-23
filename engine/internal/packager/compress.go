// Package packager constructs OCI artifacts from sanitized capture output.
package packager

import (
	"bytes"
	"errors"
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

var errDecompressedLayerTooLarge = errors.New("packager: decompressed layer exceeds size limit")

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

func decompressZstdLimited(contents []byte, limit int64) ([]byte, error) {
	reader, err := zstd.NewReader(bytes.NewReader(contents))
	if err != nil {
		return nil, fmt.Errorf("packager: open zstd layer: %w", err)
	}
	defer reader.Close()
	decoded, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, fmt.Errorf("packager: decompress layer: %w", err)
	}
	if int64(len(decoded)) > limit {
		return nil, fmt.Errorf("%w (%d bytes)", errDecompressedLayerTooLarge, limit)
	}
	return decoded, nil
}
