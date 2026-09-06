package converters_test

import (
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/stretchr/testify/assert"
)

func TestLevelConverters(t *testing.T) {
	t.Run("ToDomainCreateLevel converts request", func(t *testing.T) {
		req := pkgdashboard.CreateLevelRequest{
			Name:      "Rez-de-chaussée",
			IsOutdoor: false,
		}
		domainLevel := converters.ToDomainCreateLevel(req)
		assert.Equal(t, req.Name, domainLevel.Name)
		assert.Equal(t, req.IsOutdoor, domainLevel.IsOutdoor)
	})

	t.Run("ToDomainUpdateLevel converts request with ID", func(t *testing.T) {
		req := pkgdashboard.UpdateLevelRequest{
			Name:      "Étage 1",
			IsOutdoor: true,
		}
		domainLevel := converters.ToDomainUpdateLevel("lvl-123", req)
		assert.Equal(t, "lvl-123", domainLevel.ID)
		assert.Equal(t, req.Name, domainLevel.Name)
		assert.Equal(t, req.IsOutdoor, domainLevel.IsOutdoor)
	})

	t.Run("ToPublicLevel converts domain level", func(t *testing.T) {
		now := time.Now()
		domainLevel := domain.Level{
			ID:        "lvl-abc",
			Name:      "Salon",
			Order:     2,
			IsOutdoor: false,
			CreatedAt: now,
			UpdatedAt: now,
		}
		publicLevel := converters.ToPublicLevel(domainLevel)
		assert.Equal(t, domainLevel.ID, publicLevel.ID)
		assert.Equal(t, domainLevel.Name, publicLevel.Name)
		assert.Equal(t, domainLevel.Order, publicLevel.Order)
		assert.Equal(t, domainLevel.IsOutdoor, publicLevel.IsOutdoor)
		assert.Equal(t, domainLevel.CreatedAt, publicLevel.CreatedAt)
		assert.Equal(t, domainLevel.UpdatedAt, publicLevel.UpdatedAt)
	})

	t.Run("ToPublicLevels converts slice of domain levels", func(t *testing.T) {
		now := time.Now()
		list := []domain.Level{
			{ID: "1", Name: "L1", Order: 0, CreatedAt: now, UpdatedAt: now},
			{ID: "2", Name: "L2", Order: 1, CreatedAt: now, UpdatedAt: now},
		}
		publicList := converters.ToPublicLevels(list)
		assert.Len(t, publicList, 2)
		assert.Equal(t, "1", publicList[0].ID)
		assert.Equal(t, "2", publicList[1].ID)
	})
}

func TestPlanConverters(t *testing.T) {
	t.Run("ToDomainPlan converts save request", func(t *testing.T) {
		req := pkgdashboard.SavePlanRequest{
			Walls: []pkgdashboard.WallSegmentDTO{
				{
					ID:        "w1",
					X1:        10,
					Y1:        20,
					X2:        30,
					Y2:        40,
					Thickness: 12,
				},
			},
			Zones: []pkgdashboard.ZoneDTO{
				{
					ID:    "z1",
					Name:  "Terrasse",
					Color: "#22c55e",
					Points: []pkgdashboard.Point2DDTO{
						{X: 0, Y: 0},
						{X: 50, Y: 0},
						{X: 50, Y: 50},
					},
				},
			},
		}

		plan := converters.ToDomainPlan("lvl-100", req)
		assert.Equal(t, "lvl-100", plan.LevelID)
		assert.Len(t, plan.Walls, 1)
		assert.Equal(t, "w1", plan.Walls[0].ID)
		assert.Equal(t, 10.0, plan.Walls[0].X1)
		assert.Len(t, plan.Zones, 1)
		assert.Equal(t, "z1", plan.Zones[0].ID)
		assert.Len(t, plan.Zones[0].Points, 3)
	})

	t.Run("ToPublicPlan converts domain plan", func(t *testing.T) {
		plan := domain.Plan{
			LevelID: "lvl-200",
			Walls: []domain.WallSegment{
				{
					ID:        "w2",
					X1:        100,
					Y1:        100,
					X2:        200,
					Y2:        100,
					Thickness: 15,
				},
			},
			Zones: []domain.Zone{
				{
					ID:    "z2",
					Name:  "Chambre",
					Color: "#ec4899",
					Points: []domain.Point2D{
						{X: 10, Y: 10},
						{X: 60, Y: 10},
						{X: 60, Y: 60},
					},
				},
			},
		}

		pub := converters.ToPublicPlan(plan)
		assert.Equal(t, "lvl-200", pub.LevelID)
		assert.Len(t, pub.Walls, 1)
		assert.Equal(t, "w2", pub.Walls[0].ID)
		assert.Len(t, pub.Zones, 1)
		assert.Equal(t, "z2", pub.Zones[0].ID)
		assert.Len(t, pub.Zones[0].Points, 3)
	})
}
