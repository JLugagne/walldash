package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/JLugagne/walldash/internal/pkg/logger"
	"github.com/sirupsen/logrus"
)

// SameOrigin rejects cross-origin state-changing requests by comparing the whole canonical origin
// of the request — scheme, host and port — against the Origin header (or the Referer fallback).
//
// It replaces a host-only comparison, which cannot tell a page served over plain HTTP on the same
// host from the HTTPS origin it is impersonating. Safe methods pass through, so reads keep working
// cross-origin; an unsafe request carrying neither Origin nor Referer is refused, because the
// browser has told us nothing to check.
//
// Allow-list entries may be full origins ("https://walldash.example.com") or bare hosts
// ("walldash.example.com"). A full origin is matched exactly. A bare host carries no scheme, so it
// is honoured only on the scheme the request itself arrived on; that keeps the convenient form
// without blessing the other scheme.
func SameOrigin(trusted []string) func(http.Handler) http.Handler {
	trustedOrigins := make([]string, 0, len(trusted))
	trustedHosts := make([]string, 0, len(trusted))
	for _, entry := range trusted {
		entry = strings.TrimSpace(entry)
		switch {
		case entry == "":
			continue
		case strings.Contains(entry, "://"):
			if canonical := canonicalOrigin(entry); canonical != "" {
				trustedOrigins = append(trustedOrigins, canonical)
			}
		default:
			if host := originHostPort(entry); host != "" {
				trustedHosts = append(trustedHosts, host)
			}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}

			raw := strings.TrimSpace(r.Header.Get("Origin"))
			if raw == "" {
				raw = refererOrigin(r.Header.Get("Referer"))
			}
			origin := canonicalOrigin(raw)
			hostPort := originHostPort(raw)
			if origin == "" || hostPort == "" {
				crossSiteBlocked(w)
				return
			}

			own := requestOrigin(r)
			if origin == own {
				next.ServeHTTP(w, r)
				return
			}

			for _, allowed := range trustedOrigins {
				if origin == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			ownScheme, _, _ := strings.Cut(own, "://")
			for _, allowed := range trustedHosts {
				if hostPort == allowed && strings.HasPrefix(origin, ownScheme+"://") {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Logged because the usual cause is a reverse proxy that terminates TLS without
			// forwarding X-Forwarded-Proto: the request then looks like plain HTTP while the
			// browser reports an https origin, so every state change is refused.
			logger.LoggerFromContext(r.Context()).WithFields(logrus.Fields{
				"request_origin": own,
				"origin_header":  origin,
				"path":           r.URL.Path,
				"method":         r.Method,
			}).Warn("cross-site request blocked: origin does not match the request origin or the trusted list")
			crossSiteBlocked(w)
		})
	}
}

// requestOrigin returns the canonical origin the request was addressed to. Behind a reverse proxy
// that terminates TLS the scheme only survives in X-Forwarded-Proto; trusting it can only make the
// comparison stricter, because a mismatch with the Origin header is what gets rejected.
func requestOrigin(r *http.Request) string {
	scheme := "http"
	if isHTTPS(r) {
		scheme = "https"
	}
	return canonicalOrigin(scheme + "://" + r.Host)
}

// refererOrigin returns the origin part of a Referer header, or "" when it carries none.
func refererOrigin(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func crossSiteBlocked(w http.ResponseWriter) {
	http.Error(w, "cross_site_blocked", http.StatusForbidden)
}
