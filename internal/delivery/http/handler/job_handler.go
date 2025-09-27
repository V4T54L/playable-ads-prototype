package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"backend/internal/delivery/http/middleware"
	"backend/internal/usecase"
	"backend/pkg/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type JobHandler struct {
	jobUsecase *usecase.JobUsecase
	logger     *slog.Logger
}

func NewJobHandler(uc *usecase.JobUsecase, logger *slog.Logger) *JobHandler {
	return &JobHandler{
		jobUsecase: uc,
		logger:     logger,
	}
}

// EnqueueRenderJob godoc
// @Summary Enqueue a rendering job
// @Description Creates and enqueues a rendering job for a specific asset in a project.
// @Tags jobs
// @Produce  json
// @Param   projectId   path      string  true  "Project ID"
// @Param   assetId     path      string  true  "Asset ID"
// @Success 202 {object} model.Job
// @Failure 400 {object} utils.ErrorResponse "Invalid project or asset ID"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Access denied"
// @Failure 404 {object} utils.ErrorResponse "Project or asset not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{projectId}/render/{assetId} [post]
func (h *JobHandler) EnqueueRenderJob(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectIDStr := chi.URLParam(r, "projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	assetIDStr := chi.URLParam(r, "assetId")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid asset ID")
		return
	}

	job, err := h.jobUsecase.CreateRenderJob(r.Context(), userID, projectID, assetID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrProjectNotFound) || errors.Is(err, usecase.ErrAssetNotFound):
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, usecase.ErrProjectAccessDenied) || errors.Is(err, usecase.ErrAssetAccessDenied):
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		default:
			h.logger.Error("Failed to enqueue render job", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to enqueue job")
		}
		return
	}

	h.logger.Info("Render job enqueued", "job_id", job.ID, "asset_id", job.AssetID)
	utils.RespondWithJSON(w, http.StatusAccepted, job)
}

// ListJobs godoc
// @Summary List user's rendering jobs
// @Description Retrieves a list of all rendering jobs for the authenticated user.
// @Tags jobs
// @Produce  json
// @Success 200 {array} model.Job
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /jobs [get]
func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	jobs, err := h.jobUsecase.ListUserJobs(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to list jobs", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list jobs")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, jobs)
}

// GetJobStatus godoc
// @Summary Get job status
// @Description Retrieves the status of a specific rendering job.
// @Tags jobs
// @Produce  json
// @Param   id   path      string  true  "Job ID"
// @Success 200 {object} model.Job
// @Failure 400 {object} utils.ErrorResponse "Invalid job ID"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Access denied"
// @Failure 404 {object} utils.ErrorResponse "Job not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /jobs/{id} [get]
func (h *JobHandler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	jobIDStr := chi.URLParam(r, "id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	job, err := h.jobUsecase.GetJobByID(r.Context(), userID, jobID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrJobNotFound):
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, usecase.ErrJobAccessDenied):
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		default:
			h.logger.Error("Failed to get job status", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get job status")
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, job)
}

// ListOutputs godoc
// @Summary List user's rendered outputs
// @Description Retrieves a list of all completed rendering jobs (outputs) for the authenticated user.
// @Tags jobs
// @Produce  json
// @Success 200 {array} model.Job
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /outputs [get]
func (h *JobHandler) ListOutputs(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	outputs, err := h.jobUsecase.ListUserOutputs(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to list outputs", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list outputs")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, outputs)
}
