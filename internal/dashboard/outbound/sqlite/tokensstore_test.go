package sqlite_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JLugagne/egauth/tokens"
	"github.com/JLugagne/egauth/tokens/basic"
	"github.com/JLugagne/egauth/tokens/storetest"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteTokensStoreContract(t *testing.T) {
	adapter := setupTestDB(t)
	store := sqlite.NewTokensStore(adapter.DB(), 30*time.Second)
	storetest.StoreContractTesting(t, store, false, struct{}{})
}

func TestSQLiteTokensStoreConcurrentConsume(t *testing.T) {
	adapter := setupTestDB(t)
	store := sqlite.NewTokensStore(adapter.DB(), 30*time.Second)
	ctx := context.Background()

	rt := &tokens.RefreshToken{
		Hash:      "concurrent-consume",
		FamilyID:  uuid.Must(uuid.NewV7()),
		UserID:    uuid.Must(uuid.NewV7()),
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}
	require.NoError(t, store.SaveRefreshToken(ctx, "", rt))

	const workers = 12
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- store.ConsumeRefreshToken(ctx, "", "concurrent-consume")
		}()
	}
	wg.Wait()
	close(errs)

	var successes, reused, notFound int
	for err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, tokens.ErrRefreshTokenReused):
			reused++
		case errors.Is(err, tokens.ErrRefreshTokenNotFound):
			notFound++
		default:
			t.Fatalf("unexpected consume error: %v", err)
		}
	}
	assert.Equal(t, 1, successes)
	assert.Equal(t, workers-1, reused)
	assert.Equal(t, 0, notFound)
}

func TestSQLiteTokensStoreConsumeWithinGraceReportsConcurrent(t *testing.T) {
	adapter := setupTestDB(t)
	store := sqlite.NewTokensStore(adapter.DB(), 30*time.Second)
	ctx := context.Background()

	rt := &tokens.RefreshToken{
		Hash:      "grace-concurrent",
		FamilyID:  uuid.Must(uuid.NewV7()),
		UserID:    uuid.Must(uuid.NewV7()),
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}
	require.NoError(t, store.SaveRefreshToken(ctx, "", rt))
	require.NoError(t, store.ConsumeRefreshToken(ctx, "", rt.Hash))

	err := store.ConsumeRefreshToken(ctx, "", rt.Hash)
	require.Error(t, err)
	assert.ErrorIs(t, err, tokens.ErrRefreshConcurrent)
	assert.ErrorIs(t, err, tokens.ErrRefreshTokenReused)
}

func TestSQLiteTokensStoreConsumeAfterGraceReportsReused(t *testing.T) {
	adapter := setupTestDB(t)
	store := sqlite.NewTokensStore(adapter.DB(), 30*time.Second)
	ctx := context.Background()

	rt := &tokens.RefreshToken{
		Hash:      "grace-reused",
		FamilyID:  uuid.Must(uuid.NewV7()),
		UserID:    uuid.Must(uuid.NewV7()),
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}
	require.NoError(t, store.SaveRefreshToken(ctx, "", rt))
	require.NoError(t, store.ConsumeRefreshToken(ctx, "", rt.Hash))

	backdated := time.Now().UTC().Add(-time.Hour)
	_, err := adapter.DB().ExecContext(ctx, `UPDATE auth_refresh_tokens SET consumed_at = ? WHERE hash = ?`, backdated, rt.Hash)
	require.NoError(t, err)

	err = store.ConsumeRefreshToken(ctx, "", rt.Hash)
	require.Error(t, err)
	assert.ErrorIs(t, err, tokens.ErrRefreshTokenReused)
	assert.False(t, errors.Is(err, tokens.ErrRefreshConcurrent), "backdated replay must not be reported as benign concurrency")
}

func TestBasicIssuerConcurrentRotateKeepsFamily(t *testing.T) {
	adapter := setupTestDB(t)
	store := sqlite.NewTokensStore(adapter.DB(), 30*time.Second)
	ctx := context.Background()

	const tenant = "tenant-rotate"
	userID := uuid.Must(uuid.NewV7())
	issuer := basic.NewIssuer(basic.Config{
		Store:      store,
		Issuer:     "walldash",
		SecretKey:  "0123456789abcdef0123456789abcdef",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
		ClaimsProvider: basic.ClaimsProviderFunc(func(ctx context.Context, uid uuid.UUID, tid string) (basic.Claims, error) {
			return basic.Claims{Subject: uid, TenantID: tid}, nil
		}),
	})

	pair, err := issuer.IssueTokenPair(ctx, basic.Claims{Subject: userID, TenantID: tenant})
	require.NoError(t, err)

	type rotateResult struct {
		pair *tokens.TokenPair[struct{}]
		err  error
	}
	results := make(chan rotateResult, 2)
	start := make(chan struct{})
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			p, rerr := issuer.Rotate(ctx, tenant, pair.RefreshToken)
			results <- rotateResult{pair: p, err: rerr}
		}()
	}
	close(start)

	var winner *tokens.TokenPair[struct{}]
	losers := 0
	for i := 0; i < 2; i++ {
		r := <-results
		if r.err == nil {
			winner = r.pair
			continue
		}
		losers++
		assert.ErrorIs(t, r.err, tokens.ErrRefreshConcurrent)
	}
	require.Equal(t, 1, losers, "exactly one rotate must lose the race")
	require.NotNil(t, winner)

	next, err := issuer.Rotate(ctx, tenant, winner.RefreshToken)
	require.NoError(t, err, "winner's refresh token must still rotate; family must not be revoked")
	require.NotEmpty(t, next.RefreshToken)
}

func TestSQLiteTokensStoreHonorsConfiguredGrace(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		adapter := setupTestDB(t)
		store := sqlite.NewTokensStore(adapter.DB(), 30*time.Second)
		ctx := t.Context()

		rt := &tokens.RefreshToken{
			Hash:      "configured-grace",
			FamilyID:  uuid.Must(uuid.NewV7()),
			UserID:    uuid.Must(uuid.NewV7()),
			ExpiresAt: time.Now().Add(time.Hour),
			CreatedAt: time.Now(),
		}
		require.NoError(t, store.SaveRefreshToken(ctx, "", rt))
		require.NoError(t, store.ConsumeRefreshToken(ctx, "", rt.Hash))

		// A device replaying its consumed token 15s later is still inside the configured
		// 30s grace: this is benign concurrency, not theft.
		time.Sleep(15 * time.Second)
		err := store.ConsumeRefreshToken(ctx, "", rt.Hash)
		require.Error(t, err)
		assert.ErrorIs(t, err, tokens.ErrRefreshConcurrent, "replay inside the configured grace must be benign")

		// Past the grace window the same replay is genuine reuse and must be reported as such.
		time.Sleep(16 * time.Second)
		err = store.ConsumeRefreshToken(ctx, "", rt.Hash)
		require.Error(t, err)
		assert.ErrorIs(t, err, tokens.ErrRefreshTokenReused, "replay past the configured grace must be treated as reuse")
		assert.False(t, errors.Is(err, tokens.ErrRefreshConcurrent))
	})
}

// TestBasicIssuerReuseGraceKeepsFamilySynctest guards the end-to-end contract a waking
// device depends on: replaying a just-consumed refresh token inside the configured grace
// window must not poison the token family. It uses synctest so the grace boundary is
// exercised deterministically on a fake clock instead of a real 30s wait.
func TestBasicIssuerReuseGraceKeepsFamilySynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		adapter := setupTestDB(t)
		store := sqlite.NewTokensStore(adapter.DB(), 30*time.Second)
		ctx := t.Context()

		const tenant = "tenant-grace"
		userID := uuid.Must(uuid.NewV7())
		issuer := basic.NewIssuer(basic.Config{
			Store:            store,
			Issuer:           "walldash",
			SecretKey:        "0123456789abcdef0123456789abcdef",
			AccessTTL:        time.Minute,
			RefreshTTL:       time.Hour,
			ReuseGracePeriod: 30 * time.Second,
			ClaimsProvider: basic.ClaimsProviderFunc(func(ctx context.Context, uid uuid.UUID, tid string) (basic.Claims, error) {
				return basic.Claims{Subject: uid, TenantID: tid}, nil
			}),
		})

		pair, err := issuer.IssueTokenPair(ctx, basic.Claims{Subject: userID, TenantID: tenant})
		require.NoError(t, err)

		refreshed, err := issuer.Rotate(ctx, tenant, pair.RefreshToken)
		require.NoError(t, err)

		// The device replays its old refresh token 15s later: inside the configured grace
		// this is the benign race the fix exists to absorb, not a logout trigger.
		time.Sleep(15 * time.Second)
		_, err = issuer.Rotate(ctx, tenant, pair.RefreshToken)
		require.Error(t, err)
		assert.ErrorIs(t, err, tokens.ErrRefreshConcurrent)

		// The family must survive: the rotation winner's token still rotates.
		next, err := issuer.Rotate(ctx, tenant, refreshed.RefreshToken)
		require.NoError(t, err, "family must not be revoked by a within-grace replay")
		require.NotEmpty(t, next.RefreshToken)
	})
}

func TestSQLiteTokensStorePersistsRememberMeAndAMR(t *testing.T) {
	adapter := setupTestDB(t)
	store := sqlite.NewTokensStore(adapter.DB(), 30*time.Second)
	ctx := context.Background()

	rt := &tokens.RefreshToken{
		Hash:       "remember-amr",
		FamilyID:   uuid.Must(uuid.NewV7()),
		UserID:     uuid.Must(uuid.NewV7()),
		RememberMe: true,
		AMR:        []string{"device", "mfa"},
		ExpiresAt:  time.Now().Add(time.Hour),
		CreatedAt:  time.Now(),
	}
	require.NoError(t, store.SaveRefreshToken(ctx, "", rt))

	found, err := store.FindRefreshToken(ctx, "", rt.Hash)
	require.NoError(t, err)
	assert.True(t, found.RememberMe, "remember_me must round-trip so a rotation keeps the cookie persistence choice")
	assert.Equal(t, []string{"device", "mfa"}, found.AMR, "amr must round-trip so a stepped-up session keeps its assurance across a silent rotation")
}

func TestSQLiteTokensStoreReadsLegacyRefreshTokenRow(t *testing.T) {
	adapter := setupTestDB(t)
	store := sqlite.NewTokensStore(adapter.DB(), 30*time.Second)
	ctx := context.Background()

	familyID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())
	now := time.Now().UTC().Truncate(time.Second)
	expiresAt := now.Add(time.Hour)

	// Simulate a row written by a pre-v0.16 binary: remember_me/amr are not mentioned, so they
	// take their migration defaults. The row must stay readable and rotatable after the upgrade.
	_, err := adapter.DB().ExecContext(ctx,
		`INSERT INTO auth_refresh_tokens (hash, family_id, user_id, tenant_id, auth_time, must_change_password, expires_at, created_at, consumed_at)
		 VALUES (?, ?, ?, '', ?, 0, ?, ?, NULL)`,
		"legacy-rt", familyID.String(), userID.String(), now, expiresAt, now)
	require.NoError(t, err)

	found, err := store.FindRefreshToken(ctx, "", "legacy-rt")
	require.NoError(t, err, "a refresh token issued before the migration must still be found")
	assert.False(t, found.RememberMe, "a legacy row reads back as a non-remember-me family")
	assert.Empty(t, found.AMR, "a legacy row reads back with no recorded assurance")

	// It must still rotate: consuming the legacy token succeeds like any live one.
	require.NoError(t, store.ConsumeRefreshToken(ctx, "", "legacy-rt"))
}
