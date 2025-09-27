package repository

import (
	"context"

	"backend/internal/domain/model"

	"github.com/google/uuid"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *model.Project) error
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Project, error)
	FindByID(ctx context.Context, projectID uuid.UUID) (*model.Project, error)
}
