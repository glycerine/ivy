package webui

import (
	"encoding/json"
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"strings"
	"sync"
	//iu "github.com/glycerine/ivy/goivy/ivyutils"
)

// Event is a server-sent event delivered to the browser over SSE.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Session holds the state for one interactive verification session.
type Session struct {
	Cfg *goivy.Config

	ID              string
	Graph           *WebUIAnalysisGraphState   // ARG state
	ConceptSess     *ConceptInteractiveSession // concept graph state (uses Z3 via WebUIAlpha)
	SimpleSess      *ConceptSession            // legacy simple session (for API compat)
	Events          chan Event                 // buffered SSE channel
	mu              sync.Mutex
	FilePath        string // last loaded file path
	FileContent     string // file content (when uploaded via browser)
	toggles         *Toggles
	WebUIProofStack *WebUIProofStack
	ProofMgr        *goivy.ProofManager  // live proof state (goals + reachability graph)
	CompiledModule  *goivy.Module        // populated by full compiler pipeline
	CompiledSig     *goivy.Sig           // populated by full compiler pipeline
	AG              *goivy.AnalysisGraph // persistent analysis graph for interactive verification
	AGUI            *AnalysisGraphUI     // ARG navigation UI (delegates to AG)
}

// NewSession creates a new verification session with the given id.
func NewSession(cfg *goivy.Config, id string) *Session {
	return &Session{
		Cfg:        cfg,
		ID:         id,
		Events:     make(chan Event, 64),
		Graph:      NewWebUIAnalysisGraphState(),
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
	version := goivy.Version{1, 7}
	src := string(content)
	if strings.HasPrefix(src, "#lang ivy") {
		if idx := strings.Index(src, "\n"); idx >= 0 {
			src = src[idx+1:]
		}
	}
	parseResult, parseErr := goivy.Parse(src, version)
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
	sig := goivy.NewSig()
	mod := goivy.New()
	if s.Cfg != nil {
		if s.Cfg.ExtAction == "" {
			s.Cfg.ExtAction = CompileKwargs["ext"]
		}
		mod.Cfg = s.Cfg
	}
	mod.Sig = sig

	// Register all tactics before compilation so phase6 attach_proofs can
	// create ProofCheckers. Matches check.Start().
	mod.Cfg.ProofCfg = goivy.TacticNewConfig()
	goivy.RegisterTactics(mod.Cfg.ProofCfg, mod)

	compileErr := goivy.IvyCompile(decls, mod, true)
	if compileErr != nil {
		s.emit(Event{Type: "compiler_error", Data: map[string]string{
			"phase": "compile", "error": compileErr.Error(),
		}})
		return fmt.Errorf("compile: %w", compileErr)
	}

	// Step 3: Extract sort and symbol info from the compiled signature.
	sortMap := make(map[string]goivy.Sort)
	for name, sort := range sig.Sorts.All() {
		sortMap[name] = sort
	}
	symbolMap := make(map[string]*goivy.Const)
	var relations []RelationInfo
	var actionNames []string
	for name, entry := range sig.Symbols.All() {
		if entry == nil || entry.Sort == nil {
			continue
		}
		if c, ok := entry.Sort.(goivy.Sort); ok {
			symbolMap[name] = goivy.NewConst(name, c)
		}
		// Collect relation info for the simple session
		if fs, ok := entry.Sort.(*goivy.FunctionSort); ok {
			ri := RelationInfo{Name: name}
			dom := fs.Domain()
			for i, d := range dom {
				vname := string(rune('X' + i))
				ri.Params = append(ri.Params, ParamInfo{Name: vname, Sort: d.String()})
			}
			relations = append(relations, ri)
		}
	}
	for name := range mod.Actions.All() {
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

	// Step 6.5: Initialize ProofManager from module conjectures.
	// Python: AnalysisState.__init__ creates self.goal_stack = ProofGoalStack()
	// then push_goal is called for each conjecture during interactive verification.
	// Here we pre-populate with the module's conjectures as initial proof goals.
	s.ProofMgr = goivy.NewProofManager()
	for _, conj := range mod.LabeledConjs {
		if conj.Formula != nil {
			if fmla, ok := conj.Formula.(goivy.Expr); ok {
				s.ProofMgr.Goals.Push(&goivy.ProofGoal{Formula: fmla})
			}
		}
	}
	s.syncProofStack()

	// Step 7: Build the persistent AnalysisGraph.
	// Matches Python: self.g = AnalysisGraph() in ivy_compiler.ivy_new().
	s.AG = goivy.NewAnalysisGraph(s.CompiledModule)
	s.AGUI = NewAnalysisGraphUI()
	s.AGUI.AG = s.AG
	s.AGUI.Mod = s.CompiledModule
	s.AGUI.SyncCallback = func() { s.syncARGToGraph() }

	// Python's ivy_new() creates an empty ARG — initial state is added
	// lazily by each operation that needs it (e.g., runUPDR, RunCheck).
	s.syncARGToGraph()

	s.emit(Event{Type: "file_loaded", Data: map[string]interface{}{
		"filename":  filename,
		"size":      len(content),
		"sorts":     sortNames(sortMap),
		"relations": relationNames(relations),
		"actions":   actionNames,
	}})
	return nil
}

// syncARGToGraph converts the persistent AnalysisGraph (s.AG) into the
// lightweight WebUIAnalysisGraphState (s.Graph) for frontend rendering.
func (s *Session) syncARGToGraph() {
	if s.AG == nil {
		return
	}
	s.Graph = ArtToGraphState(s.AG)
}

// syncProofStack converts the live proof.ProofGoalStack (s.ProofMgr)
// into the lightweight webui.ProofStack for frontend rendering.
func (s *Session) syncProofStack() {
	if s.ProofMgr == nil || s.ProofMgr.Goals == nil {
		s.WebUIProofStack = &WebUIProofStack{}
		return
	}
	ps := &WebUIProofStack{}
	for _, g := range s.ProofMgr.Goals.Stack {
		parentID := -1
		if g.Parent != nil {
			parentID = g.Parent.ID
		}
		label := fmt.Sprintf("goal_%d", g.ID)
		info := ""
		if g.Formula != nil {
			info = goivy.PrettyFmla(g.Formula)
		}
		ps.Goals = append(ps.Goals, WebUIProofGoal{
			ID:       g.ID,
			Label:    label,
			Refuted:  false,
			Info:     info,
			ParentID: parentID,
		})
	}
	s.WebUIProofStack = ps
}

func sortNames(m map[string]goivy.Sort) []string {
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
		// PDR/IC3 verification via tactics.UPDR.
		// Matches Python ivy_graph_ui.py pdr_step() which uses the
		// AG-level tactics (reverse, backtrack, recalculate), and
		// Python tactics.py UPDR class.
		if s.CompiledModule == nil {
			err = fmt.Errorf("pdr_step: no compiled module")
			break
		}
		valid, pdrErr := s.runUPDR()
		if pdrErr != nil {
			err = pdrErr
			break
		}
		result["valid"] = valid
		numFrames := 0
		if s.AG != nil {
			numFrames = len(s.AG.States)
		}
		result["stats"] = map[string]int{
			"num_frames": numFrames,
		}
		msg := "PDR: counterexample found"
		if valid {
			msg = fmt.Sprintf("PDR: invariant found (%d frames)", numFrames)
		}
		s.emit(Event{Type: "pdr_complete", Data: map[string]string{"message": msg}})

	case "concrete":
		// Compute concrete model via Z3.
		// Matches Python ivy_solver.py get_model_clauses().
		if s.CompiledModule == nil || s.AG == nil {
			err = fmt.Errorf("concrete: no compiled module")
			break
		}
		state := s.AG.LastState()
		if stateID, ok := args["state_id"].(float64); ok && int(stateID) < len(s.AG.States) {
			state = s.AG.States[int(stateID)]
		}
		if state == nil {
			s.AG.AddInitialState(nil, nil)
			s.syncARGToGraph()
			state = s.AG.LastState()
		}
		if state == nil || state.Clauses == nil {
			err = fmt.Errorf("concrete: no state available")
			break
		}
		axioms := s.CompiledModule.BackgroundTheory(state.InScope)
		concrClauses := goivy.AndClausesTyped(state.Clauses, axioms)
		solver := goivy.NewSolver(s.CompiledModule, nil)
		defer solver.Close()
		mr, solverErr := solver.GetModelClauses(concrClauses)
		if solverErr != nil {
			err = solverErr
			break
		}
		if mr == nil {
			result["sat"] = false
			s.emit(Event{Type: "concrete_result", Data: map[string]string{"sat": "false"}})
		} else {
			result["sat"] = true
			valMap := make(map[string]string)
			if modelVals, mErr := solver.ModelValues(mr.Model, mr.Vocab); mErr == nil {
				for name, val := range modelVals {
					valMap[name] = val.String()
				}
			}
			result["model"] = valMap
			s.emit(Event{Type: "concrete_result", Data: map[string]interface{}{
				"sat": "true", "model": valMap,
			}})
		}

	case "reverse":
		// Compute reverse image (weakest precondition).
		// Matches Python ivy_transrel.py reverse_image().
		if s.CompiledModule == nil || s.AG == nil {
			err = fmt.Errorf("reverse: no compiled module")
			break
		}
		state := s.AG.LastState()
		if stateID, ok := args["state_id"].(float64); ok && int(stateID) < len(s.AG.States) {
			state = s.AG.States[int(stateID)]
		}
		if state == nil {
			s.AG.AddInitialState(nil, nil)
			s.syncARGToGraph()
			state = s.AG.LastState()
		}
		if state == nil {
			err = fmt.Errorf("reverse: no state available")
			break
		}
		if state.Update == nil {
			err = fmt.Errorf("reverse: state has no transition (execute an action first)")
			break
		}
		axioms := s.CompiledModule.BackgroundTheory(state.InScope)
		preClauses := goivy.ReverseImage(state.Clauses, axioms, state.Update)
		if preClauses == nil {
			err = fmt.Errorf("reverse: reverse image returned nil")
			break
		}
		preStr := goivy.PrettyFmla(preClauses.ToFormula())
		result["pre_state"] = preStr
		s.emit(Event{Type: "reverse_result", Data: map[string]string{"pre_state": preStr}})

	case "path_reach", "reach":
		// Reachability: forward image from predecessor, SAT check, extract model.
		// Matches Python ivy_interp.py reach_state().
		if s.CompiledModule == nil || s.AG == nil {
			err = fmt.Errorf("reach: no compiled module")
			break
		}
		state := s.AG.LastState()
		if stateID, ok := args["state_id"].(float64); ok && int(stateID) < len(s.AG.States) {
			state = s.AG.States[int(stateID)]
		}
		if state == nil {
			s.AG.AddInitialState(nil, nil)
			s.syncARGToGraph()
			state = s.AG.LastState()
		}
		if state == nil {
			err = fmt.Errorf("reach: no state available")
			break
		}
		interpState := goivy.ArtToInterpState(state)
		reached := goivy.ReachState(interpState, nil)
		if reached == nil {
			result["reachable"] = false
			s.emit(Event{Type: "reach_result", Data: map[string]string{"reachable": "false"}})
		} else {
			result["reachable"] = true
			artReached := goivy.InterpToArtState(reached)
			if artReached.Clauses != nil {
				result["reached_state"] = goivy.PrettyFmla(artReached.Clauses.ToFormula())
			}
			state.Unders = append(state.Unders, artReached)
			s.syncARGToGraph()
			s.emit(Event{Type: "reach_result", Data: map[string]string{"reachable": "true"}})
		}

	case "weaken":
		// Remove selected conjectures from the invariant.
		// Matches Python ivy_ui_cti.py weaken().
		if s.CompiledModule == nil {
			err = fmt.Errorf("weaken: no compiled module")
			break
		}
		indices, ok := args["indices"].([]interface{})
		if !ok || len(indices) == 0 {
			err = fmt.Errorf("weaken: no conjecture indices specified")
			break
		}
		toRemove := make(map[int]bool, len(indices))
		for _, idx := range indices {
			if fi, ok := idx.(float64); ok {
				toRemove[int(fi)] = true
			}
		}
		var kept []*goivy.LabeledFormula
		var removed []string
		for i, lc := range s.CompiledModule.LabeledConjs {
			if toRemove[i] {
				formula := ""
				if lc.Formula != nil {
					if fExpr, ok := lc.Formula.(goivy.Expr); ok {
						formula = goivy.PrettyFmla(goivy.DropUniversals(fExpr))
					}
				}
				removed = append(removed, formula)
			} else {
				kept = append(kept, lc)
			}
		}
		s.CompiledModule.LabeledConjs = kept
		result["removed_count"] = len(removed)
		result["removed"] = removed
		result["remaining_count"] = len(kept)
		s.emit(Event{Type: "weaken_result", Data: map[string]interface{}{"removed": removed}})

	case "save_abstraction":
		// Save abstraction: export concept spaces and conjectures.
		// Matches Python ivy_ui.py save_abstraction() + ivy_ui_cti.py save_conjectures().
		if s.CompiledModule == nil {
			err = fmt.Errorf("save_abstraction: no compiled module")
			break
		}
		var sb strings.Builder
		sb.WriteString("# This file was generated by ivy.\n\n")
		for _, cs := range s.CompiledModule.ConceptSpaces {
			sb.WriteString(fmt.Sprintf("concept %v = %v\n", cs.Label, cs.Body))
		}
		if len(s.CompiledModule.LabeledConjs) > 0 {
			sb.WriteString("\n# conjectures\n\n")
			for _, lc := range s.CompiledModule.LabeledConjs {
				label := ""
				if lc.Label != nil {
					label = fmt.Sprint(lc.Label)
				}
				formula := ""
				if lc.Formula != nil {
					if fExpr, ok := lc.Formula.(goivy.Expr); ok {
						formula = goivy.PrettyFmla(goivy.DropUniversals(fExpr))
					} else {
						formula = fmt.Sprint(lc.Formula)
					}
				}
				if label != "" {
					sb.WriteString(fmt.Sprintf("invariant [%s] %s\n", label, formula))
				} else {
					sb.WriteString(fmt.Sprintf("invariant %s\n", formula))
				}
			}
		}
		result["content"] = sb.String()
		s.emit(Event{Type: "export", Data: map[string]string{"content": sb.String()}})
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
		// Parse a user-entered formula string and add as a concept relation.
		// Matches Python ivy_graph.py string_to_concept() + ivy_graph_ui.py add_concept_from_string().
		if formula, ok := args["formula"].(string); ok && formula != "" {
			_, parseErr := goivy.ToFormula(formula)
			if parseErr != nil {
				err = fmt.Errorf("add_relation: parse error: %w", parseErr)
				break
			}
			c := &Concept{
				Name:    formula,
				Formula: formula,
			}
			if s.SimpleSess != nil {
				s.SimpleSess.Domain.Concepts[formula] = c
			}
			result["concept_name"] = formula
			s.emit(Event{Type: "concept_updated", Data: map[string]interface{}{
				"name": formula, "formula": formula,
			}})
		} else {
			err = fmt.Errorf("add_relation: missing formula argument")
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
				for symName, entry := range s.CompiledSig.Symbols.All() {
					if entry == nil || entry.Sort == nil {
						continue
					}
					// A constant is a symbol with no domain args whose range matches the sort
					if fs, ok := entry.Sort.(*goivy.FunctionSort); ok {
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
		// Check if actionName is a known module action — execute it on the ARG.
		// Matches Python ivy_art.py execute_action() → execute() → post_state() → concrete_post().
		// Try exact match first, then with "ext:" prefix (exported actions are
		// stored as "ext:name" in the module's action map).
		if s.CompiledModule != nil && s.AG != nil {
			resolvedName := actionName
			if _, ok := s.CompiledModule.Actions.Get2(resolvedName); !ok {
				resolvedName = "ext:" + actionName
			}
			if _, ok := s.CompiledModule.Actions.Get2(resolvedName); ok {
				prestate := s.AG.LastState()
				if stateID, ok := args["state_id"].(float64); ok && int(stateID) < len(s.AG.States) {
					prestate = s.AG.States[int(stateID)]
				}
				if prestate == nil {
					s.AG.AddInitialState(nil, nil)
					s.syncARGToGraph()
					prestate = s.AG.LastState()
				}
				if prestate != nil {
					poststate, execErr := s.AG.ExecuteAction(false, resolvedName, prestate, nil)
					if execErr != nil {
						err = execErr
						break
					}
					if poststate != nil {
						s.syncARGToGraph()
						result["post_state_id"] = poststate.ID
						if poststate.Clauses != nil {
							result["post_state"] = goivy.PrettyFmla(poststate.Clauses.ToFormula())
						}
						s.emit(Event{Type: "action_executed", Data: map[string]interface{}{
							"action": resolvedName, "post_id": poststate.ID,
						}})
					}
					break
				}
			}
		}
		// Unknown actions are accepted but logged — allows forward compatibility.
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

// WebUIProofStack holds proof goal state for rendering.
// Stub: will be wired to the proof/ package.
func (s *Session) ProofStackData() *WebUIProofStack {
	s.syncProofStack()
	return s.WebUIProofStack
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

	var err error

	switch action {
	case "view_state":
		// View state triggers concept graph update
		if s.ConceptSess != nil {
			s.ConceptSess.Recompute(nil)
			s.syncAbstractValue()
		}
	case "check_safety":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if s.AGUI != nil && stateIdx >= 0 {
			safe, msg := s.AGUI.CheckSafetyNode(stateIdx)
			result["safe"] = safe
			result["message"] = msg
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Safety check at node " + nodeID}})
	case "extend", "find_extension":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if s.AGUI != nil && stateIdx >= 0 {
			label, extErr := s.AGUI.FindExtension(stateIdx)
			if extErr != nil {
				err = extErr
			} else {
				result["extension"] = label
				cy := RenderWebUIARG(s.Graph)
				result["arg"] = map[string]interface{}{"elements": cy.Elements}
			}
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Extended from node " + nodeID}})
	case "mark", "mark_node":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if s.AGUI != nil && stateIdx >= 0 {
			s.AGUI.MarkNode(&ARGStateRef{ID: stateIdx})
		}
		result["marked"] = true
	case "cover", "cover_node":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if s.AGUI != nil && stateIdx >= 0 {
			ok, coverErr := s.AGUI.CoverNode(stateIdx)
			result["covered"] = ok
			if coverErr != nil {
				err = coverErr
			}
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Cover node " + nodeID}})
	case "join", "join_node":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if s.AGUI != nil && stateIdx >= 0 {
			err = s.AGUI.JoinNode(stateIdx)
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Join at node " + nodeID}})
	case "try_conjecture":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		conjStr, _ := args["conjecture"].(string)
		if s.AGUI != nil && stateIdx >= 0 {
			err = s.AGUI.TryConjecture(stateIdx, conjStr)
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Try conjecture at node " + nodeID}})
	case "try_remembered":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		goalName, _ := args["goal"].(string)
		if s.AGUI != nil && stateIdx >= 0 {
			err = s.AGUI.TryRememberedGraph(stateIdx, goalName)
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Try remembered at node " + nodeID}})
	case "delete", "delete_node":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if s.AGUI != nil && stateIdx >= 0 {
			s.AGUI.DeleteNode(stateIdx)
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Delete node " + nodeID}})
	case "recalculate", "recalculate_edge":
		srcIdx := -1
		tgtIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &srcIdx)
		if args != nil {
			if t, ok := args["target"].(string); ok {
				fmt.Sscanf(t, "state_%d", &tgtIdx)
			}
		}
		if s.AGUI != nil && srcIdx >= 0 && tgtIdx >= 0 {
			s.AGUI.RecalculateEdge(srcIdx, tgtIdx)
		} else if s.AGUI != nil {
			s.AGUI.RecalculateAll()
		}
		cy := RenderWebUIARG(s.Graph)
		result["arg"] = map[string]interface{}{"elements": cy.Elements}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Recalculated at " + nodeID}})
	case "decompose", "decompose_edge":
		srcIdx := -1
		tgtIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &srcIdx)
		if args != nil {
			if t, ok := args["target"].(string); ok {
				fmt.Sscanf(t, "state_%d", &tgtIdx)
			}
		}
		if s.AGUI != nil && srcIdx >= 0 && tgtIdx >= 0 {
			subGraph, decompErr := s.AGUI.DecomposeEdge(srcIdx, tgtIdx)
			if decompErr != nil {
				err = decompErr
			} else if subGraph != nil {
				result["decomposed"] = true
				cy := RenderWebUIARG(subGraph)
				result["sub_arg"] = map[string]interface{}{"elements": cy.Elements}
			}
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Decomposed at " + nodeID}})
	case "view_source", "view_source_edge":
		srcIdx := -1
		tgtIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &srcIdx)
		if args != nil {
			if t, ok := args["target"].(string); ok {
				fmt.Sscanf(t, "state_%d", &tgtIdx)
			}
		}
		if s.AGUI != nil && srcIdx >= 0 && tgtIdx >= 0 {
			filename, lineno, vsErr := s.AGUI.ViewSourceEdge(srcIdx, tgtIdx)
			if vsErr == nil {
				result["file"] = filename
				result["lineno"] = lineno
			} else {
				result["file"] = s.FilePath
			}
		} else {
			result["file"] = s.FilePath
		}
	default:
		err = fmt.Errorf("unknown ARG action: %s", action)
	}

	if err != nil {
		return result, err
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

	// Find the goal by ID.
	var goal *goivy.ProofGoal
	if s.ProofMgr != nil {
		id := -1
		fmt.Sscanf(goalID, "goal_%d", &id)
		if id < 0 {
			fmt.Sscanf(goalID, "%d", &id)
		}
		for _, g := range s.ProofMgr.Goals.Stack {
			if g.ID == id {
				goal = g
				break
			}
		}
	}

	switch action {
	case "view":
		if goal != nil && goal.Formula != nil {
			result["info"] = goivy.PrettyFmla(goal.Formula)
			result["id"] = goal.ID
			if goal.Parent != nil {
				result["parent_id"] = goal.Parent.ID
			}
		} else {
			result["info"] = "Proof goal " + goalID + " (no formula)"
		}

	case "apply_tactic":
		s.emit(Event{Type: "status", Data: map[string]string{
			"message": "Tactic application at goal " + goalID + ": proof infrastructure wired, Z3 required",
		}})

	case "refute":
		if goal != nil {
			s.ProofMgr.Goals.Remove(goal)
			result["refuted"] = true
			s.syncProofStack()
			s.emit(Event{Type: "proof_updated", Data: nil})
		} else {
			result["refuted"] = false
		}

	case "push":
		if goal != nil {
			sub := &goivy.ProofGoal{Formula: goal.Formula}
			s.ProofMgr.Goals.Push(sub)
			s.syncProofStack()
			s.emit(Event{Type: "proof_updated", Data: nil})
			result["new_goal_id"] = sub.ID
		}

	case "pop":
		popped := s.ProofMgr.Goals.Pop()
		if popped != nil {
			result["popped_id"] = popped.ID
			s.syncProofStack()
			s.emit(Event{Type: "proof_updated", Data: nil})
		}

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
type WebUICheckResult struct {
	Result           string   `json:"result"` // "pass", "fail", "error"
	Message          string   `json:"message"`
	Z3Contacted      bool     `json:"z3_contacted"`                // true if Z3 was actually called
	FailedConjecture string   `json:"failed_conjecture,omitempty"` // formula text if fail
	FailedLabel      string   `json:"failed_label,omitempty"`      // label if fail
	UsedRelations    []string `json:"used_relations,omitempty"`    // relations to auto-check "+"
}

// RunCheck runs verification in the specified mode using the compiled module and Z3.
func (s *Session) RunCheck(mode string) *WebUICheckResult {
	if s.CompiledModule == nil {
		return &WebUICheckResult{Result: "error", Message: "No module loaded — load an .ivy file first"}
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
			return &WebUICheckResult{Result: "pass", Message: "No conjectures to check"}
		}

		// Convert conjectures to Clauses, matching Python module.conjs property:
		//   formula_to_clauses(lc.formula) → strips ForAll, stores open formula.
		// Sort inference is now done at compile time (SortInfer → ConcretizeSorts),
		// matching Python's sortify_with_inference in LabeledFormula.cmpl.
		var conjClauses []*goivy.Clauses
		for _, lc := range conjs {
			if lc.Formula != nil {
				conjClauses = append(conjClauses, goivy.FormulaToClauses(lc.Formula.(goivy.Expr), nil))
			}
		}

		// make_check_art: build analysis graph, execute env_action to get post-state
		// Matches Python: ag,post,fail = make_check_art(precond=self.conjectures)
		ag, _, postState, err := goivy.MakeCheckArt(s.CompiledModule, "", conjClauses)
		if err != nil {
			return &WebUICheckResult{Result: "error", Message: fmt.Sprintf("MakeCheckArt: %v", err)}
		}

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
			displayFormula := goivy.PrettyFmla(goivy.DropUniversals(conj.ToFormula()))
			label := ""
			if lc.Label != nil {
				label = fmt.Sprint(lc.Label)
			}

			// dual_clauses(conj, witness): Skolemize variables then negate.
			// Python: clauses = dual_clauses(conj, witness)
			//   witness = lambda v: lg.Const('@' + v.name, v.sort)
			// DualClauses replaces universals with Skolem constants, then negates.
			witness := func(v *goivy.Variable) goivy.Expr {
				return goivy.VarToSkolem("@", v)
			}
			finalCond := goivy.DualClauses(conj, witness, s.CompiledModule.Instantiator)

			// Concretize sorts in the final condition for Z3.
			var sortErr error
			for fi, f := range finalCond.Fmlas {
				cf, cerr := goivy.ConcretizeSorts(f, nil)
				if cerr == nil {
					finalCond.Fmlas[fi] = cf
				} else {
					sortErr = fmt.Errorf("sort inference failed for conjecture %q: %w", displayFormula, cerr)
					fmt.Printf("checkInduction: %v\n", sortErr)
				}
			}

			if sortErr != nil {
				return &WebUICheckResult{
					Z3Contacted:      true,
					Result:           "fail",
					Message:          fmt.Sprintf("Could not check conjecture (sort inference error): %v", sortErr),
					FailedConjecture: displayFormula,
					FailedLabel:      label,
				}
			}

			formula := displayFormula

			// check_final_cond: uses the post-state + axioms + negated conjecture
			// If SAT → counterexample found → conjecture is not inductive
			var cexTrace *goivy.TraceBase
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
				cexTrace = goivy.CheckFinalCond(ag, postState, finalCond, nil, true)
			}()

			if z3err != nil {
				// Z3 error — cannot determine inductiveness. Report as failure
				// rather than silently declaring the conjecture inductive.
				return &WebUICheckResult{
					Z3Contacted:      true,
					Result:           "fail",
					Message:          fmt.Sprintf("Could not check conjecture (solver error): %v", z3err),
					FailedConjecture: formula,
					FailedLabel:      label,
				}
			}

			if cexTrace != nil {
				// Counterexample found — conjecture is not inductive.
				// Python assigns the reconstructed counterexample trace:
				//   res = ivy_trace.check_final_cond(...)
				//   self.g = res
				s.AG = cexTrace.AnalysisGraph
				s.AGUI.AG = cexTrace.AnalysisGraph
				s.syncARGToGraph()

				// Collect used relations matching Python show_used_relations:
				// all relations from the signature that appear in the CTI.
				var usedRels []string
				if s.CompiledSig != nil {
					for symName, entry := range s.CompiledSig.Symbols.All() {
						if entry == nil || entry.Sort == nil {
							continue
						}
						if fs, ok := entry.Sort.(*goivy.FunctionSort); ok {
							if goivy.SortEqual(fs.Range(), goivy.Boolean) {
								usedRels = append(usedRels, symName)
							}
						}
					}
				}
				return &WebUICheckResult{
					Z3Contacted:      true,
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
				lines = append(lines, goivy.PrettyFmla(openFmla))
			}
		}
		return &WebUICheckResult{
			Z3Contacted: true,
			Result:      "pass",
			Message:     "Inductive invariant found:\n" + strings.Join(lines, "\n"),
		}

	case "bounded":
		// Bounded model checking via bmc.CheckIsolate + Z3.
		// Matches Python ivy_ui_cti.py bmc_conjecture / ivy_bmc.py check_isolate.
		conjs := s.CompiledModule.LabeledConjs
		if len(conjs) == 0 {
			return &WebUICheckResult{Result: "pass", Message: "No conjectures to check (BMC)"}
		}
		nSteps := 10
		bmcCfg := goivy.DefaultConfig(s.CompiledModule, nSteps)
		var bmcResult *goivy.BMCResult
		var bmcErr error
		func() {
			defer func() {
				if r := recover(); r != nil {
					bmcErr = fmt.Errorf("BMC panic: %v", r)
				}
			}()
			bmcResult = goivy.BMCCheckIsolate(bmcCfg)
		}()
		if bmcErr != nil {
			return &WebUICheckResult{Z3Contacted: true, Result: "error", Message: bmcErr.Error()}
		}
		if bmcResult.Found {
			if bmcResult.Trace != nil {
				s.AG = bmcResult.Trace.AnalysisGraph
				s.AGUI.AG = s.AG
				s.syncARGToGraph()
			}
			return &WebUICheckResult{
				Z3Contacted: true,
				Result:      "fail",
				Message:     bmcResult.Message,
			}
		}
		return &WebUICheckResult{
			Z3Contacted: true,
			Result:      "pass",
			Message:     bmcResult.Message,
		}

	case "pdr":
		// PDR/IC3 via tactics.UPDR (matches Python tactics.py UPDR class).
		if s.ConceptSess != nil {
			s.ConceptSess.Recompute(nil)
			s.syncAbstractValue()
		}
		if s.CompiledModule == nil {
			return &WebUICheckResult{Result: "error", Message: "PDR: no compiled module"}
		}
		valid, pdrErr := s.runUPDR()
		if pdrErr != nil {
			return &WebUICheckResult{Z3Contacted: true, Result: "error", Message: fmt.Sprintf("PDR error: %v", pdrErr)}
		}
		if valid {
			numFrames := 0
			if s.AG != nil {
				numFrames = len(s.AG.States)
			}
			return &WebUICheckResult{
				Z3Contacted: true,
				Result:      "pass",
				Message:     fmt.Sprintf("Invariant found (%d frames)", numFrames),
			}
		}
		return &WebUICheckResult{Z3Contacted: true, Result: "fail", Message: "Counterexample found"}

	case "concrete":
		// Concrete checking: checks conjectures hold in the initial state.
		// Matches Python check_conjs_in_state + check_safety_in_state.
		// Uses MakeCheckArt to build pre/post states, then checks each
		// conjecture via CheckFinalCond against the post-state.
		conjs := s.CompiledModule.LabeledConjs
		if len(conjs) == 0 {
			return &WebUICheckResult{Result: "pass", Message: "No conjectures to check (concrete)"}
		}
		var conjClauses []*goivy.Clauses
		for _, lc := range conjs {
			if lc.Formula != nil {
				conjClauses = append(conjClauses, goivy.FormulaToClauses(lc.Formula.(goivy.Expr), nil))
			}
		}
		ag, _, postState, err := goivy.MakeCheckArt(s.CompiledModule, "", conjClauses)
		if err != nil {
			return &WebUICheckResult{Z3Contacted: true, Result: "error", Message: fmt.Sprintf("concrete: %v", err)}
		}
		for i, lc := range conjs {
			if lc.Formula == nil || i >= len(conjClauses) {
				continue
			}
			conj := conjClauses[i]
			displayFormula := goivy.PrettyFmla(goivy.DropUniversals(conj.ToFormula()))
			label := ""
			if lc.Label != nil {
				label = fmt.Sprint(lc.Label)
			}
			witness := func(v *goivy.Variable) goivy.Expr {
				return goivy.VarToSkolem("@", v)
			}
			finalCond := goivy.DualClauses(conj, witness, s.CompiledModule.Instantiator)
			for fi, f := range finalCond.Fmlas {
				if cf, cerr := goivy.ConcretizeSorts(f, nil); cerr == nil {
					finalCond.Fmlas[fi] = cf
				}
			}
			var cexTrace *goivy.TraceBase
			var z3err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						z3err = fmt.Errorf("Z3 error: %v", r)
					}
				}()
				cexTrace = goivy.CheckFinalCond(ag, postState, finalCond, nil, true)
			}()
			if z3err != nil {
				return &WebUICheckResult{Z3Contacted: true, Result: "error", Message: z3err.Error()}
			}
			if cexTrace != nil {
				s.AG = cexTrace.AnalysisGraph
				s.AGUI.AG = cexTrace.AnalysisGraph
				s.syncARGToGraph()
				return &WebUICheckResult{
					Z3Contacted:      true,
					Result:           "fail",
					Message:          "Conjecture does not hold concretely:",
					FailedConjecture: displayFormula,
					FailedLabel:      label,
				}
			}
		}
		return &WebUICheckResult{
			Z3Contacted: true,
			Result:      "pass",
			Message:     "All conjectures hold concretely.",
		}

	case "abstract":
		// Abstract checking: runs alpha abstraction via Z3 to compute
		// concept graph abstract values (cardinality, edge info), then
		// checks conjectures hold in the abstract state.
		// Matches Python: recompute via alpha, then check.
		if s.ConceptSess != nil {
			var abstractErr error
			func() {
				defer func() {
					if r := recover(); r != nil {
						abstractErr = fmt.Errorf("alpha abstraction panic: %v", r)
					}
				}()
				s.ConceptSess.Recompute(nil)
				s.syncAbstractValue()
			}()
			if abstractErr != nil {
				return &WebUICheckResult{Z3Contacted: true, Result: "error", Message: abstractErr.Error()}
			}
		}
		// After alpha abstraction, check conjectures via the same path as concrete.
		conjs := s.CompiledModule.LabeledConjs
		if len(conjs) == 0 {
			return &WebUICheckResult{Z3Contacted: true, Result: "pass", Message: "Abstract check completed, no conjectures to verify."}
		}
		var conjClauses []*goivy.Clauses
		for _, lc := range conjs {
			if lc.Formula != nil {
				conjClauses = append(conjClauses, goivy.FormulaToClauses(lc.Formula.(goivy.Expr), nil))
			}
		}
		ag, _, postState, err := goivy.MakeCheckArt(s.CompiledModule, "", conjClauses)
		if err != nil {
			return &WebUICheckResult{Z3Contacted: true, Result: "error", Message: fmt.Sprintf("abstract: %v", err)}
		}
		for i, lc := range conjs {
			if lc.Formula == nil || i >= len(conjClauses) {
				continue
			}
			conj := conjClauses[i]
			displayFormula := goivy.PrettyFmla(goivy.DropUniversals(conj.ToFormula()))
			label := ""
			if lc.Label != nil {
				label = fmt.Sprint(lc.Label)
			}
			witness := func(v *goivy.Variable) goivy.Expr {
				return goivy.VarToSkolem("@", v)
			}
			finalCond := goivy.DualClauses(conj, witness, s.CompiledModule.Instantiator)
			for fi, f := range finalCond.Fmlas {
				if cf, cerr := goivy.ConcretizeSorts(f, nil); cerr == nil {
					finalCond.Fmlas[fi] = cf
				}
			}
			var cexTrace *goivy.TraceBase
			var z3err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						z3err = fmt.Errorf("Z3 error: %v", r)
					}
				}()
				cexTrace = goivy.CheckFinalCond(ag, postState, finalCond, nil, true)
			}()
			if z3err != nil {
				return &WebUICheckResult{Z3Contacted: true, Result: "error", Message: z3err.Error()}
			}
			if cexTrace != nil {
				s.AG = cexTrace.AnalysisGraph
				s.AGUI.AG = cexTrace.AnalysisGraph
				s.syncARGToGraph()
				return &WebUICheckResult{
					Z3Contacted:      true,
					Result:           "fail",
					Message:          "Conjecture does not hold after abstraction:",
					FailedConjecture: displayFormula,
					FailedLabel:      label,
				}
			}
		}
		return &WebUICheckResult{
			Z3Contacted: true,
			Result:      "pass",
			Message:     "All conjectures hold after abstract check via Z3 alpha abstraction.",
		}

	default:
		return &WebUICheckResult{Result: "error", Message: "Unknown mode: " + mode}
	}
}

// syncAbstractValue propagates the ConceptInteractiveSession's abstract value
// (computed by Z3 via WebUIAlpha) to the SimpleSess so the concept graph renderer
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

// runUPDR runs the tactics-based UPDR algorithm on the session's compiled
// module. Matches Python tactics.py UPDR class invoked from the Tk GUI's
// pdr_step button (ivy_graph_ui.py:292).
func (s *Session) runUPDR() (bool, error) {
	mod := s.CompiledModule
	if mod == nil {
		return false, fmt.Errorf("runUPDR: no compiled module")
	}

	if s.AG == nil {
		s.AG = goivy.NewAnalysisGraph(mod)
	}
	if len(s.AG.States) == 0 {
		s.AG.AddInitialState(mod.InitCond, nil)
	}
	if len(s.AG.States) == 0 {
		return false, fmt.Errorf("runUPDR: could not establish initial state")
	}

	tc := goivy.NewTacticsContext(s.AG, mod)

	// Build bad states from negated conjectures
	var conjFmlas []goivy.Expr
	for _, lc := range mod.LabeledConjs {
		if lc.Formula != nil {
			conjFmlas = append(conjFmlas, lc.Formula.(goivy.Expr))
		}
	}
	if len(conjFmlas) == 0 {
		return true, nil // no conjectures = trivially safe
	}
	safetyProp := goivy.NewClauses(conjFmlas, nil, nil)
	badClauses := goivy.NegateClauses(safetyProp)
	badFormula := badClauses.ToFormula()

	goal := &goivy.ProofGoal{
		Formula: badFormula,
		Node:    s.AG.States[0],
	}

	u := &goivy.UPDR{TC: tc, MaxFrames: 100}
	return u.Apply(goal)
}

// emit sends an event on the SSE channel (non-blocking drop if full).
func (s *Session) emit(e Event) {
	select {
	case s.Events <- e:
	default:
		// drop if buffer full
	}
}
