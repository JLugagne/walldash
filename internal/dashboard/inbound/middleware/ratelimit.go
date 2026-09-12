package middleware

import (
	"math"
	"net/http"
	"strconv"

	"github.com/JLugagne/egauth/ratelimit"
)

// RateLimit rejects requests with a JSend 429 envelope once the limiter reports the caller
// has exhausted its budget. The limiter key defaults to the client IP.
func RateLimit(limiter ratelimit.Limiter, key ratelimit.KeyFunc) func(http.Handler) http.Handler {
	if limiter == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	if key == nil {
		key = ratelimit.ClientIP
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, retryAfter := limiter.Allow(r.Context(), key(r))
			if !allowed {
				if retryAfter > 0 {
					w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
				}
				WriteJSendError(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
