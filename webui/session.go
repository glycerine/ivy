package webui

import (
	"fmt"
	"sync"
)

// Event is a server-sent event delivered to the browser over SSE.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Session holds the state for one interactive verification session.
type Session struct {
	ID          string
	Graph       *AnalysisGraphState // ARG state
	ConceptSess *ConceptSession     // concept graph state
	Events      chan Event          // buffered SSE channel
	mu          sync.Mutex
	FilePath    string // last loaded file path
	FileContent string // file content (when uploaded via browser)
}

// NewSession creates a new verification session with the given id.
func NewSession(id string) *Session {
	return &Session{
		ID:     id,
		Events: make(chan Event, 64),
		Graph:  NewAnalysisGraphState(),
		ConceptSess: NewConceptSession(),
	}
}

// LoadFile loads an Ivy source file by path into this session.
func (s *Session) LoadFile(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if path == "" {
		return fmt.Errorf("empty file path")
	}
	s.FilePath = path
	s.emit(Event{Type: "file_loaded", Data: map[string]string{"path": path}})
	return nil
}

// LoadFileContent loads an Ivy source file from in-memory content
// (used when the browser uploads a file via multipart form).
func (s *Session) LoadFileContent(filename string, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if filename == "" {
		return fmt.Errorf("empty filename")
	}
	s.FilePath = filename
	s.FileContent = string(content)
	s.emit(Event{Type: "file_loaded", Data: map[string]string{
		"filename": filename,
		"size":     fmt.Sprintf("%d", len(content)),
	}})
	return nil
}

// ExecuteAction runs a named verification action with the given arguments.
// This is a stub: real logic will dispatch to the checker / art packages.
func (s *Session) ExecuteAction(actionName string, args map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if actionName == "" {
		return fmt.Errorf("empty action name")
	}
	s.emit(Event{
		Type: "action_started",
		Data: map[string]interface{}{
			"action": actionName,
			"args":   args,
		},
	})
	// TODO: dispatch real verification actions
	s.emit(Event{
		Type: "action_completed",
		Data: map[string]interface{}{
			"action": actionName,
			"status": "ok",
		},
	})
	return nil
}

// emit sends an event on the SSE channel (non-blocking drop if full).
func (s *Session) emit(e Event) {
	select {
	case s.Events <- e:
	default:
		// drop if buffer full
	}
}
