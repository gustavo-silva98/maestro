package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"maestro/internal/config"
	"maestro/internal/domain"
	"maestro/internal/executor"
	"maestro/internal/repository"
	"os/exec"
	"time"

	"github.com/google/uuid"
)

type Orchestrator interface {
	ExecuteJob(ctx context.Context, job domain.Job, jt config.JobType) (domain.Job, error)
}

type JobOrchestrator struct {
	DB       repository.JobTaskRepo
	Executor executor.Executor
	BaseDir  string
}

func (jo *JobOrchestrator) ExecuteJob(ctx context.Context, job domain.Job, jt config.JobType) (domain.Job, error) {
	if err := jo.DB.SetJobPending(ctx, job); err != nil {
		return domain.Job{}, err
	}

	task, err := jo.runTask(ctx, job.ID, jt, job.ItemCount)
	if err != nil {
		job.Status = domain.StatusFailed
		job.FinishedAt = time.Now()
		if dbErr := jo.DB.SetJobFailed(ctx, job); dbErr != nil {
			return domain.Job{}, fmt.Errorf("job falhou e não foi possível persistir: %w", dbErr)
		}
	}
	if dbErr := jo.DB.CreateTask(ctx, task); dbErr != nil {
		return domain.Job{}, fmt.Errorf("erro ao persistir task: %w", dbErr)
	}
	job.Status = task.Status
	job.FinishedAt = task.FinishedAt
	if err := jo.DB.FinishJob(ctx, job); err != nil {
		return domain.Job{}, err
	}
	return job, nil
}

func buildVolumes(baseDir, scriptsDir, jobID string) (string, string) {
	inputFile := fmt.Sprintf("%s/%s/input-%s.csv:/app/input-%s.csv", baseDir, scriptsDir, jobID, jobID)
	outputDir := fmt.Sprintf("%s/%s/output-%s:/app/output-%s", baseDir, scriptsDir, jobID, jobID)

	return inputFile, outputDir
}

func (jo *JobOrchestrator) runTask(ctx context.Context, jobID string, jt config.JobType, itemCount int) (domain.Task, error) {
	inputFile, outputDir := buildVolumes(jo.BaseDir, jt.Container.ContainerDir, jobID)
	createdDate := time.Now()
	args := []string{"run", "--rm", "-v", inputFile, "-v", outputDir, jt.Container.ImageName}
	stdout, stderr, err := jo.Executor.Execute(ctx, "docker", args)

	task := domain.Task{
		ID:           uuid.New().String(),
		JobID:        jobID,
		Image:        jt.Container.ImageName,
		Payload:      stdout,
		Logs:         stderr,
		TaskType:     jt.JobType,
		SavedMinutes: jt.Indicators.TimeSaved * itemCount,
		CreatedAt:    createdDate,
		FinishedAt:   time.Now(),
	}

	if err != nil {
		task.Status = domain.StatusFailed
		task.ExitCode = 1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			task.ExitCode = exitErr.ExitCode()
		}
		return task, fmt.Errorf(
			"erro ao executar %s (exit code %d): %w\nstderr: %s",
			jt.Container.ImageName, task.ExitCode, err, task.Logs,
		)
	}
	task.Status = domain.StatusSuccess
	task.ExitCode = 0
	return task, nil
}
