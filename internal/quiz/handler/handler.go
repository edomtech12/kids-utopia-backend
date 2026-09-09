package handler

import (
	"net/http"
	"time"

	"github.com/bellapacx/kids-utopia/internal/quiz/dto"
	"github.com/bellapacx/kids-utopia/internal/quiz/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{
		service: s,
	}
}

// POST /books/variants/:variantID/quiz/generate
func (h *Handler) GenerateQuiz(c *gin.Context) {

	variantID := c.Param("variantID")

	if variantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "variant_id is required",
		})
		return
	}

	var req dto.GenerateQuizRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	if req.MultipleChoiceCount < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "multiple_choice_count must be at least 1",
		})
		return
	}

	if req.FillBlankCount < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "fill_blank_count must be at least 1",
		})
		return
	}

	result, err := h.service.GenerateQuiz(
		c.Request.Context(),
		variantID,
		req.MultipleChoiceCount,
		req.FillBlankCount,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": result,
	})
}

// GET /books/variants/:variantID/quiz
func (h *Handler) GetQuizByVariant(c *gin.Context) {

	variantID := c.Param("variantID")

	if variantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "variant_id is required",
		})
		return
	}

	result, err := h.service.GetQuizByVariant(
		c.Request.Context(),
		variantID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}
func (h *Handler) SubmitQuizAttempt(c *gin.Context) {
	quizID := c.Param("quizID")

	if quizID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "quiz_id is required",
		})
		return
	}

	var req dto.SubmitQuizAttemptRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	attempt, err := h.service.SubmitQuizAttempt(
		c.Request.Context(),
		quizID,
		req,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": dto.QuizAttemptResponse{
			ID:             attempt.ID,
			QuizID:         attempt.QuizID,
			ChildID:        attempt.ChildID,
			Score:          attempt.Score,
			TotalQuestions: attempt.TotalQuestions,
			Percentage:     attempt.Percentage,
			CompletedAt:    attempt.CompletedAt.Format(time.RFC3339),
		},
	})
}
func (h *Handler) GetAttemptsByChild(c *gin.Context) {
	quizID := c.Param("quizID")
	childID := c.Param("childId")

	if quizID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "quiz_id is required",
		})
		return
	}

	if childID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "child_id is required",
		})
		return
	}

	attempts, err := h.service.GetAttemptsByChild(
		c.Request.Context(),
		quizID,
		childID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": attempts,
	})
}
func (h *Handler) GetQuizResultsByChildID(c *gin.Context) {

	childID := c.Param("childId")

	if childID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "child_id is required",
		})
		return
	}

	results, err := h.service.GetQuizResultsByChildID(
		c.Request.Context(),
		childID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}