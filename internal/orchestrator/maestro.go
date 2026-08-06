package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maestro/internal/config"
	"maestro/internal/domain"
	"maestro/internal/repository"
	"os"
	"os/exec"
	"time"
)

type Orchestrator interface {
	ExecuteJob(result ExecutionResult) (ExecutionResult, error)
}

type ExecutionResult struct {
	JobID     string
	Status    string
	ExitCode  int
	StartedAt time.Time
	EndedAt   time.Time
	Logs      string          // stderr do container
	Payload   json.RawMessage // stdout do Python — preservado como-está
	Job       domain.Job
	Task      domain.Task
}

type FootballOrchestrator struct {
	Config *config.Config
	DB     repository.MaestroRepository
}

func (f *FootballOrchestrator) ExecuteJob(result ExecutionResult) (ExecutionResult, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	ctx := context.Background()
	if err := f.DB.SetJobPending(ctx, result.Job); err != nil {
		return ExecutionResult{}, err
	}
	inputFile, outputDir, err := buildVolumes(result.JobID)

	cmd := exec.Command("docker", "run", "--rm", "-v", inputFile, "-v", outputDir, "football-rpa")
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err = cmd.Run()

	result.EndedAt = time.Now()
	result.Job.FinishedAt = result.EndedAt
	result.Task.FinishedAt = result.EndedAt
	result.Payload = stdout.Bytes()
	result.Logs = stderr.String()

	if err != nil {
		result.ExitCode = 1

		// Se o processo retornou exit code != 0
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		return result, fmt.Errorf(
			"erro ao executar football-rpa (exit code %d): %w\nstderr: %s",
			result.ExitCode,
			err,
			result.Logs,
		)
	}

	result.ExitCode = 0
	if err := f.DB.FinishJob(ctx, result.Job); err != nil {
		return ExecutionResult{}, err
	}
	return result, nil
}

func buildVolumes(jobID string) (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	inputFile := fmt.Sprintf("%s/scripts/football/input-%s.csv:/app/input-%s.csv", cwd, jobID, jobID)
	outputDir := fmt.Sprintf("%s/scripts/football/output-%s:/app/Prints-%s", cwd, jobID, jobID)

	return inputFile, outputDir, nil
}
