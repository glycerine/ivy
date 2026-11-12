package kit

import (
	"fmt"
	"io"
	"math/bits"
	"os"
	"path/filepath"
	"sort"
)

func kitRemapTruth(out, in []uint32, nVars int32, index func(int) int) {
	count := int(Kit_TruthWordNum(nVars))
	source := append([]uint32(nil), in...)
	clear(out[:count])
	for x := 0; x < count*32; x++ {
		old := index(x)
		if source[old>>5]&(1<<uint(old&31)) != 0 {
			out[x>>5] |= 1 << uint(x&31)
		}
	}
}
func Kit_TruthSwapAdjacentVars(out, in []uint32, nVars, v int32) {
	if v < 0 || v >= nVars-1 {
		panic("kit: adjacent truth variable out of range")
	}
	mask := 1<<uint(v) | 1<<uint(v+1)
	kitRemapTruth(out, in, nVars, func(x int) int { return x&^mask | (x>>uint(v)&1)<<uint(v+1) | (x>>uint(v+1)&1)<<uint(v) })
}
func Kit_TruthSwapAdjacentVars2(in, out []uint32, nVars, start int32) {
	Kit_TruthSwapAdjacentVars(out, in, nVars, start)
}
func kitTruthMove(out, in []uint32, nVars, total int32, phase uint32, shrink bool, returnIn int32) {
	if nVars < 0 || total < nVars || total > 30 || bits.OnesCount32(phase) != int(nVars) || phase&^Kit_BitMask(total) != 0 {
		panic("kit: invalid truth support mapping")
	}
	order := make([]int32, 0, total)
	for v := int32(0); v < total; v++ {
		if phase&(1<<uint(v)) != 0 {
			order = append(order, v)
		}
	}
	for v := int32(0); v < total; v++ {
		if phase&(1<<uint(v)) == 0 {
			order = append(order, v)
		}
	}
	kitRemapTruth(out, in, total, func(x int) int {
		old := x &^ int(Kit_BitMask(total))
		for new, original := range order {
			if shrink {
				old |= (x >> uint(new) & 1) << uint(original)
			} else {
				old |= (x >> uint(original) & 1) << uint(new)
			}
		}
		return old
	})
	if returnIn != 0 {
		Kit_TruthCopy(in, out, total)
	}
}
func Kit_TruthStretch(out, in []uint32, nVars, total int32, phase uint32, returnIn int32) {
	kitTruthMove(out, in, nVars, total, phase, false, returnIn)
}
func Kit_TruthShrink(out, in []uint32, nVars, total int32, phase uint32, returnIn int32) {
	kitTruthMove(out, in, nVars, total, phase, true, returnIn)
}
func Kit_TruthPermute(out, in []uint32, nVars int32, permutation []byte, returnIn int32) {
	if len(permutation) < int(nVars) {
		panic("kit: short truth permutation")
	}
	order := make([]int, nVars)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool { return permutation[order[i]] < permutation[order[j]] })
	for i := 1; i < len(order); i++ {
		if permutation[order[i]] == permutation[order[i-1]] {
			panic("kit: duplicate truth permutation label")
		}
	}
	kitRemapTruth(out, in, nVars, func(x int) int {
		old := x &^ int(Kit_BitMask(nVars))
		for new, original := range order {
			old |= (x >> uint(new) & 1) << uint(original)
		}
		return old
	})
	sort.Slice(permutation[:nVars], func(i, j int) bool { return permutation[i] < permutation[j] })
	if returnIn != 0 {
		Kit_TruthCopy(in, out, nVars)
	}
}
func Kit_TruthVarInSupport(truth []uint32, nVars, v int32) int32 {
	if v < 0 || v >= nVars {
		panic("kit: support variable out of range")
	}
	if kitTruthHasVar(truth, nVars, v) {
		return 1
	}
	return 0
}
func Kit_TruthSupportSize(truth []uint32, nVars int32) int32 {
	return int32(bits.OnesCount32(Kit_TruthSupport(truth, nVars)))
}
func Kit_TruthSupport(truth []uint32, nVars int32) uint32 {
	var support uint32
	for v := int32(0); v < nVars; v++ {
		if kitTruthHasVar(truth, nVars, v) {
			support |= 1 << uint(v)
		}
	}
	return support
}
func kitTruthCofactor(out, in []uint32, nVars, v int32, value int) {
	if v < 0 || v >= nVars {
		panic("kit: cofactor variable out of range")
	}
	words := int(Kit_TruthWordNum(nVars))
	if v < 5 {
		shift := uint(1) << uint(v)
		mask := kitVarMasks[v]
		for i := 0; i < words; i++ {
			x := in[i]
			if value == 0 {
				x &^= mask
				x |= x << shift
			} else {
				x &= mask
				x |= x >> shift
			}
			out[i] = x
		}
		return
	}
	block := 1 << uint(v-5)
	for start := 0; start < words; start += 2 * block {
		for i := 0; i < block; i++ {
			x := in[start+value*block+i]
			out[start+i] = x
			out[start+block+i] = x
		}
	}
}
func Kit_TruthCofactor0(truth []uint32, nVars, v int32)      { kitTruthCofactor(truth, truth, nVars, v, 0) }
func Kit_TruthCofactor1(truth []uint32, nVars, v int32)      { kitTruthCofactor(truth, truth, nVars, v, 1) }
func Kit_TruthCofactor0New(out, in []uint32, nVars, v int32) { kitTruthCofactor(out, in, nVars, v, 0) }
func Kit_TruthCofactor1New(out, in []uint32, nVars, v int32) { kitTruthCofactor(out, in, nVars, v, 1) }
func Kit_TruthCofactor0Count(truth []uint32, nVars, v int32) int32 {
	if v < 0 || v >= nVars {
		panic("kit: cofactor count variable out of range")
	}
	var count int32
	for i, x := range truth[:Kit_TruthWordNum(nVars)] {
		if v < 5 {
			count += int32(bits.OnesCount32(x &^ kitVarMasks[v]))
		} else if i&(1<<uint(v-5)) == 0 {
			count += int32(bits.OnesCount32(x))
		}
	}
	return count
}

func Kit_TruthVarIsVacuous(on, off []uint32, nVars, v int32) int32 {
	a, b := make([]uint32, Kit_TruthWordNum(nVars)), make([]uint32, Kit_TruthWordNum(nVars))
	Kit_TruthCofactor0New(a, on, nVars, v)
	Kit_TruthCofactor1New(b, off, nVars, v)
	if Kit_TruthIsDisjoint(a, b, nVars) == 0 {
		return 0
	}
	Kit_TruthCofactor1New(a, on, nVars, v)
	Kit_TruthCofactor0New(b, off, nVars, v)
	return Kit_TruthIsDisjoint(a, b, nVars)
}
func kitTruthQuantify(out, in []uint32, nVars, v int32, operation byte) {
	a, b := make([]uint32, Kit_TruthWordNum(nVars)), make([]uint32, Kit_TruthWordNum(nVars))
	Kit_TruthCofactor0New(a, in, nVars, v)
	Kit_TruthCofactor1New(b, in, nVars, v)
	switch operation {
	case '|':
		Kit_TruthOr(out, a, b, nVars)
	case '&':
		Kit_TruthAnd(out, a, b, nVars)
	case '^':
		Kit_TruthXor(out, a, b, nVars)
	}
}
func Kit_TruthExist(truth []uint32, nVars, v int32)       { kitTruthQuantify(truth, truth, nVars, v, '|') }
func Kit_TruthExistNew(out, in []uint32, nVars, v int32)  { kitTruthQuantify(out, in, nVars, v, '|') }
func Kit_TruthForall(truth []uint32, nVars, v int32)      { kitTruthQuantify(truth, truth, nVars, v, '&') }
func Kit_TruthForallNew(out, in []uint32, nVars, v int32) { kitTruthQuantify(out, in, nVars, v, '&') }
func Kit_TruthUniqueNew(out, in []uint32, nVars, v int32) { kitTruthQuantify(out, in, nVars, v, '^') }
func Kit_TruthExistSet(out, in []uint32, nVars int32, mask uint32) {
	Kit_TruthCopy(out, in, nVars)
	for v := int32(0); v < nVars; v++ {
		if mask&(1<<uint(v)) != 0 {
			Kit_TruthExist(out, nVars, v)
		}
	}
}
func Kit_TruthForallSet(out, in []uint32, nVars int32, mask uint32) {
	Kit_TruthCopy(out, in, nVars)
	for v := int32(0); v < nVars; v++ {
		if mask&(1<<uint(v)) != 0 {
			Kit_TruthForall(out, nVars, v)
		}
	}
}
func Kit_TruthBooleanDiffCount(truth []uint32, nVars, v int32) int32 {
	out := make([]uint32, Kit_TruthWordNum(nVars))
	Kit_TruthUniqueNew(out, truth, nVars, v)
	return Kit_TruthCountOnes(out, nVars) / 2
}
func Kit_TruthXorCount(a, b []uint32, nVars int32) int32 {
	var count int32
	for i := int32(0); i < Kit_TruthWordNum(nVars); i++ {
		count += int32(bits.OnesCount32(a[i] ^ b[i]))
	}
	return count
}
func Kit_TruthMuxVar(out, low, high []uint32, nVars, v int32) {
	Kit_TruthMuxVarPhase(out, low, high, nVars, v, 0)
}
func Kit_TruthMuxVarPhase(out, low, high []uint32, nVars, v, compl int32) {
	if v < 0 || v >= nVars {
		panic("kit: MUX variable out of range")
	}
	for i := range out[:Kit_TruthWordNum(nVars)] {
		a, b := low[i], high[i]
		if compl != 0 {
			a = ^a
		}
		if v < 5 {
			mask := kitVarMasks[v]
			out[i] = a&^mask | b&mask
		} else if i&(1<<uint(v-5)) == 0 {
			out[i] = a
		} else {
			out[i] = b
		}
	}
}

func Kit_TruthVarsSymm(truth []uint32, nVars, a, b int32, cof0, cof1 []uint32) int32 {
	return kitTruthSymmetry(truth, nVars, a, b, false)
}
func Kit_TruthVarsAntiSymm(truth []uint32, nVars, a, b int32, cof0, cof1 []uint32) int32 {
	return kitTruthSymmetry(truth, nVars, a, b, true)
}
func kitTruthSymmetry(truth []uint32, nVars, a, b int32, inverse bool) int32 {
	if a < 0 || b < 0 || a >= nVars || b >= nVars {
		panic("kit: symmetry variable out of range")
	}
	mask := 1<<uint(a) | 1<<uint(b)
	for x := 0; x < int(Kit_TruthWordNum(nVars))*32; x++ {
		y := x&^mask | (x>>uint(a)&1)<<uint(b) | (x>>uint(b)&1)<<uint(a)
		if inverse {
			y ^= mask
		}
		if Kit_TruthHasBit(truth, int32(x)) != Kit_TruthHasBit(truth, int32(y)) {
			return 0
		}
	}
	return 1
}
func Kit_TruthChangePhase(truth []uint32, nVars, v int32) {
	if v < 0 || v >= nVars {
		panic("kit: phase variable out of range")
	}
	kitRemapTruth(truth, truth, nVars, func(x int) int { return x ^ (1 << uint(v)) })
}
func Kit_TruthMinCofSuppOverlap(truth []uint32, nVars int32, varMin *int32) int32 {
	a, b := make([]uint32, Kit_TruthWordNum(nVars)), make([]uint32, Kit_TruthWordNum(nVars))
	best, index := int32(32), int32(-1)
	for v := int32(0); v < nVars; v++ {
		Kit_TruthCofactor0New(a, truth, nVars, v)
		Kit_TruthCofactor1New(b, truth, nVars, v)
		s0, s1 := Kit_TruthSupport(a, nVars), Kit_TruthSupport(b, nVars)
		overlap := int32(bits.OnesCount32(s0 & s1))
		if overlap < best && bits.OnesCount32(s0) <= 5 && bits.OnesCount32(s1) <= 5 {
			best, index = overlap, v
		}
		if best == 0 {
			break
		}
	}
	if varMin != nil {
		*varMin = index
	}
	return best
}
func Kit_TruthBestCofVar(truth []uint32, nVars int32, a, b []uint32) int32 {
	if Kit_TruthIsConst0(truth, nVars) != 0 || Kit_TruthIsConst1(truth, nVars) != 0 {
		return -1
	}
	index, best := int32(-1), int32(KIT_INFINITY)
	for v := int32(0); v < nVars; v++ {
		Kit_TruthCofactor0New(a, truth, nVars, v)
		Kit_TruthCofactor1New(b, truth, nVars, v)
		size := Kit_TruthSupportSize(a, nVars) + Kit_TruthSupportSize(b, nVars)
		if size < best {
			index, best = v, size
		}
	}
	Kit_TruthCofactor0New(a, truth, nVars, index)
	Kit_TruthCofactor1New(b, truth, nVars, index)
	return index
}
func Kit_TruthCountOnesInCofs(truth []uint32, nVars int32, store []int32) {
	clear(store[:2*nVars])
	for x := int32(0); x < 32*Kit_TruthWordNum(nVars); x++ {
		if Kit_TruthHasBit(truth, x) == 0 {
			continue
		}
		for v := int32(0); v < nVars; v++ {
			store[2*v+(x>>uint(v)&1)]++
		}
	}
}
func Kit_TruthCountOnesInCofs0(truth []uint32, nVars int32, store []int32) {
	counts := make([]int32, 2*nVars)
	Kit_TruthCountOnesInCofs(truth, nVars, counts)
	for v := int32(0); v < nVars; v++ {
		store[v] = counts[2*v]
	}
}
func Kit_TruthCountOnesInCofsSlow(truth []uint32, nVars int32, store []int32, aux []uint32) {
	Kit_TruthCountOnesInCofs(truth, nVars, store)
}

var kitHashPrimes = func() (primes [1024]uint32) {
	count := 0
	for candidate := uint32(2); count < len(primes); candidate++ {
		prime := true
		for divisor := uint32(2); divisor*divisor <= candidate; divisor++ {
			if candidate%divisor == 0 {
				prime = false
				break
			}
		}
		if prime {
			primes[count] = candidate
			count++
		}
	}
	return
}()

func Kit_TruthHash(truth []uint32, nWords int32) uint32 {
	if nWords < 0 || nWords > 1024 {
		panic("kit: unsupported truth hash size")
	}
	var hash uint32
	for i := int32(0); i < nWords; i++ {
		hash ^= kitHashPrimes[i] * truth[i]
	}
	return hash
}
func Kit_TruthSemiCanonicize(truth, aux []uint32, nVars int32, permutation []byte) uint32 {
	if len(permutation) < int(nVars) {
		panic("kit: short canonical permutation")
	}
	for i := int32(0); i < nVars; i++ {
		permutation[i] = byte(i)
	}
	var phase uint32
	if Kit_TruthCountOnes(truth, nVars) > Kit_TruthWordNum(nVars)*16 {
		phase |= 1 << uint(nVars)
		Kit_TruthNot(truth, truth, nVars)
	}
	counts := make([]int32, 2*nVars)
	Kit_TruthCountOnesInCofs(truth, nVars, counts)
	for i := int32(0); i < nVars; i++ {
		if counts[2*i] < counts[2*i+1] {
			phase |= 1 << uint(i)
			counts[2*i], counts[2*i+1] = counts[2*i+1], counts[2*i]
			Kit_TruthChangePhase(truth, nVars, i)
		}
	}
	for changed := true; changed; {
		changed = false
		for i := int32(0); i < nVars-1; i++ {
			if counts[2*i] >= counts[2*(i+1)] {
				continue
			}
			changed = true
			permutation[i], permutation[i+1] = permutation[i+1], permutation[i]
			counts[2*i], counts[2*(i+1)] = counts[2*(i+1)], counts[2*i]
			counts[2*i+1], counts[2*(i+1)+1] = counts[2*(i+1)+1], counts[2*i+1]
			if (phase >> uint(i) & 1) != (phase >> uint(i+1) & 1) {
				phase ^= 1<<uint(i) | 1<<uint(i+1)
			}
			Kit_TruthSwapAdjacentVars(aux, truth, nVars, i)
			Kit_TruthCopy(truth, aux, nVars)
		}
	}
	return phase
}
func Kit_TruthCountMinterms(truth []uint32, nVars int32, result, bytes []int32) int32 {
	Kit_TruthCountOnesInCofs0(truth, nVars, result)
	return Kit_TruthCountOnes(truth, nVars)
}
func Kit_PrintHexadecimal(w io.Writer, truth []uint32, nVars int32) {
	digits := (1<<uint(nVars) + 3) / 4
	for i := int(digits) - 1; i >= 0; i-- {
		fmt.Fprintf(w, "%x", truth[i/8]>>uint(4*(i%8))&15)
	}
}
func Kit_TruthCountMintermsPrecomp() {
	for x := 0; x < 256; x++ {
		value := uint32(bits.OnesCount8(uint8(x)))
		for v := 0; v < 3; v++ {
			var count uint32
			for bit := 0; bit < 8; bit++ {
				if bit&(1<<uint(v)) == 0 && x&(1<<uint(bit)) != 0 {
					count++
				}
			}
			value |= count << uint(8*(v+1))
		}
		fmt.Printf("0x%08x, ", value)
		if x%8 == 7 {
			fmt.Println()
		}
	}
}
func Kit_TruthDumpToFile(truth []uint32, nVars, file int32) string {
	name := filepath.Join("tt", fmt.Sprintf("s%04d", file))
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		panic(err)
	}
	out, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	fmt.Fprint(out, "rt ")
	Kit_PrintHexadecimal(out, truth, nVars)
	fmt.Fprintln(out, "; bdd; sop; ps")
	return name
}
func Kit_TruthPrintProfile_int(truth []uint32, nVars int32) {
	if nVars < 6 || nVars > 20 {
		panic("kit: truth profiling supports 6..20 variables")
	}
	counts := make([]int32, 2*nVars)
	Kit_TruthCountOnesInCofs(truth, nVars, counts)
	unique, common, total := make([][]int64, nVars), make([][]int64, nVars), make([][]int64, nVars)
	for i := range unique {
		unique[i] = make([]int64, nVars)
		common[i] = make([]int64, nVars)
		total[i] = make([]int64, nVars)
		for j := range unique[i] {
			unique[i][j] = -1
			common[i][j] = -1
			total[i][j] = -1
		}
	}
	words := Kit_TruthWordNum(nVars)
	for v := int32(0); v < nVars; v++ {
		for u := v + 1; u < nVars; u++ {
			cofactors := [4][]uint32{}
			ones := [4]int64{}
			zeros := [4]int64{}
			for i := range cofactors {
				cofactors[i] = make([]uint32, words)
				kitTruthCofactor(cofactors[i], truth, nVars, v, i>>1)
				kitTruthCofactor(cofactors[i], cofactors[i], nVars, u, i&1)
				ones[i] = int64(Kit_TruthCountOnes(cofactors[i], nVars)) / 4
				zeros[i] = (int64(1) << uint(nVars-2)) - ones[i]
			}
			var uniqueCount, totalCount int64
			for i := 0; i < 4; i++ {
				for j := i + 1; j < 4; j++ {
					uniqueCount += int64(Kit_TruthXorCount(cofactors[i], cofactors[j], nVars)) / 2
					totalCount += ones[i]*zeros[j] + zeros[i]*ones[j]
				}
			}
			commonCount := ones[0]*zeros[3] + zeros[0]*ones[3] + ones[1]*zeros[2] + zeros[1]*ones[2]
			unique[v][u], unique[u][v] = uniqueCount, uniqueCount
			common[v][u], common[u][v] = commonCount, commonCount
			total[v][u], total[u][v] = totalCount, totalCount
		}
	}
	fmt.Print("\n V: ")
	for v := int32(0); v < nVars; v++ {
		fmt.Printf("%8c  ", 'a'+v)
	}
	fmt.Print("\n M: ")
	for v := int32(0); v < nVars; v++ {
		fmt.Printf("%8d  ", counts[2*v])
	}
	fmt.Print("\n U: ")
	for v := int32(0); v < nVars; v++ {
		fmt.Printf("%8d  ", Kit_TruthBooleanDiffCount(truth, nVars, v))
	}
	fmt.Print("\n\n")
	for _, matrix := range []struct {
		name   string
		values [][]int64
	}{{"Unique", unique}, {"Common", common}, {"Total", total}} {
		fmt.Println(matrix.name + ":")
		for i, row := range matrix.values {
			fmt.Printf(" %2d ", i)
			for _, value := range row {
				fmt.Printf("%8d  ", value)
			}
			fmt.Println()
		}
	}
}
func Kit_TruthPrintProfile(truth []uint32, nVars int32) {
	if nVars < 2 {
		panic("kit: truth profiling requires at least two variables")
	}
	if nVars < 6 {
		Kit_TruthPrintProfile_int([]uint32{truth[0], truth[0]}, 6)
		return
	}
	Kit_TruthPrintProfile_int(truth, nVars)
}
