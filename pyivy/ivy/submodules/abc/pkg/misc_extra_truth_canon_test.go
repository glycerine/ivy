package abc

// Go-only exhaustive and model tests for extraUtilCanon.c.

import (
	"errors"
	"slices"
	"testing"
)

func extraPhaseTruthModel(truth uint64, variables, phase int) uint64 {
	var result uint64
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		if truth&(uint64(1)<<uint(assignment^phase)) != 0 {
			result |= uint64(1) << uint(assignment)
		}
	}
	return result
}

func extraTruthCanonModel(truth uint64, realVariables int) (uint64, []uint8) {
	enumerated := max(realVariables, 3)
	best := ^uint64(0)
	phases := []uint8{}
	for phase := 0; phase < 1<<uint(enumerated); phase++ {
		candidate := extraPhaseTruthModel(truth, realVariables, phase&((1<<uint(realVariables))-1))
		if candidate < best {
			best = candidate
			phases = phases[:0]
			phases = append(phases, uint8(phase))
		} else if candidate == best {
			phases = append(phases, uint8(phase))
		}
	}
	return best, phases
}

func extraCanonTable(t testing.TB, maxVariables, realVariables int, truth uint64) *TruthTable {
	t.Helper()
	table, err := NewConstantTruthTable(maxVariables, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 1<<uint(maxVariables); assignment++ {
		realAssignment := assignment & ((1 << uint(realVariables)) - 1)
		if truth&(uint64(1)<<uint(realAssignment)) != 0 {
			if err = table.Set(uint64(assignment), true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return table
}

func TestExtraTruthCanonThreeVariableTablesExhaustive(t *testing.T) {
	rows, err := ExtraTruth3CanonicalRows(ExtraTruthNCanonicalOptions{})
	if err != nil || len(rows) != 256 {
		t.Fatal(len(rows), err)
	}
	for truth, row := range rows {
		want, phases := extraTruthCanonModel(uint64(truth), 3)
		if row.Truth != uint8(truth) || row.Canonical != uint8(want) || !slices.Equal(row.Phases, phases) {
			t.Fatal(truth, row, want, phases)
		}
	}
	rows[0].Phases[0] = 99
	again, err := ExtraTruth3CanonicalRows(ExtraTruthNCanonicalOptions{})
	if err != nil || again[0].Phases[0] != 0 {
		t.Fatal("ownership", err)
	}
}

func TestExtraTruthCanonFourVariableExhaustive(t *testing.T) {
	for truth := 0; truth < 1<<16; truth++ {
		table := extraCanonTable(t, 4, 4, uint64(truth))
		got, err := ExtraTruthCanonFastN(table, 4, ExtraTruthNCanonicalOptions{})
		if err != nil {
			t.Fatal(truth, err)
		}
		want, phases := extraTruthCanonModel(uint64(truth), 4)
		if uint16(got.Table.words[0]) != uint16(want) || !slices.Equal(got.Phases, phases) || got.Candidates != 16 || got.EnumeratedVariables != 4 {
			t.Fatal(truth, got.Table.words[0], want, got.Phases, phases)
		}
	}
}

func TestExtraTruthCanonEmbeddingLimitsCancellationAndDimensions(t *testing.T) {
	table := extraCanonTable(t, 6, 4, 0x963c)
	result, err := ExtraTruthCanonFastN(table, 4, ExtraTruthNCanonicalOptions{})
	if err != nil || result.Table.Variables() != 6 || result.RealVariables != 4 {
		t.Fatal(result, err)
	}
	want, phases := extraTruthCanonModel(0x963c, 4)
	for assignment := 0; assignment < 64; assignment++ {
		got, _ := result.Table.Get(uint64(assignment))
		wantBit := want&(uint64(1)<<uint(assignment&15)) != 0
		if got != wantBit {
			t.Fatal(assignment, got, wantBit)
		}
	}
	if !slices.Equal(result.Phases, phases) {
		t.Fatal(result.Phases, phases)
	}
	result.Phases[0] = 99
	again, _ := ExtraTruthCanonFastN(table, 4, ExtraTruthNCanonicalOptions{})
	if again.Phases[0] == 99 || again.Table == result.Table {
		t.Fatal("ownership")
	}
	activeHigh, _ := NewVariableTruthTable(6, 5, 0)
	if _, err = ExtraTruthCanonFastN(activeHigh, 4, ExtraTruthNCanonicalOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if _, err = ExtraTruthCanonFastN(table, 4, ExtraTruthNCanonicalOptions{MaxCandidates: 15}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = ExtraTruthCanonFastN(table, 4, ExtraTruthNCanonicalOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err = ExtraTruthCanonFastN(table, 4, ExtraTruthNCanonicalOptions{MaxCandidates: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	for _, real := range []int{-1, 7} {
		if _, err = ExtraTruthCanonFastN(table, real, ExtraTruthNCanonicalOptions{}); !errors.Is(err, ErrDimension) {
			t.Fatal(real, err)
		}
	}
	wide, _ := NewConstantTruthTable(7, false, 0)
	if _, err = ExtraTruthCanonFastN(wide, 7, ExtraTruthNCanonicalOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = ExtraTruth3CanonicalRows(ExtraTruthNCanonicalOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func FuzzExtraTruthCanonFastN(f *testing.F) {
	f.Add(uint8(6), uint64(0x6996966996696996))
	f.Add(uint8(1), uint64(2))
	f.Fuzz(func(t *testing.T, width uint8, truth uint64) {
		variables := int(width % 7)
		if variables < 6 {
			truth &= (uint64(1) << uint(1<<uint(variables))) - 1
		}
		table := extraCanonTable(t, variables, variables, truth)
		result, err := ExtraTruthCanonFastN(table, variables, ExtraTruthNCanonicalOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want, phases := extraTruthCanonModel(truth, variables)
		if !slices.Equal(result.Phases, phases) {
			t.Fatal(variables, truth, result.Phases, phases)
		}
		for assignment := 0; assignment < 1<<uint(variables); assignment++ {
			got, getErr := result.Table.Get(uint64(assignment))
			wantBit := want&(uint64(1)<<uint(assignment)) != 0
			if getErr != nil || got != wantBit {
				t.Fatal(variables, truth, assignment, got, wantBit, getErr)
			}
		}
		for _, phase := range result.Phases {
			if extraPhaseTruthModel(truth, variables, int(phase)&((1<<uint(variables))-1)) != want {
				t.Fatal("invalid minimizing phase", variables, truth, phase)
			}
		}
	})
}
