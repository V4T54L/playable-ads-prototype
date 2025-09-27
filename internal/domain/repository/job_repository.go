package repository

import (
	"context"

	"backend/internal/domain/model"

	"github.com/google/uuid"
)

type JobRepository interface {
	Create(ctx context.Context, job *model.Job) error
	FindByID(ctx context.Context, jobID uuid.UUID) (*model.Job, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Job, error)
	FindDoneByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Job, error)
	UpdateToProcessing(ctx context.Context, jobID uuid.UUID) error
	UpdateToComplete(ctx context.Context, jobID uuid.UUID, outputPath string) error
	UpdateToFailed(ctx context.Context, jobID uuid.UUID, errorMessage string) error
}
