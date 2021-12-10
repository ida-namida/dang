package view

type RuleRequest struct {
	Type string      `json:"type"`
	Args []string    `json:"args"`
	Rule interface{} `json:"rule"`
}