package dto

type GenerateQuizRequest struct {
	MultipleChoiceCount int `json:"multiple_choice_count"`
	FillBlankCount      int `json:"fill_blank_count"`
}