package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestResolveHAConfigUsesSupervisorWhenNothingConfigured(t *testing.T) {
	url, token := resolveHAConfig("", "", "supervisor-secret")
	if url != "http://supervisor/core/api" {
		t.Fatalf("expected supervisor URL, got %q", url)
	}
	if token != "supervisor-secret" {
		t.Fatalf("expected supervisor token, got %q", token)
	}
}

func TestResolveHAConfigKeepsManualConfiguration(t *testing.T) {
	url, token := resolveHAConfig("http://homeassistant.local:8123", "manual-token", "supervisor-secret")
	if url != "http://homeassistant.local:8123" {
		t.Fatalf("expected manual URL to win, got %q", url)
	}
	if token != "manual-token" {
		t.Fatalf("expected manual token to win, got %q", token)
	}
}

func TestResolveHAConfigNeverLeaksSupervisorTokenToCustomURL(t *testing.T) {
	url, token := resolveHAConfig("http://192.168.1.50:8123", "", "supervisor-secret")
	if url != "http://192.168.1.50:8123" {
		t.Fatalf("expected custom URL to be kept, got %q", url)
	}
	if token != "" {
		t.Fatalf("supervisor token must not be attached to a custom URL, got %q", token)
	}
}

func TestResolveHAConfigWithoutSupervisorToken(t *testing.T) {
	url, token := resolveHAConfig("", "", "")
	if url != "" || token != "" {
		t.Fatalf("expected empty result without any configuration, got %q/%q", url, token)
	}
}

func TestConfigValuePrefersEnvironmentOverOptionsFile(t *testing.T) {
	t.Setenv("WALLDASH_TEST_KEY", "from-env")
	options := map[string]string{"walldash_test_key": "from-options"}
	if got := configValue("WALLDASH_TEST_KEY", "walldash_test_key", "fallback", options); got != "from-env" {
		t.Fatalf("expected env to win, got %q", got)
	}
}

func TestConfigValueFallsBackToOptionsFileThenDefault(t *testing.T) {
	options := map[string]string{"ha_url": "http://from-options:8123"}
	if got := configValue("WALLDASH_MISSING_ENV_URL", "ha_url", "fallback", options); got != "http://from-options:8123" {
		t.Fatalf("expected options file value, got %q", got)
	}
	if got := configValue("WALLDASH_MISSING_ENV_OTHER", "missing_key", "fallback", options); got != "fallback" {
		t.Fatalf("expected fallback value, got %q", got)
	}
}

func TestLoadAddonOptionsParsesFlatScalars(t *testing.T) {
	path := filepath.Join(t.TempDir(), "options.json")
	content := `{"ha_url": "http://ha:8123", "port": 8080, "debug": true, "nested": {"a": 1}}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	options := loadAddonOptions(path)
	if options["ha_url"] != "http://ha:8123" {
		t.Fatalf("expected ha_url, got %q", options["ha_url"])
	}
	if options["port"] != "8080" {
		t.Fatalf("expected numeric port to be stringified, got %q", options["port"])
	}
	if options["debug"] != "true" {
		t.Fatalf("expected boolean to be stringified, got %q", options["debug"])
	}
	if _, ok := options["nested"]; ok {
		t.Fatal("nested objects must be ignored")
	}
}

func TestLoadAddonOptionsMissingFileYieldsEmptyMap(t *testing.T) {
	options := loadAddonOptions(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if len(options) != 0 {
		t.Fatalf("expected empty map, got %v", options)
	}
}

func TestResolveDBPathKeepsExplicitValue(t *testing.T) {
	if got := resolveDBPath("/custom/path.db"); got != "/custom/path.db" {
		t.Fatalf("expected explicit path to win, got %q", got)
	}
	if got := resolveDBPath(""); got != defaultDBPath() {
		t.Fatalf("expected persistent default, got %q", got)
	}
}

func TestParseLogLevel(t *testing.T) {
	cases := map[string]logrus.Level{
		"debug":   logrus.DebugLevel,
		"info":    logrus.InfoLevel,
		"warning": logrus.WarnLevel,
		"warn":    logrus.WarnLevel,
		"error":   logrus.ErrorLevel,
		"bogus":   logrus.InfoLevel,
		"":        logrus.InfoLevel,
	}
	for input, expected := range cases {
		if got := parseLogLevel(input); got != expected {
			t.Fatalf("parseLogLevel(%q) = %v, expected %v", input, got, expected)
		}
	}
}
