package view

type FormFieldRequest struct {
	Key         string      `json:"key"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Unit        string      `json:"unit"`
	Type        string      `json:"type"`
	Value       string      `json:"value"`
	Rule        RuleRequest `json:"rule"`
}