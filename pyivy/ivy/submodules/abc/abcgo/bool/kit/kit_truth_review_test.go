package kit

import (
	"math/rand"
	"os"
	"slices"
	"testing"
)

func TestCofactorCountDoesNotModifyTruth(t *testing.T) {
	for _, truth := range [][]uint32{{0xaaaaaaaa}, {0x01234567, 0x89abcdef}} {
		n := int32(5)
		if len(truth) > 1 {
			n = 6
		}
		original := slices.Clone(truth)
		for v := int32(0); v < n; v++ {
			var expected int32
			for x := int32(0); x < 32*int32(len(truth)); x++ {
				if x&(1<<uint(v)) == 0 && Kit_TruthHasBit(truth, x) != 0 {
					expected++
				}
			}
			if got := Kit_TruthCofactor0Count(truth, n, v); got != expected {
				t.Fatalf("v=%d got=%d expected=%d", v, got, expected)
			}
			if !slices.Equal(truth, original) {
				t.Fatal("cofactor counting changed input")
			}
		}
	}
}

func TestTruthVariableTransforms(t *testing.T) {
	rng := rand.New(rand.NewSource(91))
	for n := int32(2); n <= 10; n++ {
		input := make([]uint32, Kit_TruthWordNum(n))
		for i := range input {
			input[i] = rng.Uint32()
		}
		if n < 5 {
			mask := Kit_BitMask(1 << uint(n))
			input[0] &= mask
			for shift := uint(1) << uint(n); shift < 32; shift *= 2 {
				input[0] |= input[0] << shift
			}
		}
		for v := int32(0); v < n; v++ {
			low, high := make([]uint32, len(input)), make([]uint32, len(input))
			Kit_TruthCofactor0New(low, input, n, v)
			Kit_TruthCofactor1New(high, input, n, v)
			for x := int32(0); x < int32(len(input))*32; x++ {
				if Kit_TruthHasBit(low, x) != Kit_TruthHasBit(input, x&^(1<<uint(v))) || Kit_TruthHasBit(high, x) != Kit_TruthHasBit(input, x|1<<uint(v)) {
					t.Fatal("cofactor", n, v, x)
				}
			}
			reconstructed := make([]uint32, len(input))
			Kit_TruthMuxVar(reconstructed, low, high, n, v)
			if !slices.Equal(reconstructed, input) {
				t.Fatal("cofactor recomposition", n, v)
			}
			inPlace := slices.Clone(input)
			Kit_TruthCofactor1(inPlace, n, v)
			if !slices.Equal(inPlace, high) {
				t.Fatal("aliased cofactor")
			}
			Kit_TruthChangePhase(inPlace, n, v)
			Kit_TruthChangePhase(inPlace, n, v)
			if !slices.Equal(inPlace, high) {
				t.Fatal("phase involution")
			}
			if v < n-1 {
				Kit_TruthSwapAdjacentVars(inPlace, input, n, v)
				Kit_TruthSwapAdjacentVars(inPlace, inPlace, n, v)
				if !slices.Equal(inPlace, input) {
					t.Fatal("swap involution")
				}
			}
			Kit_TruthExistNew(inPlace, input, n, v)
			expected := make([]uint32, len(input))
			Kit_TruthOr(expected, low, high, n)
			if !slices.Equal(inPlace, expected) {
				t.Fatal("existential quantification")
			}
		}
	}
}
func TestTruthSupportStretchShrink(t *testing.T) {
	original := make([]uint32, 8)
	for i := range original {
		original[i] = 0x66666666
	}
	stretched, shrunk := make([]uint32, 8), make([]uint32, 8)
	phase := uint32(1<<3 | 1<<7)
	Kit_TruthStretch(stretched, original, 2, 8, phase, 0)
	if support := Kit_TruthSupport(stretched, 8); support != phase {
		t.Fatal("stretched support", support)
	}
	Kit_TruthShrink(shrunk, stretched, 2, 8, phase, 0)
	if !slices.Equal(original, shrunk) {
		t.Fatal("support shrink round trip")
	}
}
func TestSemiCanonicalTransformMetadata(t *testing.T) {
	rng := rand.New(rand.NewSource(29))
	for trial := 0; trial < 100; trial++ {
		original := []uint32{rng.Uint32()}
		truth, aux := slices.Clone(original), make([]uint32, 1)
		permutation := make([]byte, 5)
		phase := Kit_TruthSemiCanonicize(truth, aux, 5, permutation)
		for x := 0; x < 32; x++ {
			old := 0
			for v, p := range permutation {
				value := x >> uint(v) & 1
				if phase&(1<<uint(v)) != 0 {
					value ^= 1
				}
				old |= value << p
			}
			want := Kit_TruthHasBit(original, int32(old))
			if phase&(1<<5) != 0 {
				want ^= 1
			}
			if got := Kit_TruthHasBit(truth, int32(x)); got != want {
				t.Fatalf("trial=%d x=%d phase=%x permutation=%v", trial, x, phase, permutation)
			}
		}
	}
}

func TestTruthDumpProducesAbcScript(t *testing.T) {
	t.Chdir(t.TempDir())
	name := Kit_TruthDumpToFile([]uint32{0x88888888}, 2, 7)
	data, err := os.ReadFile(name)
	if err != nil || string(data) != "rt 8; bdd; sop; ps\n" {
		t.Fatal(name, string(data), err)
	}
}
