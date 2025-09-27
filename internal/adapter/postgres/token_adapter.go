package postgres

import (
	"context"
	"errors"

	"backend/internal/domain/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTokenNotFound = errors.New("token not found")

type TokenPostgresRepository struct {
	db *pgxpool.Pool
}

func NewTokenPostgresRepository(db *pgxpool.Pool) *TokenPostgresRepository {
	return &TokenPostgresRepository{db: db}
}

func (r *TokenPostgresRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	// Upsert logic to enforce single active token per user
	query := `
        INSERT INTO refresh_tokens (token, user_id, expires_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) DO UPDATE SET
            token = EXCLUDED.token,
            expires_at = EXCLUDED.expires_at;
    `
	_, err := r.db.Exec(ctx, query, token.Token, token.UserID, token.ExpiresAt)
	return err
}

func (r *TokenPostgresRepository) Find(ctx context.Context, token uuid.UUID) (*model.RefreshToken, error) {
	query := `SELECT token, user_id, expires_at FROM refresh_tokens WHERE token = $1`
	row := r.db.QueryRow(ctx, query, token)

	var rt model.RefreshToken
	err := row.Scan(&rt.Token, &rt.UserID, &rt.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}
	return &rt, nil
}

func (r *TokenPostgresRepository) Delete(ctx context.Context, token uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	cmdTag, err := r.db.Exec(ctx, query, token)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrTokenNotFound
	}
	return nil
}
