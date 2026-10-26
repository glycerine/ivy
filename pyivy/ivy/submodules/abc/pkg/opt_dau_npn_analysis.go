package abc

// Symmetry and input-complement pruning from opt/dau/dauNpn.c.

type DauNPNAnalysisOptions struct {
	MaxComparisons int
	Stop           func() bool
}

type DauNPNAnalysis struct {
	SymmetricPairs           int
	SymmetricSecondVariables []bool
	ComplementEquivalent     []bool
}

func (analysis DauNPNAnalysis) ComplementCount() int {
	count := 0
	for _, equivalent := range analysis.ComplementEquivalent {
		if equivalent {
			count++
		}
	}
	return count
}

func (analysis DauNPNAnalysis) SymmetricSecondMask() (uint64, error) {
	if len(analysis.SymmetricSecondVariables) > 64 {
		return 0, ErrDimension
	}
	var result uint64
	for variable, symmetric := range analysis.SymmetricSecondVariables {
		if symmetric {
			result |= 1 << uint(variable)
		}
	}
	return result, nil
}

type dauNPNComparisonBudget struct {
	used, maximum int
	stop          func() bool
}

func (budget *dauNPNComparisonBudget) take() error {
	if budget.stop != nil && budget.stop() {
		return ErrCancelled
	}
	if budget.used >= budget.maximum {
		return ErrLimit
	}
	budget.used++
	return nil
}

// AnalyzeDauNPN computes the pruning data used by the source enumerator.
// SymmetricSecondVariables marks j when any i<j is directly symmetric with j.
// ComplementEquivalent[v] is true when flipping v complements the output or
// produces the same function as flipping another input.
func AnalyzeDauNPN(table *TruthTable, options DauNPNAnalysisOptions) (DauNPNAnalysis, error) {
	if err := table.Check(); err != nil {
		return DauNPNAnalysis{}, err
	}
	if options.MaxComparisons < 0 {
		return DauNPNAnalysis{}, ErrDimension
	}
	maximum := options.MaxComparisons
	if maximum == 0 {
		maximum = int(^uint(0) >> 1)
	}
	variables := table.variables
	result := DauNPNAnalysis{
		SymmetricSecondVariables: make([]bool, variables),
		ComplementEquivalent:     make([]bool, variables),
	}
	budget := &dauNPNComparisonBudget{maximum: maximum, stop: options.Stop}
	for first := 0; first < variables; first++ {
		for second := first + 1; second < variables; second++ {
			if err := budget.take(); err != nil {
				return DauNPNAnalysis{}, err
			}
			symmetry, err := table.InputSymmetry(first, second, false)
			if err != nil {
				return DauNPNAnalysis{}, err
			}
			if symmetry&TruthSymmetryDirect != 0 {
				result.SymmetricPairs++
				result.SymmetricSecondVariables[second] = true
			}
		}
	}
	complement := table.Not()
	flipped := make([]*TruthTable, variables)
	for variable := 0; variable < variables; variable++ {
		if err := budget.take(); err != nil {
			return DauNPNAnalysis{}, err
		}
		var err error
		flipped[variable], err = table.FlipVariable(variable)
		if err != nil {
			return DauNPNAnalysis{}, err
		}
		if flipped[variable].Equal(complement) {
			result.ComplementEquivalent[variable] = true
		}
	}
	for variable := 0; variable < variables; variable++ {
		if result.ComplementEquivalent[variable] {
			continue
		}
		for other := 0; other < variables; other++ {
			if other == variable {
				continue
			}
			if err := budget.take(); err != nil {
				return DauNPNAnalysis{}, err
			}
			if flipped[variable].Equal(flipped[other]) {
				result.ComplementEquivalent[variable] = true
				break
			}
		}
	}
	return result, nil
}
