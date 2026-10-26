package abc

// Ranked Boolean formula parsing from src/misc/parse/parseCore.c and parse.h.

import (
	"fmt"
	"strings"
)

type RankedBooleanParseOptions struct {
	MaxTokens, MaxDepth, MaxNodes, MaxWords int
	Stop                                    func() bool
}

type RankedBooleanFormula struct {
	Table            *TruthTable
	Variables, Ranks int
	Tokens, Nodes    int
	VariableNames    []string
}

type rankedFormulaNode struct {
	kind        byte
	variable    int
	left, right int
}

type rankedFormulaOperator struct {
	kind, delimiter byte
	offset          int
}

const (
	rankedNodeVariable byte = 'v'
	rankedNodeFalse    byte = '0'
	rankedNodeTrue     byte = '1'
	rankedNodeNot      byte = '!'
	rankedNodeAnd      byte = '&'
	rankedNodeXor      byte = '^'
	rankedNodeOr       byte = '|'
	rankedNodeEquiv    byte = 'e'
	rankedNodeForward  byte = 'r'
	rankedNodeBackward byte = 'l'
	rankedNodeMark     byte = '('
)

func normalizeRankedBooleanOptions(options RankedBooleanParseOptions) (RankedBooleanParseOptions, error) {
	if options.MaxTokens < 0 || options.MaxDepth < 0 || options.MaxNodes < 0 || options.MaxWords < 0 {
		return options, ErrDimension
	}
	if options.MaxTokens == 0 {
		options.MaxTokens = 1 << 20
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 1 << 16
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 20
	}
	if options.MaxWords == 0 {
		options.MaxWords = 1 << 20
	}
	return options, nil
}

func validRankedBooleanName(name string) bool {
	if name == "" || !((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z') || name[0] == '_') {
		return false
	}
	for index := 1; index < len(name); index++ {
		character := name[index]
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '_') {
			return false
		}
	}
	return true
}

func rankedFormulaPrecedence(kind byte) int {
	switch kind {
	case rankedNodeNot:
		return 10
	case rankedNodeAnd:
		return 9
	case rankedNodeXor:
		return 8
	case rankedNodeOr:
		return 7
	case rankedNodeEquiv:
		return 6
	case rankedNodeForward:
		return 5
	case rankedNodeBackward:
		return 4
	}
	return 0
}

func rankedFormulaDelimiter(character byte) bool {
	return booleanFormulaSpace(character) || strings.IndexByte("()[]!~'&*^+|<=>", character) >= 0
}

func parseRankedBooleanNodes(formula string, names []string, ranks int, options RankedBooleanParseOptions) ([]rankedFormulaNode, int, int, error) {
	if ranks < 0 || strings.IndexByte(formula, 0) >= 0 {
		if ranks < 0 {
			return nil, 0, 0, ErrDimension
		}
		return nil, 0, 0, booleanFormulaError(strings.IndexByte(formula, 0), "embedded NUL")
	}
	lookup := make(map[string]int, len(names))
	for index, name := range names {
		if !validRankedBooleanName(name) {
			return nil, 0, 0, fmt.Errorf("%w: invalid variable name at index %d", ErrDimension, index)
		}
		if _, exists := lookup[name]; !exists {
			lookup[name] = index
		}
	}
	nodes := make([]rankedFormulaNode, 0, min(len(formula), options.MaxNodes))
	values := make([]int, 0, min(len(formula), options.MaxDepth))
	operators := make([]rankedFormulaOperator, 0, min(len(formula), options.MaxDepth))
	tokens, depth, hasOperand, implicitAllowed := 0, 0, false, false
	countToken := func(offset int) error {
		if tokens == options.MaxTokens {
			return fmt.Errorf("%w at byte %d: token limit", ErrLimit, offset)
		}
		tokens++
		return nil
	}
	addNode := func(node rankedFormulaNode, offset int) (int, error) {
		if len(nodes) == options.MaxNodes {
			return 0, fmt.Errorf("%w at byte %d: node limit", ErrLimit, offset)
		}
		nodes = append(nodes, node)
		return len(nodes) - 1, nil
	}
	apply := func(operator rankedFormulaOperator) error {
		if operator.kind == rankedNodeNot {
			if len(values) < 1 {
				return booleanFormulaError(operator.offset, "negation has no operand")
			}
			child := values[len(values)-1]
			node, err := addNode(rankedFormulaNode{kind: rankedNodeNot, left: child, right: -1}, operator.offset)
			if err != nil {
				return err
			}
			values[len(values)-1] = node
			return nil
		}
		if len(values) < 2 {
			return booleanFormulaError(operator.offset, "binary operator has too few operands")
		}
		right, left := values[len(values)-1], values[len(values)-2]
		values = values[:len(values)-2]
		node, err := addNode(rankedFormulaNode{kind: operator.kind, left: left, right: right, variable: -1}, operator.offset)
		if err != nil {
			return err
		}
		values = append(values, node)
		return nil
	}
	popNegations := func() error {
		for len(operators) > 0 && operators[len(operators)-1].kind == rankedNodeNot {
			operator := operators[len(operators)-1]
			operators = operators[:len(operators)-1]
			if err := apply(operator); err != nil {
				return err
			}
		}
		return nil
	}
	pushBinary := func(kind byte, offset int) error {
		precedence := rankedFormulaPrecedence(kind)
		for len(operators) > 0 {
			top := operators[len(operators)-1]
			if top.kind == rankedNodeMark || rankedFormulaPrecedence(top.kind) < precedence {
				break
			}
			operators = operators[:len(operators)-1]
			if err := apply(top); err != nil {
				return err
			}
		}
		operators = append(operators, rankedFormulaOperator{kind: kind, offset: offset})
		return nil
	}
	lower := func(root, offset int) error {
		work := []int{root}
		for len(work) > 0 {
			index := work[len(work)-1]
			work = work[:len(work)-1]
			node := &nodes[index]
			if node.kind == rankedNodeVariable {
				if node.variable < len(names) {
					return booleanFormulaError(offset, "rank underflow")
				}
				node.variable -= len(names)
			} else {
				if node.left >= 0 {
					work = append(work, node.left)
				}
				if node.right >= 0 {
					work = append(work, node.right)
				}
			}
		}
		return nil
	}

	for offset := 0; offset < len(formula); {
		if options.Stop != nil && options.Stop() {
			return nil, 0, 0, ErrCancelled
		}
		character := formula[offset]
		if booleanFormulaSpace(character) {
			offset++
			continue
		}
		if err := countToken(offset); err != nil {
			return nil, 0, 0, err
		}
		switch character {
		case '0', '1':
			if hasOperand {
				return nil, 0, 0, booleanFormulaError(offset, "missing operator before constant")
			}
			kind := rankedNodeFalse
			if character == '1' {
				kind = rankedNodeTrue
			}
			node, err := addNode(rankedFormulaNode{kind: kind, variable: -1, left: -1, right: -1}, offset)
			if err != nil {
				return nil, 0, 0, err
			}
			values = append(values, node)
			hasOperand = true
			implicitAllowed = false
			if err = popNegations(); err != nil {
				return nil, 0, 0, err
			}
			offset++
		case '!', '~':
			if hasOperand {
				if !implicitAllowed {
					return nil, 0, 0, booleanFormulaError(offset, "missing operator after constant")
				}
				if err := pushBinary(rankedNodeAnd, offset); err != nil {
					return nil, 0, 0, err
				}
				hasOperand = false
			}
			operators = append(operators, rankedFormulaOperator{kind: rankedNodeNot, offset: offset})
			offset++
		case '\'':
			if !hasOperand || len(values) == 0 {
				return nil, 0, 0, booleanFormulaError(offset, "suffix negation has no operand")
			}
			node, err := addNode(rankedFormulaNode{kind: rankedNodeNot, left: values[len(values)-1], right: -1, variable: -1}, offset)
			if err != nil {
				return nil, 0, 0, err
			}
			values[len(values)-1] = node
			offset++
		case '&', '*', '^', '+', '|':
			if !hasOperand {
				return nil, 0, 0, booleanFormulaError(offset, "binary operator has no left operand")
			}
			kind := rankedNodeAnd
			if character == '^' {
				kind = rankedNodeXor
			} else if character == '+' || character == '|' {
				kind = rankedNodeOr
			}
			if err := pushBinary(kind, offset); err != nil {
				return nil, 0, 0, err
			}
			hasOperand = false
			offset++
		case '<', '=':
			if !hasOperand {
				return nil, 0, 0, booleanFormulaError(offset, "relation has no left operand")
			}
			kind, width := byte(0), 0
			if character == '=' && offset+1 < len(formula) && formula[offset+1] == '>' {
				kind, width = rankedNodeForward, 2
			}
			if character == '<' && offset+1 < len(formula) {
				if formula[offset+1] == '=' {
					kind, width = rankedNodeBackward, 2
					if offset+2 < len(formula) && formula[offset+2] == '>' {
						kind, width = rankedNodeEquiv, 3
					}
				} else if formula[offset+1] == '+' && offset+2 < len(formula) && formula[offset+2] == '>' {
					kind, width = rankedNodeXor, 3
				}
			}
			if width == 0 {
				return nil, 0, 0, booleanFormulaError(offset, "malformed relation operator")
			}
			if err := pushBinary(kind, offset); err != nil {
				return nil, 0, 0, err
			}
			hasOperand = false
			offset += width
		case '(', '[':
			if hasOperand {
				if !implicitAllowed {
					return nil, 0, 0, booleanFormulaError(offset, "missing operator after constant")
				}
				if err := pushBinary(rankedNodeAnd, offset); err != nil {
					return nil, 0, 0, err
				}
			}
			if depth == options.MaxDepth {
				return nil, 0, 0, fmt.Errorf("%w at byte %d: nesting limit", ErrLimit, offset)
			}
			depth++
			operators = append(operators, rankedFormulaOperator{kind: rankedNodeMark, delimiter: character, offset: offset})
			hasOperand = false
			offset++
		case ')', ']':
			if !hasOperand {
				return nil, 0, 0, booleanFormulaError(offset, "closing delimiter has no operand")
			}
			opening := byte('(')
			if character == ']' {
				opening = '['
			}
			found := false
			for len(operators) > 0 {
				top := operators[len(operators)-1]
				operators = operators[:len(operators)-1]
				if top.kind == rankedNodeMark {
					if top.delimiter != opening {
						return nil, 0, 0, booleanFormulaError(offset, "mismatched delimiter")
					}
					found, depth = true, depth-1
					break
				}
				if err := apply(top); err != nil {
					return nil, 0, 0, err
				}
			}
			if !found {
				return nil, 0, 0, booleanFormulaError(offset, "closing delimiter has no opening delimiter")
			}
			if character == ']' {
				if len(values) == 0 {
					return nil, 0, 0, ErrInvariant
				}
				if err := lower(values[len(values)-1], offset); err != nil {
					return nil, 0, 0, err
				}
			}
			if err := popNegations(); err != nil {
				return nil, 0, 0, err
			}
			hasOperand = true
			implicitAllowed = true
			offset++
		default:
			if hasOperand {
				if !implicitAllowed {
					return nil, 0, 0, booleanFormulaError(offset, "missing operator after constant")
				}
				if err := pushBinary(rankedNodeAnd, offset); err != nil {
					return nil, 0, 0, err
				}
				hasOperand = false
			}
			end := offset
			for end < len(formula) && !rankedFormulaDelimiter(formula[end]) {
				end++
			}
			if end == offset {
				return nil, 0, 0, booleanFormulaError(offset, "unknown symbol %q", character)
			}
			name := formula[offset:end]
			variable, found := lookup[name]
			if !found {
				return nil, 0, 0, booleanFormulaError(offset, "unknown variable %q", name)
			}
			node, err := addNode(rankedFormulaNode{kind: rankedNodeVariable, variable: ranks*len(names) + variable, left: -1, right: -1}, offset)
			if err != nil {
				return nil, 0, 0, err
			}
			values = append(values, node)
			hasOperand = true
			implicitAllowed = true
			if err = popNegations(); err != nil {
				return nil, 0, 0, err
			}
			offset = end
		}
	}
	if !hasOperand {
		return nil, 0, 0, booleanFormulaError(len(formula), "formula has no final operand")
	}
	if depth != 0 {
		return nil, 0, 0, booleanFormulaError(len(formula), "opening delimiter has no closing delimiter")
	}
	for len(operators) > 0 {
		top := operators[len(operators)-1]
		operators = operators[:len(operators)-1]
		if top.kind == rankedNodeMark {
			return nil, 0, 0, booleanFormulaError(top.offset, "opening delimiter has no closing delimiter")
		}
		if err := apply(top); err != nil {
			return nil, 0, 0, err
		}
	}
	if len(values) != 1 {
		return nil, 0, 0, booleanFormulaError(0, "formula leaves %d values", len(values))
	}
	return nodes, values[0], tokens, nil
}

func evaluateRankedBooleanNodes(nodes []rankedFormulaNode, root, variables int, options RankedBooleanParseOptions) (*TruthTable, error) {
	values := make([]*TruthTable, len(nodes))
	for index, node := range nodes {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		var table *TruthTable
		var err error
		switch node.kind {
		case rankedNodeFalse, rankedNodeTrue:
			table, err = NewConstantTruthTable(variables, node.kind == rankedNodeTrue, options.MaxWords)
		case rankedNodeVariable:
			table, err = NewVariableTruthTable(variables, node.variable, options.MaxWords)
		case rankedNodeNot:
			if node.left < 0 || values[node.left] == nil {
				return nil, ErrInvariant
			}
			table = values[node.left].Not()
			values[node.left] = nil
		default:
			if node.left < 0 || node.right < 0 || values[node.left] == nil || values[node.right] == nil {
				return nil, ErrInvariant
			}
			left, right := values[node.left], values[node.right]
			switch node.kind {
			case rankedNodeAnd:
				table, err = left.And(right, false, false)
			case rankedNodeXor:
				table, err = left.Xor(right, false)
			case rankedNodeOr:
				table, err = left.Or(right)
			case rankedNodeEquiv:
				table, err = left.Xor(right, true)
			case rankedNodeForward:
				table, err = left.Not().Or(right)
			case rankedNodeBackward:
				table, err = left.Or(right.Not())
			default:
				return nil, ErrInvariant
			}
			values[node.left], values[node.right] = nil, nil
		}
		if err != nil {
			return nil, err
		}
		values[index] = table
	}
	if root < 0 || root >= len(values) || values[root] == nil {
		return nil, ErrInvariant
	}
	return values[root], nil
}

// ParseRankedBooleanFormula parses the full parseCore.c grammar into a packed
// truth table. ranks is the number of lower rank blocks below bare variables.
func ParseRankedBooleanFormula(formula string, variableNames []string, ranks int, options RankedBooleanParseOptions) (RankedBooleanFormula, error) {
	options, err := normalizeRankedBooleanOptions(options)
	if err != nil {
		return RankedBooleanFormula{}, err
	}
	if ranks < 0 || len(variableNames) > 0 && ranks > (36/len(variableNames))-1 {
		return RankedBooleanFormula{}, ErrDimension
	}
	totalVariables := len(variableNames) * (ranks + 1)
	if _, err = TruthTableWordCount(totalVariables); err != nil {
		return RankedBooleanFormula{}, err
	}
	nodes, root, tokens, err := parseRankedBooleanNodes(formula, variableNames, ranks, options)
	if err != nil {
		return RankedBooleanFormula{}, err
	}
	table, err := evaluateRankedBooleanNodes(nodes, root, totalVariables, options)
	if err != nil {
		return RankedBooleanFormula{}, err
	}
	return RankedBooleanFormula{Table: table, Variables: len(variableNames), Ranks: ranks, Tokens: tokens, Nodes: len(nodes), VariableNames: append([]string(nil), variableNames...)}, nil
}
