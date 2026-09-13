package middleware

import (
	"net/http"
	"net/url"
	"strings"
)

// CORSConfig defines configuration options for CORS middleware.
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// DefaultCORSConfig returns a default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: nil,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With"},
	}
}

// CORS returns an HTTP middleware handling Cross-Origin Resource Sharing (CORS).
func CORS(cfg ...CORSConfig) func(http.Handler) http.Handler {
	config := DefaultCORSConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	methodsStr := strings.Join(config.AllowedMethods, ", ")
	headersStr := strings.Join(config.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			wildcard := len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*"
			switch {
			case wildcard:
				// A wildcard allow-list never reflects the caller origin and never grants credentials.
				w.Header().Set("Access-Control-Allow-Origin", "*")
			case origin != "":
				canonical := canonicalOrigin(origin)
				originHost := originHostPort(origin)
				for _, o := range config.AllowedOrigins {
					if matchesAllowedOrigin(o, canonical, originHost) {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Add("Vary", "Origin")
						break
					}
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", methodsStr)
			w.Header().Set("Access-Control-Allow-Headers", headersStr)

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// canonicalOrigin returns the lowercase scheme://host[:port] form of an origin,
// with default ports removed, or "" when raw is empty, opaque ("null") or lacks
// a usable scheme and host. Used for entries configured with an explicit scheme.
func canonicalOrigin(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "null") {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Host)
	if (scheme == "https" && strings.HasSuffix(host, ":443")) || (scheme == "http" && strings.HasSuffix(host, ":80")) {
		host = host[:len(host)-4]
	}
	return scheme + "://" + host
}

// originHostPort returns the lowercase host[:port] of an origin, or the lowercased
// raw value when it carries no scheme (bare host or host:port). This mirrors the
// historical scheme-insensitive comparison kept for entries configured without a scheme.
func originHostPort(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "null") {
		return ""
	}
	if u, err := url.Parse(raw); err == nil && u.Host != "" {
		return strings.ToLower(u.Host)
	}
	return strings.ToLower(raw)
}

// matchesAllowedOrigin reports whether an allow-list entry covers the request
// origin. Entries carrying an explicit "://" scheme require an exact canonical
// origin match (scheme, host and port); bare host or host:port entries keep the
// legacy scheme-insensitive host match. Opaque "null" entries never match.
func matchesAllowedOrigin(allowed, canonical, hostPort string) bool {
	allowed = strings.TrimSpace(allowed)
	if allowed == "" || strings.EqualFold(allowed, "null") {
		return false
	}
	if strings.Contains(allowed, "://") {
		return canonical != "" && canonicalOrigin(allowed) == canonical
	}
	return hostPort != "" && originHostPort(allowed) == hostPort
}
