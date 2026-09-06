package domain

import (
	"context"
	"errors"
)

// Error represents a domain error.
type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

// IsDomainError checks if any error in the chain is a *domain.Error.
func IsDomainError(err error) bool {
	var domainErr *Error
	return errors.As(err, &domainErr)
}

// AsDomainError extracts the first *domain.Error from the error chain.
func AsDomainError(err error) (*Error, bool) {
	var domainErr *Error
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return nil, false
}

type UserID string
type TeamID string
type TenantID string

// Actor represents the authenticated caller context.
type Actor struct {
	UserID   UserID
	TeamID   TeamID
	TenantID TenantID
}

type actorKey struct{}

// ActorFromContext retrieves the Actor from context or returns an empty Actor.
func ActorFromContext(ctx context.Context) Actor {
	if ctx != nil {
		if a, ok := ctx.Value(actorKey{}).(Actor); ok {
			return a
		}
	}
	return Actor{}
}

// ContextWithActor injects the Actor into context.
func ContextWithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}
