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

	"github.com/JLugagne/egauth/otp"
	"github.com/JLugagne/egauth/revocation"
	"github.com/JLugagne/egauth/tokens/basic"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/accounts"
	"github.com/JLugagne/walldash/internal/pkg/logger"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const enrollmentPurpose = "device-enroll"

// ErrInvalidCode is returned when an enrollment pending id or OTP code cannot be verified.
var ErrInvalidCode = errors.New("invalid or expired enrollment code")

// PendingEnrollment is an in-flight device enrollment waiting for OTP verification.
type PendingEnrollment struct {
	PendingID string
	DeviceID  string
	Label     string
	Code      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Auth implements the authentication use-cases: device enrollment, OTP verification and
// account lookup. Pending enrollment state is kept in memory and is intentionally lost on
// restart so a fresh code is required.
type Auth struct {
	otp         otp.Service
	accounts    accounts.AccountRepository
	issuer      *basic.Issuer
	now         func() time.Time
	ttl         time.Duration
	mu          sync.Mutex
	pending     map[string]*PendingEnrollment
	revoker     RefreshRevoker
	revocations revocation.Bus
}

// NewAuth builds the authentication use-cases over its outbound dependencies.
func NewAuth(otpSvc otp.Service, accountRepo accounts.AccountRepository, issuer *basic.Issuer, revoker RefreshRevoker, revocations revocation.Bus) *Auth {
	return &Auth{
		otp:         otpSvc,
		accounts:    accountRepo,
		issuer:      issuer,
		revoker:     revoker,
		revocations: revocations,
		now:         time.Now,
		ttl:         15 * time.Minute,
		pending:     make(map[string]*PendingEnrollment),
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
		AMR:     []string{"otp"},
	}
}

// StartEnrollment issues a fresh OTP for a device lacking authentication and returns the
// pending enrollment record. The plaintext code is never returned to the caller of the HTTP
// endpoint; it is exposed through ListPending and the logs.
func (a *Auth) StartEnrollment(ctx context.Context, label, ip string) (*PendingEnrollment, error) {
	a.sweep()
	pendingID, err := randomID()
	if err != nil {
		return nil, err
	}
	deviceID := uuid.NewString()
	ch, err := a.otp.Issue(ctx, "", uuid.MustParse(deviceID), enrollmentPurpose)
	if err != nil {
		return nil, err
	}
	rec := &PendingEnrollment{
		PendingID: pendingID,
		DeviceID:  deviceID,
		Label:     label,
		Code:      ch.Code,
		ExpiresAt: ch.ExpiresAt,
		CreatedAt: a.now().UTC(),
	}
	a.mu.Lock()
	a.pending[pendingID] = rec
	a.mu.Unlock()

	logger.LoggerFromContext(ctx).WithFields(logrus.Fields{
		"event":     "otp_issued",
		"device_id": deviceID,
		"label":     label,
		"code":      ch.Code,
		"ip":        ip,
	}).Info("otp_issued")

	return rec, nil
}

// VerifyEnrollment exchanges a pending enrollment and its OTP code for an account and a token
// pair. The first account created becomes the owner; later devices are plain devices.
func (a *Auth) VerifyEnrollment(ctx context.Context, pendingID, code string) (domain.Account, *basic.TokenPair, error) {
	a.sweep()
	a.mu.Lock()
	rec, ok := a.pending[pendingID]
	a.mu.Unlock()
	if !ok || a.now().After(rec.ExpiresAt) {
		return domain.Account{}, nil, ErrInvalidCode
	}

	subject, err := uuid.Parse(rec.DeviceID)
	if err != nil {
		return domain.Account{}, nil, ErrInvalidCode
	}
	if err := a.otp.Verify(ctx, "", subject, enrollmentPurpose, code); err != nil {
		return domain.Account{}, nil, ErrInvalidCode
	}

	acct, err := a.accounts.Create(ctx, domain.Account{
		ID:     rec.DeviceID,
		Role:   domain.RoleDevice,
		Status: domain.StatusActive,
		Label:  rec.Label,
	})
	if err != nil {
		return domain.Account{}, nil, err
	}

	pair, err := a.issuer.IssueTokenPair(ctx, ClaimsFor(acct))
	if err != nil {
		return domain.Account{}, nil, err
	}

	a.mu.Lock()
	delete(a.pending, pendingID)
	a.mu.Unlock()

	return acct, pair, nil
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

func (a *Auth) RevokeDevice(ctx context.Context, id string) (domain.Account, error) {
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
	return acct, nil
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

func (a *Auth) SetRole(ctx context.Context, actor domain.Account, id string, role domain.Role) (domain.Account, error) {
	if !role.Valid() {
		return domain.Account{}, domain.ErrInvalidRole
	}
	target, err := a.accounts.FindByID(ctx, id)
	if err != nil {
		return domain.Account{}, err
	}
	if actor.Role != domain.RoleOwner && (role == domain.RoleOwner || target.Role == domain.RoleOwner) {
		return domain.Account{}, domain.ErrForbidden
	}
	return a.accounts.SetRole(ctx, id, role)
}

func (a *Auth) SetLabel(ctx context.Context, id string, label string) error {
	trimmed := strings.TrimSpace(label)
	if n := utf8.RuneCountInString(trimmed); n < 1 || n > 64 {
		return domain.ErrInvalidLabel
	}
	return a.accounts.UpdateLabel(ctx, id, trimmed)
}
