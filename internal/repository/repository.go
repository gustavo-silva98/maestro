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
	//UpdateTask(ctx context.Context, task domain.Task) error
	//GetPending(ctx context.Context) ([]domain.Job, error)
}
