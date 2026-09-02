package middleware

import (
	"net/http"
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
		AllowedOrigins: []string{"*"},
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
			if origin != "" {
				allowed := false
				for _, o := range config.AllowedOrigins {
					if o == "*" || strings.EqualFold(o, origin) {
						allowed = true
						break
					}
				}
				if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			} else if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
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
