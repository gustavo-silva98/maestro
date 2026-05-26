package jira

type JiraSearchUser []struct {
	AccountID string `json:"accountId"`
}

type JiraTransitions struct {
	Transitions []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		To   struct {
			Self           string `json:"self"`
			Description    string `json:"description"`
			IconURL        string `json:"iconUrl"`
			Name           string `json:"name"`
			ID             string `json:"id"`
			StatusCategory struct {
				Self      string `json:"self"`
				ID        int    `json:"id"`
				Key       string `json:"key"`
				ColorName string `json:"colorName"`
				Name      string `json:"name"`
			} `json:"statusCategory"`
		} `json:"to"`
		HasScreen     bool `json:"hasScreen"`
		IsGlobal      bool `json:"isGlobal"`
		IsInitial     bool `json:"isInitial"`
		IsAvailable   bool `json:"isAvailable"`
		IsConditional bool `json:"isConditional"`
		IsLooped      bool `json:"isLooped"`
	} `json:"transitions"`
}

type JiraWebhookBody struct {
	Issue struct {
		Key string `json:"key"`
	} `json:"issue"`
}
