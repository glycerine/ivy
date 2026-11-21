package util

import "sync/atomic"

const isopByteRepeat word = 0x0101010101010101

// Each of the eight assignments is off, unspecified or on. Adding the
// base-three expansions of its on-set and allowed-set masks gives its key.
var isop3MaskIndex = func() (index [256]uint16) {
	for mask := 1; mask < len(index); mask++ {
		index[mask] = 3*index[mask>>1] + uint16(mask&1)
	}
	return index
}()

var isop3Counts [6561]atomic.Uint32

func isop3Count(on, dc byte, count *int32) byte {
	if on == 0 {
		return 0
	}
	if dc == 0xff {
		(*count)++
		return 0xff
	}
	slot := &isop3Counts[isop3MaskIndex[on]+isop3MaskIndex[dc]]
	entry := slot.Load()
	if entry == 0 {
		entry = isop3Fill(on, dc)
		slot.Store(entry)
	}
	*count += int32(entry >> 8)
	return byte(entry)
}

func isop3Fill(on, dc byte) uint32 {
	// Each cube covers a private assignment, so eight slots suffice.
	var cover [8]int32
	var cubes int32
	truth := ttIsop6(word(on)*isopByteRepeat, word(dc)*isopByteRepeat, 3, cover[:], &cubes)
	return uint32(byte(truth)) | uint32(cubes)<<8
}

const isopHalfRepeat word = 0x0001000100010001

// Exact four-input functions always return their input truth. Each count fits
// in a byte, and four independent functions share one atomic word.
var isop4ExactCounts [1 << 14]atomic.Uint32

func isop4Exact(on uint16, count *int32) uint16 {
	slot := &isop4ExactCounts[on>>2]
	shift := (on & 3) * 8
	cubes := int32(byte(slot.Load() >> shift))
	if cubes == 0 {
		isop4Split(on, on, &cubes)
		slot.Or(uint32(cubes) << shift)
	}
	*count += cubes
	return on
}

func isop4Split(on, dc uint16, count *int32) uint16 {
	a, b := byte(on), byte(on>>8)
	da, db := byte(dc), byte(dc>>8)
	// If one allowed cofactor is empty, every surviving cube uses this variable.
	if da == 0 {
		return uint16(isop3Count(b, db, count)) << 8
	}
	if db == 0 {
		return uint16(isop3Count(a, da, count))
	}
	r0 := isop3Count(a&^db, da, count)
	r1 := isop3Count(b&^da, db, count)
	r2 := isop3Count((a&^r0)|(b&^r1), da&db, count)
	return uint16(r2|r0) | uint16(r2|r1)<<8
}

func isop4Count(on, dc uint16, count *int32) uint16 {
	if on == 0 {
		return 0
	}
	if dc == 0xffff {
		(*count)++
		return 0xffff
	}
	if on == dc {
		return isop4Exact(on, count)
	}
	return isop4Split(on, dc, count)
}

const isop32Repeat word = 0x0000000100000001

// Count-only cofactors fit in sixteen bits for a five-variable function.
// When both sets repeat, the upper variable contributes no literals.
func isop5Count(on, dc uint32, count *int32) uint32 {
	if on == 0 {
		return 0
	}
	if dc == 0xffffffff {
		(*count)++
		return 0xffffffff
	}
	a, b := uint16(on), uint16(on>>16)
	da, db := uint16(dc), uint16(dc>>16)
	if a == b && da == db {
		return uint32(isop4Count(a, da, count)) * 0x00010001
	}
	if da == 0 {
		return uint32(isop4Count(b, db, count)) << 16
	}
	if db == 0 {
		return uint32(isop4Count(a, da, count))
	}
	r0 := isop4Count(a&^db, da, count)
	r1 := isop4Count(b&^da, db, count)
	r2 := isop4Count((a&^r0)|(b&^r1), da&db, count)
	return uint32(r2|r0) | uint32(r2|r1)<<16
}

func isop6Count(on, dc word, count *int32) word {
	a, b := uint32(on), uint32(on>>32)
	da, db := uint32(dc), uint32(dc>>32)
	if a == b && da == db {
		return word(isop5Count(a, da, count)) * isop32Repeat
	}
	if da == 0 {
		return word(isop5Count(b, db, count)) << 32
	}
	if db == 0 {
		return word(isop5Count(a, da, count))
	}
	r0 := isop5Count(a&^db, da, count)
	r1 := isop5Count(b&^da, db, count)
	r2 := isop5Count((a&^r0)|(b&^r1), da&db, count)
	return word(r2|r0) | word(r2|r1)<<32
}

// ISOP follows the original three-way recursive split: cubes forced into each
// cofactor, followed by cubes shared by both cofactors. Literal codes use two
// bits per variable: 1 for the negative literal, 2 for the positive literal.
func ttIsop6(on, dc word, n int32, cover []int32, count *int32) word {
	if n < 0 || n > 6 || on&^dc != 0 {
		panic("truth: invalid ISOP bounds")
	}
	if on == 0 {
		return 0
	}
	if dc == ^word(0) {
		if cover != nil {
			cover[*count] = 0
		}
		(*count)++
		return ^word(0)
	}
	if cover == nil {
		switch n {
		case 3:
			if word(byte(on))*isopByteRepeat == on && word(byte(dc))*isopByteRepeat == dc {
				return word(isop3Count(byte(on), byte(dc), count)) * isopByteRepeat
			}
		case 4:
			if word(uint16(on))*isopHalfRepeat == on && word(uint16(dc))*isopHalfRepeat == dc {
				return word(isop4Count(uint16(on), uint16(dc), count)) * isopHalfRepeat
			}
		case 5:
			if word(uint32(on))*isop32Repeat == on && word(uint32(dc))*isop32Repeat == dc {
				return word(isop5Count(uint32(on), uint32(dc), count)) * isop32Repeat
			}
		case 6:
			return isop6Count(on, dc, count)
		}
	}
	v := n - 1
	for ; v >= 0; v-- {
		if Abc_Tt6HasVar(on, v) != 0 || Abc_Tt6HasVar(dc, v) != 0 {
			break
		}
	}
	if v < 0 {
		panic("truth: ISOP input not stretched to six variables")
	}
	a, b := Abc_Tt6Cofactor0(on, v), Abc_Tt6Cofactor1(on, v)
	da, db := Abc_Tt6Cofactor0(dc, v), Abc_Tt6Cofactor1(dc, v)
	begin := *count
	r0 := ttIsop6(a&^db, da, v, cover, count)
	end0 := *count
	r1 := ttIsop6(b&^da, db, v, cover, count)
	end1 := *count
	r2 := ttIsop6((a&^r0)|(b&^r1), da&db, v, cover, count)
	markIsopCofactors(cover, begin, end0, end1, v)
	return r2 | (r0 &^ s_Truths6[v]) | (r1 & s_Truths6[v])
}
func Abc_Tt6Isop(on, dc word, n int32, count *int32) word { return ttIsop6(on, dc, n, nil, count) }
func Abc_Tt6IsopCover(on, dc word, n int32, cover []int32, count *int32) word {
	return ttIsop6(on, dc, n, cover, count)
}

func markIsopCofactors(cover []int32, begin, end0, end1, variable int32) {
	if cover == nil {
		return
	}
	for i := begin; i < end0; i++ {
		cover[i] |= 1 << uint(2*variable)
	}
	for i := end0; i < end1; i++ {
		cover[i] |= 1 << uint(2*variable+1)
	}
}

// The two halves of a seven-input function each fit in one word.
func ttIsop7(on, dc [2]word, cover []int32, count *int32) [2]word {
	if on[0] == on[1] && dc[0] == dc[1] {
		result := ttIsop6(on[0], dc[0], 6, cover, count)
		return [2]word{result, result}
	}
	begin := *count
	lo := ttIsop6(on[0]&^dc[1], dc[0], 6, cover, count)
	end0 := *count
	hi := ttIsop6(on[1]&^dc[0], dc[1], 6, cover, count)
	end1 := *count
	both := ttIsop6((on[0]&^lo)|(on[1]&^hi), dc[0]&dc[1], 6, cover, count)
	markIsopCofactors(cover, begin, end0, end1, 6)
	return [2]word{lo | both, hi | both}
}

func ttIsopWide(on, dc []word, n int32, out []word, cover []int32, count *int32) {
	if n < 0 || n > 8 {
		panic("truth: ISOP supports at most eight variables")
	}
	if n <= 6 {
		result := ttIsop6(on[0], dc[0], n, cover, count)
		for i := range out {
			out[i] = result
		}
		return
	}
	lo, dcLo := [2]word(on), [2]word(dc)
	if n == 7 || (lo == [2]word(on[2:]) && dcLo == [2]word(dc[2:])) {
		result := ttIsop7(lo, dcLo, cover, count)
		out[0], out[1] = result[0], result[1]
		if len(out) == 4 {
			out[2], out[3] = result[0], result[1]
		}
		return
	}
	hi, dcHi := [2]word(on[2:]), [2]word(dc[2:])
	begin := *count
	r0 := ttIsop7([2]word{lo[0] &^ dcHi[0], lo[1] &^ dcHi[1]}, dcLo, cover, count)
	end0 := *count
	r1 := ttIsop7([2]word{hi[0] &^ dcLo[0], hi[1] &^ dcLo[1]}, dcHi, cover, count)
	end1 := *count
	shared := [2]word{(lo[0] &^ r0[0]) | (hi[0] &^ r1[0]), (lo[1] &^ r0[1]) | (hi[1] &^ r1[1])}
	dcShared := [2]word{dcLo[0] & dcHi[0], dcLo[1] & dcHi[1]}
	both := ttIsop7(shared, dcShared, cover, count)
	out[0], out[1] = r0[0]|both[0], r0[1]|both[1]
	out[2], out[3] = r1[0]|both[0], r1[1]|both[1]
	markIsopCofactors(cover, begin, end0, end1, 7)
}
func Abc_Tt7Isop(on, dc []word, n int32, out []word) int32 {
	if n > 7 {
		panic("truth: seven-variable ISOP")
	}
	count := int32(0)
	ttIsopWide(on, dc, n, out[:2], nil, &count)
	return count
}
func Abc_Tt8Isop(on, dc []word, n int32, out []word) int32 {
	count := int32(0)
	ttIsopWide(on, dc, n, out[:4], nil, &count)
	return count
}
func Abc_Tt7IsopCover(on, dc []word, n int32, out []word, cover []int32, count *int32) {
	if n > 7 {
		panic("truth: seven-variable ISOP")
	}
	ttIsopWide(on, dc, n, out[:2], cover, count)
}
func Abc_Tt8IsopCover(on, dc []word, n int32, out []word, cover []int32, count *int32) {
	ttIsopWide(on, dc, n, out[:4], cover, count)
}
func Abc_Tt6CnfSize(t word, n int32) int32 {
	count := int32(0)
	Abc_Tt6Isop(t, t, n, &count)
	Abc_Tt6Isop(^t, ^t, n, &count)
	return count
}
func Abc_Tt8CnfSize(t []word, n int32) int32 {
	tc := []word{^t[0], ^t[1], ^t[2], ^t[3]}
	out := make([]word, 4)
	return Abc_Tt8Isop(t, t, n, out) + Abc_Tt8Isop(tc, tc, n, out)
}
func Abc_Tt6Cnf(t word, n int32, cover []int32) int32 {
	count := int32(0)
	Abc_Tt6IsopCover(t, t, n, cover, &count)
	end := count
	for i := int32(0); i < end; i++ {
		cover[i] |= 1 << uint(2*n)
	}
	Abc_Tt6IsopCover(^t, ^t, n, cover, &count)
	for i := end; i < count; i++ {
		cover[i] |= 1 << uint(2*n+1)
	}
	return count
}
func Abc_Tt8Cnf(t []word, n int32, cover []int32) int32 {
	count := int32(0)
	out := make([]word, 4)
	Abc_Tt8IsopCover(t, t, n, out, cover, &count)
	end := count
	for i := int32(0); i < end; i++ {
		cover[i] |= 1 << uint(2*n)
	}
	tc := []word{^t[0], ^t[1], ^t[2], ^t[3]}
	Abc_Tt8IsopCover(tc, tc, n, out, cover, &count)
	for i := end; i < count; i++ {
		cover[i] |= 1 << uint(2*n+1)
	}
	return count
}

// ESOP chooses the least expensive Shannon/Davio split, retaining the C
// algorithm's tie order (discard r0, then r1, then r2).
func ttEsop(t word, n int32) []int32 {
	if n < 0 || n > 6 {
		panic("truth: ESOP supports at most six variables")
	}
	if t == 0 {
		return nil
	}
	if t == ^word(0) {
		return []int32{0}
	}
	v := n - 1
	for ; v >= 0; v-- {
		if Abc_Tt6HasVar(t, v) != 0 {
			break
		}
	}
	if v < 0 {
		panic("truth: ESOP input not stretched")
	}
	a, b := Abc_Tt6Cofactor0(t, v), Abc_Tt6Cofactor1(t, v)
	r0, r1, r2 := ttEsop(a, v), ttEsop(b, v), ttEsop(a^b, v)
	switch max(len(r0), len(r1), len(r2)) {
	case len(r0):
		for i := range r2 {
			r2[i] |= 1 << uint(2*v)
		}
		return append(r1, r2...)
	case len(r1):
		for i := range r2 {
			r2[i] |= 1 << uint(2*v+1)
		}
		return append(r0, r2...)
	default:
		for i := range r0 {
			r0[i] |= 1 << uint(2*v)
		}
		for i := range r1 {
			r1[i] |= 1 << uint(2*v+1)
		}
		return append(r0, r1...)
	}
}
func Abc_Tt6Esop(t word, n int32, cover []int32) int32 {
	cubes := ttEsop(t, n)
	if cover != nil {
		if len(cover) < len(cubes) {
			panic("truth: ESOP cover buffer too short")
		}
		copy(cover, cubes)
	}
	return int32(len(cubes))
}
func Abc_Tt6EsopBuild(n int32, cover []int32, count int32) word {
	var result word
	for _, cube := range cover[:count] {
		p := ^word(0)
		for v := int32(0); v < n; v++ {
			switch cube >> uint(2*v) & 3 {
			case 1:
				p &^= s_Truths6[v]
			case 2:
				p &= s_Truths6[v]
			}
		}
		result ^= p
	}
	return result
}
func Abc_Tt6EsopVerify(t word, n int32) int32 {
	cover := make([]int32, 64)
	count := Abc_Tt6Esop(t, n, cover)
	if Abc_Tt6EsopBuild(n, cover, count) != t {
		panic("truth: ESOP verification failed")
	}
	return count
}

func Abc_Tt6CheckOutDec(t word, v int32, out []word) int32 {
	a, b := Abc_Tt6Cofactor0(t, v), Abc_Tt6Cofactor1(t, v)
	if a == b {
		panic("truth: decomposition variable is absent")
	}
	var result word
	kind := int32(-1)
	switch {
	case a == 0:
		result, kind = b, 0
	case b == 0:
		result, kind = a, 1
	case a == ^word(0):
		result, kind = b, 2
	case b == ^word(0):
		result, kind = a, 3
	case a == ^b:
		result, kind = a, 4
	}
	if kind >= 0 && out != nil {
		out[0] = result
	}
	return kind
}
func ttOutCof(t []word, n, v int32, one bool) []word {
	size := int32(1)
	if n > 6 {
		size = 1 << uint(n-6)
	}
	out := make([]word, size)
	for bit := 0; bit < int(size)*64; bit++ {
		source := bit &^ (1 << uint(v))
		if one {
			source |= 1 << uint(v)
		}
		if t[source>>6]&(word(1)<<uint(source&63)) != 0 {
			out[bit>>6] |= word(1) << uint(bit&63)
		}
	}
	return out
}
func Abc_TtCheckOutDec(t []word, n, v int32, out []word) int32 {
	if n > 8 || v < 0 || v >= n {
		panic("truth: invalid decomposition dimensions")
	}
	a, b := ttOutCof(t, n, v, false), ttOutCof(t, n, v, true)
	equal, opposite, zeroA, zeroB, fullA, fullB := true, true, true, true, true, true
	for i := range a {
		equal = equal && a[i] == b[i]
		opposite = opposite && a[i] == ^b[i]
		zeroA = zeroA && a[i] == 0
		zeroB = zeroB && b[i] == 0
		fullA = fullA && a[i] == ^word(0)
		fullB = fullB && b[i] == ^word(0)
	}
	if equal {
		panic("truth: decomposition variable is absent")
	}
	kind := int32(-1)
	var result []word
	switch {
	case zeroA:
		kind, result = 0, b
	case zeroB:
		kind, result = 1, a
	case fullA:
		kind, result = 2, b
	case fullB:
		kind, result = 3, a
	case opposite:
		kind, result = 4, a
	}
	if kind >= 0 && out != nil {
		copy(out, result)
	}
	return kind
}
func Abc_TtCheckDecOutOne7(t []word, variable, kind *int32) word {
	out := make([]word, 2)
	for v := int32(6); v >= 0; v-- {
		k := Abc_TtCheckOutDec(t, 7, v, nil)
		if k >= 0 {
			Abc_TtSwapVars(t, 7, 6, v)
			Abc_TtCheckOutDec(t, 7, 6, out)
			*variable, *kind = v, k
			return out[0]
		}
	}
	return 0
}
func Abc_TtCheckDecOutOne8(t []word, var1, var2, kind1, kind2 *int32) word {
	out := make([]word, 4)
	found := false
	for v := int32(7); v >= 0; v-- {
		k := Abc_TtCheckOutDec(t, 8, v, nil)
		if k >= 0 {
			Abc_TtSwapVars(t, 8, 7, v)
			Abc_TtCheckOutDec(t, 8, 7, out)
			*var1, *kind1 = v, k
			found = true
			break
		}
	}
	if !found {
		return 0
	}
	for v := int32(6); v >= 0; v-- {
		k := Abc_TtCheckOutDec(out, 7, v, nil)
		if k >= 0 && k>>1 == *kind1>>1 {
			Abc_TtSwapVars(out, 7, 6, v)
			out2 := make([]word, 2)
			Abc_TtCheckOutDec(out, 7, 6, out2)
			*var2, *kind2 = v, k
			return out2[0]
		}
	}
	return 0
}
func Abc_TtCheckDsdAnd(t word, i, j int32, out []word) int32 {
	a, b := Abc_Tt6Cofactor0(t, i), Abc_Tt6Cofactor1(t, i)
	a0, a1, b0, b1 := Abc_Tt6Cofactor0(a, j), Abc_Tt6Cofactor1(a, j), Abc_Tt6Cofactor0(b, j), Abc_Tt6Cofactor1(b, j)
	var low, high word
	kind := int32(-1)
	switch {
	case a0 == a1 && a0 == b0:
		low, high, kind = a0, b1, 0
	case b1 == a0 && b1 == b0:
		low, high, kind = b1, a1, 1
	case b1 == a0 && b1 == a1:
		low, high, kind = b1, b0, 2
	case b1 == a1 && b1 == b0:
		low, high, kind = b1, a0, 3
	case a0 == b1 && a1 == b0:
		low, high, kind = b1, b0, 4
	}
	if kind >= 0 && out != nil {
		out[0] = (low &^ s_Truths6[i]) | (high & s_Truths6[i])
	}
	return kind
}
func Abc_TtCheckDsdMux(t word, i int32, out []word) int32 {
	a, b := Abc_Tt6Cofactor0(t, i), Abc_Tt6Cofactor1(t, i)
	v0, v1 := int32(-1), int32(-1)
	for k := int32(0); k < 6; k++ {
		if k == i {
			continue
		}
		has0, has1 := Abc_Tt6HasVar(a, k) != 0, Abc_Tt6HasVar(b, k) != 0
		if has0 && !has1 {
			if v0 >= 0 {
				return -1
			}
			v0 = k
		}
		if !has0 && has1 {
			if v1 >= 0 {
				return -1
			}
			v1 = k
		}
	}
	if v0 < 0 || v1 < 0 {
		return -1
	}
	a0, a1 := Abc_Tt6Cofactor0(a, v0), Abc_Tt6Cofactor1(a, v0)
	b0, b1 := Abc_Tt6Cofactor0(b, v1), Abc_Tt6Cofactor1(b, v1)
	if a0 == b0 && a1 == b1 {
		if out != nil {
			out[0] = (b0 &^ s_Truths6[i]) | (b1 & s_Truths6[i])
		}
		return (2*v1)<<16 | 2*v0
	}
	if a0 == ^b0 && a1 == ^b1 {
		if out != nil {
			out[0] = (b0 &^ s_Truths6[i]) | (b1 & s_Truths6[i])
		}
		return (2*v1)<<16 | (2*v0 + 1)
	}
	return -1
}
