package abc

// Go-only dataset tests for bmcMaj3's compact exact-solution verifier.

import (
	"errors"
	"os"
	"slices"
	"testing"
)

func TestParseExactSolutionName(t *testing.T) {
	metadata, err := ParseExactSolutionName("6-2-1.bool", ExactSynthesisLimits{})
	if err != nil || metadata.Inputs != 2 || metadata.LUTSize != 2 || metadata.Nodes != 1 || !slices.Equal(metadata.Target, []bool{false, true, true, false}) {
		t.Fatal(metadata, err)
	}
	metadata, err = ParseExactSolutionName("80000000000000000000000000000000-3-4.bool", ExactSynthesisLimits{})
	if err != nil || metadata.Inputs != 7 || len(metadata.Target) != 128 || !metadata.Target[127] || metadata.Target[0] {
		t.Fatal(metadata.Inputs, len(metadata.Target), err)
	}
	for _, name := range []string{"", "66-2.bool", "6-x-1.bool", "6-2-x.bool", "6-1-1.bool", "6-2-30.bool", "6-2-1.txt"} {
		if _, err := ParseExactSolutionName(name, ExactSynthesisLimits{}); !errors.Is(err, ErrFormat) {
			t.Fatal(name, err)
		}
	}
}

func TestExactSolutionMarshalVerifyAndPermanentVector(t *testing.T) {
	network := ExactNetwork{Inputs: 2, Gates: []ExactGate{{Fanins: []int{0, 1}, Truth: 0b0110}}}
	encoded, err := MarshalExactSolutionNetwork(network, ExactSynthesisLimits{})
	if err != nil || string(encoded) != "C0110ab\n\n" {
		t.Fatal(string(encoded), err)
	}
	combined, err := MarshalExactSolutionDataset([]ExactNetwork{network, network.Clone()}, ExactSynthesisLimits{})
	verified, verifyErr := VerifyExactSolutionDataset("6-2-1.bool", combined, ExactSynthesisLimits{})
	if err != nil || verifyErr != nil || len(verified) != 2 {
		t.Fatal(string(combined), len(verified), err, verifyErr)
	}
	data, err := os.ReadFile("test_vectors/bmcMaj3/6-2-1.bool")
	if err != nil {
		t.Fatal(err)
	}
	verified, err = VerifyExactSolutionDataset("6-2-1.bool", data, ExactSynthesisLimits{})
	if err != nil || len(verified) != 1 {
		t.Fatal(len(verified), err)
	}
	truth, _ := verified[0].Truth(ExactSynthesisLimits{})
	if !slices.Equal(truth, []bool{false, true, true, false}) {
		t.Fatal(truth)
	}
}

func TestExactSolutionVerifierErrorsLimitsCancellation(t *testing.T) {
	valid := []byte("C0110ab\n\n")
	for _, malformed := range [][]byte{
		[]byte("C0110ab\n"), []byte("D0110ab\n\n"), []byte("C0210ab\n\n"),
		[]byte("C0110aa\n\n"), []byte("C0110ac\n\n"), []byte("C1000ab\n\n"),
	} {
		if _, err := VerifyExactSolutionDataset("6-2-1.bool", malformed, ExactSynthesisLimits{}); err == nil {
			t.Fatalf("accepted %q", malformed)
		}
	}
	if _, err := VerifyExactSolutionDataset("6-2-1.bool", valid, ExactSynthesisLimits{MaxBytes: 4}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := VerifyExactSolutionDataset("6-2-1.bool", valid, ExactSynthesisLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	complement := ExactNetwork{Inputs: 2, Gates: []ExactGate{{Fanins: []int{0, 1}, Truth: 6}}, ComplementOutput: true}
	encoded, err := MarshalExactSolutionNetwork(complement, ExactSynthesisLimits{})
	if err != nil || string(encoded) != "C1001ab\n\n" {
		t.Fatal(string(encoded), err)
	}
	if _, err = VerifyExactSolutionDataset("9-2-1.bool", encoded, ExactSynthesisLimits{}); err != nil {
		t.Fatal(err)
	}
}

func FuzzExactSolutionDataset(f *testing.F) {
	for _, seed := range [][]byte{[]byte("C0110ab\n\n"), nil, []byte("C1000ab\n\n"), []byte("D0110ab\n\n")} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			return
		}
		networks, err := VerifyExactSolutionDataset("6-2-1.bool", data, ExactSynthesisLimits{MaxBytes: 64})
		if err != nil {
			return
		}
		for _, network := range networks {
			truth, truthErr := network.Truth(ExactSynthesisLimits{})
			if truthErr != nil || !slices.Equal(truth, []bool{false, true, true, false}) {
				t.Fatal(data, truth, truthErr)
			}
		}
	})
}
