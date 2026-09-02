package packager

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

type layerSource struct {
	name        string
	mediaType   dawgtypes.MediaType
	directories []string
	compressed  bool
}

type layerBlob struct {
	spec     dawgtypes.LayerSpec
	contents []byte
}

func buildLayers(sessionDirectory string) ([]layerBlob, error) {
	sources := []layerSource{
		{name: "environment", mediaType: dawgtypes.MediaTypeEnvironment, directories: []string{"env"}},
		{name: "database-fixture", mediaType: dawgtypes.MediaTypeDatabaseFixture, directories: []string{"db"}, compressed: true},
		{name: "trace", mediaType: dawgtypes.MediaTypeTrace, directories: []string{"traces", "http", "logs"}, compressed: true},
		{name: "cassette", mediaType: dawgtypes.MediaTypeCassette, directories: []string{"cassettes"}, compressed: true},
	}

	layers := make([]layerBlob, 0, len(sources))
	for _, source := range sources {
		contents, included, err := archiveDirectories(sessionDirectory, source.directories)
		if err != nil {
			return nil, fmt.Errorf("packager: build %s layer: %w", source.name, err)
		}
		if !included {
			continue
		}
		if source.compressed {
			contents, err = compressZstd(contents)
			if err != nil {
				return nil, fmt.Errorf("packager: compress %s layer: %w", source.name, err)
			}
		}
		layers = append(layers, layerBlob{spec: layerSpec(source.mediaType, contents), contents: contents})
	}
	if len(layers) == 0 {
		return nil, fmt.Errorf("packager: capture directory has no packageable layer content")
	}
	return layers, nil
}

func layerSpec(mediaType dawgtypes.MediaType, contents []byte) dawgtypes.LayerSpec {
	digest := sha256.Sum256(contents)
	return dawgtypes.LayerSpec{
		MediaType: mediaType,
		Digest:    fmt.Sprintf("sha256:%x", digest),
		Size:      int64(len(contents)),
	}
}

func archiveDirectories(sessionDirectory string, directories []string) ([]byte, bool, error) {
	paths := []string{}
	for _, directory := range directories {
		root := filepath.Join(sessionDirectory, filepath.FromSlash(directory))
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, false, fmt.Errorf("inspect %s: %w", root, err)
		}
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			if entry.Type()&fs.ModeSymlink != 0 {
				return fmt.Errorf("refuse symbolic link %s", path)
			}
			if !entry.Type().IsRegular() {
				return fmt.Errorf("refuse non-regular file %s", path)
			}
			paths = append(paths, path)
			return nil
		})
		if err != nil {
			return nil, false, err
		}
	}
	if len(paths) == 0 {
		return nil, false, nil
	}
	sort.Strings(paths)

	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)
	for _, path := range paths {
		relativePath, err := filepath.Rel(sessionDirectory, path)
		if err != nil {
			writer.Close()
			return nil, false, fmt.Errorf("resolve archive path %s: %w", path, err)
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			writer.Close()
			return nil, false, fmt.Errorf("read %s: %w", path, err)
		}
		header := &tar.Header{
			Name:       filepath.ToSlash(relativePath),
			Mode:       0o600,
			Size:       int64(len(contents)),
			ModTime:    time.Unix(0, 0).UTC(),
			AccessTime: time.Time{},
			ChangeTime: time.Time{},
			Format:     tar.FormatPAX,
		}
		if err := writer.WriteHeader(header); err != nil {
			writer.Close()
			return nil, false, fmt.Errorf("write archive header %s: %w", path, err)
		}
		if _, err := writer.Write(contents); err != nil {
			writer.Close()
			return nil, false, fmt.Errorf("write archive contents %s: %w", path, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, false, fmt.Errorf("finish archive: %w", err)
	}
	return buffer.Bytes(), true, nil
}

func layerDigestPath(digest string) (string, error) {
	const prefix = "sha256:"
	if !strings.HasPrefix(digest, prefix) || len(digest) != len(prefix)+64 {
		return "", fmt.Errorf("invalid SHA-256 digest %q", digest)
	}
	return strings.TrimPrefix(digest, prefix), nil
}
