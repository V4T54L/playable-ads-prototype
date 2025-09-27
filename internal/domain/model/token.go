package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	Token     uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
}
