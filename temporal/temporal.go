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
	Invars    []*module.LabeledFormula
	Asms      []*module.LabeledFormula
	Calls     []string
	Postconds map[string][]*module.LabeledFormula // optional postconditions
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
	var initNodes []lg.Node
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
	var branches []lg.Node
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
func PropEvent(gprop lg.Node, lineno ast.Location) actions.Action {
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
func IsGloballyFormula(n lg.Node) bool {
	_, ok := n.(*lg.Globally)
	return ok
}

// IsEventuallyFormula returns true if the given node is a logic.Eventually formula.
func IsEventuallyFormula(n lg.Node) bool {
	_, ok := n.(*lg.Eventually)
	return ok
}

// IsTemporalFormula returns true if the formula is a temporal operator.
func IsTemporalFormula(n lg.Node) bool {
	switch n.(type) {
	case *lg.Globally, *lg.Eventually, *lg.WhenOperator:
		return true
	}
	return false
}

// HasTemporalOperator returns true if the formula or any subformula
// contains a temporal operator.
func HasTemporalOperator(n lg.Node) bool {
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
func IsGprop(n lg.Node) bool {
	g, ok := n.(*lg.Globally)
	if !ok {
		return false
	}
	return !HasTemporalOperator(g.Body)
}

// GetEnviron returns the environment label of a temporal formula, or nil.
func GetEnviron(n lg.Node) *string {
	switch g := n.(type) {
	case *lg.Globally:
		return g.Environ
	case *lg.Eventually:
		return g.Environ
	}
	return nil
}

// EnvironStr returns the environment as a string or empty string if nil.
func EnvironStr(n lg.Node) string {
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
	invars := make([]*module.LabeledFormula, len(np.Invars))
	copy(invars, np.Invars)
	asms := make([]*module.LabeledFormula, len(np.Asms))
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
