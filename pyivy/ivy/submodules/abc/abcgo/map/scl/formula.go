package scl

import (
	"fmt"
	"strings"
	"unicode"
)

type formulaNode struct {
	op       byte
	variable int
	a, b     *formulaNode
}
type formulaParser struct {
	tokens    []string
	pos       int
	variables map[string]int
}

func parseFormula(text string, pins []*SC_Pin) (*formulaNode, error) {
	p := formulaParser{variables: map[string]int{}}
	for i, pin := range pins {
		p.variables[pin.Name] = i
	}
	for i := 0; i < len(text); {
		if unicode.IsSpace(rune(text[i])) {
			i++
			continue
		}
		if strings.ContainsRune("!~'()*&+|^", rune(text[i])) {
			p.tokens = append(p.tokens, text[i:i+1])
			i++
			continue
		}
		start := i
		for i < len(text) && !unicode.IsSpace(rune(text[i])) && !strings.ContainsRune("!~'()*&+|^", rune(text[i])) {
			i++
		}
		p.tokens = append(p.tokens, text[start:i])
	}
	n, err := p.expression(0)
	if err != nil {
		return nil, err
	}
	if p.pos != len(p.tokens) {
		return nil, fmt.Errorf("unexpected formula token %q", p.tokens[p.pos])
	}
	return n, nil
}
func (p *formulaParser) atom() (*formulaNode, error) {
	if p.pos == len(p.tokens) {
		return nil, fmt.Errorf("incomplete formula")
	}
	t := p.tokens[p.pos]
	p.pos++
	var n *formulaNode
	switch t {
	case "!", "~":
		a, e := p.atom()
		if e != nil {
			return nil, e
		}
		n = &formulaNode{op: '!', a: a}
	case "(":
		a, e := p.expression(0)
		if e != nil {
			return nil, e
		}
		if p.pos == len(p.tokens) || p.tokens[p.pos] != ")" {
			return nil, fmt.Errorf("unclosed formula parenthesis")
		}
		p.pos++
		n = a
	case "0", "CONST0":
		n = &formulaNode{op: '0'}
	case "1", "CONST1":
		n = &formulaNode{op: '1'}
	default:
		i, ok := p.variables[t]
		if !ok {
			return nil, fmt.Errorf("unknown formula variable %q", t)
		}
		n = &formulaNode{op: 'v', variable: i}
	}
	for p.pos < len(p.tokens) && p.tokens[p.pos] == "'" {
		p.pos++
		n = &formulaNode{op: '!', a: n}
	}
	return n, nil
}
func (p *formulaParser) expression(minimum int) (*formulaNode, error) {
	n, err := p.atom()
	if err != nil {
		return nil, err
	}
	for p.pos < len(p.tokens) {
		t := p.tokens[p.pos]
		if t == ")" {
			break
		}
		op := byte('&')
		prec := 3
		explicit := true
		switch t {
		case "+", "|":
			op = '|'
			prec = 1
		case "^":
			op = '^'
			prec = 2
		case "*", "&":
		default:
			explicit = false
		}
		if prec < minimum {
			break
		}
		if explicit {
			p.pos++
		}
		right, e := p.expression(prec + 1)
		if e != nil {
			return nil, e
		}
		n = &formulaNode{op: op, a: n, b: right}
	}
	return n, nil
}
func (n *formulaNode) eval(bits int) bool {
	switch n.op {
	case '0':
		return false
	case '1':
		return true
	case 'v':
		return bits&(1<<n.variable) != 0
	case '!':
		return !n.a.eval(bits)
	case '&':
		return n.a.eval(bits) && n.b.eval(bits)
	case '|':
		return n.a.eval(bits) || n.b.eval(bits)
	case '^':
		return n.a.eval(bits) != n.b.eval(bits)
	}
	panic("invalid formula node")
}
func formulaTruth(text string, pins []*SC_Pin) ([]uint64, error) {
	if len(pins) > 20 {
		return nil, fmt.Errorf("truth table exceeds 20 inputs")
	}
	n, e := parseFormula(text, pins)
	if e != nil {
		return nil, e
	}
	bits := max(64, 1<<len(pins))
	v := make([]uint64, (bits+63)/64)
	for i := 0; i < bits; i++ {
		if n.eval(i) {
			v[i/64] |= uint64(1) << uint(i%64)
		}
	}
	return v, nil
}
