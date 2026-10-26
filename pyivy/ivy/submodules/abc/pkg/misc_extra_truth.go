package abc

// Source-specific truth utilities from src/misc/extra/extraUtilTruth.c.

import (
	"math/bits"
	"slices"
)

// ExtraElementaryTruthTables returns independently owned elementary tables.
func ExtraElementaryTruthTables(variables, maxWords int) ([]*TruthTable, error) {
	if variables < 0 {
		return nil, ErrDimension
	}
	result := make([]*TruthTable, variables)
	for variable := range result {
		table, err := NewVariableTruthTable(variables, variable, maxWords)
		if err != nil {
			return nil, err
		}
		result[variable] = table
	}
	return result, nil
}

// ExtraStretchTruth embeds each source input at the corresponding set bit of
// phase. Set bits are considered from low to high, as in Extra_TruthStretch.
func (t *TruthTable) ExtraStretchTruth(variables int, phase uint64) (*TruthTable, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	if variables < t.variables || variables > 64 || bits.OnesCount64(phase) != t.variables {
		return nil, ErrDimension
	}
	if variables < 64 && phase>>uint(variables) != 0 {
		return nil, ErrDimension
	}
	mapping := make([]int, 0, t.variables)
	for variable := 0; variable < variables; variable++ {
		if phase&(uint64(1)<<uint(variable)) != 0 {
			mapping = append(mapping, variable)
		}
	}
	return t.ExpandVariables(variables, mapping)
}

// ExtraShrinkTruth retains the phase-selected inputs in increasing source
// order. Inputs outside phase must be functionally unused; unchecked loss is
// rejected instead of silently truncating an active dimension.
func (t *TruthTable) ExtraShrinkTruth(phase uint64) (*TruthTable, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	if t.variables > 64 || (t.variables < 64 && phase>>uint(t.variables) != 0) {
		return nil, ErrDimension
	}
	keep := make([]int, 0, bits.OnesCount64(phase))
	for variable := 0; variable < t.variables; variable++ {
		if phase&(uint64(1)<<uint(variable)) != 0 {
			keep = append(keep, variable)
		}
	}
	return t.ProjectVariables(keep)
}

// ExtraRecombineTruthCofactors selects the zero and one tables using variable.
// Bits in the irrelevant half of either cofactor are ignored, matching the
// masks used by Extra_TruthMux.
func ExtraRecombineTruthCofactors(zero, one *TruthTable, variable int) (*TruthTable, error) {
	if e := zero.checkPair(one); e != nil {
		return nil, e
	}
	if variable < 0 || variable >= zero.variables {
		return nil, ErrDimension
	}
	control, e := NewVariableTruthTable(zero.variables, variable, zero.limit)
	if e != nil {
		return nil, e
	}
	return MuxTruthTables(control, one, zero)
}

// ExtraMinimumCofactorSupportOverlap returns the first input with minimum
// overlap between its cofactor supports, considering only cofactors with at
// most five support inputs. A variable of -1 and overlap 32 preserve the
// source sentinel when no input qualifies.
func (t *TruthTable) ExtraMinimumCofactorSupportOverlap() (overlap, variable int, err error) {
	if e := t.Check(); e != nil {
		return 0, -1, e
	}
	overlap, variable = 32, -1
	for candidate := 0; candidate < t.variables; candidate++ {
		zero, e := t.Cofactor(candidate, false)
		if e != nil {
			return 0, -1, e
		}
		one, e := t.Cofactor(candidate, true)
		if e != nil {
			return 0, -1, e
		}
		zeroSupport, e := zero.Support()
		if e != nil {
			return 0, -1, e
		}
		oneSupport, e := one.Support()
		if e != nil {
			return 0, -1, e
		}
		if len(zeroSupport) > 5 || len(oneSupport) > 5 {
			continue
		}
		common := 0
		for _, first := range zeroSupport {
			if slices.Contains(oneSupport, first) {
				common++
			}
		}
		if common < overlap {
			overlap, variable = common, candidate
		}
		if overlap == 0 {
			break
		}
	}
	return overlap, variable, nil
}

type ExtraTruthCofactorCounts struct {
	Zero, One []int
}

// ExtraCofactorCounts uses the source's 32-bit packed counting space. Below
// six inputs this intentionally counts replicated minterms rather than only
// the logical assignments.
func (t *TruthTable) ExtraCofactorCounts() (ExtraTruthCofactorCounts, error) {
	populations, e := t.CofactorPopulations()
	if e != nil {
		return ExtraTruthCofactorCounts{}, e
	}
	zero, one := slices.Clone(populations.Zero), slices.Clone(populations.One)
	if t.variables <= 5 {
		for variable := range zero {
			zero[variable] /= 2
			one[variable] /= 2
		}
	}
	return ExtraTruthCofactorCounts{Zero: zero, One: one}, nil
}

func (t *TruthTable) extraPacked32Words() ([]uint32, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	count := 1
	if t.variables > 5 {
		count = 1 << uint(t.variables-5)
	}
	if count > 1024 {
		return nil, ErrLimit
	}
	words := make([]uint32, count)
	for i := range words {
		word := t.words[i/2]
		words[i] = uint32(word >> uint(32*(i&1)))
	}
	return words, nil
}

func extraFirstPrimes(count int) []uint32 {
	primes := make([]uint32, 0, count)
	for candidate := uint32(2); len(primes) < count; candidate++ {
		prime := true
		for _, divisor := range primes {
			if divisor > candidate/divisor {
				break
			}
			if candidate%divisor == 0 {
				prime = false
				break
			}
		}
		if prime {
			primes = append(primes, candidate)
		}
	}
	return primes
}

// ExtraTruthHash is the exact wrapping 32-bit prime-weighted XOR hash.
func (t *TruthTable) ExtraTruthHash() (uint32, error) {
	words, e := t.extraPacked32Words()
	if e != nil {
		return 0, e
	}
	primes := extraFirstPrimes(len(words))
	var hash uint32
	for i, word := range words {
		hash ^= primes[i] * word
	}
	return hash, nil
}

type ExtraTruthSemiCanonicalization struct {
	Table              *TruthTable
	SourceToTarget     []int
	TargetToSource     []int
	InputPhases        []bool
	OutputPhase        bool
	PackedOnes         int
	ZeroCofactorCounts []int
}

// ExtraSemiCanonicalize preserves the older Extra implementation, including
// complementing a balanced function when assignment zero is one and the
// stable adjacent-swap ordering of equal cofactor signatures.
func (t *TruthTable) ExtraSemiCanonicalize() (ExtraTruthSemiCanonicalization, error) {
	fail := func(e error) (ExtraTruthSemiCanonicalization, error) {
		return ExtraTruthSemiCanonicalization{}, e
	}
	packed, e := t.extraPacked32Words()
	if e != nil {
		return fail(e)
	}
	result := t.Clone()
	ones := 0
	for _, word := range packed {
		ones += bits.OnesCount32(word)
	}
	outputPhase := ones > len(packed)*16 || ones == len(packed)*16 && packed[0]&1 != 0
	if outputPhase {
		result = result.Not()
		ones = len(packed)*32 - ones
	}
	counts, e := result.ExtraCofactorCounts()
	if e != nil {
		return fail(e)
	}
	phases := make([]bool, t.variables)
	for variable := 0; variable < t.variables; variable++ {
		if counts.Zero[variable] <= counts.One[variable] {
			continue
		}
		phases[variable] = true
		counts.Zero[variable], counts.One[variable] = counts.One[variable], counts.Zero[variable]
		result.flipVariable(variable)
	}
	targetToSource := make([]int, t.variables)
	for i := range targetToSource {
		targetToSource[i] = i
	}
	for {
		changed := false
		for variable := 0; variable+1 < t.variables; variable++ {
			if counts.Zero[variable] <= counts.Zero[variable+1] {
				continue
			}
			changed = true
			counts.Zero[variable], counts.Zero[variable+1] = counts.Zero[variable+1], counts.Zero[variable]
			counts.One[variable], counts.One[variable+1] = counts.One[variable+1], counts.One[variable]
			targetToSource[variable], targetToSource[variable+1] = targetToSource[variable+1], targetToSource[variable]
			result.swapVariables(variable, variable+1)
		}
		if !changed {
			break
		}
	}
	sourceToTarget := make([]int, t.variables)
	for target, source := range targetToSource {
		sourceToTarget[source] = target
	}
	return ExtraTruthSemiCanonicalization{
		Table:              result,
		SourceToTarget:     sourceToTarget,
		TargetToSource:     targetToSource,
		InputPhases:        phases,
		OutputPhase:        outputPhase,
		PackedOnes:         ones,
		ZeroCofactorCounts: slices.Clone(counts.Zero),
	}, nil
}
