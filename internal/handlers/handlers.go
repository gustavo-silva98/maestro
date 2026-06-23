package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maestro/internal/config"
	"maestro/internal/domain"
	"maestro/internal/integration/jira"
	"maestro/internal/orchestrator"
	"maestro/internal/repository"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
)

type Backend struct {
	Port         string
	ID           int
	Ready        bool
	JiraApi      *jira.JiraIntegration
	Config       *config.Config
	DB           repository.MaestroRepository
	Orchestrator orchestrator.Orchestrator
}

type JobStatsResponse struct {
	TotalJobs     int            `json:"totalJobs"`
	StatusLast24h map[string]int `json:"statusLast24h"`
}

func (api *Backend) GetJobStatusCounts(w http.ResponseWriter, r *http.Request) {
	statusCounts, err := api.DB.GetJobStatusCountsLast24h(r.Context())
	if err != nil {
		http.Error(w, "erro ao buscar contagem de status", http.StatusInternalServerError)
		return
	}

	totalJobs, err := api.DB.GetTotalJobsCount(r.Context())
	if err != nil {
		http.Error(w, "erro ao buscar total de jobs", http.StatusInternalServerError)
		return
	}

	response := JobStatsResponse{
		TotalJobs:     totalJobs,
		StatusLast24h: statusCounts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("erro ao serializar resposta de status: %v", err)
	}
}

func (api *Backend) ReadyEndpoint(w http.ResponseWriter, r *http.Request) {
	api.Ready = true
	resp := map[string]interface{}{
		"apiId": api.ID,
		"ready": api.Ready,
	}
	jsonData, err := sonic.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(jsonData)
}

func (api *Backend) TestAutomation(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	xhub := strings.Split(r.Header.Get("X-Hub-Signature"), "=")
	if len(xhub) != 2 {
		http.Error(w, "Erro ao validar signature webhook", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Falha ao ler body: %v", err), http.StatusBadRequest)
		return
	}

	if !checkHmac(api.Config.Jira.WebhookSecret, xhub[1], body) {
		log.Println("Autenticação do Webhook falhou")
		http.Error(w, "Autenticação do Webhook falhou!", http.StatusBadRequest)
		return
	} else {
		log.Println("HMAC Signature validado!")
	}

	var webhook_body jira.JiraWebhookBody
	r.Body = io.NopCloser(bytes.NewReader(body))
	if err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&webhook_body); err != nil {
		log.Printf("decode error: %v", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	log.Printf("Requisição recebida - Issue %v", webhook_body.Issue.Key)

	issue := webhook_body.Issue.Key
	if issue == "" {
		http.Error(w, "missing issue key", http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	job := domain.Job{
		ID:          uuid.NewString(),
		IssueKey:    issue,
		TentantName: api.Config.Jira.TenantName,
		Status:      "Pending",
		CreatedAt:   time.Now().UTC(),
		FinishedAt:  time.Now().UTC(),
		JobType:     "Futebol",
	}
	task := domain.Task{
		ID:         uuid.NewString(),
		JobID:      job.ID,
		Status:     "Pending",
		CreatedAt:  time.Now().UTC(),
		FinishedAt: time.Now().UTC(),
		TaskType:   "Pesquisa Futebol",
	}

	accountId, err := api.JiraApi.SearchUserQuery(api.Config.Jira.UserName)
	if err != nil {
		http.Error(w, "missing user", http.StatusBadGateway)
		return
	}
	_, err = api.JiraApi.AssignUser(issue, accountId)
	if err != nil {
		http.Error(w, "error on assign user", http.StatusBadGateway)
		return
	}
	transitions, err := api.JiraApi.GetIssueTransitions(issue)
	if err != nil {
		http.Error(w, "Error get issueTransition", http.StatusBadGateway)
		return
	}

	for _, val := range transitions.Transitions {
		if val.To.Name == api.Config.Jira.StatusAllowed.InitialStatus {
			transitionProcess, err := api.JiraApi.DoTransition(issue, val.ID)
			if err != nil {
				http.Error(w, "Error transitioning issue In Progress", http.StatusBadGateway)
				return
			}
			if transitionProcess {
				api.JiraApi.Comment(issue, "Transição feita: Em progresso.")
			}
			break
		}
	}

	getIssue, err := api.JiraApi.GetIssue(issue)
	if err != nil {
		http.Error(w, "Error get Issue", http.StatusBadGateway)
		return
	}

	attachment, err := api.JiraApi.GetAttachmentContent(getIssue.Fields.JiraAttachment[0].ID)
	if err != nil {
		http.Error(w, "Error get IssueAttachment", http.StatusBadGateway)
		return
	}
	err = os.RemoveAll("scripts/football/output")
	if err != nil {
		log.Printf("Erro ao excluir pasta %v", err)
		return
	}
	os.WriteFile("scripts/football/input.csv", attachment, 0644)
	if err := api.JiraApi.Comment(issue, "Iniciando job"); err != nil {
		http.Error(w, "Falha ao comentar chamado", http.StatusBadGateway)
		return
	}
	job.SavedMinutes = float64(countLinesFast(attachment) - 1)
	if err := api.DB.CreateJob(ctx, job); err != nil {
		log.Printf("erro ao criar Job %v", err)
	} else {
		log.Printf("job criado para issue %v - ID: %v", job.IssueKey, job.ID)
	}

	if err := api.DB.CreateTask(ctx, task); err != nil {
		log.Printf("erro ao criar task %v", err)
	} else {
		log.Printf("task criada para JobId %v", task.JobID)
	}
	execution := orchestrator.ExecutionResult{
		JobID:     job.ID,
		StartedAt: time.Now().UTC(),
		Job:       job,
		Task:      task,
	}
	log.Println("Iniciando execução de Job")
	if err := api.DB.SetJobRunning(ctx, job); err != nil {
		http.Error(w, "erro ao setar job running", http.StatusInternalServerError)
		return
	}
	executionResult, err := api.Orchestrator.ExecuteJob(execution)
	fmt.Println(executionResult)
	transitions, err = api.JiraApi.GetIssueTransitions(issue)

	if err := ZipFolder("scripts/football/output", "scripts/football/output.zip"); err != nil {
		http.Error(w, "Error ziping output", http.StatusBadGateway)
		return
	}

	if err := api.JiraApi.AddAttachment(issue, "scripts/football/output.zip"); err != nil {
		http.Error(w, "Error adding output", http.StatusBadGateway)
		return
	}
	comment := fmt.Sprintf("Job Finalizado. Output: \n\n!%s!\n\n", "output.zip")
	if err := api.JiraApi.Comment(issue, comment); err != nil {
		log.Printf("erro ao comentar anexo %v", err)
	}

	if err != nil {
		http.Error(w, "Error get issueTransition", http.StatusBadGateway)
		return
	}
	for _, val := range transitions.Transitions {
		if val.To.Name == api.Config.Jira.StatusAllowed.FinalStatus {
			transitionProcess, err := api.JiraApi.DoTransition(issue, val.ID)
			if err != nil {
				http.Error(w, "Error transitioning issue Done", http.StatusBadGateway)
				return
			}
			if transitionProcess {
				api.JiraApi.Comment(issue, "Transição feita: Done.")
			}
			break
		}
	}
	err = api.DB.FinishJob(ctx, job)
	if err != nil {
		http.Error(w, "erro ao finalizar job", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func checkHmac(secret, received string, data []byte) bool {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(data))
	expected := hash.Sum(nil)

	receivedHex, err := hex.DecodeString(received)
	if err != nil {
		return false
	}
	return hmac.Equal(expected, []byte(receivedHex))
}

func ZipFolder(srcDir, destZipPath string) error {
	// Garante que o diretório de origem existe
	info, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("erro ao acessar srcDir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s não é um diretório", srcDir)
	}

	zipFile, err := os.Create(destZipPath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo zip: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(srcDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Caminho relativo dentro do zip (mantém a estrutura de pastas)
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Pula a própria raiz
		if relPath == "." {
			return nil
		}

		// Normaliza separadores para "/" (padrão do formato zip)
		relPath = filepath.ToSlash(relPath)

		if fi.IsDir() {
			// Cria entrada de diretório (com "/" no final)
			_, err := zipWriter.Create(relPath + "/")
			return err
		}

		// Cria entrada do header preservando metadados (data, permissões)
		header, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate // compressão (use zip.Store para só empacotar sem compactar)

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(writer, srcFile)
		return err
	})
}

func (api *Backend) GetJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := api.DB.GetJobs(r.Context(), 50)
	if err != nil {
		http.Error(w, "erro ao buscar jobs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func countLinesFast(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	n := bytes.Count(b, []byte{'\n'})
	if b[len(b)-1] != '\n' {
		n++
	}
	return n
}

func (api *Backend) DeleteTables(w http.ResponseWriter, r *http.Request) {
	if err := api.DB.ClearTables(r.Context()); err != nil {
		http.Error(w, "erro ao deletar tabelas", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (api *Backend) GetSavedMinutes(w http.ResponseWriter, r *http.Request) {
	minutes, err := api.DB.GetSavedMinutes(r.Context())
	if err != nil {
		http.Error(w, "erro ao consultar tempo economizado", http.StatusInternalServerError)
		return
	}
	resp := map[string]float64{
		"savedMinutes": minutes,
	}
	jsonData, err := sonic.Marshal(resp)
	if err != nil {
		http.Error(w, "erro ao serializar resposta", http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
	w.WriteHeader(http.StatusOK)
}
