package ast

import (
	"fmt"
	"strings"
)

// --- AST-level sort types ---
// These represent sort annotations in the parsed syntax, distinct from logic.Sort.

// ConstantSort is an uninterpreted sort (also aliased as UninterpretedSort).
type ConstantSort struct {
	Base
	Elems []Node // from AST args
}

func NewConstantSort(elems ...Node) *ConstantSort { return &ConstantSort{Elems: elems} }

func (s *ConstantSort) Args() []Node           { return s.Elems }
func (s *ConstantSort) Clone(args []Node) Node { return &ConstantSort{Base: s.Base, Elems: args} }
func (s *ConstantSort) String() string          { return "uninterpreted" }
func (s *ConstantSort) Defines() []string       { return nil }
func (s *ConstantSort) Rng() Node               { return s }
func (s *ConstantSort) Dom() []Node             { return nil }

// EnumeratedSort is a sort with named elements like {a, b, c}.
type EnumeratedSort struct {
	Base
	Elems []Node // Symbol nodes representing the extension values
}

func NewEnumeratedSort(elems ...Node) *EnumeratedSort {
	return &EnumeratedSort{Elems: elems}
}

func (s *EnumeratedSort) Args() []Node           { return s.Elems }
func (s *EnumeratedSort) Clone(args []Node) Node { return &EnumeratedSort{Base: s.Base, Elems: args} }
func (s *EnumeratedSort) String() string {
	return "{" + strings.Join(s.Extension(), ",") + "}"
}
func (s *EnumeratedSort) Extension() []string {
	ext := make([]string, len(s.Elems))
	for i, e := range s.Elems {
		if sym, ok := e.(*Symbol); ok {
			ext[i] = sym.Rep
		} else {
			ext[i] = fmt.Sprint(e)
		}
	}
	return ext
}
func (s *EnumeratedSort) Defines() []string { return s.Extension() }
func (s *EnumeratedSort) Rng() Node         { return s }
func (s *EnumeratedSort) Dom() []Node       { return nil }

// StructSort represents a struct type with named fields.
type StructSort struct {
	Base
	Fields []Node // field declarations
}

func NewStructSort(fields ...Node) *StructSort { return &StructSort{Fields: fields} }

func (s *StructSort) Args() []Node           { return s.Fields }
func (s *StructSort) Clone(args []Node) Node { return &StructSort{Base: s.Base, Fields: args} }
func (s *StructSort) String() string {
	parts := make([]string, len(s.Fields))
	for i, f := range s.Fields {
		parts[i] = fmt.Sprint(f)
	}
	return "struct {" + strings.Join(parts, ",") + "}"
}
func (s *StructSort) Defines() []string {
	defs := make([]string, len(s.Fields))
	for i, f := range s.Fields {
		if a, ok := f.(*Atom); ok {
			defs[i] = a.Rep
		} else {
			defs[i] = fmt.Sprint(f)
		}
	}
	return defs
}

// FunctionSort represents a function sort: domain -> range.
type FunctionSort struct {
	Base
	Dom []Node
	Rng Node
}

func NewFunctionSort(dom []Node, rng Node) *FunctionSort {
	return &FunctionSort{Dom: dom, Rng: rng}
}

func (s *FunctionSort) Args() []Node { return nil }
func (s *FunctionSort) Clone(args []Node) Node {
	return &FunctionSort{Base: s.Base, Dom: s.Dom, Rng: s.Rng}
}
func (s *FunctionSort) String() string {
	parts := make([]string, len(s.Dom))
	for i, d := range s.Dom {
		parts[i] = fmt.Sprint(d)
	}
	return strings.Join(parts, " * ") + " -> " + fmt.Sprint(s.Rng)
}
func (s *FunctionSort) Defines() []string { return nil }

// RelationSort represents a relation sort (domain only, boolean range).
type RelationSort struct {
	Base
	Dom []Node
}

func NewRelationSort(dom []Node) *RelationSort {
	return &RelationSort{Dom: dom}
}

func (s *RelationSort) Args() []Node { return nil }
func (s *RelationSort) Clone(args []Node) Node {
	return &RelationSort{Base: s.Base, Dom: s.Dom}
}
func (s *RelationSort) String() string {
	parts := make([]string, len(s.Dom))
	for i, d := range s.Dom {
		parts[i] = fmt.Sprint(d)
	}
	return strings.Join(parts, " * ")
}
func (s *RelationSort) Defines() []string { return nil }

// Range represents a numeric range sort {lo..hi}.
type Range struct {
	Base
	Lo Node
	Hi Node
}

func NewRange(lo, hi Node) *Range { return &Range{Lo: lo, Hi: hi} }

func (r *Range) Args() []Node { return []Node{r.Lo, r.Hi} }
func (r *Range) Clone(args []Node) Node {
	return &Range{Base: r.Base, Lo: args[0], Hi: args[1]}
}
func (r *Range) String() string {
	return "{" + fmt.Sprint(r.Lo) + ".." + fmt.Sprint(r.Hi) + "}"
}
