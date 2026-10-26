package abc

// Go-only semantic model and fuzz tests for src/misc/espresso/sharp.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoCubeTruth(layout *EspressoCubeLayout, mask uint64) uint64 {
	total := 1
	for _, size := range layout.partSizes {
		total *= size
	}
	var truth uint64
	for assignment := 0; assignment < total; assignment++ {
		remaining, valid := assignment, true
		for variable, size := range layout.partSizes {
			part := remaining % size
			remaining /= size
			if mask&(uint64(1)<<uint(layout.firstParts[variable]+part)) == 0 {
				valid = false
				break
			}
		}
		if valid {
			truth |= uint64(1) << uint(assignment)
		}
	}
	return truth
}

func espressoFamilyTruth(t testing.TB, layout *EspressoCubeLayout, family *EspressoSetFamily) uint64 {
	t.Helper()
	var result uint64
	for _, mask := range espressoFamilyMasks(t, family) {
		result |= espressoCubeTruth(layout, mask)
	}
	return result
}

func espressoFamilyPairwiseDisjoint(t testing.TB, layout *EspressoCubeLayout, family *EspressoSetFamily) bool {
	t.Helper()
	masks := espressoFamilyMasks(t, family)
	for first := range masks {
		for second := first + 1; second < len(masks); second++ {
			if espressoCubeTruth(layout, masks[first])&espressoCubeTruth(layout, masks[second]) != 0 {
				return false
			}
		}
	}
	return true
}

func TestEspressoSharpCubeOperationsAndMetadata(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 2)
	first := espressoSetFromMask(t, 6, 0b11|0b11<<2|0b11<<4)
	first.Rank, first.Flags = 7, EspressoSetPrime
	second := espressoSetFromMask(t, 6, 0b01|0b01<<2|0b01<<4)
	sharp, err := layout.SharpCubes(first, second, EspressoSharpOptions{})
	disjoint, disjointErr := layout.DisjointSharpCubes(first, second, EspressoSharpOptions{})
	wantTruth := espressoCubeTruth(layout, espressoSetMask(t, first)) &^ espressoCubeTruth(layout, espressoSetMask(t, second))
	if err != nil || disjointErr != nil || espressoFamilyTruth(t, layout, sharp) != wantTruth || espressoFamilyTruth(t, layout, disjoint) != wantTruth || !espressoFamilyPairwiseDisjoint(t, layout, disjoint) || len(espressoFamilyMasks(t, sharp)) != 3 || len(espressoFamilyMasks(t, disjoint)) != 3 {
		t.Fatal(espressoFamilyMasks(t, sharp), espressoFamilyMasks(t, disjoint), wantTruth, err, disjointErr)
	}
	disjointCube := espressoSetFromMask(t, 6, 0b01|0b01<<2|0b10<<4)
	unchanged, err := layout.SharpCubes(second, disjointCube, EspressoSharpOptions{})
	kept, _ := unchanged.Set(0)
	if err != nil || espressoSetMask(t, kept) != espressoSetMask(t, second) || kept.Rank != 0 || kept.Flags != 0 {
		// second has zero metadata; the disjoint source path must preserve it.
		t.Fatal(kept, err)
	}
	empty := espressoSetFromMask(t, 6, 0)
	emptyResult, _ := layout.SharpCubes(empty, second, EspressoSharpOptions{})
	removeEmpty, _ := layout.SharpCubes(first, empty, EspressoSharpOptions{})
	if count, _, _ := emptyResult.Counts(); count != 0 {
		t.Fatal(count)
	}
	if espressoFamilyTruth(t, layout, removeEmpty) != espressoCubeTruth(layout, espressoSetMask(t, first)) {
		t.Fatal(espressoFamilyMasks(t, removeEmpty))
	}
}

func TestEspressoSharpCoverOperationsAndOwnership(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 2)
	firstMasks := []uint64{0b11 | 0b01<<2 | 0b11<<4, 0b10 | 0b11<<2 | 0b10<<4}
	secondMasks := []uint64{0b01 | 0b11<<2 | 0b01<<4, 0b10 | 0b10<<2 | 0b11<<4}
	first := espressoMapFamily(t, 6, firstMasks...)
	second := espressoMapFamily(t, 6, secondMasks...)
	firstTruth, secondTruth := espressoFamilyTruth(t, layout, first), espressoFamilyTruth(t, layout, second)
	sharp, err := layout.SharpCovers(first, second, EspressoSharpOptions{})
	disjointSharp, disjointErr := layout.DisjointSharpCovers(first, second, EspressoSharpOptions{})
	intersection, intersectionErr := layout.IntersectCovers(first, second, EspressoSharpOptions{})
	cubeSharp, cubeErr := layout.SharpCubeByCover(first.sets[0], second, EspressoSharpOptions{})
	disjointCubeSharp, disjointCubeErr := layout.DisjointSharpCubeByCover(first.sets[0], second, EspressoSharpOptions{})
	byCube, byCubeErr := layout.DisjointSharpFamilyByCube(first, second.sets[0], EspressoSharpOptions{})
	if err != nil || disjointErr != nil || intersectionErr != nil || cubeErr != nil || disjointCubeErr != nil || byCubeErr != nil {
		t.Fatal(err, disjointErr, intersectionErr, cubeErr, disjointCubeErr, byCubeErr)
	}
	if espressoFamilyTruth(t, layout, sharp) != firstTruth&^secondTruth || espressoFamilyTruth(t, layout, disjointSharp) != firstTruth&^secondTruth || espressoFamilyTruth(t, layout, intersection) != firstTruth&secondTruth || espressoFamilyTruth(t, layout, cubeSharp) != espressoCubeTruth(layout, firstMasks[0])&^secondTruth || espressoFamilyTruth(t, layout, disjointCubeSharp) != espressoCubeTruth(layout, firstMasks[0])&^secondTruth || espressoFamilyTruth(t, layout, byCube) != firstTruth&^espressoCubeTruth(layout, secondMasks[0]) || !espressoFamilyPairwiseDisjoint(t, layout, disjointCubeSharp) {
		t.Fatal(espressoFamilyMasks(t, sharp), espressoFamilyMasks(t, disjointSharp), espressoFamilyMasks(t, intersection), espressoFamilyMasks(t, cubeSharp), espressoFamilyMasks(t, disjointCubeSharp), espressoFamilyMasks(t, byCube))
	}
	disjointized, err := layout.MakeDisjoint(first, EspressoSharpOptions{})
	if err != nil || espressoFamilyTruth(t, layout, disjointized) != firstTruth || !espressoFamilyPairwiseDisjoint(t, layout, disjointized) || !slices.Equal(espressoFamilyMasks(t, first), firstMasks) || !slices.Equal(espressoFamilyMasks(t, second), secondMasks) {
		t.Fatal(espressoFamilyMasks(t, disjointized), err)
	}
}

func TestEspressoSharpErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 1)
	first := espressoSetFromMask(t, 4, 0b1111)
	second := espressoSetFromMask(t, 4, 0b0101)
	if result, err := layout.SharpCubes(first, second, EspressoSharpOptions{MaxOperations: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := layout.SharpCubes(first, second, EspressoSharpOptions{MaxCubes: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := layout.SharpCubes(first, second, EspressoSharpOptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	wrong := espressoSetFromMask(t, 3, 0b111)
	if result, err := layout.SharpCubes(first, wrong, EspressoSharpOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if _, err := layout.SharpCubes(first, second, EspressoSharpOptions{MaxCubes: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzEspressoSharpSemantics(f *testing.F) {
	f.Add([]byte{2, 2, 0x55, 0x2a, 0x3f})
	f.Add([]byte{3, 1, 0xff, 0x01})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		variables := int(data[0]%4) + 1
		parts := slices.Repeat([]int{2}, variables)
		layout, err := NewEspressoCubeLayout(parts, variables)
		if err != nil {
			t.Fatal(err)
		}
		cubeCount := int(data[1]%6) + 1
		masks := make([]uint64, 2*cubeCount)
		for cube := range masks {
			for variable := 0; variable < variables; variable++ {
				choice := data[(2+cube*variables+variable)%len(data)] % 3
				part := 1
				if choice == 0 {
					part = 1
				} else if choice == 1 {
					part = 2
				} else {
					part = 3
				}
				masks[cube] |= uint64(part) << uint(2*variable)
			}
		}
		first := espressoMapFamily(t, layout.BitCount(), masks[:cubeCount]...)
		second := espressoMapFamily(t, layout.BitCount(), masks[cubeCount:]...)
		firstTruth, secondTruth := espressoFamilyTruth(t, layout, first), espressoFamilyTruth(t, layout, second)
		sharp, sharpErr := layout.SharpCovers(first, second, EspressoSharpOptions{})
		disjointSharp, disjointErr := layout.DisjointSharpCovers(first, second, EspressoSharpOptions{})
		intersection, intersectionErr := layout.IntersectCovers(first, second, EspressoSharpOptions{})
		disjointized, makeErr := layout.MakeDisjoint(first, EspressoSharpOptions{})
		if sharpErr != nil || disjointErr != nil || intersectionErr != nil || makeErr != nil {
			t.Fatal(sharpErr, disjointErr, intersectionErr, makeErr)
		}
		if espressoFamilyTruth(t, layout, sharp) != firstTruth&^secondTruth || espressoFamilyTruth(t, layout, disjointSharp) != firstTruth&^secondTruth || espressoFamilyTruth(t, layout, intersection) != firstTruth&secondTruth || espressoFamilyTruth(t, layout, disjointized) != firstTruth || !espressoFamilyPairwiseDisjoint(t, layout, disjointized) {
			t.Fatal(firstTruth, secondTruth, espressoFamilyMasks(t, sharp), espressoFamilyMasks(t, disjointSharp), espressoFamilyMasks(t, intersection), espressoFamilyMasks(t, disjointized))
		}
	})
}
