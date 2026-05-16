package webui

// Full port of ivy_ui_cti.py — CTI (counterexample to induction) variant
// of AnalysisGraphUI with invariant checking, bounded checking,
// diagram abstraction, weakening, and transitive relation autodetection.

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"sort"
	"strings"
	"sync"
)

// CTIAnalysisGraphUI extends AnalysisGraphUI with CTI-specific operations
// (Python: class AnalysisGraphUI in ivy_ui_cti.py).
type CTIAnalysisGraphUI struct {
	*AnalysisGraphUI

	mu2 sync.Mutex

	// Mod is the compiled module providing axioms, sig, actions.
	Mod *goivy.Module

	// Solver is the Z3 solver instance for this session.
	Solver *goivy.Solver

	// AG is the stored analysis graph from the last CheckInductiveness.
	AG *goivy.AnalysisGraph

	// TransitiveRelations tracks detected transitive relation names.
	TransitiveRelations []string

	// TransitiveRelationConcepts tracks the concepts for transitive relations.
	TransitiveRelationConcepts []*CDConcept

	// RelationsToMinimize is the set of relations to minimize during checking.
	RelationsToMinimize string

	// Conjectures is the current list of conjectures (Python: self.conjectures = im.module.conjs).
	Conjectures []*goivy.Clauses

	// HaveCTI indicates whether a counterexample to induction exists.
	HaveCTI bool

	// CurrentConjecture is the conjecture currently being checked.
	CurrentConjecture *goivy.Clauses

	// CurrentBound is the last BMC bound used.
	CurrentBound int
}

// NewCTIAnalysisGraphUI creates a new CTI-variant UI.
func NewCTIAnalysisGraphUI(mod *goivy.Module) *CTIAnalysisGraphUI {
	var solver *goivy.Solver
	if mod != nil {
		solver = goivy.NewSolver(mod, nil)
	}
	return &CTIAnalysisGraphUI{
		AnalysisGraphUI:     NewAnalysisGraphUI(),
		Mod:                 mod,
		Solver:              solver,
		RelationsToMinimize: "relations to minimize",
		CurrentBound:        -1,
	}
}

// CTIMenus returns the CTI-specific menu structure (Python: AnalysisGraphUI.menus in ivy_ui_cti.py).
func (ui *CTIAnalysisGraphUI) CTIMenus() []MenuDef {
	return []MenuDef{
		{
			Type:  "menu",
			Label: "File",
			Items: []MenuItem{
				{Type: "button", Label: "Remove tab", Action: "remove_tab"},
				{Type: "button", Label: "Save invariant", Action: "save_conjectures"},
				{Type: "button", Label: "Exit", Action: "exit"},
			},
		},
		{
			Type:  "menu",
			Label: "Invariant",
			Items: []MenuItem{
				{Type: "button", Label: "Check induction", Action: "check_inductiveness"},
				{Type: "button", Label: "Bounded check", Action: "bmc_conjecture"},
				{Type: "button", Label: "Diagram", Action: "diagram"},
				{Type: "button", Label: "Weaken", Action: "weaken"},
			},
		},
	}
}

// StartCTI initializes the CTI UI, detecting transitive relations
// (Python: AnalysisGraphUI.start in ivy_ui_cti.py).
func (ui *CTIAnalysisGraphUI) StartCTI(conjectures []*goivy.Clauses) {
	ui.AnalysisGraphUI.Start()
	ui.TransitiveRelations = nil
	ui.TransitiveRelationConcepts = nil
	ui.RelationsToMinimize = "relations to minimize"
	ui.Conjectures = conjectures
	ui.ViewState(0, "", false)
	ui.AutodetectTransitive()
}

// AutodetectTransitive detects binary transitive relations in the background theory
// (Python: AnalysisGraphUI.autodetect_transitive).
func (ui *CTIAnalysisGraphUI) AutodetectTransitive() {
	ui.mu2.Lock()
	defer ui.mu2.Unlock()
	ui.TransitiveRelations = nil
	ui.TransitiveRelationConcepts = nil

	if ui.Mod == nil || ui.Solver == nil {
		return
	}

	axioms := ui.Mod.BackgroundTheory(nil)
	if axioms == nil {
		return
	}

	for _, c := range ui.Mod.Sig.AllSymbols() {
		fs, ok := c.CSort.(*goivy.LogicFunctionSort)
		if !ok || fs.Arity() != 2 {
			continue
		}
		if !goivy.SortEqual(fs.Domain()[0], fs.Domain()[1]) {
			continue
		}
		if !goivy.SortEqual(fs.Range(), goivy.Boolean) {
			continue
		}

		domSort := fs.Domain()[0]
		xv, _ := goivy.NewVariable("X", domSort)
		yv, _ := goivy.NewVariable("Y", domSort)
		zv, _ := goivy.NewVariable("Z", domSort)

		cxy := goivy.MustApply(c, xv, yv)
		cyz := goivy.MustApply(c, yv, zv)
		cxz := goivy.MustApply(c, xv, zv)
		cxx := goivy.MustApply(c, xv, xv)
		cyy := goivy.MustApply(c, yv, yv)

		notCxy, _ := goivy.NewNot(cxy)
		notCyz, _ := goivy.NewNot(cyz)
		notCyy, _ := goivy.NewNot(cyy)

		// transitive = ForAll([X,Y,Z], Or(Not(c(X,Y)), Not(c(Y,Z)), c(X,Z)))
		transOr, _ := goivy.NewOr(notCxy, notCyz, cxz)
		transitive, _ := goivy.NewForAll([]*goivy.LogicVariable{xv, yv, zv}, transOr)

		// defined_symmetry = ForAll([X,Y], Or(c(X,X), Not(c(Y,Y))))
		symOr, _ := goivy.NewOr(cxx, notCyy)
		definedSym, _ := goivy.NewForAll([]*goivy.LogicVariable{xv, yv}, symOr)

		t := goivy.NewClauses([]goivy.Expr{transitive, definedSym}, nil, nil)
		implied, err := ui.Solver.ClausesImply(axioms, t)
		if err != nil || !implied {
			continue
		}

		ui.TransitiveRelations = append(ui.TransitiveRelations, c.Name)

		concept := FormulaToConceptl(cxy)
		ui.TransitiveRelationConcepts = append(ui.TransitiveRelationConcepts, concept)

		if ui.CurrentConceptGraph != nil {
			ui.CurrentConceptGraph.ShowRelation(
				&Concept{Name: concept.Name, Arity: concept.Arity()},
				"T", true, false,
			)
		}
	}

	if len(ui.TransitiveRelations) > 0 && ui.CurrentConceptGraph != nil {
		ui.CurrentConceptGraph.Update()
	}
}

// ctiWitness returns a Skolem witness function that creates '@'-prefixed constants.
// (Python: def witness(v): c = lg.Const('@'+v.name, v.sort))
func ctiWitness(usedNames map[string]bool) goivy.Skolemizer {
	return func(v *goivy.LogicVariable) goivy.Expr {
		name := "@" + v.Name
		if usedNames != nil {
			if _, exists := usedNames[name]; exists {
				panic(fmt.Sprintf("witness name collision: %s", name))
			}
		}
		return goivy.NewConst(name, v.VSort)
	}
}

// collectUsedNames returns the set of symbol names from the module signature.
func (ui *CTIAnalysisGraphUI) collectUsedNames() map[string]bool {
	names := make(map[string]bool)
	if ui.Mod != nil && ui.Mod.Sig != nil {
		for _, sym := range ui.Mod.Sig.AllSymbols() {
			names[sym.Name] = true
		}
	}
	return names
}

// CheckInductiveness checks all conjectures for relative inductiveness
// (Python: AnalysisGraphUI.check_inductiveness).
func (ui *CTIAnalysisGraphUI) CheckInductiveness() (bool, string) {
	ui.mu2.Lock()
	defer ui.mu2.Unlock()
	return ui.checkInductivenessUnlocked()
}

func (ui *CTIAnalysisGraphUI) checkInductivenessUnlocked() (bool, string) {
	if ui.Mod == nil {
		ui.HaveCTI = false
		return true, "Inductive invariant found (no module)"
	}

	// Auto-detect relations to minimize from sig.
	// Python: self.relations_to_minimize.value = ' '.join(sorted(k for k,v ...))
	if ui.RelationsToMinimize == "relations to minimize" {
		transSet := make(map[string]bool)
		for _, tr := range ui.TransitiveRelations {
			transSet[tr] = true
		}
		var relNames []string
		for _, sym := range ui.Mod.Sig.AllSymbols() {
			fs, ok := sym.CSort.(*goivy.LogicFunctionSort)
			if !ok {
				continue
			}
			if !goivy.SortEqual(fs.Range(), goivy.Boolean) {
				continue
			}
			if transSet[sym.Name] {
				continue
			}
			relNames = append(relNames, sym.Name)
		}
		sort.Strings(relNames)
		ui.RelationsToMinimize = strings.Join(relNames, " ")
	}

	// Python: ag, succeed, fail = ivy_trace.make_check_art(precond=self.conjectures)
	ag, succeed, fail, err := goivy.MakeCheckArt(ui.Mod, "", ui.Conjectures)
	if err != nil {
		return true, fmt.Sprintf("CheckInductiveness: make_check_art failed: %v", err)
	}

	usedNames := ui.collectUsedNames()
	witness := ctiWitness(usedNames)

	// Python: to_test = [None] + list(self.conjectures)
	type testEntry struct {
		conj *goivy.Clauses // nil = safety check
	}
	toTest := make([]testEntry, 0, len(ui.Conjectures)+1)
	toTest = append(toTest, testEntry{nil})
	for _, c := range ui.Conjectures {
		toTest = append(toTest, testEntry{c})
	}

	// Build relations to minimize list.
	var relsToMin []string
	if ui.RelationsToMinimize != "" {
		relsToMin = strings.Fields(ui.RelationsToMinimize)
	}

	for _, entry := range toTest {
		var clauses *goivy.Clauses
		var post *goivy.State

		if entry.conj == nil {
			clauses = goivy.TrueClauses(nil)
			post = fail
		} else {
			clauses = goivy.DualClauses(entry.conj, witness, nil)
			post = succeed
		}
		clauses.Annot = goivy.EmptyAnnotation{}

		res := goivy.CheckFinalCond(ag, post, clauses, relsToMin, true)
		if res != nil {
			ui.CurrentConjecture = entry.conj
			ui.AG = res.AnalysisGraph
			ui.HaveCTI = true

			if entry.conj != nil {
				ui.showUsedRelationsUnlocked(clauses, false)
			}

			if entry.conj == nil {
				return false, "An assertion failed. A failing state is displayed."
			}
			fmla := goivy.DropUniversals(entry.conj.ToFormula())
			return false, fmt.Sprintf("The following conjecture is not relatively inductive:\n%v", fmla)
		}
	}

	ui.HaveCTI = false
	var conjStrs []string
	for _, conj := range ui.Conjectures {
		conjStrs = append(conjStrs, fmt.Sprintf("%v", conj.ToFormula()))
	}
	return true, fmt.Sprintf("Inductive invariant found:\n%s", strings.Join(conjStrs, "\n"))
}

// BoundedCheck performs bounded model checking for a conjecture
// (Python: AnalysisGraphUI.bmc_conjecture).
func (ui *CTIAnalysisGraphUI) BoundedCheck(bound int, conjecture *goivy.Clauses) (bool, string) {
	found, msg, _ := ui.BoundedCheckTrace(bound, conjecture)
	return found, msg
}

// BoundedCheckTrace is BoundedCheck plus the counterexample trace object.
// Python passes this trace to ui_parent.add(res, ui_class=ivy_ui.AnalysisGraphUI)
// when the dialog's View button is pressed.
func (ui *CTIAnalysisGraphUI) BoundedCheckTrace(bound int, conjecture *goivy.Clauses) (bool, string, *goivy.TraceBase) {
	ui.mu2.Lock()
	defer ui.mu2.Unlock()

	ui.CurrentBound = bound

	if ui.Mod == nil {
		return false, "no module loaded", nil
	}

	// Python: if conj is None: conj = and_clauses(*self.conjectures)
	conj := conjecture
	if conj == nil && len(ui.Conjectures) > 0 {
		conj = ui.Conjectures[0]
		for _, c := range ui.Conjectures[1:] {
			conj = goivy.AndClausesTyped(conj, c)
		}
	}
	if conj == nil {
		return false, "no conjectures to check", nil
	}

	usedNames := ui.collectUsedNames()
	witness := ctiWitness(usedNames)
	clauses := goivy.DualClauses(conj, witness, nil)

	// Python: ag = self.new_ag()
	ag := goivy.NewAnalysisGraph(ui.Mod)
	post := ag.AddInitialState(nil, nil)
	if len(ag.States) > 0 {
		post = ag.States[0]
	}

	// Python: if 'initialize' in im.module.actions: ...
	if initAct, ok := ui.Mod.Actions.Get2("initialize"); ok {
			if act, ok2 := initAct.(goivy.ActionsAction); ok2 {
				var err error
				post, err = ag.Execute(true, act, post, nil, "initialize")
				if err != nil {
					return false, fmt.Sprintf("initialize action failed: %v", err), nil
				}
			}
		}

	stepAction := goivy.BMCEnvAction(ui.Mod)

	for n := 0; n <= bound; n++ {
			res := goivy.CheckFinalCond(ag, post, clauses, nil, true)
			if res != nil {
				fmla := conj.ToFormula()
				return true, fmt.Sprintf("BMC with bound %d found a counter-example to:\n%v", n, fmla), res
			}
			if n < bound && stepAction != nil {
				var err error
				post, err = ag.Execute(true, stepAction, post, nil, "")
				if err != nil {
					return false, fmt.Sprintf("step %d failed: %v", n, err), nil
				}
			}
		}

	fmla := conj.ToFormula()
	return false, fmt.Sprintf("BMC with bound %d did not find a counter-example to:\n%v", bound, fmla), nil
}

// Diagram computes a diagram abstraction of the current CTI
// (Python: AnalysisGraphUI.diagram).
func (ui *CTIAnalysisGraphUI) Diagram() (string, error) {
	ui.mu2.Lock()
	defer ui.mu2.Unlock()

	if ui.Mod == nil {
		return "", fmt.Errorf("no module loaded")
	}

	if !ui.HaveCTI {
		ok, msg := ui.checkInductivenessUnlocked()
		if ok || ui.AG == nil || len(ui.AG.States) < 2 {
			return msg, nil
		}
	}

	if ui.AG == nil || len(ui.AG.States) < 2 {
		return "", fmt.Errorf("no CTI states available")
	}

	// Python: post = dual_clauses(conj) if conj else true_clauses()
	var post *goivy.Clauses
	if ui.CurrentConjecture != nil {
		post = goivy.DualClauses(ui.CurrentConjecture, nil, nil)
	} else {
		post = goivy.TrueClauses(nil)
	}

	pre := ui.AG.States[0].Clauses
	axioms := ui.Mod.BackgroundTheory(nil)

	// Python: uc = universe_constraint(self.g.states[0])
	interpState := &goivy.InterpState{Universe: ui.AG.States[0].Universe}
	uc := goivy.UniverseConstraint(interpState)
	axiomsUc := goivy.AndClausesTyped(axioms, uc)

	// Python: rev = reverse_image(post, axioms, self.g.states[1].update)
	update := goivy.StateUpdate(ui.AG.States[1])
	if update == nil {
		return "", fmt.Errorf("CTI successor state has no update")
	}
	rev := goivy.ReverseImage(post, axioms, update)

	// Python: clauses = and_clauses(and_clauses(pre, rev), axioms_uc)
	combined := goivy.AndClausesTyped(goivy.AndClausesTyped(pre, rev), axiomsUc)

	if ui.Solver == nil {
		return "", fmt.Errorf("no solver available")
	}

	modelResult, err := ui.Solver.GetModelClauses(combined)
	if err != nil {
		return "", fmt.Errorf("get_model_clauses failed: %w", err)
	}
	if modelResult == nil {
		return "", fmt.Errorf("no model found (UNSAT)")
	}

	isSkolemConst := func(c *goivy.Const) bool { return goivy.TraceIsSkolem(c.Name) }
	diag, err := ui.Solver.ClausesModelToDiagram(rev, isSkolemConst, axioms)
	if err != nil {
		return "", fmt.Errorf("clauses_model_to_diagram failed: %w", err)
	}

	// Python: self.view_state(self.g.states[0], clauses=diag, reset=True)
	// Python: self.show_used_relations(diag, both=True)
	// Python: self.current_concept_graph.gather_facts()
	ui.showUsedRelationsUnlocked(diag, true)

	return fmt.Sprintf("%v", diag.ToFormula()), nil
}

// Weaken removes conjectures from the current set
// (Python: AnalysisGraphUI.weaken).
func (ui *CTIAnalysisGraphUI) Weaken(indices []int) ([]*goivy.Clauses, error) {
	ui.mu2.Lock()
	defer ui.mu2.Unlock()

	if len(indices) == 0 {
		return nil, fmt.Errorf("no conjectures selected")
	}

	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	var removed []*goivy.Clauses
	for _, idx := range indices {
		if idx < 0 || idx >= len(ui.Conjectures) {
			continue
		}
		removed = append(removed, ui.Conjectures[idx])
		ui.Conjectures = append(ui.Conjectures[:idx], ui.Conjectures[idx+1:]...)
	}
	ui.HaveCTI = false
	return removed, nil
}

// SaveConjectures formats conjectures for saving to an Ivy file
// (Python: AnalysisGraphUI.save_conjectures).
func (ui *CTIAnalysisGraphUI) SaveConjectures() string {
	ui.mu2.Lock()
	defer ui.mu2.Unlock()

	var sb strings.Builder
	sb.WriteString("# This file was generated by ivy.\n\n")
	if len(ui.Conjectures) > 0 {
		sb.WriteString("# conjectures\n\n")
		for _, conj := range ui.Conjectures {
			fmla := goivy.DropUniversals(conj.ToFormula())
			sb.WriteString(fmt.Sprintf("invariant %v\n", fmla))
		}
	}
	return sb.String()
}

// ShowUsedRelations enables display of relations used in given clauses
// (Python: AnalysisGraphUI.show_used_relations).
func (ui *CTIAnalysisGraphUI) ShowUsedRelations(clauses *goivy.Clauses, both bool) {
	ui.mu2.Lock()
	defer ui.mu2.Unlock()
	ui.showUsedRelationsUnlocked(clauses, both)
}

func (ui *CTIAnalysisGraphUI) showUsedRelationsUnlocked(clauses *goivy.Clauses, both bool) {
	if ui.CurrentConceptGraph == nil || clauses == nil {
		return
	}
	ui.CurrentConceptGraph.ClearEdges()

	// Python: used = set(il.normalize_symbol(s) for s in lu.used_constants(clauses.to_formula()))
	fmla := clauses.ToFormula()
	usedSyms := goivy.UsedSymbolsAST(fmla)
	usedNames := make(map[string]bool)
	if usedSyms != nil {
		for _, sym := range usedSyms.All() {
			if cnst, ok := sym.(*goivy.Const); ok {
				usedNames[cnst.Name] = true
			}
		}
	}

	g := ui.CurrentConceptGraph.G()
	if g == nil {
		return
	}

	// Python: for rel in rels: if any(c in used ...) for c in used_constants(fmla)
	for _, relName := range g.RelationIDs() {
		concept := g.ConceptFromID(relName)
		if concept == nil {
			continue
		}
		if hasUsedNonSkolem(concept.Formula, usedNames) {
			ui.CurrentConceptGraph.ShowRelation(concept, "+", true, false)
			if both {
				ui.CurrentConceptGraph.ShowRelation(concept, "-", true, false)
			}
		}
	}

	// Python: handle arity-3 applications
	needUpdateRelations := false
	for _, app := range goivy.AppsClauses(clauses) {
		applyExpr, ok := app.(*goivy.Apply)
		if !ok || len(applyExpr.Terms) != 3 {
			continue
		}
		if !goivy.IsNumeral(applyExpr.Terms[0]) {
			continue
		}
		xv, _ := goivy.NewVariable("X", applyExpr.Terms[1].NodeSort())
		yv, _ := goivy.NewVariable("Y", applyExpr.Terms[2].NodeSort())
		newFmla := goivy.MustApply(applyExpr.Func, applyExpr.Terms[0], xv, yv)
		newConcept := FormulaToConceptl(newFmla)
		simpleConcept := &Concept{Name: newConcept.Name, Arity: newConcept.Arity()}
		g.NewRelation(simpleConcept)
		needUpdateRelations = true
		ui.CurrentConceptGraph.ShowRelation(simpleConcept, "+", true, false)
		if both {
			ui.CurrentConceptGraph.ShowRelation(simpleConcept, "-", true, false)
		}
	}
	_ = needUpdateRelations

	ui.CurrentConceptGraph.Update()
}

// hasUsedNonSkolem checks if a formula string contains any used non-Skolem constants.
func hasUsedNonSkolem(formulaStr string, usedNames map[string]bool) bool {
	for name := range usedNames {
		if strings.HasPrefix(name, "@") {
			continue
		}
		if strings.Contains(formulaStr, name) {
			return true
		}
	}
	return false
}

// ConceptGraphUIMenus returns the CTI concept-graph-specific menus
// (Python: class ConceptGraphUI.menus in ivy_ui_cti.py).
func ConceptGraphUIMenus() []MenuDef {
	return []MenuDef{
		{
			Type:  "menu",
			Label: "Conjecture",
			Items: []MenuItem{
				{Type: "button", Label: "Undo", Action: "undo"},
				{Type: "button", Label: "Redo", Action: "redo"},
				{Type: "button", Label: "Gather", Action: "gather_facts"},
				{Type: "button", Label: "Bounded check", Action: "bmc_conjecture"},
				{Type: "button", Label: "Minimize", Action: "minimize_conjecture"},
				{Type: "button", Label: "Check sufficient", Action: "is_sufficient"},
				{Type: "button", Label: "Check relative induction", Action: "is_inductive"},
				{Type: "button", Label: "Strengthen", Action: "strengthen"},
				{Type: "button", Label: "Export", Action: "export"},
			},
		},
		{
			Type:  "menu",
			Label: "View",
			Items: []MenuItem{
				{Type: "button", Label: "Add relation", Action: "add_relation"},
			},
		},
	}
}

// CTIConceptGraphWidget extends GraphWidget with CTI-specific operations
// (Python: class ConceptGraphUI in ivy_ui_cti.py).
type CTIConceptGraphWidget struct {
	*GraphWidget

	// ParentCTI is the CTIAnalysisGraphUI that owns this widget.
	ParentCTI *CTIAnalysisGraphUI

	// CISess is the full interactive concept session (Python: self.g.concept_session).
	CISess *ConceptInteractiveSession

	// ActiveFactExprs stores gathered fact expressions for GetSelectedConjecture.
	ActiveFactExprs []goivy.Expr
}

// NewCTIConceptGraphWidget creates a new CTI concept graph widget.
func NewCTIConceptGraphWidget(gs *GraphStack, parent *CTIAnalysisGraphUI) *CTIConceptGraphWidget {
	return &CTIConceptGraphWidget{
		GraphWidget: NewGraphWidget(gs),
		ParentCTI:   parent,
	}
}

// GatherFacts gathers facts from selected nodes and visible edges
// (Python: ConceptGraphUI.gather_facts).
func (w *CTIConceptGraphWidget) GatherFacts() {
	g := w.G()
	if g == nil || w.CISess == nil {
		return
	}

	selectedNodes := make([]string, 0)
	w.mu.Lock()
	for n := range w.NodeSelection {
		selectedNodes = append(selectedNodes, n)
	}
	w.mu.Unlock()
	sort.Strings(selectedNodes)

	if len(selectedNodes) == 0 {
		selectedNodes = g.NodeIDs()
	}

	selectedSet := make(map[string]bool, len(selectedNodes))
	for _, n := range selectedNodes {
		selectedSet[n] = true
	}

	type factEntry struct {
		formula goivy.Expr
	}

	var facts []factEntry

	// Python: for node in selected_nodes: facts += [(f, elems) for f in g.concept_session.get_node_facts(node)]
	for _, node := range selectedNodes {
		for _, f := range w.CISess.GetNodeFacts(node) {
			facts = append(facts, factEntry{f})
		}
	}

	// Python: edges = set(tag[-3:] for tag, value in g.concept_session.abstract_value
	//                     if tag[0] == 'edge_info' and tag[-2] in selected_nodes and tag[-1] in selected_nodes)
	type edgeTriple struct{ edge, source, target string }
	edgeSet := make(map[edgeTriple]bool)
	for _, tv := range w.CISess.AbstractValue {
		tag := tv.Tag
		if len(tag) < 4 || tag[0] != "edge_info" {
			continue
		}
		src := tag[len(tag)-2]
		tgt := tag[len(tag)-1]
		edge := tag[len(tag)-3]
		if selectedSet[src] && selectedSet[tgt] {
			edgeSet[edgeTriple{edge, src, tgt}] = true
		}
	}

	// Sort edges for determinism.
	var edges []edgeTriple
	for e := range edgeSet {
		edges = append(edges, e)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].edge != edges[j].edge {
			return edges[i].edge < edges[j].edge
		}
		if edges[i].source != edges[j].source {
			return edges[i].source < edges[j].source
		}
		return edges[i].target < edges[j].target
	})

	for _, e := range edges {
		checks := g.Checks

		// Python: if g.edge_display_checkboxes[edge]['all_to_all'].value:
		if checks.EdgeVisible(e.edge, EdgeDisplayAllToAll) {
			// Python: if transitive and source==target: continue
			if checks.EdgeVisible(e.edge, EdgeDisplayTransitive) && e.source == e.target {
				continue
			}
			polTrue := true
			for _, f := range w.CISess.GetEdgeFacts(e.edge, e.source, e.target, &polTrue) {
				facts = append(facts, factEntry{f})
			}
		}

		// Python: if g.edge_display_checkboxes[edge]['none_to_none'].value or edge == '=':
		if checks.EdgeVisible(e.edge, EdgeDisplayNoneToNone) || e.edge == "=" {
			polFalse := false
			for _, f := range w.CISess.GetEdgeFacts(e.edge, e.source, e.target, &polFalse) {
				facts = append(facts, factEntry{f})
			}
		}
	}

	// Python: filter double equalities: Not(Eq(t1,t2)) where t1 >= t2
	var filtered []factEntry
	for _, fe := range facts {
		if shouldFilterFact(fe.formula) {
			continue
		}
		filtered = append(filtered, fe)
	}

	// Extract formulas and set on graph.
	w.ActiveFactExprs = make([]goivy.Expr, len(filtered))
	factStrs := make([]string, len(filtered))
	for i, fe := range filtered {
		w.ActiveFactExprs[i] = fe.formula
		factStrs[i] = fmt.Sprintf("%v", fe.formula)
	}
	g.SetFacts(factStrs)
	w.Update()
	w.HighlightSelectedFacts()
}

// shouldFilterFact returns true for facts that should be filtered out.
// Python: (type(f) is Not and type(f.body) is Eq and f.body.t1 >= f.body.t2)
func shouldFilterFact(f goivy.Expr) bool {
	notExpr, ok := f.(*goivy.LogicNot)
	if !ok {
		return false
	}
	eq, ok := notExpr.Body.(*goivy.Eq)
	if !ok {
		return false
	}
	return fmt.Sprintf("%v", eq.T1) >= fmt.Sprintf("%v", eq.T2)
}

// Strengthen adds a new conjecture from selected facts
// (Python: ConceptGraphUI.strengthen).
func (w *CTIConceptGraphWidget) Strengthen() (*goivy.Clauses, error) {
	conj := w.GetSelectedConjecture()
	if conj == nil {
		return nil, fmt.Errorf("no facts selected")
	}
	if w.ParentCTI != nil {
		w.ParentCTI.HaveCTI = false
		w.ParentCTI.Conjectures = append(w.ParentCTI.Conjectures, conj)
	}
	return conj, nil
}

// GetSelectedConjecture returns a positive universal conjecture from selected facts
// (Python: ConceptGraphUI.get_selected_conjecture).
func (w *CTIConceptGraphWidget) GetSelectedConjecture() *goivy.Clauses {
	facts := w.ActiveFactExprs

	// Python: assert len(free_variables(*facts)) == 0
	for _, f := range facts {
		fv := goivy.FreeVariables(f)
		if fv != nil && fv.Len() > 0 {
			return nil
		}
	}

	// Python: collect constants, substitute numerals of uninterpreted sorts
	rn := goivy.NewVariableGenerator()
	subs := make(map[goivy.NodeKey]goivy.Expr)

	var allConsts []*goivy.Const
	for _, f := range facts {
		syms := goivy.UsedSymbolsAST(f)
		if syms == nil {
			continue
		}
		for _, sym := range syms.All() {
			if c, ok := sym.(*goivy.Const); ok {
				allConsts = append(allConsts, c)
			}
		}
	}

	// Sort for determinism.
	sort.Slice(allConsts, func(i, j int) bool {
		return allConsts[i].Name < allConsts[j].Name
	})

	seen := make(map[string]bool)
	for _, c := range allConsts {
		if seen[c.Name] {
			continue
		}
		seen[c.Name] = true
		if goivy.IsNumeralName(c.Name) && w.ParentCTI != nil && w.ParentCTI.Mod != nil {
			if goivy.IsUninterpretedSort(w.ParentCTI.Mod.Sig, c.CSort) {
				varName := rn.Generate(fmt.Sprintf("%v", c.CSort))
				v, _ := goivy.NewVariable(varName, c.CSort)
				subs[goivy.Key(c)] = v
			}
		}
	}

	// Python: literals = [negate(substitute(f, subs)) for f in facts]
	var literals []goivy.Expr
	for _, f := range facts {
		substituted, err := goivy.Substitute(f, subs)
		if err != nil {
			substituted = f
		}
		negated := goivy.Negate(substituted)
		literals = append(literals, negated)
	}

	// Python: result = Clauses([Or(*literals)])
	var orExpr goivy.Expr
	if len(literals) == 1 {
		orExpr = literals[0]
	} else {
		orExpr, _ = goivy.NewOr(literals...)
	}

	result := goivy.NewClauses([]goivy.Expr{orExpr}, nil, nil)
	result = goivy.SimplifyClauses(result)

	// Python: convert Or to Not(And(negate(each_lit)))
	if len(result.Fmlas) == 1 {
		if orNode, ok := result.Fmlas[0].(*goivy.LogicOr); ok {
			var innerLits []goivy.Expr
			for _, lit := range orNode.Terms {
				innerLits = append(innerLits, goivy.Negate(lit))
			}
			andExpr, _ := goivy.NewAnd(innerLits...)
			notExpr, _ := goivy.NewNot(andExpr)
			result = goivy.NewClauses([]goivy.Expr{notExpr}, nil, nil)
		}
	}

	return result
}

// MinimizeConjecture minimizes the active conjecture using unsat cores
// (Python: ConceptGraphUI.minimize_conjecture).
func (w *CTIConceptGraphWidget) MinimizeConjecture(bound int) (*goivy.Clauses, error) {
	if w.ParentCTI == nil || w.ParentCTI.Mod == nil {
		return nil, fmt.Errorf("no module loaded")
	}

	// Python: if self.bmc_conjecture(bound=bound, tell_unsat=False): return
	conj := w.GetSelectedConjecture()
	if conj != nil {
		found, _ := w.ParentCTI.BoundedCheck(bound, conj)
		if found {
			return nil, fmt.Errorf("BMC found a counter-example")
		}
	}

	mod := w.ParentCTI.Mod
	nSteps := w.ParentCTI.CurrentBound
	if nSteps < 0 {
		nSteps = bound
	}

	// Python: ag = self.parent.new_ag(); execute env_action n_steps times
	ag := goivy.NewAnalysisGraph(mod)
	ag.Add(goivy.NewState(mod, ag.InitCond), nil)
	post := ag.States[0]

	stepAction := goivy.BMCEnvAction(mod)
	for n := 0; n < nSteps; n++ {
		if stepAction == nil {
			break
		}
		var err error
		post, err = ag.Execute(true, stepAction, post, nil, "")
		if err != nil {
			return nil, fmt.Errorf("step %d failed: %v", n, err)
		}
	}

	axioms := mod.BackgroundTheory(nil)
	postClauses := goivy.AndClausesTyped(post.Clauses, axioms)

	facts := w.ActiveFactExprs
	if len(facts) == 0 {
		return nil, fmt.Errorf("no facts to minimize")
	}

	factsClauses := goivy.NewClauses(facts, nil, nil)

	if w.ParentCTI.Solver == nil {
		return nil, fmt.Errorf("no solver available")
	}

	core, err := w.ParentCTI.Solver.UnsatCore(factsClauses, postClauses, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("unsat_core failed: %w", err)
	}
	if core == nil {
		core = goivy.NewClauses(nil, nil, nil)
	}

	// Python: core_formulas = frozenset(core.fmlas)
	coreSet := make(map[string]bool)
	for _, cf := range core.Fmlas {
		coreSet[fmt.Sprintf("%v", cf)] = true
	}

	// Python: self.set_facts([fact for fact in facts if fact in core_formulas])
	var filteredFacts []goivy.Expr
	var filteredStrs []string
	for _, f := range facts {
		key := fmt.Sprintf("%v", f)
		if coreSet[key] {
			filteredFacts = append(filteredFacts, f)
			filteredStrs = append(filteredStrs, key)
		}
	}
	w.ActiveFactExprs = filteredFacts

	g := w.G()
	if g != nil {
		g.SetFacts(filteredStrs)
	}
	w.HighlightSelectedFacts()

	minimized := w.GetSelectedConjecture()
	return minimized, nil
}

// checkInductionHelper is the shared implementation for IsSufficient and IsInductive.
// (Python: is_sufficient and is_inductive share the same structure)
func (w *CTIConceptGraphWidget) checkInductionHelper(conj, targetConj *goivy.Clauses) (bool, string) {
	if w.ParentCTI == nil || w.ParentCTI.Mod == nil {
		return false, "no module loaded"
	}

	mod := w.ParentCTI.Mod

	// Python: pre.clauses = and_clauses(conj, *self.parent.conjectures)
	pre := conj
	for _, c := range w.ParentCTI.Conjectures {
		pre = goivy.AndClausesTyped(pre, c)
	}
	pre.Annot = goivy.EmptyAnnotation{}

	preState := goivy.NewState(mod, pre)
	ag := goivy.NewAnalysisGraph(mod)
	ag.Add(preState, nil)

	// Python: action = ia.env_action(None); post = ag.execute(action, pre)
	stepAction := goivy.BMCEnvAction(mod)
	if stepAction == nil {
		return false, "no actions available"
	}
	post, err := ag.Execute(true, stepAction, preState, nil, "")
	if err != nil {
		return false, fmt.Sprintf("execute failed: %v", err)
	}
	trueCl := goivy.TrueClauses(nil)
	trueCl.Annot = goivy.EmptyAnnotation{}
	post.Clauses = trueCl

	// Python: clauses = dual_clauses(target_conj, witness)
	usedNames := w.ParentCTI.collectUsedNames()
	witness := ctiWitness(usedNames)
	clauses := goivy.DualClauses(targetConj, witness, nil)
	clauses.Annot = goivy.EmptyAnnotation{}

	res := goivy.CheckFinalCond(ag, post, clauses, nil, true)

	conjStr := fmt.Sprintf("%v", conj.ToFormula())
	targetStr := fmt.Sprintf("%v", targetConj.ToFormula())

	if res != nil {
		return false, fmt.Sprintf("(1) %s\n(2) %s\n(1) does not imply (2) at the next time.", conjStr, targetStr)
	}
	return true, fmt.Sprintf("(1) %s\n(2) %s\n(1) implies (2) at the next time.", conjStr, targetStr)
}

// IsSufficient checks if the active conjecture implies the current CTI conjecture
// (Python: ConceptGraphUI.is_sufficient).
func (w *CTIConceptGraphWidget) IsSufficient() (bool, string) {
	conj := w.GetSelectedConjecture()
	if conj == nil {
		return false, "no conjecture selected"
	}
	if w.ParentCTI == nil || w.ParentCTI.CurrentConjecture == nil {
		return false, "no current CTI conjecture"
	}
	return w.checkInductionHelper(conj, w.ParentCTI.CurrentConjecture)
}

// IsInductive checks if the active conjecture is relatively inductive
// (Python: ConceptGraphUI.is_inductive).
func (w *CTIConceptGraphWidget) IsInductive() (bool, string) {
	conj := w.GetSelectedConjecture()
	if conj == nil {
		return false, "no conjecture selected"
	}
	// Python: target_conj = conj (self-induction check)
	return w.checkInductionHelper(conj, conj)
}

// FormulaToConceptl creates a CDConcept from a formula.
// (Python: concept_from_formula in ivy_graph.py:23-28)
func FormulaToConceptl(fmla goivy.Expr) *CDConcept {
	vs := goivy.UsedVariablesAsts([]goivy.Expr{fmla})
	sort.Slice(vs, func(i, j int) bool {
		return fmt.Sprintf("%v", vs[i]) < fmt.Sprintf("%v", vs[j])
	})

	var parts []string
	for _, v := range vs {
		parts = append(parts, fmt.Sprintf("%v:%v", v, v.VSort))
	}
	name := strings.Join(parts, ",") + "." + fmt.Sprintf("%v", fmla)
	c, _ := NewCDConcept(name, vs, fmla)
	return c
}

// WriteConjecture formats a conjecture for file output.
func WriteConjecture(label, formula string) string {
	if label != "" {
		return fmt.Sprintf("invariant [%s] %s\n", label, formula)
	}
	return fmt.Sprintf("invariant %s\n", formula)
}
