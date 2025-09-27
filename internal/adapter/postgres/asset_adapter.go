package postgres

import (
	"context"
	"errors"

	"backend/internal/domain/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAssetNotFound = errors.New("asset not found")

type AssetPostgresRepository struct {
	db *pgxpool.Pool
}

func NewAssetPostgresRepository(db *pgxpool.Pool) *AssetPostgresRepository {
	return &AssetPostgresRepository{db: db}
}

func (r *AssetPostgresRepository) Create(ctx context.Context, asset *model.Asset) error {
	query := `INSERT INTO assets (id, project_id, original_filename, stored_filename, mime_type, size) VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at`
	err := r.db.QueryRow(ctx, query, asset.ID, asset.ProjectID, asset.OriginalFilename, asset.StoredFilename, asset.MIMEType, asset.Size).Scan(&asset.CreatedAt)
	return err
}

func (r *AssetPostgresRepository) FindAllByProjectID(ctx context.Context, projectID uuid.UUID) ([]*model.Asset, error) {
	query := `SELECT id, project_id, original_filename, stored_filename, mime_type, size, created_at FROM assets WHERE project_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := []*model.Asset{}
	for rows.Next() {
		var a model.Asset
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.OriginalFilename, &a.StoredFilename, &a.MIMEType, &a.Size, &a.CreatedAt); err != nil {
			return nil, err
		}
		assets = append(assets, &a)
	}

	return assets, nil
}

func (r *AssetPostgresRepository) FindByID(ctx context.Context, assetID uuid.UUID) (*model.Asset, error) {
	query := `SELECT id, project_id, original_filename, stored_filename, mime_type, size, created_at FROM assets WHERE id = $1`
	row := r.db.QueryRow(ctx, query, assetID)

	var a model.Asset
	err := row.Scan(&a.ID, &a.ProjectID, &a.OriginalFilename, &a.StoredFilename, &a.MIMEType, &a.Size, &a.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}

	return &a, nil
}
