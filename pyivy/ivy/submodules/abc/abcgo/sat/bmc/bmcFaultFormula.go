package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
)

const FFTEST_MAX_VARS = 2
const FFTEST_MAX_PARS = 8

// Gia_FormStrCount validates a fault formula's alphabet and reports the
// highest-numbered input and parameter used by the expression.
func Gia_FormStrCount(form string, vars, params *int32) int32 {
	if len(form) == 0 || form[0] != '(' {
		fmt.Println("The first symbol should be the opening parenthesis \"(\".")
		return 1
	}
	if form[len(form)-1] != ')' {
		fmt.Println("The last symbol should be the closing parenthesis \")\".")
		return 1
	}
	depth := 0
	for i := 0; i < len(form); i++ {
		switch form[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				fmt.Println("Unbalanced parentheses in formula.")
				return 1
			}
		}
	}
	if depth != 0 {
		fmt.Println("The number of opening and closing parentheses is not equal.")
		return 1
	}
	*vars, *params = 0, 0
	for i := 0; i < len(form); i++ {
		c := form[i]
		switch {
		case c >= 'a' && c <= 'b':
			if n := int32(c - 'a' + 1); n > *vars {
				*vars = n
			}
		case c >= 'p' && c <= 's':
			if n := int32(c - 'p' + 1); n > *params {
				*params = n
			}
		case c == '(' || c == ')' || c == '&' || c == '|' || c == '^' || c == '?' || c == ':':
		case c == '~':
			if i+1 >= len(form) || form[i+1] < 'a' || form[i+1] > 'z' {
				fmt.Println("Expecting alphabetic symbol after negation (~)")
				return 1
			}
		default:
			fmt.Printf("Unknown symbol (%c) in the formula (%s)\n", c, form)
			return 1
		}
	}
	if *vars != FFTEST_MAX_VARS {
		fmt.Printf("The number of input variables (%d) should be 2\n", *vars)
		return 1
	}
	if *params < 1 || *params > FFTEST_MAX_PARS {
		fmt.Printf("The number of parameters should be between 1 and %d\n", FFTEST_MAX_PARS)
		return 1
	}
	return 0
}

// Gia_FormStrTransform represents complemented variables with uppercase letters.
func Gia_FormStrTransform(form string) string {
	out := make([]byte, 0, len(form))
	for i := 0; i < len(form); i++ {
		if form[i] == '~' {
			i++
			if i >= len(form) || form[i] < 'a' || form[i] > 'z' {
				panic("bmc: invalid formula negation")
			}
			out = append(out, form[i]-'a'+'A')
		} else {
			out = append(out, form[i])
		}
	}
	return string(out)
}

// Gia_ManFormulaEndToken returns the suffix following one atomic or parenthesized token.
func Gia_ManFormulaEndToken(form string) string {
	if len(form) == 0 {
		panic("bmc: missing formula token")
	}
	if form[0] != '(' {
		return form[1:]
	}
	depth := 0
	for i := range form {
		if form[i] == '(' {
			depth++
		}
		if form[i] == ')' {
			depth--
			if depth == 0 {
				return form[i+1:]
			}
		}
	}
	panic("bmc: unclosed formula token")
}

// Gia_ManRealizeFormula constructs a hashed graph for an ABC fault formula.
func Gia_ManRealizeFormula(p *gia.Gia_Man_t, vars, params []int32, form string, nPars int32) int32 {
	if nPars < 0 || int(nPars) > len(params) {
		panic("bmc: invalid formula parameter count")
	}
	var eval func(string) int32
	eval = func(s string) int32 {
		if len(s) == 0 {
			panic("bmc: empty formula")
		}
		if len(s) == 1 {
			c := s[0]
			inv := int32(0)
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
				inv = 1
			}
			if c >= 'a' && c <= 'b' && int(c-'a') < len(vars) {
				return vars[c-'a'] ^ inv
			}
			if c >= 'p' && c <= 'w' && int32(c-'p') < nPars {
				return params[c-'p'] ^ inv
			}
			panic("bmc: invalid formula variable")
		}
		rest := Gia_ManFormulaEndToken(s)
		if s[0] == '(' && len(rest) == 0 {
			return eval(s[1 : len(s)-1])
		}
		if len(rest) < 2 {
			panic("bmc: missing formula operator")
		}
		lhs := eval(s[:len(s)-len(rest)])
		op := rest[0]
		tail := rest[1:]
		next := Gia_ManFormulaEndToken(tail)
		rhs := eval(tail[:len(tail)-len(next)])
		switch op {
		case '&':
			if len(next) == 0 {
				return gia.Gia_ManHashAnd(p, lhs, rhs)
			}
		case '|':
			if len(next) == 0 {
				return gia.Gia_ManHashOr(p, lhs, rhs)
			}
		case '^':
			if len(next) == 0 {
				return gia.Gia_ManHashXor(p, lhs, rhs)
			}
		case '?':
			if len(next) > 1 && next[0] == ':' {
				return gia.Gia_ManHashMux(p, lhs, rhs, eval(next[1:]))
			}
		}
		panic("bmc: invalid formula operator or operands")
	}
	return eval(form)
}

// Gia_ManFormulaUnfold replaces selected AND nodes with the fault formula.
// Each selected node gets its own parameter inputs, in node order.
func Gia_ManFormulaUnfold(p *gia.Gia_Man_t, form string, ffOnly int32) *gia.Gia_Man_t {
	if p == nil {
		return nil
	}
	var nvars, nparams int32
	if len(form) >= 100 || Gia_FormStrCount(form, &nvars, &nparams) != 0 {
		panic("bmc: invalid fault formula")
	}
	form = Gia_FormStrTransform(form)
	marked := make([]bool, gia.Gia_ManObjNum(p))
	if ffOnly != 0 {
		outputs := p.OutputLiterals()
		for _, lit := range outputs[gia.Gia_ManPoNum(p):] {
			marked[lit>>1] = true
		}
	}
	q := gia.Gia_ManStart(5 * gia.Gia_ManObjNum(p))
	gia.Gia_ManHashStart(q)
	values := make([]int32, gia.Gia_ManObjNum(p))
	for _, lit := range p.InputLiterals() {
		values[lit>>1] = gia.Gia_ManAppendCi(q)
	}
	read := func(lit int32) int32 { return values[lit>>1] ^ (lit & 1) }
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, _, kind := p.Gate(id)
		if kind == 1 || kind == 6 {
			continue
		}
		if kind != 2 {
			panic("bmc: formula faults require an AND-inverter graph")
		}
		if ffOnly != 0 && !marked[id] {
			values[id] = gia.Gia_ManHashAnd(q, read(a), read(b))
			continue
		}
		params := make([]int32, nparams)
		for i := range params {
			params[i] = gia.Gia_ManAppendCi(q)
		}
		values[id] = Gia_ManRealizeFormula(q, []int32{read(a), read(b)}, params, form, nparams)
	}
	for _, lit := range p.OutputLiterals() {
		gia.Gia_ManAppendCo(q, read(lit))
	}
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}

// Gia_DeriveFormula describes the first output in terms of the named inputs.
func Gia_DeriveFormula(p *gia.Gia_Man_t, names []string) string {
	if p == nil || len(p.OutputLiterals()) == 0 || len(names) < len(p.InputLiterals()) {
		panic("bmc: invalid formula graph or input names")
	}
	inputNames := make(map[int32]string, len(names))
	for i, lit := range p.InputLiterals() {
		inputNames[lit>>1] = names[i]
	}
	var derive func(int32) string
	derive = func(lit int32) string {
		a, b, c, kind := p.Gate(lit >> 1)
		inv := lit & 1
		switch kind {
		case 0:
			if inv != 0 {
				return "1"
			}
			return "0"
		case 1:
			if inv != 0 {
				return "~" + inputNames[lit>>1]
			}
			return inputNames[lit>>1]
		case 2:
			op := "&"
			if inv != 0 {
				op = "|"
				a ^= 1
				b ^= 1
			}
			return "(" + derive(a) + op + derive(b) + ")"
		case 3:
			return "(" + derive(a^inv) + "^" + derive(b) + ")"
		case 4:
			return "(" + derive(a) + "?" + derive(b^inv) + ":" + derive(c^inv) + ")"
		case 5, 6:
			return derive(a ^ inv)
		default:
			panic("bmc: invalid formula graph")
		}
	}
	return "(" + derive(p.OutputLiterals()[0]) + ")"
}
