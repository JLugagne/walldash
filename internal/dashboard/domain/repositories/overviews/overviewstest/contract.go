package overviewstest

import (
	"context"
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/overviews"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockOverviewRepository is a function-based mock implementation of overviews.OverviewRepository.
type MockOverviewRepository struct {
	CreateOverviewFunc   func(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error)
	FindOverviewByIDFunc func(ctx context.Context, id string) (domain.OverviewDashboard, error)
	FindAllOverviewsFunc func(ctx context.Context) ([]domain.OverviewDashboard, error)
	UpdateOverviewFunc   func(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error)
	DeleteOverviewFunc   func(ctx context.Context, id string) error
}

func (m *MockOverviewRepository) CreateOverview(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
	if m.CreateOverviewFunc == nil {
		panic("called not defined CreateOverviewFunc")
	}
	return m.CreateOverviewFunc(ctx, overview)
}

func (m *MockOverviewRepository) FindOverviewByID(ctx context.Context, id string) (domain.OverviewDashboard, error) {
	if m.FindOverviewByIDFunc == nil {
		panic("called not defined FindOverviewByIDFunc")
	}
	return m.FindOverviewByIDFunc(ctx, id)
}

func (m *MockOverviewRepository) FindAllOverviews(ctx context.Context) ([]domain.OverviewDashboard, error) {
	if m.FindAllOverviewsFunc == nil {
		panic("called not defined FindAllOverviewsFunc")
	}
	return m.FindAllOverviewsFunc(ctx)
}

func (m *MockOverviewRepository) UpdateOverview(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
	if m.UpdateOverviewFunc == nil {
		panic("called not defined UpdateOverviewFunc")
	}
	return m.UpdateOverviewFunc(ctx, overview)
}

func (m *MockOverviewRepository) DeleteOverview(ctx context.Context, id string) error {
	if m.DeleteOverviewFunc == nil {
		panic("called not defined DeleteOverviewFunc")
	}
	return m.DeleteOverviewFunc(ctx, id)
}

// MockWidgetRepository is a function-based mock implementation of overviews.WidgetRepository.
type MockWidgetRepository struct {
	CreateWidgetFunc               func(ctx context.Context, widget domain.Widget) (domain.Widget, error)
	FindWidgetByIDFunc             func(ctx context.Context, id string) (domain.Widget, error)
	FindWidgetsByDashboardIDFunc   func(ctx context.Context, dashboardID string) ([]domain.Widget, error)
	UpdateWidgetFunc               func(ctx context.Context, widget domain.Widget) (domain.Widget, error)
	DeleteWidgetFunc               func(ctx context.Context, id string) error
	DeleteWidgetsByDashboardIDFunc func(ctx context.Context, dashboardID string) error
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

// OverviewRepositoryContractTesting runs all contract tests for an OverviewRepository implementation.
func OverviewRepositoryContractTesting(t *testing.T, repo overviews.OverviewRepository) {
	ctx := context.Background()

	t.Run("Contract: CreateOverview and FindOverviewByID retrieves stored overview", func(t *testing.T) {
		ov := domain.OverviewDashboard{
			ID:        "ov-contract-1",
			Name:      "Tableau Général",
			Order:     1,
			CreatedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt: time.Now().UTC().Truncate(time.Second),
		}

		created, err := repo.CreateOverview(ctx, ov)
		require.NoError(t, err)
		assert.Equal(t, ov.ID, created.ID)
		assert.Equal(t, ov.Name, created.Name)
		assert.Equal(t, ov.Order, created.Order)

		found, err := repo.FindOverviewByID(ctx, ov.ID)
		require.NoError(t, err)
		assert.Equal(t, ov.ID, found.ID)
		assert.Equal(t, ov.Name, found.Name)
		assert.Equal(t, ov.Order, found.Order)
	})

	t.Run("Contract: FindOverviewByID returns ErrOverviewNotFound for missing overview", func(t *testing.T) {
		_, err := repo.FindOverviewByID(ctx, "non-existent-overview")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOverviewNotFound)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: FindAllOverviews returns overviews in ascending order", func(t *testing.T) {
		o1 := domain.OverviewDashboard{
			ID:        "ov-sort-2",
			Name:      "Overview 2",
			Order:     20,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		o2 := domain.OverviewDashboard{
			ID:        "ov-sort-1",
			Name:      "Overview 1",
			Order:     10,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		_, err := repo.CreateOverview(ctx, o1)
		require.NoError(t, err)
		_, err = repo.CreateOverview(ctx, o2)
		require.NoError(t, err)

		all, err := repo.FindAllOverviews(ctx)
		require.NoError(t, err)
		require.True(t, len(all) >= 2)

		for i := 0; i < len(all)-1; i++ {
			assert.True(t, all[i].Order <= all[i+1].Order, "Overviews must be sorted by Order ascending")
		}
	})

	t.Run("Contract: UpdateOverview modifies overview details", func(t *testing.T) {
		ov := domain.OverviewDashboard{
			ID:        "ov-update-1",
			Name:      "Overview Original",
			Order:     5,
			CreatedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt: time.Now().UTC().Truncate(time.Second),
		}
		_, err := repo.CreateOverview(ctx, ov)
		require.NoError(t, err)

		ov.Name = "Overview Modifié"
		ov.UpdatedAt = time.Now().UTC().Truncate(time.Second)

		updated, err := repo.UpdateOverview(ctx, ov)
		require.NoError(t, err)
		assert.Equal(t, "Overview Modifié", updated.Name)

		found, err := repo.FindOverviewByID(ctx, ov.ID)
		require.NoError(t, err)
		assert.Equal(t, "Overview Modifié", found.Name)
	})

	t.Run("Contract: UpdateOverview returns ErrOverviewNotFound for non-existent overview", func(t *testing.T) {
		missing := domain.OverviewDashboard{
			ID:        "ov-missing-update",
			Name:      "Inconnu",
			Order:     99,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_, err := repo.UpdateOverview(ctx, missing)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOverviewNotFound)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: DeleteOverview removes an existing overview", func(t *testing.T) {
		ov := domain.OverviewDashboard{
			ID:        "ov-del-1",
			Name:      "À Supprimer",
			Order:     50,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_, err := repo.CreateOverview(ctx, ov)
		require.NoError(t, err)

		err = repo.DeleteOverview(ctx, ov.ID)
		require.NoError(t, err)

		_, err = repo.FindOverviewByID(ctx, ov.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOverviewNotFound)
	})

	t.Run("Contract: DeleteOverview returns ErrOverviewNotFound for missing overview", func(t *testing.T) {
		err := repo.DeleteOverview(ctx, "ov-missing-delete")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOverviewNotFound)
		assert.True(t, domain.IsDomainError(err))
	})
}

// WidgetRepositoryContractTesting runs all contract tests for a WidgetRepository implementation.
func WidgetRepositoryContractTesting(t *testing.T, repo overviews.WidgetRepository, overviewRepo overviews.OverviewRepository) {
	ctx := context.Background()

	parent := domain.OverviewDashboard{
		ID:        "ov-parent-widget-test",
		Name:      "Dashboard For Widgets",
		Order:     0,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, _ = overviewRepo.CreateOverview(ctx, parent)

	t.Run("Contract: CreateWidget and FindWidgetByID retrieves stored widget", func(t *testing.T) {
		w := domain.Widget{
			ID:          "widget-contract-1",
			DashboardID: parent.ID,
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Automatisations Salon",
			Order:       1,
			Config: domain.WidgetConfig{
				EntityIDs: []string{"automation.eteindre_tout", "automation.cinema"},
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
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		w1 := domain.Widget{
			ID:          "widget-dash-sort-1",
			DashboardID: parent.ID,
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Widget 1",
			Order:       5,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
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
			Config: domain.WidgetConfig{
				EntityIDs: []string{"automation.one"},
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
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
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
			Title:       "À détruire",
			Order:       9,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
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
}
