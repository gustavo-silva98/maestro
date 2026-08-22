package jira

type FakeJiraReader struct {
	Issue JiraIssue
	Err   error
}

func (fjr FakeJiraReader) GetIssue(issueId string) (JiraIssue, error) {
	return fjr.Issue, fjr.Err
}
