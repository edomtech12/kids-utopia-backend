package model

import "time"

type QuizStatus string

const (
	QuizStatusPending    QuizStatus = "pending"
	QuizStatusGenerating QuizStatus = "generating"
	QuizStatusReady      QuizStatus = "ready"
	QuizStatusFailed     QuizStatus = "failed"
)

type QuestionType string

const (
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeFillBlank      QuestionType = "fill_blank"
)

type Quiz struct {
	ID            string
	BookID        string
	VariantID     string
	Title         string
	Status        QuizStatus
	QuestionCount int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
type QuizResult struct {
	ID             string    `json:"id"`
	QuizID         string    `json:"quiz_id"`
	ChildID        string    `json:"child_id"`
	Score          int       `json:"score"`
	TotalQuestions int       `json:"total_questions"`
	Percentage     float64   `json:"percentage"`
	CompletedAt    time.Time `json:"completed_at"`
}