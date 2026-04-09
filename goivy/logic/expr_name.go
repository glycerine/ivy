package logic

import "fmt"

// ExprName returns the name of an Expr node.
// Matches Python's duck-typed .name field on Var, Const, sorts,
// WhenOperator, and NamedBinder. Types without a .name in Python
// will panic to catch buggy uses.
func ExprName(x Expr) string {
	switch t := x.(type) {
	case *Const:
		return t.Name
	case *Variable:
		return t.Name
	case *UninterpretedSort:
		return t.Name
	case *EnumeratedSort:
		return t.Name
	case *RangeSort:
		return t.Name
	case *TopSort:
		return t.Name
	case *WhenOperator:
		return t.Name
	case *NamedBinder:
		return t.Name
	default:
		panic(fmt.Sprintf("ExprName not implemented for %T", x))
	}
}
