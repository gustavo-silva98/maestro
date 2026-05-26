package main

import (
	"log"
	"maestro/internal/config"
	"maestro/internal/handlers"
	"maestro/internal/integration/jira"
	"net/http"
)

var ready bool

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	jiraApi, err := jira.NewJiraApp(&cfg)
	api := handlers.Backend{
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
