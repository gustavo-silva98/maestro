package jira

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maestro/internal/config"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type JiraFieldReader interface {
	GetIssue(issueId string) (JiraIssue, error)
}

type JiraIntegration struct {
	UserName   string
	Token      string
	TenantName string
	BaseUrl    string
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
		BaseUrl:    fmt.Sprintf("https://%s.atlassian.net", cfg.Jira.TenantName),
		Client:     &http.Client{},
	}
	return app, nil
}

func (jira *JiraIntegration) GetIssueTransitions(issueId string) (JiraTransitions, error) {
	url := fmt.Sprintf("%v/rest/api/2/issue/%v/transitions", jira.BaseUrl, issueId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return JiraTransitions{}, fmt.Errorf("Falha ao buscar transição de issue: %w", err)
	}
	req.SetBasicAuth(jira.UserName, jira.Token)

	req.Header.Add("Accept", "application/json")
	resp, err := jira.Client.Do(req)
	if err != nil {
		return JiraTransitions{}, fmt.Errorf("Falha ao buscar transição de issue: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return JiraTransitions{}, fmt.Errorf("Falha ao buscar transição de issue: %w", err)
	}
	var transitionsResp JiraTransitions
	if err = json.Unmarshal(body, &transitionsResp); err != nil {
		return JiraTransitions{}, fmt.Errorf("Falha ao buscar transição de issue: %w", err)
	}
	return transitionsResp, nil

}

func (jira *JiraIntegration) SearchUserQuery(userEmail string) (string, error) {
	url := fmt.Sprintf("%v/rest/api/2/user/search?query=%v", jira.BaseUrl, userEmail)

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
	url := fmt.Sprintf("%v/rest/api/2/issue/%v/assignee", jira.BaseUrl, issueId)

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

func (jira *JiraIntegration) DoTransition(issueId string, transitionId string) (bool, error) {
	url := fmt.Sprintf("%v/rest/api/2/issue/%v/transitions", jira.BaseUrl, issueId)

	data := map[string]interface{}{
		"transition": map[string]string{"id": transitionId}}
	body, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("Falha ao atrelar usuário: %v", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
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

func (jira *JiraIntegration) Comment(issueId string, commentText string) error {
	url := fmt.Sprintf("%v/rest/api/2/issue/%v/comment", jira.BaseUrl, issueId)
	data := map[string]string{"body": commentText}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.SetBasicAuth(jira.UserName, jira.Token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := jira.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		return nil
	} else {
		return fmt.Errorf("Erro ao comentar. Código HTTP %v", resp.StatusCode)
	}
}

func (jira *JiraIntegration) GetIssue(issueId string) (JiraIssue, error) {
	var jiraIssue JiraIssue

	url := fmt.Sprintf("%v/rest/api/2/issue/%v", jira.BaseUrl, issueId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return JiraIssue{}, err
	}
	req.SetBasicAuth(jira.UserName, jira.Token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := jira.Client.Do(req)
	if err != nil {
		return JiraIssue{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return JiraIssue{}, err
	}
	err = json.Unmarshal(body, &jiraIssue)
	return jiraIssue, nil
}

func (jira *JiraIntegration) GetAttachmentContent(attachmentId string) ([]byte, error) {
	url := fmt.Sprintf("%v/rest/api/2/attachment/content/%v", jira.BaseUrl, attachmentId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}
	req.SetBasicAuth(jira.UserName, jira.Token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := jira.Client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, nil
}

func (jira *JiraIntegration) AddAttachment(issueId string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo : %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("erro ao criar form file %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("erro ao copiar arquivo %w", err)
	}

	writer.Close()

	url := fmt.Sprintf("%v/rest/api/2/issue/%v/attachments", jira.BaseUrl, issueId)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("erro ao criar request %v", err)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("X-Atlassian-Token", "no-check")
	req.SetBasicAuth(jira.UserName, jira.Token)

	resp, err := jira.Client.Do(req)
	if err != nil {
		return fmt.Errorf("erro na request %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("jira retornou %d: %s", resp.StatusCode, respBody)
	}
	return nil

}
