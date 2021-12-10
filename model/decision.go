package model

type Decision struct {
	Name       string      `json:"name"`
	InputForm  []FormField `json:"input_form"`
	OutputForm []FormField `json:"output_form"`
}