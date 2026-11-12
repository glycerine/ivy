package mio

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func mioFactorial(n int32) int {
	result := 1
	for i := int32(2); i <= n; i++ {
		result *= int(i)
	}
	return result
}
func mioFlip(truth uint64, variable int32) uint64 {
	var result uint64
	for m := 0; m < 64; m++ {
		if truth>>uint(m)&1 != 0 {
			result |= 1 << uint(m^(1<<uint(variable)))
		}
	}
	return result
}
func mioSwap(truth uint64, variable int32) uint64 {
	var result uint64
	for m := 0; m < 64; m++ {
		target := m
		a, b := m>>uint(variable)&1, m>>uint(variable+1)&1
		if a != b {
			target ^= 3 << uint(variable)
		}
		if truth>>uint(m)&1 != 0 {
			result |= 1 << uint(target)
		}
	}
	return result
}
func Nf_ManPrepareGate(n int32, truth uint64, complements, permutations []int32, result *Vec_Wrd_t) {
	if n < 2 || n > 6 || len(complements) < 1<<uint(n) || len(permutations) < mioFactorial(n) {
		panic("invalid NPN transformation schedule")
	}
	vec.Vec_WrdClear(result)
	for output := 0; output < 2; output++ {
		current := truth
		if output != 0 {
			current = ^current
		}
		initial := current
		for _, swap := range permutations[:mioFactorial(n)] {
			if swap < 0 || swap >= n-1 {
				panic("invalid permutation swap")
			}
			before := current
			for _, flip := range complements[:1<<uint(n)] {
				if flip < 0 || flip >= n {
					panic("invalid complement flip")
				}
				vec.Vec_WrdPush(result, current)
				current = mioFlip(current, flip)
			}
			if current != before {
				panic("complement schedule is not cyclic")
			}
			current = mioSwap(current, swap)
		}
		if current != initial {
			panic("permutation schedule is not cyclic")
		}
	}
}
func Nf_ManPreparePrint(n int32, complements, permutations []int32, lines [][8]byte) {
	count := 2 * mioFactorial(n) * (1 << uint(n))
	if n < 2 || n > 6 || len(lines) < count+1 || len(complements) < 1<<uint(n) || len(permutations) < mioFactorial(n) {
		panic("invalid transformation labels")
	}
	var label [8]byte
	for i := int32(0); i < n; i++ {
		label[i] = byte('A' + n - 1 - i)
	}
	index := 0
	for output := 0; output < 2; output++ {
		label[n] = '+'
		if output != 0 {
			label[n] = '-'
		}
		for _, swap := range permutations[:mioFactorial(n)] {
			for _, flip := range complements[:1<<uint(n)] {
				lines[index] = label
				fmt.Printf("%8d : %d : %s\n", index, output, string(label[:n+1]))
				index++
				label[flip] ^= 32
			}
			label[swap], label[swap+1] = label[swap+1], label[swap]
		}
	}
	lines[index] = label
}

// PrepareLibrary reports the exact NPN orbit sizes without requiring a global
// library. Enumeration is bounded by six inputs, matching the original helper.
func PrepareLibrary(lib *Mio_Library_t) map[string]int {
	counts := make(map[string]int)
	roots := Mio_CollectRoots(lib, 6, 1e20, 1, nil, 0)
	for _, gate := range roots {
		n := int(gate.nInputs)
		perm := make([]int, n)
		for i := range perm {
			perm[i] = i
		}
		unique := make(map[uint64]struct{})
		var enumerate func(int)
		enumerate = func(start int) {
			if start < n {
				for i := start; i < n; i++ {
					perm[start], perm[i] = perm[i], perm[start]
					enumerate(start + 1)
					perm[start], perm[i] = perm[i], perm[start]
				}
				return
			}
			var transformed uint64
			for m := 0; m < 64; m++ {
				source := m &^ ((1 << uint(n)) - 1)
				for i, v := range perm {
					source |= (m >> uint(i) & 1) << uint(v)
				}
				transformed |= (gate.uTruth >> uint(source) & 1) << uint(m)
			}
			current := transformed
			for phase := 0; phase < 1<<uint(n); phase++ {
				unique[current] = struct{}{}
				unique[^current] = struct{}{}
				next := (phase + 1) % (1 << uint(n))
				grayChange := (phase ^ (phase >> 1)) ^ (next ^ (next >> 1))
				current = mioFlip(current, int32(bits.TrailingZeros(uint(grayChange))))
			}
		}
		enumerate(0)
		counts[gate.pName] = len(unique)
	}
	return counts
}
func Nf_ManPrepareLibrary(lib *Mio_Library_t) {
	counts := PrepareLibrary(lib)
	for g := lib.pGates; g != nil; g = g.pNext {
		if count, ok := counts[g.pName]; ok {
			fmt.Printf("%s: Classes = %d\n", g.pName, count)
		}
	}
}
func Nf_ManPrepareLibraryTest2() {
	lib := mioCurrentLibrary()
	if lib == nil {
		fmt.Println("Standard cell library is not available.")
		return
	}
	Nf_ManPrepareLibrary(lib)
}
