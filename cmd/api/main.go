package main

import (
	"fmt"
	"log"
	"maestro/internal/config"
	"maestro/internal/integration/jira"
)

func main() {
	config, _ := config.LoadConfig()
	api, err := jira.NewJiraApp(&config)
	if err != nil {
		log.Fatal(err)
	}

	//resp, _ := api.GetIssue("MAE-1")
	//fmt.Println(resp)

	user, err := api.SearchUserQuery(config.Jira.UserName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user)
}
