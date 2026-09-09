package model

import "time"

type QuizQuestion struct {
	ID         string
	QuizID     string
	Type       QuestionType
	Question   string
	Options    []string
	Answer     string
	Explanation string
	Position   int
	CreatedAt  time.Time
}