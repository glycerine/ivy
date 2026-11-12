package lucky

import (
	"fmt"
	"slices"
	"sort"
)

func memCompare(a, b []uint64, vars int32) int32 {
	for i := Kit_TruthWordNum_64bit(vars) - 1; i >= 0; i-- {
		if a[i] > b[i] {
			return 1
		}
		if a[i] < b[i] {
			return -1
		}
	}
	return 0
}
func compareWords1(a, b uint64) int32   { return CompareWords(a, b) }
func compareWords2(a, b []uint64) int32 { return CompareWords(a[0], b[0]) }
func compareWords3(a, b []uint64) int32 { return memCompare(a, b, 16) }
func compareWords(a, b []uint64) int32 {
	for i := uint(0); i < 64; i += 8 {
		aa, bb := byte(a[0]>>i), byte(b[0]>>i)
		if aa < bb {
			return -1
		}
		if aa > bb {
			return 1
		}
	}
	return 0
}
func sortAndUnique1(data []uint64, store *Abc_TtStore_t) {
	data = data[:store.NFuncs]
	slices.Sort(data)
	store.NFuncs = int32(len(slices.Compact(data)))
}
func sortAndUnique(data [][]uint64, store *Abc_TtStore_t) {
	data = data[:store.NFuncs]
	sort.Slice(data, func(i, j int) bool { return memCompare(data[i], data[j], store.NVars) < 0 })
	count := 0
	for _, table := range data {
		if count == 0 || !slices.Equal(data[count-1][:store.NWords], table[:store.NWords]) {
			data[count] = table
			count++
		}
	}
	store.NFuncs = int32(count)
}

type cycleCtr struct{ totalCycles, maxNCycles, minNCycles int32 }

func setCycleCtrPtr() *cycleCtr { return &cycleCtr{minNCycles: 111111111} }
func freeCycleCtr(state *cycleCtr) {
	if state != nil {
		*state = cycleCtr{}
	}
}
func makeArray(store *Abc_TtStore_t) [][]uint64 {
	result := make([][]uint64, store.NFuncs)
	for i := range result {
		result[i] = slices.Clone(store.Funcs[i][:store.NWords])
	}
	return result
}
func freeArray(data [][]uint64, store *Abc_TtStore_t) { clear(data[:store.NFuncs]) }
func makeArrayB(data [][]uint64, n int32) []uint64 {
	result := make([]uint64, n)
	for i := range result {
		result[i] = data[i][0]
	}
	return result
}
func freeArrayB(data []uint64) { clear(data) }
func printCCtrInfo(state *cycleCtr, n int32) {
	fmt.Printf("maxNCycles = %d\nminNCycles = %d\naverage NCycles = %.3f\n", state.maxNCycles, state.minNCycles, float64(state.totalCycles)/float64(n))
}
func minimalInitialFlip1(table []uint64, vars int32) int32 {
	if table[Kit_TruthWordNum_64bit(vars)-1]>>63 != 0 {
		Kit_TruthNot_64bit(table, vars)
		return 1
	}
	return 0
}
func minimalInitialFlip(table []uint64, vars int32, phase *uint32) int32 {
	if minimalInitialFlip1(table, vars) != 0 {
		*phase ^= 1 << uint(vars)
		return 1
	}
	return 0
}
func swapInfoAdjacentVars(variable int32, permutation []byte, phase *uint32) {
	permutation[variable], permutation[variable+1] = permutation[variable+1], permutation[variable]
	*phase = swapPhaseBits(*phase, variable)
}
func minimizeSingle(table, minimal, duplicate []uint64, vars int32, permutation, temp []byte, phase *uint32, swapping bool) int32 {
	limit := vars
	if swapping {
		limit--
	}
	if limit <= 0 {
		return 0
	}
	Kit_TruthCopy_64bit(minimal, table, vars)
	Kit_TruthCopy_64bit(duplicate, table, vars)
	bestPhase := *phase
	copy(temp, permutation)
	apply := func(index int32) {
		if swapping {
			Kit_TruthSwapAdjacentVars_64bit(table, vars, index)
			if permutation != nil {
				swapInfoAdjacentVars(index, permutation, phase)
			}
		} else {
			Kit_TruthChangePhase_64bit(table, vars, index)
			*phase ^= 1 << uint(index)
		}
	}
	apply(0)
	for i := int32(1); i < limit; i++ {
		if memCompare(minimal, table, vars) > 0 {
			Kit_TruthCopy_64bit(minimal, table, vars)
			bestPhase = *phase
			copy(temp, permutation)
		} else {
			Kit_TruthCopy_64bit(table, minimal, vars)
			*phase = bestPhase
			copy(permutation, temp)
		}
		apply(i)
	}
	if memCompare(minimal, table, vars) < 0 {
		Kit_TruthCopy_64bit(table, minimal, vars)
		*phase = bestPhase
		copy(permutation, temp)
	}
	if memCompare(table, duplicate, vars) == 0 {
		return 0
	}
	return 1
}
func minimalFlip(table, minimal, duplicate []uint64, vars int32, phase *uint32) int32 {
	return minimizeSingle(table, minimal, duplicate, vars, nil, nil, phase, false)
}
func minimalSwap(table, minimal, duplicate []uint64, vars int32, permutation, temp []byte, phase *uint32) int32 {
	return minimizeSingle(table, minimal, duplicate, vars, permutation, temp, phase, true)
}
func minimalFlip1(table, minimal, duplicate []uint64, vars int32) int32 {
	var phase uint32
	return minimalFlip(table, minimal, duplicate, vars, &phase)
}
func minimalSwap1(table, minimal, duplicate []uint64, vars int32) int32 {
	var phase uint32
	return minimizeSingle(table, minimal, duplicate, vars, nil, nil, &phase, true)
}
func luckyCanonicizer(table, aux, duplicate []uint64, vars int32, permutation, temp []byte, phase *uint32) {
	for {
		changes := minimalInitialFlip(table, vars, phase)
		changes += minimalFlip(table, aux, duplicate, vars, phase)
		changes += minimalSwap(table, aux, duplicate, vars, permutation, temp, phase)
		if changes == 0 {
			return
		}
	}
}
func luckyCanonicizer1_simple(table, aux, duplicate []uint64, vars int32, permutation []byte, phase uint32) uint32 {
	for {
		changes := minimalInitialFlip1(table, vars)
		changes += minimalFlip1(table, aux, duplicate, vars)
		changes += minimalSwap1(table, aux, duplicate, vars)
		if changes == 0 {
			return phase
		}
	}
}
func luckyCanonicizer_final(table, aux, duplicate []uint64, vars int32) {
	Kit_TruthSemiCanonicize_Yasha_simple(table, vars, make([]int32, vars))
	luckyCanonicizer1_simple(table, aux, duplicate, vars, nil, 0)
}
func Kit_TruthSemiCanonicize_new_internal(table []uint64, vars int32, permutation []byte) uint32 {
	words := Kit_TruthWordNum_64bit(vars)
	aux, duplicate := make([]uint64, words), make([]uint64, words)
	phase := Kit_TruthSemiCanonicize_Yasha(table, vars, permutation)
	luckyCanonicizer(table, aux, duplicate, vars, permutation, make([]byte, vars), &phase)
	return phase
}
func Kit_TruthSemiCanonicize_new(table, aux []uint32, vars int32, permutation []byte) uint32 {
	words := Kit_TruthWordNum_64bit(vars)
	packed := make([]uint64, words)
	if vars < 6 {
		packed[0] = uint64(table[0])<<32 | uint64(table[0])
	} else {
		for i := range packed {
			packed[i] = uint64(table[2*i]) | uint64(table[2*i+1])<<32
		}
	}
	phase := Kit_TruthSemiCanonicize_new_internal(packed, vars, permutation)
	if vars < 6 {
		table[0] = uint32(packed[0])
	} else {
		for i, x := range packed {
			table[2*i], table[2*i+1] = uint32(x), uint32(x>>32)
		}
	}
	return phase
}
