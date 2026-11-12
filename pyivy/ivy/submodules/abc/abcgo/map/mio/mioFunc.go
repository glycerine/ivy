// Gate formula, pin ordering and derived functionality from mioFunc.c.
package mio

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/mem"
)

func Mio_SopRegister(_ *mem.Mem_Flex_t, name string) string { return name }
func formulaNames(formula string) ([]string, error) {
	tokens, err := formulaTokens(formula)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var names []string
	for _, token := range tokens {
		if token.text == "0" || token.text == "1" || token.text == MIO_STRING_CONST0 || token.text == MIO_STRING_CONST1 {
			continue
		}
		if len(token.text) == 1 && stringsOperator(token.text[0]) {
			continue
		}
		if !seen[token.text] {
			names = append(names, token.text)
			seen[token.text] = true
		}
	}
	return names, nil
}
func stringsOperator(c byte) bool {
	switch c {
	case '(', ')', '!', '*', '&', '^', '+', '|', '\'':
		return true
	}
	return false
}
func Mio_GateCollectNames(formula string, names []string) int32 {
	collected, err := formulaNames(formula)
	if err != nil || len(collected) > len(names) {
		return -1
	}
	copy(names, collected)
	return int32(len(collected))
}
func deriveGate(g *Mio_Gate_t) error {
	names, err := formulaNames(g.pForm)
	if err != nil {
		return err
	}
	if len(names) > 16 {
		return fmt.Errorf("gate %s exceeds compact SOP's 16-input limit", g.pName)
	}
	if g.pPins != nil && g.pPins.pName == "*" {
		if g.pPins.pNext != nil {
			return fmt.Errorf("gate %s mixes wildcard and explicit pins", g.pName)
		}
		template := *g.pPins
		var tail *Mio_Pin_t
		g.pPins = nil
		for _, name := range names {
			pin := template
			pin.pName = name
			pin.pNext = nil
			if tail == nil {
				g.pPins = &pin
			} else {
				tail.pNext = &pin
			}
			tail = &pin
		}
	} else {
		expected := make(map[string]bool)
		for _, name := range names {
			expected[name] = true
		}
		names = nil
		for p := g.pPins; p != nil; p = p.pNext {
			if !expected[p.pName] {
				return fmt.Errorf("gate %s has duplicate or unused pin %s", g.pName, p.pName)
			}
			delete(expected, p.pName)
			names = append(names, p.pName)
		}
		if len(expected) > 0 {
			return fmt.Errorf("gate %s formula references undeclared pins", g.pName)
		}
	}
	expression, err := ParseFormula(g.pForm, names)
	if err != nil {
		return fmt.Errorf("gate %s: %w", g.pName, err)
	}
	g.nInputs = int32(len(names))
	g.vExpr = expression
	g.dDelayMax = 0
	for p := g.pPins; p != nil; p = p.pNext {
		p.dDelayBlockMax = max(p.dDelayBlockRise, p.dDelayBlockFall)
		g.dDelayMax = max(g.dDelayMax, p.dDelayBlockMax)
	}
	g.pSop = Mio_LibDeriveSop(g.nInputs, expression, nil)
	if g.nInputs <= 6 {
		g.uTruth = Exp_Truth6(g.nInputs, expression, nil)
		g.pTruth = nil
	} else {
		words := 1 << uint(g.nInputs-6)
		g.pTruth = make([]uint64, words)
		Exp_Truth(g.nInputs, expression, g.pTruth)
	}
	return nil
}
func Mio_GateParseFormula(g *Mio_Gate_t) int32 { return mioBool(deriveGate(g) != nil) }
func Mio_LibraryParseFormulas(lib *Mio_Library_t) int32 {
	lib.nGates = 0
	for g := lib.pGates; g != nil; g = g.pNext {
		if deriveGate(g) != nil {
			return 1
		}
		lib.nGates++
	}
	return 0
}
