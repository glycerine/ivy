package gogen

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"strings"
)

// ExprEmitter converts Ivy logic nodes (formulas and terms) to Go source expressions.
type ExprEmitter struct {
	// HelperSorts tracks which sorts need forAll/exists helper functions.
	HelperSorts map[string]goivy.Sort
}

// NewExprEmitter creates a new expression emitter.
func NewExprEmitter() *ExprEmitter {
	return &ExprEmitter{
		HelperSorts: make(map[string]goivy.Sort),
	}
}

// EmitExpr converts a logic.Expr to a Go expression string.
func (e *ExprEmitter) EmitExpr(node goivy.Expr) (string, error) {
	if node == nil {
		return "", fmt.Errorf("gogen: nil node")
	}
	switch n := node.(type) {
	case *goivy.Variable:
		return goUnexportedName(n.Name), nil

	case *goivy.Const:
		return goExportedName(n.Name), nil

	case *goivy.Apply:
		return e.emitApply(n)

	case *goivy.Eq:
		lhs, err := e.EmitExpr(n.T1)
		if err != nil {
			return "", err
		}
		rhs, err := e.EmitExpr(n.T2)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s == %s)", lhs, rhs), nil

	case *goivy.Not:
		body, err := e.EmitExpr(n.Body)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(!%s)", body), nil

	case *goivy.And:
		if len(n.Terms) == 0 {
			return "true", nil
		}
		parts := make([]string, len(n.Terms))
		for i, t := range n.Terms {
			s, err := e.EmitExpr(t)
			if err != nil {
				return "", err
			}
			parts[i] = s
		}
		return "(" + strings.Join(parts, " && ") + ")", nil

	case *goivy.Or:
		if len(n.Terms) == 0 {
			return "false", nil
		}
		parts := make([]string, len(n.Terms))
		for i, t := range n.Terms {
			s, err := e.EmitExpr(t)
			if err != nil {
				return "", err
			}
			parts[i] = s
		}
		return "(" + strings.Join(parts, " || ") + ")", nil

	case *goivy.Implies:
		lhs, err := e.EmitExpr(n.T1)
		if err != nil {
			return "", err
		}
		rhs, err := e.EmitExpr(n.T2)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(!%s || %s)", lhs, rhs), nil

	case *goivy.Iff:
		lhs, err := e.EmitExpr(n.T1)
		if err != nil {
			return "", err
		}
		rhs, err := e.EmitExpr(n.T2)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s == %s)", lhs, rhs), nil

	case *goivy.Ite:
		return e.emitIte(n)

	case *goivy.ForAll:
		return e.emitForAll(n)

	case *goivy.Exists:
		return e.emitExists(n)

	case *goivy.Lambda:
		return e.emitLambda(n)

	default:
		return "", fmt.Errorf("gogen: unsupported expression type %T", node)
	}
}

// emitApply generates code for function/relation application.
func (e *ExprEmitter) emitApply(a *goivy.Apply) (string, error) {
	funcStr, err := e.EmitExpr(a.Func)
	if err != nil {
		return "", err
	}

	if len(a.Terms) == 0 {
		return funcStr, nil
	}

	// Determine if this is a map lookup (relation/function) or a call.
	// For relations (result is bool) and functions, we emit map lookups.
	args := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		s, err := e.EmitExpr(t)
		if err != nil {
			return "", err
		}
		args[i] = s
	}

	if len(a.Terms) == 1 {
		// Single-arg: map lookup s.Func[arg]
		return fmt.Sprintf("%s[%s]", funcStr, args[0]), nil
	}
	// Multi-arg: use composite key.
	// Check if it looks like a map with array key.
	return fmt.Sprintf("%s[[%d]int{%s}]", funcStr, len(args), strings.Join(args, ", ")), nil
}

// emitIte generates code for if-then-else expressions.
// Go has no ternary, so we use a helper function.
func (e *ExprEmitter) emitIte(ite *goivy.Ite) (string, error) {
	cond, err := e.EmitExpr(ite.Cond)
	if err != nil {
		return "", err
	}
	then, err := e.EmitExpr(ite.Then)
	if err != nil {
		return "", err
	}
	els, err := e.EmitExpr(ite.Else)
	if err != nil {
		return "", err
	}
	// For bool results, use a simple conditional helper.
	_, isBool := ite.ISort.(*goivy.BooleanSort)
	if isBool {
		return fmt.Sprintf("ite(%s, %s, %s)", cond, then, els), nil
	}
	goType := GoType(ite.ISort)
	return fmt.Sprintf("iteVal[%s](%s, %s, %s)", goType, cond, then, els), nil
}

// emitForAll generates a forAll helper call for universally quantified formulas.
func (e *ExprEmitter) emitForAll(fa *goivy.ForAll) (string, error) {
	body, err := e.EmitExpr(fa.Body)
	if err != nil {
		return "", err
	}

	// For single-variable quantifiers, emit a direct helper call.
	// For multi-variable, nest the calls.
	result := body
	for i := len(fa.Variables) - 1; i >= 0; i-- {
		v := fa.Variables[i]
		vName := goUnexportedName(v.Name)
		sortName := sortHelperName(v.VSort)
		e.HelperSorts[sortName] = v.VSort
		result = fmt.Sprintf("forAll_%s(func(%s %s) bool { return %s })",
			sortName, vName, GoType(v.VSort), result)
	}
	return result, nil
}

// emitExists generates an exists helper call for existentially quantified formulas.
func (e *ExprEmitter) emitExists(ex *goivy.Exists) (string, error) {
	body, err := e.EmitExpr(ex.Body)
	if err != nil {
		return "", err
	}

	result := body
	for i := len(ex.Variables) - 1; i >= 0; i-- {
		v := ex.Variables[i]
		vName := goUnexportedName(v.Name)
		sortName := sortHelperName(v.VSort)
		e.HelperSorts[sortName] = v.VSort
		result = fmt.Sprintf("exists_%s(func(%s %s) bool { return %s })",
			sortName, vName, GoType(v.VSort), result)
	}
	return result, nil
}

// emitLambda generates a Go function literal for a Lambda expression.
func (e *ExprEmitter) emitLambda(lam *goivy.Lambda) (string, error) {
	body, err := e.EmitExpr(lam.Body)
	if err != nil {
		return "", err
	}

	params := make([]string, len(lam.Variables))
	for i, v := range lam.Variables {
		params[i] = fmt.Sprintf("%s %s", goUnexportedName(v.Name), GoType(v.VSort))
	}
	returnType := GoType(lam.Body.NodeSort())
	return fmt.Sprintf("func(%s) %s { return %s }",
		strings.Join(params, ", "), returnType, body), nil
}

// EmitForAllHelper generates a forAll helper function for a given sort.
//
//	func forAll_Color(f func(Color) bool) bool {
//	    for _, v := range allColor { if !f(v) { return false } }
//	    return true
//	}
func EmitForAllHelper(w *CodeWriter, s goivy.Sort) {
	sortName := sortHelperName(s)
	goT := GoType(s)
	allVals := allValsExpr(s)

	w.OpenBlock(fmt.Sprintf("func forAll_%s(f func(%s) bool) bool {", sortName, goT))
	w.OpenBlock(fmt.Sprintf("for _, v := range %s {", allVals))
	w.OpenBlock("if !f(v) {")
	w.Line("return false")
	w.CloseBlock()
	w.CloseBlock()
	w.Line("return true")
	w.CloseBlock()
}

// EmitExistsHelper generates an exists helper function for a given sort.
//
//	func exists_Color(f func(Color) bool) bool {
//	    for _, v := range allColor { if f(v) { return true } }
//	    return false
//	}
func EmitExistsHelper(w *CodeWriter, s goivy.Sort) {
	sortName := sortHelperName(s)
	goT := GoType(s)
	allVals := allValsExpr(s)

	w.OpenBlock(fmt.Sprintf("func exists_%s(f func(%s) bool) bool {", sortName, goT))
	w.OpenBlock(fmt.Sprintf("for _, v := range %s {", allVals))
	w.OpenBlock("if f(v) {")
	w.Line("return true")
	w.CloseBlock()
	w.CloseBlock()
	w.Line("return false")
	w.CloseBlock()
}

// EmitIteHelper generates the ite helper function for bool values.
func EmitIteHelper(w *CodeWriter) {
	w.OpenBlock("func ite(cond, a, b bool) bool {")
	w.OpenBlock("if cond {")
	w.Line("return a")
	w.CloseBlock()
	w.Line("return b")
	w.CloseBlock()
}

// EmitIteValHelper generates a generic ite helper using type parameters (Go 1.18+).
func EmitIteValHelper(w *CodeWriter) {
	w.OpenBlock("func iteVal[T any](cond bool, a, b T) T {")
	w.OpenBlock("if cond {")
	w.Line("return a")
	w.CloseBlock()
	w.Line("return b")
	w.CloseBlock()
}

// sortHelperName returns a sanitized name for helper function suffixes.
func sortHelperName(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "Bool"
	case *goivy.EnumeratedSort:
		return goExportedName(st.Name)
	case *goivy.RangeSort:
		return goExportedName(st.Name)
	case *goivy.UninterpretedSort:
		return goExportedName(st.Name)
	default:
		return "Unknown"
	}
}

// allValsExpr returns the Go expression for iterating all values of a sort.
func allValsExpr(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "[2]bool{false, true}"
	case *goivy.EnumeratedSort:
		return "all" + goExportedName(st.Name)
	case *goivy.RangeSort:
		name := goExportedName(st.Name)
		return fmt.Sprintf("rangeValues(%sLo, %sHi)", name, name)
	case *goivy.UninterpretedSort:
		return "all" + goExportedName(st.Name)
	default:
		return "nil"
	}
}
