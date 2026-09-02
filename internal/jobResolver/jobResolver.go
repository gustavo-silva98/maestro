package jobResolver

import (
	"context"
	"errors"
	"fmt"
	"maestro/internal/config"
	"maestro/internal/domain"
	"maestro/internal/integration/jira"
	"maestro/internal/orchestrator"
)

type JobResolver struct {
	jobTypes   map[string]config.JobType
	jiraClient jira.JiraFieldReader
	orch       orchestrator.Orchestrator
}

func NewJobResolver(jiraClient jira.JiraFieldReader, jt map[string]config.JobType, orch orchestrator.Orchestrator) *JobResolver {
	return &JobResolver{
		jobTypes:   jt,
		jiraClient: jiraClient,
		orch:       orch,
	}
}

func (jr *JobResolver) GetJobInfo(issueId string) (jira.JiraIssue, error) {
	return jr.jiraClient.GetIssue(issueId)
}

func (jr *JobResolver) ResolveJob(job *domain.Job) (config.JobType, error) {
	issueJson, err := jr.GetJobInfo(job.IssueKey)
	if err != nil {
		return config.JobType{}, err
	}
	for _, jt := range jr.jobTypes {
		if jt.JiraFields.System == issueJson.Fields.Sistema.Value {
			if jt.JiraFields.Need == issueJson.Fields.Necessidade.Value {
				for _, att := range issueJson.Fields.JiraAttachment {
					if att.Filename == jt.JiraFields.AttachmentFilename {
						job.InputFileId = att.ID
						return jt, nil
					}
				}
			}
		}
	}
	return config.JobType{}, errors.New("Nenhum job identificado")
}

func (jr *JobResolver) DispatchJob(ctx context.Context, job domain.Job) error {
	jobType, err := jr.ResolveJob(&job)
	if err != nil {
		return fmt.Errorf("Erro ao achar tipo de job: %v", err)
	}
	inputBytes, err := jr.jiraClient.GetAttachmentContent(job.InputFileId)

	job.Type = jobType.JobType
	// Implementar a definição de ItemCount

	_, err = jr.orch.ExecuteJob(ctx, job, jobType, inputBytes)
	return err
}
