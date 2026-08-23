package jobResolver

import (
	"errors"
	"maestro/internal/config"
	"maestro/internal/integration/jira"
	"path/filepath"
	"runtime"
	"testing"
)

type fakeJiraReader struct {
	Issue jira.JiraIssue
	Err   error
}

func (fjr fakeJiraReader) GetIssue(issueId string) (jira.JiraIssue, error) {
	return fjr.Issue, fjr.Err
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
		_, err := jr.ResolveJob("issue")
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
		types, err := jr.ResolveJob("issue")
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
		jobs, err := jr.ResolveJob("issue")
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

func testDataDir(_ *testing.T) (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("erro ao receber caller")
	}
	return filepath.Join(filepath.Dir(file), "testdata", "jobTypes"), nil

}
