package server

import (
	"encoding/json"
	"net/http"
)

type magicRequest struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}

func (s *Server) magicRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req magicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	_ = s.magic.Request(r.Context(), req.Email, req.Purpose)
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
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    randomURLToken(),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   !s.cfg.DevMode,
	})
	http.Redirect(w, r, "/app", http.StatusFound)
}
