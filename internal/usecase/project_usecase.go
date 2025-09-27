package usecase

import (
	"context"
	"errors"

	"backend/internal/adapter/postgres"
	"backend/internal/domain/model"
	"backend/internal/domain/repository"

	"github.com/google/uuid"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrProjectAccessDenied = errors.New("access to project denied")
)

type ProjectUsecase struct {
	projectRepo repository.ProjectRepository
}

func NewProjectUsecase(projectRepo repository.ProjectRepository) *ProjectUsecase {
	return &ProjectUsecase{projectRepo: projectRepo}
}

func (uc *ProjectUsecase) CreateProject(ctx context.Context, userID uuid.UUID, title, description string) (*model.Project, error) {
	project := &model.Project{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       title,
		Description: description,
	}

	err := uc.projectRepo.Create(ctx, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (uc *ProjectUsecase) ListUserProjects(ctx context.Context, userID uuid.UUID) ([]*model.Project, error) {
	return uc.projectRepo.FindAllByUserID(ctx, userID)
}

func (uc *ProjectUsecase) GetProjectByID(ctx context.Context, userID, projectID uuid.UUID) (*model.Project, error) {
	project, err := uc.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, postgres.ErrProjectNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	if project.UserID != userID {
		return nil, ErrProjectAccessDenied
	}

	return project, nil
}
