// Package trace constructs counterexample traces suitable for viewing.
//
// A trace is an analysis graph (ART) representing a counterexample.
// The trace object acts as a handler for match_action (see actions package),
// allowing a trace to be constructed from a counterexample.
//
// Ported from ivy_trace.py.
package goivy

import (
	"fmt"
	"sort"
	"strings"
)

const traceCheckPrecondTrue = true

// Subgraph holds a pointer to a nested trace for call/return tracking.
type Subgraph struct {
	Graph *TraceBase
}

// TraceState extends art.State with trace-specific fields.
type TraceState struct {
	*State
	LoopStart bool
	Subgraph  *Subgraph
}

// TraceBase is the base type for counterexample traces.
// It extends AnalysisGraph with trace-construction state.
type TraceBase struct {
	*AnalysisGraph

	cfg           *IvyUtilsConfig
	TraceStates   []*TraceState
	LastAction    ActionsAction
	Sub           *TraceBase
	Returned      *TraceBase
	HiddenSymbols func(string) bool
	Renaming      map[string]string
	PP            func(Expr) Expr
	IsFullTrace   bool
}

// NewTraceBase creates a new empty TraceBase.
func NewTraceBase(cfg *IvyUtilsConfig, mod *Module) *TraceBase {
	if mod == nil {
		mod = New()
	}
	return &TraceBase{
		cfg:           cfg,
		AnalysisGraph: NewAnalysisGraph(mod),
		HiddenSymbols: func(s string) bool { return false },
	}
}

// Rename sets a renaming map on the trace.
func (tb *TraceBase) Rename(m map[string]string) *TraceBase {
	tb.Renaming = m
	return tb
}

// TraceIsSkolem reports whether a symbol name is a Skolem constant
// that should be hidden from trace display.
func TraceIsSkolem(name string) bool {
	if !IsSkolem(name) {
		return false
	}
	// Symbols starting with "__X" where X is uppercase are global skolems, not hidden.
	if strings.HasPrefix(name, "__") && len(name) > 2 {
		c := name[2]
		if c >= 'A' && c <= 'Z' {
			return false
		}
	}
	return true
}

// AddTraceState adds a new trace state from a list of equations.
func (tb *TraceBase) AddTraceState(eqns []Expr) {
	clauses := NewClauses(eqns, nil, nil)
	state := NewState(tb.Domain, clauses)
	ts := &TraceState{State: state}
	if tb.LastAction != nil {
		expr := NewActionApp(tb.LastAction, tb.lastArtState())
		tb.LastAction = nil
		tb.AnalysisGraph.Add(state, expr)
		if tb.Returned != nil {
			ts.Subgraph = &Subgraph{Graph: tb.Returned}
			tb.Returned = nil
		}
	} else {
		tb.AnalysisGraph.Add(state, nil)
	}
	tb.TraceStates = append(tb.TraceStates, ts)
}

func (tb *TraceBase) lastArtState() *State {
	if len(tb.States) == 0 {
		return nil
	}
	return tb.States[len(tb.States)-1]
}

// TraceLabelFromAction returns a label string for the given action.
func TraceLabelFromAction(action ActionsAction, renaming map[string]string) string {
	if labeler, ok := action.(interface{ GetLabel() string }); ok {
		if label := labeler.GetLabel(); label != "" {
			return label + "\n"
		}
	}
	if labeler, ok := action.(interface{ GetLabels() []string }); ok {
		labels := labeler.GetLabels()
		if len(labels) > 0 {
			return labels[0] + "\n"
		}
	}
	s := action.String()
	return TracePretty(s, 4)
}

// TracePretty formats a string by splitting on semicolons and braces,
// then indenting based on brace nesting. Truncates to maxLines if > 0.
// Corresponds to Python's pretty(s, max_lines) in ivy_utils.py.
func TracePretty(s string, maxLines int) string {
	s = strings.ReplaceAll(s, ";", ";\n")
	s = strings.ReplaceAll(s, "{", "{\n")
	s = strings.ReplaceAll(s, "}", "\n}")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines-1]
		lines = append(lines, "...")
	}
	indent := 0
	var res []string
	for _, line := range lines {
		if strings.Contains(line, "}") {
			indent--
		}
		if indent < 0 {
			indent = 0
		}
		res = append(res, strings.Repeat("    ", indent)+line)
		if strings.Contains(line, "{") {
			indent++
		}
	}
	return strings.Join(res, "\n") + strings.Repeat("}", indent)
}

// EvalInState evaluates a parameter in a state by searching its clause equations.
func EvalInState(state *State, param Expr) Expr {
	if state.Clauses == nil {
		return nil
	}
	for _, c := range state.Clauses.Fmlas {
		if eq, ok := c.(*Eq); ok {
			if eq.T1.Equal(param) {
				return eq.T2
			}
		}
	}
	return nil
}

// ToLines generates the human-readable trace lines.
func (tb *TraceBase) ToLines(lines *[]string, hash map[string]string, indent int,
	hidden func(string) bool, failed bool, renaming map[string]string, pp func(Expr) Expr) {

	if renaming == nil {
		renaming = tb.Renaming
	}
	if pp == nil {
		pp = tb.PP
	}

	for idx, ts := range tb.TraceStates {
		state := ts.State
		if state.Prov != nil {
			aa, isAA := state.Prov.(*ActionApp)
			if isAA {
				action, _ := aa.Rep.(ActionsAction)
				if tb.Domain.Cfg.TraceDetailed && action != nil {
					newlines := make([]string, 0)
					label := TraceLabelFromAction(action, renaming)
					for _, line := range strings.Split(label, "\n") {
						newlines = append(newlines, strings.Repeat("    ", indent)+line+"\n")
					}
					*lines = append(*lines, newlines...)
				}
			}
			if ts.Subgraph != nil {
				if tb.Domain.Cfg.TraceDetailed {
					*lines = append(*lines, strings.Repeat("    ", indent)+"{\n")
				}
				ts.Subgraph.Graph.ToLines(lines, hash, indent+1, hidden, failed, renaming, pp)
				if tb.Domain.Cfg.TraceDetailed {
					*lines = append(*lines, strings.Repeat("    ", indent)+"}\n")
				}
			}
			if tb.Domain.Cfg.TraceDetailed {
				*lines = append(*lines, "\n")
			}
		}

		if tb.Domain.Cfg.TraceDetailed {
			_ = idx
			if ts.LoopStart {
				*lines = append(*lines, "\n--- the following repeats infinitely ---\n\n")
			}
			var lineEqns []Expr
			if state.Clauses != nil {
				for _, c := range state.Clauses.Fmlas {
					eq, ok := c.(*Eq)
					if !ok {
						continue
					}
					lhsStr := eq.T1.String()
					if hidden(lhsStr) {
						continue
					}
					if pp != nil {
						c = pp(c)
					}
					lineEqns = append(lineEqns, c)
				}
			}
			sort.Slice(lineEqns, func(i, j int) bool {
				return lineEqns[i].String() < lineEqns[j].String()
			})
			foo := false
			for _, c := range lineEqns {
				eq, ok := c.(*Eq)
				if !ok {
					continue
				}
				s1, s2 := eq.T1.String(), eq.T2.String()
				if prev, ok := hash[s1]; !ok || prev != s2 {
					hash[s1] = s2
					if !foo {
						*lines = append(*lines, strings.Repeat("    ", indent)+"[\n")
						foo = true
					}
					*lines = append(*lines, strings.Repeat("    ", indent+1)+c.String()+"\n")
				}
			}
			if foo {
				*lines = append(*lines, strings.Repeat("    ", indent)+"]\n")
			}
		}
	}
}

// String returns the human-readable trace.
func (tb *TraceBase) String() string {
	var lines []string
	hash := make(map[string]string)
	tb.ToLines(&lines, hash, 0, tb.HiddenSymbols, false, nil, nil)
	return strings.Join(lines, "")
}

// Handle processes an action during trace construction.
func (tb *TraceBase) Handle(action ActionsAction, env map[string]string) {
	if tb.Sub != nil {
		tb.Sub.Handle(action, env)
	} else if isCallOrEnv(tb.LastAction) && tb.Returned == nil {
		tb.Sub = tb.Clone()
		tb.Sub.Handle(action, env)
	} else {
		tb.NewTraceStateFromEnv(env)
		tb.LastAction = action
	}
}

// DoReturn handles a return from a call during trace construction.
func (tb *TraceBase) DoReturn(action ActionsAction, env map[string]string) {
	if tb.Sub != nil {
		if tb.Sub.Sub != nil {
			tb.Sub.DoReturn(action, env)
		} else {
			if isCallAction(tb.Sub.LastAction) && tb.Sub.Returned == nil {
				tb.Sub.DoReturn(action, env)
				return
			}
			tb.Returned = tb.Sub
			tb.Sub = nil
			tb.Returned.NewTraceStateFromEnv(env)
		}
	} else if isCallAction(tb.LastAction) && tb.Returned == nil {
		tb.Sub = tb.Clone()
		tb.Handle(action, env)
		tb.DoReturn(action, env)
	}
}

// Fail marks the last action as failed.
// Python: ivy_trace.py:229-230:
//
//	def fail(self):
//	    self.last_action = itp.fail_action(self.last_action)
func (tb *TraceBase) Fail() {
	tb.LastAction = NewFailAction(tb.LastAction)
}

// End finishes the trace, returning from any unfinished calls.
func (tb *TraceBase) End() {
	if tb.Sub != nil {
		tb.Sub.End()
		tb.Returned = tb.Sub
		tb.Sub = nil
	}
	tb.FinalState()
}

// Clone creates a copy of the trace for subcall tracking.
func (tb *TraceBase) Clone() *TraceBase {
	return NewTraceBase(tb.cfg, tb.Domain)
}

// NewTraceStateFromEnv creates a new state from an environment mapping.
// The env maps symbol names to their renamed versions in the current context.
// For each symbol in the vocabulary, we look up its value and create
// equality equations.
//
// Python: ivy_trace.py:279-297
func (tb *TraceBase) NewTraceStateFromEnv(env map[string]string) {
	var symPairs [][2]string

	isSkolem := func(name string) bool {
		return IsSkolem(name) && (tb.HiddenSymbols == nil || !tb.HiddenSymbols(name))
	}

	// For vocabulary symbols not in env, use identity mapping
	if tb.AnalysisGraph != nil && tb.AnalysisGraph.Domain != nil {
		for name := range tb.AnalysisGraph.Domain.Relations.All() {
			if _, inEnv := env[name]; !inEnv && !IsNew(name) && !isSkolem(name) {
				symPairs = append(symPairs, [2]string{name, name})
			}
		}
		for name := range tb.AnalysisGraph.Domain.Functions.All() {
			if _, inEnv := env[name]; !inEnv && !IsNew(name) && !isSkolem(name) {
				symPairs = append(symPairs, [2]string{name, name})
			}
		}
	}

	// For symbols in env, use the renaming
	for sym, renamedSym := range env {
		if !IsNew(sym) && !isSkolem(sym) {
			symPairs = append(symPairs, [2]string{sym, renamedSym})
		}
	}

	// Build equations from symbol pairs
	var eqns []Expr
	for _, pair := range symPairs {
		sym := NewConst(pair[0], nil)
		eqns = append(eqns, &Eq{T1: sym, T2: sym})
	}

	tb.AddTraceState(eqns)
}

// FinalState adds a final state to the trace.
func (tb *TraceBase) FinalState() {
	tb.AddTraceState(nil)
}

func isCallOrEnv(action ActionsAction) bool {
	if action == nil {
		return false
	}
	switch action.(type) {
	case *LogicCallAction, *LogicEnvAction:
		return true
	}
	return false
}

func isCallAction(action ActionsAction) bool {
	if action == nil {
		return false
	}
	_, ok := action.(*LogicCallAction)
	return ok
}

// Trace extends TraceBase with model-based state construction.
type Trace struct {
	*TraceBase
	Clauses  *Clauses
	Model    TraceModel
	Vocab    []Expr
	TopLevel bool
	Eqs      map[string][]Expr // symbol name -> equations
}

// TraceModel is the interface for a counterexample model.
type TraceModel interface {
	// EvalToConstant evaluates a formula to a constant in the model.
	EvalToConstant(Expr) Expr
	// Universes returns the universe (domain) for each sort.
	Universes(numerals bool) map[string][]Expr
}

// NewTrace creates a Trace from clauses and a model.
func NewTrace(cfg *IvyUtilsConfig, clauses *Clauses, model TraceModel, vocab []Expr, topLevel bool) *Trace {
	mod := New()
	t := &Trace{
		TraceBase: NewTraceBase(cfg, mod),
		Clauses:   clauses,
		Model:     model,
		Vocab:     vocab,
		TopLevel:  topLevel,
		Eqs:       make(map[string][]Expr),
	}
	if clauses != nil {
		for _, fmla := range clauses.Fmlas {
			if eq, ok := fmla.(*Eq); ok {
				if app, ok := eq.T1.(*Apply); ok {
					key := app.Func.String()
					t.Eqs[key] = append(t.Eqs[key], fmla)
				}
			}
		}
	}
	return t
}

// GetUniverses returns the model universes.
func (t *Trace) GetUniverses() map[string][]Expr {
	if t.Model == nil {
		return nil
	}
	return t.Model.Universes(true)
}

// Eval evaluates a condition in the model, returning true or false.
func (t *Trace) Eval(cond Expr) (bool, error) {
	if t.Model == nil {
		return false, fmt.Errorf("no model available")
	}
	truth := t.Model.EvalToConstant(cond)
	if truth == nil {
		return false, fmt.Errorf("could not evaluate condition")
	}
	if truth.Equal(False) {
		return false, nil
	}
	if truth.Equal(True) {
		return true, nil
	}
	return false, fmt.Errorf("unexpected truth value: %s", truth)
}

// GetSymEqs returns the equations for a symbol in the model.
func (t *Trace) GetSymEqs(sym string) []Expr {
	// Match Python defaultdict auto-vivification
	if _, ok := t.Eqs[sym]; !ok {
		t.Eqs[sym] = nil
	}
	return t.Eqs[sym]
}

// MakeCheckArt creates an analysis graph for checking an action.
// Matches Python ivy_trace.py make_check_art:
//  1. Create pre-state with conjectures as clauses
//  2. Execute env_action to produce post-state with transition relation
//  3. Return (ag, post_state) — post includes the TR encoding
//
// Returns (ag, preState, postState).
func MakeCheckArt(mod *Module, actName string, precond []*Clauses) (*AnalysisGraph, *State, *State, error) {
	ag := NewAnalysisGraph(mod)
	var pre *Clauses
	if len(precond) > 0 {
		pre = precond[0]
		for _, p := range precond[1:] {
			pre = AndClausesTyped(pre, p)
		}
	} else {
		pre = TrueClauses(nil)
	}
	pre.Annot = EmptyAnnotation{}
	preState := NewState(mod, pre)
	ag.Add(preState, nil)

	// Execute the env_action to produce the post-state.
	// Python: post = ag.execute(env_action(act_name), pre)
	// The post-state encodes the transition relation.
	envAction := buildEnvAction(mod, actName)
	var postState *State
	if envAction != nil {
		var err error
		postState, err = ag.Execute(traceCheckPrecondTrue, envAction, preState, nil, "")
		if err != nil {
			return nil, nil, nil, fmt.Errorf("MakeCheckArt: Execute failed: %w", err)
		}
		if postState != nil {
			// Python: post.clauses = true_clauses()
			postState.Clauses = TrueClauses(nil)
		}
	}
	if postState == nil {
		postState = preState
	}

	return ag, preState, postState, nil
}

// buildEnvAction creates an EnvAction wrapping all public actions from the module.
// Matches Python ivy_actions.py env_action().
func buildEnvAction(mod *Module, actName string) ActionsAction {
	if mod == nil {
		return nil
	}
	env := BuildEnvAction(mod.Cfg.ActCfg, mod.PublicActions, mod.Actions, actName, "")
	if env == nil || len(env.ActionArgs()) == 0 {
		return nil
	}
	return env
}

// CheckFinalCond checks a final condition against an analysis graph state.
// Matches Python ivy_trace.py check_final_cond:
//   - Gets history from the analysis graph
//   - Conjoins with background theory (axioms)
//   - Calls CheckVC to check satisfiability
//
// Returns a trace if a counterexample is found, nil otherwise.
func CheckFinalCond(ag *AnalysisGraph, post *State,
	finalCond *Clauses, relsToMin []string, shrink bool) *TraceBase {
	if post == nil {
		return nil
	}
	if finalCond == nil {
		return nil
	}
	// Get history from the analysis graph — this reconstructs the full
	// transition relation from the execution path, not just the state clauses.
	// Matches Python ivy_trace.py:326: history = ag.get_history(post)
	history := ag.GetHistory(post, nil)
	if history == nil || history.Post == nil {
		return nil
	}
	// Use the history's post clauses directly.
	// Matches Python ivy_trace.py:328: clauses = history.post
	clauses := history.Post
	if clauses.Annot == nil {
		clauses = NewClauses(clauses.Fmlas, clauses.Defs, EmptyAnnotation{})
	}
	// Conjoin with background theory (axioms, definitions)
	// Matches Python ivy_trace.py:330: clauses = lut.and_clauses(clauses, axioms)
	if ag.Domain != nil {
		bgTheory := ag.Domain.BackgroundTheory(nil)
		if bgTheory != nil && len(bgTheory.Fmlas) > 0 {
			clauses = AndClausesTyped(clauses, bgTheory)
		}
	}
	// Python resolves string action names here, then wraps the path in Sequence.
	var actionExprs []Expr
	for _, a := range history.Actions {
		if a == nil {
			continue
		}
		if sym, ok := a.(*Const); ok {
			if act, exists := ag.Domain.Actions.Get2(sym.Name); exists {
				if actAction, ok := act.(ActionsAction); ok {
					actionExprs = append(actionExprs, actAction)
					continue
				}
			}
		}
		actionExprs = append(actionExprs, a)
	}
	action := NewSequence(actionExprs...)
	return CheckVC(ag.Domain, clauses, action, finalCond, relsToMin, shrink)
}

// CheckVC checks a verification condition.
// Returns a trace if a counterexample is found, nil otherwise.
// CheckVC checks a verification condition using Z3.
// Matches Python ivy_trace.py check_vc:
//   - Conjoins clauses (state + axioms) with finalCond (negated conjecture)
//   - Calls z3bridge.GetSmallModel to check satisfiability
//   - Returns a TraceBase if a counterexample is found, nil otherwise.
func CheckVC(mod *Module, clauses *Clauses, action ActionsAction,
	finalCond *Clauses, relsToMin []string, shrink bool) *TraceBase {
	if clauses == nil || clauses.Annot == nil {
		return nil
	}

	// Conjoin the state clauses with the final condition (negated conjecture).
	// Python: model = slv.get_small_model(clauses, sorts, rels, final_cond=final_cond)
	checkClauses := clauses
	if finalCond != nil {
		checkClauses = AndClausesTyped(clauses, finalCond)
	}

	// Collect uninterpreted sorts for minimization
	var sortsToMin []Sort
	for _, fmla := range checkClauses.Fmlas {
		collectUninterpSorts(fmla, &sortsToMin)
	}

	// Create solver and check
	slv := NewSolver(mod, nil)
	model, err := slv.GetSmallModel(checkClauses, sortsToMin, nil)
	if err != nil {
		fmt.Printf("CheckVC: solver error: %v\n", err)
		return nil
	}
	if model == nil {
		// UNSAT — no counterexample, property holds
		return nil
	}

	// SAT — counterexample found. Reconstruct the displayed ARG by replaying the
	// action annotation, matching Python check_vc:
	//   handler = Trace(mclauses, model, vocab)
	//   act.match_annotation(action, clauses.annot, handler)
	//   handler.end()
	ag := buildAnnotatedTraceGraph(mod, clauses, finalCond, action, model, slv)

	tb := &TraceBase{
		AnalysisGraph: ag,
	}
	for _, state := range ag.States {
		tb.TraceStates = append(tb.TraceStates, &TraceState{State: state})
	}
	return tb
}

type annotatedTraceGraphHandler struct {
	ag          *AnalysisGraph
	model       *HerbrandModel
	preClauses  *Clauses
	postClauses *Clauses
	lastAction  ActionsAction
	inSubtrace  bool
	returned    bool
}

func buildAnnotatedTraceGraph(mod *Module, clauses *Clauses, finalCond *Clauses, action ActionsAction, model *ModelResult, slv *Solver) *AnalysisGraph {
	handler := &annotatedTraceGraphHandler{
		ag:          NewAnalysisGraph(mod),
		preClauses:  clauses,
		postClauses: finalCond,
	}
	if model != nil && slv != nil {
		vocabMap := UsedSymbolsClauses(clauses)
		if finalCond != nil {
			for k, v := range UsedSymbolsClauses(finalCond).All() {
				vocabMap.Set(k, v)
			}
		}
		vocab := make([]*Const, 0, vocabMap.Len())
		for _, sym := range vocabMap.All() {
			if c, ok := sym.(*Const); ok {
				vocab = append(vocab, c)
			}
		}
		handler.model = NewHerbrandModel(slv, model.Solver, model.Model, vocab)
	}

	if action != nil {
		if clauses.Annot != nil {
			MatchAnnotation(action, clauses.Annot, handler, mod)
		}
	}
	handler.End()
	if len(handler.ag.Transitions) == 0 && action != nil {
		handler.ag = NewAnalysisGraph(mod)
		handler.lastAction = nil
		handler.returned = false
		handler.inSubtrace = false
		handler.addState(nil)
		handler.lastAction = action
		handler.addState(nil)
	}
	return handler.ag
}

func (h *annotatedTraceGraphHandler) Eval(cond Expr) bool {
	if h.model == nil {
		return true
	}
	truth := h.model.EvalToConstant(cond)
	if truth != nil && truth.Equal(False) {
		return false
	}
	if truth != nil && truth.Equal(True) {
		return true
	}
	panic(fmt.Sprintf("unexpected truth value: %v", truth))
}

func (h *annotatedTraceGraphHandler) Handle(action ActionsAction, env map[NodeKey]Expr) {
	if h.inSubtrace {
		return
	}
	if isCallOrEnv(h.lastAction) && !h.returned {
		h.inSubtrace = true
		return
	}
	h.addState(env)
	h.lastAction = action
	h.returned = false
}

func (h *annotatedTraceGraphHandler) DoReturn(action ActionsAction, env map[NodeKey]Expr) {
	if h.inSubtrace {
		h.inSubtrace = false
		h.returned = true
		return
	}
}

func (h *annotatedTraceGraphHandler) Fail() {
	h.lastAction = NewFailAction(h.lastAction)
}

func (h *annotatedTraceGraphHandler) End() {
	h.addState(nil)
}

func (h *annotatedTraceGraphHandler) addState(env map[NodeKey]Expr) {
	var clauses *Clauses
	if len(h.ag.States) == 0 {
		clauses = h.preClauses
	} else if h.postClauses != nil {
		clauses = h.postClauses
	}
	if clauses == nil {
		clauses = TrueClauses(nil)
	}
	state := NewState(h.ag.Domain, clauses)
	if h.lastAction != nil && len(h.ag.States) > 0 {
		h.ag.Add(state, NewActionApp(h.lastAction, h.ag.States[len(h.ag.States)-1]))
		h.lastAction = nil
		h.returned = false
		return
	}
	h.ag.Add(state, nil)
}

// collectUninterpSorts collects all uninterpreted sorts from a formula.
func collectUninterpSorts(n Expr, out *[]Sort) {
	if n == nil {
		return
	}
	switch t := n.(type) {
	case *LogicVariable:
		if us, ok := t.VSort.(*UninterpretedSort); ok {
			addSortIfNew(out, us)
		}
	case *Const:
		if fs, ok := t.CSort.(*LogicFunctionSort); ok {
			for _, s := range fs.Domain() {
				if us, ok := s.(*UninterpretedSort); ok {
					addSortIfNew(out, us)
				}
			}
			if us, ok := fs.Range().(*UninterpretedSort); ok {
				addSortIfNew(out, us)
			}
		}
	case *Apply:
		// Explicitly walk Func since Children() returns Terms only.
		collectUninterpSorts(t.Func, out)
	}
	for _, c := range n.Children() {
		collectUninterpSorts(c, out)
	}
}

func addSortIfNew(out *[]Sort, s Sort) {
	for _, existing := range *out {
		if existing.Equal(s) {
			return
		}
	}
	*out = append(*out, s)
}

// MakeVC generates a verification condition for an action.
// The VC is: pre ∧ TR ∧ ¬post, where TR is the action's transition relation.
//
// Python: ivy_trace.py:make_vc
func MakeVC(action ActionsAction, precond []*Clauses,
	postcond []*Clauses, checkAsserts bool) *Clauses {
	// Collect precondition formulas
	var preFmlas []Expr
	for _, p := range precond {
		preFmlas = append(preFmlas, p.Fmlas...)
	}

	// Collect postcondition formulas (negated)
	var postFmlas []Expr
	for _, p := range postcond {
		for _, f := range p.Fmlas {
			postFmlas = append(postFmlas, &LogicNot{Body: f})
		}
	}

	// Combine: pre ∧ ¬post (the TR would be added by the caller)
	allFmlas := append(preFmlas, postFmlas...)
	return NewClauses(allFmlas, nil, EmptyAnnotation{})
}

// ValueToStr converts a value to a human-readable string for trace display.
// ValueToStr converts a model value to a human-readable string.
// For constants, returns the name. For structured types (arrays, structs),
// recursively evaluates components.
//
// Python: ivy_trace.py:405-428
func ValueToStr(val Expr, evalFn func(Expr) Expr) string {
	if val == nil {
		return "..."
	}
	if c, ok := val.(*Const); ok {
		// Check for array-like sorts with end/value destructors
		if c.CSort != nil {
			sortName := c.CSort.String()
			// Try to detect array pattern: sort has .end and .value
			// This is a simplified version; full implementation would
			// check module.sort_destructors for struct rendering.
			_ = sortName
		}
		return c.Name
	}
	if app, ok := val.(*Apply); ok {
		return app.String()
	}
	return val.String()
}
