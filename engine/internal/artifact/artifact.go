// Package artifact validates and derives DAWG OCI image layouts.
package artifact

import (
	"archive/tar"
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
	"github.com/klauspost/compress/zstd"
)

const (
	ociImageManifestMediaType = "application/vnd.oci.image.manifest.v1+json"
	ociImageIndexMediaType    = "application/vnd.oci.image.index.v1+json"
)

// Descriptor is an OCI content descriptor.
type Descriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

type imageManifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType"`
	Config        Descriptor   `json:"config"`
	Layers        []Descriptor `json:"layers"`
}

type index struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType"`
	Manifests     []Descriptor `json:"manifests"`
}

// Identity returns the deterministic DAWG logical ID for a manifest.
func Identity(value manifest.Manifest) (string, error) {
	value.ID = ""
	contents, err := manifest.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("serialize logical identity: %w", err)
	}
	digest := sha256.Sum256(contents)
	return fmt.Sprintf("sha256:%x", digest), nil
}

// Validate verifies a complete OCI image layout and its DAWG logical identity.
func Validate(directory string) (manifest.Manifest, error) {
	info, err := os.Lstat(directory)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("inspect artifact directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return manifest.Manifest{}, fmt.Errorf("artifact path is not a real directory")
	}
	value, err := manifest.ReadValidated(filepath.Join(directory, "dawg-manifest.json"))
	if err != nil {
		return manifest.Manifest{}, err
	}
	expectedID, err := Identity(value)
	if err != nil || value.ID != expectedID {
		return manifest.Manifest{}, fmt.Errorf("logical DAWG artifact ID does not match manifest")
	}
	layoutContents, err := os.ReadFile(filepath.Join(directory, "oci-layout"))
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("read OCI layout: %w", err)
	}
	var layout struct {
		ImageLayoutVersion string `json:"imageLayoutVersion"`
	}
	if json.Unmarshal(layoutContents, &layout) != nil || layout.ImageLayoutVersion != "1.0.0" {
		return manifest.Manifest{}, fmt.Errorf("invalid OCI layout")
	}
	var layoutIndex index
	if err := readJSON(filepath.Join(directory, "index.json"), &layoutIndex); err != nil {
		return manifest.Manifest{}, fmt.Errorf("read OCI index: %w", err)
	}
	if layoutIndex.SchemaVersion != 2 || layoutIndex.MediaType != ociImageIndexMediaType || len(layoutIndex.Manifests) != 1 {
		return manifest.Manifest{}, fmt.Errorf("invalid OCI index")
	}
	imageBytes, err := verifyDescriptor(directory, layoutIndex.Manifests[0])
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("validate OCI manifest descriptor: %w", err)
	}
	var image imageManifest
	if err := json.Unmarshal(imageBytes, &image); err != nil {
		return manifest.Manifest{}, fmt.Errorf("decode OCI manifest: %w", err)
	}
	if image.SchemaVersion != 2 || image.MediaType != ociImageManifestMediaType || image.Config.MediaType != string(dawgtypes.MediaTypeManifest) {
		return manifest.Manifest{}, fmt.Errorf("invalid OCI manifest")
	}
	config, err := verifyDescriptor(directory, image.Config)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("validate OCI config descriptor: %w", err)
	}
	manifestJSON, err := manifest.Marshal(value)
	if err != nil || !bytes.Equal(config, manifestJSON) {
		return manifest.Manifest{}, fmt.Errorf("OCI config does not match dawg-manifest.json")
	}
	if len(image.Layers) != len(value.Layers) {
		return manifest.Manifest{}, fmt.Errorf("OCI layer descriptors do not match DAWG manifest")
	}
	for position, descriptor := range image.Layers {
		layer := value.Layers[position]
		if descriptor.MediaType != string(layer.MediaType) || descriptor.Digest != layer.Digest || descriptor.Size != layer.Size {
			return manifest.Manifest{}, fmt.Errorf("OCI layer descriptor %d does not match DAWG manifest", position)
		}
		if _, err := verifyDescriptor(directory, descriptor); err != nil {
			return manifest.Manifest{}, fmt.Errorf("validate OCI layer descriptor: %w", err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(directory, "blobs", "sha256"))
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("read OCI blobs: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || len(entry.Name()) != 64 || entry.Type()&fs.ModeSymlink != 0 {
			return manifest.Manifest{}, fmt.Errorf("invalid OCI blob entry: %s", entry.Name())
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return manifest.Manifest{}, fmt.Errorf("invalid OCI blob entry: %s", entry.Name())
		}
		if _, err := verifyBlob(directory, "sha256:"+entry.Name(), info.Size()); err != nil {
			return manifest.Manifest{}, err
		}
	}
	return value, nil
}

func readJSON(path string, target any) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(contents, target)
}

func verifyDescriptor(directory string, descriptor Descriptor) ([]byte, error) {
	if descriptor.MediaType == "" || descriptor.Size < 0 {
		return nil, fmt.Errorf("invalid descriptor")
	}
	return verifyBlob(directory, descriptor.Digest, descriptor.Size)
}

func verifyBlob(directory, digest string, size int64) ([]byte, error) {
	if size < 0 || !validDigest(digest) {
		return nil, fmt.Errorf("invalid descriptor digest %q", digest)
	}
	path := filepath.Join(directory, "blobs", "sha256", strings.TrimPrefix(digest, "sha256:"))
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("OCI blob is not a regular file")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	actual := sha256.Sum256(contents)
	if digest != fmt.Sprintf("sha256:%x", actual) || size != int64(len(contents)) {
		return nil, fmt.Errorf("descriptor digest or size mismatch: %s", digest)
	}
	return contents, nil
}

func validDigest(digest string) bool {
	if len(digest) != 71 || !strings.HasPrefix(digest, "sha256:") {
		return false
	}
	for _, character := range digest[7:] {
		if !(character >= '0' && character <= '9' || character >= 'a' && character <= 'f') {
			return false
		}
	}
	return true
}

// DeriveReviewRequest creates an immutable, metadata-only flagged revision.
type DeriveReviewRequest struct {
	InputDirectory  string
	OutputDirectory string
	ArtifactTitle   string
	Review          manifest.Review
}

// DeriveReview copies verified evidence blobs byte-for-byte and publishes a new flagged OCI layout.
func DeriveReview(request DeriveReviewRequest) (dawgtypes.PackagedArtifact, manifest.Manifest, error) {
	input, err := filepath.Abs(request.InputDirectory)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	output, err := filepath.Abs(request.OutputDirectory)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	if input == output || strings.HasPrefix(output, input+string(os.PathSeparator)) {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, fmt.Errorf("review output must not be the source artifact or its descendant")
	}
	if _, err := os.Lstat(output); !os.IsNotExist(err) {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, fmt.Errorf("review output directory already exists: %s", output)
	}
	source, err := Validate(input)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, fmt.Errorf("validate source artifact: %w", err)
	}
	duration, err := ReplayDuration(input, source)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	expectedReview := NewReview(source, append([]manifest.ReviewFlag(nil), request.Review.Flags...), request.Review.ReviewedAt)
	if request.Review.FormatVersion != expectedReview.FormatVersion || request.Review.Kind != expectedReview.Kind || request.Review.RootArtifactID != expectedReview.RootArtifactID || request.Review.ParentArtifactID != expectedReview.ParentArtifactID || request.Review.Revision != expectedReview.Revision {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, fmt.Errorf("review lineage does not match source artifact")
	}
	if err := request.Review.Validate(duration); err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, fmt.Errorf("validate review: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o700); err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, fmt.Errorf("create review output parent: %w", err)
	}
	staging, err := os.MkdirTemp(filepath.Dir(output), ".dawg-review-")
	if err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	defer os.RemoveAll(staging)
	if err := copyLayout(input, staging); err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	derived := source
	derived.SchemaVersion = manifest.SchemaVersion
	derived.Title = request.ArtifactTitle
	derived.Provenance = nil
	derived.Review = &request.Review
	derived.ID, err = Identity(derived)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	schemaPath, err := manifest.DefaultSchemaPath()
	if err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	if err := manifest.Validate(schemaPath, derived); err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	config, err := manifest.Marshal(derived)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	if err := os.WriteFile(filepath.Join(staging, "dawg-manifest.json"), config, 0o600); err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	configDescriptor := descriptor(dawgtypes.MediaTypeManifest, config)
	if err := writeBlob(staging, configDescriptor.Digest, config); err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	layers := make([]Descriptor, 0, len(derived.Layers))
	for _, layer := range derived.Layers {
		layers = append(layers, Descriptor{MediaType: string(layer.MediaType), Digest: layer.Digest, Size: layer.Size})
	}
	imageBytes, err := json.Marshal(imageManifest{SchemaVersion: 2, MediaType: ociImageManifestMediaType, Config: configDescriptor, Layers: layers})
	if err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	imageDescriptor := descriptor(dawgtypes.MediaType(ociImageManifestMediaType), imageBytes)
	if err := writeBlob(staging, imageDescriptor.Digest, imageBytes); err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	indexBytes, err := json.MarshalIndent(index{SchemaVersion: 2, MediaType: ociImageIndexMediaType, Manifests: []Descriptor{imageDescriptor}}, "", "  ")
	if err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	if err := os.WriteFile(filepath.Join(staging, "index.json"), append(indexBytes, '\n'), 0o600); err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, err
	}
	if _, err := Validate(staging); err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, fmt.Errorf("validate derived artifact: %w", err)
	}
	if err := os.Rename(staging, output); err != nil {
		return dawgtypes.PackagedArtifact{}, manifest.Manifest{}, fmt.Errorf("publish review artifact: %w", err)
	}
	return dawgtypes.PackagedArtifact{Directory: output, ManifestPath: filepath.Join(output, "dawg-manifest.json"), OCIManifestDigest: imageDescriptor.Digest}, derived, nil
}

func descriptor(mediaType dawgtypes.MediaType, contents []byte) Descriptor {
	digest := sha256.Sum256(contents)
	return Descriptor{MediaType: string(mediaType), Digest: fmt.Sprintf("sha256:%x", digest), Size: int64(len(contents))}
}

func writeBlob(directory, digest string, contents []byte) error {
	if !validDigest(digest) {
		return fmt.Errorf("invalid blob digest")
	}
	return os.WriteFile(filepath.Join(directory, "blobs", "sha256", strings.TrimPrefix(digest, "sha256:")), contents, 0o600)
}

func copyLayout(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("artifact contains symbolic link: %s", path)
		}
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("artifact contains non-regular file: %s", path)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			_ = in.Close()
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := firstError(out.Close(), in.Close())
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func firstError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

// ReplayDuration returns the rrweb timeline duration used to validate flag offsets.
func ReplayDuration(directory string, value manifest.Manifest) (int64, error) {
	for _, layer := range value.Layers {
		if layer.MediaType != dawgtypes.MediaTypeTrace {
			continue
		}
		contents, err := verifyBlob(directory, layer.Digest, layer.Size)
		if err != nil {
			return 0, err
		}
		decoder, err := zstd.NewReader(bytes.NewReader(contents))
		if err != nil {
			return 0, fmt.Errorf("read review trace: %w", err)
		}
		decoded, err := io.ReadAll(io.LimitReader(decoder, (128<<20)+1))
		decoder.Close()
		if err != nil || len(decoded) > 128<<20 {
			return 0, fmt.Errorf("read review trace: invalid or oversized trace")
		}
		reader := tar.NewReader(bytes.NewReader(decoded))
		for {
			header, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return 0, fmt.Errorf("read review trace archive: %w", err)
			}
			if header.Typeflag != tar.TypeReg || header.Size < 0 || header.Size > 128<<20 || strings.Contains(header.Name, "..") {
				return 0, fmt.Errorf("unsafe review trace entry")
			}
			if filepath.ToSlash(header.Name) == "traces/rrweb.jsonl" {
				return traceDuration(io.LimitReader(reader, header.Size))
			}
		}
	}
	return 0, fmt.Errorf("artifact has no replay trace")
}

func traceDuration(reader io.Reader) (int64, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64<<10), dawgtypes.MaxJSONLRecordBytes+1)
	var first, last int64
	found := false
	for scanner.Scan() {
		if len(scanner.Bytes()) > dawgtypes.MaxJSONLRecordBytes {
			return 0, fmt.Errorf("rrweb event exceeds %d bytes", dawgtypes.MaxJSONLRecordBytes)
		}
		var event struct {
			Timestamp int64 `json:"timestamp"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return 0, fmt.Errorf("invalid rrweb event")
		}
		if event.Timestamp <= 0 {
			continue
		}
		if !found || event.Timestamp < first {
			first = event.Timestamp
		}
		if !found || event.Timestamp > last {
			last = event.Timestamp
		}
		found = true
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("replay trace has no timestamps")
	}
	return last - first, nil
}

// NewReview builds lineage metadata from a source manifest and validated flags.
func NewReview(source manifest.Manifest, flags []manifest.ReviewFlag, reviewedAt time.Time) manifest.Review {
	root := source.ID
	revision := 1
	if source.Review != nil {
		root = source.Review.RootArtifactID
		revision = source.Review.Revision + 1
	}
	manifest.SortFlags(flags)
	return manifest.Review{FormatVersion: "1", Kind: "flagged", RootArtifactID: root, ParentArtifactID: source.ID, Revision: revision, ReviewedAt: reviewedAt.UTC(), Flags: flags}
}
