package dto

type QuestionResponse struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Question string   `json:"question"`
	Options  []string `json:"options,omitempty"`
	Position int      `json:"position"`
}

type QuizResponse struct {
	ID            string             `json:"id"`
	BookID        string             `json:"book_id"`
	VariantID     string             `json:"variant_id"`
	Title         string             `json:"title"`
	Status        string             `json:"status"`
	QuestionCount int                `json:"question_count"`
	Questions     []QuestionResponse `json:"questions"`
}