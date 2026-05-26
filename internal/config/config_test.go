package config

import (
	"os"
	"path/filepath"
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
