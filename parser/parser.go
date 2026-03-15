// Package parser implements a hand-written recursive descent parser for Ivy.
// It converts a token stream from the lexer into an AST.
package parser

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// ParseError represents a parser error with location info.
type ParseError struct {
	Message string
	Line    int
	Column  int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at %d:%d: %s", e.Line, e.Column, e.Message)
}

// Parser converts tokens to AST nodes.
type Parser struct {
	lex          *lexer.Lexer
	current      lexer.Token
	version      lexer.Version
	errors       []ParseError
	labelCounter int // auto-label counter, matches Python's label_counter
}

// New creates a parser for the given input and language version.
func New(input string, version lexer.Version) *Parser {
	p := &Parser{
		lex:     lexer.New(input, version),
		version: version,
	}
	p.advance() // prime the current token
	return p
}

// Parse parses the input and returns the top-level AST.
// Matches Python behavior: each grammar rule can produce one or more
// declarations (e.g., "after init" produces both ActionDecl and MixinDecl).
func (p *Parser) Parse() ([]ast.Node, error) {
	var decls []ast.Node
	for p.current.Type != lexer.EOF {
		ds := p.parseTopLevel()
		decls = append(decls, ds...)
		if len(p.errors) > 0 {
			return decls, &p.errors[0]
		}
	}
	return decls, nil
}

// Errors returns all accumulated parse errors.
func (p *Parser) Errors() []ParseError {
	return p.errors
}

// --- Token management ---

func (p *Parser) advance() lexer.Token {
	prev := p.current
	p.current = p.lex.NextToken()
	return prev
}

func (p *Parser) peek() lexer.Token {
	return p.current
}

func (p *Parser) at(tt lexer.TokenType) bool {
	return p.current.Type == tt
}

func (p *Parser) match(tt lexer.TokenType) bool {
	if p.current.Type == tt {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) expect(tt lexer.TokenType) lexer.Token {
	if p.current.Type == tt {
		return p.advance()
	}
	p.errorf("expected %s, got %s (%q)", tt, p.current.Type, p.current.Value)
	return p.current
}

func (p *Parser) errorf(format string, args ...interface{}) {
	p.errors = append(p.errors, ParseError{
		Message: fmt.Sprintf(format, args...),
		Line:    p.current.Line,
		Column:  p.current.Column,
	})
}

func (p *Parser) loc() ast.Location {
	return ast.Location{Line: p.current.Line}
}

func (p *Parser) setLoc(n ast.Node, tok lexer.Token) ast.Node {
	n.SetLineno(ast.Location{Line: tok.Line})
	return n
}

// vle checks if version <= (major, minor).
func (p *Parser) vle(major, minor int) bool {
	return p.version[0] < major || (p.version[0] == major && p.version[1] <= minor)
}

// --- Helper parsers ---

// parseSymbol parses a SYMBOL token and returns an ast.Symbol.
func (p *Parser) parseSymbol() *ast.Symbol {
	tok := p.expect(lexer.SYMBOL)
	return ast.NewSymbol(tok.Value, nil)
}

// parseAtomName parses a symbol name (SYMBOL or THIS).
func (p *Parser) parseAtomName() (string, lexer.Token) {
	tok := p.current
	switch tok.Type {
	case lexer.SYMBOL:
		p.advance()
		return tok.Value, tok
	case lexer.THIS:
		p.advance()
		return "this", tok
	default:
		p.expect(lexer.SYMBOL) // will error
		return "", tok
	}
}

// parseCallatom parses a callable atom (may have dots).
func (p *Parser) parseCallatom() ast.Node {
	name, tok := p.parseAtomName()
	var args []ast.Node
	if p.match(lexer.LPAREN) {
		args = p.parseTermList()
		p.expect(lexer.RPAREN)
	}
	result := ast.Node(ast.NewAtom(name, args...))
	p.setLoc(result, tok)

	// Handle dot chaining: a.b.c(...)
	for p.match(lexer.DOT) {
		name2, tok2 := p.parseAtomName()
		var args2 []ast.Node
		if p.match(lexer.LPAREN) {
			args2 = p.parseTermList()
			p.expect(lexer.RPAREN)
		}
		right := ast.NewAtom(name2, args2...)
		p.setLoc(right, tok2)
		result = ast.NewDot(result, right)
	}
	return result
}

// parseTermList parses comma-separated terms.
func (p *Parser) parseTermList() []ast.Node {
	if p.at(lexer.RPAREN) || p.at(lexer.RB) {
		return nil
	}
	var terms []ast.Node
	terms = append(terms, p.parseExpr(0))
	for p.match(lexer.COMMA) {
		terms = append(terms, p.parseExpr(0))
	}
	return terms
}

// parseAType parses a type annotation (just a name, possibly dotted).
func (p *Parser) parseAType() ast.Node {
	tok := p.current
	var result ast.Node

	switch tok.Type {
	case lexer.SYMBOL:
		p.advance()
		result = ast.NewSymbol(tok.Value, nil)
	case lexer.THIS:
		p.advance()
		result = &ast.This{}
	default:
		p.errorf("expected type name, got %s", tok.Type)
		return ast.NewSymbol("?", nil)
	}
	p.setLoc(result, tok)

	// Handle dotted types: mod.type
	for p.match(lexer.DOT) {
		tok2 := p.expect(lexer.SYMBOL)
		right := ast.NewSymbol(tok2.Value, nil)
		result = ast.NewDot(result, right)
	}
	return result
}

// parseTTerm parses a typed term: name or name : type or name(params) : type
func (p *Parser) parseTTerm() ast.Node {
	tok := p.current
	ca := p.parseCallatom()
	if p.match(lexer.COLON) {
		atype := p.parseAType()
		if a, ok := ca.(*ast.Atom); ok {
			a.ASort = atype
			return a
		}
	}
	p.setLoc(ca, tok)
	return ca
}

// parseTTermList parses comma-separated typed terms.
func (p *Parser) parseTTermList() []ast.Node {
	var terms []ast.Node
	terms = append(terms, p.parseTTerm())
	for p.match(lexer.COMMA) {
		terms = append(terms, p.parseTTerm())
	}
	return terms
}

// parseLabel parses an optional [label] prefix. Returns nil if no label.
func (p *Parser) parseLabel() ast.Node {
	if p.at(lexer.LB) {
		p.advance()
		name, tok := p.parseAtomName()
		var args []ast.Node
		// Label can have arguments: [label(x,y)]
		if p.match(lexer.LPAREN) {
			args = p.parseTermList()
			p.expect(lexer.RPAREN)
		}
		p.expect(lexer.RB)
		a := ast.NewAtom(name, args...)
		p.setLoc(a, tok)
		return a
	}
	return nil
}

// parseLabeledFmla parses an optional label followed by a formula.
func (p *Parser) parseLabeledFmla() *ast.LabeledFormula {
	tok := p.current
	label := p.parseLabel()
	fmla := p.parseExpr(0)
	lf := ast.NewLabeledFormula(label, fmla)
	p.setLoc(lf, tok)
	return lf
}

// parseSimpleVars parses comma-separated simple variables: X:S, Y:T
// Uses simple type names (no dotted types) to avoid consuming the DOT
// that terminates "forall X:t. body".
func (p *Parser) parseSimpleVars() []ast.Node {
	var vars []ast.Node
	for p.at(lexer.VARIABLE) {
		tok := p.advance()
		var sort ast.Node
		if p.match(lexer.COLON) {
			// Use simple type name (SYMBOL only, no dots)
			stok := p.expect(lexer.SYMBOL)
			sort = ast.NewSymbol(stok.Value, nil)
		}
		v := ast.NewVariable(tok.Value, sort)
		p.setLoc(v, tok)
		vars = append(vars, v)
		if !p.match(lexer.COMMA) {
			break
		}
	}
	return vars
}

// parseParams parses parenthesized or bare parameter lists.
func (p *Parser) parseParams() []ast.Node {
	if p.match(lexer.LPAREN) {
		params := p.parseTTermList()
		p.expect(lexer.RPAREN)
		return params
	}
	return nil
}

// parseSort parses a sort definition: {a,b,c} | struct{...} | uninterpreted | range
func (p *Parser) parseSort() ast.Node {
	tok := p.current
	switch tok.Type {
	case lexer.LCB:
		p.advance()
		// Could be {a,b,c} enum or {lo..hi} range
		first := p.parseExpr(0)
		if p.match(lexer.DOTS) {
			// Range: {lo..hi}
			hi := p.parseExpr(0)
			p.expect(lexer.RCB)
			return p.setLoc(ast.NewRange(first, hi), tok)
		}
		// Enum: {a, b, c}
		elems := []ast.Node{first}
		for p.match(lexer.COMMA) {
			elems = append(elems, p.parseExpr(0))
		}
		p.expect(lexer.RCB)
		return p.setLoc(ast.NewEnumeratedSort(elems...), tok)

	case lexer.STRUCT:
		p.advance()
		p.expect(lexer.LCB)
		var fields []ast.Node
		if !p.at(lexer.RCB) {
			fields = p.parseTTermList()
		}
		p.expect(lexer.RCB)
		return p.setLoc(ast.NewStructSort(fields...), tok)

	default:
		// Just a type name
		return p.parseAType()
	}
}
