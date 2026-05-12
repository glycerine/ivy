package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
)

type opaqueStartRequest struct {
	UserID   string `json:"userId"`
	Message  string `json:"message"`
	Password string `json:"password,omitempty"`
}

type opaqueFinishRequest struct {
	FlowID   string `json:"flowId"`
	Message  string `json:"message"`
	Password string `json:"password,omitempty"`
}

type opaqueMessageResponse struct {
	FlowID   string `json:"flowId,omitempty"`
	Message  string `json:"message,omitempty"`
	Complete bool   `json:"complete,omitempty"`
}

func (s *Server) opaqueRegistrationStart(w http.ResponseWriter, r *http.Request) {
	var req opaqueStartRequest
	if !decodeOpaqueJSON(w, r, &req) {
		return
	}
	message, err := base64.StdEncoding.DecodeString(req.Message)
	if err != nil {
		http.Error(w, "invalid opaque message", http.StatusBadRequest)
		return
	}
	flowID, response, err := s.opaque.RegistrationStart(req.UserID, message)
	if err != nil {
		http.Error(w, "opaque registration failed", http.StatusBadRequest)
		return
	}
	writeOpaqueJSON(w, opaqueMessageResponse{FlowID: flowID, Message: base64.StdEncoding.EncodeToString(response)})
}

func (s *Server) opaqueRegistrationFinish(w http.ResponseWriter, r *http.Request) {
	var req opaqueFinishRequest
	if !decodeOpaqueJSON(w, r, &req) {
		return
	}
	message, err := base64.StdEncoding.DecodeString(req.Message)
	if err != nil {
		http.Error(w, "invalid opaque message", http.StatusBadRequest)
		return
	}
	if err := s.opaque.RegistrationFinish(req.FlowID, message); err != nil {
		http.Error(w, "opaque registration failed", http.StatusBadRequest)
		return
	}
	writeOpaqueJSON(w, opaqueMessageResponse{Complete: true})
}

func (s *Server) opaqueLoginStart(w http.ResponseWriter, r *http.Request) {
	var req opaqueStartRequest
	if !decodeOpaqueJSON(w, r, &req) {
		return
	}
	message, err := base64.StdEncoding.DecodeString(req.Message)
	if err != nil {
		http.Error(w, "invalid opaque message", http.StatusBadRequest)
		return
	}
	flowID, response, err := s.opaque.LoginStart(req.UserID, message)
	if err != nil {
		http.Error(w, "opaque login failed", http.StatusBadRequest)
		return
	}
	writeOpaqueJSON(w, opaqueMessageResponse{FlowID: flowID, Message: base64.StdEncoding.EncodeToString(response)})
}

func (s *Server) opaqueLoginFinish(w http.ResponseWriter, r *http.Request) {
	var req opaqueFinishRequest
	if !decodeOpaqueJSON(w, r, &req) {
		return
	}
	message, err := base64.StdEncoding.DecodeString(req.Message)
	if err != nil {
		http.Error(w, "invalid opaque message", http.StatusBadRequest)
		return
	}
	if _, err := s.opaque.LoginFinish(req.FlowID, message); err != nil {
		http.Error(w, "opaque login failed", http.StatusUnauthorized)
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
	writeOpaqueJSON(w, opaqueMessageResponse{Complete: true})
}

func decodeOpaqueJSON(w http.ResponseWriter, r *http.Request, out interface{}) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return false
	}
	switch req := out.(type) {
	case *opaqueStartRequest:
		if req.Password != "" {
			http.Error(w, "raw passwords are not accepted by OPAQUE endpoints", http.StatusBadRequest)
			return false
		}
	case *opaqueFinishRequest:
		if req.Password != "" {
			http.Error(w, "raw passwords are not accepted by OPAQUE endpoints", http.StatusBadRequest)
			return false
		}
	}
	return true
}

func writeOpaqueJSON(w http.ResponseWriter, value interface{}) {
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
