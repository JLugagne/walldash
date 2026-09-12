package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/dashboards/dashboardstest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/health/healthtest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/levels/levelstest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/placements/placementstest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/plans/planstest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow/uowtest"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/sqlite"
	"github.com/stretchr/testify/assert"
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

func TestSQLiteDashboardRepositoryContract(t *testing.T) {
	adapter := setupTestDB(t)
	dashboardstest.DashboardRepositoryContractTesting(t, adapter)
}

func TestSQLiteWidgetRepositoryContract(t *testing.T) {
	adapter := setupTestDB(t)
	dashboardstest.WidgetRepositoryContractTesting(t, adapter, adapter)
}

func TestFreshDatabaseSeedsDefaultWeatherDashboard(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = adapter.Close() })

	dashboards, err := adapter.FindAllDashboards(ctx)
	require.NoError(t, err)
	require.Len(t, dashboards, 1)

	seeded := dashboards[0]
	assert.Equal(t, "Home", seeded.Name)
	assert.Equal(t, domain.DefaultGridCols, seeded.Cols)
	assert.Equal(t, domain.DefaultGridRows, seeded.Rows)
	require.Len(t, seeded.Widgets, 1)

	widget := seeded.Widgets[0]
	assert.Equal(t, domain.WidgetTypeWeather, widget.Type)
	assert.Equal(t, domain.WeatherModeNDays, widget.Config.WeatherMode)
	assert.Equal(t, 5, widget.Config.WeatherDays)
	require.NoError(t, widget.ValidateIn(seeded.Cols, seeded.Rows))

	// The seed is applied at most once: deleting it and reopening the same database must
	// not resurrect a dashboard the user chose to remove.
	require.NoError(t, adapter.DeleteDashboard(ctx, seeded.ID))
	require.NoError(t, adapter.Close())

	reopened, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = reopened.Close() })

	after, err := reopened.FindAllDashboards(ctx)
	require.NoError(t, err)
	assert.Empty(t, after)
}
