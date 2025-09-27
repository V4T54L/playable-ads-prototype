package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/configs"
	"backend/internal/adapter/postgres"
	"backend/internal/adapter/redis"
	router "backend/internal/delivery/http"
	"backend/internal/delivery/http/handler"
	"backend/internal/delivery/worker"
	"backend/internal/usecase"
	"backend/pkg/auth"

	"github.com/joho/godotenv"
)

// @title Playable Ads Manager Backend API
// @version 1.0
// @description This is the backend API for the Playable Ads Manager Platform.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {

	if os.Getenv("ENVIRONMENT") != "Production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal(".env not found. Set 'ENVIRONMENT'='Production' if required.")
		}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := configs.LoadConfig()
	if err != nil {
		logger.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbPool, err := postgres.NewDBPool(cfg)
	if err != nil {
		logger.Error("cannot connect to db", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	logger.Info("database connected successfully")

	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		logger.Error("cannot connect to redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	logger.Info("connected to redis")

	userRepo := postgres.NewUserPostgresRepository(dbPool)
	tokenRepo := postgres.NewTokenPostgresRepository(dbPool)
	projectRepo := postgres.NewProjectPostgresRepository(dbPool)
	assetRepo := postgres.NewAssetPostgresRepository(dbPool)
	jobRepo := postgres.NewJobPostgresRepository(dbPool)
	analyticsRepo := postgres.NewAnalyticsPostgresRepository(dbPool)

	jobQueue := redis.NewJobQueueAdapter(redisClient)

	jwtMaker, err := auth.NewJWTMaker(cfg.JWTSecret)
	if err != nil {
		logger.Error("cannot create JWT maker", "error", err)
		os.Exit(1)
	}

	userUsecase := usecase.NewUserUsecase(userRepo, tokenRepo, jwtMaker, cfg)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)
	assetUsecase := usecase.NewAssetUsecase(assetRepo, projectRepo, cfg)
	jobUsecase := usecase.NewJobUsecase(jobRepo, assetRepo, projectRepo, jobQueue)
	analyticsUsecase := usecase.NewAnalyticsUsecase(analyticsRepo, projectRepo)

	authHandler := handler.NewAuthHandler(userUsecase, logger)
	projectHandler := handler.NewProjectHandler(projectUsecase, logger)
	assetHandler := handler.NewAssetHandler(assetUsecase, cfg, logger)
	jobHandler := handler.NewJobHandler(jobUsecase, logger)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsUsecase, logger)

	mainRouter := router.NewRouter(authHandler, projectHandler, assetHandler, jobHandler, analyticsHandler, jwtMaker, logger)

	jobWorker := worker.NewJobWorker(jobRepo, assetRepo, jobQueue, logger)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: mainRouter,
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info("Shutting down server...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Server forced to shutdown", "error", err)
		}
	}()


	go func() {
		logger.Info("Starting job worker")
		jobWorker.Start(ctx)
		logger.Info("Job worker stopped")
	}()

	logger.Info(fmt.Sprintf("Starting server on port %s", cfg.ServerPort))
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exiting")
}
