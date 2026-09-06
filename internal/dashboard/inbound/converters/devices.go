package converters

import (
	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
)

// ToPublicDevice converts a domain Device to public DeviceResponse.
func ToPublicDevice(d domain.Device) pkgdashboard.DeviceResponse {
	return pkgdashboard.DeviceResponse{
		ID:          d.ID,
		Name:        d.Name,
		Domain:      d.Domain,
		State:       d.State,
		Attributes:  d.Attributes,
		LastUpdated: d.LastUpdated,
	}
}

// ToPublicDevices converts a slice of domain Devices to public DeviceResponses.
func ToPublicDevices(devices []domain.Device) []pkgdashboard.DeviceResponse {
	if len(devices) == 0 {
		return []pkgdashboard.DeviceResponse{}
	}
	res := make([]pkgdashboard.DeviceResponse, len(devices))
	for i, d := range devices {
		res[i] = ToPublicDevice(d)
	}
	return res
}

// ToDomainSavePlacement converts levelID and public SavePlacementRequest into a domain DevicePlacement.
func ToDomainSavePlacement(levelID string, req pkgdashboard.SavePlacementRequest) domain.DevicePlacement {
	layer := req.Layer
	if layer == "" {
		layer = domain.DefaultPlacementLayer
	}
	return domain.DevicePlacement{
		ID:           req.ID,
		LevelID:      levelID,
		DeviceID:     req.DeviceID,
		X:            req.X,
		Y:            req.Y,
		Icon:         req.Icon,
		RenderDomain: req.RenderDomain,
		CustomName:   req.CustomName,
		Layer:        layer,
	}
}

// ToPublicPlacement converts a domain DevicePlacement into public DevicePlacementResponse.
func ToPublicPlacement(p domain.DevicePlacement) pkgdashboard.DevicePlacementResponse {
	layer := p.Layer
	if layer == "" {
		layer = domain.DefaultPlacementLayer
	}
	return pkgdashboard.DevicePlacementResponse{
		ID:           p.ID,
		LevelID:      p.LevelID,
		DeviceID:     p.DeviceID,
		X:            p.X,
		Y:            p.Y,
		Icon:         p.Icon,
		RenderDomain: p.RenderDomain,
		CustomName:   p.CustomName,
		Layer:        layer,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// ToPublicPlacements converts a slice of domain DevicePlacements into public DevicePlacementResponses.
func ToPublicPlacements(placements []domain.DevicePlacement) []pkgdashboard.DevicePlacementResponse {
	if len(placements) == 0 {
		return []pkgdashboard.DevicePlacementResponse{}
	}
	res := make([]pkgdashboard.DevicePlacementResponse, len(placements))
	for i, p := range placements {
		res[i] = ToPublicPlacement(p)
	}
	return res
}
