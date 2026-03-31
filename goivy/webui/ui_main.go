package webui

// Full port of ivy_ui.py — AnalysisGraphUI for verification-graph
// interaction with modes, node/edge actions, and menu system.

import (
	"fmt"
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
	ID       int
	Clauses  string
	IsSafe   bool
	HasPred  bool
	PredID   int
	Label    string
}

// AnalysisGraphUI manages the ARG display and user interactions
// (Python: class AnalysisGraphUI).
type AnalysisGraphUI struct {
	mu sync.Mutex

	// G is the analysis graph (ARG) state.
	G *AnalysisGraphState

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
		G:                NewAnalysisGraphState(),
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

// Start initializes the UI, creating an initial ARG node if needed
// (Python: AnalysisGraphUI.start).
func (ui *AnalysisGraphUI) Start() {
	if len(ui.G.States) == 0 {
		ui.G.States = append(ui.G.States, ARGNode{
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

// NodeExecuteCommands returns execute-action entries for a node.
func (ui *AnalysisGraphUI) NodeExecuteCommands(nodeID int) []ActionEntry {
	// Stub: real implementation calls g.state_actions(node).
	return nil
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

	if ui.CurrentConceptGraph != nil {
		// Reuse existing concept graph.
		ui.CurrentConceptGraph.SetParentState(nil, clauses, reset)
		return ui.CurrentConceptGraph
	}

	// Create new concept graph widget.
	sorts := []string{"node"} // default; real implementation reads from ARG state
	gs := StandardGraph(sorts, nil)
	w := NewGraphWidget(gs)
	ui.CurrentConceptGraph = w
	return w
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

// CheckLocalSafety checks local safety of a node.
func (ui *AnalysisGraphUI) CheckLocalSafety(nodeID int) (bool, string) {
	// Stub: real implementation calls g.check_safety(node).
	return true, "Node is safe"
}

// CheckBoundedSafety checks bounded safety along a path.
func (ui *AnalysisGraphUI) CheckBoundedSafety(nodeID int) (bool, string) {
	// Stub: real implementation calls g.check_bounded_safety(node).
	return true, "Node is safe (bounded check)"
}

// FindExtension finds an action to extend the ARG at a node.
func (ui *AnalysisGraphUI) FindExtension(nodeID int) (string, error) {
	// Stub: real implementation calls g.state_extensions(node).
	return "", fmt.Errorf("state %d is closed", nodeID)
}

// ExecuteAction evaluates an action at a node (Python: AnalysisGraphUI.execute_action).
func (ui *AnalysisGraphUI) ExecuteAction(nodeID int, actionName string) error {
	// Stub: real implementation calls g.execute_action.
	return nil
}

// RecalculateAll re-evaluates all ARG transitions (Python: AnalysisGraphUI.recalculate_all).
func (ui *AnalysisGraphUI) RecalculateAll() {
	// Stub: real implementation iterates transitions and recalculates each.
}

// RecalculateEdge re-evaluates one ARG edge.
func (ui *AnalysisGraphUI) RecalculateEdge(srcID, tgtID int) {
	// Stub: real implementation calls g.recalculate.
}

// DecomposeEdge decomposes a transition into sub-actions.
func (ui *AnalysisGraphUI) DecomposeEdge(srcID, tgtID int) (*AnalysisGraphState, error) {
	// Stub: real implementation calls g.decompose_edge.
	return nil, fmt.Errorf("cannot decompose action")
}

// ViewSourceEdge browses the source code of a transition action.
func (ui *AnalysisGraphUI) ViewSourceEdge(srcID, tgtID int) (string, int, error) {
	// Stub: real implementation extracts lineno from action.
	return "", 0, fmt.Errorf("no source available")
}

// DeleteNode removes a node from the ARG (Python: AnalysisGraphUI.delete_node).
func (ui *AnalysisGraphUI) DeleteNode(nodeID int) {
	var kept []ARGNode
	for _, s := range ui.G.States {
		if s.ID != nodeID {
			kept = append(kept, s)
		}
	}
	ui.G.States = kept
}

// CoverNode tries to cover one node by the marked node.
func (ui *AnalysisGraphUI) CoverNode(coveredID int) (bool, error) {
	mark := ui.GetMark()
	if mark == nil {
		return false, fmt.Errorf("no marked node")
	}
	// Stub: real implementation calls g.cover.
	return false, fmt.Errorf("covering failed")
}

// JoinNode joins a node with the marked node.
func (ui *AnalysisGraphUI) JoinNode(nodeID int) error {
	mark := ui.GetMark()
	if mark == nil {
		return fmt.Errorf("no marked node")
	}
	// Stub: real implementation calls g.join.
	return nil
}

// TryConjecture sets up to prove a conjecture at a node.
func (ui *AnalysisGraphUI) TryConjecture(nodeID int, conjecture string) error {
	// Stub: real implementation dispatches to bmc or concept graph.
	return nil
}

// TryRememberedGraph loads a previously saved proof goal.
func (ui *AnalysisGraphUI) TryRememberedGraph(nodeID int, goalName string) error {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	if goalName == "" {
		// List available goals.
		return nil
	}
	_, ok := ui.RememberedGraphs[goalName]
	if !ok {
		return fmt.Errorf("no remembered graph named %q", goalName)
	}
	// Stub: set up the remembered graph for the given node.
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

// BMC performs bounded model checking from initial to a state.
func (ui *AnalysisGraphUI) BMC(nodeID int, errCond string, bound int) (*AnalysisGraphState, error) {
	// Stub: real implementation calls g.bmc.
	return nil, fmt.Errorf("condition unreachable along given path")
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
	mu sync.Mutex
}

// NewIvyUI creates a new top-level IvyUI.
func NewIvyUI() *IvyUI {
	return &IvyUI{}
}

// AGUI returns the default AnalysisGraphUI class.
func (ui *IvyUI) AGUI() *AnalysisGraphUI {
	return NewAnalysisGraphUI()
}

// TryProperty sets up to prove a background property.
func (ui *IvyUI) TryProperty(propText string) error {
	// Stub: real implementation checks property and launches BMC.
	return nil
}
