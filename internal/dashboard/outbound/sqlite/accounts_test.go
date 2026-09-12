package sqlite_test

import (
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/accounts/accountstest"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/sqlite"
)

func TestSQLiteAccountRepositoryContract(t *testing.T) {
	adapter := setupTestDB(t)
	accountstest.AccountRepositoryContractTesting(t, sqlite.NewAccountRepository(adapter.DB()))
}
