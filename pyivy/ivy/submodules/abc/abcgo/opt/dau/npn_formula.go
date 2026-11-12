package dau

import (
	"bufio"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"os"
	"strconv"
	"strings"
)

func Dau_ParseFormulaEndToken(s string) int {
	depth := 0
	for i, c := range []byte(s) {
		if c == '~' {
			continue
		}
		if c == '(' {
			depth++
		} else if c == ')' {
			depth--
		}
		if depth == 0 {
			return i + 1
		}
		if depth < 0 {
			break
		}
	}
	panic("dau: malformed formula token")
}
func parseNPNFormula[T any](text string, variable func(byte) T, negate func(T) T, combine func(byte, T, T) T) T {
	negative := false
	for len(text) > 0 && text[0] == '~' {
		text = text[1:]
		negative = !negative
	}
	if text == "" {
		panic("dau: empty formula")
	}
	var result T
	if len(text) == 1 && text[0] >= 'a' && text[0] <= 'f' {
		result = variable(text[0])
	} else if text[0] == '(' && Dau_ParseFormulaEndToken(text) == len(text) {
		result = parseNPNFormula(text[1:len(text)-1], variable, negate, combine)
	} else {
		end := Dau_ParseFormulaEndToken(text)
		if end >= len(text) {
			panic("dau: missing formula operator")
		}
		operation := text[end]
		if operation != '&' && operation != '^' {
			panic("dau: invalid formula operator")
		}
		second := text[end+1:]
		if second == "" || Dau_ParseFormulaEndToken(second) != len(second) {
			panic("dau: invalid trailing formula text")
		}
		// A leading complement applies to the first token, unless parentheses
		// enclose the whole expression, exactly as in the source parser.
		left := parseNPNFormula(text[:end], variable, negate, combine)
		if negative {
			left = negate(left)
			negative = false
		}
		result = combine(operation, left, parseNPNFormula(second, variable, negate, combine))
	}
	if negative {
		return negate(result)
	}
	return result
}
func Dau_ParseFormula(s string) uint64 {
	return parseNPNFormula(s, func(v byte) uint64 { return npnVariables[v-'a'] }, func(t uint64) uint64 { return ^t }, func(op byte, a, b uint64) uint64 {
		if op == '&' {
			return a & b
		}
		return a ^ b
	})
}
func Dau_ParseFormula_rec(begin, end string) uint64 {
	if len(end) > len(begin) {
		panic("dau: invalid formula range")
	}
	return Dau_ParseFormula(begin[:len(begin)-len(end)])
}

const npnExampleFormula = "~((~~d&~(~~b&c))^(~(~a&~d)&~(~c^~b)))"

func Dau_ParseFormulaTest() {
	if Dau_ParseFormula(npnExampleFormula) != 0x037d037d037d037d {
		panic("dau: formula self-test failed")
	}
}

// FormulaAIG is implemented by graph-manager adapters; literals preserve the
// source input numbering and complement bit without importing the GIA facade.
type FormulaAIG interface {
	HashAnd(int32, int32) int32
	HashXor(int32, int32) int32
}

func Dau_ParseFormulaAig(p FormulaAIG, s string) int32 {
	return parseNPNFormula(s, func(v byte) int32 { return 2 * (1 + int32(v-'a')) }, func(t int32) int32 { return t ^ 1 }, func(op byte, a, b int32) int32 {
		if op == '&' {
			return p.HashAnd(a, b)
		}
		return p.HashXor(a, b)
	})
}

type npnGraphBuilder struct{ graph *aig.Graph }

func (p npnGraphBuilder) HashAnd(a, b int32) int32 {
	return int32(p.graph.AND(aig.Edge(a), aig.Edge(b)))
}
func (p npnGraphBuilder) HashXor(a, b int32) int32 {
	return int32(p.graph.XOR(aig.Edge(a), aig.Edge(b)))
}
func Dau_ParseFormulaAigTest() *aig.Graph {
	p := aig.NewGraph()
	for i := 0; i < 5; i++ {
		p.NewInput()
	}
	p.AddOutput(aig.Edge(Dau_ParseFormulaAig(npnGraphBuilder{p}, npnExampleFormula)))
	return p
}
func readNPNFormulas(name string, visit func(uint32, string)) int {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 1<<20)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if len(line) < 12 {
			panic("dau: malformed formula library line")
		}
		truth, err := strconv.ParseUint(line[:8], 16, 32)
		if err != nil {
			panic(err)
		}
		visit(uint32(truth), line[11:])
		count++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return count
}
func Dau_VerifyFile(name string) {
	failures := 0
	lines := readNPNFormulas(name, func(truth uint32, formula string) {
		if uint32(Dau_ParseFormula(formula)) != truth {
			fmt.Printf("Verification failed: %08x = %s\n", truth, formula)
			failures++
		}
	})
	fmt.Printf("Verification succeeded for %d functions and failed for %d functions.\n", lines-failures, failures)
}
func Dau_VerifyFileTest() { Dau_VerifyFile("lib4var.txt") }
func Dau_ConstructAigFromFile(name string) *aig.Graph {
	p := aig.NewGraph()
	for i := 0; i < 5; i++ {
		p.NewInput()
	}
	count := readNPNFormulas(name, func(_ uint32, formula string) {
		p.AddOutput(aig.Edge(Dau_ParseFormulaAig(npnGraphBuilder{p}, formula)))
	})
	fmt.Printf("Finish constructing AIG for %d structures.\n", count)
	return p
}

func Dau_ParseFormulaAig_rec(p FormulaAIG, begin, end string) int32 {
	if len(end) > len(begin) {
		panic("dau: invalid formula range")
	}
	return Dau_ParseFormulaAig(p, begin[:len(begin)-len(end)])
}
