package abc

// Go-only minimization and assignment models for src/misc/espresso/espresso.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestEspressoMinimizeStrategiesPreserveFunction(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 3)
	onTruth, dcTruth := uint64(0b01111110), uint64(0)
	universe := espressoTruthUniverse(layout)
	offTruth := universe &^ (onTruth | dcTruth)
	on := espressoMintermCover(t, 3, onTruth)
	dc := espressoMintermCover(t, 3, dcTruth)
	off := espressoMintermCover(t, 3, offTruth)
	before := espressoFamilyMasks(t, on)
	strategies := []EspressoMinimizeOptions{
		{MaxIterations: 16},
		{SingleExpand: true, MaxIterations: 16},
		{RemoveEssential: true, MaxIterations: 16},
		{UseSuperGasp: true, ExactIrredundant: true, MaxIterations: 16},
		{RecomputeOnset: true, ReduceOrder: EspressoReduceColumnOrder, MaxIterations: 16},
		{SkipMakeSparse: true, MaxIterations: 16},
	}
	for index, options := range strategies {
		result, err := layout.MinimizeEspressoCover(on, dc, off, options)
		if err != nil {
			t.Fatal(index, err)
		}
		truth := espressoFamilyTruth(t, layout, result.Cover)
		if truth|dcTruth != onTruth|dcTruth || truth&offTruth != 0 || len(result.Cover.sets) > len(on.sets) || result.Cost.Cubes != len(result.Cover.sets) {
			t.Fatal(index, result.Cost, espressoFamilyMasks(t, result.Cover))
		}
	}
	if !slices.Equal(espressoFamilyMasks(t, on), before) {
		t.Fatal("minimization mutated input")
	}
}

func TestEspressoMinimizeRandomOrderDeterministic(t *testing.T) {
	layout, err := NewEspressoCubeLayout([]int{2, 2, 2}, 3)
	if err != nil {
		t.Fatal(err)
	}
	onTruth := uint64(0b01111110)
	dcTruth := uint64(0)
	offTruth := espressoTruthUniverse(layout) &^ (onTruth | dcTruth)
	on := espressoMintermCover(t, 3, onTruth)
	dc := espressoMintermCover(t, 3, dcTruth)
	off := espressoMintermCover(t, 3, offTruth)
	options := EspressoMinimizeOptions{UseRandomOrder: true, RandomSeed: 314973, MaxIterations: 32}
	first, err := layout.MinimizeEspressoCover(on, dc, off, options)
	second, secondErr := layout.MinimizeEspressoCover(on, dc, off, options)
	if err != nil || secondErr != nil || !slices.Equal(espressoFamilyMasks(t, first.Cover), espressoFamilyMasks(t, second.Cover)) || espressoFamilyTruth(t, layout, first.Cover) != espressoFamilyTruth(t, layout, on) {
		t.Fatal(espressoFamilyMasks(t, first.Cover), espressoFamilyMasks(t, second.Cover), err, secondErr)
	}
}

func TestEspressoMinimizeOutputUnravel(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 3}, 2)
	on := espressoMapFamily(t, 7, 0b01|0b11<<2|0b011<<4)
	dc := espressoMapFamily(t, 7)
	off, err := layout.ComplementCover(on, EspressoSharpOptions{})
	if err != nil {
		t.Fatal(err)
	}
	options, _ := normalizeEspressoMinimizeOptions(layout, EspressoMinimizeOptions{UnwrapOnset: true, SingleExpand: true})
	unravelled, err := layout.unravelOutput(on, options)
	if err != nil || len(unravelled.sets) != 2 || espressoFamilyTruth(t, layout, unravelled) != espressoFamilyTruth(t, layout, on) {
		t.Fatal(espressoFamilyMasks(t, unravelled), err)
	}
	result, err := layout.MinimizeEspressoCover(on, dc, off, EspressoMinimizeOptions{UnwrapOnset: true, SingleExpand: true})
	if err != nil || result.Fallback || !result.Unwrapped || len(result.Cover.sets) != 1 || espressoFamilyTruth(t, layout, result.Cover) != espressoFamilyTruth(t, layout, on) {
		t.Fatal(result.Cost, result.Unwrapped, result.Fallback, espressoFamilyMasks(t, result.Cover), err)
	}
}

func TestEspressoMinimizeErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	on := espressoMintermCover(t, 2, 0b0011)
	dc := espressoMintermCover(t, 2, 0)
	off := espressoMintermCover(t, 2, 0b1100)
	for _, options := range []EspressoMinimizeOptions{
		{MaxIterations: -1},
		{MaxCubes: -1},
		{MaxOperations: -1},
		{MaxRows: -1},
		{MaxDepth: -1},
		{ReduceOrder: 9},
		{SparseVariables: []bool{true}},
		{MaxCubes: 1},
		{MaxOperations: 1},
		{Stop: func() bool { return true }},
	} {
		result, err := layout.MinimizeEspressoCover(on, dc, off, options)
		if result.Cover != nil || err == nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	overlap := espressoMintermCover(t, 2, 0b0001)
	if result, err := layout.MinimizeEspressoCover(on, dc, overlap, EspressoMinimizeOptions{}); result.Cover != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	wrong := espressoMapFamily(t, 3)
	if result, err := layout.MinimizeEspressoCover(on, wrong, off, EspressoMinimizeOptions{}); result.Cover != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
}

func FuzzEspressoMinimize(f *testing.F) {
	f.Add([]byte{3, 0x7e, 0, 0})
	f.Add([]byte{2, 0x07, 0x08, 0x1f})
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
		layout, _ := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), variables)
		on := espressoMintermCover(t, variables, onTruth)
		dc := espressoMintermCover(t, variables, dcTruth)
		off := espressoMintermCover(t, variables, offTruth)
		flags := data[3]
		options := EspressoMinimizeOptions{
			RecomputeOnset: flags&1 != 0, RemoveEssential: flags&2 != 0,
			SingleExpand: flags&4 != 0, UseSuperGasp: flags&8 != 0,
			SkipMakeSparse: flags&16 != 0, ExactIrredundant: flags&32 != 0,
			UseRandomOrder: flags&64 != 0, RandomSeed: uint64(data[1])<<8 | uint64(data[2]),
			ReduceOrder: EspressoReduceOrder(flags >> 7), MaxIterations: 16,
			MaxCubes: 1 << 14, MaxOperations: 1 << 20, MaxRows: 1 << 14, MaxDepth: 1 << 12,
		}
		result, err := layout.MinimizeEspressoCover(on, dc, off, options)
		if err != nil {
			t.Fatal(variables, onTruth, dcTruth, flags, err)
		}
		truth := espressoFamilyTruth(t, layout, result.Cover)
		if truth|dcTruth != onTruth|dcTruth || truth&offTruth != 0 || len(result.Cover.sets) > len(on.sets) || result.Cost.Cubes != len(result.Cover.sets) {
			t.Fatal(variables, onTruth, dcTruth, flags, result.Cost, espressoFamilyMasks(t, result.Cover))
		}
	})
}
