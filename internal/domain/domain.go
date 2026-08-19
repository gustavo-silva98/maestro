package domain

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusPending        Status = "Pending"
	StatusRunning        Status = "Running"
	StatusSuccess        Status = "Success"
	StatusFailed         Status = "Failed"
	StatusPartialSuccess Status = "PartialSuccess"
)

type Job struct {
	ID         string
	IssueKey   string
	TenantName string
	Type       string
	Status     Status
	InputFile  string
	ItemCount  int
	CreatedAt  time.Time
	FinishedAt time.Time
}

type Task struct {
	ID           string
	JobID        string
	Image        string
	Status       Status
	ExitCode     int
	Payload      json.RawMessage
	Logs         []byte
	TaskType     string
	SavedMinutes int
	CreatedAt    time.Time
	FinishedAt   time.Time
}
