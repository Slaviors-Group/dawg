package replay

import (
	"context"
	"fmt"
)

// DBRestorer provides a mechanism to restore database fixtures into a sandbox.
type DBRestorer struct {
	Runner CommandRunner
	dbName string
}

// NewDBRestorer creates a new DBRestorer for the specified database name.
func NewDBRestorer(runner CommandRunner, dbName string) *DBRestorer {
	if runner == nil {
		runner = ExecRunner{}
	}
	return &DBRestorer{
		Runner: runner,
		dbName: dbName,
	}
}

// Restore copies a database fixture into the running container and applies it via pg_restore.
func (r *DBRestorer) Restore(ctx context.Context, containerName string, fixturePath string) error {
	// Copy fixture into container
	output, err := r.Runner.Run(ctx, "docker", "cp", fixturePath, containerName+":/tmp/fixture.sql")
	if err != nil {
		return fmt.Errorf("dbrestore: failed to copy fixture: %s: %w", string(output), err)
	}

	// Execute pg_restore inside container using the generic CommandRunner instead of direct exec to match pattern
	// Wait, the integration pattern uses exec directly, but we have CommandRunner.
	// Let's use CommandRunner for consistency if we want, but the pattern says exec.CommandContext.
	// Actually, using the runner makes it testable.
	restoreOutput, err := r.Runner.Run(ctx, "docker", "exec", containerName,
		"pg_restore",
		"--data-only",
		"--no-owner",
		"--no-privileges",
		"-d", r.dbName,
		"/tmp/fixture.sql",
	)
	if err != nil {
		return fmt.Errorf("dbrestore: pg_restore failed: %s: %w", string(restoreOutput), err)
	}

	return nil
}
