// Two-LUT4 decomposition from ifDec07.c. Cofactor matrices express the same
// bound/free-variable search without aliasing truth words or mutating inputs.
package if_

import (
	"fmt"
	"math/bits"
)

func decBit(t []uint64, m int) uint64 { return t[m>>6] >> uint(m&63) & 1 }
func decWords(n int32) int {
	if n < 0 || n > 10 {
		panic("decomposition variable count outside 0..10")
	}
	if n <= 6 {
		return 1
	}
	return 1 << uint(n-6)
}
func decCheck(t []uint64, n int32) {
	if len(t) < decWords(n) {
		panic("short decomposition truth table")
	}
}
func decSet(t []uint64, m int) { t[m>>6] |= 1 << uint(m&63) }
func decCofactor(t []uint64, n, v int32, value int) []uint64 {
	decCheck(t, n)
	if v < 0 || v >= n {
		panic("cofactor variable outside truth table")
	}
	out := make([]uint64, decWords(n))
	for m := 0; m < len(out)*64; m++ {
		source := m &^ (1 << uint(v))
		if value != 0 {
			source |= 1 << uint(v)
		}
		if decBit(t, source) != 0 {
			decSet(out, m)
		}
	}
	return out
}
func decSupport(t []uint64, n int32) int32 {
	decCheck(t, n)
	var support int32
	for v := int32(0); v < n; v++ {
		step := 1 << uint(v)
		for m := 0; m < 1<<uint(n); m++ {
			if m&step == 0 && decBit(t, m) != decBit(t, m|step) {
				support |= 1 << uint(v)
				break
			}
		}
	}
	return support
}
func decCombinations(n, k int, visit func([]int) bool) bool {
	choice := make([]int, k)
	var choose func(int, int) bool
	choose = func(next, depth int) bool {
		if depth == k {
			return visit(choice)
		}
		for v := next; v <= n-(k-depth); v++ {
			choice[depth] = v
			if choose(v+1, depth+1) {
				return true
			}
		}
		return false
	}
	return choose(0, 0)
}
func decInsert(assignment int, variables []int) int {
	result := 0
	for bit, v := range variables {
		result |= (assignment >> uint(bit) & 1) << uint(v)
	}
	return result
}
func decProject(minterm int, variables []int) int {
	result := 0
	for bit, v := range variables {
		result |= (minterm >> uint(v) & 1) << uint(bit)
	}
	return result
}
func decComplement(n int, variables []int) []int {
	mask := 0
	for _, v := range variables {
		mask |= 1 << uint(v)
	}
	var result []int
	for v := 0; v < n; v++ {
		if mask>>uint(v)&1 == 0 {
			result = append(result, v)
		}
	}
	return result
}

// decCascade derives H(B), G(F,H). Shared pins partition the cofactor matrix;
// each partition must contain at most two distinct column functions.
func decCascade(t []uint64, n int, bound, free []int) (uint64, bool) {
	var shared, outside []int
	boundMask := 0
	for _, v := range bound {
		boundMask |= 1 << uint(v)
	}
	for _, v := range free {
		if boundMask>>uint(v)&1 != 0 {
			shared = append(shared, v)
		} else {
			outside = append(outside, v)
		}
	}
	type group struct {
		patterns [2]uint64
		count    int
	}
	groups := make([]group, 1<<uint(len(shared)))
	var h uint64
	for b := 0; b < 16; b++ {
		base := decInsert(b, bound)
		var pattern uint64
		for f := 0; f < 1<<uint(len(outside)); f++ {
			pattern |= decBit(t, base|decInsert(f, outside)) << uint(f)
		}
		group := &groups[decProject(base, shared)]
		code := 0
		if group.count == 0 {
			group.patterns[0] = pattern
			group.count = 1
		} else if pattern != group.patterns[0] {
			if group.count == 1 {
				group.patterns[1] = pattern
				group.count = 2
			} else if pattern != group.patterns[1] {
				return 0, false
			}
			code = 1
		}
		h |= uint64(code) << uint(b)
	}
	var g uint64
	for assignment := 0; assignment < 16; assignment++ {
		minterm := decInsert(assignment&7, free)
		group := groups[decProject(minterm, shared)]
		code := assignment >> 3
		if group.count == 1 {
			code = 0
		}
		g |= (group.patterns[code] >> uint(decProject(minterm, outside)) & 1) << uint(assignment)
	}
	z := h | g<<32
	for pin, v := range bound {
		z |= uint64(v) << uint(16+4*pin)
	}
	for pin, v := range free {
		z |= uint64(v) << uint(48+4*pin)
	}
	z |= uint64(7) << 60
	return z, true
}
func decPerformSmall(t []uint64, n int, derive int32) uint64 {
	decCheck(t, int32(n))
	var result uint64
	decCombinations(n, 4, func(bound []int) bool {
		outside := decComplement(n, bound)
		return decCombinations(4, 7-n, func(sharedIndices []int) bool {
			free := append([]int(nil), outside...)
			for _, index := range sharedIndices {
				free = append(free, bound[index])
			}
			z, ok := decCascade(t, n, bound, free)
			if !ok {
				return false
			}
			result = z
			if derive == 0 {
				result = 1
			}
			return true
		})
	})
	return result
}
func If_Dec5Perform(t uint64, derive int32) uint64   { return decPerformSmall([]uint64{t}, 5, derive) }
func If_Dec6Perform(t uint64, derive int32) uint64   { return decPerformSmall([]uint64{t}, 6, derive) }
func If_Dec7Perform(t []uint64, derive int32) uint64 { return decPerformSmall(t, 7, derive) }
func decConfigTruth(z uint64, n int32) []uint64 {
	if z == 0 {
		panic("zero LUT configuration")
	}
	out := make([]uint64, decWords(n))
	for minterm := 0; minterm < len(out)*64; minterm++ {
		lower := 0
		for pin := 0; pin < 4; pin++ {
			v := int(z >> uint(16+4*pin) & 7)
			if v == 7 {
				panic("LUT references its own output")
			}
			if v == 6 && n <= 6 {
				continue
			}
			if v >= int(n) {
				panic("LUT pin outside truth support")
			}
			lower |= (minterm >> uint(v) & 1) << uint(pin)
		}
		h := int(z >> uint(lower) & 1)
		upper := 0
		for pin := 0; pin < 4; pin++ {
			v := int(z >> uint(48+4*pin) & 7)
			value := 0
			if v == 7 {
				value = h
			} else if v != 6 || n > 6 {
				if v >= int(n) {
					panic("LUT pin outside truth support")
				}
				value = minterm >> uint(v) & 1
			}
			upper |= value << uint(pin)
		}
		if z>>uint(32+upper)&1 != 0 {
			decSet(out, minterm)
		}
	}
	return out
}
func If_Dec6Truth(z uint64) uint64 { return decConfigTruth(z, 6)[0] }
func If_Dec6Verify(t, z uint64) {
	if If_Dec6Truth(z) != t {
		panic("six-input LUT decomposition verification failed")
	}
}
func If_Dec7Verify(t []uint64, z uint64) {
	decCheck(t, 7)
	out := decConfigTruth(z, 7)
	if out[0] != t[0] || out[1] != t[1] {
		panic("seven-input LUT decomposition verification failed")
	}
}
func If_DecPrintConfig(z uint64) {
	for lut := 0; lut < 2; lut++ {
		part := uint32(z >> uint(32*lut))
		fmt.Printf("%016b", part&0xffff)
		for pin := 0; pin < 4; pin++ {
			fmt.Printf(" %d", part>>uint(16+4*pin)&7)
		}
		if lut == 0 {
			fmt.Print("   ")
		}
	}
	fmt.Println()
}
func If_Dec6HasVar(t uint64, v int32) int32 {
	if v < 0 || v >= 6 {
		panic("truth variable outside 0..5")
	}
	if decSupport([]uint64{t}, 6)>>uint(v)&1 != 0 {
		return 1
	}
	return 0
}
func If_Dec7HasVar(t []uint64, v int32) int32 {
	if v < 0 || v >= 7 {
		panic("truth variable outside 0..6")
	}
	if decSupport(t, 7)>>uint(v)&1 != 0 {
		return 1
	}
	return 0
}
func If_Dec6Cofactor(t uint64, v, value int32) uint64 {
	return decCofactor([]uint64{t}, 6, v, int(value))[0]
}
func If_Dec7Cofactor(t []uint64, v, value int32, out []uint64) {
	if len(out) < 2 {
		panic("short seven-variable cofactor output")
	}
	copy(out, decCofactor(t, 7, v, int(value)))
}
func If_Dec6CountOnes(t uint64) int32 { return int32(bits.OnesCount64(t)) }
func If_Dec6SuppSize(t uint64) int32 {
	return int32(bits.OnesCount32(uint32(decSupport([]uint64{t}, 6))))
}
func If_Dec7SuppSize(t []uint64) int32 { return int32(bits.OnesCount32(uint32(decSupport(t, 7)))) }
func If_DecSuppIsMinBase(support int32) int32 {
	if support&(support+1) == 0 {
		return 1
	}
	return 0
}
func decMinimumBase(t []uint64, n int32, support []int32, count *int32) []uint64 {
	mask := decSupport(t, n)
	var variables []int
	for v := int32(0); v < n; v++ {
		if mask>>uint(v)&1 != 0 {
			if support != nil {
				support[len(variables)] = support[v]
			}
			variables = append(variables, int(v))
		}
	}
	if count != nil {
		*count = int32(len(variables))
	}
	out := make([]uint64, decWords(n))
	for minterm := 0; minterm < len(out)*64; minterm++ {
		if decBit(t, decInsert(minterm, variables)) != 0 {
			decSet(out, minterm)
		}
	}
	return out
}
func If_Dec6MinimumBase(t uint64, support []int32, n int32, count *int32) uint64 {
	if n < 0 || n > 6 {
		panic("six-input minimum base exceeds support")
	}
	return decMinimumBase([]uint64{t}, n, support, count)[0]
}
func If_Dec7MinimumBase(t []uint64, support []int32, n int32, count *int32) {
	if n < 0 || n > 7 {
		panic("seven-input minimum base exceeds support")
	}
	copy(t, decMinimumBase(t, n, support, count))
}
func decPickMux(t []uint64, n int32, zero, one []uint64) int32 {
	best := int32(-1)
	cost := 1000
	for v := int32(0); v < n; v++ {
		a, b := decCofactor(t, n, v, 0), decCofactor(t, n, v, 1)
		ca, cb := bits.OnesCount32(uint32(decSupport(a, n))), bits.OnesCount32(uint32(decSupport(b, n)))
		if ca < 5 && cb < 5 && ca+cb < cost {
			best, cost = v, ca+cb
			copy(zero, a)
			copy(one, b)
		}
	}
	return best
}
func If_Dec6PickBestMux(t uint64, cofactors []uint64) int32 {
	if len(cofactors) < 2 {
		panic("short MUX cofactor output")
	}
	var a, b [1]uint64
	best := decPickMux([]uint64{t}, 6, a[:], b[:])
	if best >= 0 {
		cofactors[0], cofactors[1] = a[0], b[0]
	}
	return best
}
func If_Dec7PickBestMux(t, zero, one []uint64) int32 {
	if len(zero) < 2 || len(one) < 2 {
		panic("short MUX cofactor output")
	}
	return decPickMux(t, 7, zero, one)
}
func If_Dec6CheckMux(t uint64) int32   { var out [2]uint64; return If_Dec6PickBestMux(t, out[:]) }
func If_Dec7CheckMux(t []uint64) int32 { var a, b [2]uint64; return If_Dec7PickBestMux(t, a[:], b[:]) }
func If_Dec5PerformEx() uint64 {
	return uint64(0x17ac) | uint64(3)<<16 | uint64(4)<<20 | uint64(1)<<24 | uint64(2)<<28 | uint64(0x179a)<<32 | uint64(7)<<52 | uint64(1)<<56 | uint64(2)<<60
}
func If_Dec5PerformTest() {
	truth := uint64(0xB0F3B0FFB0F3B0FF)
	If_Dec6Verify(truth, If_Dec5Perform(truth, 1))
}
func decFrom32(truth []uint32, n int32) []uint64 {
	words := decWords(n)
	required := (1<<uint(n) + 31) / 32
	if len(truth) < required {
		panic("short 32-bit truth table")
	}
	out := make([]uint64, words)
	for i := range out {
		out[i] = uint64(truth[2*i])
		if 2*i+1 < len(truth) {
			out[i] |= uint64(truth[2*i+1]) << 32
		} else {
			out[i] |= uint64(truth[2*i]) << 32
		}
	}
	return out
}
func If_CutPerformDerive07(_ *If_Man_t, truth []uint32, nVars, leaves int32, _ string) uint64 {
	if leaves < 0 || leaves > 7 || nVars < leaves {
		return 0
	}
	if leaves < 5 {
		return 1
	}
	t := decFrom32(truth, leaves)
	switch leaves {
	case 5:
		return If_Dec5Perform(t[0], 1)
	case 6:
		return If_Dec6Perform(t[0], 1)
	default:
		return If_Dec7Perform(t, 1)
	}
}
func If_CutPerformCheck07(_ *If_Man_t, truth []uint32, nVars, leaves int32, _ string) int32 {
	if leaves == 0 && nVars >= 0 {
		return 1
	}
	if leaves < 0 || leaves > 7 || nVars < leaves {
		return 0
	}
	t := decFrom32(truth, leaves)
	if decSupport(t, leaves) != (1<<uint(leaves))-1 {
		return 0
	}
	if leaves < 5 {
		return 1
	}
	var z uint64
	switch leaves {
	case 5:
		z = If_Dec5Perform(t[0], 0)
	case 6:
		z = If_Dec6Perform(t[0], 0)
	default:
		z = If_Dec7Perform(t, 0)
	}
	if z != 0 {
		return 1
	}
	return 0
}
