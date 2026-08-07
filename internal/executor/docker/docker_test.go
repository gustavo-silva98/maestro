package docker

import (
	"context"
	"strings"
	"testing"
)

func TestExecute(t *testing.T) {
	t.Run("stderr e stdout ok sem erros", func(t *testing.T) {
		docker := Docker{}
		_, stdout, err := docker.Execute(context.Background(), "echo", []string{"hello"})
		if err != nil {
			t.Fatalf("Erro ao executar echo: %v", err)
		}
		if strings.TrimSpace(string(stdout)) != "hello" {
			t.Fatalf("Valor esperado: hello - Valor Recebido: %v", stdout)
		}
	})
}
