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
		INSERT INTO jobs (id,issue_key,tenant_name,status,created_at, finished_at,job_type,saved_minutes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, job.ID, job.IssueKey, job.TentantName, job.Status, job.CreatedAt, job.CreatedAt, job.JobType, job.SavedMinutes)
	return err
}

func (p *Postgres) CreateTask(ctx context.Context, task domain.Task) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO tasks (id,job_id,status,created_at,finished_at, task_type)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, task.ID, task.JobID, task.Status, task.CreatedAt, task.CreatedAt, task.TaskType)
	return err
}

func (p *Postgres) SetJobPending(ctx context.Context, job domain.Job) error {
	_, err := p.pool.Exec(ctx, `
		UPDATE jobs SET status = $1 WHERE id = $2
	`, "Running", job.ID)
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

func (p *Postgres) GetJobs(ctx context.Context, numberRows int) ([]domain.Job, error) {
	rows, err := p.pool.Query(ctx, `
	SELECT * FROM jobs ORDER BY created_at DESC LIMIT $1
	`, numberRows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := make([]domain.Job, 0)

	for rows.Next() {
		var job domain.Job
		if err := rows.Scan(
			&job.ID,
			&job.IssueKey,
			&job.TentantName,
			&job.Status,
			&job.CreatedAt,
			&job.FinishedAt,
			&job.JobType,
			&job.SavedMinutes,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (p *Postgres) ClearTables(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM jobs`)
	if err != nil {
		return err
	}
	_, err = p.pool.Exec(ctx, `DELETE FROM tasks`)
	if err != nil {
		return err
	}
	return nil
}

func (p *Postgres) GetSavedMinutes(ctx context.Context) (float64, error) {
	var totalMinutes float64

	err := p.pool.QueryRow(ctx, "SELECT COALESCE(SUM(saved_minutes), 0) FROM jobs WHERE created_at >= NOW() - INTERVAL '24 hours'").Scan(&totalMinutes)
	if err != nil {
		return 0, err
	}
	return totalMinutes, nil

}

func (p *Postgres) GetJobStatusCountsLast24h(ctx context.Context) (map[string]int, error) {
	rows, err := p.pool.Query(ctx, `
        SELECT status, COUNT(*) as count
        FROM jobs
        WHERE created_at >= NOW() - INTERVAL '24 hours'
        GROUP BY status
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statusCounts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		statusCounts[status] = count
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return statusCounts, nil
}

func (p *Postgres) GetTotalJobsCount(ctx context.Context) (int, error) {
	var total int
	err := p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM jobs WHERE created_at >= NOW() - INTERVAL '24 hours'").Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}
