package converters_test

import (
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/stretchr/testify/assert"
)

func TestLevelConverters(t *testing.T) {
	t.Run("ToDomainCreateLevel converts request", func(t *testing.T) {
		req := pkgdashboard.CreateLevelRequest{
			Name:      "Ground Floor",
			IsOutdoor: false,
			Layers:    []pkgdashboard.LayerRequest{{Name: "controls", HideGauges: false}, {Name: "sensors", HideGauges: false}},
		}
		domainLevel := converters.ToDomainCreateLevel(req)
		assert.Equal(t, req.Name, domainLevel.Name)
		assert.Equal(t, req.IsOutdoor, domainLevel.IsOutdoor)
		assert.Len(t, domainLevel.Layers, 2)
		assert.Equal(t, "controls", domainLevel.Layers[0].Name)
		assert.False(t, domainLevel.Layers[0].HideGauges)
		assert.Equal(t, "sensors", domainLevel.Layers[1].Name)
		assert.False(t, domainLevel.Layers[1].HideGauges)
	})

	t.Run("ToDomainUpdateLevel converts request with ID", func(t *testing.T) {
		req := pkgdashboard.UpdateLevelRequest{
			Name:      "1st Floor",
			IsOutdoor: true,
			Layers:    []pkgdashboard.LayerRequest{{Name: "controls", HideGauges: false}, {Name: "hvac", HideGauges: false}},
		}
		domainLevel := converters.ToDomainUpdateLevel("lvl-123", req)
		assert.Equal(t, "lvl-123", domainLevel.ID)
		assert.Equal(t, req.Name, domainLevel.Name)
		assert.Equal(t, req.IsOutdoor, domainLevel.IsOutdoor)
		assert.Len(t, domainLevel.Layers, 2)
		assert.Equal(t, "controls", domainLevel.Layers[0].Name)
		assert.False(t, domainLevel.Layers[0].HideGauges)
		assert.Equal(t, "hvac", domainLevel.Layers[1].Name)
		assert.False(t, domainLevel.Layers[1].HideGauges)
	})

	t.Run("ToPublicLevel converts domain level", func(t *testing.T) {
		now := time.Now()
		domainLevel := domain.Level{
			ID:        "lvl-abc",
			Name:      "Living Room",
			Order:     2,
			IsOutdoor: false,
			Layers:    []domain.Layer{{Name: "controls", HideGauges: false}, {Name: "custom", HideGauges: false}},
			CreatedAt: now,
			UpdatedAt: now,
		}
		publicLevel := converters.ToPublicLevel(domainLevel)
		assert.Equal(t, domainLevel.ID, publicLevel.ID)
		assert.Equal(t, domainLevel.Name, publicLevel.Name)
		assert.Equal(t, domainLevel.Order, publicLevel.Order)
		assert.Equal(t, domainLevel.IsOutdoor, publicLevel.IsOutdoor)
		assert.Equal(t, []pkgdashboard.LayerResponse{{Name: "controls", HideGauges: false}, {Name: "custom", HideGauges: false}}, publicLevel.Layers)
		assert.Equal(t, domainLevel.CreatedAt, publicLevel.CreatedAt)
		assert.Equal(t, domainLevel.UpdatedAt, publicLevel.UpdatedAt)
	})

	t.Run("ToPublicLevel defaults empty layers to controls and sensors", func(t *testing.T) {
		now := time.Now()
		domainLevel := domain.Level{
			ID:        "lvl-def",
			Name:      "Living Room",
			Order:     0,
			CreatedAt: now,
			UpdatedAt: now,
		}
		publicLevel := converters.ToPublicLevel(domainLevel)
		assert.Equal(t, []pkgdashboard.LayerResponse{{Name: "controls", HideGauges: false}, {Name: "sensors", HideGauges: false}}, publicLevel.Layers)
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
					Openings: []pkgdashboard.WallOpeningDTO{
						{
							ID:        "win-1",
							Type:      "window",
							Offset:    15,
							Width:     10,
							FlipSide:  true,
							FlipHinge: true,
							HideDoor:  true,
						},
					},
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
					TempSensor:     "sensor.terrasse_temp",
					HumiditySensor: "sensor.terrasse_hum",
				},
			},
		}

		plan := converters.ToDomainPlan("lvl-100", req)
		assert.Equal(t, "lvl-100", plan.LevelID)
		assert.Len(t, plan.Walls, 1)
		assert.Equal(t, "w1", plan.Walls[0].ID)
		assert.Equal(t, 10.0, plan.Walls[0].X1)
		assert.Len(t, plan.Walls[0].Openings, 1)
		assert.Equal(t, "win-1", plan.Walls[0].Openings[0].ID)
		assert.Equal(t, "window", plan.Walls[0].Openings[0].Type)
		assert.Equal(t, 15.0, plan.Walls[0].Openings[0].Offset)
		assert.Equal(t, 10.0, plan.Walls[0].Openings[0].Width)
		assert.True(t, plan.Walls[0].Openings[0].FlipSide)
		assert.True(t, plan.Walls[0].Openings[0].FlipHinge)
		assert.True(t, plan.Walls[0].Openings[0].HideDoor)
		assert.Len(t, plan.Zones, 1)
		assert.Equal(t, "z1", plan.Zones[0].ID)
		assert.Len(t, plan.Zones[0].Points, 3)
		assert.Equal(t, "sensor.terrasse_temp", plan.Zones[0].TempSensor)
		assert.Equal(t, "sensor.terrasse_hum", plan.Zones[0].HumiditySensor)
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
					Openings: []domain.WallOpening{
						{
							ID:        "door-1",
							Type:      "door",
							Offset:    50,
							Width:     20,
							FlipSide:  true,
							FlipHinge: true,
							HideDoor:  true,
						},
					},
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
					TempSensor:     "sensor.chambre_temp",
					HumiditySensor: "sensor.chambre_hum",
				},
			},
		}

		pub := converters.ToPublicPlan(plan)
		assert.Equal(t, "lvl-200", pub.LevelID)
		assert.Len(t, pub.Walls, 1)
		assert.Equal(t, "w2", pub.Walls[0].ID)
		assert.Len(t, pub.Walls[0].Openings, 1)
		assert.Equal(t, "door-1", pub.Walls[0].Openings[0].ID)
		assert.Equal(t, "door", pub.Walls[0].Openings[0].Type)
		assert.Equal(t, 50.0, pub.Walls[0].Openings[0].Offset)
		assert.Equal(t, 20.0, pub.Walls[0].Openings[0].Width)
		assert.True(t, pub.Walls[0].Openings[0].FlipSide)
		assert.True(t, pub.Walls[0].Openings[0].FlipHinge)
		assert.True(t, pub.Walls[0].Openings[0].HideDoor)
		assert.Len(t, pub.Zones, 1)
		assert.Equal(t, "z2", pub.Zones[0].ID)
		assert.Len(t, pub.Zones[0].Points, 3)
		assert.Equal(t, "sensor.chambre_temp", pub.Zones[0].TempSensor)
		assert.Equal(t, "sensor.chambre_hum", pub.Zones[0].HumiditySensor)
	})
}

func TestZoneConverters(t *testing.T) {
	tempMin := 17.5
	tempMax := 23.0
	humMin := 40.0
	humMax := 60.0

	t.Run("ToDomainZone converts dto to domain with sensors and thresholds", func(t *testing.T) {
		dto := pkgdashboard.ZoneDTO{
			ID:             "zone-1",
			Name:           "Chambre Parentale",
			Color:          "#6366f1",
			Points:         []pkgdashboard.Point2DDTO{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}},
			TempSensor:     "sensor.chambre_temperature",
			TempMin:        &tempMin,
			TempMax:        &tempMax,
			HumiditySensor: "sensor.chambre_humidity",
			HumidityMin:    &humMin,
			HumidityMax:    &humMax,
		}

		z := converters.ToDomainZone(dto)
		assert.Equal(t, dto.ID, z.ID)
		assert.Equal(t, dto.Name, z.Name)
		assert.Equal(t, dto.Color, z.Color)
		assert.Len(t, z.Points, 3)
		assert.Equal(t, 0.0, z.Points[0].X)
		assert.Equal(t, "sensor.chambre_temperature", z.TempSensor)
		assert.Equal(t, &tempMin, z.TempMin)
		assert.Equal(t, &tempMax, z.TempMax)
		assert.Equal(t, "sensor.chambre_humidity", z.HumiditySensor)
		assert.Equal(t, &humMin, z.HumidityMin)
		assert.Equal(t, &humMax, z.HumidityMax)
	})

	t.Run("ToPublicZone converts domain to dto with sensors and thresholds", func(t *testing.T) {
		zone := domain.Zone{
			ID:             "zone-2",
			Name:           "Living Room",
			Color:          "#10b981",
			Points:         []domain.Point2D{{X: 0, Y: 0}, {X: 20, Y: 0}, {X: 20, Y: 20}},
			TempSensor:     "sensor.salon_temp",
			TempMin:        &tempMin,
			TempMax:        &tempMax,
			HumiditySensor: "sensor.salon_humidity",
			HumidityMin:    &humMin,
			HumidityMax:    &humMax,
		}

		dto := converters.ToPublicZone(zone)
		assert.Equal(t, zone.ID, dto.ID)
		assert.Equal(t, zone.Name, dto.Name)
		assert.Equal(t, zone.Color, dto.Color)
		assert.Len(t, dto.Points, 3)
		assert.Equal(t, "sensor.salon_temp", dto.TempSensor)
		assert.Equal(t, &tempMin, dto.TempMin)
		assert.Equal(t, &tempMax, dto.TempMax)
		assert.Equal(t, "sensor.salon_humidity", dto.HumiditySensor)
		assert.Equal(t, &humMin, dto.HumidityMin)
		assert.Equal(t, &humMax, dto.HumidityMax)
	})
}
