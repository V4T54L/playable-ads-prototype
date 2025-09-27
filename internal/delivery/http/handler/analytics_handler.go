package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"backend/internal/delivery/http/middleware"
	"backend/internal/domain/model"
	"backend/internal/usecase"
	"backend/pkg/utils"

	"github.com/google/uuid"
)

type AnalyticsHandler struct {
	analyticsUsecase *usecase.AnalyticsUsecase
	logger           *slog.Logger
}

type LogEventRequest struct {
	ProjectID string          `json:"project_id" example:"f47ac10b-58cc-4372-a567-0e02b2c3d479"`
	EventType model.EventType `json:"event_type" example:"play"`
}

func NewAnalyticsHandler(uc *usecase.AnalyticsUsecase, logger *slog.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsUsecase: uc,
		logger:           logger,
	}
}

// LogEvent godoc
// @Summary Log an analytics event
// @Description Logs an analytics event for a project owned by the user.
// @Tags analytics
// @Accept  json
// @Produce  json
// @Param   event  body      LogEventRequest  true  "Analytics Event Info"
// @Success 201 {object} model.AnalyticsEvent
// @Failure 400 {object} utils.ErrorResponse "Invalid request body or event type"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Access denied"
// @Failure 404 {object} utils.ErrorResponse "Project not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /analytics [post]
func (h *AnalyticsHandler) LogEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req LogEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
		return
	}

	event, err := h.analyticsUsecase.LogEvent(r.Context(), userID, projectID, req.EventType)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrProjectNotFound):
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, usecase.ErrProjectAccessDenied):
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, usecase.ErrInvalidEventType):
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			h.logger.Error("Failed to log event", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to log event")
		}
		return
	}

	h.logger.Info("Analytics event logged", "event_id", event.ID, "project_id", event.ProjectID, "user_id", event.UserID)
	utils.RespondWithJSON(w, http.StatusCreated, event)
}

// ListEvents godoc
// @Summary List all analytics events
// @Description Retrieves a list of all analytics events across all projects. This is a public endpoint.
// @Tags analytics
// @Produce  json
// @Success 200 {array} model.AnalyticsEvent
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /analytics [get]
func (h *AnalyticsHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.analyticsUsecase.ListAllEvents(r.Context())
	if err != nil {
		h.logger.Error("Failed to list analytics events", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve events")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, events)
}
