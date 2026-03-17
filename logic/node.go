package logic

// Node is the interface for all AST nodes (sorts, terms, formulas).
type Node interface {
	NodeSort() Sort
	Children() []Node
	String() string
	Equal(Node) bool
	// Sexp returns an S-expression that uniquely identifies this node
	// by structure. Two nodes with the same Sexp() are structurally
	// equal, matching Python's recstruct == and hash behavior.
	Sexp() string
}

// Sort types implement Node: they are leaf nodes whose sort is themselves.

func (s *UninterpretedSort) NodeSort() Sort   { return s }
func (s *UninterpretedSort) Children() []Node { return nil }
func (s *UninterpretedSort) Equal(n Node) bool {
	if o, ok := n.(*UninterpretedSort); ok {
		return s.Name == o.Name
	}
	return false
}

func (s *BooleanSort) NodeSort() Sort   { return s }
func (s *BooleanSort) Children() []Node { return nil }
func (s *BooleanSort) Equal(n Node) bool {
	_, ok := n.(*BooleanSort)
	return ok
}

func (s *FunctionSort) NodeSort() Sort   { return s }
func (s *FunctionSort) Children() []Node {
	nodes := make([]Node, len(s.Sorts))
	for i, sub := range s.Sorts {
		nodes[i] = sub
	}
	return nodes
}
func (s *FunctionSort) Equal(n Node) bool {
	o, ok := n.(*FunctionSort)
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

func (s *EnumeratedSort) NodeSort() Sort   { return s }
func (s *EnumeratedSort) Children() []Node { return nil }
func (s *EnumeratedSort) Equal(n Node) bool {
	o, ok := n.(*EnumeratedSort)
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
func (s *TopSort) Children() []Node { return nil }
func (s *TopSort) Equal(n Node) bool {
	if o, ok := n.(*TopSort); ok {
		return s.Name == o.Name
	}
	return false
}

func (s *RangeSort) NodeSort() Sort   { return s }
func (s *RangeSort) Children() []Node { return nil }
func (s *RangeSort) Equal(n Node) bool {
	o, ok := n.(*RangeSort)
	if !ok {
		return false
	}
	return s.Name == o.Name && s.Lb == o.Lb && s.Ub == o.Ub
}
