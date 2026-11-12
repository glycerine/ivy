package dau

import "math/bits"

var dauTruthVariables = [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}

func Abc_TtNormalizeSmallTruth(truth []uint64, nVars int32) {
	if nVars < 0 {
		panic("negative DAU variable count")
	}
	if nVars >= 6 {
		return
	}
	size := uint(1) << uint(nVars)
	base := truth[0] & ((uint64(1) << size) - 1)
	truth[0] = base
	for s := size; s < 64; s += size {
		truth[0] |= base << s
	}
}
func Abc_TtVerifySmallTruth(truth []uint64, nVars int32) {
	if nVars < 6 {
		copy := []uint64{truth[0]}
		Abc_TtNormalizeSmallTruth(copy, nVars)
		if copy[0] != truth[0] {
			panic("DAU truth is not normalized")
		}
	}
}
func Abc_TtCountOnesInTruth(truth []uint64, nVars int32) int32 {
	Abc_TtVerifySmallTruth(truth, nVars)
	total := 0
	for _, w := range truth[:dauCountWords(nVars)] {
		total += bits.OnesCount64(w)
	}
	return int32(total)
}
func Abc_TtCountOnesInCofs(truth []uint64, nVars int32, store []int32) {
	if nVars < 0 || len(store) < int(nVars) {
		panic("short DAU cofactor result")
	}
	clear(store[:nVars])
	for block, w := range truth[:dauCountWords(nVars)] {
		for v := int32(0); v < nVars; v++ {
			if v < 6 {
				store[v] += int32(bits.OnesCount64(w &^ dauTruthVariables[v]))
			} else if block&(1<<uint(v-6)) == 0 {
				store[v] += int32(bits.OnesCount64(w))
			}
		}
	}
}
func Abc_TtCountOnesInCofsSimple(truth []uint64, nVars int32, store []int32) int32 {
	Abc_TtCountOnesInCofs(truth, nVars, store)
	return Abc_TtCountOnesInTruth(truth, nVars)
}
func Abc_TtCountOnesInCofsSlow(truth []uint64, nVars int32, store []int32) {
	Abc_TtCountOnesInCofs(truth, nVars, store)
}

// Shifted cofactor coefficients weigh an asserted minterm by 2^popcount.
func shiftFunc(ci int32) int32 { return 1 << uint(ci) }
func Abc_TtScc6(truth uint64, ck int32) int32 {
	sum := int32(0)
	for truth != 0 {
		i := bits.TrailingZeros64(truth)
		sum += shiftFunc(int32(bits.OnesCount8(uint8(i))) + ck)
		truth &= truth - 1
	}
	return sum
}
func Abc_TtScc(truth []uint64, nVars int32) int32 {
	Abc_TtNormalizeSmallTruth(truth, nVars)
	sum := int32(0)
	for i, w := range truth[:dauCountWords(nVars)] {
		sum += Abc_TtScc6(w, int32(bits.OnesCount(uint(i))))
	}
	return sum
}
func Abc_TtSccInCofs6(truth uint64, nVars, ck int32, store []int32) {
	if nVars > 6 {
		panic("DAU single-word SCC width exceeds six")
	}
	for v := int32(0); v < nVars; v++ {
		for i := 0; i < 64; i++ {
			if i&(1<<uint(v)) != 0 || truth&(1<<uint(i)) == 0 {
				continue
			}
			compressed := (i & ((1 << uint(v)) - 1)) | (i>>uint(v+1))<<uint(v)
			store[v] += shiftFunc(int32(bits.OnesCount(uint(compressed))) + ck)
		}
	}
}
func Abc_TtSccInCofs(truth []uint64, nVars int32, store []int32) {
	clear(store[:nVars])
	if nVars <= 6 {
		Abc_TtNormalizeSmallTruth(truth, nVars)
		Abc_TtSccInCofs6(truth[0], nVars, 0, store)
		return
	}
	for block, w := range truth[:dauCountWords(nVars)] {
		Abc_TtSccInCofs6(w, 6, int32(bits.OnesCount(uint(block))), store)
		for v := int32(6); v < nVars; v++ {
			if block&(1<<uint(v-6)) != 0 {
				continue
			}
			b := v - 6
			compressed := (block & ((1 << uint(b)) - 1)) | (block>>uint(b+1))<<uint(b)
			store[v] += Abc_TtScc6(w, int32(bits.OnesCount(uint(compressed))))
		}
	}
}

func dauCofactorWord(t uint64, v int32, one bool) uint64 {
	mask := dauTruthVariables[v]
	shift := uint(1) << uint(v)
	if one {
		x := t & mask
		return x | x>>shift
	}
	x := t &^ mask
	return x | x<<shift
}
func Abc_TtCountOnesInCofsFast6_rec(truth uint64, iVar, nBytes int32, store []int32) int32 {
	if truth == 0 {
		return 0
	}
	if truth == ^uint64(0) {
		for i := int32(0); i <= iVar; i++ {
			store[i] += nBytes * 4
		}
		return nBytes * 8
	}
	if nBytes == 1 {
		if iVar != 2 {
			panic("invalid DAU fast byte cofactor index")
		}
		store[0] += int32(bits.OnesCount8(byte(truth) & 0x55))
		store[1] += int32(bits.OnesCount8(byte(truth) & 0x33))
		store[2] += int32(bits.OnesCount8(byte(truth) & 0x0f))
		return int32(bits.OnesCount8(byte(truth)))
	}
	a := Abc_TtCountOnesInCofsFast6_rec(dauCofactorWord(truth, iVar, false), iVar-1, nBytes/2, store)
	b := Abc_TtCountOnesInCofsFast6_rec(dauCofactorWord(truth, iVar, true), iVar-1, nBytes/2, store)
	store[iVar] += a
	return a + b
}
func Abc_TtCountOnesInCofsFast_rec(truth []uint64, iVar, nWords int32, store []int32) int32 {
	if nWords == 1 {
		return Abc_TtCountOnesInCofsFast6_rec(truth[0], iVar, 8, store)
	}
	allZero, allOne := true, true
	for _, w := range truth[:nWords] {
		allZero = allZero && w == 0
		allOne = allOne && w == ^uint64(0)
	}
	if allZero {
		return 0
	}
	if allOne {
		for i := int32(0); i <= iVar; i++ {
			store[i] += nWords * 32
		}
		return nWords * 64
	}
	a := Abc_TtCountOnesInCofsFast_rec(truth[:nWords/2], iVar-1, nWords/2, store)
	b := Abc_TtCountOnesInCofsFast_rec(truth[nWords/2:], iVar-1, nWords/2, store)
	store[iVar] += a
	return a + b
}
func Abc_TtCountOnesInCofsFast(truth []uint64, nVars int32, store []int32) int32 {
	if nVars < 3 || nVars > 30 || len(store) < int(nVars) || len(truth) < dauCountWords(nVars) {
		panic("invalid DAU fast cofactor dimensions")
	}
	clear(store[:nVars])
	if nVars <= 6 {
		return Abc_TtCountOnesInCofsFast6_rec(truth[0], nVars-1, 1<<uint(nVars-3), store)
	}
	return Abc_TtCountOnesInCofsFast_rec(truth, nVars-1, int32(dauCountWords(nVars)), store)
}
