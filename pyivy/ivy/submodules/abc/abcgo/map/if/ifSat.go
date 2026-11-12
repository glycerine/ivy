package if_

import (
	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// ifSatPlan retains the public build/unbuild lifecycle. The source SAT clauses
// express a two-level functional decomposition; the Go implementation solves
// that exact relation from truth-table cofactors instead of constructing CNF.
type ifSatPlan struct{ lutSize int32 }

func If_ManSatBuildXY(nLutSize int32) any {
	if nLutSize < 2 || nLutSize > 6 {
		panic("if: two-LUT decomposition supports LUT sizes 2..6")
	}
	return &ifSatPlan{nLutSize}
}
func If_ManSatBuildXYZ(nLutSize int32) any {
	if nLutSize < 2 || nLutSize > 6 {
		panic("if: three-LUT decomposition supports LUT sizes 2..6")
	}
	return &ifSatPlan{nLutSize}
}
func If_ManSatUnbuild(p any) {}

func If_ManSat6ComposeLut4(table uint32, functions []word, nVars int32) word {
	if nVars < 0 || nVars > 4 || len(functions) < int(nVars) {
		panic("if: invalid 4-LUT composition inputs")
	}
	var result word
	for assignment := 0; assignment < 64; assignment++ {
		index := 0
		for v := int32(0); v < nVars; v++ {
			if functions[v]&(word(1)<<uint(assignment)) != 0 {
				index |= 1 << uint(v)
			}
		}
		if table&(1<<uint(index)) != 0 {
			result |= word(1) << uint(assignment)
		}
	}
	return result
}

func If_ManSat6Truth(bound, free word, boundSet, sharedSet, freeSet []int32) word {
	boundVars := make([]word, 0, len(boundSet)+len(sharedSet))
	for _, v := range sharedSet {
		boundVars = append(boundVars, truth6Variable(v))
	}
	for _, v := range boundSet {
		boundVars = append(boundVars, truth6Variable(v))
	}
	q := If_ManSat6ComposeLut4(uint32(bound), boundVars, int32(len(boundVars)))
	freeVars := []word{q}
	for _, v := range sharedSet {
		freeVars = append(freeVars, truth6Variable(v))
	}
	for _, v := range freeSet {
		freeVars = append(freeVars, truth6Variable(v))
	}
	return If_ManSat6ComposeLut4(uint32(free), freeVars, int32(len(freeVars)))
}

func truth6Variable(v int32) word {
	if v < 0 || v >= 6 {
		panic("if: truth-table variable outside 0..5")
	}
	var truth word
	for assignment := 0; assignment < 64; assignment++ {
		if assignment&(1<<uint(v)) != 0 {
			truth |= word(1) << uint(assignment)
		}
	}
	return truth
}

func truthBit(truth []word, assignment int) bool {
	return truth[assignment>>6]&(word(1)<<uint(assignment&63)) != 0
}
func setTruthBit(truth *word, assignment int, value bool) {
	if value {
		*truth |= word(1) << uint(assignment)
	}
}

// If_ManSatCheckXY derives bound/free LUT truth tables for a proposed variable
// partition. For each shared assignment, bound assignments may select at most
// two distinct free-variable cofactors; their selector is the bound LUT output.
func If_ManSatCheckXY(sat any, lutSize int32, truth []word, nVars int32, set uint32, boundOut, freeOut *word, scratch *Vec_Int_t) int32 {
	return ifManSatCheckXY(sat, lutSize, truth, nVars, set, boundOut, freeOut, scratch)
}

// ifManSatCheckXY accepts all truth words, including cuts wider than six inputs.
func ifManSatCheckXY(_ any, lutSize int32, truth []word, nVars int32, set uint32, boundOut, freeOut *word, scratch *Vec_Int_t) int32 {
	if nVars < 1 || nVars > 11 || lutSize < 2 || lutSize > 6 || nVars > 2*lutSize-1 || len(truth) < (1<<uint(nVars)+63)/64 {
		return 0
	}
	var boundSet, sharedSet, freeSet []int32
	for v := int32(0); v < nVars; v++ {
		switch set >> (2 * v) & 3 {
		case 0:
			freeSet = append(freeSet, v)
		case 1:
			boundSet = append(boundSet, v)
		case 3:
			sharedSet = append(sharedSet, v)
		default:
			return 0
		}
	}
	if len(boundSet)+len(sharedSet) > int(lutSize) || len(sharedSet)+len(freeSet)+1 > int(lutSize) {
		return 0
	}
	if len(boundSet) == 0 || len(freeSet) == 0 {
		return 0
	}
	sharedCount, boundCount, freeCount := 1<<uint(len(sharedSet)), 1<<uint(len(boundSet)), 1<<uint(len(freeSet))
	var boundTable, freeTable word
	if boundOut != nil {
		*boundOut = 0
	}
	if freeOut != nil {
		*freeOut = 0
	}
	for shared := 0; shared < sharedCount; shared++ {
		cofactorID := make(map[uint64]int)
		patterns := make([]uint64, boundCount)
		for b := 0; b < boundCount; b++ {
			var pattern uint64
			for f := 0; f < freeCount; f++ {
				assignment := 0
				for i, v := range sharedSet {
					if shared&(1<<uint(i)) != 0 {
						assignment |= 1 << uint(v)
					}
				}
				for i, v := range boundSet {
					if b&(1<<uint(i)) != 0 {
						assignment |= 1 << uint(v)
					}
				}
				for i, v := range freeSet {
					if f&(1<<uint(i)) != 0 {
						assignment |= 1 << uint(v)
					}
				}
				if truthBit(truth, assignment) {
					pattern |= uint64(1) << uint(f)
				}
			}
			patterns[b] = pattern
			if _, ok := cofactorID[pattern]; !ok {
				cofactorID[pattern] = len(cofactorID)
			}
		}
		if len(cofactorID) > 2 {
			return 0
		}
		for b, pattern := range patterns {
			selector := cofactorID[pattern]
			boundIndex := shared | (b << uint(len(sharedSet)))
			setTruthBit(&boundTable, boundIndex, selector != 0)
			for f := 0; f < freeCount; f++ {
				freeIndex := selector | (shared << 1) | (f << uint(1+len(sharedSet)))
				setTruthBit(&freeTable, freeIndex, pattern&(uint64(1)<<uint(f)) != 0)
			}
		}
	}
	if boundOut != nil {
		*boundOut = abcutil.Abc_Tt6Stretch(boundTable, int32(len(sharedSet)+len(boundSet)))
	}
	if freeOut != nil {
		*freeOut = abcutil.Abc_Tt6Stretch(freeTable, int32(1+len(sharedSet)+len(freeSet)))
	}
	if scratch != nil {
		vec.Vec_IntClear(scratch)
	}
	return 1
}

// If_ManSatCheckXYall searches bound sets and increasing numbers of shared
// inputs in the same order as the source implementation.
func If_ManSatCheckXYall(pSat any, lutSize int32, truth []word, nVars int32, scratch *Vec_Int_t) uint32 {
	if lutSize < 2 || lutSize > 6 || nVars <= lutSize || nVars > 2*lutSize-1 || len(truth) < (1<<uint(nVars)+63)/64 {
		return 0
	}
	maxShared := 2*lutSize - 1 - nVars
	var result uint32
	bound := make([]int32, lutSize)
	for sharedCount := int32(0); sharedCount <= maxShared; sharedCount++ {
		var chooseBound func(int32, int32) bool
		chooseBound = func(depth, start int32) bool {
			if depth == lutSize {
				var chooseShared func(int32, int32, uint32) bool
				chooseShared = func(remaining, start int32, set uint32) bool {
					if remaining == 0 {
						if ifManSatCheckXY(pSat, lutSize, truth, nVars, set, nil, nil, scratch) != 0 {
							result = set
							return true
						}
						return false
					}
					for i := start; i <= lutSize-remaining; i++ {
						if chooseShared(remaining-1, i+1, set|(3<<uint(2*bound[i]))) {
							return true
						}
					}
					return false
				}
				var set uint32
				for _, variable := range bound {
					set |= 1 << uint(2*variable)
				}
				return chooseShared(sharedCount, 0, set)
			}
			for i := start; i <= nVars-(lutSize-depth); i++ {
				bound[depth] = i
				if chooseBound(depth+1, i+1) {
					return true
				}
			}
			return false
		}
		if chooseBound(0, 0) {
			return result
		}
	}
	return 0
}

func If_ManSatCheckXYall_int(pSat any, lutSize int32, truth []word, nVars int32, scratch *Vec_Int_t) uint32 {
	return If_ManSatCheckXYall(pSat, lutSize, truth, nVars, scratch)
}

func If_ManSatTest2() {}
func If_ManSatTest3() {}
