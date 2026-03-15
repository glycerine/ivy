package webui

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/parser"
)

// Event is a server-sent event delivered to the browser over SSE.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Session holds the state for one interactive verification session.
type Session struct {
	ID          string
	Graph       *AnalysisGraphState          // ARG state
	ConceptSess *ConceptInteractiveSession   // concept graph state (uses Z3 via Alpha)
	SimpleSess  *ConceptSession              // legacy simple session (for API compat)
	Events      chan Event                   // buffered SSE channel
	mu          sync.Mutex
	FilePath    string // last loaded file path
	FileContent string // file content (when uploaded via browser)
	toggles     *Toggles
	ProofStack  *ProofStack
}

// NewSession creates a new verification session with the given id.
func NewSession(id string) *Session {
	return &Session{
		ID:         id,
		Events:     make(chan Event, 64),
		Graph:      NewAnalysisGraphState(),
		SimpleSess: NewConceptSession(),
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
// It parses the file to extract types, relations, and actions
// to populate the concept domain and ARG.
func (s *Session) LoadFileContent(filename string, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if filename == "" {
		return fmt.Errorf("empty filename")
	}
	s.FilePath = filename
	s.FileContent = string(content)

	// Parse the Ivy file to extract declarations.
	version := lexer.Version{1, 7}
	// Check for #lang directive
	src := string(content)
	if strings.HasPrefix(src, "#lang ivy") {
		// skip the #lang line for the parser
		if idx := strings.Index(src, "\n"); idx >= 0 {
			src = src[idx+1:]
		}
	}

	p := parser.New(src, version)
	decls, _ := p.Parse() // best-effort: ignore parse errors

	// Extract types, relations, and actions from declarations.
	var sorts []string
	var relations []RelationInfo
	var actions []string

	for _, d := range decls {
		switch decl := d.(type) {
		case *ast.TypeDecl:
			// type client, type server
			// The first arg is a TypeDef whose Name is a Symbol
			if len(decl.Args()) > 0 {
				if td, ok := decl.Args()[0].(*ast.TypeDef); ok {
					if sym, ok := td.Name.(*ast.Symbol); ok {
						sorts = append(sorts, sym.Rep)
					} else if atom, ok := td.Name.(*ast.Atom); ok {
						sorts = append(sorts, atom.Rep)
					}
				} else if sym, ok := decl.Args()[0].(*ast.Symbol); ok {
					sorts = append(sorts, sym.Rep)
				} else if atom, ok := decl.Args()[0].(*ast.Atom); ok {
					sorts = append(sorts, atom.Rep)
				}
			}
		case *ast.RelationDecl:
			// relation link(X:client, Y:server)
			if len(decl.Args()) > 0 {
				if atom, ok := decl.Args()[0].(*ast.Atom); ok {
					ri := RelationInfo{Name: atom.Rep}
					for _, t := range atom.Terms {
						if v, ok := t.(*ast.Variable); ok {
							sortName := ""
							if v.VSort != nil {
								sortName = v.VSort.String()
							}
							ri.Params = append(ri.Params, ParamInfo{
								Name: v.Rep,
								Sort: sortName,
							})
						}
					}
					relations = append(relations, ri)
				}
			}
		case *ast.ActionDecl:
			if len(decl.Args()) > 0 {
				if ad, ok := decl.Args()[0].(*ast.ActionDef); ok {
					if len(ad.Args()) > 0 {
						if atom, ok := ad.Args()[0].(*ast.Atom); ok {
							actions = append(actions, atom.Rep)
						}
					}
				}
			}
		}
	}

	// Build BOTH concept domain types:
	// 1. Simple ConceptSession (for API JSON responses — string-based)
	// 2. ConceptInteractiveSession with CDConceptDomain (for Z3 — logic.Node-based)

	// --- Simple session (for API compat) ---
	s.SimpleSess = NewConceptSession()
	for _, sortName := range sorts {
		s.SimpleSess.Domain.Concepts[sortName] = &Concept{
			Name: sortName, Variables: []string{"X"},
			Formula: "X = X", Sorts: []string{sortName}, Arity: 1,
		}
		s.SimpleSess.Domain.Nodes = append(s.SimpleSess.Domain.Nodes, sortName)
	}
	for _, rel := range relations {
		var vars, sortList []string
		for _, p := range rel.Params {
			vars = append(vars, p.Name)
			sortList = append(sortList, p.Sort)
		}
		s.SimpleSess.Domain.Concepts[rel.Name] = &Concept{
			Name: rel.Name, Variables: vars, Arity: len(vars),
			Formula: rel.Name + "(" + strings.Join(vars, ", ") + ")",
			Sorts:   sortList,
		}
		if len(vars) == 1 {
			s.SimpleSess.Domain.NodeLabels = append(s.SimpleSess.Domain.NodeLabels, rel.Name)
		} else if len(vars) == 2 {
			s.SimpleSess.Domain.Edges = append(s.SimpleSess.Domain.Edges, rel.Name)
		}
	}

	// --- Real session with logic.Node formulas (for Z3) ---
	// Build sort and symbol maps for GetInitialConceptDomain
	sortMap := make(map[string]logic.Sort)
	for _, sortName := range sorts {
		sortMap[sortName] = &logic.UninterpretedSort{Name: sortName}
	}
	symbolMap := make(map[string]*logic.Const)
	for _, rel := range relations {
		var domSorts []logic.Sort
		for _, p := range rel.Params {
			s, ok := sortMap[p.Sort]
			if !ok {
				s = &logic.UninterpretedSort{Name: p.Sort}
				sortMap[p.Sort] = s
			}
			domSorts = append(domSorts, s)
		}
		domSorts = append(domSorts, logic.Boolean)
		fs, err := logic.NewFunctionSort(domSorts...)
		if err == nil {
			symbolMap[rel.Name] = logic.NewConst(rel.Name, fs)
		}
	}
	cdDomain := GetInitialConceptDomain(sortMap, symbolMap)
	s.ConceptSess = NewConceptInteractiveSession(
		cdDomain,
		nil,  // state: no state formula yet (will be set when viewing a state)
		nil,  // axioms
		nil,  // goal constraints
		nil,  // suppose constraints
		nil,  // widget
		nil,  // analysis session
		nil,  // cache
		false, // don't recompute yet (no state to abstract against)
	)

	// Build initial ARG with state 0
	s.Graph = NewAnalysisGraphState()
	s.Graph.States = append(s.Graph.States, ARGNode{
		ID:    0,
		Label: "0",
	})

	s.emit(Event{Type: "file_loaded", Data: map[string]interface{}{
		"filename":  filename,
		"size":      len(content),
		"sorts":     sorts,
		"relations": relationNames(relations),
		"actions":   actions,
	}})
	return nil
}

// RelationInfo describes a relation declaration.
type RelationInfo struct {
	Name   string
	Params []ParamInfo
}

// ParamInfo describes a parameter of a relation.
type ParamInfo struct {
	Name string
	Sort string
}

func relationNames(rels []RelationInfo) []string {
	names := make([]string, len(rels))
	for i, r := range rels {
		names[i] = r.Name
	}
	return names
}

// ExecuteAction runs a named verification action with the given arguments.
// Dispatches to the appropriate verification package based on action name.
func (s *Session) ExecuteAction(actionName string, args map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if actionName == "" {
		return fmt.Errorf("empty action name")
	}
	s.emit(Event{Type: "action_started", Data: map[string]interface{}{"action": actionName, "args": args}})

	var err error
	switch actionName {
	// --- Concept graph operations (ConceptInteractiveSession) ---
	case "undo":
		if s.ConceptSess != nil {
			err = s.ConceptSess.Undo()
		}
	case "redo":
		// Redo not yet implemented in ConceptInteractiveSession
		// Would need a redo stack
		err = fmt.Errorf("redo not yet implemented")
	case "recalculate":
		if s.ConceptSess != nil {
			s.ConceptSess.Recompute(nil)
		}
	case "gather":
		if s.ConceptSess != nil {
			// Gather facts from the current concept graph state
			_ = s.ConceptSess.GetFacts(nil)
		}
	case "backtrack":
		if s.ConceptSess != nil {
			err = s.ConceptSess.Undo() // backtrack = undo all to checkpoint
		}
	case "conjecture":
		// Generate a conjecture from the gathered facts
		if s.ConceptSess != nil {
			facts := s.ConceptSess.GetFacts(nil)
			if len(facts) > 0 {
				s.emit(Event{Type: "conjecture_generated", Data: map[string]interface{}{
					"facts": len(facts),
				}})
			}
		}
	case "remember":
		// Store current concept graph for later recall
		if s.ConceptSess != nil && s.ConceptSess.AnalysisSession != nil {
			s.ConceptSess.SaveDomain("remembered")
		}

	// --- Verification operations (check/art packages) ---
	case "pdr_step":
		// Single PDR strengthening step — uses Z3 via solver
		s.emit(Event{Type: "status", Data: map[string]string{"message": "PDR step: not yet wired to solver"}})
	case "concrete":
		// Compute concrete model — uses Z3
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Concrete: not yet wired to solver"}})
	case "reverse":
		// Compute reverse image — uses transrel + Z3
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Reverse: not yet wired to solver"}})
	case "path_reach", "reach":
		// Reachability analysis — uses art + Z3
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Reach: not yet wired to solver"}})
	case "weaken":
		// Weaken invariant
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Weaken: not yet wired"}})
	case "save_abstraction":
		// Save abstraction to file — export concept spaces
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Save abstraction: not yet wired"}})
	case "export":
		// Export current conjecture
		if s.ConceptSess != nil {
			facts := s.ConceptSess.GetFacts(nil)
			s.emit(Event{Type: "export", Data: map[string]interface{}{"facts": len(facts)}})
		}
	case "add_relation":
		// Add a relation from a user-entered formula string
		if formula, ok := args["formula"].(string); ok && formula != "" {
			s.emit(Event{Type: "status", Data: map[string]string{
				"message": "Add relation '" + formula + "': not yet wired to parser",
			}})
		}
	case "splatter":
		// Splatter: materialize all universe elements
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Splatter: not yet wired"}})

	default:
		// Unknown actions are accepted but logged — allows forward compatibility
		// as new verification operations are added.
		s.emit(Event{Type: "status", Data: map[string]string{
			"message": "Action '" + actionName + "' accepted (not yet wired to engine)",
		}})
	}

	status := "ok"
	if err != nil {
		status = err.Error()
	}
	s.emit(Event{Type: "action_completed", Data: map[string]interface{}{"action": actionName, "status": status}})
	return err
}

// ProofStack holds proof goal state for rendering.
// Stub: will be wired to the proof/ package.
func (s *Session) ProofStackData() *ProofStack {
	return s.ProofStack
}

// AddProjection adds a projection concept to the domain.
// Matches Python ConceptInteractiveSession.add_edge().
func (s *Session) AddProjection(name, concept string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ConceptSess != nil {
		// AddCustomEdge creates a projection in the concept domain
		s.ConceptSess.AddCustomEdge(name, concept, concept)
	}
	s.emit(Event{Type: "concept_updated", Data: nil})
	return nil
}

// ArgNodeAction executes an action on an ARG node.
// Dispatches to art/check/interp packages based on action name.
func (s *Session) ArgNodeAction(nodeID, action string, args map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := map[string]interface{}{
		"status": "ok",
		"node":   nodeID,
		"action": action,
	}

	switch action {
	case "view_state":
		// View state triggers concept graph update
		if s.ConceptSess != nil {
			s.ConceptSess.Recompute(nil)
		}
	case "check_safety":
		// Check safety at this node — would use check.CheckSafetyInState
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Safety check at node " + nodeID}})
	case "extend":
		// Find extension — would use art.AnalysisGraph.Execute
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Extend from node " + nodeID}})
	case "mark":
		// Mark node for covering
		result["marked"] = true
	case "cover":
		// Cover node by marked — would use art.AnalysisGraph.Cover
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Cover node " + nodeID}})
	case "join":
		// Join with marked node
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Join at node " + nodeID}})
	case "try_conjecture":
		// Try current conjecture at node
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Try conjecture at node " + nodeID}})
	case "try_remembered":
		// Try remembered graph at node
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Try remembered at node " + nodeID}})
	case "delete":
		// Delete node from ARG
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Delete node " + nodeID}})
	case "recalculate":
		// Recalculate edge
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Recalculate at " + nodeID}})
	case "decompose":
		// Step into / decompose edge
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Decompose at " + nodeID}})
	case "view_source":
		// View source code for this transition
		result["source"] = "// Source code viewer not yet wired"
		result["file"] = s.FilePath
	default:
		return result, fmt.Errorf("unknown ARG action: %s", action)
	}

	s.emit(Event{Type: "action_completed", Data: map[string]string{"action": action}})
	return result, nil
}

// ProofGoalAction executes an action on a proof goal node.
// Dispatches to the proof/ package based on action name.
func (s *Session) ProofGoalAction(goalID, action string) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := map[string]interface{}{
		"status": "ok",
		"goal":   goalID,
		"action": action,
	}

	switch action {
	case "view":
		// View proof goal details
		result["info"] = "Proof goal " + goalID
	case "apply_tactic":
		// Apply a proof tactic — requires proof.ProofChecker
		s.emit(Event{Type: "status", Data: map[string]string{
			"message": "Apply tactic at goal " + goalID + ": requires proof pipeline",
		}})
	case "refute":
		// Mark goal as refuted
		result["refuted"] = true
	default:
		return result, fmt.Errorf("unknown proof action: %s", action)
	}

	return result, nil
}

// SaveState serializes the current session state as JSON.
func (s *Session) SaveState() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, _ := json.Marshal(map[string]interface{}{
		"session_id":   s.ID,
		"file_path":    s.FilePath,
		"file_content": s.FileContent,
		"toggles":      s.toggles,
	})
	return data
}

// Toggles stores edge/label visibility checkbox state.
type Toggles struct {
	Edges map[string]map[string]bool `json:"edges"` // edge_name → {display_class → checked}
}

// GetToggles returns the current toggle state.
func (s *Session) GetToggles() *Toggles {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.toggles == nil {
		s.toggles = &Toggles{Edges: make(map[string]map[string]bool)}
	}
	return s.toggles
}

// SetToggle updates a single toggle value.
func (s *Session) SetToggle(edge, displayClass string, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.toggles == nil {
		s.toggles = &Toggles{Edges: make(map[string]map[string]bool)}
	}
	if s.toggles.Edges[edge] == nil {
		s.toggles.Edges[edge] = make(map[string]bool)
	}
	s.toggles.Edges[edge][displayClass] = value
}

// emit sends an event on the SSE channel (non-blocking drop if full).
func (s *Session) emit(e Event) {
	select {
	case s.Events <- e:
	default:
		// drop if buffer full
	}
}
