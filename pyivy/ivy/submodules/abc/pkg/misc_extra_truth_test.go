package abc

// Go-only unit, invariant, and model tests for extraUtilTruth.c.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func extraTruthModelTable(t testing.TB, variables int, value func(int) bool) *TruthTable {
	t.Helper()
	table, err := NewConstantTruthTable(variables, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		if value(assignment) {
			if err := table.Set(uint64(assignment), true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return table
}

func TestExtraTruthElementaryStretchShrinkAndSwap(t *testing.T) {
	for variables := 0; variables <= 8; variables++ {
		tables, err := ExtraElementaryTruthTables(variables, 0)
		if err != nil || len(tables) != variables {
			t.Fatal(variables, len(tables), err)
		}
		for variable, table := range tables {
			for assignment := 0; assignment < 1<<uint(variables); assignment++ {
				got, getErr := table.Get(uint64(assignment))
				if getErr != nil || got != (assignment&(1<<uint(variable)) != 0) {
					t.Fatal(variables, variable, assignment, got, getErr)
				}
			}
		}
	}
	if _, err := ExtraElementaryTruthTables(-1, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	source := extraTruthModelTable(t, 3, func(a int) bool {
		return a&1 != 0 != (a&2 != 0 && a&4 != 0)
	})
	const phase = uint64(1<<0 | 1<<2 | 1<<5)
	stretched, err := source.ExtraStretchTruth(6, phase)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 64; assignment++ {
		mapped := (assignment & 1) | ((assignment >> 1) & 2) | ((assignment >> 3) & 4)
		got, _ := stretched.Get(uint64(assignment))
		want, _ := source.Get(uint64(mapped))
		if got != want {
			t.Fatal(assignment, mapped, got, want)
		}
	}
	shrunk, err := stretched.ExtraShrinkTruth(phase)
	if err != nil || !shrunk.Equal(source) {
		t.Fatal(shrunk, err)
	}
	if _, err = source.ExtraStretchTruth(5, 3); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = source.ExtraShrinkTruth(3); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	swapped, err := stretched.SwapVariables(2, 5)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 64; assignment++ {
		sourceAssignment := assignment
		first, second := assignment&(1<<2), assignment&(1<<5)
		sourceAssignment &^= 1<<2 | 1<<5
		if first != 0 {
			sourceAssignment |= 1 << 5
		}
		if second != 0 {
			sourceAssignment |= 1 << 2
		}
		got, _ := swapped.Get(uint64(assignment))
		want, _ := stretched.Get(uint64(sourceAssignment))
		if got != want {
			t.Fatal(assignment, got, want)
		}
	}
}

func TestExtraTruthCofactorsQuantificationMuxSymmetryAndOverlap(t *testing.T) {
	table := extraTruthModelTable(t, 4, func(a int) bool {
		return a&4 != 0 || a&1 != 0 && a&2 != 0
	})
	for variable := 0; variable < 4; variable++ {
		zero, err := table.Cofactor(variable, false)
		if err != nil {
			t.Fatal(err)
		}
		one, _ := table.Cofactor(variable, true)
		exist, _ := table.Existential(variable)
		forall, _ := table.Universal(variable)
		recombined, err := ExtraRecombineTruthCofactors(zero, one, variable)
		if err != nil || !recombined.Equal(table) {
			t.Fatal(variable, err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			zeroAssignment := assignment &^ (1 << uint(variable))
			oneAssignment := zeroAssignment | 1<<uint(variable)
			wantZero, _ := table.Get(uint64(zeroAssignment))
			wantOne, _ := table.Get(uint64(oneAssignment))
			for gotTable, want := range map[*TruthTable]bool{zero: wantZero, one: wantOne, exist: wantZero || wantOne, forall: wantZero && wantOne} {
				got, _ := gotTable.Get(uint64(assignment))
				if got != want {
					t.Fatal(variable, assignment, got, want)
				}
			}
		}
	}
	symmetry, err := table.InputSymmetry(0, 1, true)
	if err != nil || symmetry&TruthSymmetryDirect == 0 {
		t.Fatal(symmetry, err)
	}
	xor := extraTruthModelTable(t, 2, func(a int) bool { return bits.OnesCount(uint(a)) == 1 })
	symmetry, err = xor.InputSymmetry(0, 1, true)
	if err != nil || symmetry != TruthSymmetryBoth {
		t.Fatal(symmetry, err)
	}
	overlap, variable, err := table.ExtraMinimumCofactorSupportOverlap()
	if err != nil || overlap != 0 || variable != 2 {
		t.Fatal(overlap, variable, err)
	}
	if _, err = ExtraRecombineTruthCofactors(table, table, 4); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestExtraTruthCountsHashAndSemiCanonicalization(t *testing.T) {
	variable, _ := NewVariableTruthTable(2, 0, 0)
	counts, err := variable.ExtraCofactorCounts()
	if err != nil || !slices.Equal(counts.Zero, []int{0, 8}) || !slices.Equal(counts.One, []int{16, 8}) {
		t.Fatal(counts, err)
	}
	word := uint64(0x0123456789abcdef)
	table, err := TruthTableFromWords(6, []uint64{word}, 0)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := table.ExtraTruthHash()
	wantHash := uint32(2)*uint32(word) ^ uint32(3)*uint32(word>>32)
	if err != nil || hash != wantHash {
		t.Fatal(hash, wantHash, err)
	}
	tooWide, _ := NewConstantTruthTable(16, false, 0)
	if _, err = tooWide.ExtraTruthHash(); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	xnor := extraTruthModelTable(t, 2, func(a int) bool { return a&1 == (a>>1)&1 })
	semi, err := xnor.ExtraSemiCanonicalize()
	if err != nil || !semi.OutputPhase || semi.PackedOnes != 16 || semi.Table.words[0]&1 != 0 {
		t.Fatal(semi, err)
	}
	replayed, err := xnor.TransformNPN(semi.SourceToTarget, semi.InputPhases, semi.OutputPhase)
	if err != nil || !replayed.Equal(semi.Table) {
		t.Fatal(replayed, err)
	}
	if !slices.IsSorted(semi.ZeroCofactorCounts) || !slices.Equal(semi.TargetToSource, []int{0, 1}) {
		t.Fatal(semi.ZeroCofactorCounts, semi.TargetToSource)
	}
}

func FuzzExtraTruthUtilities(f *testing.F) {
	f.Add(uint8(4), []byte{0x96, 0x3c})
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5, 6, 7, 8})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		variables := 1 + int(flags%8)
		assignments := 1 << uint(variables)
		model := make([]bool, assignments)
		for assignment := range model {
			model[assignment] = data[(assignment/8)%len(data)]&(1<<uint(assignment&7)) != 0
		}
		table := extraTruthModelTable(t, variables, func(a int) bool { return model[a] })
		variable := int(flags>>3) % variables
		other := int(flags>>5) % variables
		cofactor := [2]*TruthTable{}
		for value := 0; value < 2; value++ {
			var err error
			cofactor[value], err = table.Cofactor(variable, value != 0)
			if err != nil {
				t.Fatal(err)
			}
		}
		exist, _ := table.Existential(variable)
		forall, _ := table.Universal(variable)
		recombined, err := ExtraRecombineTruthCofactors(cofactor[0], cofactor[1], variable)
		if err != nil || !recombined.Equal(table) {
			t.Fatal(err)
		}
		flipped, _ := table.FlipVariable(variable)
		swapped, _ := table.SwapVariables(variable, other)
		for assignment := range model {
			zeroAssignment := assignment &^ (1 << uint(variable))
			oneAssignment := zeroAssignment | 1<<uint(variable)
			flipAssignment := assignment ^ 1<<uint(variable)
			swapAssignment := assignment
			if ((assignment>>uint(variable))^(assignment>>uint(other)))&1 != 0 {
				swapAssignment ^= 1<<uint(variable) | 1<<uint(other)
			}
			checks := []struct {
				table *TruthTable
				want  bool
			}{
				{cofactor[0], model[zeroAssignment]},
				{cofactor[1], model[oneAssignment]},
				{exist, model[zeroAssignment] || model[oneAssignment]},
				{forall, model[zeroAssignment] && model[oneAssignment]},
				{flipped, model[flipAssignment]},
				{swapped, model[swapAssignment]},
			}
			for _, check := range checks {
				got, getErr := check.table.Get(uint64(assignment))
				if getErr != nil || got != check.want {
					t.Fatal(variables, variable, other, assignment, got, check.want, getErr)
				}
			}
		}
		counts, err := table.ExtraCofactorCounts()
		if err != nil {
			t.Fatal(err)
		}
		scale := 1
		if variables <= 5 {
			scale = 1 << uint(5-variables)
		}
		for input := 0; input < variables; input++ {
			want := [2]int{}
			for assignment, value := range model {
				if value {
					want[assignment>>uint(input)&1] += scale
				}
			}
			if counts.Zero[input] != want[0] || counts.One[input] != want[1] {
				t.Fatal(input, counts, want)
			}
		}
		semi, err := table.ExtraSemiCanonicalize()
		if err != nil {
			t.Fatal(err)
		}
		replayed, err := table.TransformNPN(semi.SourceToTarget, semi.InputPhases, semi.OutputPhase)
		if err != nil || !replayed.Equal(semi.Table) || !slices.IsSorted(semi.ZeroCofactorCounts) {
			t.Fatal(err, semi.ZeroCofactorCounts)
		}
		packed, _ := table.extraPacked32Words()
		ones := 0
		for _, word := range packed {
			ones += bits.OnesCount32(word)
		}
		wantOutputPhase := ones > len(packed)*16 || ones == len(packed)*16 && packed[0]&1 != 0
		wantPackedOnes := ones
		if wantOutputPhase {
			wantPackedOnes = len(packed)*32 - ones
		}
		if semi.OutputPhase != wantOutputPhase || semi.PackedOnes != wantPackedOnes {
			t.Fatal("output phase decision", semi.OutputPhase, wantOutputPhase, semi.PackedOnes, wantPackedOnes)
		}
	})
}
