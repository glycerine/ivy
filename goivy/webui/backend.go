package webui

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"

	"github.com/glycerine/goivy/module"
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
	NewSession(cfg *module.Config) ([]byte, error)
	Load(sessionID, filename string, content []byte) ([]byte, error)
	LoadPath(sessionID, path string) ([]byte, error)
	Action(sessionID, action string, args map[string]interface{}) ([]byte, error)
	GetARG(sessionID string) ([]byte, error)
	GetConcept(sessionID string) ([]byte, error)
	ConceptSplit(sessionID, concept, splitBy string) ([]byte, error)
	ConceptEmpty(sessionID, concept string) ([]byte, error)
	ConceptRemove(sessionID, concept string) ([]byte, error)
	ConceptUndo(sessionID string) ([]byte, error)
	ConceptMaterialize(sessionID, concept string) ([]byte, error)
	ConceptReset(sessionID string) ([]byte, error)
	ConceptDiagram(sessionID string) ([]byte, error)
	ConceptProjection(sessionID, name, concept string) ([]byte, error)
	GetToggles(sessionID string) ([]byte, error)
	SetToggle(sessionID, edge, displayClass string, value bool) ([]byte, error)
	Check(sessionID, mode string) ([]byte, error)
	GetProof(sessionID string) ([]byte, error)
	ArgAction(sessionID, node, action string, args map[string]interface{}) ([]byte, error)
	ProofAction(sessionID, goal, action string) ([]byte, error)
	Save(sessionID string) ([]byte, error)
	Events(sessionID string) (<-chan Event, error)
	Close() error
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
