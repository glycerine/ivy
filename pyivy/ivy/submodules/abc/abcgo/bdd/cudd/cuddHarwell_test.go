package cudd

import (
	"strings"
	"testing"
)

func TestHarwellReadsAssembledMatrix(t *testing.T) {
	manager := Cudd_Init(0, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0)
	input := "small matrix                 small01\n" +
		"5 1 1 1 0\n" +
		"RUA 2 2 2 0\n" +
		"(3I3) (2I3) (2E10.2)\n" +
		"1 2 3\n" +
		"1 2\n" +
		"2.0 3.0\n"
	got, err := Cudd_addHarwell(strings.NewReader(input), manager, 0, 1, 2, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Rows != 2 || got.Cols != 2 || got.RightHandSides != 0 {
		t.Fatalf("matrix dimensions = %dx%d rhs=%d", got.Rows, got.Cols, got.RightHandSides)
	}
	if got.Function == nil || len(got.X) != 1 || len(got.Y) != 1 || len(got.XNeg) != 1 || len(got.YNeg) != 1 {
		t.Fatalf("incomplete ADD result: %+v", got)
	}
}

func TestHarwellRejectsUnsupportedType(t *testing.T) {
	manager := Cudd_Init(0, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0)
	input := "matrix\n1 1 1 1 0\nCUA 1 1 1 0\nfmt\n1 2\n1\n1.0\n"
	if _, err := Cudd_addHarwell(strings.NewReader(input), manager, 0, 1, 1, 1, nil); err == nil {
		t.Fatal("expected unsupported matrix type error")
	}
}
