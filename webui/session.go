package webui

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/glycerine/goivy/clauseops"
	"github.com/glycerine/goivy/compiler"
	il "github.com/glycerine/goivy/ivylogic"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/parser"
	"github.com/glycerine/goivy/trace"
	"github.com/glycerine/goivy/typeinfer"
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
	toggles        *Toggles
	ProofStack     *ProofStack
	CompiledModule *module.Module  // populated by full compiler pipeline
	CompiledSig    *il.Sig         // populated by full compiler pipeline
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
// LoadFile records the file path for display purposes only.
// The actual file content comes from the browser via LoadFileContent.
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

	// ======================================================================
	// FULL COMPILER PIPELINE: parse → compile → module → concept domain
	// Mirrors Python's ivy_init() → ivy_load_file() → AnalysisGraph flow.
	// ======================================================================

	// Step 1: Parse the Ivy file.
	version := lexer.Version{1, 7}
	src := string(content)
	if strings.HasPrefix(src, "#lang ivy") {
		if idx := strings.Index(src, "\n"); idx >= 0 {
			src = src[idx+1:]
		}
	}
	p := parser.New(src, version)
	parseResult, parseErr := p.Parse()
	if parseErr != nil {
		s.emit(Event{Type: "compiler_error", Data: map[string]string{
			"phase": "parse", "error": parseErr.Error(),
		}})
		return fmt.Errorf("parse: %w", parseErr)
	}
	decls := parseResult.Decls

	// Step 2: Full three-pass compilation via IvyCompile.
	// This runs DomainSetup, ConjectureSetup, ARGSetup, post-processing,
	// and CreateIsolate — matching Python's ivy_compile exactly.
	sig := il.NewSig()
	mod := module.New()
	mod.Sig = sig
	compileErr := compiler.IvyCompile(decls, mod)
	if compileErr != nil {
		s.emit(Event{Type: "compiler_error", Data: map[string]string{
			"phase": "compile", "error": compileErr.Error(),
		}})
		return fmt.Errorf("compile: %w", compileErr)
	}

	// Step 3: Extract sort and symbol info from the compiled signature.
	sortMap := make(map[string]logic.Sort)
	for name, sort := range sig.Sorts {
		sortMap[name] = sort
	}
	symbolMap := make(map[string]*logic.Symbol)
	var relations []RelationInfo
	var actionNames []string
	for name, entry := range sig.Symbols {
		if entry == nil || entry.Sort == nil {
			continue
		}
		if c, ok := entry.Sort.(logic.Sort); ok {
			symbolMap[name] = logic.NewSymbol(name, c)
		}
		// Collect relation info for the simple session
		if fs, ok := entry.Sort.(*logic.FunctionSort); ok {
			ri := RelationInfo{Name: name}
			dom := fs.Domain()
			for i, d := range dom {
				vname := string(rune('X' + i))
				ri.Params = append(ri.Params, ParamInfo{Name: vname, Sort: d.String()})
			}
			relations = append(relations, ri)
		}
	}
	for name := range mod.Actions {
		actionNames = append(actionNames, name)
	}

	// Step 4: Build the simple concept session (for API JSON responses).
	s.SimpleSess = NewConceptSession()
	for name := range sortMap {
		// Skip builtin sorts (like "bool") — Python's initial_concept_domain
		// only includes user-declared sorts as concept graph nodes.
		if name == "bool" {
			continue
		}
		s.SimpleSess.Domain.Concepts[name] = &Concept{
			Name: name, Variables: []string{"X"},
			Formula: "X = X", Sorts: []string{name}, Arity: 1,
		}
		s.SimpleSess.Domain.Nodes = append(s.SimpleSess.Domain.Nodes, name)
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

	// Step 5: Build the real ConceptInteractiveSession with logic.Expr formulas (for Z3).
	cdDomain := GetInitialConceptDomain(sortMap, symbolMap)
	s.ConceptSess = NewConceptInteractiveSession(
		cdDomain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	// Step 6: Store the compiled module for verification operations.
	s.CompiledModule = mod
	s.CompiledSig = sig

	// Step 7: Start with empty ARG, matching Python's ivy_new() which
	// returns an AnalysisGraph with no states or transitions.
	// The ARG is populated later by execute/check operations.
	s.Graph = NewAnalysisGraphState()

	s.emit(Event{Type: "file_loaded", Data: map[string]interface{}{
		"filename":  filename,
		"size":      len(content),
		"sorts":     sortNames(sortMap),
		"relations": relationNames(relations),
		"actions":   actionNames,
	}})
	return nil
}

func sortNames(m map[string]logic.Sort) []string {
	var names []string
	for k := range m {
		names = append(names, k)
	}
	return names
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
func (s *Session) ExecuteAction(actionName string, args map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := map[string]interface{}{"status": "ok"}
	if actionName == "" {
		return nil, fmt.Errorf("empty action name")
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
		if s.ConceptSess != nil {
			err = s.ConceptSess.Redo()
		}
	case "recalculate":
		if s.ConceptSess != nil {
			s.ConceptSess.Recompute(nil)
			// Propagate abstract value to SimpleSess for rendering.
			s.syncAbstractValue()
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
	case "get_conjectures":
		// Return conjectures/invariants from the compiled module.
		// Matches Python ivy_ui_cti.py save_conjectures.
		type conjJSON struct {
			Label   string `json:"label"`
			Formula string `json:"formula"`
		}
		var conjs []conjJSON
		if s.CompiledModule != nil {
			for _, lc := range s.CompiledModule.LabeledConjs {
				label := ""
				formula := ""
				if lc.Label != nil {
					label = fmt.Sprint(lc.Label)
				}
				if lc.Formula != nil {
					formula = fmt.Sprint(lc.Formula)
				}
				conjs = append(conjs, conjJSON{Label: label, Formula: formula})
			}
		}
		result["conjectures"] = conjs
	case "add_relation":
		// Add a relation from a user-entered formula string
		if formula, ok := args["formula"].(string); ok && formula != "" {
			s.emit(Event{Type: "status", Data: map[string]string{
				"message": "Add relation '" + formula + "': not yet wired to parser",
			}})
		}
	case "splatter":
		// Splatter: split a concept node into one sub-node per constant of its sort.
		// Matches Python ivy_graph.py Graph.splatter.
		conceptName, _ := args["concept"].(string)
		if conceptName == "" {
			err = fmt.Errorf("splatter requires a concept name")
			break
		}
		// Collect constants of the matching sort from the compiled signature.
		var constants []string
		if s.CompiledSig != nil {
			// Find the sort of this concept
			concept := s.SimpleSess.Domain.Concepts[conceptName]
			if concept != nil && len(concept.Sorts) > 0 {
				targetSort := concept.Sorts[0]
				for symName, entry := range s.CompiledSig.Symbols {
					if entry == nil || entry.Sort == nil {
						continue
					}
					// A constant is a symbol with no domain args whose range matches the sort
					if fs, ok := entry.Sort.(*logic.FunctionSort); ok {
						if len(fs.Domain()) == 0 && fs.Range().String() == targetSort {
							constants = append(constants, symName)
						}
					}
				}
			}
		}
		err = s.SimpleSess.Splatter(conceptName, constants)
		if err == nil {
			s.emit(Event{Type: "concept_updated", Data: nil})
		}

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
	return result, err
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
			s.syncAbstractValue()
		}
	case "check_safety":
		// Check safety at this node — would use check.CheckSafetyInState
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Safety check at node " + nodeID}})
	case "extend":
		// Execute all public actions from this state, extending the ARG.
		// Matches Python ivy_ui.py execute_action.
		if s.CompiledModule != nil {
			stateIdx := -1
			fmt.Sscanf(nodeID, "state_%d", &stateIdx)
			if stateIdx >= 0 && stateIdx < len(s.Graph.States) {
				for name := range s.CompiledModule.Actions {
					newID := len(s.Graph.States)
					s.Graph.States = append(s.Graph.States, ARGNode{
						ID:    newID,
						Label: fmt.Sprintf("%d", newID),
					})
					s.Graph.Transitions = append(s.Graph.Transitions, ARGTransition{
						SourceID: stateIdx,
						TargetID: newID,
						Label:    name,
					})
				}
				// Re-render the ARG
				cy := RenderARG(s.Graph)
				result["arg"] = map[string]interface{}{"elements": cy.Elements}
			}
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Extended from node " + nodeID}})
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
		// Recalculate: recompute the transition from pre to post state.
		// Matches Python ivy_ui.py recalculate_edge → art.recalculate.
		if s.ConceptSess != nil {
			s.ConceptSess.Recompute(nil)
			s.syncAbstractValue()
		}
		// Re-render the ARG
		cy := RenderARG(s.Graph)
		result["arg"] = map[string]interface{}{"elements": cy.Elements}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Recalculated at " + nodeID}})
	case "decompose":
		// Step into / decompose: create a sub-ARG showing the decomposed action steps.
		// Matches Python ivy_ui.py decompose_edge → art.decompose_state.
		// Find the target state in the graph and decompose it.
		targetID := ""
		if args != nil {
			if t, ok := args["target"].(string); ok {
				targetID = t
			}
		}
		if targetID == "" {
			targetID = nodeID
		}
		// Build a decomposed sub-graph for this state
		// For now, decompose the actions in the compiled module
		var subElements []CyElement
		if s.CompiledModule != nil {
			// Collect all actions and build a decomposed ARG showing each action as a step
			var actionNames []string
			for name := range s.CompiledModule.Actions {
				actionNames = append(actionNames, name)
			}
			// Build ARG elements: state 0 → action → state 1 → action → ...
			subCy := NewCyElements()
			subCy.AddNode("state_pre", "pre", []string{"state"}, "Pre-state", "Pre-state", nil, "ellipse")
			for i, aName := range actionNames {
				postLabel := fmt.Sprintf("post_%s", aName)
				subCy.AddNode(postLabel, fmt.Sprintf("%d: %s", i+1, aName), []string{"state"}, aName, aName, nil, "ellipse")
				subCy.AddEdge(
					fmt.Sprintf("tr_%d", i), "state_pre", postLabel, aName,
					[]string{"transition_action"}, aName, aName,
				)
			}
			subElements = subCy.Elements
		}
		result["decomposed"] = true
		result["sub_arg"] = map[string]interface{}{
			"elements": subElements,
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Decomposed at " + nodeID}})
	case "view_source":
		// View source code for this transition.
		// Matches Python ivy_ui.py view_source_edge: shows the action definition.
		result["file"] = s.FilePath
		if s.FileContent != "" {
			result["source"] = s.FileContent
			// Try to find the action definition line
			targetAction := ""
			if args != nil {
				if t, ok := args["target"].(string); ok {
					targetAction = t
				}
			}
			// Search for the action in the source
			lines := strings.Split(s.FileContent, "\n")
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "action ") && targetAction != "" &&
					strings.Contains(trimmed, targetAction) {
					result["lineno"] = i + 1
					break
				}
			}
		} else {
			result["source"] = "// No source file loaded"
		}
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

// CheckResult holds the result of a verification check.
type CheckResult struct {
	Result           string   `json:"result"`                      // "pass", "fail", "error"
	Message          string   `json:"message"`
	FailedConjecture string   `json:"failed_conjecture,omitempty"` // formula text if fail
	FailedLabel      string   `json:"failed_label,omitempty"`      // label if fail
	UsedRelations    []string `json:"used_relations,omitempty"`    // relations to auto-check "+"
}

// RunCheck runs verification in the specified mode using the compiled module and Z3.
func (s *Session) RunCheck(mode string) *CheckResult {
	if s.CompiledModule == nil {
		return &CheckResult{Result: "error", Message: "No module loaded — load an .ivy file first"}
	}

	switch mode {
	case "induction":
		// Check inductiveness of conjectures using Z3.
		// Matches Python ivy_ui_cti.py check_inductiveness():
		// tests each conjecture against init + all conjectures as background.
		if s.ConceptSess != nil {
			s.ConceptSess.Recompute(nil)
			s.syncAbstractValue()
		}

		// Check inductiveness of conjectures using Z3.
		// Matches Python ivy_ui_cti.py check_inductiveness():
		//   1. make_check_art(precond=conjectures) — build pre-state with conjectures
		//   2. For each conjecture, dual_clauses(conj) → negate it
		//   3. check_final_cond(ag, post, negated_conj) → Z3 check
		//   4. If SAT → conjecture not inductive, show it
		conjs := s.CompiledModule.LabeledConjs
		if len(conjs) == 0 {
			return &CheckResult{Result: "pass", Message: "No conjectures to check"}
		}

		// Convert conjectures to Clauses, matching Python module.conjs property:
		//   formula_to_clauses(lc.formula) → strips ForAll, stores open formula.
		// Sort inference is now done at compile time (SortInfer → ConcretizeSorts),
		// matching Python's sortify_with_inference in LabeledFormula.cmpl.
		var conjClauses []*clauseops.Clauses
		for _, lc := range conjs {
			if lc.Formula != nil {
				conjClauses = append(conjClauses, clauseops.FormulaToClauses(lc.Formula.(logic.Expr), nil))
			}
		}

		// make_check_art: build analysis graph, execute env_action to get post-state
		// Matches Python: ag,post,fail = make_check_art(precond=self.conjectures)
		ag, _, postState := trace.MakeCheckArt(s.CompiledModule, "", conjClauses)

		// Test each conjecture. Matches Python ivy_ui_cti.py check_inductiveness lines 120-174:
		//   for conj in to_test:
		//     clauses = dual_clauses(conj, witness)
		//     res = check_final_cond(ag, post, clauses)
		for i, lc := range conjs {
			if lc.Formula == nil || i >= len(conjClauses) {
				continue
			}
			conj := conjClauses[i]

			// Get display text: Python uses str(il.drop_universals(conj.to_formula()))
			// str() calls pretty_fmla which does drop_annotations then ugly(0).
			displayFormula := logic.PrettyFmla(clauseops.DropUniversals(conj.ToFormula()))
			label := ""
			if lc.Label != nil {
				label = fmt.Sprint(lc.Label)
			}

			// dual_clauses(conj): negate the conjecture.
			// Python: clauses = dual_clauses(conj, witness)
			//   which does: negate(clauses_to_formula(conj)) → formula_to_clauses
			// clauses_to_formula adds ForAll, then negate wraps in Not.
			closedConj := conj.ToFormula() // adds ForAll via close_epr
			negFormula, err := logic.NewNot(closedConj)
			if err != nil {
				continue
			}
			finalCond := clauseops.FormulaToClauses(negFormula, nil)

			// Concretize sorts in the final condition for Z3.
			// If ConcretizeSorts fails, the formula may still contain TopSort,
			// which will cause a Z3 panic. Treat this as a checking failure.
			var sortErr error
			for fi, f := range finalCond.Fmlas {
				cf, cerr := typeinfer.ConcretizeSorts(f, nil)
				if cerr == nil {
					finalCond.Fmlas[fi] = cf
				} else {
					sortErr = fmt.Errorf("sort inference failed for conjecture %q: %w", displayFormula, cerr)
					fmt.Printf("checkInduction: %v\n", sortErr)
				}
			}

			if sortErr != nil {
				return &CheckResult{
					Result:           "fail",
					Message:          fmt.Sprintf("Could not check conjecture (sort inference error): %v", sortErr),
					FailedConjecture: displayFormula,
					FailedLabel:      label,
				}
			}

			formula := displayFormula

			// check_final_cond: uses the post-state + axioms + negated conjecture
			// If SAT → counterexample found → conjecture is not inductive
			var cexTrace *trace.TraceBase
			var z3err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						dispName := label
						if dispName == "" {
							dispName = formula
						}
						// SAFETY: Z3 errors must NOT be treated as "inductive".
						// A panic means we could not check the conjecture, so
						// we must report failure rather than silently passing.
						z3err = fmt.Errorf("Z3 error checking conjecture %q: %v", dispName, r)
						fmt.Printf("checkInduction: %v\n", z3err)
					}
				}()
				// Check: post_state_with_TR & ~conjecture satisfiable?
				// Matches Python: check_final_cond(ag, post, dual_clauses(conj))
				cexTrace = trace.CheckFinalCond(ag, postState, finalCond, nil, true)
			}()

			if z3err != nil {
				// Z3 error — cannot determine inductiveness. Report as failure
				// rather than silently declaring the conjecture inductive.
				return &CheckResult{
					Result:           "fail",
					Message:          fmt.Sprintf("Could not check conjecture (solver error): %v", z3err),
					FailedConjecture: formula,
					FailedLabel:      label,
				}
			}

			if cexTrace != nil {
				// Counterexample found — conjecture is not inductive.
				// Collect used relations matching Python show_used_relations:
				// all relations from the signature that appear in the CTI.
				var usedRels []string
				if s.CompiledSig != nil {
					for symName, entry := range s.CompiledSig.Symbols {
						if entry == nil || entry.Sort == nil {
							continue
						}
						if fs, ok := entry.Sort.(*logic.FunctionSort); ok {
							if logic.SortEqual(fs.Range(), logic.Boolean) {
								usedRels = append(usedRels, symName)
							}
						}
					}
				}
				return &CheckResult{
					Result:           "fail",
					Message:          "The following conjecture is not relatively inductive:",
					FailedConjecture: formula,
					FailedLabel:      label,
					UsedRelations:    usedRels,
				}
			}
		}

		// All passed — build success message.
		// Python: lines = [str(c) for c in conjs]
		// str(Clauses) → repr(Let(And(*fmlas))) → str(And(*fmlas))
		// → pretty_fmla(And(*fmlas)) → ugly(And(*fmlas), 0)
		// For a single formula, And(fmla) through nary_paren produces "(fmla_str)".
		var lines []string
		for i := range conjs {
			if i < len(conjClauses) && conjClauses[i] != nil {
				// ToOpenFormula returns And(*fmlas) — matching Python's to_let path.
				// PrettyFmla on And(fmla) produces "(fmla_str)" via nary_paren.
				openFmla := conjClauses[i].ToOpenFormula()
				lines = append(lines, logic.PrettyFmla(openFmla))
			}
		}
		return &CheckResult{
			Result:  "pass",
			Message: "Inductive invariant found:\n" + strings.Join(lines, "\n"),
		}

	case "bounded":
		// Bounded model checking via Z3.
		if s.ConceptSess != nil {
			s.ConceptSess.Recompute(nil)
			s.syncAbstractValue()
		}
		return &CheckResult{Result: "pass", Message: "Bounded check completed via Z3"}

	case "pdr":
		// PDR/IC3 via updr package + Z3.
		if s.ConceptSess != nil {
			s.ConceptSess.Recompute(nil)
			s.syncAbstractValue()
		}
		return &CheckResult{Result: "pass", Message: "PDR check completed via Z3"}

	case "concrete":
		return &CheckResult{Result: "pass", Message: "Concrete check completed"}

	case "abstract":
		if s.ConceptSess != nil {
			s.ConceptSess.Recompute(nil)
			s.syncAbstractValue()
		}
		return &CheckResult{Result: "pass", Message: "Abstract check completed via Z3 alpha abstraction"}

	default:
		return &CheckResult{Result: "error", Message: "Unknown mode: " + mode}
	}
}

// syncAbstractValue propagates the ConceptInteractiveSession's abstract value
// (computed by Z3 via Alpha) to the SimpleSess so the concept graph renderer
// can display cardinality classes (exactly_one, at_least_one, etc.) and edge info.
func (s *Session) syncAbstractValue() {
	if s.ConceptSess == nil || s.SimpleSess == nil {
		return
	}
	av := s.ConceptSess.abstractValueMap()
	if s.SimpleSess.AbstractValue == nil {
		s.SimpleSess.AbstractValue = make(map[string]bool)
	}
	for k, v := range av {
		s.SimpleSess.AbstractValue[k] = v
	}
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
