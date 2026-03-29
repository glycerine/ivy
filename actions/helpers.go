package actions

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/xtracer"
)

// shortTypeName returns just the struct name without package prefix or pointer star,
// matching Python's type(x).__name__ output.
func shortTypeName(v interface{}) string {
	s := fmt.Sprintf("%T", v)
	if i := strings.LastIndex(s, "."); i >= 0 {
		s = s[i+1:]
	}
	// Map Go type names to Python class names where they differ.
	// Python: Var, Const; Go: Variable, Symbol.
	switch s {
	case "Variable":
		return "Var"
	case "Symbol":
		return "Const"
	}
	return s
}

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
	sortStrings(names)
	return names
}

// sortStrings sorts a string slice in place.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
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
func ParamsToStr(params []*lg.Symbol) string {
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
	formals1 := make([]*lg.Symbol, 0, len(fp1)+len(fr1))
	formals1 = append(formals1, fp1...)
	formals1 = append(formals1, fr1...)
	formals2 := make([]*lg.Symbol, 0, len(fp2)+len(fr2))
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
	if ab, ok := action2.(interface{ SetLabels([]string) }); ok {
		_ = ab // labels handled by CopyFormalsTo
	}
	xtracer.Trace("actions.apply_mixin EXIT")
	return res
}

// SubstituteConstantsAction recursively applies a constant substitution
// to all lg.Expr children of an action. Matches Python's
// substitute_constants_ast applied to action nodes.
//
// Python's substitute_constants_ast always clones every non-leaf node
// and recurses into LabeledFormula children (label, formula). We
// replicate both behaviors here.
func SubstituteConstantsAction(action Action, subs map[lg.NodeKey]lg.Expr) Action {
	// No early return for empty subs — Python's substitute_constants_ast
	// has no such guard. It always traverses and clones, which is needed
	// to keep UniqueID counters (CallAction, ChoiceAction, LocalAction)
	// in sync between Go and Python.
	xtracer.Trace("actions.substitute_constants_action ENTER type=%s nargs=%d", shortTypeName(action), len(action.ActionArgs()))
	args := action.ActionArgs()
	newArgs := make([]lg.Expr, len(args))

	// Step 1: Handle LabeledFormula FIRST, matching Python's recursion order.
	// In Python, AssumeAction.args = [LabeledFormula]. substitute_constants_ast
	// sees args[0] = LF and recurses into lf.args = [label, formula] BEFORE
	// processing any other children. We must do the same to produce traces
	// in the same depth-first order.
	var clonedLF *ast.LabeledFormula
	if bearer, ok := action.(LFBearer); ok {
		if lf := bearer.GetLF(); lf != nil {
			// Emit trace matching Python entering the LF node
			xtracer.Trace("actions.substitute_constants_action ENTER type=LabeledFormula nargs=%d", len(lf.Args()))

			// Process lf.args in order: [label, formula] — matching Python's
			// substitute_constants_ast which recurses into lf.args[0] (label)
			// first, then lf.args[1] (formula).

			// Substitute in label (Python: substitute_constants_ast(label, subs))
			newLabel := substituteConstantsNode(lf.Label, subs)
			// Substitute in formula (args[0] is the unwrapped formula)
			newFormula := substituteConstantsExpr(args[0], subs)

			// Clone LF with substituted children — triggers LF.clone PRESERVE
			clonedLF = lf.Clone([]ast.Node{newLabel, newFormula}).(*ast.LabeledFormula)

			// Use the substituted formula as newArgs[0]
			newArgs[0] = newFormula
		}
	}

	// Step 2: Process remaining ActionArgs (skip index 0 if LF handled it)
	for i, arg := range args {
		if clonedLF != nil && i == 0 {
			continue // already handled by LF processing above
		}
		if child, ok := arg.(Action); ok {
			newArgs[i] = SubstituteConstantsAction(child, subs)
		} else if arg != nil {
			newArgs[i] = substituteConstantsExpr(arg, subs)
		} else {
			newArgs[i] = arg
		}
	}

	// Step 3: Always clone action — Python always calls ast.clone(new_args)
	result := action.ActionClone(newArgs)

	// Step 4: Set cloned LF on result
	if clonedLF != nil {
		result.(LFBearer).SetLF(clonedLF)
	}

	return result
}

// substituteConstantsExpr applies constant substitution to an lg.Expr.
// Matches Python's substitute_constants_ast from ivy_logic_utils.py:172.
// Python: is_constant(ast) checks isinstance(ast, lg.Const). Only constants
// get the subs.get(rep, ast) short-circuit. Everything else (including Var)
// falls through to recurse + clone + trace.
func substituteConstantsExpr(expr lg.Expr, subs map[lg.NodeKey]lg.Expr) lg.Expr {
	// Python: if is_constant(ast): return subs.get(ast.rep, ast)
	// is_constant checks isinstance(term, lg.Const). Constants are lg.Symbol
	// in Go, Variables are lg.Variable. Only constants short-circuit.
	if sym, ok := expr.(*lg.Symbol); ok {
		// This is a constant (Python lg.Const). Short-circuit: lookup in subs.
		return substituteLookup(sym, subs)
	}
	// lg.Variable (Python lg.Var) is NOT a constant — falls through to trace+clone.
	children := expr.Children()
	xtracer.Trace("actions.substitute_constants_action ENTER type=%s nargs=%d", shortTypeName(expr), len(children))
	if len(children) == 0 {
		// Leaf non-constant (e.g. Var): Python traces then clones with empty args.
		return expr
	}
	// Python ALWAYS clones: ast.clone(substitute_constants_ast(x,subs) for x in ast.args)
	// No "changed" optimization — every non-leaf gets cloned.
	newChildren := make([]lg.Expr, len(children))
	for i, c := range children {
		newChildren[i] = substituteConstantsExpr(c, subs)
	}
	return cloneExpr(expr, newChildren)
}

// cloneExpr creates a shallow copy of an expression with new children.
func cloneExpr(expr lg.Expr, children []lg.Expr) lg.Expr {
	switch t := expr.(type) {
	case *lg.Apply:
		if len(children) > 0 {
			result, err := lg.NewApply(children[0], children[1:]...)
			if err != nil {
				return &lg.Apply{Func: children[0], Terms: children[1:]}
			}
			return result
		}
		return t
	case *lg.And:
		a, _ := lg.NewAnd(children...)
		if a != nil {
			return a
		}
		return &lg.And{Terms: children}
	case *lg.Or:
		o, _ := lg.NewOr(children...)
		if o != nil {
			return o
		}
		return &lg.Or{Terms: children}
	case *lg.Not:
		if len(children) > 0 {
			return &lg.Not{Body: children[0]}
		}
		return t
	case *lg.Implies:
		if len(children) >= 2 {
			return &lg.Implies{T1: children[0], T2: children[1]}
		}
		return t
	case *lg.Iff:
		if len(children) >= 2 {
			return &lg.Iff{T1: children[0], T2: children[1]}
		}
		return t
	case *lg.Eq:
		if len(children) >= 2 {
			return &lg.Eq{T1: children[0], T2: children[1]}
		}
		return t
	case *lg.ForAll:
		if len(children) > 0 {
			fa, _ := lg.NewForAll(t.Variables, children[0])
			return fa
		}
		return t
	case *lg.Exists:
		if len(children) > 0 {
			ex, _ := lg.NewExists(t.Variables, children[0])
			return ex
		}
		return t
	default:
		return expr
	}
}

// substituteLookup implements Python's is_constant short-circuit:
// if the symbol is in subs, return the replacement; otherwise return as-is.
// No trace is emitted (Python's is_constant returns True for lg.Const).
func substituteLookup(sym *lg.Symbol, subs map[lg.NodeKey]lg.Expr) lg.Expr {
	if rep, found := subs[lg.Key(sym)]; found {
		return rep
	}
	return sym
}

// substituteConstantsNode applies constant substitution to an ast.Node,
// matching Python's substitute_constants_ast for non-Action, non-Expr nodes
// (e.g. label Atoms inside LabeledFormulas). Python's is_constant(x) checks
// isinstance(x, lg.Const); everything else recurses into x.args and clones.
func substituteConstantsNode(node ast.Node, subs map[lg.NodeKey]lg.Expr) ast.Node {
	if node == nil {
		return nil
	}
	// If it's an lg.Expr, delegate to substituteConstantsExpr
	if expr, ok := node.(lg.Expr); ok {
		return substituteConstantsExpr(expr, subs)
	}
	// For plain ast.Node (like ast.Atom labels): recurse into Args and clone.
	// Python: substitute_constants_ast enters the else branch for non-constants.
	args := node.Args()
	if len(args) == 0 {
		// Leaf node with no children — Python would check is_constant (which
		// only matches lg.Const). ast.Atom is NOT lg.Const, so Python enters
		// else branch, traces, and clones with empty args.
		xtracer.Trace("actions.substitute_constants_action ENTER type=%s nargs=%d", shortTypeName(node), len(args))
		return node.Clone(args)
	}
	xtracer.Trace("actions.substitute_constants_action ENTER type=%s nargs=%d", shortTypeName(node), len(args))
	newArgs := make([]ast.Node, len(args))
	for i, a := range args {
		newArgs[i] = substituteConstantsNode(a, subs)
	}
	return node.Clone(newArgs)
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
}
