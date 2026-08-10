package docker

import (
	"bytes"
	"context"
	"maestro/internal/executor"
	"os/exec"
)

var _ executor.Executor = Docker{}

type Docker struct{}

func (Docker) Execute(ctx context.Context, name string, args []string) ([]byte, []byte, error) {
	var stderr, stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()

	return stdout.Bytes(), stderr.Bytes(), err
}
