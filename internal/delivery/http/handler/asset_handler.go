package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"backend/configs"
	"backend/internal/delivery/http/middleware"
	"backend/internal/usecase"
	"backend/pkg/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AssetHandler struct {
	assetUsecase *usecase.AssetUsecase
	cfg          configs.Config
	logger       *slog.Logger
}

func NewAssetHandler(uc *usecase.AssetUsecase, cfg configs.Config, logger *slog.Logger) *AssetHandler {
	return &AssetHandler{
		assetUsecase: uc,
		cfg:          cfg,
		logger:       logger,
	}
}

// UploadAsset godoc
// @Summary Upload an asset to a project
// @Description Uploads an image or video file to a specific project.
// @Tags assets
// @Accept  multipart/form-data
// @Produce  json
// @Param   id   path      string  true  "Project ID"
// @Param   file formData  file    true  "Asset file"
// @Success 201 {object} model.Asset
// @Failure 400 {object} utils.ErrorResponse "Invalid project ID or file error"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Access denied"
// @Failure 404 {object} utils.ErrorResponse "Project not found"
// @Failure 413 {object} utils.ErrorResponse "File size exceeds limit"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{id}/assets [post]
func (h *AssetHandler) UploadAsset(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	maxUploadSize := int64(h.cfg.MaxFileSizeMB) * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		utils.RespondWithError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("File size exceeds the limit of %dMB", h.cfg.MaxFileSizeMB))
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid file")
		return
	}
	defer file.Close()

	asset, err := h.assetUsecase.UploadAsset(r.Context(), userID, projectID, file, fileHeader)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrProjectNotFound):
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, usecase.ErrProjectAccessDenied):
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, usecase.ErrFileSizeExceeded):
			utils.RespondWithError(w, http.StatusRequestEntityTooLarge, err.Error())
		case errors.Is(err, usecase.ErrInvalidMIMEType):
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			h.logger.Error("Failed to upload asset", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to upload asset")
		}
		return
	}

	h.logger.Info("Asset uploaded successfully", "asset_id", asset.ID, "project_id", projectID)
	utils.RespondWithJSON(w, http.StatusCreated, asset)
}

// ListAssets godoc
// @Summary List assets in a project
// @Description Retrieves a list of all assets for a specific project.
// @Tags assets
// @Produce  json
// @Param   id   path      string  true  "Project ID"
// @Success 200 {array} model.Asset
// @Failure 400 {object} utils.ErrorResponse "Invalid project ID"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Access denied"
// @Failure 404 {object} utils.ErrorResponse "Project not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{id}/assets [get]
func (h *AssetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	assets, err := h.assetUsecase.ListProjectAssets(r.Context(), userID, projectID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrProjectNotFound):
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, usecase.ErrProjectAccessDenied):
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		default:
			h.logger.Error("Failed to list assets", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list assets")
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, assets)
}

// GetAsset godoc
// @Summary Retrieve a specific asset file
// @Description Serves the asset file if it belongs to the requesting user.
// @Tags assets
// @Produce  octet-stream
// @Param   id   path      string  true  "Asset ID"
// @Success 200 {file} file
// @Failure 400 {object} utils.ErrorResponse "Invalid asset ID"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Access denied"
// @Failure 404 {object} utils.ErrorResponse "Asset not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /assets/{id} [get]
func (h *AssetHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	assetIDStr := chi.URLParam(r, "id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid asset ID")
		return
	}

	asset, filePath, err := h.assetUsecase.GetAssetByID(r.Context(), userID, assetID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrAssetNotFound):
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, usecase.ErrAssetAccessDenied):
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		default:
			h.logger.Error("Failed to get asset", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get asset")
		}
		return
	}

	// Check if file exists before serving
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.logger.Error("Asset file not found on disk", "path", filePath, "asset_id", asset.ID)
		utils.RespondWithError(w, http.StatusNotFound, "Asset file not found")
		return
	}

	w.Header().Set("Content-Type", asset.MIMEType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filepath.Base(asset.OriginalFilename)))
	http.ServeFile(w, r, filePath)
}
