package testsupport

import "sync"

type Email struct {
	To      string
	Subject string
	Body    string
}

type EmailSink struct {
	mu     sync.Mutex
	emails []Email
}

func (s *EmailSink) Send(email Email) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emails = append(s.emails, email)
}

func (s *EmailSink) Emails() []Email {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Email, len(s.emails))
	copy(out, s.emails)
	return out
}

func (s *EmailSink) LastTo(to string) (Email, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.emails) - 1; i >= 0; i-- {
		if s.emails[i].To == to {
			return s.emails[i], true
		}
	}
	return Email{}, false
}
