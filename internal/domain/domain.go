package domain

import "time"

type Job struct {
	ID          string
	IssueKey    string
	TentantName string
	Status      string
	CreatedAt   time.Time
	FinishedAt  time.Time
	JobType     string
}

type Task struct {
	ID         string
	JobID      string
	Status     string
	CreatedAt  time.Time
	FinishedAt time.Time
	TaskType   string
}
