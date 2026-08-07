package executor

import (
	"context"
)

type Executor interface {
	Execute(ctx context.Context, name string, args []string) (stdout []byte, stderr []byte, err error)
}
