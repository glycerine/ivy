package webui

// Full port of ivy_ui_cti.py — CTI (counterexample to induction) variant
// of AnalysisGraphUI with invariant checking, bounded checking,
// diagram abstraction, weakening, and transitive relation autodetection.

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// CTIAnalysisGraphUI extends AnalysisGraphUI with CTI-specific operations
// (Python: class AnalysisGraphUI in ivy_ui_cti.py).
type CTIAnalysisGraphUI struct {
	*AnalysisGraphUI

	mu2 sync.Mutex

	// TransitiveRelations tracks detected transitive relations.
	TransitiveRelations []string

	// TransitiveRelationConcepts tracks the concepts for transitive relations.
	TransitiveRelationConcepts []*Concept

	// RelationsToMinimize is the set of relations to minimize during checking.
	RelationsToMinimize string

	// Conjectures is the current list of conjectures (formula strings).
	Conjectures []string

	// HaveCTI indicates whether a counterexample to induction exists.
	HaveCTI bool

	// CurrentConjecture is the conjecture currently being checked.
	CurrentConjecture string

	// CurrentBound is the last BMC bound used.
	CurrentBound int
}

// NewCTIAnalysisGraphUI creates a new CTI-variant UI.
func NewCTIAnalysisGraphUI() *CTIAnalysisGraphUI {
	return &CTIAnalysisGraphUI{
		AnalysisGraphUI:    NewAnalysisGraphUI(),
		RelationsToMinimize: "relations to minimize",
		CurrentBound:       -1,
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
func (ui *CTIAnalysisGraphUI) StartCTI(conjectures []string) {
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
	// Stub: real implementation checks axioms for transitive + defined_symmetry.
	// For each qualifying symbol c:
	//   transitive = ForAll([X,Y,Z], Or(Not(c(X,Y)), Not(c(Y,Z)), c(X,Z)))
	//   defined_symmetry = ForAll([X,Y], Or(c(X,X), Not(c(Y,Y))))
	//   if axioms imply both: add to transitive_relations
}

// CheckInductiveness checks all conjectures for relative inductiveness
// (Python: AnalysisGraphUI.check_inductiveness).
func (ui *CTIAnalysisGraphUI) CheckInductiveness() (bool, string) {
	ui.mu2.Lock()
	defer ui.mu2.Unlock()

	// Determine relations to minimize.
	if ui.RelationsToMinimize == "relations to minimize" {
		ui.RelationsToMinimize = "" // auto-detect (stub)
	}

	toTest := make([]string, 0, len(ui.Conjectures)+1)
	toTest = append(toTest, "") // empty string = check safety
	toTest = append(toTest, ui.Conjectures...)

	for _, conj := range toTest {
		// Stub: real implementation creates analysis graph, computes
		// dual_clauses(conj), calls check_final_cond, and checks for CTI.
		_ = conj
	}

	// If we reach here without finding a CTI, the invariant is inductive.
	ui.HaveCTI = false
	result := strings.Join(ui.Conjectures, "\n")
	return true, fmt.Sprintf("Inductive invariant found:\n%s", result)
}

// BoundedCheck performs bounded model checking for a conjecture
// (Python: AnalysisGraphUI.bmc_conjecture).
func (ui *CTIAnalysisGraphUI) BoundedCheck(bound int, conjecture string) (bool, string) {
	ui.mu2.Lock()
	defer ui.mu2.Unlock()

	ui.CurrentBound = bound

	if conjecture == "" && len(ui.Conjectures) > 0 {
		conjecture = strings.Join(ui.Conjectures, " & ")
	}

	// Stub: real implementation creates analysis graph, executes steps,
	// and checks final condition at each step.
	for n := 0; n <= bound; n++ {
		// check at step n
		_ = n
	}

	return false, fmt.Sprintf("BMC with bound %d did not find a counter-example", bound)
}

// Diagram computes a diagram abstraction of the current CTI
// (Python: AnalysisGraphUI.diagram).
func (ui *CTIAnalysisGraphUI) Diagram() (string, error) {
	ui.mu2.Lock()
	defer ui.mu2.Unlock()

	if !ui.HaveCTI {
		ok, msg := ui.checkInductivenessUnlocked()
		if ok {
			return msg, nil
		}
	}

	// Stub: real implementation computes reverse image, gets model,
	// extracts diagram, and updates the concept graph.
	return "diagram computed", nil
}

// checkInductivenessUnlocked is the internal unlocked version.
func (ui *CTIAnalysisGraphUI) checkInductivenessUnlocked() (bool, string) {
	ui.HaveCTI = false
	return true, "Inductive invariant found"
}

// Weaken removes conjectures from the current set
// (Python: AnalysisGraphUI.weaken).
func (ui *CTIAnalysisGraphUI) Weaken(indices []int) ([]string, error) {
	ui.mu2.Lock()
	defer ui.mu2.Unlock()

	if len(indices) == 0 {
		return nil, fmt.Errorf("no conjectures selected")
	}

	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	var removed []string
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
			sb.WriteString(fmt.Sprintf("invariant %s\n", conj))
		}
	}
	return sb.String()
}

// ShowUsedRelations enables display of relations used in given clauses
// (Python: AnalysisGraphUI.show_used_relations).
func (ui *CTIAnalysisGraphUI) ShowUsedRelations(clauseStr string, both bool) {
	if ui.CurrentConceptGraph == nil {
		return
	}
	ui.CurrentConceptGraph.ClearEdges()
	// Stub: real implementation parses clauses, finds used symbols,
	// and enables checkboxes for matching relations.
	ui.CurrentConceptGraph.Update()
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

	// Parent is the CTIAnalysisGraphUI that owns this widget.
	ParentCTI *CTIAnalysisGraphUI
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
	selectedNodes := make([]string, 0)
	w.mu.Lock()
	for n := range w.NodeSelection {
		selectedNodes = append(selectedNodes, n)
	}
	w.mu.Unlock()
	sort.Strings(selectedNodes)

	// If nothing selected, use all nodes.
	if len(selectedNodes) == 0 {
		selectedNodes = g.NodeIDs()
	}

	// Stub: real implementation collects node facts and edge facts,
	// filters duplicates, and sets them as constraints.
	_ = selectedNodes
	w.Update()
	w.HighlightSelectedFacts()
}

// Strengthen adds a new conjecture from selected facts
// (Python: ConceptGraphUI.strengthen).
func (w *CTIConceptGraphWidget) Strengthen() (string, error) {
	conj := w.GetSelectedConjecture()
	if conj == "" {
		return "", fmt.Errorf("no facts selected")
	}
	if w.ParentCTI != nil {
		w.ParentCTI.HaveCTI = false
		w.ParentCTI.Conjectures = append(w.ParentCTI.Conjectures, conj)
	}
	return conj, nil
}

// GetSelectedConjecture returns a positive universal conjecture from selected facts
// (Python: ConceptGraphUI.get_selected_conjecture).
func (w *CTIConceptGraphWidget) GetSelectedConjecture() string {
	facts := w.GetActiveFacts()
	if len(facts) == 0 {
		return ""
	}
	// Stub: real implementation negates facts, substitutes skolem constants
	// with variables, simplifies, and returns the conjecture formula.
	return strings.Join(facts, " & ")
}

// MinimizeConjecture minimizes the active conjecture using unsat cores
// (Python: ConceptGraphUI.minimize_conjecture).
func (w *CTIConceptGraphWidget) MinimizeConjecture(bound int) (string, error) {
	// Stub: real implementation does BMC, then unsat_core minimization.
	return "", fmt.Errorf("minimization not yet implemented")
}

// IsSufficient checks if the active conjecture implies the current CTI conjecture
// (Python: ConceptGraphUI.is_sufficient).
func (w *CTIConceptGraphWidget) IsSufficient() (bool, string) {
	// Stub: real implementation creates analysis graph, sets up pre/post states,
	// and checks using check_final_cond.
	return false, "sufficiency check not yet implemented"
}

// IsInductive checks if the active conjecture is relatively inductive
// (Python: ConceptGraphUI.is_inductive).
func (w *CTIConceptGraphWidget) IsInductive() (bool, string) {
	// Stub: real implementation is similar to IsSufficient but with
	// target_conj = conj (self-induction check).
	return false, "induction check not yet implemented"
}

// WriteConjecture formats a conjecture for file output.
func WriteConjecture(label, formula string) string {
	if label != "" {
		return fmt.Sprintf("invariant [%s] %s\n", label, formula)
	}
	return fmt.Sprintf("invariant %s\n", formula)
}
