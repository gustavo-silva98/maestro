package repository

import (
	"context"
	"maestro/internal/domain"
)

type MaestroRepository interface {
	CreateJob(ctx context.Context, job domain.Job) error
	SetJobPending(ctx context.Context, job domain.Job) error
	FinishJob(ctx context.Context, job domain.Job) error
	CreateTask(ctx context.Context, task domain.Task) error
	SetTaskPending(ctx context.Context, task domain.Task) error
	FinishTask(ctx context.Context, task domain.Task) error
	GetJobs(ctx context.Context, numberRows int) ([]domain.Job, error)
	SetJobRunning(ctx context.Context, job domain.Job) error
	ClearTables(ctx context.Context) error
	GetSavedMinutes(ctx context.Context) (float64, error)
	GetJobStatusCountsLast24h(ctx context.Context) (map[string]int, error)
	GetTotalJobsCount(ctx context.Context) (int, error)
	//UpdateTask(ctx context.Context, task domain.Task) error
	//GetPending(ctx context.Context) ([]domain.Job, error)
}
