package middleware

import (
	"net/http"

	"github.com/JLugagne/egauth/tokens"
	"github.com/JLugagne/egauth/tokens/basic"
)

// RequireSetupScopeOnWrites enforces the setup:manage scope on state-changing requests,
// exempting safe methods and the listed public/device-safe paths. Read endpoints stay
// available to every authenticated account.
func RequireSetupScopeOnWrites(issuer *basic.Issuer, cookies tokens.Cookies, revocation tokens.AccessTokenRevocationChecker, exemptPaths []string) func(http.Handler) http.Handler {
	scoped := RequireSetupScope(issuer, cookies, revocation)
	exempt := make(map[string]struct{}, len(exemptPaths))
	for _, p := range exemptPaths {
		exempt[p] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		guarded := scoped(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}
			if _, ok := exempt[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}
			guarded.ServeHTTP(w, r)
		})
	}
}
