package sqlite_test

import (
	"context"
	"errors"
	"sync"
	"testing"
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
	store := sqlite.NewTokensStore(adapter.DB())
	storetest.StoreContractTesting(t, store, false, struct{}{})
}

func TestSQLiteTokensStoreConcurrentConsume(t *testing.T) {
	adapter := setupTestDB(t)
	store := sqlite.NewTokensStore(adapter.DB())
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
	store := sqlite.NewTokensStore(adapter.DB())
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
	store := sqlite.NewTokensStore(adapter.DB())
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
	store := sqlite.NewTokensStore(adapter.DB())
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
