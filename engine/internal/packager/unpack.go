package packager

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
)

// Unpack unrolls an OCI Image Layout artifact into a target directory.
func Unpack(layoutDirectory, targetDirectory string) error {
	manifestPath := filepath.Join(layoutDirectory, "dawg-manifest.json")
	contents, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("packager: read dawg manifest %s: %w", manifestPath, err)
	}

	manifestValue, err := manifest.Unmarshal(contents)
	if err != nil {
		return fmt.Errorf("packager: parse dawg manifest: %w", err)
	}

	if err := os.MkdirAll(targetDirectory, 0o700); err != nil {
		return fmt.Errorf("packager: create unpack target %s: %w", targetDirectory, err)
	}

	for _, layer := range manifestValue.Layers {
		digestPath, err := layerDigestPath(layer.Digest)
		if err != nil {
			return fmt.Errorf("packager: parse layer digest: %w", err)
		}

		blobPath := filepath.Join(layoutDirectory, "blobs", "sha256", digestPath)
		blobContents, err := os.ReadFile(blobPath)
		if err != nil {
			return fmt.Errorf("packager: read blob %s: %w", blobPath, err)
		}

		if err := unpackLayer(layer.MediaType, blobContents, targetDirectory); err != nil {
			return fmt.Errorf("packager: unpack layer %s: %w", layer.Digest, err)
		}
	}
	return nil
}

func unpackLayer(mediaType dawgtypes.MediaType, contents []byte, targetDirectory string) error {
	if strings.HasSuffix(string(mediaType), "+tar.zstd") || strings.HasSuffix(string(mediaType), "+jsonl.zstd") {
		decompressed, err := decompressZstd(contents)
		if err != nil {
			return err
		}
		contents = decompressed
	}

	if !strings.Contains(string(mediaType), "+tar") {
		// Not an archive, nothing to unpack directly into filesystem structure.
		// Usually trace layer is jsonl.zstd, wait, in contracts.md:
		// "application/vnd.dawg.trace.rrweb+jsonl.zstd" -> wait, trace layer includes multiple files in our packager?
		// Let's check layers.go: trace layer has directories []string{"traces", "http", "logs"}
		// and uses archiveDirectories. Wait, archiveDirectories creates a tar!
		// BUT the media type is application/vnd.dawg.trace.rrweb+jsonl.zstd. That's a naming mismatch in architecture vs implementation.
		// Actually, if it was tarred, it must be untarred.
	}

	// For MVP, all our packaged capture data is actually tarred.
	// layers.go: archiveDirectories is called for all layers.
	reader := tar.NewReader(bytes.NewReader(contents))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		// Prevent path traversal
		if strings.Contains(header.Name, "..") {
			return fmt.Errorf("refuse unsafe tar path %s", header.Name)
		}

		targetPath := filepath.Join(targetDirectory, header.Name)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o700); err != nil {
			return fmt.Errorf("create directory %s: %w", filepath.Dir(targetPath), err)
		}

		if header.Typeflag == tar.TypeReg {
			f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create file %s: %w", targetPath, err)
			}
			if _, err := io.Copy(f, reader); err != nil {
				f.Close()
				return fmt.Errorf("write file %s: %w", targetPath, err)
			}
			f.Close()
		}
	}
	return nil
}
