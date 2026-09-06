package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/health"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/levels"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/overviews"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/placements"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/plans"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow"
	"github.com/JLugagne/walldash/internal/pkg/logger"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Adapter provides the SQLite storage adapter implementing domain repositories and Unit of Work.
type Adapter struct {
	db *sql.DB
}

// Ensure Adapter implements domain repositories and uow.UnitOfWork.
var (
	_ health.Repository                    = (*Adapter)(nil)
	_ levels.LevelRepository               = (*Adapter)(nil)
	_ plans.PlanRepository                 = (*Adapter)(nil)
	_ placements.DevicePlacementRepository = (*Adapter)(nil)
	_ overviews.OverviewRepository         = (*Adapter)(nil)
	_ overviews.WidgetRepository           = (*Adapter)(nil)
	_ uow.UnitOfWork                       = (*Adapter)(nil)
)

// New initializes an Adapter connected to the SQLite database at dbPath and runs migrations.
func New(ctx context.Context, dbPath string) (*Adapter, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	// Basic SQLite connection pooling tuning for single-file concurrency
	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	adapter := &Adapter{db: db}
	if err := adapter.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return adapter, nil
}

// Close closes the underlying SQLite database connection.
func (a *Adapter) Close() error {
	return a.db.Close()
}

// DB returns the underlying sql.DB instance.
func (a *Adapter) DB() *sql.DB {
	return a.db
}

// Ping verifies connectivity to the SQLite database.
func (a *Adapter) Ping(ctx context.Context) error {
	if err := a.db.PingContext(ctx); err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	var dummy int
	if err := a.db.QueryRowContext(ctx, "SELECT 1").Scan(&dummy); err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	return nil
}

// Do executes operations within a database transaction.
func (a *Adapter) Do(ctx context.Context, fn func(repos uow.Repositories) error) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	repos := uow.Repositories{
		Health:     &txHealthRepo{tx: tx},
		Levels:     &levelRepo{db: tx},
		Plans:      &planRepo{db: tx},
		Placements: &placementRepo{db: tx},
		Overviews:  &overviewRepo{db: tx},
		Widgets:    &widgetRepo{db: tx},
	}

	if err := fn(repos); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	return nil
}

// txHealthRepo wraps a transaction for health checks within a unit of work.
type txHealthRepo struct {
	tx *sql.Tx
}

func (r *txHealthRepo) Ping(ctx context.Context) error {
	var dummy int
	if err := r.tx.QueryRowContext(ctx, "SELECT 1").Scan(&dummy); err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return nil
}

func (a *Adapter) migrate(ctx context.Context) error {
	log := logger.LoggerFromContext(ctx)

	_, err := a.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL
		);
	`)
	if err != nil {
		log.WithError(err).Error("failed to create schema_migrations table")
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		log.WithError(err).Error("failed to read embedded migrations")
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		var count int
		err := a.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = ?", name).Scan(&count)
		if err != nil {
			return errors.Join(domain.ErrDatabaseUnavailable, err)
		}

		if count > 0 {
			continue
		}

		content, err := fs.ReadFile(migrationsFS, "migrations/"+name)
		if err != nil {
			return errors.Join(domain.ErrDatabaseUnavailable, err)
		}

		tx, err := a.db.BeginTx(ctx, nil)
		if err != nil {
			return errors.Join(domain.ErrDatabaseUnavailable, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			log.WithError(err).WithField("migration", name).Error("failed to apply migration")
			return errors.Join(domain.ErrDatabaseUnavailable, fmt.Errorf("migration %s failed: %w", name, err))
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version, applied_at) VALUES (?, CURRENT_TIMESTAMP)", name); err != nil {
			_ = tx.Rollback()
			return errors.Join(domain.ErrDatabaseUnavailable, err)
		}

		if err := tx.Commit(); err != nil {
			return errors.Join(domain.ErrDatabaseUnavailable, err)
		}

		log.WithField("migration", name).Info("applied migration successfully")
	}

	return nil
}
