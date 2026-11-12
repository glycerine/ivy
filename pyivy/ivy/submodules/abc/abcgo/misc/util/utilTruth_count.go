package util

import "math/bits"

// Population counts and bit searches preserve ABC's 64-bit truth-table layout.
func Abc_TtCountOnesSlow(t uint64) int32 { return int32(bits.OnesCount64(t)) }
func Abc_TtCountOnes(t uint64) int32     { return int32(bits.OnesCount64(t)) }
func Abc_TtCountOnesVec(x []uint64, n int32) int32 {
	var c int32
	for _, v := range x[:n] {
		c += Abc_TtCountOnes(v)
	}
	return c
}
func Abc_TtCountOnesVecMask(x, mask []uint64, n, compl int32) int32 {
	var c int32
	for i := int32(0); i < n; i++ {
		v := x[i]
		if compl != 0 {
			v = ^v
		}
		c += Abc_TtCountOnes(v & mask[i])
	}
	return c
}
func Abc_TtCountOnesVecMask2(x, y []uint64, cx, cy int32, mask []uint64, n int32) int32 {
	var c int32
	for i := int32(0); i < n; i++ {
		a, b := x[i], y[i]
		if cx != 0 {
			a = ^a
		}
		if cy != 0 {
			b = ^b
		}
		c += Abc_TtCountOnes(mask[i] & a & b)
	}
	return c
}
func Abc_TtCountOnesVecXor(x, y []uint64, n int32) int32 {
	var c int32
	for i := int32(0); i < n; i++ {
		c += Abc_TtCountOnes(x[i] ^ y[i])
	}
	return c
}
func Abc_TtCountOnesVecXorMask(x, y []uint64, compl int32, mask []uint64, n int32) int32 {
	var c int32
	for i := int32(0); i < n; i++ {
		v := x[i] ^ y[i]
		if compl != 0 {
			v = ^v
		}
		c += Abc_TtCountOnes(mask[i] & v)
	}
	return c
}
func Abc_TtAndXorSum(out, x, y []uint64, n int32) int32 {
	var c int32
	for i := int32(0); i < n; i++ {
		out[i] &= x[i] ^ y[i]
		c += Abc_TtCountOnes(out[i])
	}
	return c
}
func Abc_Tt6FirstBit(t uint64) int32 {
	if t == 0 {
		return -1
	}
	return int32(bits.TrailingZeros64(t))
}
func Abc_Tt6LastBit(t uint64) int32 {
	if t == 0 {
		return -1
	}
	return int32(bits.Len64(t) - 1)
}
func truthCountWords(n int32) int32 {
	if n < 0 || n > 30 {
		panic("invalid truth variable count")
	}
	if n <= 6 {
		return 1
	}
	return 1 << uint(n-6)
}
func Abc_TtFindFirstBit(x []uint64, nVars int32) int32 {
	return Abc_TtFindFirstBit2(x, truthCountWords(nVars))
}
func Abc_TtFindFirstBit2(x []uint64, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if x[i] != 0 {
			return 64*i + Abc_Tt6FirstBit(x[i])
		}
	}
	return -1
}
func Abc_TtFindLastBit(x []uint64, nVars int32) int32 {
	return Abc_TtFindLastBit2(x, truthCountWords(nVars))
}
func Abc_TtFindLastBit2(x []uint64, n int32) int32 {
	for i := n - 1; i >= 0; i-- {
		if x[i] != 0 {
			return 64*i + Abc_Tt6LastBit(x[i])
		}
	}
	return -1
}
func Abc_TtFindFirstDiffBit(x, y []uint64, nVars int32) int32 {
	return Abc_TtFindFirstDiffBit2(x, y, truthCountWords(nVars))
}
func Abc_TtFindFirstDiffBit2(x, y []uint64, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if v := x[i] ^ y[i]; v != 0 {
			return 64*i + Abc_Tt6FirstBit(v)
		}
	}
	return -1
}
func Abc_TtFindLastDiffBit(x, y []uint64, nVars int32) int32 {
	return Abc_TtFindLastDiffBit2(x, y, truthCountWords(nVars))
}
func Abc_TtFindLastDiffBit2(x, y []uint64, n int32) int32 {
	for i := n - 1; i >= 0; i-- {
		if v := x[i] ^ y[i]; v != 0 {
			return 64*i + Abc_Tt6LastBit(v)
		}
	}
	return -1
}
func Abc_TtFindFirstZero(x []uint64, nVars int32) int32 {
	n := truthCountWords(nVars)
	for i := int32(0); i < n; i++ {
		if x[i] != ^uint64(0) {
			return 64*i + Abc_Tt6FirstBit(^x[i])
		}
	}
	return -1
}
func Abc_TtFindLastZero(x []uint64, nVars int32) int32 {
	n := truthCountWords(nVars)
	for i := n - 1; i >= 0; i-- {
		if x[i] != ^uint64(0) {
			return 64*i + Abc_Tt6LastBit(^x[i])
		}
	}
	return -1
}

// ReverseVars reverses variable significance (a permutation of truth bit indices).
func Abc_TtReverseVars(t []uint64, nVars int32) {
	n := truthCountWords(nVars)
	out := make([]uint64, n)
	limit := int64(n) * 64
	for i := int64(0); i < limit; i++ {
		j := i
		for k := int32(0); k < nVars/2; k++ {
			a, b := int64(1)<<k, int64(1)<<(nVars-1-k)
			if (j&a != 0) != (j&b != 0) {
				j ^= a | b
			}
		}
		if t[i>>6]&(uint64(1)<<uint(i&63)) != 0 {
			out[j>>6] |= uint64(1) << uint(j&63)
		}
	}
	copy(t, out)
}
func Abc_TtReverseBits(t []uint64, nVars int32) {
	n := truthCountWords(nVars)
	for i := int32(0); i < (n+1)/2; i++ {
		j := n - 1 - i
		a, b := bits.Reverse64(t[i]), bits.Reverse64(t[j])
		t[i], t[j] = b, a
	}
}
func truthCountUnate(t []uint64, nVars, iVar int32, positive bool) int32 {
	if iVar < 0 || iVar >= nVars {
		panic("invalid variable")
	}
	n := truthCountWords(nVars)
	if iVar < 6 {
		shift := uint(1) << uint(iVar)
		var mask uint64
		for i := uint(0); i < 64; i++ {
			if i&shift == 0 {
				mask |= uint64(1) << i
			}
		}
		for _, w := range t[:n] {
			lo := w & mask
			hi := (w >> shift) & mask
			if positive {
				if lo&^hi != 0 {
					return 0
				}
			} else if hi&^lo != 0 {
				return 0
			}
		}
		return 1
	}
	step := int32(1) << uint(iVar-6)
	for base := int32(0); base < n; base += 2 * step {
		for i := int32(0); i < step; i++ {
			lo, hi := t[base+i], t[base+step+i]
			if positive {
				if lo&^hi != 0 {
					return 0
				}
			} else if hi&^lo != 0 {
				return 0
			}
		}
	}
	return 1
}
func Abc_Tt6PosVar(t uint64, i int32) int32         { return truthCountUnate([]uint64{t}, 6, i, true) }
func Abc_Tt6NegVar(t uint64, i int32) int32         { return truthCountUnate([]uint64{t}, 6, i, false) }
func Abc_TtPosVar(t []uint64, nVars, i int32) int32 { return truthCountUnate(t, nVars, i, true) }
func Abc_TtNegVar(t []uint64, nVars, i int32) int32 { return truthCountUnate(t, nVars, i, false) }
func Abc_TtIsUnate(t []uint64, nVars int32) int32 {
	for i := int32(0); i < nVars; i++ {
		if Abc_TtPosVar(t, nVars, i) == 0 && Abc_TtNegVar(t, nVars, i) == 0 {
			return 0
		}
	}
	return 1
}
func Abc_TtIsPosUnate(t []uint64, nVars int32) int32 {
	for i := int32(0); i < nVars; i++ {
		if Abc_TtPosVar(t, nVars, i) == 0 {
			return 0
		}
	}
	return 1
}
func Abc_TtMakePosUnate(t []uint64, nVars int32) {
	n := truthCountWords(nVars)
	for v := int32(0); v < nVars; v++ {
		if Abc_TtNegVar(t, nVars, v) == 0 {
			if Abc_TtPosVar(t, nVars, v) == 0 {
				panic("function is not unate")
			}
			continue
		}
		if v < 6 {
			shift := uint(1) << uint(v)
			var mask uint64
			for i := uint(0); i < 64; i++ {
				if i&shift == 0 {
					mask |= uint64(1) << i
				}
			}
			for j := int32(0); j < n; j++ {
				t[j] = ((t[j] & mask) << shift) | ((t[j] >> shift) & mask)
			}
		} else {
			step := int32(1) << uint(v-6)
			for base := int32(0); base < n; base += 2 * step {
				for i := int32(0); i < step; i++ {
					a, b := base+i, base+step+i
					t[a], t[b] = t[b], t[a]
				}
			}
		}
	}
}
