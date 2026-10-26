package abc

// Go-only prime enumeration and assignment models for src/misc/espresso/expand.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoExpansionIsMaximal(layout *EspressoCubeLayout, cube *EspressoSet, offTruth uint64, options EspressoExpandOptions) bool {
	mask := espressoSetMaskNoTest(cube)
	for variable, size := range layout.partSizes {
		if options.Nonsparse && options.SparseVariables[variable] {
			continue
		}
		start := layout.firstParts[variable]
		for part := start; part < start+size; part++ {
			if mask>>uint(part)&1 != 0 {
				continue
			}
			if espressoCubeTruth(layout, mask|uint64(1)<<uint(part))&offTruth == 0 {
				return false
			}
		}
	}
	return true
}

func espressoSetMaskNoTest(set *EspressoSet) uint64 {
	var result uint64
	for wordIndex, word := range set.words {
		result |= uint64(word) << uint(32*wordIndex)
	}
	return result
}

func TestEspressoExpandCubeAllPrimesAndOwnership(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	onTruth, offTruth := uint64(0b0111), uint64(0b1000)
	on := espressoMintermCover(t, 2, onTruth)
	off := espressoMintermCover(t, 2, offTruth)
	source, _ := on.Set(0)
	expansion, err := layout.ExpandCube(source, on, off, EspressoExpandOptions{})
	if err != nil || espressoCubeTruth(layout, espressoSetMask(t, expansion.Cube))&offTruth != 0 || espressoCubeTruth(layout, espressoSetMask(t, source))&^espressoCubeTruth(layout, espressoSetMask(t, expansion.Cube)) != 0 || !espressoExpansionIsMaximal(layout, expansion.Cube, offTruth, EspressoExpandOptions{SparseVariables: []bool{false, false}}) {
		t.Fatal(expansion, err)
	}
	all, err := layout.FindAllPrimesCoveringCube(source, off, EspressoExpandOptions{})
	brute := espressoBrutePrimes(layout, onTruth, 0)
	want := make([]uint64, 0)
	sourceMask := espressoSetMask(t, source)
	for _, prime := range brute {
		if sourceMask&^prime == 0 {
			want = append(want, prime)
		}
	}
	if err != nil || !slices.Equal(espressoSortedMasks(t, all), want) {
		t.Fatal(espressoSortedMasks(t, all), want, err)
	}
	expansion.Cube.words[0] = 0
	if espressoFamilyTruth(t, layout, on) != onTruth {
		t.Fatal("source cover mutated")
	}
}

func TestEspressoExpandCoverNonsparseFlagsAndSemantics(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 3)
	onTruth, offTruth := uint64(0b00111111), uint64(0b11000000)
	on := espressoMintermCover(t, 3, onTruth)
	off := espressoMintermCover(t, 3, offTruth)
	before := espressoFamilyMasks(t, on)
	options := EspressoExpandOptions{Nonsparse: true, SparseVariables: []bool{false, false, true}}
	result, err := layout.ExpandCover(on, off, options)
	if err != nil || espressoFamilyTruth(t, layout, result) != onTruth || !slices.Equal(espressoFamilyMasks(t, on), before) {
		t.Fatal(espressoFamilyMasks(t, result), err)
	}
	for _, cube := range result.sets {
		if cube.Flags&(EspressoSetPrime|EspressoSetActive) != EspressoSetPrime|EspressoSetActive || !espressoExpansionIsMaximal(layout, cube, offTruth, options) {
			t.Fatal(espressoSetMask(t, cube), cube.Flags)
		}
	}
	all, err := layout.AllPrimes(on, off, EspressoExpandOptions{})
	if err != nil || len(all.sets) == 0 {
		t.Fatal(len(all.sets), err)
	}
	for _, prime := range all.sets {
		if espressoCubeTruth(layout, espressoSetMask(t, prime))&offTruth != 0 || prime.Flags&EspressoSetPrime == 0 {
			t.Fatal(espressoSetMask(t, prime), prime.Flags)
		}
	}
}

func TestEspressoExpandErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	on := espressoMintermCover(t, 2, 0b0111)
	off := espressoMintermCover(t, 2, 0b1000)
	checks := []EspressoExpandOptions{{MaxCubes: -1}, {MaxCubes: 1}, {MaxOperations: 1}, {MaxDepth: -1}, {Stop: func() bool { return true }}, {SparseVariables: []bool{true}}}
	for _, options := range checks {
		result, err := layout.ExpandCover(on, off, options)
		if err == nil || result != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	overlap := espressoMintermCover(t, 2, 0b0001)
	if result, err := layout.ExpandCover(on, overlap, EspressoExpandOptions{}); !errors.Is(err, ErrInvariant) || result != nil {
		t.Fatal(result, err)
	}
	wrong := espressoMapFamily(t, 3)
	if result, err := layout.ExpandCover(on, wrong, EspressoExpandOptions{}); !errors.Is(err, ErrDimension) || result != nil {
		t.Fatal(result, err)
	}
}

func FuzzEspressoExpand(f *testing.F) {
	f.Add([]byte{3, 0x3f, 0xc0, 1})
	f.Add([]byte{1, 1, 2, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 4 {
			return
		}
		variables := int(data[0]%3) + 1
		assignments := 1 << uint(variables)
		universe := uint64(1)<<uint(assignments) - 1
		onTruth := uint64(data[1]) & universe
		offTruth := uint64(data[2]) & universe &^ onTruth
		if offTruth != universe&^onTruth {
			offTruth = universe &^ onTruth
		}
		layout, _ := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), variables)
		on := espressoMintermCover(t, variables, onTruth)
		off := espressoMintermCover(t, variables, offTruth)
		sparse := make([]bool, variables)
		for variable := range sparse {
			sparse[variable] = data[(3+variable)%len(data)]&1 != 0
		}
		for _, nonsparse := range []bool{false, true} {
			options := EspressoExpandOptions{Nonsparse: nonsparse, SparseVariables: sparse, MaxCubes: 1 << 16, MaxOperations: 1 << 20}
			result, err := layout.ExpandCover(on, off, options)
			if err != nil {
				t.Fatal(err)
			}
			gotTruth := espressoFamilyTruth(t, layout, result)
			if gotTruth != onTruth || !slices.Equal(espressoFamilyMasks(t, on), espressoFamilyMasks(t, espressoMintermCover(t, variables, onTruth))) {
				t.Fatal(variables, onTruth, nonsparse, espressoFamilyMasks(t, result), gotTruth)
			}
			for _, cube := range result.sets {
				if !espressoExpansionIsMaximal(layout, cube, offTruth, options) || espressoCubeTruth(layout, espressoSetMask(t, cube))&offTruth != 0 {
					t.Fatal(variables, onTruth, nonsparse, espressoSetMask(t, cube))
				}
			}
		}
		if onTruth != 0 {
			source, _ := on.Set(0)
			all, err := layout.FindAllPrimesCoveringCube(source, off, EspressoExpandOptions{MaxCubes: 1 << 16, MaxOperations: 1 << 20})
			if err != nil {
				t.Fatal(err)
			}
			brute := espressoBrutePrimes(layout, onTruth, 0)
			want := make([]uint64, 0)
			sourceMask := espressoSetMask(t, source)
			for _, prime := range brute {
				if sourceMask&^prime == 0 {
					want = append(want, prime)
				}
			}
			if !slices.Equal(espressoSortedMasks(t, all), want) {
				t.Fatal(sourceMask, espressoSortedMasks(t, all), want)
			}
		}
	})
}
