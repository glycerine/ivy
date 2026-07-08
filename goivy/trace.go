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
	"strconv"
	"strings"
)

const traceCheckPrecondTrue = true
const traceCheckPrecondFalse = false

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
		tb.registerReturnedTraceActionNames(tb.LastAction, tb.Returned)
		expr := NewActionApp(tb.LastAction, tb.lastArtState())
		tb.LastAction = nil
		tb.AnalysisGraph.addActionState(state, expr)
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
	if label, ok := traceActionLabel(action); ok {
		return label + "\n"
	}
	rendered := action
	if len(renaming) > 0 {
		if _, failed := action.(*FailAction); !failed {
			if renamed, ok := RenameASTByName(action, renaming).(ActionsAction); ok {
				if reduced, ok := ReduceNamedBinders(renamed, nil).(ActionsAction); ok {
					rendered = reduced
				} else {
					rendered = renamed
				}
			}
		}
	}
	s := rendered.String()
	return TracePretty(s, 4)
}

func traceActionLabel(action ActionsAction) (string, bool) {
	if labeler, ok := action.(interface{ GetLabel() string }); ok {
		if label := labeler.GetLabel(); label != "" {
			return label, true
		}
	}
	return "", false
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

func traceStateEqMap(state *State) map[NodeKey]Expr {
	res := make(map[NodeKey]Expr)
	if state == nil || state.Clauses == nil {
		return res
	}
	for _, c := range state.Clauses.Fmlas {
		if eq, ok := c.(*Eq); ok {
			res[Key(eq.T1)] = eq.T2
		}
	}
	return res
}

func traceEvalInStateMap(stateHash map[NodeKey]Expr, param Expr) Expr {
	if param == nil {
		return nil
	}
	if app, ok := param.(*Apply); ok {
		args := make([]Expr, len(app.Terms))
		for i, arg := range app.Terms {
			if val := traceEvalInStateMap(stateHash, arg); val != nil {
				args[i] = val
			} else {
				args[i] = arg
			}
		}
		param = CloneApplyTerms(app, args)
	}
	return stateHash[Key(param)]
}

func traceActionLinenoPrefix(action ActionsAction) string {
	if action != nil && action.HasLineno() {
		return action.GetLineno().String()
	}
	return ""
}

func traceHideDetailedAction(action ActionsAction) bool {
	if action == nil || action.HasLineno() {
		return false
	}
	assign, ok := action.(*LogicAssignAction)
	if !ok {
		return false
	}
	lhs, ok := assign.LHS.(*Const)
	if !ok {
		return false
	}
	switch lhs.Name {
	case "err_flag":
		return IsFalse(assign.RHS)
	case "__init":
		return IsTrue(assign.RHS)
	default:
		return false
	}
}

func traceIsAssertAction(action ActionsAction) bool {
	switch action.(type) {
	case *LogicAssertAction, *LogicRequiresAction, *LogicEnsuresAction:
		return true
	default:
		return false
	}
}

func traceExprName(expr Expr) string {
	switch t := expr.(type) {
	case *Const:
		return t.Name
	case *Apply:
		return traceExprName(t.Func)
	default:
		return fmt.Sprint(expr)
	}
}

func traceImportName(n Node) string {
	switch t := n.(type) {
	case *ImportDef:
		return fmt.Sprint(t.Imported)
	default:
		return fmt.Sprint(n)
	}
}

func traceImportedActionName(mod *Module, name string) string {
	if mod == nil {
		return ""
	}
	for _, imp := range mod.Imports {
		if traceImportName(imp) == name {
			return name
		}
	}
	return ""
}

func traceFirstSubgraphState(ts *TraceState, fallback *State) *State {
	if ts != nil && ts.Subgraph != nil && ts.Subgraph.Graph != nil && len(ts.Subgraph.Graph.TraceStates) > 0 {
		return ts.Subgraph.Graph.TraceStates[0].State
	}
	return fallback
}

func traceQuoteEvent(event string) string {
	if !strings.HasPrefix(event, "\"") {
		return "\"" + event + "\""
	}
	return event
}

func traceDebugRep(expr Expr) string {
	switch t := expr.(type) {
	case *Const:
		return t.Name
	case *Apply:
		return traceExprName(t)
	default:
		return fmt.Sprint(expr)
	}
}

func tracePrintDebugExpr(mod *Module, expr Expr, stateHash map[NodeKey]Expr) string {
	subst := make(map[string]Expr)
	for _, v := range VariablesAST(expr) {
		for _, prefix := range []string{"@", "__"} {
			sym := NewConst(prefix+v.Name, v.VSort)
			if val, ok := stateHash[Key(sym)]; ok {
				subst[v.Name] = val
				break
			}
		}
	}
	if len(subst) > 0 {
		expr = SubstituteByName(expr, subst)
	}
	return ValueToStrWithModule(mod, traceEvalInStateMap(stateHash, expr), func(param Expr) Expr {
		return traceEvalInStateMap(stateHash, param)
	})
}

func (tb *TraceBase) appendNonDetailedAction(lines *[]string, idx int, ts *TraceState, state *State, action ActionsAction, failed bool) bool {
	if action == nil {
		return failed
	}
	current := action
	if fail, ok := action.(*FailAction); ok {
		if !isCallOrEnv(fail.Inner) {
			*lines = append(*lines, traceActionLinenoPrefix(fail)+"error: assertion failed\n")
		} else {
			current = fail.Inner
			failed = true
		}
	} else if failed && idx == len(tb.TraceStates)-1 && traceIsAssertAction(action) {
		*lines = append(*lines, traceActionLinenoPrefix(action)+"error: assertion failed\n")
	}

	switch a := current.(type) {
	case *LogicCallAction:
		calleeName := traceExprName(a.Callee)
		calleeName = traceImportedActionName(tb.Domain, calleeName)
		if calleeName != "" {
			var params []*Const
			if callee, ok := tb.Domain.Actions.Get2(calleeName); ok && callee != nil {
				params = callee.GetFormalParams()
			}
			tb.appendCallEnvSummary(lines, "< ", calleeName, params, traceFirstSubgraphState(ts, state))
		}
	case *LogicEnvAction:
		if len(a.Branches) == 0 {
			break
		}
		branchAction, _ := a.Branches[0].(ActionsAction)
		if branchAction == nil {
			break
		}
		labeler, ok := branchAction.(interface{ GetLabel() string })
		if !ok || labeler.GetLabel() == "" {
			break
		}
		tb.appendCallEnvSummary(lines, "> ", labeler.GetLabel(), branchAction.GetFormalParams(), traceFirstSubgraphState(ts, state))
	case *LogicDebugAction:
		stateHash := traceStateEqMap(state)
		*lines = append(*lines, "{\n")
		*lines = append(*lines, fmt.Sprintf("\"event\" : %s\n", traceQuoteEvent(traceDebugRep(a.DebugExpr))))
		for _, eqn := range a.WithExprs {
			if eq, ok := eqn.(*Eq); ok {
				*lines = append(*lines, fmt.Sprintf("%s : %s,\n",
					traceQuoteEvent(traceDebugRep(eq.T1)),
					tracePrintDebugExpr(tb.Domain, eq.T2, stateHash)))
			}
		}
		*lines = append(*lines, "}\n")
	}
	return failed
}

func (tb *TraceBase) appendCallEnvSummary(lines *[]string, arrow, name string, params []*Const, state *State) {
	if name == "" {
		return
	}
	out := arrow + name
	if len(params) > 0 {
		stateHash := traceStateEqMap(state)
		vals := make([]string, len(params))
		for i, param := range params {
			vals[i] = ValueToStrWithModule(tb.Domain, traceEvalInStateMap(stateHash, param), func(expr Expr) Expr {
				return traceEvalInStateMap(stateHash, expr)
			})
		}
		out += "(" + strings.Join(vals, ",") + ")"
	}
	*lines = append(*lines, out+"\n")
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
			var action ActionsAction
			hideAction := false
			if isAA {
				action, _ = aa.Rep.(ActionsAction)
				hideAction = traceHideDetailedAction(action)
				if tb.Domain.Cfg.TraceDetailed && action != nil && !hideAction {
					if _, labeled := traceActionLabel(action); !labeled && action.HasLineno() {
						*lines = append(*lines, action.GetLineno().String()+"\n")
					}
					newlines := make([]string, 0)
					label := TraceLabelFromAction(action, renaming)
					for _, line := range strings.Split(label, "\n") {
						newlines = append(newlines, strings.Repeat("    ", indent)+line+"\n")
					}
					*lines = append(*lines, newlines...)
				} else if action != nil && !hideAction {
					failed = tb.appendNonDetailedAction(lines, idx, ts, state, action, failed)
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
			if tb.Domain.Cfg.TraceDetailed && !hideAction {
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
					lhsStr := traceLHSRepName(eq.T1)
					if hidden(lhsStr) {
						continue
					}
					if renaming != nil {
						c = RenameASTByName(c, renaming)
					}
					c = ReduceNamedBinders(c, nil)
					if pp != nil {
						c = pp(c)
					}
					eq, ok = c.(*Eq)
					if !ok {
						continue
					}
					if ReduceNumerically(eq.T1).Equal(eq.T2) {
						continue
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

func traceLHSRepName(lhs Expr) string {
	switch t := lhs.(type) {
	case *Apply:
		return t.Func.String()
	case *Const:
		return t.Name
	default:
		return lhs.String()
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
	Clauses       *Clauses
	Model         TraceModel
	Vocab         []Expr
	TopLevel      bool
	Eqs           map[NodeKey][]Expr // symbol rep -> equations
	SubTrace      *Trace
	ReturnedTrace *Trace
	IsCti         *Clauses
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
	return NewTraceForModule(cfg, New(), clauses, model, vocab, topLevel)
}

// NewTraceForModule creates a Trace using the given module as its analysis
// graph domain, matching Python TraceBase's use of the current ivy module.
func NewTraceForModule(cfg *IvyUtilsConfig, mod *Module, clauses *Clauses, model TraceModel, vocab []Expr, topLevel bool) *Trace {
	if mod == nil {
		mod = New()
	}
	t := &Trace{
		TraceBase: NewTraceBase(cfg, mod),
		Clauses:   clauses,
		Model:     model,
		Vocab:     vocab,
		TopLevel:  topLevel,
		Eqs:       make(map[NodeKey][]Expr),
	}
	t.SetEqsFromClauses(clauses)
	return t
}

// SetEqsFromClauses indexes model equations by their application symbol,
// matching Python Trace.__init__'s self.eqs defaultdict.
func (t *Trace) SetEqsFromClauses(clauses *Clauses) {
	t.Eqs = make(map[NodeKey][]Expr)
	if clauses == nil {
		return
	}
	for _, fmla := range clauses.Fmlas {
		switch f := fmla.(type) {
		case *Eq:
			if rep := traceEqRep(f.T1); rep != nil {
				t.Eqs[Key(rep)] = append(t.Eqs[Key(rep)], fmla)
			}
		case *LogicNot:
			if rep := traceEqRep(f.Body); rep != nil {
				t.Eqs[Key(rep)] = append(t.Eqs[Key(rep)], &Eq{T1: f.Body, T2: &LogicOr{}})
			}
		default:
			if rep := traceEqRep(fmla); rep != nil {
				t.Eqs[Key(rep)] = append(t.Eqs[Key(rep)], &Eq{T1: fmla, T2: &LogicAnd{}})
			}
		}
	}
}

// GetUniverses returns the model universes.
func (t *Trace) GetUniverses() map[string][]Expr {
	if t.Model == nil {
		return nil
	}
	return t.Model.Universes(true)
}

// Eval evaluates a condition in the model, returning true or false.
func (t *Trace) Eval(cond Expr) bool {
	truth := cond
	if t.Model != nil {
		truth = t.Model.EvalToConstant(cond)
	}
	if IsFalse(truth) {
		return false
	}
	if IsTrue(truth) {
		return true
	}
	panic(fmt.Sprintf("unexpected truth value: %v", truth))
}

// GetSymEqs returns the equations for a symbol in the model.
func (t *Trace) GetSymEqs(sym Expr) []Expr {
	// Match Python defaultdict auto-vivification
	key := Key(sym)
	if _, ok := t.Eqs[key]; !ok {
		t.Eqs[key] = nil
	}
	return t.Eqs[key]
}

// Clone creates a model-backed subtrace, matching Python Trace.clone.
func (t *Trace) Clone() *Trace {
	clone := NewTraceForModule(t.cfg, t.Domain, t.Clauses, t.Model, t.Vocab, false)
	clone.Eqs = make(map[NodeKey][]Expr, len(t.Eqs))
	for key, eqs := range t.Eqs {
		clone.Eqs[key] = append([]Expr(nil), eqs...)
	}
	return clone
}

// Handle processes an action during annotation-guided trace construction.
func (t *Trace) Handle(action ActionsAction, env map[NodeKey]Expr) {
	if t.SubTrace != nil {
		t.SubTrace.Handle(action, env)
	} else if isCallOrEnv(t.LastAction) && t.ReturnedTrace == nil {
		t.SubTrace = t.Clone()
		t.SubTrace.Handle(action, env)
	} else {
		if traceIsNowhereAction(action) {
			return
		}
		t.NewState(env)
		t.LastAction = action
	}
}

// DoReturn handles a return from a call during annotation-guided trace construction.
func (t *Trace) DoReturn(action ActionsAction, env map[NodeKey]Expr) {
	if t.SubTrace != nil {
		if t.SubTrace.SubTrace != nil {
			t.SubTrace.DoReturn(action, env)
		} else {
			if isCallAction(t.SubTrace.LastAction) && t.SubTrace.ReturnedTrace == nil {
				t.SubTrace.DoReturn(action, env)
				return
			}
			t.ReturnedTrace = t.SubTrace
			t.SubTrace = nil
			t.ReturnedTrace.NewState(env)
		}
	} else if isCallAction(t.LastAction) && t.ReturnedTrace == nil {
		t.SubTrace = t.Clone()
		t.Handle(action, env)
		t.DoReturn(action, env)
	}
}

// End finishes the trace, returning from any unfinished calls.
func (t *Trace) End() {
	if t.SubTrace != nil {
		t.SubTrace.End()
		t.ReturnedTrace = t.SubTrace
		t.SubTrace = nil
	}
	t.FinalState()
}

// NewState builds a displayed state from model equations and the annotation env.
func (t *Trace) NewState(env map[NodeKey]Expr) {
	var symPairs [][2]Expr
	for _, sym := range t.Vocab {
		if sym == nil {
			continue
		}
		if _, inEnv := env[Key(sym)]; !inEnv && t.ShouldTrackStateSymbol(sym) {
			symPairs = append(symPairs, [2]Expr{sym, sym})
		}
	}

	vocabByKey := make(map[NodeKey]Expr, len(t.Vocab))
	for _, sym := range t.Vocab {
		if sym != nil {
			vocabByKey[Key(sym)] = sym
		}
	}
	for symKey, renamedSym := range env {
		sym := vocabByKey[symKey]
		if sym == nil {
			continue
		}
		if t.ShouldTrackStateSymbol(sym) {
			symPairs = append(symPairs, [2]Expr{sym, renamedSym})
		}
	}
	t.NewStatePairs(symPairs, env)
}

// NewStatePairs materializes state equations for symbol/renamed-symbol pairs.
func (t *Trace) NewStatePairs(symPairs [][2]Expr, env map[NodeKey]Expr) {
	var eqns []Expr
	for _, pair := range symPairs {
		sym, renamedSym := pair[0], pair[1]
		symConst, symOK := sym.(*Const)
		renamedConst, renamedOK := renamedSym.(*Const)
		for _, fmla := range t.GetSymEqs(renamedSym) {
			rfmla := fmla
			if symOK && renamedOK {
				rfmla = RenameAST(fmla, map[NodeKey]*Const{Key(renamedConst): symConst})
			}
			eqns = append(eqns, rfmla)
		}
	}
	t.AddState(eqns)
}

// FinalState adds the final model-backed state.
func (t *Trace) FinalState() {
	var symPairs [][2]Expr
	for _, sym := range t.Vocab {
		if sym == nil {
			continue
		}
		if t.ShouldTrackStateSymbol(sym) {
			symPairs = append(symPairs, [2]Expr{sym, sym})
		}
	}
	t.NewStatePairs(symPairs, nil)
}

// AddState adds a trace state, including model universes and call subgraphs.
func (t *Trace) AddState(eqns []Expr) {
	clauses := NewClauses(eqns, nil, nil)
	state := NewState(t.Domain, clauses)
	if univs := t.GetUniverses(); univs != nil {
		state.Universe = univs
	}
	ts := &TraceState{State: state}
	if t.LastAction != nil {
		var returned *TraceBase
		if t.ReturnedTrace != nil {
			returned = t.ReturnedTrace.TraceBase
		}
		t.registerReturnedTraceActionNames(t.LastAction, returned)
		expr := NewActionApp(t.LastAction, t.lastArtState())
		t.LastAction = nil
		t.AnalysisGraph.addActionState(state, expr)
		if t.ReturnedTrace != nil {
			ts.Subgraph = &Subgraph{Graph: t.ReturnedTrace.TraceBase}
			t.ReturnedTrace = nil
		}
	} else {
		t.AnalysisGraph.Add(state, nil)
	}
	t.TraceStates = append(t.TraceStates, ts)
}

func (tb *TraceBase) registerReturnedTraceActionNames(action ActionsAction, returned *TraceBase) {
	if tb == nil || tb.AnalysisGraph == nil || action == nil || returned == nil || returned.AnalysisGraph == nil {
		return
	}
	returned.AnalysisGraph.CanonicalizeTransitionActionNames()
	seen := make(map[string]bool)
	var names []string
	for _, tr := range returned.AnalysisGraph.Transitions {
		if tr.IsJoin() {
			continue
		}
		name := strings.TrimSpace(tr.ActionName)
		if !tb.AnalysisGraph.validTransitionActionName(name) || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	tb.AnalysisGraph.RegisterActionNames(action, names)
}

// IsSkolem reports whether a trace symbol should be treated as a Skolem.
func (t *Trace) IsSkolem(sym Expr) bool {
	c, ok := sym.(*Const)
	if !ok {
		return false
	}
	return TraceIsSkolem(c.Name)
}

func (t *Trace) ShouldTrackStateSymbol(sym Expr) bool {
	return !traceIsNew(sym) && !t.IsSkolem(sym) && !IsNumeral(sym)
}

func traceIsNew(sym Expr) bool {
	c, ok := sym.(*Const)
	return ok && IsNew(c.Name)
}

func traceIsNowhereAction(action ActionsAction) bool {
	return action != nil && action.HasLineno() && action.GetLineno().Filename == "nowhere"
}

func traceEqRep(lhs Expr) Expr {
	switch t := lhs.(type) {
	case *Apply:
		return t.Func
	case *Const:
		return t
	default:
		return nil
	}
}

func traceModelIgnore(slv *Solver) func(*Const) bool {
	return func(sym *Const) bool {
		if slv == nil || sym == nil {
			return false
		}
		return slv.SolverName(sym) == ""
	}
}

func traceRelationsToMinimize(mod *Module, clauses *Clauses, names []string) []*Const {
	if len(names) == 0 {
		return nil
	}
	byName := make(map[string]*Const)
	if clauses != nil {
		for _, sym := range clauses.Symbols().All() {
			if c, ok := sym.(*Const); ok {
				byName[c.Name] = c
			}
		}
	}
	if mod != nil {
		for key, sort := range mod.Relations.All() {
			name := SymbolNameFromKey(key)
			if _, ok := byName[name]; !ok {
				byName[name] = NewConst(name, sort)
			}
		}
		for key, sort := range mod.Functions.All() {
			name := SymbolNameFromKey(key)
			if _, ok := byName[name]; !ok {
				byName[name] = NewConst(name, sort)
			}
		}
	}
	rels := make([]*Const, 0, len(names))
	seen := make(map[string]bool)
	for _, name := range names {
		if seen[name] {
			continue
		}
		if rel, ok := byName[name]; ok {
			rels = append(rels, rel)
			seen[name] = true
		}
	}
	return rels
}

func traceHistoryRelationsToMinimize(mod *Module, history *History, names []string) []string {
	if len(names) == 0 {
		return nil
	}
	result := make([]string, 0, len(names))
	if history == nil || len(history.Maps) == 0 {
		return append(result, names...)
	}
	firstMap := history.Maps[0]
	for _, name := range names {
		rel := traceRelationSymbolByName(mod, name)
		if rel != nil {
			if renamed, ok := firstMap.Get(rel); ok && renamed != nil {
				result = append(result, renamed.Name)
				continue
			}
		}
		result = append(result, name)
	}
	return result
}

func traceRelationSymbolByName(mod *Module, name string) *Const {
	if mod == nil || mod.Sig == nil {
		return nil
	}
	for _, sym := range mod.Sig.AllSymbolsNamed(name) {
		key := Key(sym)
		if mod.Relations != nil {
			if _, ok := mod.Relations.Get2(key); ok {
				return sym
			}
		}
		if mod.Functions != nil {
			if _, ok := mod.Functions.Get2(key); ok {
				return sym
			}
		}
	}
	return nil
}

// MakeCheckArt creates an analysis graph for checking an action.
// Matches Python ivy_trace.py make_check_art:
//  1. Create a standalone pre-state with conjectures as clauses
//  2. Execute env_action with safety checking disabled to produce graph post-state
//  3. Create a standalone fail state with fail_action(env_action) provenance
//  4. Return (ag, post, fail)
//
// Returns (ag, postState, failState).
func MakeCheckArt(mod *Module, actName string, precond []*Clauses) (*AnalysisGraph, *State, *State, error) {
	ag := NewAnalysisGraph(mod)
	mod = ag.Domain
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

	// Execute the env_action to produce the post-state.
	// Python: post = ag.execute(env_action(act_name), pre)
	// The post-state encodes the transition relation.
	envAction, err := buildEnvAction(mod, actName)
	if err != nil {
		return nil, nil, nil, err
	}
	postState, err := ag.Execute(traceCheckPrecondFalse, envAction, preState, nil, "")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("MakeCheckArt: Execute failed: %w", err)
	}
	if postState == nil {
		postState = preState
	}
	// Python: post.clauses = true_clauses()
	postState.Clauses = TrueClauses(nil)

	failAction := NewFailAction(envAction)
	failPred := postState.Pred
	if failPred == nil {
		failPred = preState
	}
	failState := NewState(mod, TrueClauses(nil))
	failState.Pred = failPred
	failState.Prov = NewActionApp(failAction, failPred)
	failState.Action = failAction

	return ag, postState, failState, nil
}

// buildEnvAction creates an EnvAction wrapping all public actions from the module.
// Matches Python ivy_actions.py env_action().
func buildEnvAction(mod *Module, actName string) (ActionsAction, error) {
	if mod == nil {
		return nil, fmt.Errorf("buildEnvAction: nil module")
	}
	if actName != "" {
		if _, ok := mod.Actions.Get2(actName); !ok {
			return nil, fmt.Errorf("buildEnvAction: action %q not found", actName)
		}
	}
	env := BuildEnvAction(mod.Cfg.ActCfg, mod.PublicActions, mod.Actions, actName, "")
	if env == nil {
		return nil, fmt.Errorf("buildEnvAction: env_action returned nil")
	}
	return env, nil
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
		panic("CheckFinalCond: history post clauses missing annotation")
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
	relsToMin = traceHistoryRelationsToMinimize(ag.Domain, history, relsToMin)
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

	var sortsToMin []Sort
	if mod != nil && mod.Sig != nil {
		sortsToMin = UninterpretedSorts(mod.Sig)
	} else {
		for _, fmla := range checkClauses.Fmlas {
			collectUninterpSorts(fmla, &sortsToMin)
		}
	}

	// Create solver and check
	slv := NewSolver(mod, nil)
	relsToMinimize := traceRelationsToMinimize(mod, checkClauses, relsToMin)
	model, err := slv.GetSmallModelWithCond(checkClauses, sortsToMin, relsToMinimize, nil, shrink)
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
	//   mclauses = and_clauses(clauses, failed)
	//   handler = Trace(mclauses, model, vocab)
	//   act.match_annotation(action, clauses.annot, handler)
	//   handler.end()
	mclauses := checkClauses
	vocabMap := mclauses.Symbols()
	vocabConsts := make([]*Const, 0, vocabMap.Len())
	vocab := make([]Expr, 0, vocabMap.Len())
	for _, sym := range vocabMap.All() {
		if c, ok := sym.(*Const); ok {
			vocabConsts = append(vocabConsts, c)
			vocab = append(vocab, c)
		}
	}
	modClauses, hm, err := slv.ClausesModelToClausesWithModelAndHerbrand(mclauses, model, traceModelIgnore(slv), true)
	if err != nil {
		fmt.Printf("CheckVC: model conversion error: %v\n", err)
		return nil
	}
	if hm == nil {
		return nil
	}
	var cfg *IvyUtilsConfig
	if mod != nil && mod.Cfg != nil {
		cfg = mod.Cfg.IuCfg
	}
	handler := NewTraceForModule(cfg, mod, mclauses, hm, vocab, true)
	handler.SetEqsFromClauses(modClauses)
	if action != nil {
		MatchAnnotation(action, clauses.Annot, handler, mod)
	}
	handler.End()
	return handler.TraceBase
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
	return MakeVCWithModule(nil, action, precond, postcond, checkAsserts)
}

func MakeVCWithModule(mod *Module, action ActionsAction, precond []*Clauses,
	postcond []*Clauses, checkAsserts bool) *Clauses {
	ag := NewAnalysisGraph(mod)
	mod = ag.Domain

	var preFmlas []Expr
	for _, p := range precond {
		if p != nil {
			preFmlas = append(preFmlas, p.Fmlas...)
		}
	}
	pre := NewClauses(preFmlas, nil, EmptyAnnotation{})
	preState := NewState(mod, pre)

	if action == nil {
		action = NewSequence()
	}
	postState, err := ag.executeStateOnly(traceCheckPrecondFalse, action.Name(), action, preState, nil)
	if err != nil {
		panic(fmt.Sprintf("MakeVC: Execute failed: %v", err))
	}
	if postState == nil {
		postState = preState
	}
	postState.Clauses = TrueClauses(nil)

	history := ag.GetHistory(postState, nil)
	if history == nil || history.Post == nil {
		return TrueClauses(EmptyAnnotation{})
	}
	clauses := history.Post
	clauses.Annot = makeVCActionAnnot(clauses.Annot)

	axioms := mod.BackgroundTheory(nil)
	if axioms == nil {
		axioms = TrueClauses(nil)
	}
	clauses = AndClausesTyped(clauses, axioms)

	var postFmlas []Expr
	for _, p := range postcond {
		if p != nil {
			postFmlas = append(postFmlas, p.Fmlas...)
		}
	}
	fc := NewClauses(postFmlas, nil, EmptyAnnotation{})
	usedNames := make(map[string]bool)
	if mod.Sig != nil && mod.Sig.Symbols != nil {
		for name := range mod.Sig.Symbols.All() {
			usedNames[name] = true
		}
	}
	witness := func(v *LogicVariable) Expr {
		name := "@" + v.Name
		if usedNames[name] {
			panic(fmt.Sprintf("MakeVC: witness symbol %q already exists", name))
		}
		return NewConst(name, v.VSort)
	}
	fcc := DualClauses(fc, witness, mod.Instantiator)
	return AndClausesTyped(clauses, fcc)
}

func makeVCActionAnnot(annot Annotation) Annotation {
	var stack []map[NodeKey]Expr
	for {
		ren, ok := annot.(*RenameAnnotation)
		if !ok {
			break
		}
		stack = append(stack, ren.Map)
		annot = ren.Arg
	}
	if comp, ok := annot.(*ComposeAnnotation); ok && len(comp.Args) > 1 {
		annot = comp.Args[1]
	}
	for i := len(stack) - 1; i >= 0; i-- {
		annot = newRenameAnnotation(annot, stack[i])
	}
	if annot == nil {
		return EmptyAnnotation{}
	}
	return annot
}

// ValueToStr converts a value to a human-readable string for trace display.
// ValueToStr converts a model value to a human-readable string.
// For constants, returns the name. For structured types (arrays, structs),
// recursively evaluates components.
//
// Python: ivy_trace.py:405-428
func ValueToStr(val Expr, evalFn func(Expr) Expr) string {
	return ValueToStrWithModule(nil, val, evalFn)
}

// ValueToStrWithModule converts a model value to the Python trace display form.
//
// Python: ivy_trace.py:value_to_str.
func ValueToStrWithModule(mod *Module, val Expr, evalFn func(Expr) Expr) string {
	if val == nil {
		return "..."
	}
	if c, ok := val.(*Const); ok {
		if rendered, ok := valueToStrArray(mod, c, evalFn); ok {
			return rendered
		}
		if rendered, ok := valueToStrDestructors(mod, c, evalFn); ok {
			return rendered
		}
		return c.Name
	}
	if app, ok := val.(*Apply); ok {
		return app.String()
	}
	return val.String()
}

func valueToStrArray(mod *Module, val *Const, evalFn func(Expr) Expr) (string, bool) {
	if mod == nil || val == nil || val.CSort == nil || evalFn == nil {
		return "", false
	}
	sortName := IvySortName(val.CSort)
	end := valueToStrLookupSymbol(mod, valueToStrComposeNames(mod, sortName, "end"))
	value := valueToStrLookupSymbol(mod, valueToStrComposeNames(mod, sortName, "value"))
	if end == nil || value == nil {
		return "", false
	}
	endSort, ok := end.CSort.(*LogicFunctionSort)
	if !ok {
		return "", false
	}
	valueSort, ok := value.CSort.(*LogicFunctionSort)
	if !ok {
		return "", false
	}
	endDom := endSort.Domain()
	valueDom := valueSort.Domain()
	if len(endDom) != 1 || len(valueDom) != 2 {
		return "", false
	}
	if !SortEqual(endDom[0], val.CSort) || !SortEqual(valueDom[0], val.CSort) || !SortEqual(valueDom[1], endSort.Range()) {
		return "", false
	}
	endVal := evalFn(MustApply(end, val))
	endConst, ok := endVal.(*Const)
	if !ok || !IsNumeral(endConst) {
		return "", false
	}
	endNum, err := strconv.Atoi(endConst.Name)
	if err != nil {
		return "", false
	}
	if endNum < 0 {
		endNum = 0
	}
	vals := make([]string, endNum)
	for i := 0; i < endNum; i++ {
		idx := NewConst(strconv.Itoa(i), endConst.CSort)
		vals[i] = ValueToStrWithModule(mod, evalFn(MustApply(value, val, idx)), evalFn)
	}
	return "[" + strings.Join(vals, ",") + "]", true
}

func valueToStrDestructors(mod *Module, val *Const, evalFn func(Expr) Expr) (string, bool) {
	if mod == nil || val == nil || val.CSort == nil || evalFn == nil {
		return "", false
	}
	destrs, ok := mod.SortDestructors.Get2(IvySortName(val.CSort))
	if !ok {
		return "", false
	}
	fields := make([]string, len(destrs))
	for i, destr := range destrs {
		rendered := "..."
		if fs, ok := destr.CSort.(*LogicFunctionSort); ok && len(fs.Domain()) == 1 {
			rendered = ValueToStrWithModule(mod, evalFn(MustApply(destr, val)), evalFn)
		}
		fields[i] = destr.Name + ":" + rendered
	}
	return "{" + strings.Join(fields, ",") + "}", true
}

func valueToStrLookupSymbol(mod *Module, name string) *Const {
	if mod == nil || name == "" {
		return nil
	}
	if mod.Sig != nil && mod.Sig.Symbols != nil {
		if syms := mod.Sig.AllSymbolsNamed(name); len(syms) > 0 {
			return syms[0]
		}
	}
	return nil
}

func valueToStrComposeNames(mod *Module, names ...string) string {
	if mod != nil && mod.Cfg != nil && mod.Cfg.IuCfg != nil {
		return mod.Cfg.IuCfg.ComposeNames(names...)
	}
	if len(names) > 0 && names[0] == "this" {
		names = names[1:]
	}
	return strings.Join(names, ".")
}
