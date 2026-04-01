package dbtest

import (
	"context"
	"database/sql"
	"sync"
	"testing"
)

func TestEnsureMigratedRunsMigrationsOnce(t *testing.T) {
	originalRunMigrations := runMigrations
	originalMigrateErr := migrateErr
	t.Cleanup(func() {
		runMigrations = originalRunMigrations
		migrateOnce = sync.Once{}
		migrateErr = originalMigrateErr
	})

	migrateOnce = sync.Once{}
	migrateErr = nil

	callCount := 0
	runMigrations = func(context.Context, *sql.DB) (int, error) {
		callCount++
		return 0, nil
	}

	if err := ensureMigrated(context.Background(), nil); err != nil {
		t.Fatalf("first ensureMigrated call failed: %v", err)
	}
	if err := ensureMigrated(context.Background(), nil); err != nil {
		t.Fatalf("second ensureMigrated call failed: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected one migration call, got %d", callCount)
	}
}
