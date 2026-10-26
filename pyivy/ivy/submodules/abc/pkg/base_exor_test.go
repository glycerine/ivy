package abc

// Tests for the Go port of src/base/exor.

import (
	"errors"
	"math/rand"
	"slices"
	"strings"
	"testing"
)

func exorTestCube(inputs []ExorValue, output uint64) ExorCube {
	return NewExorCube(inputs, []uint64{output})
}

func exorModelEvaluate(c ExorCover, assignment uint64) []uint64 {
	result := make([]uint64, exorWordCount(c.Outputs))
	for _, cube := range c.Cubes {
		match := true
		for variable, literal := range cube.Inputs {
			bit := assignment>>variable&1 != 0
			if literal == ExorNegative && bit || literal == ExorPositive && !bit {
				match = false
				break
			}
		}
		if match {
			for i := range result {
				result[i] ^= cube.Outputs[i]
			}
		}
	}
	return result
}

func exorAssertEquivalent(t *testing.T, a, b ExorCover) {
	t.Helper()
	if a.Inputs != b.Inputs || a.Outputs != b.Outputs {
		t.Fatalf("dimension mismatch: %#v %#v", a, b)
	}
	for assignment := uint64(0); assignment < uint64(1)<<a.Inputs; assignment++ {
		gotA := exorModelEvaluate(a, assignment)
		gotB := exorModelEvaluate(b, assignment)
		if !slices.Equal(gotA, gotB) {
			t.Fatalf("assignment %0*b: %x != %x\na=%#v\nb=%#v", a.Inputs, assignment, gotA, gotB, a, b)
		}
		bitsAssignment := make([]bool, a.Inputs)
		for i := range bitsAssignment {
			bitsAssignment[i] = assignment>>i&1 != 0
		}
		production, err := a.Evaluate(bitsAssignment)
		if err != nil || !slices.Equal(production, gotA) {
			t.Fatalf("production evaluator: %x, %v; model %x", production, err, gotA)
		}
	}
}

func TestExorCostsAndPrimitiveOperations(t *testing.T) {
	want := [][]int{{1}, {1, 2}, {5, 5, 6}, {14, 14, 16, 18}, {20, 20, 20, 22, 24}, {32, 32, 32, 34, 36, 38}}
	for literals, row := range want {
		for negative, cost := range row {
			if got := ExorQCost(literals, negative); got != cost {
				t.Fatalf("qcost(%d,%d)=%d want %d", literals, negative, got, cost)
			}
		}
	}
	if ExorQCost(2, 3) != -1 {
		t.Fatal("invalid cost was accepted")
	}
	for _, tc := range []struct{ controls, lines, want int }{{0, 1, 0}, {2, 1, 1}, {3, 4, 4}, {4, 7, 8}, {4, 3, 10}, {6, 11, 16}, {6, 3, 24}} {
		got, err := ExorToffoliGateCount(tc.controls, tc.lines)
		if err != nil || got != tc.want {
			t.Fatalf("Toffoli(%d,%d)=%d,%v want %d", tc.controls, tc.lines, got, err, tc.want)
		}
	}
	if _, err := ExorToffoliGateCount(-1, 1); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative controls: %v", err)
	}
	if got, err := exorTestCube([]ExorValue{ExorNegative, ExorPositive, ExorAbsent}, 1).TCount(4); err != nil || got != 7 {
		t.Fatalf("T count=%d,%v", got, err)
	}

	a := exorTestCube([]ExorValue{ExorNegative, ExorPositive, ExorAbsent}, 1)
	b := exorTestCube([]ExorValue{ExorPositive, ExorPositive, ExorNegative}, 3)
	if d, err := ExorDistance(a, b, 4); err != nil || d != 3 {
		t.Fatalf("distance=%d,%v", d, err)
	}
	diff, err := ExorDifferentVariables(a, b, 4)
	if err != nil || !slices.Equal(diff, []int{-1, 0, 2}) {
		t.Fatalf("diff=%v,%v", diff, err)
	}
}

func TestExorNormalizeAdjacentAndOwnership(t *testing.T) {
	c := ExorCover{Inputs: 2, Outputs: 2, Cubes: []ExorCube{
		exorTestCube([]ExorValue{ExorNegative, ExorAbsent}, 1),
		exorTestCube([]ExorValue{ExorNegative, ExorAbsent}, 2),
		exorTestCube([]ExorValue{ExorPositive, ExorAbsent}, 3),
		exorTestCube([]ExorValue{ExorAbsent, ExorPositive}, 1),
		exorTestCube([]ExorValue{ExorAbsent, ExorPositive}, 1),
	}}
	n, err := c.Normalize()
	if err != nil || len(n.Cubes) != 2 || n.Cubes[0].Outputs[0] != 3 {
		t.Fatalf("normalize=%#v,%v", n, err)
	}
	r, stats, err := MinimizeExorCover(c, ExorOptions{})
	if err != nil || len(r.Cubes) != 1 || r.Cubes[0].Inputs[0] != ExorAbsent || r.Cubes[0].Outputs[0] != 3 {
		t.Fatalf("reduce=%#v stats=%+v err=%v", r, stats, err)
	}
	exorAssertEquivalent(t, c, r)
	r.Cubes[0].Inputs[0] = ExorNegative
	if c.Cubes[0].Inputs[0] != ExorNegative || c.Cubes[1].Outputs[0] != 2 {
		t.Fatal("result aliases source")
	}
}

func TestExorLinkGroupsPreserveFunction(t *testing.T) {
	for distance := 2; distance <= 4; distance++ {
		inputs := distance
		aInputs := make([]ExorValue, inputs)
		bInputs := make([]ExorValue, inputs)
		for i := range aInputs {
			aInputs[i] = ExorNegative
			bInputs[i] = ExorPositive
		}
		a := exorTestCube(aInputs, 1)
		b := exorTestCube(bInputs, 1)
		groups, err := ExorLinkCandidates(a, b)
		if err != nil || len(groups) != []int{0, 0, 2, 6, 24}[distance] {
			t.Fatalf("distance %d groups=%d err=%v", distance, len(groups), err)
		}
		original := ExorCover{Inputs: inputs, Outputs: 1, Cubes: []ExorCube{a, b}}
		for groupNo, group := range groups {
			candidate := ExorCover{Inputs: inputs, Outputs: 1, Cubes: group}
			if err := candidate.Validate(0); err != nil {
				t.Fatalf("distance %d group %d invalid: %v", distance, groupNo, err)
			}
			exorAssertEquivalent(t, original, candidate)
		}
	}

	// Output difference occupies the first ExorLink dimension, as in FindDiffVars.
	a := exorTestCube([]ExorValue{ExorNegative, ExorAbsent}, 1)
	b := exorTestCube([]ExorValue{ExorPositive, ExorAbsent}, 2)
	groups, err := ExorLinkCandidates(a, b)
	if err != nil {
		t.Fatal(err)
	}
	original := ExorCover{Inputs: 2, Outputs: 2, Cubes: []ExorCube{a, b}}
	for _, group := range groups {
		exorAssertEquivalent(t, original, ExorCover{Inputs: 2, Outputs: 2, Cubes: group})
	}
}

func TestExorLinkMixedValuesIndependentModel(t *testing.T) {
	rng := rand.New(rand.NewSource(0x4c494e4b))
	for run := 0; run < 5000; run++ {
		inputs := 4
		aInputs := make([]ExorValue, inputs)
		bInputs := make([]ExorValue, inputs)
		for i := range aInputs {
			aInputs[i] = ExorValue(1 + rng.Intn(3))
			bInputs[i] = aInputs[i]
		}
		distance := 2 + rng.Intn(3)
		outputDiff := rng.Intn(2) == 0
		needInputs := distance
		if outputDiff {
			needInputs--
		}
		permutation := rng.Perm(inputs)
		for _, variable := range permutation[:needInputs] {
			for bInputs[variable] == aInputs[variable] {
				bInputs[variable] = ExorValue(1 + rng.Intn(3))
			}
		}
		aOutput, bOutput := uint64(1+rng.Intn(3)), uint64(1+rng.Intn(3))
		if !outputDiff {
			bOutput = aOutput
		} else if bOutput == aOutput {
			bOutput = aOutput ^ 3
			if bOutput == 0 {
				bOutput = 1
			}
		}
		a := exorTestCube(aInputs, aOutput)
		b := exorTestCube(bInputs, bOutput)
		groups, err := ExorLinkCandidates(a, b)
		if err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		original := ExorCover{Inputs: inputs, Outputs: 2, Cubes: []ExorCube{a, b}}
		for groupNo, group := range groups {
			candidate := ExorCover{Inputs: inputs, Outputs: 2, Cubes: group}
			for assignment := uint64(0); assignment < 1<<inputs; assignment++ {
				if !slices.Equal(exorModelEvaluate(original, assignment), exorModelEvaluate(candidate, assignment)) {
					t.Fatalf("run %d group %d assignment %d is not equivalent", run, groupNo, assignment)
				}
			}
		}
	}
}

func exorRandomCover(rng *rand.Rand, maxInputs, maxOutputs, maxCubes int) ExorCover {
	inputs := 1 + rng.Intn(maxInputs)
	outputs := 1 + rng.Intn(maxOutputs)
	c := ExorCover{Inputs: inputs, Outputs: outputs}
	for i := 0; i < 1+rng.Intn(maxCubes); i++ {
		cube := ExorCube{Inputs: make([]ExorValue, inputs), Outputs: make([]uint64, exorWordCount(outputs))}
		for j := range cube.Inputs {
			cube.Inputs[j] = ExorValue(1 + rng.Intn(3))
		}
		for j := range cube.Outputs {
			cube.Outputs[j] = rng.Uint64()
		}
		if rem := outputs & 63; rem != 0 {
			cube.Outputs[len(cube.Outputs)-1] &= uint64(1)<<rem - 1
		}
		if exorOutputsZero(cube.Outputs) {
			cube.Outputs[0] = 1
		}
		c.Cubes = append(c.Cubes, cube)
	}
	return c
}

func TestExorDeterministicIndependentModel5000(t *testing.T) {
	rng := rand.New(rand.NewSource(0x45584f52))
	for run := 0; run < 5000; run++ {
		input := exorRandomCover(rng, 5, 5, 10)
		got, stats, err := MinimizeExorCover(input, ExorOptions{Quality: run & 1, MaxCubes: 32, MaxPairChecks: 100000, UseQCost: run&2 != 0})
		if err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		exorAssertEquivalent(t, input, got)
		if stats.CubesAfter > stats.CubesBefore {
			t.Fatalf("run %d grew from %d to %d", run, stats.CubesBefore, stats.CubesAfter)
		}
		again, _, err := MinimizeExorCover(input, ExorOptions{Quality: run & 1, MaxCubes: 32, MaxPairChecks: 100000, UseQCost: run&2 != 0})
		if err != nil {
			t.Fatal(err)
		}
		aDigest, _ := got.ExorCoverDigest()
		bDigest, _ := again.ExorCoverDigest()
		if !slices.Equal(aDigest, bDigest) {
			t.Fatalf("run %d is nondeterministic", run)
		}
	}
}

func TestExorLimitsCancellationAndValidationAreAtomic(t *testing.T) {
	input := ExorCover{Inputs: 2, Outputs: 1, Cubes: []ExorCube{
		exorTestCube([]ExorValue{ExorNegative, ExorNegative}, 1),
		exorTestCube([]ExorValue{ExorPositive, ExorPositive}, 1),
	}}
	before, _ := input.ExorCoverDigest()
	if got, _, err := MinimizeExorCover(input, ExorOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) || got.Cubes != nil {
		t.Fatalf("cancel result=%#v err=%v", got, err)
	}
	if got, _, err := MinimizeExorCover(input, ExorOptions{MaxPairChecks: 1}); !errors.Is(err, ErrLimit) || got.Cubes != nil {
		t.Fatalf("limit result=%#v err=%v", got, err)
	}
	after, _ := input.ExorCoverDigest()
	if !slices.Equal(before, after) {
		t.Fatal("failed operation mutated input")
	}
	bad := input.Clone()
	bad.Cubes[0].Inputs[0] = 0
	if err := bad.Validate(0); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad literal accepted: %v", err)
	}
	bad = input.Clone()
	bad.Cubes[0].Outputs[0] = 2
	if err := bad.Validate(0); !errors.Is(err, ErrDimension) {
		t.Fatalf("padding accepted: %v", err)
	}
	tooWide := ExorCover{Inputs: 1, Outputs: ExorMaxOutputs + 1}
	if err := tooWide.Validate(0); !errors.Is(err, ErrLimit) {
		t.Fatalf("oversized output count accepted: %v", err)
	}
}

func TestExorPLAFormatting(t *testing.T) {
	c := ExorCover{Inputs: 3, Outputs: 2, Cubes: []ExorCube{
		exorTestCube([]ExorValue{ExorNegative, ExorPositive, ExorAbsent}, 1),
		exorTestCube([]ExorValue{ExorAbsent, ExorAbsent, ExorPositive}, 2),
	}}
	got, err := c.MarshalExorPLA()
	want := ".i 3\n.o 2\n.p 2\n.type esop\n01- 10\n--1 01\n.e\n"
	if err != nil || string(got) != want {
		t.Fatalf("PLA=%q,%v want %q", got, err, want)
	}
	if strings.Contains(string(got), "#") {
		t.Fatal("nondeterministic metadata in output")
	}
}

func FuzzExorMinimize(f *testing.F) {
	f.Add([]byte{2, 1, 3, 1, 2, 1, 1, 2})
	f.Add([]byte{4, 3, 8, 1, 1, 2, 2, 3, 3, 7})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		inputs := 1 + int(data[0]%6)
		outputs := 1
		if len(data) > 1 {
			outputs += int(data[1] % 6)
		}
		count := 1
		if len(data) > 2 {
			count += int(data[2] % 12)
		}
		cover := ExorCover{Inputs: inputs, Outputs: outputs}
		pos := 3
		for cubeNo := 0; cubeNo < count; cubeNo++ {
			cube := ExorCube{Inputs: make([]ExorValue, inputs), Outputs: make([]uint64, 1)}
			for i := range cube.Inputs {
				value := byte(cubeNo + i)
				if pos < len(data) {
					value = data[pos]
					pos++
				}
				cube.Inputs[i] = ExorValue(1 + value%3)
			}
			value := byte(cubeNo + 1)
			if pos < len(data) {
				value = data[pos]
				pos++
			}
			cube.Outputs[0] = uint64(value) & (uint64(1)<<outputs - 1)
			if cube.Outputs[0] == 0 {
				cube.Outputs[0] = 1
			}
			cover.Cubes = append(cover.Cubes, cube)
		}
		got, _, err := MinimizeExorCover(cover, ExorOptions{Quality: 1, MaxCubes: 32, MaxPairChecks: 200000, UseQCost: len(data)&1 != 0})
		if err != nil {
			t.Fatalf("minimize: %v", err)
		}
		exorAssertEquivalent(t, cover, got)
		if len(got.Cubes) > len(cover.Cubes) {
			t.Fatalf("cover grew: %d > %d", len(got.Cubes), len(cover.Cubes))
		}
	})
}
