package repository

import (
	"context"

	"backend/internal/domain/model"

	"github.com/google/uuid"
)

type TokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	Find(ctx context.Context, token uuid.UUID) (*model.RefreshToken, error)
	Delete(ctx context.Context, token uuid.UUID) error
}
