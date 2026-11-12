package util

import (
	"math/rand"
	"testing"
)

func TestTruthCountsAgainstBits(t *testing.T) {
	r := rand.New(rand.NewSource(17))
	for trial := 0; trial < 200; trial++ {
		x, y, m := make([]uint64, 3), make([]uint64, 3), make([]uint64, 3)
		for i := range x {
			x[i] = r.Uint64()
			y[i] = r.Uint64()
			m[i] = r.Uint64()
		}
		for a := int32(0); a < 2; a++ {
			for b := int32(0); b < 2; b++ {
				var want int32
				for i := range x {
					for bit := uint(0); bit < 64; bit++ {
						xx, yy := x[i]>>bit&1, y[i]>>bit&1
						if a != 0 {
							xx ^= 1
						}
						if b != 0 {
							yy ^= 1
						}
						if xx&yy&(m[i]>>bit&1) != 0 {
							want++
						}
					}
				}
				if got := Abc_TtCountOnesVecMask2(x, y, a, b, m, 3); got != want {
					t.Fatal(got, want)
				}
			}
		}
		var total int32
		for _, v := range x {
			var n int32
			for bit := uint(0); bit < 64; bit++ {
				n += int32(v >> bit & 1)
			}
			if n != Abc_TtCountOnes(v) || n != Abc_TtCountOnesSlow(v) {
				t.Fatal("scalar count")
			}
			total += n
		}
		if total != Abc_TtCountOnesVec(x, 3) {
			t.Fatal("vector count")
		}
		out := append([]uint64(nil), m...)
		got := Abc_TtAndXorSum(out, x, y, 3)
		var want int32
		for i := range out {
			if out[i] != m[i]&(x[i]^y[i]) {
				t.Fatal("and xor")
			}
			want += Abc_TtCountOnes(out[i])
		}
		if got != want {
			t.Fatal("sum")
		}
	}
}
func TestTruthSearches(t *testing.T) {
	x := []uint64{0, 0, 0, 0}
	if Abc_TtFindFirstBit(x, 8) != -1 || Abc_TtFindLastBit(x, 8) != -1 {
		t.Fatal("empty")
	}
	for b := int32(0); b < 256; b++ {
		for i := range x {
			x[i] = 0
		}
		x[b/64] = 1 << uint(b%64)
		if Abc_TtFindFirstBit2(x, 4) != b || Abc_TtFindLastBit2(x, 4) != b {
			t.Fatal(b)
		}
		zero := make([]uint64, 4)
		if Abc_TtFindFirstDiffBit(x, zero, 8) != b || Abc_TtFindLastDiffBit2(x, zero, 4) != b {
			t.Fatal("diff", b)
		}
		for i := range x {
			x[i] = ^x[i]
		}
		if Abc_TtFindFirstZero(x, 8) != b || Abc_TtFindLastZero(x, 8) != b {
			t.Fatal("zero", b)
		}
	}
}
func TestTruthUnateness(t *testing.T) {
	for n := int32(1); n <= 9; n++ {
		for v := int32(0); v < n; v++ {
			x := make([]uint64, truthCountWords(n))
			for b := 0; b < len(x)*64; b++ {
				if b&(1<<v) != 0 {
					x[b/64] |= 1 << uint(b%64)
				}
			}
			if Abc_TtPosVar(x, n, v) != 1 || Abc_TtNegVar(x, n, v) != 0 {
				t.Fatal(n, v)
			}
			for i := range x {
				x[i] = ^x[i]
			}
			if Abc_TtNegVar(x, n, v) != 1 || Abc_TtIsUnate(x, n) != 1 {
				t.Fatal("negative")
			}
			Abc_TtMakePosUnate(x, n)
			if Abc_TtIsPosUnate(x, n) != 1 {
				t.Fatal("normalize")
			}
		}
	}
	xor := uint64(0x6666666666666666)
	if Abc_TtIsUnate([]uint64{xor}, 2) != 0 {
		t.Fatal("xor unate")
	}
}
func TestTruthReversal(t *testing.T) {
	for n := int32(0); n <= 9; n++ {
		r := rand.New(rand.NewSource(int64(n)))
		x := make([]uint64, truthCountWords(n))
		for i := range x {
			x[i] = r.Uint64()
		}
		original := append([]uint64(nil), x...)
		Abc_TtReverseVars(x, n)
		Abc_TtReverseVars(x, n)
		for i := range x {
			if x[i] != original[i] {
				t.Fatal("vars involution")
			}
		}
		Abc_TtReverseBits(x, n)
		Abc_TtReverseBits(x, n)
		for i := range x {
			if x[i] != original[i] {
				t.Fatal("bits involution")
			}
		}
	}
	x := []uint64{0xaaaaaaaaaaaaaaaa}
	Abc_TtReverseVars(x, 3)
	if x[0] != 0xf0f0f0f0f0f0f0f0 {
		t.Fatalf("%x", x[0])
	}
}

func TestTruthMaskedAndXorCounts(t *testing.T) {
	r := rand.New(rand.NewSource(89))
	for k := 0; k < 100; k++ {
		x, y, m := []uint64{r.Uint64(), r.Uint64()}, []uint64{r.Uint64(), r.Uint64()}, []uint64{r.Uint64(), r.Uint64()}
		var xor int32
		for i := range x {
			xor += Abc_TtCountOnes(x[i] ^ y[i])
		}
		if Abc_TtCountOnesVecXor(x, y, 2) != xor {
			t.Fatal("xor")
		}
		for c := int32(0); c < 2; c++ {
			var one, maskedXor int32
			for i := range x {
				a, b := x[i], x[i]^y[i]
				if c != 0 {
					a = ^a
					b = ^b
				}
				one += Abc_TtCountOnes(a & m[i])
				maskedXor += Abc_TtCountOnes(b & m[i])
			}
			if Abc_TtCountOnesVecMask(x, m, 2, c) != one || Abc_TtCountOnesVecXorMask(x, y, c, m, 2) != maskedXor {
				t.Fatal("mask")
			}
		}
	}
	if Abc_Tt6FirstBit(0) != -1 || Abc_Tt6LastBit(0) != -1 {
		t.Fatal("empty scalar")
	}
	for b := int32(0); b < 64; b++ {
		w := uint64(1) << uint(b)
		if Abc_Tt6FirstBit(w) != b || Abc_Tt6LastBit(w) != b {
			t.Fatal(b)
		}
	}
}
