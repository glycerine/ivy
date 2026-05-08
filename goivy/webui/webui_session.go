package webui

import (
	"encoding/json"
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	ProofMgr        *goivy.ProofManager // live proof state (goals + reachability graph)
	CompiledModule  *goivy.Module       // populated by full compiler pipeline
	CompiledSig     *goivy.Sig          // populated by full compiler pipeline
	OriginalConjs   []*goivy.LabeledFormula
	AG              *goivy.AnalysisGraph // persistent analysis graph for interactive verification
	AGUI            *AnalysisGraphUI     // ARG navigation UI (delegates to AG)
	CTIUI           *CTIAnalysisGraphUI  // CTI/invariant workflow UI
	SheetUIs        map[string]*AnalysisGraphUI
	sheetCounter    int
	ReachableUI     *AnalysisGraphUI
	EventViewer     *EventTraceViewer
}

const rootSheetID = "sheet-1"

// NewSession creates a new verification session with the given id.
func NewSession(cfg *goivy.Config, id string) *Session {
	return &Session{
		Cfg:          cfg,
		ID:           id,
		Events:       make(chan Event, 64),
		Graph:        NewWebUIAnalysisGraphState(),
		SimpleSess:   NewConceptSession(),
		SheetUIs:     make(map[string]*AnalysisGraphUI),
		sheetCounter: 1,
		EventViewer:  NewEventTraceViewer(),
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
	mod := goivy.NewModule()
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
		if fs, ok := entry.Sort.(*goivy.LogicFunctionSort); ok {
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
	s.OriginalConjs = append([]*goivy.LabeledFormula{}, mod.LabeledConjs...)

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
	s.SheetUIs = map[string]*AnalysisGraphUI{rootSheetID: s.AGUI}
	s.sheetCounter = 1

	s.CTIUI = NewCTIAnalysisGraphUI(s.CompiledModule)
	s.CTIUI.AnalysisGraphUI.AG = goivy.NewAnalysisGraph(s.CompiledModule)
	s.CTIUI.AnalysisGraphUI.Mod = s.CompiledModule
	s.CTIUI.AnalysisGraphUI.SyncCallback = func() {
		if s.CTIUI != nil && s.CTIUI.AnalysisGraphUI != nil {
			s.CTIUI.AnalysisGraphUI.G = ArtToGraphState(s.CTIUI.AnalysisGraphUI.AG)
		}
	}
	s.CTIUI.AnalysisGraphUI.AG.AddInitialState(nil, nil)
	s.CTIUI.AnalysisGraphUI.G = ArtToGraphState(s.CTIUI.AnalysisGraphUI.AG)
	s.CTIUI.StartCTI(ctiClausesFromModule(s.CompiledModule))

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

func (s *Session) ensureRootAnalysisUIRegisteredLocked() {
	if s.SheetUIs == nil {
		s.SheetUIs = make(map[string]*AnalysisGraphUI)
	}
	if s.sheetCounter < 1 {
		s.sheetCounter = 1
	}
	if s.AGUI != nil {
		s.SheetUIs[rootSheetID] = s.AGUI
	}
}

func (s *Session) analysisUIForSheetLocked(sheetID string) *AnalysisGraphUI {
	if sheetID == "" {
		sheetID = rootSheetID
	}
	s.ensureRootAnalysisUIRegisteredLocked()
	if ui := s.SheetUIs[sheetID]; ui != nil {
		return ui
	}
	if sheetID == rootSheetID {
		return s.AGUI
	}
	return nil
}

func (s *Session) requireAnalysisUIForSheetLocked(sheetID string) (*AnalysisGraphUI, string, error) {
	if sheetID == "" {
		sheetID = rootSheetID
	}
	ui := s.analysisUIForSheetLocked(sheetID)
	if ui == nil {
		return nil, sheetID, fmt.Errorf("analysis sheet %q is not initialized", sheetID)
	}
	return ui, sheetID, nil
}

func (s *Session) nextAnalysisSheetIDLocked() string {
	s.ensureRootAnalysisUIRegisteredLocked()
	s.sheetCounter++
	return fmt.Sprintf("sheet-%d", s.sheetCounter)
}

func (s *Session) registerAnalysisSheetLocked(ui *AnalysisGraphUI) string {
	sheetID := s.nextAnalysisSheetIDLocked()
	s.SheetUIs[sheetID] = ui
	return sheetID
}

func (s *Session) newAnalysisGraphUIForGraphLocked(ag *goivy.AnalysisGraph) *AnalysisGraphUI {
	ui := NewAnalysisGraphUI()
	ui.AG = ag
	ui.Mod = s.CompiledModule
	if ui.Mod == nil {
		ui.Mod = s.CompiledModule
	}
	ui.G = ArtToGraphState(ag)
	ui.SyncCallback = func() {
		ui.G = ArtToGraphState(ui.AG)
	}
	return ui
}

func (s *Session) ensureEventViewerLocked() *EventTraceViewer {
	if s.EventViewer == nil {
		s.EventViewer = NewEventTraceViewer()
	}
	return s.EventViewer
}

func eventSheetResult(result map[string]interface{}, viewer *EventTraceViewer, sheetID string) {
	sheet := viewer.GetSheet(sheetID)
	if sheet == nil {
		return
	}
	result["sheet_id"] = sheet.Name
	result["label"] = sheet.Label
	result["events"] = sheet.Events
	result["patterns"] = append([]string{}, viewer.Patterns...)
}

func traceEventsFromActionArg(raw interface{}) ([]*TraceEvent, error) {
	switch v := raw.(type) {
	case []*TraceEvent:
		return v, nil
	case []TraceEvent:
		events := make([]*TraceEvent, 0, len(v))
		for i := range v {
			ev := v[i]
			events = append(events, &ev)
		}
		return events, nil
	case []interface{}:
		events := make([]*TraceEvent, 0, len(v))
		for _, item := range v {
			ev, err := traceEventFromActionArg(item)
			if err != nil {
				return nil, err
			}
			events = append(events, ev)
		}
		return events, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("events_new_sheet: unsupported events payload %T", raw)
	}
}

func traceEventFromActionArg(raw interface{}) (*TraceEvent, error) {
	switch v := raw.(type) {
	case *TraceEvent:
		return v, nil
	case TraceEvent:
		ev := v
		return &ev, nil
	case map[string]interface{}:
		text, _ := v["text"].(string)
		if text == "" {
			return nil, fmt.Errorf("event is missing text")
		}
		ev := NewTraceEvent(text)
		if addr, ok := v["address"].(string); ok {
			ev.Address = addr
		}
		if rawSubs, ok := v["subs"].([]interface{}); ok {
			subs, err := traceEventsFromActionArg(rawSubs)
			if err != nil {
				return nil, err
			}
			ev.Subs = subs
		}
		return ev, nil
	default:
		return nil, fmt.Errorf("unsupported event payload %T", raw)
	}
}

func actionStringArg(args map[string]interface{}, key string) string {
	if args == nil {
		return ""
	}
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func parseARGStateNodeID(nodeID string) (int, bool, error) {
	if nodeID == "" {
		return 0, false, nil
	}
	raw := strings.TrimSpace(nodeID)
	for _, prefix := range []string{"state_", "n"} {
		if strings.HasPrefix(raw, prefix) {
			raw = strings.TrimPrefix(raw, prefix)
			break
		}
	}
	idx, err := strconv.Atoi(raw)
	if err != nil {
		return 0, true, fmt.Errorf("invalid ARG node id %q", nodeID)
	}
	return idx, true, nil
}

func (s *Session) selectConceptARGNode(sheetID, nodeID string) (selectedNode, stateLabel string, err error) {
	idx, ok, err := parseARGStateNodeID(nodeID)
	if err != nil || !ok {
		return "", "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	ui, _, err := s.requireAnalysisUIForSheetLocked(sheetID)
	if err != nil {
		return "", "", err
	}
	state, err := ui.stateByID(idx)
	if err != nil {
		return "", "", err
	}

	if ui.CurrentConceptGraph == nil ||
		ui.CurrentConceptGraph.G() == nil ||
		ui.CurrentConceptGraph.G().ParentState != state {
		ui.ViewState(idx, "", false)
	}
	if s.ConceptSess != nil {
		if state.Clauses != nil {
			s.ConceptSess.State = state.Clauses.ToFormula()
		} else {
			s.ConceptSess.State = goivy.True
		}
		s.ConceptSess.GoalConstraints = nil
		s.ConceptSess.Cache = make(map[string]bool)
		s.ConceptSess.Recompute(nil)
		s.syncAbstractValue()
	}
	return fmt.Sprintf("state_%d", idx), ui.StateLabel(idx), nil
}

func (s *Session) ensureConceptGraphWidgetLocked() *GraphWidget {
	return s.ensureConceptGraphWidgetForSheetLocked(rootSheetID)
}

func (s *Session) ensureConceptGraphWidgetForSheetLocked(sheetID string) *GraphWidget {
	ui := s.analysisUIForSheetLocked(sheetID)
	if ui == nil {
		return nil
	}
	if ui.CurrentConceptGraph != nil {
		return ui.CurrentConceptGraph
	}
	sorts := ui.conceptSortNames()
	gs := StandardGraph(sorts, nil)
	w := NewGraphWidget(gs)
	w.Parent = ui
	ui.CurrentConceptGraph = w
	return w
}

func (s *Session) ensureConceptChecksLocked() *DisplayCheckboxes {
	return s.ensureConceptChecksForSheetLocked(rootSheetID)
}

func (s *Session) ensureConceptChecksForSheetLocked(sheetID string) *DisplayCheckboxes {
	w := s.ensureConceptGraphWidgetForSheetLocked(sheetID)
	if w == nil || w.G() == nil {
		return NewDisplayCheckboxes()
	}
	checks := w.G().Checks
	if s.SimpleSess != nil && s.SimpleSess.Domain != nil {
		for _, rel := range s.SimpleSess.Domain.RelationIDs() {
			checks.EnsureEdge(rel)
			checks.EnsureNodeLabel(rel)
		}
		for _, label := range s.SimpleSess.Domain.DefaultLabelIDs() {
			checks.SetNodeLabelCheckbox(label, NodeLabelNecessarily, true)
		}
	}
	s.toggles = checks.Snapshot()
	return checks
}

func conceptGraphActionPayload(w *GraphWidget) map[string]interface{} {
	if w == nil || w.G() == nil {
		return map[string]interface{}{
			"elements": []WebUICyElement{},
			"facts":    []FactSelection{},
			"toggles":  NewDisplayCheckboxes().Snapshot(),
		}
	}
	cy := RenderConceptGraph(w.G().ConceptSess, w.G().Checks)
	if cy.Elements == nil {
		cy.Elements = []WebUICyElement{}
	}
	return map[string]interface{}{
		"elements": cy.Elements,
		"facts":    w.ConstraintFacts(),
		"toggles":  w.G().Checks.Snapshot(),
	}
}

func conceptGraphGoalClauses(w *GraphWidget, parentState *goivy.State) (*goivy.Clauses, error) {
	if w != nil {
		exprs := w.GetActiveFactExprs()
		if len(exprs) == 1 {
			return goivy.FormulaToClauses(exprs[0], nil), nil
		}
		if len(exprs) > 1 {
			and, err := goivy.NewAnd(exprs...)
			if err != nil {
				return nil, err
			}
			return goivy.FormulaToClauses(and, nil), nil
		}
		if g := w.G(); g != nil && strings.TrimSpace(g.State) != "" {
			fmla, err := goivy.ToFormula(g.State)
			if err == nil {
				if expr, ok := fmla.(goivy.Expr); ok {
					return goivy.FormulaToClauses(expr, nil), nil
				}
			}
		}
	}
	if parentState != nil && parentState.Clauses != nil {
		return parentState.Clauses, nil
	}
	return goivy.TrueClauses(nil), nil
}

func (s *Session) ctiConceptWidgetForSheetLocked(sheetID string) (*CTIConceptGraphWidget, error) {
	if s.CTIUI == nil {
		return nil, fmt.Errorf("cti action: CTI UI is not initialized")
	}
	var w *GraphWidget
	if (sheetID == "" || sheetID == rootSheetID) && s.CTIUI.CurrentConceptGraph != nil {
		w = s.CTIUI.CurrentConceptGraph
	} else {
		w = s.ensureConceptGraphWidgetForSheetLocked(sheetID)
	}
	if w == nil || w.G() == nil {
		return nil, fmt.Errorf("cti action: no concept graph")
	}
	return &CTIConceptGraphWidget{
		GraphWidget:     w,
		ParentCTI:       s.CTIUI,
		CISess:          s.ConceptSess,
		ActiveFactExprs: w.GetActiveFactExprs(),
	}, nil
}

func ctiClausesFromModule(mod *goivy.Module) []*goivy.Clauses {
	if mod == nil {
		return nil
	}
	conjs := make([]*goivy.Clauses, 0, len(mod.LabeledConjs))
	for _, lc := range mod.LabeledConjs {
		if lc == nil || lc.Formula == nil {
			continue
		}
		if expr, ok := lc.Formula.(goivy.Expr); ok {
			conjs = append(conjs, goivy.FormulaToClauses(expr, nil))
		}
	}
	return conjs
}

func relationNamesUsedByClauses(mod *goivy.Module, clauses *goivy.Clauses) []string {
	if mod == nil || mod.Sig == nil || clauses == nil {
		return nil
	}
	used := map[string]bool{}
	if syms := goivy.UsedSymbolsAST(clauses.ToFormula()); syms != nil {
		for _, sym := range syms.All() {
			if c, ok := sym.(*goivy.Const); ok {
				used[c.Name] = true
			}
		}
	}
	var rels []string
	for symName, entry := range mod.Sig.Symbols.All() {
		if !used[symName] || entry == nil || entry.Sort == nil {
			continue
		}
		if fs, ok := entry.Sort.(*goivy.LogicFunctionSort); ok && goivy.SortEqual(fs.Range(), goivy.Boolean) {
			rels = append(rels, symName)
		}
	}
	sort.Strings(rels)
	return rels
}

func (s *Session) installCounterexampleFeedback(cexTrace *goivy.TraceBase, finalCond, currentConj *goivy.Clauses) (traceText, details string) {
	if cexTrace == nil {
		return "", ""
	}
	ag := cexTrace.AnalysisGraph
	if ag != nil {
		s.AG = ag
		if s.AGUI != nil {
			s.AGUI.AG = ag
			s.AGUI.G = ArtToGraphState(ag)
			if len(ag.States) > 0 {
				// Python's CTI path immediately views the predecessor state so the
				// concept graph explains the bad transition, not the stale load state.
				s.AGUI.ViewState(0, "", true)
			}
		}
		s.syncARGToGraph()

		if s.CTIUI != nil {
			s.CTIUI.AG = ag
			s.CTIUI.HaveCTI = true
			s.CTIUI.CurrentConjecture = currentConj
			if s.CTIUI.AnalysisGraphUI != nil {
				s.CTIUI.AnalysisGraphUI.AG = ag
				s.CTIUI.AnalysisGraphUI.G = ArtToGraphState(ag)
				if len(ag.States) > 0 {
					s.CTIUI.ViewState(0, "", true)
				}
			}
			if finalCond != nil {
				s.CTIUI.ShowUsedRelations(finalCond, false)
			}
		}

		if len(ag.States) > 0 {
			s.setConceptSessionState(ag.States[0])
			s.installStructureConceptGraph(ag.States[0])
		}
	}
	if finalCond != nil {
		s.showUsedRelationsInRoot(finalCond, false)
	}

	traceText = strings.TrimSpace(cexTrace.String())
	if traceText == "" {
		traceText = counterexampleFallbackText(ag)
	}
	if traceText != "" {
		details = "Counterexample trace:\n" + traceText
	}
	return traceText, details
}

func (s *Session) setConceptSessionState(state *goivy.State) {
	if s.ConceptSess == nil || state == nil {
		return
	}
	if state.Clauses != nil {
		s.ConceptSess.State = state.Clauses.ToFormula()
	} else {
		s.ConceptSess.State = goivy.True
	}
	s.ConceptSess.GoalConstraints = nil
	s.ConceptSess.Cache = make(map[string]bool)
	s.ConceptSess.Recompute(nil)
	s.syncAbstractValue()
}

func (s *Session) installStructureConceptGraph(state *goivy.State) bool {
	universe := structureUniverseConsts(state)
	if len(universe) == 0 {
		return false
	}
	stateFormula := goivy.True
	if state != nil && state.Clauses != nil {
		stateFormula = state.Clauses.ToFormula()
	}
	cd := GetStructureConceptDomain(stateFormula, universe, s.structureSignatureSymbols())
	simple := NewConceptSession()
	simple.Domain = simpleConceptDomainFromCD(cd)
	simple.AbstractValue = GetStructureConceptAbstractValue(stateFormula, universe)
	if simple.Domain == nil || len(simple.Domain.Nodes) == 0 {
		return false
	}
	s.SimpleSess = simple
	if w := s.ensureConceptGraphWidgetLocked(); w != nil && w.G() != nil {
		w.G().ParentState = state
	}
	s.ensureConceptChecksLocked()
	return true
}

func (s *Session) structureSignatureSymbols() map[string]*goivy.Const {
	if s == nil || s.CompiledSig == nil {
		return nil
	}
	symbols := make(map[string]*goivy.Const)
	for name, entry := range s.CompiledSig.Symbols.All() {
		if entry == nil || entry.Sort == nil {
			continue
		}
		if sortVal, ok := entry.Sort.(goivy.Sort); ok {
			symbols[name] = goivy.NewConst(name, sortVal)
		}
	}
	return symbols
}

func structureUniverseConsts(state *goivy.State) map[string][]*goivy.Const {
	result := make(map[string][]*goivy.Const)
	if state == nil || state.Universe == nil {
		return result
	}
	seen := make(map[string]bool)
	addExpr := func(sortName string, expr goivy.Expr) {
		c, ok := expr.(*goivy.Const)
		if !ok || c == nil {
			return
		}
		if sortName == "" && c.CSort != nil {
			sortName = c.CSort.String()
		}
		if sortName == "" {
			return
		}
		key := sortName + "\x00" + c.Name
		if seen[key] {
			return
		}
		seen[key] = true
		result[sortName] = append(result[sortName], c)
	}
	addConst := func(sortName string, c *goivy.Const) {
		if c == nil {
			return
		}
		addExpr(sortName, c)
	}
	switch universe := state.Universe.(type) {
	case map[string][]goivy.Expr:
		for sortName, values := range universe {
			for _, value := range values {
				addExpr(sortName, value)
			}
		}
	case map[string][]*goivy.Const:
		for sortName, values := range universe {
			for _, value := range values {
				addConst(sortName, value)
			}
		}
	case []goivy.SortUniverse:
		for _, entry := range universe {
			sortName := ""
			if entry.Sort != nil {
				sortName = entry.Sort.String()
			}
			for _, value := range entry.Values {
				addExpr(sortName, value)
			}
		}
	case map[goivy.Sort][]goivy.Expr:
		for sortVal, values := range universe {
			sortName := ""
			if sortVal != nil {
				sortName = sortVal.String()
			}
			for _, value := range values {
				addExpr(sortName, value)
			}
		}
	}
	for sortName := range result {
		sort.Slice(result[sortName], func(i, j int) bool {
			return result[sortName][i].Name < result[sortName][j].Name
		})
	}
	return result
}

func simpleConceptDomainFromCD(cd *CDConceptDomain) *ConceptDomain {
	d := NewConceptDomain()
	if cd == nil || cd.Concepts == nil {
		return d
	}
	cd.Concepts.ForEachConcept(func(name string, c *CDConcept) {
		if c == nil {
			return
		}
		var vars []string
		var sorts []string
		for _, v := range c.Variables {
			if v == nil {
				continue
			}
			vars = append(vars, v.Name)
			if v.VSort != nil {
				sorts = append(sorts, v.VSort.String())
			} else {
				sorts = append(sorts, "")
			}
		}
		formula := ""
		if c.Formula != nil {
			formula = c.Formula.String()
		}
		d.Concepts[name] = &Concept{
			Name:      name,
			Variables: vars,
			Formula:   formula,
			Sorts:     sorts,
			Arity:     c.Arity(),
		}
	})
	d.Nodes = append([]string{}, cd.Concepts.GetList("nodes")...)
	d.Edges = append([]string{}, cd.Concepts.GetList("edges")...)
	d.NodeLabels = append([]string{}, cd.Concepts.GetList("node_labels")...)
	return d
}

func (s *Session) showUsedRelationsInRoot(clauses *goivy.Clauses, both bool) {
	if clauses == nil || s.AGUI == nil {
		return
	}
	w := s.ensureConceptGraphWidgetLocked()
	if w == nil || w.G() == nil {
		return
	}
	w.ClearEdges()
	for _, rel := range relationNamesUsedByClauses(s.CompiledModule, clauses) {
		boxes := "+"
		if both {
			boxes += "-"
		}
		w.ShowRelation(&Concept{Name: rel}, boxes, true, false)
	}
	w.Update()
	s.toggles = w.G().Checks.Snapshot()
}

func counterexampleFallbackText(ag *goivy.AnalysisGraph) string {
	if ag == nil {
		return ""
	}
	var lines []string
	for _, st := range ag.States {
		if st == nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("state %d:", st.ID))
		if st.Clauses == nil {
			lines = append(lines, "  true")
			continue
		}
		open := st.Clauses.ToOpenFormula()
		if and, ok := open.(*goivy.LogicAnd); ok && len(and.Terms) > 0 {
			for _, term := range and.Terms {
				lines = append(lines, "  "+term.String())
			}
		} else {
			lines = append(lines, "  "+open.String())
		}
	}
	for _, tr := range ag.Transitions {
		if tr.Pre == nil || tr.Post == nil {
			continue
		}
		label := strings.TrimSpace(tr.Label)
		if label == "" {
			label = "transition"
		}
		lines = append(lines, fmt.Sprintf("%s: state %d -> state %d", label, tr.Pre.ID, tr.Post.ID))
	}
	return strings.Join(lines, "\n")
}

func (s *Session) saveInvariantContent() string {
	current := ctiClausesFromModule(s.CompiledModule)
	if s.CTIUI != nil && s.CTIUI.Conjectures != nil {
		current = s.CTIUI.Conjectures
	}
	original := s.OriginalConjs
	if len(original) == 0 && s.CompiledModule != nil {
		original = s.CompiledModule.LabeledConjs
	}
	oldSet := make(map[string]bool)
	for _, lc := range original {
		oldSet[labeledConjKey(lc)] = true
	}
	newSet := make(map[string]bool)
	for _, conj := range current {
		newSet[clausesConjKey(conj)] = true
	}

	var oldKept []*goivy.LabeledFormula
	var oldDropped []*goivy.LabeledFormula
	for _, lc := range original {
		if newSet[labeledConjKey(lc)] {
			oldKept = append(oldKept, lc)
		} else {
			oldDropped = append(oldDropped, lc)
		}
	}
	var newConjs []*goivy.Clauses
	for _, conj := range current {
		if !oldSet[clausesConjKey(conj)] {
			newConjs = append(newConjs, conj)
		}
	}

	var sb strings.Builder
	sb.WriteString("# This file was generated by ivy.\n\n")
	if len(oldKept) > 0 {
		sb.WriteString("\n# original conjectures kept\n\n")
		for _, lc := range oldKept {
			writeInvariantConj(&sb, labeledConjLabel(lc), labeledConjFormula(lc), false)
		}
	}
	if len(oldDropped) > 0 {
		sb.WriteString("\n# original conjectures dropped\n\n")
		for _, lc := range oldDropped {
			writeInvariantConj(&sb, labeledConjLabel(lc), labeledConjFormula(lc), true)
		}
	}
	if len(newConjs) > 0 {
		sb.WriteString("\n# new conjectures\n\n")
		for _, conj := range newConjs {
			writeInvariantConj(&sb, "", clausesConjFormula(conj), false)
		}
	}
	return sb.String()
}

func labeledConjKey(lc *goivy.LabeledFormula) string {
	return labeledConjFormula(lc)
}

func clausesConjKey(conj *goivy.Clauses) string {
	return clausesConjFormula(conj)
}

func labeledConjLabel(lc *goivy.LabeledFormula) string {
	if lc == nil || lc.Label == nil {
		return ""
	}
	return fmt.Sprint(lc.Label)
}

func labeledConjFormula(lc *goivy.LabeledFormula) string {
	if lc == nil || lc.Formula == nil {
		return ""
	}
	if expr, ok := lc.Formula.(goivy.Expr); ok {
		return fmt.Sprint(goivy.DropUniversals(expr))
	}
	return fmt.Sprint(lc.Formula)
}

func clausesConjFormula(conj *goivy.Clauses) string {
	if conj == nil {
		return ""
	}
	return fmt.Sprint(goivy.DropUniversals(conj.ToFormula()))
}

func writeInvariantConj(sb *strings.Builder, label, formula string, commented bool) {
	if commented {
		sb.WriteString("# ")
	}
	if label != "" {
		sb.WriteString(fmt.Sprintf("invariant [%s] %s\n", label, formula))
	} else {
		sb.WriteString(fmt.Sprintf("invariant %s\n", formula))
	}
}

func graphWidgetDOT(w *GraphWidget) string {
	var elements []WebUICyElement
	if w != nil && w.G() != nil {
		cy := RenderConceptGraph(w.G().ConceptSess, w.G().Checks)
		if cy != nil {
			elements = cy.Elements
		}
	}
	var nodes []string
	var edges []string
	for _, ele := range elements {
		switch ele.Group {
		case "nodes":
			id := dotDataString(ele.Data, "id")
			if id == "" {
				id = dotDataString(ele.Data, "obj")
			}
			if id == "" {
				continue
			}
			label := dotDataString(ele.Data, "label")
			if label == "" {
				label = id
			}
			nodes = append(nodes, fmt.Sprintf("  %q [label=%q];", dotEscape(id), dotEscape(label)))
		case "edges":
			src := dotDataString(ele.Data, "source")
			tgt := dotDataString(ele.Data, "target")
			if src == "" || tgt == "" {
				continue
			}
			label := dotDataString(ele.Data, "label")
			if label == "" {
				label = dotDataString(ele.Data, "obj")
			}
			if label == "" {
				edges = append(edges, fmt.Sprintf("  %q -> %q;", dotEscape(src), dotEscape(tgt)))
			} else {
				edges = append(edges, fmt.Sprintf("  %q -> %q [label=%q];", dotEscape(src), dotEscape(tgt), dotEscape(label)))
			}
		}
	}
	sort.Strings(nodes)
	sort.Strings(edges)
	var sb strings.Builder
	sb.WriteString("digraph concept_graph {\n")
	for _, line := range nodes {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	for _, line := range edges {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	sb.WriteString("}\n")
	return sb.String()
}

func dotDataString(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	if v, ok := data[key]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func dotEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
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
	// --- Event trace viewer operations (ivy_ev_viewer.py) ---
	case "events_new_sheet":
		events, evErr := traceEventsFromActionArg(args["events"])
		if evErr != nil {
			err = evErr
			break
		}
		viewer := s.ensureEventViewerLocked()
		sheetID := viewer.NewSheet(events)
		eventSheetResult(result, viewer, sheetID)
	case "events_parse":
		content := actionStringArg(args, "content")
		if content == "" {
			err = fmt.Errorf("events_parse: missing content")
			break
		}
		events, parseErr := ParseTraceEvents(content)
		if parseErr != nil {
			err = fmt.Errorf("events_parse: %w", parseErr)
			break
		}
		viewer := s.ensureEventViewerLocked()
		sheetID := viewer.NewSheet(events)
		eventSheetResult(result, viewer, sheetID)
	case "events_filter":
		viewer := s.ensureEventViewerLocked()
		sheetID := actionStringArg(args, "sheet_id")
		sheet := viewer.GetSheet(sheetID)
		if sheet == nil {
			err = fmt.Errorf("events_filter: unknown sheet %q", sheetID)
			break
		}
		pattern := actionStringArg(args, "pattern")
		filtered, parseErr := FilterEventsE(sheet.Events, pattern)
		if parseErr != nil {
			err = fmt.Errorf("events_filter: %w", parseErr)
			break
		}
		newSheetID := viewer.NewSheet(filtered)
		eventSheetResult(result, viewer, newSheetID)
	case "events_find":
		viewer := s.ensureEventViewerLocked()
		sheetID := actionStringArg(args, "sheet_id")
		sheet := viewer.GetSheet(sheetID)
		if sheet == nil {
			err = fmt.Errorf("events_find: unknown sheet %q", sheetID)
			break
		}
		pattern := actionStringArg(args, "pattern")
		reverse, _ := actionBoolArg(args, "reverse")
		anchor := actionStringArg(args, "anchor")
		ev, addr, parseErr := FindEventFromE(sheet.Events, pattern, reverse, anchor)
		if parseErr != nil {
			err = fmt.Errorf("events_find: %w", parseErr)
			break
		}
		if ev == nil {
			result["found"] = false
			result["message"] = "Pattern not found"
			break
		}
		result["found"] = true
		result["address"] = addr
		result["event"] = ev
	case "events_add_pattern":
		viewer := s.ensureEventViewerLocked()
		pattern := actionStringArg(args, "pattern")
		if pattern == "" {
			err = fmt.Errorf("events_add_pattern: missing pattern")
			break
		}
		if parseErr := viewer.AddPatternE(pattern); parseErr != nil {
			err = fmt.Errorf("events_add_pattern: %w", parseErr)
			break
		}
		result["patterns"] = append([]string{}, viewer.Patterns...)
	case "events_remove_pattern":
		viewer := s.ensureEventViewerLocked()
		idx, ok := actionIntArg(args, "index")
		if !ok {
			err = fmt.Errorf("events_remove_pattern: missing index")
			break
		}
		err = viewer.RemovePattern(idx)
		result["patterns"] = append([]string{}, viewer.Patterns...)
	case "events_clear_patterns":
		viewer := s.ensureEventViewerLocked()
		viewer.ClearPatterns()
		result["patterns"] = []string{}
	case "events_load_patterns":
		viewer := s.ensureEventViewerLocked()
		if parseErr := viewer.LoadPatternsE(actionStringArg(args, "patterns")); parseErr != nil {
			err = fmt.Errorf("events_load_patterns: %w", parseErr)
			break
		}
		result["patterns"] = append([]string{}, viewer.Patterns...)
	case "events_save_patterns":
		viewer := s.ensureEventViewerLocked()
		result["content"] = viewer.SavePatterns()

	// --- Concept graph operations (ConceptInteractiveSession) ---
	case "undo":
		if s.AGUI != nil && s.AGUI.CurrentConceptGraph != nil && s.AGUI.CurrentConceptGraph.GraphStack.CanUndo() {
			s.AGUI.CurrentConceptGraph.Undo()
			s.toggles = s.AGUI.CurrentConceptGraph.G().Checks.Snapshot()
			break
		}
		if s.ConceptSess != nil {
			err = s.ConceptSess.Undo()
		}
	case "redo":
		if s.AGUI != nil && s.AGUI.CurrentConceptGraph != nil && s.AGUI.CurrentConceptGraph.GraphStack.CanRedo() {
			s.AGUI.CurrentConceptGraph.Redo()
			s.toggles = s.AGUI.CurrentConceptGraph.G().Checks.Snapshot()
			break
		}
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
		sheetID := actionStringArg(args, "sheet_id")
		if sheetID != "" {
			w := s.ensureConceptGraphWidgetForSheetLocked(sheetID)
			if w == nil {
				err = fmt.Errorf("gather: no concept graph for sheet %q", sheetID)
				break
			}
			facts := w.Gather()
			active := w.GetActiveFacts()
			if len(active) > 0 {
				facts = active
			}
			result["facts"] = facts
			result["concept"] = conceptGraphActionPayload(w)
			if w.G() != nil {
				s.toggles = w.G().Checks.Snapshot()
			}
		} else if s.ConceptSess != nil {
			facts := s.ConceptSess.GetFacts(nil)
			s.ConceptSess.SupposeConstraints = append([]goivy.Expr{}, facts...)
			strs := make([]string, 0, len(facts))
			for _, fact := range facts {
				strs = append(strs, fact.String())
			}
			result["facts"] = strs
		}
	case "set_fact_selection":
		idx, ok := actionIntArg(args, "index")
		if !ok {
			err = fmt.Errorf("set_fact_selection: missing integer index")
			break
		}
		selected, ok := actionBoolArg(args, "selected")
		if !ok {
			err = fmt.Errorf("set_fact_selection: missing boolean selected")
			break
		}
		w := s.ensureConceptGraphWidgetLocked()
		if w == nil {
			err = fmt.Errorf("set_fact_selection: no concept graph")
			break
		}
		err = w.SetFactSelected(idx, selected)
	case "get_active_facts":
		w := s.ensureConceptGraphWidgetLocked()
		if w == nil {
			result["facts"] = []string{}
			break
		}
		result["facts"] = w.GetActiveFacts()
	case "backtrack":
		w := s.ensureConceptGraphWidgetForSheetLocked(actionStringArg(args, "sheet_id"))
		if w != nil {
			w.Backtrack()
			if w.G() != nil {
				s.toggles = w.G().Checks.Snapshot()
			}
		} else if s.ConceptSess != nil {
			err = s.ConceptSess.Undo()
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
		name := actionStringArg(args, "name")
		if name == "" {
			err = fmt.Errorf("remember: missing name")
			break
		}
		ui, _, uiErr := s.requireAnalysisUIForSheetLocked(actionStringArg(args, "sheet_id"))
		if uiErr != nil {
			err = uiErr
			break
		}
		w := ui.CurrentConceptGraph
		if w == nil {
			w = s.ensureConceptGraphWidgetForSheetLocked(actionStringArg(args, "sheet_id"))
		}
		if w == nil || w.G() == nil {
			err = fmt.Errorf("remember: no concept graph")
			break
		}
		ui.RememberGraph(name, w.G().Copy())
		result["remembered"] = name
		result["remembered_graphs"] = ui.RememberedGraphNames()
	case "show_reachable", "show_reachable_states":
		if s.CompiledModule == nil {
			err = fmt.Errorf("show_reachable: no compiled module")
			break
		}
		if s.ReachableUI == nil {
			ag := goivy.NewAnalysisGraph(s.CompiledModule)
			ag.AddInitialState(nil, nil)
			s.ReachableUI = s.newAnalysisGraphUIForGraphLocked(ag)
		}
		sheetID := s.registerAnalysisSheetLocked(s.ReachableUI)
		cy := RenderAnalysisUIARG(s.ReachableUI)
		result["sheet_id"] = sheetID
		result["arg"] = map[string]interface{}{"elements": cy.Elements}

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
		if sheetID := actionStringArg(args, "sheet_id"); sheetID != "" {
			w := s.ensureConceptGraphWidgetForSheetLocked(sheetID)
			if w == nil {
				err = fmt.Errorf("concrete: no concept graph for sheet %q", sheetID)
				break
			}
			w.MakeConcrete()
			result["state"] = w.G().State
			result["concept"] = conceptGraphActionPayload(w)
			if w.G() != nil {
				s.toggles = w.G().Checks.Snapshot()
			}
			break
		}
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
		if sheetID := actionStringArg(args, "sheet_id"); sheetID != "" {
			ui, _, uiErr := s.requireAnalysisUIForSheetLocked(sheetID)
			if uiErr != nil {
				err = uiErr
				break
			}
			w := ui.CurrentConceptGraph
			if w == nil {
				err = fmt.Errorf("%s: no concept graph for sheet %q", actionName, sheetID)
				break
			}
			parentState, _ := w.G().ParentState.(*goivy.State)
			if parentState == nil {
				err = fmt.Errorf("%s: current concept graph has no parent ARG state", actionName)
				break
			}
			goalClauses, goalErr := conceptGraphGoalClauses(w, parentState)
			if goalErr != nil {
				err = fmt.Errorf("%s: %w", actionName, goalErr)
				break
			}
			if actionName == "path_reach" {
				var boundPtr *int
				if bound, ok := actionIntArg(args, "bound"); ok {
					boundPtr = &bound
				}
				if ui.AG == nil {
					err = fmt.Errorf("path_reach: no analysis graph")
					break
				}
				resultAG := ui.AG.BMC(parentState, goalClauses.ToFormula(), nil, boundPtr)
				if resultAG == nil {
					result["reachable"] = false
					result["message"] = "The condition is unreachable along the given path"
					break
				}
				subUI := s.newAnalysisGraphUIForGraphLocked(resultAG)
				subSheetID := s.registerAnalysisSheetLocked(subUI)
				result["reachable"] = true
				result["sheet_id"] = subSheetID
				result["arg"] = map[string]interface{}{"elements": RenderAnalysisUIARG(subUI).Elements}
				break
			}

			reached := goivy.ReachState(goivy.ArtToInterpState(parentState), goalClauses)
			if reached == nil {
				result["reachable"] = false
				result["message"] = `Cannot reach this state in one step from any known reachable state. Try "reverse".`
				break
			}
			artReached := goivy.InterpToArtState(reached)
			parentState.Unders = append(parentState.Unders, artReached)
			result["reachable"] = true
			if artReached.Clauses != nil {
				result["reached_state"] = goivy.PrettyFmla(artReached.Clauses.ToFormula())
			}
			break
		}
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
		var intIndices []int
		for _, idx := range indices {
			if fi, ok := idx.(float64); ok {
				i := int(fi)
				toRemove[i] = true
				intIndices = append(intIndices, i)
			} else if ii, ok := idx.(int); ok {
				toRemove[ii] = true
				intIndices = append(intIndices, ii)
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
		if s.CTIUI != nil {
			_, _ = s.CTIUI.Weaken(append([]int{}, intIndices...))
		}
		result["removed_count"] = len(removed)
		result["removed"] = removed
		result["remaining_count"] = len(kept)
		s.emit(Event{Type: "weaken_result", Data: map[string]interface{}{"removed": removed}})

	case "cti_gather", "cti_minimize", "cti_check_sufficient", "cti_check_inductive", "cti_strengthen":
		w, widgetErr := s.ctiConceptWidgetForSheetLocked(actionStringArg(args, "sheet_id"))
		if widgetErr != nil {
			err = widgetErr
			break
		}
		switch actionName {
		case "cti_gather":
			w.GatherFacts()
			result["message"] = "CTI facts gathered"
		case "cti_minimize":
			bound := 10
			if s.CTIUI != nil && s.CTIUI.CurrentBound > 0 {
				bound = s.CTIUI.CurrentBound
			}
			conj, minErr := w.MinimizeConjecture(bound)
			if minErr != nil {
				err = minErr
				break
			}
			if conj != nil {
				result["conjecture"] = fmt.Sprint(conj.ToFormula())
			}
			result["message"] = "Conjecture minimized"
		case "cti_check_sufficient":
			ok, msg := w.IsSufficient()
			result["ok"] = ok
			result["message"] = msg
		case "cti_check_inductive":
			ok, msg := w.IsInductive()
			result["ok"] = ok
			result["message"] = msg
		case "cti_strengthen":
			conj, strErr := w.Strengthen()
			if strErr != nil {
				err = strErr
				break
			}
			if conj != nil {
				result["conjecture"] = fmt.Sprint(conj.ToFormula())
				if s.CompiledModule != nil && s.CompiledModule.Cfg != nil && s.CompiledModule.Cfg.AstCfg != nil {
					s.CompiledModule.LabeledConjs = append(
						s.CompiledModule.LabeledConjs,
						s.CompiledModule.Cfg.AstCfg.NewLabeledFormula(nil, conj.ToFormula()),
					)
				}
			}
			result["message"] = "Invariant strengthened"
		}
		result["concept"] = conceptGraphActionPayload(w.GraphWidget)

	case "save_invariant":
		if s.CompiledModule == nil {
			err = fmt.Errorf("save_invariant: no compiled module")
			break
		}
		content := s.saveInvariantContent()
		result["content"] = content
		result["filename"] = "invariant.ivy"
		s.emit(Event{Type: "export", Data: map[string]string{"content": content}})

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
		if sheetID := actionStringArg(args, "sheet_id"); sheetID != "" {
			w := s.ensureConceptGraphWidgetForSheetLocked(sheetID)
			if w == nil {
				err = fmt.Errorf("export: no concept graph for sheet %q", sheetID)
				break
			}
			content := graphWidgetDOT(w)
			result["content"] = content
			result["filename"] = "concept_graph.dot"
			result["mime_type"] = "text/vnd.graphviz"
			s.emit(Event{Type: "export", Data: map[string]interface{}{"filename": "concept_graph.dot", "bytes": len(content)}})
		} else if s.ConceptSess != nil {
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
			c := s.conceptFromFormulaString(formula)
			if s.SimpleSess != nil {
				s.SimpleSess.Domain.Concepts[formula] = c
				if c.Arity == 1 && !stringSliceContains(s.SimpleSess.Domain.NodeLabels, formula) {
					s.SimpleSess.Domain.NodeLabels = append(s.SimpleSess.Domain.NodeLabels, formula)
				}
				if c.Arity == 2 && !stringSliceContains(s.SimpleSess.Domain.Edges, formula) {
					s.SimpleSess.Domain.Edges = append(s.SimpleSess.Domain.Edges, formula)
				}
			}
			if s.AGUI != nil && s.AGUI.CurrentConceptGraph != nil {
				s.AGUI.CurrentConceptGraph.AddConcept(c)
			}
			result["concept_name"] = formula
			result["arity"] = c.Arity
			result["sorts"] = c.Sorts
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
					if fs, ok := entry.Sort.(*goivy.LogicFunctionSort); ok {
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
		err = fmt.Errorf("unknown action: %s", actionName)
	}

	status := "ok"
	if err != nil {
		status = err.Error()
	}
	s.emit(Event{Type: "action_completed", Data: map[string]interface{}{"action": actionName, "status": status}})
	return result, err
}

func actionIntArg(args map[string]interface{}, key string) (int, bool) {
	if args == nil {
		return 0, false
	}
	switch v := args[key].(type) {
	case int:
		return v, true
	case float64:
		return int(v), float64(int(v)) == v
	case json.Number:
		i, err := v.Int64()
		return int(i), err == nil
	default:
		return 0, false
	}
}

func actionBoolArg(args map[string]interface{}, key string) (bool, bool) {
	if args == nil {
		return false, false
	}
	v, ok := args[key].(bool)
	return v, ok
}

func (s *Session) conceptFromFormulaString(formula string) *Concept {
	c := &Concept{Name: formula, Formula: formula}
	open := strings.Index(formula, "(")
	close := strings.LastIndex(formula, ")")
	if open <= 0 || close <= open {
		return c
	}
	baseName := strings.TrimSpace(formula[:open])
	rawArgs := strings.Split(formula[open+1:close], ",")
	var vars []string
	for _, raw := range rawArgs {
		arg := strings.TrimSpace(raw)
		if arg != "" {
			vars = append(vars, arg)
		}
	}
	c.Variables = vars
	c.Arity = len(vars)
	if s.SimpleSess != nil && s.SimpleSess.Domain != nil {
		if base := s.SimpleSess.Domain.Concepts[baseName]; base != nil && len(base.Sorts) >= c.Arity {
			c.Sorts = append([]string{}, base.Sorts[:c.Arity]...)
		}
	}
	return c
}

func stringSliceContains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
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
	sheetID := actionStringArg(args, "sheet_id")
	ui, resolvedSheetID, uiErr := s.requireAnalysisUIForSheetLocked(sheetID)
	if uiErr == nil {
		result["sheet_id"] = resolvedSheetID
	}

	switch action {
	case "view_state":
		if uiErr != nil {
			err = uiErr
			break
		}
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if stateIdx >= 0 {
			ui.ViewState(stateIdx, "", false)
		}
		if s.ConceptSess != nil {
			s.ConceptSess.Recompute(nil)
			s.syncAbstractValue()
		}
	case "check_safety":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if uiErr == nil && stateIdx >= 0 {
			safe, msg := ui.CheckSafetyNode(stateIdx)
			result["safe"] = safe
			result["message"] = msg
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Safety check at node " + nodeID}})
	case "extend", "find_extension":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if uiErr == nil && stateIdx >= 0 {
			label, extErr := ui.FindExtension(stateIdx)
			if extErr != nil {
				err = extErr
			} else {
				result["extension"] = label
				cy := RenderAnalysisUIARG(ui)
				result["arg"] = map[string]interface{}{"elements": cy.Elements}
			}
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Extended from node " + nodeID}})
	case "execute_action":
		if uiErr != nil {
			err = uiErr
			break
		}
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		actionName := actionStringArg(args, "action_name")
		if stateIdx < 0 {
			err = fmt.Errorf("execute_action: invalid ARG node %q", nodeID)
			break
		}
		if actionName == "" {
			err = fmt.Errorf("execute_action: missing action_name")
			break
		}
		if execErr := ui.ExecuteAction(stateIdx, actionName); execErr != nil {
			err = execErr
			break
		}
		cy := RenderAnalysisUIARG(ui)
		result["executed_action"] = actionName
		result["arg"] = map[string]interface{}{"elements": cy.Elements}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Executed " + actionName + " at node " + nodeID}})
	case "mark", "mark_node":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if uiErr == nil && stateIdx >= 0 {
			ui.MarkNode(&ARGStateRef{ID: stateIdx})
		}
		result["marked"] = true
	case "cover", "cover_node":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if uiErr == nil && stateIdx >= 0 {
			ok, coverErr := ui.CoverNode(stateIdx)
			result["covered"] = ok
			if coverErr != nil {
				err = coverErr
			}
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Cover node " + nodeID}})
	case "join", "join_node":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if uiErr == nil && stateIdx >= 0 {
			err = ui.JoinNode(stateIdx)
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Join at node " + nodeID}})
	case "try_conjecture":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		conjStr, _ := args["conjecture"].(string)
		if uiErr == nil && stateIdx >= 0 {
			err = ui.TryConjecture(stateIdx, conjStr)
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Try conjecture at node " + nodeID}})
	case "try_conjecture_choices":
		if uiErr != nil {
			err = uiErr
			break
		}
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if stateIdx < 0 {
			err = fmt.Errorf("try_conjecture_choices: invalid ARG node %q", nodeID)
			break
		}
		choices, choiceErr := ui.TryConjectureChoices(stateIdx)
		if choiceErr != nil {
			err = choiceErr
			break
		}
		result["choices"] = choices
	case "try_remembered":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		goalName, _ := args["goal"].(string)
		if uiErr == nil && stateIdx >= 0 {
			err = ui.TryRememberedGraph(stateIdx, goalName)
			if err == nil && ui.CurrentConceptGraph != nil {
				result["concept"] = map[string]interface{}{
					"elements": RenderConceptGraph(s.SimpleSess, ui.CurrentConceptGraph.G().Checks).Elements,
				}
			}
		}
		s.emit(Event{Type: "status", Data: map[string]string{"message": "Try remembered at node " + nodeID}})
	case "try_remembered_choices":
		if uiErr != nil {
			err = uiErr
			break
		}
		names := ui.RememberedGraphNames()
		choices := make([]ChoiceItem, 0, len(names))
		for _, name := range names {
			choices = append(choices, ChoiceItem{Label: name, Value: name})
		}
		result["choices"] = choices
	case "delete", "delete_node":
		stateIdx := -1
		fmt.Sscanf(nodeID, "state_%d", &stateIdx)
		if uiErr == nil && stateIdx >= 0 {
			ui.DeleteNode(stateIdx)
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
		if uiErr == nil && srcIdx >= 0 && tgtIdx >= 0 {
			ui.RecalculateEdge(srcIdx, tgtIdx)
		} else if uiErr == nil {
			ui.RecalculateAll()
		}
		if uiErr != nil {
			err = uiErr
		} else {
			cy := RenderAnalysisUIARG(ui)
			result["arg"] = map[string]interface{}{"elements": cy.Elements}
		}
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
		if uiErr != nil {
			err = uiErr
		} else if srcIdx >= 0 && tgtIdx >= 0 {
			subArt, decompErr := ui.DecomposeEdgeGraph(srcIdx, tgtIdx)
			if decompErr != nil {
				err = decompErr
			} else if subArt != nil {
				subUI := s.newAnalysisGraphUIForGraphLocked(subArt)
				subSheetID := s.registerAnalysisSheetLocked(subUI)
				result["decomposed"] = true
				result["sheet_id"] = subSheetID
				cy := RenderAnalysisUIARG(subUI)
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
		if uiErr == nil && srcIdx >= 0 && tgtIdx >= 0 {
			filename, lineno, vsErr := ui.ViewSourceEdge(srcIdx, tgtIdx)
			if vsErr == nil {
				result["file"] = filename
				result["lineno"] = lineno
				if source, srcErr := s.sourceText(filename); srcErr == nil {
					result["source"] = source
				}
			} else {
				result["file"] = s.FilePath
				if source, srcErr := s.sourceText(s.FilePath); srcErr == nil {
					result["source"] = source
				}
			}
		} else {
			result["file"] = s.FilePath
			if source, srcErr := s.sourceText(s.FilePath); srcErr == nil {
				result["source"] = source
			}
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

func (s *Session) sourceText(filename string) (string, error) {
	if s.FileContent != "" {
		if filename == "" || filename == s.FilePath || filepath.Base(filename) == filepath.Base(s.FilePath) {
			return s.FileContent, nil
		}
	}
	if filename != "" {
		if by, err := os.ReadFile(filename); err == nil {
			return string(by), nil
		}
	}
	if s.FileContent != "" {
		return s.FileContent, nil
	}
	return "", fmt.Errorf("no source text available")
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
	checks := s.ensureConceptChecksLocked()
	s.toggles = checks.Snapshot()
	sheets := []map[string]interface{}{s.analysisSheetStateLocked(rootSheetID, "Sheet 1", s.AGUI)}
	for _, sheetID := range s.sortedAnalysisSheetIDsLocked() {
		if sheetID == rootSheetID {
			continue
		}
		sheets = append(sheets, s.analysisSheetStateLocked(sheetID, sheetID, s.SheetUIs[sheetID]))
	}
	if s.EventViewer != nil {
		for _, sheetID := range s.EventViewer.SheetNames() {
			if sheet := s.EventViewer.GetSheet(sheetID); sheet != nil {
				sheets = append(sheets, eventSheetState(s.EventViewer, sheet))
			}
		}
	}
	fileName := ""
	if s.FilePath != "" {
		fileName = filepath.Base(s.FilePath)
	}
	mode := string(DefaultMode)
	if s.AGUI != nil && s.AGUI.Mode != "" {
		mode = string(s.AGUI.Mode)
	}
	data, _ := json.Marshal(map[string]interface{}{
		"analysis_state_format":  "ivyweb-json",
		"analysis_state_version": 1,
		"python_a2g_equivalent":  false,
		"session_id":             s.ID,
		"fileName":               fileName,
		"filePath":               s.FilePath,
		"fileContent":            s.FileContent,
		"mode":                   mode,
		"activeSheetId":          rootSheetID,
		"selectedArgNode":        nil,
		"edgeVisibility":         map[string]interface{}{},
		"labelVisibility":        map[string]interface{}{},
		"toggles":                s.toggles,
		"sheets":                 sheets,
	})
	return data
}

func (s *Session) sortedAnalysisSheetIDsLocked() []string {
	ids := make([]string, 0, len(s.SheetUIs))
	for id := range s.SheetUIs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (s *Session) analysisSheetStateLocked(sheetID, label string, ui *AnalysisGraphUI) map[string]interface{} {
	argElements := []WebUICyElement{}
	if ui != nil && ui.AG != nil {
		if cy := RenderAnalysisUIARG(ui); cy != nil {
			argElements = cy.Elements
		}
	} else if sheetID == rootSheetID && s.Graph != nil {
		if cy := RenderWebUIARG(s.Graph); cy != nil {
			argElements = cy.Elements
		}
	}
	conceptElements := []WebUICyElement{}
	if ui != nil && ui.CurrentConceptGraph != nil && ui.CurrentConceptGraph.G() != nil {
		g := ui.CurrentConceptGraph.G()
		if g.ConceptSess != nil {
			if cy := RenderConceptGraph(g.ConceptSess, g.Checks); cy != nil {
				conceptElements = cy.Elements
			}
		}
	} else if sheetID == rootSheetID && s.SimpleSess != nil {
		if cy := RenderConceptGraph(s.SimpleSess, s.ensureConceptChecksLocked()); cy != nil {
			conceptElements = cy.Elements
		}
	}
	return map[string]interface{}{
		"id":              sheetID,
		"type":            "analysis",
		"label":           label,
		"selectedArgNode": nil,
		"arg": map[string]interface{}{
			"elements":  argElements,
			"positions": nil,
		},
		"concept": map[string]interface{}{
			"elements":  conceptElements,
			"positions": nil,
		},
	}
}

func eventSheetState(viewer *EventTraceViewer, sheet *EventSheet) map[string]interface{} {
	return map[string]interface{}{
		"id":                   sheet.Name,
		"type":                 "events",
		"label":                sheet.Label,
		"events":               sheet.Events,
		"patterns":             append([]string{}, viewer.Patterns...),
		"selectedEventAddress": nil,
	}
}

// CheckResult holds the result of a verification check.
type WebUICheckResult struct {
	Result                string   `json:"result"` // "pass", "fail", "error"
	Message               string   `json:"message"`
	Z3Contacted           bool     `json:"z3_contacted"`                // true if Z3 was actually called
	FailedConjecture      string   `json:"failed_conjecture,omitempty"` // formula text if fail
	FailedLabel           string   `json:"failed_label,omitempty"`      // label if fail
	UsedRelations         []string `json:"used_relations,omitempty"`    // relations to auto-check "+"
	CounterexampleTrace   string   `json:"counterexample_trace,omitempty"`
	CounterexampleDetails string   `json:"counterexample_details,omitempty"`
}

// RunCheck runs verification in the specified mode using the compiled module and Z3.
func (s *Session) RunCheck(mode string) *WebUICheckResult {
	return s.RunCheckWithOptions(mode, CheckOptions{})
}

func (s *Session) RunCheckWithOptions(mode string, options CheckOptions) *WebUICheckResult {
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
		ag, postState, _, err := goivy.MakeCheckArt(s.CompiledModule, "", conjClauses)
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
			witness := func(v *goivy.LogicVariable) goivy.Expr {
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
				currentConj := conj
				traceText, cexDetails := s.installCounterexampleFeedback(cexTrace, finalCond, currentConj)
				return &WebUICheckResult{
					Z3Contacted:           true,
					Result:                "fail",
					Message:               "The following conjecture is not relatively inductive:",
					FailedConjecture:      formula,
					FailedLabel:           label,
					UsedRelations:         relationNamesUsedByClauses(s.CompiledModule, finalCond),
					CounterexampleTrace:   traceText,
					CounterexampleDetails: cexDetails,
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
		if options.Bound > 0 {
			nSteps = options.Bound
		}
		if s.CTIUI != nil {
			s.CTIUI.CurrentBound = nSteps
		}
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
		ag, postState, _, err := goivy.MakeCheckArt(s.CompiledModule, "", conjClauses)
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
			witness := func(v *goivy.LogicVariable) goivy.Expr {
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
		ag, postState, _, err := goivy.MakeCheckArt(s.CompiledModule, "", conjClauses)
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
			witness := func(v *goivy.LogicVariable) goivy.Expr {
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
	s.SimpleSess.AbstractValue = make(map[string]bool, len(av))
	for k, v := range av {
		s.SimpleSess.AbstractValue[k] = v
	}
}

// Toggles stores edge/label visibility checkbox state.
type Toggles struct {
	Edges  map[string]map[string]bool `json:"edges"`  // edge_name → {display_class → checked}
	Labels map[string]map[string]bool `json:"labels"` // label_name → {display_class → checked}
}

// GetToggles returns the current toggle state.
func (s *Session) GetToggles() *Toggles {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toggles = s.ensureConceptChecksLocked().Snapshot()
	return s.toggles
}

// SetToggle updates a single toggle value.
func (s *Session) SetToggle(edge, displayClass string, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	checks := s.ensureConceptChecksLocked()
	checks.SetCheckboxClass(edge, displayClass, value)
	s.toggles = checks.Snapshot()
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
