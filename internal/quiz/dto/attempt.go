package dto

type SubmitQuizAnswer struct {
	QuestionID string `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
}

type SubmitQuizAttemptRequest struct {
	ChildID string              `json:"child_id" binding:"required"`
	Answers []SubmitQuizAnswer `json:"answers" binding:"required,min=1"`
}

type QuizAttemptResponse struct {
	ID             string  `json:"id"`
	QuizID         string  `json:"quiz_id"`
	ChildID        string  `json:"child_id"`
	Score          int     `json:"score"`
	TotalQuestions int     `json:"total_questions"`
	Percentage     float64 `json:"percentage"`
	CompletedAt    string  `json:"completed_at"`
}