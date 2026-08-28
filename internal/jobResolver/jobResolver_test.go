package jobResolver

import (
	"context"
	"errors"
	"maestro/internal/config"
	"maestro/internal/domain"
	"maestro/internal/integration/jira"
	"path/filepath"
	"runtime"
	"testing"
)

type fakeJiraReader struct {
	Issue      jira.JiraIssue
	Err        error
	InputBytes []byte
}

type fakeOrchestrator struct {
	called       bool
	receivedJob  domain.Job
	receivedType config.JobType
	err          error
}

func (fo *fakeOrchestrator) ExecuteJob(
	ctx context.Context,
	job domain.Job,
	jt config.JobType,
	inputBytes []byte,
) (domain.Job, error) {
	fo.called = true
	fo.receivedJob = job
	fo.receivedType = jt

	return job, fo.err
}

func (fjr fakeJiraReader) GetIssue(issueId string) (jira.JiraIssue, error) {
	return fjr.Issue, fjr.Err
}

func (fjr fakeJiraReader) GetAttachmentContent(attachmentId string) ([]byte, error) {
	return fjr.InputBytes, fjr.Err
}

func TestResolveJob(t *testing.T) {
	t.Run("Erro ao receber GetJobInfo", func(t *testing.T) {
		fjr := fakeJiraReader{
			Issue: jira.JiraIssue{},
			Err:   errors.New("Erro fabricado pro teste"),
		}
		jr := JobResolver{
			jiraClient: fjr,
		}
		job := domain.Job{IssueKey: "issue"}
		_, err := jr.ResolveJob(&job)
		if err == nil {
			t.Errorf("Falha ao gerar erro no GetJobInfo")
		}
	})
	t.Run("Erro nenhum job identificado", func(t *testing.T) {
		fjr := fakeJiraReader{
			Issue: jira.JiraIssue{},
			Err:   nil,
		}
		jr := JobResolver{
			jiraClient: fjr,
			jobTypes:   map[string]config.JobType{},
		}
		job := domain.Job{IssueKey: "issue"}
		types, err := jr.ResolveJob(&job)
		if err == nil {
			t.Error("Falha ao simular job de não identificado")
		}
		if err.Error() != "Nenhum job identificado" {
			t.Errorf("Falha de retorno de erro. Recebido>: %v", err.Error())
		}
		if len(types.JobType) > 0 {
			t.Errorf("Esperado Len de types 0. Recebido: %v", len(types.JobType))
		}
	})
	t.Run("Happy Path", func(t *testing.T) {
		issue := jira.JiraIssue{
			ID: "id",
			Fields: jira.JiraIssueFields{
				Necessidade:    jira.JiraCustomField{Value: "Necessidade"},
				Sistema:        jira.JiraCustomField{Value: "Sistema"},
				JiraAttachment: []jira.JiraAttachment{{ID: "id", Filename: "NomeDoAnexo.csv"}},
			},
		}
		dir, err := testDataDir(t)
		if err != nil {
			t.Fatalf("Falha ao setar diretório de pasta: %v", err)
		}
		types, err := config.LoadJobTypes(dir)
		if err != nil {
			t.Fatalf("Falha ao carregar tipos de jobs")
		}
		jr := JobResolver{
			jobTypes: types,
			jiraClient: fakeJiraReader{
				Issue: issue,
				Err:   nil,
			},
		}

		job := domain.Job{IssueKey: "issue"}
		jobs, err := jr.ResolveJob(&job)
		if err != nil {
			t.Errorf("Erro ao testar resolveJob: %v", err)
		}
		switch {
		case jobs.JiraFields.Need != "Necessidade":
			t.Errorf("Erro jobs Campo Need: Rec. %v", jobs.JiraFields.Need)
		case jobs.JiraFields.System != "Sistema":
			t.Errorf("Erro jobs Campo Sistema: Rec. %v", jobs.JiraFields.System)
		case jobs.JiraFields.AttachmentFilename != "NomeDoAnexo.csv":
			t.Errorf("Erro ao validar nome de anexo: %v", jobs.JiraFields.AttachmentFilename)

		}
	})

}
func TestDispatchJob(t *testing.T) {

	jobType := config.JobType{
		JobType: "football",
	}
	jobType.JiraFields.System = "Sistema"
	jobType.JiraFields.Need = "Necessidade"
	jobType.JiraFields.AttachmentFilename = "input.csv"

	issue := jira.JiraIssue{
		Fields: jira.JiraIssueFields{
			Sistema:     jira.JiraCustomField{Value: "Sistema"},
			Necessidade: jira.JiraCustomField{Value: "Necessidade"},
			JiraAttachment: []jira.JiraAttachment{
				{Filename: "input.csv"},
			},
		},
	}
	t.Run("Falha no ResolveJob", func(t *testing.T) {
		fjr := jira.FakeJiraReader{
			Issue: jira.JiraIssue{},
			Err:   errors.New("Erro fabricado pro teste"),
		}
		jr := JobResolver{
			jiraClient: fjr,
		}
		job := domain.Job{IssueKey: "issue"}
		err := jr.DispatchJob(context.Background(), job)
		if err == nil {
			t.Error("Falha ao gerar erro Dispatch Job")
		}
	})
	t.Run("Despacha job com sucesso", func(t *testing.T) {
		orch := &fakeOrchestrator{}

		jr := JobResolver{
			jiraClient: fakeJiraReader{Issue: issue},
			jobTypes: map[string]config.JobType{
				jobType.JobType: jobType,
			},
			orch: orch,
		}

		job := domain.Job{
			ID:       "job-1",
			IssueKey: "PROJ-1",
		}

		err := jr.DispatchJob(context.Background(), job)
		if err != nil {
			t.Fatalf("não esperava erro: %v", err)
		}

		if !orch.called {
			t.Fatal("esperava que o orquestrador fosse chamado")
		}

		if orch.receivedJob.Type != "football" {
			t.Errorf("Type recebido: %q, esperado %q",
				orch.receivedJob.Type, "football")
		}

		if orch.receivedJob.IssueKey != "PROJ-1" {
			t.Errorf("IssueKey recebido: %q, esperado %q",
				orch.receivedJob.IssueKey, "PROJ-1")
		}

		if orch.receivedType.JobType != "football" {
			t.Errorf("JobType recebido: %q, esperado %q",
				orch.receivedType.JobType, "football")
		}
	})
	t.Run("Propaga erro do orquestrador", func(t *testing.T) {
		expectedErr := errors.New("erro ao executar job")
		orch := &fakeOrchestrator{err: expectedErr}

		jr := JobResolver{
			jiraClient: fakeJiraReader{Issue: issue},
			jobTypes: map[string]config.JobType{
				jobType.JobType: jobType,
			},
			orch: orch,
		}

		err := jr.DispatchJob(context.Background(), domain.Job{
			IssueKey: "PROJ-1",
		})

		if !errors.Is(err, expectedErr) {
			t.Errorf("erro recebido: %v, esperado: %v", err, expectedErr)
		}

		if !orch.called {
			t.Fatal("esperava que o orquestrador fosse chamado")
		}
	})
}

func testDataDir(_ *testing.T) (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("erro ao receber caller")
	}
	return filepath.Join(filepath.Dir(file), "testdata", "jobTypes"), nil

}
