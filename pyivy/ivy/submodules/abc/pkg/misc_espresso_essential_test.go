package abc

// Go-only semantic and fuzz tests for src/misc/espresso/essen.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoSemanticEssential(layout *EspressoCubeLayout, onsetMasks, dontCareMasks []uint64, index int) bool {
	candidate := espressoCubeTruth(layout, onsetMasks[index])
	var covered uint64
	for other, mask := range onsetMasks {
		if other != index {
			covered |= espressoCubeTruth(layout, mask)
		}
	}
	for _, mask := range dontCareMasks {
		covered |= espressoCubeTruth(layout, mask)
	}
	return candidate&^covered != 0
}

func TestEspressoConsensusSharpDistanceCases(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 3}, 2)
	cube := espressoSetFromMask(t, 7, 0b01|0b01<<2|0b011<<4)
	distanceOne := espressoSetFromMask(t, 7, 0b10|0b01<<2|0b011<<4)
	distanceZero := espressoSetFromMask(t, 7, 0b11|0b11<<2|0b101<<4)
	contained := espressoSetFromMask(t, 7, 0b01|0b01<<2|0b001<<4)
	cover := espressoMapFamily(t, 7, espressoSetMask(t, cube), espressoSetMask(t, distanceOne), espressoSetMask(t, distanceZero), espressoSetMask(t, contained))
	result, err := layout.ConsensusSharp(cover, cube, 0, EspressoEssentialOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantDistanceOne, _ := layout.Consensus(distanceOne, cube)
	intersection, _ := distanceZero.Intersection(cube)
	mask, _ := layout.VariableMask(2)
	wantDistanceZero, _ := cube.Merge(intersection, mask)
	if got := espressoFamilyMasks(t, result); !slices.Equal(got, []uint64{espressoSetMask(t, wantDistanceOne), espressoSetMask(t, wantDistanceZero)}) {
		t.Fatal(got)
	}
	direct, err := layout.ConsensusSharpDistanceZero(distanceZero, cube, EspressoEssentialOptions{})
	if err != nil || !slices.Equal(espressoFamilyMasks(t, direct), []uint64{espressoSetMask(t, wantDistanceZero)}) {
		t.Fatal(espressoFamilyMasks(t, direct), err)
	}
}

func TestEspressoEssentialExtractionFlagsAndOwnership(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	onsetMasks := []uint64{0b10 | 0b11<<2, 0b11 | 0b10<<2}
	var dontCareMasks []uint64
	onset := espressoMapFamily(t, 4, onsetMasks...)
	dontCare := espressoMapFamily(t, 4, dontCareMasks...)
	for _, cube := range onset.sets {
		cube.Flags = EspressoSetRelEssential
	}
	onset.sets[1].Flags |= EspressoSetNonessential
	result, err := layout.ExtractEssentials(onset, dontCare, EspressoEssentialOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantEssential := make([]uint64, 0)
	wantRemaining := make([]uint64, 0)
	for index, mask := range onsetMasks {
		selected := index != 1 && espressoSemanticEssential(layout, onsetMasks, dontCareMasks, index)
		if selected {
			wantEssential = append(wantEssential, mask)
		} else {
			wantRemaining = append(wantRemaining, mask)
		}
	}
	if !slices.Equal(espressoFamilyMasks(t, result.Essential), wantEssential) || !slices.Equal(espressoFamilyMasks(t, result.Remaining), wantRemaining) || !slices.Equal(espressoFamilyMasks(t, result.DontCare), append(slices.Clone(dontCareMasks), wantEssential...)) || !slices.Equal(espressoFamilyMasks(t, onset), onsetMasks) || !slices.Equal(espressoFamilyMasks(t, dontCare), dontCareMasks) {
		t.Fatal(espressoFamilyMasks(t, result.Essential), wantEssential, espressoFamilyMasks(t, result.Remaining), wantRemaining, espressoFamilyMasks(t, result.DontCare))
	}
	for index, cube := range result.Remaining.sets {
		if cube.Flags&EspressoSetActive == 0 {
			t.Fatalf("remaining cube %d inactive", index)
		}
	}
}

func TestEspressoEssentialErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	onset := espressoMapFamily(t, 4, 0b0111, 0b1101)
	dontCare := espressoMapFamily(t, 4)
	cube, _ := onset.Set(0)
	if _, err := layout.ConsensusSharp(onset, cube, 2, EspressoEssentialOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := layout.ConsensusSharp(onset, cube, -1, EspressoEssentialOptions{MaxCubes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := layout.ConsensusSharp(onset, cube, -1, EspressoEssentialOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := layout.EssentialCube(onset, dontCare, cube, EspressoEssentialOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err := layout.ExtractEssentials(onset, dontCare, EspressoEssentialOptions{MaxCubes: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	wrong := espressoMapFamily(t, 3, 0b111)
	if _, err := layout.ExtractEssentials(onset, wrong, EspressoEssentialOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzEspressoEssentialExtraction(f *testing.F) {
	f.Add([]byte{2, 3, 1, 0x55, 0xaa, 0x03})
	f.Add([]byte{4, 5, 0, 0xff, 0x01})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 4 {
			return
		}
		variables := int(data[0]%4) + 1
		layout, err := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), variables)
		if err != nil {
			t.Fatal(err)
		}
		onsetCount := int(data[1]%7) + 1
		dontCareCount := int(data[2] % 4)
		makeMasks := func(count, offset int) []uint64 {
			masks := make([]uint64, 0, count)
			seen := make(map[uint64]struct{})
			for candidate := 0; candidate < count*3 && len(masks) < count; candidate++ {
				var mask uint64
				for variable := 0; variable < variables; variable++ {
					choice := data[(offset+candidate*variables+variable)%len(data)] % 3
					mask |= uint64(choice+1) << uint(2*variable)
				}
				if _, exists := seen[mask]; !exists {
					seen[mask] = struct{}{}
					masks = append(masks, mask)
				}
			}
			return masks
		}
		rawOnsetMasks := makeMasks(onsetCount, 3)
		var function uint64
		for _, mask := range rawOnsetMasks {
			function |= espressoCubeTruth(layout, mask)
		}
		dontCareMasks := makeMasks(dontCareCount, 4+onsetCount*variables)
		var dontCareFunction uint64
		for _, mask := range dontCareMasks {
			dontCareFunction |= espressoCubeTruth(layout, mask)
		}
		allPrimes := espressoPrimeModel(layout, function|dontCareFunction)
		onsetMasks := make([]uint64, 0, len(allPrimes))
		for _, prime := range allPrimes {
			if espressoCubeTruth(layout, prime)&function != 0 {
				onsetMasks = append(onsetMasks, prime)
			}
		}
		onset := espressoMapFamily(t, layout.BitCount(), onsetMasks...)
		dontCare := espressoMapFamily(t, layout.BitCount(), dontCareMasks...)
		eligible := make([]bool, len(onsetMasks))
		for index, cube := range onset.sets {
			flags := data[(3+index)%len(data)]
			if flags&1 != 0 {
				cube.Flags |= EspressoSetRelEssential
				eligible[index] = true
			}
			if flags&2 != 0 {
				cube.Flags |= EspressoSetNonessential
				eligible[index] = false
			}
		}
		result, err := layout.ExtractEssentials(onset, dontCare, EspressoEssentialOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantEssential, wantRemaining := make([]uint64, 0), make([]uint64, 0)
		for index, mask := range onsetMasks {
			if eligible[index] && espressoSemanticEssential(layout, onsetMasks, dontCareMasks, index) {
				wantEssential = append(wantEssential, mask)
			} else {
				wantRemaining = append(wantRemaining, mask)
			}
		}
		if !slices.Equal(espressoFamilyMasks(t, result.Essential), wantEssential) || !slices.Equal(espressoFamilyMasks(t, result.Remaining), wantRemaining) || !slices.Equal(espressoFamilyMasks(t, result.DontCare), append(slices.Clone(dontCareMasks), wantEssential...)) || !slices.Equal(espressoFamilyMasks(t, onset), onsetMasks) {
			t.Fatal(onsetMasks, dontCareMasks, espressoFamilyMasks(t, result.Essential), wantEssential, espressoFamilyMasks(t, result.Remaining), wantRemaining)
		}
	})
}
