package dashboard

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/JLugagne/ha-dash/internal/dashboard/app"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/commands"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/queries"
	"github.com/JLugagne/ha-dash/internal/dashboard/outbound/sqlite"
	"github.com/gorilla/mux"
)

// Config contains runtime configuration for the dashboard application.
type Config struct {
	DBPath      string
	HAUrl       string
	HAToken     string
	Version     string
	FrontendDir string
}

// Dashboard represents the initialized composition root for the dashboard service.
type Dashboard struct {
	App     *app.App
	Adapter *sqlite.Adapter
}

// Close gracefully shuts down dashboard resources such as database connections.
func (d *Dashboard) Close() error {
	if d.Adapter != nil {
		return d.Adapter.Close()
	}
	return nil
}

// New initializes the composition root, wiring outbound adapters, core app, and inbound routers.
func New(ctx context.Context, conf Config, router *mux.Router) (*Dashboard, error) {
	if conf.DBPath == "" {
		conf.DBPath = "ha-dash.db"
	}
	if conf.Version == "" {
		conf.Version = "0.1.0"
	}
	if conf.FrontendDir == "" {
		conf.FrontendDir = "frontend/dist"
	}

	adapter, err := sqlite.New(ctx, conf.DBPath)
	if err != nil {
		return nil, err
	}

	application := app.New(adapter, adapter, conf.Version)
	controller := inbound.NewController()

	// Register inbound routes
	queries.SetupRoutes(router, controller, application)
	commands.SetupRoutes(router, controller, application)

	// Setup static files / SPA fallback if frontend directory exists
	setupFrontendServing(router, conf.FrontendDir)

	return &Dashboard{
		App:     application,
		Adapter: adapter,
	}, nil
}

func setupFrontendServing(router *mux.Router, frontendDir string) {
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		return
	}

	fileServer := http.FileServer(http.Dir(frontendDir))

	router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(frontendDir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	}))
}
