package main

import (
	"encoding/json"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var Version = "1.2.2"

const (
	// defaultSupervisorURL is the Home Assistant Supervisor API proxy base URL.
	// It is reachable from inside an add-on container when homeassistant_api is enabled.
	defaultSupervisorURL = "http://supervisor/core"
	// supervisorTokenEnv carries the Supervisor API token, injected by the Supervisor.
	supervisorTokenEnv = "SUPERVISOR_TOKEN"
	// addonOptionsPath is where the Supervisor writes the user configuration of the add-on.
	addonOptionsPath = "/data/options.json"
	// addonDataDir is the persistent volume mapped into every add-on container.
	addonDataDir = "/data"
)

// loadAddonOptions reads a flat JSON options file (e.g. /data/options.json)
// into a normalized key/value map. Missing, unreadable or invalid files yield
// an empty map. Nested objects are ignored; scalar values are stringified.
func loadAddonOptions(path string) map[string]string {
	options := map[string]string{}
	data, err := os.ReadFile(path)
	if err != nil {
		return options
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return options
	}
	for key, value := range raw {
		normalized := strings.ToLower(strings.TrimSpace(key))
		switch val := value.(type) {
		case string:
			options[normalized] = val
		case bool:
			options[normalized] = strconv.FormatBool(val)
		case float64:
			options[normalized] = strconv.FormatFloat(val, 'f', -1, 64)
		default:
			continue
		}
	}
	return options
}

// configValue resolves one setting with the following precedence:
// environment variable first, then add-on options file, then fallback.
func configValue(envKey, optionKey, fallback string, options map[string]string) string {
	if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
		return value
	}
	if value, ok := options[strings.ToLower(optionKey)]; ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

// resolveHAConfig selects the Home Assistant endpoint and token.
// When no URL is configured but a Supervisor token is available (i.e. the app
// runs as a Home Assistant add-on), the Supervisor API proxy base is used so users
// do not have to create a long-lived access token manually.
// The Supervisor token is never attached to a user-provided URL, as it would
// only be valid against the Supervisor proxy.
func resolveHAConfig(haURL, haToken, supervisorToken string) (string, string) {
	if haURL == "" && supervisorToken != "" {
		return defaultSupervisorURL, supervisorToken
	}
	return haURL, haToken
}

// defaultDBPath prefers the persistent add-on volume when present,
// so the database survives add-on updates and host reboots.
func defaultDBPath() string {
	if info, err := os.Stat(addonDataDir); err == nil && info.IsDir() {
		return addonDataDir + "/walldash.db"
	}
	return "walldash.db"
}

// resolveDBPath returns the explicit database path when set,
// otherwise the persistent default.
func resolveDBPath(explicit string) string {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit)
	}
	return defaultDBPath()
}

// parseLogLevel converts a log level name to a logrus level,
// defaulting to info on unknown or empty input.
func parseLogLevel(name string) logrus.Level {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "debug":
		return logrus.DebugLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

// haSourceLabel reports where the Home Assistant endpoint comes from:
// "supervisor" when using the Supervisor API proxy, "manual" otherwise.
// It is logged at startup to diagnose connectivity issues (never with tokens).
func haSourceLabel(haURL string) string {
	if haURL == defaultSupervisorURL {
		return "supervisor"
	}
	return "manual"
}

// haHost extracts the hostname from an endpoint URL for safe startup logging.
// It returns "unknown" when the URL cannot be parsed.
func haHost(haURL string) string {
	parsed, err := url.Parse(haURL)
	if err != nil || parsed.Hostname() == "" {
		return "unknown"
	}
	return parsed.Hostname()
}

// resolveAllowedOrigins merges the comma-separated ALLOWED_ORIGINS setting with the
// normalized DOMAIN host, deduplicating case-insensitively while preserving order.
func resolveAllowedOrigins(options map[string]string) []string {
	origins := splitOrigins(configValue("ALLOWED_ORIGINS", "allowed_origins", "", options))
	domain := normalizeDomain(configValue("DOMAIN", "domain", "", options))
	return mergeOrigins(origins, domain)
}

// splitOrigins splits a comma-separated origin list, dropping blank entries.
func splitOrigins(value string) []string {
	var origins []string
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

// normalizeDomain reduces a domain setting to a bare host, accepting an optional
// scheme and dropping any path or query while keeping an optional port.
func normalizeDomain(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	u, err := url.Parse(value)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Host
}

// mergeOrigins appends domain to origins, deduplicating by host case-insensitively.
func mergeOrigins(origins []string, domain string) []string {
	merged := make([]string, 0, len(origins)+1)
	seen := make(map[string]struct{}, len(origins)+1)
	add := func(origin string) {
		key := originHostKey(origin)
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		merged = append(merged, origin)
	}
	for _, origin := range origins {
		add(origin)
	}
	if domain != "" {
		add(domain)
	}
	return merged
}

// originHostKey returns a case-insensitive dedup key: the host of a URL, or the raw value.
func originHostKey(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" || origin == "*" {
		return origin
	}
	if host := normalizeDomain(origin); host != "" {
		return strings.ToLower(host)
	}
	return strings.ToLower(origin)
}
