package fake

import (
	"context"
	"maestro/internal/executor"
)

var _ executor.Executor = Fake{}

type Fake struct {
	StdOut []byte
	StdErr []byte
	Err    error
}

func (f Fake) Execute(ctx context.Context, name string, args []string) ([]byte, []byte, error) {
	return f.StdOut, f.StdErr, f.Err
}
