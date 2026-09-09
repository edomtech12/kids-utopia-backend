package repository

import (
	"context"
	"fmt"

	"github.com/bellapacx/kids-utopia/internal/quiz/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VariantPage struct {
	PageNumber int
	Content    string
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) GetVariantContent(
	ctx context.Context,
	variantID string,
) (
	string,
	string,
	[]VariantPage,
	error,
) {

	var bookID string
	var title string

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			book_id,
			title
		FROM book_variants
		WHERE id = $1
		`,
		variantID,
	).Scan(
		&bookID,
		&title,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return "", "", nil, fmt.Errorf(
				"variant not found",
			)
		}

		return "", "", nil, err
	}

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			page_number,
			content
		FROM book_pages
		WHERE variant_id = $1
		ORDER BY page_number ASC
		`,
		variantID,
	)

	if err != nil {
		return "", "", nil, err
	}

	defer rows.Close()

	pages := make([]VariantPage, 0)

	for rows.Next() {

		var page VariantPage

		if err := rows.Scan(
			&page.PageNumber,
			&page.Content,
		); err != nil {
			return "", "", nil, err
		}

		pages = append(pages, page)
	}

	if err := rows.Err(); err != nil {
		return "", "", nil, err
	}

	return bookID, title, pages, nil
}
func (r *PostgresRepository) CreateQuestions(
	ctx context.Context,
	questions []model.QuizQuestion,
) error {

	if len(questions) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	for _, q := range questions {

		_, err := tx.Exec(
			ctx,
			`
			INSERT INTO quiz_questions (
				id,
				quiz_id,
				type,
				question,
				options,
				answer,
				explanation,
				position,
				created_at
			)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				NOW()
			)
			`,
			q.ID,
			q.QuizID,
			q.Type,
			q.Question,
			q.Options,
			q.Answer,
			q.Explanation,
			q.Position,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
func (r *PostgresRepository) CreateQuizWithQuestions(
	ctx context.Context,
	quiz *model.Quiz,
	questions []model.QuizQuestion,
) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin quiz transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	// =========================
	// CREATE QUIZ
	// =========================

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO quizzes (
			id,
			book_id,
			variant_id,
			title,
			status,
			question_count,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8
		)
		`,
		quiz.ID,
		quiz.BookID,
		quiz.VariantID,
		quiz.Title,
		quiz.Status,
		quiz.QuestionCount,
		quiz.CreatedAt,
		quiz.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("create quiz: %w", err)
	}

	// =========================
	// CREATE QUESTIONS
	// =========================

	for _, q := range questions {

		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO quiz_questions (
				id,
				quiz_id,
				type,
				question,
				options,
				answer,
				explanation,
				position,
				created_at
			)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9
			)
			`,
			q.ID,
			q.QuizID,
			q.Type,
			q.Question,
			q.Options, // PostgreSQL TEXT[]
			q.Answer,
			q.Explanation,
			q.Position,
			q.CreatedAt,
		)

		if err != nil {
			return fmt.Errorf(
				"create quiz question position=%d: %w",
				q.Position,
				err,
			)
		}
	}

	// =========================
	// COMMIT
	// =========================

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"commit quiz transaction: %w",
			err,
		)
	}

	return nil
}
func (r *PostgresRepository) GetQuizByVariant(
	ctx context.Context,
	variantID string,
) (*model.Quiz, []model.QuizQuestion, error) {

	var quiz model.Quiz

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			book_id,
			variant_id,
			title,
			status,
			question_count,
			created_at,
			updated_at
		FROM quizzes
		WHERE variant_id = $1
		  AND status = 'ready'
		ORDER BY created_at DESC
		LIMIT 1
		`,
		variantID,
	).Scan(
		&quiz.ID,
		&quiz.BookID,
		&quiz.VariantID,
		&quiz.Title,
		&quiz.Status,
		&quiz.QuestionCount,
		&quiz.CreatedAt,
		&quiz.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, fmt.Errorf("quiz not found")
		}

		return nil, nil, fmt.Errorf(
			"get quiz: %w",
			err,
		)
	}

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			quiz_id,
			type,
			question,
			options,
			answer,
			explanation,
			position,
			created_at
		FROM quiz_questions
		WHERE quiz_id = $1
		ORDER BY position ASC
		`,
		quiz.ID,
	)

	if err != nil {
		return nil, nil, fmt.Errorf(
			"get quiz questions: %w",
			err,
		)
	}

	defer rows.Close()

	questions := make(
		[]model.QuizQuestion,
		0,
		quiz.QuestionCount,
	)

	for rows.Next() {

		var question model.QuizQuestion

		err := rows.Scan(
			&question.ID,
			&question.QuizID,
			&question.Type,
			&question.Question,
			&question.Options,
			&question.Answer,
			&question.Explanation,
			&question.Position,
			&question.CreatedAt,
		)

		if err != nil {
			return nil, nil, fmt.Errorf(
				"scan quiz question: %w",
				err,
			)
		}

		questions = append(
			questions,
			question,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return &quiz, questions, nil
}
func (r *PostgresRepository) CreateQuiz(
	ctx context.Context,
	quiz *model.Quiz,
) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO quizzes (
			id,
			book_id,
			variant_id,
			title,
			status,
			question_count,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
		quiz.ID,
		quiz.BookID,
		quiz.VariantID,
		quiz.Title,
		quiz.Status,
		quiz.QuestionCount,
		quiz.CreatedAt,
		quiz.UpdatedAt,
	)

	return err
}
func (r *PostgresRepository) CreateAttempt(
	ctx context.Context,
	attempt *model.QuizAttempt,
) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO quiz_attempts (
			id,
			quiz_id,
			child_id,
			score,
			total_questions,
			percentage,
			completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		attempt.ID,
		attempt.QuizID,
		attempt.ChildID,
		attempt.Score,
		attempt.TotalQuestions,
		attempt.Percentage,
		attempt.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to create quiz attempt: %w",
			err,
		)
	}

	return nil
}
func (r *PostgresRepository) GetQuestions(
	ctx context.Context,
	quizID string,
) ([]model.QuizQuestion, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			quiz_id,
			type,
			question,
			options,
			answer,
			explanation,
			position,
			created_at
		FROM quiz_questions
		WHERE quiz_id = $1
		ORDER BY position ASC
		`,
		quizID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to query quiz questions: %w",
			err,
		)
	}

	defer rows.Close()

	questions := make(
		[]model.QuizQuestion,
		0,
	)

	for rows.Next() {
		var question model.QuizQuestion

		if err := rows.Scan(
			&question.ID,
			&question.QuizID,
			&question.Type,
			&question.Question,
			&question.Options,
			&question.Answer,
			&question.Explanation,
			&question.Position,
			&question.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan quiz question: %w",
				err,
			)
		}

		questions = append(
			questions,
			question,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate quiz questions: %w",
			err,
		)
	}

	return questions, nil
}
func (r *PostgresRepository) GetAttemptsByChild(
	ctx context.Context,
	quizID string,
	childID string,
) ([]model.QuizAttempt, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			quiz_id,
			child_id,
			score,
			total_questions,
			percentage,
			completed_at
		FROM quiz_attempts
		WHERE quiz_id = $1
		  AND child_id = $2
		ORDER BY completed_at DESC
		`,
		quizID,
		childID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to query quiz attempts: %w",
			err,
		)
	}

	defer rows.Close()

	attempts := make(
		[]model.QuizAttempt,
		0,
	)

	for rows.Next() {
		var attempt model.QuizAttempt

		if err := rows.Scan(
			&attempt.ID,
			&attempt.QuizID,
			&attempt.ChildID,
			&attempt.Score,
			&attempt.TotalQuestions,
			&attempt.Percentage,
			&attempt.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan quiz attempt: %w",
				err,
			)
		}

		attempts = append(
			attempts,
			attempt,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate quiz attempts: %w",
			err,
		)
	}

	return attempts, nil
}
func (r *PostgresRepository) GetQuizResultsByChildID(
    ctx context.Context,
    childID string,
) ([]model.QuizResult, error) {

    if childID == "" {
        return nil, fmt.Errorf("child_id is required")
    }

    const query = `
        SELECT
            id,
            quiz_id,
            child_id,
            score,
            total_questions,
            percentage,
            completed_at
        FROM quiz_attempts
        WHERE child_id = $1
        ORDER BY completed_at DESC
    `

    rows, err := r.db.Query(ctx, query, childID)
    if err != nil {
        return nil, fmt.Errorf(
            "failed to get quiz results for child: %w",
            err,
        )
    }
    defer rows.Close()

    results := make([]model.QuizResult, 0)

    for rows.Next() {

        var result model.QuizResult

        err := rows.Scan(
            &result.ID,
            &result.QuizID,
            &result.ChildID,
            &result.Score,
            &result.TotalQuestions,
            &result.Percentage,
            &result.CompletedAt,
        )

        if err != nil {
            return nil, fmt.Errorf(
                "failed to scan quiz result: %w",
                err,
            )
        }

        results = append(results, result)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf(
            "failed while reading quiz results: %w",
            err,
        )
    }

    return results, nil
}