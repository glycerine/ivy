package dau

import (
	_ "embed"
	"encoding/json"
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
		Truth                   []uint64
		Formula                 string
	}
	if err := json.Unmarshal(dsdOracle, &fixtures); err != nil {
		t.Fatal(err)
	}
	for trial, fixture := range fixtures {
		truth := slices.Clone(fixture.Truth)
		var result []byte
		status := Dau_DsdDecompose(truth, fixture.N, fixture.Split, fixture.Write, &result)
		if status != fixture.Status || string(result) != fixture.Formula {
			t.Fatalf("trial%d n%d split%d status%d want%d formula %s want %s", trial, fixture.N, fixture.Split, status, fixture.Status, result, fixture.Formula)
		}
		if !slices.Equal(Dau_DsdToTruth(string(result), fixture.N), fixture.Truth) {
			t.Fatalf("trial%d C/native formula changed function %s", trial, result)
		}
	}
}
