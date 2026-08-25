package memory

import (
	"context"
	"errors"
	"maestro/internal/domain"
	"maestro/internal/repository"
	"sync"
	"testing"
	"time"
)

func newTestRepo() *Memory {
	return NewMemoryRepo(
		&sync.Mutex{},
		make(map[string]domain.Job),
		make(map[string]domain.Task),
	)
}

func TestCreateJob(t *testing.T) {
	repo := newTestRepo()
	job := domain.Job{ID: "job-1", IssueKey: "MAE-1"}

	if err := repo.CreateJob(context.Background(), job); err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}

	if repo.jobs["job-1"] != job {
		t.Errorf("job não foi salvo corretamente")
	}

	err := repo.CreateJob(context.Background(), job)
	if err == nil {
		t.Fatal("esperava erro de duplicidade")
	}

	if err.Error() != string(repository.DuplicateError) {
		t.Errorf("erro = %q, esperado %q", err.Error(), repository.DuplicateError)
	}
}

func TestSetJobPending(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	err := repo.SetJobPending(ctx, domain.Job{ID: "missing"})
	if err == nil {
		t.Fatal("esperava erro para job inexistente")
	}

	job := domain.Job{
		ID:     "job-1",
		Status: domain.StatusRunning,
	}

	repo.jobs[job.ID] = job

	if err := repo.SetJobPending(ctx, job); err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}

	if repo.jobs[job.ID].Status != domain.StatusPending {
		t.Errorf("status = %q, esperado %q",
			repo.jobs[job.ID].Status, domain.StatusPending)
	}

}

func TestFinishJob(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()
	finishedAt := time.Now()

	job := domain.Job{
		ID:     "job-1",
		Status: domain.StatusPending,
	}
	repo.jobs[job.ID] = job

	updatedJob := domain.Job{
		ID:         job.ID,
		Status:     domain.StatusSuccess,
		FinishedAt: finishedAt,
	}

	if err := repo.FinishJob(ctx, updatedJob); err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}

	savedJob := repo.jobs[job.ID]

	if savedJob.Status != domain.StatusSuccess {
		t.Errorf("status = %q, esperado %q",
			savedJob.Status, domain.StatusSuccess)
	}

	if !savedJob.FinishedAt.Equal(finishedAt) {
		t.Errorf("FinishedAt não foi atualizado")
	}

	if err := repo.FinishJob(ctx, domain.Job{ID: "missing"}); err == nil {
		t.Fatal("esperava erro para job inexistente")
	}
}

func TestSetJobFailed(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()
	finishedAt := time.Now()

	repo.jobs["job-1"] = domain.Job{
		ID: "job-1",
	}

	job := domain.Job{
		ID:         "job-1",
		Status:     domain.StatusFailed,
		FinishedAt: finishedAt,
	}

	if err := repo.SetJobFailed(ctx, job); err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}

	savedJob := repo.jobs["job-1"]

	if savedJob.Status != domain.StatusFailed {
		t.Errorf("status = %q, esperado %q",
			savedJob.Status, domain.StatusFailed)
	}

	if !savedJob.FinishedAt.Equal(finishedAt) {
		t.Errorf("FinishedAt não foi atualizado")
	}

	if err := repo.SetJobFailed(ctx, domain.Job{ID: "missing"}); err == nil {
		t.Fatal("esperava erro para job inexistente")
	}
}

func TestCreateTask(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	task := domain.Task{
		ID:     "task-1",
		JobID:  "job-1",
		Status: domain.StatusSuccess,
	}

	if err := repo.CreateTask(ctx, task); err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}

	if repo.tasks[task.ID].ID != task.ID {
		t.Errorf("task não foi salva corretamente")
	}

	err := repo.CreateTask(ctx, task)
	if err == nil {
		t.Fatal("esperava erro de duplicidade")
	}

	if !errors.Is(err, errors.New(string(repository.DuplicateError))) &&
		err.Error() != string(repository.DuplicateError) {
		t.Errorf("erro inesperado: %v", err)
	}
}
