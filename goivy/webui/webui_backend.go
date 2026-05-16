package webui

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	goivy "github.com/glycerine/ivy/goivy"
)

// content holds our static web server content.
//
//go:embed static/*
var staticContent embed.FS

// Backend abstracts the Ivy verification engine behind the web UI.
// Each method returns canonical JSON bytes on success or an error.
// Canonical JSON uses sorted keys and compact formatting so that
// conformance checking can compare outputs bit-for-bit.
type Backend interface {
	NewSession(cfg *goivy.Config) ([]byte, error)
	Load(sessionID, filename string, content []byte, isolate string) ([]byte, error)
	LoadPath(sessionID, path string) ([]byte, error)
	Action(sessionID, action string, args map[string]interface{}) ([]byte, error)
	GetARG(sessionID string, full bool) ([]byte, error)
	GetCTIARG(sessionID string) ([]byte, error)
	GetConcept(sessionID, sheetID, nodeID string) ([]byte, error)
	ConceptSplit(sessionID, concept, splitBy string) ([]byte, error)
	ConceptEmpty(sessionID, concept string) ([]byte, error)
	ConceptRemove(sessionID, concept string) ([]byte, error)
	ConceptUndo(sessionID string) ([]byte, error)
	ConceptMaterialize(sessionID string, req ConceptMaterializeRequest) ([]byte, error)
	ConceptReset(sessionID string) ([]byte, error)
	ConceptDiagram(sessionID string) ([]byte, error)
	ConceptProjection(sessionID, name, concept string) ([]byte, error)
	GetToggles(sessionID string) ([]byte, error)
	SetToggle(sessionID, edge, displayClass string, value bool) ([]byte, error)
	Check(sessionID, mode string, options CheckOptions) ([]byte, error)
	GetProof(sessionID string) ([]byte, error)
	ArgAction(sessionID, node, action string, args map[string]interface{}) ([]byte, error)
	ProofAction(sessionID, goal, action string) ([]byte, error)
	Save(sessionID string) ([]byte, error)
	Events(sessionID string) (<-chan Event, error)
	Close() error
}

type ConceptMaterializeRequest struct {
	Concept  string `json:"concept,omitempty"`
	Type     string `json:"type,omitempty"`
	Relation string `json:"relation,omitempty"`
	Source   string `json:"source,omitempty"`
	Target   string `json:"target,omitempty"`
	Positive bool   `json:"positive,omitempty"`
}

type CheckOptions struct {
	Bound   int             `json:"bound,omitempty"`
	Context context.Context `json:"-"`
}

// ErrSessionNotFound is returned when a session ID is not recognized.
var ErrSessionNotFound = errors.New("session not found")

// canonicalJSON marshals v to compact JSON with sorted map keys.
// Both Go and Python backends must produce identical output for the
// same logical data when using this format.
// Uses SetEscapeHTML(false) to avoid escaping <, >, & which Python
// json.dumps does not escape, ensuring byte-identical output.
func canonicalJSON(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	// Encode appends a newline; trim it for compatibility with json.Marshal
	b := buf.Bytes()
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	return b, nil
}

// okJSON is the canonical encoding of {"status":"ok"}.
var okJSON, _ = canonicalJSON(map[string]string{"status": "ok"})
