package main

import (
	"bytes"
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
var stdout bytes.Buffer
var stderr bytes.Buffer

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	/*
		jiraApi, err := jira.NewJiraApp(&cfg)
		oi, err := jiraApi.GetIssue("MAE-19")
		if err != nil {
			log.Println(err)
		}
		attachment, err := jiraApi.GetAttachmentContent(oi.Fields.JiraAttachment[0].ID)
		if err != nil {
			log.Println(err)
		}
		os.WriteFile("inputReceived.csv", attachment, 0644)
		if err := jiraApi.AddAttachment("MAE-19", "inputReceived.csv"); err != nil {
			log.Printf("erro ao colocar anexo %v", err)
		}
		comment := fmt.Sprintf("Csv Recebido \n\n!%s!\n\n", "inputReceived.csv")
		if err := jiraApi.Comment("MAE-19", comment); err != nil {
			log.Printf("erro ao comentar anexo %v", err)
		}
	*/
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
	orch := orchestrator.FootballOrchestrator{
		Config: &cfg,
		DB:     db,
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
