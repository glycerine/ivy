package abc

// Disjoint-support decomposition from src/opt/dsc/dsc.c and dsc.h.

import "slices"

const (
	dscSourceMaxVariables = 16
	dscSourceMaxFormula   = 64
)

// DSCOptions bounds decomposition. Zero limits select the source bounds.
type DSCOptions struct {
	MaxVariables    int
	MaxWords        int
	MaxFormulaBytes int
	Stop            func() bool
}

func normalizeDSCOptions(options DSCOptions) (DSCOptions, error) {
	if options.MaxVariables < 0 || options.MaxWords < 0 || options.MaxFormulaBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = dscSourceMaxVariables
	}
	if options.MaxVariables > dscSourceMaxVariables {
		options.MaxVariables = dscSourceMaxVariables
	}
	if options.MaxWords == 0 {
		options.MaxWords = 1 << 20
	}
	if options.MaxFormulaBytes == 0 {
		// DSC_MAX_STR includes the terminating NUL in the source.
		options.MaxFormulaBytes = dscSourceMaxFormula - 1
	}
	return options, nil
}

// DSCWorkspace is the owned replacement for Dsc_alloc_pool/Dsc_free_pool.
// The C pool changes allocation only; it does not carry algorithmic state.
type DSCWorkspace struct {
	variables int
	words     int
	closed    bool
}

func NewDSCWorkspace(variables, maxWords int) (*DSCWorkspace, error) {
	if variables < 0 || variables > dscSourceMaxVariables || maxWords < 0 {
		return nil, ErrDimension
	}
	words, err := TruthTableWordCount(variables)
	if err != nil {
		return nil, err
	}
	if maxWords == 0 {
		maxWords = 1 << 20
	}
	needed, err := checkedSize(0, 3*variables, words, maxWords)
	if err != nil {
		return nil, err
	}
	return &DSCWorkspace{variables: variables, words: needed}, nil
}

func (workspace *DSCWorkspace) CapacityWords() (int, error) {
	if workspace == nil || workspace.closed {
		return 0, ErrInvariant
	}
	return workspace.words, nil
}

func (workspace *DSCWorkspace) Close() {
	if workspace != nil {
		workspace.closed = true
		workspace.variables = 0
		workspace.words = 0
	}
}

// DSCResult reports the source's zero/nonzero disposition without using an
// empty formula as an error sentinel. Formula is empty when Complete is false.
type DSCResult struct {
	Formula  string
	Complete bool
	Support  int
}

type dscLiteral struct {
	variable int
	positive bool
}

type dscNode struct {
	negative, positive, difference *TruthTable
	on, off                        []dscLiteral
	expression                     string
}

func dscStopped(options DSCOptions) bool {
	return options.Stop != nil && options.Stop()
}

func dscFormulaJoin(begin, end byte, first string, firstPositive bool, second string, secondPositive bool, limit int) (string, error) {
	length := 2 + len(first) + len(second)
	if !firstPositive {
		length++
	}
	if !secondPositive {
		length++
	}
	if length > limit {
		return "", ErrLimit
	}
	result := make([]byte, 0, length)
	result = append(result, begin)
	if !firstPositive {
		result = append(result, '!')
	}
	result = append(result, first...)
	if !secondPositive {
		result = append(result, '!')
	}
	result = append(result, second...)
	result = append(result, end)
	return string(result), nil
}

func dscCofactorCube(truth *TruthTable, cube []dscLiteral, options DSCOptions) (*TruthTable, error) {
	result := truth.Clone()
	var err error
	for _, literal := range cube {
		if dscStopped(options) {
			return nil, ErrCancelled
		}
		result, err = result.Cofactor(literal.variable, literal.positive)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func dscAndTest(first, second *dscNode) (firstPositive, secondPositive, found bool) {
	switch {
	case first.negative.Equal(second.negative):
		return true, true, true
	case first.negative.Equal(second.positive):
		return true, false, true
	case first.positive.Equal(second.negative):
		return false, true, true
	case first.positive.Equal(second.positive):
		return false, false, true
	default:
		return false, false, false
	}
}

func dscAndGroup(first *dscNode, firstPositive bool, second *dscNode, secondPositive bool, options DSCOptions) (*dscNode, error) {
	expression, err := dscFormulaJoin('(', ')', first.expression, firstPositive, second.expression, secondPositive, options.MaxFormulaBytes)
	if err != nil {
		return nil, err
	}
	firstOff, firstOn := first.off, first.on
	if !firstPositive {
		firstOff, firstOn = first.on, first.off
	}
	secondOff, secondOn := second.off, second.on
	if !secondPositive {
		secondOff, secondOn = second.on, second.off
	}
	result := &dscNode{expression: expression}
	if len(firstOff) <= len(secondOff) {
		result.off = slices.Clone(firstOff)
	} else {
		result.off = slices.Clone(secondOff)
	}
	if firstPositive {
		result.negative = first.negative.Clone()
	} else {
		result.negative = first.positive.Clone()
	}
	result.on = append(slices.Clone(firstOn), secondOn...)
	if len(firstOn) >= len(secondOn) {
		base := first.positive
		if !firstPositive {
			base = first.negative
		}
		result.positive, err = dscCofactorCube(base, secondOn, options)
	} else {
		base := second.positive
		if !secondPositive {
			base = second.negative
		}
		result.positive, err = dscCofactorCube(base, firstOn, options)
	}
	if err != nil {
		return nil, err
	}
	base := second.negative
	if !secondPositive {
		base = second.positive
	}
	result.difference, err = base.Xor(result.positive, false)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func dscXorGroup(first, second *dscNode, options DSCOptions) (*dscNode, error) {
	expression, err := dscFormulaJoin('[', ']', first.expression, true, second.expression, true, options.MaxFormulaBytes)
	if err != nil {
		return nil, err
	}
	type choice struct {
		node     *dscNode
		positive bool
		cube     []dscLiteral
	}
	choices := []choice{{first, false, first.off}, {first, true, first.on}, {second, false, second.off}, {second, true, second.on}}
	selected := choices[0]
	for _, candidate := range choices[1:] {
		if len(candidate.cube) < len(selected.cube) {
			selected = candidate
		}
	}
	result := &dscNode{expression: expression, difference: first.difference.Clone()}
	if selected.node == first {
		if selected.positive {
			result.negative, err = dscCofactorCube(second.positive, first.on, options)
			if err == nil {
				result.positive, err = dscCofactorCube(second.negative, first.on, options)
			}
		} else {
			result.negative, err = dscCofactorCube(second.negative, first.off, options)
			if err == nil {
				result.positive, err = dscCofactorCube(second.positive, first.off, options)
			}
		}
	} else if selected.positive {
		result.negative, err = dscCofactorCube(first.positive, second.on, options)
		if err == nil {
			result.positive, err = dscCofactorCube(first.negative, second.on, options)
		}
	} else {
		result.negative, err = dscCofactorCube(first.negative, second.off, options)
		if err == nil {
			result.positive, err = dscCofactorCube(first.positive, second.off, options)
		}
	}
	if err != nil {
		return nil, err
	}
	if len(first.off)+len(second.off) <= len(first.on)+len(second.on) {
		result.off = append(slices.Clone(first.off), second.off...)
	} else {
		result.off = append(slices.Clone(first.on), second.on...)
	}
	if len(first.off)+len(second.on) <= len(first.on)+len(second.off) {
		result.on = append(slices.Clone(first.off), second.on...)
	} else {
		result.on = append(slices.Clone(first.on), second.off...)
	}
	return result, nil
}

// DecomposeDSC implements Dsc_Decompose. A non-decomposable function returns a
// zero-error result with Complete false, matching the source's negative status.
func DecomposeDSC(truth *TruthTable, workspace *DSCWorkspace, options DSCOptions) (DSCResult, error) {
	options, err := normalizeDSCOptions(options)
	if err != nil {
		return DSCResult{}, err
	}
	if truth == nil {
		return DSCResult{}, ErrDimension
	}
	if err = truth.Check(); err != nil {
		return DSCResult{}, err
	}
	variables := truth.Variables()
	if variables < 0 || variables > options.MaxVariables {
		return DSCResult{}, ErrLimit
	}
	words, err := TruthTableWordCount(variables)
	if err != nil {
		return DSCResult{}, err
	}
	if words > options.MaxWords {
		return DSCResult{}, ErrLimit
	}
	if workspace != nil {
		if workspace.closed || workspace.variables != variables || workspace.words < 3*variables*words {
			return DSCResult{}, ErrDimension
		}
	}
	if dscStopped(options) {
		return DSCResult{}, ErrCancelled
	}
	newNodes := make([]*dscNode, 0, variables)
	for variable := 0; variable < variables; variable++ {
		if dscStopped(options) {
			return DSCResult{}, ErrCancelled
		}
		negative, cofactorErr := truth.Cofactor(variable, false)
		if cofactorErr != nil {
			return DSCResult{}, cofactorErr
		}
		if negative.Equal(truth) {
			continue
		}
		positive, cofactorErr := truth.Cofactor(variable, true)
		if cofactorErr != nil {
			return DSCResult{}, cofactorErr
		}
		difference, cofactorErr := negative.Xor(positive, false)
		if cofactorErr != nil {
			return DSCResult{}, cofactorErr
		}
		newNodes = append(newNodes, &dscNode{
			negative: negative, positive: positive, difference: difference,
			on: []dscLiteral{{variable, true}}, off: []dscLiteral{{variable, false}},
			expression: string(byte('a' + variable)),
		})
	}
	support := len(newNodes)
	if support == 0 {
		if truth.IsConstant(false) {
			return DSCResult{Formula: "0", Complete: true}, nil
		}
		if truth.IsConstant(true) {
			return DSCResult{Formula: "1", Complete: true}, nil
		}
		return DSCResult{}, ErrInvariant
	}
	oldNodes := make([]*dscNode, 0, support)
	groups := 0
	for len(newNodes) > 0 {
		next := make([]*dscNode, 0, len(newNodes))
		for _, first := range newNodes {
			if dscStopped(options) {
				return DSCResult{}, ErrCancelled
			}
			var group *dscNode
			for index := 0; index < len(oldNodes); index++ {
				if dscStopped(options) {
					return DSCResult{}, ErrCancelled
				}
				second := oldNodes[index]
				firstPositive, secondPositive, found := dscAndTest(first, second)
				if found {
					group, err = dscAndGroup(first, firstPositive, second, secondPositive, options)
				} else if first.difference.Equal(second.difference) {
					group, err = dscXorGroup(first, second, options)
				}
				if err != nil {
					return DSCResult{}, err
				}
				if group != nil {
					groups++
					if groups >= dscSourceMaxVariables {
						return DSCResult{}, ErrInvariant
					}
					oldNodes[index] = oldNodes[len(oldNodes)-1]
					oldNodes = oldNodes[:len(oldNodes)-1]
					break
				}
			}
			if group != nil {
				next = append(next, group)
			} else {
				oldNodes = append(oldNodes, first)
			}
		}
		newNodes = next
	}
	if len(oldNodes) != 1 {
		return DSCResult{Support: support}, nil
	}
	solution := oldNodes[0]
	if solution.negative.IsConstant(false) && solution.positive.IsConstant(true) {
		return DSCResult{Formula: solution.expression, Complete: true, Support: support}, nil
	}
	if solution.negative.IsConstant(true) && solution.positive.IsConstant(false) {
		if len(solution.expression)+1 > options.MaxFormulaBytes {
			return DSCResult{}, ErrLimit
		}
		return DSCResult{Formula: "!" + solution.expression, Complete: true, Support: support}, nil
	}
	return DSCResult{}, ErrInvariant
}

// DSCMatches is the owned checked replacement for Dsc_ComputeMatches.
func DSCMatches(formula string, maxBytes int) ([]int, error) {
	matches, err := DauDSDMatches(formula, maxBytes)
	if err != nil {
		return nil, err
	}
	// Dsc_ComputeMatches stores only opener-to-closer entries. All ordinary
	// bytes and closing delimiters remain zero in its static result array.
	result := make([]int, len(matches))
	for index := range formula {
		switch formula[index] {
		case '(', '[', '<', '{':
			result[index] = matches[index]
		}
	}
	return result, nil
}

// DSCAndCount implements Dsc_CountAnds and its recursive helper.
func DSCAndCount(formula string, maxBytes int) (int, error) {
	parsed, err := ParseDauDSD(formula, DauDSDOptions{MaxBytes: maxBytes, MaxVariables: dscSourceMaxVariables})
	if err != nil {
		return 0, err
	}
	return parsed.AndCount()
}
