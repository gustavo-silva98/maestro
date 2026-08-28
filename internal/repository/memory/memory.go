package memory

import (
	"context"
	"fmt"
	"maestro/internal/domain"
	"maestro/internal/repository"
	"sync"
)

var _ repository.JobTaskRepo = &Memory{}
var _ repository.QueryData = &Memory{}

type Memory struct {
	mu    *sync.Mutex
	jobs  map[string]domain.Job
	tasks map[string]domain.Task
}

func NewMemoryRepo(mu *sync.Mutex, jobs map[string]domain.Job, tasks map[string]domain.Task) *Memory {
	return &Memory{
		mu:    mu,
		jobs:  jobs,
		tasks: tasks,
	}
}

func (m *Memory) CreateJob(ctx context.Context, job domain.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.elementIsPresent(job.ID) {
		return fmt.Errorf("%v", repository.DuplicateError)
	}
	m.jobs[job.ID] = job
	return nil
}

func (m *Memory) SetJobPending(ctx context.Context, job domain.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.elementIsPresent(job.ID) {
		return fmt.Errorf(string(repository.NotFoundError))
	}
	newJob := m.jobs[job.ID]
	newJob.Status = domain.StatusPending

	m.jobs[job.ID] = newJob

	return nil
}

func (m *Memory) FinishJob(ctx context.Context, job domain.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.elementIsPresent(job.ID) {
		return fmt.Errorf(string(repository.NotFoundError))
	}
	newjob := m.jobs[job.ID]
	newjob.Status = job.Status
	newjob.FinishedAt = job.FinishedAt
	m.jobs[job.ID] = newjob
	return nil
}

func (m *Memory) SetJobFailed(ctx context.Context, job domain.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.elementIsPresent(job.ID) {
		return fmt.Errorf(string(repository.NotFoundError))
	}
	newjob := m.jobs[job.ID]
	newjob.Status = job.Status
	newjob.FinishedAt = job.FinishedAt
	m.jobs[job.ID] = newjob
	return nil
}

func (m *Memory) CreateTask(ctx context.Context, task domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[task.ID]; exists {
		return fmt.Errorf("%v", repository.DuplicateError)
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *Memory) elementIsPresent(id string) bool {
	_, ok := m.jobs[id]
	if ok {
		return true
	}
	return false
}

func (m *Memory) ListJobs(ctx context.Context) ([]domain.Job, error) {
	result := make([]domain.Job, 0, len(m.jobs))
	for _, val := range m.jobs {
		result = append(result, val)
	}
	return result, nil
}
