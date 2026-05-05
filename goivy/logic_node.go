package goivy

// Expr is the interface for all sorted logic expressions (sorts, terms, formulas).
// It embeds ast.Node so every Expr is statically known to be an ast.Node —
// no runtime type assertions needed when assigning Expr values to ast.Node slots.
type Expr interface {
	Node
	NodeSort() Sort
	Children() []Expr
	Equal(Expr) bool
	// Sexp returns a NodeKey that uniquely identifies this node
	// by structure. Two nodes with the same Sexp() are structurally
	// equal, matching Python's recstruct == and hash behavior.
	// Returns NodeKey (distinct type from string) to enforce
	// compile-time separation of structural keys from plain strings.
	Sexp() NodeKey
}

// ExprName returns the name of an Expr node.
// Matches Python's duck-typed .name field on Var, Const, sorts,
// WhenOperator, and NamedBinder. Types without a .name in Python
// will panic to catch buggy uses.
func ExprName(x Expr) string {
	switch t := x.(type) {
	case *Const:
		return t.Name
	case *LogicVariable:
		return t.Name
	case *UninterpretedSort:
		return t.Name
	case *LogicEnumeratedSort:
		return t.Name
	case *RangeSort:
		return t.Name
	case *TopSort:
		return t.Name
	case *LogicWhenOperator:
		return t.Name
	case *LogicNamedBinder:
		return t.Name
	case *Apply:
		return ExprName(t.Func)
	}
	panicf("ExprName not implemented for %T", x)
	return ""
}

// Sort types implement Expr: they are leaf nodes whose sort is themselves.

func (s *UninterpretedSort) NodeSort() Sort   { return s }
func (s *UninterpretedSort) Children() []Expr { return nil }
func (s *UninterpretedSort) Equal(n Expr) bool {
	if o, ok := n.(*UninterpretedSort); ok {
		return s.Name == o.Name
	}
	return false
}

func (s *BooleanSort) NodeSort() Sort   { return s }
func (s *BooleanSort) Children() []Expr { return nil }
func (s *BooleanSort) Equal(n Expr) bool {
	_, ok := n.(*BooleanSort)
	return ok
}

func (s *LogicFunctionSort) NodeSort() Sort { return s }
func (s *LogicFunctionSort) Children() []Expr {
	nodes := make([]Expr, len(s.Sorts))
	for i, sub := range s.Sorts {
		nodes[i] = sub
	}
	return nodes
}
func (s *LogicFunctionSort) Equal(n Expr) bool {
	o, ok := n.(*LogicFunctionSort)
	if !ok || len(s.Sorts) != len(o.Sorts) {
		return false
	}
	for i := range s.Sorts {
		if !s.Sorts[i].Equal(o.Sorts[i]) {
			return false
		}
	}
	return true
}

func (s *LogicEnumeratedSort) NodeSort() Sort   { return s }
func (s *LogicEnumeratedSort) Children() []Expr { return nil }
func (s *LogicEnumeratedSort) Equal(n Expr) bool {
	o, ok := n.(*LogicEnumeratedSort)
	if !ok || s.Name != o.Name || len(s.Extension) != len(o.Extension) {
		return false
	}
	for i := range s.Extension {
		if s.Extension[i] != o.Extension[i] {
			return false
		}
	}
	return true
}

func (s *TopSort) NodeSort() Sort   { return s }
func (s *TopSort) Children() []Expr { return nil }
func (s *TopSort) Equal(n Expr) bool {
	if o, ok := n.(*TopSort); ok {
		return s.Name == o.Name
	}
	return false
}

func (s *RangeSort) NodeSort() Sort   { return s }
func (s *RangeSort) Children() []Expr { return nil }
func (s *RangeSort) Equal(n Expr) bool {
	o, ok := n.(*RangeSort)
	if !ok {
		return false
	}
	return s.Name == o.Name && s.Lb.BoundString() == o.Lb.BoundString() && s.Ub.BoundString() == o.Ub.BoundString()
}
