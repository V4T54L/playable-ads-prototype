package handler

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"

	"backend/internal/usecase"
	"backend/pkg/utils"
)

type AuthHandler struct {
	userUsecase *usecase.UserUsecase
	logger      *slog.Logger
}

func NewAuthHandler(uc *usecase.UserUsecase, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		userUsecase: uc,
		logger:      logger,
	}
}

type RegisterRequest struct {
	Username string `json:"username" example:"newuser"`
	Password string `json:"password" example:"password123"`
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account with a username and password.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   user  body      RegisterRequest  true  "User Registration Info"
// @Success 201 {object} model.User
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 409 {object} utils.ErrorResponse "Username already exists"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.userUsecase.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			utils.RespondWithError(w, http.StatusConflict, err.Error())
		} else {
			log.Println("RefisterHandler Error: ", err)
			h.logger.Error("Failed to register user", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to register user")
		}
		return
	}

	h.logger.Info("User registered successfully", "username", user.Username, "user_id", user.ID)
	utils.RespondWithJSON(w, http.StatusCreated, user)
}

type LoginRequest struct {
	Username string `json:"username" example:"newuser"`
	Password string `json:"password" example:"password123"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login godoc
// @Summary Log in a user
// @Description Authenticates a user and returns JWT and refresh tokens.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   credentials  body      LoginRequest  true  "User Login Credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 401 {object} utils.ErrorResponse "Invalid credentials"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	accessToken, refreshToken, err := h.userUsecase.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		} else {
			h.logger.Error("Failed to login user", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to login")
		}
		return
	}

	h.logger.Info("User logged in successfully", "username", req.Username)
	utils.RespondWithJSON(w, http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

// Refresh godoc
// @Summary Refresh access token
// @Description Provides a new access token using a valid refresh token.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   token  body      RefreshRequest  true  "Refresh Token"
// @Success 200 {object} RefreshResponse
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 401 {object} utils.ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	accessToken, err := h.userUsecase.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, usecase.ErrRefreshTokenExpired) || errors.Is(err, usecase.ErrRefreshTokenRevoked) {
			utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		} else {
			h.logger.Error("Failed to refresh token", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to refresh token")
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, RefreshResponse{
		AccessToken: accessToken,
	})
}
