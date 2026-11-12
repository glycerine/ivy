// Boolean genlib formula parser. AND binds before XOR, which binds before OR;
// juxtaposition is implicit AND and apostrophe is postfix negation.
package mio

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
	"unicode"
)

type formulaToken struct {
	text     string
	position int
}
type formulaParser struct {
	tokens    []formulaToken
	next      int
	variables map[string]int32
	n         int32
}

func formulaTokens(formula string) ([]formulaToken, error) {
	var tokens []formulaToken
	for i := 0; i < len(formula); {
		if unicode.IsSpace(rune(formula[i])) {
			i++
			continue
		}
		start := i
		if strings.ContainsRune("()!*&^+|'", rune(formula[i])) {
			i++
		} else {
			for i < len(formula) && !unicode.IsSpace(rune(formula[i])) && !strings.ContainsRune("()!*&^+|'", rune(formula[i])) {
				if formula[i] == '=' || formula[i] == ';' {
					return nil, fmt.Errorf("unexpected %q at byte %d", formula[i], i)
				}
				i++
			}
		}
		tokens = append(tokens, formulaToken{formula[start:i], start})
	}
	return tokens, nil
}
func (p *formulaParser) peek() string {
	if p.next >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.next].text
}
func (p *formulaParser) primary() (*Vec_Int_t, error) {
	token := p.peek()
	if token == "" {
		return nil, fmt.Errorf("missing expression operand")
	}
	p.next++
	var result *Vec_Int_t
	var err error
	switch token {
	case "!":
		result, err = p.primary()
		if err == nil {
			Exp_Not(result)
		}
	case "(":
		result, err = p.expression(0)
		if err == nil {
			if p.peek() != ")" {
				return nil, fmt.Errorf("unclosed expression parenthesis")
			}
			p.next++
		}
	case "0", MIO_STRING_CONST0:
		result = Exp_Const0()
	case "1", MIO_STRING_CONST1:
		result = Exp_Const1()
	case ")", "*", "&", "^", "+", "|", "'":
		return nil, fmt.Errorf("unexpected operator %q", token)
	default:
		index, ok := p.variables[token]
		if !ok {
			return nil, fmt.Errorf("unknown formula variable %q", token)
		}
		result = Exp_Var(index)
	}
	if err != nil {
		return nil, err
	}
	for p.peek() == "'" {
		p.next++
		Exp_Not(result)
	}
	return result, nil
}
func formulaPrecedence(token string) (int, bool) {
	switch token {
	case "+", "|":
		return 1, true
	case "^":
		return 2, true
	case "*", "&":
		return 3, true
	case "", ")":
		return 0, false
	default:
		return 3, false
	}
}
func (p *formulaParser) expression(minimum int) (*Vec_Int_t, error) {
	left, err := p.primary()
	if err != nil {
		return nil, err
	}
	for {
		operator := p.peek()
		precedence, explicit := formulaPrecedence(operator)
		if precedence == 0 || precedence < minimum {
			return left, nil
		}
		if explicit {
			p.next++
		}
		right, err := p.expression(precedence + 1)
		if err != nil {
			return nil, err
		}
		switch operator {
		case "+", "|":
			left = Exp_Or(nil, p.n, left, right)
		case "^":
			left = Exp_Xor(nil, p.n, left, right)
		default:
			left = Exp_And(nil, p.n, left, right, 0, 0)
		}
	}
}

// ParseFormula returns the original topological compact expression encoding.
func ParseFormula(formula string, names []string) (*Vec_Int_t, error) {
	tokens, err := formulaTokens(formula)
	if err != nil {
		return nil, err
	}
	p := formulaParser{tokens: tokens, variables: make(map[string]int32), n: int32(len(names))}
	for i, name := range names {
		if _, exists := p.variables[name]; exists {
			return nil, fmt.Errorf("duplicate variable %q", name)
		}
		p.variables[name] = int32(i)
	}
	result, err := p.expression(0)
	if err != nil {
		return nil, err
	}
	if p.next != len(tokens) {
		return nil, fmt.Errorf("unexpected token %q at byte %d", p.peek(), tokens[p.next].position)
	}
	return Exp_Reverse(result), nil
}
func Mio_ParseFormula(formula string, names []string, n int32) *Vec_Int_t {
	if n < 0 || int(n) > len(names) {
		return nil
	}
	p, err := ParseFormula(formula, names[:n])
	if err != nil {
		return nil
	}
	return p
}
func Mio_ParseFormulaTruth(formula string, names []string, n int32) *Vec_Wrd_t {
	p := Mio_ParseFormula(formula, names, n)
	if p == nil || n > 20 {
		return nil
	}
	words := 1
	if n > 6 {
		words = 1 << uint(n-6)
	}
	truth := make([]uint64, words)
	Exp_Truth(n, p, truth)
	return vec.Vec_WrdAllocArrayCopy(truth, len(truth))
}

// Mio_ParseCheckFormula checks syntax and exact pin-name coverage; it does not
// require equivalence to the gate's current function.
func Mio_ParseCheckFormula(gate *Mio_Gate_t, formula string) int32 {
	names, err := formulaNames(formula)
	if err != nil {
		return 0
	}
	if gate.pPins == nil || gate.pPins.pName == "*" {
		_, err = ParseFormula(formula, names)
		return mioBool(err == nil)
	}
	expected := make(map[string]bool)
	for _, name := range names {
		expected[name] = true
	}
	var pins []string
	for p := gate.pPins; p != nil; p = p.pNext {
		if !expected[p.pName] {
			return 0
		}
		delete(expected, p.pName)
		pins = append(pins, p.pName)
	}
	if len(expected) != 0 {
		return 0
	}
	_, err = ParseFormula(formula, pins)
	return mioBool(err == nil)
}
func Mio_ParseCheckName(gate *Mio_Gate_t, text *string) int32 {
	tokens, err := formulaTokens(*text)
	if err != nil || len(tokens) == 0 {
		return -1
	}
	token := tokens[0]
	for index, p := int32(0), gate.pPins; p != nil; index, p = index+1, p.pNext {
		if p.pName == token.text {
			*text = (*text)[token.position+len(token.text):]
			return index
		}
	}
	return -1
}
func Mio_ParseFormulaTruthTest(formula string, names []string, n int32) int32 {
	expression := Mio_ParseFormula(formula, names, n)
	if expression == nil {
		return 0
	}
	truth := Mio_ParseFormulaTruth(formula, names, n)
	if truth == nil {
		return 0
	}
	if n <= 6 {
		return mioBool(vec.Vec_WrdEntry(truth, 0) == Exp_Truth6(n, expression, nil))
	}
	return 1
}
