package webui

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/glycerine/idem"
)

// GoBackend is the native Go implementation of Backend.
// It wraps the existing Session-based logic.
type GoBackend struct {
	cfg      *goivy.Config
	sessions map[string]*Session

	// can probably delete mu, it is overkill now that we do().
	// but: LaunchUI() would still seem to need it!
	mu sync.RWMutex

	counter uint64

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
func NewGoBackend(cfg *goivy.Config) *GoBackend {
	b := &GoBackend{
		cfg:      cfg,
		sessions: make(map[string]*Session),
	}
	b.sst = newSameSingleThread(b)
	//vv("NewGoBackend() with b=%p ; sst=%p", b, b.sst)
	b.sst.start()
	return b
}

func (b *GoBackend) do(f func(gbe *GoBackend) error) error {
	tkt := newTkt(f)
	b.sst.doChan <- tkt
	//vv("about to wait on <-tkt.done = %p", tkt.done) // [goID 7] 2026-03-19 07:10:40.300693000 +0000 UTC about to wait on <-tkt.done = 0x214e59806690        [goID 7] 2026-03-19 07:10:40.302926000 +0000 UTC about to wait on <-tkt.done = 0x214e59806af0
	<-tkt.done // hung here.
	return tkt.err
}

// must only called by sst goro
func (b *GoBackend) getSession(id string) (sess *Session, err error) {

	var ok bool
	b.mu.RLock()
	sess, ok = b.sessions[id]
	b.mu.RUnlock()
	if !ok {
		err = ErrSessionNotFound
	}
	return
}

func (gbe *GoBackend) NewSession(cfg *goivy.Config) (by []byte, err error) {
	// note well this pattern: if the closure
	// returns a non-nil error, this shuts down
	// the sameSingleThread. Currently we do
	// not shutdown the thread on regular API
	// level errors, but that might change in the future.
	gbe.do(func(b *GoBackend) error {

		id := fmt.Sprintf("s%d", atomic.AddUint64(&b.counter, 1))
		sess := NewSession(cfg, id)
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
		by, err = canonicalJSON(AnalysisUIARGPayload(sess.AGUI))
		return nil
	})
	return
}

func (gbe *GoBackend) GetConcept(sessionID, sheetID, nodeID string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		var selectedNode string
		var stateLabel string
		if nodeID != "" {
			selectedNode, stateLabel, err = sess.selectConceptARGNode(sheetID, nodeID)
			if err != nil {
				return nil
			}
		}

		var checks *Toggles
		var facts []FactSelection
		var displayChecks *DisplayCheckboxes
		sess.mu.Lock()
		widget := sess.ensureConceptGraphWidgetForSheetLocked(sheetID)
		displayChecks = sess.ensureConceptChecksForSheetLocked(sheetID)
		checks = displayChecks.Snapshot()
		sess.toggles = checks
		if widget != nil {
			facts = widget.ConstraintFacts()
		}
		sess.mu.Unlock()

		// Render concept graph with octagon nodes, edges, and per-sort colors.
		cy := RenderConceptGraph(sess.SimpleSess, displayChecks)
		if cy.Elements == nil {
			cy.Elements = []WebUICyElement{}
		}

		// Gather metadata for the state checkbox panel and JS rendering.
		var nodes []string
		var edges []string
		var nodeLabels []string
		var relations []string
		labelSorts := make(map[string]string)
		edgeSorts := make(map[string][]string)

		if sess.SimpleSess != nil && sess.SimpleSess.Domain != nil {
			d := sess.SimpleSess.Domain
			nodes = append(nodes, d.Nodes...)
			edges = append(edges, d.Edges...)
			nodeLabels = append(nodeLabels, d.NodeLabels...)
			relations = append(relations, sess.SimpleSess.RelationNames()...)
			for _, lbl := range d.NodeLabels {
				c := d.Concepts[lbl]
				if c != nil && len(c.Sorts) > 0 {
					labelSorts[lbl] = c.Sorts[0]
				}
			}
			for _, edge := range d.Edges {
				c := d.Concepts[edge]
				if c != nil && len(c.Sorts) >= 2 {
					edgeSorts[edge] = append([]string{}, c.Sorts...)
				}
			}
		}

		sort.Strings(relations)
		sort.Strings(edges)
		sort.Strings(nodeLabels)
		sort.Strings(nodes)

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

		// Include real abstract_value from the concept session (for node labels).
		abstractValue := make(map[string]bool)
		if sess.SimpleSess != nil {
			for k, v := range sess.SimpleSess.AbstractValue {
				if strings.HasPrefix(k, "node_label|") {
					abstractValue[k] = v
				}
			}
		}

		conceptDomain := NewConceptDomain() // non-nil fallback; SimpleSess is always set
		if sess.SimpleSess != nil {
			conceptDomain = sess.SimpleSess.Domain
		}
		response := map[string]interface{}{
			"abstract_value":              abstractValue,
			"concept_domain":              conceptDomain,
			"concept_interactive_session": conceptInteractiveSessionPayload(sess.ConceptSess),
			"concept_session":             sess.SimpleSess,
			"display_checkboxes":          checks,
			"edges":                       edges,
			"edge_sorts":                  edgeSorts,
			"elements":                    cy.Elements,
			"facts":                       facts,
			"graph":                       conceptGraphPayload(nil, nil),
			"graph_stack":                 conceptGraphStackPayload(nil),
			"label_sorts":                 labelSorts,
			"node_labels":                 nodeLabels,
			"nodes":                       nodes,
			"relations":                   relations,
			"selected_node":               selectedNode,
			"state_label":                 stateLabel,
			"toggles":                     checks,
		}
		if widget != nil && widget.G() != nil {
			response["graph"] = conceptGraphPayload(widget.G(), widget.GraphStack)
			response["graph_stack"] = conceptGraphStackPayload(widget.GraphStack)
		}
		if sheetID != "" {
			response["sheet_id"] = sheetID
		}
		by, err = canonicalJSON(response)
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

func (gbe *GoBackend) ConceptMaterialize(sessionID string, req ConceptMaterializeRequest) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		if req.Type == "edge" {
			relation := req.Relation
			if relation == "" {
				relation = req.Concept
			}
			if relation == "" || req.Source == "" || req.Target == "" {
				err = fmt.Errorf("materialize edge: relation, source, and target are required")
				return nil
			}
			widget := sess.ensureConceptGraphWidgetLocked()
			if widget == nil {
				err = fmt.Errorf("materialize edge: no concept graph")
				return nil
			}
			var witnesses []string
			witnesses, err = widget.MaterializeEdge(relation, req.Source, req.Target, req.Positive)
			if err != nil {
				return nil
			}
			by, err = canonicalJSON(map[string]interface{}{"status": "ok", "witnesses": witnesses})
			return nil
		}
		concept := req.Concept
		if concept == "" {
			err = fmt.Errorf("materialize node: concept is required")
			return nil
		}
		if sess.ConceptSess != nil {
			sess.ConceptSess.MaterializeNode(concept)
		}
		if sess.AGUI != nil && sess.AGUI.CurrentConceptGraph != nil {
			_, _ = sess.AGUI.CurrentConceptGraph.MaterializeNode(concept)
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
			sortMap := make(map[string]goivy.Sort)
			for name, sort := range sess.CompiledSig.Sorts.All() {
				if name != "bool" {
					sortMap[name] = sort
				}
			}
			symbolMap := make(map[string]*goivy.Const)
			for name, entry := range sess.CompiledSig.Symbols.All() {
				if entry != nil && entry.Sort != nil {
					if c, ok := entry.Sort.(goivy.Sort); ok {
						symbolMap[name] = goivy.NewConst(name, c)
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
		if sess.CTIUI != nil {
			diagram, diagErr := sess.CTIUI.Diagram()
			if diagErr != nil {
				err = diagErr
				return nil
			}
			response := map[string]interface{}{
				"status":  "ok",
				"diagram": diagram,
			}
			if sess.CTIUI.CurrentConceptGraph != nil {
				response["concept"] = conceptGraphActionPayload(sess.CTIUI.CurrentConceptGraph)
			}
			by, err = canonicalJSON(response)
			return nil
		}
		sess.SimpleSess.Diagram()
		by = okJSON
		return nil
	})
	return
}

func (gbe *GoBackend) ConceptProjection(sessionID, name, concept string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		err = sess.AddProjection(name, concept)
		if err != nil {
			return nil
		}
		by = okJSON
		return nil
	})
	return
}

func (gbe *GoBackend) GetToggles(sessionID string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		by, err = canonicalJSON(sess.GetToggles())
		return nil
	})
	return
}

func (gbe *GoBackend) SetToggle(sessionID, edge, displayClass string, value bool) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		sess.SetToggle(edge, displayClass, value)
		by = okJSON
		return nil
	})
	return
}

func (gbe *GoBackend) Check(sessionID, mode string, options CheckOptions) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		sess.emit(Event{Type: "check_started", Data: map[string]string{"mode": mode}})
		cr := sess.RunCheckWithOptions(mode, options)
		sess.emit(Event{Type: "check_completed", Data: map[string]interface{}{
			"result":  cr.Result,
			"mode":    mode,
			"message": cr.Message,
		}})

		m := map[string]interface{}{
			"status":            "ok",
			"result":            cr.Result,
			"mode":              mode,
			"message":           cr.Message,
			"failed_conjecture": cr.FailedConjecture,
			"failed_label":      cr.FailedLabel,
			"used_relations":    cr.UsedRelations,
		}
		if !gbe.cfg.WebUIConformCheck {
			// python does not have this new/extra flag
			// and we are comparing byte-for-byte.
			// TODO: better solution would be to add to python too.
			m["z3_contacted"] = cr.Z3Contacted
		}
		if !gbe.cfg.WebUIConformCheck && cr.Result == "fail" && sess.AGUI != nil && sess.AGUI.AG != nil && len(sess.AGUI.AG.States) > 0 {
			m["trace_arg"] = AnalysisUIARGPayload(sess.AGUI)
		}
		if !gbe.cfg.WebUIConformCheck {
			if cr.CounterexampleTrace != "" {
				m["counterexample_trace"] = cr.CounterexampleTrace
			}
			if cr.CounterexampleDetails != "" {
				m["counterexample_details"] = cr.CounterexampleDetails
			}
		}
		by, err = canonicalJSON(m)
		return nil
	})
	return
}

func (gbe *GoBackend) GetProof(sessionID string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		cy := RenderProofStack(sess.ProofStackData())
		by, err = canonicalJSON(cy)
		return nil
	})
	return
}

func (gbe *GoBackend) ArgAction(sessionID, node, action string, args map[string]interface{}) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		result, err := sess.ArgNodeAction(node, action, args)
		if err != nil {
			return nil
		}
		by, err = canonicalJSON(result)
		return nil
	})
	return
}

func (gbe *GoBackend) ProofAction(sessionID, goal, action string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		var result map[string]any
		result, err = sess.ProofGoalAction(goal, action)
		if err != nil {
			return nil
		}
		by, err = canonicalJSON(result)
		return nil
	})
	return
}

func (gbe *GoBackend) Save(sessionID string) (by []byte, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		by = sess.SaveState()
		return nil
	})
	return
}

func (gbe *GoBackend) Events(sessionID string) (ch <-chan Event, err error) {
	gbe.do(func(b *GoBackend) error {
		var sess *Session
		sess, err = b.getSession(sessionID)
		if err != nil {
			return nil
		}
		ch = sess.Events
		return nil
	})
	return
}

func (gbe *GoBackend) Close() error {
	//vv("gbe = %p Close()", gbe)
	return gbe.sst.Close()
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
	if b == nil {
		return nil
	}
	b.halt.ReqStop.Close()
	//vv("closed halt = %p ReqStop", b.halt)
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

func (sst *sameSingleThread) start() {
	//vv(" sst.start(gbe = %p); sst=%p", sst.gbe, sst)

	go func() {
		runtime.LockOSThread()
		defer func() {
			//vv("sst = %p, defer running", sst) // not seen
			sst.halt.ReqStop.Close()
			sst.halt.Done.Close()
			runtime.UnlockOSThread()
			//vv("sst = %p, defer ran", sst)
		}()

		for {
			////vv("about to wait on sst.halt = %p .ReqStop.Chan", sst.halt)
			select {
			case tkt := <-sst.doChan:
				//vv("got tkt = %p about to call f()", tkt)
				tkt.err = tkt.f(sst.gbe)
				//vv("about to close tkt.done = %p", tkt.done) // only 1x: backend_go.go:627 [goID 10] 2026-03-19 07:15:42.981212000 +0000 UTC about to close tkt.done = 0x14ba9bc5aaf0
				close(tkt.done)

				if tkt.err != nil {
					// shut down on error
					//vv("ran a func, shutting down on err='%v'", tkt.err) //not seen
					return
				}
				//vv("ran a func, stayed up.") // seen once.
			case <-sst.halt.ReqStop.Chan:
				//vv("sst=%p got ReqStop", sst) // not seen
				return
			}
		}
	}()
}
