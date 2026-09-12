package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JLugagne/egauth/tokens"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/google/uuid"
)

type TokensStore struct {
	db *sql.DB
}

var _ tokens.Store[struct{}] = (*TokensStore)(nil)

func NewTokensStore(db *sql.DB) *TokensStore {
	return &TokensStore{db: db}
}

func scanRefreshToken(row accountRowScanner) (*tokens.RefreshToken, error) {
	var rt tokens.RefreshToken
	var familyID, userID string
	var consumedAt sql.NullTime
	if err := row.Scan(&rt.Hash, &familyID, &userID, &rt.TenantID, &rt.AuthTime, &rt.MustChangePassword, &rt.ExpiresAt, &rt.CreatedAt, &consumedAt); err != nil {
		return nil, err
	}
	fid, err := uuid.Parse(familyID)
	if err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	rt.FamilyID = fid
	rt.UserID = uid
	if consumedAt.Valid {
		value := consumedAt.Time
		rt.ConsumedAt = &value
	}
	return &rt, nil
}

func scanAPIKey(row accountRowScanner) (*tokens.APIKey[struct{}], error) {
	var key tokens.APIKey[struct{}]
	var id, hash, prefix, keyType, createdBy string
	var claims []byte
	var expiresAt, revokedAt sql.NullTime
	if err := row.Scan(&id, &key.TenantID, &hash, &prefix, &keyType, &createdBy, &claims, &expiresAt, &revokedAt); err != nil {
		return nil, err
	}
	keyID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	creator, err := uuid.Parse(createdBy)
	if err != nil {
		return nil, err
	}
	key.ID = keyID
	key.Hash = hash
	key.Prefix = prefix
	key.Type = tokens.KeyType(keyType)
	key.CreatedBy = creator
	if err := json.Unmarshal(claims, &key.Claims); err != nil {
		return nil, err
	}
	if expiresAt.Valid {
		value := expiresAt.Time
		key.ExpiresAt = &value
	}
	if revokedAt.Valid {
		value := revokedAt.Time
		key.RevokedAt = &value
	}
	return &key, nil
}

func (s *TokensStore) SaveRefreshToken(ctx context.Context, tenantID string, rt *tokens.RefreshToken) error {
	if rt.TenantID != tenantID {
		return fmt.Errorf("tokensstore: %w", tokens.ErrTenantMismatch)
	}
	var consumedAt any
	if rt.ConsumedAt != nil {
		consumedAt = rt.ConsumedAt.UTC()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO auth_refresh_tokens (hash, family_id, user_id, tenant_id, auth_time, must_change_password, expires_at, created_at, consumed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rt.Hash, rt.FamilyID.String(), rt.UserID.String(), rt.TenantID, rt.AuthTime.UTC(), rt.MustChangePassword, rt.ExpiresAt.UTC(), rt.CreatedAt.UTC(), consumedAt)
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return nil
}

func (s *TokensStore) FindRefreshToken(ctx context.Context, tenantID string, tokenHash string) (*tokens.RefreshToken, error) {
	query := `SELECT hash, family_id, user_id, tenant_id, auth_time, must_change_password, expires_at, created_at, consumed_at FROM auth_refresh_tokens WHERE hash = ? AND tenant_id = ?`
	rt, err := scanRefreshToken(s.db.QueryRowContext(ctx, query, tokenHash, tenantID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(tokens.ErrRefreshTokenNotFound, err)
		}
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return rt, nil
}

func (s *TokensStore) ConsumeRefreshToken(ctx context.Context, tenantID string, tokenHash string) error {
	now := time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx, `UPDATE auth_refresh_tokens SET consumed_at = ? WHERE hash = ? AND tenant_id = ? AND consumed_at IS NULL`, now, tokenHash, tenantID)
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if affected == 1 {
		return nil
	}

	var consumedAt sql.NullTime
	err = s.db.QueryRowContext(ctx, `SELECT consumed_at FROM auth_refresh_tokens WHERE hash = ? AND tenant_id = ?`, tokenHash, tenantID).Scan(&consumedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.Join(tokens.ErrRefreshTokenNotFound, err)
		}
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return tokens.ErrRefreshTokenReused
}

func (s *TokensStore) RevokeRefreshToken(ctx context.Context, tenantID string, tokenHash string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM auth_refresh_tokens WHERE hash = ? AND tenant_id = ?`, tokenHash, tenantID)
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if affected == 0 {
		return errors.Join(tokens.ErrRefreshTokenNotFound, sql.ErrNoRows)
	}
	return nil
}

func (s *TokensStore) RevokeFamily(ctx context.Context, tenantID string, familyID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_refresh_tokens WHERE family_id = ? AND tenant_id = ?`, familyID.String(), tenantID)
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return nil
}

func (s *TokensStore) RevokeAllRefreshTokensForUser(ctx context.Context, tenantID string, userID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_refresh_tokens WHERE user_id = ? AND tenant_id = ?`, userID.String(), tenantID)
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return nil
}

func (s *TokensStore) SaveAPIKey(ctx context.Context, tenantID string, key *tokens.APIKey[struct{}]) error {
	if key.TenantID != tenantID {
		return fmt.Errorf("tokensstore: %w", tokens.ErrTenantMismatch)
	}
	claims, err := json.Marshal(key.Claims)
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	var expiresAt, revokedAt any
	if key.ExpiresAt != nil {
		expiresAt = key.ExpiresAt.UTC()
	}
	if key.RevokedAt != nil {
		revokedAt = key.RevokedAt.UTC()
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO auth_api_keys (id, tenant_id, hash, prefix, type, created_by, claims, expires_at, revoked_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		key.ID.String(), key.TenantID, key.Hash, key.Prefix, string(key.Type), key.CreatedBy.String(), claims, expiresAt, revokedAt)
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return nil
}

func (s *TokensStore) FindAPIKeyByHash(ctx context.Context, tenantID string, tokenHash string) (*tokens.APIKey[struct{}], error) {
	query := `SELECT id, tenant_id, hash, prefix, type, created_by, claims, expires_at, revoked_at FROM auth_api_keys WHERE hash = ? AND tenant_id = ?`
	key, err := scanAPIKey(s.db.QueryRowContext(ctx, query, tokenHash, tenantID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(tokens.ErrAPIKeyNotFound, err)
		}
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return key, nil
}

func (s *TokensStore) RevokeAPIKey(ctx context.Context, tenantID string, keyID uuid.UUID) error {
	now := time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx, `UPDATE auth_api_keys SET revoked_at = COALESCE(revoked_at, ?) WHERE id = ? AND tenant_id = ?`, now, keyID.String(), tenantID)
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if affected == 0 {
		return errors.Join(tokens.ErrAPIKeyNotFound, sql.ErrNoRows)
	}
	return nil
}

func (s *TokensStore) ListAPIKeysByCreator(ctx context.Context, tenantID string, createdBy uuid.UUID) ([]*tokens.APIKey[struct{}], error) {
	query := `SELECT id, tenant_id, hash, prefix, type, created_by, claims, expires_at, revoked_at FROM auth_api_keys WHERE created_by = ? AND tenant_id = ? ORDER BY created_at ASC, id ASC`
	rows, err := s.db.QueryContext(ctx, query, createdBy.String(), tenantID)
	if err != nil {
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	defer rows.Close()

	keys := []*tokens.APIKey[struct{}]{}
	for rows.Next() {
		key, err := scanAPIKey(rows)
		if err != nil {
			return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return keys, nil
}

func (s *TokensStore) RevokeAllAPIKeysForUser(ctx context.Context, tenantID string, userID uuid.UUID) error {
	now := time.Now().UTC().Truncate(time.Second)
	_, err := s.db.ExecContext(ctx, `UPDATE auth_api_keys SET revoked_at = COALESCE(revoked_at, ?) WHERE created_by = ? AND tenant_id = ?`, now, userID.String(), tenantID)
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return nil
}

func (s *TokensStore) DeleteExpired(ctx context.Context, tenantID string) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM auth_refresh_tokens WHERE tenant_id = ? AND expires_at < ?`, tenantID, time.Now().UTC())
	if err != nil {
		return 0, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return affected, nil
}
