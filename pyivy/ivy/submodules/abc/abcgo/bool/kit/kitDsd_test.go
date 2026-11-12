package kit

import (
	"math/rand"
	"testing"
)

func TestDsdDecompositionRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(22))
	for n := int32(0); n <= 9; n++ {
		trials := 30
		if n == 3 {
			trials = 256
		}
		for trial := 0; trial < trials; trial++ {
			truth := make([]uint32, Kit_TruthWordNum(n))
			for i := range truth {
				truth[i] = rng.Uint32()
			}
			if n == 3 {
				truth[0] = uint32(trial)
			}
			if n < 5 {
				mask := Kit_BitMask(1 << uint(n))
				truth[0] &= mask
				for shift := uint(1) << uint(n); shift < 32; shift *= 2 {
					truth[0] |= truth[0] << shift
				}
			}
			for _, limit := range []int32{0, 3} {
				ntk := Kit_DsdDecomposeMux(truth, n, limit)
				out := make([]uint32, len(truth))
				Kit_DsdTruth(ntk, out)
				if Kit_TruthIsEqual(out, truth, n) == 0 {
					t.Fatalf("vars=%d trial=%d limit=%d: %x != %x", n, trial, limit, out, truth)
				}
				expanded := Kit_DsdExpand(ntk)
				Kit_DsdTruth(expanded, out)
				if Kit_TruthIsEqual(out, truth, n) == 0 {
					t.Fatal("expansion changed truth", n, trial)
				}
				priorities := make([]int32, n)
				for i := range priorities {
					priorities[i] = int32(i)
				}
				shrunk := Kit_DsdShrink(expanded, priorities)
				Kit_DsdTruth(shrunk, out)
				if Kit_TruthIsEqual(out, truth, n) == 0 {
					t.Fatal("shrink changed truth", n, trial)
				}
			}
		}
	}
}
func TestDsdPrimeAndSimpleSupports(t *testing.T) {
	and := Kit_DsdDecompose([]uint32{0x80808080}, 3)
	if Kit_DsdNonDsdPrimeMax(and) != nil {
		t.Fatal("AND retained a prime")
	}
	if Kit_DsdGetSupports(and) != 7 {
		t.Fatal("AND support")
	}
	threshold := uint32(0)
	for x := 0; x < 32; x++ {
		count := 0
		for bit := 0; bit < 4; bit++ {
			count += x >> uint(bit) & 1
		}
		if count >= 2 {
			threshold |= 1 << uint(x)
		}
	}
	ntk := Kit_DsdDecompose([]uint32{threshold}, 4)
	prime := Kit_DsdNonDsdPrimeMax(ntk)
	if prime == nil || Kit_DsdObjFanNum(prime) != 4 || len(Kit_DsdObjTruth(prime)) != 1 {
		t.Fatal("threshold prime missing")
	}
	before := make([]uint32, 1)
	Kit_DsdTruth(ntk, before)
	Kit_DsdRotate(ntk, []int32{1, 2, 3, 4})
	after := make([]uint32, 1)
	Kit_DsdTruth(ntk, after)
	if after[0] != before[0] {
		t.Fatal("rotation changed truth")
	}
}
func TestDsdPartialComposition(t *testing.T) {
	// (a XOR b) AND c: factor the {a,b} bound set into a replacement at a.
	truth := []uint32{0x60606060}
	ntk := Kit_DsdDecompose(truth, 3)
	p := Kit_DsdManAlloc(3, 32)
	co, dec := make([]uint32, 1), make([]uint32, 1)
	Kit_DsdTruthPartialTwo(p, ntk, 3, 0, co, dec)
	for x := 0; x < 32; x++ {
		bound := Kit_TruthHasBit(dec, int32(x))
		mapped := int32(x)&^1 | bound
		if Kit_TruthHasBit(co, mapped) != Kit_TruthHasBit(truth, int32(x)) {
			t.Fatal("partial composition", x, co, dec)
		}
	}
}

func TestDsdLargeTruthRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(83))
	for n := int32(10); n <= 16; n++ {
		truth := make([]uint32, Kit_TruthWordNum(n))
		for i := range truth {
			truth[i] = rng.Uint32()
		}
		ntk := Kit_DsdDecompose(truth, n)
		got := make([]uint32, len(truth))
		Kit_DsdTruth(ntk, got)
		if Kit_TruthIsEqual(got, truth, n) == 0 {
			t.Fatal("large DSD round trip", n)
		}
	}
}
func TestDsdCofactoringFindsPrimeReduction(t *testing.T) {
	// Four-input threshold >=2 is prime; one cofactoring variable reduces it to
	// 3-input AND/OR decompositions and no further cofactoring is needed.
	var truth uint32
	for x := 0; x < 32; x++ {
		count := 0
		for v := 0; v < 4; v++ {
			count += x >> uint(v) & 1
		}
		if count >= 2 {
			truth |= 1 << uint(x)
		}
	}
	variables := make([]int32, 4)
	if steps := Kit_DsdCofactoring([]uint32{truth}, 4, variables, 4, 0); steps != 1 {
		t.Fatal(steps, variables)
	}
}
