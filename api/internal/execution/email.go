package execution

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// EmailPayload represents an email that would be sent
type EmailPayload struct {
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Timestamp time.Time `json:"timestamp"`
}

// InMemoryEmailService tracks email payloads in memory without actually sending them
type InMemoryEmailService struct {
	sentEmails []EmailPayload
}

// NewInMemoryEmailService creates a new in-memory email service
func NewInMemoryEmailService() *InMemoryEmailService {
	return &InMemoryEmailService{
		sentEmails: make([]EmailPayload, 0),
	}
}

// SendEmail tracks the email payload in memory
func (s *InMemoryEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	// Validate email parameters
	if to == "" {
		return fmt.Errorf("recipient email is required")
	}
	
	if subject == "" {
		return fmt.Errorf("email subject is required")
	}
	
	// Create email payload
	payload := EmailPayload{
		To:        to,
		Subject:   subject,
		Body:      body,
		Timestamp: time.Now(),
	}
	
	// Store in memory
	s.sentEmails = append(s.sentEmails, payload)
	
	// Log to avoid unused variable warnings and for visibility
	slog.Info("Email payload tracked in memory", 
		"to", payload.To,
		"subject", payload.Subject,
		"body", payload.Body,
		"timestamp", payload.Timestamp,
		"totalEmails", len(s.sentEmails),
	)
	
	return nil
}

// GetSentEmails returns all tracked email payloads (for testing/debugging)
func (s *InMemoryEmailService) GetSentEmails() []EmailPayload {
	return s.sentEmails
}

// ClearSentEmails clears all tracked email payloads
func (s *InMemoryEmailService) ClearSentEmails() {
	s.sentEmails = make([]EmailPayload, 0)
}