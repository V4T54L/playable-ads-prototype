package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"backend/internal/domain/repository"

	"github.com/google/uuid"
)

const (
	uploadDir  = "uploads"
	outputDir  = "outputs"
	workerName = "JobWorker"
)

type JobQueue interface {
	Dequeue(ctx context.Context) (uuid.UUID, error)
}

type JobWorker struct {
	jobRepo   repository.JobRepository
	assetRepo repository.AssetRepository
	jobQueue  JobQueue
	logger    *slog.Logger
}

func NewJobWorker(jobRepo repository.JobRepository, assetRepo repository.AssetRepository, jobQueue JobQueue, logger *slog.Logger) *JobWorker {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Error("failed to create output directory", "error", err, "path", outputDir)
	}
	return &JobWorker{
		jobRepo:   jobRepo,
		assetRepo: assetRepo,
		jobQueue:  jobQueue,
		logger:    logger.With("worker", workerName),
	}
}

func (w *JobWorker) Start(ctx context.Context) {
	w.logger.Info("starting job worker")
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping job worker")
			return
		default:
			w.processJob(ctx)
		}
	}
}

func (w *JobWorker) processJob(ctx context.Context) {
	jobID, err := w.jobQueue.Dequeue(ctx)
	if err != nil {
		if ctx.Err() == nil { // Don't log error if context was cancelled
			w.logger.Error("failed to dequeue job", "error", err)
		}
		// Sleep briefly to prevent tight loop on persistent errors
		time.Sleep(5 * time.Second)
		return
	}

	w.logger.Info("processing job", "job_id", jobID)

	// Mark job as processing
	if err := w.jobRepo.UpdateToProcessing(ctx, jobID); err != nil {
		w.logger.Error("failed to update job status to processing", "job_id", jobID, "error", err)
		return
	}

	// Fetch job and asset details
	job, err := w.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		w.logger.Error("failed to find job", "job_id", jobID, "error", err)
		return
	}
	asset, err := w.assetRepo.FindByID(ctx, job.AssetID)
	if err != nil {
		w.logger.Error("failed to find asset for job", "job_id", jobID, "asset_id", job.AssetID, "error", err)
		_ = w.jobRepo.UpdateToFailed(ctx, jobID, "asset not found")
		return
	}

	// Perform rendering
	outputPath, err := w.renderAsset(ctx, asset.StoredFilename)
	if err != nil {
		w.logger.Error("failed to render asset", "job_id", jobID, "asset_id", asset.ID, "error", err)
		_ = w.jobRepo.UpdateToFailed(ctx, jobID, err.Error())
		return
	}

	// Mark job as complete
	if err := w.jobRepo.UpdateToComplete(ctx, jobID, outputPath); err != nil {
		w.logger.Error("failed to update job status to done", "job_id", jobID, "error", err)
		return
	}

	w.logger.Info("job completed successfully", "job_id", jobID, "output_path", outputPath)
}

func (w *JobWorker) renderAsset(ctx context.Context, storedFilename string) (string, error) {
	// Create a context with a timeout for the rendering process
	renderCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	inputPath := filepath.Join(uploadDir, storedFilename)
	outputFilename := fmt.Sprintf("rendered-%s", storedFilename)
	outputPath := filepath.Join(outputDir, outputFilename)

	// Example FFmpeg command: compress video
	// This is a simulation. A real-world scenario would be more complex.
	cmd := exec.CommandContext(renderCtx, "ffmpeg",
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "28",
		"-y", // Overwrite output file if it exists
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if renderCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("rendering timed out: %w", err)
		}
		return "", fmt.Errorf("ffmpeg error: %s\n%s", err, string(output))
	}

	return outputPath, nil
}
