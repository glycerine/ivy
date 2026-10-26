package abc

// Go-only semantic and cost models for src/misc/espresso/sparse.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestEspressoMVReduceSparsePartsFlagsAndOwnership(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 3}, 2)
	onMasks := []uint64{
		0b01 | 0b11<<2 | 0b011<<4,
		0b10 | 0b01<<2 | 0b101<<4,
		0b11 | 0b01<<2 | 0b001<<4,
	}
	dcMasks := []uint64{0b01 | 0b10<<2 | 0b100<<4}
	on := espressoMapFamily(t, 7, onMasks...)
	for _, cube := range on.sets {
		cube.Flags |= EspressoSetPrime
	}
	dc := espressoMapFamily(t, 7, dcMasks...)
	before := espressoFamilyMasks(t, on)
	result, err := layout.MVReduce(on, dc, EspressoMakeSparseOptions{SparseVariables: []bool{false, false, true}, ExactIrredundant: true})
	if err != nil || espressoFamilyTruth(t, layout, result)|espressoFamilyTruth(t, layout, dc) != espressoFamilyTruth(t, layout, on)|espressoFamilyTruth(t, layout, dc) || !slices.Equal(espressoFamilyMasks(t, on), before) {
		t.Fatal(espressoFamilyMasks(t, result), err)
	}
	for _, cube := range result.sets {
		if !layout.variableIntersection(cube, cube, 2) || cube.Flags&EspressoSetActive == 0 {
			t.Fatal(espressoSetMask(t, cube), cube.Flags)
		}
	}
}

func TestEspressoMakeSparseCostSemanticsAndIteration(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 2)
	onTruth, dcTruth := uint64(0b00110101), uint64(0b00001010)
	universe := espressoTruthUniverse(layout)
	on := espressoMintermCover(t, 3, onTruth)
	dc := espressoMintermCover(t, 3, dcTruth)
	off := espressoMintermCover(t, 3, universe&^(onTruth|dcTruth))
	options := EspressoMakeSparseOptions{SparseVariables: []bool{false, false, true}, ForceIrredundant: true, ExactIrredundant: true}
	before, _ := layout.CoverCost(on, EspressoCoverCostOptions{SparseVariables: options.SparseVariables})
	result, err := layout.MakeSparse(on, dc, off, options)
	after, afterErr := layout.CoverCost(result, EspressoCoverCostOptions{SparseVariables: options.SparseVariables})
	if err != nil || afterErr != nil || after.Total > before.Total || espressoFamilyTruth(t, layout, result)|dcTruth != onTruth|dcTruth || espressoFamilyTruth(t, layout, result)&espressoFamilyTruth(t, layout, off) != 0 {
		t.Fatal(before, after, espressoFamilyMasks(t, result), err, afterErr)
	}
}

func TestEspressoMakeSparseErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	on := espressoMintermCover(t, 2, 0b0111)
	dc := espressoMintermCover(t, 2, 0)
	off := espressoMintermCover(t, 2, 0b1000)
	checks := []EspressoMakeSparseOptions{{MaxIterations: -1}, {MaxCubes: 1}, {MaxOperations: 1}, {MaxRows: -1}, {MaxDepth: -1}, {SparseVariables: []bool{true}}, {Stop: func() bool { return true }}}
	for _, options := range checks {
		result, err := layout.MakeSparse(on, dc, off, options)
		if err == nil || result != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	wrong := espressoMapFamily(t, 3)
	if result, err := layout.MVReduce(on, wrong, EspressoMakeSparseOptions{}); !errors.Is(err, ErrDimension) || result != nil {
		t.Fatal(result, err)
	}
}

func FuzzEspressoMakeSparse(f *testing.F) {
	f.Add([]byte{3, 0x35, 0x0a, 0x01})
	f.Add([]byte{1, 1, 0, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 4 {
			return
		}
		variables := int(data[0]%3) + 1
		assignments := 1 << uint(variables)
		universe := uint64(1)<<uint(assignments) - 1
		onTruth := uint64(data[1]) & universe
		dcTruth := uint64(data[2]) & universe &^ onTruth
		offTruth := universe &^ (onTruth | dcTruth)
		layout, _ := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), max(0, variables-1))
		on := espressoMintermCover(t, variables, onTruth)
		dc := espressoMintermCover(t, variables, dcTruth)
		off := espressoMintermCover(t, variables, offTruth)
		sparse := make([]bool, variables)
		for variable := range sparse {
			sparse[variable] = data[(3+variable)%len(data)]&1 != 0
		}
		options := EspressoMakeSparseOptions{SparseVariables: sparse, ExactIrredundant: true, ForceIrredundant: data[3]&2 != 0, MaxIterations: 16, MaxCubes: 1 << 15, MaxOperations: 1 << 20, MaxRows: 1 << 14}
		before, err := layout.CoverCost(on, EspressoCoverCostOptions{SparseVariables: sparse})
		if err != nil {
			t.Fatal(err)
		}
		result, err := layout.MakeSparse(on, dc, off, options)
		if err != nil {
			t.Fatal(err)
		}
		after, err := layout.CoverCost(result, EspressoCoverCostOptions{SparseVariables: sparse})
		if err != nil || after.Total > before.Total || espressoFamilyTruth(t, layout, result)|dcTruth != onTruth|dcTruth || espressoFamilyTruth(t, layout, result)&offTruth != 0 {
			t.Fatal(variables, onTruth, dcTruth, sparse, before, after, espressoFamilyMasks(t, result), err)
		}
		for _, cube := range result.sets {
			for variable, yes := range sparse {
				if yes && !layout.variableIntersection(cube, cube, variable) {
					t.Fatal("empty sparse variable", variable, espressoSetMask(t, cube))
				}
			}
		}
	})
}
