package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"backend/internal/delivery/http/middleware"
	"backend/internal/usecase"
	"backend/pkg/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectUsecase *usecase.ProjectUsecase
	logger         *slog.Logger
}

func NewProjectHandler(uc *usecase.ProjectUsecase, logger *slog.Logger) *ProjectHandler {
	return &ProjectHandler{
		projectUsecase: uc,
		logger:         logger,
	}
}

type CreateProjectRequest struct {
	Title       string `json:"title" example:"My First Ad"`
	Description string `json:"description" example:"A cool playable ad project."`
}

// CreateProject godoc
// @Summary Create a new project
// @Description Creates a new project for the authenticated user.
// @Tags projects
// @Accept  json
// @Produce  json
// @Param   project  body      CreateProjectRequest  true  "Project Info"
// @Success 201 {object} model.Project
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /projects [post]
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	project, err := h.projectUsecase.CreateProject(r.Context(), userID, req.Title, req.Description)
	if err != nil {
		h.logger.Error("Failed to create project", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create project")
		return
	}

	h.logger.Info("Project created successfully", "project_id", project.ID, "user_id", userID)
	utils.RespondWithJSON(w, http.StatusCreated, project)
}

// ListProjects godoc
// @Summary List user's projects
// @Description Retrieves a list of all projects belonging to the authenticated user.
// @Tags projects
// @Produce  json
// @Success 200 {array} model.Project
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /projects [get]
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projects, err := h.projectUsecase.ListUserProjects(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to list projects", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list projects")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, projects)
}

// GetProject godoc
// @Summary Get a specific project
// @Description Retrieves a single project by its ID, if owned by the user.
// @Tags projects
// @Produce  json
// @Param   id   path      string  true  "Project ID"
// @Success 200 {object} model.Project
// @Failure 400 {object} utils.ErrorResponse "Invalid project ID"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Access denied"
// @Failure 404 {object} utils.ErrorResponse "Project not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{id} [get]
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
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

	project, err := h.projectUsecase.GetProjectByID(r.Context(), userID, projectID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrProjectNotFound):
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, usecase.ErrProjectAccessDenied):
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		default:
			h.logger.Error("Failed to get project", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get project")
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, project)
}
