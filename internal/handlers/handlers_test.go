package handlers

import (
	"bytes"
	"errors"
	"io"
	"log"
	"maestro/internal/config"
	"maestro/internal/integration/jira"
	"maestro/internal/jobResolver"
	FakeDB "maestro/internal/repository/fakeDB"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type errorReadCloser struct{}

func (errorReadCloser) Read([]byte) (int, error) {
	return 0, errors.New("erro fabricado ao ler body")
}

func (errorReadCloser) Close() error {
	return nil
}

func setupFakeJira(t *testing.T) *httptest.Server {
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
	return fakeJira
}

func TestReadyEndpoint(t *testing.T) {
	api := &Backend{ID: 42, Ready: false}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ready", nil)
	api.ReadyEndpoint(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperava 200, recebeu %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)

	if !bytes.Contains(body, []byte(`"ready":true`)) {
		t.Fatalf("esperava ready true, body: %s", string(body))
	}
}

func TestHandleJiraWebhook(t *testing.T) {
	t.Run("Erro ao criar Job", func(t *testing.T) {
		fakedb := FakeDB.FakeJobDB{
			Error: errors.New("Erro fabricado ao criar Job"),
		}
		jr := jobResolver.JobResolver{}
		h := NewHandler(jr, fakedb)
		payload := `{"issue":{"key":"MAE-456"}}`
		req := httptest.NewRequest(
			http.MethodPost,
			"/webhook",
			strings.NewReader(payload),
		)
		recorder := httptest.NewRecorder()
		h.HandleJiraWebhook(recorder, req)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("Esperado InternalServerError, recebido: %v", recorder.Code)
		}
	})
	t.Run("Erro ao parsear JSON", func(t *testing.T) {
		repo := FakeDB.FakeJobDB{}
		h := NewHandler(jobResolver.JobResolver{}, repo)

		r := httptest.NewRequest(
			http.MethodPost,
			"/webhook",
			nil,
		)
		r.Body = errorReadCloser{}

		w := httptest.NewRecorder()
		h.HandleJiraWebhook(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Esperado BadRequest, recebido: %v", w.Code)
		}
	})
	t.Run("Happy Path", func(t *testing.T) {

		repo := FakeDB.FakeJobDB{
			Error: nil,
		}

		jr := jobResolver.NewJobResolver(
			jira.FakeJiraReader{},
			map[string]config.JobType{},
			nil,
		)
		h := NewHandler(*jr, repo)
		payload := `{"issue":{"key":"MAE-456"}}`
		r := httptest.NewRequest(
			http.MethodPost,
			"/webhook",
			strings.NewReader(payload),
		)

		w := httptest.NewRecorder()

		h.HandleJiraWebhook(w, r)
		if w.Code != http.StatusAccepted {
			t.Errorf("Status Expected: 202. Expected: %v", w.Code)
		}
	})
}

func TestLoggingMiddleware(t *testing.T) {
	var logs bytes.Buffer

	originalWriter := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
	})

	handlerCalled := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	h := LoggingMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	recorder := httptest.NewRecorder()

	h.ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Error("esperava o proximo handler ser chamado")
	}

	if recorder.Code != http.StatusNoContent {
		t.Errorf(
			"status = %d, esperado %d",
			recorder.Code,
			http.StatusNoContent,
		)
	}

	output := logs.String()

	if !strings.Contains(output, "Method: GET") {
		t.Errorf("log não contém o método HTTP: %q", output)
	}

	if !strings.Contains(output, "URI: /jobs") {
		t.Errorf("log nãoi contém a URI: %q", output)
	}

	if !strings.Contains(output, "Time:") {
		t.Errorf("log não contém o tempo de execução: %q", output)
	}
}

/*
func TestTestAutomation(t *testing.T) {
	fake := setupFakeJira(t)
	defer fake.Close()

	cfg := &config.Config{}
	cfg.Jira.WebhookSecret = "mysecret"
	cfg.Jira.UserName = "bot@ex"
	cfg.Jira.StatusAllowed.InitialStatus = "indeterminate"
	cfg.Jira.StatusAllowed.FinalStatus = "done"

	jc := &jira.JiraIntegration{
		UserName:   "bot@ex",
		Token:      "token",
		TenantName: "tenant",
		BaseUrl:    fake.URL,
		Client:     &http.Client{},
	}

	api := &Backend{JiraApi: jc, Config: cfg}

	payload := []byte(`{"issue":{"key":"MAE-1"}}`)
	computeSig := func(secret string, body []byte) string {
		h := hmac.New(sha256.New, []byte(secret))
		h.Write(body)
		return hex.EncodeToString(h.Sum(nil))
	}

	t.Run("success", func(t *testing.T) {
		sig := computeSig(cfg.Jira.WebhookSecret, payload)
		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
		req.Header.Set("X-Hub-Signature", "sha256="+sig)

		w := httptest.NewRecorder()
		api.TestAutomation(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, recebeu %d; body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("bad signature", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
		req.Header.Set("X-Hub-Signature", "sha256=deadbeef")

		w := httptest.NewRecorder()
		api.TestAutomation(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400 para assinatura inválida, recebeu %d", w.Code)
		}
	})
}
*/
