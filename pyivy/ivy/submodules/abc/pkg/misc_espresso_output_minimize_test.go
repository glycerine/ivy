package abc

// Go-only per-output truth models for src/misc/espresso/cvrm.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoPLAFromOutputTruths(t testing.TB, variables int, truths []uint64) *EspressoPLA {
	t.Helper()
	parts := append(slices.Repeat([]int{2}, variables), len(truths))
	layout, _ := NewEspressoCubeLayout(parts, variables)
	on, _ := NewEspressoSetFamily(1<<uint(variables), layout.bitCount)
	off, _ := NewEspressoSetFamily(1<<uint(variables), layout.bitCount)
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		onCube, _ := NewEspressoSet(layout.bitCount)
		offCube, _ := NewEspressoSet(layout.bitCount)
		for variable := 0; variable < variables; variable++ {
			part := 2*variable + ((assignment >> uint(variable)) & 1)
			_ = onCube.Set(part, true)
			_ = offCube.Set(part, true)
		}
		for output, truth := range truths {
			if truth>>uint(assignment)&1 != 0 {
				_ = onCube.Set(2*variables+output, true)
			} else {
				_ = offCube.Set(2*variables+output, true)
			}
		}
		onEmpty, _ := onCube.Empty()
		offEmpty, _ := offCube.Empty()
		// Input coordinates make both rows nonempty, so inspect the output field.
		onOutputs, offOutputs := false, false
		for output := range truths {
			onPart, _ := onCube.Has(2*variables + output)
			offPart, _ := offCube.Has(2*variables + output)
			onOutputs = onOutputs || onPart
			offOutputs = offOutputs || offPart
		}
		if !onEmpty && onOutputs {
			_ = on.Add(onCube)
		}
		if !offEmpty && offOutputs {
			_ = off.Add(offCube)
		}
	}
	dc, _ := NewEspressoSetFamily(0, layout.bitCount)
	pla, err := NewEspressoPLA(layout, on, dc, off, nil)
	if err != nil {
		t.Fatal(err)
	}
	return pla
}

func espressoStoredOutputTruth(t testing.TB, pla *EspressoPLA, output int) uint64 {
	t.Helper()
	variables := pla.Layout.BinaryVariableCount()
	truth := uint64(0)
	part := pla.Layout.firstParts[pla.Layout.OutputVariable()] + output
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		for _, cube := range pla.On.sets {
			selected, _ := cube.Has(part)
			if !selected {
				continue
			}
			matches := true
			for variable := 0; variable < variables; variable++ {
				value := (assignment >> uint(variable)) & 1
				present, _ := cube.Has(2*variable + value)
				if !present {
					matches = false
					break
				}
			}
			if matches {
				truth |= uint64(1) << uint(assignment)
				break
			}
		}
	}
	return truth
}

func TestEspressoMinimizeOutputsStrategiesAndPolarity(t *testing.T) {
	truths := []uint64{0b0001, 0b0111, 0b1010}
	pla := espressoPLAFromOutputTruths(t, 2, truths)
	for index, options := range []EspressoOutputMinimizeOptions{
		{Strategy: EspressoOutputHeuristic, Heuristic: EspressoMinimizeOptions{MaxIterations: 16}},
		{Strategy: EspressoOutputExact, Exact: EspressoExactOptions{MaxSearchNodes: 1 << 16}},
		{Strategy: EspressoOutputHeuristic, ChoosePolarity: true, Heuristic: EspressoMinimizeOptions{MaxIterations: 16}},
		{Strategy: EspressoOutputExact, ChoosePolarity: true, Exact: EspressoExactOptions{MaxSearchNodes: 1 << 16}},
	} {
		result, err := pla.MinimizeOutputs(options)
		if err != nil {
			t.Fatal(index, err)
		}
		if len(result.PositivePhase) != len(truths) || len(result.CubeCounts) != len(truths) {
			t.Fatal(index, result)
		}
		universe := uint64(0b1111)
		for output, truth := range truths {
			want := truth
			if !result.PositivePhase[output] {
				want = universe &^ truth
			}
			if got := espressoStoredOutputTruth(t, result.PLA, output); got != want {
				t.Fatal(index, output, got, want, result.PositivePhase)
			}
		}
	}
	if !slices.Equal(espressoFamilyMasks(t, pla.On), espressoFamilyMasks(t, espressoPLAFromOutputTruths(t, 2, truths).On)) {
		t.Fatal("input PLA mutated")
	}
}

func TestEspressoMinimizeOutputsErrorsAndCancellation(t *testing.T) {
	pla := espressoPLAFromOutputTruths(t, 1, []uint64{1, 2})
	if result, err := pla.MinimizeOutputs(EspressoOutputMinimizeOptions{Strategy: 9}); result.PLA != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := pla.MinimizeOutputs(EspressoOutputMinimizeOptions{Manipulation: EspressoCoverManipulationOptions{MaxCubes: 1}}); result.PLA != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := pla.MinimizeOutputs(EspressoOutputMinimizeOptions{Manipulation: EspressoCoverManipulationOptions{Stop: func() bool { return true }}}); result.PLA != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	layout, _ := NewEspressoCubeLayout([]int{2}, 1)
	empty := espressoMapFamily(t, 2)
	withoutOutput, _ := NewEspressoPLA(layout, empty, empty, empty, nil)
	if result, err := withoutOutput.MinimizeOutputs(EspressoOutputMinimizeOptions{}); result.PLA != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
}

func FuzzEspressoMinimizeOutputs(f *testing.F) {
	f.Add([]byte{2, 3, 0x1, 0x7, 0xa, 0})
	f.Add([]byte{1, 2, 1, 2, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 4 {
			return
		}
		variables := int(data[0]%2) + 1
		outputs := int(data[1]%3) + 1
		assignments := 1 << uint(variables)
		universe := uint64(1)<<uint(assignments) - 1
		truths := make([]uint64, outputs)
		for output := range truths {
			truths[output] = uint64(data[(2+output)%len(data)]) & universe
		}
		flags := data[len(data)-1]
		strategy := EspressoOutputHeuristic
		if flags&1 != 0 {
			strategy = EspressoOutputExact
		}
		pla := espressoPLAFromOutputTruths(t, variables, truths)
		result, err := pla.MinimizeOutputs(EspressoOutputMinimizeOptions{
			Strategy: strategy, ChoosePolarity: flags&2 != 0,
			Manipulation: EspressoCoverManipulationOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 20},
			Heuristic:    EspressoMinimizeOptions{MaxIterations: 16, MaxCubes: 1 << 14, MaxOperations: 1 << 20, MaxRows: 1 << 14, MaxDepth: 1 << 12},
			Exact:        EspressoExactOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 20, MaxDepth: 1 << 12, MaxSearchNodes: 1 << 16},
		})
		if err != nil {
			t.Fatal(variables, truths, flags, err)
		}
		for output, truth := range truths {
			want := truth
			if !result.PositivePhase[output] {
				want = universe &^ truth
			}
			if got := espressoStoredOutputTruth(t, result.PLA, output); got != want {
				t.Fatal(variables, truths, flags, output, got, want, result.PositivePhase)
			}
		}
	})
}
