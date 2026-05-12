package mail

import (
	"context"

	mailgun "github.com/mailgun/mailgun-go/v4"
)

type Sender interface {
	Send(ctx context.Context, to, subject, text string) error
}

type NoopSender struct{}

func (NoopSender) Send(context.Context, string, string, string) error { return nil }

type MailgunSender struct {
	from string
	mg   *mailgun.MailgunImpl
}

func NewMailgunSender(domain, apiKey, from string) *MailgunSender {
	return &MailgunSender{
		from: from,
		mg:   mailgun.NewMailgun(domain, apiKey),
	}
}

func (s *MailgunSender) Send(ctx context.Context, to, subject, text string) error {
	msg := s.mg.NewMessage(s.from, subject, text, to)
	_, _, err := s.mg.Send(ctx, msg)
	return err
}
