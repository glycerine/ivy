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

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/proof"
)

// ActionTerm represents an action term with inputs, outputs, labels, and a body statement.
//
// An action term is of the form:
//
//	action (inputs) returns (outputs) { stmt }
type ActionTerm struct {
	Inputs  []*lg.Symbol
	Outputs []*lg.Symbol
	Labels  []string
	Stmt    actions.Action
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

// Clone creates a copy of the ActionTerm with a new statement but preserving
// all other attributes.
func (at *ActionTerm) Clone(stmt actions.Action) *ActionTerm {
	return &ActionTerm{
		Inputs:  at.Inputs,
		Outputs: at.Outputs,
		Labels:  at.Labels,
		Stmt:    stmt,
	}
}

// ActionTermBinding binds an action term to a name.
type ActionTermBinding struct {
	Name   string
	Action *ActionTerm
}

// String returns a human-readable representation of the binding.
func (b *ActionTermBinding) String() string {
	return fmt.Sprintf("%s = %s", b.Name, b.Action)
}

// Clone creates a copy of the binding with a new action.
func (b *ActionTermBinding) Clone(action *ActionTerm) *ActionTermBinding {
	return &ActionTermBinding{
		Name:   b.Name,
		Action: action,
	}
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
	Bindings  []*ActionTermBinding
	Init      actions.Action
	Invars    []*ast.LabeledFormula
	Asms      []*ast.LabeledFormula
	Calls     []string
	Postconds map[string][]*ast.LabeledFormula // optional postconditions
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
	for name, actIface := range mod.Actions {
		if act, ok := actIface.(actions.Action); ok {
			bindings = append(bindings, &ActionTermBinding{
				Name:   name,
				Action: OldActionToNew(act),
			})
		}
	}
	// Sort bindings by name for determinism
	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].Name < bindings[j].Name
	})

	// Build init from initializers
	var initNodes []lg.Expr
	for _, na := range mod.Initializers {
		if act, ok := na.Action.(actions.Action); ok {
			initNodes = append(initNodes, actions.WrapAction(act))
		}
	}
	var init actions.Action
	if len(initNodes) > 0 {
		init = actions.NewSequence(initNodes...)
	} else {
		init = actions.NewSequence()
	}

	// Sort public actions
	calls := make([]string, 0, len(mod.PublicActions))
	for name := range mod.PublicActions {
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
			actions.WrapAction(act.Stmt),
			actions.WrapAction(actions.NewReturnAction()),
		)
		ract.SetFormalParams(act.Inputs)
		ract.SetFormalReturns(act.Outputs)
		if strings.HasPrefix(name, "ext:") {
			name = name[4:]
		}
		ract.SetLabels([]string{name})
		branches = append(branches, actions.WrapAction(ract))
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
	return at.Clone(newStmt)
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
	return &NormalProgram{
		Bindings: bindings,
		Init:     np.Init,
		Invars:   invars,
		Asms:     asms,
		Calls:    calls,
	}
}

// InvarianceTactic proves a "globally phi" formula using the invariance rule.
// It instruments all actions with property events and converts "M |= G phi" to "M |= true".
//
// Python: ivy_temporal.py:invariance_tactic (lines 257-393)
func InvarianceTactic(pc *proof.ProofChecker, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
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

	// Get the model from the module
	var model *NormalProgram
	if CurrentModule != nil {
		model = NormalProgramClone(NormalProgramFromModule(CurrentModule))
	} else {
		model = &NormalProgram{Init: actions.NewSequence()}
	}

	// Add the invariant phi to the model's invariants
	model.Invars = append(model.Invars, &ast.LabeledFormula{Formula: invar})

	// Collect assumed globally properties from prover axioms
	var gprops []lg.Expr
	var gpropLines []ast.Location
	if pc != nil {
		for _, ax := range pc.Axioms {
			if !ax.Explicit && ax.Temporal {
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
	symprops := make(map[string][]lg.Expr)
	propLines := make(map[string]ast.Location) // prop string -> lineno

	for i, prop := range gprops {
		env := EnvironStr(prop)
		envprops[env] = append(envprops[env], prop)
		propLines[fmt.Sprint(prop)] = gpropLines[i]
		for _, sym := range symbolsAst(prop) {
			symprops[sym.Name] = append(symprops[sym.Name], prop)
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
		args := stmt.Args()
		newArgs := make([]lg.Expr, len(args))
		changed := false
		for i, a := range args {
			if sub := actions.UnwrapAction(a); sub != nil {
				newSub := instrStmt(sub, labels)
				newArgs[i] = actions.WrapAction(newSub)
				if newSub != sub {
					changed = true
				}
			} else {
				newArgs[i] = a
			}
		}
		var res actions.Action
		if changed {
			res = stmt.Clone(newArgs)
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
		for sym := range mods {
			for _, prop := range symprops[sym] {
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
		model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))
	}

	// Add assumed G-properties as model assumptions
	if pc != nil {
		for _, ax := range pc.Axioms {
			if !ax.Explicit && ax.Temporal {
				if f, ok := ax.Formula.(lg.Expr); ok {
					if g, ok := f.(*lg.Globally); ok {
						model.Asms = append(model.Asms, &ast.LabeledFormula{Formula: g.Body})
					}
				}
			}
		}
	}

	// Change conclusion to M |= true
	newConc := &ast.TemporalModels{Model: tm.Model, Fmla: wrapLogicAsAST(lg.True)}

	// Build new goal
	prems := proof.GoalPrems(goal)
	newGoal := cloneGoalWithASTConc(goal, prems, newConc)

	result := make([]*ast.LabeledFormula, len(goals))
	result[0] = newGoal
	copy(result[1:], goals[1:])
	return result, nil
}

// CurrentModule is set by the caller before invoking the tactic.
var CurrentModule *module.Module

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

// cloneGoalWithASTConc clones a goal with an ast.Node conclusion.
func cloneGoalWithASTConc(goal *ast.LabeledFormula, prems []ast.Node, conc ast.Node) *ast.LabeledFormula {
	var formula ast.Node
	if len(prems) > 0 {
		elems := make([]ast.Node, len(prems)+1)
		copy(elems, prems)
		elems[len(prems)] = conc
		formula = ast.NewSchemaBody(elems...)
	} else {
		formula = conc
	}
	return goal.CloneWithFreshID([]ast.Node{goal.Label, formula})
}

// logicASTAdapter wraps lg.Expr as ast.Node.
type logicASTAdapter struct {
	ast.Base
	Node lg.Expr
}

func (a *logicASTAdapter) Args() []ast.Node        { return nil }
func (a *logicASTAdapter) Clone([]ast.Node) ast.Node { return a }
func (a *logicASTAdapter) String() string           { return a.Node.String() }

func wrapLogicAsAST(n lg.Expr) ast.Node {
	if an, ok := n.(ast.Node); ok {
		return an
	}
	return &logicASTAdapter{Node: n}
}

// symbolsAst collects symbols from a logic node (wrapper for package access).
func symbolsAst(n lg.Expr) []*lg.Symbol {
	var result []*lg.Symbol
	symbolsAstRec(n, &result, make(map[string]bool))
	return result
}

func symbolsAstRec(n lg.Expr, result *[]*lg.Symbol, seen map[string]bool) {
	if c, ok := n.(*lg.Symbol); ok {
		if !seen[c.Name] {
			seen[c.Name] = true
			*result = append(*result, c)
		}
	}
	if app, ok := n.(*lg.Apply); ok {
		if c, ok := app.Func.(*lg.Symbol); ok {
			if !seen[c.Name] {
				seen[c.Name] = true
				*result = append(*result, c)
			}
		}
	}
	for _, child := range n.Children() {
		symbolsAstRec(child, result, seen)
	}
}

func init() {
	proof.RegisterTactic("invariance", InvarianceTactic)
}
