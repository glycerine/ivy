package goivy

import (
	"fmt"

	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// ConcatActions concatenates actions into a single Sequence.
// If an action is already a Sequence, its children are flattened.
func ConcatActions(actions ...ActionsAction) *LogicSequence {
	var all []Expr
	for _, a := range actions {
		if seq, ok := a.(*LogicSequence); ok {
			all = append(all, seq.Elems...)
		} else {
			all = append(all, a)
		}
	}
	return NewSequence(all...)
}

// HasCode returns true if the action contains any non-Sequence subaction,
// indicating it has actual executable content.
func HasCode(action ActionsAction) bool {
	for _, a := range action.IterSubactions() {
		if _, isSeq := a.(*LogicSequence); !isSeq {
			return true
		}
	}
	return false
}

// CallSetRec recursively collects all action names reachable from actionName.
func CallSetRec(actionName string, env map[string]ActionsAction, res map[string]bool) {
	if res[actionName] {
		return
	}
	res[actionName] = true
	if act, ok := env[actionName]; ok {
		for _, c := range act.IterCalls() {
			CallSetRec(c, env, res)
		}
	}
}

// CallSet returns a sorted list of all action names reachable from actionName.
func CallSet(actionName string, env map[string]ActionsAction) []string {
	res := make(map[string]bool)
	CallSetRec(actionName, env, res)
	names := make([]string, 0, len(res))
	for n := range res {
		names = append(names, n)
	}
	// Sort for deterministic output.
	sort.Strings(names)
	return names
}

// PrefixAction adds a list of statements at the beginning of an action.
func PrefixAction(action ActionsAction, stmts []ActionsAction) ActionsAction {
	// Python prefix_action has NO empty check — always wraps in Sequence.
	// Do NOT add `if len(stmts) == 0 { return action }` here.
	nodes := make([]Expr, 0, len(stmts)+1)
	for _, s := range stmts {
		nodes = append(nodes, s)
	}
	nodes = append(nodes, action)
	res := NewSequence(nodes...)
	CopyFormalsTo(action, res)
	if action.GetLineno().Line != 0 || action.GetLineno().Filename != "" {
		res.SetLineno(action.GetLineno())
	}
	return res
}

// PostfixAction adds a list of statements at the end of an action.
func PostfixAction(action ActionsAction, stmts []ActionsAction) ActionsAction {
	if len(stmts) == 0 {
		return action
	}
	nodes := make([]Expr, 0, 1+len(stmts))
	nodes = append(nodes, action)
	for _, s := range stmts {
		nodes = append(nodes, s)
	}
	res := NewSequence(nodes...)
	CopyFormalsTo(action, res)
	if action.GetLineno().Line != 0 || action.GetLineno().Filename != "" {
		res.SetLineno(action.GetLineno())
	}
	return res
}

// ParamsToStr formats a list of formal parameters as "(name:sort, ...)".
func ParamsToStr(params []*Const) string {
	parts := make([]string, len(params))
	for i, p := range params {
		name := p.Name
		if strings.HasPrefix(name, "fml:") {
			name = name[4:]
		}
		parts[i] = fmt.Sprintf("%s:%s", name, p.CSort)
	}
	return "(" + strings.Join(parts, ",") + ")"
}

// ActionDefToStr formats an action definition for display.
func ActionDefToStr(name string, action ActionsAction) string {
	res := "action " + name
	if fp := action.GetFormalParams(); len(fp) > 0 {
		res += ParamsToStr(fp)
	}
	if fr := action.GetFormalReturns(); len(fr) > 0 {
		res += " returns" + ParamsToStr(fr)
	}
	res += " = "
	if _, ok := action.(*LogicSequence); ok {
		res += action.String()
	} else {
		res += "{" + action.String() + "}"
	}
	return res
}

// ApplyMixin combines two actions as a mixin. If isAfter is true,
// action1 is appended after action2; otherwise it is prepended before.
//
// Matches Python ivy_actions.py:1338-1367 apply_mixin:
// validates param/return counts and sorts, substitutes action1's formals
// to match action2's, then concatenates.
func ApplyMixin(action1, action2 ActionsAction, isAfter bool) ActionsAction {
	xtracer.Trace("actions.apply_mixin ENTER")
	fp1, fp2 := action1.GetFormalParams(), action2.GetFormalParams()
	fr1, fr2 := action1.GetFormalReturns(), action2.GetFormalReturns()
	xtracer.Trace("actions.apply_mixin fp1=%d fp2=%d fr1=%d fr2=%d isAfter=%v",
		len(fp1), len(fp2), len(fr1), len(fr2), isAfter)

	// Python: raise IvyError on param/return count or sort mismatch.
	if len(fp1) != len(fp2) {
		xtracer.Trace("actions.apply_mixin EARLY_RETURN fp_mismatch")
		panic(fmt.Sprintf("mixin has wrong number of input parameters: %d vs %d", len(fp1), len(fp2)))
	}
	if len(fr1) != len(fr2) {
		xtracer.Trace("actions.apply_mixin EARLY_RETURN fr_mismatch")
		panic(fmt.Sprintf("mixin has wrong number of output parameters: %d vs %d", len(fr1), len(fr2)))
	}

	formals1 := make([]*Const, 0, len(fp1)+len(fr1))
	formals1 = append(formals1, fp1...)
	formals1 = append(formals1, fr1...)
	formals2 := make([]*Const, 0, len(fp2)+len(fr2))
	formals2 = append(formals2, fp2...)
	formals2 = append(formals2, fr2...)

	for i, x := range formals1 {
		y := formals2[i]
		if x.CSort != nil && y.CSort != nil && SortKey(x.CSort) != SortKey(y.CSort) {
			xtracer.Trace("actions.apply_mixin EARLY_RETURN sort_mismatch param=%s", x.Name)
			panic(fmt.Sprintf("parameter %s of mixin has wrong sort", x.Name))
		}
	}

	// Build substitution: formals1 -> formals2
	subs := make(map[NodeKey]Expr, len(formals1))
	for i, f1 := range formals1 {
		subs[Key(f1)] = formals2[i]
	}
	xtracer.Trace("actions.apply_mixin subs=%d", len(subs))

	// Apply substitution to action1
	action1Renamed := SubstituteConstantsAction(action1, subs)

	var res *LogicSequence
	if isAfter {
		res = ConcatActions(action2, action1Renamed)
	} else {
		res = ConcatActions(action1Renamed, action2)
	}
	res.SetLineno(action1.GetLineno())
	res.SetFormalParams(action2.GetFormalParams())
	res.SetFormalReturns(action2.GetFormalReturns())
	if labeler, ok := action2.(interface{ GetLabels() []string }); ok {
		if labels := labeler.GetLabels(); labels != nil {
			res.SetLabels(labels)
		}
	}
	xtracer.Trace("actions.apply_mixin EXIT")
	return res
}

// SubstituteConstantsAction is the entry point for callers expecting Action return type.
// Delegates to module.SubstituteConstantsAST.
func SubstituteConstantsAction(action ActionsAction, subs map[NodeKey]Expr) ActionsAction {
	return SubstituteConstantsAST(action, subs).(ActionsAction)
}

// AppendToAction appends action2 at the end of action1, preserving
// action1's formals and labels.
func AppendToAction(action1, action2 ActionsAction) ActionsAction {
	res := ConcatActions(action1, action2)
	res.SetLineno(action1.GetLineno())
	if fp := action1.GetFormalParams(); fp != nil {
		res.SetFormalParams(fp)
	}
	if fr := action1.GetFormalReturns(); fr != nil {
		res.SetFormalReturns(fr)
	}
	if labeler, ok := action1.(interface{ GetLabels() []string }); ok {
		if labels := labeler.GetLabels(); labels != nil {
			res.SetLabels(labels)
		}
	}
	return res
}

// CopyFormalsTo copies formal parameters, returns, and labels
// from src to dst, provided src is an Action. This is a convenience
// wrapper around ActionBase.CopyFormalsTo.
func CopyFormalsTo(src, dst ActionsAction) {
	// Python: if not isinstance(res, EnvAction): copy params/returns
	if _, isEnvAction := dst.(*LogicEnvAction); !isEnvAction {
		if fp := src.GetFormalParams(); fp != nil {
			dst.SetFormalParams(fp)
		}
		if fr := src.GetFormalReturns(); fr != nil {
			dst.SetFormalReturns(fr)
		}
	}
	if labeler, ok := src.(interface{ GetLabels() []string }); ok {
		if labels := labeler.GetLabels(); labels != nil {
			if setter, ok2 := dst.(interface{ SetLabels([]string) }); ok2 {
				setter.SetLabels(labels)
			}
		}
	}
}
