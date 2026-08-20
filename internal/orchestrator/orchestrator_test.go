package orchestrator

import (
	"context"
	"maestro/internal/config"
	"maestro/internal/domain"
	"maestro/internal/executor/fake"
	FakeDB "maestro/internal/repository/fakeDB"
	"testing"
	"time"
)

func TestBuildVolumes(t *testing.T) {
	tests := []struct {
		name           string
		baseDir        string
		scriptsDir     string
		jobID          string
		expectedInput  string
		expectedOutput string
	}{
		{
			name:           "monta paths corretamente",
			baseDir:        "/app",
			scriptsDir:     "scripts/football",
			jobID:          "abc-123",
			expectedInput:  "/app/scripts/football/input-abc-123.csv:/app/input-abc-123.csv",
			expectedOutput: "/app/scripts/football/output-abc-123:/app/output-abc-123",
		},
		{
			name:           "funciona com outro tipo de job",
			baseDir:        "/app",
			scriptsDir:     "scripts/volleyball",
			jobID:          "xyz-789",
			expectedInput:  "/app/scripts/volleyball/input-xyz-789.csv:/app/input-xyz-789.csv",
			expectedOutput: "/app/scripts/volleyball/output-xyz-789:/app/output-xyz-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputFile, outputDir := buildVolumes(tt.baseDir, tt.scriptsDir, tt.jobID)
			if inputFile != tt.expectedInput {
				t.Errorf("inputFile = %q, esperado %q", inputFile, tt.expectedInput)
			}
			if outputDir != tt.expectedOutput {
				t.Errorf("outputDir = %q, esperado %q", outputDir, tt.expectedOutput)
			}
		})
	}
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
		result := domain.Job{
			ID:         "id",
			IssueKey:   "issueKey",
			TenantName: "tenant",
			Type:       "Type",
			Status:     domain.StatusSuccess,
			InputFile:  "inputFile",
			ItemCount:  10,
			CreatedAt:  now,
			FinishedAt: now.Add(10 * time.Minute),
		}

		orch := JobOrchestrator{
			DB:       db,
			Executor: exe,
			BaseDir:  "baseDir",
		}
		jt := config.JobType{
			JobType: "jobType",
		}
		resultOut, err := orch.ExecuteJob(context.Background(), result, jt)
		if err != nil {
			t.Fatalf("Erro ao testar ExecuteJob: %v", err)
		}
		switch {
		case resultOut.Status != "Success":
			t.Fatalf("Erro ao setar o resultado da automação como sucesso : Recebido: %v", resultOut.Status)
		}
	})
}
