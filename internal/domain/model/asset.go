package model

import (
	"time"

	"github.com/google/uuid"
)

type Asset struct {
	ID               uuid.UUID `json:"id"`
	ProjectID        uuid.UUID `json:"project_id"`
	OriginalFilename string    `json:"original_filename"`
	StoredFilename   string    `json:"stored_filename"`
	MIMEType         string    `json:"mime_type"`
	Size             int64     `json:"size"`
	CreatedAt        time.Time `json:"created_at"`
}
