package repository

import (
	"context"

	"backend/internal/domain/model"
)

type AnalyticsRepository interface {
	Create(ctx context.Context, event *model.AnalyticsEvent) error
	FindAll(ctx context.Context) ([]*model.AnalyticsEvent, error)
}
