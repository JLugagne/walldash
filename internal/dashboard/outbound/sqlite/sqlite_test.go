package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/health/healthtest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/levels/levelstest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/placements/placementstest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/plans/planstest"
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

func TestSQLiteLevelRepositoryContract(t *testing.T) {
	adapter := setupTestDB(t)
	levelstest.LevelRepositoryContractTesting(t, adapter)
}

func TestSQLitePlanRepositoryContract(t *testing.T) {
	adapter := setupTestDB(t)
	planstest.PlanRepositoryContractTesting(t, adapter)
}

func TestSQLiteDevicePlacementRepositoryContract(t *testing.T) {
	adapter := setupTestDB(t)
	ctx := context.Background()
	testLevelID := "level-contract-test"
	_, err := adapter.Create(ctx, domain.Level{
		ID:        testLevelID,
		Name:      "Test Level",
		Order:     1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	require.NoError(t, err)

	placementstest.DevicePlacementRepositoryContractTesting(t, adapter, testLevelID)
}
