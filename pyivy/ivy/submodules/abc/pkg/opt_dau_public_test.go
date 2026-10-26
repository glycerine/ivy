package abc

// Go-only checks for the public DAU constants and DSD token predicates.

import (
	"errors"
	"testing"
)

func TestDauPublicConstantsAndKinds(t *testing.T) {
	if DauMaxVariables != 12 || DauMaxFormulaBytes != 2000 || DauMaxTruthWords != 64 {
		t.Fatal(DauMaxVariables, DauMaxFormulaBytes, DauMaxTruthWords)
	}
	want := []DauDSDNodeKind{DauDSDNone, DauDSDConstant, DauDSDVariable, DauDSDAnd, DauDSDXor, DauDSDMux, DauDSDPrime}
	for value, kind := range want {
		if int(kind) != value {
			t.Fatal(value, kind)
		}
	}
}

func TestDauDSDTokenPredicates(t *testing.T) {
	for _, test := range []struct {
		text                          string
		constant, zero, one, variable bool
		index                         int
	}{
		{"0", true, true, false, false, -1},
		{"1", true, false, true, false, -1},
		{"a", false, false, false, true, 0},
		{"!z", false, false, false, true, 25},
		{"a trailing", false, false, false, true, 0},
		{"", false, false, false, false, -1},
		{"!", false, false, false, false, -1},
		{"A", false, false, false, false, -1},
	} {
		if DauDSDIsConstant(test.text) != test.constant || DauDSDIsConstantZero(test.text) != test.zero || DauDSDIsConstantOne(test.text) != test.one || DauDSDIsVariable(test.text) != test.variable {
			t.Fatal(test)
		}
		index, err := DauDSDReadVariable(test.text)
		if test.index < 0 {
			if !errors.Is(err, ErrDimension) {
				t.Fatal(test.text, index, err)
			}
		} else if err != nil || index != test.index {
			t.Fatal(test.text, index, err)
		}
	}
}
