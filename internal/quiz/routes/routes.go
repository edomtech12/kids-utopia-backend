package routes

import (
	"github.com/bellapacx/kids-utopia/internal/quiz/handler"
	"github.com/gin-gonic/gin"
)
func Register(
	router *gin.RouterGroup,
	h *handler.Handler,
) {
	variantRoutes := router.Group("/books/variants")

	variantRoutes.POST(
		"/:variantID/quiz/generate",
		h.GenerateQuiz,
	)

	variantRoutes.GET(
		"/:variantID/quiz",
		h.GetQuizByVariant,
	)

	quizRoutes := router.Group("/quizzes")

	quizRoutes.POST(
		"/:quizID/attempts",
		h.SubmitQuizAttempt,
	)
	quizRoutes.GET(
	"/:quizID/attempts/child/:childId",
	h.GetAttemptsByChild,
)
router.GET(
		"/children/:childId/quizzes",
		h.GetQuizResultsByChildID,
	)
}