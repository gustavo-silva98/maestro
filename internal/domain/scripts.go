package domain

type FootballRpaResults struct {
	Tasks []struct {
		Atleta  string `json:"atleta"`
		Success bool   `json:"success"`
	} `json:"tasks"`
}
