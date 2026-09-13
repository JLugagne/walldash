package domain

import "time"

// Invite is a single-use, time-bounded credential minted by an owner or admin that lets a new
// device enroll with a predetermined role.
type Invite struct {
	Selector     string
	VerifierHash string
	Role         Role
	CreatedBy    string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	ConsumedAt   *time.Time
	ConsumedBy   string
	RevokedAt    *time.Time
}

// InvitableRole reports whether role may be granted through an invitation. The owner role is
// reserved for first-run bootstrap and rescue mode.
func InvitableRole(role Role) bool {
	return role == RoleDevice || role == RoleAdmin
}
