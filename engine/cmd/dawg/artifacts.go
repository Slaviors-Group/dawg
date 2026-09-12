package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
	"github.com/spf13/cobra"
)

const (
	artifactCatalogName             = "artifact-catalog.json"
	maxArchiveSize            int64 = 512 << 20
	maxExtractedSize          int64 = 1 << 30
	maxArchiveEntries               = 10000
	maxCompressionRatio       int64 = 100
	ociImageManifestMediaType       = "application/vnd.oci.image.manifest.v1+json"
	ociImageIndexMediaType          = "application/vnd.oci.image.index.v1+json"
)

type artifactCatalog struct {
	Artifacts map[string]artifactCatalogEntry `json:"artifacts"`
}

type artifactCatalogEntry struct {
	Origin       string    `json:"origin"`
	ImportedAt   time.Time `json:"importedAt,omitempty"`
	ImportSource string    `json:"importSource,omitempty"`
	StoredPath   string    `json:"storedPath"`
	TargetURL    string    `json:"targetUrl,omitempty"`
}

type artifactMetadata struct {
	ID            string    `json:"id"`
	Path          string    `json:"path"`
	TargetURL     string    `json:"targetUrl"`
	CreatedAt     time.Time `json:"createdAt"`
	Title         string    `json:"title"`
	SchemaVersion string    `json:"schemaVersion"`
	Origin        string    `json:"origin"`
	Status        string    `json:"status"`
	Components    []string  `json:"components,omitempty"`
}

type ociDescriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

type ociIndex struct {
	SchemaVersion int             `json:"schemaVersion"`
	MediaType     string          `json:"mediaType"`
	Manifests     []ociDescriptor `json:"manifests"`
}

type ociImageManifest struct {
	SchemaVersion int             `json:"schemaVersion"`
	MediaType     string          `json:"mediaType"`
	Config        ociDescriptor   `json:"config"`
	Layers        []ociDescriptor `json:"layers"`
}

func newArtifactsCommand() *cobra.Command {
	command := &cobra.Command{Use: "artifacts", Short: "Manage local DAWG artifacts"}
	command.AddCommand(newArtifactsListCommand())
	command.AddCommand(newArtifactsExportCommand())
	command.AddCommand(newArtifactsImportCommand())
	return command
}

func newArtifactsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List validated local DAWG artifacts",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			artifacts, err := listArtifacts()
			if err != nil {
				return err
			}
			return json.NewEncoder(command.OutOrStdout()).Encode(artifacts)
		},
	}
}

func newArtifactsExportCommand() *cobra.Command {
	var output string
	var force bool
	command := &cobra.Command{
		Use:   "export <artifact-directory>",
		Short: "Export a DAWG OCI artifact as a portable ZIP",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			if output == "" {
				return fmt.Errorf("artifacts export: --output is required")
			}
			if _, err := validateOCIArtifact(arguments[0]); err != nil {
				return fmt.Errorf("artifacts export: %w", err)
			}
			outputPath, err := filepath.Abs(output)
			if err != nil {
				return fmt.Errorf("artifacts export: resolve output: %w", err)
			}
			if force {
				if err := os.Remove(outputPath); err != nil && !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("artifacts export: replace output: %w", err)
				}
			}
			if err := exportArtifact(arguments[0], outputPath); err != nil {
				return err
			}
			return json.NewEncoder(command.OutOrStdout()).Encode(map[string]string{"status": "exported", "output": outputPath})
		},
	}
	command.Flags().StringVar(&output, "output", "", "Portable .dawg ZIP file")
	command.Flags().BoolVar(&force, "force", false, "Replace an existing output file")
	return command
}

func newArtifactsImportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import <file.dawg>",
		Short: "Import a portable DAWG OCI artifact",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			metadata, err := importArtifact(arguments[0])
			if err != nil {
				return err
			}
			return json.NewEncoder(command.OutOrStdout()).Encode(metadata)
		},
	}
}

func listArtifacts() ([]artifactMetadata, error) {
	root := defaultDawgDir("artifacts")
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("artifacts list: create artifact directory: %w", err)
	}
	catalog, err := readArtifactCatalog()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("artifacts list: read artifact directory: %w", err)
	}
	result := make([]artifactMetadata, 0)
	changed := false
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name())
		value, err := validateOCIArtifact(path)
		if err != nil {
			continue
		}
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("artifacts list: resolve artifact path: %w", err)
		}
		catalogEntry, ok := catalog.Artifacts[value.ID]
		if !ok {
			catalogEntry = artifactCatalogEntry{Origin: "legacy", StoredPath: absolutePath}
			catalog.Artifacts[value.ID] = catalogEntry
			changed = true
		} else if catalogEntry.Origin == "" {
			catalogEntry.Origin = "legacy"
			catalogEntry.StoredPath = absolutePath
			catalog.Artifacts[value.ID] = catalogEntry
			changed = true
		}
		result = append(result, artifactMetadataFrom(value, absolutePath, catalogEntry))
	}
	if changed {
		if err := writeArtifactCatalog(catalog); err != nil {
			return nil, err
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func artifactMetadataFrom(value manifest.Manifest, path string, entry artifactCatalogEntry) artifactMetadata {
	components := make([]string, 0, len(value.Layers))
	for _, layer := range value.Layers {
		components = append(components, string(layer.MediaType))
	}
	return artifactMetadata{ID: value.ID, Path: path, TargetURL: entry.TargetURL, CreatedAt: value.CreatedAt, Title: value.Title, SchemaVersion: value.SchemaVersion, Origin: entry.Origin, Status: "valid", Components: components}
}

func readArtifactCatalog() (artifactCatalog, error) {
	path := defaultDawgDir(artifactCatalogName)
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return artifactCatalog{Artifacts: map[string]artifactCatalogEntry{}}, nil
	}
	if err != nil {
		return artifactCatalog{}, fmt.Errorf("artifacts: read catalog: %w", err)
	}
	var catalog artifactCatalog
	if err := json.Unmarshal(contents, &catalog); err != nil {
		return artifactCatalog{}, fmt.Errorf("artifacts: decode catalog: %w", err)
	}
	if catalog.Artifacts == nil {
		catalog.Artifacts = map[string]artifactCatalogEntry{}
	}
	return catalog, nil
}

func writeArtifactCatalog(catalog artifactCatalog) error {
	if err := os.MkdirAll(defaultDawgDir(), 0o700); err != nil {
		return fmt.Errorf("artifacts: create catalog directory: %w", err)
	}
	contents, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("artifacts: encode catalog: %w", err)
	}
	temporary, err := os.CreateTemp(defaultDawgDir(), ".artifact-catalog-")
	if err != nil {
		return fmt.Errorf("artifacts: create temporary catalog: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(append(contents, '\n')); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("artifacts: write catalog: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("artifacts: close catalog: %w", err)
	}
	if err := os.Rename(temporaryPath, defaultDawgDir(artifactCatalogName)); err != nil {
		return fmt.Errorf("artifacts: publish catalog: %w", err)
	}
	return nil
}

func registerCapturedArtifact(directory, targetURL string) error {
	value, err := validateOCIArtifact(directory)
	if err != nil {
		return err
	}
	path, err := filepath.Abs(directory)
	if err != nil {
		return fmt.Errorf("artifacts: resolve captured artifact path: %w", err)
	}
	catalog, err := readArtifactCatalog()
	if err != nil {
		return err
	}
	catalog.Artifacts[value.ID] = artifactCatalogEntry{Origin: "captured", StoredPath: path, TargetURL: targetURL}
	return writeArtifactCatalog(catalog)
}

func exportArtifact(directory, output string) error {
	if _, err := os.Stat(output); err == nil {
		return fmt.Errorf("artifacts export: output already exists: %s", output)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("artifacts export: inspect output: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o700); err != nil {
		return fmt.Errorf("artifacts export: create output directory: %w", err)
	}
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("artifacts export: create output: %w", err)
	}
	defer func() { _ = file.Close() }()
	writer := zip.NewWriter(file)
	err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("artifacts export: symbolic links are not supported: %s", path)
		}
		relative, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relative)
		header.Method = zip.Deflate
		out, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := in.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if err == nil {
		err = writer.Close()
	} else {
		_ = writer.Close()
	}
	if err != nil {
		_ = os.Remove(output)
		return fmt.Errorf("artifacts export: write archive: %w", err)
	}
	return nil
}

func importArtifact(source string) (artifactMetadata, error) {
	info, err := os.Stat(source)
	if err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts import: inspect archive: %w", err)
	}
	if info.Size() > maxArchiveSize {
		return artifactMetadata{}, fmt.Errorf("artifacts import: archive exceeds %d bytes", maxArchiveSize)
	}
	archive, err := zip.OpenReader(source)
	if err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts import: open archive: %w", err)
	}
	defer archive.Close()
	stagingParent := defaultDawgDir("artifacts")
	if err := os.MkdirAll(stagingParent, 0o700); err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts import: create artifact directory: %w", err)
	}
	staging, err := os.MkdirTemp(stagingParent, ".dawg-import-")
	if err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts import: create staging directory: %w", err)
	}
	defer os.RemoveAll(staging)
	if err := extractArchive(archive.File, staging); err != nil {
		return artifactMetadata{}, err
	}
	value, err := validateOCIArtifact(staging)
	if err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts import: validate archive: %w", err)
	}
	destination, err := uniqueArtifactPath(stagingParent, artifactDirectoryNameForTitle(value.Title, value.CreatedAt))
	if err != nil {
		return artifactMetadata{}, err
	}
	if err := os.Rename(staging, destination); err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts import: publish artifact: %w", err)
	}
	absoluteSource, err := filepath.Abs(source)
	if err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts import: resolve archive path: %w", err)
	}
	absoluteDestination, err := filepath.Abs(destination)
	if err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts import: resolve artifact path: %w", err)
	}
	catalog, err := readArtifactCatalog()
	if err != nil {
		return artifactMetadata{}, err
	}
	entry := artifactCatalogEntry{Origin: "imported", ImportedAt: time.Now().UTC(), ImportSource: absoluteSource, StoredPath: absoluteDestination}
	catalog.Artifacts[value.ID] = entry
	if err := writeArtifactCatalog(catalog); err != nil {
		return artifactMetadata{}, err
	}
	return artifactMetadataFrom(value, absoluteDestination, entry), nil
}

func extractArchive(files []*zip.File, destination string) error {
	if len(files) > maxArchiveEntries {
		return fmt.Errorf("artifacts import: archive has too many entries")
	}
	var extracted int64
	for _, file := range files {
		if file.FileInfo().IsDir() {
			continue
		}
		if file.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("artifacts import: symbolic links are not allowed: %s", file.Name)
		}
		if !safeArchivePath(file.Name) {
			return fmt.Errorf("artifacts import: unsafe archive path: %s", file.Name)
		}
		if file.UncompressedSize64 > uint64(maxExtractedSize) || file.CompressedSize64 == 0 && file.UncompressedSize64 > 0 || (file.CompressedSize64 > 0 && file.UncompressedSize64 > uint64(maxCompressionRatio)*file.CompressedSize64) {
			return fmt.Errorf("artifacts import: unsafe compression ratio or file size: %s", file.Name)
		}
		extracted += int64(file.UncompressedSize64)
		if extracted > maxExtractedSize {
			return fmt.Errorf("artifacts import: extracted data exceeds %d bytes", maxExtractedSize)
		}
		path := filepath.Join(destination, filepath.FromSlash(file.Name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return fmt.Errorf("artifacts import: create extraction directory: %w", err)
		}
		in, err := file.Open()
		if err != nil {
			return fmt.Errorf("artifacts import: open entry %s: %w", file.Name, err)
		}
		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			_ = in.Close()
			return fmt.Errorf("artifacts import: create entry %s: %w", file.Name, err)
		}
		_, copyErr := io.Copy(out, io.LimitReader(in, int64(file.UncompressedSize64)+1))
		closeOutErr, closeInErr := out.Close(), in.Close()
		if copyErr != nil || closeOutErr != nil || closeInErr != nil {
			return fmt.Errorf("artifacts import: extract entry %s: %v", file.Name, firstError(copyErr, closeOutErr, closeInErr))
		}
	}
	return nil
}

func firstError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
func safeArchivePath(name string) bool {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(name)))
	return name != "" && !strings.HasPrefix(name, "/") && !strings.Contains(name, "\\") && clean != "." && !strings.HasPrefix(clean, "../") && clean != ".."
}

func uniqueArtifactPath(parent, name string) (string, error) {
	for i := 0; ; i++ {
		suffix := ""
		if i > 0 {
			suffix = fmt.Sprintf("-%d", i)
		}
		path := filepath.Join(parent, name+suffix)
		_, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			return path, nil
		}
		if err != nil {
			return "", fmt.Errorf("artifacts import: inspect destination: %w", err)
		}
	}
}

func validateOCIArtifact(directory string) (manifest.Manifest, error) {
	value, err := inspectArtifact(directory)
	if err != nil {
		return manifest.Manifest{}, err
	}
	layoutContents, err := os.ReadFile(filepath.Join(directory, "oci-layout"))
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("read OCI layout: %w", err)
	}
	var layout struct {
		ImageLayoutVersion string `json:"imageLayoutVersion"`
	}
	if err := json.Unmarshal(layoutContents, &layout); err != nil || layout.ImageLayoutVersion != "1.0.0" {
		return manifest.Manifest{}, fmt.Errorf("invalid OCI layout")
	}
	var index ociIndex
	if err := readJSON(filepath.Join(directory, "index.json"), &index); err != nil {
		return manifest.Manifest{}, fmt.Errorf("read OCI index: %w", err)
	}
	if index.SchemaVersion != 2 || index.MediaType != ociImageIndexMediaType || len(index.Manifests) != 1 {
		return manifest.Manifest{}, fmt.Errorf("invalid OCI index")
	}
	manifestContents, err := verifyDescriptor(directory, index.Manifests[0])
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("validate OCI manifest descriptor: %w", err)
	}
	var image ociImageManifest
	if err := json.Unmarshal(manifestContents, &image); err != nil {
		return manifest.Manifest{}, fmt.Errorf("decode OCI manifest: %w", err)
	}
	if image.SchemaVersion != 2 || image.MediaType != ociImageManifestMediaType {
		return manifest.Manifest{}, fmt.Errorf("invalid OCI manifest")
	}
	if image.Config.MediaType != string(dawgtypes.MediaTypeManifest) {
		return manifest.Manifest{}, fmt.Errorf("invalid OCI config media type")
	}
	config, err := verifyDescriptor(directory, image.Config)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("validate OCI config descriptor: %w", err)
	}
	if string(config) != mustManifestJSON(value) {
		return manifest.Manifest{}, fmt.Errorf("OCI config does not match dawg-manifest.json")
	}
	if len(image.Layers) != len(value.Layers) {
		return manifest.Manifest{}, fmt.Errorf("OCI layer descriptors do not match DAWG manifest")
	}
	for index, descriptor := range image.Layers {
		layer := value.Layers[index]
		if descriptor.MediaType != string(layer.MediaType) || descriptor.Digest != layer.Digest || descriptor.Size != layer.Size {
			return manifest.Manifest{}, fmt.Errorf("OCI layer descriptor %d does not match DAWG manifest", index)
		}
		if _, err := verifyDescriptor(directory, descriptor); err != nil {
			return manifest.Manifest{}, fmt.Errorf("validate OCI layer descriptor: %w", err)
		}
	}
	blobRoot := filepath.Join(directory, "blobs", "sha256")
	entries, err := os.ReadDir(blobRoot)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("read OCI blobs: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || len(entry.Name()) != 64 {
			return manifest.Manifest{}, fmt.Errorf("invalid OCI blob entry: %s", entry.Name())
		}
		info, err := entry.Info()
		if err != nil {
			return manifest.Manifest{}, fmt.Errorf("inspect OCI blob %s: %w", entry.Name(), err)
		}
		if _, err := verifyBlob(directory, "sha256:"+entry.Name(), info.Size()); err != nil {
			return manifest.Manifest{}, err
		}
	}
	return value, nil
}

func mustManifestJSON(value manifest.Manifest) string {
	contents, err := manifest.Marshal(value)
	if err != nil {
		return ""
	}
	return string(contents)
}
func readJSON(path string, target any) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(contents, target)
}
func verifyDescriptor(directory string, descriptor ociDescriptor) ([]byte, error) {
	if descriptor.MediaType == "" || descriptor.Size < 0 {
		return nil, fmt.Errorf("invalid descriptor")
	}
	return verifyBlob(directory, descriptor.Digest, descriptor.Size)
}

func verifyBlob(directory, digest string, size int64) ([]byte, error) {
	if size < 0 || !strings.HasPrefix(digest, "sha256:") || len(digest) != len("sha256:")+64 {
		return nil, fmt.Errorf("invalid descriptor digest %q", digest)
	}
	hexDigest := strings.TrimPrefix(digest, "sha256:")
	if _, err := hex.DecodeString(hexDigest); err != nil {
		return nil, fmt.Errorf("invalid descriptor digest %q", digest)
	}
	contents, err := os.ReadFile(filepath.Join(directory, "blobs", "sha256", hexDigest))
	if err != nil {
		return nil, err
	}
	actual := sha256.Sum256(contents)
	if digest != fmt.Sprintf("sha256:%x", actual) || size != int64(len(contents)) {
		return nil, fmt.Errorf("descriptor digest or size mismatch: %s", digest)
	}
	return contents, nil
}
