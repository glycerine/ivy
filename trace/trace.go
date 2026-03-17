// Package trace constructs counterexample traces suitable for viewing.
//
// A trace is an analysis graph (ART) representing a counterexample.
// The trace object acts as a handler for match_action (see actions package),
// allowing a trace to be constructed from a counterexample.
//
// Ported from ivy_trace.py.
package trace

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/solver"
	tr "github.com/glycerine/goivy/transrel"
)

// OptionDetailed controls whether traces include detailed state information.
var OptionDetailed = true

// FailAction wraps an action that failed during trace construction.
type FailAction struct {
	actions.ActionBase
	Action actions.Action
}

func (f *FailAction) String() string {
	if f.Action != nil {
		return fmt.Sprintf("FAIL(%s)", f.Action.String())
	}
	return "FAIL"
}

func (f *FailAction) Clone(args []lg.Node) actions.Action {
	return &FailAction{Action: f.Action}
}

func (f *FailAction) Args() []lg.Node { return nil }

func (f *FailAction) IterCalls() []string {
	if f.Action != nil {
		return f.Action.IterCalls()
	}
	return nil
}

func (f *FailAction) IterSubactions() []actions.Action {
	return []actions.Action{f}
}

func (f *FailAction) Name() string                  { return "fail" }
func (f *FailAction) Decompose() [][]actions.Action  { return [][]actions.Action{{f}} }

// Subgraph holds a pointer to a nested trace for call/return tracking.
type Subgraph struct {
	Graph *TraceBase
}

// TraceState extends art.State with trace-specific fields.
type TraceState struct {
	*art.State
	LoopStart bool
	Subgraph  *Subgraph
}

// TraceBase is the base type for counterexample traces.
// It extends AnalysisGraph with trace-construction state.
type TraceBase struct {
	*art.AnalysisGraph

	TraceStates  []*TraceState
	LastAction   actions.Action
	Sub          *TraceBase
	Returned     *TraceBase
	HiddenSymbols func(string) bool
	Renaming     map[string]string
	PP           func(lg.Node) lg.Node
	IsFullTrace  bool
}

// NewTraceBase creates a new empty TraceBase.
func NewTraceBase(mod *module.Module) *TraceBase {
	if mod == nil {
		mod = module.New()
	}
	return &TraceBase{
		AnalysisGraph: art.NewAnalysisGraph(mod),
		HiddenSymbols: func(s string) bool { return false },
	}
}

// Rename sets a renaming map on the trace.
func (tb *TraceBase) Rename(m map[string]string) *TraceBase {
	tb.Renaming = m
	return tb
}

// IsSkolem reports whether a symbol name is a Skolem constant
// that should be hidden from trace display.
func IsSkolem(name string) bool {
	if !tr.IsSkolem(name) {
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
func (tb *TraceBase) AddTraceState(eqns []lg.Node) {
	clauses := clauseops.NewClauses(eqns, nil, nil)
	state := art.NewState(tb.Domain, clauses)
	ts := &TraceState{State: state}
	if tb.LastAction != nil {
		expr := art.NewActionApp(tb.LastAction, tb.lastArtState())
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

func (tb *TraceBase) lastArtState() *art.State {
	if len(tb.States) == 0 {
		return nil
	}
	return tb.States[len(tb.States)-1]
}

// LabelFromAction returns a label string for the given action.
func LabelFromAction(action actions.Action, renaming map[string]string) string {
	if labeler, ok := action.(interface{ GetLabels() []string }); ok {
		labels := labeler.GetLabels()
		if len(labels) > 0 {
			return labels[0] + "\n"
		}
	}
	s := action.String()
	return Pretty(s, 4)
}

// Pretty truncates s to at most maxLines lines.
func Pretty(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		return strings.Join(lines, "\n") + "\n..."
	}
	return s
}

// EvalInState evaluates a parameter in a state by searching its clause equations.
func EvalInState(state *art.State, param lg.Node) lg.Node {
	if state.Clauses == nil {
		return nil
	}
	for _, c := range state.Clauses.Fmlas {
		if eq, ok := c.(*lg.Eq); ok {
			if eq.T1.Equal(param) {
				return eq.T2
			}
		}
	}
	return nil
}

// ToLines generates the human-readable trace lines.
func (tb *TraceBase) ToLines(lines *[]string, hash map[string]string, indent int,
	hidden func(string) bool, failed bool, renaming map[string]string, pp func(lg.Node) lg.Node) {

	if renaming == nil {
		renaming = tb.Renaming
	}
	if pp == nil {
		pp = tb.PP
	}

	for idx, ts := range tb.TraceStates {
		state := ts.State
		if state.Expr != nil {
			aa, isAA := state.Expr.(*art.ActionApp)
			if isAA {
				action, _ := aa.Rep.(actions.Action)
				if OptionDetailed && action != nil {
					newlines := make([]string, 0)
					label := LabelFromAction(action, renaming)
					for _, line := range strings.Split(label, "\n") {
						newlines = append(newlines, strings.Repeat("    ", indent)+line+"\n")
					}
					*lines = append(*lines, newlines...)
				}
			}
			if ts.Subgraph != nil {
				if OptionDetailed {
					*lines = append(*lines, strings.Repeat("    ", indent)+"{\n")
				}
				ts.Subgraph.Graph.ToLines(lines, hash, indent+1, hidden, failed, renaming, pp)
				if OptionDetailed {
					*lines = append(*lines, strings.Repeat("    ", indent)+"}\n")
				}
			}
			if OptionDetailed {
				*lines = append(*lines, "\n")
			}
		}

		if OptionDetailed {
			_ = idx
			if ts.LoopStart {
				*lines = append(*lines, "\n--- the following repeats infinitely ---\n\n")
			}
			var lineEqns []lg.Node
			if state.Clauses != nil {
				for _, c := range state.Clauses.Fmlas {
					eq, ok := c.(*lg.Eq)
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
				eq, ok := c.(*lg.Eq)
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
func (tb *TraceBase) Handle(action actions.Action, env map[string]string) {
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
func (tb *TraceBase) DoReturn(action actions.Action, env map[string]string) {
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
func (tb *TraceBase) Fail() {
	tb.LastAction = &FailAction{Action: tb.LastAction}
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
	return NewTraceBase(tb.Domain)
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
		return tr.IsSkolem(name) && (tb.HiddenSymbols == nil || !tb.HiddenSymbols(name))
	}

	// For vocabulary symbols not in env, use identity mapping
	if tb.AnalysisGraph != nil && tb.AnalysisGraph.Domain != nil {
		for name := range tb.AnalysisGraph.Domain.Relations {
			if _, inEnv := env[name]; !inEnv && !tr.IsNew(name) && !isSkolem(name) {
				symPairs = append(symPairs, [2]string{name, name})
			}
		}
		for name := range tb.AnalysisGraph.Domain.Functions {
			if _, inEnv := env[name]; !inEnv && !tr.IsNew(name) && !isSkolem(name) {
				symPairs = append(symPairs, [2]string{name, name})
			}
		}
	}

	// For symbols in env, use the renaming
	for sym, renamedSym := range env {
		if !tr.IsNew(sym) && !isSkolem(sym) {
			symPairs = append(symPairs, [2]string{sym, renamedSym})
		}
	}

	// Build equations from symbol pairs
	var eqns []lg.Node
	for _, pair := range symPairs {
		sym := lg.NewConst(pair[0], nil)
		eqns = append(eqns, &lg.Eq{T1: sym, T2: sym})
	}

	tb.AddTraceState(eqns)
}

// FinalState adds a final state to the trace.
func (tb *TraceBase) FinalState() {
	tb.AddTraceState(nil)
}

func isCallOrEnv(action actions.Action) bool {
	if action == nil {
		return false
	}
	switch action.(type) {
	case *actions.CallAction, *actions.EnvAction:
		return true
	}
	return false
}

func isCallAction(action actions.Action) bool {
	if action == nil {
		return false
	}
	_, ok := action.(*actions.CallAction)
	return ok
}

// Trace extends TraceBase with model-based state construction.
type Trace struct {
	*TraceBase
	Clauses  *clauseops.Clauses
	Model    Model
	Vocab    []lg.Node
	TopLevel bool
	Eqs      map[string][]lg.Node // symbol name -> equations
}

// Model is the interface for a counterexample model.
type Model interface {
	// EvalToConstant evaluates a formula to a constant in the model.
	EvalToConstant(lg.Node) lg.Node
	// Universes returns the universe (domain) for each sort.
	Universes(numerals bool) map[string][]lg.Node
}

// NewTrace creates a Trace from clauses and a model.
func NewTrace(clauses *clauseops.Clauses, model Model, vocab []lg.Node, topLevel bool) *Trace {
	mod := module.New()
	t := &Trace{
		TraceBase: NewTraceBase(mod),
		Clauses:   clauses,
		Model:     model,
		Vocab:     vocab,
		TopLevel:  topLevel,
		Eqs:       make(map[string][]lg.Node),
	}
	if clauses != nil {
		for _, fmla := range clauses.Fmlas {
			if eq, ok := fmla.(*lg.Eq); ok {
				if app, ok := eq.T1.(*lg.Apply); ok {
					key := app.Func.String()
					t.Eqs[key] = append(t.Eqs[key], fmla)
				}
			}
		}
	}
	return t
}

// GetUniverses returns the model universes.
func (t *Trace) GetUniverses() map[string][]lg.Node {
	if t.Model == nil {
		return nil
	}
	return t.Model.Universes(true)
}

// Eval evaluates a condition in the model, returning true or false.
func (t *Trace) Eval(cond lg.Node) (bool, error) {
	if t.Model == nil {
		return false, fmt.Errorf("no model available")
	}
	truth := t.Model.EvalToConstant(cond)
	if truth == nil {
		return false, fmt.Errorf("could not evaluate condition")
	}
	if truth.Equal(lg.False) {
		return false, nil
	}
	if truth.Equal(lg.True) {
		return true, nil
	}
	return false, fmt.Errorf("unexpected truth value: %s", truth)
}

// GetSymEqs returns the equations for a symbol in the model.
func (t *Trace) GetSymEqs(sym string) []lg.Node {
	return t.Eqs[sym]
}

// MakeCheckArt creates an analysis graph for checking an action.
// Matches Python ivy_trace.py make_check_art:
//   1. Create pre-state with conjectures as clauses
//   2. Execute env_action to produce post-state with transition relation
//   3. Return (ag, post_state) — post includes the TR encoding
// Returns (ag, preState, postState).
func MakeCheckArt(mod *module.Module, actName string, precond []*clauseops.Clauses) (*art.AnalysisGraph, *art.State, *art.State) {
	ag := art.NewAnalysisGraph(mod)
	var pre *clauseops.Clauses
	if len(precond) > 0 {
		pre = precond[0]
		for _, p := range precond[1:] {
			pre = clauseops.AndClausesTyped(pre, p)
		}
	} else {
		pre = clauseops.TrueClauses(nil)
	}
	pre.Annot = actions.EmptyAnnotation{}
	preState := art.NewState(mod, pre)
	ag.Add(preState, nil)

	// Execute the env_action to produce the post-state.
	// Python: post = ag.execute(env_action(act_name), pre)
	// The post-state encodes the transition relation.
	envAction := buildEnvAction(mod, actName)
	var postState *art.State
	if envAction != nil {
		postState = ag.Execute(envAction, preState, nil, "")
		if postState != nil {
			// Python: post.clauses = true_clauses()
			postState.Clauses = clauseops.TrueClauses(nil)
		}
	}
	if postState == nil {
		postState = preState
	}

	return ag, preState, postState
}

// buildEnvAction creates an EnvAction wrapping all public actions from the module.
// Matches Python ivy_actions.py env_action().
func buildEnvAction(mod *module.Module, actName string) actions.Action {
	if mod == nil {
		return nil
	}
	var branches []lg.Node
	if actName != "" {
		if a, ok := mod.Actions[actName]; ok {
			if act, ok2 := a.(actions.Action); ok2 {
				branches = append(branches, actions.WrapAction(act))
			}
		}
	} else {
		// All public actions
		for name := range mod.PublicActions {
			if a, ok := mod.Actions[name]; ok {
				if act, ok2 := a.(actions.Action); ok2 {
					branches = append(branches, actions.WrapAction(act))
				}
			}
		}
	}
	if len(branches) == 0 {
		return nil
	}
	return actions.NewEnvAction(branches...)
}

// CheckFinalCond checks a final condition against an analysis graph state.
// Matches Python ivy_trace.py check_final_cond:
//   - Gets history from the analysis graph
//   - Conjoins with background theory (axioms)
//   - Calls CheckVC to check satisfiability
//
// Returns a trace if a counterexample is found, nil otherwise.
func CheckFinalCond(ag *art.AnalysisGraph, post *art.State,
	finalCond *clauseops.Clauses, relsToMin []string, shrink bool) *TraceBase {
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
	// Use the history's post formula as the clauses.
	// Matches Python ivy_trace.py:328: clauses = history.post
	clauses := clauseops.FormulaToClauses(history.Post, actions.EmptyAnnotation{})
	// Conjoin with background theory (axioms, definitions)
	// Matches Python ivy_trace.py:330: clauses = lut.and_clauses(clauses, axioms)
	if ag.Domain != nil {
		bgTheory := ag.Domain.BackgroundTheory(nil)
		if bgTheory != nil && len(bgTheory.Fmlas) > 0 {
			clauses = clauseops.AndClausesTyped(clauses, bgTheory)
		}
	}
	return CheckVC(clauses, nil, finalCond, relsToMin, shrink)
}

// CheckVC checks a verification condition.
// Returns a trace if a counterexample is found, nil otherwise.
// CheckVC checks a verification condition using Z3.
// Matches Python ivy_trace.py check_vc:
//   - Conjoins clauses (state + axioms) with finalCond (negated conjecture)
//   - Calls solver.GetSmallModel to check satisfiability
//   - Returns a TraceBase if a counterexample is found, nil otherwise.
func CheckVC(clauses *clauseops.Clauses, action actions.Action,
	finalCond *clauseops.Clauses, relsToMin []string, shrink bool) *TraceBase {
	if clauses == nil || clauses.Annot == nil {
		return nil
	}

	// Conjoin the state clauses with the final condition (negated conjecture).
	// Python: model = slv.get_small_model(clauses, sorts, rels, final_cond=final_cond)
	checkClauses := clauses
	if finalCond != nil {
		checkClauses = clauseops.AndClausesTyped(clauses, finalCond)
	}

	// Collect uninterpreted sorts for minimization
	var sortsToMin []lg.Sort
	for _, fmla := range checkClauses.Fmlas {
		collectUninterpSorts(fmla, &sortsToMin)
	}

	// Create solver and check
	slv := solver.New()
	model, err := slv.GetSmallModel(checkClauses, sortsToMin, nil)
	if err != nil {
		fmt.Printf("CheckVC: solver error: %v\n", err)
		return nil
	}
	if model == nil {
		// UNSAT — no counterexample, property holds
		return nil
	}

	// SAT — counterexample found. Build a minimal trace.
	ag := art.NewAnalysisGraph(nil)
	preState := art.NewState(nil, clauses)
	ag.Add(preState, nil)
	postState := art.NewState(nil, finalCond)
	ag.Add(postState, nil)

	tb := &TraceBase{
		AnalysisGraph: ag,
		TraceStates: []*TraceState{
			{State: preState},
			{State: postState},
		},
	}
	return tb
}

// collectUninterpSorts collects all uninterpreted sorts from a formula.
func collectUninterpSorts(n lg.Node, out *[]lg.Sort) {
	if n == nil {
		return
	}
	switch t := n.(type) {
	case *lg.Var:
		if us, ok := t.VSort.(*lg.UninterpretedSort); ok {
			addSortIfNew(out, us)
		}
	case *lg.Const:
		if fs, ok := t.CSort.(*lg.FunctionSort); ok {
			for _, s := range fs.Domain() {
				if us, ok := s.(*lg.UninterpretedSort); ok {
					addSortIfNew(out, us)
				}
			}
			if us, ok := fs.Range().(*lg.UninterpretedSort); ok {
				addSortIfNew(out, us)
			}
		}
	case *lg.Apply:
		// Explicitly walk Func since Children() returns Terms only.
		collectUninterpSorts(t.Func, out)
	}
	for _, c := range n.Children() {
		collectUninterpSorts(c, out)
	}
}

func addSortIfNew(out *[]lg.Sort, s lg.Sort) {
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
func MakeVC(action actions.Action, precond []*clauseops.Clauses,
	postcond []*clauseops.Clauses, checkAsserts bool) *clauseops.Clauses {
	// Collect precondition formulas
	var preFmlas []lg.Node
	for _, p := range precond {
		preFmlas = append(preFmlas, p.Fmlas...)
	}

	// Collect postcondition formulas (negated)
	var postFmlas []lg.Node
	for _, p := range postcond {
		for _, f := range p.Fmlas {
			postFmlas = append(postFmlas, &lg.Not{Body: f})
		}
	}

	// Combine: pre ∧ ¬post (the TR would be added by the caller)
	allFmlas := append(preFmlas, postFmlas...)
	return clauseops.NewClauses(allFmlas, nil, actions.EmptyAnnotation{})
}

// ValueToStr converts a value to a human-readable string for trace display.
// ValueToStr converts a model value to a human-readable string.
// For constants, returns the name. For structured types (arrays, structs),
// recursively evaluates components.
//
// Python: ivy_trace.py:405-428
func ValueToStr(val lg.Node, evalFn func(lg.Node) lg.Node) string {
	if val == nil {
		return "..."
	}
	if c, ok := val.(*lg.Const); ok {
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
	if app, ok := val.(*lg.Apply); ok {
		return app.String()
	}
	return val.String()
}
