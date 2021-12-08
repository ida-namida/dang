package model

type DecisionChainLink struct {
	DecisionName   string `json:"decision_name"`
	SourceKey      string `json:"source_key"`
	DestinationKey string `json:"destination_key"`
}