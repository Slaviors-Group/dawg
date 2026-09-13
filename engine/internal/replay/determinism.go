package replay

import (
	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

// ConfigureDeterminism returns replay environment variables for configured controls.
func ConfigureDeterminism(config dawgtypes.DeterminismConfig) []string {
	var env []string

	if !config.ClockFrozenAt.IsZero() {
		// The target environment must consume FAKETIME for clock freezing to apply.
		env = append(env, "FAKETIME=@"+config.ClockFrozenAt.Format("2006-01-02 15:04:05"))
	}

	return env
}
