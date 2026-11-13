// Port of the expression parser in verFormula.c by Alan Mishchenko (2006).
package ver

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
)

const (
	verOpMark    int32 = 1
	verOpMux     int32 = 2
	verOpEqu     int32 = 3
	verOpOr      int32 = 4
	verOpXor     int32 = 5
	verOpAnd     int32 = 6
	verOpNeg     int32 = 7
	verFlagStart       = 1
	verFlagVar         = 2
	verFlagOper        = 3
	verFlagError       = 4
)

func Ver_FormulaParserTopOper(man *hop.Hop_Man_t, stack *vec.Vec_Ptr_t, op int32) *hop.Hop_Obj_t {
	b := vec.Vec_PtrPop(stack).(*hop.Hop_Obj_t)
	a := vec.Vec_PtrPop(stack).(*hop.Hop_Obj_t)
	var result *hop.Hop_Obj_t
	switch op {
	case verOpAnd:
		result = hop.Hop_And(man, a, b)
	case verOpXor:
		result = hop.Hop_Exor(man, a, b)
	case verOpOr:
		result = hop.Hop_Or(man, a, b)
	case verOpEqu:
		result = hop.Hop_Not(hop.Hop_Exor(man, a, b))
	case verOpMux:
		result = hop.Hop_Mux(man, vec.Vec_PtrPop(stack).(*hop.Hop_Obj_t), a, b)
	default:
		return nil
	}
	vec.Vec_PtrPush(stack, result)
	return result
}

// Names retain the C vector's alternating length/name entries.
func Ver_FormulaParserFindVar(s string, names *vec.Vec_Ptr_t) int32 {
	end, start := 0, 0
	if len(s) > 0 && s[0] == '\\' {
		start = 1
		for end < len(s) && s[end] != ' ' {
			end++
		}
	} else {
		for end < len(s) && !strings.ContainsRune(" \t\r\n,}()!~&|^?:", rune(s[end])) {
			end++
		}
	}
	name := s[start:end]
	for i := 0; i < vec.Vec_PtrSize(names)/2; i++ {
		if vec.Vec_PtrEntry(names, 2*i).(int) == len(name) && vec.Vec_PtrEntry(names, 2*i+1).(string) == name {
			return int32(i)
		}
	}
	index := int32(vec.Vec_PtrSize(names) / 2)
	vec.Vec_PtrPush(names, len(name))
	vec.Vec_PtrPush(names, name)
	return index
}

func verFormulaError(dst []byte, message string) { clear(dst); copy(dst, message) }

func Ver_FormulaParser(formula string, man *hop.Hop_Man_t, names, functions *vec.Vec_Ptr_t, operators *vec.Vec_Int_t, errorMessage []byte) *hop.Hop_Obj_t {
	vec.Vec_PtrClear(names)
	vec.Vec_PtrClear(functions)
	vec.Vec_IntClear(operators)
	if formula == "0" || formula == "1'b0" {
		return hop.Hop_ManConst0(man)
	}
	if formula == "1" || formula == "1'b1" {
		return hop.Hop_ManConst1(man)
	}
	depth := 0
	for _, c := range formula {
		if c == '(' {
			depth++
		} else if c == ')' {
			depth--
		}
	}
	if depth != 0 {
		verFormulaError(errorMessage, "Parse_FormulaParser(): Different number of opening and closing parentheses ().")
		return nil
	}
	formula = "(" + formula + ")"
	flag := verFlagStart
	for pos := 0; pos < len(formula); pos++ {
		c := formula[pos]
		switch c {
		case ' ', '\t', '\r', '\n':
			continue
		case '!', '~':
			if flag == verFlagVar {
				verFormulaError(errorMessage, "Parse_FormulaParser(): Variable before negation.")
				flag = verFlagError
				break
			}
			vec.Vec_IntPush(operators, verOpNeg)
		case '&', '|', '^', '?', ':':
			if flag != verFlagVar {
				verFormulaError(errorMessage, "Parse_FormulaParser(): There is no variable before AND, EXOR, or OR.")
				flag = verFlagError
				break
			}
			switch c {
			case '&':
				vec.Vec_IntPush(operators, verOpAnd)
			case '|':
				vec.Vec_IntPush(operators, verOpOr)
			case '^':
				vec.Vec_IntPush(operators, verOpXor)
			case '?':
				vec.Vec_IntPush(operators, verOpMux)
			}
			flag = verFlagOper
		case '(':
			if flag == verFlagVar {
				verFormulaError(errorMessage, "Parse_FormulaParser(): Variable before a parenthesis.")
				flag = verFlagError
				break
			}
			vec.Vec_IntPush(operators, verOpMark)
			flag = verFlagStart
		case ')':
			if vec.Vec_IntSize(operators) == 0 {
				verFormulaError(errorMessage, "Parse_FormulaParser(): There is no opening parenthesis\n")
				flag = verFlagError
				break
			}
			for {
				if vec.Vec_IntSize(operators) == 0 {
					verFormulaError(errorMessage, "Parse_FormulaParser(): There is no opening parenthesis\n")
					flag = verFlagError
					break
				}
				op := vec.Vec_IntPop(operators)
				if op == verOpMark {
					break
				}
				if Ver_FormulaParserTopOper(man, functions, op) == nil {
					verFormulaError(errorMessage, "Parse_FormulaParser(): Unknown operation\n")
					return nil
				}
			}
			if flag != verFlagError {
				flag = verFlagVar
			}
		default:
			v := Ver_FormulaParserFindVar(formula[pos:], names)
			if c == '\\' {
				pos++
			}
			pos += vec.Vec_PtrEntry(names, 2*int(v)).(int) - 1
			if flag == verFlagVar {
				verFormulaError(errorMessage, "Parse_FormulaParser(): Incorrect state.")
				return nil
			}
			vec.Vec_PtrPush(functions, hop.Hop_IthVar(man, v))
			flag = verFlagVar
		}
		if flag == verFlagError {
			break
		}
		if flag == verFlagStart {
			continue
		}
		if flag == verFlagVar {
			for vec.Vec_IntSize(operators) > 0 {
				op := vec.Vec_IntPop(operators)
				if op != verOpNeg {
					vec.Vec_IntPush(operators, op)
					break
				}
				vec.Vec_PtrPush(functions, hop.Hop_Not(vec.Vec_PtrPop(functions).(*hop.Hop_Obj_t)))
			}
		} else {
			for {
				op1 := vec.Vec_IntPop(operators)
				if vec.Vec_IntSize(operators) == 0 {
					vec.Vec_IntPush(operators, op1)
					break
				}
				op2 := vec.Vec_IntPop(operators)
				if op2 >= op1 && !(op1 == op2 && op1 == verOpMux) {
					if Ver_FormulaParserTopOper(man, functions, op2) == nil {
						verFormulaError(errorMessage, "Parse_FormulaParser(): Unknown operation\n")
						return nil
					}
					vec.Vec_IntPush(operators, op1)
				} else {
					vec.Vec_IntPush(operators, op2)
					vec.Vec_IntPush(operators, op1)
					break
				}
			}
		}
	}
	if flag != verFlagError {
		if vec.Vec_PtrSize(functions) > 0 {
			result := vec.Vec_PtrPop(functions).(*hop.Hop_Obj_t)
			if vec.Vec_PtrSize(functions) == 0 {
				if vec.Vec_IntSize(operators) == 0 {
					return result
				}
				verFormulaError(errorMessage, "Parse_FormulaParser(): Something is left in the operation stack\n")
			} else {
				verFormulaError(errorMessage, "Parse_FormulaParser(): Something is left in the function stack\n")
			}
		} else {
			verFormulaError(errorMessage, "Parse_FormulaParser(): The input string is empty\n")
		}
	}
	return nil
}

func Ver_FormulaReduction(formula string, man *hop.Hop_Man_t, names *vec.Vec_Ptr_t, errorMessage []byte) *hop.Hop_Obj_t {
	pos := 0
	symbol := formula[pos]
	pos++
	complement := int32(0)
	if symbol == '~' {
		complement = 1
		symbol = formula[pos]
		pos++
	}
	if symbol != '&' && symbol != '|' && symbol != '^' {
		verFormulaError(errorMessage, fmt.Sprintf("Ver_FormulaReduction(): Unknown operation (%c)\n", symbol))
		return nil
	}
	for formula[pos] != '{' {
		pos++
	}
	pos++
	vec.Vec_PtrClear(names)
	for formula[pos] != '}' {
		v := Ver_FormulaParserFindVar(formula[pos:], names)
		pos += vec.Vec_PtrEntry(names, 2*int(v)).(int)
		for formula[pos] == ' ' || formula[pos] == ',' {
			pos++
		}
	}
	var result *hop.Hop_Obj_t
	switch symbol {
	case '&':
		result = hop.Hop_CreateAnd(man, int32(vec.Vec_PtrSize(names)/2))
	case '|':
		result = hop.Hop_CreateOr(man, int32(vec.Vec_PtrSize(names)/2))
	case '^':
		result = hop.Hop_CreateExor(man, int32(vec.Vec_PtrSize(names)/2))
	}
	return hop.Hop_NotCond(result, complement)
}
