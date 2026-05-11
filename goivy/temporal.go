// Package temporal defines structures for temporal proof goals and provides
// tactics for liveness proofs in the Ivy verification system.
//
// A temporal proof goal is of the form M |= phi where M is a program term
// and phi is a temporal formula. This package handles the conversion of
// modules to normal program form and the invariance tactic for proving
// globally formulas using the invariance rule.
//
// Ported from ivy_temporal.py.
package goivy

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// ActionTerm represents an action term with inputs, outputs, labels, and a body statement.
//
// An action term is of the form:
//
//	action (inputs) returns (outputs) { stmt }
type ActionTerm struct {
	Inputs  []*Const
	Outputs []*Const
	Labels  []string
	Stmt    ActionsAction
	// ast.Node support (matches Python ia.AST base: lineno, config)
	Loc    Location
	HasLoc bool
	Cfg    *AstConfig
}

// String returns a human-readable representation of the action term.
func (at *ActionTerm) String() string {
	res := "action"
	if len(at.Inputs) > 0 {
		res += ParamsToStr(at.Inputs)
	}
	if len(at.Outputs) > 0 {
		res += " returns" + ParamsToStr(at.Outputs)
	}
	res += "{" + at.Stmt.String() + "}"
	return res
}

// CloneStmt creates a copy of the ActionTerm with a new statement but preserving
// all other attributes. Used by callers that build a new Stmt manually.
func (at *ActionTerm) CloneStmt(stmt ActionsAction) *ActionTerm {
	return &ActionTerm{
		Inputs:  at.Inputs,
		Outputs: at.Outputs,
		Labels:  at.Labels,
		Stmt:    stmt,
		Loc:     at.Loc,
		HasLoc:  at.HasLoc,
		Cfg:     at.Cfg,
	}
}

// --- ast.Node interface for ActionTerm ---
// Python: ActionTerm(ia.AST) with args=[self.stmt]

func (at *ActionTerm) Args() []Node             { return []Node{at.Stmt} }
func (at *ActionTerm) GetLineno() Location      { return at.Loc }
func (at *ActionTerm) SetLineno(l Location)     { at.Loc = l; at.HasLoc = true }
func (at *ActionTerm) GetAstConfig() *AstConfig { return at.Cfg }

// Clone creates a copy with transformed children (ast.Node interface).
// Python: ActionTerm.clone(args) replaces stmt with args[0], copies attrs.
func (at *ActionTerm) Clone(args []Node) Node {
	return &ActionTerm{
		Inputs:  at.Inputs,
		Outputs: at.Outputs,
		Labels:  at.Labels,
		Stmt:    args[0].(ActionsAction),
		Loc:     at.Loc,
		HasLoc:  at.HasLoc,
		Cfg:     at.Cfg,
	}
}

// Canon returns a canonical s-expression with all fields.
func (at *ActionTerm) Canon() Canonical {
	return canonString(func(w io.Writer) { writeActionTermCanon(w, at) })
}

// ActionTermBinding binds an action term to a name.
// Python: ActionTermBinding(ia.AST) with args=[self.action]
type ActionTermBinding struct {
	Name   string
	Action *ActionTerm
	// ast.Node support (matches Python ia.AST base: lineno, config)
	Loc    Location
	HasLoc bool
	Cfg    *AstConfig
}

// String returns a human-readable representation of the binding.
func (b *ActionTermBinding) String() string {
	return fmt.Sprintf("%s = %s", b.Name, b.Action)
}

// CloneAction creates a copy of the binding with a new action.
// Used by callers that build a new ActionTerm manually.
func (b *ActionTermBinding) CloneAction(action *ActionTerm) *ActionTermBinding {
	return &ActionTermBinding{
		Name:   b.Name,
		Action: action,
		Loc:    b.Loc,
		HasLoc: b.HasLoc,
		Cfg:    b.Cfg,
	}
}

// --- ast.Node interface for ActionTermBinding ---

func (b *ActionTermBinding) Args() []Node             { return []Node{b.Action} }
func (b *ActionTermBinding) GetLineno() Location      { return b.Loc }
func (b *ActionTermBinding) SetLineno(l Location)     { b.Loc = l; b.HasLoc = true }
func (b *ActionTermBinding) GetAstConfig() *AstConfig { return b.Cfg }

// Clone creates a copy with transformed children (ast.Node interface).
// Python: ActionTermBinding.clone(args) replaces action with args[0], copies attrs.
func (b *ActionTermBinding) Clone(args []Node) Node {
	return &ActionTermBinding{
		Name:   b.Name,
		Action: args[0].(*ActionTerm),
		Loc:    b.Loc,
		HasLoc: b.HasLoc,
		Cfg:    b.Cfg,
	}
}

// Canon returns a canonical s-expression with all fields.
func (b *ActionTermBinding) Canon() Canonical {
	return canonString(func(w io.Writer) { writeActionTermBindingCanon(w, b) })
}

// --- canon helpers ---

// constSliceCanon returns canonical form for []*lg.Const.
func constSliceCanon(cs []*Const) string {
	if len(cs) == 0 {
		return "[]"
	}
	parts := make([]string, len(cs))
	for i, c := range cs {
		parts[i] = string(c.Canon())
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// stringSliceCanon returns canonical form for []string.
func temporalStringSliceCanon(ss []string) string {
	var b strings.Builder
	appendTemporalStringSliceCanon(&b, ss)
	return b.String()
}

func appendTemporalStringSliceCanon(b *strings.Builder, ss []string) {
	if len(ss) == 0 {
		b.WriteString("[]")
		return
	}
	b.WriteByte('[')
	for i, s := range ss {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(strconv.Quote(s))
	}
	b.WriteByte(']')
}

// NormalProgram represents a normal program consisting of bindings, an
// initial statement, invariants, assumptions, and calls.
//
// A normal program has the form:
//
//	let
//	    name1 = action1
//	    ...
//	in
//	    init;
//	    while *
//	    invariant inv1;
//	    ...
//	    {
//	        diverge;
//	        assume asm1;
//	        ...
//	        if * { call1 }
//	        else if * { call2 }
//	        ...
//	    }
type NormalProgram struct {
	Base      // Python: extends ia.AST — provides GetLineno/SetLineno
	Bindings  []*ActionTermBinding
	Init      ActionsAction
	Invars    []*LabeledFormula
	Asms      []*LabeledFormula
	Calls     []string
	Postconds map[string][]*LabeledFormula // optional postconditions
}

// Fmlas returns all component formulas of the program.
// Matches Python NormalProgram.fmlas property (ivy_temporal.py:188-193).
func (np *NormalProgram) Fmlas() []Node {
	var res []Node
	for _, b := range np.Bindings {
		res = append(res, b)
	}
	if np.Init != nil {
		res = append(res, np.Init)
	}
	for _, inv := range np.Invars {
		res = append(res, inv)
	}
	for _, asm := range np.Asms {
		res = append(res, asm)
	}
	for _, pcs := range np.Postconds {
		for _, pc := range pcs {
			res = append(res, pc)
		}
	}
	return res
}

// Args returns the child AST nodes. Python: args property returns [].
func (np *NormalProgram) Args() []Node { return nil }

// Clone creates a copy of this node. Python: clone just copies.
func (np *NormalProgram) Clone(args []Node) Node {
	return NormalProgramClone(np)
}

// BindingMap returns a map from binding names to their underlying actions.
func (np *NormalProgram) BindingMap() map[string]ActionsAction {
	m := make(map[string]ActionsAction, len(np.Bindings))
	for _, b := range np.Bindings {
		m[b.Name] = b.Action.Stmt
	}
	return m
}

// Formulas returns all formulas (bindings' actions, init, invariants, assumptions)
// contained in the normal program.
func (np *NormalProgram) Formulas() []interface{} {
	var res []interface{}
	for _, b := range np.Bindings {
		res = append(res, b)
	}
	res = append(res, np.Init)
	for _, inv := range np.Invars {
		res = append(res, inv)
	}
	for _, asm := range np.Asms {
		res = append(res, asm)
	}
	if np.Postconds != nil {
		for _, pcs := range np.Postconds {
			for _, pc := range pcs {
				res = append(res, pc)
			}
		}
	}
	return res
}

// Canon returns a canonical s-expression for the normal program.
// Postconds map keys are sorted to keep output deterministic.
func (np *NormalProgram) Canon() Canonical {
	return canonString(func(w io.Writer) { writeNormalProgramCanon(w, np) })
}

// bindingSliceCanon canonicalizes []*ActionTermBinding.
func bindingSliceCanon(bs []*ActionTermBinding) string {
	var b strings.Builder
	appendBindingSliceCanon(&b, bs)
	return b.String()
}

func appendBindingSliceCanon(b *strings.Builder, bs []*ActionTermBinding) {
	if len(bs) == 0 {
		b.WriteString("[]")
		return
	}
	b.WriteByte('[')
	for i, binding := range bs {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(string(binding.Canon()))
	}
	b.WriteByte(']')
}

// lfSliceCanon canonicalizes []*ast.LabeledFormula.
func lfSliceCanon(lfs []*LabeledFormula) string {
	var b strings.Builder
	appendLFSliceCanon(&b, lfs)
	return b.String()
}

func appendLFSliceCanon(b *strings.Builder, lfs []*LabeledFormula) {
	if len(lfs) == 0 {
		b.WriteString("[]")
		return
	}
	b.WriteByte('[')
	for i, lf := range lfs {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(string(lf.Canon()))
	}
	b.WriteByte(']')
}

// postcondsHashCanon canonicalizes map[string][]*ast.LabeledFormula
// as "(hash "k1":[...] "k2":[...])" with keys sorted lexicographically.
func postcondsHashCanon(m map[string][]*LabeledFormula) string {
	var b strings.Builder
	appendPostcondsHashCanon(&b, m)
	return b.String()
}

func appendPostcondsHashCanon(b *strings.Builder, m map[string][]*LabeledFormula) {
	if len(m) == 0 {
		b.WriteString("(hash)")
		return
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b.WriteString("(hash ")
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(strconv.Quote(k))
		b.WriteByte(':')
		appendLFSliceCanon(b, m[k])
	}
	b.WriteByte(')')
}

// String returns a human-readable representation of the normal program.
func (np *NormalProgram) String() string {
	var b strings.Builder
	b.WriteString("\nlet\n")
	for _, bind := range np.Bindings {
		b.WriteString("    " + bind.String() + "\n")
	}
	b.WriteString("in\n")
	b.WriteString("    " + np.Init.String() + "\n")
	b.WriteString("    while *\n")
	for _, invar := range np.Invars {
		b.WriteString("        invariant " + fmt.Sprint(invar) + "\n")
	}
	b.WriteString("    {\n")
	b.WriteString("        diverge;\n")
	for _, asm := range np.Asms {
		b.WriteString("        assume " + fmt.Sprint(asm) + ";\n")
	}
	callStrs := make([]string, len(np.Calls))
	for i, c := range np.Calls {
		callStrs[i] = c
	}
	b.WriteString("        call one of {" + strings.Join(callStrs, ",") + "}\n")
	b.WriteString("    }\n")
	return b.String()
}

// OldActionToNew converts an actions.ActionsAction (the "old" representation)
// to an ActionTerm (the "new" representation) by extracting formals and labels.
func OldActionToNew(act ActionsAction) *ActionTerm {
	return &ActionTerm{
		Inputs:  act.GetFormalParams(),
		Outputs: act.GetFormalReturns(),
		Labels:  getLabels(act),
		Stmt:    act,
	}
}

// getLabels extracts labels from an action if available.
func getLabels(act ActionsAction) []string {
	if ab, ok := act.(interface{ GetLabels() []string }); ok {
		return ab.GetLabels()
	}
	// Try to access the Labels field via the ActionBase
	type labeler interface {
		SetLabels([]string)
	}
	// Access Labels field directly if possible through ActionBase
	switch a := act.(type) {
	case *LogicSequence:
		return a.Labels
	case *LogicAssumeAction:
		return a.Labels
	case *LogicAssertAction:
		return a.Labels
	case *LogicCallAction:
		return a.Labels
	case *LogicChoiceAction:
		return a.Labels
	case *LogicEnvAction:
		return a.LogicChoiceAction.Labels
	case *LogicIfAction:
		return a.Labels
	}
	return nil
}

// NewActionToOld converts an ActionTerm back to its underlying action.
func NewActionToOld(act *ActionTerm) ActionsAction {
	return act.Stmt
}

// NormalProgramFromModule creates a NormalProgram from a module by
// extracting its actions, initializers, conjectures, and public actions.
func NormalProgramFromModule(mod *Module) *NormalProgram {
	var bindings []*ActionTermBinding
	for name, actIface := range mod.Actions.All() {
		if act, ok := actIface.(ActionsAction); ok {
			bindings = append(bindings, &ActionTermBinding{
				Name:   name,
				Action: OldActionToNew(act),
			})
		}
	}
	// python uses dict insert order, which we also do without this:
	//sort.Slice(bindings, func(i, j int) bool {
	//	return bindings[i].Name < bindings[j].Name
	//})

	// Build init from initializers
	var initNodes []Expr
	for _, na := range mod.Initializers {
		if act, ok := na.Action.(ActionsAction); ok {
			initNodes = append(initNodes, act)
		}
	}
	var init ActionsAction
	if len(initNodes) > 0 {
		init = NewSequence(initNodes...)
	} else {
		init = NewSequence()
	}

	// Sort public actions
	calls := make([]string, 0, mod.PublicActions.Len())
	for name := range mod.PublicActions.All() {
		calls = append(calls, name)
	}
	sort.Strings(calls)

	xtracer.Trace("temporal.NormalProgramFromModule EXIT nInvars=%d nAsms=%d nBindings=%d",
		len(mod.LabeledConjs), len(mod.AssumedInvs), len(bindings))
	return &NormalProgram{
		Bindings: bindings,
		Init:     init,
		Invars:   mod.LabeledConjs,
		Asms:     mod.AssumedInvs,
		Calls:    calls,
	}
}

// EnvAction creates an environment action from a list of action bindings.
// This represents the environment nondeterministically calling one of the actions.
func TemporalEnvAction(cfg *ActionsConfig, bindings []*ActionTermBinding) *LogicEnvAction {
	var branches []Expr
	for _, b := range bindings {
		name := b.Name
		act := b.Action
		ract := NewSequence(
			act.Stmt,
			NewReturnAction(),
		)
		ract.SetFormalParams(act.Inputs)
		ract.SetFormalReturns(act.Outputs)
		if strings.HasPrefix(name, "ext:") {
			name = name[4:]
		}
		ract.SetLabel(name) // Python: ract.label = name (singular)
		branches = append(branches, ract)
	}
	return NewEnvActionOn(cfg, branches...)
}

// TemporalModels represents M |= phi, a temporal proof goal.
// It wraps the ast.TemporalModels type for convenience.
// PropEvent computes the event action for a temporal property.
// For G phi (Globally) formulas, the event is "assume phi".
// For F ~phi (Eventually with negation) formulas, the event is "assert phi".
func TemporalPropEvent(gprop Expr, lineno Location) ActionsAction {
	switch g := gprop.(type) {
	case *LogicEventually:
		// Formula of the form F ~phi translates to "assert phi"
		if not, ok := g.Body.(*LogicNot); ok {
			res := NewAssertAction(not.Body)
			res.SetLineno(lineno)
			return res
		}
		// If body is not negated, just assert the body
		res := NewAssertAction(g.Body)
		res.SetLineno(lineno)
		return res
	case *LogicGlobally:
		// Formula of the form G phi translates to "assume phi"
		res := NewAssumeAction(g.Body)
		res.SetLineno(lineno)
		return res
	default:
		// Fallback: assume
		res := NewAssumeAction(gprop)
		res.SetLineno(lineno)
		return res
	}
}

// PrefixActionTerm creates a new ActionTerm with statements prepended
// before the original body statement.
func PrefixActionTerm(at *ActionTerm, stmts []ActionsAction) *ActionTerm {
	newStmt := PrefixAction(at.Stmt, stmts)
	return at.CloneStmt(newStmt)
}

// IsGloballyFormula returns true if the given node is a logic.Globally formula.
func IsGloballyFormula(n Expr) bool {
	_, ok := n.(*LogicGlobally)
	return ok
}

// IsEventuallyFormula returns true if the given node is a logic.Eventually formula.
func IsEventuallyFormula(n Expr) bool {
	_, ok := n.(*LogicEventually)
	return ok
}

// IsTemporalFormula returns true if the formula is a temporal operator.
func IsTemporalFormula(n Expr) bool {
	switch n.(type) {
	case *LogicGlobally, *LogicEventually, *LogicWhenOperator:
		return true
	}
	return false
}

// HasTemporalOperator returns true if the formula or any subformula
// contains a temporal operator.
func HasTemporalOperator(n Expr) bool {
	if IsTemporalFormula(n) {
		return true
	}
	for _, c := range n.Children() {
		if HasTemporalOperator(c) {
			return true
		}
	}
	return false
}

// IsGprop returns true if the formula is Globally(phi) where phi
// has no temporal operators.
func TemporalIsGprop(n Expr) bool {
	g, ok := n.(*LogicGlobally)
	if !ok {
		return false
	}
	return !HasTemporalOperator(g.Body)
}

// GetEnviron returns the environment label of a temporal formula, or nil.
func GetEnviron(n Expr) *string {
	switch g := n.(type) {
	case *LogicGlobally:
		return g.Environ
	case *LogicEventually:
		return g.Environ
	}
	return nil
}

// EnvironStr returns the environment as a string or empty string if nil.
func EnvironStr(n Expr) string {
	env := GetEnviron(n)
	if env != nil {
		return *env
	}
	return ""
}

// NormalProgramClone mirrors Python NormalProgram.clone (ivy_temporal.py:179-183),
// which constructs a new NormalProgram with the SAME slice/map references
// as the original — appending to a clone's slice mutates the original.
// goivy must preserve this aliasing because Python's check_isolate /
// l2s_tactic_int flow relies on it (e.g. ivy_l2s.py:159 model.asms.append
// mutates im.module.assumed_invariants in place; ivy_check.py:761 then
// extends the same shared list).
func NormalProgramClone(np *NormalProgram) *NormalProgram {
	result := &NormalProgram{
		Base:     np.Base,
		Bindings: np.Bindings,
		Init:     np.Init,
		Invars:   np.Invars,
		Asms:     np.Asms,
		Calls:    np.Calls,
	}
	if np.Postconds != nil {
		result.Postconds = np.Postconds
	}
	return result
}

// InvarianceTactic proves a "globally phi" formula using the invariance rule.
// It instruments all actions with property events and converts "M |= G phi" to "M |= true".
//
// Python: ivy_temporal.py:invariance_tactic (lines 257-393)
func InvarianceTactic(pc ProofCheckerInterface, goals []*LabeledFormula, pf Node) ([]*LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, fmt.Errorf("invariance: no proof goals")
	}
	goal := goals[0]

	// Find the TemporalModels conclusion
	tm := temporalFindTemporalModels(goal)
	if tm == nil {
		return nil, fmt.Errorf("temporal/temporal: invariance: [4]proof goal is not temporal")
	}

	fmla, _ := tm.Fmla.(Expr)
	if fmla == nil {
		return nil, fmt.Errorf("invariance: could not extract formula from goal")
	}

	glob, ok := fmla.(*LogicGlobally)
	if !ok {
		return nil, fmt.Errorf("invariance: tactic applies only to globally formulas")
	}

	invar := glob.Body
	if HasTemporalOperator(invar) {
		return nil, fmt.Errorf("invariance: tactic applies only to formulas 'globally p' where p is non-temporal")
	}

	// Get the model from the TemporalModels conclusion (Python line 262: model = conc.model)
	// Clone before mutating (Python line 278: model = model.clone([]))
	var model *NormalProgram
	if tm.Model != nil {
		model = NormalProgramClone(tm.Model.(*NormalProgram))
	} else if pc.GetModule() != nil {
		model = NormalProgramClone(NormalProgramFromModule(pc.GetModule()))
	} else {
		model = &NormalProgram{Init: NewSequence()}
	}

	// Add the invariant phi to the model's invariants. Python ivy_temporal.py:275:
	//   invars.append(ipr.clone_goal(goal,[],invar))
	// where clone_goal calls goal.clone_with_fresh_id([goal.label, invar]) — fresh
	// id is intentional, but the label must be the goal's label, not nil.
	model.Invars = append(model.Invars, goal.CloneWithFreshID([]Node{goal.Label, invar}))

	// Collect assumed globally properties from prover axioms
	var gprops []Expr
	var gpropLines []Location
	if pc != nil {
		for _, ax := range pc.GetAxioms() {
			if !ax.Explicit && ax.IsTemporal() {
				if f, ok := ax.Formula.(Expr); ok {
					if TemporalIsGprop(f) {
						gprops = append(gprops, f)
						gpropLines = append(gpropLines, ax.GetLineno())
					}
				}
			}
		}
	}

	// Add the negation of the property: F ~phi
	gprops = append(gprops, &LogicEventually{Environ: glob.Environ, Body: &LogicNot{Body: invar}})
	gpropLines = append(gpropLines, goal.GetLineno())

	// Build memo tables: environ -> props, symbol -> props
	envprops := make(map[string][]Expr)
	symprops := make(map[NodeKey][]Expr)
	propLines := make(map[string]Location) // prop string -> lineno

	for i, prop := range gprops {
		env := EnvironStr(prop)
		envprops[env] = append(envprops[env], prop)
		propLines[fmt.Sprint(prop)] = gpropLines[i]
		for _, sym := range symbolsAst(prop) {
			symprops[Key(sym)] = append(symprops[Key(sym)], prop)
		}
	}

	// Build action map for callee resolution
	actionMap := make(map[string]*ActionTerm)
	for _, b := range model.Bindings {
		actionMap[b.Name] = b.Action
	}

	// instrStmt instruments a statement with property events
	var instrStmt func(stmt ActionsAction, labels []string) ActionsAction
	instrStmt = func(stmt ActionsAction, labels []string) ActionsAction {
		// Recur on sub-statements
		args := stmt.ActionArgs()
		newArgs := make([]Expr, len(args))
		for i, a := range args {
			if sub, ok := a.(ActionsAction); ok {
				newArgs[i] = instrStmt(sub, labels)
			} else {
				newArgs[i] = a
			}
		}
		res := stmt.ActionClone(newArgs)

		eventProps := make(map[string]Expr) // deduped by string

		// If it is a call, check for events on return
		if call, ok := stmt.(*LogicCallAction); ok {
			calleeName := call.Callee.String()
			if at, ok := actionMap[calleeName]; ok {
				labelSet := make(map[string]bool)
				for _, l := range labels {
					labelSet[l] = true
				}
				for _, l := range at.Labels {
					if !labelSet[l] {
						// Match Python defaultdict auto-vivification
						if _, ok := envprops[l]; !ok {
							envprops[l] = nil
						}
						for _, prop := range envprops[l] {
							eventProps[fmt.Sprint(prop)] = prop
						}
					}
				}
			}
		}

		// If a symbol is modified, add events for properties depending on it
		mods := Modifies(stmt)
		labelSet := make(map[string]bool)
		for _, l := range labels {
			labelSet[l] = true
		}
		for _, sym := range mods {
			// Match Python defaultdict auto-vivification
			sk := Key(sym)
			if _, ok := symprops[sk]; !ok {
				symprops[sk] = nil
			}
			for _, prop := range symprops[sk] {
				env := EnvironStr(prop)
				if !labelSet[env] {
					eventProps[fmt.Sprint(prop)] = prop
				}
			}
		}

		// Add property events
		var events []ActionsAction
		for key, prop := range eventProps {
			loc := propLines[key]
			events = append(events, TemporalPropEvent(prop, loc))
		}

		res = PostfixAction(res, events)
		CopyFormalsTo(stmt, res)
		return res
	}

	// Instrument all bindings
	for i, b := range model.Bindings {
		newStmt := instrStmt(b.Action.Stmt, b.Action.Labels)
		model.Bindings[i] = b.CloneAction(b.Action.CloneStmt(newStmt))
	}

	// Add assumed G-properties as model assumptions. Python ivy_temporal.py:378:
	//   model.asms.extend([p.clone([p.label,p.formula.args[0]]) for p in assumed_gprops])
	// Use Clone (not NewLabeledFormula(nil,...)) so the axiom's label, id, and
	// metadata flags are preserved.
	if pc != nil {
		for _, ax := range pc.GetAxioms() {
			if !ax.Explicit && ax.IsTemporal() {
				if f, ok := ax.Formula.(Expr); ok {
					if g, ok := f.(*LogicGlobally); ok {
						model.Asms = append(model.Asms, ax.Clone([]Node{ax.Label, g.Body}).(*LabeledFormula))
					}
				}
			}
		}
	}

	// Change conclusion to M |= true (Python line 385: conc = TemporalModels(model, il.And()))
	newConc := pc.GetAstCfg().NewTemporalModels(model, True)

	// Build new goal.
	prems := GoalPrems(goal)
	newGoal := CloneGoal(pc.GetAstCfg(), goal, prems, newConc)

	result := make([]*LabeledFormula, len(goals))
	result[0] = newGoal
	copy(result[1:], goals[1:])
	return result, nil
}

// findTemporalModels looks for a TemporalModels in the goal.
func temporalFindTemporalModels(goal *LabeledFormula) *TemporalModels {
	if goal == nil || goal.Formula == nil {
		return nil
	}
	if tm, ok := goal.Formula.(*TemporalModels); ok {
		return tm
	}
	if sb, ok := goal.Formula.(*SchemaBody); ok {
		if c := sb.Conc(); c != nil {
			if tm, ok := c.(*TemporalModels); ok {
				return tm
			}
		}
	}
	return nil
}

// symbolsAst collects symbols from a logic node (wrapper for package access).
func symbolsAst(n Expr) []*Const {
	var result []*Const
	temporalSymbolsAstRec(n, &result, make(map[NodeKey]bool))
	return result
}

func temporalSymbolsAstRec(n Expr, result *[]*Const, seen map[NodeKey]bool) {
	if c, ok := n.(*Const); ok {
		if !seen[Key(c)] {
			seen[Key(c)] = true
			*result = append(*result, c)
		}
	}
	if app, ok := n.(*Apply); ok {
		if c, ok := app.Func.(*Const); ok {
			if !seen[Key(c)] {
				seen[Key(c)] = true
				*result = append(*result, c)
			}
		}
	}
	for _, child := range n.Children() {
		temporalSymbolsAstRec(child, result, seen)
	}
}

// RegisterTactics registers the invariance tactic on the given proof config.
// Replaces the old init()-based global registration.
func RegisterTemporalTactics(proofCfg *ProofConfig) {
	proofCfg.RegisterTactic("invariance", InvarianceTactic)
}
