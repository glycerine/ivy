package webui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
)

var sessionCounter uint64

// apiNewSession handles POST /api/session/new.
func (s *Server) apiNewSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	id := fmt.Sprintf("s%d", atomic.AddUint64(&sessionCounter, 1))
	sess := NewSession(id)

	s.mu.Lock()
	s.sessions[id] = sess
	s.mu.Unlock()

	writeJSON(w, map[string]string{"session_id": id})
}

// apiLoad handles POST /api/session/{id}/load.
// Accepts either a multipart file upload (field name "file") from the browser,
// or a JSON body with {"path": "/some/file.ivy"} for programmatic use.
func (s *Server) apiLoad(w http.ResponseWriter, r *http.Request, sess *Session) {
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
		data, err := io.ReadAll(file)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "reading file: "+err.Error())
			return
		}
		if err := sess.LoadFileContent(header.Filename, data); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, map[string]string{"status": "ok", "filename": header.Filename})
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
	if err := sess.LoadFile(req.Path); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiAction handles POST /api/session/{id}/action.
func (s *Server) apiAction(w http.ResponseWriter, r *http.Request, sess *Session) {
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
	if err := sess.ExecuteAction(req.Action, req.Args); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiARG handles GET /api/session/{id}/arg.
func (s *Server) apiARG(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	cy := RenderARG(sess.Graph)
	writeJSON(w, cy)
}

// apiConcept handles GET /api/session/{id}/concept.
func (s *Server) apiConcept(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	cy := RenderConceptGraph(sess.ConceptSess, nil)
	// Include relation names for the state checkbox panel.
	// Matches Python: relation_ids = edges + node_labels.
	writeJSON(w, map[string]interface{}{
		"elements":    cy.Elements,
		"relations":   sess.ConceptSess.RelationNames(),
		"edges":       sess.ConceptSess.Domain.Edges,
		"node_labels": sess.ConceptSess.Domain.NodeLabels,
		"nodes":       sess.ConceptSess.Domain.Nodes,
	})
}

// apiConceptSplit handles POST /api/session/{id}/concept/split.
func (s *Server) apiConceptSplit(w http.ResponseWriter, r *http.Request, sess *Session) {
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
	if err := sess.ConceptSess.Split(req.Concept, req.SplitBy); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiConceptEmpty handles POST /api/session/{id}/concept/empty.
func (s *Server) apiConceptEmpty(w http.ResponseWriter, r *http.Request, sess *Session) {
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
	if err := sess.ConceptSess.SupposeEmpty(req.Concept); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiConceptRemove handles POST /api/session/{id}/concept/remove.
func (s *Server) apiConceptRemove(w http.ResponseWriter, r *http.Request, sess *Session) {
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
	if err := sess.ConceptSess.RemoveConcept(req.Concept); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiConceptUndo handles POST /api/session/{id}/concept/undo.
func (s *Server) apiConceptUndo(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	if err := sess.ConceptSess.Undo(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiConceptMaterialize handles POST /api/session/{id}/concept/materialize.
func (s *Server) apiConceptMaterialize(w http.ResponseWriter, r *http.Request, sess *Session) {
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
	if err := sess.ConceptSess.Materialize(req.Concept); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiConceptReset handles POST /api/session/{id}/concept/reset.
func (s *Server) apiConceptReset(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	sess.ConceptSess.Reset()
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiConceptDiagram handles POST /api/session/{id}/concept/diagram.
func (s *Server) apiConceptDiagram(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	sess.ConceptSess.Diagram()
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiCheck handles POST /api/session/{id}/check.
func (s *Server) apiCheck(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	// Stub: real implementation will invoke check package.
	sess.emit(Event{Type: "check_started", Data: nil})
	sess.emit(Event{Type: "check_completed", Data: map[string]string{"result": "pass"}})
	writeJSON(w, map[string]string{"status": "ok", "result": "pass"})
}

// apiEvents handles GET /api/session/{id}/events — SSE stream.
func (s *Server) apiEvents(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET required")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, http.StatusInternalServerError, "streaming not supported")
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
		case evt, ok := <-sess.Events:
			if !ok {
				return
			}
			data, _ := json.Marshal(evt)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
