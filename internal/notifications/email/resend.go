package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Resend struct {
	apiKey string
	from   string
}

func NewResend(apiKey, from string) *Resend {
	return &Resend{
		apiKey: apiKey,
		from:   from,
	}
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

type resendError struct {
	Message string `json:"message"`
}

func (r *Resend) Send(to string, subject string, html string) error {
	payload := resendRequest{
		From:    r.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal resend request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.resend.com/emails",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("failed to create resend request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email through resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var resendErr resendError

		if err := json.NewDecoder(resp.Body).Decode(&resendErr); err == nil {
			return fmt.Errorf(
				"resend API error: %s",
				resendErr.Message,
			)
		}

		return fmt.Errorf(
			"resend API returned status code %d",
			resp.StatusCode,
		)
	}

	return nil
}