package fake

import (
	"context"
	"log"
	"maestro/internal/executor"
	"time"
)

var _ executor.Executor = Fake{}

type Fake struct {
	StdOut             []byte
	StdErr             []byte
	Err                error
	JobDurationSeconds int
}

func (f Fake) Execute(ctx context.Context, name string, args []string) ([]byte, []byte, error) {
	if f.JobDurationSeconds > 0 {
		time.Sleep(time.Duration(f.JobDurationSeconds) * time.Second)
		log.Printf("Job simulado demorando %v segundos.", f.JobDurationSeconds)
	}
	return f.StdOut, f.StdErr, f.Err
}
