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
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With", "X-CSRF-Token"},
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
				originKey := normalizeOrigin(origin)
				for _, o := range config.AllowedOrigins {
					if strings.EqualFold(normalizeOrigin(o), originKey) {
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

// normalizeOrigin reduces an origin or allowed entry to a comparable host: the host
// of a URL, or the raw value when no host can be parsed (bare host, "*").
func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	if u, err := url.Parse(origin); err == nil && u.Host != "" {
		return u.Host
	}
	return origin
}
