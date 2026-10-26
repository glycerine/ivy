package abc

// Contiguous bound-set decomposition from opt/dau/dauDsd2.c.

type DuaRangeOptions struct {
	MaxWords int
	Stop     func() bool
}

type DuaRangeDecomposition struct {
	Found   bool
	Mask    uint64
	Reduced *TruthTable
}

func normalizeDuaRangeOptions(o DuaRangeOptions) (DuaRangeOptions, error) {
	if o.MaxWords < 0 {
		return o, ErrDimension
	}
	if o.MaxWords == 0 {
		o.MaxWords = 1 << 20
	}
	return o, nil
}

func duaRangeCofactor(t *TruthTable, first, last int, high, bound uint64) ([]bool, error) {
	lowCount := uint64(1) << uint(first)
	result := make([]bool, lowCount)
	for low := uint64(0); low < lowCount; low++ {
		assignment := low | bound<<uint(first) | high<<uint(last)
		value, err := t.Get(assignment)
		if err != nil {
			return nil, err
		}
		result[low] = value
	}
	return result, nil
}

func duaBoolSlicesEqual(left, right []bool) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

// DuaDSDRange analyzes the contiguous bound set [first,last). A successful
// result replaces that set with one selector variable in Reduced; Mask is the
// selector's truth function over the original bound variables.
func DuaDSDRange(t *TruthTable, first, last int, options DuaRangeOptions) (DuaRangeDecomposition, error) {
	var result DuaRangeDecomposition
	o, err := normalizeDuaRangeOptions(options)
	if err != nil {
		return result, err
	}
	if t == nil || first < 0 || last > t.Variables() || last-first <= 2 || last-first >= 7 {
		return result, ErrDimension
	}
	if err = t.Check(); err != nil {
		return result, err
	}
	if o.Stop != nil && o.Stop() {
		return result, ErrCancelled
	}
	boundCount := uint64(1) << uint(last-first)
	highCount := uint64(1) << uint(t.Variables()-last)
	lowCount := uint64(1) << uint(first)
	reduced, err := NewConstantTruthTable(t.Variables()-(last-first)+1, false, o.MaxWords)
	if err != nil {
		return result, err
	}
	var commonMask uint64
	for high := uint64(0); high < highCount; high++ {
		if o.Stop != nil && o.Stop() {
			return DuaRangeDecomposition{}, ErrCancelled
		}
		firstCofactor, err := duaRangeCofactor(t, first, last, high, 0)
		if err != nil {
			return result, err
		}
		var secondCofactor []bool
		var mask uint64
		for bound := uint64(1); bound < boundCount; bound++ {
			cofactor, err := duaRangeCofactor(t, first, last, high, bound)
			if err != nil {
				return result, err
			}
			if duaBoolSlicesEqual(cofactor, firstCofactor) {
				continue
			}
			if secondCofactor == nil {
				secondCofactor = cofactor
			}
			if !duaBoolSlicesEqual(cofactor, secondCofactor) {
				return result, nil
			}
			mask |= uint64(1) << bound
		}
		if secondCofactor == nil || mask == 0 {
			return result, nil
		}
		if high == 0 {
			commonMask = mask
		} else if mask != commonMask {
			return result, nil
		}
		for selector, cofactor := range [][]bool{firstCofactor, secondCofactor} {
			for low := uint64(0); low < lowCount; low++ {
				assignment := low | uint64(selector)<<uint(first) | high<<uint(first+1)
				if cofactor[low] {
					if err = reduced.Set(assignment, true); err != nil {
						return result, err
					}
				}
			}
		}
	}
	result.Found, result.Mask, result.Reduced = true, commonMask, reduced
	return result, nil
}

// DuaComposeRange is the exact inverse representation used to validate and
// consume DuaDSDRange results.
func DuaComposeRange(reduced *TruthTable, first, last int, mask uint64, maxWords int) (*TruthTable, error) {
	if reduced == nil || first < 0 || last-first <= 2 || last-first >= 7 || first >= reduced.Variables() || maxWords < 0 {
		return nil, ErrDimension
	}
	if err := reduced.Check(); err != nil {
		return nil, err
	}
	boundVariables := last - first
	variables := reduced.Variables() + boundVariables - 1
	if last > variables || mask == 0 || mask>>(uint64(1)<<uint(boundVariables)) != 0 {
		return nil, ErrDimension
	}
	result, err := NewConstantTruthTable(variables, false, maxWords)
	if err != nil {
		return nil, err
	}
	assignments := uint64(1) << uint(variables)
	lowMask := uint64(1)<<uint(first) - 1
	boundMask := uint64(1)<<uint(boundVariables) - 1
	for assignment := uint64(0); assignment < assignments; assignment++ {
		low := assignment & lowMask
		bound := assignment >> uint(first) & boundMask
		high := assignment >> uint(last)
		selector := mask >> bound & 1
		reducedAssignment := low | selector<<uint(first) | high<<uint(first+1)
		value, err := reduced.Get(reducedAssignment)
		if err != nil {
			return nil, err
		}
		if value {
			if err = result.Set(assignment, true); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

type DuaRangeZeroDecomposition struct {
	Found         bool
	BoundFunction *TruthTable
	Phase         *TruthTable
}

// DuaDSDRangeZero recognizes F(bound,outer)=D(bound) XOR P(outer), the
// coherent behavior of Dua_DsdRangeVars0 and Dua_DsdRangeVars0Derive.
func DuaDSDRangeZero(t *TruthTable, boundVariables int, options DuaRangeOptions) (DuaRangeZeroDecomposition, error) {
	var result DuaRangeZeroDecomposition
	o, err := normalizeDuaRangeOptions(options)
	if err != nil {
		return result, err
	}
	if t == nil || boundVariables <= 2 || boundVariables > 6 || boundVariables >= t.Variables() {
		return result, ErrDimension
	}
	if err = t.Check(); err != nil {
		return result, err
	}
	bound, err := NewConstantTruthTable(boundVariables, false, o.MaxWords)
	if err != nil {
		return result, err
	}
	phase, err := NewConstantTruthTable(t.Variables()-boundVariables, false, o.MaxWords)
	if err != nil {
		return result, err
	}
	boundCount := uint64(1) << uint(boundVariables)
	highCount := uint64(1) << uint(t.Variables()-boundVariables)
	for assignment := uint64(0); assignment < boundCount; assignment++ {
		value, err := t.Get(assignment)
		if err != nil {
			return result, err
		}
		if value {
			if err = bound.Set(assignment, true); err != nil {
				return result, err
			}
		}
	}
	for high := uint64(0); high < highCount; high++ {
		if o.Stop != nil && o.Stop() {
			return DuaRangeZeroDecomposition{}, ErrCancelled
		}
		equal, opposite := true, true
		for assignment := uint64(0); assignment < boundCount; assignment++ {
			want, _ := bound.Get(assignment)
			got, _ := t.Get(assignment | high<<uint(boundVariables))
			equal = equal && got == want
			opposite = opposite && got != want
		}
		if !equal && !opposite {
			return result, nil
		}
		if opposite {
			if err = phase.Set(high, true); err != nil {
				return result, err
			}
		}
	}
	result.Found, result.BoundFunction, result.Phase = true, bound, phase
	return result, nil
}

func DuaComposeRangeZero(bound, phase *TruthTable, maxWords int) (*TruthTable, error) {
	if bound == nil || phase == nil || bound.Variables() <= 2 || bound.Variables() > 6 || maxWords < 0 {
		return nil, ErrDimension
	}
	if err := bound.Check(); err != nil {
		return nil, err
	}
	if err := phase.Check(); err != nil {
		return nil, err
	}
	variables := bound.Variables() + phase.Variables()
	result, err := NewConstantTruthTable(variables, false, maxWords)
	if err != nil {
		return nil, err
	}
	boundMask := uint64(1)<<uint(bound.Variables()) - 1
	for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
		left, _ := bound.Get(assignment & boundMask)
		right, _ := phase.Get(assignment >> uint(bound.Variables()))
		if left != right {
			if err = result.Set(assignment, true); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}
