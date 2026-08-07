package FakeDB

import (
	"context"
	"maestro/internal/domain"
	"maestro/internal/repository"
)

var _ repository.JobRepository = FakeJobDB{}

type FakeJobDB struct {
	Job     domain.Job
	Context context.Context
	Error   error
}

func (f FakeJobDB) CreateJob(ctx context.Context, job domain.Job) error {
	return f.Error
}

func (f FakeJobDB) SetJobPending(ctx context.Context, job domain.Job) error {
	return f.Error
}

func (f FakeJobDB) FinishJob(ctx context.Context, job domain.Job) error {
	return f.Error
}
