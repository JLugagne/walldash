package health

import (
	"context"
)

// Repository defines the contract for health check storage and connectivity verification.
type Repository interface {
	Ping(ctx context.Context) error
}
