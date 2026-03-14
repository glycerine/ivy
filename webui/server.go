// Package webui provides an HTTP server with SSE push for Ivy's
// interactive verification UI.  It replaces the Python Tcl/Tk frontend
// with a Go HTTP + JSON backend, keeping the Cytoscape.js graph library
// on the browser side.
package webui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

// Server is the HTTP backend for the Ivy verification UI.
type Server struct {
	addr     string
	sessions map[string]*Session
	mu       sync.RWMutex
	mux      *http.ServeMux
}

// NewServer creates a Server that will listen on addr (e.g. ":8080").
func NewServer(addr string) *Server {
	s := &Server{
		addr:     addr,
		sessions: make(map[string]*Session),
		mux:      http.NewServeMux(),
	}
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/static/", s.handleStatic)
	s.mux.HandleFunc("/api/", s.handleAPI)
	return s
}

// Start begins serving HTTP.  It blocks until the server is shut down.
func (s *Server) Start() error {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}
	log.Printf("webui: listening on %s", s.addr)
	return srv.ListenAndServe()
}

// ServeHTTP implements http.Handler so the server can be used in tests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// handleIndex serves the main SPA page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}

// handleStatic serves embedded static assets (JS, CSS).
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// For now return 404; real assets will be embedded later.
	http.NotFound(w, r)
}

// handleAPI routes /api/* requests to the correct handler.
func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	parts := strings.Split(path, "/")

	w.Header().Set("Content-Type", "application/json")

	// POST /api/session/new
	if len(parts) == 2 && parts[0] == "session" && parts[1] == "new" {
		s.apiNewSession(w, r)
		return
	}

	// All remaining routes require a session id at parts[1].
	if len(parts) < 3 || parts[0] != "session" {
		writeErr(w, http.StatusNotFound, "unknown api endpoint")
		return
	}

	sid := parts[1]
	s.mu.RLock()
	sess, ok := s.sessions[sid]
	s.mu.RUnlock()
	if !ok {
		writeErr(w, http.StatusNotFound, "session not found")
		return
	}

	rest := strings.Join(parts[2:], "/")
	switch rest {
	case "load":
		s.apiLoad(w, r, sess)
	case "action":
		s.apiAction(w, r, sess)
	case "arg":
		s.apiARG(w, r, sess)
	case "concept":
		s.apiConcept(w, r, sess)
	case "concept/split":
		s.apiConceptSplit(w, r, sess)
	case "concept/empty":
		s.apiConceptEmpty(w, r, sess)
	case "concept/remove":
		s.apiConceptRemove(w, r, sess)
	case "concept/undo":
		s.apiConceptUndo(w, r, sess)
	case "concept/materialize":
		s.apiConceptMaterialize(w, r, sess)
	case "check":
		s.apiCheck(w, r, sess)
	case "events":
		s.apiEvents(w, r, sess)
	default:
		writeErr(w, http.StatusNotFound, "unknown api endpoint")
	}
}

// getSession returns the session for the given id, or nil.
func (s *Server) getSession(id string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[id]
}

// writeJSON marshals v as JSON to w.
func writeJSON(w http.ResponseWriter, v interface{}) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		log.Printf("webui: json encode: %v", err)
	}
}

// writeErr writes a JSON error response.
func writeErr(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	writeJSON(w, map[string]string{"error": msg})
}

// indexHTML is a minimal SPA shell served at "/".
const indexHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Ivy Verification UI</title>
<style>
body { font-family: sans-serif; margin: 0; padding: 1em; }
#cy { width: 100%; height: 600px; border: 1px solid #ccc; }
</style>
</head>
<body>
<h1>Ivy Verification UI</h1>
<div id="cy"></div>
<script src="https://unpkg.com/cytoscape@3/dist/cytoscape.min.js"></script>
</body>
</html>
`
