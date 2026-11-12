package util

import (
	"math/rand"
	"strings"
	"testing"
)

func cofactorFixture(n int32, fn func(int) bool) []word {
	a := make([]word, ttWords(n))
	for i := 0; i < len(a)*64; i++ {
		if fn(i) {
			a[i/64] |= word(1) << uint(i%64)
		}
	}
	return a
}
func TestTruthCofactorReference(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	for n := int32(6); n <= 9; n++ {
		a := make([]word, ttWords(n))
		for i := range a {
			a[i] = r.Uint64()
		}
		for v := int32(0); v < n; v++ {
			for phase := 0; phase < 2; phase++ {
				b := append([]word(nil), a...)
				if phase == 0 {
					Abc_TtCofactor0(b, int32(len(b)), v)
				} else {
					Abc_TtCofactor1(b, int32(len(b)), v)
				}
				for i := 0; i < len(a)*64; i++ {
					j := i&^(1<<v) | phase<<v
					if ttBit(b, i) != ttBit(a, j) {
						t.Fatal(n, v, phase, i)
					}
				}
			}
			b := append([]word(nil), a...)
			Abc_TtFlip(b, int32(len(b)), v)
			for i := 0; i < len(a)*64; i++ {
				if ttBit(b, i) != ttBit(a, i^(1<<v)) {
					t.Fatal("flip", n, v, i)
				}
			}
		}
	}
}
func TestTruthSwapsAndSupport(t *testing.T) {
	for n := int32(2); n <= 9; n++ {
		a := cofactorFixture(n, func(i int) bool { return i&1 != 0 != (i&(1<<(n-1)) != 0) })
		if Abc_TtSupport(a, n) != (1 | 1<<(n-1)) {
			t.Fatal("support", n)
		}
		for i := int32(0); i < n; i++ {
			for j := i + 1; j < n; j++ {
				b := append([]word(nil), a...)
				Abc_TtSwapVars(b, n, i, j)
				for k := 0; k < len(a)*64; k++ {
					q := k
					if (k>>i)&1 != (k>>j)&1 {
						q ^= 1<<i | 1<<j
					}
					if ttBit(a, q) != ttBit(b, k) {
						t.Fatal("swap", n, i, j, k)
					}
				}
			}
		}
	}
}
func TestTruthSmallPredicatesAndStretch(t *testing.T) {
	for v := int32(0); v < 6; v++ {
		a := ttMask(v)
		if Abc_Tt6Cof0IsConst0(a, v) != 1 || Abc_Tt6Cof1IsConst1(a, v) != 1 || Abc_Tt6CofsOpposite(a, v) != 1 {
			t.Fatal(v)
		}
		if Abc_Tt6HasVar(a, v) != 1 {
			t.Fatal(v)
		}
	}
	for n := int32(0); n <= 6; n++ {
		a := Abc_Tt6Stretch(1, n)
		for i := 0; i < 64; i++ {
			if (a>>i)&1 != word(ttFlag(i%(1<<n) == 0)) {
				t.Fatal("stretch", n, i)
			}
		}
	}
	if Abc_Tt6IsAndType(1, 3) != 1 || Abc_Tt6IsOrType(^word(1), 3) != 1 {
		t.Fatal("elementary")
	}
	var parity word
	for i := 0; i < 64; i++ {
		v := i
		ones := 0
		for v > 0 {
			ones += v & 1
			v >>= 1
		}
		if ones&1 != 0 {
			parity |= 1 << i
		}
	}
	if Abc_Tt6IsXorType(parity, 6) != 1 {
		t.Fatal("parity")
	}
}
func TestTruthMinBaseExpandAndNpn(t *testing.T) {
	a := cofactorFixture(6, func(i int) bool { return i&1 != 0 != (i&4 != 0) })
	vars := []int32{10, 11, 12, 13, 14, 15}
	var size int32
	if Abc_TtMinimumBase(a, vars, 6, &size) != 1 || size != 2 || vars[0] != 10 || vars[1] != 12 {
		t.Fatal(size, vars)
	}
	if Abc_TtSupport(a, 6) != 3 {
		t.Fatal("compressed support")
	}
	Abc_TtExpand(a, 6, []int32{10, 12}, 2, []int32{10, 11, 12}, 3)
	if Abc_TtSupport(a, 6) != 5 {
		t.Fatal("expanded support")
	}
	original := cofactorFixture(6, func(i int) bool { return i&1 != 0 && i&2 == 0 || i&4 != 0 })[0]
	b := []word{original}
	perm := []byte{2, 0, 1}
	phase := uint32(1 | 1<<3)
	Abc_TtImplementNpnConfig(b, 3, perm, phase)
	for i := 0; i < 64; i++ {
		q := i&^7 | ((i >> 2) & 1) | ((i & 1) << 1) | (((i >> 1) & 1) << 2)
		want := ttBit([]word{original}, q^1)
		if ttBit(b, i) == want {
			t.Fatal("npn", i)
		}
	}
	if perm[0] != 0 || perm[1] != 1 || perm[2] != 2 {
		t.Fatal(perm)
	}
}
func TestTruthHexRoundTrip(t *testing.T) {
	for _, s := range []string{"0", "F", "5", "A", "6", "96", "6996", "0123456789ABCDEF", "0123456789ABCDEFFEDCBA9876543210"} {
		a := make([]word, 4)
		n := Abc_TtReadHex(a, s)
		if n >= 2 {
			var out strings.Builder
			Abc_TtPrintHexRev(&out, a, n)
			if !strings.EqualFold(strings.TrimLeft(out.String(), "0"), strings.TrimLeft(s, "0")) {
				t.Fatal(s, n, out.String())
			}
			buf := make([]byte, 128)
			count := Abc_TtWriteHexRev(buf, a, n)
			if string(buf[:count]) != out.String() {
				t.Fatal("buffer")
			}
		}
	}
	a := []word{^word(0)}
	if Abc_TtReadHexNumber(a, "12 ") != 2 || a[0] != ^word(0) {
		t.Fatal(a)
	}
}
