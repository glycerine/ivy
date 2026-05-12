package server

import (
	"encoding/json"
	"net/http"
)

type passkeyRequest struct {
	UserID       string `json:"userId"`
	Challenge    string `json:"challenge,omitempty"`
	CredentialID string `json:"credentialId,omitempty"`
}

func (s *Server) passkeyRegistrationOptions(w http.ResponseWriter, r *http.Request) {
	var req passkeyRequest
	if !decodePasskeyJSON(w, r, &req) {
		return
	}
	writeOpaqueJSON(w, map[string]string{"challenge": s.passkeys.RegistrationOptions(req.UserID)})
}

func (s *Server) passkeyRegistrationFinish(w http.ResponseWriter, r *http.Request) {
	var req passkeyRequest
	if !decodePasskeyJSON(w, r, &req) {
		return
	}
	if err := s.passkeys.FinishRegistration(req.UserID, req.Challenge, req.CredentialID); err != nil {
		http.Error(w, "passkey registration failed", http.StatusBadRequest)
		return
	}
	writeOpaqueJSON(w, map[string]bool{"ok": true})
}

func (s *Server) passkeyLoginOptions(w http.ResponseWriter, r *http.Request) {
	var req passkeyRequest
	if !decodePasskeyJSON(w, r, &req) {
		return
	}
	writeOpaqueJSON(w, map[string]string{"challenge": s.passkeys.LoginOptions(req.UserID)})
}

func (s *Server) passkeyLoginFinish(w http.ResponseWriter, r *http.Request) {
	var req passkeyRequest
	if !decodePasskeyJSON(w, r, &req) {
		return
	}
	if err := s.passkeys.FinishLogin(req.UserID, req.Challenge, req.CredentialID); err != nil {
		http.Error(w, "passkey login failed", http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    randomURLToken(),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   !s.cfg.DevMode,
	})
	writeOpaqueJSON(w, map[string]bool{"ok": true})
}

func decodePasskeyJSON(w http.ResponseWriter, r *http.Request, out *passkeyRequest) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return false
	}
	return true
}
