// Package registry implements PRD §6.4 OCI artifact push and pull.
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"
)

// Client transfers OCI layouts through an OCI registry.
type Client struct{}

// Push copies the OCI manifest selected by the layout index to ref.
func (Client) Push(ctx context.Context, layoutDirectory, ref string) error {
	source, err := oci.NewFromFS(ctx, os.DirFS(layoutDirectory))
	if err != nil {
		return fmt.Errorf("registry: open OCI layout: %w", err)
	}
	digest, err := indexedManifestDigest(layoutDirectory)
	if err != nil {
		return err
	}
	target, err := remote.NewRepository(ref)
	if err != nil {
		return fmt.Errorf("registry: invalid reference %s: %w", ref, err)
	}
	if _, err := oras.Copy(ctx, source, digest, target, ref, oras.DefaultCopyOptions); err != nil {
		return fmt.Errorf("registry: push %s: %w", ref, err)
	}
	return nil
}

// Pull copies ref into a newly created OCI layout directory.
func (Client) Pull(ctx context.Context, ref, destination string) error {
	source, err := remote.NewRepository(ref)
	if err != nil {
		return fmt.Errorf("registry: invalid reference %s: %w", ref, err)
	}
	target, err := oci.New(destination)
	if err != nil {
		return fmt.Errorf("registry: create OCI layout: %w", err)
	}
	if _, err := oras.Copy(ctx, source, ref, target, ref, oras.DefaultCopyOptions); err != nil {
		return fmt.Errorf("registry: pull %s: %w", ref, err)
	}
	return nil
}

func indexedManifestDigest(layoutDirectory string) (string, error) {
	contents, err := os.ReadFile(filepath.Join(layoutDirectory, "index.json"))
	if err != nil {
		return "", fmt.Errorf("registry: read OCI index: %w", err)
	}
	var index struct {
		Manifests []struct {
			Digest string `json:"digest"`
		} `json:"manifests"`
	}
	if err := json.Unmarshal(contents, &index); err != nil {
		return "", fmt.Errorf("registry: decode OCI index: %w", err)
	}
	if len(index.Manifests) != 1 || index.Manifests[0].Digest == "" {
		return "", fmt.Errorf("registry: OCI index must contain one manifest")
	}
	return index.Manifests[0].Digest, nil
}
