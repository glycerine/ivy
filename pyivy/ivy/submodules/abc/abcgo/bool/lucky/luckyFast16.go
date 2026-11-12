package lucky

// The sixteen-variable minimizer uses the same eight adjacent signed
// permutations as the source's separate intra-word, half-word and block paths.
// One shared word-parallel evaluator avoids type-punning and fixed scratch arrays.
import "slices"

func minimalInitialFlip_fast_16Vars(table []uint64, vars int32, phase *uint32) {
	minimalInitialFlip(table, vars, phase)
}
func minimumAllOnce(table []uint64, vars int32, store []int32, permutation []byte, phase *uint32, allowFlip bool) int32 {
	before := slices.Clone(table[:Kit_TruthWordNum_64bit(vars)])
	for i := int32(0); i < vars-1; i++ {
		if store[i] == store[i+1] {
			minimumPair(table, vars, i, permutation, phase, allowFlip)
		}
	}
	if memCompare(table, before, vars) != 0 {
		return 1
	}
	return 0
}
func minimalSwapAndFlipIVar_superFast_all(table []uint64, vars, words int32, store []int32, permutation []byte, phase *uint32) int32 {
	return minimumAllOnce(table, vars, store, permutation, phase, true)
}
func minimalSwapAndFlipIVar_superFast_all_noEBFC(table []uint64, vars, words int32, store []int32, permutation []byte, phase *uint32) int32 {
	return minimumAllOnce(table, vars, store, permutation, phase, false)
}
func luckyCanonicizerS_F_first_16Vars1(table []uint64, vars, words int32, store []int32, permutation []byte, phase *uint32) {
	allowFlip := *phase>>uint(vars+1)&1 != 0
	minimumRounds(table, vars, store, permutation, phase, allowFlip)
}
func luckyCanonicizerS_F_first_16Vars11(table []uint64, vars, words int32, store []int32, permutation []byte, phase *uint32) {
	if *phase>>uint(vars+2) == 0 {
		luckyCanonicizerS_F_first_16Vars1(table, vars, words, store, permutation, phase)
		return
	}
	other := slices.Clone(table[:words])
	Kit_TruthNot_64bit(other, vars)
	otherPermutation := slices.Clone(permutation)
	otherPhase := *phase ^ (1 << uint(vars))
	luckyCanonicizerS_F_first_16Vars1(table, vars, words, store, permutation, phase)
	luckyCanonicizerS_F_first_16Vars1(other, vars, words, store, otherPermutation, &otherPhase)
	if memCompare(other, table, vars) < 0 {
		copy(table, other)
		copy(permutation, otherPermutation)
		*phase = otherPhase
	}
}
func luckyCanonicizer_final_fast_16Vars(table []uint64, vars, words int32, store []int32, permutation []byte, phase *uint32) {
	*phase = Kit_TruthSemiCanonicize_Yasha1(table, vars, permutation, store)
	luckyCanonicizerS_F_first_16Vars1(table, vars, words, store, permutation, phase)
}
func bitReverceOrder(table []uint64, vars int32) {
	for i := vars - 1; i >= 0; i-- {
		Kit_TruthChangePhase_64bit(table, vars, i)
	}
}
func luckyCanonicizer_final_fast_16Vars1(table []uint64, vars, words int32, store []int32, permutation []byte, phase *uint32) {
	*phase = Kit_TruthSemiCanonicize_Yasha1(table, vars, permutation, store)
	luckyCanonicizerS_F_first_16Vars11(table, vars, words, store, permutation, phase)
	bitReverceOrder(table, vars)
	*phase ^= (1 << uint(vars)) - 1
	luckyCanonicizerS_F_first_16Vars11(table, vars, words, store, permutation, phase)
}
func luckyFast(table []uint64, vars int32, permutation []byte, second bool) uint32 {
	words := Kit_TruthWordNum_64bit(vars)
	store := make([]int32, 16)
	var phase uint32
	if vars <= 6 {
		if len(permutation) < 6 {
			panic("single-word fast canonicalization needs six permutation entries")
		}
		if second {
			table[0] = luckyCanonicizer_final_fast_6Vars1(table[0], store, permutation, &phase)
		} else {
			table[0] = luckyCanonicizer_final_fast_6Vars(table[0], store, permutation, &phase)
		}
	} else if second {
		luckyCanonicizer_final_fast_16Vars1(table, vars, words, store, permutation, &phase)
	} else {
		luckyCanonicizer_final_fast_16Vars(table, vars, words, store, permutation, &phase)
	}
	return phase
}
func luckyCanonicizer_final_fast(table []uint64, vars int32, permutation []byte) uint32 {
	return luckyFast(table, vars, permutation, false)
}
func luckyCanonicizer_final_fast1(table []uint64, vars int32, permutation []byte) uint32 {
	return luckyFast(table, vars, permutation, true)
}
func LuckyCanonicizerFinalFast(table []uint64, vars int32, permutation []byte) uint32 {
	return luckyCanonicizer_final_fast(table, vars, permutation)
}
func LuckyCanonicizerFinalFast1(table []uint64, vars int32, permutation []byte) uint32 {
	return luckyCanonicizer_final_fast1(table, vars, permutation)
}
func ResetCanonPermutation(permutation []byte, vars int32) { resetPCanonPermArray(permutation, vars) }

// luckyCheck restores canonical variable positions/phases in place, as in the
// source verification helper. Zero indicates a successful reconstruction.
func luckyCheck(after, before []uint64, vars int32, permutation []byte, phase uint32) int32 {
	for position := int32(0); position < vars; position++ {
		wanted := byte('a' + position)
		found := position
		for found < vars && permutation[found] != wanted {
			found++
		}
		if found == vars {
			return 1
		}
		swapAndFlip(after, vars, position, found, permutation, &phase)
	}
	if phase>>uint(vars)&1 != 0 {
		Kit_TruthNot_64bit(after, vars)
	}
	if memCompare(after, before, vars) == 0 {
		return 0
	}
	return 1
}
func swapAndFlip(table []uint64, vars, position, other int32, permutation []byte, phase *uint32) {
	swap_ij(table, vars, position, other)
	permutation[position], permutation[other] = permutation[other], permutation[position]
	if ((*phase>>uint(position))^(*phase>>uint(other)))&1 != 0 {
		*phase ^= 1<<uint(position) | 1<<uint(other)
	}
	if *phase>>uint(position)&1 != 0 {
		Kit_TruthChangePhase_64bit(table, vars, position)
	}
}
