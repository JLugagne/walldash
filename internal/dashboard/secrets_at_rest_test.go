package dashboard

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JLugagne/egauth/keystore"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/sqlite"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestGeneratedTokenSecretWithoutKEKIsRejected(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "secrets.db")
	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	defer func() { _ = adapter.Close() }()
	store := sqlite.NewSecretsStore(adapter.DB())

	_, err = resolveTokenSecret(ctx, store, "", nil)
	require.Error(t, err, "storing a generated token secret without a key-encryption key must fail instead of writing plaintext")

	raw, err := store.Get(ctx, sqlite.TokenSecretName)
	require.NoError(t, err)
	require.Empty(t, raw, "no token secret must be persisted without a key-encryption key")
}

func TestLegacyPlaintextTokenSecretResealWithoutKEKIsRejected(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "secrets.db")
	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	defer func() { _ = adapter.Close() }()
	store := sqlite.NewSecretsStore(adapter.DB())

	legacy := bytes.Repeat([]byte{3}, 32)
	require.NoError(t, store.Set(ctx, sqlite.TokenSecretName, legacy))

	_, err = resolveTokenSecret(ctx, store, "", nil)
	require.Error(t, err, "a legacy plaintext secret must not be served without a key-encryption key")

	raw, err := store.Get(ctx, sqlite.TokenSecretName)
	require.NoError(t, err)
	require.Equal(t, legacy, raw, "the legacy plaintext must be left untouched rather than silently served")
}

func TestLegacyPlaintextResealWriteFailureIsFatal(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "secrets.db")
	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	legacy := bytes.Repeat([]byte{5}, 32)
	require.NoError(t, sqlite.NewSecretsStore(adapter.DB()).Set(ctx, sqlite.TokenSecretName, legacy))
	require.NoError(t, adapter.Close())

	ro, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro")
	require.NoError(t, err)
	defer func() { _ = ro.Close() }()
	store := sqlite.NewSecretsStore(ro)

	kek, err := keystore.NewKEK(bytes.Repeat([]byte{7}, 32))
	require.NoError(t, err)

	_, err = resolveTokenSecret(ctx, store, "", kek)
	require.Error(t, err, "a failed plaintext-to-KEK migration must fail startup")
	require.Contains(t, err.Error(), "re-encrypting legacy token secret")

	raw, err := store.Get(ctx, sqlite.TokenSecretName)
	require.NoError(t, err)
	require.Equal(t, legacy, raw, "the failed migration must not pretend the secret was re-sealed")
}

func TestConfiguredTokenSecretWinsOverStoredSecret(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "secrets.db")
	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	defer func() { _ = adapter.Close() }()
	store := sqlite.NewSecretsStore(adapter.DB())

	configured := strings.Repeat("k", 32)
	stored := bytes.Repeat([]byte{9}, 32)
	require.NoError(t, store.Set(ctx, sqlite.TokenSecretName, stored))

	got, err := resolveTokenSecret(ctx, store, configured, nil)
	require.NoError(t, err)
	require.Equal(t, configured, got)

	raw, err := store.Get(ctx, sqlite.TokenSecretName)
	require.NoError(t, err)
	require.Equal(t, stored, raw, "an explicit TOKEN_SECRET must not rewrite the stored secret")
}

func TestGeneratedTokenSecretIsSealedWithAutoKEKFile(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "secrets.db")
	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	defer func() { _ = adapter.Close() }()
	store := sqlite.NewSecretsStore(adapter.DB())

	kek, err := loadOrCreateAutoKEK(dbPath)
	require.NoError(t, err)

	secret, err := resolveTokenSecret(ctx, store, "", kek)
	require.NoError(t, err)
	require.Len(t, secret, 32)

	raw, err := store.Get(ctx, sqlite.TokenSecretName)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(raw), secretEnvelopePrefix), "stored secret must be sealed")
	require.NotEqual(t, secret, string(raw), "stored value must not be the live signing key")
	require.NotContains(t, string(raw), secret, "stored value must not contain the plaintext key")

	kekPath := dbPath + ".kek"
	info, err := os.Stat(kekPath)
	require.NoError(t, err, "auto-KEK file must exist next to the database")
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "auto-KEK file must be owner-only")
	keyBytes, err := os.ReadFile(kekPath)
	require.NoError(t, err)
	require.Len(t, keyBytes, 32)
}

func TestAutoKEKSurvivesRestarts(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "secrets.db")
	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	store := sqlite.NewSecretsStore(adapter.DB())

	firstKEK, err := loadOrCreateAutoKEK(dbPath)
	require.NoError(t, err)
	first, err := resolveTokenSecret(ctx, store, "", firstKEK)
	require.NoError(t, err)

	kekPath := dbPath + ".kek"
	keyBytes, err := os.ReadFile(kekPath)
	require.NoError(t, err)
	require.NoError(t, adapter.Close())

	reopened, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	defer func() { _ = reopened.Close() }()
	secondKEK, err := loadOrCreateAutoKEK(dbPath)
	require.NoError(t, err)
	second, err := resolveTokenSecret(ctx, sqlite.NewSecretsStore(reopened.DB()), "", secondKEK)
	require.NoError(t, err)
	require.Equal(t, first, second, "the auto-generated KEK must survive restarts")

	keyBytesAfter, err := os.ReadFile(kekPath)
	require.NoError(t, err)
	require.Equal(t, keyBytes, keyBytesAfter, "restarting must not rotate the auto-KEK file")
}

func TestLegacyPlaintextTokenSecretIsResealedWithKEK(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "secrets.db")
	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	defer func() { _ = adapter.Close() }()
	store := sqlite.NewSecretsStore(adapter.DB())

	legacy := bytes.Repeat([]byte{6}, 32)
	require.NoError(t, store.Set(ctx, sqlite.TokenSecretName, legacy))

	kek, err := keystore.NewKEK(bytes.Repeat([]byte{7}, 32))
	require.NoError(t, err)

	got, err := resolveTokenSecret(ctx, store, "", kek)
	require.NoError(t, err)
	require.Equal(t, string(legacy), got, "the legacy signing key must keep signing existing sessions")

	raw, err := store.Get(ctx, sqlite.TokenSecretName)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(raw), secretEnvelopePrefix), "legacy plaintext must be re-sealed on startup")
	require.NotContains(t, string(raw), string(legacy))
}

func TestSealedTokenSecretFailsClosedWhenAutoKEKRegenerated(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "secrets.db")
	adapter, err := sqlite.New(ctx, dbPath)
	require.NoError(t, err)
	defer func() { _ = adapter.Close() }()
	store := sqlite.NewSecretsStore(adapter.DB())

	kek, err := loadOrCreateAutoKEK(dbPath)
	require.NoError(t, err)
	_, err = resolveTokenSecret(ctx, store, "", kek)
	require.NoError(t, err)

	require.NoError(t, os.Remove(dbPath+".kek"))
	regenerated, err := loadOrCreateAutoKEK(dbPath)
	require.NoError(t, err)

	_, err = resolveTokenSecret(ctx, store, "", regenerated)
	require.Error(t, err, "a sealed secret must fail closed when the KEK is lost or regenerated")
	require.Contains(t, err.Error(), "wrong SECRET_KEY")
}

func TestNewSealsTokenSecretAndFailsClosedWhenAutoKEKIsLost(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "walldash.db")

	dash, err := New(ctx, Config{DBPath: dbPath, Version: "test"}, mux.NewRouter())
	require.NoError(t, err)

	kekPath := dbPath + ".kek"
	info, err := os.Stat(kekPath)
	require.NoError(t, err, "New must create the auto-KEK next to the database")
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	stored, err := sqlite.NewSecretsStore(dash.Adapter.DB()).Get(ctx, sqlite.TokenSecretName)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(stored), secretEnvelopePrefix), "New must seal the generated token secret")
	require.NoError(t, dash.Close())

	require.NoError(t, os.Remove(kekPath))
	_, err = New(ctx, Config{DBPath: dbPath, Version: "test"}, mux.NewRouter())
	require.Error(t, err, "New must fail closed when the auto-KEK no longer matches the sealed secret")
}

func TestNewTOKENSecretWinsAndSkipsAutoKEK(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "walldash.db")
	configured := "0123456789abcdefghijklmnopqrstuv"

	dash, err := New(ctx, Config{DBPath: dbPath, Version: "test", TokenSecret: configured}, mux.NewRouter())
	require.NoError(t, err)
	defer func() { _ = dash.Close() }()

	_, err = os.Stat(dbPath + ".kek")
	require.True(t, errors.Is(err, os.ErrNotExist), "an explicit TOKEN_SECRET must not create a KEK file")
	stored, err := sqlite.NewSecretsStore(dash.Adapter.DB()).Get(ctx, sqlite.TokenSecretName)
	require.NoError(t, err)
	require.Empty(t, stored, "an explicit TOKEN_SECRET must not persist anything")
}
