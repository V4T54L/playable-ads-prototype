package postgres

import (
	"context"
	"errors"

	"backend/internal/domain/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProjectNotFound = errors.New("project not found")

type ProjectPostgresRepository struct {
	db *pgxpool.Pool
}

func NewProjectPostgresRepository(db *pgxpool.Pool) *ProjectPostgresRepository {
	return &ProjectPostgresRepository{db: db}
}

func (r *ProjectPostgresRepository) Create(ctx context.Context, project *model.Project) error {
	query := `INSERT INTO projects (id, user_id, title, description) VALUES ($1, $2, $3, $4) RETURNING created_at, updated_at`
	err := r.db.QueryRow(ctx, query, project.ID, project.UserID, project.Title, project.Description).Scan(&project.CreatedAt, &project.UpdatedAt)
	return err
}

func (r *ProjectPostgresRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Project, error) {
	query := `SELECT id, user_id, title, description, created_at, updated_at FROM projects WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []*model.Project{}
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}

	return projects, nil
}

func (r *ProjectPostgresRepository) FindByID(ctx context.Context, projectID uuid.UUID) (*model.Project, error) {
	query := `SELECT id, user_id, title, description, created_at, updated_at FROM projects WHERE id = $1`
	row := r.db.QueryRow(ctx, query, projectID)

	var p model.Project
	err := row.Scan(&p.ID, &p.UserID, &p.Title, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return &p, nil
}
