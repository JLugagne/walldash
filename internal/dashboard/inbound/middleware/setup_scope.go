package middleware

import (
	"net/http"
	"strings"

	"github.com/JLugagne/egauth/tokens"
	"github.com/JLugagne/egauth/tokens/basic"
	"github.com/gorilla/mux"
)

// DeviceSafeRoutePrefix tags a state-changing route as reachable by any authenticated
// device, not just owner accounts holding the setup:manage scope. Tag a route by naming it
// with this prefix at registration time, e.g.:
//
//	r.HandleFunc("/api/automations/{id}/trigger", h.TriggerAutomation).
//		Name(DeviceSafeRoutePrefix + "automation-trigger")
//
// Tagging routes instead of matching literal paths keeps device-safe exemptions correct for
// routes with dynamic segments, which an exact path list can never capture.
const DeviceSafeRoutePrefix = "device-safe:"

// RequireSetupScopeOnWrites enforces the setup:manage scope on state-changing requests,
// exempting safe methods, the listed public paths, and any route tagged as device-safe by
// naming it with DeviceSafeRoutePrefix at registration. Read endpoints stay available to every
// authenticated account.
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
			if route := mux.CurrentRoute(r); route != nil && strings.HasPrefix(route.GetName(), DeviceSafeRoutePrefix) {
				next.ServeHTTP(w, r)
				return
			}
			guarded.ServeHTTP(w, r)
		})
	}
}
