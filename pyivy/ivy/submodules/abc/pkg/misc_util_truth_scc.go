package abc

// Shifted cofactor coefficients from opt/dau/dauCanon.c:
// shiftFunc, Abc_TtScc6/Scc and Abc_TtSccInCofs6/InCofs.

import "math/bits"

// TruthShiftedCoefficients records weighted packed minterm populations.
// Each set minterm at packed index x contributes 2^popcount(x) to Total,
// and to Zero[v] when input v is zero. Small tables include the six-input
// packed repetitions; these are not logical assignment counts. uint64 avoids
// C's signed-int overflow and covers every supported TruthTable width.
type TruthShiftedCoefficients struct {
	Total uint64
	Zero  []uint64
}

// ShiftedCofactorCoefficients computes C's SCC signature without mutating truth
// or allocating cofactor tables. Removing a zero input preserves the minterm's
// set-bit count, so one packed traversal computes all zero-cofactor weights.
func (t *TruthTable) ShiftedCofactorCoefficients() (TruthShiftedCoefficients, error) {
	if e := t.Check(); e != nil {
		return TruthShiftedCoefficients{}, e
	}
	r := TruthShiftedCoefficients{Zero: make([]uint64, t.variables)}
	for wordIndex, word := range t.words {
		for word != 0 {
			index := uint64(wordIndex)*64 + uint64(bits.TrailingZeros64(word))
			weight := uint64(1) << uint(bits.OnesCount64(index))
			r.Total += weight
			for v := range r.Zero {
				if index&(uint64(1)<<uint(v)) == 0 {
					r.Zero[v] += weight
				}
			}
			word &= word - 1
		}
	}
	return r, nil
}

// shiftedGroupCoefficients follows Abc_TgCalcScc on applied group order.
// Each linked symmetry chain occupies consecutive truth inputs with equal
// cofactor coefficients. sorted sorts values within each signature group,
// leaving representatives and truth unchanged.
func (s *truthAdaptiveState) shiftedGroupCoefficients(sorted bool) ([]uint64, error) {
	coefficients, e := s.table.ShiftedCofactorCoefficients()
	if e != nil {
		return nil, e
	}
	result := make([]uint64, len(s.representatives))
	position := 0
	for i, head := range s.representatives {
		result[i] = coefficients.Zero[position]
		for v := head; v >= 0; v = s.links[v] {
			if coefficients.Zero[position] != result[i] {
				return nil, ErrDimension
			}
			position++
		}
	}
	if sorted {
		for _, group := range s.groups {
			for i := group.Start + 1; i < group.Start+group.Size; i++ {
				value := result[i]
				j := i
				for j > group.Start && result[j-1] > value {
					result[j] = result[j-1]
					j--
				}
				result[j] = value
			}
		}
	}
	return result, nil
}
