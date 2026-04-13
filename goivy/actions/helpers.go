package actions

import (
	"fmt"
	"os"
	"sort"
	"strings"

	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// shortTypeName delegates to module.ShortTypeName.
//func shortTypeName(v interface{}) string { return iu.ShortTypeName(v) }

// ShortTypeName delegates to module.ShortTypeName.
//func ShortTypeName(v interface{}) string { return iu.ShortTypeName(v) }

// ConcatActions concatenates actions into a single Sequence.
// If an action is already a Sequence, its children are flattened.
func ConcatActions(actions ...Action) *Sequence {
	var all []lg.Expr
	for _, a := range actions {
		if seq, ok := a.(*Sequence); ok {
			all = append(all, seq.Elems...)
		} else {
			all = append(all, a)
		}
	}
	return NewSequence(all...)
}

// HasCode returns true if the action contains any non-Sequence subaction,
// indicating it has actual executable content.
func HasCode(action Action) bool {
	for _, a := range action.IterSubactions() {
		if _, isSeq := a.(*Sequence); !isSeq {
			return true
		}
	}
	return false
}

// CallSetRec recursively collects all action names reachable from actionName.
func CallSetRec(actionName string, env map[string]Action, res map[string]bool) {
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
func CallSet(actionName string, env map[string]Action) []string {
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
func PrefixAction(action Action, stmts []Action) Action {
	nodes := make([]lg.Expr, 0, len(stmts)+1)
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
func PostfixAction(action Action, stmts []Action) Action {
	if len(stmts) == 0 {
		return action
	}
	nodes := make([]lg.Expr, 0, 1+len(stmts))
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
func ParamsToStr(params []*lg.Const) string {
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
func ActionDefToStr(name string, action Action) string {
	res := "action " + name
	if fp := action.GetFormalParams(); len(fp) > 0 {
		res += ParamsToStr(fp)
	}
	if fr := action.GetFormalReturns(); len(fr) > 0 {
		res += " returns" + ParamsToStr(fr)
	}
	res += " = "
	if _, ok := action.(*Sequence); ok {
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
func ApplyMixin(action1, action2 Action, isAfter bool) Action {
	xtracer.Trace("actions.apply_mixin ENTER")
	fp1, fp2 := action1.GetFormalParams(), action2.GetFormalParams()
	fr1, fr2 := action1.GetFormalReturns(), action2.GetFormalReturns()
	xtracer.Trace("actions.apply_mixin fp1=%d fp2=%d fr1=%d fr2=%d isAfter=%v",
		len(fp1), len(fp2), len(fr1), len(fr2), isAfter)

	// Validate param/return counts match.
	// Python: raise IvyError (caught upstream, compilation continues).
	// We skip the mixin and return action2 unchanged instead of panicking.
	if len(fp1) != len(fp2) {
		xtracer.Trace("actions.apply_mixin EARLY_RETURN fp_mismatch")
		fmt.Fprintf(os.Stderr, "warning: mixin has wrong number of input parameters: %d vs %d, skipping\n", len(fp1), len(fp2))
		return action2
	}
	if len(fr1) != len(fr2) {
		xtracer.Trace("actions.apply_mixin EARLY_RETURN fr_mismatch")
		fmt.Fprintf(os.Stderr, "warning: mixin has wrong number of output parameters: %d vs %d, skipping\n", len(fr1), len(fr2))
		return action2
	}

	// Build combined formals lists and validate sorts match
	formals1 := make([]*lg.Const, 0, len(fp1)+len(fr1))
	formals1 = append(formals1, fp1...)
	formals1 = append(formals1, fr1...)
	formals2 := make([]*lg.Const, 0, len(fp2)+len(fr2))
	formals2 = append(formals2, fp2...)
	formals2 = append(formals2, fr2...)

	for i, x := range formals1 {
		y := formals2[i]
		if x.CSort != nil && y.CSort != nil && lg.SortKey(x.CSort) != lg.SortKey(y.CSort) {
			xtracer.Trace("actions.apply_mixin EARLY_RETURN sort_mismatch param=%s", x.Name)
			fmt.Fprintf(os.Stderr, "warning: parameter %s of mixin has wrong sort, skipping\n", x.Name)
			return action2
		}
	}

	// Build substitution: formals1 -> formals2
	subs := make(map[lg.NodeKey]lg.Expr, len(formals1))
	for i, f1 := range formals1 {
		subs[lg.Key(f1)] = formals2[i]
	}
	xtracer.Trace("actions.apply_mixin subs=%d", len(subs))

	// Apply substitution to action1
	action1Renamed := SubstituteConstantsAction(action1, subs)

	var res *Sequence
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
func SubstituteConstantsAction(action Action, subs map[lg.NodeKey]lg.Expr) Action {
	return module.SubstituteConstantsAST(action, subs).(Action)
}

// AppendToAction appends action2 at the end of action1, preserving
// action1's formals and labels.
func AppendToAction(action1, action2 Action) Action {
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
func CopyFormalsTo(src, dst Action) {
	if fp := src.GetFormalParams(); fp != nil {
		dst.SetFormalParams(fp)
	}
	if fr := src.GetFormalReturns(); fr != nil {
		dst.SetFormalReturns(fr)
	}
	if labeler, ok := src.(interface{ GetLabels() []string }); ok {
		if labels := labeler.GetLabels(); labels != nil {
			if setter, ok2 := dst.(interface{ SetLabels([]string) }); ok2 {
				setter.SetLabels(labels)
			}
		}
	}
}
