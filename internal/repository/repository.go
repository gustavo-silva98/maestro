package repository

import (
	"context"
	"maestro/internal/domain"
)

type MaestroRepository interface {
	CreateJob(ctx context.Context, job domain.Job) error
	//UpdateJob(ctx context.Context, job domain.Job) error
	CreateTask(ctx context.Context, task domain.Task) error
	//UpdateTask(ctx context.Context, task domain.Task) error
	//GetPending(ctx context.Context) ([]domain.Job, error)
}
