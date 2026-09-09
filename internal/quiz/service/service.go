package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bellapacx/kids-utopia/internal/quiz/dto"
	"github.com/bellapacx/kids-utopia/internal/quiz/model"
	"github.com/bellapacx/kids-utopia/internal/quiz/repository"
	geminiClient "github.com/bellapacx/kids-utopia/pkg/gemini"
)

type Service struct {
	repo   repository.Repository
	gemini *geminiClient.Client
}

func NewService(
	repo repository.Repository,
	gemini *geminiClient.Client,
) *Service {
	return &Service{
		repo:   repo,
		gemini: gemini,
	}
}
func (s *Service) GenerateQuiz(
	ctx context.Context,
	variantID string,
	multipleChoiceCount int,
	fillBlankCount int,
) (*model.Quiz, error) {

	log.Printf(
		"🧠 QUIZ GENERATION START variant=%s mc=%d fill=%d",
		variantID,
		multipleChoiceCount,
		fillBlankCount,
	)

	// =========================
	// VALIDATE INPUT
	// =========================

	if variantID == "" {
		return nil, fmt.Errorf("variant_id is required")
	}

	if multipleChoiceCount < 1 {
		return nil, fmt.Errorf(
			"multiple choice count must be at least 1",
		)
	}

	if fillBlankCount < 1 {
		return nil, fmt.Errorf(
			"fill blank count must be at least 1",
		)
	}

	// =========================
	// LOAD VARIANT CONTENT
	// =========================

	bookID, variantTitle, pages, err :=
		s.repo.GetVariantContent(
			ctx,
			variantID,
		)

	if err != nil {
		log.Printf(
			"❌ failed to load variant=%s err=%v",
			variantID,
			err,
		)

		return nil, err
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf(
			"variant has no pages",
		)
	}

	log.Printf(
		"📖 variant loaded variant=%s book=%s title=%q pages=%d",
		variantID,
		bookID,
		variantTitle,
		len(pages),
	)

	// =========================
	// BUILD STORY CONTENT
	// =========================

	var contentBuilder strings.Builder

	for _, page := range pages {

		if strings.TrimSpace(page.Content) == "" {
			continue
		}

		fmt.Fprintf(
			&contentBuilder,
			"\n\n[Page %d]\n%s",
			page.PageNumber,
			page.Content,
		)
	}

	content := strings.TrimSpace(
		contentBuilder.String(),
	)

	if content == "" {
		return nil, fmt.Errorf(
			"variant pages contain no readable content",
		)
	}

	log.Printf(
		"📝 quiz content prepared chars=%d",
		len(content),
	)

	// =========================
	// CALL OPENAI
	// =========================

	result, err := s.gemini.GenerateQuiz(
		ctx,
		geminiClient.QuizRequest{
			Content:             content,
			MultipleChoiceCount: multipleChoiceCount,
			FillBlankCount:      fillBlankCount,
		},
	)

	if err != nil {
		log.Printf(
			"❌ OpenAI quiz generation failed variant=%s err=%v",
			variantID,
			err,
		)

		return nil, err
	}

	log.Printf(
		"🤖 OpenAI generated title=%q questions=%d",
		result.Title,
		len(result.Questions),
	)

	// =========================
	// CREATE QUIZ
	// =========================

	now := time.Now()

	quiz := &model.Quiz{
		ID:            uuid.NewString(),
		BookID:        bookID,
		VariantID:     variantID,
		Title:         result.Title,
		Status:         model.QuizStatusReady,
		QuestionCount: len(result.Questions),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// =========================
	// BUILD QUESTIONS
	// =========================

	questions := make(
		[]model.QuizQuestion,
		0,
		len(result.Questions),
	)

	for i, generated := range result.Questions {

		questionType := model.QuestionType(
			generated.Type,
		)

		question := model.QuizQuestion{
			ID:          uuid.NewString(),
			QuizID:      quiz.ID,
			Type:        questionType,
			Question:    generated.Question,
			Options:     generated.Options,
			Answer:      generated.Answer,
			Explanation: generated.Explanation,
			Position:    i + 1,
			CreatedAt:   now,
		}

		questions = append(
			questions,
			question,
		)
	}

	log.Printf(
		"📝 questions prepared quiz=%s count=%d",
		quiz.ID,
		len(questions),
	)

	// =========================
	// SAVE QUIZ + QUESTIONS
	// =========================

	if err := s.repo.CreateQuizWithQuestions(
		ctx,
		quiz,
		questions,
	); err != nil {

		log.Printf(
			"❌ failed to save quiz quiz=%s err=%v",
			quiz.ID,
			err,
		)

		return nil, err
	}

	log.Printf(
		"✅ QUIZ GENERATION COMPLETE quiz=%s questions=%d",
		quiz.ID,
		len(questions),
	)

	return quiz, nil
}
func (s *Service) GetQuizByVariant(
	ctx context.Context,
	variantID string,
) (*dto.QuizResponse, error) {

	if variantID == "" {
		return nil, fmt.Errorf("variant_id is required")
	}

	quiz, questions, err := s.repo.GetQuizByVariant(
		ctx,
		variantID,
	)

	if err != nil {
		return nil, err
	}

	response := &dto.QuizResponse{
		ID:            quiz.ID,
		BookID:        quiz.BookID,
		VariantID:     quiz.VariantID,
		Title:         quiz.Title,
		Status:        string(quiz.Status),
		QuestionCount: quiz.QuestionCount,
		Questions: make(
			[]dto.QuestionResponse,
			0,
			len(questions),
		),
	}

	for _, q := range questions {

		response.Questions = append(
			response.Questions,
			dto.QuestionResponse{
				ID:       q.ID,
				Type:     string(q.Type),
				Question: q.Question,
				Options:  q.Options,
				Position: q.Position,
			},
		)
	}

	return response, nil
}
func (s *Service) SubmitQuizAttempt(
	ctx context.Context,
	quizID string,
	req dto.SubmitQuizAttemptRequest,
) (*model.QuizAttempt, error) {

	if quizID == "" {
		return nil, fmt.Errorf("quiz_id is required")
	}

	if req.ChildID == "" {
		return nil, fmt.Errorf("child_id is required")
	}

	if len(req.Answers) == 0 {
		return nil, fmt.Errorf("answers are required")
	}

	// =========================
	// LOAD QUESTIONS
	// =========================

	questions, err := s.repo.GetQuestions(
		ctx,
		quizID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to load quiz questions: %w",
			err,
		)
	}

	if len(questions) == 0 {
		return nil, fmt.Errorf(
			"quiz has no questions",
		)
	}

	// =========================
	// INDEX QUESTIONS
	// =========================

	questionMap := make(
		map[string]model.QuizQuestion,
		len(questions),
	)

	for _, question := range questions {
		questionMap[question.ID] = question
	}

	// =========================
	// CALCULATE SCORE
	// =========================

	score := 0

	for _, submitted := range req.Answers {

		question, exists := questionMap[
			submitted.QuestionID,
		]

		if !exists {
			return nil, fmt.Errorf(
				"question %s does not belong to quiz",
				submitted.QuestionID,
			)
		}

		userAnswer := strings.TrimSpace(
			submitted.Answer,
		)

		correctAnswer := strings.TrimSpace(
			question.Answer,
		)

		if strings.EqualFold(
			userAnswer,
			correctAnswer,
		) {
			score++
		}
	}

	// =========================
	// CALCULATE RESULT
	// =========================

	totalQuestions := len(questions)

	percentage := 0.0

	if totalQuestions > 0 {
		percentage =
			(float64(score) /
				float64(totalQuestions)) *
				100
	}

	// =========================
	// CREATE ATTEMPT
	// =========================

	now := time.Now()

	attempt := &model.QuizAttempt{
		ID:             uuid.NewString(),
		QuizID:         quizID,
		ChildID:        req.ChildID,
		Score:          score,
		TotalQuestions: totalQuestions,
		Percentage:     percentage,
		CompletedAt:    now,
	}

	// =========================
	// SAVE ATTEMPT
	// =========================

	if err := s.repo.CreateAttempt(
		ctx,
		attempt,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to save quiz attempt: %w",
			err,
		)
	}

	return attempt, nil
}
func (s *Service) GetAttemptsByChild(
	ctx context.Context,
	quizID string,
	childID string,
) ([]model.QuizAttempt, error) {

	if quizID == "" {
		return nil, fmt.Errorf("quiz_id is required")
	}

	if childID == "" {
		return nil, fmt.Errorf("child_id is required")
	}

	attempts, err := s.repo.GetAttemptsByChild(
		ctx,
		quizID,
		childID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get quiz attempts: %w",
			err,
		)
	}

	return attempts, nil
}
func (s *Service) GetQuizResultsByChildID(
	ctx context.Context,
	childID string,
) ([]model.QuizResult, error) {

	if childID == "" {
		return nil, fmt.Errorf("child_id is required")
	}

	results, err := s.repo.GetQuizResultsByChildID(
		ctx,
		childID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get quiz results: %w",
			err,
		)
	}

	return results, nil
}