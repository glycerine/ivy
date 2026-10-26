package abc

// Go-only output-phase truth and subset models for src/misc/espresso/opo.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestEspressoOutputPhaseSetupAndApplication(t *testing.T) {
	truths := []uint64{0b0001, 0b0111}
	pla := espressoPLAFromOutputTruths(t, 2, truths)
	doubled, err := pla.OutputPhaseSetup(1, EspressoPhaseOptions{})
	if err != nil || doubled.Layout.partSizes[doubled.Layout.OutputVariable()] != 3 {
		t.Fatal(doubled, err)
	}
	if got := espressoStoredOutputTruth(t, doubled, 0); got != truths[0] {
		t.Fatal(0, got, truths[0])
	}
	if got := espressoStoredOutputTruth(t, doubled, 1); got != truths[1] {
		t.Fatal(1, got, truths[1])
	}
	if got := espressoStoredOutputTruth(t, doubled, 2); got != 0b1111&^truths[1] {
		t.Fatal(2, got, 0b1111&^truths[1])
	}
	phased, err := pla.ApplyOutputPhase([]bool{true, false}, EspressoPhaseOptions{})
	if err != nil || espressoStoredOutputTruth(t, phased, 0) != truths[0] || espressoStoredOutputTruth(t, phased, 1) != 0b1111&^truths[1] {
		t.Fatal(espressoFamilyMasks(t, phased.On), err)
	}
	if !slices.Equal(espressoFamilyMasks(t, pla.On), espressoFamilyMasks(t, espressoPLAFromOutputTruths(t, 2, truths).On)) {
		t.Fatal("phase operations mutated input")
	}
}

func TestEspressoSelectOutputPhasePrimes(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{4}, 0)
	on := espressoMapFamily(t, 4, 0b0001, 0b1100, 0b0010)
	empty := espressoMapFamily(t, 4)
	pla, _ := NewEspressoPLA(layout, on, empty, empty, nil)
	result, err := pla.SelectOutputPhasePrimes(0, EspressoPhaseOptions{})
	if err != nil || !slices.Equal(result.PositivePhase, []bool{false, false}) || !slices.Equal(espressoFamilyMasks(t, result.PLA.On), []uint64{0b1100}) || result.Evaluated != 4 {
		t.Fatal(result.PositivePhase, espressoFamilyMasks(t, result.PLA.On), result.Evaluated, err)
	}
}

func TestEspressoPhaseAssignmentStrategiesAndEnumeration(t *testing.T) {
	truths := []uint64{0b0001, 0b0111, 0b1010}
	pla := espressoPLAFromOutputTruths(t, 2, truths)
	strategies := []EspressoPhaseOptions{
		{Heuristic: EspressoMinimizeOptions{SingleExpand: true, MaxIterations: 16}},
		{Repeated: true, Heuristic: EspressoMinimizeOptions{SingleExpand: true, MaxIterations: 16}},
		{Strategy: EspressoPhaseExact, Exact: EspressoExactOptions{MaxSearchNodes: 1 << 16}},
	}
	for index, options := range strategies {
		result, err := pla.AssignOutputPhase(options)
		if err != nil {
			t.Fatal(index, err)
		}
		assertEspressoPhaseTruth(t, result, truths)
	}
	enumerated, err := pla.EnumerateOutputPhases(0, 2, EspressoPhaseOptions{Heuristic: EspressoMinimizeOptions{SingleExpand: true, MaxIterations: 16}})
	if err != nil || enumerated.Evaluated != 8 {
		t.Fatal(enumerated.Evaluated, err)
	}
	assertEspressoPhaseTruth(t, enumerated, truths)
}

func assertEspressoPhaseTruth(t testing.TB, result EspressoPhaseResult, truths []uint64) {
	t.Helper()
	variables := result.PLA.Layout.BinaryVariableCount()
	universe := uint64(1)<<uint(1<<uint(variables)) - 1
	if len(result.PositivePhase) != len(truths) || result.Cost.Cubes != len(result.PLA.On.sets) {
		t.Fatal(result.PositivePhase, result.Cost)
	}
	for output, truth := range truths {
		want := truth
		if !result.PositivePhase[output] {
			want = universe &^ truth
		}
		if got := espressoStoredOutputTruth(t, result.PLA, output); got != want {
			t.Fatal(output, got, want, result.PositivePhase)
		}
	}
}

func TestEspressoPhaseErrorsLimitsAndCancellation(t *testing.T) {
	pla := espressoPLAFromOutputTruths(t, 1, []uint64{1, 2})
	if result, err := pla.OutputPhaseSetup(-1, EspressoPhaseOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := pla.ApplyOutputPhase([]bool{true}, EspressoPhaseOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	doubled, _ := pla.OutputPhaseSetup(0, EspressoPhaseOptions{})
	if result, err := doubled.SelectOutputPhasePrimes(0, EspressoPhaseOptions{MaxPhases: 1}); result.PLA != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := pla.EnumerateOutputPhases(0, 1, EspressoPhaseOptions{MaxPhases: 1}); result.PLA != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := pla.AssignOutputPhase(EspressoPhaseOptions{Stop: func() bool { return true }}); result.PLA != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	layout, _ := NewEspressoCubeLayout([]int{2}, 1)
	empty := espressoMapFamily(t, 2)
	withoutOutput, _ := NewEspressoPLA(layout, empty, empty, empty, nil)
	if result, err := withoutOutput.AssignOutputPhase(EspressoPhaseOptions{}); result.PLA != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
}

func FuzzEspressoPhaseTransforms(f *testing.F) {
	f.Add([]byte{2, 3, 0x1, 0x7, 0xa, 0x05})
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
		pla := espressoPLAFromOutputTruths(t, variables, truths)
		phase := make([]bool, outputs)
		for output := range phase {
			phase[output] = data[(2+outputs+output)%len(data)]&1 != 0
		}
		phased, err := pla.ApplyOutputPhase(phase, EspressoPhaseOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		assertEspressoPhaseTruth(t, EspressoPhaseResult{PLA: phased, PositivePhase: phase, Cost: EspressoCoverCost{Cubes: len(phased.On.sets)}}, truths)
		first := int(data[len(data)-1]) % outputs
		doubled, err := pla.OutputPhaseSetup(first, EspressoPhaseOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		for output, truth := range truths {
			if got := espressoStoredOutputTruth(t, doubled, output); got != truth {
				t.Fatal(first, output, got, truth)
			}
			if output >= first {
				duplicate := outputs + output - first
				if got := espressoStoredOutputTruth(t, doubled, duplicate); got != universe&^truth {
					t.Fatal(first, duplicate, got, universe&^truth)
				}
			}
		}
	})
}

func FuzzEspressoPhaseAssignment(f *testing.F) {
	f.Add([]byte{2, 3, 0x1, 0x7, 0xa, 0})
	f.Add([]byte{1, 2, 1, 2, 3})
	f.Add([]byte("12107"))
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
		strategy := EspressoPhaseHeuristic
		if flags&1 != 0 {
			strategy = EspressoPhaseExact
		}
		pla := espressoPLAFromOutputTruths(t, variables, truths)
		result, err := pla.AssignOutputPhase(EspressoPhaseOptions{
			Strategy: strategy, Repeated: flags&2 != 0, MaxPhases: 1 << 12, MaxCubes: 1 << 14, MaxOperations: 1 << 20,
			Heuristic: EspressoMinimizeOptions{SingleExpand: true, MaxIterations: 16, MaxCubes: 1 << 14, MaxOperations: 1 << 20, MaxRows: 1 << 14, MaxDepth: 1 << 12},
			Exact:     EspressoExactOptions{MaxCubes: 1 << 16, MaxOperations: 16 << 20, MaxDepth: 1 << 14, MaxSearchNodes: 1 << 20},
		})
		if err != nil {
			t.Fatal(variables, truths, flags, err)
		}
		assertEspressoPhaseTruth(t, result, truths)
	})
}
