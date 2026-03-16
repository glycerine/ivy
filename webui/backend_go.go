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
	return canonicalJSON(cy)
}

func (b *GoBackend) GetConcept(sessionID string) ([]byte, error) {
	sess, err := b.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	cy := RenderConceptGraph(sess.SimpleSess, nil)

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

	// Sort all string slices for canonical output.
	sort.Strings(relations)
	sort.Strings(edges)
	sort.Strings(nodeLabels)
	sort.Strings(nodes)

	labelSorts := make(map[string]string)
	if sess.SimpleSess != nil && sess.SimpleSess.Domain != nil {
		for _, lbl := range sess.SimpleSess.Domain.NodeLabels {
			c := sess.SimpleSess.Domain.Concepts[lbl]
			if c != nil && len(c.Sorts) > 0 {
				labelSorts[lbl] = c.Sorts[0]
			}
		}
	}

	abstractValue := make(map[string]bool)
	if sess.SimpleSess != nil {
		for k, v := range sess.SimpleSess.AbstractValue {
			if strings.HasPrefix(k, "node_label|") {
				abstractValue[k] = v
			}
		}
	}

	return canonicalJSON(map[string]interface{}{
		"elements":       cy.Elements,
		"relations":      relations,
		"edges":          edges,
		"node_labels":    nodeLabels,
		"nodes":          nodes,
		"label_sorts":    labelSorts,
		"abstract_value": abstractValue,
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
		symbolMap := make(map[string]*lg.Const)
		for name, entry := range sess.CompiledSig.Symbols {
			if entry != nil && entry.Sort != nil {
				if c, ok := entry.Sort.(lg.Sort); ok {
					symbolMap[name] = lg.NewConst(name, c)
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
