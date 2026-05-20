package main

import (
	"log"
	"maestro/internal/config"
	"maestro/internal/integration/jira"
	"net/http"

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
	var webhook_body jira.JiraWebhookBody
	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(webhook_body)
	issue := webhook_body.Issue.Key
	if issue != "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	transitions, err := api.JiraApi.GetIssueTransitions(issue)
	if err != nil {
		log.Fatal(err)
	}
	for _, val := range transitions.Transitions {
		if val.To.StatusCategory.Key == api.Config.Jira.StatusAllowed.InitialStatus {
			transitionProcess, err := api.JiraApi.DoTransition(issue, val.ID)
			if err != nil {
				log.Fatal(err)
			}
			if transitionProcess {
				api.JiraApi.Comment(issue, "Transição feita: Em progresso.")
			}
			break
		}
	}
	transitions, err = api.JiraApi.GetIssueTransitions(issue)
	if err != nil {
		log.Fatal(err)
	}
	for _, val := range transitions.Transitions {
		if val.To.StatusCategory.Key == api.Config.Jira.StatusAllowed.FinalStatus {
			transitionProcess, err := api.JiraApi.DoTransition(issue, val.ID)
			if err != nil {
				log.Fatalf("ERRO AO TRANSICIONAR DONE: %v", err)
			}
			if transitionProcess {
				api.JiraApi.Comment(issue, "Transição feita: Done.")
			}
			break
		}
	}
	w.WriteHeader(http.StatusOK)
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
