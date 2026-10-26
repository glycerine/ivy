package abc

// Recursive constant-block populations from opt/dau/dauCanon.c:
// Abc_TtCountOnesInCofsFast6_rec, Abc_TtCountOnesInCofsFast_rec and
// Abc_TtCountOnesInCofsFast. Counts below six inputs are logical, not packed.

import "math/bits"

// TruthLogicalCofactorPopulations counts logical satisfying assignments of the
// function and its zero/one-input cofactors. Small-table packed repetitions are
// excluded. uint64 covers every supported input width without signed overflow.
type TruthLogicalCofactorPopulations struct {
	Ones      uint64
	Zero, One []uint64
}

func truthCountCofactorWord(word uint64, variable, count int, zero []uint64) uint64 {
	if word == 0 {
		return 0
	}
	if word == ^uint64(0) {
		for i := 0; i <= variable; i++ {
			zero[i] += uint64(count / 2)
		}
		return uint64(count)
	}
	if count <= 8 {
		word &= (uint64(1) << uint(count)) - 1
		for i := 0; i <= variable; i++ {
			zero[i] += uint64(bits.OnesCount64(word &^ truthVariablePattern(i)))
		}
		return uint64(bits.OnesCount64(word))
	}
	mask := truthVariablePattern(variable)
	shift := uint(1) << uint(variable)
	low := word &^ mask
	high := word & mask
	a := truthCountCofactorWord(low|(low<<shift), variable-1, count/2, zero)
	b := truthCountCofactorWord(high|(high>>shift), variable-1, count/2, zero)
	zero[variable] += a
	return a + b
}
func truthCountCofactorWords(words []uint64, variable int, zero []uint64) uint64 {
	if len(words) == 1 {
		return truthCountCofactorWord(words[0], 5, 64, zero)
	}
	value := uint64(0)
	if words[0]&1 != 0 {
		value = ^uint64(0)
	}
	constant := true
	for _, word := range words {
		if word != value {
			constant = false
			break
		}
	}
	if constant {
		if value == 0 {
			return 0
		}
		half := uint64(len(words)) * 32
		for i := 0; i <= variable; i++ {
			zero[i] += half
		}
		return half * 2
	}
	half := len(words) / 2
	a := truthCountCofactorWords(words[:half], variable-1, zero)
	b := truthCountCofactorWords(words[half:], variable-1, zero)
	zero[variable] += a
	return a + b
}

// CofactorPopulationsFast uses C's recursive constant-block shortcuts and byte
// leaves. Unlike CofactorPopulations, counts exclude packed repetitions below
// six inputs, matching C's Fast entry point. Zero through two inputs are safe
// extensions of C's nVars>=3 assertion. No cofactor tables or borrowed byte
// views are allocated; results own their count vectors and source is unchanged.
func (t *TruthTable) CofactorPopulationsFast() (TruthLogicalCofactorPopulations, error) {
	if e := t.Check(); e != nil {
		return TruthLogicalCofactorPopulations{}, e
	}
	r := TruthLogicalCofactorPopulations{Zero: make([]uint64, t.variables), One: make([]uint64, t.variables)}
	if t.variables <= 6 {
		r.Ones = truthCountCofactorWord(t.words[0], t.variables-1, 1<<uint(t.variables), r.Zero)
	} else {
		r.Ones = truthCountCofactorWords(t.words, t.variables-1, r.Zero)
	}
	for i, count := range r.Zero {
		r.One[i] = r.Ones - count
	}
	return r, nil
}
