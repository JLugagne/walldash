package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/accounts"
	"github.com/JLugagne/walldash/internal/pkg/logger"
)

type accountRepo struct {
	db *sql.DB
}

var _ accounts.AccountRepository = (*accountRepo)(nil)

func NewAccountRepository(db *sql.DB) accounts.AccountRepository {
	return &accountRepo{db: db}
}

type accountRowScanner interface {
	Scan(dest ...any) error
}

func scanAccount(row accountRowScanner) (domain.Account, error) {
	var account domain.Account
	var status, role string
	var lastSeen, revokedAt sql.NullTime
	if err := row.Scan(&account.ID, &status, &role, &account.Label, &account.CreatedAt, &lastSeen, &revokedAt); err != nil {
		return domain.Account{}, err
	}
	account.Status = domain.Status(status)
	account.Role = domain.Role(role)
	if lastSeen.Valid {
		value := lastSeen.Time
		account.LastSeen = &value
	}
	if revokedAt.Valid {
		value := revokedAt.Time
		account.RevokedAt = &value
	}
	return account, nil
}

func (r *accountRepo) Create(ctx context.Context, account domain.Account) (domain.Account, error) {
	log := logger.LoggerFromContext(ctx)
	if account.Role == "" {
		account.Role = domain.RoleDevice
	}
	if account.Status == "" {
		account.Status = domain.StatusActive
	}
	if account.CreatedAt.IsZero() {
		account.CreatedAt = time.Now().UTC().Truncate(time.Second)
	}
	account.CreatedAt = account.CreatedAt.UTC()

	query := `INSERT INTO auth_accounts (id, status, role, label, created_at)
		VALUES (?, ?, CASE WHEN NOT EXISTS (SELECT 1 FROM auth_accounts) THEN ? ELSE ? END, ?, ?)
		RETURNING id, status, role, label, created_at, last_seen_at, revoked_at`
	row := r.db.QueryRowContext(ctx, query, account.ID, account.Status, domain.RoleOwner, account.Role, account.Label, account.CreatedAt)
	created, err := scanAccount(row)
	if err != nil {
		log.WithError(err).WithField("account_id", account.ID).Error("failed to create account")
		return domain.Account{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return created, nil
}

func (r *accountRepo) FindByID(ctx context.Context, id string) (domain.Account, error) {
	log := logger.LoggerFromContext(ctx)
	query := `SELECT id, status, role, label, created_at, last_seen_at, revoked_at FROM auth_accounts WHERE id = ?`
	account, err := scanAccount(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Account{}, errors.Join(domain.ErrAccountNotFound, err)
		}
		log.WithError(err).WithField("account_id", id).Error("failed to query account by id")
		return domain.Account{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return account, nil
}

func (r *accountRepo) FindAll(ctx context.Context) ([]domain.Account, error) {
	log := logger.LoggerFromContext(ctx)
	query := `SELECT id, status, role, label, created_at, last_seen_at, revoked_at FROM auth_accounts ORDER BY created_at ASC, id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.WithError(err).Error("failed to query accounts")
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	defer rows.Close()

	all := []domain.Account{}
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			log.WithError(err).Error("failed to scan account")
			return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
		}
		all = append(all, account)
	}
	if err := rows.Err(); err != nil {
		log.WithError(err).Error("failed to iterate accounts")
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return all, nil
}

func (r *accountRepo) SetRole(ctx context.Context, id string, role domain.Role) (domain.Account, error) {
	log := logger.LoggerFromContext(ctx)
	if !role.Valid() {
		return domain.Account{}, domain.ErrInvalidRole
	}

	res, err := r.db.ExecContext(ctx, `UPDATE auth_accounts SET role = ? WHERE id = ?`, string(role), id)
	if err != nil {
		log.WithError(err).WithField("account_id", id).Error("failed to update account role")
		return domain.Account{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return domain.Account{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if affected == 0 {
		return domain.Account{}, errors.Join(domain.ErrAccountNotFound, errors.New("no account found to update role"))
	}
	return r.FindByID(ctx, id)
}

func (r *accountRepo) Revoke(ctx context.Context, id string) (domain.Account, error) {
	log := logger.LoggerFromContext(ctx)
	now := time.Now().UTC().Truncate(time.Second)
	res, err := r.db.ExecContext(ctx, `UPDATE auth_accounts SET status = ?, revoked_at = COALESCE(revoked_at, ?) WHERE id = ?`, string(domain.StatusRevoked), now, id)
	if err != nil {
		log.WithError(err).WithField("account_id", id).Error("failed to revoke account")
		return domain.Account{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return domain.Account{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if affected == 0 {
		return domain.Account{}, errors.Join(domain.ErrAccountNotFound, errors.New("no account found to revoke"))
	}
	return r.FindByID(ctx, id)
}

func (r *accountRepo) Touch(ctx context.Context, id string) error {
	log := logger.LoggerFromContext(ctx)
	now := time.Now().UTC().Truncate(time.Second)
	res, err := r.db.ExecContext(ctx, `UPDATE auth_accounts SET last_seen_at = ? WHERE id = ?`, now, id)
	if err != nil {
		log.WithError(err).WithField("account_id", id).Error("failed to touch account")
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if affected == 0 {
		return errors.Join(domain.ErrAccountNotFound, errors.New("no account found to touch"))
	}
	return nil
}
