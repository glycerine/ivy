package webui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// writeBackendErr writes an error response, using 404 for session-not-found
// and 400 for all other backend errors.
func writeBackendErr(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrSessionNotFound) {
		writeErr(w, http.StatusNotFound, err.Error())
	} else {
		writeErr(w, http.StatusBadRequest, err.Error())
	}
}

// apiNewSession handles POST /api/session/new.
func (s *Server) apiNewSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	data, err := s.backend.NewSession(s.cfg)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeBackend(w, data)
}

// apiLoad handles POST /api/session/{id}/load.
// Accepts either a multipart file upload (field name "file") from the browser,
// or a JSON body with {"path": "/some/file.ivy"} for programmatic use.
func (s *Server) apiLoad(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	ct := r.Header.Get("Content-Type")

	// Multipart file upload from the browser.
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
			writeErr(w, http.StatusBadRequest, "invalid multipart form: "+err.Error())
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			writeErr(w, http.StatusBadRequest, "missing file field: "+err.Error())
			return
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "reading file: "+err.Error())
			return
		}
		data, err := s.backend.Load(sessionID, header.Filename, content)
		if err != nil {
			writeBackendErr(w, err)
			return
		}
		writeBackend(w, data)
		return
	}

	// JSON body with a file path (programmatic use).
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	data, err := s.backend.LoadPath(sessionID, req.Path)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiAction handles POST /api/session/{id}/action.
func (s *Server) apiAction(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Action string                 `json:"action"`
		Args   map[string]interface{} `json:"args"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	data, err := s.backend.Action(sessionID, req.Action, req.Args)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiARG handles GET /api/session/{id}/arg.
func (s *Server) apiARG(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	data, err := s.backend.GetARG(sessionID)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiConcept handles GET /api/session/{id}/concept.
func (s *Server) apiConcept(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	data, err := s.backend.GetConcept(sessionID)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiConceptSplit handles POST /api/session/{id}/concept/split.
func (s *Server) apiConceptSplit(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Concept string `json:"concept"`
		SplitBy string `json:"split_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	data, err := s.backend.ConceptSplit(sessionID, req.Concept, req.SplitBy)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiConceptEmpty handles POST /api/session/{id}/concept/empty.
func (s *Server) apiConceptEmpty(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Concept string `json:"concept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	data, err := s.backend.ConceptEmpty(sessionID, req.Concept)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiConceptRemove handles POST /api/session/{id}/concept/remove.
func (s *Server) apiConceptRemove(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Concept string `json:"concept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	data, err := s.backend.ConceptRemove(sessionID, req.Concept)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiConceptUndo handles POST /api/session/{id}/concept/undo.
func (s *Server) apiConceptUndo(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	data, err := s.backend.ConceptUndo(sessionID)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiConceptMaterialize handles POST /api/session/{id}/concept/materialize.
func (s *Server) apiConceptMaterialize(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Concept string `json:"concept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	data, err := s.backend.ConceptMaterialize(sessionID, req.Concept)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiToggles handles GET/POST /api/session/{id}/toggles.
// GET returns current toggle state; POST updates it.
func (s *Server) apiToggles(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method == http.MethodGet {
		data, err := s.backend.GetToggles(sessionID)
		if err != nil {
			writeBackendErr(w, err)
			return
		}
		writeBackend(w, data)
		return
	}
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "GET or POST required")
		return
	}
	var req struct {
		Edge         string `json:"edge"`
		DisplayClass string `json:"display_class"`
		Value        bool   `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	data, err := s.backend.SetToggle(sessionID, req.Edge, req.DisplayClass, req.Value)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiConceptReset handles POST /api/session/{id}/concept/reset.
func (s *Server) apiConceptReset(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	data, err := s.backend.ConceptReset(sessionID)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiConceptDiagram handles POST /api/session/{id}/concept/diagram.
func (s *Server) apiConceptDiagram(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	data, err := s.backend.ConceptDiagram(sessionID)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiProof handles GET /api/session/{id}/proof.
func (s *Server) apiProof(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	data, err := s.backend.GetProof(sessionID)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiConceptProjection handles POST /api/session/{id}/concept/projection.
func (s *Server) apiConceptProjection(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Name    string `json:"name"`
		Concept string `json:"concept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	data, err := s.backend.ConceptProjection(sessionID, req.Name, req.Concept)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiArgAction handles POST /api/session/{id}/arg/action.
func (s *Server) apiArgAction(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Node   string                 `json:"node"`
		Action string                 `json:"action"`
		Args   map[string]interface{} `json:"args"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	data, err := s.backend.ArgAction(sessionID, req.Node, req.Action, req.Args)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiProofAction handles POST /api/session/{id}/proof/action.
func (s *Server) apiProofAction(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Goal   string `json:"goal"`
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	data, err := s.backend.ProofAction(sessionID, req.Goal, req.Action)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiSave handles GET /api/session/{id}/save.
func (s *Server) apiSave(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	data, err := s.backend.Save(sessionID)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=ivy_session.json")
	w.Write(data)
}

// apiCheck handles POST /api/session/{id}/check.
// Dispatches to the appropriate verification mode via check/art/updr packages.
func (s *Server) apiCheck(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to the current mode if no body
		req.Mode = "induction"
	}
	data, err := s.backend.Check(sessionID, req.Mode)
	if err != nil {
		writeBackendErr(w, err)
		return
	}
	writeBackend(w, data)
}

// apiEvents handles GET /api/session/{id}/events — SSE stream.
func (s *Server) apiEvents(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET required")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	events, err := s.backend.Events(sessionID)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-events:
			if !ok {
				return
			}
			data, _ := json.Marshal(evt)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
