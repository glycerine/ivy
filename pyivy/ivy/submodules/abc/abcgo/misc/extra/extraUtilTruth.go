package extra

import (
	"math/bits"
	"slices"
)

/**CFile****************************************************************

  FileName    [extraUtilMisc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Various procedures for truth table manipulation.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilMisc.c,v 1.0 2003/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

var truthVariableMasks = [5]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}

func truthWords(n int32) int {
	if n < 0 || n > 30 {
		panic("truth variable count outside supported range")
	}
	if n <= 5 {
		return 1
	}
	return 1 << uint(n-5)
}
func truthTable(table []uint32, n int32) []uint32 { return table[:truthWords(n)] }
func truthVariable(n, v int32) {
	truthWords(n)
	if v < 0 || v >= n {
		panic("truth variable outside support domain")
	}
}
func Extra_TruthElementary(n int32) [][]uint32 {
	count := truthWords(n)
	result := make([][]uint32, n)
	for v := range result {
		result[v] = make([]uint32, count)
		for w := range result[v] {
			if v < 5 {
				result[v][w] = truthVariableMasks[v]
			} else if w&(1<<uint(v-5)) != 0 {
				result[v][w] = ^uint32(0)
			}
		}
	}
	return result
}
func Extra_TruthSwapAdjacentVars(out, in []uint32, n, v int32) {
	truthVariable(n, v)
	if v+1 >= n {
		panic("adjacent truth variable outside domain")
	}
	in = slices.Clone(truthTable(in, n))
	out = truthTable(out, n)
	switch {
	case v < 4:
		masks := [4][3]uint32{{0x99999999, 0x22222222, 0x44444444}, {0xc3c3c3c3, 0x0c0c0c0c, 0x30303030}, {0xf00ff00f, 0x00f000f0, 0x0f000f00}, {0xff0000ff, 0x0000ff00, 0x00ff0000}}
		mask := masks[v]
		shift := uint(1) << uint(v)
		for w, value := range in {
			out[w] = value&mask[0] | (value&mask[1])<<shift | (value&mask[2])>>shift
		}
	case v == 4:
		for w := 0; w < len(in); w += 2 {
			a, b := in[w], in[w+1]
			out[w] = a&0xffff | b<<16
			out[w+1] = b&0xffff0000 | a>>16
		}
	default:
		step := 1 << uint(v-5)
		for base := 0; base < len(in); base += 4 * step {
			copy(out[base:base+step], in[base:base+step])
			copy(out[base+step:base+2*step], in[base+2*step:base+3*step])
			copy(out[base+2*step:base+3*step], in[base+step:base+2*step])
			copy(out[base+3*step:base+4*step], in[base+3*step:base+4*step])
		}
	}
}
func Extra_TruthSwapAdjacentVars2(in, out []uint32, n, v int32) {
	Extra_TruthSwapAdjacentVars(out, in, n, v)
}
func truthPlacement(n, all int32, phase uint32) {
	truthWords(all)
	if n < 0 || n > all || phase>>uint(all) != 0 || bits.OnesCount32(phase) != int(n) {
		panic("invalid truth support placement")
	}
}
func Extra_TruthStretch(out, in []uint32, n, all int32, phase uint32) {
	truthPlacement(n, all, phase)
	work := slices.Clone(truthTable(in, all))
	v := n - 1
	for i := all - 1; i >= 0; i-- {
		if phase&(1<<uint(i)) != 0 {
			for k := v; k < i; k++ {
				Extra_TruthSwapAdjacentVars(work, work, all, k)
			}
			v--
		}
	}
	copy(truthTable(out, all), work)
}
func Extra_TruthShrink(out, in []uint32, n, all int32, phase uint32) {
	truthPlacement(n, all, phase)
	work := slices.Clone(truthTable(in, all))
	v := int32(0)
	for i := int32(0); i < all; i++ {
		if phase&(1<<uint(i)) != 0 {
			for k := i - 1; k >= v; k-- {
				Extra_TruthSwapAdjacentVars(work, work, all, k)
			}
			v++
		}
	}
	copy(truthTable(out, all), work)
}
func Extra_TruthVarInSupport(table []uint32, n, v int32) int32 {
	truthVariable(n, v)
	table = truthTable(table, n)
	if v < 5 {
		mask := truthVariableMasks[v]
		shift := uint(1) << uint(v)
		for _, word := range table {
			if word&^mask != (word&mask)>>shift {
				return 1
			}
		}
		return 0
	}
	step := 1 << uint(v-5)
	for base := 0; base < len(table); base += 2 * step {
		if !slices.Equal(table[base:base+step], table[base+step:base+2*step]) {
			return 1
		}
	}
	return 0
}
func Extra_TruthSupport(table []uint32, n int32) int32 {
	truthWords(n)
	var support int32
	for v := int32(0); v < n; v++ {
		if Extra_TruthVarInSupport(table, n, v) != 0 {
			support |= 1 << uint(v)
		}
	}
	return support
}
func Extra_TruthSupportSize(table []uint32, n int32) int32 {
	return int32(bits.OnesCount32(uint32(Extra_TruthSupport(table, n))))
}

// Cofactors replicate their selected branch across both values of the variable.
func truthCofactor(table []uint32, n, v int32, positive bool) {
	truthVariable(n, v)
	table = truthTable(table, n)
	if v < 5 {
		mask := truthVariableMasks[v]
		shift := uint(1) << uint(v)
		for w, value := range table {
			if positive {
				value &= mask
				table[w] = value | value>>shift
			} else {
				value &^= mask
				table[w] = value | value<<shift
			}
		}
		return
	}
	step := 1 << uint(v-5)
	for base := 0; base < len(table); base += 2 * step {
		if positive {
			copy(table[base:base+step], table[base+step:base+2*step])
		} else {
			copy(table[base+step:base+2*step], table[base:base+step])
		}
	}
}
func Extra_TruthCofactor0(table []uint32, n, v int32) { truthCofactor(table, n, v, false) }
func Extra_TruthCofactor1(table []uint32, n, v int32) { truthCofactor(table, n, v, true) }
func truthQuantify(table []uint32, n, v int32, forall bool) {
	truthVariable(n, v)
	table = truthTable(table, n)
	if v < 5 {
		mask := truthVariableMasks[v]
		shift := uint(1) << uint(v)
		for w, value := range table {
			swapped := (value&mask)>>shift | (value&^mask)<<shift
			if forall {
				table[w] = value & swapped
			} else {
				table[w] = value | swapped
			}
		}
		return
	}
	step := 1 << uint(v-5)
	for base := 0; base < len(table); base += 2 * step {
		for w := 0; w < step; w++ {
			a, b := base+w, base+step+w
			value := table[a] | table[b]
			if forall {
				value = table[a] & table[b]
			}
			table[a], table[b] = value, value
		}
	}
}
func Extra_TruthExist(table []uint32, n, v int32)  { truthQuantify(table, n, v, false) }
func Extra_TruthForall(table []uint32, n, v int32) { truthQuantify(table, n, v, true) }
func Extra_TruthMux(out, zero, one []uint32, n, v int32) {
	truthVariable(n, v)
	out = truthTable(out, n)
	zero = slices.Clone(truthTable(zero, n))
	one = slices.Clone(truthTable(one, n))
	if v < 5 {
		mask := truthVariableMasks[v]
		for w := range out {
			out[w] = zero[w]&^mask | one[w]&mask
		}
		return
	}
	step := 1 << uint(v-5)
	for base := 0; base < len(out); base += 2 * step {
		copy(out[base:base+step], zero[base:base+step])
		copy(out[base+step:base+2*step], one[base+step:base+2*step])
	}
}
func truthSymmetry(table []uint32, n, a, b int32, anti bool) int32 {
	truthVariable(n, a)
	truthVariable(n, b)
	zero := slices.Clone(truthTable(table, n))
	one := slices.Clone(zero)
	Extra_TruthCofactor0(zero, n, a)
	Extra_TruthCofactor1(one, n, a)
	if anti {
		Extra_TruthCofactor0(zero, n, b)
		Extra_TruthCofactor1(one, n, b)
	} else {
		Extra_TruthCofactor1(zero, n, b)
		Extra_TruthCofactor0(one, n, b)
	}
	if slices.Equal(zero, one) {
		return 1
	}
	return 0
}
func Extra_TruthVarsSymm(table []uint32, n, a, b int32) int32 {
	return truthSymmetry(table, n, a, b, false)
}
func Extra_TruthVarsAntiSymm(table []uint32, n, a, b int32) int32 {
	return truthSymmetry(table, n, a, b, true)
}
func Extra_TruthChangePhase(table []uint32, n, v int32) {
	truthVariable(n, v)
	table = truthTable(table, n)
	if v < 5 {
		mask := truthVariableMasks[v]
		shift := uint(1) << uint(v)
		for w, value := range table {
			table[w] = (value&mask)>>shift | (value&^mask)<<shift
		}
		return
	}
	step := 1 << uint(v-5)
	for base := 0; base < len(table); base += 2 * step {
		for w := 0; w < step; w++ {
			a, b := base+w, base+step+w
			table[a], table[b] = table[b], table[a]
		}
	}
}
func Extra_TruthMinCofSuppOverlap(table []uint32, n int32) (minimum, variable int32) {
	truthWords(n)
	minimum, variable = 32, -1
	for v := int32(0); v < n; v++ {
		zero := slices.Clone(truthTable(table, n))
		one := slices.Clone(zero)
		Extra_TruthCofactor0(zero, n, v)
		Extra_TruthCofactor1(one, n, v)
		a, b := uint32(Extra_TruthSupport(zero, n)), uint32(Extra_TruthSupport(one, n))
		count := int32(bits.OnesCount32(a & b))
		if count < minimum && bits.OnesCount32(a) <= 5 && bits.OnesCount32(b) <= 5 {
			minimum, variable = count, v
		}
		if minimum == 0 {
			break
		}
	}
	return
}

// Counts use int32 rather than C short, avoiding overflow for larger tables.
func Extra_TruthCountOnesInCofs(table []uint32, n int32, store []int32) {
	table = truthTable(table, n)
	store = store[:2*n]
	clear(store)
	for w, value := range table {
		for v := int32(0); v < n; v++ {
			if v < 5 {
				mask := truthVariableMasks[v]
				store[2*v] += int32(bits.OnesCount32(value &^ mask))
				store[2*v+1] += int32(bits.OnesCount32(value & mask))
			} else {
				phase := (w >> uint(v-5)) & 1
				store[2*int(v)+phase] += int32(bits.OnesCount32(value))
			}
		}
	}
}

var truthHashPrimes = truthFirstPrimes()

func truthFirstPrimes() [1024]uint32 {
	var primes [1024]uint32
	count := 0
	for candidate := uint32(2); count < len(primes); candidate++ {
		prime := true
		for i := 0; i < count && primes[i]*primes[i] <= candidate; i++ {
			if candidate%primes[i] == 0 {
				prime = false
				break
			}
		}
		if prime {
			primes[count] = candidate
			count++
		}
	}
	return primes
}
func Extra_TruthHash(table []uint32, count int32) uint32 {
	if count < 0 || count > 1024 {
		panic("truth hash word count outside supported range")
	}
	var hash uint32
	for i, value := range table[:count] {
		hash ^= truthHashPrimes[i] * value
	}
	return hash
}
func Extra_TruthSemiCanonicize(inout, aux []uint32, n int32, perm []byte, store []int32) uint32 {
	inout = truthTable(inout, n)
	aux = truthTable(aux, n)
	perm = perm[:n]
	store = store[:2*n]
	var phase uint32
	ones := 0
	for _, word := range inout {
		ones += bits.OnesCount32(word)
	}
	if ones > len(inout)*16 || ones == len(inout)*16 && inout[0]&1 != 0 {
		phase |= 1 << uint(n)
		for i := range inout {
			inout[i] = ^inout[i]
		}
	}
	Extra_TruthCountOnesInCofs(inout, n, store)
	for v := int32(0); v < n; v++ {
		if store[2*v] > store[2*v+1] {
			phase |= 1 << uint(v)
			store[2*v], store[2*v+1] = store[2*v+1], store[2*v]
			Extra_TruthChangePhase(inout, n, v)
		}
	}
	// Stable adjacent sorting preserves ABC's tie order and permutation encoding.
	for changed := true; changed; {
		changed = false
		for v := int32(0); v < n-1; v++ {
			if store[2*v] <= store[2*v+2] {
				continue
			}
			changed = true
			perm[v], perm[v+1] = perm[v+1], perm[v]
			store[2*v], store[2*v+2] = store[2*v+2], store[2*v]
			store[2*v+1], store[2*v+3] = store[2*v+3], store[2*v+1]
			Extra_TruthSwapAdjacentVars(aux, inout, n, v)
			copy(inout, aux)
		}
	}
	return phase
}
