package ast

import (
	"fmt"
	"strings"

	iu "github.com/glycerine/ivy/goivy/ivyutils"
)

// --- AST-level sort types ---
// These represent sort annotations in the parsed syntax, distinct from logic.Sort.

// ConstantSort is an uninterpreted sort (also aliased as UninterpretedSort).
type AstConstantSort struct {
	Base
	Elems []Node // from AST args
}

func (r *AstConstantSort) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(constantSort%v elems:%v)", r.Base.canonFields(), SliceCanon(r.Elems)))
}

func (cfg *AstConfig) NewConstantSort(elems ...Node) *AstConstantSort {
	result := &AstConstantSort{Elems: elems}
	result.Cfg = cfg
	return result
}

func (s *AstConstantSort) Args() []Node           { return s.Elems }
func (s *AstConstantSort) Clone(args []Node) Node { return &AstConstantSort{Base: s.Base, Elems: args} }
func (s *AstConstantSort) String() string         { return "uninterpreted" }
func (s *AstConstantSort) Defines() []string      { return nil }
func (s *AstConstantSort) Rng() Node              { return s }
func (s *AstConstantSort) Dom() []Node            { return nil }

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

func (r *UninterpretedSortAST) Canon() iu.Canonical {
	// Python: UninterpretedSort = ConstantSort, so canon as constantSort to match.
	return iu.Canonical(fmt.Sprintf("(constantSort%v elems:[])", r.Base.canonFields()))
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
type AstEnumeratedSort struct {
	Base
	Elems []Node // Symbol nodes representing the extension values
}

func (r *AstEnumeratedSort) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(enumeratedSort%v elems:%v)", r.Base.canonFields(), SliceCanon(r.Elems)))
}

func (cfg *AstConfig) NewEnumeratedSort(elems ...Node) *AstEnumeratedSort {
	result := &AstEnumeratedSort{Elems: elems}
	result.Cfg = cfg
	return result
}

func (s *AstEnumeratedSort) Args() []Node { return s.Elems }
func (s *AstEnumeratedSort) Clone(args []Node) Node {
	return &AstEnumeratedSort{Base: s.Base, Elems: args}
}
func (s *AstEnumeratedSort) String() string {
	return "{" + strings.Join(s.Extension(), ",") + "}"
}
func (s *AstEnumeratedSort) Extension() []string {
	ext := make([]string, len(s.Elems))
	for i, e := range s.Elems {
		ext[i] = NodeRep(e)
	}
	return ext
}
func (s *AstEnumeratedSort) Defines() []string { return s.Extension() }
func (s *AstEnumeratedSort) Rng() Node         { return s }
func (s *AstEnumeratedSort) Dom() []Node       { return nil }

// StructSort represents a struct type with named fields.
type AstStructSort struct {
	Base
	Fields []Node // field declarations
}

func (r *AstStructSort) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(structSort%v fields:%v)", r.Base.canonFields(), SliceCanon(r.Fields)))
}

func (cfg *AstConfig) NewStructSort(fields ...Node) *AstStructSort {
	result := &AstStructSort{Fields: fields}
	result.Cfg = cfg
	return result
}

func (s *AstStructSort) Args() []Node           { return s.Fields }
func (s *AstStructSort) Clone(args []Node) Node { return &AstStructSort{Base: s.Base, Fields: args} }
func (s *AstStructSort) String() string {
	parts := make([]string, len(s.Fields))
	for i, f := range s.Fields {
		parts[i] = fmt.Sprint(f)
	}
	return "struct {" + strings.Join(parts, ",") + "}"
}
func (s *AstStructSort) Defines() []string {
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
type AstFunctionSort struct {
	Base
	Dom []Node
	Rng Node
}

func (r *AstFunctionSort) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(functionSort%v dom:%v range:%v)", r.Base.canonFields(), SliceCanon(r.Dom), nodeCanon(r.Rng)))
}

func (cfg *AstConfig) NewFunctionSort(dom []Node, rng Node) *AstFunctionSort {
	result := &AstFunctionSort{Dom: dom, Rng: rng}
	result.Cfg = cfg
	return result
}

func (s *AstFunctionSort) Args() []Node { return nil }
func (s *AstFunctionSort) Clone(args []Node) Node {
	return &AstFunctionSort{Base: s.Base, Dom: s.Dom, Rng: s.Rng}
}
func (s *AstFunctionSort) String() string {
	parts := make([]string, len(s.Dom))
	for i, d := range s.Dom {
		parts[i] = fmt.Sprint(d)
	}
	return strings.Join(parts, " * ") + " -> " + fmt.Sprint(s.Rng)
}
func (s *AstFunctionSort) Defines() []string { return nil }

// RelationSort represents a relation sort (domain only, boolean range).
type AstRelationSort struct {
	Base
	Dom []Node
}

func (r *AstRelationSort) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(relationSort%v dom:%v)", r.Base.canonFields(), SliceCanon(r.Dom)))
}

func (cfg *AstConfig) NewRelationSort(dom []Node) *AstRelationSort {
	result := &AstRelationSort{Dom: dom}
	result.Cfg = cfg
	return result
}

func (s *AstRelationSort) Args() []Node { return nil }
func (s *AstRelationSort) Clone(args []Node) Node {
	return &AstRelationSort{Base: s.Base, Dom: s.Dom}
}
func (s *AstRelationSort) String() string {
	parts := make([]string, len(s.Dom))
	for i, d := range s.Dom {
		parts[i] = fmt.Sprint(d)
	}
	return strings.Join(parts, " * ")
}
func (s *AstRelationSort) Defines() []string { return nil }

// Range represents a numeric range sort {lo..hi}.
type AstRange struct {
	Base
	Lo Node
	Hi Node
}

func (cfg *AstConfig) NewRange(lo, hi Node) *AstRange {
	result := &AstRange{Lo: lo, Hi: hi}
	result.Cfg = cfg
	return result
}

func (r *AstRange) Canon() iu.Canonical {
	return iu.Canonical(
		fmt.Sprintf("(range%v lo:%v hi:%v)",
			r.Base.canonFields(),
			r.Lo.Canon(),
			r.Hi.Canon(),
		))
}

func (r *AstRange) Args() []Node { return []Node{r.Lo, r.Hi} }
func (r *AstRange) Clone(args []Node) Node {
	return &AstRange{Base: r.Base, Lo: args[0], Hi: args[1]}
}
func (r *AstRange) String() string {
	return "{" + fmt.Sprint(r.Lo) + ".." + fmt.Sprint(r.Hi) + "}"
}
