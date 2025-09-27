package repository

import (
	"context"

	"backend/internal/domain/model"

	"github.com/google/uuid"
)

type AssetRepository interface {
	Create(ctx context.Context, asset *model.Asset) error
	FindAllByProjectID(ctx context.Context, projectID uuid.UUID) ([]*model.Asset, error)
	FindByID(ctx context.Context, assetID uuid.UUID) (*model.Asset, error)
}
