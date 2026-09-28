package main

import (
	"testing"
	"time"
)

func TestArtifactDirectoryNameForTitleUsesOnlySlug(t *testing.T) {
	if got, want := artifactDirectoryNameForTitle(" Checkout Validation Regression! "), "checkout-validation-regression"; got != want {
		t.Fatalf("artifact directory name = %q, want %q", got, want)
	}
}

func TestDefaultArtifactDirectoryNameIncludesTimestampAndHost(t *testing.T) {
	capturedAt := time.Date(2026, time.September, 28, 12, 34, 56, 0, time.UTC)
	if got, want := defaultArtifactDirectoryName("https://staging.example.com/orders", capturedAt), "20260928-123456-staging-example-com"; got != want {
		t.Fatalf("default artifact directory name = %q, want %q", got, want)
	}
}
