package abc

// Go-only group, layer, and fuzz tests for extraUtilCube.c.

import (
	"errors"
	"os"
	"slices"
	"testing"
)

func extraCubeValidPermutation(state ExtraCubeState) bool {
	seen := [24]bool{}
	for _, value := range state {
		if value >= 24 || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func TestExtraCubeGeneratorsCompositionSignatureAndFormat(t *testing.T) {
	identity := ExtraCubeIdentity()
	if !extraCubeValidPermutation(identity) {
		t.Fatal(identity)
	}
	generators := ExtraCubeGenerators()
	for axis := 0; axis < 3; axis++ {
		state := identity
		for power := 1; power <= 4; power++ {
			var err error
			state, err = PermuteExtraCubeState(state, generators[axis])
			if err != nil || !extraCubeValidPermutation(state) {
				t.Fatal(axis, power, state, err)
			}
			if power < 4 && state != generators[3*(power-1)+axis] {
				t.Fatal("power", axis, power)
			}
		}
		if state != identity {
			t.Fatal("fourth power", axis, state)
		}
	}
	signatures := map[uint64]struct{}{}
	for _, state := range append([]ExtraCubeState{identity}, generators[:]...) {
		signature, err := ExtraCubeSignature(state)
		if err != nil {
			t.Fatal(err)
		}
		if _, duplicate := signatures[signature]; duplicate {
			t.Fatal("duplicate signature", signature)
		}
		signatures[signature] = struct{}{}
	}
	wantText := "  0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16 17 18 19 20 21 22 23\n"
	if got := FormatExtraCubeState(identity); got != wantText {
		t.Fatalf("%q", got)
	}
	bad := identity
	bad[1] = bad[0]
	if _, err := PermuteExtraCubeState(identity, bad); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	bad = identity
	bad[0] = 18
	if _, err := ExtraCubeSignature(bad); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestExtraCubeEnumerationKnownLayersAndModes(t *testing.T) {
	want := []int{1, 10, 64, 385, 2232, 12224}
	for _, signatureOnly := range []bool{false, true} {
		result, err := EnumerateExtraCubeStates(ExtraCubeEnumerationOptions{
			MaxStates: 20000, MaxDepth: 5, SignatureOnly: signatureOnly, ReturnStates: true,
		})
		if !errors.Is(err, ErrLimit) || result.Converged || !slices.Equal(result.LayerTotals, want) || len(result.States) != want[len(want)-1] {
			t.Fatal(signatureOnly, result.LayerTotals, len(result.States), result.Converged, err)
		}
		for _, state := range result.States {
			if !extraCubeValidPermutation(state) {
				t.Fatal("invalid state", state)
			}
		}
		result.States[0][0] = 99
		again, _ := EnumerateExtraCubeStates(ExtraCubeEnumerationOptions{MaxStates: 100, MaxDepth: 2, ReturnStates: true})
		if again.States[0] != ExtraCubeIdentity() {
			t.Fatal("ownership")
		}
	}
}

func TestExtraCubeFullGroup(t *testing.T) {
	if os.Getenv("ABC_FULL_CUBE") != "1" {
		t.Skip("set ABC_FULL_CUBE=1 to enumerate all 3,674,160 states")
	}
	want := []int{1, 10, 64, 385, 2232, 12224, 62360, 289896, 1159968, 3047716, 3671516, 3674160, 3674160}
	result, err := EnumerateExtraCubeStates(ExtraCubeEnumerationOptions{SignatureOnly: true})
	if err != nil || !result.Converged || !slices.Equal(result.LayerTotals, want) {
		t.Fatal(result.LayerTotals, result.Converged, err)
	}
}

func TestExtraCubeEnumerationLimitsAndCancellation(t *testing.T) {
	if result, err := EnumerateExtraCubeStates(ExtraCubeEnumerationOptions{MaxStates: 1, ReturnStates: true}); !errors.Is(err, ErrLimit) || len(result.States) != 1 || !slices.Equal(result.LayerTotals, []int{1}) {
		t.Fatal(result, err)
	}
	if _, err := EnumerateExtraCubeStates(ExtraCubeEnumerationOptions{MaxStates: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := EnumerateExtraCubeStates(ExtraCubeEnumerationOptions{MaxDepth: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	calls := 0
	result, err := EnumerateExtraCubeStates(ExtraCubeEnumerationOptions{ReturnStates: true, Stop: func() bool {
		calls++
		return calls == 5
	}})
	if !errors.Is(err, ErrCancelled) || len(result.States) != 4 {
		t.Fatal(calls, len(result.States), err)
	}
}

func FuzzExtraCubeEnumeration(f *testing.F) {
	f.Add(uint8(3), []byte{0, 1, 2, 8})
	f.Add(uint8(1), []byte{7})
	f.Fuzz(func(t *testing.T, flags uint8, moves []byte) {
		if len(moves) > 64 {
			moves = moves[:64]
		}
		generators := ExtraCubeGenerators()
		state := ExtraCubeIdentity()
		for _, move := range moves {
			var err error
			state, err = PermuteExtraCubeState(state, generators[int(move)%len(generators)])
			if err != nil || !extraCubeValidPermutation(state) {
				t.Fatal(move, state, err)
			}
			if _, err = ExtraCubeSignature(state); err != nil {
				t.Fatal(err)
			}
		}
		depth := 1 + int(flags%4)
		full, fullErr := EnumerateExtraCubeStates(ExtraCubeEnumerationOptions{MaxStates: 5000, MaxDepth: depth, ReturnStates: true})
		signatures, signatureErr := EnumerateExtraCubeStates(ExtraCubeEnumerationOptions{MaxStates: 5000, MaxDepth: depth, SignatureOnly: true, ReturnStates: true})
		if !errors.Is(fullErr, ErrLimit) || !errors.Is(signatureErr, ErrLimit) || !slices.Equal(full.LayerTotals, signatures.LayerTotals) || len(full.States) != len(signatures.States) {
			t.Fatal(depth, full.LayerTotals, signatures.LayerTotals, fullErr, signatureErr)
		}
		seen := make(map[uint64]struct{}, len(full.States))
		for _, cube := range full.States {
			signature, err := ExtraCubeSignature(cube)
			if err != nil {
				t.Fatal(err)
			}
			if _, duplicate := seen[signature]; duplicate {
				t.Fatal("signature collision", depth, signature)
			}
			seen[signature] = struct{}{}
		}
	})
}
