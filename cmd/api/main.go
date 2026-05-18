package main

import (
	"fmt"
	"log"
	"maestro/internal/integration/jira"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	api, err := jira.NewJiraApp()
	if err != nil {
		log.Fatal(err)
	}

	resp, _ := api.GetIssue("MAE-1")
	fmt.Println(resp)
}
