package model

type FormField struct {
	Key         string             `json:"key"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Unit        string             `json:"unit"`
	Type        string             `json:"type"`
	ChainLink   *DecisionChainLink `json:"chain,omitempty"`
	Rule        *string            `json:"rule,omitempty"`
}