// nativeexpr.go: NativeExpr holds compiled children of a native expression.
// Corresponds to Python's NativeExpr AST node after compilation,
// which clones the node with compiled args and sets sort = TopS.
package logic

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
)

// NativeExpr is a compiled native expression that preserves all children
// with TopSort. Python: res = self.clone([a.compile() for a in self.args]); res.sort = TopS
type NativeExpr struct {
	ast.Base
	CompiledChildren []Expr
}

func (n *NativeExpr) NodeSort() Sort   { return TopS }
func (n *NativeExpr) Children() []Expr { return n.CompiledChildren }
func (n *NativeExpr) String() string   { return fmt.Sprintf("native(%d)", len(n.CompiledChildren)) }

func (n *NativeExpr) Equal(other Expr) bool {
	o, ok := other.(*NativeExpr)
	if !ok || len(n.CompiledChildren) != len(o.CompiledChildren) {
		return false
	}
	for i, c := range n.CompiledChildren {
		if !c.Equal(o.CompiledChildren[i]) {
			return false
		}
	}
	return true
}

func (n *NativeExpr) Sexp() string {
	s := "(NativeExpr"
	for _, c := range n.CompiledChildren {
		s += " " + c.Sexp()
	}
	return s + ")"
}

func (n *NativeExpr) Args() []ast.Node {
	r := make([]ast.Node, len(n.CompiledChildren))
	for i, c := range n.CompiledChildren {
		r[i] = c
	}
	return r
}

func (n *NativeExpr) Clone(args []ast.Node) ast.Node {
	children := make([]Expr, len(args))
	for i, a := range args {
		children[i] = a.(Expr)
	}
	return &NativeExpr{Base: n.Base, CompiledChildren: children}
}
