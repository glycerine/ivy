package abc

// Read-polarity-once factorization from src/bool/rpo/literal.h and rpo.c.

import (
	"strconv"
	"strings"
)

type RPOOperator uint8

const (
	RPOOperatorNone RPOOperator = iota
	RPOOperatorAND
	RPOOperatorOR
	RPOOperatorXOR
)

// RPOLiteral owns one Boolean function, its transition set, and the source
// expression. Returned truth tables are cloned so callers cannot mutate it.
type RPOLiteral struct {
	function   *TruthTable
	transition *TruthTable
	expression string
}

func (literal *RPOLiteral) check() error {
	if literal == nil || literal.function == nil || literal.transition == nil || literal.expression == "" {
		return ErrDimension
	}
	if err := literal.function.Check(); err != nil {
		return err
	}
	if err := literal.transition.Check(); err != nil {
		return err
	}
	if literal.function.Variables() != literal.transition.Variables() {
		return ErrInvariant
	}
	return nil
}

func (literal *RPOLiteral) Function() (*TruthTable, error) {
	if err := literal.check(); err != nil {
		return nil, err
	}
	return literal.function.Clone(), nil
}

func (literal *RPOLiteral) Transition() (*TruthTable, error) {
	if err := literal.check(); err != nil {
		return nil, err
	}
	return literal.transition.Clone(), nil
}

func (literal *RPOLiteral) Expression() (string, error) {
	if err := literal.check(); err != nil {
		return "", err
	}
	return literal.expression, nil
}

func (literal *RPOLiteral) Clone() (*RPOLiteral, error) {
	if err := literal.check(); err != nil {
		return nil, err
	}
	return &RPOLiteral{function: literal.function.Clone(), transition: literal.transition.Clone(), expression: literal.expression}, nil
}

func RPOPositiveTransition(function *TruthTable, variable int) (*TruthTable, error) {
	if function == nil {
		return nil, ErrDimension
	}
	zero, err := function.Cofactor(variable, false)
	if err != nil {
		return nil, err
	}
	one, err := function.Cofactor(variable, true)
	if err != nil {
		return nil, err
	}
	return zero.And(one, true, false)
}

func RPONegativeTransition(function *TruthTable, variable int) (*TruthTable, error) {
	if function == nil {
		return nil, ErrDimension
	}
	zero, err := function.Cofactor(variable, false)
	if err != nil {
		return nil, err
	}
	one, err := function.Cofactor(variable, true)
	if err != nil {
		return nil, err
	}
	return one.And(zero, true, false)
}

// NewRPOLiteral is Lit_Alloc. The Boolean result reports whether the source
// transition is nonempty and therefore produces a literal.
func NewRPOLiteral(target *TruthTable, variable int, positive bool) (*RPOLiteral, bool, error) {
	if target == nil || variable < 0 || variable >= target.Variables() {
		return nil, false, ErrDimension
	}
	var transition *TruthTable
	var err error
	if positive {
		transition, err = RPOPositiveTransition(target, variable)
	} else {
		transition, err = RPONegativeTransition(target, variable)
	}
	if err != nil {
		return nil, false, err
	}
	if transition.IsConstant(false) {
		return nil, false, nil
	}
	function, err := NewVariableTruthTable(target.Variables(), variable, len(target.words))
	if err != nil {
		return nil, false, err
	}
	expression := string(rune('a' + variable))
	if !positive {
		function = function.Not()
		expression = "!" + expression
	}
	return &RPOLiteral{function: function, transition: transition, expression: expression}, true, nil
}

func newRPOConstant(target *TruthTable, value bool) (*RPOLiteral, error) {
	if target == nil {
		return nil, ErrDimension
	}
	if !target.IsConstant(value) {
		return nil, ErrDimension
	}
	expression := "0"
	if value {
		expression = "1"
	}
	// The source aliases transition to function for constants. Owned clones
	// retain that observable value without sharing mutable storage.
	return &RPOLiteral{function: target.Clone(), transition: target.Clone(), expression: expression}, nil
}

func GroupRPOLiterals(first, second *RPOLiteral, operator RPOOperator, maxExpressionBytes int) (*RPOLiteral, error) {
	if err := first.check(); err != nil {
		return nil, err
	}
	if err := second.check(); err != nil {
		return nil, err
	}
	if first.function.Variables() != second.function.Variables() {
		return nil, ErrDimension
	}
	if maxExpressionBytes < 0 {
		return nil, ErrDimension
	}
	if maxExpressionBytes == 0 {
		maxExpressionBytes = 1 << 20
	}
	needed := len(first.expression) + len(second.expression) + 3
	if needed > maxExpressionBytes {
		return nil, ErrLimit
	}
	var function *TruthTable
	var err error
	operatorByte := byte(0)
	switch operator {
	case RPOOperatorAND:
		function, err = first.function.And(second.function, false, false)
		operatorByte = '*'
	case RPOOperatorOR:
		function, err = first.function.Or(second.function)
		operatorByte = '+'
	default:
		return nil, ErrUnsupported
	}
	if err != nil {
		return nil, err
	}
	transition, err := first.transition.Or(second.transition)
	if err != nil {
		return nil, err
	}
	var builder strings.Builder
	builder.Grow(needed)
	builder.WriteByte('(')
	builder.WriteString(first.expression)
	builder.WriteByte(operatorByte)
	builder.WriteString(second.expression)
	builder.WriteByte(')')
	return &RPOLiteral{function: function, transition: transition, expression: builder.String()}, nil
}

func RPOCheckANDGroup(first, second *RPOLiteral) (bool, error) {
	if err := first.check(); err != nil {
		return false, err
	}
	if err := second.check(); err != nil {
		return false, err
	}
	firstTest, err := first.transition.And(second.function, false, true)
	if err != nil {
		return false, err
	}
	if !firstTest.IsConstant(false) {
		return false, nil
	}
	secondTest, err := second.transition.And(first.function, false, true)
	return err == nil && secondTest.IsConstant(false), err
}

func RPOCheckORGroup(first, second *RPOLiteral) (bool, error) {
	if err := first.check(); err != nil {
		return false, err
	}
	if err := second.check(); err != nil {
		return false, err
	}
	firstTest, err := first.transition.And(second.function, false, false)
	if err != nil {
		return false, err
	}
	if !firstTest.IsConstant(false) {
		return false, nil
	}
	secondTest, err := second.transition.And(first.function, false, false)
	return err == nil && secondTest.IsConstant(false), err
}

type RPOEdge struct {
	Visited  bool
	Operator RPOOperator
	First    int
	Second   int
}

func NewRPOEdge(operator RPOOperator, first, second int, degrees []int) (RPOEdge, error) {
	if operator != RPOOperatorAND && operator != RPOOperatorOR || first < 0 || second < 0 || first >= len(degrees) || second >= len(degrees) || first == second {
		return RPOEdge{}, ErrDimension
	}
	edge := RPOEdge{Operator: operator, First: first, Second: second}
	degrees[first]++
	degrees[second]++
	return edge, nil
}

func rpoMinimumEdge(edges []RPOEdge, degrees []int) (int, error) {
	minimum, minimumCost, minimumVertex := -1, int(^uint(0)>>1), int(^uint(0)>>1)
	for index, edge := range edges {
		if edge.First < 0 || edge.Second < 0 || edge.First >= len(degrees) || edge.Second >= len(degrees) || edge.First == edge.Second {
			return -1, ErrInvariant
		}
		if edge.Visited {
			continue
		}
		cost := degrees[edge.First] + degrees[edge.Second]
		vertex := min(edge.First, edge.Second)
		if cost < minimumCost || cost == minimumCost && vertex < minimumVertex {
			minimum, minimumCost, minimumVertex = index, cost, vertex
		}
	}
	return minimum, nil
}

func FormatRPOEdge(edge RPOEdge) (string, error) {
	if edge.First < 0 || edge.Second < 0 || edge.First == edge.Second || edge.Operator != RPOOperatorAND && edge.Operator != RPOOperatorOR {
		return "", ErrDimension
	}
	return "Edge (" + strconv.Itoa(edge.First) + "," + strconv.Itoa(edge.Second) + ")/" + strconv.Itoa(int(edge.Operator)) + "\n", nil
}

// FormatRPOTruth replaces Lit_PrintTT with owned text. It prints ABC's 32-bit
// truth words from high to low. The source loops over nVars rather than the word
// count and can read outside the table; this checked form uses the actual count.
func FormatRPOTruth(table *TruthTable) (string, error) {
	if err := table.Check(); err != nil {
		return "", err
	}
	words := make([]uint32, 0, len(table.words)*2)
	for _, word := range table.words {
		words = append(words, uint32(word), uint32(word>>32))
	}
	if table.variables < 6 {
		words = words[:1]
	}
	var builder strings.Builder
	builder.Grow(8 * len(words))
	for index := len(words) - 1; index >= 0; index-- {
		text := strconv.FormatUint(uint64(words[index]), 16)
		for padding := len(text); padding < 8; padding++ {
			builder.WriteByte('0')
		}
		builder.WriteString(strings.ToUpper(text))
	}
	return builder.String(), nil
}

type RPOFactorOptions struct {
	Threshold          int
	MaxVariables       int
	MaxLiterals        int
	MaxEdges           int
	MaxSteps           int
	MaxExpressionBytes int
	Stop               func() bool
}

type RPOFactorStatistics struct {
	Steps            int
	InitialLiterals  int
	EdgesConsidered  int
	ThresholdReached bool
}

type rpoFactorizer struct {
	target  *TruthTable
	options RPOFactorOptions
	stats   RPOFactorStatistics
}

func normalizeRPOFactorOptions(options RPOFactorOptions) (RPOFactorOptions, error) {
	if options.Threshold < -1 || options.MaxVariables < 0 || options.MaxLiterals < 0 || options.MaxEdges < 0 || options.MaxSteps < 0 || options.MaxExpressionBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 16
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 64
	}
	if options.MaxEdges == 0 {
		options.MaxEdges = 1 << 16
	}
	if options.MaxSteps == 0 {
		options.MaxSteps = 1 << 20
	}
	if options.MaxExpressionBytes == 0 {
		options.MaxExpressionBytes = 1 << 20
	}
	return options, nil
}

// FactorizeRPO implements Rpo_Factorize/Rpo_Recursion. A nil literal with nil
// error means no read-polarity-once factorization was found before Threshold.
func FactorizeRPO(target *TruthTable, options RPOFactorOptions) (*RPOLiteral, RPOFactorStatistics, error) {
	options, err := normalizeRPOFactorOptions(options)
	if err != nil {
		return nil, RPOFactorStatistics{}, err
	}
	if target == nil {
		return nil, RPOFactorStatistics{}, ErrDimension
	}
	if err = target.Check(); err != nil {
		return nil, RPOFactorStatistics{}, err
	}
	if target.Variables() > options.MaxVariables {
		return nil, RPOFactorStatistics{}, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return nil, RPOFactorStatistics{}, ErrCancelled
	}
	if target.IsConstant(false) {
		literal, constantErr := newRPOConstant(target, false)
		return literal, RPOFactorStatistics{}, constantErr
	}
	if target.IsConstant(true) {
		literal, constantErr := newRPOConstant(target, true)
		return literal, RPOFactorStatistics{}, constantErr
	}
	literals := make([]*RPOLiteral, 0, 2*target.Variables())
	for variable := target.Variables() - 1; variable >= 0; variable-- {
		for _, positive := range []bool{true, false} {
			literal, present, literalErr := NewRPOLiteral(target, variable, positive)
			if literalErr != nil {
				return nil, RPOFactorStatistics{}, literalErr
			}
			if present {
				literals = append(literals, literal)
			}
		}
	}
	if len(literals) > options.MaxLiterals {
		return nil, RPOFactorStatistics{}, ErrLimit
	}
	factorizer := &rpoFactorizer{target: target, options: options}
	factorizer.stats.InitialLiterals = len(literals)
	result, err := factorizer.recurse(literals, len(literals))
	return result, factorizer.stats, err
}

func (factorizer *rpoFactorizer) recurse(literals []*RPOLiteral, active int) (*RPOLiteral, error) {
	if factorizer.options.Stop != nil && factorizer.options.Stop() {
		return nil, ErrCancelled
	}
	if factorizer.stats.Steps >= factorizer.options.MaxSteps {
		return nil, ErrLimit
	}
	factorizer.stats.Steps++
	if factorizer.options.Threshold > 0 && factorizer.stats.Steps == factorizer.options.Threshold {
		factorizer.stats.ThresholdReached = true
		return nil, nil
	}
	if active == 1 {
		for _, literal := range literals {
			if literal != nil && literal.function.Equal(factorizer.target) {
				return literal.Clone()
			}
		}
		return nil, nil
	}
	degrees := make([]int, len(literals))
	edges := make([]RPOEdge, 0)
	for first := 0; first < len(literals); first++ {
		if literals[first] == nil {
			continue
		}
		// The C loop starts second at first, but self-edges overflow its
		// n*(n-1)/2 allocation and cannot reduce the active literal count.
		// Starting at first+1 is the only defined form of the intended graph.
		for second := first + 1; second < len(literals); second++ {
			if literals[second] == nil {
				continue
			}
			if factorizer.options.Stop != nil && factorizer.options.Stop() {
				return nil, ErrCancelled
			}
			andGrouped, err := RPOCheckANDGroup(literals[first], literals[second])
			if err != nil {
				return nil, err
			}
			orGrouped, err := RPOCheckORGroup(literals[first], literals[second])
			if err != nil {
				return nil, err
			}
			for _, candidate := range []struct {
				present  bool
				operator RPOOperator
			}{{andGrouped, RPOOperatorAND}, {orGrouped, RPOOperatorOR}} {
				if !candidate.present {
					continue
				}
				if len(edges) >= factorizer.options.MaxEdges {
					return nil, ErrLimit
				}
				edge, edgeErr := NewRPOEdge(candidate.operator, first, second, degrees)
				if edgeErr != nil {
					return nil, edgeErr
				}
				edges = append(edges, edge)
				factorizer.stats.EdgesConsidered++
			}
		}
	}
	for {
		index, err := rpoMinimumEdge(edges, degrees)
		if err != nil {
			return nil, err
		}
		if index < 0 {
			return nil, nil
		}
		edges[index].Visited = true
		edge := edges[index]
		low, high := min(edge.First, edge.Second), max(edge.First, edge.Second)
		grouped, err := GroupRPOLiterals(literals[low], literals[high], edge.Operator, factorizer.options.MaxExpressionBytes)
		if err != nil {
			return nil, err
		}
		oldLow, oldHigh := literals[low], literals[high]
		literals[low], literals[high] = grouped, nil
		result, recurseErr := factorizer.recurse(literals, active-1)
		literals[low], literals[high] = oldLow, oldHigh
		if recurseErr != nil || result != nil || factorizer.stats.ThresholdReached {
			return result, recurseErr
		}
	}
}
