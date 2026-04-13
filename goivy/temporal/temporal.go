// Package temporal defines structures for temporal proof goals and provides
// tactics for liveness proofs in the Ivy verification system.
//
// A temporal proof goal is of the form M |= phi where M is a program term
// and phi is a temporal formula. This package handles the conversion of
// modules to normal program form and the invariance tactic for proving
// globally formulas using the invariance rule.
//
// Ported from ivy_temporal.py.
package temporal

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/proof"
)

// ActionTerm represents an action term with inputs, outputs, labels, and a body statement.
//
// An action term is of the form:
//
//	action (inputs) returns (outputs) { stmt }
type ActionTerm struct {
	Inputs  []*lg.Const
	Outputs []*lg.Const
	Labels  []string
	Stmt    actions.Action
	// ast.Node support (matches Python ia.AST base: lineno, config)
	Loc    ast.Location
	HasLoc bool
	Cfg    *ast.AstConfig
}

// String returns a human-readable representation of the action term.
func (at *ActionTerm) String() string {
	res := "action"
	if len(at.Inputs) > 0 {
		res += actions.ParamsToStr(at.Inputs)
	}
	if len(at.Outputs) > 0 {
		res += " returns" + actions.ParamsToStr(at.Outputs)
	}
	res += "{" + at.Stmt.String() + "}"
	return res
}

// CloneStmt creates a copy of the ActionTerm with a new statement but preserving
// all other attributes. Used by callers that build a new Stmt manually.
func (at *ActionTerm) CloneStmt(stmt actions.Action) *ActionTerm {
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

func (at *ActionTerm) Args() []ast.Node           { return []ast.Node{at.Stmt} }
func (at *ActionTerm) GetLineno() ast.Location     { return at.Loc }
func (at *ActionTerm) SetLineno(l ast.Location)    { at.Loc = l; at.HasLoc = true }
func (at *ActionTerm) GetAstConfig() *ast.AstConfig { return at.Cfg }

// Clone creates a copy with transformed children (ast.Node interface).
// Python: ActionTerm.clone(args) replaces stmt with args[0], copies attrs.
func (at *ActionTerm) Clone(args []ast.Node) ast.Node {
	return &ActionTerm{
		Inputs:  at.Inputs,
		Outputs: at.Outputs,
		Labels:  at.Labels,
		Stmt:    args[0].(actions.Action),
		Loc:     at.Loc,
		HasLoc:  at.HasLoc,
		Cfg:     at.Cfg,
	}
}

// Canon returns a canonical s-expression with all fields.
func (at *ActionTerm) Canon() iu.Canonical {
	var lf string
	if at.HasLoc {
		lf = fmt.Sprintf(" lineno:%d", at.Loc.Line)
	}
	return iu.Canonical(fmt.Sprintf("(actionTerm%s inputs:%s outputs:%s labels:%s stmt:%s)",
		lf, constSliceCanon(at.Inputs), constSliceCanon(at.Outputs),
		stringSliceCanon(at.Labels), at.Stmt.Canon()))
}

// ActionTermBinding binds an action term to a name.
// Python: ActionTermBinding(ia.AST) with args=[self.action]
type ActionTermBinding struct {
	Name   string
	Action *ActionTerm
	// ast.Node support (matches Python ia.AST base: lineno, config)
	Loc    ast.Location
	HasLoc bool
	Cfg    *ast.AstConfig
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

func (b *ActionTermBinding) Args() []ast.Node            { return []ast.Node{b.Action} }
func (b *ActionTermBinding) GetLineno() ast.Location      { return b.Loc }
func (b *ActionTermBinding) SetLineno(l ast.Location)     { b.Loc = l; b.HasLoc = true }
func (b *ActionTermBinding) GetAstConfig() *ast.AstConfig { return b.Cfg }

// Clone creates a copy with transformed children (ast.Node interface).
// Python: ActionTermBinding.clone(args) replaces action with args[0], copies attrs.
func (b *ActionTermBinding) Clone(args []ast.Node) ast.Node {
	return &ActionTermBinding{
		Name:   b.Name,
		Action: args[0].(*ActionTerm),
		Loc:    b.Loc,
		HasLoc: b.HasLoc,
		Cfg:    b.Cfg,
	}
}

// Canon returns a canonical s-expression with all fields.
func (b *ActionTermBinding) Canon() iu.Canonical {
	var lf string
	if b.HasLoc {
		lf = fmt.Sprintf(" lineno:%d", b.Loc.Line)
	}
	return iu.Canonical(fmt.Sprintf("(actionTermBinding%s name:%q action:%s)",
		lf, b.Name, b.Action.Canon()))
}

// --- canon helpers ---

// constSliceCanon returns canonical form for []*lg.Const.
func constSliceCanon(cs []*lg.Const) string {
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
func stringSliceCanon(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	parts := make([]string, len(ss))
	for i, s := range ss {
		parts[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(parts, " ") + "]"
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
	ast.Base  // Python: extends ia.AST — provides GetLineno/SetLineno
	Bindings  []*ActionTermBinding
	Init      actions.Action
	Invars    []*ast.LabeledFormula
	Asms      []*ast.LabeledFormula
	Calls     []string
	Postconds map[string][]*ast.LabeledFormula // optional postconditions
}

// Args returns the child AST nodes. Python: args property returns [].
func (np *NormalProgram) Args() []ast.Node { return nil }

// Clone creates a copy of this node. Python: clone just copies.
func (np *NormalProgram) Clone(args []ast.Node) ast.Node {
	return NormalProgramClone(np)
}

// BindingMap returns a map from binding names to their underlying actions.
func (np *NormalProgram) BindingMap() map[string]actions.Action {
	m := make(map[string]actions.Action, len(np.Bindings))
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

// OldActionToNew converts an actions.Action (the "old" representation)
// to an ActionTerm (the "new" representation) by extracting formals and labels.
func OldActionToNew(act actions.Action) *ActionTerm {
	return &ActionTerm{
		Inputs:  act.GetFormalParams(),
		Outputs: act.GetFormalReturns(),
		Labels:  getLabels(act),
		Stmt:    act,
	}
}

// getLabels extracts labels from an action if available.
func getLabels(act actions.Action) []string {
	if ab, ok := act.(interface{ GetLabels() []string }); ok {
		return ab.GetLabels()
	}
	// Try to access the Labels field via the ActionBase
	type labeler interface {
		SetLabels([]string)
	}
	// Access Labels field directly if possible through ActionBase
	switch a := act.(type) {
	case *actions.Sequence:
		return a.Labels
	case *actions.AssumeAction:
		return a.Labels
	case *actions.AssertAction:
		return a.Labels
	case *actions.CallAction:
		return a.Labels
	case *actions.ChoiceAction:
		return a.Labels
	case *actions.EnvAction:
		return a.ChoiceAction.Labels
	case *actions.IfAction:
		return a.Labels
	}
	return nil
}

// NewActionToOld converts an ActionTerm back to its underlying action.
func NewActionToOld(act *ActionTerm) actions.Action {
	return act.Stmt
}

// NormalProgramFromModule creates a NormalProgram from a module by
// extracting its actions, initializers, conjectures, and public actions.
func NormalProgramFromModule(mod *module.Module) *NormalProgram {
	var bindings []*ActionTermBinding
	for name, actIface := range mod.Actions.All() {
		if act, ok := actIface.(actions.Action); ok {
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
	var initNodes []lg.Expr
	for _, na := range mod.Initializers {
		if act, ok := na.Action.(actions.Action); ok {
			initNodes = append(initNodes, act)
		}
	}
	var init actions.Action
	if len(initNodes) > 0 {
		init = actions.NewSequence(initNodes...)
	} else {
		init = actions.NewSequence()
	}

	// Sort public actions
	calls := make([]string, 0, mod.PublicActions.Len())
	for name := range mod.PublicActions.All() {
		calls = append(calls, name)
	}
	sort.Strings(calls)

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
func EnvAction(bindings []*ActionTermBinding) *actions.EnvAction {
	var branches []lg.Expr
	for _, b := range bindings {
		name := b.Name
		act := b.Action
		ract := actions.NewSequence(
			act.Stmt,
			actions.NewReturnAction(),
		)
		ract.SetFormalParams(act.Inputs)
		ract.SetFormalReturns(act.Outputs)
		if strings.HasPrefix(name, "ext:") {
			name = name[4:]
		}
		ract.SetLabel(name) // Python: ract.label = name (singular)
		branches = append(branches, ract)
	}
	return actions.NewEnvAction(branches...)
}

// TemporalModels represents M |= phi, a temporal proof goal.
// It wraps the ast.TemporalModels type for convenience.
type TemporalModels = ast.TemporalModels

// PropEvent computes the event action for a temporal property.
// For G phi (Globally) formulas, the event is "assume phi".
// For F ~phi (Eventually with negation) formulas, the event is "assert phi".
func PropEvent(gprop lg.Expr, lineno ast.Location) actions.Action {
	switch g := gprop.(type) {
	case *lg.Eventually:
		// Formula of the form F ~phi translates to "assert phi"
		if not, ok := g.Body.(*lg.Not); ok {
			res := actions.NewAssertAction(not.Body)
			res.SetLineno(lineno)
			return res
		}
		// If body is not negated, just assert the body
		res := actions.NewAssertAction(g.Body)
		res.SetLineno(lineno)
		return res
	case *lg.Globally:
		// Formula of the form G phi translates to "assume phi"
		res := actions.NewAssumeAction(g.Body)
		res.SetLineno(lineno)
		return res
	default:
		// Fallback: assume
		res := actions.NewAssumeAction(gprop)
		res.SetLineno(lineno)
		return res
	}
}

// PrefixActionTerm creates a new ActionTerm with statements prepended
// before the original body statement.
func PrefixActionTerm(at *ActionTerm, stmts []actions.Action) *ActionTerm {
	newStmt := actions.PrefixAction(at.Stmt, stmts)
	return at.CloneStmt(newStmt)
}

// IsGloballyFormula returns true if the given node is a logic.Globally formula.
func IsGloballyFormula(n lg.Expr) bool {
	_, ok := n.(*lg.Globally)
	return ok
}

// IsEventuallyFormula returns true if the given node is a logic.Eventually formula.
func IsEventuallyFormula(n lg.Expr) bool {
	_, ok := n.(*lg.Eventually)
	return ok
}

// IsTemporalFormula returns true if the formula is a temporal operator.
func IsTemporalFormula(n lg.Expr) bool {
	switch n.(type) {
	case *lg.Globally, *lg.Eventually, *lg.WhenOperator:
		return true
	}
	return false
}

// HasTemporalOperator returns true if the formula or any subformula
// contains a temporal operator.
func HasTemporalOperator(n lg.Expr) bool {
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
func IsGprop(n lg.Expr) bool {
	g, ok := n.(*lg.Globally)
	if !ok {
		return false
	}
	return !HasTemporalOperator(g.Body)
}

// GetEnviron returns the environment label of a temporal formula, or nil.
func GetEnviron(n lg.Expr) *string {
	switch g := n.(type) {
	case *lg.Globally:
		return g.Environ
	case *lg.Eventually:
		return g.Environ
	}
	return nil
}

// EnvironStr returns the environment as a string or empty string if nil.
func EnvironStr(n lg.Expr) string {
	env := GetEnviron(n)
	if env != nil {
		return *env
	}
	return ""
}

// NormalProgramClone creates a shallow copy of a NormalProgram,
// allowing modification of specific fields.
func NormalProgramClone(np *NormalProgram) *NormalProgram {
	bindings := make([]*ActionTermBinding, len(np.Bindings))
	copy(bindings, np.Bindings)
	invars := make([]*ast.LabeledFormula, len(np.Invars))
	copy(invars, np.Invars)
	asms := make([]*ast.LabeledFormula, len(np.Asms))
	copy(asms, np.Asms)
	calls := make([]string, len(np.Calls))
	copy(calls, np.Calls)
	result := &NormalProgram{
		Base:     np.Base,
		Bindings: bindings,
		Init:     np.Init,
		Invars:   invars,
		Asms:     asms,
		Calls:    calls,
	}
	if np.Postconds != nil {
		result.Postconds = make(map[string][]*ast.LabeledFormula, len(np.Postconds))
		for k, v := range np.Postconds {
			result.Postconds[k] = v
		}
	}
	return result
}

// InvarianceTactic proves a "globally phi" formula using the invariance rule.
// It instruments all actions with property events and converts "M |= G phi" to "M |= true".
//
// Python: ivy_temporal.py:invariance_tactic (lines 257-393)
func InvarianceTactic(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, fmt.Errorf("invariance: no proof goals")
	}
	goal := goals[0]

	// Find the TemporalModels conclusion
	tm := findTemporalModels(goal)
	if tm == nil {
		return nil, fmt.Errorf("invariance: proof goal is not temporal")
	}

	fmla, _ := tm.Fmla.(lg.Expr)
	if fmla == nil {
		return nil, fmt.Errorf("invariance: could not extract formula from goal")
	}

	glob, ok := fmla.(*lg.Globally)
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
		model = &NormalProgram{Init: actions.NewSequence()}
	}

	// Add the invariant phi to the model's invariants. Python ivy_temporal.py:275:
	//   invars.append(ipr.clone_goal(goal,[],invar))
	// where clone_goal calls goal.clone_with_fresh_id([goal.label, invar]) — fresh
	// id is intentional, but the label must be the goal's label, not nil.
	model.Invars = append(model.Invars, goal.CloneWithFreshID([]ast.Node{goal.Label, invar}))

	// Collect assumed globally properties from prover axioms
	var gprops []lg.Expr
	var gpropLines []ast.Location
	if pc != nil {
		for _, ax := range pc.GetAxioms() {
			if !ax.Explicit && ax.IsTemporal() {
				if f, ok := ax.Formula.(lg.Expr); ok {
					if IsGprop(f) {
						gprops = append(gprops, f)
						gpropLines = append(gpropLines, ax.GetLineno())
					}
				}
			}
		}
	}

	// Add the negation of the property: F ~phi
	gprops = append(gprops, &lg.Eventually{Environ: glob.Environ, Body: &lg.Not{Body: invar}})
	gpropLines = append(gpropLines, goal.GetLineno())

	// Build memo tables: environ -> props, symbol -> props
	envprops := make(map[string][]lg.Expr)
	symprops := make(map[lg.NodeKey][]lg.Expr)
	propLines := make(map[string]ast.Location) // prop string -> lineno

	for i, prop := range gprops {
		env := EnvironStr(prop)
		envprops[env] = append(envprops[env], prop)
		propLines[fmt.Sprint(prop)] = gpropLines[i]
		for _, sym := range symbolsAst(prop) {
			symprops[lg.Key(sym)] = append(symprops[lg.Key(sym)], prop)
		}
	}

	// Build action map for callee resolution
	actionMap := make(map[string]*ActionTerm)
	for _, b := range model.Bindings {
		actionMap[b.Name] = b.Action
	}

	// instrStmt instruments a statement with property events
	var instrStmt func(stmt actions.Action, labels []string) actions.Action
	instrStmt = func(stmt actions.Action, labels []string) actions.Action {
		// Recur on sub-statements
		args := stmt.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		changed := false
		for i, a := range args {
			if sub, ok := a.(actions.Action); ok {
				newSub := instrStmt(sub, labels)
				newArgs[i] = newSub
				if newSub != sub {
					changed = true
				}
			} else {
				newArgs[i] = a
			}
		}
		var res actions.Action
		if changed {
			res = stmt.ActionClone(newArgs)
		} else {
			res = stmt
		}

		eventProps := make(map[string]lg.Expr) // deduped by string

		// If it is a call, check for events on return
		if call, ok := stmt.(*actions.CallAction); ok {
			calleeName := call.Callee.String()
			if at, ok := actionMap[calleeName]; ok {
				labelSet := make(map[string]bool)
				for _, l := range labels {
					labelSet[l] = true
				}
				for _, l := range at.Labels {
					if !labelSet[l] {
						for _, prop := range envprops[l] {
							eventProps[fmt.Sprint(prop)] = prop
						}
					}
				}
			}
		}

		// If a symbol is modified, add events for properties depending on it
		mods := actions.Modifies(stmt)
		labelSet := make(map[string]bool)
		for _, l := range labels {
			labelSet[l] = true
		}
		for _, sym := range mods {
			for _, prop := range symprops[lg.Key(sym)] {
				env := EnvironStr(prop)
				if !labelSet[env] {
					eventProps[fmt.Sprint(prop)] = prop
				}
			}
		}

		// Add property events
		var events []actions.Action
		for key, prop := range eventProps {
			loc := propLines[key]
			events = append(events, PropEvent(prop, loc))
		}

		res = actions.PostfixAction(res, events)
		actions.CopyFormalsTo(stmt, res)
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
				if f, ok := ax.Formula.(lg.Expr); ok {
					if g, ok := f.(*lg.Globally); ok {
						model.Asms = append(model.Asms, ax.Clone([]ast.Node{ax.Label, g.Body}).(*ast.LabeledFormula))
					}
				}
			}
		}
	}

	// Change conclusion to M |= true (Python line 385: conc = TemporalModels(model, il.And()))
	newConc := pc.GetAstCfg().NewTemporalModels(model, lg.True)

	// Build new goal.
	prems := proof.GoalPrems(goal)
	newGoal := proof.CloneGoal(pc.GetAstCfg(), goal, prems, newConc)

	result := make([]*ast.LabeledFormula, len(goals))
	result[0] = newGoal
	copy(result[1:], goals[1:])
	return result, nil
}

// findTemporalModels looks for a TemporalModels in the goal.
func findTemporalModels(goal *ast.LabeledFormula) *ast.TemporalModels {
	if goal == nil || goal.Formula == nil {
		return nil
	}
	if tm, ok := goal.Formula.(*ast.TemporalModels); ok {
		return tm
	}
	if sb, ok := goal.Formula.(*ast.SchemaBody); ok {
		if c := sb.Conc(); c != nil {
			if tm, ok := c.(*ast.TemporalModels); ok {
				return tm
			}
		}
	}
	return nil
}

// symbolsAst collects symbols from a logic node (wrapper for package access).
func symbolsAst(n lg.Expr) []*lg.Const {
	var result []*lg.Const
	symbolsAstRec(n, &result, make(map[lg.NodeKey]bool))
	return result
}

func symbolsAstRec(n lg.Expr, result *[]*lg.Const, seen map[lg.NodeKey]bool) {
	if c, ok := n.(*lg.Const); ok {
		if !seen[lg.Key(c)] {
			seen[lg.Key(c)] = true
			*result = append(*result, c)
		}
	}
	if app, ok := n.(*lg.Apply); ok {
		if c, ok := app.Func.(*lg.Const); ok {
			if !seen[lg.Key(c)] {
				seen[lg.Key(c)] = true
				*result = append(*result, c)
			}
		}
	}
	for _, child := range n.Children() {
		symbolsAstRec(child, result, seen)
	}
}

// RegisterTactics registers the invariance tactic on the given proof config.
// Replaces the old init()-based global registration.
func RegisterTactics(proofCfg *module.ProofConfig) {
	proofCfg.RegisterTactic("invariance", InvarianceTactic)
}
