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
	cwd, err := os.Getwd()
	if err != nil {
		return ExecutionResult{}, err
	}
	inputFile := fmt.Sprintf("%s/scripts/football/input.csv:/app/input.csv", cwd)
	outputDir := fmt.Sprintf("%s/scripts/football/output:/app/Prints", cwd)

	stdoutBytes := stdout.Bytes()
	stderrStr := stderr.String()

	cmd := exec.Command("docker", "run", "--rm", "-v", inputFile, "-v", outputDir, "football-rpa")
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err = cmd.Run()
	result.EndedAt = time.Now()
	result.Job.FinishedAt = result.EndedAt
	result.Task.FinishedAt = result.EndedAt
	result.Payload = stdoutBytes
	result.Logs = stderrStr
	result.ExitCode = 0
	if err := f.DB.FinishJob(ctx, result.Job); err != nil {
		return ExecutionResult{}, err
	}
	return result, nil
}
