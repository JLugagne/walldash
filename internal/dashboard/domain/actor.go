package domain

import (
	"context"
)

type UserID string
type TeamID string
type TenantID string

// Actor defines the identity and tenant context for dashboard operations.
type Actor struct {
	UserID   UserID
	TeamID   TeamID
	TenantID TenantID
}

type actorContextKey struct{}

// ActorFromContext retrieves the Actor from context or returns an empty Actor.
func ActorFromContext(ctx context.Context) Actor {
	if ctx != nil {
		if actor, ok := ctx.Value(actorContextKey{}).(Actor); ok {
			return actor
		}
	}
	return Actor{}
}

// ContextWithActor injects the Actor into the context.
func ContextWithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}
