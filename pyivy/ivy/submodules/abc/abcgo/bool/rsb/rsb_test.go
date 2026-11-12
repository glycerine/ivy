package rsb

import (
	"math/rand"
	"slices"
	"testing"
)

var testVars = [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}

func composeTruth(function uint64, fanins []int32, divisors []uint64) uint64 {
	var result uint64
	for a := 0; a < 64; a++ {
		mint := uint(0)
		for i, index := range fanins {
			mint |= uint(divisors[index]>>uint(a)&1) << uint(i)
		}
		if function&(1<<mint) != 0 {
			result |= 1 << uint(a)
		}
	}
	return result
}
func TestAllThreeVariableResubstitutions(t *testing.T) {
	p := Rsb_ManAlloc(6, 64, 4, 0)
	for f := uint64(0); f < 256; f++ {
		truth := rsbStretch(f, 3)
		var off, on uint64
		if Rsb_ManPerformResub6(p, 3, truth, testVars[:3], &off, &on, 0) != 1 {
			t.Fatal("missing decomposition", f)
		}
		if composeTruth(on, p.fanins, testVars[:3]) != truth || off&on != 0 {
			t.Fatal("wrong decomposition", f, p.fanins)
		}
		if len(p.fanins) > 3 {
			t.Fatal("unnecessary divisor")
		}
	}
}
func TestCounterexampleCheckWithinAndAcrossWords(t *testing.T) {
	for _, tc := range []struct {
		n    int32
		f    []uint64
		g    [][]uint64
		a, b int32
	}{{6, []uint64{testVars[0]}, [][]uint64{{0}}, 0, 1}, {7, []uint64{0, ^uint64(0)}, [][]uint64{{testVars[0], testVars[0]}}, 0, 64}} {
		a, b := int32(-1), int32(-1)
		pat := make([]uint32, rsbWords(tc.n))
		if Rsb_DecCheck(tc.n, tc.f, tc.g, 1, pat, &a, &b) != 0 || a != tc.a || b != tc.b {
			t.Fatal("unexpected CEX", a, b, tc.a, tc.b)
		}
		if !Rsb_DecVerifyCex(tc.f, tc.g, 1, a, b) {
			t.Fatal("invalid counterexample")
		}
	}
	a, b := int32(-1), int32(-1)
	if got := Rsb_DecCheck(7, []uint64{0, ^uint64(0)}, [][]uint64{{0, ^uint64(0)}}, 1, make([]uint32, 2), &a, &b); got != 9 {
		t.Fatal("pattern orientation", got)
	}
}
func TestPatternPrimitivesAgainstMinterms(t *testing.T) {
	r := rand.New(rand.NewSource(37))
	for n := 0; n <= 5; n++ {
		for trial := 0; trial < 30; trial++ {
			care := r.Uint64()
			f := make([]uint64, n)
			for i := range f {
				f[i] = r.Uint64()
			}
			want := uint32(0)
			for a := 0; a < 64; a++ {
				if care&(1<<uint(a)) == 0 {
					continue
				}
				index := uint(0)
				for i, x := range f {
					index |= uint(x>>uint(a)&1) << uint(i)
				}
				want |= 1 << index
			}
			if got := decTry(care, f); got != want {
				t.Fatal("pattern", n, trial)
			}
		}
	}
}
func TestReusableManagerFailureClearsFanins(t *testing.T) {
	p := Rsb_ManAlloc(6, 64, 4, 0)
	var off, on uint64
	if Rsb_ManPerformResub6(p, 6, testVars[0], testVars[:1], &off, &on, 0) != 1 || !slices.Equal(p.fanins, []int32{0}) {
		t.Fatal("initial success")
	}
	off, on = 123, 456
	if Rsb_ManPerformResub6(p, 6, testVars[0], []uint64{0}, &off, &on, 0) != 0 || len(p.fanins) != 0 || off != 123 || on != 456 {
		t.Fatal("stale failed result", p.fanins, off, on)
	}
}
func TestVerifyDetectsWrongResult(t *testing.T) {
	p := Rsb_ManAlloc(6, 64, 4, 0)
	f := []uint64{testVars[0] | testVars[1]}
	g := [][]uint64{{testVars[0]}, {testVars[1]}}
	pattern := Rsb_DecPerformInt(p, 6, f, g, 2, 2, 0)
	if pattern == 0 {
		t.Fatal("missing OR")
	}
	a, b := make([]uint64, 1), make([]uint64, 1)
	if Rsb_DecVerify(p, 6, f, g, 2, pattern, a, b) != 1 {
		t.Fatal("valid result rejected")
	}
	if Rsb_DecVerify(p, 6, f, g, 2, 0, a, b) != 0 {
		t.Fatal("wrong result accepted")
	}
}
func TestDivisorLimitAndSelectionOrder(t *testing.T) {
	p := Rsb_ManAlloc(6, 64, 2, 0)
	truth := testVars[0] ^ testVars[1] ^ testVars[2]
	var off, on uint64
	if Rsb_ManPerformResub6(p, 3, truth, testVars[:3], &off, &on, 0) != 0 {
		t.Fatal("decomposition limit ignored")
	}
	p = Rsb_ManAlloc(6, 64, 4, 0)
	if Rsb_ManPerformResub6(p, 3, truth, testVars[:3], &off, &on, 0) != 1 || !slices.Equal(p.fanins, []int32{2, 1, 0}) {
		t.Fatal("enumeration order", p.fanins)
	}
	divisors := append(append([]uint64(nil), testVars[:]...), testVars[0]&testVars[1], testVars[2]&testVars[3], testVars[4]&testVars[5])
	truth = divisors[6] | divisors[7] | divisors[8]
	if Rsb_ManPerformResub6(p, 6, truth, divisors, &off, &on, 0) != 1 || composeTruth(on, p.fanins, divisors) != truth {
		t.Fatal("original example", p.fanins)
	}
}
func TestCexInitializationAndConstantCase(t *testing.T) {
	f := []uint64{testVars[0]}
	g := [][]uint64{{0}, {testVars[0]}}
	cexes := make([]uint64, 2)
	var tries []int32
	if Rsb_DecInitCexes(6, f, g, 2, 2, cexes, &tries) != 4 || cexes[0] != 15 || cexes[1] != 0 || len(tries) != 4 {
		t.Fatal("initial CEXes", cexes, tries)
	}
	if Rsb_DecInitCexes(6, []uint64{0}, g, 2, 2, cexes, nil) != 0 {
		t.Fatal("constant has no distinguishing pair")
	}
}
func TestGeneralMultiwordSearch(t *testing.T) {
	p := Rsb_ManAlloc(8, 64, 4, 0)
	f := []uint64{0, testVars[0], testVars[1], testVars[0] | testVars[1]}
	g := [][]uint64{{testVars[0], testVars[0], testVars[0], testVars[0]}, {testVars[1], testVars[1], testVars[1], testVars[1]}, {0, ^uint64(0), 0, ^uint64(0)}, {0, 0, ^uint64(0), ^uint64(0)}}
	pattern := Rsb_ManPerform(p, 8, f, g, 4, 4, 0)
	if pattern == 0 {
		t.Fatal("general search failed")
	}
	if Rsb_DecVerify(p, 8, f, g, 4, pattern, make([]uint64, 4), make([]uint64, 4)) != 1 {
		t.Fatal("general verification")
	}
}

// This hash was produced by compiling the unchanged original Rsb_DecTry* /
// Rsb_DecCheck routines and checking all 256 three-variable truth functions
// against the first 1,2,3,4 elementary divisors. Each case hashes both its
// returned pattern and its exact ordered counterexample pair.
func TestOriginalCPatternAndCounterexampleOracle(t *testing.T) {
	hash := uint64(1469598103934665603)
	g := [][]uint64{{testVars[0]}, {testVars[1]}, {testVars[2]}, {testVars[3]}}
	for f := uint64(0); f < 256; f++ {
		truth := []uint64{rsbStretch(f, 3)}
		for n := int32(1); n <= 4; n++ {
			a, b := int32(-1), int32(-1)
			pattern := Rsb_DecCheck(6, truth, g, n, make([]uint32, 1), &a, &b)
			for _, value := range []uint32{pattern, uint32(a), uint32(b)} {
				hash ^= uint64(value)
				hash *= 1099511628211
			}
		}
	}
	if hash != 8994635971493543787 {
		t.Fatalf("Go pattern/CEX hash %d does not match original C", hash)
	}
}

func TestCounterexampleBudgetStopsAt64(t *testing.T) {
	p := Rsb_ManAlloc(8, 100, 4, 1)
	f := []uint64{^uint64(1), ^uint64(0), ^uint64(0), ^uint64(0)}
	g := make([][]uint64, 100)
	for i := range g {
		g[i] = append([]uint64(nil), f...)
		a := i + 1
		g[i][a>>6] &^= 1 << uint(a&63)
	}
	if Rsb_DecPerformInt(p, 8, f, g, 100, 100, 0) != 0 || len(p.fanins) != 0 {
		t.Fatal("budget result")
	}
	if len(p.tries) != 128 {
		t.Fatal("expected exactly 64 recorded attempts", len(p.tries))
	}
	for i := 0; i < 64; i++ {
		if p.cexes[i] != 1<<uint(i) {
			t.Fatal("counterexample recording overflow", i, p.cexes[i])
		}
	}
}
func TestVerboseDiagnosticHasValidIsopCounter(t *testing.T) {
	p := Rsb_ManAlloc(6, 64, 4, 1)
	var off, on uint64
	if Rsb_ManPerformResub6(p, 6, testVars[0]|testVars[1], testVars[:2], &off, &on, 1) != 1 {
		t.Fatal("verbose decomposition failed")
	}
}

// Native original rsbMan.c/rsbDec6.c, compiled with LTO and no source edits,
// produced this hash over status, both returned truth words, fanin count and
// ordered fanin IDs for all 256 three-variable functions with three divisors.
func TestOriginalCFullSearchOracle(t *testing.T) {
	hash := uint64(1469598103934665603)
	p := Rsb_ManAlloc(6, 64, 4, 0)
	for f := uint64(0); f < 256; f++ {
		var off, on uint64
		status := Rsb_ManPerformResub6(p, 3, rsbStretch(f, 3), testVars[:3], &off, &on, 0)
		values := []uint64{uint64(status), off, on, uint64(len(p.fanins))}
		for _, index := range p.fanins {
			values = append(values, uint64(index))
		}
		for _, value := range values {
			hash ^= value
			hash *= 1099511628211
		}
	}
	if hash != 11774239727511085057 {
		t.Fatalf("Go full-search hash %d differs from original C", hash)
	}
}

func TestFourDivisorSearchAndRandomComposition(t *testing.T) {
	p := Rsb_ManAlloc(6, 64, 4, 0)
	var off, on uint64
	truth := testVars[0] ^ testVars[1] ^ testVars[2] ^ testVars[3]
	if Rsb_ManPerformResub6(p, 4, truth, testVars[:4], &off, &on, 0) != 1 || !slices.Equal(p.fanins, []int32{3, 2, 1, 0}) || composeTruth(on, p.fanins, testVars[:4]) != truth {
		t.Fatal("quadruple order", p.fanins)
	}
	r := rand.New(rand.NewSource(71))
	for trial := 0; trial < 100; trial++ {
		divisors := []uint64{r.Uint64(), r.Uint64(), r.Uint64(), r.Uint64()}
		lut := rsbStretch(uint64(r.Uint32()), 4)
		truth = composeTruth(lut, []int32{0, 1, 2, 3}, divisors)
		if Rsb_ManPerformResub6(p, 6, truth, divisors, &off, &on, 0) != 1 || composeTruth(on, p.fanins, divisors) != truth || off&on != 0 {
			t.Fatal("random composition", trial)
		}
	}
}

func TestLargeDecompositionBudgetStillUsesFourInputs(t *testing.T) {
	p := Rsb_ManAlloc(6, 64, 99, 0)
	var off, on uint64
	truth := testVars[0] ^ testVars[1] ^ testVars[2] ^ testVars[3]
	if Rsb_ManPerformResub6(p, 4, truth, testVars[:4], &off, &on, 0) != 1 || len(p.fanins) != 4 {
		t.Fatal("source four-input budget behavior")
	}
}

func TestOldFaninSnapshotAndVerification(t *testing.T) {
	p := Rsb_ManAlloc(6, 64, 4, 0)
	old := []int32{1}
	Rsb_ManSetFaninsOld(p, old)
	old[0] = 0
	if !slices.Equal(Rsb_ManGetFaninsOld(p), []int32{1}) {
		t.Fatal("old fanins not copied")
	}
	g := [][]uint64{{testVars[0]}, {testVars[1]}}
	if Rsb_ManPerform(p, 6, []uint64{testVars[1]}, g, 2, 2, 0) == 0 || !slices.Equal(Rsb_ManGetFanins(p), []int32{1}) {
		t.Fatal("old fanin verification altered search")
	}
}

func TestGeneralConsistencyAgainstDirectPartition(t *testing.T) {
	random := rand.New(rand.NewSource(81))
	for n := int32(7); n <= 9; n++ {
		for count := int32(1); count <= 4; count++ {
			for trial := 0; trial < 12; trial++ {
				words := rsbWords(n)
				f := make([]uint64, words)
				g := make([][]uint64, count)
				for i := range g {
					g[i] = make([]uint64, words)
					for w := range g[i] {
						g[i][w] = random.Uint64()
					}
				}
				for w := range f {
					f[w] = random.Uint64()
				}
				if trial&1 == 0 {
					function := random.Uint32()
					clear(f)
					for a := 0; a < words*64; a++ {
						index := uint(0)
						for i, t := range g {
							index |= uint(rsbBit(t, int32(a))) << uint(i)
						}
						if function&(1<<index) != 0 {
							f[a>>6] |= 1 << uint(a&63)
						}
					}
				}
				seen := map[int]uint64{}
				conflict := false
				var expected uint32
				for a := 0; a < words*64; a++ {
					index := 0
					for i, t := range g {
						index |= int(rsbBit(t, int32(a))) << uint(i)
					}
					value := rsbBit(f, int32(a))
					if previous, ok := seen[index]; ok && previous != value {
						conflict = true
					}
					seen[index] = value
					expected |= 1 << uint(index+int(value)*(1<<uint(count)))
				}
				a, b := int32(-1), int32(-1)
				got := Rsb_DecCheck(n, f, g, count, make([]uint32, words), &a, &b)
				if conflict {
					if got != 0 || !Rsb_DecVerifyCex(f, g, count, a, b) {
						t.Fatal("partition conflict not witnessed", n, count, trial)
					}
				} else if got != expected {
					t.Fatal("reachable pattern differs", n, count, trial)
				}
			}
		}
	}
}
