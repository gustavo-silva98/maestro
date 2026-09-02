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
	"maestro/internal/jobResolver"
	"maestro/internal/orchestrator"
	"maestro/internal/repository"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Hub-Signature")

		// Se for uma requisição pre-flight OPTIONS, apenas retorne
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("Time: %s - Method: %s - URI: %s", time.Since(start), r.Method, r.URL)
	})
}

type Backend struct {
	Port         string
	ID           int
	Ready        bool
	JiraApi      *jira.JiraIntegration
	Config       *config.Config
	Orchestrator orchestrator.Orchestrator
}

type Handler struct {
	jr *jobResolver.JobResolver
	db repository.JobRepository
}

type QueryHandler struct {
	db repository.QueryData
}

func NewQueryHandler(db repository.QueryData) *QueryHandler {
	return &QueryHandler{db: db}
}

func NewHandler(jr jobResolver.JobResolver, db repository.JobRepository) *Handler {
	return &Handler{
		jr: &jr,
		db: db,
	}
}

func (h *QueryHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.db.ListJobs(r.Context())
	if err != nil {
		http.Error(w, "erro ao listar jobs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (h *Handler) HandleJiraWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Falha ao ler Body: %v", err)
		http.Error(w, "BadRequest", http.StatusBadRequest)
		return
	}

	var webhookBody jira.JiraWebhookBody
	if err := json.Unmarshal(body, &webhookBody); err != nil {
		log.Printf("Erro em enfileirar job: %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	job := domain.Job{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		IssueKey:  webhookBody.Issue.Key,
		Status:    domain.StatusPending,
	}

	if err := h.db.CreateJob(r.Context(), job); err != nil {
		log.Printf("Erro em decodificar body: %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	go func() {
		if err := h.jr.DispatchJob(context.Background(), job); err != nil {
			log.Printf("Falha ao validar job %v: %v", webhookBody.Issue.Key, err)
		} else {
			log.Printf("Simulação Job criado OK!")
		}
	}()

}

type JobStatsResponse struct {
	TotalJobs     int            `json:"totalJobs"`
	StatusLast24h map[string]int `json:"statusLast24h"`
}

func verifyHMAC(secret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Println("Falha ao ler body da request")
				http.Error(w, "erro ao ler body da request", http.StatusBadRequest)
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(body)
			expected := hex.EncodeToString(mac.Sum(nil))

			got := r.Header.Get("X-Hub-Signature-256")
			if !hmac.Equal([]byte(got), []byte(expected)) {
				log.Println("Assinatura inválida")
				http.Error(w, "Falha na autenticação", http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
}

/*
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
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Job recebido e agendado para execução."))

	// --- Inicia o processamento em background ---
	go func() {
		ctx := context.Background()
		job := domain.Job{
			ID:         uuid.NewString(),
			IssueKey:   issue,
			TenantName: api.Config.Jira.TenantName,
			Status:     "Running",
			CreatedAt:  time.Now().UTC(),
			FinishedAt: time.Now().UTC(),
			Type:       "Futebol",
		}
		task := domain.Task{
			ID:         uuid.NewString(),
			JobID:      job.ID,
			Status:     "Pending",
			CreatedAt:  time.Now().UTC(),
			FinishedAt: time.Now().UTC(), // Mantendo como você pediu
			TaskType:   "Pesquisa Futebol",
		}

		accountId, err := api.JiraApi.SearchUserQuery(api.Config.Jira.UserName)
		if err != nil {
			log.Printf("ERRO no job %s: missing user: %v", job.ID, err)
			return
		}
		_, err = api.JiraApi.AssignUser(issue, accountId)
		if err != nil {
			log.Printf("ERRO no job %s: error on assign user: %v", job.ID, err)
			return
		}
		transitions, err := api.JiraApi.GetIssueTransitions(issue)
		if err != nil {
			log.Printf("ERRO no job %s: Error get issueTransition: %v", job.ID, err)
			return
		}

		for _, val := range transitions.Transitions {
			if val.To.Name == api.Config.Jira.StatusAllowed.InitialStatus {
				transitionProcess, err := api.JiraApi.DoTransition(issue, val.ID)
				if err != nil {
					log.Printf("ERRO no job %s: Error transitioning issue In Progress: %v", job.ID, err)
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
			log.Printf("ERRO no job %s: Error get Issue: %v", job.ID, err)
			return
		}

		attachment, err := api.JiraApi.GetAttachmentContent(getIssue.Fields.JiraAttachment[0].ID)
		if err != nil {
			log.Printf("ERRO no job %s: Error get IssueAttachment: %v", job.ID, err)
			return
		}
		err = os.RemoveAll("scripts/football/output")
		if err != nil {
			log.Printf("ERRO no job %s: Erro ao excluir pasta %v", job.ID, err)
			return
		}
		os.WriteFile("scripts/football/input.csv", attachment, 0644)
		if err := api.JiraApi.Comment(issue, "Iniciando job"); err != nil {
			log.Printf("ERRO no job %s: Falha ao comentar chamado: %v", job.ID, err)
			return
		}
		task.SavedMinutes = countLinesFast(attachment) - 1
		if err := api.DB.CreateJob(ctx, job); err != nil {
			log.Printf("ERRO no job %s: erro ao criar Job %v", job.ID, err)
		} else {
			log.Printf("job criado para issue %v - ID: %v", job.IssueKey, job.ID)
		}

		if err := api.DB.CreateTask(ctx, task); err != nil {
			log.Printf("ERRO no job %s: erro ao criar task %v", job.ID, err)
		} else {
			log.Printf("task criada para JobId %v", task.JobID)
		}

		log.Printf("Iniciando execução do Job %s", job.ID)
		if err := api.DB.SetJobRunning(ctx, job); err != nil {
			log.Printf("ERRO no job %s: erro ao setar job running: %v", job.ID, err)
			return
		}

		// A execução longa acontece aqui
		_, err = api.Orchestrator.ExecuteJob()
		if err != nil {
			log.Printf("ERRO na execução do Job %s: %v", job.ID, err)
			// Aqui você poderia implementar uma lógica para marcar o job como "Failed"
			return
		}
		log.Printf("Job %s executado com sucesso.", job.ID)

		if err := ZipFolder("scripts/football/output", "scripts/football/output.zip"); err != nil {
			log.Printf("ERRO no job %s: Error ziping output: %v", job.ID, err)
			return
		}

		if err := api.JiraApi.AddAttachment(issue, "scripts/football/output.zip"); err != nil {
			log.Printf("ERRO no job %s: Error adding output: %v", job.ID, err)
			return
		}
		comment := fmt.Sprintf("Job Finalizado. Output: \n\n!%s!\n\n", "output.zip")
		if err := api.JiraApi.Comment(issue, comment); err != nil {
			log.Printf("ERRO no job %s: erro ao comentar anexo %v", job.ID, err)
		}

		transitions, err = api.JiraApi.GetIssueTransitions(issue)
		if err != nil {
			log.Printf("ERRO no job %s: Error get issueTransition on finish: %v", job.ID, err)
			return
		}
		for _, val := range transitions.Transitions {
			if val.To.Name == api.Config.Jira.StatusAllowed.FinalStatus {
				transitionProcess, err := api.JiraApi.DoTransition(issue, val.ID)
				if err != nil {
					log.Printf("ERRO no job %s: Error transitioning issue Done: %v", job.ID, err)
					return
				}
				if transitionProcess {
					api.JiraApi.Comment(issue, "Transição feita: Done.")
				}
				break
			}
		}

		job.FinishedAt = time.Now().UTC()
		err = api.DB.FinishJob(ctx, job)
		if err != nil {
			log.Printf("ERRO no job %s: erro ao finalizar job: %v", job.ID, err)
			return
		}
		log.Printf("Job %s finalizado com sucesso.", job.ID)
	}() // A `()` no final executa a função anônima
}
*/

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
