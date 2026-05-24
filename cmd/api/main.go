package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"maestro/internal/config"
	"maestro/internal/integration/jira"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
)

// var testIssue string = "MAE-1"
var ready bool

type Backend struct {
	Port    string
	ID      int
	Ready   bool
	JiraApi *jira.JiraIntegration
	Config  *config.Config
}

func (api *Backend) ReadyEndpoint(w http.ResponseWriter, r *http.Request) {
	api.Ready = true
	resp := map[string]interface{}{
		"apiId": api.ID,
		"ready": api.Ready,
	}
	jsonData, err := sonic.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(jsonData)
}

func (api *Backend) TestAutomation(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	xhub := strings.Split(r.Header.Get("X-Hub-Signature"), "=")
	if len(xhub) != 2 {
		http.Error(w, "Erro ao validar signature webhook", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Falha ao ler body: %v", err), http.StatusBadRequest)
		return
	}

	if !checkHmac(api.Config.Jira.WebhookSecret, xhub[1], body) {
		log.Println("Autenticação do Webhook falhou")
		http.Error(w, "Autenticação do Webhook falhou!", http.StatusBadRequest)
		return
	} else {
		log.Println("HMAC Signature validado!")
	}

	var webhook_body jira.JiraWebhookBody
	r.Body = io.NopCloser(bytes.NewReader(body))
	if err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&webhook_body); err != nil {
		log.Printf("decode error: %v", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	log.Printf("Requisição recebida - Issue %v", webhook_body.Issue.Key)

	issue := webhook_body.Issue.Key
	if issue == "" {
		http.Error(w, "missing issue key", http.StatusBadRequest)
		return
	}
	accountId, err := api.JiraApi.SearchUserQuery(api.Config.Jira.UserName)
	if err != nil {
		http.Error(w, "missing user", http.StatusBadGateway)
		return
	}
	_, err = api.JiraApi.AssignUser(issue, accountId)
	if err != nil {
		http.Error(w, "error on assign user", http.StatusBadGateway)
		return
	}
	transitions, err := api.JiraApi.GetIssueTransitions(issue)
	if err != nil {
		http.Error(w, "Error get issueTransition", http.StatusBadGateway)
		return
	}
	for _, val := range transitions.Transitions {
		if val.To.Name == api.Config.Jira.StatusAllowed.InitialStatus {
			transitionProcess, err := api.JiraApi.DoTransition(issue, val.ID)
			if err != nil {
				http.Error(w, "Error transitioning issue In Progress", http.StatusBadGateway)
				return
			}
			if transitionProcess {
				api.JiraApi.Comment(issue, "Transição feita: Em progresso.")
			}
			break
		}
	}
	transitions, err = api.JiraApi.GetIssueTransitions(issue)
	if err != nil {
		http.Error(w, "Error get issueTransition", http.StatusBadGateway)
		return
	}
	for _, val := range transitions.Transitions {
		if val.To.Name == api.Config.Jira.StatusAllowed.FinalStatus {
			transitionProcess, err := api.JiraApi.DoTransition(issue, val.ID)
			if err != nil {
				http.Error(w, "Error transitioning issue Done", http.StatusBadGateway)
				return
			}
			if transitionProcess {
				api.JiraApi.Comment(issue, "Transição feita: Done.")
			}
			break
		}
	}
	w.WriteHeader(http.StatusOK)
}

func checkHmac(secret, received string, data []byte) bool {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(data))
	expected := hash.Sum(nil)

	receivedHex, err := hex.DecodeString(received)
	if err != nil {
		return false
	}
	return hmac.Equal(expected, []byte(receivedHex))
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	jiraApi, err := jira.NewJiraApp(&cfg)
	api := Backend{
		Port:    cfg.API.Port,
		ID:      1,
		Ready:   ready,
		JiraApi: &jiraApi,
		Config:  &cfg,
	}
	router := http.NewServeMux()
	router.HandleFunc("/ready", api.ReadyEndpoint)
	router.HandleFunc("/test-automation", api.TestAutomation)

	server := &http.Server{
		Addr:    ":" + api.Port,
		Handler: router,
	}
	log.Printf("Server interno up : ID %v", api.ID)
	log.Fatal(server.ListenAndServe())
}
