package postgres

import (
	"context"
	"errors"

	"backend/internal/domain/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrJobNotFound = errors.New("job not found")

type JobPostgresRepository struct {
	db *pgxpool.Pool
}

func NewJobPostgresRepository(db *pgxpool.Pool) *JobPostgresRepository {
	return &JobPostgresRepository{db: db}
}

func (r *JobPostgresRepository) Create(ctx context.Context, job *model.Job) error {
	query := `INSERT INTO jobs (id, asset_id, user_id, status)
			   VALUES ($1, $2, $3, $4)
			   RETURNING created_at, updated_at`

	err := r.db.QueryRow(ctx, query, job.ID, job.AssetID, job.UserID, job.Status).Scan(&job.CreatedAt, &job.UpdatedAt)
	return err
}

func (r *JobPostgresRepository) FindByID(ctx context.Context, jobID uuid.UUID) (*model.Job, error) {
	query := `SELECT id, asset_id, user_id, status, error_message, output_path, created_at, updated_at, completed_at
			   FROM jobs WHERE id = $1`

	job := &model.Job{}
	err := r.db.QueryRow(ctx, query, jobID).Scan(
		&job.ID, &job.AssetID, &job.UserID, &job.Status,
		&job.ErrorMessage, &job.OutputPath, &job.CreatedAt, &job.UpdatedAt, &job.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, err
	}
	return job, nil
}

func (r *JobPostgresRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Job, error) {
	query := `SELECT id, asset_id, user_id, status, error_message, output_path, created_at, updated_at, completed_at
			   FROM jobs WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*model.Job
	for rows.Next() {
		job := &model.Job{}
		err := rows.Scan(
			&job.ID, &job.AssetID, &job.UserID, &job.Status,
			&job.ErrorMessage, &job.OutputPath, &job.CreatedAt, &job.UpdatedAt, &job.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (r *JobPostgresRepository) FindDoneByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Job, error) {
	query := `SELECT id, asset_id, user_id, status, error_message, output_path, created_at, updated_at, completed_at
			   FROM jobs WHERE user_id = $1 AND status = 'done' ORDER BY completed_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*model.Job
	for rows.Next() {
		job := &model.Job{}
		err := rows.Scan(
			&job.ID, &job.AssetID, &job.UserID, &job.Status,
			&job.ErrorMessage, &job.OutputPath, &job.CreatedAt, &job.UpdatedAt, &job.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (r *JobPostgresRepository) UpdateToProcessing(ctx context.Context, jobID uuid.UUID) error {
	query := `UPDATE jobs SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, model.JobStatusProcessing, jobID)
	return err
}

func (r *JobPostgresRepository) UpdateToComplete(ctx context.Context, jobID uuid.UUID, outputPath string) error {
	query := `UPDATE jobs SET status = $1, output_path = $2, completed_at = NOW(), updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(ctx, query, model.JobStatusDone, outputPath, jobID)
	return err
}

func (r *JobPostgresRepository) UpdateToFailed(ctx context.Context, jobID uuid.UUID, errorMessage string) error {
	query := `UPDATE jobs SET status = $1, error_message = $2, completed_at = NOW(), updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(ctx, query, model.JobStatusFailed, errorMessage, jobID)
	return err
}
