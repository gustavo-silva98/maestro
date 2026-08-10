package docker

import (
	"context"
	"strings"
	"testing"
)

func TestExecute(t *testing.T) {
	t.Run("stderr e stdout ok sem erros", func(t *testing.T) {
		d := Docker{}
		stdout, _, err := d.Execute(context.Background(), "echo", []string{"hello"})
		if err != nil {
			t.Fatalf("Erro ao executar echo: %v", err)
		}
		if strings.TrimSpace(string(stdout)) != "hello" {
			t.Fatalf("Valor esperado: hello - Valor Recebido: %v", stdout)
		}
	})
	t.Run("stderr com erros", func(t *testing.T) {
		d := Docker{}
		_, stderr, err := d.Execute(context.Background(), "bash", []string{"-c", "echo erro >&2; exit 42"})
		if err == nil {
			t.Error("Falha ao gerar erro no teste")
		}
		if strings.TrimSpace(string(stderr)) != "erro" {
			t.Errorf("Falha ao ler stderr. Esperado: erro - Recebido: %s", stderr)
		}
	})
}
