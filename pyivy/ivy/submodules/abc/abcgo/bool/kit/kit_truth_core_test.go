package kit

import (
	"math/rand"
	"testing"
)

func TestTruthCoreOperations(t *testing.T) {
	rng := rand.New(rand.NewSource(9))
	for vars := int32(0); vars <= 12; vars++ {
		size := Kit_TruthWordNum(vars)
		a, b, c, out := make([]uint32, size), make([]uint32, size), make([]uint32, size), make([]uint32, size)
		for i := range a {
			a[i] = rng.Uint32()
			b[i] = rng.Uint32()
			c[i] = rng.Uint32()
		}
		for x := 0; x < 4; x++ {
			Kit_TruthAndPhase(out, a, b, vars, int32(x&1), int32(x>>1))
			for i := range a {
				av, bv := a[i], b[i]
				if x&1 != 0 {
					av = ^av
				}
				if x&2 != 0 {
					bv = ^bv
				}
				if out[i] != av&bv {
					t.Fatal("phase AND", vars, x, i)
				}
			}
		}
		Kit_TruthMux(out, a, b, c, vars)
		for i := range out {
			if out[i] != (a[i]&^c[i] | b[i]&c[i]) {
				t.Fatal("mux")
			}
		}
		Kit_TruthCopy(out, a, vars)
		if Kit_TruthIsEqual(out, a, vars) != 1 {
			t.Fatal("copy")
		}
		Kit_TruthNot(out, out, vars)
		if Kit_TruthIsOpposite(out, a, vars) != 1 || Kit_TruthIsEqualWithPhase(out, a, vars) != 1 {
			t.Fatal("complement")
		}
		Kit_TruthClear(out, vars)
		if Kit_TruthIsConst0(out, vars) != 1 || Kit_TruthCountOnes(out, vars) != 0 {
			t.Fatal("clear")
		}
		Kit_TruthFill(out, vars)
		if Kit_TruthIsConst1(out, vars) != 1 || Kit_TruthCountOnes(out, vars) != size*32 {
			t.Fatal("fill")
		}
		for v := int32(0); v < vars; v++ {
			Kit_TruthIthVar(out, vars, v)
			for bit := int32(0); bit < size*32; bit++ {
				if Kit_TruthHasBit(out, bit) != (bit>>uint(v))&1 {
					t.Fatal("elementary variable", vars, v, bit)
				}
			}
		}
	}
}
func TestTruthBitHelpers(t *testing.T) {
	p := make([]uint32, 3)
	Kit_TruthSetBit(p, 66)
	if Kit_TruthFindFirstBit(p, 6) != -1 {
		t.Fatal("only first two words are in six-var truth")
	}
	if Kit_TruthHasBit(p, 66) != 1 {
		t.Fatal("set bit")
	}
	Kit_TruthXorBit(p, 66)
	if p[2] != 0 {
		t.Fatal("xor bit")
	}
	for i := int32(0); i <= 32; i++ {
		m := Kit_BitMask(i)
		if Kit_WordCountOnes(m) != i {
			t.Fatal("mask", i, m)
		}
	}
	if Kit_WordFindFirstBit(0) != -1 || Kit_WordHasOneBit(0) != 1 || Kit_BitWordNum(2147483647) != 67108864 {
		t.Fatal("word helper edge case")
	}
}
