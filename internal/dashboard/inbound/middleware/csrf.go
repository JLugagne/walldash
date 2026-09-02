package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TokenManager manages issuance and validation of anti-CSRF tokens.
type TokenManager interface {
	GenerateToken() string
	ValidateToken(token string) bool
}

// CSRFTokenManager is an in-memory, thread-safe TokenManager implementation.
type CSRFTokenManager struct {
	tokens sync.Map
	ttl    time.Duration
}

// NewCSRFTokenManager creates a new CSRFTokenManager with a default 24-hour TTL.
func NewCSRFTokenManager(ttl ...time.Duration) *CSRFTokenManager {
	tokenTTL := 24 * time.Hour
	if len(ttl) > 0 && ttl[0] > 0 {
		tokenTTL = ttl[0]
	}
	return &CSRFTokenManager{
		ttl: tokenTTL,
	}
}

// GenerateToken generates a cryptographically secure random token and records its expiration.
func (m *CSRFTokenManager) GenerateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	m.tokens.Store(token, time.Now().Add(m.ttl))
	return token
}

// ValidateToken checks whether the provided token exists and has not expired.
func (m *CSRFTokenManager) ValidateToken(token string) bool {
	if token == "" {
		return false
	}
	val, ok := m.tokens.Load(token)
	if !ok {
		return false
	}
	exp, ok := val.(time.Time)
	if !ok || time.Now().After(exp) {
		m.tokens.Delete(token)
		return false
	}
	return true
}

// CSRFConfig defines configuration options for CSRF middleware.
type CSRFConfig struct {
	TokenManager   TokenManager
	AllowedOrigins []string
}

// CSRF returns an HTTP middleware enforcing CSRF protection on state-changing methods.
func CSRF(manager TokenManager, opts ...CSRFConfig) func(http.Handler) http.Handler {
	var allowedOrigins []string
	if len(opts) > 0 {
		allowedOrigins = opts[0].AllowedOrigins
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Exempt safe HTTP methods
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}

			// 1. Origin / Referer check against Host
			origin := r.Header.Get("Origin")
			if origin != "" {
				u, err := url.Parse(origin)
				if err != nil || !isHostAllowed(u.Host, r.Host, allowedOrigins) {
					rejectCSRF(w, "CSRF origin mismatch")
					return
				}
			} else if referer := r.Header.Get("Referer"); referer != "" {
				u, err := url.Parse(referer)
				if err != nil || !isHostAllowed(u.Host, r.Host, allowedOrigins) {
					rejectCSRF(w, "CSRF referer mismatch")
					return
				}
			}

			// 2. Custom header or Token check
			if strings.EqualFold(r.Header.Get("X-Requested-With"), "XMLHttpRequest") {
				next.ServeHTTP(w, r)
				return
			}

			token := r.Header.Get("X-CSRF-Token")
			if token != "" && manager != nil && manager.ValidateToken(token) {
				next.ServeHTTP(w, r)
				return
			}

			rejectCSRF(w, "CSRF token or custom header missing or invalid")
		})
	}
}

func isHostAllowed(targetHost, requestHost string, allowedOrigins []string) bool {
	if strings.EqualFold(targetHost, requestHost) {
		return true
	}
	for _, allowed := range allowedOrigins {
		if allowed == "*" || strings.EqualFold(targetHost, allowed) {
			return true
		}
		if u, err := url.Parse(allowed); err == nil && strings.EqualFold(targetHost, u.Host) {
			return true
		}
	}
	return false
}

func rejectCSRF(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "fail",
		"data": map[string]string{
			"message": message,
		},
		"error": map[string]string{
			"code":    "CSRF_FORBIDDEN",
			"message": message,
		},
	})
}
