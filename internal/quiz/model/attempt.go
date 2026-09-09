package model

import "time"

type QuizAttempt struct {
	ID             string
	QuizID         string
	ChildID        string
	Score          int
	TotalQuestions int
	Percentage     float64
	CompletedAt    time.Time
}