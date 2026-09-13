package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/invites"
	"github.com/JLugagne/walldash/internal/pkg/logger"
)

// inviteRepo is the SQLite implementation of invites.Repository.
type inviteRepo struct {
	db *sql.DB
}

var _ invites.Repository = (*inviteRepo)(nil)

// NewInviteRepository builds a SQLite-backed invitation repository.
func NewInviteRepository(db *sql.DB) invites.Repository {
	return &inviteRepo{db: db}
}

type inviteRowScanner interface {
	Scan(dest ...any) error
}

const inviteColumns = `selector, verifier_hash, role, created_by, created_at, expires_at, consumed_at, consumed_by, revoked_at`

func scanInvite(row inviteRowScanner) (domain.Invite, error) {
	var invite domain.Invite
	var role string
	var consumedAt, revokedAt sql.NullTime
	var consumedBy sql.NullString
	if err := row.Scan(&invite.Selector, &invite.VerifierHash, &role, &invite.CreatedBy, &invite.CreatedAt, &invite.ExpiresAt, &consumedAt, &consumedBy, &revokedAt); err != nil {
		return domain.Invite{}, err
	}
	invite.Role = domain.Role(role)
	if consumedAt.Valid {
		value := consumedAt.Time
		invite.ConsumedAt = &value
	}
	if consumedBy.Valid {
		invite.ConsumedBy = consumedBy.String
	}
	if revokedAt.Valid {
		value := revokedAt.Time
		invite.RevokedAt = &value
	}
	return invite, nil
}

func (r *inviteRepo) Create(ctx context.Context, invite domain.Invite) (domain.Invite, error) {
	log := logger.LoggerFromContext(ctx)
	if invite.CreatedAt.IsZero() {
		invite.CreatedAt = time.Now().UTC().Truncate(time.Second)
	}
	query := `INSERT INTO auth_invites (selector, verifier_hash, role, created_by, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING ` + inviteColumns
	row := r.db.QueryRowContext(ctx, query, invite.Selector, invite.VerifierHash, string(invite.Role), invite.CreatedBy, invite.CreatedAt.UTC(), invite.ExpiresAt.UTC())
	created, err := scanInvite(row)
	if err != nil {
		log.WithError(err).WithField("selector", invite.Selector).Error("failed to create invite")
		return domain.Invite{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return created, nil
}

func (r *inviteRepo) FindBySelector(ctx context.Context, selector string) (domain.Invite, error) {
	log := logger.LoggerFromContext(ctx)
	query := `SELECT ` + inviteColumns + ` FROM auth_invites WHERE selector = ?`
	invite, err := scanInvite(r.db.QueryRowContext(ctx, query, selector))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Invite{}, domain.ErrInviteNotFound
		}
		log.WithError(err).WithField("selector", selector).Error("failed to query invite")
		return domain.Invite{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return invite, nil
}

func (r *inviteRepo) FindAll(ctx context.Context) ([]domain.Invite, error) {
	log := logger.LoggerFromContext(ctx)
	query := `SELECT ` + inviteColumns + ` FROM auth_invites ORDER BY created_at DESC, selector ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.WithError(err).Error("failed to query invites")
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	defer rows.Close()
	all := []domain.Invite{}
	for rows.Next() {
		invite, err := scanInvite(rows)
		if err != nil {
			log.WithError(err).Error("failed to scan invite")
			return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
		}
		all = append(all, invite)
	}
	if err := rows.Err(); err != nil {
		log.WithError(err).Error("failed to iterate invites")
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return all, nil
}

func (r *inviteRepo) Consume(ctx context.Context, selector string, consumedBy string) (domain.Invite, error) {
	log := logger.LoggerFromContext(ctx)
	// The conditional UPDATE is the single-use gate: it only flips a row that is still
	// active, so two concurrent redemptions cannot both succeed.
	query := `UPDATE auth_invites
		SET consumed_at = CURRENT_TIMESTAMP, consumed_by = ?
		WHERE selector = ? AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP
		RETURNING ` + inviteColumns
	invite, err := scanInvite(r.db.QueryRowContext(ctx, query, consumedBy, selector))
	if err == nil {
		return invite, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		log.WithError(err).WithField("selector", selector).Error("failed to consume invite")
		return domain.Invite{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	// No row was updated: load the current state to report the precise reason.
	current, findErr := r.FindBySelector(ctx, selector)
	if findErr != nil {
		return domain.Invite{}, findErr
	}
	switch {
	case current.ConsumedAt != nil:
		return domain.Invite{}, domain.ErrInviteConsumed
	case current.RevokedAt != nil:
		return domain.Invite{}, domain.ErrInviteRevoked
	case !current.ExpiresAt.After(time.Now().UTC()):
		return domain.Invite{}, domain.ErrInviteExpired
	default:
		return domain.Invite{}, domain.ErrInviteConsumed
	}
}

func (r *inviteRepo) Revoke(ctx context.Context, selector string) (domain.Invite, error) {
	log := logger.LoggerFromContext(ctx)
	query := `UPDATE auth_invites SET revoked_at = CURRENT_TIMESTAMP
		WHERE selector = ? AND revoked_at IS NULL
		RETURNING ` + inviteColumns
	invite, err := scanInvite(r.db.QueryRowContext(ctx, query, selector))
	if err == nil {
		return invite, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		log.WithError(err).WithField("selector", selector).Error("failed to revoke invite")
		return domain.Invite{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	current, findErr := r.FindBySelector(ctx, selector)
	if findErr != nil {
		return domain.Invite{}, findErr
	}
	return current, nil
}

func (r *inviteRepo) DeleteExpired(ctx context.Context) (int64, error) {
	log := logger.LoggerFromContext(ctx)
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_invites WHERE expires_at <= CURRENT_TIMESTAMP`)
	if err != nil {
		log.WithError(err).Error("failed to delete expired invites")
		return 0, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return res.RowsAffected()
}
