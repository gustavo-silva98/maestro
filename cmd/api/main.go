package main

import (
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
	if !assign {
		log.Fatal("ERRO: Chamado não foi atribuído ao usuário")
	}
	transitions, err := api.GetIssueTransitions(testIssue)
	if err != nil {
		log.Fatal(err)
	}
	for _, val := range transitions.Transitions {
		if val.To.StatusCategory.Key == config.Jira.StatusAllowed.InitialStatus {
			do, err := api.DoTransition(testIssue, val.ID)
			if err != nil {
				log.Fatal(err)
			}
			if !do {
				log.Fatal("ERRO: Transição não foi concluída")
			}
			break
		}
	}
}
