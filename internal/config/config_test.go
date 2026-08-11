package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	dir := t.TempDir()
	prevDir, _ := os.Getwd()
	defer os.Chdir(prevDir)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// .env must exist because LoadConfig chama godotenv.Load() e faz log.Fatal() em erro
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(""), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfgYAML := `jira:
  username: test@test.com
  token: token
  tenant_name: tenant-test
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfgYAML), 0644); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("esperava sucesso, recebeu erro: %v", err)
	}

	if cfg.Jira.UserName != "test@test.com" {
		t.Errorf("esperava Jira.UserName=%q, recebeu=%q", "test@test.com", cfg.Jira.UserName)
	}
	if cfg.Jira.Token != "token" {
		t.Errorf("esperava Jira.Token=%q, recebeu=%q", "token", cfg.Jira.Token)
	}
	if cfg.Jira.TenantName != "tenant-test" {
		t.Errorf("esperava Jira.TenantName=%q, recebeu=%q", "tenant-test", cfg.Jira.TenantName)
	}
}

func TestLoadConfig_MissingConfigFile(t *testing.T) {
	dir := t.TempDir()
	prevDir, _ := os.Getwd()
	defer os.Chdir(prevDir)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(""), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("esperava erro por falta de config.yaml, recebeu nil")
	}
}

func TestLoadJobTypes(t *testing.T) {
	t.Run("Carregamento Ok com happy path", func(t *testing.T) {
		dir, err := testDataDir(t)
		if err != nil {
			t.Fatalf("Falha ao setar diretório de pasta: %v", err)
		}
		types, err := LoadJobTypes(dir)
		if err != nil {
			t.Fatalf("Falha ao carregar arquivos: %v", err)
		}
		if len(types) == 0 {
			t.Fatalf("Falha ao carregar jobs. Lenght 0")
		}
		football, ok := types["football"]
		if !ok {
			t.Fatal("Não carregado tipo 'football'")
		}
		if football.Container.ImageName != "football-rpa" {
			t.Errorf("Image name esperado era football-rpa: Recebido %v", football.Container.ImageName)
		}
		if football.JiraFields.AnswerCommentTemplate != "Modelo de resposta" {
			t.Errorf("Falha ao ler template de comentário")
		}

	})
}

func testDataDir(_ *testing.T) (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("erro ao receber caller")
	}
	return filepath.Join(filepath.Dir(file), "testdata", "jobTypes"), nil

}
