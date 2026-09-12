package queries

import (
	"net/http"

	"github.com/JLugagne/egauth/tokens"
	"github.com/JLugagne/egauth/tokens/basic"
	"github.com/JLugagne/walldash/internal/dashboard/app"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/middleware"
	"github.com/gorilla/mux"
)

// AuthHandler handles the authentication query HTTP requests.
type AuthHandler struct {
	controller *inbound.Controller
	auth       *app.Auth
	cookies    tokens.Cookies
	issuer     *basic.Issuer
}

// NewAuthHandler builds the authentication query handler.
func NewAuthHandler(controller *inbound.Controller, auth *app.Auth, cookies tokens.Cookies, issuer *basic.Issuer) *AuthHandler {
	return &AuthHandler{
		controller: controller,
		auth:       auth,
		cookies:    cookies,
		issuer:     issuer,
	}
}

// SetupAuthRoutes registers the authentication query endpoints. Status is public; me is
// protected by the global auth middleware.
func SetupAuthRoutes(r *mux.Router, h *AuthHandler) {
	r.HandleFunc("/api/auth/status", h.Status).Methods(http.MethodGet)
	r.HandleFunc("/api/auth/me", h.Me).Methods(http.MethodGet)
}

// Status reports whether the caller holds a valid access token. It never fails.
func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	authenticated := false
	if token, ok := h.cookies.Access(r); ok {
		if _, err := h.issuer.VerifyAccessTokenForTenant(r.Context(), "", token); err == nil {
			authenticated = true
		}
	}
	h.controller.SendSuccess(w, r, map[string]bool{"authenticated": authenticated})
}

// Me returns the account behind the verified access token.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	actor, ok := tokens.ActorFromContext(r.Context())
	if !ok {
		middleware.WriteJSendError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	acct, err := h.auth.GetAccount(r.Context(), actor.UserID.String())
	if err != nil {
		h.controller.SendFail(w, r, nil, err)
		return
	}
	h.controller.SendSuccess(w, r, map[string]any{
		"id":     acct.ID,
		"label":  acct.Label,
		"role":   acct.Role,
		"status": acct.Status,
	})
}

// Pending lists the live device enrollments and their OTP codes for the setup tab.
func (h *AuthHandler) Pending(w http.ResponseWriter, r *http.Request) {
	pending := h.auth.ListPending(r.Context())
	data := make([]map[string]any, 0, len(pending))
	for _, entry := range pending {
		data = append(data, map[string]any{
			"device_id":  entry.DeviceID,
			"label":      entry.Label,
			"code":       entry.Code,
			"expires_at": entry.ExpiresAt,
		})
	}
	h.controller.SendSuccess(w, r, data)
}

// Devices lists every known device account for the setup tab.
func (h *AuthHandler) Devices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.auth.ListDevices(r.Context())
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}
	data := make([]map[string]any, 0, len(devices))
	for _, device := range devices {
		data = append(data, map[string]any{
			"id":         device.ID,
			"label":      device.Label,
			"role":       device.Role,
			"status":     device.Status,
			"created_at": device.CreatedAt,
			"last_seen":  device.LastSeen,
		})
	}
	h.controller.SendSuccess(w, r, data)
}

// SetupSetupAuthRoutes registers the protected setup query endpoints on a router whose
// base path is /api/setup/auth.
func SetupSetupAuthRoutes(r *mux.Router, h *AuthHandler) {
	r.HandleFunc("/pending", h.Pending).Methods(http.MethodGet)
	r.HandleFunc("/devices", h.Devices).Methods(http.MethodGet)
}
