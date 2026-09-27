package webui

// Full port of ivy_ui.py — AnalysisGraphUI for verification-graph
// interaction with modes, node/edge actions, and menu system.

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"reflect"
	"sort"
	"strconv"
	"strings"
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

type ChoiceItem struct {
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

type NodeSafetyCheckResult struct {
	Safe    bool
	Message string
	Trace   *goivy.AnalysisGraph
}

type TryConjectureResult struct {
	Mode       VerificationMode
	View       string
	Conjecture string
	SourceLoc  goivy.Location
	Concept    *GraphWidget
	Trace      *goivy.AnalysisGraph
	Reachable  bool
	Message    string
}

type ClosedStateError struct {
	StateID int
}

func (e *ClosedStateError) Error() string {
	return fmt.Sprintf("state %d is closed", e.StateID)
}

func (e *ClosedStateError) DialogMessage() string {
	return fmt.Sprintf("State %d is closed.", e.StateID)
}

type ExhaustedARGActionError struct {
	StateID    int
	ActionName string
	Label      string
}

func (e *ExhaustedARGActionError) Error() string {
	if e == nil {
		return "ARG action exhausted"
	}
	if e.Label != "" {
		return fmt.Sprintf("ARG action %q at state %d only reproduced existing transition %q", e.ActionName, e.StateID, e.Label)
	}
	return fmt.Sprintf("ARG action %q at state %d produced no new transition", e.ActionName, e.StateID)
}

func (e *ExhaustedARGActionError) DialogMessage() string {
	if e == nil {
		return "No new transition was found."
	}
	if e.Label != "" {
		return fmt.Sprintf("No new %s transition from state %d was found.", cleanARGActionDisplayName(e.ActionName), e.StateID)
	}
	return fmt.Sprintf("No new transition from state %d was found.", e.StateID)
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

	// ConceptDomainFn returns a freshly initialized concept domain for ARG state
	// views that need a solver-backed interactive session.
	ConceptDomainFn func() (*CDConceptDomain, error)

	// CurrentConceptGraph is the currently displayed concept graph widget.
	CurrentConceptGraph *GraphWidget

	// Mode is the current verification mode.
	Mode VerificationMode

	// Mark is the currently marked ARG node (for cover, join).
	Mark *ARGStateRef

	// SafeNodes records nodes whose safety checks succeeded.
	SafeNodes map[int]bool

	// Radios stores radio-button state.
	Radios map[string]string

	// RememberedGraphs stores named proof goals.
	RememberedGraphs map[string]*Graph

	// UIParent is used for dialog interactions.
	UIParent interface{}

	// ExtensionConfig holds dynamic extension-point actions exposed to browsers.
	ExtensionConfig *ExtConfig
}

// NewAnalysisGraphUI creates a new AnalysisGraphUI.
func NewAnalysisGraphUI() *AnalysisGraphUI {
	return &AnalysisGraphUI{
		G:                NewWebUIAnalysisGraphState(),
		Mode:             DefaultMode,
		SafeNodes:        make(map[int]bool),
		Radios:           map[string]string{"mode": string(DefaultMode)},
		RememberedGraphs: make(map[string]*Graph),
		ExtensionConfig:  NewExtConfig(),
	}
}

// Menus returns the full menu structure (Python: AnalysisGraphUI.menus).
func (ui *AnalysisGraphUI) Menus() []MenuDef {
	return []MenuDef{
		{
			Type:  "menu",
			Label: "File",
			Items: []MenuItem{
				{Type: "button", Label: "Save", Action: "save_model"},
				{Type: "button", Label: "Save analysis state", Action: "save_analysis_state"},
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

func (ui *AnalysisGraphUI) initialConceptDomain() (*CDConceptDomain, error) {
	if ui != nil && ui.ConceptDomainFn != nil {
		return ui.ConceptDomainFn()
	}
	sorts := make(map[string]goivy.Sort)
	symbols := make(map[string]*goivy.Const)
	if ui != nil && ui.Mod != nil && ui.Mod.Sig != nil {
		if ui.Mod.Sig.Sorts != nil {
			for name, sortVal := range ui.Mod.Sig.Sorts.All() {
				if sortVal != nil {
					sorts[name] = sortVal
				}
			}
		}
		if ui.Mod.Sig.Symbols != nil {
			for name, entry := range ui.Mod.Sig.Symbols.All() {
				if entry == nil || entry.Sort == nil {
					continue
				}
				if sortVal, ok := entry.Sort.(goivy.Sort); ok {
					symbols[name] = goivy.NewConst(name, sortVal)
				}
			}
		}
	}
	return GetInitialConceptDomainE(sorts, symbols)
}

func (ui *AnalysisGraphUI) ensureInteractiveConceptSession(w *GraphWidget, state *goivy.State) error {
	if w == nil || w.G() == nil {
		return fmt.Errorf("no concept graph")
	}
	g := w.G()
	if g.InteractiveSess != nil {
		return nil
	}
	domain, err := ui.initialConceptDomain()
	if err != nil {
		return err
	}
	stateExpr := goivy.True
	if state != nil && state.Clauses != nil {
		stateExpr = state.Clauses.ToFormula()
	}
	axioms := goivy.True
	if ui != nil && ui.Mod != nil {
		axiomTerms := ui.Mod.Axioms()
		if len(axiomTerms) > 0 {
			var axiomErr error
			axioms, axiomErr = goivy.NewAnd(axiomTerms...)
			if axiomErr != nil {
				return axiomErr
			}
		}
	}
	sess, err := NewConceptInteractiveSessionE(domain, stateExpr, axioms, nil, nil, nil, nil, nil, false)
	if err != nil {
		return err
	}
	g.InteractiveSess = sess
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

func (ui *AnalysisGraphUI) applyARGMetadata(gs *WebUIAnalysisGraphState) {
	if ui == nil || gs == nil {
		return
	}
	safeNodes, markedID, hasMark := ui.argRenderMetadata()
	for i := range gs.States {
		id := gs.States[i].ID
		gs.States[i].IsSafe = safeNodes[id]
		gs.States[i].IsMarked = hasMark && id == markedID
	}
}

func (ui *AnalysisGraphUI) applyFullARGMetadata(gs *goivy.FullAnalysisGraphState) {
	if ui == nil || gs == nil {
		return
	}
	safeNodes, markedID, hasMark := ui.argRenderMetadata()
	for i := range gs.States {
		id := gs.States[i].ID
		gs.States[i].IsSafe = safeNodes[id]
		gs.States[i].IsMarked = hasMark && id == markedID
	}
}

func (ui *AnalysisGraphUI) argRenderMetadata() (map[int]bool, int, bool) {
	safeNodes := make(map[int]bool)
	if ui == nil {
		return safeNodes, 0, false
	}
	ui.mu.Lock()
	defer ui.mu.Unlock()
	for id, safe := range ui.SafeNodes {
		if safe {
			safeNodes[id] = true
		}
	}
	if ui.Mark == nil {
		return safeNodes, 0, false
	}
	return safeNodes, ui.Mark.ID, true
}

func (ui *AnalysisGraphUI) SetNodeSafety(nodeID int, safe bool) {
	if ui == nil {
		return
	}
	ui.mu.Lock()
	defer ui.mu.Unlock()
	if ui.SafeNodes == nil {
		ui.SafeNodes = make(map[int]bool)
	}
	if safe {
		ui.SafeNodes[nodeID] = true
		return
	}
	delete(ui.SafeNodes, nodeID)
}

func (ui *AnalysisGraphUI) IsNodeSafe(nodeID int) bool {
	if ui == nil {
		return false
	}
	ui.mu.Lock()
	defer ui.mu.Unlock()
	return ui.SafeNodes[nodeID]
}

// ArtToGraphState converts an art.AnalysisGraph to a lightweight WebUIAnalysisGraphState.
func ArtToGraphState(ag *goivy.AnalysisGraph) *WebUIAnalysisGraphState {
	gs := NewWebUIAnalysisGraphState()
	if ag != nil {
		ag.CanonicalizeTransitionActionNames()
	}
	for _, st := range ag.States {
		label := st.Label
		if label == "" {
			label = fmt.Sprintf("%d", st.ID)
		}
		gs.States = append(gs.States, WebUIARGNode{
			ID:       st.ID,
			Label:    label,
			IsBottom: st.IsBottom(),
			Info:     argNodeInfo(st),
		})
	}
	for _, t := range ag.Transitions {
		gs.Transitions = append(gs.Transitions, WebUIARGTransition{
			SourceID: t.Pre.ID,
			TargetID: t.Post.ID,
			Label:    argTransitionDisplayLabel(t),
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

func argNodeInfo(st *goivy.State) string {
	if st.Clauses != nil {
		return st.Clauses.String()
	}
	return fmt.Sprintf("State %d", st.ID)
}

// ArtToFullGraphState converts an AnalysisGraph to FullAnalysisGraphState,
// including formula strings and universe data for CTI inspection.
func ArtToFullGraphState(ag *goivy.AnalysisGraph) *goivy.FullAnalysisGraphState {
	gs := goivy.NewFullAnalysisGraphState()
	if ag != nil {
		ag.CanonicalizeTransitionActionNames()
	}
	for _, st := range ag.States {
		label := st.Label
		if label == "" {
			label = fmt.Sprintf("%d", st.ID)
		}
		gs.States = append(gs.States, goivy.FullARGNode{
			ID:         st.ID,
			Label:      label,
			IsBottom:   st.IsBottom(),
			Info:       argNodeInfo(st),
			Clauses:    argNodeClauses(st),
			ActionName: st.ActionName,
			Universe:   argNodeUniverse(st),
		})
	}
	for _, t := range ag.Transitions {
		gs.Transitions = append(gs.Transitions, goivy.ARGTransition{
			SourceID: t.Pre.ID,
			TargetID: t.Post.ID,
			Label:    argTransitionDisplayLabel(t),
		})
	}
	for _, c := range ag.Covering {
		gs.Covering = append(gs.Covering, goivy.ARGCover{
			CoveredID:  c.Covered.ID,
			CoveringID: c.Covering.ID,
		})
	}
	return gs
}

func argNodeClauses(st *goivy.State) string {
	if st.Clauses == nil {
		return ""
	}
	if f := st.Clauses.ToOpenFormula(); f != nil {
		return fmt.Sprint(f)
	}
	return st.Clauses.String()
}

func argTransitionDisplayLabel(t goivy.Transition) string {
	if t.IsJoin() {
		return "join"
	}
	return mustARGTransitionDisplayLabel(argTransitionDisplayLabelCandidate(t), t)
}

func argTransitionDisplayLabelCandidate(t goivy.Transition) string {
	actionName := strings.TrimSpace(t.ActionName)
	if actionName == "" {
		return ""
	}
	var base string
	if display := actionDisplayNameForModelActionName(argTransitionModule(t), actionName); display != "" {
		base = display
	} else if display := cleanARGActionDisplayName(actionName); display != "" {
		base = display
	} else {
		base = actionName
	}
	if withActuals := argTransitionDisplayLabelWithActuals(t, base); withActuals != "" {
		return withActuals
	}
	return base
}

func mustARGTransitionDisplayLabel(label string, t goivy.Transition) string {
	actionName := strings.TrimSpace(t.ActionName)
	if actionName == "" {
		panic(fmt.Sprintf("ARG transition edge is missing canonical action name: raw=%q op=%T pre=%s post=%s", t.Label, t.Op, argTransitionStateID(t.Pre), argTransitionStateID(t.Post)))
	}
	if strings.EqualFold(actionName, "sequence") {
		panic(fmt.Sprintf("ARG transition edge action name resolved to internal action name %q: raw=%q op=%T pre=%s post=%s", actionName, t.Label, t.Op, argTransitionStateID(t.Pre), argTransitionStateID(t.Post)))
	}
	if argLabelLooksLikeGoImplementation(actionName) {
		panic(fmt.Sprintf("ARG transition edge action name leaked Go implementation detail %q: raw=%q op=%T pre=%s post=%s", actionName, t.Label, t.Op, argTransitionStateID(t.Pre), argTransitionStateID(t.Post)))
	}
	if strings.EqualFold(actionName, "ext") || strings.EqualFold(actionName, "call ext") || strings.EqualFold(actionName, "call:ext") {
		panic(fmt.Sprintf("ARG transition edge action name resolved to generic external dispatcher %q: raw=%q op=%T pre=%s post=%s", actionName, t.Label, t.Op, argTransitionStateID(t.Pre), argTransitionStateID(t.Post)))
	}
	if mod := argTransitionModule(t); mod != nil && mod.Actions != nil {
		if _, ok := mod.Actions.Get2(actionName); !ok {
			panic(fmt.Sprintf("ARG transition edge action name %q is not in the model: raw=%q op=%T pre=%s post=%s", actionName, t.Label, t.Op, argTransitionStateID(t.Pre), argTransitionStateID(t.Post)))
		}
	}
	trimmed := strings.TrimSpace(label)
	if trimmed == "" {
		panic(fmt.Sprintf("ARG transition edge label resolved to empty string: raw=%q op=%T pre=%s post=%s", t.Label, t.Op, argTransitionStateID(t.Pre), argTransitionStateID(t.Post)))
	}
	if strings.EqualFold(trimmed, "sequence") {
		panic(fmt.Sprintf("ARG transition edge label resolved to internal action name %q: raw=%q op=%T pre=%s post=%s", trimmed, t.Label, t.Op, argTransitionStateID(t.Pre), argTransitionStateID(t.Post)))
	}
	return label
}

func argTransitionModule(t goivy.Transition) *goivy.Module {
	if t.Post != nil && t.Post.Domain != nil {
		return t.Post.Domain
	}
	if t.Pre != nil {
		return t.Pre.Domain
	}
	return nil
}

func argTransitionStateID(st *goivy.State) string {
	if st == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d", st.ID)
}

func actionDisplayNameFromTransitionContext(t goivy.Transition) string {
	if display := actionDisplayNameFromPostState(t.Post, t.Pre, t.Op); display != "" {
		return display
	}
	if display := actionDisplayNameFromStateModule(t.Post, t.Op); display != "" {
		return display
	}
	if display := actionDisplayNameFromStateModule(t.Pre, t.Op); display != "" {
		return display
	}
	return ""
}

func actionDisplayNameFromPostState(post, pre *goivy.State, op goivy.ActionsAction) string {
	if post == nil {
		return ""
	}
	if display := actionDisplayNameForModelActionName(post.Domain, post.ActionName); display != "" {
		return display
	}
	if post.Action != nil {
		if display := actionDisplayNameInModule(post.Domain, post.Action); display != "" {
			return display
		}
		if op == nil || sameARGActionIdentity(post.Action, op) {
			if display := actionDisplayNameForARG(post.Action); display != "" {
				return display
			}
		}
	}
	aa, ok := post.Prov.(*goivy.ActionApp)
	if !ok || aa == nil {
		return ""
	}
	if pre != nil && (len(aa.Args) == 0 || aa.Args[0] != pre) {
		return ""
	}
	switch rep := aa.Rep.(type) {
	case string:
		if display := actionDisplayNameForModelActionName(post.Domain, rep); display != "" {
			return display
		}
		if post.Domain == nil || post.Domain.Actions == nil {
			return cleanARGActionDisplayName(rep)
		}
	case *goivy.Const:
		if display := actionDisplayNameForModelActionName(post.Domain, rep.Name); display != "" {
			return display
		}
		if post.Domain == nil || post.Domain.Actions == nil {
			return cleanARGActionDisplayName(rep.Name)
		}
	case goivy.ActionsAction:
		if display := actionDisplayNameInModule(post.Domain, rep); display != "" {
			return display
		}
		if op == nil || sameARGActionIdentity(rep, op) {
			return actionDisplayNameForARG(rep)
		}
	}
	return ""
}

func actionDisplayNameFromStateModule(st *goivy.State, action goivy.ActionsAction) string {
	if st == nil {
		return ""
	}
	return actionDisplayNameInModule(st.Domain, action)
}

func actionDisplayNameForModelActionName(mod *goivy.Module, name string) string {
	display := cleanARGActionDisplayName(name)
	if display == "" {
		return ""
	}
	if mod == nil || mod.Actions == nil {
		return display
	}
	if _, ok := mod.Actions.Get2(name); ok {
		return display
	}
	if strings.HasPrefix(name, "ext:") {
		if _, ok := mod.Actions.Get2(strings.TrimPrefix(name, "ext:")); ok {
			return display
		}
		return ""
	}
	if _, ok := mod.Actions.Get2("ext:" + name); ok {
		return display
	}
	return ""
}

func actionDisplayNameInModule(mod *goivy.Module, action goivy.ActionsAction) string {
	if mod == nil || mod.Actions == nil || action == nil {
		return ""
	}
	for name, candidate := range mod.Actions.All() {
		if sameARGActionIdentity(candidate, action) {
			if display := cleanARGActionDisplayName(name); display != "" {
				return display
			}
		}
	}
	var structuralDisplay string
	for name, candidate := range mod.Actions.All() {
		if !sameARGActionStructure(candidate, action) {
			continue
		}
		display := cleanARGActionDisplayName(name)
		if display == "" {
			continue
		}
		if structuralDisplay != "" && structuralDisplay != display {
			return ""
		}
		structuralDisplay = display
	}
	if structuralDisplay != "" {
		return structuralDisplay
	}
	return ""
}

func sameARGActionIdentity(a, b goivy.ActionsAction) bool {
	if a == nil || b == nil {
		return false
	}
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	if !av.IsValid() || !bv.IsValid() || av.Type() != bv.Type() {
		return false
	}
	if av.Kind() == reflect.Ptr {
		if av.IsNil() || bv.IsNil() {
			return false
		}
		return av.Pointer() == bv.Pointer()
	}
	if av.Type().Comparable() {
		return a == b
	}
	return false
}

func sameARGActionStructure(a, b goivy.ActionsAction) (same bool) {
	if a == nil || b == nil {
		return false
	}
	defer func() {
		if recover() != nil {
			same = false
		}
	}()
	return a.Equal(b) || b.Equal(a)
}

func argLabelLooksLikeGoImplementation(label string) bool {
	return strings.Contains(label, "goivy.") || strings.Contains(label, "@0x")
}

func argTransitionLabelIsGeneric(label string) bool {
	trimmed := strings.TrimSpace(label)
	lower := strings.ToLower(trimmed)
	return lower == "call ext" ||
		lower == "call:ext" ||
		lower == "ext" ||
		strings.HasPrefix(lower, "ext:") ||
		argLabelLooksLikeGoImplementation(trimmed)
}

func actionDisplayNameForARG(action goivy.ActionsAction) string {
	switch a := action.(type) {
	case nil:
		return ""
	case *goivy.FailAction:
		return actionDisplayNameForARG(a.Inner)
	case *goivy.LogicEnvAction:
		if len(a.Branches) == 1 {
			for _, branch := range a.Branches {
				if branchAction, ok := branch.(goivy.ActionsAction); ok {
					if display := actionDisplayNameForARG(branchAction); display != "" {
						return display
					}
				}
			}
		}
		return cleanARGActionDisplayName(a.GetLabel())
	case *goivy.LogicSequence:
		if display := cleanARGActionDisplayName(a.GetLabel()); display != "" {
			return display
		}
		for _, elem := range a.Elems {
			if elemAction, ok := elem.(goivy.ActionsAction); ok {
				if display := actionDisplayNameForARG(elemAction); display != "" {
					return display
				}
			}
		}
		return ""
	case *goivy.LogicCallAction:
		return cleanARGActionDisplayName(a.CalleeName())
	default:
		if labeler, ok := action.(interface{ GetLabel() string }); ok {
			return cleanARGActionDisplayName(labeler.GetLabel())
		}
		return ""
	}
}

func cleanARGActionDisplayName(label string) string {
	label = strings.TrimSpace(label)
	if argLabelLooksLikeGoImplementation(label) {
		return ""
	}
	label = strings.TrimPrefix(label, "call:")
	label = strings.TrimPrefix(label, "call ")
	label = strings.TrimPrefix(label, "ext:")
	label = strings.TrimSpace(label)
	if label == "" ||
		strings.EqualFold(label, "ext") ||
		strings.EqualFold(label, "call ext") ||
		argLabelLooksLikeGoImplementation(label) {
		return ""
	}
	return label
}

func argTransitionDisplayLabelWithActuals(t goivy.Transition, base string) string {
	if base == "" {
		return ""
	}
	formals := argTransitionFormalParams(t)
	if len(formals) == 0 {
		return ""
	}
	if actuals := argTransitionActualsFromFormalEqualities(t.Post, formals); len(actuals) == len(formals) {
		return argTransitionFormatActualLabel(base, actuals)
	}
	consts := argTransitionPostConstants(t)
	if len(consts) == 0 {
		return ""
	}
	if actuals := argTransitionActualsByFormalName(formals, consts); len(actuals) == len(formals) {
		return argTransitionFormatActualLabel(base, actuals)
	}
	if actuals := argTransitionActualsFromStateDiff(t, formals); len(actuals) == len(formals) {
		return argTransitionFormatActualLabel(base, actuals)
	}
	return ""
}

func argTransitionActualsFromFormalEqualities(state *goivy.State, formals []*goivy.Const) []*goivy.Const {
	if state == nil || state.Clauses == nil || len(formals) == 0 {
		return nil
	}
	actualByFormal := make(map[goivy.NodeKey]*goivy.Const, len(formals))
	visit := func(expr goivy.Expr) {
		if eq, ok := expr.(*goivy.Eq); ok {
			argTransitionRecordFormalEquality(eq.T1, eq.T2, formals, actualByFormal)
			argTransitionRecordFormalEquality(eq.T2, eq.T1, formals, actualByFormal)
		}
	}
	for _, fmla := range state.Clauses.Fmlas {
		visit(fmla)
	}
	for _, def := range state.Clauses.Defs {
		if def != nil {
			visit(def)
		}
	}
	actuals := make([]*goivy.Const, 0, len(formals))
	for _, formal := range formals {
		actual := actualByFormal[goivy.Key(formal)]
		if actual == nil {
			return nil
		}
		actuals = append(actuals, actual)
	}
	return actuals
}

func argTransitionRecordFormalEquality(lhs, rhs goivy.Expr, formals []*goivy.Const, actualByFormal map[goivy.NodeKey]*goivy.Const) {
	lhsConst, ok := lhs.(*goivy.Const)
	if !ok || lhsConst == nil {
		return
	}
	rhsConst, ok := rhs.(*goivy.Const)
	if !ok || rhsConst == nil {
		return
	}
	for _, formal := range formals {
		if formal == nil || !lhsConst.Equal(formal) || !sameARGSortName(lhsConst.CSort, rhsConst.CSort) {
			continue
		}
		actualByFormal[goivy.Key(formal)] = rhsConst
		return
	}
}

func argTransitionActualsByFormalName(formals []*goivy.Const, consts []*goivy.Const) []*goivy.Const {
	resolved := make([]*goivy.Const, 0, len(formals))
	for _, formal := range formals {
		actual := argTransitionFindFormalActual(formal, consts)
		if actual == nil {
			return nil
		}
		resolved = append(resolved, actual)
	}
	return resolved
}

func argTransitionFormatActualLabel(base string, actuals []*goivy.Const) string {
	labels := make([]string, 0, len(actuals))
	aliases := make(map[goivy.NodeKey]string)
	nextBySort := make(map[string]int)
	for _, actual := range actuals {
		labels = append(labels, argTransitionActualDisplay(actual, aliases, nextBySort))
	}
	return fmt.Sprintf("%s(%s)", base, strings.Join(labels, ", "))
}

type argTransitionRelationFact struct {
	Key      goivy.NodeKey
	Args     []*goivy.Const
	Positive bool
}

func argTransitionActualsFromStateDiff(t goivy.Transition, formals []*goivy.Const) []*goivy.Const {
	if t.Pre == nil || t.Post == nil || t.Pre.Clauses == nil || t.Post.Clauses == nil {
		return nil
	}
	preFacts := make(map[goivy.NodeKey]bool)
	for _, fact := range argTransitionRelationFacts(t.Pre.Clauses) {
		preFacts[fact.Key] = fact.Positive
	}
	for _, fact := range argTransitionRelationFacts(t.Post.Clauses) {
		if positive, ok := preFacts[fact.Key]; ok && positive == fact.Positive {
			continue
		}
		if actuals := argTransitionFactActuals(formals, fact.Args); len(actuals) == len(formals) {
			return actuals
		}
	}
	return nil
}

func argTransitionRelationFacts(clauses *goivy.Clauses) []argTransitionRelationFact {
	if clauses == nil {
		return nil
	}
	var facts []argTransitionRelationFact
	for _, fmla := range clauses.Fmlas {
		if fact, ok := argTransitionRelationFactFromExpr(fmla); ok {
			facts = append(facts, fact)
		}
	}
	for _, def := range clauses.Defs {
		if def == nil {
			continue
		}
		if fact, ok := argTransitionRelationFactFromExpr(def); ok {
			facts = append(facts, fact)
		}
	}
	return facts
}

func argTransitionRelationFactFromExpr(expr goivy.Expr) (argTransitionRelationFact, bool) {
	switch e := expr.(type) {
	case *goivy.Eq:
		if app, ok := argTransitionRelationApp(e.T1); ok {
			if value, ok := argTransitionBoolValue(e.T2); ok {
				return argTransitionBuildRelationFact(app, value)
			}
		}
		if app, ok := argTransitionRelationApp(e.T2); ok {
			if value, ok := argTransitionBoolValue(e.T1); ok {
				return argTransitionBuildRelationFact(app, value)
			}
		}
	case *goivy.LogicDefinition:
		if app, ok := argTransitionRelationApp(e.Lhs); ok {
			if value, ok := argTransitionBoolValue(e.Rhs); ok {
				return argTransitionBuildRelationFact(app, value)
			}
		}
	case *goivy.Apply:
		if app, ok := argTransitionRelationApp(e); ok {
			return argTransitionBuildRelationFact(app, true)
		}
	case *goivy.LogicNot:
		if app, ok := argTransitionRelationApp(e.Body); ok {
			return argTransitionBuildRelationFact(app, false)
		}
	}
	return argTransitionRelationFact{}, false
}

func argTransitionRelationApp(expr goivy.Expr) (*goivy.Apply, bool) {
	app, ok := expr.(*goivy.Apply)
	if !ok || app == nil || !goivy.SortEqual(app.NodeSort(), goivy.Boolean) {
		return nil, false
	}
	for _, term := range app.Terms {
		if _, ok := term.(*goivy.Const); !ok {
			return nil, false
		}
	}
	return app, true
}

func argTransitionBuildRelationFact(app *goivy.Apply, positive bool) (argTransitionRelationFact, bool) {
	if app == nil {
		return argTransitionRelationFact{}, false
	}
	args := make([]*goivy.Const, 0, len(app.Terms))
	for _, term := range app.Terms {
		c, ok := term.(*goivy.Const)
		if !ok {
			return argTransitionRelationFact{}, false
		}
		args = append(args, c)
	}
	return argTransitionRelationFact{
		Key:      goivy.Key(app),
		Args:     args,
		Positive: positive,
	}, true
}

func argTransitionBoolValue(expr goivy.Expr) (bool, bool) {
	if goivy.IsTrue(expr) {
		return true, true
	}
	if goivy.IsFalse(expr) {
		return false, true
	}
	return false, false
}

func argTransitionFactActuals(formals []*goivy.Const, args []*goivy.Const) []*goivy.Const {
	if len(formals) == 0 || len(args) < len(formals) {
		return nil
	}
	actuals := make([]*goivy.Const, 0, len(formals))
	used := make([]bool, len(args))
	for _, formal := range formals {
		found := -1
		for i, arg := range args {
			if used[i] || arg == nil || formal == nil || !sameARGSortName(arg.CSort, formal.CSort) {
				continue
			}
			found = i
			break
		}
		if found < 0 {
			return nil
		}
		used[found] = true
		actuals = append(actuals, args[found])
	}
	return actuals
}

func argTransitionFormalParams(t goivy.Transition) []*goivy.Const {
	if t.Op != nil {
		if formals := t.Op.GetFormalParams(); len(formals) > 0 {
			return formals
		}
	}
	mod := argTransitionModule(t)
	if mod == nil || mod.Actions == nil {
		return nil
	}
	for _, name := range argTransitionActionNameCandidates(t.ActionName) {
		if action, ok := mod.Actions.Get2(name); ok && action != nil {
			if formals := action.GetFormalParams(); len(formals) > 0 {
				return formals
			}
		}
	}
	return nil
}

func argTransitionActionNameCandidates(actionName string) []string {
	actionName = strings.TrimSpace(actionName)
	if actionName == "" {
		return nil
	}
	raw := []string{actionName}
	if strings.HasPrefix(actionName, "ext:") {
		raw = append(raw, strings.TrimPrefix(actionName, "ext:"))
	} else {
		raw = append(raw, "ext:"+actionName)
	}
	seen := make(map[string]bool, len(raw))
	var out []string
	for _, name := range raw {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

func argTransitionPostConstants(t goivy.Transition) []*goivy.Const {
	var consts []*goivy.Const
	seen := make(map[goivy.NodeKey]bool)
	add := func(expr goivy.Expr) {
		argTransitionCollectConstants(expr, &consts, seen)
	}
	if t.Post != nil && t.Post.Clauses != nil {
		for _, fmla := range t.Post.Clauses.Fmlas {
			add(fmla)
		}
		for _, def := range t.Post.Clauses.Defs {
			add(def)
		}
	}
	sort.Slice(consts, func(i, j int) bool {
		if consts[i].Name == consts[j].Name {
			return goivy.IvySortName(consts[i].CSort) < goivy.IvySortName(consts[j].CSort)
		}
		return consts[i].Name < consts[j].Name
	})
	return consts
}

func argTransitionCollectConstants(expr goivy.Expr, out *[]*goivy.Const, seen map[goivy.NodeKey]bool) {
	if expr == nil {
		return
	}
	if c, ok := expr.(*goivy.Const); ok {
		key := goivy.Key(c)
		if !seen[key] {
			seen[key] = true
			*out = append(*out, c)
		}
	}
	for _, child := range expr.Children() {
		argTransitionCollectConstants(child, out, seen)
	}
}

func argTransitionFindFormalActual(formal *goivy.Const, consts []*goivy.Const) *goivy.Const {
	if formal == nil {
		return nil
	}
	formalName := strings.TrimSpace(formal.Name)
	shortFormalName := strings.TrimPrefix(formalName, "fml:")
	var fallback *goivy.Const
	for _, c := range consts {
		if c == nil || !sameARGSortName(c.CSort, formal.CSort) {
			continue
		}
		name := strings.TrimSpace(c.Name)
		if name == formalName || name == "__"+formalName {
			return c
		}
		if strings.HasSuffix(name, formalName) && strings.Contains(name, "fml:") {
			if fallback == nil {
				fallback = c
			}
			continue
		}
		if shortFormalName != "" && strings.HasSuffix(name, shortFormalName) && strings.Contains(name, "fml:") {
			if fallback == nil {
				fallback = c
			}
		}
	}
	return fallback
}

func sameARGSortName(a, b goivy.Sort) bool {
	if a == nil || b == nil {
		return a == b
	}
	return goivy.IvySortName(a) == goivy.IvySortName(b)
}

func argTransitionActualDisplay(actual *goivy.Const, aliases map[goivy.NodeKey]string, nextBySort map[string]int) string {
	if actual == nil {
		return ""
	}
	key := goivy.Key(actual)
	if alias := aliases[key]; alias != "" {
		return alias
	}
	sortName := goivy.IvySortName(actual.CSort)
	name := strings.TrimSpace(actual.Name)
	if argTransitionConstNeedsDiagramAlias(name) {
		idx := nextBySort[sortName]
		nextBySort[sortName] = idx + 1
		alias := fmt.Sprintf("%d:%s", idx, sortName)
		aliases[key] = alias
		return alias
	}
	if sortName != "" && sortName != "Top" && sortName != "alpha" && !strings.Contains(name, ":") {
		return fmt.Sprintf("%s:%s", name, sortName)
	}
	return fmt.Sprint(actual)
}

func argTransitionConstNeedsDiagramAlias(name string) bool {
	return strings.Contains(name, "fml:") ||
		strings.HasPrefix(name, "__") ||
		strings.HasPrefix(name, "new_") ||
		strings.HasPrefix(name, "old_")
}

func applyCounterexampleTraceTransitionLabels(trace *goivy.TraceBase) {
	if trace == nil || trace.AnalysisGraph == nil {
		return
	}
	trace.AnalysisGraph.CanonicalizeTransitionActionNames()
	exported := exportedActionNamesByDisplay(trace.AnalysisGraph.Domain)
	if len(exported) == 0 {
		return
	}
	traceStates := make(map[*goivy.State]*goivy.TraceState, len(trace.TraceStates))
	for _, ts := range trace.TraceStates {
		if ts != nil && ts.State != nil {
			traceStates[ts.State] = ts
		}
	}
	for i := range trace.AnalysisGraph.Transitions {
		t := &trace.AnalysisGraph.Transitions[i]
		current := argTransitionDisplayLabelCandidate(*t)
		if actionName := exported[current]; actionName != "" {
			t.ActionName = actionName
			t.Label = current
			continue
		}
		if !argTransitionLabelIsGeneric(t.Label) && !argTransitionLabelIsGeneric(t.ActionName) {
			continue
		}
		if display := traceStateExportedActionDisplayName(traceStates[t.Post], exported); display != "" {
			t.ActionName = exported[display]
			t.Label = display
			continue
		}
		if display, actionName, ok := soleExportedAction(exported); ok {
			t.ActionName = actionName
			t.Label = display
		}
	}
}

func soleExportedAction(exported map[string]string) (display string, actionName string, ok bool) {
	if len(exported) != 1 {
		return "", "", false
	}
	for display, actionName := range exported {
		return display, actionName, true
	}
	return "", "", false
}

func exportedActionNamesByDisplay(mod *goivy.Module) map[string]string {
	names := make(map[string]string)
	if mod == nil || mod.PublicActions == nil {
		return names
	}
	for name := range mod.PublicActions.All() {
		if display := cleanARGActionDisplayName(name); display != "" {
			names[display] = name
		}
	}
	return names
}

func traceStateExportedActionDisplayName(ts *goivy.TraceState, exported map[string]string) string {
	if ts == nil || ts.Subgraph == nil || ts.Subgraph.Graph == nil {
		return ""
	}
	return traceGraphExportedActionDisplayName(ts.Subgraph.Graph, exported, make(map[*goivy.TraceBase]bool))
}

func traceGraphExportedActionDisplayName(trace *goivy.TraceBase, exported map[string]string, seen map[*goivy.TraceBase]bool) string {
	if trace == nil || seen[trace] {
		return ""
	}
	seen[trace] = true
	if trace.AnalysisGraph != nil {
		for _, tr := range trace.AnalysisGraph.Transitions {
			for _, candidate := range []string{
				actionDisplayNameForARG(tr.Op),
				cleanARGActionDisplayName(tr.Label),
				argTransitionDisplayLabelCandidate(tr),
			} {
				if exported[candidate] != "" {
					return candidate
				}
			}
		}
	}
	for _, ts := range trace.TraceStates {
		if ts == nil || ts.Subgraph == nil {
			continue
		}
		if display := traceGraphExportedActionDisplayName(ts.Subgraph.Graph, exported, seen); display != "" {
			return display
		}
	}
	return ""
}

func argNodeUniverse(st *goivy.State) map[string][]string {
	if st.Universe == nil {
		return nil
	}
	switch u := st.Universe.(type) {
	case map[string][]goivy.Expr:
		out := make(map[string][]string, len(u))
		for sort, exprs := range u {
			strs := make([]string, len(exprs))
			for i, e := range exprs {
				strs[i] = fmt.Sprint(e)
			}
			out[sort] = strs
		}
		return out
	default:
		return nil
	}
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

func stateEquationActionName(eq *goivy.Definition) string {
	if eq == nil || eq.Rhs == nil {
		return ""
	}
	if atom, ok := eq.Rhs.(*goivy.Atom); ok {
		return atom.Rep
	}
	return ""
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
	if node != nil && (node.IsSafe || ui.IsNodeSafe(node.ID)) {
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
	if extActions := ui.NodeExtensionCommands(nodeID); len(extActions) > 0 {
		actions = append(actions, ActionEntry{Label: "---", Action: ""})
		actions = append(actions, extActions...)
	}
	return actions
}

// NodeCommands returns the standard node command entries.
func (ui *AnalysisGraphUI) NodeCommands() []ActionEntry {
	return []ActionEntry{
		{Label: "Check safety", Action: "check_safety"},
		{Label: "Bounded check", Action: "bmc"},
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
		actionName := stateEquationActionName(eq)
		entries = append(entries, ActionEntry{
			Label:  label,
			Action: "execute_action",
			Args: map[string]interface{}{
				"action_name":  actionName,
				"action_label": label,
			},
		})
	}
	return entries
}

// NodeExtensionCommands returns dynamic extension-point entries for a node.
func (ui *AnalysisGraphUI) NodeExtensionCommands(nodeID int) []ActionEntry {
	if ui == nil || ui.ExtensionConfig == nil || ui.ExtensionConfig.ArgNodeActions == nil {
		return nil
	}
	extActions, _ := ui.ExtensionConfig.ArgNodeActions.Invoke(ui, nodeID)
	var entries []ActionEntry
	seen := make(map[string]bool)
	for _, ext := range extActions {
		label := strings.TrimSpace(ext.Label)
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		entries = append(entries, ActionEntry{
			Label:  label,
			Action: "extension_arg_node",
			Args: map[string]interface{}{
				"extension_label": label,
			},
		})
	}
	return entries
}

func actionSelectionStrings(raw interface{}) []string {
	switch v := raw.(type) {
	case nil:
		return nil
	case []string:
		return append([]string{}, v...)
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, fmt.Sprint(item))
		}
		return out
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	default:
		return []string{fmt.Sprint(v)}
	}
}

// RunArgNodeExtension dispatches a browser-selected ARG node extension action.
func (ui *AnalysisGraphUI) RunArgNodeExtension(nodeID int, label string, args map[string]interface{}) error {
	if ui == nil || ui.ExtensionConfig == nil || ui.ExtensionConfig.ArgNodeActions == nil {
		return fmt.Errorf("no ARG node extension actions registered")
	}
	label = strings.TrimSpace(label)
	if label == "" {
		return fmt.Errorf("missing extension label")
	}
	extActions, errs := ui.ExtensionConfig.ArgNodeActions.Invoke(ui, nodeID)
	if len(errs) > 0 && len(extActions) == 0 {
		return errs[0]
	}
	for _, ext := range extActions {
		if strings.TrimSpace(ext.Label) != label {
			continue
		}
		if ext.Callback == nil {
			return nil
		}
		callbackArgs := []interface{}{nodeID}
		if args != nil {
			if selection, ok := args["selection"]; ok {
				callbackArgs = append(callbackArgs, actionSelectionStrings(selection))
			}
		}
		return ext.Callback(callbackArgs...)
	}
	return fmt.Errorf("ARG node extension action %q not found", label)
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
func (ui *AnalysisGraphUI) ViewState(nodeID int, clauses string, reset bool) (*GraphWidget, error) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	state, err := ui.stateByID(nodeID)
	if err != nil {
		return nil, err
	}
	if clauses == "" && state.Clauses != nil {
		clauses = state.Clauses.String()
	}

	if ui.CurrentConceptGraph != nil {
		// Reuse existing concept graph.
		if err := ui.CurrentConceptGraph.SetParentState(state, clauses, reset); err != nil {
			return nil, err
		}
		return ui.CurrentConceptGraph, nil
	}

	// Create new concept graph widget.
	sorts := ui.conceptSortNames()
	gs := StandardGraph(sorts, state)
	w := NewGraphWidget(gs)
	w.Parent = ui
	if clauses != "" {
		if err := w.G().SetState(clauses, true, false, reset); err != nil {
			return nil, err
		}
	}
	ui.CurrentConceptGraph = w
	return w, nil
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
	result := ui.CheckSafetyNodeResult(nodeID, "")
	return result.Safe, result.Message
}

func (ui *AnalysisGraphUI) CheckSafetyNodeResult(nodeID int, mode VerificationMode) NodeSafetyCheckResult {
	if mode == "" {
		mode = ui.GetMode()
	}
	var result NodeSafetyCheckResult
	if mode != ModeBounded && mode != ModeInduction {
		result = ui.CheckLocalSafetyResult(nodeID)
	} else {
		result = ui.CheckBoundedSafetyResult(nodeID)
	}
	ui.SetNodeSafety(nodeID, result.Safe)
	return result
}

// CheckLocalSafety checks local safety of a node
// (Python: AnalysisGraphUI.check_local_safety).
func (ui *AnalysisGraphUI) CheckLocalSafety(nodeID int) (bool, string) {
	result := ui.CheckLocalSafetyResult(nodeID)
	return result.Safe, result.Message
}

func (ui *AnalysisGraphUI) CheckLocalSafetyResult(nodeID int) NodeSafetyCheckResult {
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return NodeSafetyCheckResult{Safe: false, Message: err.Error()}
	}
	result := ui.AG.CheckSafety(true, state)
	if result.Safe {
		return NodeSafetyCheckResult{Safe: true, Message: "Node is safe"}
	}
	msg := "The node is not proved safe"
	if result.Cex != nil && result.Cex.Msg != "" {
		msg = fmt.Sprintf("The node is not proved safe: %s", result.Cex.Msg)
	}
	return NodeSafetyCheckResult{Safe: false, Message: msg}
}

// CheckBoundedSafety checks bounded safety along a path
// (Python: AnalysisGraphUI.check_bounded_safety).
func (ui *AnalysisGraphUI) CheckBoundedSafety(nodeID int) (bool, string) {
	result := ui.CheckBoundedSafetyResult(nodeID)
	return result.Safe, result.Message
}

func (ui *AnalysisGraphUI) CheckBoundedSafetyResult(nodeID int) NodeSafetyCheckResult {
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return NodeSafetyCheckResult{Safe: false, Message: err.Error()}
	}
	result := ui.AG.CheckBoundedSafety(state, nil)
	if result.Safe {
		return NodeSafetyCheckResult{Safe: true, Message: "Node is safe (bounded check)"}
	}
	msg := "The node is unsafe"
	if result.Art != nil {
		msg = "The node is unsafe: View error trace?"
	} else if result.Cex != nil && result.Cex.Msg != "" {
		msg = fmt.Sprintf("The node is unsafe: %s", result.Cex.Msg)
	}
	return NodeSafetyCheckResult{Safe: false, Message: msg, Trace: result.Art}
}

func (ui *AnalysisGraphUI) markFinalTraceState() {
	if ui == nil || ui.AG == nil || len(ui.AG.States) == 0 {
		return
	}
	finalState := ui.AG.States[len(ui.AG.States)-1]
	ui.MarkNode(&ARGStateRef{ID: finalState.ID})
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
		return "", &ClosedStateError{StateID: nodeID}
	}
	s := ui.AG.DoStateAction(false, extensions[0], ui.getAlpha())
	if s == nil {
		return "", fmt.Errorf("state action evaluation failed")
	}
	actionName := stateEquationActionName(extensions[0])
	if actionName != "" {
		ui.AG.Add(s, goivy.NewActionApp(ui.resolveActionName(actionName), state))
	} else {
		ui.AG.Add(s, nil)
	}
	if _, err := ui.ViewState(s.ID, "", true); err != nil {
		return "", err
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
	if handled, err := ui.executeConcreteFreshAction(nodeID, state, actionName); handled || err != nil {
		return err
	}
	existingLabels := ui.transitionLabelsFromState(state)
	beforeStates := len(ui.AG.States)
	beforeTransitions := len(ui.AG.Transitions)
	_, err = ui.AG.ExecuteAction(false, ui.resolveActionName(actionName), state, ui.getAlpha())
	if err != nil {
		return err
	}
	if duplicateLabel, ok := ui.newTransitionDuplicatesExistingLabel(state, beforeTransitions, existingLabels); ok {
		ui.AG.States = ui.AG.States[:beforeStates]
		ui.AG.Transitions = ui.AG.Transitions[:beforeTransitions]
		ui.sync()
		return &ExhaustedARGActionError{StateID: nodeID, ActionName: actionName, Label: duplicateLabel}
	}
	ui.sync()
	return nil
}

func (ui *AnalysisGraphUI) executeConcreteFreshAction(nodeID int, state *goivy.State, actionName string) (bool, error) {
	if ui == nil || ui.AG == nil || state == nil {
		return false, nil
	}
	domain := ui.Mod
	if domain == nil {
		domain = ui.AG.Domain
	}
	if domain == nil || domain.Actions == nil || domain.Cfg == nil {
		return false, nil
	}
	resolvedName := ui.resolveActionName(actionName)
	action, ok := domain.Actions.Get2(resolvedName)
	if !ok || action == nil {
		return false, nil
	}
	formals := action.GetFormalParams()
	if len(formals) == 0 {
		return false, nil
	}
	query, update, err := ui.concreteActionQuery(state, action, formals)
	if err != nil {
		return true, err
	}
	if nonSelf, ok := concreteActionNonSelfConstraint(update); ok {
		query = goivy.AndClausesTyped(query, goivy.FormulaToClauses(nonSelf, nil))
	}
	blockers := ui.concreteActionTupleBlockers(state, resolvedName, formals)
	if len(blockers) > 0 {
		query = goivy.AndClausesTyped(query, goivy.FormulaToClauses(&goivy.LogicAnd{Terms: blockers}, nil))
	}
	slv := goivy.NewSolver(domain, nil)
	model, err := slv.GetModelClauses(query)
	if err != nil {
		return true, err
	}
	if model == nil {
		return true, &ExhaustedARGActionError{StateID: nodeID, ActionName: actionName}
	}
	knownActuals := ui.knownConcreteActionActuals(state, resolvedName, formals)
	postClauses := concretePostClausesFromModel(slv, domain, query, model, update.Modified, formals, state.Clauses, knownActuals)
	usedDirectPostClauses := postClauses != nil
	if !usedDirectPostClauses {
		_, postClauses = goivy.ExtractPrePostModel(domain, domain.Cfg.IuCfg, query, model, update.Modified)
	}
	postState := goivy.NewState(domain, postClauses)
	postState.Pred = state
	postState.Action = action
	postState.ActionName = resolvedName
	if !usedDirectPostClauses {
		postState.Clauses = ui.stabilizeConcreteActionActuals(slv, model, postState, state, resolvedName, formals)
	}
	app := goivy.NewActionApp(action, state)
	app.ActionName = resolvedName
	ui.AG.Add(postState, app)
	ui.sync()
	return true, nil
}

func (ui *AnalysisGraphUI) concreteActionQuery(state *goivy.State, action goivy.ActionsAction, formals []*goivy.Const) (*goivy.Clauses, *goivy.Update, error) {
	if ui == nil || state == nil || action == nil {
		return nil, nil, fmt.Errorf("concrete action query: missing action context")
	}
	domain := ui.Mod
	if domain == nil && ui.AG != nil {
		domain = ui.AG.Domain
	}
	if domain == nil || domain.Cfg == nil {
		return nil, nil, fmt.Errorf("concrete action query: missing module")
	}
	ctx := &goivy.UpdateContext{
		Domain:          domain,
		PVars:           state.InScope,
		ActCfg:          domain.Cfg.ActCfg,
		Instantiator:    domain.Instantiator,
		CheckUnprovable: domain.Cfg.OnlyCheckUnprovable,
		CheckedAssert:   domain.Cfg.CheckLineno,
		GetAction: func(name string) goivy.ActionsAction {
			if domain.Actions == nil {
				return nil
			}
			if a, ok := domain.Actions.Get2(name); ok {
				return a
			}
			return nil
		},
	}
	update := goivy.BindOldsUpdate(goivy.IntUpdate(action, ctx))
	if update == nil || update.TR == nil {
		return nil, nil, fmt.Errorf("concrete action query: action update is empty")
	}
	parts := []*goivy.Clauses{state.Clauses, update.TR, domain.BackgroundTheory(state.InScope)}
	return goivy.AndClausesTyped(parts...), update, nil
}

func concretePostClausesFromModel(slv *goivy.Solver, domain *goivy.Module, query *goivy.Clauses, model *goivy.ModelResult, modified []*goivy.Const, formals []*goivy.Const, preClauses *goivy.Clauses, knownActuals map[string][]*goivy.Const) *goivy.Clauses {
	if slv == nil || domain == nil || query == nil || model == nil || model.Model == nil {
		return nil
	}
	modifiedByName := make(map[string]bool, len(modified))
	for _, sym := range modified {
		if sym != nil {
			modifiedByName[sym.Name] = true
		}
	}
	known, evalTerms := concreteKnownConstantsBySort(preClauses)
	preKnownKeys := concreteKnownConstKeySet(known)
	for sortName, actuals := range knownActuals {
		for _, actual := range actuals {
			known[sortName] = appendKnownConcreteConst(known[sortName], actual)
			evalTerms[goivy.Key(actual)] = actual
		}
	}
	nextBySort := nextConcreteActualIndexBySort(known)
	var fmlas []goivy.Expr
	var formalEqFacts []goivy.Expr
	seenFmlas := make(map[goivy.NodeKey]bool)
	addFact := func(f goivy.Expr) {
		if f == nil {
			return
		}
		key := goivy.Key(f)
		if seenFmlas[key] {
			return
		}
		seenFmlas[key] = true
		fmlas = append(fmlas, f)
	}
	for _, formal := range formals {
		if formal == nil {
			continue
		}
		sortName := goivy.IvySortName(formal.CSort)
		actual := concreteStableActualForModelValue(slv, model, formal, known[sortName], &nextBySort)
		if actual == nil {
			continue
		}
		actualKey := goivy.Key(actual)
		if _, ok := evalTerms[actualKey]; !ok {
			evalTerms[actualKey] = formal
		}
		known[sortName] = appendKnownConcreteConst(known[sortName], actual)
		eq, err := goivy.NewEq(formal, actual)
		if err == nil {
			addFact(eq)
			formalEqFacts = append(formalEqFacts, eq)
		}
	}
	pinnedQuery := query
	if len(formalEqFacts) > 0 {
		pinnedQuery = goivy.AndClausesTyped(query, goivy.NewClauses(formalEqFacts, nil, nil))
	}
	for _, rel := range concretePostRelationSymbols(domain, modified) {
		evalRel := rel
		if modifiedByName[rel.Name] {
			evalRel = goivy.NewActionConst(rel)
		}
		for _, fact := range concreteRelationFactsFromModel(slv, model, pinnedQuery, rel, evalRel, known, evalTerms, preKnownKeys) {
			addFact(fact)
		}
	}
	if len(fmlas) == 0 {
		return nil
	}
	clauses := goivy.NewClauses(fmlas, nil, nil)
	return goivy.RemoveTautEqsClauses(clauses)
}

func concreteKnownConstantsBySort(clauses *goivy.Clauses) (map[string][]*goivy.Const, map[goivy.NodeKey]goivy.Expr) {
	known := make(map[string][]*goivy.Const)
	evalTerms := make(map[goivy.NodeKey]goivy.Expr)
	if clauses == nil {
		return known, evalTerms
	}
	for _, sym := range clauses.Symbols().All() {
		c, ok := sym.(*goivy.Const)
		if !ok {
			continue
		}
		if !concreteDisplayActualConst(c) {
			continue
		}
		sortName := goivy.IvySortName(c.CSort)
		known[sortName] = appendKnownConcreteConst(known[sortName], c)
		evalTerms[goivy.Key(c)] = c
	}
	return known, evalTerms
}

func concreteKnownConstKeySet(known map[string][]*goivy.Const) map[goivy.NodeKey]bool {
	keys := make(map[goivy.NodeKey]bool)
	for _, constants := range known {
		for _, c := range constants {
			if c != nil {
				keys[goivy.Key(c)] = true
			}
		}
	}
	return keys
}

func concreteDisplayActualConst(c *goivy.Const) bool {
	if c == nil || goivy.IsFunctionSort(c.CSort) || goivy.SortEqual(c.CSort, goivy.Boolean) {
		return false
	}
	name := strings.TrimSpace(c.Name)
	if name == "" || strings.Contains(name, "@") || strings.HasPrefix(name, "fml:") {
		return false
	}
	if goivy.IsSkolem(name) || goivy.IsNew(name) || goivy.IsOld(name) {
		return false
	}
	return true
}

func appendKnownConcreteConst(constants []*goivy.Const, c *goivy.Const) []*goivy.Const {
	if c == nil {
		return constants
	}
	key := goivy.Key(c)
	for _, existing := range constants {
		if existing != nil && goivy.Key(existing) == key {
			return constants
		}
	}
	return append(constants, c)
}

func concretePostRelationSymbols(domain *goivy.Module, modified []*goivy.Const) []*goivy.Const {
	var out []*goivy.Const
	seen := make(map[string]bool)
	add := func(sym *goivy.Const) {
		if sym == nil || !goivy.IsRelationalSort(sym.CSort) {
			return
		}
		if seen[sym.Name] {
			return
		}
		seen[sym.Name] = true
		out = append(out, sym)
	}
	if domain != nil {
		for _, relExpr := range domain.AllRelations {
			if rel, ok := relExpr.(*goivy.Const); ok {
				add(rel)
			}
		}
	}
	for _, sym := range modified {
		add(sym)
	}
	return out
}

func concreteRelationFactsFromModel(slv *goivy.Solver, model *goivy.ModelResult, query *goivy.Clauses, rel, evalRel *goivy.Const, known map[string][]*goivy.Const, evalTerms map[goivy.NodeKey]goivy.Expr, preKnownKeys map[goivy.NodeKey]bool) []goivy.Expr {
	if slv == nil || model == nil || model.Model == nil || rel == nil || evalRel == nil || !goivy.IsRelationalSort(rel.CSort) {
		return nil
	}
	fs, ok := rel.CSort.(*goivy.LogicFunctionSort)
	if !ok || fs.Arity() == 0 {
		return concreteBoolRelationFactFromModel(slv, model, rel, evalRel)
	}
	domain := fs.Domain()
	tuples := concreteKnownTuplesForSorts(domain, known)
	var facts []goivy.Expr
	for _, tuple := range tuples {
		evalTuple := concreteEvalTuple(tuple, evalTerms)
		evalApp, err := goivy.NewApply(evalRel, evalTuple...)
		if err != nil {
			continue
		}
		baseApp, err := goivy.NewApply(rel, tuple...)
		if err != nil {
			continue
		}
		value, err := slv.EvalFormula(model.Model, evalApp)
		if err != nil {
			continue
		}
		if value {
			if concreteTupleHasFreshActual(tuple, preKnownKeys) && !concreteQueryForcesFormula(slv, query, evalApp) {
				neg, err := goivy.NewNot(baseApp)
				if err == nil {
					facts = append(facts, neg)
				}
				continue
			}
			facts = append(facts, baseApp)
			continue
		}
		neg, err := goivy.NewNot(baseApp)
		if err == nil {
			facts = append(facts, neg)
		}
	}
	return facts
}

func concreteTupleHasFreshActual(tuple []goivy.Expr, preKnownKeys map[goivy.NodeKey]bool) bool {
	for _, term := range tuple {
		c, ok := term.(*goivy.Const)
		if !ok || !concreteDisplayActualConst(c) {
			continue
		}
		if !preKnownKeys[goivy.Key(c)] {
			return true
		}
	}
	return false
}

func concreteQueryForcesFormula(slv *goivy.Solver, query *goivy.Clauses, fmla goivy.Expr) bool {
	if slv == nil || query == nil || fmla == nil {
		return false
	}
	neg, err := goivy.NewNot(fmla)
	if err != nil {
		return false
	}
	sat, err := slv.ClausesSat(goivy.AndClausesTyped(query, goivy.FormulaToClauses(neg, nil)))
	return err == nil && !sat
}

func concreteEvalTuple(tuple []goivy.Expr, evalTerms map[goivy.NodeKey]goivy.Expr) []goivy.Expr {
	if len(tuple) == 0 {
		return nil
	}
	out := make([]goivy.Expr, len(tuple))
	for i, term := range tuple {
		out[i] = term
		if evalTerms == nil {
			continue
		}
		if replacement := evalTerms[goivy.Key(term)]; replacement != nil {
			out[i] = replacement
		}
	}
	return out
}

func concreteBoolRelationFactFromModel(slv *goivy.Solver, model *goivy.ModelResult, rel, evalRel *goivy.Const) []goivy.Expr {
	value, err := slv.EvalFormula(model.Model, evalRel)
	if err != nil {
		return nil
	}
	if value {
		return []goivy.Expr{rel}
	}
	neg, err := goivy.NewNot(rel)
	if err != nil {
		return nil
	}
	return []goivy.Expr{neg}
}

func concreteKnownTuplesForSorts(sorts []goivy.Sort, known map[string][]*goivy.Const) [][]goivy.Expr {
	if len(sorts) == 0 {
		return [][]goivy.Expr{{}}
	}
	first := known[goivy.IvySortName(sorts[0])]
	if len(first) == 0 {
		return nil
	}
	rest := concreteKnownTuplesForSorts(sorts[1:], known)
	if len(rest) == 0 {
		return nil
	}
	var out [][]goivy.Expr
	for _, c := range first {
		if c == nil {
			continue
		}
		for _, tail := range rest {
			tuple := make([]goivy.Expr, 0, 1+len(tail))
			tuple = append(tuple, c)
			tuple = append(tuple, tail...)
			out = append(out, tuple)
		}
	}
	return out
}

func concreteActionNonSelfConstraint(update *goivy.Update) (goivy.Expr, bool) {
	if update == nil || update.ModifiedAll || len(update.Modified) == 0 {
		return nil, false
	}
	diffs := make([]goivy.Expr, 0, len(update.Modified))
	for i, sym := range update.Modified {
		if diff, ok := concreteSymbolChangedConstraint(sym, i); ok {
			diffs = append(diffs, diff)
		}
	}
	if len(diffs) == 0 {
		return nil, false
	}
	if len(diffs) == 1 {
		return diffs[0], true
	}
	disj, err := goivy.NewOr(diffs...)
	if err != nil {
		return nil, false
	}
	return disj, true
}

func concreteSymbolChangedConstraint(sym *goivy.Const, symbolIndex int) (goivy.Expr, bool) {
	if sym == nil {
		return nil, false
	}
	newSym := goivy.NewActionConst(sym)
	domain := goivy.SortDomain(sym.CSort)
	if len(domain) == 0 {
		return concreteTermsDifferConstraint(sym, newSym)
	}
	vars := make([]*goivy.LogicVariable, 0, len(domain))
	args := make([]goivy.Expr, 0, len(domain))
	for i, sort := range domain {
		v, err := goivy.NewVariable(fmt.Sprintf("__arg_diff_%d_%d", symbolIndex, i), sort)
		if err != nil {
			return nil, false
		}
		vars = append(vars, v)
		args = append(args, v)
	}
	oldApp, err := goivy.NewApply(sym, args...)
	if err != nil {
		return nil, false
	}
	newApp, err := goivy.NewApply(newSym, args...)
	if err != nil {
		return nil, false
	}
	body, ok := concreteTermsDifferConstraint(oldApp, newApp)
	if !ok {
		return nil, false
	}
	exists, err := goivy.NewExists(vars, body)
	if err != nil {
		return nil, false
	}
	return exists, true
}

func concreteTermsDifferConstraint(pre, post goivy.Expr) (goivy.Expr, bool) {
	eq, err := goivy.NewEq(pre, post)
	if err != nil {
		return nil, false
	}
	neq, err := goivy.NewNot(eq)
	if err != nil {
		return nil, false
	}
	return neq, true
}

func (ui *AnalysisGraphUI) concreteActionTupleBlockers(state *goivy.State, actionName string, formals []*goivy.Const) []goivy.Expr {
	if ui == nil || ui.AG == nil || state == nil || len(formals) == 0 {
		return nil
	}
	var blockers []goivy.Expr
	for _, tr := range ui.AG.Transitions {
		if !argTransitionStartsAtState(tr, state) || strings.TrimSpace(tr.ActionName) != actionName {
			continue
		}
		actuals := argTransitionActualsFromFormalEqualities(tr.Post, formals)
		if len(actuals) != len(formals) {
			continue
		}
		blocker, ok := concreteActionTupleBlocker(formals, actuals)
		if ok {
			blockers = append(blockers, blocker)
		}
	}
	return blockers
}

func concreteActionTupleBlocker(formals, actuals []*goivy.Const) (goivy.Expr, bool) {
	if len(formals) == 0 || len(formals) != len(actuals) {
		return nil, false
	}
	eqs := make([]goivy.Expr, 0, len(formals))
	for i := range formals {
		eq, err := goivy.NewEq(formals[i], actuals[i])
		if err != nil {
			return nil, false
		}
		eqs = append(eqs, eq)
	}
	conj, err := goivy.NewAnd(eqs...)
	if err != nil {
		return nil, false
	}
	neg, err := goivy.NewNot(conj)
	if err != nil {
		return nil, false
	}
	return neg, true
}

func (ui *AnalysisGraphUI) stabilizeConcreteActionActuals(slv *goivy.Solver, model *goivy.ModelResult, postState *goivy.State, preState *goivy.State, actionName string, formals []*goivy.Const) *goivy.Clauses {
	if slv == nil || model == nil || model.Model == nil || postState == nil || postState.Clauses == nil || len(formals) == 0 {
		if postState != nil {
			return postState.Clauses
		}
		return nil
	}
	actuals := argTransitionActualsFromFormalEqualities(postState, formals)
	if len(actuals) != len(formals) {
		return postState.Clauses
	}
	known := ui.knownConcreteActionActuals(preState, actionName, formals)
	nextBySort := nextConcreteActualIndexBySort(known)
	subs := make(map[goivy.NodeKey]goivy.Expr)
	for i, formal := range formals {
		stable := concreteStableActualForModelValue(slv, model, formal, known[goivy.IvySortName(formal.CSort)], &nextBySort)
		if stable == nil {
			continue
		}
		subs[goivy.Key(actuals[i])] = stable
		known[goivy.IvySortName(formal.CSort)] = append(known[goivy.IvySortName(formal.CSort)], stable)
	}
	if len(subs) == 0 {
		return postState.Clauses
	}
	return goivy.SubstituteConstantsClauses(postState.Clauses, subs)
}

func (ui *AnalysisGraphUI) knownConcreteActionActuals(state *goivy.State, actionName string, formals []*goivy.Const) map[string][]*goivy.Const {
	known := make(map[string][]*goivy.Const)
	if ui == nil || ui.AG == nil || state == nil {
		return known
	}
	seen := make(map[goivy.NodeKey]bool)
	for _, tr := range ui.AG.Transitions {
		if !argTransitionStartsAtState(tr, state) || strings.TrimSpace(tr.ActionName) != actionName {
			continue
		}
		actuals := argTransitionActualsFromFormalEqualities(tr.Post, formals)
		for _, actual := range actuals {
			if actual == nil {
				continue
			}
			key := goivy.Key(actual)
			if seen[key] {
				continue
			}
			seen[key] = true
			sortName := goivy.IvySortName(actual.CSort)
			known[sortName] = append(known[sortName], actual)
		}
	}
	return known
}

func nextConcreteActualIndexBySort(known map[string][]*goivy.Const) map[string]int {
	next := make(map[string]int, len(known))
	for sortName, actuals := range known {
		for _, actual := range actuals {
			if actual == nil {
				continue
			}
			idx, err := strconv.Atoi(strings.TrimSpace(actual.Name))
			if err != nil {
				continue
			}
			if idx >= next[sortName] {
				next[sortName] = idx + 1
			}
		}
	}
	return next
}

func concreteStableActualForModelValue(slv *goivy.Solver, model *goivy.ModelResult, formal *goivy.Const, known []*goivy.Const, nextBySort *map[string]int) *goivy.Const {
	formalValue := concreteModelValueString(slv, model, formal)
	if formalValue == "" {
		return nil
	}
	for _, actual := range known {
		if actual == nil || !sameARGSortName(actual.CSort, formal.CSort) {
			continue
		}
		if concreteModelValueString(slv, model, actual) == formalValue {
			return actual
		}
	}
	sortName := goivy.IvySortName(formal.CSort)
	idx := (*nextBySort)[sortName]
	(*nextBySort)[sortName] = idx + 1
	return goivy.NewConst(strconv.Itoa(idx), formal.CSort)
}

func concreteModelValueString(slv *goivy.Solver, model *goivy.ModelResult, sym *goivy.Const) string {
	if slv == nil || model == nil || model.Model == nil || sym == nil {
		return ""
	}
	values, err := slv.ModelValues(model.Model, []*goivy.Const{sym})
	if err != nil {
		return ""
	}
	if value, ok := values[sym.Name]; ok {
		return value.String()
	}
	return ""
}

func (ui *AnalysisGraphUI) transitionLabelsFromState(state *goivy.State) map[string]bool {
	labels := make(map[string]bool)
	if ui == nil || ui.AG == nil || state == nil {
		return labels
	}
	for _, tr := range ui.AG.Transitions {
		if !argTransitionStartsAtState(tr, state) {
			continue
		}
		labels[argTransitionDisplayLabel(tr)] = true
	}
	return labels
}

func (ui *AnalysisGraphUI) newTransitionDuplicatesExistingLabel(state *goivy.State, firstNewTransition int, existing map[string]bool) (string, bool) {
	if ui == nil || ui.AG == nil || state == nil || firstNewTransition >= len(ui.AG.Transitions) {
		return "", false
	}
	for _, tr := range ui.AG.Transitions[firstNewTransition:] {
		if !argTransitionStartsAtState(tr, state) {
			continue
		}
		label := argTransitionDisplayLabel(tr)
		if existing[label] {
			return label, true
		}
		existing[label] = true
	}
	return "", false
}

func argTransitionStartsAtState(tr goivy.Transition, state *goivy.State) bool {
	if tr.Pre == nil || state == nil {
		return false
	}
	return tr.Pre == state || tr.Pre.ID == state.ID
}

func (ui *AnalysisGraphUI) resolveActionName(actionName string) string {
	if ui == nil || ui.AG == nil || ui.AG.Actions == nil {
		return actionName
	}
	if _, ok := ui.AG.Actions.Get2(actionName); ok {
		return actionName
	}
	if strings.HasPrefix(actionName, "ext:") {
		trimmed := strings.TrimPrefix(actionName, "ext:")
		if _, ok := ui.AG.Actions.Get2(trimmed); ok {
			return trimmed
		}
		return actionName
	}
	extended := "ext:" + actionName
	if _, ok := ui.AG.Actions.Get2(extended); ok {
		return extended
	}
	return actionName
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

// RecalculateState re-evaluates one ARG state from its equation or join sources
// (Python: AnalysisGraphUI.recalculate_state).
func (ui *AnalysisGraphUI) RecalculateState(state *goivy.State) error {
	if ui == nil || ui.AG == nil {
		return fmt.Errorf("no analysis graph")
	}
	if state == nil {
		return fmt.Errorf("recalculate state: nil state")
	}
	ui.AG.RecalculateState(false, state, ui.getAlpha())
	ui.sync()
	return nil
}

// DecomposeEdge decomposes a transition into sub-actions
// (Python: AnalysisGraphUI.decompose_edge).
func (ui *AnalysisGraphUI) DecomposeEdge(srcID, tgtID int) (*WebUIAnalysisGraphState, error) {
	subArt, err := ui.DecomposeEdgeGraph(srcID, tgtID)
	if err != nil {
		return nil, err
	}
	return ArtToGraphState(subArt), nil
}

func (ui *AnalysisGraphUI) DecomposeEdgeGraph(srcID, tgtID int) (*goivy.AnalysisGraph, error) {
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
	return subArt, nil
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
	expr := goivy.NewActionApp(t.Op, pre)
	expr.ActionName = t.ActionName
	subArt.Add(post, expr)
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
	ui.mu.Lock()
	if ui.Mark != nil && ui.Mark.ID == nodeID {
		ui.Mark = nil
	}
	if ui.SafeNodes != nil {
		delete(ui.SafeNodes, nodeID)
	}
	ui.mu.Unlock()
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

func (ui *AnalysisGraphUI) resolveConjectureFormula(conjecture string) (goivy.Expr, goivy.Location, string, error) {
	selected := strings.TrimSpace(conjecture)
	if selected == "" {
		return nil, goivy.Location{}, "", fmt.Errorf("no conjecture specified")
	}
	if ui != nil && ui.Mod != nil {
		for _, lf := range ui.Mod.LabeledConjs {
			if lf == nil || lf.Formula == nil {
				continue
			}
			expr, ok := lf.Formula.(goivy.Expr)
			if !ok {
				continue
			}
			pretty := goivy.PrettyFmla(expr)
			if selected == pretty || selected == strings.TrimSpace(fmt.Sprint(expr)) {
				return expr, lf.GetLineno(), pretty, nil
			}
		}
	}
	fmla, parseErr := goivy.ToFormula(conjecture)
	if parseErr != nil {
		return nil, goivy.Location{}, "", fmt.Errorf("parse conjecture: %w", parseErr)
	}
	fExpr, ok := fmla.(goivy.Expr)
	if !ok {
		return nil, goivy.Location{}, "", fmt.Errorf("conjecture is not a logic expression")
	}
	return fExpr, goivy.Location{}, goivy.PrettyFmla(fExpr), nil
}

// TryConjectureResult sets up to prove a conjecture at a node and returns the
// user-visible result view (Python: AnalysisGraphUI.try_conjecture).
func (ui *AnalysisGraphUI) TryConjectureResult(nodeID int, conjecture string) (*TryConjectureResult, error) {
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return nil, err
	}
	fExpr, sourceLoc, displayConjecture, resolveErr := ui.resolveConjectureFormula(conjecture)
	if resolveErr != nil {
		return nil, resolveErr
	}
	conj := goivy.FormulaToClauses(fExpr, nil)
	dual := goivy.DualClauses(conj, nil, nil)

	mode := ui.GetMode()
	result := &TryConjectureResult{
		Mode:       mode,
		Conjecture: displayConjecture,
		SourceLoc:  sourceLoc,
	}
	if mode == ModeInduction || mode == ModeBounded {
		bmcResult := ui.AG.BMC(state, dual.ToFormula(), nil, nil)
		if bmcResult == nil {
			result.View = "message"
			result.Reachable = false
			result.Message = "The condition is unreachable along the given path."
			return result, nil
		}
		result.View = "trace"
		result.Reachable = true
		result.Trace = bmcResult
		result.Message = "The condition is reachable along the given path."
		return result, nil
	}
	w, err := ui.ViewState(nodeID, "", true)
	if err != nil {
		return nil, err
	}
	if err := ui.ensureInteractiveConceptSession(w, state); err != nil {
		return nil, err
	}
	w.G().SetFactsExpr(nil)
	if err := w.G().AddConstraintsExpr(dual.Fmlas, true); err != nil {
		return nil, err
	}
	result.View = "concept"
	result.Concept = w
	result.Message = "Conjecture goal opened."
	return result, nil
}

// TryConjecture sets up to prove a conjecture at a node
// (Python: AnalysisGraphUI.try_conjecture).
func (ui *AnalysisGraphUI) TryConjecture(nodeID int, conjecture string) error {
	_, err := ui.TryConjectureResult(nodeID, conjecture)
	return err
}

func (ui *AnalysisGraphUI) TryConjectureChoices(nodeID int) ([]ChoiceItem, error) {
	state, err := ui.stateByID(nodeID)
	if err != nil {
		return nil, err
	}
	interpState := goivy.ArtToInterpState(state)
	conjs := goivy.UndecidedConjectures(interpState)
	choices := make([]ChoiceItem, 0, len(conjs))
	for _, conj := range conjs {
		if conj == nil {
			continue
		}
		text := goivy.PrettyFmla(conj.ToFormula())
		choices = append(choices, ChoiceItem{Label: text, Value: text})
	}
	if len(choices) == 0 && ui.Mod != nil {
		for _, lf := range ui.Mod.LabeledConjs {
			if lf == nil || lf.Formula == nil {
				continue
			}
			if expr, ok := lf.Formula.(goivy.Expr); ok {
				text := goivy.PrettyFmla(expr)
				choices = append(choices, ChoiceItem{Label: text, Value: text})
			}
		}
	}
	return choices, nil
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
	if state.Clauses != nil {
		sgCopy.SetState(state.Clauses.String(), true, false, true)
	}
	w := NewGraphWidget(NewGraphStack(sgCopy))
	w.Parent = ui
	ui.CurrentConceptGraph = w
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
	fExpr, parseErr := parseBMCErrorConditionString(errCond)
	if parseErr != nil {
		return nil, parseErr
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
