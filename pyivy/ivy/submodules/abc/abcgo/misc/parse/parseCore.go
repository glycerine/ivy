package parse

import (
	"fmt"
	"strings"
)

type SyntaxError struct {
	Offset  int
	Message string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("parse: column %d: %s", e.Offset+1, e.Message)
}

type formulaParser struct {
	source   string
	pos      int
	names    map[string]int
	equation bool
	depth    int
}

func ParseFormula(source string, names []string) (*Expression, error) {
	return parseFormula(source, names, false)
}
func parseFormula(source string, names []string, equation bool) (*Expression, error) {
	p := &formulaParser{source: source, names: map[string]int{}, equation: equation}
	for i, name := range names {
		if name == "" {
			return nil, fmt.Errorf("parse: empty variable name")
		}
		if _, ok := p.names[name]; ok {
			return nil, fmt.Errorf("parse: duplicate variable %q", name)
		}
		p.names[name] = i
	}
	e, err := p.expr(0)
	if err != nil {
		return nil, err
	}
	p.space()
	if p.pos != len(source) {
		return nil, p.fail("unexpected trailing token")
	}
	return e, nil
}
func (p *formulaParser) fail(message string) error { return &SyntaxError{p.pos, message} }
func (p *formulaParser) space() {
	for p.pos < len(p.source) && strings.ContainsRune(" \t\r\n", rune(p.source[p.pos])) {
		p.pos++
	}
}
func (p *formulaParser) prefix() (*Expression, error) {
	p.space()
	if p.pos >= len(p.source) {
		return nil, p.fail("expected an operand")
	}
	if p.depth >= 4096 {
		return nil, p.fail("formula nesting limit exceeded")
	}
	p.depth++
	defer func() { p.depth-- }()
	c := p.source[p.pos]
	p.pos++
	switch c {
	case '!', '~':
		if c == '~' && p.equation {
			return nil, p.fail("unsupported equation operator")
		}
		e, err := p.prefix()
		if err != nil {
			return nil, err
		}
		return &Expression{Op: Not, Left: e}, nil
	case '(', '[':
		if c == '[' && p.equation {
			return nil, p.fail("rank shifts are not equation operators")
		}
		e, err := p.expr(0)
		if err != nil {
			return nil, err
		}
		p.space()
		close := byte(')')
		if c == '[' {
			close = ']'
		}
		if p.pos >= len(p.source) || p.source[p.pos] != close {
			return nil, p.fail("missing closing delimiter")
		}
		p.pos++
		if c == '[' {
			e = &Expression{Op: Shift, Left: e}
		}
		return p.postfix(e)
	case '0', '1':
		return p.postfix(&Expression{Op: Constant, Value: c == '1'})
	}
	p.pos--
	start := p.pos
	for p.pos < len(p.source) && !strings.ContainsRune(" \t\r\n()[]!~'&*^+|<=>", rune(p.source[p.pos])) {
		p.pos++
	}
	if p.pos == start {
		return nil, p.fail("expected a variable or constant")
	}
	name := p.source[start:p.pos]
	index, ok := p.names[name]
	if !ok {
		return nil, &SyntaxError{start, fmt.Sprintf("unknown variable %q", name)}
	}
	return p.postfix(&Expression{Op: Variable, Index: index})
}
func (p *formulaParser) postfix(e *Expression) (*Expression, error) {
	for {
		p.space()
		if p.pos >= len(p.source) || p.source[p.pos] != '\'' {
			return e, nil
		}
		if p.equation {
			return nil, p.fail("suffix negation is not an equation operator")
		}
		p.pos++
		e = &Expression{Op: Not, Left: e}
	}
}
func (p *formulaParser) operator() (Operator, int, int) {
	p.space()
	s := p.source[p.pos:]
	if !p.equation {
		for _, op := range []struct {
			s    string
			op   Operator
			prec int
		}{{"<=>", Equivalent, 6}, {"<+>", Xor, 8}, {"=>", Implies, 5}, {"<=", ImpliedBy, 4}} {
			if strings.HasPrefix(s, op.s) {
				return op.op, op.prec, len(op.s)
			}
		}
	}
	if len(s) == 0 {
		return 0, -1, 0
	}
	switch s[0] {
	case '*':
		return And, 9, 1
	case '&':
		if !p.equation {
			return And, 9, 1
		}
	case '^':
		return Xor, 8, 1
	case '+':
		return Or, 7, 1
	case '|':
		if !p.equation {
			return Or, 7, 1
		}
	}
	return 0, -1, 0
}
func (p *formulaParser) expr(min int) (*Expression, error) {
	left, err := p.prefix()
	if err != nil {
		return nil, err
	}
	for {
		op, prec, width := p.operator()
		if prec < 0 && !p.equation && p.pos < len(p.source) {
			c := p.source[p.pos]
			if c == '(' || c == '[' || c == '!' || c == '~' || !strings.ContainsRune(")]'&*^+|<=>", rune(c)) {
				op, prec, width = And, 9, 0
			}
		}
		if prec < min {
			return left, nil
		}
		p.pos += width
		right, err := p.expr(prec + 1)
		if err != nil {
			return nil, err
		}
		left = &Expression{Op: op, Left: left, Right: right}
	}
}
