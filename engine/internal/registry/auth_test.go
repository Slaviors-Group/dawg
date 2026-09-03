package registry

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetCredentials(t *testing.T) {
	// Temporarily override the home directory
	tmpHome := t.TempDir()
	_, err := os.UserHomeDir()
	if err == nil {
		t.Setenv("USERPROFILE", tmpHome) // Windows
		t.Setenv("HOME", tmpHome)        // Linux/Mac
	}

	// Create a mock .docker/config.json
	dockerDir := filepath.Join(tmpHome, ".docker")
	if err := os.MkdirAll(dockerDir, 0o755); err != nil {
		t.Fatalf("failed to create docker config dir: %v", err)
	}
	
	configPath := filepath.Join(dockerDir, "config.json")
	
	// Create mock credentials
	authPayload := base64.StdEncoding.EncodeToString([]byte("testuser:testpass123"))
	mockConfig := DockerConfig{
		Auths: map[string]DockerAuth{
			"ghcr.io": {Auth: authPayload},
		},
	}
	
	contents, _ := json.Marshal(mockConfig)
	if err := os.WriteFile(configPath, contents, 0o600); err != nil {
		t.Fatalf("failed to write mock docker config: %v", err)
	}

	// Test 1: Successful retrieval
	username, password, err := GetCredentials("ghcr.io")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if username != "testuser" || password != "testpass123" {
		t.Errorf("expected testuser:testpass123, got %s:%s", username, password)
	}

	// Test 2: Missing registry returns empty but no error
	username, password, err = GetCredentials("unknown.registry.com")
	if err != nil {
		t.Fatalf("expected no error for missing registry, got %v", err)
	}
	if username != "" || password != "" {
		t.Errorf("expected empty credentials, got %s:%s", username, password)
	}
	
	// Test 3: No config.json file should return no error (anonymous fallback)
	os.Remove(configPath)
	username, password, err = GetCredentials("ghcr.io")
	if err != nil {
		t.Fatalf("expected no error when config is missing, got %v", err)
	}
	if username != "" || password != "" {
		t.Errorf("expected empty credentials when config is missing, got %s:%s", username, password)
	}
}
