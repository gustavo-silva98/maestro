package repository

import (
	"context"
	"fmt"
	"maestro/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, connString string) (*Postgres, error) {

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres não respondeu %w", err)
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) CreateTables(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS jobs (
		id UUID PRIMARY KEY,
		issue_key TEXT NOT NULL,
		tenant_name TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		finished_at TIMESTAMPTZ NOT NULL,
		job_type TEXT NOT NULL,
		saved_minutes NUMERIC(10,2) NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS tasks (
		id UUID PRIMARY KEY,
		job_id UUID NOT NULL REFERENCES jobs(id),
		status TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		finished_at TIMESTAMPTZ NOT NULL,
		task_type TEXT NOT NULL	
		)	
	`)

	return err
}

func (p *Postgres) CreateJob(ctx context.Context, job domain.Job) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO jobs (id,issue_key,tenant_name,type,status,input_file,created_at, finished_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, job.ID, job.IssueKey, job.TenantName, job.Type, job.Status, job.InputFile, job.CreatedAt, job.FinishedAt)
	return err
}

func (p *Postgres) CreateTask(ctx context.Context, task domain.Task) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO tasks (id,job_id,image,status,exit_code,payload,logs,task_type,saved_minutes,created_at,finished_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, task.ID, task.JobID, task.Image, task.Status, task.ExitCode, task.Payload,
		task.Logs, task.TaskType, task.SavedMinutes, task.CreatedAt, task.FinishedAt)
	return err
}

func (p *Postgres) SetJobPending(ctx context.Context, job domain.Job) error {
	_, err := p.pool.Exec(ctx, `
		UPDATE jobs SET status = $1 WHERE id = $2
	`, "Running", job.ID)
	return err
}

func (p *Postgres) SetJobFailed(ctx context.Context, job domain.Job) error {
	_, err := p.pool.Exec(ctx, `
	UPDATE jobs SET status = $1 WHERE id = $2`, domain.StatusFailed, job.ID)
	return err
}

func (p *Postgres) SetJobRunning(ctx context.Context, job domain.Job) error {
	_, err := p.pool.Exec(ctx, `
		UPDATE jobs SET status = $1 WHERE id = $2
	`, "Running", job.ID)
	return err
}

func (p *Postgres) FinishJob(ctx context.Context, job domain.Job) error {
	_, err := p.pool.Exec(ctx, `
	UPDATE jobs SET 
		status = $1, finished_at = $2
	WHERE id = $3
	`, "Success", job.FinishedAt, job.ID)

	return err
}

func (p *Postgres) SetTaskPending(ctx context.Context, task domain.Task) error {
	_, err := p.pool.Exec(ctx, `
		UPDATE tasks SET status = $1 WHERE id = $2
	`, "Running", task.ID)
	return err
}

func (p *Postgres) FinishTask(ctx context.Context, task domain.Task) error {
	_, err := p.pool.Exec(ctx, `
	UPDATE tasks SET 
		status = $1, finished_at = $2
	WHERE id = $3
	`, "Success", task.FinishedAt, task.ID)

	return err
}
