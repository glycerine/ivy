package abc

// Cofactor populations and semicanonicalization from opt/dau/dauCanon.c:
// Abc_TtCountOnesInTruth/CountOnesInCofs/Simple and Abc_TtSemiCanonicize.

import "math/bits"

// TruthCofactorPopulations counts packed minterms, including small-table
// repeats. Zero/One count only each selected half (not a replicated cofactor).
// In particular these counts differ from TruthTable.CountOnes below six vars.
type TruthCofactorPopulations struct {
	Ones      int
	Zero, One []int
}

// NormalizeSmallTruthWord masks a zero-through-five-input logical truth word
// and repeats its active pattern across 64 bits. Six or more inputs leave the
// supplied word unchanged, matching Abc_TtNormalizeSmallTruth.
func NormalizeSmallTruthWord(word uint64, variables int) (uint64, error) {
	if _, e := TruthTableWordCount(variables); e != nil {
		return 0, e
	}
	if variables >= 6 {
		return word, nil
	}
	period := 1 << uint(variables)
	base := word & (uint64(1)<<uint(period) - 1)
	word = 0
	for shift := 0; shift < 64; shift += period {
		word |= base << uint(shift)
	}
	return word, nil
}

func (t *TruthTable) CofactorPopulations() (TruthCofactorPopulations, error) {
	if e := t.Check(); e != nil {
		return TruthCofactorPopulations{}, e
	}
	zero := make([]int, t.variables)
	ones := 0
	for index, word := range t.words {
		count := bits.OnesCount64(word)
		ones += count
		for variable := 0; variable < min(6, t.variables); variable++ {
			zero[variable] += bits.OnesCount64(word &^ truthVariablePattern(variable))
		}
		for variable := 6; variable < t.variables; variable++ {
			if index&(1<<uint(variable-6)) == 0 {
				zero[variable] += count
			}
		}
	}
	one := make([]int, t.variables)
	for variable, count := range zero {
		one[variable] = ones - count
	}
	return TruthCofactorPopulations{ones, zero, one}, nil
}

// CofactorPopulationsSlow follows Abc_TtCountOnesInCofsSlow's direct byte
// scan. Counts use the packed convention, including repeated small tables.
// The implementation is intentionally independent of CofactorPopulations so
// it can serve as a simple cross-check for optimized population algorithms.
func (t *TruthTable) CofactorPopulationsSlow() (TruthCofactorPopulations, error) {
	if e := t.Check(); e != nil {
		return TruthCofactorPopulations{}, e
	}
	zero := make([]int, t.variables)
	ones := 0
	for wordIndex, word := range t.words {
		for byteInWord := 0; byteInWord < 8; byteInWord++ {
			value := uint8(word >> uint(8*byteInWord))
			ones += bits.OnesCount8(value)
			if t.variables > 0 {
				zero[0] += bits.OnesCount8(value & 0x55)
			}
			if t.variables > 1 {
				zero[1] += bits.OnesCount8(value & 0x33)
			}
			if t.variables > 2 {
				zero[2] += bits.OnesCount8(value & 0x0f)
			}
			byteIndex := 8*wordIndex + byteInWord
			for variable := 3; variable < t.variables; variable++ {
				if byteIndex&(1<<uint(variable-3)) == 0 {
					zero[variable] += bits.OnesCount8(value)
				}
			}
		}
	}
	one := make([]int, t.variables)
	for variable, count := range zero {
		one[variable] = ones - count
	}
	return TruthCofactorPopulations{Ones: ones, Zero: zero, One: one}, nil
}

// TruthSemiCanonicalization supplies a FORWARD configuration from source to
// Table: TransformNPN(Permutation,InputPhases,OutputPhase). Permutation maps
// source variable to target position; phases are indexed by source variable.
// C instead returns target-to-source permutation and target-indexed phases.
// ZeroCofactorCounts are indexed by target position and include packed repeats.
type TruthSemiCanonicalization struct {
	Table              *TruthTable
	Permutation        []int
	InputPhases        []bool
	OutputPhase        bool
	PackedOnes         int
	ZeroCofactorCounts []int
}

// SemiCanonicalize matches C's population signature pass, including its
// nonstable selection-sort ties. Unless onlyPermutation, output is oriented to
// at most half packed minterms, and each input's zero cofactor to at least as
// many minterms as its one cofactor. Variables sort by ascending zero-cofactor
// count. This signature normalization is not full NPN canonicalization.
// All results own storage; source stays unchanged. Explicit vectors extend C's
// fixed 16-variable permutation/count buffers and unsigned phase mask.
func (t *TruthTable) SemiCanonicalize(onlyPermutation bool) (TruthSemiCanonicalization, error) {
	fail := func(e error) (TruthSemiCanonicalization, error) { return TruthSemiCanonicalization{}, e }
	populations, e := t.CofactorPopulations()
	if e != nil {
		return fail(e)
	}
	result := t.Clone()
	ones := populations.Ones
	zero := populations.Zero
	output := !onlyPermutation && ones > len(t.words)*32
	if output {
		for i, word := range result.words {
			result.words[i] = ^word
		}
		ones = len(t.words)*64 - ones
		for i, count := range zero {
			zero[i] = len(t.words)*32 - count
		}
	}
	phases := make([]bool, t.variables)
	if !onlyPermutation {
		for variable, count := range zero {
			if count < ones-count {
				result.flipVariable(variable)
				phases[variable] = true
				zero[variable] = ones - count
			}
		}
	}
	labels := make([]int, t.variables)
	for i := range labels {
		labels[i] = i
	}
	for i := 0; i < t.variables-1; i++ {
		best := i + 1
		for k := i + 2; k < t.variables; k++ {
			if zero[best] > zero[k] {
				best = k
			}
		}
		if zero[i] <= zero[best] {
			continue
		}
		labels[i], labels[best] = labels[best], labels[i]
		zero[i], zero[best] = zero[best], zero[i]
		result.swapVariables(i, best)
	}
	permutation := make([]int, t.variables)
	for target, source := range labels {
		permutation[source] = target
	}
	return TruthSemiCanonicalization{result, permutation, phases, output, ones, zero}, nil
}
