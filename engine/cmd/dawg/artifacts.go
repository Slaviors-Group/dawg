package main

import (
	"archive/zip"
	"crypto/rand"
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

	"github.com/Slaviors-Group/dawg/engine/internal/artifact"

	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
	"github.com/spf13/cobra"
)

const (
	artifactCatalogName             = "artifact-catalog.json"
	artifactCatalogVersion          = 2
	maxArchiveSize            int64 = 512 << 20
	maxExtractedSize          int64 = 1 << 30
	maxArchiveEntries               = 10000
	maxCompressionRatio       int64 = 100
	ociImageManifestMediaType       = "application/vnd.oci.image.manifest.v1+json"
	ociImageIndexMediaType          = "application/vnd.oci.image.index.v1+json"
)

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

type artifactCatalog struct {
	Version   int                             `json:"version"`
	Artifacts map[string]artifactCatalogEntry `json:"artifacts"`
}

type artifactCatalogEntry struct {
	InstanceID   string    `json:"instanceId"`
	ArtifactID   string    `json:"artifactId"`
	Origin       string    `json:"origin"`
	AddedAt      time.Time `json:"addedAt"`
	ImportedAt   time.Time `json:"importedAt,omitempty"`
	ImportSource string    `json:"importSource,omitempty"`
	StoredPath   string    `json:"storedPath"`
	TargetURL    string    `json:"targetUrl,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
}

type artifactDiagnosticsSummary struct {
	Total   int `json:"total"`
	Console int `json:"console"`
	Network int `json:"network"`
	Errors  int `json:"errors"`
}

type artifactMetadata struct {
	ID                   string                      `json:"id"`
	Path                 string                      `json:"path"`
	TargetURL            string                      `json:"targetUrl"`
	CreatedAt            time.Time                   `json:"createdAt"`
	Title                string                      `json:"title"`
	SchemaVersion        string                      `json:"schemaVersion"`
	Origin               string                      `json:"origin"`
	AcquisitionOrigin    string                      `json:"acquisitionOrigin"`
	Status               string                      `json:"status"`
	Components           []string                    `json:"components,omitempty"`
	InstanceID           string                      `json:"instanceId"`
	AddedAt              time.Time                   `json:"addedAt"`
	Flagged              bool                        `json:"flagged"`
	FlagCount            int                         `json:"flagCount"`
	FlagTitles           []string                    `json:"flagTitles"`
	Revision             int                         `json:"revision,omitempty"`
	ReviewRootID         string                      `json:"reviewRootId,omitempty"`
	RootArtifactID       string                      `json:"rootArtifactId,omitempty"`
	SupersedesArtifactID string                      `json:"supersedesArtifactId,omitempty"`
	Tags                 []string                    `json:"tags,omitempty"`
	DiagnosticsSummary   *artifactDiagnosticsSummary `json:"diagnosticsSummary,omitempty"`
}

type deleteArtifactResult struct {
	Status     string `json:"status"`
	InstanceID string `json:"instanceId"`
	Path       string `json:"path"`
}

func newArtifactsCommand() *cobra.Command {
	command := &cobra.Command{Use: "artifacts", Short: "Manage local DAWG artifacts"}
	command.AddCommand(newArtifactsListCommand())
	command.AddCommand(newArtifactsExportCommand())
	command.AddCommand(newArtifactsImportCommand())
	command.AddCommand(newArtifactsDeleteCommand())
	command.AddCommand(newArtifactsReviewCommand())
	return command
}

func newArtifactsListCommand() *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List validated local DAWG artifacts", Args: cobra.NoArgs, RunE: func(command *cobra.Command, _ []string) error {
		artifacts, err := listArtifacts()
		if err != nil {
			return err
		}
		return json.NewEncoder(command.OutOrStdout()).Encode(artifacts)
	}}
}

func newArtifactsExportCommand() *cobra.Command {
	var output string
	var force bool
	command := &cobra.Command{Use: "export <artifact-directory>", Short: "Export a DAWG OCI artifact as a portable ZIP", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, arguments []string) error {
		if output == "" {
			return fmt.Errorf("artifacts export: --output is required")
		}
		if _, err := artifact.Validate(arguments[0]); err != nil {
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
	}}
	command.Flags().StringVar(&output, "output", "", "Portable .dawg ZIP file")
	command.Flags().BoolVar(&force, "force", false, "Replace an existing output file")
	return command
}

func newArtifactsImportCommand() *cobra.Command {
	return &cobra.Command{Use: "import <file.dawg>", Short: "Import a portable DAWG OCI artifact", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, arguments []string) error {
		metadata, err := importArtifact(arguments[0])
		if err != nil {
			return err
		}
		return json.NewEncoder(command.OutOrStdout()).Encode(metadata)
	}}
}

func newArtifactsDeleteCommand() *cobra.Command {
	return &cobra.Command{Use: "delete <artifact-directory>", Short: "Delete one local artifact instance", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, arguments []string) error {
		result, err := deleteArtifact(arguments[0])
		if err != nil {
			return err
		}
		return json.NewEncoder(command.OutOrStdout()).Encode(result)
	}}
}

func newArtifactsReviewCommand() *cobra.Command {
	var reviewFile string
	command := &cobra.Command{Use: "review <artifact-directory>", Short: "Publish an immutable flagged artifact revision", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, arguments []string) error {
		if reviewFile == "" {
			return fmt.Errorf("artifacts review: --review-file is required")
		}
		metadata, err := reviewArtifact(arguments[0], reviewFile)
		if err != nil {
			return err
		}
		return json.NewEncoder(command.OutOrStdout()).Encode(metadata)
	}}
	command.Flags().StringVar(&reviewFile, "review-file", "", "JSON review metadata or review envelope")
	return command
}

func listArtifacts() ([]artifactMetadata, error) {
	unlock, err := lockArtifactStore()
	if err != nil {
		return nil, err
	}
	defer unlock()
	root, err := artifactRoot()
	if err != nil {
		return nil, err
	}
	catalog, err := readArtifactCatalog()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("artifacts list: read artifact directory: %w", err)
	}
	result := make([]artifactMetadata, 0, len(entries))
	changed := false
	for _, entry := range entries {
		if !entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 {
			continue
		}
		path := filepath.Join(root, entry.Name())
		value, err := artifact.Validate(path)
		if err != nil {
			continue
		}
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("artifacts list: resolve artifact path: %w", err)
		}
		catalogEntry, found := catalog.entryForPath(absolutePath)
		if !found {
			catalogEntry, err = newCatalogEntry(value.ID, "legacy", absolutePath, "")
			if err != nil {
				return nil, err
			}
			catalogEntry.AddedAt = value.CreatedAt.UTC()
			catalog.Artifacts[catalogEntry.InstanceID] = catalogEntry
			changed = true
		}
		if catalogEntry.ArtifactID != value.ID || catalogEntry.Origin == "" || catalogEntry.AddedAt.IsZero() {
			catalogEntry.ArtifactID = value.ID
			if catalogEntry.Origin == "" {
				catalogEntry.Origin = "legacy"
			}
			if catalogEntry.AddedAt.IsZero() {
				catalogEntry.AddedAt = value.CreatedAt.UTC()
			}
			catalog.Artifacts[catalogEntry.InstanceID] = catalogEntry
			changed = true
		}
		result = append(result, artifactMetadataFrom(value, absolutePath, catalogEntry))
	}
	if changed {
		if err := writeArtifactCatalog(catalog); err != nil {
			return nil, err
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].AddedAt.After(result[j].AddedAt) })
	return result, nil
}

func artifactMetadataFrom(value manifest.Manifest, path string, entry artifactCatalogEntry) artifactMetadata {
	components := make([]string, 0, len(value.Layers))
	for _, layer := range value.Layers {
		components = append(components, string(layer.MediaType))
	}
	metadata := artifactMetadata{ID: value.ID, Path: path, TargetURL: entry.TargetURL, CreatedAt: value.CreatedAt, Title: value.Title, SchemaVersion: value.SchemaVersion, Origin: entry.Origin, AcquisitionOrigin: entry.Origin, Status: "valid", Components: components, InstanceID: entry.InstanceID, AddedAt: entry.AddedAt, FlagTitles: []string{}, Tags: append([]string{}, entry.Tags...)}
	if value.Diagnostics != nil {
		metadata.DiagnosticsSummary = &artifactDiagnosticsSummary{Console: value.Diagnostics.ConsoleEvents, Network: value.Diagnostics.NetworkEvents, Errors: value.Diagnostics.ErrorEvents}
		metadata.DiagnosticsSummary.Total = metadata.DiagnosticsSummary.Console + metadata.DiagnosticsSummary.Network + metadata.DiagnosticsSummary.Errors
	}
	if value.Review != nil {
		metadata.Flagged = true
		metadata.FlagCount = len(value.Review.Flags)
		metadata.Revision = value.Review.Revision
		metadata.ReviewRootID = value.Review.RootArtifactID
		metadata.RootArtifactID = value.Review.RootArtifactID
		metadata.SupersedesArtifactID = value.Review.ParentArtifactID
		metadata.FlagTitles = make([]string, 0, len(value.Review.Flags))
		for _, flag := range value.Review.Flags {
			metadata.FlagTitles = append(metadata.FlagTitles, flag.Title)
		}
	}
	return metadata
}

func artifactRoot() (string, error) {
	root, err := filepath.Abs(defaultDawgDir("artifacts"))
	if err != nil {
		return "", fmt.Errorf("artifacts: resolve artifact directory: %w", err)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", fmt.Errorf("artifacts: create artifact directory: %w", err)
	}
	return root, nil
}

func lockArtifactStore() (func(), error) {
	if err := os.MkdirAll(defaultDawgDir(), 0o700); err != nil {
		return nil, fmt.Errorf("artifacts: create state directory: %w", err)
	}
	path := defaultDawgDir(".artifact-store.lock")
	deadline := time.Now().Add(10 * time.Second)
	for {
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			if closeErr := file.Close(); closeErr != nil {
				_ = os.Remove(path)
				return nil, fmt.Errorf("artifacts: close store lock: %w", closeErr)
			}
			return func() { _ = os.Remove(path) }, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("artifacts: acquire store lock: %w", err)
		}
		if info, statErr := os.Stat(path); statErr == nil && time.Since(info.ModTime()) > 2*time.Minute {
			_ = os.Remove(path)
			continue
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("artifacts: timed out waiting for store lock")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (catalog artifactCatalog) entryForPath(path string) (artifactCatalogEntry, bool) {
	for _, entry := range catalog.Artifacts {
		if entry.StoredPath == path {
			return entry, true
		}
	}
	return artifactCatalogEntry{}, false
}

func readArtifactCatalog() (artifactCatalog, error) {
	path := defaultDawgDir(artifactCatalogName)
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return artifactCatalog{Version: artifactCatalogVersion, Artifacts: map[string]artifactCatalogEntry{}}, nil
	}
	if err != nil {
		return artifactCatalog{}, fmt.Errorf("artifacts: read catalog: %w", err)
	}
	var raw struct {
		Version   int                             `json:"version"`
		Artifacts map[string]artifactCatalogEntry `json:"artifacts"`
	}
	if err := json.Unmarshal(contents, &raw); err != nil {
		return artifactCatalog{}, fmt.Errorf("artifacts: decode catalog: %w", err)
	}
	if raw.Version != 0 && raw.Version != artifactCatalogVersion {
		return artifactCatalog{}, fmt.Errorf("artifacts: unsupported catalog version %d", raw.Version)
	}
	catalog := artifactCatalog{Version: artifactCatalogVersion, Artifacts: map[string]artifactCatalogEntry{}}
	migrated := raw.Version == 0
	for legacyID, entry := range raw.Artifacts {
		if entry.InstanceID == "" {
			instanceID, idErr := newInstanceID()
			if idErr != nil {
				return artifactCatalog{}, idErr
			}
			entry.InstanceID = instanceID
		}
		if entry.ArtifactID == "" {
			entry.ArtifactID = legacyID
		}
		if entry.Origin == "" {
			entry.Origin = "legacy"
		}
		if entry.AddedAt.IsZero() {
			entry.AddedAt = entry.ImportedAt.UTC()
			if entry.AddedAt.IsZero() {
				entry.AddedAt = time.Now().UTC()
			}
		}
		if entry.StoredPath != "" {
			absolutePath, pathErr := filepath.Abs(entry.StoredPath)
			if pathErr != nil {
				return artifactCatalog{}, fmt.Errorf("artifacts: resolve catalog path: %w", pathErr)
			}
			entry.StoredPath = absolutePath
		}
		catalog.Artifacts[entry.InstanceID] = entry
	}
	if migrated {
		if err := writeArtifactCatalog(catalog); err != nil {
			return artifactCatalog{}, err
		}
	}
	return catalog, nil
}

func writeArtifactCatalog(catalog artifactCatalog) error {
	catalog.Version = artifactCatalogVersion
	if catalog.Artifacts == nil {
		catalog.Artifacts = map[string]artifactCatalogEntry{}
	}
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
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("artifacts: secure catalog: %w", err)
	}
	if _, err := temporary.Write(append(contents, '\n')); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("artifacts: write catalog: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("artifacts: sync catalog: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("artifacts: close catalog: %w", err)
	}
	if err := os.Rename(temporaryPath, defaultDawgDir(artifactCatalogName)); err != nil {
		return fmt.Errorf("artifacts: publish catalog: %w", err)
	}
	return nil
}

func newInstanceID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("artifacts: generate instance ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func newCatalogEntry(artifactID, origin, storedPath, targetURL string) (artifactCatalogEntry, error) {
	instanceID, err := newInstanceID()
	if err != nil {
		return artifactCatalogEntry{}, err
	}
	return artifactCatalogEntry{InstanceID: instanceID, ArtifactID: artifactID, Origin: origin, AddedAt: time.Now().UTC(), StoredPath: storedPath, TargetURL: targetURL}, nil
}

func registerCapturedArtifact(directory, targetURL string) error {
	unlock, err := lockArtifactStore()
	if err != nil {
		return err
	}
	defer unlock()
	value, err := artifact.Validate(directory)
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
	entry, err := newCatalogEntry(value.ID, "captured", path, targetURL)
	if err != nil {
		return err
	}
	catalog.Artifacts[entry.InstanceID] = entry
	return writeArtifactCatalog(catalog)
}

func deleteArtifact(requestedPath string) (deleteArtifactResult, error) {
	unlock, err := lockArtifactStore()
	if err != nil {
		return deleteArtifactResult{}, err
	}
	defer unlock()
	root, err := artifactRoot()
	if err != nil {
		return deleteArtifactResult{}, err
	}
	path, err := filepath.Abs(requestedPath)
	if err != nil {
		return deleteArtifactResult{}, fmt.Errorf("artifacts delete: resolve artifact path: %w", err)
	}
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == "." || filepath.Dir(path) != root || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) || relative == ".." {
		return deleteArtifactResult{}, fmt.Errorf("artifacts delete: only direct local artifact instances may be deleted")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return deleteArtifactResult{}, fmt.Errorf("artifacts delete: inspect artifact: %w", err)
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return deleteArtifactResult{}, fmt.Errorf("artifacts delete: artifact path is not a real directory")
	}
	catalog, err := readArtifactCatalog()
	if err != nil {
		return deleteArtifactResult{}, err
	}
	entry, found := catalog.entryForPath(path)
	if !found {
		return deleteArtifactResult{}, fmt.Errorf("artifacts delete: artifact is not a managed local instance")
	}
	if _, err := artifact.Validate(path); err != nil {
		return deleteArtifactResult{}, fmt.Errorf("artifacts delete: validate artifact: %w", err)
	}
	quarantineRoot := defaultDawgDir("quarantine")
	if err := os.MkdirAll(quarantineRoot, 0o700); err != nil {
		return deleteArtifactResult{}, fmt.Errorf("artifacts delete: create quarantine: %w", err)
	}
	quarantine := filepath.Join(quarantineRoot, entry.InstanceID)
	if _, err := os.Lstat(quarantine); err == nil {
		return deleteArtifactResult{}, fmt.Errorf("artifacts delete: quarantine entry already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return deleteArtifactResult{}, fmt.Errorf("artifacts delete: inspect quarantine: %w", err)
	}
	if err := os.Rename(path, quarantine); err != nil {
		return deleteArtifactResult{}, fmt.Errorf("artifacts delete: quarantine artifact: %w", err)
	}
	delete(catalog.Artifacts, entry.InstanceID)
	if err := writeArtifactCatalog(catalog); err != nil {
		if rollbackErr := os.Rename(quarantine, path); rollbackErr != nil {
			return deleteArtifactResult{}, fmt.Errorf("artifacts delete: publish catalog: %w; restore artifact: %v", err, rollbackErr)
		}
		return deleteArtifactResult{}, err
	}
	if err := os.RemoveAll(quarantine); err != nil {
		return deleteArtifactResult{}, fmt.Errorf("artifacts delete: remove quarantined artifact: %w", err)
	}
	return deleteArtifactResult{Status: "deleted", InstanceID: entry.InstanceID, Path: path}, nil
}

func reviewArtifact(sourcePath, reviewFile string) (artifactMetadata, error) {
	source, err := artifact.Validate(sourcePath)
	if err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts review: validate source artifact: %w", err)
	}
	review, err := readReviewFile(reviewFile, source)
	if err != nil {
		return artifactMetadata{}, err
	}
	return publishReviewArtifact(sourcePath, source, review, defaultReviewArtifactTitle(source.Title, review.Revision))
}

func defaultReviewArtifactTitle(sourceTitle string, revision int) string {
	return fmt.Sprintf("%s review %d", sourceTitle, revision)
}

func publishReviewArtifact(sourcePath string, source manifest.Manifest, review manifest.Review, artifactTitle string) (artifactMetadata, error) {
	unlock, err := lockArtifactStore()
	if err != nil {
		return artifactMetadata{}, err
	}
	defer unlock()
	expected := artifact.NewReview(source, append([]manifest.ReviewFlag(nil), review.Flags...), review.ReviewedAt)
	if review.FormatVersion != expected.FormatVersion || review.Kind != expected.Kind || review.RootArtifactID != expected.RootArtifactID || review.ParentArtifactID != expected.ParentArtifactID || review.Revision != expected.Revision {
		return artifactMetadata{}, fmt.Errorf("artifacts review: review lineage does not match source artifact")
	}
	root, err := artifactRoot()
	if err != nil {
		return artifactMetadata{}, err
	}
	output, err := uniqueArtifactPath(root, artifactDirectoryNameForTitle(artifactTitle))
	if err != nil {
		return artifactMetadata{}, err
	}
	packaged, derived, err := artifact.DeriveReview(artifact.DeriveReviewRequest{InputDirectory: sourcePath, OutputDirectory: output, ArtifactTitle: artifactTitle, Review: review})
	if err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts review: publish revision: %w", err)
	}
	catalog, err := readArtifactCatalog()
	if err != nil {
		_ = os.RemoveAll(packaged.Directory)
		return artifactMetadata{}, err
	}
	absoluteSource, _ := filepath.Abs(sourcePath)
	sourceEntry, found := catalog.entryForPath(absoluteSource)
	origin, targetURL := "legacy", ""
	if found {
		origin, targetURL = sourceEntry.Origin, sourceEntry.TargetURL
	}
	entry, err := newCatalogEntry(derived.ID, origin, packaged.Directory, targetURL)
	if err != nil {
		_ = os.RemoveAll(packaged.Directory)
		return artifactMetadata{}, err
	}
	entry.Tags = []string{"reviewed"}
	catalog.Artifacts[entry.InstanceID] = entry
	if err := writeArtifactCatalog(catalog); err != nil {
		_ = os.RemoveAll(packaged.Directory)
		return artifactMetadata{}, err
	}
	return artifactMetadataFrom(derived, packaged.Directory, entry), nil
}

func readReviewFile(path string, source manifest.Manifest) (manifest.Review, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return manifest.Review{}, fmt.Errorf("artifacts review: read review file: %w", err)
	}
	var envelope struct {
		Review *manifest.Review      `json:"review"`
		Flags  []manifest.ReviewFlag `json:"flags"`
	}
	if err := json.Unmarshal(contents, &envelope); err != nil {
		return manifest.Review{}, fmt.Errorf("artifacts review: decode review file: %w", err)
	}
	requested := envelope.Review
	flags := envelope.Flags
	if requested != nil {
		flags = requested.Flags
	}
	if len(flags) == 0 {
		return manifest.Review{}, fmt.Errorf("artifacts review: review file must contain at least one flag")
	}
	reviewedAt := time.Now().UTC()
	if requested != nil && !requested.ReviewedAt.IsZero() {
		reviewedAt = requested.ReviewedAt.UTC()
	}
	review := artifact.NewReview(source, append([]manifest.ReviewFlag(nil), flags...), reviewedAt)
	return review, nil
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
	defer file.Close()
	writer := zip.NewWriter(file)
	err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
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
		header.Name, header.Method = filepath.ToSlash(relative), zip.Deflate
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
		return firstError(copyErr, closeErr)
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
	unlock, err := lockArtifactStore()
	if err != nil {
		return artifactMetadata{}, err
	}
	defer unlock()
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
	root, err := artifactRoot()
	if err != nil {
		return artifactMetadata{}, err
	}
	staging, err := os.MkdirTemp(root, ".dawg-import-")
	if err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts import: create staging directory: %w", err)
	}
	defer os.RemoveAll(staging)
	if err := extractArchive(archive.File, staging); err != nil {
		return artifactMetadata{}, err
	}
	value, err := artifact.Validate(staging)
	if err != nil {
		return artifactMetadata{}, fmt.Errorf("artifacts import: validate archive: %w", err)
	}
	destination, err := uniqueArtifactPath(root, artifactDirectoryNameForTitle(value.Title))
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
	catalog, err := readArtifactCatalog()
	if err != nil {
		_ = os.RemoveAll(destination)
		return artifactMetadata{}, err
	}
	entry, err := newCatalogEntry(value.ID, "imported", destination, "")
	if err != nil {
		_ = os.RemoveAll(destination)
		return artifactMetadata{}, err
	}
	entry.ImportSource = absoluteSource
	entry.ImportedAt = entry.AddedAt
	catalog.Artifacts[entry.InstanceID] = entry
	if err := writeArtifactCatalog(catalog); err != nil {
		_ = os.RemoveAll(destination)
		return artifactMetadata{}, err
	}
	return artifactMetadataFrom(value, destination, entry), nil
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
		if file.Mode()&os.ModeSymlink != 0 || !safeArchivePath(file.Name) {
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
		closeErr := firstError(out.Close(), in.Close())
		if copyErr != nil || closeErr != nil {
			return fmt.Errorf("artifacts import: extract entry %s: %v", file.Name, firstError(copyErr, closeErr))
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
			return "", fmt.Errorf("artifacts: inspect destination: %w", err)
		}
	}
}

func validateOCIArtifact(directory string) (manifest.Manifest, error) {
	return artifact.Validate(directory)
}
