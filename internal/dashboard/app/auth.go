package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/JLugagne/egauth/revocation"
	"github.com/JLugagne/egauth/tokens/basic"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/accounts"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/invites"
	"github.com/google/uuid"
)

const (
	// defaultPendingTTL is how long a device may wait for owner approval before its
	// pending enrollment is swept.
	defaultPendingTTL = 15 * time.Minute
	// defaultInviteTTL is how long an administrator-minted invitation stays valid.
	defaultInviteTTL = 15 * time.Minute
	// reasonPrivilegeGranted marks an access-token-only revocation published after a role
	// promotion: the target's current access token is rejected so its next request re-issues
	// one with the new scopes, while its refresh family and live socket stay valid.
	reasonPrivilegeGranted revocation.Reason = "privilege_granted"
)

var (
	// ErrInvalidEnrollment is returned when a pending enrollment id is unknown or expired.
	ErrInvalidEnrollment = errors.New("invalid or expired enrollment")
	// ErrPendingApproval is returned when a pending enrollment has not been approved yet.
	ErrPendingApproval = errors.New("enrollment is awaiting approval")
	// ErrInvalidInvite is returned when an invitation is unknown, expired, revoked or used.
	ErrInvalidInvite = errors.New("invalid or expired invitation")
)

// PendingEnrollment is an in-flight device enrollment waiting for owner approval. It carries no
// secret: the device that started it redeems an approved enrollment through its opaque pending
// cookie, and the operator approves it from the setup panel.
type PendingEnrollment struct {
	PendingID string
	DeviceID  string
	Label     string
	Approved  bool
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Auth implements the authentication use-cases: first-run owner bootstrap, owner-approved
// pending enrollments, single-use invitations, rescue recovery and account lookup. Pending
// enrollment state is kept in memory and is intentionally lost on restart.
type Auth struct {
	accounts       accounts.AccountRepository
	invites        invites.Repository
	issuer         *basic.Issuer
	now            func() time.Time
	ttl            time.Duration
	inviteTTL      time.Duration
	rescueEnabled  bool
	rescueConsumed bool
	mu             sync.Mutex
	pending        map[string]*PendingEnrollment
	revoker        RefreshRevoker
	revocations    revocation.Bus
	sessionCloser  SessionCloser
}

// NewAuth builds the authentication use-cases over its outbound dependencies. When
// rescueEnabled is true, the next unauthenticated device to connect may claim the owner role
// even when accounts already exist, and the mode is consumed after that single use.
func NewAuth(accountRepo accounts.AccountRepository, inviteRepo invites.Repository, issuer *basic.Issuer, revoker RefreshRevoker, revocations revocation.Bus, rescueEnabled bool) *Auth {
	return &Auth{
		accounts:      accountRepo,
		invites:       inviteRepo,
		issuer:        issuer,
		revoker:       revoker,
		revocations:   revocations,
		now:           time.Now,
		ttl:           defaultPendingTTL,
		inviteTTL:     defaultInviteTTL,
		rescueEnabled: rescueEnabled,
		pending:       make(map[string]*PendingEnrollment),
	}
}

// ScopesFor maps an account role to the capability scopes carried by its tokens.
func ScopesFor(role domain.Role) []string {
	scopes := []string{"dashboard"}
	if role == domain.RoleOwner || role == domain.RoleAdmin {
		scopes = append(scopes, "setup:manage")
	}
	return scopes
}

// ClaimsFor builds the JWT claims for an account.
func ClaimsFor(acct domain.Account) basic.Claims {
	return basic.Claims{
		Subject: uuid.MustParse(acct.ID),
		Roles:   []string{string(acct.Role)},
		Scopes:  ScopesFor(acct.Role),
		AMR:     []string{"device"},
	}
}

// ListPending returns the live pending enrollments, oldest first.
func (a *Auth) ListPending(_ context.Context) []PendingEnrollment {
	a.sweep()
	a.mu.Lock()
	defer a.mu.Unlock()

	out := make([]PendingEnrollment, 0, len(a.pending))
	for _, rec := range a.pending {
		out = append(out, *rec)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

// GetAccount loads an account by device id.
func (a *Auth) GetAccount(ctx context.Context, id string) (domain.Account, error) {
	return a.accounts.FindByID(ctx, id)
}

func (a *Auth) sweep() {
	now := a.now()
	a.mu.Lock()
	defer a.mu.Unlock()
	for id, rec := range a.pending {
		if now.After(rec.ExpiresAt) {
			delete(a.pending, id)
		}
	}
}

func randomID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

type RefreshRevoker interface {
	RevokeAllRefreshTokensForUser(ctx context.Context, tenantID string, userID uuid.UUID) error
}

func (a *Auth) ListDevices(ctx context.Context) ([]domain.Account, error) {
	return a.accounts.FindAll(ctx)
}

// publishRevocation publishes an account-scoped revocation so already-issued access tokens
// are rejected before their TTL expires, not just the refresh family.
func (a *Auth) publishRevocation(ctx context.Context, subject uuid.UUID, reason revocation.Reason) {
	if a.revocations == nil {
		return
	}
	_ = a.revocations.Publish(ctx, revocation.Revocation{
		TargetType: revocation.TargetUser,
		TargetID:   subject.String(),
		Scope:      revocation.ScopeAll,
		Reason:     reason,
		CutoffTime: a.now().UTC(),
	})
}

type SessionCloser interface {
	CloseUser(subject string)
}

func (a *Auth) SetSessionCloser(c SessionCloser) {
	a.sessionCloser = c
}

func (a *Auth) closeUser(subject string) {
	if a.sessionCloser != nil {
		a.sessionCloser.CloseUser(subject)
	}
}

func canManageAccount(actor, target domain.Account) error {
	switch actor.Role {
	case domain.RoleOwner:
		return nil
	case domain.RoleAdmin:
		if target.Role == domain.RoleOwner {
			return domain.ErrForbidden
		}
		return nil
	default:
		return domain.ErrForbidden
	}
}

func (a *Auth) guardLastOwner(ctx context.Context, target domain.Account) error {
	if target.Role != domain.RoleOwner || target.Status != domain.StatusActive {
		return nil
	}
	all, err := a.accounts.FindAll(ctx)
	if err != nil {
		return err
	}
	owners := 0
	for _, acct := range all {
		if acct.Role == domain.RoleOwner && acct.Status == domain.StatusActive {
			owners++
		}
	}
	if owners <= 1 {
		return domain.ErrForbidden
	}
	return nil
}

func (a *Auth) RevokeDevice(ctx context.Context, actor domain.Account, id string) (domain.Account, error) {
	target, err := a.accounts.FindByID(ctx, id)
	if err != nil {
		return domain.Account{}, err
	}
	if err := canManageAccount(actor, target); err != nil {
		return domain.Account{}, err
	}
	if err := a.guardLastOwner(ctx, target); err != nil {
		return domain.Account{}, err
	}
	acct, err := a.accounts.Revoke(ctx, id)
	if err != nil {
		return domain.Account{}, err
	}
	subject, err := uuid.Parse(id)
	if err != nil {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	if err := a.revoker.RevokeAllRefreshTokensForUser(ctx, "", subject); err != nil {
		return domain.Account{}, err
	}
	a.publishRevocation(ctx, subject, revocation.ReasonAccountDisabled)
	a.closeUser(subject.String())
	return acct, nil
}

func (a *Auth) SetRole(ctx context.Context, actor domain.Account, id string, role domain.Role) (domain.Account, error) {
	if !role.Valid() {
		return domain.Account{}, domain.ErrInvalidRole
	}
	if actor.Role != domain.RoleOwner && actor.Role != domain.RoleAdmin {
		return domain.Account{}, domain.ErrForbidden
	}
	target, err := a.accounts.FindByID(ctx, id)
	if err != nil {
		return domain.Account{}, err
	}
	if err := canManageAccount(actor, target); err != nil {
		return domain.Account{}, err
	}
	if actor.Role != domain.RoleOwner && role == domain.RoleOwner {
		return domain.Account{}, domain.ErrForbidden
	}
	if role != domain.RoleOwner {
		if err := a.guardLastOwner(ctx, target); err != nil {
			return domain.Account{}, err
		}
	}
	updated, err := a.accounts.SetRole(ctx, id, role)
	if err != nil {
		return domain.Account{}, err
	}
	subject, err := uuid.Parse(id)
	if err != nil {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	if role == target.Role {
		return updated, nil
	}
	if roleRank(role) > roleRank(target.Role) {
		// Promotion: keep the device signed in. Reject the already-issued access token so the
		// next request re-issues one carrying the new scopes, but leave the refresh family
		// and the live socket untouched.
		a.publishRevocation(ctx, subject, reasonPrivilegeGranted)
		return updated, nil
	}
	// Demotion: privileges are reduced, so end the session immediately — revoke the refresh
	// family, reject issued access tokens and close live sockets.
	if err := a.revoker.RevokeAllRefreshTokensForUser(ctx, "", subject); err != nil {
		return domain.Account{}, err
	}
	a.publishRevocation(ctx, subject, revocation.ReasonLogoutEverywhere)
	a.closeUser(subject.String())
	return updated, nil
}

func (a *Auth) SetLabel(ctx context.Context, actor domain.Account, id string, label string) error {
	trimmed := strings.TrimSpace(label)
	if n := utf8.RuneCountInString(trimmed); n < 1 || n > 64 {
		return domain.ErrInvalidLabel
	}
	target, err := a.accounts.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := canManageAccount(actor, target); err != nil {
		return err
	}
	return a.accounts.UpdateLabel(ctx, id, trimmed)
}

// roleRank orders roles by privilege so a change can be classified as a promotion or a
// demotion. Unknown roles rank lowest.
func roleRank(role domain.Role) int {
	switch role {
	case domain.RoleOwner:
		return 3
	case domain.RoleAdmin:
		return 2
	case domain.RoleDevice:
		return 1
	default:
		return 0
	}
}
