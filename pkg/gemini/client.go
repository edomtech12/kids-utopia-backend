package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/genai"
)

type Client struct {
	client *genai.Client
	model  string
}

func NewClient(ctx context.Context) (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		client: client,
		model:  "gemini-3.6-flash",
	}, nil
}

type QuizRequest struct {
	Content             string
	MultipleChoiceCount int
	FillBlankCount      int
}

type QuizResponse struct {
	Title     string              `json:"title"`
	Questions []GeneratedQuestion `json:"questions"`
}

type GeneratedQuestion struct {
	Type        string   `json:"type"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	Answer      string   `json:"answer"`
	Explanation string   `json:"explanation"`
}

func (c *Client) GenerateQuiz(
	ctx context.Context,
	req QuizRequest,
) (*QuizResponse, error) {

	if req.Content == "" {
		return nil, fmt.Errorf("quiz content is empty")
	}

	if req.MultipleChoiceCount < 1 {
		return nil, fmt.Errorf(
			"multiple choice count must be at least 1",
		)
	}

	if req.FillBlankCount < 1 {
		return nil, fmt.Errorf(
			"fill blank count must be at least 1",
		)
	}

	prompt := buildQuizPrompt(req)

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",

		ResponseJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title": map[string]any{
					"type": "string",
				},
				"questions": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"type": map[string]any{
								"type": "string",
								"enum": []string{
									"multiple_choice",
									"fill_blank",
								},
							},
							"question": map[string]any{
								"type": "string",
							},
							"options": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "string",
								},
							},
							"answer": map[string]any{
								"type": "string",
							},
							"explanation": map[string]any{
								"type": "string",
							},
						},
						"required": []string{
							"type",
							"question",
							"options",
							"answer",
							"explanation",
						},
					},
				},
			},
			"required": []string{
				"title",
				"questions",
			},
		},
	}

	response, err := c.client.Models.GenerateContent(
		ctx,
		c.model,
		genai.Text(prompt),
		config,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"gemini quiz generation failed: %w",
			err,
		)
	}

	if response == nil {
		return nil, fmt.Errorf(
			"gemini returned empty response",
		)
	}

	content := response.Text()

	if content == "" {
		return nil, fmt.Errorf(
			"gemini returned empty content",
		)
	}

	var result QuizResponse

	if err := json.Unmarshal(
		[]byte(content),
		&result,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to parse Gemini quiz JSON: %w; content=%s",
			err,
			content,
		)
	}

	if err := validateQuizResponse(
		&result,
		req.MultipleChoiceCount,
		req.FillBlankCount,
	); err != nil {
		return nil, err
	}

	return &result, nil
}

func buildQuizPrompt(req QuizRequest) string {
	return fmt.Sprintf(`
You are an educational quiz generator for children.

Create a quiz from the children's story below.

Generate exactly:

- %d multiple-choice questions
- %d fill-in-the-blank questions

Rules:

1. Every question MUST be answerable from the provided story.
2. Do not invent facts.
3. Use simple, child-friendly language.
4. Avoid duplicate questions.
5. Multiple-choice questions MUST have exactly 4 options.
6. Multiple-choice questions MUST have exactly one correct answer.
7. The answer must exactly match one of the options.
8. Fill-in-the-blank questions must have one clear answer.
9. The blank should replace an important word or short phrase from the story.
10. Keep questions appropriate for children.
11. Return ONLY valid JSON.
12. Do not include markdown.
13. Do not include text outside the JSON object.

Return this exact structure:

{
  "title": "Quiz title",
  "questions": [
    {
      "type": "multiple_choice",
      "question": "Question text",
      "options": [
        "Option 1",
        "Option 2",
        "Option 3",
        "Option 4"
      ],
      "answer": "Correct option",
      "explanation": "Short explanation"
    },
    {
      "type": "fill_blank",
      "question": "The boy went to the ____.",
      "options": [],
      "answer": "forest",
      "explanation": "The story says that the boy went to the forest."
    }
  ]
}

STORY:

%s
`,
		req.MultipleChoiceCount,
		req.FillBlankCount,
		req.Content,
	)
}

func validateQuizResponse(
	result *QuizResponse,
	expectedMultipleChoice int,
	expectedFillBlank int,
) error {

	if result.Title == "" {
		return fmt.Errorf(
			"generated quiz has no title",
		)
	}

	expectedTotal :=
		expectedMultipleChoice +
			expectedFillBlank

	if len(result.Questions) != expectedTotal {
		return fmt.Errorf(
			"expected %d questions, got %d",
			expectedTotal,
			len(result.Questions),
		)
	}

	multipleChoiceCount := 0
	fillBlankCount := 0

	for i, q := range result.Questions {

		if q.Question == "" {
			return fmt.Errorf(
				"question %d has empty question text",
				i+1,
			)
		}

		switch q.Type {

		case "multiple_choice":

			multipleChoiceCount++

			if len(q.Options) != 4 {
				return fmt.Errorf(
					"multiple choice question %d must have exactly 4 options",
					i+1,
				)
			}

			if q.Answer == "" {
				return fmt.Errorf(
					"multiple choice question %d has no answer",
					i+1,
				)
			}

			found := false

			for _, option := range q.Options {
				if option == q.Answer {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf(
					"answer for question %d does not match any option",
					i+1,
				)
			}

		case "fill_blank":

			fillBlankCount++

			if q.Answer == "" {
				return fmt.Errorf(
					"fill blank question %d has no answer",
					i+1,
				)
			}

		default:

			return fmt.Errorf(
				"invalid question type %q at question %d",
				q.Type,
				i+1,
			)
		}
	}

	if multipleChoiceCount != expectedMultipleChoice {
		return fmt.Errorf(
			"expected %d multiple choice questions, got %d",
			expectedMultipleChoice,
			multipleChoiceCount,
		)
	}

	if fillBlankCount != expectedFillBlank {
		return fmt.Errorf(
			"expected %d fill blank questions, got %d",
			expectedFillBlank,
			fillBlankCount,
		)
	}

	return nil
}

