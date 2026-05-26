package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"maestro/internal/config"
	"maestro/internal/integration/jira"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
