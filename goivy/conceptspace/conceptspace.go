// Package conceptspace implements the concept space parser and AST.
// This corresponds to Python's ivy_concept_space.py.
//
// A concept space is a structured description of a set of clauses, used
// in the concept graph model for interactive proof visualization.
//
// Grammar (from PLY):
//   expr  : lit | '(' prod ')' | '(' sum ')'
//   term  : SYMBOL
//   terms : /* empty */ | term | terms ',' term
//   atom  : SYMBOL '(' terms ')'
//   lit   : atom | '~' atom
//   prod  : expr '*' expr | prod '*' expr
//   sum   : expr '+' expr | sum '+' expr
//
// SYMBOL: [a-zA-Z_=][_a-zA-Z0-9]*
//   If SYMBOL starts with uppercase → Variable, else Constant.
package conceptspace

import (
	"fmt"
	"strings"
	"unicode"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
)

// -----------------------------------------------------------------------
// AST types
// -----------------------------------------------------------------------

// Space is the interface for concept space expressions.
type Space interface {
	String() string
	// Enumerate returns all valid clause combinations from this space.
	Enumerate(memo map[string]MemoEntry, test func([]*il.Literal) bool) [][]*il.Literal
}

// MemoEntry stores a cached (params, clauses) pair for a relation name.
type MemoEntry struct {
	Params []lg.Expr
	Value  [][]*il.Literal
}

// NamedSpace is a leaf concept space containing a single literal.
type NamedSpace struct {
	Lit *il.Literal
}

func (ns *NamedSpace) String() string {
	return ns.Lit.String()
}

func (ns *NamedSpace) Enumerate(memo map[string]MemoEntry, test func([]*il.Literal) bool) [][]*il.Literal {
	// Check if the atom's relation name is in memo
	if ns.Lit.Polarity == 1 {
		if app, ok := ns.Lit.Atom.(*lg.Apply); ok {
			if c, ok2 := app.Func.(*lg.Const); ok2 {
				if entry, found := memo[c.Name]; found {
					if len(entry.Params) == len(app.Terms) {
						var result [][]*il.Literal
						subs := make(map[lg.NodeKey]lg.Expr)
						for i, p := range entry.Params {
							subs[lg.Key(p)] = app.Terms[i]
						}
						for _, cl := range entry.Value {
							newCl := substituteLiterals(cl, subs)
							result = append(result, newCl)
						}
						return result
					}
				}
			}
		}
	}
	clause := []*il.Literal{ns.Lit}
	if test(clause) {
		return [][]*il.Literal{clause}
	}
	return nil
}

// SumSpace represents a disjunction (union) of concept spaces.
type SumSpace struct {
	Spaces []Space
}

func (ss *SumSpace) String() string {
	parts := make([]string, len(ss.Spaces))
	for i, s := range ss.Spaces {
		parts[i] = s.String()
	}
	return "(" + strings.Join(parts, " + ") + ")"
}

func (ss *SumSpace) Enumerate(memo map[string]MemoEntry, test func([]*il.Literal) bool) [][]*il.Literal {
	var result [][]*il.Literal
	for _, s := range ss.Spaces {
		result = append(result, s.Enumerate(memo, test)...)
	}
	return result
}

// ProductSpace represents a conjunction (Cartesian product) of concept spaces.
type ProductSpace struct {
	Spaces []Space
}

func (ps *ProductSpace) String() string {
	parts := make([]string, len(ps.Spaces))
	for i, s := range ps.Spaces {
		parts[i] = s.String()
	}
	return "(" + strings.Join(parts, " * ") + ")"
}

func (ps *ProductSpace) Enumerate(memo map[string]MemoEntry, test func([]*il.Literal) bool) [][]*il.Literal {
	if len(ps.Spaces) == 0 {
		return [][]*il.Literal{{}}
	}
	fs := ps.Spaces[0].Enumerate(memo, test)
	for _, s := range ps.Spaces[1:] {
		fs2 := s.Enumerate(memo, test)
		var prod [][]*il.Literal
		for _, x := range fs {
			for _, y := range fs2 {
				combined := make([]*il.Literal, 0, len(x)+len(y))
				combined = append(combined, x...)
				combined = append(combined, y...)
				if test(combined) {
					prod = append(prod, combined)
				}
			}
		}
		fs = prod
	}
	return fs
}

// -----------------------------------------------------------------------
// Eval method with relational algebra (for concept graph rendering)
// -----------------------------------------------------------------------

// RelAlg is the relational algebra interface used by Eval.
// Corresponds to Python's relalg parameter in concept space eval.
type RelAlg interface {
	// Prim computes the relational value for a primitive literal.
	Prim(lit *il.Literal) interface{}
	// Empty checks if a relational value is empty.
	Empty(v interface{}) bool
	// Top returns the universe (non-empty) relational value.
	Top() interface{}
	// Prod computes the product of two relational values.
	Prod(v1, v2 interface{}) interface{}
	// Subst applies a substitution to a relational value.
	Subst(v interface{}, subs map[lg.NodeKey]lg.Expr) interface{}
}

// EvalEntry is a (clause, relational-value) pair returned by Eval.
type EvalEntry struct {
	Clause []*il.Literal
	Value  interface{}
}

// EvalMemoEntry stores a cached (params, values) pair for eval.
type EvalMemoEntry struct {
	Params []lg.Expr
	Value  []EvalEntry
}

// Eval evaluates the concept space against a relational algebra.
// This is used by the concept graph for rendering.
// Corresponds to Python's Space.eval(memo, relalg).

func (ns *NamedSpace) Eval(memo map[string]EvalMemoEntry, ra RelAlg) []EvalEntry {
	if ns.Lit.Polarity == 1 {
		if app, ok := ns.Lit.Atom.(*lg.Apply); ok {
			if c, ok2 := app.Func.(*lg.Const); ok2 {
				if entry, found := memo[c.Name]; found {
					if len(entry.Params) == len(app.Terms) {
						subs := make(map[lg.NodeKey]lg.Expr)
						for i, p := range entry.Params {
							subs[lg.Key(p)] = app.Terms[i]
						}
						var result []EvalEntry
						for _, ev := range entry.Value {
							newCl := substituteLiterals(ev.Clause, subs)
							newV := ra.Subst(ev.Value, subs)
							result = append(result, EvalEntry{Clause: newCl, Value: newV})
						}
						return result
					}
				}
			}
		}
	}
	v := ra.Prim(ns.Lit)
	if !ra.Empty(v) {
		return []EvalEntry{{Clause: []*il.Literal{ns.Lit}, Value: v}}
	}
	return nil
}

func (ss *SumSpace) Eval(memo map[string]EvalMemoEntry, ra RelAlg) []EvalEntry {
	var result []EvalEntry
	for _, s := range ss.Spaces {
		result = append(result, evalSpace(s, memo, ra)...)
	}
	return result
}

func (ps *ProductSpace) Eval(memo map[string]EvalMemoEntry, ra RelAlg) []EvalEntry {
	if len(ps.Spaces) == 0 {
		return []EvalEntry{{Clause: nil, Value: ra.Top()}}
	}
	fs := evalSpace(ps.Spaces[0], memo, ra)
	for _, s := range ps.Spaces[1:] {
		fs2 := evalSpace(s, memo, ra)
		var prod []EvalEntry
		for _, x := range fs {
			for _, y := range fs2 {
				combined := make([]*il.Literal, 0, len(x.Clause)+len(y.Clause))
				combined = append(combined, x.Clause...)
				combined = append(combined, y.Clause...)
				v := ra.Prod(x.Value, y.Value)
				if !ra.Empty(v) {
					prod = append(prod, EvalEntry{Clause: combined, Value: v})
				}
			}
		}
		fs = prod
	}
	return fs
}

// evalSpace dispatches Eval to the correct type.
func evalSpace(s Space, memo map[string]EvalMemoEntry, ra RelAlg) []EvalEntry {
	switch sp := s.(type) {
	case *NamedSpace:
		return sp.Eval(memo, ra)
	case *SumSpace:
		return sp.Eval(memo, ra)
	case *ProductSpace:
		return sp.Eval(memo, ra)
	default:
		return nil
	}
}

// -----------------------------------------------------------------------
// Substitution helpers
// -----------------------------------------------------------------------

// substituteLiterals applies a node substitution to a slice of literals.
func substituteLiterals(lits []*il.Literal, subs map[lg.NodeKey]lg.Expr) []*il.Literal {
	result := make([]*il.Literal, len(lits))
	for i, lit := range lits {
		newAtom := substituteNode(lit.Atom, subs)
		result[i] = il.NewLiteral(lit.Polarity, newAtom)
	}
	return result
}

func substituteNode(n lg.Expr, subs map[lg.NodeKey]lg.Expr) lg.Expr {
	if r, ok := subs[lg.Key(n)]; ok {
		return r
	}
	switch t := n.(type) {
	case *lg.Apply:
		newTerms := make([]lg.Expr, len(t.Terms))
		for i, term := range t.Terms {
			newTerms[i] = substituteNode(term, subs)
		}
		return &lg.Apply{Func: t.Func, Terms: newTerms}
	case *lg.Variable:
		if r, ok := subs[lg.Key(t)]; ok {
			return r
		}
		return t
	case *lg.Const:
		if r, ok := subs[lg.Key(t)]; ok {
			return r
		}
		return t
	default:
		return n
	}
}

// -----------------------------------------------------------------------
// Lexer
// -----------------------------------------------------------------------

type csTokType int

const (
	csTokEOF csTokType = iota
	csTokSYMBOL
	csTokCOMMA
	csTokLPAREN
	csTokRPAREN
	csTokLBR
	csTokRBR
	csTokPLUS
	csTokTIMES
	csTokTILDA
)

type csToken struct {
	typ csTokType
	val string
}

type csLexerState struct {
	input  string
	pos    int
	tokens []csToken
	cur    int
}

func newCSLexer(input string) *csLexerState {
	l := &csLexerState{input: input}
	l.tokenize()
	return l
}

func isCSSymbolStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == '='
}

func isCSSymbolChar(c byte) bool {
	return isCSSymbolStart(c) || (c >= '0' && c <= '9')
}

func (l *csLexerState) tokenize() {
	for l.pos < len(l.input) {
		c := l.input[l.pos]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			l.pos++
		case c == '~':
			l.tokens = append(l.tokens, csToken{csTokTILDA, "~"})
			l.pos++
		case c == ',':
			l.tokens = append(l.tokens, csToken{csTokCOMMA, ","})
			l.pos++
		case c == '+':
			l.tokens = append(l.tokens, csToken{csTokPLUS, "+"})
			l.pos++
		case c == '*':
			l.tokens = append(l.tokens, csToken{csTokTIMES, "*"})
			l.pos++
		case c == '(':
			l.tokens = append(l.tokens, csToken{csTokLPAREN, "("})
			l.pos++
		case c == ')':
			l.tokens = append(l.tokens, csToken{csTokRPAREN, ")"})
			l.pos++
		case c == '[':
			l.tokens = append(l.tokens, csToken{csTokLBR, "["})
			l.pos++
		case c == ']':
			l.tokens = append(l.tokens, csToken{csTokRBR, "]"})
			l.pos++
		case isCSSymbolStart(c):
			start := l.pos
			for l.pos < len(l.input) && isCSSymbolChar(l.input[l.pos]) {
				l.pos++
			}
			l.tokens = append(l.tokens, csToken{csTokSYMBOL, l.input[start:l.pos]})
		default:
			l.pos++ // skip unknown
		}
	}
}

func (l *csLexerState) next() csToken {
	if l.cur >= len(l.tokens) {
		return csToken{csTokEOF, ""}
	}
	t := l.tokens[l.cur]
	l.cur++
	return t
}

// -----------------------------------------------------------------------
// Parse entry point (uses goyacc-generated parser)
// -----------------------------------------------------------------------

// ToConceptSpace parses a concept space string.
// Corresponds to Python's to_concept_space(s).
func ToConceptSpace(s string) (Space, error) {
	lex := newCSLexer(s)
	adapter := &csLexAdapter{lex: lex}
	ret := csParse(adapter)
	if ret != 0 || adapter.err != "" {
		if adapter.err != "" {
			return nil, fmt.Errorf("parse error: %s", adapter.err)
		}
		return nil, fmt.Errorf("parse error (code %d)", ret)
	}
	return adapter.result, nil
}

// keep compiler happy
var _ = unicode.IsUpper
