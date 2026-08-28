package main

import (
	"log"
	"maestro/internal/config"
	"maestro/internal/executor/fake"
	"maestro/internal/handlers"
	"maestro/internal/integration/jira"
	"maestro/internal/jobResolver"
	"maestro/internal/orchestrator"
	FakeDB "maestro/internal/repository/fakeDB"
	"net/http"
	"os"
	"path/filepath"
)

var ready bool

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
		return
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return
	}
	path := filepath.Join(dir, "/config/jobTypes/")
	jobs, err := config.LoadJobTypes(path)
	// Cenario DemoMode para teste de job template valido
	if cfg.DemoMode {
		db := FakeDB.FakeJobDB{}
		executor := fake.Fake{}
		jiraReader := jira.FakeJiraReader{
			Issue: jira.JiraIssue{
				Fields: jira.JiraIssueFields{
					Sistema:        jira.JiraCustomField{Value: "Sistema"},
					Necessidade:    jira.JiraCustomField{Value: "Necessidade"},
					JiraAttachment: []jira.JiraAttachment{{Filename: "NomeDoAnexo.csv"}},
				},
			},
		}
		orch := orchestrator.NewJobOrchestrator(db, executor, "basedir", make(chan struct{}, cfg.API.ConcurrentJobs))
		jobRes := jobResolver.NewJobResolver(jiraReader, jobs, orch)
		h := handlers.NewHandler(*jobRes, db)
		mux := http.NewServeMux()
		mux.HandleFunc("/jira-webhook", h.HandleJiraWebhook)

		server := &http.Server{
			Addr:    ":8080",
			Handler: handlers.LoggingMiddleware(mux),
		}
		log.Println("API subindo")
		log.Fatal(server.ListenAndServe())
	}
}

/*
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
	//router.HandleFunc("/test-automation", api.TestAutomation)
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
*/
