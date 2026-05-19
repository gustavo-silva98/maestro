package jira

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maestro/internal/config"
	"net/http"
)

type JiraIntegration struct {
	UserName   string
	Token      string
	TenantName string
	Client     *http.Client
}

func NewJiraApp(cfg *config.Config) (JiraIntegration, error) {
	if cfg.Jira.UserName == "" || cfg.Jira.Token == "" || cfg.Jira.TenantName == "" {
		return JiraIntegration{}, errors.New("Variáveis de ambiente não foram carregadas corretamente")
	}
	app := JiraIntegration{
		UserName:   cfg.Jira.UserName,
		Token:      cfg.Jira.Token,
		TenantName: cfg.Jira.TenantName,
		Client:     &http.Client{},
	}
	return app, nil
}

func (jira *JiraIntegration) GetIssueTransitions(issueId string) (string, error) {
	url := fmt.Sprintf("https://%v.atlassian.net/rest/api/2/issue/%v/transitions", jira.TenantName, issueId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("Falha ao buscar transição de issue: %w", err)
	}
	req.SetBasicAuth(jira.UserName, jira.Token)

	req.Header.Add("Accept", "application/json")
	resp, err := jira.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Falha ao buscar transição de issue: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Falha ao buscar transição de issue: %w", err)
	}
	respString := string(body)
	return respString, nil

}

func (jira *JiraIntegration) GetIssue(issueId string) (string, error) {
	url := fmt.Sprintf("https://%v.atlassian.net/rest/api/2/issue/%v", jira.TenantName, issueId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("Falha ao buscar issue: %w", err)
	}
	req.SetBasicAuth(jira.UserName, jira.Token)

	req.Header.Add("Accept", "application/json")
	resp, err := jira.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Falha ao buscar issue: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Falha ao buscar issue: %w", err)
	}
	respString := string(body)
	return respString, nil

}

func (jira *JiraIntegration) SearchUserQuery(userEmail string) (string, error) {
	url := fmt.Sprintf("https://%v.atlassian.net/rest/api/2/user/search?query=%v", jira.TenantName, userEmail)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("Falha ao buscar usuário: %w", err)
	}
	req.SetBasicAuth(jira.UserName, jira.Token)

	req.Header.Add("Accept", "application/json")
	resp, err := jira.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Falha ao buscar usuário: %w", err)
	}
	defer resp.Body.Close()

	var searchUser JiraSearchUser

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Falha ao buscar usuário: %w", err)
	}
	err = json.Unmarshal(body, &searchUser)
	if err != nil {
		return "", err
	}
	if len(searchUser) > 0 {
		return searchUser[0].AccountID, nil
	} else {
		return "", errors.New("Falha ao buscar usuário: Usuário não encontrado.")
	}
}
func (jira *JiraIntegration) AssignUser(issueId string, accountId string) (bool, error) {
	url := fmt.Sprintf("https://%v.atlassian.net/rest/api/2/issue/%v/assignee", jira.TenantName, issueId)

	data := map[string]string{"accountId": accountId}
	body, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("Falha ao atrelar usuário: %v", err)
	}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return false, fmt.Errorf("Falha ao atrelar usuário: %v", err)
	}
	req.SetBasicAuth(jira.UserName, jira.Token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := jira.Client.Do(req)
	if err != nil {
		return false, fmt.Errorf("Falha ao atrelar usuário: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return true, nil
	} else {
		return false, nil
	}
}
