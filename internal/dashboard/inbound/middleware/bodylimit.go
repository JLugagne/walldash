package middleware

import (
	"net/http"
	"strings"
)

// DefaultBodyLimitBytes is the request body cap applied when no rule matches.
const DefaultBodyLimitBytes int64 = 1 << 20

// BodyLimitRule overrides the body size cap for requests whose path matches.
// A rule matches when every non-empty criterion (Prefix, Suffix) matches.
type BodyLimitRule struct {
	Prefix string
	Suffix string
	Limit  int64
}

func (rule BodyLimitRule) matches(path string) bool {
	if rule.Prefix != "" && !strings.HasPrefix(path, rule.Prefix) {
		return false
	}
	if rule.Suffix != "" && !strings.HasSuffix(path, rule.Suffix) {
		return false
	}
	return true
}

// BodyLimit returns middleware that caps request bodies with
// http.MaxBytesReader before the request reaches next. fallback applies to
// paths that match no rule; rules are evaluated in order and the first match
// wins. A non-positive fallback falls back to DefaultBodyLimitBytes.
func BodyLimit(fallback int64, rules ...BodyLimitRule) func(http.Handler) http.Handler {
	if fallback <= 0 {
		fallback = DefaultBodyLimitBytes
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit := fallback
			for _, rule := range rules {
				if rule.matches(r.URL.Path) {
					limit = rule.Limit
					break
				}
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

// DefaultBodyLimitRules returns the application's per-route body size overrides.
func DefaultBodyLimitRules() []BodyLimitRule {
	return []BodyLimitRule{
		{Prefix: "/api/restore", Limit: 32 << 20},
		{Prefix: "/api/levels/import/sh3d", Limit: 64 << 20},
		{Suffix: "/plan/import", Limit: 32 << 20},
	}
}

// DefaultBodyLimit returns the application body limit middleware: a 1 MiB
// default with the DefaultBodyLimitRules overrides.
func DefaultBodyLimit() func(http.Handler) http.Handler {
	return BodyLimit(DefaultBodyLimitBytes, DefaultBodyLimitRules()...)
}
