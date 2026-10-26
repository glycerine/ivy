package abc

// Go-only truth-table and cover models for src/misc/espresso/gasp.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestEspressoReduceGaspIndependentFlagsAndOwnership(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	on := espressoMapFamily(t, 4, 0b01|0b11<<2, 0b11|0b01<<2)
	before := espressoFamilyMasks(t, on)
	reduced, err := layout.ReduceGasp(on, espressoMapFamily(t, 4), EspressoGaspOptions{})
	if err != nil || !slices.Equal(espressoFamilyMasks(t, reduced), []uint64{0b1001, 0b0110}) || !slices.Equal(espressoFamilyMasks(t, on), before) {
		t.Fatal(espressoFamilyMasks(t, reduced), espressoFamilyMasks(t, on), err)
	}
	for _, cube := range reduced.sets {
		if cube.Flags&EspressoSetActive == 0 || cube.Flags&EspressoSetPrime != 0 {
			t.Fatal(cube.Flags)
		}
	}
	reduced.sets[0].words[0] = 0
	if !slices.Equal(espressoFamilyMasks(t, on), before) {
		t.Fatal("result aliases input")
	}
}

func TestEspressoGaspWorkflowsPreserveFunction(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 3)
	universe := espressoTruthUniverse(layout)
	onTruth, dcTruth := uint64(0b01111110), uint64(0)
	offTruth := universe &^ (onTruth | dcTruth)
	minterms := espressoMintermCover(t, 3, onTruth)
	dontCare := espressoMintermCover(t, 3, dcTruth)
	off := espressoMintermCover(t, 3, offTruth)
	options := EspressoGaspOptions{ExactIrredundant: true}
	primeOn, err := layout.ExpandCover(minterms, off, options.expandOptions())
	if err != nil {
		t.Fatal(err)
	}
	reduced, err := layout.ReduceGasp(primeOn, dontCare, options)
	if err != nil {
		t.Fatal(err)
	}
	perCube, err := layout.ExpandGaspCube(reduced, dontCare, off, primeOn, 0, options)
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := layout.ExpandGasp(reduced, dontCare, off, primeOn, options)
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates.sets {
		if candidate.Flags&EspressoSetPrime == 0 || espressoFamilyTruth(t, layout, espressoFamilyFromSets(t, layout.bitCount, candidate))&offTruth != 0 {
			t.Fatal(espressoSetMask(t, candidate), candidate.Flags)
		}
	}
	combined, err := layout.IrredundantGasp(primeOn, dontCare, candidates, options)
	if err != nil {
		t.Fatal(err)
	}
	last, lastErr := layout.LastGasp(primeOn, dontCare, off, options)
	super, superErr := layout.SuperGasp(primeOn, dontCare, off, options)
	for name, result := range map[string]*EspressoSetFamily{"combined": combined, "last": last, "super": super} {
		if result == nil || espressoFamilyTruth(t, layout, result)|dcTruth != onTruth|dcTruth || espressoFamilyTruth(t, layout, result)&offTruth != 0 {
			t.Fatal(name, espressoFamilyMasks(t, result), lastErr, superErr)
		}
	}
	if lastErr != nil || superErr != nil || perCube == nil {
		t.Fatal(lastErr, superErr, perCube)
	}
}

func espressoFamilyFromSets(t testing.TB, bits int, sets ...*EspressoSet) *EspressoSetFamily {
	t.Helper()
	result, err := NewEspressoSetFamily(len(sets), bits)
	if err != nil {
		t.Fatal(err)
	}
	for _, set := range sets {
		if err = result.Add(set); err != nil {
			t.Fatal(err)
		}
	}
	return result
}

func TestEspressoGaspErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	on := espressoMintermCover(t, 2, 0b0011)
	dc := espressoMintermCover(t, 2, 0)
	off := espressoMintermCover(t, 2, 0b1100)
	for _, options := range []EspressoGaspOptions{
		{MaxCubes: -1},
		{MaxOperations: -1},
		{MaxRows: -1},
		{MaxDepth: -1},
		{SparseVariables: []bool{true}},
		{MaxCubes: 1},
		{MaxOperations: 1},
		{Stop: func() bool { return true }},
	} {
		result, err := layout.LastGasp(on, dc, off, options)
		if result != nil || err == nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	overlap := espressoMintermCover(t, 2, 0b0001)
	if result, err := layout.SuperGasp(on, dc, overlap, EspressoGaspOptions{}); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	wrong := espressoMapFamily(t, 3)
	if result, err := layout.ExpandGasp(on, dc, wrong, on, EspressoGaspOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
}

func FuzzEspressoGasp(f *testing.F) {
	f.Add([]byte{3, 0x7e, 0, 1})
	f.Add([]byte{2, 0x07, 0x08, 0})
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
		options := EspressoGaspOptions{ExactIrredundant: data[3]&1 != 0, MaxCubes: 1 << 14, MaxOperations: 1 << 20, MaxRows: 1 << 14, MaxDepth: 1 << 12}
		primeOn, err := layout.ExpandCover(on, off, options.expandOptions())
		if err != nil {
			t.Fatal(err)
		}
		marked, err := layout.Irredundant(primeOn, dc, options.irredundantOptions())
		if err != nil {
			t.Fatal(err)
		}
		primeOn = marked.Cover
		last, err := layout.LastGasp(primeOn, dc, off, options)
		if err != nil {
			t.Fatal("last", variables, onTruth, dcTruth, err)
		}
		super, err := layout.SuperGasp(primeOn, dc, off, options)
		if err != nil {
			t.Fatal("super", variables, onTruth, dcTruth, err)
		}
		for name, result := range map[string]*EspressoSetFamily{"last": last, "super": super} {
			truth := espressoFamilyTruth(t, layout, result)
			if truth|dcTruth != onTruth|dcTruth || truth&offTruth != 0 {
				t.Fatal(name, variables, onTruth, dcTruth, espressoFamilyMasks(t, result))
			}
		}
	})
}
