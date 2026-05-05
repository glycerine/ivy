// nativeexpr.go: NativeExpr holds compiled children of a native expression.
// Corresponds to Python's NativeExpr AST node after compilation,
// which clones the node with compiled args and sets sort = TopS.
package goivy

import (
	"fmt"
)

// NativeExpr is a compiled native expression that preserves all children
// with TopSort. Python: res = self.clone([a.compile() for a in self.args]); res.sort = TopS
type LogicNativeExpr struct {
	Base
	CompiledChildren []Expr
}

func (n *LogicNativeExpr) NodeSort() Sort   { return TopS }
func (n *LogicNativeExpr) Children() []Expr { return n.CompiledChildren }
func (n *LogicNativeExpr) String() string   { return fmt.Sprintf("native(%d)", len(n.CompiledChildren)) }

func (n *LogicNativeExpr) Equal(other Expr) bool {
	o, ok := other.(*LogicNativeExpr)
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

func (n *LogicNativeExpr) Sexp() NodeKey {
	s := "(NativeExpr"
	for _, c := range n.CompiledChildren {
		s += " " + string(c.Sexp())
	}
	return NodeKey(s + ")")
}

func (n *LogicNativeExpr) Args() []Node {
	r := make([]Node, len(n.CompiledChildren))
	for i, c := range n.CompiledChildren {
		r[i] = c
	}
	return r
}

func (n *LogicNativeExpr) Clone(args []Node) Node {
	children := make([]Expr, len(args))
	for i, a := range args {
		children[i] = a.(Expr)
	}
	return &LogicNativeExpr{Base: n.Base, CompiledChildren: children}
}
