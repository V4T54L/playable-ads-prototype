package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/configs"
	"backend/internal/adapter/postgres"
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
	"backend/pkg/auth"
	"backend/pkg/utils"

	"github.com/google/uuid"
)

var ErrInvalidCredentials = errors.New("invalid username or password")
var ErrRefreshTokenExpired = errors.New("refresh token has expired")
var ErrRefreshTokenRevoked = errors.New("refresh token has been revoked")
var ErrUserAlreadyExists = errors.New("user already exists")

type UserUsecase struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	jwtMaker  *auth.JWTMaker
	cfg       configs.Config
}

func NewUserUsecase(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, jwtMaker *auth.JWTMaker, cfg configs.Config) *UserUsecase {
	return &UserUsecase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtMaker:  jwtMaker,
		cfg:       cfg,
	}
}

func (uc *UserUsecase) Register(ctx context.Context, username, password string) (*model.User, error) {
	if len(username) < 3 || len(password) < 3 {
		return nil, errors.New("username and password must be at least 3 characters long")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: hashedPassword,
	}

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *UserUsecase) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error) {
	user, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return "", "", ErrInvalidCredentials
	}

	accessToken, err = uc.jwtMaker.CreateToken(user.ID, uc.cfg.AccessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to create access token: %w", err)
	}

	rt := &model.RefreshToken{
		Token:     uuid.New(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(uc.cfg.RefreshTokenTTL),
	}

	err = uc.tokenRepo.Create(ctx, rt)
	if err != nil {
		return "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return accessToken, rt.Token.String(), nil
}

func (uc *UserUsecase) Refresh(ctx context.Context, refreshTokenString string) (string, error) {
	refreshTokenID, err := uuid.Parse(refreshTokenString)
	if err != nil {
		return "", ErrRefreshTokenRevoked
	}

	refreshToken, err := uc.tokenRepo.Find(ctx, refreshTokenID)
	if err != nil {
		if errors.Is(err, postgres.ErrTokenNotFound) {
			return "", ErrRefreshTokenRevoked
		}
		return "", err
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		// Clean up expired token
		_ = uc.tokenRepo.Delete(ctx, refreshToken.Token)
		return "", ErrRefreshTokenExpired
	}

	user, err := uc.userRepo.FindByID(ctx, refreshToken.UserID)
	if err != nil {
		return "", fmt.Errorf("user associated with refresh token not found: %w", err)
	}

	accessToken, err := uc.jwtMaker.CreateToken(user.ID, uc.cfg.AccessTokenTTL)
	if err != nil {
		return "", fmt.Errorf("failed to create new access token: %w", err)
	}

	return accessToken, nil
}
