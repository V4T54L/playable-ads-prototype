package model

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventTypePlay       EventType = "play"
	EventTypeClick      EventType = "click"
	EventTypeImpression EventType = "impression"
)

type AnalyticsEvent struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	EventType EventType `json:"event_type"`
	CreatedAt time.Time `json:"created_at"`
}
