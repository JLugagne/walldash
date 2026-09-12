package sqlite

import (
	"context"
	"database/sql"
	"errors"
)

const TokenSecretName = "token_secret"

type SecretsStore struct {
	db *sql.DB
}

func NewSecretsStore(db *sql.DB) *SecretsStore {
	return &SecretsStore{db: db}
}

func (s *SecretsStore) Get(ctx context.Context, name string) ([]byte, error) {
	var value []byte
	err := s.db.QueryRowContext(ctx, "SELECT value FROM app_secrets WHERE name = ?", name).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *SecretsStore) SetIfAbsent(ctx context.Context, name string, value []byte) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO app_secrets (name, value) VALUES (?, ?) ON CONFLICT(name) DO NOTHING", name, value)
	return err
}

func (s *SecretsStore) Set(ctx context.Context, name string, value []byte) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO app_secrets (name, value) VALUES (?, ?) ON CONFLICT(name) DO UPDATE SET value = excluded.value", name, value)
	return err
}
