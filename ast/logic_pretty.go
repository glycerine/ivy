package ast

import (
	"fmt"
	"strings"
)

// PrettyFmla formats a logic Expr using Python Ivy's infix notation.
// This replicates the behavior of ivy_logic.py's pretty_fmla function,
// which calls drop_annotations(False, set()) then ugly(0).
//
// For conformance testing, all user-facing formula display should use
// this function instead of Expr.String().
func PrettyFmla(n Expr) string {
	d := dropAnnotations(n, false, make(map[string]bool))
	return ugly(d, 0)
}

// PrettyFmlaAmbiguous formats without any sort annotations.
// Matches Python's fmla_to_str_ambiguous.
func PrettyFmlaAmbiguous(n Expr) string {
	return ugly(n, 0)
}

// --- ugly formatting (matches ivy_logic.py lines 1268-1344) ---

var infixSymbols = map[string]bool{
	"<": true, "<=": true, ">": true, ">=": true,
	"+": true, "-": true, "*": true, "/": true,
}

var precSymbols = map[string]int{
	"<": 7, "<=": 7, ">": 7, ">=": 7,
	"+": 12, "-": 13, "*": 14, "/": 15,
}

// ugly formats a node with precedence-based infix notation.
// Matches Python ivy_logic.py ugly methods.
func ugly(n Expr, prec int) string {
	switch t := n.(type) {
	case *Variable:
		return varUgly(t, prec)
	case *Symbol:
		return constUgly(t, prec)
	case *Apply:
		return appUgly(t, prec)
	case *Eq:
		return naryUgly("=", []Expr{t.T1, t.T2}, 7, prec)
	case *Not:
		return notUgly(t, prec)
	case *And:
		if len(t.Terms) == 0 {
			return "true"
		}
		return naryParen("&", t.Terms, 5, prec)
	case *Or:
		if len(t.Terms) == 0 {
			return "false"
		}
		return naryUgly("|", t.Terms, 4, prec)
	case *Implies:
		return naryUgly("->", []Expr{t.T1, t.T2}, 3, prec)
	case *Iff:
		return naryUgly("<->", []Expr{t.T1, t.T2}, 3, prec)
	case *Ite:
		// Python: '({} if {} else {})'.format(then.ugly(9), cond.ugly(9), else.ugly(9))
		return fmt.Sprintf("(%s if %s else %s)",
			ugly(t.Then, 9), ugly(t.Cond, 9), ugly(t.Else, 9))
	case *Cond:
		return fmt.Sprintf("(%s => %s)", ugly(t.T1, 9), ugly(t.T2, 9))
	case *Globally:
		// Python uses unicode □ but ugly_environ returns '' always
		return fmt.Sprintf("\u25A1 %s", ugly(t.Body, 2))
	case *Eventually:
		return fmt.Sprintf("\u2B26 %s", ugly(t.Body, 2))
	case *WhenOperator:
		return naryUgly("when"+t.Name, []Expr{t.T1, t.T2}, 2, prec)
	case *ForAll:
		return quantUgly("forall", t.Variables, t.Body, prec)
	case *Exists:
		return quantUgly("exists", t.Variables, t.Body, prec)
	case *Lambda:
		return quantUgly("lambda", t.Variables, t.Body, prec)
	case *NamedBinder:
		return quantUgly("$"+t.Name, t.Variables, t.Body, prec)
	case *LogicDefinition:
		// Python: LogicDefinition.ugly = nary_ugly('=', self.args, 7, prec)
		return naryUgly("=", []Expr{t.Lhs, t.Rhs}, 7, prec)
	default:
		return fmt.Sprint(n)
	}
}

// varUgly matches Python lg.Variable.ugly.
// Shows sort annotation for variables whose sort is concrete (not TopSort/SortVar).
func varUgly(v *Variable, prec int) string {
	if v.VSort != nil {
		if _, isTop := v.VSort.(*TopSort); !isTop {
			// Show sort annotation: "X:sortname"
			return v.Name + ":" + sortName(v.VSort)
		}
	}
	return v.Name
}

// constUgly matches Python lg.Symbol.ugly.
// Shows sort annotation only for numerals with concrete sorts.
func constUgly(c *Symbol, prec int) string {
	if isNumeralName(c.Name) {
		if _, isTop := c.CSort.(*TopSort); !isTop {
			return c.Name + ":" + sortName(c.CSort)
		}
	}
	return c.Name
}

// appUgly matches Python app_ugly (ivy_logic.py:1268-1289).
func appUgly(a *Apply, prec int) string {
	var name string
	if nb, ok := a.Func.(*NamedBinder); ok {
		name = PrettyFmla(nb)
	} else if c, ok := a.Func.(*Symbol); ok {
		name = c.Name
	} else if v, ok := a.Func.(*Variable); ok {
		name = v.Name
	} else {
		name = fmt.Sprint(a.Func)
	}

	myprec := 1
	if p, ok := precSymbols[name]; ok {
		myprec = p
	}
	infix := infixSymbols[name]

	var args []string
	if infix && len(a.Terms) == 2 {
		args = []string{
			ugly(a.Terms[0], myprec-1),
			ugly(a.Terms[1], myprec),
		}
	} else {
		args = make([]string, len(a.Terms))
		for i, t := range a.Terms {
			args[i] = ugly(t, myprec)
		}
	}

	if infix {
		if len(args) == 1 && name == "-" {
			res := name + args[0]
			return res
		}
		res := strings.Join(args, " "+name+" ")
		if myprec <= prec {
			res = "(" + res + ")"
		}
		return res
	}

	if len(args) == 0 {
		return name
	}
	return name + "(" + strings.Join(args, ",") + ")"
}

// notUgly matches Python lg.Not.ugly.
func notUgly(n *Not, prec int) string {
	if eq, ok := n.Body.(*Eq); ok {
		// Not(Eq(a,b)) → "a ~= b"
		return naryUgly("~=", []Expr{eq.T1, eq.T2}, 8, prec)
	}
	return "~" + ugly(n.Body, 6)
}

// naryUgly matches Python nary_ugly (ivy_logic.py:1291-1297).
func naryUgly(op string, args []Expr, myprec, prec int) string {
	var uargs []string
	if len(args) == 2 {
		uargs = []string{
			ugly(args[0], myprec-1),
			ugly(args[1], myprec),
		}
	} else {
		uargs = make([]string, len(args))
		for i, a := range args {
			uargs[i] = ugly(a, myprec)
		}
	}
	res := strings.Join(uargs, " "+op+" ")
	if len(args) > 1 && myprec <= prec {
		return "(" + res + ")"
	}
	return res
}

// naryParen matches Python nary_paren (ivy_logic.py:1299-1302).
// Always wraps in parens (used for And).
func naryParen(op string, args []Expr, myprec, prec int) string {
	uargs := make([]string, len(args))
	for i, a := range args {
		uargs[i] = ugly(a, myprec)
	}
	res := strings.Join(uargs, " "+op+" ")
	return "(" + res + ")"
}

// quantUgly matches Python quant_ugly (ivy_logic.py:1332-1341).
func quantUgly(keyword string, vars []*Variable, body Expr, prec int) string {
	vparts := make([]string, len(vars))
	for i, v := range vars {
		vparts[i] = ugly(v, 1)
	}
	res := keyword + " " + strings.Join(vparts, ",") + ". " + ugly(body, 1)
	if prec >= 1 {
		res = "(" + res + ")"
	}
	return res
}

// --- drop_annotations (matches ivy_logic.py lines 1352-1426) ---

// dropAnnotations removes sort annotations from variables and constants
// where the sort can be inferred. This matches Python's pretty_fmla
// which calls drop_annotations(False, set()) before ugly(0).
func dropAnnotations(n Expr, inferredSort bool, annotatedVars map[string]bool) Expr {
	switch t := n.(type) {
	case *Variable:
		if inferredSort || annotatedVars[t.Name] {
			annotatedVars[t.Name] = true
			return &Variable{Name: t.Name, VSort: TopS}
		}
		if _, isTop := t.VSort.(*TopSort); !isTop {
			annotatedVars[t.Name] = true
		}
		return t

	case *Symbol:
		if inferredSort && isNumeralName(t.Name) {
			return NewSymbol(t.Name, TopS)
		}
		return t

	case *Apply:
		name := ""
		if c, ok := t.Func.(*Symbol); ok {
			name = c.Name
		}
		if isPolymorphicSymbolName(name) {
			arg0 := dropAnnotations(t.Terms[0], inferredSort && !SortEqual(t.aSort, Boolean), annotatedVars)
			rest := make([]Expr, len(t.Terms)-1)
			for i, a := range t.Terms[1:] {
				rest[i] = dropAnnotations(a, name != "*>", annotatedVars)
			}
			newTerms := append([]Expr{arg0}, rest...)
			result, _ := NewApply(t.Func, newTerms...)
			return result
		}
		newTerms := make([]Expr, len(t.Terms))
		for i, a := range t.Terms {
			newTerms[i] = dropAnnotations(a, true, annotatedVars)
		}
		result, _ := NewApply(t.Func, newTerms...)
		return result

	case *Eq:
		a0 := dropAnnotations(t.T1, false, annotatedVars)
		a1 := dropAnnotations(t.T2, true, annotatedVars)
		result, _ := NewEq(a0, a1)
		return result

	case *Ite:
		a1 := dropAnnotations(t.Then, inferredSort, annotatedVars)
		a2 := dropAnnotations(t.Else, true, annotatedVars)
		a0 := dropAnnotations(t.Cond, true, annotatedVars)
		result, _ := NewIte(a0, a1, a2)
		return result

	case *Cond:
		a1 := dropAnnotations(t.T2, inferredSort, annotatedVars)
		a0 := dropAnnotations(t.T1, true, annotatedVars)
		result, _ := NewCond(a0, a1)
		return result

	case *ForAll:
		vars := make([]*Variable, len(t.Variables))
		for i, v := range t.Variables {
			dv := dropAnnotations(v, false, annotatedVars)
			if vv, ok := dv.(*Variable); ok {
				vars[i] = vv
			} else {
				vars[i] = v
			}
		}
		body := dropAnnotations(t.Body, true, annotatedVars)
		result, _ := NewForAll(vars, body)
		return result

	case *Exists:
		vars := make([]*Variable, len(t.Variables))
		for i, v := range t.Variables {
			dv := dropAnnotations(v, false, annotatedVars)
			if vv, ok := dv.(*Variable); ok {
				vars[i] = vv
			} else {
				vars[i] = v
			}
		}
		body := dropAnnotations(t.Body, true, annotatedVars)
		result, _ := NewExists(vars, body)
		return result

	case *Lambda:
		vars := make([]*Variable, len(t.Variables))
		for i, v := range t.Variables {
			dv := dropAnnotations(v, false, annotatedVars)
			if vv, ok := dv.(*Variable); ok {
				vars[i] = vv
			} else {
				vars[i] = v
			}
		}
		body := dropAnnotations(t.Body, true, annotatedVars)
		result, _ := NewLambda(vars, body)
		return result

	case *NamedBinder:
		vars := make([]*Variable, len(t.Variables))
		for i, v := range t.Variables {
			dv := dropAnnotations(v, false, annotatedVars)
			if vv, ok := dv.(*Variable); ok {
				vars[i] = vv
			} else {
				vars[i] = v
			}
		}
		body := dropAnnotations(t.Body, true, annotatedVars)
		result, _ := NewNamedBinder(t.Name, vars, t.Environ, body)
		return result

	// Default: recurse into children with inferred_sort=true
	// Matches Python default_drop_annotations for Not, And, Or, Implies, Iff, etc.
	default:
		return dropAnnotationsDefault(n, annotatedVars)
	}
}

// dropAnnotationsDefault handles Not, Globally, Eventually, WhenOperator, And, Or, Implies, Iff.
func dropAnnotationsDefault(n Expr, annotatedVars map[string]bool) Expr {
	switch t := n.(type) {
	case *Not:
		body := dropAnnotations(t.Body, true, annotatedVars)
		result, _ := NewNot(body)
		return result
	case *And:
		terms := make([]Expr, len(t.Terms))
		for i, a := range t.Terms {
			terms[i] = dropAnnotations(a, true, annotatedVars)
		}
		result, _ := NewAnd(terms...)
		return result
	case *Or:
		terms := make([]Expr, len(t.Terms))
		for i, a := range t.Terms {
			terms[i] = dropAnnotations(a, true, annotatedVars)
		}
		result, _ := NewOr(terms...)
		return result
	case *Implies:
		a0 := dropAnnotations(t.T1, true, annotatedVars)
		a1 := dropAnnotations(t.T2, true, annotatedVars)
		result, _ := NewImplies(a0, a1)
		return result
	case *Iff:
		a0 := dropAnnotations(t.T1, true, annotatedVars)
		a1 := dropAnnotations(t.T2, true, annotatedVars)
		result, _ := NewIff(a0, a1)
		return result
	case *Globally:
		body := dropAnnotations(t.Body, true, annotatedVars)
		result, _ := NewGlobally(t.Environ, body)
		return result
	case *Eventually:
		body := dropAnnotations(t.Body, true, annotatedVars)
		result, _ := NewEventually(t.Environ, body)
		return result
	case *WhenOperator:
		a0 := dropAnnotations(t.T1, true, annotatedVars)
		a1 := dropAnnotations(t.T2, true, annotatedVars)
		result, _ := NewWhenOperator(t.Name, a0, a1)
		return result
	}
	return n
}

// --- helpers ---

// sortName returns the name of a sort for display purposes.
func sortName(s Sort) string {
	switch st := s.(type) {
	case *UninterpretedSort:
		return st.Name
	case *BooleanSort:
		return "bool"
	case *EnumeratedSort:
		return st.Name
	case *TopSort:
		return st.Name
	default:
		return s.String()
	}
}

// isNumeralName matches Python is_numeral_name.
func isNumeralName(s string) bool {
	if len(s) == 0 {
		return false
	}
	if s[0] >= '0' && s[0] <= '9' {
		return true
	}
	if s[0] == '"' {
		return true
	}
	if s[0] == '-' && len(s) > 1 && s[1] >= '0' && s[1] <= '9' {
		return true
	}
	return false
}

// isPolymorphicSymbolName checks if a name is in the polymorphic symbols set.
// Matches Python's `name in polymorphic_symbols`.
func isPolymorphicSymbolName(name string) bool {
	switch name {
	case "<", "<=", ">", ">=", "+", "-", "*", "/", "*>",
		"bvand", "bvor", "bvnot",
		"l2s_waiting", "l2s_frozen", "l2s_saved", "l2s_d", "l2s_a",
		"cast", "arrsel", "arrupd", "arrcst":
		return true
	}
	if strings.HasPrefix(name, "bfe[") {
		return true
	}
	return false
}
