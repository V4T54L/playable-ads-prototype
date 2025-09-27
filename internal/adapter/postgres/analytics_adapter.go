package postgres

import (
	"context"
	"errors"

	"backend/internal/domain/model"
	"backend/internal/domain/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAnalyticsEventNotFound = errors.New("analytics event not found")

type AnalyticsPostgresRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsPostgresRepository(db *pgxpool.Pool) repository.AnalyticsRepository {
	return &AnalyticsPostgresRepository{db: db}
}

func (r *AnalyticsPostgresRepository) Create(ctx context.Context, event *model.AnalyticsEvent) error {
	query := `INSERT INTO analytics_events (id, project_id, user_id, event_type)
              VALUES ($1, $2, $3, $4)
              RETURNING created_at`
	return r.db.QueryRow(ctx, query, event.ID, event.ProjectID, event.UserID, event.EventType).Scan(&event.CreatedAt)
}

func (r *AnalyticsPostgresRepository) FindAll(ctx context.Context) ([]*model.AnalyticsEvent, error) {
	query := `SELECT id, project_id, user_id, event_type, created_at
              FROM analytics_events
              ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*model.AnalyticsEvent
	for rows.Next() {
		var event model.AnalyticsEvent
		if err := rows.Scan(&event.ID, &event.ProjectID, &event.UserID, &event.EventType, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, nil
}
