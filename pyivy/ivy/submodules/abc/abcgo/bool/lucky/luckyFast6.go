package lucky

import "slices"

func resetPCanonPermArray_6Vars(permutation []byte) { resetPCanonPermArray(permutation, 6) }
func resetPCanonPermArray(permutation []byte, vars int32) {
	for i := int32(0); i < vars; i++ {
		permutation[i] = byte('a' + i)
	}
}
func Abc_allFlip(x uint64, phase *uint32) uint64 {
	if x>>63 != 0 {
		*phase ^= 1 << 6
		return ^x
	}
	return x
}
func adjustInfoAfterSwap(permutation []byte, phase uint32, variable int32, info uint32) uint32 {
	if info < 4 {
		return phase ^ (info << uint(variable))
	}
	phase ^= (info - 4) << uint(variable)
	permutation[variable], permutation[variable+1] = permutation[variable+1], permutation[variable]
	return swapPhaseBits(phase, variable)
}
func Extra_Truth6SwapAdjacent(t uint64, variable int32) uint64 {
	if variable < 0 || variable >= 5 {
		panic("single-word adjacent variable outside [0,4]")
	}
	m := adjacentMasks[variable]
	shift := uint(1) << uint(variable)
	return t&m[0] | (t&m[1])<<shift | (t&m[2])>>shift
}
func Extra_Truth6ChangePhase(t uint64, variable int32) uint64 {
	if variable < 0 || variable >= 6 {
		panic("single-word phase variable outside [0,5]")
	}
	shift := uint(1) << uint(variable)
	return (t&mask0[variable])<<shift | (t&^mask0[variable])>>shift
}
func minimumPair(table []uint64, vars, variable int32, permutation []byte, phase *uint32, allowFlip bool) {
	words := int(Kit_TruthWordNum_64bit(vars))
	original := slices.Clone(table[:words])
	best := slices.Clone(original)
	candidate := make([]uint64, words)
	bestInfo := uint32(0)
	infos := []uint32{1, 2, 3, 4, 6, 5, 7}
	if !allowFlip {
		infos = []uint32{4}
	}
	for _, info := range infos {
		copy(candidate, original)
		flips := info
		if info >= 4 {
			Kit_TruthSwapAdjacentVars_64bit(candidate, vars, variable)
			flips = info - 4
			flips = (flips&1)<<1 | (flips&2)>>1
		}
		if flips&1 != 0 {
			Kit_TruthChangePhase_64bit(candidate, vars, variable)
		}
		if flips&2 != 0 {
			Kit_TruthChangePhase_64bit(candidate, vars, variable+1)
		}
		comparison := memCompare(candidate, best, vars)
		if comparison < 0 || !allowFlip && vars == 6 && comparison == 0 {
			copy(best, candidate)
			bestInfo = info
		}
	}
	if bestInfo != 0 {
		copy(table, best)
		*phase = adjustInfoAfterSwap(permutation, *phase, variable, bestInfo)
	}
}
func Extra_Truth6MinimumRoundOne(t uint64, variable int32, permutation []byte, phase *uint32) uint64 {
	table := []uint64{t}
	minimumPair(table, 6, variable, permutation, phase, true)
	return table[0]
}
func Extra_Truth6MinimumRoundOne_noEBFC(t uint64, variable int32, permutation []byte, phase *uint32) uint64 {
	table := []uint64{t}
	minimumPair(table, 6, variable, permutation, phase, false)
	return table[0]
}
func minimumRounds(table []uint64, vars int32, store []int32, permutation []byte, phase *uint32, allowFlip bool) {
	for {
		before := slices.Clone(table)
		for i := int32(0); i < vars-1; i++ {
			if store[i] == store[i+1] {
				minimumPair(table, vars, i, permutation, phase, allowFlip)
			}
		}
		if memCompare(before, table, vars) == 0 {
			return
		}
	}
}
func Extra_Truth6MinimumRoundMany(t uint64, store []int32, permutation []byte, phase *uint32) uint64 {
	table := []uint64{t}
	minimumRounds(table, 6, store, permutation, phase, true)
	return table[0]
}
func Extra_Truth6MinimumRoundMany_noEBFC(t uint64, store []int32, permutation []byte, phase *uint32) uint64 {
	table := []uint64{t}
	minimumRounds(table, 6, store, permutation, phase, false)
	return table[0]
}
func Extra_Truth6MinimumRoundMany1(t uint64, store []int32, permutation []byte, phase *uint32) uint64 {
	flags := *phase >> 7
	allowFlip := flags&1 != 0
	run := Extra_Truth6MinimumRoundMany_noEBFC
	if allowFlip {
		run = Extra_Truth6MinimumRoundMany
	}
	if flags&2 == 0 {
		return run(t, store, permutation, phase)
	}
	alternativePerm := slices.Clone(permutation)
	alternativePhase := *phase ^ (1 << 6)
	a := run(t, store, permutation, phase)
	b := run(^t, store, alternativePerm, &alternativePhase)
	if b < a {
		copy(permutation, alternativePerm)
		*phase = alternativePhase
		return b
	}
	return a
}
func luckyCanonicizer_final_fast_6Vars(t uint64, store []int32, permutation []byte, phase *uint32) uint64 {
	table := []uint64{t}
	*phase = Kit_TruthSemiCanonicize_Yasha1(table, 6, permutation, store)
	return Extra_Truth6MinimumRoundMany1(table[0], store, permutation, phase)
}
func luckyCanonicizer_final_fast_6Vars1(t uint64, store []int32, permutation []byte, phase *uint32) uint64 {
	t = luckyCanonicizer_final_fast_6Vars(t, store, permutation, phase)
	for i := int32(5); i >= 0; i-- {
		t = Extra_Truth6ChangePhase(t, i)
	}
	*phase ^= 0x3f
	return Extra_Truth6MinimumRoundMany1(t, store, permutation, phase)
}
