package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JLugagne/walldash/frontend"
	dashboard "github.com/JLugagne/walldash/internal/dashboard"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	loadEnvFile(".env")

	options := loadAddonOptions(addonOptionsPath)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(parseLogLevel(configValue("LOG_LEVEL", "log_level", "info", options)))
	port := configValue("PORT", "port", "8080", options)
	dbPath := resolveDBPath(configValue("DB_PATH", "db_path", "", options))
	haURL := configValue("HA_URL", "ha_url", "", options)
	haToken := configValue("HA_TOKEN", "ha_token", "", options)
	haURL, haToken = resolveHAConfig(haURL, haToken, os.Getenv(supervisorTokenEnv))
	if haURL == "" {
		haURL = "http://homeassistant.local:8123"
	}
	frontendDir := os.Getenv("FRONTEND_DIR")
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsStr != "" {
		for _, o := range strings.Split(allowedOriginsStr, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	conf := dashboard.Config{
		DBPath:         dbPath,
		HAUrl:          haURL,
		HAToken:        haToken,
		Version:        Version,
		FrontendDir:    frontendDir,
		AssetsFS:       frontend.FS(),
		AllowedOrigins: allowedOrigins,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := mux.NewRouter()

	dash, err := dashboard.New(ctx, conf, router)
	if err != nil {
		logrus.WithError(err).Fatal("failed to initialize dashboard service")
	}
	defer func() {
		if err := dash.Close(); err != nil {
			logrus.WithError(err).Warn("error closing dashboard service")
		}
	}()

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logrus.WithField("port", port).Info("starting walldash server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.WithError(err).Fatal("server encountered fatal error")
		}
	}()

	<-stop
	logrus.Info("shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Error("server forced to shutdown")
	} else {
		logrus.Info("server exited properly")
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func loadEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}
