package ast

import (
	"fmt"
	"strings"

	iu "github.com/glycerine/goivy/ivyutils"
)

// --- AST-level sort types ---
// These represent sort annotations in the parsed syntax, distinct from logic.Sort.

// ConstantSort is an uninterpreted sort (also aliased as UninterpretedSort).
type ConstantSort struct {
	Base
	Elems []Node // from AST args
}

func (r *ConstantSort) Canon() iu.Canonical {
	s := fmt.Sprintf("(constantSort base:%v", r.Base.Canon())
	if len(r.Elems) > 0 {
		s += " elems:["
		for i, d := range r.Elems {
			_ = i
			s += fmt.Sprintf("%v ", d.Canon())
		}
		s += "]"
	}
	s += ")"
	return iu.Canonical(s)
}
func NewConstantSort(elems ...Node) *ConstantSort { return &ConstantSort{Elems: elems} }

func (s *ConstantSort) Args() []Node           { return s.Elems }
func (s *ConstantSort) Clone(args []Node) Node { return &ConstantSort{Base: s.Base, Elems: args} }
func (s *ConstantSort) String() string         { return "uninterpreted" }
func (s *ConstantSort) Defines() []string      { return nil }
func (s *ConstantSort) Rng() Node              { return s }
func (s *ConstantSort) Dom() []Node            { return nil }

// UninterpretedSortAST is the AST-level (parse-time) representation of
// Python's UninterpretedSort() from logic.py:21.
//
// This is distinct from logic.UninterpretedSort which is the compiled
// semantic-level sort used after parsing. The parser creates
// UninterpretedSortAST nodes; the compiler later converts them to
// logic.UninterpretedSort during sort resolution.
//
// The "AST" suffix distinguishes this from logic.UninterpretedSort
// to prevent confusion between parse-level and compiled representations.
//
// Note: the hand-rolled parser in parser/ uses ConstantSort as a stand-in
// for uninterpreted sorts. The new lalr_full LALR parser uses
// UninterpretedSortAST to be faithful to the original Python grammar
// where UninterpretedSort() is a distinct type.
type UninterpretedSortAST struct {
	Base
}

func (r *UninterpretedSortAST) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(UninterpretedSortAST base:%v", r.Base.Canon()))
}

func NewUninterpretedSortAST() *UninterpretedSortAST { return &UninterpretedSortAST{} }

func (s *UninterpretedSortAST) Args() []Node           { return nil }
func (s *UninterpretedSortAST) Clone(args []Node) Node { return &UninterpretedSortAST{Base: s.Base} }
func (s *UninterpretedSortAST) String() string         { return "uninterpreted" }
func (s *UninterpretedSortAST) Defines() []string      { return nil }
func (s *UninterpretedSortAST) Rng() Node              { return s }
func (s *UninterpretedSortAST) Dom() []Node            { return nil }

// EnumeratedSort is a sort with named elements like {a, b, c}.
type EnumeratedSort struct {
	Base
	Elems []Node // Symbol nodes representing the extension values
}

func (r *EnumeratedSort) Canon() iu.Canonical {
	s := fmt.Sprintf("(enumeratedSort base:%v", r.Base.Canon())
	if len(r.Elems) > 0 {
		s += " elems:["
		for i, d := range r.Elems {
			_ = i
			s += fmt.Sprintf("%v ", d.Canon())
		}
		s += "]"
	}
	s += ")"
	return iu.Canonical(s)
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

func (r *StructSort) Canon() iu.Canonical {
	s := fmt.Sprintf("(structSort base:%v", r.Base.Canon())
	if len(r.Fields) > 0 {
		s += " fields:["
		for i, d := range r.Fields {
			_ = i
			s += fmt.Sprintf("%v ", d.Canon())
		}
		s += "]"
	}
	s += ")"
	return iu.Canonical(s)
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

func (r *FunctionSort) Canon() iu.Canonical {
	s := fmt.Sprintf("(functionSort base:%v", r.Base.Canon())
	if len(r.Dom) > 0 {
		s += " dom:["
		for i, d := range r.Dom {
			_ = i
			s += fmt.Sprintf("%v ", d.Canon())
		}
		s += "]"
	}
	s += fmt.Sprintf(" range:%v)", r.Rng.Canon())
	return iu.Canonical(s)
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

func (r *RelationSort) Canon() iu.Canonical {
	s := fmt.Sprintf("(relationSort base:%v", r.Base.Canon())
	if len(r.Dom) > 0 {
		s += " dom:["
		for i, d := range r.Dom {
			_ = i
			s += fmt.Sprintf("%v ", d.Canon())
		}
		s += "]"
	}
	s += ")"
	return iu.Canonical(s)
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

func (r *Range) Canon() iu.Canonical {
	return iu.Canonical(
		fmt.Sprintf("(range base:%v lo:%v hi:%v)",
			r.Base.Canon(),
			r.Lo.Canon(),
			r.Hi.Canon(),
		))
}

func (r *Range) Args() []Node { return []Node{r.Lo, r.Hi} }
func (r *Range) Clone(args []Node) Node {
	return &Range{Base: r.Base, Lo: args[0], Hi: args[1]}
}
func (r *Range) String() string {
	return "{" + fmt.Sprint(r.Lo) + ".." + fmt.Sprint(r.Hi) + "}"
}
