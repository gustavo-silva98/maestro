package repository

import (
	"context"
	"maestro/internal/domain"
)

// Refatorar essa interface urgente

type JobRepository interface {
	CreateJob(ctx context.Context, job domain.Job) error
	SetJobPending(ctx context.Context, job domain.Job) error
	FinishJob(ctx context.Context, job domain.Job) error
	SetJobFailed(ctx context.Context, job domain.Job) error
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task domain.Task) error
}

type JobTaskRepo interface {
	TaskRepository
	JobRepository
}
