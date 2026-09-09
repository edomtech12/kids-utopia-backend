package service

import (
	"context"

	"github.com/bellapacx/kids-utopia/pkg/openai"
)

type AIClient interface {
	GenerateQuiz(
		ctx context.Context,
		req openai.QuizRequest,
	) (*openai.QuizResponse, error)
}