package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/health/healthtest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/uow/uowtest"
	"github.com/JLugagne/ha-dash/internal/dashboard/outbound/sqlite"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlite.Adapter {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	adapter, err := sqlite.New(context.Background(), dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = adapter.Close()
	})
	return adapter
}

func TestSQLiteHealthRepositoryContract(t *testing.T) {
	adapter := setupTestDB(t)
	healthtest.RepositoryContractTesting(t, adapter)
}

func TestSQLiteUnitOfWorkContract(t *testing.T) {
	adapter := setupTestDB(t)
	uowtest.UnitOfWorkContractTesting(t, adapter)
}
