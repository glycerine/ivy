package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

type magicRequest struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}

func (s *Server) magicRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req magicRequest
	if strings.HasPrefix(r.Header.Get("content-type"), "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		req.Email = r.Form.Get("email")
		req.Purpose = r.Form.Get("purpose")
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
	}
	_ = s.magic.Request(r.Context(), req.Email, req.Purpose)
	if strings.HasPrefix(r.Header.Get("content-type"), "application/x-www-form-urlencoded") {
		renderPage(w, "Check your email", "If the address can receive SVK login email, a magic link is on the way.")
		return
	}
	w.Header().Set("content-type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *Server) magicConsume(w http.ResponseWriter, r *http.Request) {
	email, err := s.magic.Consume(r.URL.Query().Get("token"))
	if err != nil {
		http.Error(w, "invalid or expired magic link", http.StatusUnauthorized)
		return
	}
	_ = email
	s.setSessionCookies(w, r, "")
	http.Redirect(w, r, "/app", http.StatusFound)
}
