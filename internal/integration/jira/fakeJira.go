package jira

type FakeJiraReader struct {
	Issue      JiraIssue
	Err        error
	InputBytes []byte
}

func (fjr FakeJiraReader) GetIssue(issueId string) (JiraIssue, error) {
	return fjr.Issue, fjr.Err
}

func (fjr FakeJiraReader) GetAttachmentContent(attachmentId string) ([]byte, error) {
	return fjr.InputBytes, fjr.Err
}
