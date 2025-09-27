package usecase

import (
	"context"
	"errors"

	"backend/internal/domain/model"
	"backend/internal/domain/repository"

	"github.com/google/uuid"
)

var (
	ErrInvalidEventType = errors.New("invalid event type")
)

type AnalyticsUsecase struct {
	analyticsRepo repository.AnalyticsRepository
	projectRepo   repository.ProjectRepository
}

func NewAnalyticsUsecase(analyticsRepo repository.AnalyticsRepository, projectRepo repository.ProjectRepository) *AnalyticsUsecase {
	return &AnalyticsUsecase{
		analyticsRepo: analyticsRepo,
		projectRepo:   projectRepo,
	}
}

func (uc *AnalyticsUsecase) LogEvent(ctx context.Context, userID, projectID uuid.UUID, eventType model.EventType) (*model.AnalyticsEvent, error) {
	project, err := uc.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	if project.UserID != userID {
		return nil, ErrProjectAccessDenied
	}

	switch eventType {
	case model.EventTypePlay, model.EventTypeClick, model.EventTypeImpression:
		// valid
	default:
		return nil, ErrInvalidEventType
	}

	event := &model.AnalyticsEvent{
		ID:        uuid.New(),
		ProjectID: projectID,
		UserID:    userID,
		EventType: eventType,
	}

	if err := uc.analyticsRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (uc *AnalyticsUsecase) ListAllEvents(ctx context.Context) ([]*model.AnalyticsEvent, error) {
	return uc.analyticsRepo.FindAll(ctx)
}
