package abc

// Go-only exhaustive and fuzz checks for six-variable DSD decomposition.

import (
	"errors"
	"testing"
)

func checkDauDSD6Decomposition(t *testing.T, truth uint64) {
	t.Helper()
	formula, err := DauDSDFromTruth6(truth, DauDSDOptions{})
	if err != nil {
		t.Fatalf("truth %016x: %v", truth, err)
	}
	parsed, err := ParseDauDSD(formula, DauDSDOptions{})
	if err != nil {
		t.Fatalf("truth %016x formula %q: %v", truth, formula, err)
	}
	actual, err := parsed.TruthTable(6, 1, nil)
	if err != nil {
		t.Fatalf("truth %016x formula %q: %v", truth, formula, err)
	}
	if got := actual.Words()[0]; got != truth {
		t.Fatalf("truth %016x formula %q evaluates to %016x", truth, formula, got)
	}
}

func TestDauDSDFromTruth6Known(t *testing.T) {
	for _, truth := range []uint64{
		0, ^uint64(0), truthVariablePattern(0), ^truthVariablePattern(0),
		truthVariablePattern(0) & truthVariablePattern(1),
		truthVariablePattern(0) ^ truthVariablePattern(1),
		(truthVariablePattern(0) & truthVariablePattern(1)) | (^truthVariablePattern(0) & truthVariablePattern(2)),
		0x9ef7a8d9c7193a0f,
	} {
		checkDauDSD6Decomposition(t, truth)
	}
	if got, _ := DauDSDFromTruth6(0, DauDSDOptions{}); got != "0" {
		t.Fatalf("zero=%q", got)
	}
	if got, _ := DauDSDFromTruth6(^uint64(0), DauDSDOptions{}); got != "1" {
		t.Fatalf("one=%q", got)
	}
	if got, _ := DauDSDFromTruth6(truthVariablePattern(0), DauDSDOptions{}); got != "a" {
		t.Fatalf("variable=%q", got)
	}
}

func TestDauDSDFromTruth6ExhaustiveFourVariables(t *testing.T) {
	for small := uint64(0); small < 1<<16; small++ {
		truth := small | small<<16 | small<<32 | small<<48
		checkDauDSD6Decomposition(t, truth)
	}
}

func TestDauDSDFromTruth6Limits(t *testing.T) {
	truth := uint64(0x9ef7a8d9c7193a0f)
	for _, options := range []DauDSDOptions{{MaxBytes: 2}, {MaxDepth: 1}} {
		if result, err := DauDSDFromTruth6(truth, options); result != "" || !errors.Is(err, ErrLimit) {
			t.Fatalf("options %+v: %q, %v", options, result, err)
		}
	}
	if result, err := DauDSDFromTruth6(truth, DauDSDOptions{Stop: func() bool { return true }}); result != "" || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %q, %v", result, err)
	}
	if _, err := DauDSDFromTruth6(truth, DauDSDOptions{MaxDepth: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative: %v", err)
	}
}

func FuzzDauDSDFromTruth6(f *testing.F) {
	for _, truth := range []uint64{0, 1, ^uint64(0), 0xaaaaaaaaaaaaaaaa, 0x9ef7a8d9c7193a0f} {
		f.Add(truth)
	}
	f.Fuzz(func(t *testing.T, truth uint64) { checkDauDSD6Decomposition(t, truth) })
}
