package dashboard

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JLugagne/egauth/keystore"
	"github.com/JLugagne/egauth/origin"
	"github.com/JLugagne/egauth/otp"
	otpmemory "github.com/JLugagne/egauth/otp/memory"
	"github.com/JLugagne/egauth/revocation"
	"github.com/JLugagne/egauth/tokens"
	"github.com/JLugagne/egauth/tokens/basic"
	"github.com/JLugagne/walldash/frontend"
	"github.com/JLugagne/walldash/internal/dashboard/app"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/accounts"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/commands"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/middleware"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/queries"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/websocket"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/homeassistant"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/sqlite"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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
	TokenSecret    string
	// SecretKey is an optional 32-byte key-encryption key (KEK). When set, the
	// auto-generated JWT signing secret is envelope-encrypted (AES-256-GCM) before it is
	// persisted, so a database copy or snapshot alone never yields the signing key.
	SecretKey string
}

// Dashboard represents the initialized composition root for the dashboard service.
type Dashboard struct {
	App          *app.App
	Adapter      *sqlite.Adapter
	Hub          *websocket.Hub
	TokenManager middleware.TokenManager
	OTPService   otp.Service
	Issuer       *basic.Issuer
	TokenStore   *sqlite.TokensStore
	Cookies      tokens.Cookies
	Accounts     accounts.AccountRepository
	Auth         *app.Auth
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

	secretStore := sqlite.NewSecretsStore(adapter.DB())

	var kek *keystore.KEK
	if conf.SecretKey != "" {
		k, err := keystore.NewKEK([]byte(conf.SecretKey))
		if err != nil {
			_ = adapter.Close()
			return nil, fmt.Errorf("SECRET_KEY must be exactly 32 bytes: %w", err)
		}
		kek = k
	}
	tokenSecret, err := resolveTokenSecret(ctx, secretStore, conf.TokenSecret, kek)
	if err != nil {
		_ = adapter.Close()
		return nil, err
	}

	cookies := tokens.DefaultCookies()
	otpSvc := otp.NewService(
		otpmemory.NewStore(),
		otp.WithTTL(15*time.Minute),
		otp.WithMaxAttempts(5),
		otp.WithCooldown(30*time.Second),
	)
	tokenStore := sqlite.NewTokensStore(adapter.DB())
	accountRepo := sqlite.NewAccountRepository(adapter.DB())

	errAccountInactive := errors.New("auth: account is missing or revoked")
	issuerCfg := basic.Config{
		Store:      tokenStore,
		Issuer:     "walldash",
		SecretKey:  tokenSecret,
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 60 * 24 * time.Hour,
		// ReuseGracePeriod widens the window during which replaying an already-consumed
		// refresh token is treated as benign concurrency instead of theft. egauth's zero
		// value selects its 10s default, which is too tight for bursty mobile reconnects
		// (a phone waking from sleep can replay a consumed token seconds later); 30s
		// absorbs those retries while keeping a stolen token's replay window short.
		ReuseGracePeriod: 30 * time.Second,
		ClaimsProvider: basic.ClaimsProviderFunc(func(ctx context.Context, userID uuid.UUID, tenantID string) (basic.Claims, error) {
			acct, err := accountRepo.FindByID(ctx, userID.String())
			if err != nil || acct.Status != domain.StatusActive {
				return basic.Claims{}, errAccountInactive
			}
			return app.ClaimsFor(acct), nil
		}),
	}
	if err := issuerCfg.Validate(); err != nil {
		_ = adapter.Close()
		return nil, err
	}
	issuer := basic.NewIssuer(issuerCfg)

	revocationBus := revocation.NewMemBus()
	revocationTracker := tokens.NewRevocationTracker(revocationBus)

	authApp := app.NewAuth(otpSvc, accountRepo, issuer, tokenStore, revocationBus)

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
	router.Use(middleware.SecurityHeaders)
	router.Use(middleware.CORS(corsConfig))

	// Preflight handler so Gorilla Mux matches OPTIONS for all routes
	router.Methods(http.MethodOptions).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handled by CORS middleware
	}))

	tokenManager := middleware.NewCSRFTokenManager()
	if !conf.DisableCSRF {
		trustedOrigins := conf.AllowedOrigins
		router.Use(func(next http.Handler) http.Handler {
			return origin.Middleware(next, origin.WithTrustedOrigins(trustedOrigins...))
		})
	}

	// Register WebSocket endpoint
	router.Handle("/api/ws", basic.ContextMiddleware(
		issuer,
		http.HandlerFunc(wsHub.ServeWS),
		tokens.WithCookieAuth[struct{}](cookies),
		tokens.WithAccessTokenRevocation[struct{}](revocationTracker),
	)).Methods(http.MethodGet)

	// Register inbound routes
	queries.SetupRoutes(router, controller, application, tokenManager)
	commands.SetupRoutes(router, controller, application)
	authCommands := commands.NewAuthHandler(controller, authApp, cookies, issuer, tokenStore, revocationBus, trustedOriginHosts(conf.AllowedOrigins))
	commands.SetupAuthRoutes(router, authCommands)
	authQueries := queries.NewAuthHandler(controller, authApp, cookies, issuer)
	queries.SetupAuthRoutes(router, authQueries)

	setupRouter := router.PathPrefix("/api/setup/auth").Subrouter()
	setupRouter.Use(middleware.RequireSetupScope(issuer, cookies, revocationTracker))
	commands.SetupSetupAuthRoutes(setupRouter, authCommands)
	queries.SetupSetupAuthRoutes(setupRouter, authQueries)

	// Owner-only management operations: every state-changing request requires the
	// setup:manage scope, except the public auth endpoints and the device-safe action route.
	adminWriteExempt := []string{
		"/api/actions",
		"/api/auth/connect",
		"/api/auth/verify",
		"/api/auth/refresh",
		"/api/auth/logout",
	}
	router.Use(middleware.RequireSetupScopeOnWrites(issuer, cookies, revocationTracker, adminWriteExempt))
	queries.SetupExportRoute(router, controller, application, middleware.RequireSetupScope(issuer, cookies, revocationTracker))

	publicAPIPaths := []string{
		"/api/auth/connect",
		"/api/auth/verify",
		"/api/auth/refresh",
		"/api/auth/logout",
		"/api/auth/status",
		"/api/csrf-token",
		"/api/ws",
	}
	router.Use(middleware.Conditional(middleware.RequireAuth(issuer, cookies, revocationTracker), middleware.ExemptAuth(publicAPIPaths...)))

	// Setup static files / SPA fallback
	setupFrontendServing(router, conf.FrontendDir, conf.AssetsFS)

	return &Dashboard{
		App:          application,
		Adapter:      adapter,
		Hub:          wsHub,
		TokenManager: tokenManager,
		OTPService:   otpSvc,
		Issuer:       issuer,
		TokenStore:   tokenStore,
		Cookies:      cookies,
		Accounts:     accountRepo,
		Auth:         authApp,
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

func resolveTokenSecret(ctx context.Context, store *sqlite.SecretsStore, configured string, kek *keystore.KEK) (string, error) {
	const minBytes = 32
	if configured != "" {
		if len(configured) < minBytes {
			return "", fmt.Errorf("TOKEN_SECRET must be at least %d bytes, got %d", minBytes, len(configured))
		}
		return configured, nil
	}
	existing, err := store.Get(ctx, sqlite.TokenSecretName)
	if err != nil {
		return "", err
	}
	if len(existing) > 0 {
		plain, sealed, err := unwrapSecret(existing, kek)
		if err != nil {
			return "", err
		}
		if len(plain) > 0 {
			if kek != nil && !sealed {
				if wrapped, wrapErr := wrapSecret(plain, kek); wrapErr == nil {
					_ = store.Set(ctx, sqlite.TokenSecretName, wrapped)
				}
			}
			return string(plain), nil
		}
	}
	raw := make([]byte, minBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate token secret: %w", err)
	}
	stored, err := wrapSecret(raw, kek)
	if err != nil {
		return "", err
	}
	if err := store.SetIfAbsent(ctx, sqlite.TokenSecretName, stored); err != nil {
		return "", err
	}
	current, err := store.Get(ctx, sqlite.TokenSecretName)
	if err != nil {
		return "", err
	}
	plain, _, err := unwrapSecret(current, kek)
	if err != nil {
		return "", err
	}
	if len(plain) == 0 {
		return "", errors.New("failed to persist generated token secret")
	}
	if kek == nil {
		logrus.Warn("TOKEN_SECRET is not set: generated a JWT signing key and stored it in the database; set TOKEN_SECRET or SECRET_KEY so database snapshots do not expose it")
	}
	return string(plain), nil
}

func trustedOriginHosts(origins []string) []string {
	hosts := make([]string, 0, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if u, err := url.Parse(origin); err == nil && u.Host != "" {
			hosts = append(hosts, u.Host)
			continue
		}
		hosts = append(hosts, origin)
	}
	return hosts
}

// secretEnvelopePrefix marks values in app_secrets that were envelope-encrypted with a KEK.
const secretEnvelopePrefix = "enc:v1:"

func secretEnvelopeAAD() []byte { return []byte("walldash.app_secrets") }

// wrapSecret seals plaintext with the KEK when one is configured, otherwise it returns the
// plaintext unchanged (legacy zero-config behaviour).
func wrapSecret(plain []byte, kek *keystore.KEK) ([]byte, error) {
	if kek == nil {
		return plain, nil
	}
	sealed, err := kek.Seal(plain, secretEnvelopeAAD())
	if err != nil {
		return nil, fmt.Errorf("sealing token secret: %w", err)
	}
	return []byte(secretEnvelopePrefix + base64.StdEncoding.EncodeToString(sealed)), nil
}

// unwrapSecret returns the plaintext behind a stored secret and whether it was sealed. A
// value without the envelope prefix is treated as legacy plaintext and returned as-is.
func unwrapSecret(stored []byte, kek *keystore.KEK) ([]byte, bool, error) {
	if !bytes.HasPrefix(stored, []byte(secretEnvelopePrefix)) {
		return stored, false, nil
	}
	if kek == nil {
		return nil, true, errors.New("stored token secret is encrypted but SECRET_KEY is not set")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(string(stored), secretEnvelopePrefix))
	if err != nil {
		return nil, true, fmt.Errorf("decoding token secret: %w", err)
	}
	plain, err := kek.Open(raw, secretEnvelopeAAD())
	if err != nil {
		return nil, true, fmt.Errorf("opening token secret (wrong SECRET_KEY?): %w", err)
	}
	return plain, true, nil
}
