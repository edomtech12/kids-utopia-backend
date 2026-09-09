package repository

import (
	"context"

	"github.com/bellapacx/kids-utopia/internal/quiz/model"
)

type Repository interface {
	GetVariantContent(
		ctx context.Context,
		variantID string,
	) (
		bookID string,
		title string,
		pages []VariantPage,
		err error,
	)

	CreateQuizWithQuestions(
		ctx context.Context,
		quiz *model.Quiz,
		questions []model.QuizQuestion,
	) error

	GetQuizByVariant(
		ctx context.Context,
		variantID string,
	) (*model.Quiz, []model.QuizQuestion, error)
	GetQuestions(
		ctx context.Context,
		quizID string,
	) ([]model.QuizQuestion, error)
	// Attempts
CreateAttempt(
	ctx context.Context,
	attempt *model.QuizAttempt,
) error

GetAttemptsByChild(
	ctx context.Context,
	quizID string,
childID string,
) ([]model.QuizAttempt, error)

//GetAttemptByID(
//	ctx context.Context,
//	attemptID string,
//) (*model.QuizAttempt, error)
GetQuizResultsByChildID(
    ctx context.Context,
    childID string,
) ([]model.QuizResult, error)
}