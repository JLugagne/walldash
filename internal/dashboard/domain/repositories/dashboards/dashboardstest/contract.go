package dashboardstest

import (
	"context"
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/dashboards"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDashboardRepository is a function-based mock implementation of dashboards.DashboardRepository.
type MockDashboardRepository struct {
	CreateDashboardFunc   func(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error)
	FindDashboardByIDFunc func(ctx context.Context, id string) (domain.Dashboard, error)
	FindAllDashboardsFunc func(ctx context.Context) ([]domain.Dashboard, error)
	UpdateDashboardFunc   func(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error)
	DeleteDashboardFunc   func(ctx context.Context, id string) error
}

func (m *MockDashboardRepository) CreateDashboard(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error) {
	if m.CreateDashboardFunc == nil {
		panic("called not defined CreateDashboardFunc")
	}
	return m.CreateDashboardFunc(ctx, dashboard)
}

func (m *MockDashboardRepository) FindDashboardByID(ctx context.Context, id string) (domain.Dashboard, error) {
	if m.FindDashboardByIDFunc == nil {
		panic("called not defined FindDashboardByIDFunc")
	}
	return m.FindDashboardByIDFunc(ctx, id)
}

func (m *MockDashboardRepository) FindAllDashboards(ctx context.Context) ([]domain.Dashboard, error) {
	if m.FindAllDashboardsFunc == nil {
		panic("called not defined FindAllDashboardsFunc")
	}
	return m.FindAllDashboardsFunc(ctx)
}

func (m *MockDashboardRepository) UpdateDashboard(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error) {
	if m.UpdateDashboardFunc == nil {
		panic("called not defined UpdateDashboardFunc")
	}
	return m.UpdateDashboardFunc(ctx, dashboard)
}

func (m *MockDashboardRepository) DeleteDashboard(ctx context.Context, id string) error {
	if m.DeleteDashboardFunc == nil {
		panic("called not defined DeleteDashboardFunc")
	}
	return m.DeleteDashboardFunc(ctx, id)
}

// MockWidgetRepository is a function-based mock implementation of dashboards.WidgetRepository.
type MockWidgetRepository struct {
	CreateWidgetFunc               func(ctx context.Context, widget domain.Widget) (domain.Widget, error)
	FindWidgetByIDFunc             func(ctx context.Context, id string) (domain.Widget, error)
	FindWidgetsByDashboardIDFunc   func(ctx context.Context, dashboardID string) ([]domain.Widget, error)
	UpdateWidgetFunc               func(ctx context.Context, widget domain.Widget) (domain.Widget, error)
	DeleteWidgetFunc               func(ctx context.Context, id string) error
	DeleteWidgetsByDashboardIDFunc func(ctx context.Context, dashboardID string) error
	ReplaceWidgetPositionsFunc     func(ctx context.Context, dashboardID string, positions []domain.WidgetPosition) error
}

func (m *MockWidgetRepository) CreateWidget(ctx context.Context, widget domain.Widget) (domain.Widget, error) {
	if m.CreateWidgetFunc == nil {
		panic("called not defined CreateWidgetFunc")
	}
	return m.CreateWidgetFunc(ctx, widget)
}

func (m *MockWidgetRepository) FindWidgetByID(ctx context.Context, id string) (domain.Widget, error) {
	if m.FindWidgetByIDFunc == nil {
		panic("called not defined FindWidgetByIDFunc")
	}
	return m.FindWidgetByIDFunc(ctx, id)
}

func (m *MockWidgetRepository) FindWidgetsByDashboardID(ctx context.Context, dashboardID string) ([]domain.Widget, error) {
	if m.FindWidgetsByDashboardIDFunc == nil {
		panic("called not defined FindWidgetsByDashboardIDFunc")
	}
	return m.FindWidgetsByDashboardIDFunc(ctx, dashboardID)
}

func (m *MockWidgetRepository) UpdateWidget(ctx context.Context, widget domain.Widget) (domain.Widget, error) {
	if m.UpdateWidgetFunc == nil {
		panic("called not defined UpdateWidgetFunc")
	}
	return m.UpdateWidgetFunc(ctx, widget)
}

func (m *MockWidgetRepository) DeleteWidget(ctx context.Context, id string) error {
	if m.DeleteWidgetFunc == nil {
		panic("called not defined DeleteWidgetFunc")
	}
	return m.DeleteWidgetFunc(ctx, id)
}

func (m *MockWidgetRepository) DeleteWidgetsByDashboardID(ctx context.Context, dashboardID string) error {
	if m.DeleteWidgetsByDashboardIDFunc == nil {
		panic("called not defined DeleteWidgetsByDashboardIDFunc")
	}
	return m.DeleteWidgetsByDashboardIDFunc(ctx, dashboardID)
}

func (m *MockWidgetRepository) ReplaceWidgetPositions(ctx context.Context, dashboardID string, positions []domain.WidgetPosition) error {
	if m.ReplaceWidgetPositionsFunc == nil {
		panic("called not defined ReplaceWidgetPositionsFunc")
	}
	return m.ReplaceWidgetPositionsFunc(ctx, dashboardID, positions)
}

// DashboardRepositoryContractTesting runs all contract tests for a DashboardRepository implementation.
func DashboardRepositoryContractTesting(t *testing.T, repo dashboards.DashboardRepository) {
	ctx := context.Background()

	t.Run("Contract: CreateDashboard and FindDashboardByID retrieves stored dashboard", func(t *testing.T) {
		ov := domain.Dashboard{
			ID:        "ov-contract-1",
			Name:      "General Dashboard",
			Order:     1,
			Cols:      domain.DefaultGridCols,
			Rows:      domain.DefaultGridRows,
			CreatedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt: time.Now().UTC().Truncate(time.Second),
		}

		created, err := repo.CreateDashboard(ctx, ov)
		require.NoError(t, err)
		assert.Equal(t, ov.ID, created.ID)
		assert.Equal(t, ov.Name, created.Name)
		assert.Equal(t, ov.Order, created.Order)

		found, err := repo.FindDashboardByID(ctx, ov.ID)
		require.NoError(t, err)
		assert.Equal(t, ov.ID, found.ID)
		assert.Equal(t, ov.Name, found.Name)
		assert.Equal(t, ov.Order, found.Order)
	})

	t.Run("Contract: FindDashboardByID returns ErrDashboardNotFound for missing dashboard", func(t *testing.T) {
		_, err := repo.FindDashboardByID(ctx, "non-existent-dashboard")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDashboardNotFound)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: FindAllDashboards returns dashboards in ascending order", func(t *testing.T) {
		o1 := domain.Dashboard{
			ID:        "ov-sort-2",
			Name:      "Dashboard 2",
			Order:     20,
			Cols:      domain.DefaultGridCols,
			Rows:      domain.DefaultGridRows,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		o2 := domain.Dashboard{
			ID:        "ov-sort-1",
			Name:      "Dashboard 1",
			Order:     10,
			Cols:      domain.DefaultGridCols,
			Rows:      domain.DefaultGridRows,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		_, err := repo.CreateDashboard(ctx, o1)
		require.NoError(t, err)
		_, err = repo.CreateDashboard(ctx, o2)
		require.NoError(t, err)

		all, err := repo.FindAllDashboards(ctx)
		require.NoError(t, err)
		require.True(t, len(all) >= 2)

		for i := 0; i < len(all)-1; i++ {
			assert.True(t, all[i].Order <= all[i+1].Order, "Dashboards must be sorted by Order ascending")
		}
	})

	t.Run("Contract: UpdateDashboard modifies dashboard details", func(t *testing.T) {
		ov := domain.Dashboard{
			ID:        "ov-update-1",
			Name:      "Dashboard Original",
			Order:     5,
			Cols:      domain.DefaultGridCols,
			Rows:      domain.DefaultGridRows,
			CreatedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt: time.Now().UTC().Truncate(time.Second),
		}
		_, err := repo.CreateDashboard(ctx, ov)
		require.NoError(t, err)

		ov.Name = "Modified Dashboard"
		ov.UpdatedAt = time.Now().UTC().Truncate(time.Second)

		updated, err := repo.UpdateDashboard(ctx, ov)
		require.NoError(t, err)
		assert.Equal(t, "Modified Dashboard", updated.Name)

		found, err := repo.FindDashboardByID(ctx, ov.ID)
		require.NoError(t, err)
		assert.Equal(t, "Modified Dashboard", found.Name)
	})

	t.Run("Contract: UpdateDashboard returns ErrDashboardNotFound for non-existent dashboard", func(t *testing.T) {
		missing := domain.Dashboard{
			ID:        "ov-missing-update",
			Name:      "Inconnu",
			Order:     99,
			Cols:      domain.DefaultGridCols,
			Rows:      domain.DefaultGridRows,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_, err := repo.UpdateDashboard(ctx, missing)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDashboardNotFound)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: DeleteDashboard removes an existing dashboard", func(t *testing.T) {
		ov := domain.Dashboard{
			ID:        "ov-del-1",
			Name:      "To Delete",
			Order:     50,
			Cols:      domain.DefaultGridCols,
			Rows:      domain.DefaultGridRows,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_, err := repo.CreateDashboard(ctx, ov)
		require.NoError(t, err)

		err = repo.DeleteDashboard(ctx, ov.ID)
		require.NoError(t, err)

		_, err = repo.FindDashboardByID(ctx, ov.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDashboardNotFound)
	})

	t.Run("Contract: DeleteDashboard returns ErrDashboardNotFound for missing dashboard", func(t *testing.T) {
		err := repo.DeleteDashboard(ctx, "ov-missing-delete")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDashboardNotFound)
		assert.True(t, domain.IsDomainError(err))
	})
}

// WidgetRepositoryContractTesting runs all contract tests for a WidgetRepository implementation.
func WidgetRepositoryContractTesting(t *testing.T, repo dashboards.WidgetRepository, dashboardRepo dashboards.DashboardRepository) {
	ctx := context.Background()

	parent := domain.Dashboard{
		ID:        "ov-parent-widget-test",
		Name:      "Dashboard For Widgets",
		Order:     0,
		Cols:      domain.DefaultGridCols,
		Rows:      domain.DefaultGridRows,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, _ = dashboardRepo.CreateDashboard(ctx, parent)

	t.Run("Contract: CreateWidget and FindWidgetByID retrieves stored widget", func(t *testing.T) {
		w := domain.Widget{
			ID:          "widget-contract-1",
			DashboardID: parent.ID,
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Living Room Automations",
			Order:       1,
			ColSpan:     2,
			RowSpan:     2,
			Config: domain.WidgetConfig{
				EntityIDs: []string{"automation.eteindre_tout", "automation.cinema"},
				Display:   domain.DisplayList,
			},
			CreatedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt: time.Now().UTC().Truncate(time.Second),
		}

		created, err := repo.CreateWidget(ctx, w)
		require.NoError(t, err)
		assert.Equal(t, w.ID, created.ID)
		assert.Equal(t, w.DashboardID, created.DashboardID)
		assert.Equal(t, w.Type, created.Type)
		assert.Equal(t, w.Title, created.Title)
		assert.Equal(t, w.Config.EntityIDs, created.Config.EntityIDs)

		found, err := repo.FindWidgetByID(ctx, w.ID)
		require.NoError(t, err)
		assert.Equal(t, w.ID, found.ID)
		assert.Equal(t, w.Title, found.Title)
		assert.Equal(t, w.Config.EntityIDs, found.Config.EntityIDs)
	})

	t.Run("Contract: FindWidgetByID returns ErrWidgetNotFound for missing widget", func(t *testing.T) {
		_, err := repo.FindWidgetByID(ctx, "non-existent-widget")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWidgetNotFound)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: FindWidgetsByDashboardID returns widgets sorted by order", func(t *testing.T) {
		w2 := domain.Widget{
			ID:          "widget-dash-sort-2",
			DashboardID: parent.ID,
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Widget 2",
			Order:       10,
			ColSpan:     2,
			RowSpan:     2,
			Config: domain.WidgetConfig{
				EntityIDs: []string{"automation.two"},
				Display:   domain.DisplayList,
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		w1 := domain.Widget{
			ID:          "widget-dash-sort-1",
			DashboardID: parent.ID,
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Widget 1",
			Order:       5,
			ColSpan:     2,
			RowSpan:     2,
			Config: domain.WidgetConfig{
				EntityIDs: []string{"automation.one"},
				Display:   domain.DisplayList,
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		_, err := repo.CreateWidget(ctx, w2)
		require.NoError(t, err)
		_, err = repo.CreateWidget(ctx, w1)
		require.NoError(t, err)

		list, err := repo.FindWidgetsByDashboardID(ctx, parent.ID)
		require.NoError(t, err)
		require.True(t, len(list) >= 2)

		for i := 0; i < len(list)-1; i++ {
			assert.True(t, list[i].Order <= list[i+1].Order, "Widgets must be ordered by Order ascending")
		}
	})

	t.Run("Contract: UpdateWidget modifies widget", func(t *testing.T) {
		w := domain.Widget{
			ID:          "widget-upd-1",
			DashboardID: parent.ID,
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Ancien Titre",
			Order:       3,
			ColSpan:     2,
			RowSpan:     2,
			Config: domain.WidgetConfig{
				EntityIDs: []string{"automation.one"},
				Display:   domain.DisplayList,
			},
			CreatedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt: time.Now().UTC().Truncate(time.Second),
		}
		_, err := repo.CreateWidget(ctx, w)
		require.NoError(t, err)

		w.Title = "Nouveau Titre"
		w.Config.EntityIDs = []string{"automation.one", "automation.two"}
		w.UpdatedAt = time.Now().UTC().Truncate(time.Second)

		updated, err := repo.UpdateWidget(ctx, w)
		require.NoError(t, err)
		assert.Equal(t, "Nouveau Titre", updated.Title)
		assert.Equal(t, []string{"automation.one", "automation.two"}, updated.Config.EntityIDs)

		found, err := repo.FindWidgetByID(ctx, w.ID)
		require.NoError(t, err)
		assert.Equal(t, "Nouveau Titre", found.Title)
	})

	t.Run("Contract: UpdateWidget returns ErrWidgetNotFound for missing widget", func(t *testing.T) {
		missing := domain.Widget{
			ID:          "widget-missing",
			DashboardID: parent.ID,
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Missing",
			ColSpan:     2,
			RowSpan:     2,
			Config: domain.WidgetConfig{
				EntityIDs: []string{"automation.missing"},
				Display:   domain.DisplayList,
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_, err := repo.UpdateWidget(ctx, missing)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWidgetNotFound)
	})

	t.Run("Contract: DeleteWidget removes widget", func(t *testing.T) {
		w := domain.Widget{
			ID:          "widget-del-1",
			DashboardID: parent.ID,
			Type:        domain.WidgetTypeAutomationList,
			Title:       "To Destroy",
			Order:       9,
			ColSpan:     2,
			RowSpan:     2,
			Config: domain.WidgetConfig{
				EntityIDs: []string{"automation.del"},
				Display:   domain.DisplayList,
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_, err := repo.CreateWidget(ctx, w)
		require.NoError(t, err)

		err = repo.DeleteWidget(ctx, w.ID)
		require.NoError(t, err)

		_, err = repo.FindWidgetByID(ctx, w.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWidgetNotFound)
	})

	t.Run("Contract: DeleteWidget returns ErrWidgetNotFound for missing widget", func(t *testing.T) {
		err := repo.DeleteWidget(ctx, "widget-missing-del")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWidgetNotFound)
	})

	t.Run("Contract: ReplaceWidgetPositions updates the position of every widget", func(t *testing.T) {
		w1 := domain.Widget{
			ID:          "widget-pos-1",
			DashboardID: parent.ID,
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Position 1",
			Order:       1,
			Col:         0,
			Row:         0,
			ColSpan:     2,
			RowSpan:     2,
			Config: domain.WidgetConfig{
				EntityIDs: []string{"automation.pos_one"},
				Display:   domain.DisplayList,
			},
			CreatedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt: time.Now().UTC().Truncate(time.Second),
		}
		w2 := domain.Widget{
			ID:          "widget-pos-2",
			DashboardID: parent.ID,
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Position 2",
			Order:       2,
			Col:         2,
			Row:         0,
			ColSpan:     2,
			RowSpan:     2,
			Config: domain.WidgetConfig{
				EntityIDs: []string{"automation.pos_two"},
				Display:   domain.DisplayList,
			},
			CreatedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt: time.Now().UTC().Truncate(time.Second),
		}
		_, err := repo.CreateWidget(ctx, w1)
		require.NoError(t, err)
		_, err = repo.CreateWidget(ctx, w2)
		require.NoError(t, err)

		positions := []domain.WidgetPosition{
			{ID: w1.ID, Col: 4, Row: 4, ColSpan: 2, RowSpan: 2},
			{ID: w2.ID, Col: 6, Row: 4, ColSpan: 2, RowSpan: 2},
		}
		err = repo.ReplaceWidgetPositions(ctx, parent.ID, positions)
		require.NoError(t, err)

		found1, err := repo.FindWidgetByID(ctx, w1.ID)
		require.NoError(t, err)
		assert.Equal(t, 4, found1.Col)
		assert.Equal(t, 4, found1.Row)

		found2, err := repo.FindWidgetByID(ctx, w2.ID)
		require.NoError(t, err)
		assert.Equal(t, 6, found2.Col)
		assert.Equal(t, 4, found2.Row)
	})

	t.Run("Contract: ReplaceWidgetPositions returns ErrWidgetNotFound for a missing widget", func(t *testing.T) {
		positions := []domain.WidgetPosition{
			{ID: "widget-pos-missing", Col: 0, Row: 0, ColSpan: 2, RowSpan: 2},
		}
		err := repo.ReplaceWidgetPositions(ctx, parent.ID, positions)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWidgetNotFound)
	})
}
