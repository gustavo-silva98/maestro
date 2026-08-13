package jobResolver

import (
	"errors"
	"maestro/internal/config"
	"maestro/internal/integration/jira"
)

type JobResolver struct {
	JobTypes   map[string]config.JobType
	jiraClient jira.JiraFieldReader
}

func (jr *JobResolver) GetJobInfo(issueId string) (jira.JiraIssue, error) {
	return jr.jiraClient.GetIssue(issueId)
}

func (jr *JobResolver) ResolveJob(issue string) (config.JobType, error) {
	issueJson, err := jr.GetJobInfo(issue)
	if err != nil {
		return config.JobType{}, err
	}
	for _, jt := range jr.JobTypes {
		if jt.JiraFields.System == issueJson.Fields.Sistema.Value {
			if jt.JiraFields.Need == issueJson.Fields.Necessidade.Value {
				for _, att := range issueJson.Fields.JiraAttachment {
					if att.Filename == jt.JiraFields.AttachmentFilename {
						return jt, nil
					}
				}
			}
		}
	}
	return config.JobType{}, errors.New("Nenhum job identificado")
}
