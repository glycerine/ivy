package webui

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	lg "github.com/glycerine/goivy/logic"
)

// GoBackend is the native Go implementation of Backend.
// It wraps the existing Session-based logic.
type GoBackend struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	counter  uint64
}

// NewGoBackend creates a GoBackend.
func NewGoBackend() *GoBackend {
	return &GoBackend{
		sessions: make(map[string]*Session),
	}
}

func (b *GoBackend) getSession(id string) (*Session, error) {
	b.mu.RLock()
	sess, ok := b.sessions[id]
	b.mu.RUnlock()
	if !ok {
		return nil, ErrSessionNotFound
	}
	return sess, nil
}

func (b *GoBackend) NewSession() ([]byte, error) {
	id := fmt.Sprintf("s%d", atomic.AddUint64(&b.counter, 1))
	sess := NewSession(id)
	b.mu.Lock()
	b.sessions[id] = sess
	b.mu.Unlock()
	return canonicalJSON(map[string]string{"session_id": id})
}

func (b *GoBackend) Load(sessionID, filename string, content []byte) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := sess.LoadFileContent(filename, content); err != nil {
		return nil, err
	}
	return canonicalJSON(map[string]string{"status": "ok", "filename": filename})
}

func (b *GoBackend) LoadPath(sessionID, path string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := sess.LoadFile(path); err != nil {
		return nil, err
	}
	return canonicalJSON(map[string]string{"status": "ok"})
}

func (b *GoBackend) Action(sessionID, action string, args map[string]interface{}) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	result, err := sess.ExecuteAction(action, args)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = map[string]interface{}{"status": "ok"}
	}
	return canonicalJSON(result)
}

func (b *GoBackend) GetARG(sessionID string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	cy := RenderARG(sess.Graph)
	// Ensure empty elements is [] not null to match Python.
	if cy.Elements == nil {
		cy.Elements = []CyElement{}
	}
	return canonicalJSON(cy)
}

func (b *GoBackend) GetConcept(sessionID string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Build concept data matching Python's output exactly.
	// Python builds from im.module.sig: sorts become nodes,
	// unary boolean relations become node_labels,
	// binary boolean relations become edges.
	// Relation names are bare (no parameter lists).
	var nodes []string
	var edges []string
	var nodeLabels []string
	var relations []string
	labelSorts := make(map[string]string)

	if sess.SimpleSess != nil && sess.SimpleSess.Domain != nil {
		d := sess.SimpleSess.Domain
		nodes = append(nodes, d.Nodes...)
		edges = append(edges, d.Edges...)
		nodeLabels = append(nodeLabels, d.NodeLabels...)
		// Build relations list: all concepts that are relations (not sorts).
		for name, c := range d.Concepts {
			if c != nil && c.Arity >= 1 {
				// Check if this is a sort concept (formula is "X = X") or a relation.
				isSort := false
				for _, n := range d.Nodes {
					if n == name {
						isSort = true
						break
					}
				}
				if !isSort {
					// Format as "name(X,Y)" with parameter names, not bare "name".
					if len(c.Variables) > 0 {
						relations = append(relations, name+"("+strings.Join(c.Variables, ",")+")")
					} else {
						relations = append(relations, name)
					}
				}
			}
		}
		// Build label_sorts from node_labels.
		for _, lbl := range d.NodeLabels {
			c := d.Concepts[lbl]
			if c != nil && len(c.Sorts) > 0 {
				labelSorts[lbl] = c.Sorts[0]
			}
		}
	}

	sort.Strings(relations)
	sort.Strings(edges)
	sort.Strings(nodeLabels)
	sort.Strings(nodes)

	// Build CyElements matching Python: one node per sort, no edges.
	// Python's elements include "cluster", "locked", and use "ellipse" shape.
	var elements []map[string]interface{}
	for i, name := range nodes {
		elements = append(elements, map[string]interface{}{
			"classes": "node_unknown",
			"data": map[string]interface{}{
				"cluster":    nil,
				"id":         fmt.Sprintf("n%d", i),
				"label":      name,
				"long_info":  name,
				"obj":        name,
				"shape":      "ellipse",
				"short_info": name,
			},
			"group":  "nodes",
			"locked": true,
		})
	}

	// Ensure empty slices are [] not null, and empty maps are {}.
	if elements == nil {
		elements = []map[string]interface{}{}
	}
	if relations == nil {
		relations = []string{}
	}
	if edges == nil {
		edges = []string{}
	}
	if nodeLabels == nil {
		nodeLabels = []string{}
	}
	if nodes == nil {
		nodes = []string{}
	}

	return canonicalJSON(map[string]interface{}{
		"abstract_value": map[string]bool{},
		"edges":          edges,
		"elements":       elements,
		"label_sorts":    labelSorts,
		"node_labels":    nodeLabels,
		"nodes":          nodes,
		"relations":      relations,
	})
}

func (b *GoBackend) ConceptSplit(sessionID, concept, splitBy string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if sess.ConceptSess != nil {
		sess.ConceptSess.Split(concept, splitBy)
	}
	if err := sess.SimpleSess.Split(concept, splitBy); err != nil {
		return nil, err
	}
	return okJSON, nil
}

func (b *GoBackend) ConceptEmpty(sessionID, concept string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if sess.ConceptSess != nil {
		sess.ConceptSess.SupposeEmpty(concept)
	}
	if err := sess.SimpleSess.SupposeEmpty(concept); err != nil {
		return nil, err
	}
	return okJSON, nil
}

func (b *GoBackend) ConceptRemove(sessionID, concept string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if sess.ConceptSess != nil {
		sess.ConceptSess.RemoveConcepts(concept)
	}
	if err := sess.SimpleSess.RemoveConcept(concept); err != nil {
		return nil, err
	}
	return okJSON, nil
}

func (b *GoBackend) ConceptUndo(sessionID string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if sess.ConceptSess != nil {
		if err := sess.ConceptSess.Undo(); err != nil {
			return nil, err
		}
	}
	if err := sess.SimpleSess.Undo(); err != nil {
		return nil, err
	}
	return okJSON, nil
}

func (b *GoBackend) ConceptMaterialize(sessionID, concept string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if sess.ConceptSess != nil {
		sess.ConceptSess.MaterializeNode(concept)
	}
	if err := sess.SimpleSess.Materialize(concept); err != nil {
		return nil, err
	}
	return okJSON, nil
}

func (b *GoBackend) ConceptReset(sessionID string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	sess.SimpleSess.Reset()
	if sess.CompiledSig != nil {
		sortMap := make(map[string]lg.Sort)
		for name, sort := range sess.CompiledSig.Sorts {
			if name != "bool" {
				sortMap[name] = sort
			}
		}
		symbolMap := make(map[string]*lg.Symbol)
		for name, entry := range sess.CompiledSig.Symbols {
			if entry != nil && entry.Sort != nil {
				if c, ok := entry.Sort.(lg.Sort); ok {
					symbolMap[name] = lg.NewSymbol(name, c)
				}
			}
		}
		cdDomain := GetInitialConceptDomain(sortMap, symbolMap)
		sess.ConceptSess = NewConceptInteractiveSession(
			cdDomain, nil, nil, nil, nil, nil, nil, nil, false,
		)
	}
	return okJSON, nil
}

func (b *GoBackend) ConceptDiagram(sessionID string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	sess.SimpleSess.Diagram()
	return okJSON, nil
}

func (b *GoBackend) ConceptProjection(sessionID, name, concept string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := sess.AddProjection(name, concept); err != nil {
		return nil, err
	}
	return okJSON, nil
}

func (b *GoBackend) GetToggles(sessionID string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	return canonicalJSON(sess.GetToggles())
}

func (b *GoBackend) SetToggle(sessionID, edge, displayClass string, value bool) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	sess.SetToggle(edge, displayClass, value)
	return okJSON, nil
}

func (b *GoBackend) Check(sessionID, mode string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	sess.emit(Event{Type: "check_started", Data: map[string]string{"mode": mode}})
	cr := sess.RunCheck(mode)
	sess.emit(Event{Type: "check_completed", Data: map[string]interface{}{
		"result":  cr.Result,
		"mode":    mode,
		"message": cr.Message,
	}})
	return canonicalJSON(map[string]interface{}{
		"status":            "ok",
		"result":            cr.Result,
		"mode":              mode,
		"message":           cr.Message,
		"failed_conjecture": cr.FailedConjecture,
		"failed_label":      cr.FailedLabel,
		"used_relations":    cr.UsedRelations,
	})
}

func (b *GoBackend) GetProof(sessionID string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	cy := RenderProofStack(sess.ProofStack)
	return canonicalJSON(cy)
}

func (b *GoBackend) ArgAction(sessionID, node, action string, args map[string]interface{}) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	result, err := sess.ArgNodeAction(node, action, args)
	if err != nil {
		return nil, err
	}
	return canonicalJSON(result)
}

func (b *GoBackend) ProofAction(sessionID, goal, action string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	result, err := sess.ProofGoalAction(goal, action)
	if err != nil {
		return nil, err
	}
	return canonicalJSON(result)
}

func (b *GoBackend) Save(sessionID string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	return sess.SaveState(), nil
}

func (b *GoBackend) Events(sessionID string) (<-chan Event, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	return sess.Events, nil
}

func (b *GoBackend) Close() error {
	return nil
}
