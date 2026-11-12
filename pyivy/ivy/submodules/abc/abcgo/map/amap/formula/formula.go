// Package formula parses the Boolean expressions used by ABC's Amap library
// readers. Liberty gate functions use a small C-like notation rather than a
// complete programming language; this package preserves that notation's
// precedence and implicit-AND rules.
package formula

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/glycerine/abcgo/aig/hop"
)

// Expr is a parsed Boolean expression. The zero value is invalid.
type Expr struct {
	kind  byte
	name  string
	left  *Expr
	right *Expr
}

// Parse converts a gate equation into an expression tree.
//
// Supported operators, from highest to lowest precedence, are prefix !,
// postfix ', implicit or explicit AND (* or &), XOR (^), and OR (+ or |).
// Constants 0 and 1 and parenthesized expressions are supported.
func Parse(input string) (*Expr, error) {
	p := parser{input: input}
	p.next()
	if p.tok.kind == tokEnd {
		return nil, p.err("empty formula")
	}
	e, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.tok.kind != tokEnd {
		return nil, p.err("unexpected token %q", p.tok.text)
	}
	return e, nil
}

// Eval evaluates the expression for the supplied variable assignment.
// Missing variables return an error instead of silently becoming false.
func (e *Expr) Eval(vars map[string]bool) (bool, error) {
	if e == nil {
		return false, fmt.Errorf("nil formula expression")
	}
	switch e.kind {
	case '0':
		return false, nil
	case '1':
		return true, nil
	case 'v':
		v, ok := vars[e.name]
		if !ok {
			return false, fmt.Errorf("unknown variable %q", e.name)
		}
		return v, nil
	case '!':
		v, err := e.left.Eval(vars)
		return !v, err
	case '&', '^', '|':
		l, err := e.left.Eval(vars)
		if err != nil {
			return false, err
		}
		r, err := e.right.Eval(vars)
		if err != nil {
			return false, err
		}
		switch e.kind {
		case '&':
			return l && r, nil
		case '^':
			return l != r, nil
		default:
			return l || r, nil
		}
	default:
		return false, fmt.Errorf("invalid formula node %q", e.kind)
	}
}

// ToHop builds the corresponding structurally hashed AIG. Each referenced
// variable must be present in vars; names remain owned by the caller so they
// can follow the pin order supplied by the parsed cell.
func (e *Expr) ToHop(manager *hop.Hop_Man_t, vars map[string]*hop.Hop_Obj_t) (*hop.Hop_Obj_t, error) {
	if e == nil || manager == nil {
		return nil, fmt.Errorf("nil formula or Hop manager")
	}
	switch e.kind {
	case '0':
		return hop.Hop_ManConst0(manager), nil
	case '1':
		return hop.Hop_ManConst1(manager), nil
	case 'v':
		v := vars[e.name]
		if v == nil {
			return nil, fmt.Errorf("no Hop variable supplied for %q", e.name)
		}
		return v, nil
	case '!':
		v, err := e.left.ToHop(manager, vars)
		if err != nil {
			return nil, err
		}
		return hop.Hop_Not(v), nil
	default:
		l, err := e.left.ToHop(manager, vars)
		if err != nil {
			return nil, err
		}
		r, err := e.right.ToHop(manager, vars)
		if err != nil {
			return nil, err
		}
		switch e.kind {
		case '&':
			return hop.Hop_And(manager, l, r), nil
		case '^':
			return hop.Hop_Exor(manager, l, r), nil
		case '|':
			return hop.Hop_Or(manager, l, r), nil
		default:
			return nil, fmt.Errorf("invalid formula node %q", e.kind)
		}
	}
}

type tokenKind uint8

const (
	tokEnd tokenKind = iota
	tokName
	tokZero
	tokOne
	tokNot
	tokPostNot
	tokAnd
	tokXor
	tokOr
	tokOpen
	tokClose
)

type token struct {
	kind tokenKind
	text string
	pos  int
}

type parser struct {
	input string
	pos   int
	tok   token
}

func (p *parser) next() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
	start := p.pos
	if p.pos == len(p.input) {
		p.tok = token{kind: tokEnd, pos: p.pos}
		return
	}
	c := p.input[p.pos]
	p.pos++
	switch c {
	case '0':
		p.tok = token{kind: tokZero, text: "0", pos: start}
	case '1':
		p.tok = token{kind: tokOne, text: "1", pos: start}
	case '!':
		p.tok = token{kind: tokNot, text: "!", pos: start}
	case '\'':
		p.tok = token{kind: tokPostNot, text: "'", pos: start}
	case '*', '&':
		p.tok = token{kind: tokAnd, text: string(c), pos: start}
	case '^':
		p.tok = token{kind: tokXor, text: "^", pos: start}
	case '+', '|':
		p.tok = token{kind: tokOr, text: string(c), pos: start}
	case '(':
		p.tok = token{kind: tokOpen, text: "(", pos: start}
	case ')':
		p.tok = token{kind: tokClose, text: ")", pos: start}
	default:
		if !isNameByte(c) {
			p.tok = token{kind: tokEnd, text: string(c), pos: start}
			return
		}
		for p.pos < len(p.input) && isNameByte(p.input[p.pos]) {
			p.pos++
		}
		p.tok = token{kind: tokName, text: p.input[start:p.pos], pos: start}
	}
}

func isNameByte(c byte) bool {
	return c == '_' || c == '$' || c == '.' || c == '/' || c == '\\' ||
		(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func (p *parser) err(format string, args ...any) error {
	return fmt.Errorf("formula at byte %d: %s", p.tok.pos, fmt.Sprintf(format, args...))
}

func (p *parser) parseOr() (*Expr, error) {
	e, err := p.parseXor()
	if err != nil {
		return nil, err
	}
	for p.tok.kind == tokOr {
		p.next()
		r, err := p.parseXor()
		if err != nil {
			return nil, err
		}
		e = binary('|', e, r)
	}
	return e, nil
}

func (p *parser) parseXor() (*Expr, error) {
	e, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.tok.kind == tokXor {
		p.next()
		r, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		e = binary('^', e, r)
	}
	return e, nil
}

func startsAtom(k tokenKind) bool {
	return k == tokName || k == tokZero || k == tokOne || k == tokNot || k == tokOpen
}

func (p *parser) parseAnd() (*Expr, error) {
	e, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		explicit := p.tok.kind == tokAnd
		if !explicit && !startsAtom(p.tok.kind) {
			return e, nil
		}
		if explicit {
			p.next()
		}
		r, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		e = binary('&', e, r)
	}
}

func (p *parser) parseUnary() (*Expr, error) {
	if p.tok.kind == tokNot {
		p.next()
		e, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		e = &Expr{kind: '!', left: e}
		return p.parsePostfix(e), nil
	}
	e, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	return p.parsePostfix(e), nil
}

func (p *parser) parsePostfix(e *Expr) *Expr {
	for p.tok.kind == tokPostNot {
		p.next()
		e = &Expr{kind: '!', left: e}
	}
	return e
}

func (p *parser) parsePrimary() (*Expr, error) {
	switch p.tok.kind {
	case tokZero:
		p.next()
		return &Expr{kind: '0'}, nil
	case tokOne:
		p.next()
		return &Expr{kind: '1'}, nil
	case tokName:
		e := &Expr{kind: 'v', name: p.tok.text}
		p.next()
		return e, nil
	case tokOpen:
		p.next()
		e, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.tok.kind != tokClose {
			return nil, p.err("expected ')' ")
		}
		p.next()
		return e, nil
	default:
		return nil, p.err("expected a variable, constant, negation, or '('")
	}
}

func binary(kind byte, left, right *Expr) *Expr {
	return &Expr{kind: kind, left: left, right: right}
}

// Variables returns names referenced by e in first-occurrence order.
func (e *Expr) Variables() []string {
	seen := make(map[string]bool)
	var names []string
	var walk func(*Expr)
	walk = func(n *Expr) {
		if n == nil {
			return
		}
		if n.kind == 'v' && !seen[n.name] {
			seen[n.name] = true
			names = append(names, n.name)
		}
		walk(n.left)
		walk(n.right)
	}
	walk(e)
	return names
}

// String prints a fully parenthesized equivalent expression.
func (e *Expr) String() string {
	if e == nil {
		return "<nil>"
	}
	switch e.kind {
	case '0', '1':
		return string(e.kind)
	case 'v':
		return e.name
	case '!':
		return "!" + e.left.String()
	case '&':
		return "(" + e.left.String() + " * " + e.right.String() + ")"
	case '^':
		return "(" + e.left.String() + " ^ " + e.right.String() + ")"
	case '|':
		return "(" + e.left.String() + " + " + e.right.String() + ")"
	default:
		return strings.TrimSpace(fmt.Sprintf("<invalid:%c>", e.kind))
	}
}
