package domain

import "time"

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleDevice Role = "device"
)

func (r Role) Valid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleDevice:
		return true
	default:
		return false
	}
}

type Status string

const (
	StatusActive  Status = "active"
	StatusRevoked Status = "revoked"
)

func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusRevoked:
		return true
	default:
		return false
	}
}

type Account struct {
	ID        string
	Status    Status
	Role      Role
	Label     string
	CreatedAt time.Time
	LastSeen  *time.Time
	RevokedAt *time.Time
}
