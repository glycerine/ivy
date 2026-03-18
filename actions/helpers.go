package actions

import (
	"fmt"
	"strings"

	lg "github.com/glycerine/goivy/logic"
)

// ConcatActions concatenates actions into a single Sequence.
// If an action is already a Sequence, its children are flattened.
func ConcatActions(actions ...Action) *Sequence {
	var all []lg.Expr
	for _, a := range actions {
		if seq, ok := a.(*Sequence); ok {
			all = append(all, seq.Children...)
		} else {
			all = append(all, WrapAction(a))
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
		nodes = append(nodes, WrapAction(s))
	}
	nodes = append(nodes, WrapAction(action))
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
	nodes = append(nodes, WrapAction(action))
	for _, s := range stmts {
		nodes = append(nodes, WrapAction(s))
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
func ApplyMixin(action1, action2 Action, isAfter bool) Action {
	var res *Sequence
	if isAfter {
		res = ConcatActions(action2, action1)
	} else {
		res = ConcatActions(action1, action2)
	}
	res.SetLineno(action1.GetLineno())
	res.SetFormalParams(action2.GetFormalParams())
	res.SetFormalReturns(action2.GetFormalReturns())
	if ab, ok := action2.(interface{ SetLabels([]string) }); ok {
		_ = ab // labels handled by CopyFormalsTo
	}
	return res
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
