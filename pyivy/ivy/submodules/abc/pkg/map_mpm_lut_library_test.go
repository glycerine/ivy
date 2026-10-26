package abc

// Tests for simple LUT libraries from map/mpm/mpmLib.c.

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

type mpmLUTLibraryOracleRow struct {
	Size, Maximum, Variable int
	Areas                   [MPMMaximumVariables + 1]int
	Delays                  [MPMMaximumVariables + 1][MPMMaximumVariables + 1]int
}

func TestMPMLUTLibraryNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mpm_lut_library.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []mpmLUTLibraryOracleRow
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != MPMMaximumVariables+1 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		library, err := NewSimpleMPMLUTLibrary(row.Size)
		if err != nil || library.MaximumLUTSize != row.Maximum || library.VariablePinDelays != (row.Variable != 0) || library.Areas != row.Areas || library.Delays != row.Delays || library.Name != "" {
			t.Fatal(row.Size, library, row, err)
		}
	}
}

func TestMPMLUTLibraryErrorsAndOwnership(t *testing.T) {
	for _, size := range []int{-1, MPMMaximumVariables + 1} {
		if library, err := NewSimpleMPMLUTLibrary(size); library != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(size, library, err)
		}
	}
	first, err := NewSimpleMPMLUTLibrary(4)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewSimpleMPMLUTLibrary(4)
	if err != nil {
		t.Fatal(err)
	}
	first.Name, first.Areas[1], first.Delays[4][3] = "changed", 99, 99
	if second.Name != "" || second.Areas[1] != MPMUnitArea || second.Delays[4][3] != MPMUnitTime {
		t.Fatal(second)
	}
}

func FuzzMPMLUTLibrary(f *testing.F) {
	for size := 0; size <= MPMMaximumVariables; size++ {
		f.Add(uint8(size))
	}
	f.Fuzz(func(t *testing.T, raw uint8) {
		size := int(raw % (MPMMaximumVariables + 1))
		library, err := NewSimpleMPMLUTLibrary(size)
		if err != nil {
			t.Fatal(err)
		}
		for inputs := 0; inputs <= MPMMaximumVariables; inputs++ {
			wantArea := 0
			if inputs > 0 && inputs <= size {
				wantArea = MPMUnitArea
			}
			if library.Areas[inputs] != wantArea {
				t.Fatal(size, inputs, library.Areas[inputs], wantArea)
			}
			for pin := 0; pin <= MPMMaximumVariables; pin++ {
				wantDelay := 0
				if inputs > 0 && inputs <= size && pin < inputs {
					wantDelay = MPMUnitTime
				}
				if library.Delays[inputs][pin] != wantDelay {
					t.Fatal(size, inputs, pin, library.Delays[inputs][pin], wantDelay)
				}
			}
		}
	})
}
