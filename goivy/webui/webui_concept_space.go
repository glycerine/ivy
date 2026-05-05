// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_concept_space.py.
//
// This file implements concept spaces: NamedSpace, SumSpace, ProductSpace.
// These are AST nodes for concept-space expressions that can be enumerated
// or evaluated against a relational algebra.
//
// The Python version uses PLY for parsing; the Go version provides a simple
// recursive-descent parser for the same grammar.
package webui

import (
	"fmt"
	"strings"
	"unicode"
)

// ---------------------------------------------------------------------------
// Literal, Atom (light wrappers for concept-space expressions)
// ---------------------------------------------------------------------------

// CSAtom represents a relational atom in a concept space expression.
type CSAtom struct {
	RelName string
	Args    []CSTerm
}

func (a *CSAtom) String() string {
	if len(a.Args) == 0 {
		return a.RelName
	}
	argStrs := make([]string, len(a.Args))
	for i, t := range a.Args {
		argStrs[i] = t.String()
	}
	return fmt.Sprintf("%s(%s)", a.RelName, strings.Join(argStrs, ", "))
}

// CSTerm is a term in a concept-space atom (variable or constant).
type CSTerm struct {
	Name       string
	IsVariable bool // true if starts with uppercase
}

func (t *CSTerm) String() string { return t.Name }

// CSLiteral is a possibly-negated atom.
type CSLiteral struct {
	Polarity int // 1 = positive, 0 = negative
	Atom     *CSAtom
}

func (l *CSLiteral) String() string {
	if l.Polarity == 0 {
		return "~" + l.Atom.String()
	}
	return l.Atom.String()
}

// Negate returns a literal with flipped polarity.
func (l *CSLiteral) Negate() *CSLiteral {
	return &CSLiteral{Polarity: 1 - l.Polarity, Atom: l.Atom}
}

// ---------------------------------------------------------------------------
// Concept Space nodes: NamedSpace, SumSpace, ProductSpace
// ---------------------------------------------------------------------------

// CSNode is the interface for concept space expression nodes.
type CSNode interface {
	String() string
	Enumerate(memo map[string]CSMemoEntry, test func([]*CSLiteral) bool) [][]*CSLiteral
}

// CSMemoEntry stores precomputed results for a relation name.
type CSMemoEntry struct {
	Params []CSTerm
	Value  [][]*CSLiteral
}

// NamedSpace wraps a single literal.
type NamedSpace struct {
	Lit *CSLiteral
}

func (ns *NamedSpace) String() string { return ns.Lit.String() }

// Enumerate returns clauses matching this literal.
func (ns *NamedSpace) Enumerate(memo map[string]CSMemoEntry, test func([]*CSLiteral) bool) [][]*CSLiteral {
	atom := ns.Lit.Atom
	if ns.Lit.Polarity == 1 {
		if entry, ok := memo[atom.RelName]; ok {
			if len(entry.Params) == len(atom.Args) {
				// Substitute params -> args in value.
				var result [][]*CSLiteral
				for _, clause := range entry.Value {
					subst := make(map[string]string)
					for i, p := range entry.Params {
						subst[p.Name] = atom.Args[i].Name
					}
					newClause := make([]*CSLiteral, len(clause))
					for i, lit := range clause {
						newClause[i] = substituteLiteral(lit, subst)
					}
					result = append(result, newClause)
				}
				return result
			}
		}
	}
	lit := []*CSLiteral{ns.Lit}
	if test(lit) {
		return [][]*CSLiteral{lit}
	}
	return nil
}

// substituteLiteral applies a name substitution to a literal.
func substituteLiteral(lit *CSLiteral, subst map[string]string) *CSLiteral {
	newArgs := make([]CSTerm, len(lit.Atom.Args))
	for i, a := range lit.Atom.Args {
		if newName, ok := subst[a.Name]; ok {
			newArgs[i] = CSTerm{Name: newName, IsVariable: a.IsVariable}
		} else {
			newArgs[i] = a
		}
	}
	return &CSLiteral{
		Polarity: lit.Polarity,
		Atom:     &CSAtom{RelName: lit.Atom.RelName, Args: newArgs},
	}
}

// SumSpace represents the union (+) of concept spaces.
type SumSpace struct {
	Spaces []CSNode
}

func (ss *SumSpace) String() string {
	parts := make([]string, len(ss.Spaces))
	for i, s := range ss.Spaces {
		parts[i] = s.String()
	}
	return "(" + strings.Join(parts, " + ") + ")"
}

// Enumerate returns the union of enumerations of all sub-spaces.
func (ss *SumSpace) Enumerate(memo map[string]CSMemoEntry, test func([]*CSLiteral) bool) [][]*CSLiteral {
	var result [][]*CSLiteral
	for _, s := range ss.Spaces {
		result = append(result, s.Enumerate(memo, test)...)
	}
	return result
}

// ProductSpace represents the product (*) of concept spaces.
type ProductSpace struct {
	Spaces []CSNode
}

func (ps *ProductSpace) String() string {
	parts := make([]string, len(ps.Spaces))
	for i, s := range ps.Spaces {
		parts[i] = s.String()
	}
	return "(" + strings.Join(parts, " * ") + ")"
}

// Enumerate returns the product of enumerations, filtering by test.
func (ps *ProductSpace) Enumerate(memo map[string]CSMemoEntry, test func([]*CSLiteral) bool) [][]*CSLiteral {
	if len(ps.Spaces) == 0 {
		return [][]*CSLiteral{{}}
	}
	fs := ps.Spaces[0].Enumerate(memo, test)
	for _, s := range ps.Spaces[1:] {
		fs2 := s.Enumerate(memo, test)
		var prod [][]*CSLiteral
		for _, x := range fs {
			for _, y := range fs2 {
				merged := make([]*CSLiteral, 0, len(x)+len(y))
				merged = append(merged, x...)
				merged = append(merged, y...)
				if test(merged) {
					prod = append(prod, merged)
				}
			}
		}
		fs = prod
	}
	return fs
}

// ---------------------------------------------------------------------------
// Parser for concept space expressions
// ---------------------------------------------------------------------------

// csToken types.
const (
	csTokSymbol = iota
	csTokComma
	csTokLParen
	csTokRParen
	csTokLBr
	csTokRBr
	csTokPlus
	csTokTimes
	csTokTilda
	csTokEOF
)

type csToken struct {
	typ int
	val string
}

// csLexer tokenizes a concept space expression string.
type csLexer struct {
	input  string
	pos    int
	tokens []csToken
}

func newCSLexer(input string) *csLexer {
	l := &csLexer{input: input}
	l.tokenize()
	return l
}

func (l *csLexer) tokenize() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		switch {
		case ch == ' ' || ch == '\t' || ch == '\n':
			l.pos++
		case ch == '~':
			l.tokens = append(l.tokens, csToken{csTokTilda, "~"})
			l.pos++
		case ch == ',':
			l.tokens = append(l.tokens, csToken{csTokComma, ","})
			l.pos++
		case ch == '(':
			l.tokens = append(l.tokens, csToken{csTokLParen, "("})
			l.pos++
		case ch == ')':
			l.tokens = append(l.tokens, csToken{csTokRParen, ")"})
			l.pos++
		case ch == '[':
			l.tokens = append(l.tokens, csToken{csTokLBr, "["})
			l.pos++
		case ch == ']':
			l.tokens = append(l.tokens, csToken{csTokRBr, "]"})
			l.pos++
		case ch == '+':
			l.tokens = append(l.tokens, csToken{csTokPlus, "+"})
			l.pos++
		case ch == '*':
			l.tokens = append(l.tokens, csToken{csTokTimes, "*"})
			l.pos++
		case isSymbolStart(ch):
			start := l.pos
			for l.pos < len(l.input) && isSymbolContinue(l.input[l.pos]) {
				l.pos++
			}
			l.tokens = append(l.tokens, csToken{csTokSymbol, l.input[start:l.pos]})
		default:
			l.pos++ // skip unknown chars
		}
	}
	l.tokens = append(l.tokens, csToken{csTokEOF, ""})
}

func isSymbolStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '='
}

func isSymbolContinue(ch byte) bool {
	return isSymbolStart(ch) || (ch >= '0' && ch <= '9')
}

// csParser is a recursive-descent parser for concept space expressions.
type csParser struct {
	tokens []csToken
	pos    int
}

func newCSParser(tokens []csToken) *csParser {
	return &csParser{tokens: tokens}
}

func (p *csParser) peek() csToken {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return csToken{csTokEOF, ""}
}

func (p *csParser) advance() csToken {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *csParser) expect(typ int) (csToken, error) {
	tok := p.advance()
	if tok.typ != typ {
		return tok, fmt.Errorf("expected token type %d, got %d (%q)", typ, tok.typ, tok.val)
	}
	return tok, nil
}

// parseExpr parses: expr = lit | '(' prod ')' | '(' sum ')'
func (p *csParser) parseExpr() (CSNode, error) {
	if p.peek().typ == csTokLParen {
		p.advance() // consume '('
		first, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		switch p.peek().typ {
		case csTokTimes:
			// Product
			spaces := []CSNode{first}
			for p.peek().typ == csTokTimes {
				p.advance()
				next, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				spaces = append(spaces, next)
			}
			if _, err := p.expect(csTokRParen); err != nil {
				return nil, err
			}
			return &ProductSpace{Spaces: spaces}, nil
		case csTokPlus:
			// Sum
			spaces := []CSNode{first}
			for p.peek().typ == csTokPlus {
				p.advance()
				next, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				spaces = append(spaces, next)
			}
			if _, err := p.expect(csTokRParen); err != nil {
				return nil, err
			}
			return &SumSpace{Spaces: spaces}, nil
		case csTokRParen:
			p.advance()
			return first, nil
		default:
			return nil, fmt.Errorf("unexpected token in expression: %q", p.peek().val)
		}
	}
	return p.parseLit()
}

// parseLit parses: lit = atom | '~' atom
func (p *csParser) parseLit() (CSNode, error) {
	polarity := 1
	if p.peek().typ == csTokTilda {
		p.advance()
		polarity = 0
	}
	atom, err := p.parseAtom()
	if err != nil {
		return nil, err
	}
	return &NamedSpace{Lit: &CSLiteral{Polarity: polarity, Atom: atom}}, nil
}

// parseAtom parses: atom = SYMBOL '(' terms ')'
func (p *csParser) parseAtom() (*CSAtom, error) {
	name, err := p.expect(csTokSymbol)
	if err != nil {
		return nil, err
	}
	if p.peek().typ != csTokLParen {
		return &CSAtom{RelName: name.val}, nil
	}
	p.advance() // consume '('
	args, err := p.parseTerms()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(csTokRParen); err != nil {
		return nil, err
	}
	return &CSAtom{RelName: name.val, Args: args}, nil
}

// parseTerms parses: terms = ” | term (',' term)*
func (p *csParser) parseTerms() ([]CSTerm, error) {
	if p.peek().typ == csTokRParen {
		return nil, nil
	}
	var terms []CSTerm
	term, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	terms = append(terms, term)
	for p.peek().typ == csTokComma {
		p.advance()
		term, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		terms = append(terms, term)
	}
	return terms, nil
}

// parseTerm parses: term = SYMBOL (uppercase -> variable, else constant)
func (p *csParser) parseTerm() (CSTerm, error) {
	name, err := p.expect(csTokSymbol)
	if err != nil {
		return CSTerm{}, err
	}
	isVar := len(name.val) > 0 && unicode.IsUpper(rune(name.val[0]))
	return CSTerm{Name: name.val, IsVariable: isVar}, nil
}

// ToConceptSpace parses a concept space expression string.
func ToConceptSpace(s string) (CSNode, error) {
	lexer := newCSLexer(s)
	parser := newCSParser(lexer.tokens)
	return parser.parseExpr()
}
