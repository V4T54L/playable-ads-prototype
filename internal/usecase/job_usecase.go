package usecase

import (
	"context"
	"errors"

	"backend/internal/domain/model"
	"backend/internal/domain/repository"

	"github.com/google/uuid"
)

var (
	ErrJobNotFound         = errors.New("job not found")
	ErrJobAccessDenied     = errors.New("access to job denied")
	ErrAssetNotEligible    = errors.New("asset is not eligible for rendering")
	ErrAssetNotFound       = errors.New("asset not found")
	ErrAssetAccessDenied   = errors.New("access to asset denied")
)

type JobQueue interface {
	Enqueue(ctx context.Context, jobID uuid.UUID) error
}

type JobUsecase struct {
	jobRepo     repository.JobRepository
	assetRepo   repository.AssetRepository
	projectRepo repository.ProjectRepository
	jobQueue    JobQueue
}

func NewJobUsecase(jobRepo repository.JobRepository, assetRepo repository.AssetRepository, projectRepo repository.ProjectRepository, jobQueue JobQueue) *JobUsecase {
	return &JobUsecase{
		jobRepo:     jobRepo,
		assetRepo:   assetRepo,
		projectRepo: projectRepo,
		jobQueue:    jobQueue,
	}
}

func (uc *JobUsecase) CreateRenderJob(ctx context.Context, userID, projectID, assetID uuid.UUID) (*model.Job, error) {
	// 1. Verify asset exists and belongs to the user's project
	asset, err := uc.assetRepo.FindByID(ctx, assetID)
	if err != nil {
		return nil, ErrAssetNotFound
	}

	if asset.ProjectID != projectID {
		return nil, ErrAssetAccessDenied
	}

	project, err := uc.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	if project.UserID != userID {
		return nil, ErrProjectAccessDenied
	}

	// 2. Create a new job model
	job := &model.Job{
		ID:      uuid.New(),
		AssetID: assetID,
		UserID:  userID,
		Status:  model.JobStatusPending,
	}

	// 3. Persist the job to the database
	if err := uc.jobRepo.Create(ctx, job); err != nil {
		return nil, err
	}

	// 4. Enqueue the job for the worker
	if err := uc.jobQueue.Enqueue(ctx, job.ID); err != nil {
		return nil, err
	}

	return job, nil
}

func (uc *JobUsecase) ListUserJobs(ctx context.Context, userID uuid.UUID) ([]*model.Job, error) {
	return uc.jobRepo.FindAllByUserID(ctx, userID)
}

func (uc *JobUsecase) ListUserOutputs(ctx context.Context, userID uuid.UUID) ([]*model.Job, error) {
	return uc.jobRepo.FindDoneByUserID(ctx, userID)
}

func (uc *JobUsecase) GetJobByID(ctx context.Context, userID, jobID uuid.UUID) (*model.Job, error) {
	job, err := uc.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, ErrJobNotFound
	}

	if job.UserID != userID {
		return nil, ErrJobAccessDenied
	}

	return job, nil
}
