package dau

import (
	_ "embed"
	"encoding/json"
	"math/bits"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestDSDDecompositionRoundTrip(t *testing.T) {
	random := rand.New(rand.NewPCG(221, 663))
	for n := int32(0); n <= 9; n++ {
		for trial := 0; trial < 100; trial++ {
			truth := make([]uint64, dsWords(n))
			for i := range truth {
				truth[i] = random.Uint64()
			}
			if n < 6 {
				mask := uint64(1)<<uint(1<<uint(n)) - 1
				truth[0] &= mask
				for shift := 1 << uint(n); shift < 64; shift *= 2 {
					truth[0] |= truth[0] << uint(shift)
				}
			}
			original := slices.Clone(truth)
			text, largest := DecomposeDSD(truth, n)
			if !slices.Equal(Dau_DsdToTruth(text, n), original) {
				t.Fatalf("n%d trial%d bad %s prime%d", n, trial, text, largest)
			}
			normalized := Dau_DsdNormalize(text)
			if !slices.Equal(Dau_DsdToTruth(normalized, n), original) {
				t.Fatalf("normalize changed %s", text)
			}
			if !slices.Equal(truth, original) {
				t.Fatal("idiomatic decomposition mutated input")
			}
		}
	}
}

//go:embed test_vectors/dsd_oracle.json
var dsdOracle []byte

func TestDSDOriginalC(t *testing.T) {
	var fixtures []struct {
		N, Split, Write, Status int32
		Levels                  []int32
		Truth                   []uint64
		Formula, Perform        string
	}
	if err := json.Unmarshal(dsdOracle, &fixtures); err != nil {
		t.Fatal(err)
	}
	for trial, fixture := range fixtures {
		truth := slices.Clone(fixture.Truth)
		if fixture.N <= 6 {
			if got := Dau_DsdPerform(truth[0]); got != fixture.Perform {
				t.Fatalf("trial%d old perform %s want %s", trial, got, fixture.Perform)
			}
		}
		var result []byte
		status := Dau_DsdDecomposeLevel(truth, fixture.N, fixture.Split, fixture.Write, &result, fixture.Levels)
		if status != fixture.Status || string(result) != fixture.Formula {
			t.Fatalf("trial%d n%d split%d status%d want%d formula %s want %s", trial, fixture.N, fixture.Split, status, fixture.Status, result, fixture.Formula)
		}
		if !slices.Equal(Dau_DsdToTruth(string(result), fixture.N), fixture.Truth) {
			t.Fatalf("trial%d C/native formula changed function %s", trial, result)
		}
	}
}

func TestDSD2TrialExtraction(t *testing.T) {
	formulas := []string{"a", "!a", "(a[bc])", "[(ab)(cd)]", "<a(bc)[de]>", "6996{abcd}", "0", "1"}
	for _, formula := range formulas {
		truth := Dau_DsdToTruth(formula, 6)
		result := Dua_DsdTest(truth, 6)
		if !slices.Equal(Dau_DsdToTruth(result.Expression, 6), truth) {
			t.Fatalf("trial extraction changed %s -> %s", formula, result.Expression)
		}
		if result.nObjs != int32(len(result.pObjs)) {
			t.Fatal("trial object count")
		}
	}
}
func TestDSD2BoundSetFactor(t *testing.T) {
	// F(x0, bound x1..x3, x4) = x0 XOR majority(bound) XOR x4.
	truth := make([]uint64, 1)
	for index := 0; index < 64; index++ {
		bound := (index >> 1) & 7
		majority := bits.OnesCount(uint(bound)) >= 2
		value := (index&1 != 0) != majority
		value = value != (index&16 != 0)
		if value {
			dsSet(truth, index)
		}
	}
	original := slices.Clone(truth)
	partition := Dua_DsdRangeVars(truth, 5, 1, 4, 1)
	if partition != 0xe8 {
		t.Fatalf("bound truth %x", partition)
	}
	for index := 0; index < 32; index++ {
		bit := (partition >> uint((index>>1)&7)) & 1
		replacement := (index & 1) | int(bit)<<1 | (index>>4)<<2
		if dsBit(original, index) != dsBit(truth, replacement) {
			t.Fatal("bound substitution changed function", index)
		}
	}
	low := Dau_DsdToTruth("[<abc>de]", 5)
	if Dua_DsdRangeVars0(low, 5, 3, 0) != 1 {
		t.Fatal("complement cofactor partition")
	}
	Dua_DsdRangeVars0Derive(low, 5, 3)
	if low[0] != Dau_Dsd6ToTruth("[ab]") {
		t.Fatalf("outer XOR selector %x", low[0])
	}
}

func TestDSD2AdjacentSwapOrdering(t *testing.T) {
	// Source comparisons prioritize the numeric word, rather than its lowest
	// differing bit: 01's low bit is smaller than 10's later bit.
	truth := []uint64{1<<1 | 1<<6}
	if got := Dua_DsdTrySwap(truth, 1, 0); got != 0 {
		t.Fatalf("numeric cofactor order returned%d", got)
	}
}

//go:embed test_vectors/dsd2_primitives.json
var dsd2Primitives []byte

func TestDSD2DefinedPrimitivesOriginalC(t *testing.T) {
	var fixtures []struct {
		N, V, Selector, Zero, Opposite, Complement int32
		Truth                                      []uint64
		Half                                       uint64
	}
	if err := json.Unmarshal(dsd2Primitives, &fixtures); err != nil {
		t.Fatal(err)
	}
	for trial, f := range fixtures {
		truth := slices.Clone(f.Truth)
		words := int32(len(truth))
		if got := Abc_Tt6HalfUnShuffleVars(truth[0], f.V%6, f.Selector); got != f.Half {
			t.Fatalf("half%d got%x want%x", trial, got, f.Half)
		}
		if Dua_DsdCheckCof0Const0(truth, words, f.V) != f.Zero || Dua_DsdCheckCofsEqualNot(truth, words, f.V) != f.Opposite || Dua_DsdTryConst(truth, f.N) != f.Complement {
			t.Fatal("C primitive mismatch", trial)
		}
	}
}
func TestDSD2WideComplementBoundSet(t *testing.T) {
	random := rand.New(rand.NewPCG(771, 88))
	for bound := int32(3); bound <= 9; bound++ {
		n := bound + 3
		truth := make([]uint64, dsWords(n))
		inner := make([]bool, 1<<uint(bound))
		for i := range inner {
			inner[i] = random.IntN(2) != 0
		}
		outer := make([]bool, 8)
		for i := range outer {
			outer[i] = random.IntN(2) != 0
		}
		for index := 0; index < 64*len(truth); index++ {
			if inner[index&((1<<uint(bound))-1)] != outer[index>>uint(bound)] {
				dsSet(truth, index)
			}
		}
		if Dua_DsdRangeVars0(truth, n, bound, 0) != 1 {
			t.Fatal("wide bound recognition", bound)
		}
		Dua_DsdRangeVars0Derive(truth, n, bound)
		for index := 0; index < 64; index++ {
			if dsBit(truth, index) != (outer[index&7] != outer[0]) {
				t.Fatal("wide selector", bound, index)
			}
		}
	}
}

func TestDSDNormalizeComposedPrefixStaysOpaque(t *testing.T) {
	input := "<a(cb)d>{(fe)ghij}"
	if got := Dau_DsdNormalize(input); got != "<a(cb)d>{(ef)ghij}" {
		t.Fatalf("normalized composed prefix: %s", got)
	}
}
