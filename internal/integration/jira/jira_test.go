package jira

import (
	"maestro/internal/config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupFakeJira(t *testing.T) (*httptest.Server, *JiraIntegration) {
	fakeJira := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.Path, "/user/search"):
				w.Write([]byte(`[{"accountId":"abc123"}]`))
			case strings.Contains(r.URL.Path, "/comment"):
				w.WriteHeader(http.StatusCreated)
			case strings.Contains(r.URL.Path, "/assignee"):
				w.WriteHeader((http.StatusNoContent))
			case strings.Contains(r.URL.Path, "/transitions") && r.Method == "POST":
				w.WriteHeader(http.StatusNoContent)
			case strings.Contains(r.URL.Path, "/transitions") && r.Method == "GET":
				w.Write([]byte(`{"transitions":[
                {"id": "1", "to": {"name": "indeterminate"}},
                {"id": "2", "to": {"name": "done"}}
				]}`))
			}
		}))
	t.Cleanup(fakeJira.Close)
	client := &JiraIntegration{
		UserName:   "test@test.com",
		Token:      "token",
		TenantName: "tenant-test",
		BaseUrl:    fakeJira.URL,
		Client:     &http.Client{},
	}
	return fakeJira, client
}

func TestSearchUserQuery(t *testing.T) {
	t.Run("usuário encontrado", func(t *testing.T) {
		_, client := setupFakeJira(t)

		accountId, err := client.SearchUserQuery("test@test.com")
		if err != nil {
			t.Fatalf("Erro ao buscar cliente. Erro: %v", err)
		}
		if accountId != "abc123" {
			t.Errorf("Esperado: abc123 - Recebido: %v", accountId)
		}
	})

	t.Run("usuário não encontrado", func(t *testing.T) {
		fakeJira := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`[]`))
		}))
		t.Cleanup(func() { fakeJira.Close() })

		client := &JiraIntegration{
			BaseUrl: fakeJira.URL,
			Client:  &http.Client{},
		}

		_, err := client.SearchUserQuery("naoexiste@test.com")
		if err == nil {
			t.Fatal("esperava erro, got nil")
		}
	})
}

func TestAssignUser(t *testing.T) {
	t.Run("usuário atrelado", func(t *testing.T) {
		_, client := setupFakeJira(t)
		ok, err := client.AssignUser("MAE-1", "abc123")
		if err != nil {
			t.Fatalf("Erro ao atrelar usuario. Erro: %v", err)
		}
		if !ok {
			t.Fatal("Esperava true. Recebi Falso")
		}
	})
	t.Run("Assign Falho", func(t *testing.T) {
		fakeJira := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		t.Cleanup(func() { fakeJira.Close() })

		client := &JiraIntegration{
			BaseUrl: fakeJira.URL,
			Client:  &http.Client{},
		}
		ok, err := client.AssignUser("MAE-1", "abc123")

		if err != nil {
			t.Fatalf("esperava sem erro, recebido: %v", err)
		}
		if ok {
			t.Error("Esperava falso, recebi True")
		}
	})
}

func TestDoTransition(t *testing.T) {
	t.Run("Transição feita", func(t *testing.T) {
		_, client := setupFakeJira(t)
		ok, err := client.DoTransition("MAE-1", "11")
		if err != nil {
			t.Fatalf("Erro ao realizar transição. Erro %v", err)
		}
		if !ok {
			t.Fatal("Esperava true. Recebi Falso")
		}
	})

	t.Run("Transição não feita", func(t *testing.T) {
		fakeJira := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		t.Cleanup(func() { fakeJira.Close() })
		client := &JiraIntegration{
			BaseUrl: fakeJira.URL,
			Client:  &http.Client{},
		}
		ok, err := client.DoTransition("MAE-1", "11")
		if err != nil {
			t.Fatalf("Esperava sem erro, recebido: %v", err)
		}
		if ok {
			t.Error("Esperava falso, recebi True")
		}
	})

}

func TestComment(t *testing.T) {
	t.Run("Comentário feito", func(t *testing.T) {
		_, client := setupFakeJira(t)
		err := client.Comment("MAE-1", "Comentário Texto")
		if err != nil {
			t.Fatalf("Erro ao comentar. Erro %v", err)
		}
	})
	t.Run("Comentário não feito", func(t *testing.T) {
		fakeJira := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		t.Cleanup(func() { fakeJira.Close() })

		client := &JiraIntegration{
			BaseUrl: fakeJira.URL,
			Client:  &http.Client{},
		}
		err := client.Comment("MAE-1", "Comentario Texto")
		if err == nil {
			t.Fatalf("Esperava erro, recebi nil")
		}
	})
}

func TestGetIssueTransitions(t *testing.T) {
	t.Run("Get Transition Feitor", func(t *testing.T) {
		_, client := setupFakeJira(t)
		transitions, err := client.GetIssueTransitions("MAE-1")
		if err != nil {
			t.Fatalf("Não esperava erro. Recebi %v", err)
		}
		if len(transitions.Transitions) != 2 {
			t.Fatalf("Esperava len=2. Recebi len=%v", len(transitions.Transitions))
		}

		if transitions.Transitions[0].ID != "1" {
			t.Errorf("esperava id 1, got: %v", transitions.Transitions[0].ID)
		}
		if transitions.Transitions[0].To.Name != "indeterminate" {
			t.Errorf("esperava indeterminate, got: %v", transitions.Transitions[0].To.Name)
		}
	})

	t.Run("issue não existe", func(t *testing.T) {
		fakeJira := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errorMessages":["Issue does not exist"]}`))
		}))
		t.Cleanup(func() { fakeJira.Close() })

		client := &JiraIntegration{
			BaseUrl: fakeJira.URL,
			Client:  &http.Client{},
		}

		_, err := client.GetIssueTransitions("MAE-1")
		if err != nil {
			t.Fatalf("Não esperava erro. Recebi =%v", err)
		}
	})
}

func TestNewJiraApp(t *testing.T) {
	t.Run("cfg vazio", func(t *testing.T) {
		config := config.Config{}
		config.Jira.Token = ""

		_, err := NewJiraApp(&config)
		if err == nil {
			t.Fatal("esperava erro, recebeu nil")
		}
	})

	t.Run("cfg ok", func(t *testing.T) {
		cfg := config.Config{}
		cfg.Jira.Token = "token"
		cfg.Jira.UserName = "username"
		cfg.Jira.TenantName = "tenant-test"

		app, err := NewJiraApp(&cfg)
		if err != nil {
			t.Fatalf("esperava nil, recebi erro: %v", err)
		}
		if app.UserName != "username" {
			t.Fatalf("Esperava username. Recebido: %v", app.UserName)
		}
		if app.TenantName != "tenant-test" {
			t.Fatalf("Esperava tenant-test. Recebido: %v", app.TenantName)
		}

		if app.Token != "token" {
			t.Fatalf("Esperava token. Recebido: %v", app.Token)
		}
	})
}
