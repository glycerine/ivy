package webui

import (
	"bytes"
	"encoding/json"
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"log"
	"sync"
)

// ConformBackend sends every request to both a Go and Python backend,
// compares the responses bit-for-bit, and only returns a result to the
// browser when both backends agree. If they disagree, the browser
// receives an error — never unverified data.
type ConformBackend struct {
	goBE Backend
	pyBE Backend
	// sessionMap maps Go session ID → Python session ID.
	// Each backend generates its own IDs; we expose the Go ID to the browser
	// and translate to the Python ID when forwarding.
	sessionMap map[string]string
	mu         sync.Mutex
}

// NewConformBackend creates a ConformBackend wrapping two backends.
func NewConformBackend(goBE, pyBE Backend) *ConformBackend {
	return &ConformBackend{
		goBE:       goBE,
		pyBE:       pyBE,
		sessionMap: make(map[string]string),
	}
}

// pySessionID returns the Python session ID for a Go session ID.
func (c *ConformBackend) pySessionID(goSID string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if pySID, ok := c.sessionMap[goSID]; ok {
		return pySID
	}
	return goSID // fallback
}

// conform runs f on both backends concurrently and compares results.
// Returns the shared result only if both succeed with identical bytes.
func (c *ConformBackend) conform(op string, goFn, pyFn func() ([]byte, error)) ([]byte, error) {
	var goData, pyData []byte
	var goErr, pyErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); goData, goErr = goFn() }()
	go func() { defer wg.Done(); pyData, pyErr = pyFn() }()
	wg.Wait()

	// Both must succeed or both must fail with identical error.
	if goErr != nil && pyErr != nil {
		if goErr.Error() == pyErr.Error() {
			return nil, goErr // same error, conformance OK
		}
		log.Printf("CONFORM MISMATCH [%s] both errored but differently:\n  Go:  %v\n  Py:  %v", op, goErr, pyErr)
		return nil, fmt.Errorf("conformance error in %s: Go error=%q, Python error=%q", op, goErr, pyErr)
	}
	if goErr != nil {
		log.Printf("CONFORM MISMATCH [%s] Go errored, Python succeeded:\n  Go err: %v\n  Py data: %s", op, goErr, pyData)
		return nil, fmt.Errorf("conformance error in %s: Go failed (%v) but Python succeeded", op, goErr)
	}
	if pyErr != nil {
		log.Printf("CONFORM MISMATCH [%s] Python errored, Go succeeded:\n  Py err: %v\n  Go data: %s", op, pyErr, goData)
		return nil, fmt.Errorf("conformance error in %s: Python failed (%v) but Go succeeded", op, pyErr)
	}

	// Both succeeded — compare bytes after known sidecar-shape normalizations.
	goCompare := normalizeConformanceData(op, goData)
	pyCompare := normalizeConformanceData(op, pyData)
	if !bytes.Equal(goCompare, pyCompare) {
		diffPos := firstDiffPos(goCompare, pyCompare)
		log.Printf("CONFORM MISMATCH [%s] responses differ at byte %d:\n  Go (%d bytes): %s\n  Py (%d bytes): %s",
			op, diffPos, len(goData), goData, len(pyData), pyData)
		return nil, fmt.Errorf("conformance error in %s: Go and Python responses differ at byte %d (Go=%d bytes, Py=%d bytes)",
			op, diffPos, len(goData), len(pyData))
	}

	return goData, nil
}

func normalizeConformanceData(op string, data []byte) []byte {
	if op != "GetARG" {
		return data
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return data
	}
	if positions, ok := payload["positions"]; ok {
		if positions == nil {
			delete(payload, "positions")
		} else if positionMap, ok := positions.(map[string]interface{}); ok && len(positionMap) == 0 {
			delete(payload, "positions")
		}
	}
	normalized, err := canonicalJSON(payload)
	if err != nil {
		return data
	}
	return normalized
}

// firstDiffPos returns the byte offset of the first difference between a and b.
func firstDiffPos(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n // differ in length
}

func (c *ConformBackend) NewSession(cfg *goivy.Config) ([]byte, error) {
	// NewSession is special: both backends create sessions independently,
	// but session IDs will differ. We map Go ID → Python ID.
	goData, goErr := c.goBE.NewSession(cfg)
	pyData, pyErr := c.pyBE.NewSession(cfg)
	if goErr != nil {
		return nil, goErr
	}
	if pyErr != nil {
		return nil, fmt.Errorf("conformance error in NewSession: Python failed: %v", pyErr)
	}

	// Extract session IDs from both responses.
	var goResp, pyResp struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(goData, &goResp); err != nil {
		return nil, fmt.Errorf("conformance error: bad Go session response: %w", err)
	}
	if err := json.Unmarshal(pyData, &pyResp); err != nil {
		return nil, fmt.Errorf("conformance error: bad Python session response: %w", err)
	}

	// Map Go session ID → Python session ID.
	c.mu.Lock()
	c.sessionMap[goResp.SessionID] = pyResp.SessionID
	c.mu.Unlock()

	log.Printf("CONFORM: session mapping %s (Go) → %s (Py)", goResp.SessionID, pyResp.SessionID)

	// Return the Go response (session IDs are expected to differ).
	return goData, nil
}

func (c *ConformBackend) Load(sessionID, filename string, content []byte, isolate string) ([]byte, error) {
	return c.conform("Load",
		func() ([]byte, error) { return c.goBE.Load(sessionID, filename, content, isolate) },
		func() ([]byte, error) { return c.pyBE.Load(c.pySessionID(sessionID), filename, content, isolate) },
	)
}

func (c *ConformBackend) LoadPath(sessionID, path string) ([]byte, error) {
	return c.conform("LoadPath",
		func() ([]byte, error) { return c.goBE.LoadPath(sessionID, path) },
		func() ([]byte, error) { return c.pyBE.LoadPath(c.pySessionID(sessionID), path) },
	)
}

func (c *ConformBackend) Action(sessionID, action string, args map[string]interface{}) ([]byte, error) {
	return c.conform("Action:"+action,
		func() ([]byte, error) { return c.goBE.Action(sessionID, action, args) },
		func() ([]byte, error) { return c.pyBE.Action(c.pySessionID(sessionID), action, args) },
	)
}

func (c *ConformBackend) GetARG(sessionID string, full bool) ([]byte, error) {
	return c.conform("GetARG",
		func() ([]byte, error) { return c.goBE.GetARG(sessionID, full) },
		func() ([]byte, error) { return c.pyBE.GetARG(c.pySessionID(sessionID), full) },
	)
}

func (c *ConformBackend) GetCTIARG(sessionID string) ([]byte, error) {
	return c.conform("GetCTIARG",
		func() ([]byte, error) { return c.goBE.GetCTIARG(sessionID) },
		func() ([]byte, error) { return c.pyBE.GetCTIARG(c.pySessionID(sessionID)) },
	)
}

func (c *ConformBackend) GetConcept(sessionID, sheetID, nodeID string) ([]byte, error) {
	return c.conform("GetConcept",
		func() ([]byte, error) { return c.goBE.GetConcept(sessionID, sheetID, nodeID) },
		func() ([]byte, error) { return c.pyBE.GetConcept(c.pySessionID(sessionID), sheetID, nodeID) },
	)
}

func (c *ConformBackend) GetMenus(sessionID string, req MenuRequest) ([]byte, error) {
	return c.conform("GetMenus",
		func() ([]byte, error) { return c.goBE.GetMenus(sessionID, req) },
		func() ([]byte, error) { return c.pyBE.GetMenus(c.pySessionID(sessionID), req) },
	)
}

func (c *ConformBackend) ConceptSplit(sessionID, concept, splitBy string) ([]byte, error) {
	return c.conform("ConceptSplit",
		func() ([]byte, error) { return c.goBE.ConceptSplit(sessionID, concept, splitBy) },
		func() ([]byte, error) { return c.pyBE.ConceptSplit(c.pySessionID(sessionID), concept, splitBy) },
	)
}

func (c *ConformBackend) ConceptEmpty(sessionID, concept string) ([]byte, error) {
	return c.conform("ConceptEmpty",
		func() ([]byte, error) { return c.goBE.ConceptEmpty(sessionID, concept) },
		func() ([]byte, error) { return c.pyBE.ConceptEmpty(c.pySessionID(sessionID), concept) },
	)
}

func (c *ConformBackend) ConceptRemove(sessionID, concept string) ([]byte, error) {
	return c.conform("ConceptRemove",
		func() ([]byte, error) { return c.goBE.ConceptRemove(sessionID, concept) },
		func() ([]byte, error) { return c.pyBE.ConceptRemove(c.pySessionID(sessionID), concept) },
	)
}

func (c *ConformBackend) ConceptUndo(sessionID string) ([]byte, error) {
	return c.conform("ConceptUndo",
		func() ([]byte, error) { return c.goBE.ConceptUndo(sessionID) },
		func() ([]byte, error) { return c.pyBE.ConceptUndo(c.pySessionID(sessionID)) },
	)
}

func (c *ConformBackend) ConceptMaterialize(sessionID string, req ConceptMaterializeRequest) ([]byte, error) {
	return c.conform("ConceptMaterialize",
		func() ([]byte, error) { return c.goBE.ConceptMaterialize(sessionID, req) },
		func() ([]byte, error) { return c.pyBE.ConceptMaterialize(c.pySessionID(sessionID), req) },
	)
}

func (c *ConformBackend) ConceptReset(sessionID string) ([]byte, error) {
	return c.conform("ConceptReset",
		func() ([]byte, error) { return c.goBE.ConceptReset(sessionID) },
		func() ([]byte, error) { return c.pyBE.ConceptReset(c.pySessionID(sessionID)) },
	)
}

func (c *ConformBackend) ConceptDiagram(sessionID string) ([]byte, error) {
	return c.conform("ConceptDiagram",
		func() ([]byte, error) { return c.goBE.ConceptDiagram(sessionID) },
		func() ([]byte, error) { return c.pyBE.ConceptDiagram(c.pySessionID(sessionID)) },
	)
}

func (c *ConformBackend) ConceptProjection(sessionID, name, concept string) ([]byte, error) {
	return c.conform("ConceptProjection",
		func() ([]byte, error) { return c.goBE.ConceptProjection(sessionID, name, concept) },
		func() ([]byte, error) { return c.pyBE.ConceptProjection(c.pySessionID(sessionID), name, concept) },
	)
}

func (c *ConformBackend) GetToggles(sessionID string) ([]byte, error) {
	return c.conform("GetToggles",
		func() ([]byte, error) { return c.goBE.GetToggles(sessionID) },
		func() ([]byte, error) { return c.pyBE.GetToggles(c.pySessionID(sessionID)) },
	)
}

func (c *ConformBackend) SetToggle(sessionID, edge, displayClass string, value bool) ([]byte, error) {
	return c.conform("SetToggle",
		func() ([]byte, error) { return c.goBE.SetToggle(sessionID, edge, displayClass, value) },
		func() ([]byte, error) { return c.pyBE.SetToggle(c.pySessionID(sessionID), edge, displayClass, value) },
	)
}

func (c *ConformBackend) Check(sessionID, mode string, options CheckOptions) ([]byte, error) {
	return c.conform("Check:"+mode,
		func() ([]byte, error) { return c.goBE.Check(sessionID, mode, options) },
		func() ([]byte, error) { return c.pyBE.Check(c.pySessionID(sessionID), mode, options) },
	)
}

func (c *ConformBackend) GetProof(sessionID string) ([]byte, error) {
	return c.conform("GetProof",
		func() ([]byte, error) { return c.goBE.GetProof(sessionID) },
		func() ([]byte, error) { return c.pyBE.GetProof(c.pySessionID(sessionID)) },
	)
}

func (c *ConformBackend) ArgAction(sessionID, node, action string, args map[string]interface{}) ([]byte, error) {
	return c.conform("ArgAction:"+action,
		func() ([]byte, error) { return c.goBE.ArgAction(sessionID, node, action, args) },
		func() ([]byte, error) { return c.pyBE.ArgAction(c.pySessionID(sessionID), node, action, args) },
	)
}

func (c *ConformBackend) ProofAction(sessionID, goal, action string) ([]byte, error) {
	return c.conform("ProofAction:"+action,
		func() ([]byte, error) { return c.goBE.ProofAction(sessionID, goal, action) },
		func() ([]byte, error) { return c.pyBE.ProofAction(c.pySessionID(sessionID), goal, action) },
	)
}

func (c *ConformBackend) Save(sessionID string) ([]byte, error) {
	return c.conform("Save",
		func() ([]byte, error) { return c.goBE.Save(sessionID) },
		func() ([]byte, error) { return c.pyBE.Save(c.pySessionID(sessionID)) },
	)
}

func (c *ConformBackend) Events(sessionID string) (<-chan Event, error) {
	// For SSE events, we return the Go channel directly.
	// Conformance of events is checked separately via the conform
	// comparison on each operation's response.
	return c.goBE.Events(sessionID)
}

func (c *ConformBackend) Close() error {
	goErr := c.goBE.Close()
	pyErr := c.pyBE.Close()
	if goErr != nil {
		return goErr
	}
	return pyErr
}
