// Compatible bound-set decomposition checks from ifDec08.c.
package if_

import (
	"fmt"
	"math/bits"
)

func decCofCount(t []uint64, n int, bound []int, fixed, value int) int {
	free := decComplement(n, bound)
	if len(free) > 6 {
		panic("cofactor signature exceeds one word")
	}
	patterns := make(map[uint64]struct{})
	for b := 0; b < 1<<uint(len(bound)); b++ {
		base := decInsert(b, bound)
		var pattern uint64
		for f := 0; f < 1<<uint(len(free)); f++ {
			minterm := base | decInsert(f, free)
			if fixed >= 0 {
				minterm = minterm&^(1<<uint(fixed)) | value<<uint(fixed)
			}
			pattern |= decBit(t, minterm) << uint(f)
		}
		patterns[pattern] = struct{}{}
	}
	return len(patterns)
}

// decCompatible follows the source's evolving variable order, including its
// fourth candidate shared variable outside the selected bound set.
func decCompatible(t []uint64, n, capacity, boundSize int) int32 {
	decCheck(t, int32(n))
	if n < 6 || n > capacity || decSupport(t, int32(n)) != int32((1<<uint(n))-1) {
		return 0
	}
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	move := func(variable, position int) {
		from := 0
		for order[from] != variable {
			from++
		}
		if from > position {
			panic("invalid bound-set permutation")
		}
		copy(order[from:position], order[from+1:position+1])
		order[position] = variable
	}
	type candidate struct{ mask, count int }
	var candidates []candidate
	decCombinations(n, boundSize, func(bound []int) bool {
		mask := 0
		for i, variable := range bound {
			move(variable, n-1-i)
			mask |= 1 << uint(variable)
		}
		count := decCofCount(t, n, bound, -1, 0)
		if count >= 3 && count <= 5 {
			for i := 0; i < 4; i++ {
				variable := order[n-1-i]
				if decCofCount(t, n, bound, variable, 0) <= 2 && decCofCount(t, n, bound, variable, 1) <= 2 {
					count = -count
					break
				}
			}
		}
		candidates = append(candidates, candidate{mask, count})
		return false
	})
	for i, a := range candidates {
		for _, b := range candidates[i+1:] {
			if bits.OnesCount(uint(a.mask&b.mask)) > capacity-n {
				continue
			}
			if a.count == 2 && b.count == 2 {
				return 1
			}
			if n < capacity && (a.count == 2 && b.count < 0 || a.count < 0 && b.count == 2) {
				return 1
			}
			if n < capacity-1 && a.count < 0 && b.count < 0 {
				return 1
			}
		}
	}
	return 0
}
func If_Dec08Perform(t []uint64, n, derive int32) int32 { return decCompatible(t, int(n), 8, 3) }
func If_Dec08WordNum(n int32) int32                     { return int32(decWords(n)) }
func If_Dec08HasVar(t []uint64, n, v int32) int32 {
	if v < 0 || v >= n {
		panic("variable outside truth table")
	}
	if decSupport(t, n)>>uint(v)&1 != 0 {
		return 1
	}
	return 0
}
func If_Dec08Support(t []uint64, n int32) int32 { return decSupport(t, n) }
func If_Dec08Cofactors(t []uint64, n, v int32, zero, one []uint64) {
	words := decWords(n)
	if len(zero) < words || len(one) < words {
		panic("short cofactor output")
	}
	a, b := decCofactor(t, n, v, 0), decCofactor(t, n, v, 1)
	copy(zero, a)
	copy(one, b)
}
func If_Dec08CofCount(t []uint64, n int32) int32 {
	if n < 6 || n > 8 {
		panic("Dec08 support outside 6..8")
	}
	return int32(decCofCount(t, int(n), []int{int(n) - 3, int(n) - 2, int(n) - 1}, -1, 0))
}
func If_Dec08CofCount2(t []uint64, n int32) int32 {
	count := If_Dec08CofCount(t, n)
	if count <= 2 {
		return 2
	}
	return 3
}
func If_Dec08Copy(out, in []uint64, n int32) {
	words := decWords(n)
	if len(out) < words || len(in) < words {
		panic("short truth copy")
	}
	copy(out[:words], in[:words])
}
func decSwap(out, in []uint64, n, v int32) {
	decCheck(in, n)
	if v < 0 || v >= n-1 || len(out) < decWords(n) {
		panic("invalid adjacent variable swap")
	}
	temp := make([]uint64, decWords(n))
	for m := 0; m < len(temp)*64; m++ {
		source := m
		if (m >> uint(v) & 1) != (m >> uint(v+1) & 1) {
			source ^= 3 << uint(v)
		}
		if decBit(in, source) != 0 {
			decSet(temp, m)
		}
	}
	copy(out, temp)
}
func If_Dec08SwapAdjacent(out, in []uint64, v, n int32) { decSwap(out, in, n, v) }
func decMoveTo(t []uint64, n, v, position int32, placeToVar, varToPlace []int32) {
	if v < 0 || v >= n || position < 0 || position >= n || len(placeToVar) < int(n) || len(varToPlace) < int(n) {
		panic("invalid variable permutation")
	}
	from := varToPlace[v]
	if from > position {
		panic("move requires increasing position")
	}
	for from < position {
		decSwap(t, t, n, from)
		other := placeToVar[from+1]
		placeToVar[from], placeToVar[from+1] = other, v
		varToPlace[other], varToPlace[v] = from, from+1
		from++
	}
}
func If_Dec08MoveTo(t []uint64, n, v, position int32, placeToVar, varToPlace []int32) {
	decMoveTo(t, n, v, position, placeToVar, varToPlace)
}
func If_Dec08Count16(value int32) int32 { return int32(bits.OnesCount16(uint16(value))) }
func If_DecTruthStretch(t uint64, n int32) uint64 {
	if n < 0 || n > 6 {
		panic("stretch support outside 0..6")
	}
	width := 1 << uint(n)
	if width == 64 {
		return t
	}
	mask := uint64(1)<<uint(width) - 1
	t &= mask
	for size := width; size < 64; size *= 2 {
		t |= t << uint(size)
	}
	return t
}
func decComposeLut4(table uint16, inputs [][]uint64, out []uint64, n int32) {
	words := decWords(n)
	if len(inputs) < 4 || len(out) < words {
		panic("short LUT composition")
	}
	result := make([]uint64, words)
	for m := 0; m < 16; m++ {
		if table>>uint(m)&1 == 0 {
			continue
		}
		for w := 0; w < words; w++ {
			cube := ^uint64(0)
			for pin := 0; pin < 4; pin++ {
				if m>>uint(pin)&1 != 0 {
					cube &= inputs[pin][w]
				} else {
					cube &= ^inputs[pin][w]
				}
			}
			result[w] |= cube
		}
	}
	copy(out, result)
}
func If_Dec08ComposeLut4(table int32, inputs [][]uint64, out []uint64, n int32) {
	decComposeLut4(uint16(table), inputs, out, n)
}
func decVerifyConfig(t []uint64, n int32, configuration []uint32) {
	decCheck(t, n)
	nodes := make([][]uint64, n)
	for v := range nodes {
		nodes[v] = make([]uint64, decWords(n))
		for m := 0; m < len(nodes[v])*64; m++ {
			if m>>uint(v)&1 != 0 {
				decSet(nodes[v], m)
			}
		}
	}
	gates := 0
	for _, config := range configuration {
		if config == 0 {
			break
		}
		inputs := make([][]uint64, 4)
		for pin := 0; pin < 4; pin++ {
			index := int(config >> uint(16+4*pin) & 15)
			if index >= len(nodes) {
				panic("configuration references unavailable LUT input")
			}
			inputs[pin] = nodes[index]
		}
		out := make([]uint64, decWords(n))
		decComposeLut4(uint16(config), inputs, out, n)
		nodes = append(nodes, out)
		gates++
	}
	if gates == 0 {
		panic("empty LUT configuration")
	}
	result := nodes[len(nodes)-1]
	for i := range result {
		if result[i] != t[i] {
			panic("LUT configuration verification failed")
		}
	}
}
func If_Dec08Verify(t []uint64, n int32, z []uint32) { decVerifyConfig(t, n, z) }
func If_Dec08PrintConfigOne(z uint32) {
	fmt.Printf("%016b", z&0xffff)
	for pin := 0; pin < 4; pin++ {
		fmt.Printf(" %d", z>>uint(16+4*pin)&15)
	}
	fmt.Println()
}
func If_Dec08PrintConfig(z []uint32) {
	for _, config := range z {
		if config == 0 {
			break
		}
		If_Dec08PrintConfigOne(config)
	}
}
func If_CutPerformCheck08(_ *If_Man_t, truth []uint32, nVars, leaves int32, _ string) int32 {
	if leaves < 0 || leaves > 8 || nVars < leaves {
		return 0
	}
	if leaves <= 5 {
		return 1
	}
	return If_Dec08Perform(decFrom32(truth, leaves), leaves, 0)
}
