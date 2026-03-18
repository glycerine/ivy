// ast_compat.go provides Args() and Clone() methods so that ivylogic types
// (Some, Let, Literal) also satisfy ast.Node.
package ivylogic

import (
	"reflect"

	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
)

// isNil uses reflect to return true iff face is nil or contains
// a nil pointer, map, array, slice, or channel.
func isNil(face interface{}) bool {
	if face == nil {
		return true
	}
	switch reflect.TypeOf(face).Kind() {
	case reflect.Ptr, reflect.Array, reflect.Map, reflect.Slice, reflect.Chan:
		return reflect.ValueOf(face).IsNil()
	}
	return false
}

// exprToAstNode converts an lg.Expr to ast.Node via type assertion.
func exprToAstNode(e lg.Expr) ast.Node {
	if isNil(e) {
		return nil
	}
	return e.(ast.Node)
}

// --- Some ---

func (s *Some) Args() []ast.Node {
	r := make([]ast.Node, 0, len(s.Params)+3)
	for _, p := range s.Params {
		r = append(r, exprToAstNode(p))
	}
	r = append(r, exprToAstNode(s.Fmla))
	if s.IfVal != nil {
		r = append(r, exprToAstNode(s.IfVal))
	}
	if s.ElseVal != nil {
		r = append(r, exprToAstNode(s.ElseVal))
	}
	return r
}

func (s *Some) Clone(args []ast.Node) ast.Node {
	nParams := len(s.Params)
	params := make([]lg.Expr, nParams)
	for i := 0; i < nParams; i++ {
		params[i] = args[i].(lg.Expr)
	}
	result := &Some{Params: params, Fmla: args[nParams].(lg.Expr)}
	idx := nParams + 1
	if s.IfVal != nil && idx < len(args) {
		result.IfVal = args[idx].(lg.Expr)
		idx++
	}
	if s.ElseVal != nil && idx < len(args) {
		result.ElseVal = args[idx].(lg.Expr)
	}
	return result
}

// --- Let ---

func (l *Let) Args() []ast.Node {
	r := make([]ast.Node, 0, len(l.Defs)+1)
	for _, d := range l.Defs {
		r = append(r, exprToAstNode(d))
	}
	r = append(r, exprToAstNode(l.Body))
	return r
}

func (l *Let) Clone(args []ast.Node) ast.Node {
	nDefs := len(l.Defs)
	defs := make([]lg.Expr, nDefs)
	for i := 0; i < nDefs; i++ {
		defs[i] = args[i].(lg.Expr)
	}
	return &Let{Defs: defs, Body: args[nDefs].(lg.Expr)}
}

// --- Literal ---

func (l *Literal) Args() []ast.Node { return []ast.Node{exprToAstNode(l.Atom)} }

func (l *Literal) Clone(args []ast.Node) ast.Node {
	return &Literal{Polarity: l.Polarity, Atom: args[0].(lg.Expr)}
}

// --- Interface satisfaction compile-time checks ---

var (
	_ ast.Node = (*Some)(nil)
	_ ast.Node = (*Let)(nil)
	_ ast.Node = (*Literal)(nil)
)
