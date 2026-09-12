package dashboard

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/outbound/sqlite"
	"github.com/stretchr/testify/require"
)

func TestResolveTokenSecret(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "secrets.db")
	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	store := sqlite.NewSecretsStore(adapter.DB())

	first, err := resolveTokenSecret(ctx, store, "")
	require.NoError(t, err)
	require.Len(t, first, 32)

	second, err := resolveTokenSecret(ctx, store, "")
	require.NoError(t, err)
	require.Equal(t, first, second)

	require.NoError(t, adapter.Close())

	reopened, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	defer func() { _ = reopened.Close() }()
	third, err := resolveTokenSecret(ctx, sqlite.NewSecretsStore(reopened.DB()), "")
	require.NoError(t, err)
	require.Equal(t, first, third)

	_, err = resolveTokenSecret(ctx, store, "too-short")
	require.Error(t, err)
}
