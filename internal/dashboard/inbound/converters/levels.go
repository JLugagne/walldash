package converters

import (
	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
)

// ToDomainCreateLevel converts a public CreateLevelRequest to a domain Level.
func ToDomainCreateLevel(req pkgdashboard.CreateLevelRequest) domain.Level {
	return domain.Level{
		Name:      req.Name,
		IsOutdoor: req.IsOutdoor,
	}
}

// ToDomainUpdateLevel converts a public UpdateLevelRequest and id to a domain Level.
func ToDomainUpdateLevel(id string, req pkgdashboard.UpdateLevelRequest) domain.Level {
	return domain.Level{
		ID:        id,
		Name:      req.Name,
		IsOutdoor: req.IsOutdoor,
	}
}

// ToPublicLevel converts a domain Level into a public LevelResponse.
func ToPublicLevel(l domain.Level) pkgdashboard.LevelResponse {
	return pkgdashboard.LevelResponse{
		ID:        l.ID,
		Name:      l.Name,
		Order:     l.Order,
		IsOutdoor: l.IsOutdoor,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

// ToPublicLevels converts a slice of domain Levels into a slice of public LevelResponses.
func ToPublicLevels(levels []domain.Level) []pkgdashboard.LevelResponse {
	if len(levels) == 0 {
		return []pkgdashboard.LevelResponse{}
	}
	res := make([]pkgdashboard.LevelResponse, len(levels))
	for i, l := range levels {
		res[i] = ToPublicLevel(l)
	}
	return res
}

// ToDomainPlan converts levelID and SavePlanRequest into a domain Plan.
func ToDomainPlan(levelID string, req pkgdashboard.SavePlanRequest) domain.Plan {
	walls := make([]domain.WallSegment, len(req.Walls))
	for i, w := range req.Walls {
		walls[i] = domain.WallSegment{
			ID:        w.ID,
			X1:        w.X1,
			Y1:        w.Y1,
			X2:        w.X2,
			Y2:        w.Y2,
			Thickness: w.Thickness,
		}
	}

	zones := make([]domain.Zone, len(req.Zones))
	for i, z := range req.Zones {
		pts := make([]domain.Point2D, len(z.Points))
		for j, p := range z.Points {
			pts[j] = domain.Point2D{
				X: p.X,
				Y: p.Y,
			}
		}
		zones[i] = domain.Zone{
			ID:     z.ID,
			Name:   z.Name,
			Color:  z.Color,
			Points: pts,
		}
	}

	return domain.Plan{
		LevelID: levelID,
		Walls:   walls,
		Zones:   zones,
	}
}

// ToPublicPlan converts a domain Plan into a public PlanResponse.
func ToPublicPlan(plan domain.Plan) pkgdashboard.PlanResponse {
	walls := make([]pkgdashboard.WallSegmentDTO, len(plan.Walls))
	for i, w := range plan.Walls {
		walls[i] = pkgdashboard.WallSegmentDTO{
			ID:        w.ID,
			X1:        w.X1,
			Y1:        w.Y1,
			X2:        w.X2,
			Y2:        w.Y2,
			Thickness: w.Thickness,
		}
	}

	zones := make([]pkgdashboard.ZoneDTO, len(plan.Zones))
	for i, z := range plan.Zones {
		pts := make([]pkgdashboard.Point2DDTO, len(z.Points))
		for j, p := range z.Points {
			pts[j] = pkgdashboard.Point2DDTO{
				X: p.X,
				Y: p.Y,
			}
		}
		zones[i] = pkgdashboard.ZoneDTO{
			ID:     z.ID,
			Name:   z.Name,
			Color:  z.Color,
			Points: pts,
		}
	}

	return pkgdashboard.PlanResponse{
		LevelID: plan.LevelID,
		Walls:   walls,
		Zones:   zones,
	}
}
