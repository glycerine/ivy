package webui

// Full port of ivy_ui.py — AnalysisGraphUI for verification-graph
// interaction with modes, node/edge actions, and menu system.

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"sort"
	"sync"
)

// VerificationMode represents a verification approach.
type VerificationMode string

const (
	ModeAbstract  VerificationMode = "abstract"
	ModeConcrete  VerificationMode = "concrete"
	ModeBounded   VerificationMode = "bounded"
	ModeInduction VerificationMode = "induction"
	ModePDR       VerificationMode = "pdr"
)

// AllModes lists all verification modes.
var AllModes = []VerificationMode{
	ModeAbstract,
	ModeConcrete,
	ModeBounded,
	ModeInduction,
	ModePDR,
}

// DefaultMode is the default verification mode.
var DefaultMode = ModePDR

// RadioOption is a label/value pair for radio-button menus.
type RadioOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// ARGStateRef is a lightweight reference to an ARG state node,
// used to avoid pulling in the full art package.
type ARGStateRef struct {
	ID      int
	Clauses string
	IsSafe  bool
	HasPred bool
	PredID  int
	Label   string
}

// AnalysisGraphUI manages the ARG display and user interactions
// (Python: class AnalysisGraphUI).
type AnalysisGraphUI struct {
	mu sync.Mutex

	// G is the analysis graph (ARG) lightweight rendering state.
	G *WebUIAnalysisGraphState

	// AG is the real analysis graph (Python: self.g).
	AG *goivy.AnalysisGraph

	// Mod is the compiled module providing axioms, sig, actions.
	Mod *goivy.Module

	// SyncCallback is called after AG mutations to update G for the frontend.
	SyncCallback func()

	// AlphaFn returns the current abstractor based on mode.
	// Set by the caller (e.g., session wiring) to avoid circular import with alpha.
	AlphaFn func() goivy.Abstractor

	// CurrentConceptGraph is the currently displayed concept graph widget.
	CurrentConceptGraph *GraphWidget

	// Mode is the current verification mode.
	Mode VerificationMode

	// Mark is the currently marked ARG node (for cover, join).
	Mark *ARGStateRef

	// Radios stores radio-button state.
	Radios map[string]string

	// RememberedGraphs stores named proof goals.
	RememberedGraphs map[string]*Graph

	// UIParent is used for dialog interactions.
	UIParent interface{}
}

// NewAnalysisGraphUI creates a new AnalysisGraphUI.
func NewAnalysisGraphUI() *AnalysisGraphUI {
	return &AnalysisGraphUI{
		G:                NewWebUIAnalysisGraphState(),
		Mode:             DefaultMode,
		Radios:           map[string]string{"mode": string(DefaultMode)},
		RememberedGraphs: make(map[string]*Graph),
	}
}

// Menus returns the full menu structure (Python: AnalysisGraphUI.menus).
func (ui *AnalysisGraphUI) Menus() []MenuDef {
	return []MenuDef{
		{
			Type:  "menu",
			Label: "File",
			Items: []MenuItem{
				{Type: "button", Label: "Save", Action: "save"},
				{Type: "button", Label: "Save abstraction", Action: "save_abstraction"},
				{Type: "separator", Label: "---"},
				{Type: "button", Label: "Remove tab", Action: "remove_tab"},
				{Type: "button", Label: "Exit", Action: "exit"},
			},
		},
		{
			Type:  "menu",
			Label: "Mode",
			Items: []MenuItem{
				{Type: "button", Label: "Concrete", Action: "mode_concrete"},
				{Type: "button", Label: "Abstract", Action: "mode_abstract"},
				{Type: "button", Label: "Bounded", Action: "mode_bounded"},
				{Type: "button", Label: "Induction", Action: "mode_induction"},
				{Type: "button", Label: "Pdr", Action: "mode_pdr"},
			},
		},
		{
			Type:  "menu",
			Label: "Action",
			Items: []MenuItem{
				{Type: "button", Label: "Recalculate all", Action: "recalculate_all"},
				{Type: "button", Label: "Show reachable states", Action: "show_reachable"},
			},
		},
	}
}

// SetMode changes the verification mode.
func (ui *AnalysisGraphUI) SetMode(mode VerificationMode) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.Mode = mode
	ui.Radios["mode"] = string(mode)
}

// GetMode returns the current verification mode.
func (ui *AnalysisGraphUI) GetMode() VerificationMode {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	return ui.Mode
}

// getAlpha returns the current abstractor. Returns nil if no AlphaFn is set.
func (ui *AnalysisGraphUI) getAlpha() goivy.Abstractor {
	if ui.AlphaFn != nil {
		return ui.AlphaFn()
	}
	return nil
}

// stateByID returns the art.State for the given nodeID with bounds checking.
func (ui *AnalysisGraphUI) stateByID(nodeID int) (*goivy.State, error) {
	if ui.AG == nil {
		return nil, fmt.Errorf("no analysis graph")
	}
	if nodeID < 0 || nodeID >= len(ui.AG.States) {
		return nil, fmt.Errorf("invalid node ID: %d", nodeID)
	}
	return ui.AG.States[nodeID], nil
}

// transitionByEndpoints finds the transition matching srcID→tgtID.
func (ui *AnalysisGraphUI) transitionByEndpoints(srcID, tgtID int) (*goivy.Transition, error) {
	if ui.AG == nil {
		return nil, fmt.Errorf("no analysis graph")
	}
	for i := range ui.AG.Transitions {
		t := &ui.AG.Transitions[i]
		if t.Pre != nil && t.Post != nil && t.Pre.ID == srcID && t.Post.ID == tgtID {
			return t, nil
		}
	}
	return nil, fmt.Errorf("no transition from %d to %d", srcID, tgtID)
}

// sync calls the SyncCallback if set (updates lightweight G from AG).
func (ui *AnalysisGraphUI) sync() {
	if ui.SyncCallback != nil {
		ui.SyncCallback()
	}
}

// ArtToGraphState converts an art.AnalysisGraph to a lightweight WebUIAnalysisGraphState.
func ArtToGraphState(ag *goivy.AnalysisGraph) *WebUIAnalysisGraphState {
	gs := NewWebUIAnalysisGraphState()
	for _, st := range ag.States {
		label := st.Label
		if label == "" {
			label = fmt.Sprintf("%d", st.ID)
		}
		gs.States = append(gs.States, WebUIARGNode{
			ID:       st.ID,
			Label:    label,
			IsBottom: st.IsBottom(),
			Info:     fmt.Sprintf("State %d", st.ID),
		})
	}
	for _, t := range ag.Transitions {
		gs.Transitions = append(gs.Transitions, WebUIARGTransition{
			SourceID: t.Pre.ID,
			TargetID: t.Post.ID,
			Label:    t.Label,
		})
	}
	for _, c := range ag.Covering {
		gs.Covering = append(gs.Covering, WebUIARGCover{
			CoveredID:  c.Covered.ID,
			CoveringID: c.Covering.ID,
		})
	}
	return gs
}

// defEquationLabel extracts a display label from a state equation (ast.Definition).
// Python: state_equation_label(a) — reads a.args[0] (action name) and a.args[1].rep.
func defEquationLabel(eq *goivy.Definition) string {
	var actionName string
	if eq != nil && eq.Lhs != nil {
		if atom, ok := eq.Lhs.(*goivy.Atom); ok {
			actionName = atom.Rep
		}
	}
	var transLabel string
	if eq != nil && eq.Rhs != nil {
		if atom, ok := eq.Rhs.(*goivy.Atom); ok {
			transLabel = atom.Rep
		}
	}
	return StateEquationLabel(actionName, transLabel)
}

// Start initializes the UI, creating an initial ARG node if needed
// (Python: AnalysisGraphUI.start).
func (ui *AnalysisGraphUI) Start() {
	if len(ui.G.States) == 0 {
		ui.G.States = append(ui.G.States, WebUIARGNode{
			ID:    0,
			Label: "0",
		})
	}
}

// NodeColor returns the display color for an ARG node (Python: AnalysisGraphUI.node_color).
func (ui *AnalysisGraphUI) NodeColor(node *ARGStateRef) string {
	if node != nil && node.IsSafe {
		return "green"
	}
	return "black"
}

// GetNodeActions returns context-menu actions for an ARG node
// (Python: AnalysisGraphUI.get_node_actions).
func (ui *AnalysisGraphUI) GetNodeActions(nodeID int, click string) []ActionEntry {
	if click == "left" {
		return []ActionEntry{
			{Label: "<>", Action: "view_state"},
		}
	}
	// Right click: execute actions + node commands.
	actions := []ActionEntry{
		{Label: "Execute action:", Action: ""},
		{Label: "---", Action: ""},
	}
	actions = append(actions, ui.NodeExecuteCommands(nodeID)...)
	actions = append(actions, ActionEntry{Label: "---", Action: ""})
	actions = append(actions, ui.NodeCommands()...)
	return actions
}

// NodeCommands returns the standard node command entries.
func (ui *AnalysisGraphUI) NodeCommands() []ActionEntry {
	return []ActionEntry{
		{Label: "Check safety", Action: "check_safety"},
		{Label: "Extend", Action: "find_extension"},
		{Label: "Mark", Action: "mark_node"},
		{Label: "Cover by marked", Action: "cover_node"},
		{Label: "Join with marked", Action: "join_node"},
		{Label: "Try conjecture", Action: "try_conjecture"},
		{Label: "Try remembered goal", Action: "try_remembered"},
		{Label: "Delete", Action: "delete_node"},
	}
}

// NodeExecuteCommands returns execute-action entries for a node
// (Python: AnalysisGraphUI.node_execute_commands).
func (ui *AnalysisGraphUI) NodeExecuteCommands(nodeID int) []ActionEntry {
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return nil
	}
	equations := ui.AG.StateActions(state)
	sort.Slice(equations, func(i, j int) bool {
		return defEquationLabel(equations[i]) < defEquationLabel(equations[j])
	})
	var entries []ActionEntry
	for _, eq := range equations {
		label := defEquationLabel(eq)
		entries = append(entries, ActionEntry{
			Label:  label,
			Action: fmt.Sprintf("execute_%d_%s", nodeID, label),
		})
	}
	return entries
}

// GetEdgeActions returns context-menu actions for an ARG edge
// (Python: AnalysisGraphUI.get_edge_actions).
func (ui *AnalysisGraphUI) GetEdgeActions(click string) []ActionEntry {
	if click == "left" {
		return nil
	}
	return []ActionEntry{
		{Label: "Dismiss", Action: "dismiss"},
		{Label: "Recalculate", Action: "recalculate_edge"},
		{Label: "Step in", Action: "decompose_edge"},
		{Label: "View Source", Action: "view_source_edge"},
	}
}

// ViewState shows a state in the concept graph (Python: AnalysisGraphUI.view_state).
func (ui *AnalysisGraphUI) ViewState(nodeID int, clauses string, reset bool) *GraphWidget {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	state, err := ui.stateByID(nodeID)
	if err != nil {
		return nil
	}
	if clauses == "" && state.Clauses != nil {
		clauses = state.Clauses.String()
	}

	if ui.CurrentConceptGraph != nil {
		// Reuse existing concept graph.
		ui.CurrentConceptGraph.SetParentState(state, clauses, reset)
		return ui.CurrentConceptGraph
	}

	// Create new concept graph widget.
	sorts := ui.conceptSortNames()
	gs := StandardGraph(sorts, state)
	w := NewGraphWidget(gs)
	w.Parent = ui
	if clauses != "" {
		w.G().SetState(clauses, true, false, reset)
	}
	ui.CurrentConceptGraph = w
	return w
}

func (ui *AnalysisGraphUI) conceptSortNames() []string {
	if ui.Mod == nil || ui.Mod.Sig == nil {
		return []string{"node"}
	}
	var sorts []string
	for name := range ui.Mod.Sig.Sorts.All() {
		if name != "bool" {
			sorts = append(sorts, name)
		}
	}
	sort.Strings(sorts)
	if len(sorts) == 0 {
		return []string{"node"}
	}
	return sorts
}

// MarkNode sets the marked ARG node (Python: AnalysisGraphUI.mark_node).
func (ui *AnalysisGraphUI) MarkNode(node *ARGStateRef) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.Mark = node
}

// GetMark returns the currently marked node.
func (ui *AnalysisGraphUI) GetMark() *ARGStateRef {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	return ui.Mark
}

// CheckSafetyNode checks safety of a node according to the current mode
// (Python: AnalysisGraphUI.check_safety_node).
func (ui *AnalysisGraphUI) CheckSafetyNode(nodeID int) (bool, string) {
	mode := ui.GetMode()
	if mode != ModeBounded && mode != ModeInduction {
		return ui.CheckLocalSafety(nodeID)
	}
	return ui.CheckBoundedSafety(nodeID)
}

// CheckLocalSafety checks local safety of a node
// (Python: AnalysisGraphUI.check_local_safety).
func (ui *AnalysisGraphUI) CheckLocalSafety(nodeID int) (bool, string) {
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return false, err.Error()
	}
	result := ui.AG.CheckSafety(true, state)
	if result.Safe {
		return true, "Node is safe"
	}
	msg := "The node is not proved safe"
	if result.Cex != nil && result.Cex.Msg != "" {
		msg = fmt.Sprintf("The node is not proved safe: %s", result.Cex.Msg)
	}
	return false, msg
}

// CheckBoundedSafety checks bounded safety along a path
// (Python: AnalysisGraphUI.check_bounded_safety).
func (ui *AnalysisGraphUI) CheckBoundedSafety(nodeID int) (bool, string) {
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return false, err.Error()
	}
	result := ui.AG.CheckBoundedSafety(state, nil)
	if result.Safe {
		return true, "Node is safe (bounded check)"
	}
	msg := "The node is unsafe"
	if result.Cex != nil && result.Cex.Msg != "" {
		msg = fmt.Sprintf("The node is unsafe: %s", result.Cex.Msg)
	}
	return false, msg
}

// FindExtension finds an action to extend the ARG at a node
// (Python: AnalysisGraphUI.find_extension).
func (ui *AnalysisGraphUI) FindExtension(nodeID int) (string, error) {
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return "", err
	}
	extensions := ui.AG.StateExtensions(state, nil)
	if len(extensions) == 0 {
		return "", fmt.Errorf("state %d is closed", nodeID)
	}
	s := ui.AG.DoStateAction(false, extensions[0], ui.getAlpha())
	if s == nil {
		return "", fmt.Errorf("state action evaluation failed")
	}
	ui.sync()
	return defEquationLabel(extensions[0]), nil
}

// ExecuteAction evaluates an action at a node (Python: AnalysisGraphUI.execute_action).
func (ui *AnalysisGraphUI) ExecuteAction(nodeID int, actionName string) error {
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return err
	}
	_, err = ui.AG.ExecuteAction(false, actionName, state, ui.getAlpha())
	if err != nil {
		return err
	}
	ui.sync()
	return nil
}

// RecalculateAll re-evaluates all ARG transitions (Python: AnalysisGraphUI.recalculate_all).
func (ui *AnalysisGraphUI) RecalculateAll() {
	if ui.AG == nil {
		return
	}
	done := make(map[int]bool)
	for _, t := range ui.AG.Transitions {
		if t.Post != nil && !done[t.Post.ID] {
			ui.AG.Recalculate(false, t, ui.getAlpha())
			done[t.Post.ID] = true
		}
	}
	ui.sync()
}

// RecalculateEdge re-evaluates one ARG edge
// (Python: AnalysisGraphUI.recalculate_edge).
func (ui *AnalysisGraphUI) RecalculateEdge(srcID, tgtID int) {
	t, err := ui.transitionByEndpoints(srcID, tgtID)
	if err != nil {
		return
	}
	ui.AG.Recalculate(false, *t, ui.getAlpha())
	ui.sync()
}

// DecomposeEdge decomposes a transition into sub-actions
// (Python: AnalysisGraphUI.decompose_edge).
func (ui *AnalysisGraphUI) DecomposeEdge(srcID, tgtID int) (*WebUIAnalysisGraphState, error) {
	t, err := ui.transitionByEndpoints(srcID, tgtID)
	if err != nil {
		return nil, err
	}
	var subArt *goivy.AnalysisGraph
	func() {
		defer func() {
			if recover() != nil {
				subArt = nil
			}
		}()
		subArt = ui.AG.DecomposeEdge(*t)
	}()
	if subArt == nil {
		subArt = ui.fallbackStepInGraph(t)
	}
	if subArt == nil {
		return nil, fmt.Errorf("cannot decompose action")
	}
	return ArtToGraphState(subArt), nil
}

func (ui *AnalysisGraphUI) fallbackStepInGraph(t *goivy.Transition) *goivy.AnalysisGraph {
	if ui == nil || ui.AG == nil || t == nil || t.Pre == nil || t.Post == nil || t.Op == nil {
		return nil
	}
	subArt := goivy.NewAnalysisGraph(ui.AG.Domain)
	pre := goivy.NewState(ui.AG.Domain, t.Pre.Clauses)
	pre.Label = t.Pre.Label
	post := goivy.NewState(ui.AG.Domain, t.Post.Clauses)
	post.Label = t.Post.Label
	subArt.Add(pre, nil)
	subArt.Add(post, goivy.NewActionApp(t.Op, pre))
	return subArt
}

// ViewSourceEdge browses the source code of a transition action
// (Python: AnalysisGraphUI.view_source_edge).
func (ui *AnalysisGraphUI) ViewSourceEdge(srcID, tgtID int) (string, int, error) {
	t, err := ui.transitionByEndpoints(srcID, tgtID)
	if err != nil {
		return "", 0, err
	}
	if t.Op == nil {
		return "", 0, fmt.Errorf("no action on this edge")
	}
	if !t.Op.HasLineno() {
		return "", 0, fmt.Errorf("no source location available")
	}
	loc := t.Op.GetLineno()
	return loc.Filename, loc.Line, nil
}

// DeleteNode removes a node from the ARG (Python: AnalysisGraphUI.delete_node).
func (ui *AnalysisGraphUI) DeleteNode(nodeID int) {
	if ui.AG != nil {
		state, err := ui.stateByID(nodeID)
		if err == nil {
			ui.AG.Delete(state)
		}
	}
	if ui.Mark != nil && ui.Mark.ID == nodeID {
		ui.Mark = nil
	}
	ui.sync()
}

// CoverNode tries to cover one node by the marked node
// (Python: AnalysisGraphUI.cover_node).
func (ui *AnalysisGraphUI) CoverNode(coveredID int) (bool, error) {
	mark := ui.GetMark()
	if mark == nil {
		return false, fmt.Errorf("no marked node")
	}
	covered, err := ui.stateByID(coveredID)
	if err != nil {
		return false, err
	}
	covering, err := ui.stateByID(mark.ID)
	if err != nil {
		return false, err
	}
	ok := ui.AG.Cover(covered, covering)
	if !ok {
		return false, fmt.Errorf("covering failed")
	}
	ui.sync()
	return true, nil
}

// JoinNode joins a node with the marked node
// (Python: AnalysisGraphUI.join_node).
func (ui *AnalysisGraphUI) JoinNode(nodeID int) error {
	mark := ui.GetMark()
	if mark == nil {
		return fmt.Errorf("no marked node")
	}
	state1, err := ui.stateByID(mark.ID)
	if err != nil {
		return err
	}
	state2, err := ui.stateByID(nodeID)
	if err != nil {
		return err
	}
	joined := ui.AG.Join(state1, state2, ui.getAlpha())
	if joined == nil {
		return fmt.Errorf("join failed")
	}
	ui.sync()
	return nil
}

// TryConjecture sets up to prove a conjecture at a node
// (Python: AnalysisGraphUI.try_conjecture).
func (ui *AnalysisGraphUI) TryConjecture(nodeID int, conjecture string) error {
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return err
	}
	if conjecture == "" {
		return fmt.Errorf("no conjecture specified")
	}
	fmla, parseErr := goivy.ToFormula(conjecture)
	if parseErr != nil {
		return fmt.Errorf("parse conjecture: %w", parseErr)
	}
	fExpr, ok := fmla.(goivy.Expr)
	if !ok {
		return fmt.Errorf("conjecture is not a logic expression")
	}
	conj := goivy.FormulaToClauses(fExpr, nil)
	dual := goivy.DualClauses(conj, nil, nil)

	mode := ui.GetMode()
	if mode == ModeInduction || mode == ModeBounded {
		bmcResult := ui.AG.BMC(state, dual.ToFormula(), nil, nil)
		if bmcResult == nil {
			return fmt.Errorf("the condition is unreachable along the given path")
		}
		return nil
	}
	return nil
}

// TryRememberedGraph loads a previously saved proof goal
// (Python: AnalysisGraphUI.try_remembered_graph).
func (ui *AnalysisGraphUI) TryRememberedGraph(nodeID int, goalName string) error {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	if goalName == "" {
		return nil
	}
	sg, ok := ui.RememberedGraphs[goalName]
	if !ok {
		return fmt.Errorf("no remembered graph named %q", goalName)
	}
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return err
	}
	sgCopy := sg.Copy()
	sgCopy.ParentState = state
	return nil
}

// RememberGraph saves a proof goal under a name.
func (ui *AnalysisGraphUI) RememberGraph(name string, g *Graph) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.RememberedGraphs[name] = g
}

// RememberedGraphNames returns the names of remembered graphs in sorted order.
func (ui *AnalysisGraphUI) RememberedGraphNames() []string {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	var names []string
	for n := range ui.RememberedGraphs {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// BMC performs bounded model checking from initial to a state
// (Python: AnalysisGraphUI.bmc).
func (ui *AnalysisGraphUI) BMC(nodeID int, errCond string, bound int) (*WebUIAnalysisGraphState, error) {
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return nil, err
	}
	fmla, parseErr := goivy.ToFormula(errCond)
	if parseErr != nil {
		return nil, fmt.Errorf("parse error condition: %w", parseErr)
	}
	fExpr, ok := fmla.(goivy.Expr)
	if !ok {
		return nil, fmt.Errorf("error condition is not a logic expression")
	}
	var boundPtr *int
	if bound >= 0 {
		boundPtr = &bound
	}
	resultAG := ui.AG.BMC(state, fExpr, nil, boundPtr)
	if resultAG == nil {
		return nil, fmt.Errorf("condition unreachable along given path")
	}
	return ArtToGraphState(resultAG), nil
}

// StateLabel returns the display label for an ARG state.
func (ui *AnalysisGraphUI) StateLabel(stateID int) string {
	return fmt.Sprintf("%d", stateID)
}

// StateEquationLabel formats a state equation for display.
func StateEquationLabel(actionName, transLabel string) string {
	if actionName == "" {
		return transLabel
	}
	return transLabel + " -> " + actionName
}

// IvyUI is the top-level UI class (Python: class IvyUI).
type IvyUI struct {
	mu  sync.Mutex
	Mod *goivy.Module
}

// NewIvyUI creates a new top-level IvyUI.
func NewIvyUI() *IvyUI {
	return &IvyUI{}
}

// AGUI returns the default AnalysisGraphUI class.
func (ui *IvyUI) AGUI() *AnalysisGraphUI {
	return NewAnalysisGraphUI()
}

// TryProperty sets up to prove a background property
// (Python: IvyUI.try_property).
func (ui *IvyUI) TryProperty(propText string) error {
	if propText == "" {
		return fmt.Errorf("no property specified")
	}
	if ui.Mod == nil {
		return fmt.Errorf("no module loaded")
	}
	fmla, parseErr := goivy.ToFormula(propText)
	if parseErr != nil {
		return fmt.Errorf("parse property: %w", parseErr)
	}
	fExpr, ok := fmla.(goivy.Expr)
	if !ok {
		return fmt.Errorf("property is not a logic expression")
	}
	conj := goivy.FormulaToClauses(fExpr, nil)
	dual := goivy.DualClauses(conj, nil, nil)

	topAlpha := goivy.AbstractorFunc(func(s *goivy.State) {
		s.Clauses = goivy.TrueClauses(nil)
	})
	ag := goivy.NewAnalysisGraph(ui.Mod)
	ag.AddInitialState(nil, topAlpha)
	if len(ag.States) == 0 {
		return fmt.Errorf("failed to create initial state")
	}
	oag := ag.BMC(ag.States[0], dual.ToFormula(), nil, nil)
	if oag == nil {
		return fmt.Errorf("property holds (no counterexample found)")
	}
	return nil
}
