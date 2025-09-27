package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"backend/configs"
	"backend/internal/domain/model"
	"backend/internal/domain/repository"

	"github.com/google/uuid"
)

var (
	ErrFileSizeExceeded = errors.New("file size exceeds the limit")
	ErrInvalidMIMEType  = errors.New("invalid or unsupported MIME type")
)

const uploadDir = "uploads"

type AssetUsecase struct {
	assetRepo   repository.AssetRepository
	projectRepo repository.ProjectRepository
	cfg         configs.Config
}

func NewAssetUsecase(assetRepo repository.AssetRepository, projectRepo repository.ProjectRepository, cfg configs.Config) *AssetUsecase {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		panic(fmt.Sprintf("failed to create upload directory: %v", err))
	}
	return &AssetUsecase{
		assetRepo:   assetRepo,
		projectRepo: projectRepo,
		cfg:         cfg,
	}
}

func (uc *AssetUsecase) UploadAsset(ctx context.Context, userID, projectID uuid.UUID, file multipart.File, fileHeader *multipart.FileHeader) (*model.Asset, error) {
	// 1. Verify project ownership
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

	// 2. Validate file size
	maxSize := int64(uc.cfg.MaxFileSizeMB) * 1024 * 1024
	if fileHeader.Size > maxSize {
		return nil, fmt.Errorf("%w: max size is %dMB", ErrFileSizeExceeded, uc.cfg.MaxFileSizeMB)
	}

	// 3. Validate MIME type (simple validation for now)
	mimeType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(mimeType, "image/") && !strings.HasPrefix(mimeType, "video/") {
		return nil, fmt.Errorf("%w: got %s", ErrInvalidMIMEType, mimeType)
	}

	// 4. Generate a unique filename
	ext := filepath.Ext(fileHeader.Filename)
	storedFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(uploadDir, storedFilename)

	// 5. Save the file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		return nil, err
	}

	// 6. Create asset record in DB
	asset := &model.Asset{
		ID:               uuid.New(),
		ProjectID:        projectID,
		OriginalFilename: fileHeader.Filename,
		StoredFilename:   storedFilename,
		MIMEType:         mimeType,
		Size:             fileHeader.Size,
	}

	if err := uc.assetRepo.Create(ctx, asset); err != nil {
		// Attempt to clean up the saved file on DB error
		os.Remove(filePath)
		return nil, err
	}

	return asset, nil
}

func (uc *AssetUsecase) ListProjectAssets(ctx context.Context, userID, projectID uuid.UUID) ([]*model.Asset, error) {
	// Verify project ownership
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

	return uc.assetRepo.FindAllByProjectID(ctx, projectID)
}

func (uc *AssetUsecase) GetAssetByID(ctx context.Context, userID, assetID uuid.UUID) (*model.Asset, string, error) {
	asset, err := uc.assetRepo.FindByID(ctx, assetID)
	if err != nil {
		if errors.Is(err, ErrAssetNotFound) {
			return nil, "", ErrAssetNotFound
		}
		return nil, "", err
	}

	// Verify ownership via project
	project, err := uc.projectRepo.FindByID(ctx, asset.ProjectID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return nil, "", ErrAssetAccessDenied // Obscure the reason
		}
		return nil, "", err
	}
	if project.UserID != userID {
		return nil, "", ErrAssetAccessDenied
	}

	filePath := filepath.Join(uploadDir, asset.StoredFilename)
	return asset, filePath, nil
}
