package dashboard_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	dashboard "github.com/JLugagne/ha-dash/internal/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardNew(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	router := mux.NewRouter()
	conf := dashboard.Config{
		DBPath:  dbPath,
		Version: "0.1.0-test",
	}

	dash, err := dashboard.New(ctx, conf, router)
	require.NoError(t, err)
	require.NotNil(t, dash)
	defer func() {
		_ = dash.Close()
	}()

	// Verify route registration works
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
