package app

import (
	"context"
	"errors"
	"strings"

	"github.com/JLugagne/egauth/identity"
	"github.com/JLugagne/egauth/tokens/basic"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/pkg/logger"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EnrollmentStatus describes the outcome of Auth.Connect.
type EnrollmentStatus string

const (
	// EnrollmentAuthenticated means the device received tokens immediately (first-run
	// bootstrap or rescue recovery).
	EnrollmentAuthenticated EnrollmentStatus = "authenticated"
	// EnrollmentPending means the device must wait for an owner or admin to approve it.
	EnrollmentPending EnrollmentStatus = "pending"
)

// Enrollment is the outcome of a device connect attempt.
type Enrollment struct {
	Status  EnrollmentStatus
	Account domain.Account
	Tokens  *basic.TokenPair
	Pending PendingEnrollment
}

// Connect authenticates a device or registers it as pending. The account count is read and the
// owner account written under the auth mutex, so exactly one concurrent first connect can claim
// ownership. When rescue mode is enabled and no rescue has been consumed yet, the next device
// claims ownership even though accounts already exist.
func (a *Auth) Connect(ctx context.Context, label, ip string) (Enrollment, error) {
	label = enrollLabel(label)
	log := logger.LoggerFromContext(ctx)

	a.mu.Lock()
	count, err := a.accounts.Count(ctx)
	if err != nil {
		a.mu.Unlock()
		return Enrollment{}, err
	}
	bootstrap := count == 0
	rescue := false
	if !bootstrap && a.rescueEnabled && !a.rescueConsumed {
		rescue = true
		a.rescueConsumed = true
	}
	if bootstrap || rescue {
		acct, err := a.accounts.Create(ctx, domain.Account{
			ID:     uuid.NewString(),
			Role:   domain.RoleOwner,
			Status: domain.StatusActive,
			Label:  label,
		})
		a.mu.Unlock()
		if err != nil {
			return Enrollment{}, err
		}
		if rescue {
			log.WithFields(logrus.Fields{"event": "rescue_used", "device_id": acct.ID, "label": label, "ip": ip}).
				Warn("rescue mode claimed the owner role; set rescue_mode to false in the add-on options")
		} else {
			log.WithFields(logrus.Fields{"event": "owner_claimed", "device_id": acct.ID, "label": label, "ip": ip}).
				Warn("first device claimed the owner role; complete setup promptly")
		}
		pair, err := a.issuer.IssueTokenPair(ctx, ClaimsFor(acct))
		if err != nil {
			return Enrollment{}, err
		}
		return Enrollment{Status: EnrollmentAuthenticated, Account: acct, Tokens: pair}, nil
	}
	a.mu.Unlock()

	a.sweep()
	pendingID, err := randomID()
	if err != nil {
		return Enrollment{}, err
	}
	now := a.now().UTC()
	rec := &PendingEnrollment{
		PendingID: pendingID,
		DeviceID:  uuid.NewString(),
		Label:     label,
		ExpiresAt: now.Add(a.ttl),
		CreatedAt: now,
	}
	a.mu.Lock()
	a.pending[pendingID] = rec
	a.mu.Unlock()

	log.WithFields(logrus.Fields{"event": "enrollment_pending", "device_id": rec.DeviceID, "label": label, "ip": ip}).
		Info("device is waiting for owner approval")

	return Enrollment{Status: EnrollmentPending, Pending: *rec}, nil
}

// Redeem issues tokens for a pending enrollment once an owner or admin has approved it. It
// reports ErrPendingApproval while the enrollment is still awaiting approval and
// ErrInvalidEnrollment when the pending id is unknown or expired.
func (a *Auth) Redeem(ctx context.Context, pendingID string) (domain.Account, *basic.TokenPair, error) {
	a.sweep()
	a.mu.Lock()
	rec, ok := a.pending[pendingID]
	if !ok {
		a.mu.Unlock()
		return domain.Account{}, nil, ErrInvalidEnrollment
	}
	if !rec.Approved {
		a.mu.Unlock()
		return domain.Account{}, nil, ErrPendingApproval
	}
	deviceID := rec.DeviceID
	delete(a.pending, pendingID)
	a.mu.Unlock()

	acct, err := a.accounts.FindByID(ctx, deviceID)
	if err != nil {
		return domain.Account{}, nil, err
	}
	pair, err := a.issuer.IssueTokenPair(ctx, ClaimsFor(acct))
	if err != nil {
		return domain.Account{}, nil, err
	}
	return acct, pair, nil
}

// ApprovePending creates the device account behind a pending enrollment so the waiting device
// can redeem it. Approving twice returns the already-created account.
func (a *Auth) ApprovePending(ctx context.Context, pendingID string) (domain.Account, error) {
	a.sweep()
	a.mu.Lock()
	rec, ok := a.pending[pendingID]
	if !ok {
		a.mu.Unlock()
		return domain.Account{}, ErrInvalidEnrollment
	}
	if rec.Approved {
		deviceID := rec.DeviceID
		a.mu.Unlock()
		return a.accounts.FindByID(ctx, deviceID)
	}
	rec.Approved = true
	deviceID := rec.DeviceID
	label := rec.Label
	a.mu.Unlock()

	return a.accounts.Create(ctx, domain.Account{
		ID:     deviceID,
		Role:   domain.RoleDevice,
		Status: domain.StatusActive,
		Label:  label,
	})
}

// DenyPending drops a pending enrollment so the waiting device may try again.
func (a *Auth) DenyPending(_ context.Context, pendingID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.pending[pendingID]; !ok {
		return ErrInvalidEnrollment
	}
	delete(a.pending, pendingID)
	return nil
}

// CreateInvite mints a single-use invitation that lets a new device enroll with the given role.
// Only the device and admin roles may be invited; owner is reserved for bootstrap and rescue.
// The returned token is the only copy of the plaintext invitation secret.
func (a *Auth) CreateInvite(ctx context.Context, actorID string, role domain.Role) (domain.Invite, string, error) {
	if !domain.InvitableRole(role) {
		return domain.Invite{}, "", domain.ErrInvalidInviteRole
	}
	token, selector, verifierHash, err := identity.GenerateVerificationToken()
	if err != nil {
		return domain.Invite{}, "", err
	}
	invite, err := a.invites.Create(ctx, domain.Invite{
		Selector:     selector,
		VerifierHash: verifierHash,
		Role:         role,
		CreatedBy:    actorID,
		ExpiresAt:    a.now().UTC().Add(a.inviteTTL),
	})
	if err != nil {
		return domain.Invite{}, "", err
	}
	return invite, token, nil
}

// RedeemInvite consumes an invitation token and enrolls the device with the invited role.
func (a *Auth) RedeemInvite(ctx context.Context, token, label string) (domain.Account, *basic.TokenPair, error) {
	selector, verifier, ok := identity.SplitVerificationToken(strings.TrimSpace(token))
	if !ok {
		return domain.Account{}, nil, ErrInvalidInvite
	}
	stored, err := a.invites.FindBySelector(ctx, selector)
	if err != nil {
		if errors.Is(err, domain.ErrInviteNotFound) {
			return domain.Account{}, nil, ErrInvalidInvite
		}
		return domain.Account{}, nil, err
	}
	if stored.ConsumedAt != nil || stored.RevokedAt != nil || !stored.ExpiresAt.After(a.now().UTC()) {
		return domain.Account{}, nil, ErrInvalidInvite
	}
	if !identity.CompareVerifier(stored.VerifierHash, verifier) {
		return domain.Account{}, nil, ErrInvalidInvite
	}
	deviceID := uuid.NewString()
	consumed, err := a.invites.Consume(ctx, selector, deviceID)
	if err != nil {
		if domain.IsDomainError(err) {
			return domain.Account{}, nil, ErrInvalidInvite
		}
		return domain.Account{}, nil, err
	}
	acct, err := a.accounts.Create(ctx, domain.Account{
		ID:     deviceID,
		Role:   consumed.Role,
		Status: domain.StatusActive,
		Label:  enrollLabel(label),
	})
	if err != nil {
		return domain.Account{}, nil, err
	}
	pair, err := a.issuer.IssueTokenPair(ctx, ClaimsFor(acct))
	if err != nil {
		return domain.Account{}, nil, err
	}
	return acct, pair, nil
}

// ListInvites returns every invitation, newest first.
func (a *Auth) ListInvites(ctx context.Context) ([]domain.Invite, error) {
	return a.invites.FindAll(ctx)
}

// RevokeInvite prevents an invitation from being used.
func (a *Auth) RevokeInvite(ctx context.Context, selector string) (domain.Invite, error) {
	return a.invites.Revoke(ctx, selector)
}

// enrollLabel derives a safe, bounded device label from the submitted value (typically the
// User-Agent). It never fails: an empty value becomes a generic label.
func enrollLabel(label string) string {
	trimmed := strings.TrimSpace(label)
	if trimmed == "" {
		return "Unknown device"
	}
	runes := []rune(trimmed)
	if len(runes) > 64 {
		return string(runes[:64])
	}
	return trimmed
}
