package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

type createProjectRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type saveModelRequest struct {
	Filename     string `json:"filename"`
	Text         string `json:"text"`
	BaseRevision int64  `json:"baseRevision"`
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeOpaqueJSON(w, map[string][]ProjectRecord{"projects": s.projects.AccessibleProjects(userID)})
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	defer r.Body.Close()
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	project := s.projects.CreatePersonalProject(userID, req.Name, req.Slug)
	writeOpaqueJSON(w, map[string]ProjectRecord{"project": project})
}

func (s *Server) saveProjectModel(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	projectID := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	projectID = strings.TrimSuffix(projectID, "/models")
	defer r.Body.Close()
	var req saveModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	model, err := s.projects.SaveModelRevision(userID, projectID, req.Filename, req.Text, req.BaseRevision)
	if err != nil {
		if err.Error() == "stale model revision" {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	writeOpaqueJSON(w, map[string]HostedModel{"model": model})
}

func requestUserID(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	if strings.HasPrefix(cookie.Value, "user:") {
		return strings.TrimPrefix(cookie.Value, "user:")
	}
	return "dev-user"
}
