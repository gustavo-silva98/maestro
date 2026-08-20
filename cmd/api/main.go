package main

import (
	"context"
	"log"
	"maestro/internal/config"
	"maestro/internal/handlers"
	"maestro/internal/integration/jira"
	"maestro/internal/orchestrator"
	"maestro/internal/repository"
	"net/http"
)

var ready bool

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	ctx := context.Background()

	db, err := repository.NewPostgres(ctx, cfg.DB.URL)
	if err != nil {
		log.Println(err)
		return
	}
	if err := db.CreateTables(ctx); err != nil {
		log.Printf("erro ao criar tabelas %v", err)
	}

	jiraApi, err := jira.NewJiraApp(&cfg)
	if err != nil {
		log.Println(err)
		return
	}
	orch := orchestrator.JobOrchestrator{
		DB: db,
	}
	api := handlers.Backend{
		Port:         cfg.API.Port,
		ID:           1,
		Ready:        ready,
		JiraApi:      &jiraApi,
		Config:       &cfg,
		DB:           db,
		Orchestrator: &orch,
	}
	router := http.NewServeMux()
	router.HandleFunc("/ready", api.ReadyEndpoint)
	router.HandleFunc("/test-automation", api.TestAutomation)
	router.HandleFunc("/get-jobs", api.GetJobs)
	router.HandleFunc("/delete-table", api.DeleteTables)
	router.HandleFunc("/get-saved-minute", api.GetSavedMinutes)
	router.HandleFunc("/status-counts", api.GetJobStatusCounts)

	server := &http.Server{
		Addr:    ":" + api.Port,
		Handler: handlers.EnableCORS(router),
	}
	log.Printf("Server interno up : ID %v", api.ID)
	log.Fatal(server.ListenAndServe())
}
