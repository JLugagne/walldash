package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/JLugagne/egauth"
	"github.com/JLugagne/egauth/tokens"
	"github.com/JLugagne/egauth/tokens/basic"
)

// RequireAuth returns middleware that verifies the access token from the cookie (or
// Authorization header) and transparently refreshes it when a valid refresh cookie is present.
func RequireAuth(issuer *basic.Issuer, cookies tokens.Cookies, revocation tokens.AccessTokenRevocationChecker) func(http.Handler) http.Handler {
	return authMiddleware(issuer, authOptions(cookies, issuer, revocation, nil))
}

// RequireSetupScope returns middleware that also requires the setup:manage scope.
func RequireSetupScope(issuer *basic.Issuer, cookies tokens.Cookies, revocation tokens.AccessTokenRevocationChecker) func(http.Handler) http.Handler {
	return authMiddleware(issuer, authOptions(cookies, issuer, revocation, []string{"setup:manage"}))
}

// authOptions assembles the shared cookie/auto-refresh options, an optional access-token
// revocation checker, and an optional required-scope list.
func authOptions(cookies tokens.Cookies, issuer *basic.Issuer, revocation tokens.AccessTokenRevocationChecker, scopes []string) []tokens.AuthOption[struct{}] {
	opts := []tokens.AuthOption[struct{}]{
		tokens.WithCookieAuth[struct{}](cookies),
		tokens.WithAutoRefresh[struct{}](issuer, cookies),
	}
	if revocation != nil {
		opts = append(opts, tokens.WithAccessTokenRevocation[struct{}](revocation))
	}
	if len(scopes) > 0 {
		opts = append(opts, tokens.WithRequiredScopes[struct{}](scopes...))
	}
	return opts
}

func authMiddleware(issuer *basic.Issuer, opts []tokens.AuthOption[struct{}]) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		protected := basic.RequireAuth(
			issuer,
			func(w http.ResponseWriter, r *http.Request, actor egauth.Actor, _ struct{}) {
				next.ServeHTTP(w, r.WithContext(egauth.ContextWithActor(r.Context(), actor)))
			},
			opts...,
		)
		return JSendAuthErrors(protected)
	}
}

// JSendAuthErrors converts 401/403 responses emitted by the auth middleware into JSend
// envelopes, leaving successful responses untouched.
func JSendAuthErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &authErrorWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if rec.intercepted {
			WriteJSendError(w, rec.status, authErrorCode(rec.status), http.StatusText(rec.status))
		}
	})
}

// WriteJSendError writes a JSend error envelope with the given HTTP status.
func WriteJSendError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "error",
		"code":    code,
		"message": message,
	})
}

// Conditional applies mw to every request for which skip returns false.
func Conditional(mw func(http.Handler) http.Handler, skip func(*http.Request) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		protected := mw(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skip(r) {
				next.ServeHTTP(w, r)
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}

// ExemptAuth reports whether a request skips token authentication: non-API paths, the
// WebSocket handshake (guarded separately) and the explicitly listed public API paths.
func ExemptAuth(publicPaths ...string) func(*http.Request) bool {
	exempt := make(map[string]struct{}, len(publicPaths))
	for _, p := range publicPaths {
		exempt[p] = struct{}{}
	}
	return func(r *http.Request) bool {
		path := r.URL.Path
		if !strings.HasPrefix(path, "/api/") {
			return true
		}
		_, ok := exempt[path]
		return ok
	}
}

func authErrorCode(status int) string {
	if status == http.StatusForbidden {
		return "FORBIDDEN"
	}
	return "UNAUTHORIZED"
}

type authErrorWriter struct {
	http.ResponseWriter
	status      int
	headerSet   bool
	intercepted bool
}

func (w *authErrorWriter) WriteHeader(status int) {
	if w.headerSet {
		return
	}
	w.status = status
	w.headerSet = true
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		w.intercepted = true
		return
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *authErrorWriter) Write(b []byte) (int, error) {
	if !w.headerSet {
		w.WriteHeader(http.StatusOK)
	}
	if w.intercepted {
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}
