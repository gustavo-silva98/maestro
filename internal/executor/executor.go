package executor

import (
	"context"
)

type Executor interface {
	Execute(ctx context.Context, name string, args []string) (stderr []byte, stdout []byte, err error)
}
