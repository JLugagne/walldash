package commands

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/JLugagne/egauth/ratelimit"
	"github.com/JLugagne/egauth/revocation"
	"github.com/JLugagne/egauth/tokens"
	"github.com/JLugagne/egauth/tokens/basic"
	"github.com/JLugagne/walldash/internal/dashboard/app"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/middleware"
	"github.com/gorilla/mux"
)

const pendingCookieName = "pending"

// AuthHandler processes the authentication command HTTP requests.
type AuthHandler struct {
	controller     *inbound.Controller
	auth           *app.Auth
	cookies        tokens.Cookies
	issuer         *basic.Issuer
	revoker        tokens.FamilyRevoker
	revocations    revocation.Bus
	allowedOrigins []string
}

// NewAuthHandler builds the authentication command handler.
func NewAuthHandler(controller *inbound.Controller, auth *app.Auth, cookies tokens.Cookies, issuer *basic.Issuer, revoker tokens.FamilyRevoker, revocations revocation.Bus, allowedOrigins []string) *AuthHandler {
	return &AuthHandler{
		controller:     controller,
		auth:           auth,
		cookies:        cookies,
		issuer:         issuer,
		revoker:        revoker,
		revocations:    revocations,
		allowedOrigins: allowedOrigins,
	}
}

// SetupAuthRoutes registers the public authentication command endpoints.
func SetupAuthRoutes(r *mux.Router, h *AuthHandler) {
	limiter := ratelimit.NewTokenBucket(5, 12*time.Second)
	limited := middleware.RateLimit(limiter, nil)
	r.Handle("/api/auth/connect", limited(http.HandlerFunc(h.Connect))).Methods(http.MethodPost)
	// Redeem is deliberately not rate limited: it carries an opaque single-use
	// pending cookie (no guessable credential) and is polled while a device waits
	// for owner approval, so throttling it would break the waiting screen.
	r.HandleFunc("/api/auth/redeem", h.Redeem).Methods(http.MethodPost)
	r.Handle("/api/auth/invite/redeem", limited(http.HandlerFunc(h.RedeemInvite))).Methods(http.MethodPost)
	r.HandleFunc("/api/auth/refresh", h.Refresh).Methods(http.MethodPost)
	r.HandleFunc("/api/auth/logout", h.Logout).Methods(http.MethodPost)
}

// Connect authenticates the first device automatically (or a rescue device), or registers the
// caller as a pending enrollment waiting for owner approval. No secret is ever returned.
func (h *AuthHandler) Connect(w http.ResponseWriter, r *http.Request) {
	outcome, err := h.auth.Connect(r.Context(), r.UserAgent(), clientIP(r))
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}
	if outcome.Status == app.EnrollmentAuthenticated {
		h.cookies.SetAccess(w, outcome.Tokens.AccessToken)
		h.cookies.SetRefresh(w, outcome.Tokens.RefreshToken, outcome.Tokens.RefreshTokenExpiresAt, true)
		clearPendingCookie(w)
		h.controller.SendSuccess(w, r, map[string]any{
			"status": "authenticated",
			"device": devicePayload(outcome.Account),
		})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     pendingCookieName,
		Value:    outcome.Pending.PendingID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   900,
	})
	h.controller.SendSuccess(w, r, map[string]any{
		"status":     "pending",
		"expires_at": outcome.Pending.ExpiresAt,
	})
}

// Refresh rotates the refresh token family and rewrites the auth cookies.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	basic.RefreshHandler(
		h.issuer,
		tokens.WithCookies(h.cookies),
		tokens.WithPersistentRefresh(),
		tokens.WithTrustedOrigins(h.allowedOrigins...),
	)(w, r)
}

// Logout revokes the caller's refresh family and clears the auth cookies.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.publishRevocation(r)
	basic.LogoutHandler(
		h.revoker,
		tokens.WithCookies(h.cookies),
		tokens.WithTrustedOrigins(h.allowedOrigins...),
	)(w, r)
}

// publishRevocation invalidates already-issued access tokens for the logging-out device so
// they cannot be replayed after logout, before the access TTL elapses.
func (h *AuthHandler) publishRevocation(r *http.Request) {
	if h.revocations == nil {
		return
	}
	raw, ok := h.cookies.Refresh(r)
	if !ok || raw == "" {
		return
	}
	rt, err := h.revoker.FindRefreshToken(r.Context(), "", tokens.HashToken(raw))
	if err != nil || rt == nil {
		return
	}
	_ = h.revocations.Publish(r.Context(), revocation.Revocation{
		TenantID:   rt.TenantID,
		TargetType: revocation.TargetUser,
		TargetID:   rt.UserID.String(),
		Scope:      revocation.ScopeAll,
		Reason:     revocation.ReasonLogoutEverywhere,
		CutoffTime: time.Now().UTC(),
	})
}

func pendingCookie(r *http.Request) (string, bool) {
	c, err := r.Cookie(pendingCookieName)
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}

func clearPendingCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     pendingCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first, _, found := strings.Cut(xff, ","); found {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RevokeDevice disables a device account and kills all of its refresh tokens.
func (h *AuthHandler) RevokeDevice(w http.ResponseWriter, r *http.Request) {
	actor, ok := tokens.ActorFromContext(r.Context())
	if !ok {
		middleware.WriteJSendError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	current, err := h.auth.GetAccount(r.Context(), actor.UserID.String())
	if err != nil {
		h.controller.SendFail(w, r, nil, err)
		return
	}
	id := mux.Vars(r)["id"]
	if _, err := h.auth.RevokeDevice(r.Context(), current, id); err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			middleware.WriteJSendError(w, http.StatusForbidden, "FORBIDDEN", "operation not permitted")
		case errors.Is(err, domain.ErrAccountNotFound):
			middleware.WriteJSendError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "account not found")
		default:
			h.controller.SendError(w, r, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SetRole changes the role of a device account; only owners may grant or revoke the owner
// role.
func (h *AuthHandler) SetRole(w http.ResponseWriter, r *http.Request) {
	actor, ok := tokens.ActorFromContext(r.Context())
	if !ok {
		middleware.WriteJSendError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	current, err := h.auth.GetAccount(r.Context(), actor.UserID.String())
	if err != nil {
		h.controller.SendFail(w, r, nil, err)
		return
	}
	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteJSendError(w, http.StatusBadRequest, "INVALID_ROLE", "invalid role")
		return
	}
	updated, err := h.auth.SetRole(r.Context(), current, mux.Vars(r)["id"], domain.Role(body.Role))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			middleware.WriteJSendError(w, http.StatusForbidden, "FORBIDDEN", "operation not permitted")
		case errors.Is(err, domain.ErrInvalidRole):
			middleware.WriteJSendError(w, http.StatusBadRequest, "INVALID_ROLE", "invalid role")
		case errors.Is(err, domain.ErrAccountNotFound):
			middleware.WriteJSendError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "account not found")
		default:
			h.controller.SendError(w, r, err)
		}
		return
	}
	h.controller.SendSuccess(w, r, map[string]any{
		"id":         updated.ID,
		"label":      updated.Label,
		"role":       updated.Role,
		"status":     updated.Status,
		"created_at": updated.CreatedAt,
		"last_seen":  updated.LastSeen,
	})
}

// SetupSetupAuthRoutes registers the protected setup command endpoints on a router whose base
// path is /api/setup/auth.
func SetupSetupAuthRoutes(r *mux.Router, h *AuthHandler) {
	r.HandleFunc("/devices/{id}/revoke", h.RevokeDevice).Methods(http.MethodPost)
	r.HandleFunc("/devices/{id}/role", h.SetRole).Methods(http.MethodPost)
	r.HandleFunc("/devices/{id}/label", h.SetLabel).Methods(http.MethodPost)
	r.HandleFunc("/pending/{id}/approve", h.ApprovePending).Methods(http.MethodPost)
	r.HandleFunc("/pending/{id}/deny", h.DenyPending).Methods(http.MethodPost)
	r.HandleFunc("/invites", h.CreateInvite).Methods(http.MethodPost)
	r.HandleFunc("/invites/{selector}/revoke", h.RevokeInvite).Methods(http.MethodPost)
}

func (h *AuthHandler) SetLabel(w http.ResponseWriter, r *http.Request) {
	actor, ok := tokens.ActorFromContext(r.Context())
	if !ok {
		middleware.WriteJSendError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	current, err := h.auth.GetAccount(r.Context(), actor.UserID.String())
	if err != nil {
		h.controller.SendFail(w, r, nil, err)
		return
	}
	var body struct {
		Label string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteJSendError(w, http.StatusBadRequest, "INVALID_LABEL", "invalid label")
		return
	}
	id := mux.Vars(r)["id"]
	if err := h.auth.SetLabel(r.Context(), current, id, body.Label); err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			middleware.WriteJSendError(w, http.StatusForbidden, "FORBIDDEN", "operation not permitted")
		case errors.Is(err, domain.ErrInvalidLabel):
			middleware.WriteJSendError(w, http.StatusBadRequest, "INVALID_LABEL", "invalid label")
		case errors.Is(err, domain.ErrAccountNotFound):
			middleware.WriteJSendError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "account not found")
		default:
			h.controller.SendError(w, r, err)
		}
		return
	}
	updated, err := h.auth.GetAccount(r.Context(), id)
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}
	h.controller.SendSuccess(w, r, map[string]any{
		"id":         updated.ID,
		"label":      updated.Label,
		"role":       updated.Role,
		"status":     updated.Status,
		"created_at": updated.CreatedAt,
		"last_seen":  updated.LastSeen,
	})
}
