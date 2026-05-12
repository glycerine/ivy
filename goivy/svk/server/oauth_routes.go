package server

import (
	"errors"
	"net/http"
	"strings"
)

const oauthStateCookie = "ivysvk_oauth_state"

func (s *Server) oauthStart(w http.ResponseWriter, r *http.Request) {
	provider := strings.TrimPrefix(r.URL.Path, "/auth/oauth/")
	provider = strings.TrimSuffix(provider, "/start")
	state, _, authURL := s.oauth.Start(provider)
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/auth/oauth/" + provider,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   !s.cfg.DevMode,
	})
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *Server) oauthCallback(w http.ResponseWriter, r *http.Request) {
	provider := strings.TrimPrefix(r.URL.Path, "/auth/oauth/")
	provider = strings.TrimSuffix(provider, "/callback")
	cookie, err := r.Cookie(oauthStateCookie)
	if err != nil || cookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}
	result, err := s.oauth.Callback(provider, r.URL.Query().Get("state"), r.URL.Query().Get("code"), r.URL.Query().Get("email"))
	if errors.Is(err, ErrExplicitLinkRequired) {
		http.Error(w, "explicit identity linking required", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "oauth callback failed", http.StatusBadRequest)
		return
	}
	_ = result
	s.setSessionCookies(w, r, "")
	http.Redirect(w, r, "/app", http.StatusFound)
}
