package goivy

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

func (r *ConstantSort) Canon() Canonical {
	return Canonical(fmt.Sprintf("(constantSort%v elems:%v)", r.Base.canonFields(), SliceCanon(r.Elems)))
}

func (cfg *AstConfig) NewConstantSort(elems ...Node) *ConstantSort {
	result := &ConstantSort{Elems: elems}
	result.Cfg = cfg
	return result
}

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
// Note: The new LALR parser/ package uses
// UninterpretedSortAST to be faithful to the original Python grammar
// where UninterpretedSort() is a distinct type (from ConstantSort).
type UninterpretedSortAST struct {
	Base
}

func (r *UninterpretedSortAST) Canon() Canonical {
	// Python: UninterpretedSort = ConstantSort, so canon as constantSort to match.
	return Canonical(fmt.Sprintf("(constantSort%v elems:[])", r.Base.canonFields()))
}

func (cfg *AstConfig) NewUninterpretedSortAST() *UninterpretedSortAST {
	result := &UninterpretedSortAST{}
	result.Cfg = cfg
	return result
}

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

func (r *EnumeratedSort) Canon() Canonical {
	return Canonical(fmt.Sprintf("(enumeratedSort%v elems:%v)", r.Base.canonFields(), SliceCanon(r.Elems)))
}

func (cfg *AstConfig) NewEnumeratedSort(elems ...Node) *EnumeratedSort {
	result := &EnumeratedSort{Elems: elems}
	result.Cfg = cfg
	return result
}

func (s *EnumeratedSort) Args() []Node { return s.Elems }
func (s *EnumeratedSort) Clone(args []Node) Node {
	return &EnumeratedSort{Base: s.Base, Elems: args}
}
func (s *EnumeratedSort) String() string {
	return "{" + strings.Join(s.Extension(), ",") + "}"
}
func (s *EnumeratedSort) Extension() []string {
	ext := make([]string, len(s.Elems))
	for i, e := range s.Elems {
		ext[i] = NodeRep(e)
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

func (r *StructSort) Canon() Canonical {
	return Canonical(fmt.Sprintf("(structSort%v fields:%v)", r.Base.canonFields(), SliceCanon(r.Fields)))
}

func (cfg *AstConfig) NewStructSort(fields ...Node) *StructSort {
	result := &StructSort{Fields: fields}
	result.Cfg = cfg
	return result
}

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
		} else if app, ok := f.(*App); ok {
			// Python: a.rep for a in self.args — App.rep is a string in Python.
			// In Go, App.Rep is a Node (usually *Symbol), so extract the string.
			if sym, ok := app.Rep.(*Symbol); ok {
				defs[i] = sym.Rep
			} else {
				defs[i] = fmt.Sprint(app.Rep)
			}
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

func (r *FunctionSort) Canon() Canonical {
	return Canonical(fmt.Sprintf("(functionSort%v dom:%v range:%v)", r.Base.canonFields(), SliceCanon(r.Dom), nodeCanon(r.Rng)))
}

func (cfg *AstConfig) NewFunctionSort(dom []Node, rng Node) *FunctionSort {
	result := &FunctionSort{Dom: dom, Rng: rng}
	result.Cfg = cfg
	return result
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

func (r *RelationSort) Canon() Canonical {
	return Canonical(fmt.Sprintf("(relationSort%v dom:%v)", r.Base.canonFields(), SliceCanon(r.Dom)))
}

func (cfg *AstConfig) NewRelationSort(dom []Node) *RelationSort {
	result := &RelationSort{Dom: dom}
	result.Cfg = cfg
	return result
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

func (cfg *AstConfig) NewRange(lo, hi Node) *Range {
	result := &Range{Lo: lo, Hi: hi}
	result.Cfg = cfg
	return result
}

func (r *Range) Canon() Canonical {
	return Canonical(
		fmt.Sprintf("(range%v lo:%v hi:%v)",
			r.Base.canonFields(),
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
