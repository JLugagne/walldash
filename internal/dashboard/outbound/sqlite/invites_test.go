package sqlite_test

import (
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/invites/invitestest"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/sqlite"
)

func TestSQLiteInviteRepositoryContract(t *testing.T) {
	adapter := setupTestDB(t)
	invitestest.InviteRepositoryContractTesting(t, sqlite.NewInviteRepository(adapter.DB()))
}
