package control

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type EmailMessage struct {
	ToEmail   string    `json:"toEmail"`
	LoginURL  string    `json:"loginUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type MemoryEmailSender struct {
	mu       sync.Mutex
	messages []EmailMessage
}

func NewMemoryEmailSender() *MemoryEmailSender {
	return &MemoryEmailSender{}
}

func (s *MemoryEmailSender) SendLoginLink(ctx context.Context, toEmail, loginURL string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, EmailMessage{
		ToEmail:   toEmail,
		LoginURL:  loginURL,
		ExpiresAt: expiresAt,
	})
	return nil
}

func (s *MemoryEmailSender) Messages() []EmailMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]EmailMessage(nil), s.messages...)
}

func (s *MemoryEmailSender) LatestFor(email string) (EmailMessage, bool) {
	email = strings.ToLower(strings.TrimSpace(email))
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.messages) - 1; i >= 0; i-- {
		if s.messages[i].ToEmail == email {
			return s.messages[i], true
		}
	}
	return EmailMessage{}, false
}

type MailgunEmailSender struct {
	Domain string
	APIKey string
	From   string
}

func (s MailgunEmailSender) SendLoginLink(ctx context.Context, toEmail, loginURL string, expiresAt time.Time) error {
	if strings.TrimSpace(s.Domain) == "" {
		return errors.New("mailgun domain is required")
	}
	if strings.TrimSpace(s.APIKey) == "" {
		return errors.New("mailgun api key is required")
	}
	from := s.From
	if strings.TrimSpace(from) == "" {
		from = "Ivy <postmaster@" + s.Domain + ">"
	}
	mg := mailgun.NewMailgun(s.Domain, s.APIKey)
	message := mg.NewMessage(
		from,
		"Your Ivy sign-in link",
		fmt.Sprintf("Use this link to sign in to Ivy. It expires at %s.\n\n%s", expiresAt.Format(time.RFC1123), loginURL),
		toEmail,
	)
	_, _, err := mg.Send(ctx, message)
	return err
}
