package docker

import (
	"bytes"
	"context"
	"os/exec"
)

type Docker struct{}

func (Docker) Execute(ctx context.Context, name string, args []string) ([]byte, string, error) {
	var stderr, stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()

	return stderr.Bytes(), stdout.String(), err
}
