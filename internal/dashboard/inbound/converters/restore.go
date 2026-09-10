package converters

import (
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/restore"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
)

// ToDomainRestoreInput converts a restore request payload into domain input,
// preserving every id so placements and widgets stay linked to their parents.
func ToDomainRestoreInput(req pkgdashboard.RestoreRequest) restore.RestoreInput {
	input := restore.RestoreInput{IncludeDevices: req.IncludeDevices}
	for _, lvl := range req.Levels {
		layers := make([]domain.Layer, 0, len(lvl.Level.Layers))
		for _, l := range lvl.Level.Layers {
			layers = append(layers, domain.Layer{Name: l.Name, HideGauges: l.HideGauges})
		}
		input.Levels = append(input.Levels, domain.Level{
			ID:        lvl.Level.ID,
			Name:      lvl.Level.Name,
			Order:     lvl.Level.Order,
			IsOutdoor: lvl.Level.IsOutdoor,
			Layers:    layers,
			CreatedAt: lvl.Level.CreatedAt,
			UpdatedAt: lvl.Level.UpdatedAt,
		})
		if lvl.Plan != nil {
			walls := make([]domain.WallSegment, 0, len(lvl.Plan.Walls))
			for _, w := range lvl.Plan.Walls {
				openings := make([]domain.WallOpening, 0, len(w.Openings))
				for _, o := range w.Openings {
					openings = append(openings, domain.WallOpening{
						ID:        o.ID,
						Type:      o.Type,
						Offset:    o.Offset,
						Width:     o.Width,
						FlipSide:  o.FlipSide,
						FlipHinge: o.FlipHinge,
						HideDoor:  o.HideDoor,
					})
				}
				walls = append(walls, domain.WallSegment{
					ID:        w.ID,
					X1:        w.X1,
					Y1:        w.Y1,
					X2:        w.X2,
					Y2:        w.Y2,
					Thickness: w.Thickness,
					Openings:  openings,
				})
			}
			zones := make([]domain.Zone, 0, len(lvl.Plan.Zones))
			for _, z := range lvl.Plan.Zones {
				points := make([]domain.Point2D, 0, len(z.Points))
				for _, p := range z.Points {
					points = append(points, domain.Point2D{X: p.X, Y: p.Y})
				}
				zones = append(zones, domain.Zone{
					ID:             z.ID,
					Name:           z.Name,
					Color:          z.Color,
					Points:         points,
					TempSensor:     z.TempSensor,
					TempMin:        z.TempMin,
					TempMax:        z.TempMax,
					HumiditySensor: z.HumiditySensor,
					HumidityMin:    z.HumidityMin,
					HumidityMax:    z.HumidityMax,
				})
			}
			input.Plans = append(input.Plans, domain.Plan{
				LevelID: lvl.Plan.LevelID,
				Walls:   walls,
				Zones:   zones,
			})
		}
		for _, p := range lvl.Placements {
			input.Placements = append(input.Placements, domain.DevicePlacement{
				ID:           p.ID,
				LevelID:      p.LevelID,
				DeviceID:     p.DeviceID,
				X:            p.X,
				Y:            p.Y,
				Icon:         p.Icon,
				RenderDomain: p.RenderDomain,
				CustomName:   p.CustomName,
				Layer:        p.Layer,
				CreatedAt:    p.CreatedAt,
				UpdatedAt:    p.UpdatedAt,
			})
		}
	}
	for _, ov := range req.Overviews {
		widgets := make([]domain.Widget, 0, len(ov.Widgets))
		for _, w := range ov.Widgets {
			widgets = append(widgets, domain.Widget{
				ID:          w.ID,
				DashboardID: w.DashboardID,
				Type:        w.Type,
				Title:       w.Title,
				Order:       w.Order,
				Col:         w.Col,
				Row:         w.Row,
				ColSpan:     w.ColSpan,
				RowSpan:     w.RowSpan,
				Config: domain.WidgetConfig{
					EntityIDs:    w.Config.EntityIDs,
					Display:      w.Config.Display,
					Labels:       w.Config.Labels,
					Min:          w.Config.Min,
					Max:          w.Config.Max,
					Unit:         w.Config.Unit,
					WeatherMode:  w.Config.WeatherMode,
					WeatherDays:  w.Config.WeatherDays,
					Latitude:     w.Config.Latitude,
					Longitude:    w.Config.Longitude,
					LocationName: w.Config.LocationName,
					Units:        w.Config.Units,
				},
				CreatedAt: w.CreatedAt,
				UpdatedAt: w.UpdatedAt,
			})
		}
		input.Overviews = append(input.Overviews, domain.OverviewDashboard{
			ID:                ov.ID,
			Name:              ov.Name,
			Order:             ov.Order,
			Cols:              ov.Cols,
			Rows:              ov.Rows,
			BackgroundImage:   ov.BackgroundImage,
			BackgroundOpacity: ov.BackgroundOpacity,
			BackgroundBlur:    ov.BackgroundBlur,
			BackgroundDim:     ov.BackgroundDim,
			CreatedAt:         ov.CreatedAt,
			UpdatedAt:         ov.UpdatedAt,
			Widgets:           widgets,
		})
	}
	return input
}

// ToPublicRestore converts a restore summary into its API response.
func ToPublicRestore(summary restore.RestoreSummary) pkgdashboard.RestoreResponse {
	return pkgdashboard.RestoreResponse{
		Levels:     summary.Levels,
		Plans:      summary.Plans,
		Placements: summary.Placements,
		Overviews:  summary.Overviews,
		Widgets:    summary.Widgets,
	}
}
