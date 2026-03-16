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
	result, err := sess.ExecuteAction(req.Action, req.Args)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if result == nil {
		result = map[string]interface{}{"status": "ok"}
	}
	writeJSON(w, result)
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
	cy := RenderConceptGraph(sess.SimpleSess, nil)
	// Include relation names for the state checkbox panel.
	var relations, edges, nodeLabels, nodes []string
	if sess.ConceptSess != nil {
		relations = sess.ConceptSess.RelationNames()
		edges = sess.ConceptSess.EdgeNames()
		nodeLabels = sess.ConceptSess.NodeLabelNames()
		nodes = sess.ConceptSess.NodeNames()
	} else {
		relations = sess.SimpleSess.RelationNames()
		edges = sess.SimpleSess.Domain.Edges
		nodeLabels = sess.SimpleSess.Domain.NodeLabels
		nodes = sess.SimpleSess.Domain.Nodes
	}
	// Build label_sorts: maps each node_label name → its parameter sort name.
	// This lets the JS know which sort node each label belongs to
	// (e.g., "semaphore" → "server").
	labelSorts := make(map[string]string)
	if sess.SimpleSess != nil && sess.SimpleSess.Domain != nil {
		for _, lbl := range sess.SimpleSess.Domain.NodeLabels {
			c := sess.SimpleSess.Domain.Concepts[lbl]
			if c != nil && len(c.Sorts) > 0 {
				labelSorts[lbl] = c.Sorts[0]
			}
		}
	}

	// Build abstract_value for node labels if available.
	// Maps "node_label|<k>|<node>|<label>" → bool.
	abstractValue := make(map[string]bool)
	if sess.SimpleSess != nil {
		for k, v := range sess.SimpleSess.AbstractValue {
			if strings.HasPrefix(k, "node_label|") {
				abstractValue[k] = v
			}
		}
	}

	writeJSON(w, map[string]interface{}{
		"elements":       cy.Elements,
		"relations":      relations,
		"edges":          edges,
		"node_labels":    nodeLabels,
		"nodes":          nodes,
		"label_sorts":    labelSorts,
		"abstract_value": abstractValue,
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
	if sess.ConceptSess != nil {
		sess.ConceptSess.Split(req.Concept, req.SplitBy)
	}
	if err := sess.SimpleSess.Split(req.Concept, req.SplitBy); err != nil {
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
	if sess.ConceptSess != nil {
		sess.ConceptSess.SupposeEmpty(req.Concept)
	}
	if err := sess.SimpleSess.SupposeEmpty(req.Concept); err != nil {
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
	if sess.ConceptSess != nil {
		sess.ConceptSess.RemoveConcepts(req.Concept)
	}
	if err := sess.SimpleSess.RemoveConcept(req.Concept); err != nil {
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
	if sess.ConceptSess != nil {
		if err := sess.ConceptSess.Undo(); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if err := sess.SimpleSess.Undo(); err != nil {
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
	if sess.ConceptSess != nil {
		sess.ConceptSess.MaterializeNode(req.Concept)
	}
	if err := sess.SimpleSess.Materialize(req.Concept); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiToggles handles GET/POST /api/session/{id}/toggles.
// GET returns current toggle state; POST updates it.
func (s *Server) apiToggles(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method == http.MethodGet {
		writeJSON(w, sess.GetToggles())
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
	sess.SetToggle(req.Edge, req.DisplayClass, req.Value)
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiConceptReset handles POST /api/session/{id}/concept/reset.
func (s *Server) apiConceptReset(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	sess.SimpleSess.Reset()
	// ConceptSess.Reset needs sort/symbol maps — use empty for now
	// TODO: preserve original sort/symbol maps from file load
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiConceptDiagram handles POST /api/session/{id}/concept/diagram.
func (s *Server) apiConceptDiagram(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	sess.SimpleSess.Diagram()
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiProof handles GET /api/session/{id}/proof.
func (s *Server) apiProof(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	cy := RenderProofStack(sess.ProofStack)
	writeJSON(w, cy)
}

// apiConceptProjection handles POST /api/session/{id}/concept/projection.
func (s *Server) apiConceptProjection(w http.ResponseWriter, r *http.Request, sess *Session) {
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
	if err := sess.AddProjection(req.Name, req.Concept); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// apiArgAction handles POST /api/session/{id}/arg/action.
func (s *Server) apiArgAction(w http.ResponseWriter, r *http.Request, sess *Session) {
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
	result, err := sess.ArgNodeAction(req.Node, req.Action, req.Args)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, result)
}

// apiProofAction handles POST /api/session/{id}/proof/action.
func (s *Server) apiProofAction(w http.ResponseWriter, r *http.Request, sess *Session) {
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
	result, err := sess.ProofGoalAction(req.Goal, req.Action)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, result)
}

// apiSave handles GET /api/session/{id}/save.
func (s *Server) apiSave(w http.ResponseWriter, r *http.Request, sess *Session) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	data := sess.SaveState()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=ivy_session.json")
	w.Write(data)
}

// apiCheck handles POST /api/session/{id}/check.
// Dispatches to the appropriate verification mode via check/art/updr packages.
func (s *Server) apiCheck(w http.ResponseWriter, r *http.Request, sess *Session) {
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

	sess.emit(Event{Type: "check_started", Data: map[string]string{"mode": req.Mode}})

	// Run verification using the compiled module and Z3.
	cr := sess.RunCheck(req.Mode)

	sess.emit(Event{Type: "check_completed", Data: map[string]interface{}{
		"result":  cr.Result,
		"mode":    req.Mode,
		"message": cr.Message,
	}})
	writeJSON(w, map[string]interface{}{
		"status":             "ok",
		"result":             cr.Result,
		"mode":               req.Mode,
		"message":            cr.Message,
		"failed_conjecture":  cr.FailedConjecture,
		"failed_label":       cr.FailedLabel,
	})
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
