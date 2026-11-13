package util

import (
	"math/rand"
	"testing"
)

func TestTruthFoundationBitsAndDigits(t *testing.T) {
	table := make([]word, 4)
	for i := int32(0); i < 256; i += 3 {
		Abc_TtSetBit(table, i)
	}
	for i := int32(0); i < 256; i++ {
		if Abc_TtGetBit(table, i) != ttBool(i%3 == 0) {
			t.Fatal(i)
		}
	}
	Abc_TtXorBit(table, 0)
	if Abc_TtGetBit(table, 0) != 0 {
		t.Fatal("xor bit")
	}
	clear(table)
	Abc_TtSetHex(table, 17, 13)
	if Abc_TtGetHex(table, 17) != 13 {
		t.Fatal(table)
	}
	Abc_TtXorHex(table, 17, 3)
	if Abc_TtGetHex(table, 17) != 14 {
		t.Fatal(table)
	}
	clear(table)
	Abc_TtSetQua(table, 40, 3)
	if Abc_TtGetQua(table, 40) != 3 {
		t.Fatal(table)
	}
	clear(table)
	Abc_TtSet256(table, 9, 123)
	if Abc_TtGet256(table, 9) != 123 {
		t.Fatal(table)
	}
	for i := int32(0); i <= 64; i++ {
		mask := Abc_Tt6Mask(i)
		for bit := int32(0); bit < 64; bit++ {
			if (mask>>bit)&1 != word(ttBool(bit < i)) {
				t.Fatal(i, bit)
			}
		}
	}
	Abc_TtMask(table, 4, 130)
	if table[0] != ^word(0) || table[1] != ^word(0) || table[2] != 3 || table[3] != 0 {
		t.Fatal(table)
	}
}
func TestTruthFoundationBooleanOperations(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for trial := 0; trial < 100; trial++ {
		a, b, c := []word{rng.Uint64(), rng.Uint64()}, []word{rng.Uint64(), rng.Uint64()}, []word{rng.Uint64(), rng.Uint64()}
		out := make([]word, 2)
		for complement := int32(0); complement < 2; complement++ {
			Abc_TtAnd(out, a, b, 2, complement)
			for i := range out {
				if out[i] != ttCompl(a[i]&b[i], complement) {
					t.Fatal("and")
				}
			}
			Abc_TtXor(out, a, b, 2, complement)
			for i := range out {
				if out[i] != ttCompl(a[i]^b[i], complement) {
					t.Fatal("xor")
				}
			}
			Abc_TtXorMask(out, a, b, c, 2, complement)
			for i := range out {
				if out[i] != (a[i]^b[i])&ttCompl(c[i], complement) {
					t.Fatal("masked xor")
				}
			}
		}
		Abc_TtMux(out, c, a, b, 2)
		for i := range out {
			if out[i] != (c[i]&a[i])|(^c[i]&b[i]) {
				t.Fatal("mux")
			}
		}
		Abc_TtMaj(out, a, b, c, 2)
		for i := range out {
			if out[i] != (a[i]&b[i])|(a[i]&c[i])|(b[i]&c[i]) {
				t.Fatal("majority")
			}
		}
		copied := append([]word(nil), a...)
		Abc_TtAnd(copied, copied, b, 2, 0)
		for i := range out {
			if copied[i] != a[i]&b[i] {
				t.Fatal("aliasing")
			}
		}
	}
}
func TestTruthFoundationVariables(t *testing.T) {
	for n := int32(1); n <= 10; n++ {
		for v := int32(0); v < n; v++ {
			table := make([]word, Abc_TtWordNum(n))
			Abc_TtIthVar(table, v, n)
			for bit := int32(0); bit < int32(len(table))*64; bit++ {
				if Abc_TtGetBit(table, bit) != (bit>>v)&1 {
					t.Fatal(n, v, bit)
				}
			}
		}
	}
	for i := range s_PMasks {
		a, b := s_Truths6[i], s_Truths6[i+1]
		if s_PMasks[i][0] != ^(a^b) || s_PMasks[i][1] != a&^b || s_PMasks[i][2] != ^a&b {
			t.Fatal(i)
		}
	}
}

func TestTruthLutEvaluation(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	for n := int32(0); n <= 6; n++ {
		for repeat := 0; repeat < 20; repeat++ {
			ins := make([]word, n)
			for i := range ins {
				ins[i] = rng.Uint64()
			}
			lut := rng.Uint64()
			got := Abc_TtEvalLut6(ins, lut, n)
			var want word
			for bit := 0; bit < 64; bit++ {
				var idx word
				for i, x := range ins {
					idx |= ((x >> bit) & 1) << i
				}
				want |= ((lut >> idx) & 1) << bit
			}
			if got != want {
				t.Fatal(n, got, want)
			}
		}
	}
	for n := int32(0); n <= 4; n++ {
		ins := []int32{0xAAAA, 0xCCCC, 0xF0F0, 0xFF00}
		lut := int32(0x6996)
		got := Abc_TtEvalLut4(ins, lut, n)
		want := lut & int32(Abc_Tt6Mask(1<<n))
		if got != want {
			t.Fatal(n, got, want)
		}
	}
}
func TestTruthSymmetry(t *testing.T) {
	for n := int32(1); n <= 10; n++ {
		for polar := int32(0); polar < 1<<(n+1); polar += 17 {
			table := make([]word, Abc_TtWordNum(n))
			Abc_TtGenFullySymmetric(table, n, polar)
			if got := Abc_TtIsFullySymmetric(table, n); got != polar {
				t.Fatal(n, polar, got)
			}
		}
	}
	f := Abc_TtSymFunGenerate("00011", 4)
	if Abc_TtIsFullySymmetric(f, 4) != 24 {
		t.Fatal(f)
	}
	Abc_TtXorBit(f, 1)
	if Abc_TtIsFullySymmetric(f, 4) != -1 {
		t.Fatal("asymmetry missed")
	}
	Abc_TtTestFullySymmetric()
	for i := int32(0); i < 5; i++ {
		for j := i + 1; j < 6; j++ {
			t0 := s_Truths6[i] & s_Truths6[j]
			if Abc_Tt6VarsAreSymmetric(t0, i, j) != 1 {
				t.Fatal(i, j)
			}
			t1 := s_Truths6[i] &^ s_Truths6[j]
			if Abc_Tt6VarsAreAntiSymmetric(t1, i, j) != 1 {
				t.Fatal(i, j)
			}
		}
	}
}
func TestTruthBiDecompositionExhaustiveThreeVariables(t *testing.T) {
	for fn := word(0); fn < 256; fn++ {
		truth := fn
		for shift := uint(8); shift < 64; shift += 8 {
			truth |= fn << shift
		}
		table := []word{truth}
		result := Abc_TtProcessBiDec(table, 3, 2)
		if result == 0 {
			continue
		}
		if result&(1<<30) != 0 {
			table[0] = ^table[0]
		}
		a, b := result&65535, (result>>16)&65535
		ca, cb := Abc_TtDeriveBiDecOne(table, 3, a), Abc_TtDeriveBiDecOne(table, 3, b)
		if Abc_TtVerifyBiDec(table, 3, a, b, 2, ca, cb) != 1 {
			t.Fatalf("function %x decomposition %x", truth, result)
		}
	}
}
func TestTruthBiDecompositionSharedHighVariables(t *testing.T) {
	const n int32 = 8
	table := make([]word, Abc_TtWordNum(n))
	// F=(x0 XOR x6) AND (x1 OR x7), two disjoint two-input factors.
	for m := int32(0); m < 1<<n; m++ {
		if ((m>>0)^(m>>6))&1 != 0 && ((m>>1)|(m>>7))&1 != 0 {
			Abc_TtSetBit(table, m)
		}
	}
	a, b := int32(1<<0|1<<6), int32(1<<1|1<<7)
	if Abc_TtCheckBiDec(table, n, a, b) != 1 {
		t.Fatal("valid decomposition rejected")
	}
	ca, cb := Abc_TtDeriveBiDecOne(table, n, a), Abc_TtDeriveBiDecOne(table, n, b)
	if Abc_TtVerifyBiDec(table, n, a, b, 2, ca, cb) != 1 {
		t.Fatal(ca, cb)
	}
	Abc_TtExist(table, 6, Abc_TtWordNum(n))
	for m := int32(0); m < 1<<n; m++ {
		if Abc_TtGetBit(table, m) != ((m>>1)|(m>>7))&1 {
			t.Fatal(m)
		}
	}
}

func TestTruthWordBitMatches32BitHalves(t *testing.T) {
	words := []uint64{0x8000000180000002, 0x8000000180000001}
	halves := []uint32{uint32(words[0]), uint32(words[0] >> 32), uint32(words[1]), uint32(words[1] >> 32)}
	for i := int32(0); i < 128; i++ {
		if got, want := Abc_TtGetBit(words, i) != 0, Abc_InfoHasBit(halves, i); got != want {
			t.Fatalf("bit %d: 64-bit accessor = %v, 32-bit halves = %v", i, got, want)
		}
	}
}
