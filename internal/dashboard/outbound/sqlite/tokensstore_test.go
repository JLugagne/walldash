package sqlite_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JLugagne/egauth/tokens"
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
