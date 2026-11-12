package rpo

// Literal clusters and transitions from ABC's rpo/literal.h, by Mayler G. A.
// Martins and Vinicius Callegaro. Truth tables retain KIT's 32-bit packed and
// stretched layout: variables below five repeat through the entire first word.
import (
	"fmt"
	"slices"
)

var rpoVariableMasks = [5]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}

type Operator_t int32

const (
	LIT_NONE Operator_t = iota
	LIT_AND
	LIT_OR
	LIT_XOR
)

type Literal_t struct {
	Transition, Function []uint32
	Expression           string
}

func rpoWordNum(vars int32) int {
	if vars < 0 || vars > 30 {
		panic("RPO variable count outside [0,30]")
	}
	if vars <= 5 {
		return 1
	}
	return 1 << uint(vars-5)
}
func rpoTruthBit(table []uint32, index int) bool { return table[index>>5]&(1<<uint(index&31)) != 0 }
func rpoTruthSet(table []uint32, index int)      { table[index>>5] |= 1 << uint(index&31) }
func rpoTruthConstant(table []uint32, vars int32, value uint32) bool {
	for _, word := range table[:rpoWordNum(vars)] {
		if word != value {
			return false
		}
	}
	return true
}
func rpoTruthEqual(a, b []uint32, vars int32) bool {
	return slices.Equal(a[:rpoWordNum(vars)], b[:rpoWordNum(vars)])
}
func litTransition(input, output []uint32, vars, index int32, positive bool) {
	if index < 0 || index >= vars {
		panic("RPO transition variable out of range")
	}
	words := rpoWordNum(vars)
	// Snapshot before clearing: the C cofactor temporaries also support aliasing.
	input = slices.Clone(input[:words])
	clear(output[:words])
	if index < 5 {
		shift := uint(1) << uint(index)
		mask := rpoVariableMasks[index]
		for i, word := range input {
			low, high := word&^mask, word&mask
			zero, one := low|low<<shift, high|high>>shift
			if positive {
				output[i] = one &^ zero
			} else {
				output[i] = zero &^ one
			}
		}
		return
	}
	step := 1 << uint(index-5)
	for block := 0; block < words; block += 2 * step {
		for i := 0; i < step; i++ {
			zero, one := input[block+i], input[block+step+i]
			value := zero &^ one
			if positive {
				value = one &^ zero
			}
			output[block+i], output[block+step+i] = value, value
		}
	}
}
func Lit_TruthPositiveTransition(input, output []uint32, vars, index int32) {
	litTransition(input, output, vars, index, true)
}
func Lit_TruthNegativeTransition(input, output []uint32, vars, index int32) {
	litTransition(input, output, vars, index, false)
}
func Lit_Alloc(truth []uint32, vars, index int32, polarity byte) *Literal_t {
	if polarity != '+' && polarity != '-' {
		panic("RPO polarity must be + or -")
	}
	transition := make([]uint32, rpoWordNum(vars))
	litTransition(truth, transition, vars, index, polarity == '+')
	if rpoTruthConstant(transition, vars, 0) {
		return nil
	}
	function := make([]uint32, len(transition))
	for i := range function {
		var value uint32
		if index < 5 {
			value = rpoVariableMasks[index]
		} else if i&(1<<uint(index-5)) != 0 {
			value = ^uint32(0)
		}
		if polarity == '-' {
			value = ^value
		}
		function[i] = value
	}
	expression := string([]byte{byte('a' + index)})
	if polarity == '-' {
		expression = "!" + expression
	}
	return &Literal_t{Transition: transition, Function: function, Expression: expression}
}
func Lit_GroupLiterals(a, b *Literal_t, op Operator_t, vars int32) *Literal_t {
	result := &Literal_t{Transition: make([]uint32, rpoWordNum(vars)), Function: make([]uint32, rpoWordNum(vars))}
	symbol := ""
	switch op {
	case LIT_AND:
		symbol = "*"
		for i := range result.Function {
			result.Function[i] = a.Function[i] & b.Function[i]
		}
	case LIT_OR:
		symbol = "+"
		for i := range result.Function {
			result.Function[i] = a.Function[i] | b.Function[i]
		}
	default:
		panic("unsupported RPO grouping operator")
	}
	for i := range result.Transition {
		result.Transition[i] = a.Transition[i] | b.Transition[i]
	}
	result.Expression = "(" + a.Expression + symbol + b.Expression + ")"
	return result
}
func Lit_CreateLiteralConst(truth []uint32, vars, constant int32) *Literal_t {
	if constant != 0 && constant != 1 {
		panic("invalid literal constant")
	}
	return &Literal_t{Transition: make([]uint32, rpoWordNum(vars)), Function: slices.Clone(truth[:rpoWordNum(vars)]), Expression: fmt.Sprint(constant)}
}
func Lit_Copy(literal *Literal_t, vars int32) *Literal_t {
	return &Literal_t{Transition: slices.Clone(literal.Transition[:rpoWordNum(vars)]), Function: slices.Clone(literal.Function[:rpoWordNum(vars)]), Expression: literal.Expression}
}
func Lit_PrintTT(truth []uint32, vars int32) {
	for w := rpoWordNum(vars) - 1; w >= 0; w-- {
		fmt.Printf("%08X", truth[w])
	}
}
func Lit_PrintExp(literal *Literal_t) { fmt.Print(literal.Expression) }
func Lit_Free(literal *Literal_t) {
	if literal != nil {
		literal.Function = nil
		literal.Transition = nil
		literal.Expression = ""
	}
}
