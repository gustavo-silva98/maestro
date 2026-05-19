package main

import (
	"fmt"
	"log"
	"maestro/internal/config"
	"maestro/internal/integration/jira"
)

var testIssue string = "MAE-1"

func main() {
	config, _ := config.LoadConfig()
	api, err := jira.NewJiraApp(&config)
	if err != nil {
		log.Fatal(err)
	}

	//resp, _ := api.GetIssue("MAE-1")
	//fmt.Println(resp)

	accountId, err := api.SearchUserQuery(config.Jira.UserName)
	if err != nil {
		log.Fatal(err)
	}
	assign, err := api.AssignUser(testIssue, accountId)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(assign)
}
