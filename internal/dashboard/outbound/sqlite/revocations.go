package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/google/uuid"
)

// RevocationStore persists account-scoped access-token revocation cutoffs.
//
// The checker egauth installs is an in-process map, so without a durable copy an access token that
// was revoked, demoted or logged out becomes valid again after a restart for the remainder of its
// TTL. Cutoffs are stored as Unix nanoseconds; see migration 014 for why.
type RevocationStore struct {
	db *sql.DB
}

// NewRevocationStore wraps a database handle in a revocation store.
func NewRevocationStore(db *sql.DB) *RevocationStore {
	return &RevocationStore{db: db}
}

// RecordRevocation stores the cutoff for an account, keeping the most recent value so an older
// event can never rewind a newer revocation.
func (s *RevocationStore) RecordRevocation(ctx context.Context, userID uuid.UUID, reason string, cutoff time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO auth_revocation_cutoffs (user_id, cutoff_at, reason) VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET cutoff_at = excluded.cutoff_at, reason = excluded.reason
		WHERE excluded.cutoff_at > auth_revocation_cutoffs.cutoff_at`,
		userID.String(), cutoff.UTC().UnixNano(), reason)
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return nil
}

// RevocationCutoffs returns every stored cutoff keyed by account id.
func (s *RevocationStore) RevocationCutoffs(ctx context.Context) (map[uuid.UUID]time.Time, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id, cutoff_at FROM auth_revocation_cutoffs`)
	if err != nil {
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	defer rows.Close()

	cutoffs := make(map[uuid.UUID]time.Time)
	for rows.Next() {
		var rawID string
		var nanos int64
		if err := rows.Scan(&rawID, &nanos); err != nil {
			return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
		}
		userID, err := uuid.Parse(rawID)
		if err != nil {
			// An unparsable row can never match a token subject; skip it rather than refusing to
			// start, and let pruning drop it.
			continue
		}
		cutoffs[userID] = time.Unix(0, nanos).UTC()
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return cutoffs, nil
}

// PruneRevocations deletes cutoffs older than before. A token issued before such a cutoff has
// already outlived the access-token TTL, so the row can no longer reject anything.
func (s *RevocationStore) PruneRevocations(ctx context.Context, before time.Time) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_revocation_cutoffs WHERE cutoff_at < ?`, before.UTC().UnixNano())
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return nil
}
