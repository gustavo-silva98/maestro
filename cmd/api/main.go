package main

import (
	"context"
	"log"
	"maestro/internal/config"
	"maestro/internal/handlers"
	"maestro/internal/integration/jira"
	"maestro/internal/repository"
	"net/http"
)

var ready bool

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	/*
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			inputFile := fmt.Sprintf("%s/scripts/football/input.csv:/app/input.csv", cwd)
			outputDir := fmt.Sprintf("%s/scripts/football/output:/app/Prints", cwd)

			cmd := exec.Command("docker", "run", "--rm", "-v", inputFile, "-v", outputDir, "football-rpa")
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Erro ao rodar container %v\nOutput: %s", err, string(out))
			}

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
	}
	if err := db.CreateTables(ctx); err != nil {
		log.Printf("erro ao criar tabelas %v", err)
	}

	jiraApi, err := jira.NewJiraApp(&cfg)
	if err != nil {
		log.Println(err)
	}
	api := handlers.Backend{
		Port:    cfg.API.Port,
		ID:      1,
		Ready:   ready,
		JiraApi: &jiraApi,
		Config:  &cfg,
		DB:      db,
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
