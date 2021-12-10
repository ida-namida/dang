package view

type SaveDecisionRequest struct {
	Decisions []DecisionRequest `json:"decisions"`
}

type DecisionRequest struct {
	Name       string             `json:"name"`
	InputForm  []FormFieldRequest `json:"input_form"`
	OutputForm []FormFieldRequest `json:"output_form"`
}