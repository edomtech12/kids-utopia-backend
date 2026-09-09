package modules

import (
	"context"

	"github.com/gin-gonic/gin"

	appcontainer "github.com/bellapacx/kids-utopia/internal/app"
	quizhandler "github.com/bellapacx/kids-utopia/internal/quiz/handler"
	quizrepository "github.com/bellapacx/kids-utopia/internal/quiz/repository"
	quizroutes "github.com/bellapacx/kids-utopia/internal/quiz/routes"
	quizservice "github.com/bellapacx/kids-utopia/internal/quiz/service"

	geminiClient "github.com/bellapacx/kids-utopia/pkg/gemini"
)

func RegisterQuiz(
	r *gin.Engine,
	container *appcontainer.Container,
) error {

	// =========================
	// GEMINI CLIENT
	// =========================

	gemini, err := geminiClient.NewClient(
		context.Background(),
	)
	if err != nil {
		return err
	}

	// =========================
	// REPOSITORY
	// =========================

	quizRepo := quizrepository.NewPostgresRepository(
		container.DB,
	)

	// =========================
	// SERVICE
	// =========================

	quizService := quizservice.NewService(
		quizRepo,
		gemini,
	)

	// =========================
	// HANDLER
	// =========================

	quizHandler := quizhandler.NewHandler(
		quizService,
	)

	// =========================
	// ROUTES
	// =========================

	quizroutes.Register(
		r.Group("/api/v1"),
		quizHandler,
	)

	return nil
}