package abc

// Boolean equation parsing from src/misc/parse/parseEqn.c and parseStack.c.

import (
	"fmt"
	"strings"
)

type BooleanFormulaParseOptions struct {
	MaxTokens int
	MaxDepth  int
	Stop      func() bool
}

type booleanFormulaToken struct {
	kind   byte
	index  int
	offset int
}

const (
	booleanFormulaVariable byte = 'v'
	booleanFormulaConst0   byte = '0'
	booleanFormulaConst1   byte = '1'
	booleanFormulaNot      byte = '!'
	booleanFormulaAnd      byte = '*'
	booleanFormulaXor      byte = '^'
	booleanFormulaOr       byte = '+'
	booleanFormulaMark     byte = '('
)

func booleanFormulaPrecedence(operator byte) int {
	switch operator {
	case booleanFormulaNot:
		return 10
	case booleanFormulaAnd:
		return 9
	case booleanFormulaXor:
		return 8
	case booleanFormulaOr:
		return 7
	}
	return 0
}

func booleanFormulaSpace(character byte) bool {
	switch character {
	case ' ', '\t', '\r', '\n':
		return true
	}
	return false
}

func validBooleanFormulaName(name string) bool {
	if name == "" || name[0] == booleanFormulaConst0 || name[0] == booleanFormulaConst1 || strings.IndexByte(name, 0) >= 0 {
		return false
	}
	for index := 0; index < len(name); index++ {
		if booleanFormulaSpace(name[index]) || strings.ContainsRune("!()*^+", rune(name[index])) {
			return false
		}
	}
	return true
}

func booleanFormulaError(offset int, format string, arguments ...any) error {
	return fmt.Errorf("%w at byte %d: %s", ErrFormat, offset, fmt.Sprintf(format, arguments...))
}

func parseBooleanFormulaPostfix(formula string, variableNames []string, options BooleanFormulaParseOptions) ([]booleanFormulaToken, error) {
	if options.MaxTokens < 0 || options.MaxDepth < 0 {
		return nil, ErrDimension
	}
	if options.MaxTokens == 0 {
		options.MaxTokens = 1 << 20
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 1 << 16
	}
	if strings.IndexByte(formula, 0) >= 0 {
		return nil, booleanFormulaError(strings.IndexByte(formula, 0), "embedded NUL")
	}
	variables := make(map[string]int, len(variableNames))
	for index, name := range variableNames {
		if !validBooleanFormulaName(name) {
			return nil, fmt.Errorf("%w: invalid variable name at index %d", ErrDimension, index)
		}
		if _, found := variables[name]; !found {
			variables[name] = index
		}
	}

	program := make([]booleanFormulaToken, 0, min(len(formula), options.MaxTokens))
	operators := make([]booleanFormulaToken, 0, min(len(formula), options.MaxDepth))
	tokenCount, depth := 0, 0
	hasOperand := false
	countToken := func(offset int) error {
		if tokenCount == options.MaxTokens {
			return fmt.Errorf("%w at byte %d: token limit", ErrLimit, offset)
		}
		tokenCount++
		return nil
	}
	emitOperator := func(operator booleanFormulaToken) {
		program = append(program, operator)
	}
	popNegations := func() {
		for len(operators) > 0 && operators[len(operators)-1].kind == booleanFormulaNot {
			emitOperator(operators[len(operators)-1])
			operators = operators[:len(operators)-1]
		}
	}
	pushBinary := func(kind byte, offset int) {
		precedence := booleanFormulaPrecedence(kind)
		for len(operators) > 0 {
			top := operators[len(operators)-1]
			if top.kind == booleanFormulaMark || booleanFormulaPrecedence(top.kind) < precedence {
				break
			}
			emitOperator(top)
			operators = operators[:len(operators)-1]
		}
		operators = append(operators, booleanFormulaToken{kind: kind, offset: offset})
	}

	for offset := 0; offset < len(formula); {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		character := formula[offset]
		if booleanFormulaSpace(character) {
			offset++
			continue
		}
		switch character {
		case booleanFormulaConst0, booleanFormulaConst1:
			if err := countToken(offset); err != nil {
				return nil, err
			}
			if hasOperand {
				return nil, booleanFormulaError(offset, "missing operator before constant")
			}
			program = append(program, booleanFormulaToken{kind: character, offset: offset})
			hasOperand = true
			popNegations()
			offset++
		case booleanFormulaNot:
			if err := countToken(offset); err != nil {
				return nil, err
			}
			if hasOperand {
				pushBinary(booleanFormulaAnd, offset)
				hasOperand = false
			}
			operators = append(operators, booleanFormulaToken{kind: booleanFormulaNot, offset: offset})
			offset++
		case booleanFormulaAnd, booleanFormulaXor, booleanFormulaOr:
			if err := countToken(offset); err != nil {
				return nil, err
			}
			if !hasOperand {
				return nil, booleanFormulaError(offset, "binary operator has no left operand")
			}
			pushBinary(character, offset)
			hasOperand = false
			offset++
		case booleanFormulaMark:
			if err := countToken(offset); err != nil {
				return nil, err
			}
			if hasOperand {
				return nil, booleanFormulaError(offset, "missing operator before opening parenthesis")
			}
			if depth == options.MaxDepth {
				return nil, fmt.Errorf("%w at byte %d: nesting limit", ErrLimit, offset)
			}
			depth++
			operators = append(operators, booleanFormulaToken{kind: booleanFormulaMark, offset: offset})
			offset++
		case ')':
			if err := countToken(offset); err != nil {
				return nil, err
			}
			if !hasOperand {
				return nil, booleanFormulaError(offset, "closing parenthesis has no operand")
			}
			foundMark := false
			for len(operators) > 0 {
				top := operators[len(operators)-1]
				operators = operators[:len(operators)-1]
				if top.kind == booleanFormulaMark {
					foundMark = true
					depth--
					break
				}
				emitOperator(top)
			}
			if !foundMark {
				return nil, booleanFormulaError(offset, "closing parenthesis has no opening parenthesis")
			}
			popNegations()
			offset++
		default:
			if hasOperand {
				return nil, booleanFormulaError(offset, "missing operator before variable")
			}
			end := offset
			for end < len(formula) {
				candidate := formula[end]
				if booleanFormulaSpace(candidate) || strings.ContainsRune("!*^+()", rune(candidate)) {
					break
				}
				end++
			}
			name := formula[offset:end]
			variable, found := variables[name]
			if !found {
				return nil, booleanFormulaError(offset, "unknown variable %q", name)
			}
			if err := countToken(offset); err != nil {
				return nil, err
			}
			program = append(program, booleanFormulaToken{kind: booleanFormulaVariable, index: variable, offset: offset})
			hasOperand = true
			popNegations()
			offset = end
		}
	}
	if !hasOperand {
		return nil, booleanFormulaError(len(formula), "formula has no final operand")
	}
	if depth != 0 {
		return nil, booleanFormulaError(len(formula), "opening parenthesis has no closing parenthesis")
	}
	for len(operators) > 0 {
		top := operators[len(operators)-1]
		operators = operators[:len(operators)-1]
		if top.kind == booleanFormulaMark {
			return nil, booleanFormulaError(top.offset, "opening parenthesis has no closing parenthesis")
		}
		emitOperator(top)
	}
	return program, nil
}

func evaluateBooleanFormulaPostfix(manager *AigManager, program []booleanFormulaToken, options BooleanFormulaParseOptions) (AigLit, error) {
	stack := make([]AigLit, 0, len(program))
	for _, token := range program {
		if options.Stop != nil && options.Stop() {
			return InvalidAigLit, ErrCancelled
		}
		switch token.kind {
		case booleanFormulaConst0:
			stack = append(stack, AigConst0)
		case booleanFormulaConst1:
			stack = append(stack, AigConst1)
		case booleanFormulaVariable:
			literal, err := manager.IthVar(token.index)
			if err != nil {
				return InvalidAigLit, err
			}
			stack = append(stack, literal)
		case booleanFormulaNot:
			if len(stack) < 1 {
				return InvalidAigLit, booleanFormulaError(token.offset, "negation has no operand")
			}
			stack[len(stack)-1] = stack[len(stack)-1].Not()
		case booleanFormulaAnd, booleanFormulaXor, booleanFormulaOr:
			if len(stack) < 2 {
				return InvalidAigLit, booleanFormulaError(token.offset, "binary operator has too few operands")
			}
			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			var result AigLit
			var err error
			switch token.kind {
			case booleanFormulaAnd:
				result, err = manager.And(left, right)
			case booleanFormulaXor:
				result, err = manager.Xor(left, right)
			case booleanFormulaOr:
				result, err = manager.Or(left, right)
			}
			if err != nil {
				return InvalidAigLit, err
			}
			stack = append(stack, result)
		default:
			return InvalidAigLit, ErrInvariant
		}
	}
	if len(stack) != 1 {
		return InvalidAigLit, booleanFormulaError(0, "formula leaves %d values", len(stack))
	}
	return stack[0], nil
}

// ParseBooleanFormula compiles one native equation expression into manager.
// The manager is unchanged if parsing or AIG construction fails.
func ParseBooleanFormula(manager *AigManager, formula string, variableNames []string, options BooleanFormulaParseOptions) (AigLit, error) {
	if manager == nil {
		return InvalidAigLit, ErrDimension
	}
	if err := manager.room(); err != nil {
		return InvalidAigLit, err
	}
	program, err := parseBooleanFormulaPostfix(formula, variableNames, options)
	if err != nil {
		return InvalidAigLit, err
	}
	work := manager.Clone()
	result, err := evaluateBooleanFormulaPostfix(work, program, options)
	if err != nil {
		return InvalidAigLit, err
	}
	*manager = *work
	return result, nil
}
