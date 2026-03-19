package webui

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/idem"
)

// GoBackend is the native Go implementation of Backend.
// It wraps the existing Session-based logic.
type GoBackend struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	counter  uint64

	// For a given Z3Context, we may only
	// talk to a Z3 on one same single thread (and
	// for us, we implement that with one single
	// goroutine per context, wired to one same
	// thread using runtime.LockOSThread().
	//
	// Note that multiple contexts are fine,
	// and so a pool of *sameSingleThread would
	// fine (a thread pool) should we want
	// that. We probably do for fuzzing.
	sst *sameSingleThread
}

// NewGoBackend creates a GoBackend.
func NewGoBackend() *GoBackend {
	b := &GoBackend{
		sessions: make(map[string]*Session),
	}
	b.sst = newSameSingleThread(b)
	b.sst.start()
	return b
}

func (b *GoBackend) do(f func(gbe *GoBackend) error) error {
	tkt := newTkt(f)
	b.sst.doChan <- tkt
	<-tkt.done
	return tkt.err
}

func (gbe *GoBackend) getSession(id string) (sess *Session, err error) {
	gbe.do(func(b *GoBackend) error {
		var ok bool
		b.mu.RLock()
		sess, ok = b.sessions[id]
		b.mu.RUnlock()
		if !ok {
			err = ErrSessionNotFound
		}
		return nil
	})
	return
}

func (gbe *GoBackend) NewSession() (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {

		id := fmt.Sprintf("s%d", atomic.AddUint64(&b.counter, 1))
		sess := NewSession(id)
		b.mu.Lock()
		b.sessions[id] = sess
		b.mu.Unlock()
		by, err = canonicalJSON(map[string]string{"session_id": id})
		return nil
	})
	return
}

func (gbe *GoBackend) Load(sessionID, filename string, content []byte) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil // return nil => sst stays up, else sst shuts down.
		}
		err = sess.LoadFileContent(filename, content)
		if err != nil {
			return nil
		}
		by, err = canonicalJSON(map[string]string{"status": "ok", "filename": filename})
		return nil
	})
	return
}

func (gbe *GoBackend) LoadPath(sessionID, path string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		err = sess.LoadFile(path)
		if err != nil {
			return nil
		}
		by, err = canonicalJSON(map[string]string{"status": "ok"})
		return nil
	})
	return
}

func (gbe *GoBackend) Action(sessionID, action string, args map[string]interface{}) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		var result map[string]any
		result, err = sess.ExecuteAction(action, args)
		if err != nil {
			return nil
		}
		if result == nil {
			result = map[string]any{"status": "ok"}
		}
		by, err = canonicalJSON(result)
		return nil
	})
	return
}

func (gbe *GoBackend) GetARG(sessionID string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		cy := RenderARG(sess.Graph)
		// Ensure empty elements is [] not null to match Python.
		if cy.Elements == nil {
			cy.Elements = []CyElement{}
		}
		by, err = canonicalJSON(cy)
		return nil
	})
	return
}

func (gbe *GoBackend) GetConcept(sessionID string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
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

		by, err = canonicalJSON(map[string]interface{}{
			"abstract_value": map[string]bool{},
			"edges":          edges,
			"elements":       elements,
			"label_sorts":    labelSorts,
			"node_labels":    nodeLabels,
			"nodes":          nodes,
			"relations":      relations,
		})
		return nil
	})
	return
}

func (gbe *GoBackend) ConceptSplit(sessionID, concept, splitBy string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		if sess.ConceptSess != nil {
			sess.ConceptSess.Split(concept, splitBy)
		}
		err = sess.SimpleSess.Split(concept, splitBy)
		if err != nil {
			return nil
		}
		by = okJSON
		return nil
	})
	return
}

func (gbe *GoBackend) ConceptEmpty(sessionID, concept string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		if sess.ConceptSess != nil {
			sess.ConceptSess.SupposeEmpty(concept)
		}
		err = sess.SimpleSess.SupposeEmpty(concept)
		if err != nil {
			return nil
		}
		by = okJSON
		return nil
	})
	return
}

func (gbe *GoBackend) ConceptRemove(sessionID, concept string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		if sess.ConceptSess != nil {
			sess.ConceptSess.RemoveConcepts(concept)
		}
		err = sess.SimpleSess.RemoveConcept(concept)
		if err != nil {
			return nil
		}
		by = okJSON
		return nil
	})
	return
}

func (gbe *GoBackend) ConceptUndo(sessionID string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {

		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		if sess.ConceptSess != nil {
			err = sess.ConceptSess.Undo()
			if err != nil {
				return nil
			}
		}
		err = sess.SimpleSess.Undo()
		if err != nil {
			return nil
		}
		by = okJSON
		return nil
	})
	return
}

func (gbe *GoBackend) ConceptMaterialize(sessionID, concept string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		if sess.ConceptSess != nil {
			sess.ConceptSess.MaterializeNode(concept)
		}
		err = sess.SimpleSess.Materialize(concept)
		if err != nil {
			return nil
		}
		by = okJSON
		return nil
	})
	return
}

func (gbe *GoBackend) ConceptReset(sessionID string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
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
		by = okJSON
		return nil
	})
	return
}

func (gbe *GoBackend) ConceptDiagram(sessionID string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		sess.SimpleSess.Diagram()
		by = okJSON
		return nil
	})
	return
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
	return b.sst.Close()
}

type sameSingleThread struct {
	doChan chan *tkt
	gbe    *GoBackend
	halt   *idem.Halter
}

func newSameSingleThread(gbe *GoBackend) *sameSingleThread {
	b := &sameSingleThread{
		doChan: make(chan *tkt),
		gbe:    gbe,
	}
	b.halt = idem.NewHalterNamed(fmt.Sprintf("sameSingleThread p=%p", b))
	return b
}

func (b *sameSingleThread) Close() error {
	b.halt.ReqStop.Close()
	<-b.halt.Done.Chan
	return nil
}

func newTkt(f func(gbe *GoBackend) error) *tkt {
	return &tkt{f: f, done: make(chan struct{})}
}

type tkt struct {
	f    func(gbe *GoBackend) error
	err  error
	done chan struct{} // closed after f is run.
}

func (b *sameSingleThread) start() {
	go func() {
		runtime.LockOSThread()
		defer func() {
			b.halt.ReqStop.Close()
			b.halt.Done.Close()
			runtime.UnlockOSThread()
		}()
		for {
			select {
			case tkt := <-b.doChan:
				tkt.err = tkt.f(b.gbe)
				close(tkt.done)

				if tkt.err != nil {
					// shut down on error
					return
				}
			case <-b.halt.ReqStop.Chan:
			}
		}
	}()
}
