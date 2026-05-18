package jira

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type JiraIntegration struct {
	UserName   string
	Token      string
	TenantName string
	Client     *http.Client
}

func NewJiraApp() (JiraIntegration, error) {
	if os.Getenv("USERNAME_JIRA") == "" || os.Getenv("TOKEN_JIRA") == "" || os.Getenv("TENANT_NAME_JIRA") == "" {
		return JiraIntegration{}, errors.New("Variáveis de ambiente não foram carregadas corretamente")
	}

	app := JiraIntegration{
		UserName:   os.Getenv("USERNAME_JIRA"),
		Token:      os.Getenv("TOKEN_JIRA"),
		TenantName: os.Getenv("TENANT_NAME_JIRA"),
		Client:     &http.Client{},
	}
	return app, nil
}

func (jira *JiraIntegration) GetIssue(issueId string) (string, error) {
	url := fmt.Sprintf("https://%v.atlassian.net/rest/api/2/issue/%v", jira.TenantName, issueId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(jira.UserName, jira.Token)

	req.Header.Add("Accept", "application/json")
	resp, err := jira.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	respString := string(body)
	return respString, nil

}
