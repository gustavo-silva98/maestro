package main

import (
	"log"
	"maestro/internal/config"
	"maestro/internal/domain"
	"maestro/internal/executor/docker"
	"maestro/internal/executor/fake"
	"maestro/internal/handlers"
	"maestro/internal/integration/jira"
	"maestro/internal/jobResolver"
	"maestro/internal/orchestrator"
	FakeDB "maestro/internal/repository/fakeDB"
	"maestro/internal/repository/memory"
	"net/http"
	"os"
	"path/filepath"
	"sync"
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
	if cfg.API.ConcurrentJobs <= 0 {
		log.Fatalf("Concurrent Jobs deve ser maior que 0: %v", cfg.API.ConcurrentJobs)
		return
	}
	path := filepath.Join(dir, "config", "jobTypes")
	jobs, err := config.LoadJobTypes(path)
	// Cenario DemoMode para teste de job template valido
	if cfg.Mode == "test" {
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
		orch := orchestrator.NewJobOrchestrator(db, executor, dir, make(chan struct{}, cfg.API.ConcurrentJobs))
		jobRes := jobResolver.NewJobResolver(jiraReader, jobs, orch)
		h := handlers.NewHandler(*jobRes, db)
		mux := http.NewServeMux()
		mux.HandleFunc("/jira-webhook", h.HandleJiraWebhook)

		server := &http.Server{
			Addr:    ":8080",
			Handler: handlers.LoggingMiddleware(mux),
		}
		log.Println("API Up em Test")
		log.Fatal(server.ListenAndServe())
	}

	if cfg.Mode == "load" {
		db := memory.NewMemoryRepo(
			&sync.Mutex{},
			make(map[string]domain.Job),
			make(map[string]domain.Task),
		)
		exe := docker.Docker{}
		inputPath := filepath.Join(dir, "scripts", "football", "input.csv")

		inputBytes, err := os.ReadFile(inputPath)
		if err != nil {
			log.Fatal(err)
		}
		jiraReader := jira.FakeJiraReader{
			Issue: jira.JiraIssue{
				Fields: jira.JiraIssueFields{
					Sistema:        jira.JiraCustomField{Value: "Transfermarket"},
					Necessidade:    jira.JiraCustomField{Value: "Pesquisa"},
					JiraAttachment: []jira.JiraAttachment{{ID: "FakeID", Filename: "Input.csv"}},
				},
			},
			InputBytes: inputBytes,
		}

		sem := make(chan struct{}, cfg.API.ConcurrentJobs)
		orch := orchestrator.NewJobOrchestrator(db, exe, dir, sem)
		jobResolver := jobResolver.NewJobResolver(jiraReader, jobs, orch)
		handler := handlers.NewHandler(*jobResolver, db)
		queryHandler := handlers.NewQueryHandler(db)

		mux := http.NewServeMux()
		mux.HandleFunc("/jira-webhook", handler.HandleJiraWebhook)
		mux.HandleFunc("/jobs", queryHandler.ListJobs)

		server := http.Server{
			Addr:    ":" + cfg.API.Port,
			Handler: handlers.LoggingMiddleware(mux),
		}
		log.Println("API subindo em Load Mode")
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
