package commands

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JLugagne/egauth/tokens"
	"github.com/JLugagne/walldash/internal/dashboard/app"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/middleware"
	"github.com/gorilla/mux"
)

// Redeem exchanges an approved pending enrollment for tokens. While the enrollment is still
// awaiting approval it returns 202 so the client keeps polling; an unknown or expired pending
// cookie returns 401.
func (h *AuthHandler) Redeem(w http.ResponseWriter, r *http.Request) {
	pendingID, ok := pendingCookie(r)
	if !ok {
		middleware.WriteJSendError(w, http.StatusUnauthorized, "invalid_enrollment", "invalid or expired enrollment")
		return
	}
	acct, pair, err := h.auth.Redeem(r.Context(), pendingID)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrPendingApproval):
			writeJSendSuccess(w, http.StatusAccepted, map[string]any{"status": "pending"})
		case errors.Is(err, app.ErrInvalidEnrollment):
			clearPendingCookie(w)
			middleware.WriteJSendError(w, http.StatusUnauthorized, "invalid_enrollment", "invalid or expired enrollment")
		default:
			h.controller.SendError(w, r, err)
		}
		return
	}
	h.cookies.SetAccess(w, pair.AccessToken)
	h.cookies.SetRefresh(w, pair.RefreshToken, pair.RefreshTokenExpiresAt, true)
	clearPendingCookie(w)
	h.controller.SendSuccess(w, r, map[string]any{
		"status": "authenticated",
		"device": devicePayload(acct),
	})
}

// RedeemInvite consumes an administrator-minted invitation token and signs the device in.
func (h *AuthHandler) RedeemInvite(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteJSendError(w, http.StatusUnauthorized, "invalid_invite", "invalid or expired invitation")
		return
	}
	acct, pair, err := h.auth.RedeemInvite(r.Context(), body.Token, r.UserAgent())
	if err != nil {
		if errors.Is(err, app.ErrInvalidInvite) {
			middleware.WriteJSendError(w, http.StatusUnauthorized, "invalid_invite", "invalid or expired invitation")
			return
		}
		h.controller.SendError(w, r, err)
		return
	}
	h.cookies.SetAccess(w, pair.AccessToken)
	h.cookies.SetRefresh(w, pair.RefreshToken, pair.RefreshTokenExpiresAt, true)
	clearPendingCookie(w)
	h.controller.SendSuccess(w, r, map[string]any{
		"status": "authenticated",
		"device": devicePayload(acct),
	})
}

// CreateInvite mints a single-use invitation for a new device. The plaintext token is returned
// exactly once; only its hash is stored.
func (h *AuthHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	actor, ok := tokens.ActorFromContext(r.Context())
	if !ok {
		middleware.WriteJSendError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteJSendError(w, http.StatusBadRequest, "INVALID_ROLE", "invalid role")
		return
	}
	role := domain.Role(body.Role)
	if role == "" {
		role = domain.RoleDevice
	}
	invite, token, err := h.auth.CreateInvite(r.Context(), actor.UserID.String(), role)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInviteRole) {
			middleware.WriteJSendError(w, http.StatusBadRequest, "INVALID_ROLE", "invalid role")
			return
		}
		h.controller.SendError(w, r, err)
		return
	}
	h.controller.SendSuccess(w, r, map[string]any{
		"selector":   invite.Selector,
		"token":      token,
		"role":       invite.Role,
		"created_by": invite.CreatedBy,
		"created_at": invite.CreatedAt,
		"expires_at": invite.ExpiresAt,
	})
}

// RevokeInvite prevents an invitation from being used.
func (h *AuthHandler) RevokeInvite(w http.ResponseWriter, r *http.Request) {
	selector := mux.Vars(r)["selector"]
	if _, err := h.auth.RevokeInvite(r.Context(), selector); err != nil {
		if errors.Is(err, domain.ErrInviteNotFound) {
			middleware.WriteJSendError(w, http.StatusNotFound, "INVITE_NOT_FOUND", "invitation not found")
			return
		}
		h.controller.SendError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ApprovePending creates the account behind a pending enrollment so its device can sign in.
func (h *AuthHandler) ApprovePending(w http.ResponseWriter, r *http.Request) {
	acct, err := h.auth.ApprovePending(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		if errors.Is(err, app.ErrInvalidEnrollment) {
			middleware.WriteJSendError(w, http.StatusNotFound, "INVALID_ENROLLMENT", "enrollment not found")
			return
		}
		h.controller.SendError(w, r, err)
		return
	}
	h.controller.SendSuccess(w, r, devicePayload(acct))
}

// DenyPending drops a pending enrollment.
func (h *AuthHandler) DenyPending(w http.ResponseWriter, r *http.Request) {
	if err := h.auth.DenyPending(r.Context(), mux.Vars(r)["id"]); err != nil {
		if errors.Is(err, app.ErrInvalidEnrollment) {
			middleware.WriteJSendError(w, http.StatusNotFound, "INVALID_ENROLLMENT", "enrollment not found")
			return
		}
		h.controller.SendError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// devicePayload renders the public shape of a freshly enrolled device.
func devicePayload(acct domain.Account) map[string]any {
	return map[string]any{
		"id":    acct.ID,
		"label": acct.Label,
		"role":  acct.Role,
	}
}

// writeJSendSuccess writes a JSend success envelope with an explicit HTTP status.
func writeJSendSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": data})
}
