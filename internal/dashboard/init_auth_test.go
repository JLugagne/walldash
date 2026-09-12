package dashboard

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JLugagne/egauth/keystore"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/sqlite"
	"github.com/stretchr/testify/require"
)

func TestResolveTokenSecret(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "secrets.db")
	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	store := sqlite.NewSecretsStore(adapter.DB())

	first, err := resolveTokenSecret(ctx, store, "", nil)
	require.NoError(t, err)
	require.Len(t, first, 32)

	second, err := resolveTokenSecret(ctx, store, "", nil)
	require.NoError(t, err)
	require.Equal(t, first, second)

	require.NoError(t, adapter.Close())

	reopened, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	defer func() { _ = reopened.Close() }()
	third, err := resolveTokenSecret(ctx, sqlite.NewSecretsStore(reopened.DB()), "", nil)
	require.NoError(t, err)
	require.Equal(t, first, third)

	_, err = resolveTokenSecret(ctx, store, "too-short", nil)
	require.Error(t, err)
}

func TestResolveTokenSecretSealedWithKEK(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "secrets.db")
	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	defer func() { _ = adapter.Close() }()
	store := sqlite.NewSecretsStore(adapter.DB())

	kek, err := keystore.NewKEK(bytes.Repeat([]byte{7}, 32))
	require.NoError(t, err)

	secret, err := resolveTokenSecret(ctx, store, "", kek)
	require.NoError(t, err)
	require.Len(t, secret, 32)

	raw, err := store.Get(ctx, sqlite.TokenSecretName)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(raw), "enc:v1:"), "stored secret must be sealed")
	require.NotContains(t, string(raw), secret, "stored secret must not contain the plaintext key")

	// Same KEK on the next boot yields the same signing key.
	again, err := resolveTokenSecret(ctx, store, "", kek)
	require.NoError(t, err)
	require.Equal(t, secret, again)

	// A wrong KEK must fail loudly rather than silently rotate the signing key.
	otherKek, err := keystore.NewKEK(bytes.Repeat([]byte{9}, 32))
	require.NoError(t, err)
	_, err = resolveTokenSecret(ctx, store, "", otherKek)
	require.Error(t, err)

	// Without the KEK the app must refuse to start rather than downgrade to plaintext.
	_, err = resolveTokenSecret(ctx, store, "", nil)
	require.Error(t, err)
}
