package dashboard

import (
	"github.com/JLugagne/walldash/domain"
)

// ActionRequest specifies an action execution request on an entity.
type ActionRequest struct {
	EntityID string `json:"entity_id" validate:"required"`
	Action   string `json:"action" validate:"required"`
}

// ActionResponse represents the result of executing an action.
type ActionResponse struct {
	Status   string `json:"status"`
	EntityID string `json:"entity_id"`
	Action   string `json:"action"`
}

// Validation domain errors for actions
var (
	ErrInvalidActionRequest = &domain.Error{
		Code:    "INVALID_ACTION_REQUEST",
		Message: "invalid action request data",
	}
)
