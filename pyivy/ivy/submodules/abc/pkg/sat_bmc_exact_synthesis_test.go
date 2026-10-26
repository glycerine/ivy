package abc

// Go-only end-to-end checks for the bmcMaj.c CEGAR loops.

import (
	"errors"
	"slices"
	"testing"
)

func TestSynthesizeExactMajority(t *testing.T) {
	result, err := SynthesizeExactMajority(3, 1, false, false, ExactSynthesisOptions{})
	if err != nil || result.Status != SatSatisfiable || result.Network == nil || result.Iterations == 0 || result.Variables == 0 || result.Clauses == 0 {
		t.Fatal(result, err)
	}
	got, err := result.Network.Truth(ExactSynthesisLimits{})
	want, _ := ExactMajorityTruth(3, ExactSynthesisLimits{})
	if err != nil || !slices.Equal(got, want) {
		t.Fatal(got, want, err)
	}
	seen := make([]bool, len(want))
	for _, minterm := range result.Minterms {
		if minterm < 0 || minterm >= len(seen) || seen[minterm] {
			t.Fatal(result.Minterms)
		}
		seen[minterm] = true
	}
}

func TestSynthesizeExactLUTAndComplement(t *testing.T) {
	topology, _ := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
	for _, truthBits := range []uint64{0b0110, 0b1110, 0b0001} {
		target := exactTruthBits(2, truthBits)
		result, err := SynthesizeExactNetwork(topology, target, ExactSynthesisOptions{})
		if err != nil || result.Status != SatSatisfiable || result.Network == nil || result.ComplementedTarget != target[0] {
			t.Fatal(truthBits, result, err)
		}
		got, truthErr := result.Network.Truth(ExactSynthesisLimits{})
		if truthErr != nil || !slices.Equal(got, target) {
			t.Fatal(truthBits, got, target, truthErr)
		}
	}
}

func TestSynthesizeExactUnsatLimitAndCancel(t *testing.T) {
	topology, _ := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
	and := exactTruthBits(2, 0b1000)
	result, err := SynthesizeExactNetwork(topology, and, ExactSynthesisOptions{})
	if err != nil || result.Status != SatUnsatisfiable || result.Network != nil {
		t.Fatal(result, err)
	}
	limited, err := SynthesizeExactNetwork(topology, exactTruthBits(2, 0b0110), ExactSynthesisOptions{MaxIterations: 1})
	if err != nil || limited.Status != SatUnknown || !limited.IterationLimitReached || limited.Network != nil || limited.Iterations != 1 {
		t.Fatal(limited, err)
	}
	if result, err = SynthesizeExactNetwork(topology, exactTruthBits(2, 0b0110), ExactSynthesisOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) || result.Network != nil {
		t.Fatal(result, err)
	}
}

func FuzzSynthesizeExactLUT(f *testing.F) {
	for _, seed := range []uint8{0, 2, 6, 14} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw uint8) {
		truthBits := uint64(raw & 14)
		topology, err := BuildLUTExactTopology(2, 1, 2, false, ExactSynthesisLimits{})
		if err != nil {
			t.Fatal(err)
		}
		target := exactTruthBits(2, truthBits)
		result, err := SynthesizeExactNetwork(topology, target, ExactSynthesisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		f1, f2, f3 := truthBits&2 != 0, truthBits&4 != 0, truthBits&8 != 0
		allowed := (f1 || f2 || !f3) && (!f1 || f2 || !f3) && (f1 || !f2 || !f3)
		if allowed != (result.Status == SatSatisfiable) {
			t.Fatal(truthBits, allowed, result.Status)
		}
		if allowed {
			got, err := result.Network.Truth(ExactSynthesisLimits{})
			if err != nil || !slices.Equal(got, target) {
				t.Fatal(truthBits, got, target, err)
			}
		}
	})
}
