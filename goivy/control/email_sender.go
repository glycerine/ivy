package control

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type EmailMessage struct {
	ToEmail   string    `json:"toEmail"`
	LoginURL  string    `json:"loginUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type DatabaseEmailSender struct {
	Store *PostgresStore
}

func NewDatabaseEmailSender(store *PostgresStore) DatabaseEmailSender {
	return DatabaseEmailSender{Store: store}
}

func (s DatabaseEmailSender) SendLoginLink(ctx context.Context, toEmail, loginURL string, expiresAt time.Time) error {
	if s.Store == nil {
		return errors.New("database email sender requires a store")
	}
	return s.Store.RecordEmailDelivery(ctx, toEmail, loginURL, expiresAt, time.Now().UTC())
}

type CompositeEmailSender struct {
	Senders []EmailSender
}

func (s CompositeEmailSender) SendLoginLink(ctx context.Context, toEmail, loginURL string, expiresAt time.Time) error {
	if len(s.Senders) == 0 {
		return errors.New("composite email sender requires at least one sender")
	}
	for _, sender := range s.Senders {
		if sender == nil {
			continue
		}
		if err := sender.SendLoginLink(ctx, toEmail, loginURL, expiresAt); err != nil {
			return err
		}
	}
	return nil
}

type MailgunEmailSender struct {
	Domain string
	APIKey string
	From   string
}

func (s *MailgunEmailSender) SendLoginLink(ctx context.Context, toEmail, loginURL string, expiresAt time.Time) error {
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
	ex := expiresAt.Format(time.RFC1123)
	message := mg.NewMessage(
		from,
		"Your Ivy sign-in link",
		fmt.Sprintf("Use this link to sign in to Ivy. It expires at %s.\n\n%s", ex, loginURL),
		toEmail,
	)
	_, _, err := mg.Send(ctx, message)

	vv("end of MailgunEmailSender.SendLoginLink(): mailgun.Send() returned err = '%v'; message='%v' toEmail='%v'; loginURL = '%v'; expiresAt='%v'", err, message, toEmail, loginURL, ex)
	return err
}

func tokenHashFromLoginURL(loginURL string) ([]byte, error) {
	parsed, err := url.Parse(loginURL)
	if err != nil {
		return nil, err
	}
	values, err := url.ParseQuery(parsed.Fragment)
	if err != nil {
		return nil, err
	}
	token := values.Get("token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("login URL is missing token fragment")
	}
	return hashToken(token), nil
}
