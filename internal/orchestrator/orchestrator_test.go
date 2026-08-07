package orchestrator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maestro/internal/domain"
	"maestro/internal/executor/fake"
	FakeDB "maestro/internal/repository/fakeDB"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestBuildVolumes(t *testing.T) {
	t.Run("Build String sem erros", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("erro ao descobrir diretório")
		}
		expOutputDir := fmt.Sprintf("%s/scripts/football/output-teste:/app/Prints-teste", cwd)
		expInputFile := fmt.Sprintf("%s/scripts/football/input-teste.csv:/app/input-teste.csv", cwd)
		inputFile, outputDir, err := buildVolumes("teste")
		if err != nil {
			t.Fatalf("Erro ao gerar string de volumes: %v", err)
		}
		if inputFile != expInputFile {
			t.Errorf("erro ao gerar string de inputFile - Exp:%v - Rec: %v", expInputFile, inputFile)
		}
		if outputDir != expOutputDir {
			t.Errorf("erro ao gerar string de outputDir - Exp:%v - Rec: %v", expOutputDir, outputDir)
		}
	})
}

func TestExecuteJob(t *testing.T) {
	t.Run("Execução sem erros validando leitura de stderr,stdout,err", func(t *testing.T) {
		exe := fake.Fake{
			StdOut: []byte("stdout"),
			StdErr: []byte{},
			Err:    nil,
		}
		db := FakeDB.FakeJobDB{
			Job:     domain.Job{},
			Context: context.Background(),
			Error:   nil,
		}
		now := time.Now()
		result := ExecutionResult{
			JobID:     "jobId",
			Status:    "status",
			ExitCode:  1,
			StartedAt: now,
			Job:       domain.Job{},
			Task:      domain.Task{},
		}
		orch := FootballOrchestrator{
			DB:       db,
			Executor: exe,
		}

		resultOut, err := orch.ExecuteJob(result)
		if err != nil {
			t.Fatalf("Erro ao testar ExecuteJob: %v", err)
		}
		switch {
		case !bytes.Equal(resultOut.Logs, exe.StdErr):
			t.Fatalf("Erro ao testar Logs/Stderr do Execute: %v", resultOut.Logs)
		case !bytes.Equal(resultOut.Payload, exe.StdOut):
			t.Fatalf("Erro ao testar StdOut do Execute: %v", resultOut.Payload)
		}
	})
	t.Run("ExitCode no resultado do exec", func(t *testing.T) {
		cmd := exec.Command("bash", "-c", "exit 42")
		err := cmd.Run()
		exe := fake.Fake{
			StdOut: []byte("stdout"),
			StdErr: []byte{},
			Err:    err,
		}
		db := FakeDB.FakeJobDB{
			Job:     domain.Job{},
			Context: context.Background(),
			Error:   nil,
		}
		now := time.Now()
		result := ExecutionResult{
			JobID:     "jobId",
			Status:    "status",
			ExitCode:  1,
			StartedAt: now,
			Job:       domain.Job{},
			Task:      domain.Task{},
		}
		orch := FootballOrchestrator{
			DB:       db,
			Executor: exe,
		}
		resultOut, err := orch.ExecuteJob(result)

		if err == nil {
			t.Fatalf("Erro ao criar erro no Err")
		}
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatal("Erro ao criar tipo de erro exitErr")
		}
		if resultOut.ExitCode == 0 {
			t.Fatalf("Erro ao devolver erro. ExitCode: %v", resultOut.ExitCode)
		}
		if resultOut.ExitCode != 42 {
			t.Fatalf("Erro ao devolver erro. ExitCode exp: %v - ExitCode rec:%v", resultOut.ExitCode, 42)
		}
		if !bytes.Equal(resultOut.Logs, exe.StdErr) {
			t.Fatalf("stderr separado: %s - recebido: %s", exe.StdErr, resultOut.Logs)
		}
	})
}
