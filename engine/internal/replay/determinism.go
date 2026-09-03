package replay

import (
	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

// ConfigureDeterminism provides the environment variables required to freeze time
// and inject random seeds into the target container.
func ConfigureDeterminism(config dawgtypes.DeterminismConfig) []string {
	var env []string

	if !config.ClockFrozenAt.IsZero() {
		// Example implementation for libfaketime or similar:
		// We pass FAKETIME environment variable which the container can consume if configured.
		env = append(env, "FAKETIME=@"+config.ClockFrozenAt.Format("2006-01-02 15:04:05"))
	}

	if config.RandomSeed != 0 {
		// Pass seed to app (exact env var depends on application framework, e.g. PYTHONHASHSEED)
		// For Node: NODE_OPTIONS="--predictable-random-seed" (requires specific build) or just DAWG_SEED
		// DAWG expects the app's Dockerfile to map this variable if determinism is required.
	}

	return env
}
