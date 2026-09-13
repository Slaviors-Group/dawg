package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DockerConfig represents the structure of ~/.docker/config.json
type DockerConfig struct {
	Auths map[string]DockerAuth `json:"auths"`
}

// DockerAuth represents the auth entry for a specific registry
type DockerAuth struct {
	Auth string `json:"auth"`
}

// GetCredentials reads ~/.docker/config.json and returns the username and password
// for the specified registry reference (e.g., "index.docker.io", "ghcr.io").
func GetCredentials(registry string) (username, password string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("registry: locate home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".docker", "config.json")
	contents, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil // No credentials configured, proceed anonymously
		}
		return "", "", fmt.Errorf("registry: read docker config: %w", err)
	}

	var config DockerConfig
	if err := json.Unmarshal(contents, &config); err != nil {
		return "", "", fmt.Errorf("registry: parse docker config: %w", err)
	}

	// Try exact match and https:// match
	authEntry, ok := config.Auths[registry]
	if !ok {
		authEntry, ok = config.Auths["https://"+registry]
		if !ok && registry == "docker.io" {
			authEntry, ok = config.Auths["https://index.docker.io/v1/"]
		}
	}

	if !ok || authEntry.Auth == "" {
		return "", "", nil // No credentials for this registry
	}

	decoded, err := base64.StdEncoding.DecodeString(authEntry.Auth)
	if err != nil {
		return "", "", fmt.Errorf("registry: decode auth string for %s: %w", registry, err)
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("registry: invalid auth string format for %s", registry)
	}

	return parts[0], parts[1], nil
}
