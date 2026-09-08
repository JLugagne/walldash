package restoretest

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/restore"
)

// MockRestoreCommands is a function-based mock implementation of restore.RestoreCommands.
type MockRestoreCommands struct {
	RestoreBackupFunc func(ctx context.Context, actor domain.Actor, input restore.RestoreInput) (restore.RestoreSummary, error)
}

func (m *MockRestoreCommands) RestoreBackup(ctx context.Context, actor domain.Actor, input restore.RestoreInput) (restore.RestoreSummary, error) {
	if m.RestoreBackupFunc == nil {
		panic("called not defined RestoreBackupFunc")
	}
	return m.RestoreBackupFunc(ctx, actor, input)
}
