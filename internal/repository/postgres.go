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
		created_at TIMESTAMPTZ NOT NULL
		);

		CREATE TABLE IF NOT EXISTS tasks (
		id UUID PRIMARY KEY,
		job_id UUID NOT NULL REFERENCES jobs(id),
		status TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL		
		)	
	`)

	return err
}

func (p *Postgres) CreateJob(ctx context.Context, job domain.Job) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO jobs (id,issue_key,tenant_name,status,created_at)
		VALUES ($1,$2,$3,$4,$5)
	`, job.ID, job.IssueKey, job.TentantName, job.Status, job.CreatedAt)
	return err
}

func (p *Postgres) CreateTask(ctx context.Context, task domain.Task) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO tasks (id,job_id,status,created_at)
		VALUES ($1,$2,$3,$4)
	`, task.ID, task.JobID, task.Status, task.CreatedAt)
	return err
}
