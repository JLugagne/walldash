package dashboard

import (
	"context"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JLugagne/walldash/frontend"
	"github.com/JLugagne/walldash/internal/dashboard/app"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/commands"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/middleware"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/queries"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/websocket"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/homeassistant"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/sqlite"
	"github.com/gorilla/mux"
)

// Config contains runtime configuration for the dashboard application.
type Config struct {
	DBPath         string
	HAUrl          string
	HAToken        string
	Version        string
	FrontendDir    string
	AssetsFS       fs.FS
	AllowedOrigins []string
	DisableCSRF    bool
}

// Dashboard represents the initialized composition root for the dashboard service.
type Dashboard struct {
	App          *app.App
	Adapter      *sqlite.Adapter
	Hub          *websocket.Hub
	TokenManager middleware.TokenManager
}

// Close gracefully shuts down dashboard resources such as database connections.
func (d *Dashboard) Close() error {
	if d.Hub != nil {
		d.Hub.Stop()
	}
	if d.Adapter != nil {
		return d.Adapter.Close()
	}
	return nil
}

// New initializes the composition root, wiring outbound adapters, core app, and inbound routers.
func New(ctx context.Context, conf Config, router *mux.Router) (*Dashboard, error) {
	if conf.DBPath == "" {
		conf.DBPath = "walldash.db"
	}
	if conf.Version == "" {
		conf.Version = "0.1.0"
	}

	adapter, err := sqlite.New(ctx, conf.DBPath)
	if err != nil {
		return nil, err
	}

	haClient := homeassistant.NewClient(conf.HAUrl, conf.HAToken, nil)
	application := app.New(adapter, adapter, adapter, adapter, haClient, adapter, adapter, adapter, conf.Version)
	controller := inbound.NewController()

	wsHub := websocket.NewHub(application, conf.AllowedOrigins...)
	go wsHub.Run()
	application.SetBroadcaster(wsHub)

	// Middlewares setup
	corsConfig := middleware.DefaultCORSConfig()
	if len(conf.AllowedOrigins) > 0 {
		corsConfig.AllowedOrigins = conf.AllowedOrigins
	}
	router.Use(middleware.CORS(corsConfig))

	// Preflight handler so Gorilla Mux matches OPTIONS for all routes
	router.Methods(http.MethodOptions).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handled by CORS middleware
	}))

	tokenManager := middleware.NewCSRFTokenManager()
	if !conf.DisableCSRF {
		csrfConfig := middleware.CSRFConfig{
			AllowedOrigins: conf.AllowedOrigins,
		}
		router.Use(middleware.CSRF(tokenManager, csrfConfig))
	}

	// Register WebSocket endpoint
	router.HandleFunc("/api/ws", wsHub.ServeWS)

	// Register inbound routes
	queries.SetupRoutes(router, controller, application, tokenManager)
	commands.SetupRoutes(router, controller, application)

	// Setup static files / SPA fallback
	setupFrontendServing(router, conf.FrontendDir, conf.AssetsFS)

	return &Dashboard{
		App:          application,
		Adapter:      adapter,
		Hub:          wsHub,
		TokenManager: tokenManager,
	}, nil
}

func setupFrontendServing(router *mux.Router, frontendDir string, assetsFS fs.FS) {
	if frontendDir != "" {
		if _, err := os.Stat(frontendDir); err == nil {
			serveFromDir(router, frontendDir)
			return
		}
	}

	if assetsFS == nil {
		assetsFS = frontend.FS()
	}

	if assetsFS != nil {
		if _, err := fs.Stat(assetsFS, "index.html"); err == nil {
			serveFromFS(router, assetsFS)
			return
		}
	}
}

func serveFromFS(router *mux.Router, assetsFS fs.FS) {
	fileServer := http.FileServer(http.FS(assetsFS))

	router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}

		cleanPath := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
		if cleanPath == "" || cleanPath == "." {
			cleanPath = "index.html"
		}

		if f, err := assetsFS.Open(cleanPath); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA Fallback: serve index.html for client routes
		indexData, err := fs.ReadFile(assetsFS, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(indexData)
	}))
}

func serveFromDir(router *mux.Router, frontendDir string) {
	fileServer := http.FileServer(http.Dir(frontendDir))

	router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}

		path := filepath.Join(frontendDir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	}))
}
