package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"maestro/internal/config"
	"maestro/internal/domain"
	"maestro/internal/executor"
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
	Logs      []byte          // stderr do container
	Payload   json.RawMessage // stdout do Python — preservado como-está
	Job       domain.Job
	Task      domain.Task
}

type FootballOrchestrator struct {
	Config   *config.Config
	DB       repository.MaestroRepository
	Executor executor.Executor
}

func (f *FootballOrchestrator) ExecuteJob(result ExecutionResult) (ExecutionResult, error) {
	ctx := context.Background()
	if err := f.DB.SetJobPending(ctx, result.Job); err != nil {
		return ExecutionResult{}, err
	}
	inputFile, outputDir, err := buildVolumes(result.Job.IssueKey)
	argsString := []string{"run", "--rm", "-v", inputFile, "-v", outputDir}
	stdout, stderr, err := f.Executor.Execute(ctx, "docker", argsString)

	result.EndedAt = time.Now()
	result.Job.FinishedAt = result.EndedAt
	result.Task.FinishedAt = result.EndedAt
	result.Payload = stdout
	result.Logs = stderr

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
