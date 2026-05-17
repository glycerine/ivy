package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/glycerine/ivy/goivy/webengine"
)

type EngineSessionStore struct {
	mu       sync.Mutex
	sessions map[string]EngineSessionRecord
	nextJob  int
}

type EngineSessionRecord struct {
	ID        string
	UserID    string
	ProjectID string
}

type engineSessionRequest struct {
	ProjectID string `json:"projectId"`
}

type engineModelDocument struct {
	ID             string `json:"id"`
	ProjectID      string `json:"projectId"`
	Filename       string `json:"filename"`
	Text           string `json:"text"`
	EngineRevision int64  `json:"engineRevision"`
}

type engineLoadRequest struct {
	Model engineModelDocument `json:"model"`
}

type engineCommandIntent struct {
	ID        string         `json:"id"`
	CommandID string         `json:"commandId"`
	Args      map[string]any `json:"args"`
}

type engineCommandRequest struct {
	Intent engineCommandIntent `json:"intent"`
}

type engineSessionResponse struct {
	ID           string             `json:"id"`
	Kind         string             `json:"kind"`
	Status       string             `json:"status"`
	ProjectID    string             `json:"projectId"`
	Capabilities engineCapabilities `json:"capabilities"`
	CreatedAt    string             `json:"createdAt"`
	UpdatedAt    string             `json:"updatedAt"`
}

type engineCapabilities struct {
	Offline        bool `json:"offline"`
	PersistentJobs bool `json:"persistentJobs"`
	CancelJob      bool `json:"cancelJob"`
	EventStream    bool `json:"eventStream"`
	ParallelJobs   bool `json:"parallelJobs"`
}

type engineJobResponse struct {
	ID            string `json:"id"`
	SessionID     string `json:"sessionId"`
	EngineID      string `json:"engineId"`
	ProjectID     string `json:"projectId"`
	ModelID       string `json:"modelId"`
	ModelRevision int64  `json:"modelRevision"`
	Kind          string `json:"kind"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

func NewEngineSessionStore() *EngineSessionStore {
	return &EngineSessionStore{sessions: map[string]EngineSessionRecord{}}
}

func (s *EngineSessionStore) Put(record EngineSessionRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[record.ID] = record
}

func (s *EngineSessionStore) Get(id string) (EngineSessionRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.sessions[id]
	return record, ok
}

func (s *EngineSessionStore) NextJobID(sessionID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextJob++
	return "job-" + sessionID + "-" + stringID(s.nextJob)
}

func (s *Server) engineNewSession(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req engineSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if s.cfg.DevMode && strings.HasPrefix(req.ProjectID, "project-local-") && s.projects.RoleForUser(userID, req.ProjectID) == "" {
		s.projects.GrantUser(userID, req.ProjectID, RoleOwner)
	}
	if s.projects.RoleForUser(userID, req.ProjectID) == "" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	session, err := s.engine.NewSession(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.engineSessions.Put(EngineSessionRecord{ID: session.ID, UserID: userID, ProjectID: req.ProjectID})
	writeOpaqueJSON(w, map[string]engineSessionResponse{"session": s.engineSessionResponse(session.ID, req.ProjectID)})
}

func (s *Server) engineLoadModel(w http.ResponseWriter, r *http.Request) {
	record, ok := s.authorizeEngineSession(w, r, true)
	if !ok {
		return
	}
	var req engineLoadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Model.ProjectID != "" && req.Model.ProjectID != record.ProjectID {
		http.Error(w, "project mismatch", http.StatusForbidden)
		return
	}
	if _, err := s.projects.SaveModelRevision(record.UserID, record.ProjectID, req.Model.Filename, req.Model.Text, 0); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if _, err := s.engine.LoadModel(r.Context(), record.ID, req.Model.Filename, []byte(req.Model.Text)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	job := s.engineJob(record, req.Model, "load", "succeeded")
	writeOpaqueJSON(w, map[string]engineJobResponse{"job": job})
}

func (s *Server) engineRunCommand(w http.ResponseWriter, r *http.Request) {
	record, ok := s.authorizeEngineSession(w, r, false)
	if !ok {
		return
	}
	var req engineCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	model := engineModelDocument{ID: "model-" + record.ProjectID, ProjectID: record.ProjectID}
	kind := commandJobKind(req.Intent.CommandID)
	job := s.engineJob(record, model, kind, "succeeded")
	response := map[string]any{"job": job}
	if strings.HasPrefix(req.Intent.CommandID, "check.") {
		mode := strings.TrimPrefix(req.Intent.CommandID, "check.")
		result, err := s.engine.Check(r.Context(), record.ID, webengine.CheckRequest{Mode: mode, Bound: intArg(req.Intent.Args["bound"])})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		response["result"] = result
	} else {
		http.Error(w, "unsupported engine command", http.StatusBadRequest)
		return
	}
	writeOpaqueJSON(w, response)
}

func (s *Server) engineSnapshot(w http.ResponseWriter, r *http.Request) {
	record, ok := s.authorizeEngineSession(w, r, false)
	if !ok {
		return
	}
	arg, err := s.engine.ARG(r.Context(), record.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	concept, err := s.engine.Concept(r.Context(), record.ID, webengine.ConceptRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeOpaqueJSON(w, map[string]any{"arg": arg, "concept": concept})
}

func (s *Server) authorizeEngineSession(w http.ResponseWriter, r *http.Request, write bool) (EngineSessionRecord, bool) {
	userID := requestUserID(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return EngineSessionRecord{}, false
	}
	sessionID := r.PathValue("session")
	record, ok := s.engineSessions.Get(sessionID)
	if !ok || record.UserID != userID {
		http.Error(w, "not found", http.StatusNotFound)
		return EngineSessionRecord{}, false
	}
	role := s.projects.RoleForUser(userID, record.ProjectID)
	if role == "" || (write && !canWrite(role)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return EngineSessionRecord{}, false
	}
	return record, true
}

func (s *Server) engineSessionResponse(sessionID, projectID string) engineSessionResponse {
	return engineSessionResponse{
		ID:        sessionID,
		Kind:      "hosted-go",
		Status:    "ready",
		ProjectID: projectID,
		Capabilities: engineCapabilities{
			Offline:        false,
			PersistentJobs: false,
			CancelJob:      false,
			EventStream:    false,
			ParallelJobs:   false,
		},
		CreatedAt: "2026-05-12T00:00:00.000Z",
		UpdatedAt: "2026-05-12T00:00:00.000Z",
	}
}

func (s *Server) engineJob(record EngineSessionRecord, model engineModelDocument, kind, status string) engineJobResponse {
	return engineJobResponse{
		ID:            s.engineSessions.NextJobID(record.ID),
		SessionID:     record.ID,
		EngineID:      record.ID,
		ProjectID:     record.ProjectID,
		ModelID:       model.ID,
		ModelRevision: model.EngineRevision,
		Kind:          kind,
		Status:        status,
		CreatedAt:     "2026-05-12T00:00:00.000Z",
		UpdatedAt:     "2026-05-12T00:00:00.000Z",
	}
}

func commandJobKind(commandID string) string {
	switch commandID {
	case "check.induction":
		return "check-induction"
	case "check.bounded":
		return "check-bounded"
	case "check.pdr":
		return "check-pdr"
	case "check.concrete":
		return "check-concrete"
	case "check.abstract":
		return "check-abstract"
	default:
		return "arg-action"
	}
}

func intArg(value any) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	default:
		return 0
	}
}
