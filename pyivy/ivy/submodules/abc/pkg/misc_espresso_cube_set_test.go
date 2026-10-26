package abc

// Go-only exhaustive and fuzz tests for src/misc/espresso/setc.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoCubeDistanceModel(layout *EspressoCubeLayout, first, second uint64) int {
	distance := 0
	for variable, size := range layout.partSizes {
		mask := (uint64(1)<<uint(size) - 1) << uint(layout.firstParts[variable])
		if first&second&mask == 0 {
			distance++
		}
	}
	return distance
}

func espressoConsensusModel(layout *EspressoCubeLayout, first, second uint64) uint64 {
	var result uint64
	for variable, size := range layout.partSizes {
		mask := (uint64(1)<<uint(size) - 1) << uint(layout.firstParts[variable])
		intersection := first & second & mask
		if intersection != 0 {
			result |= intersection
		} else {
			result |= (first | second) & mask
		}
	}
	return result
}

func TestEspressoCubeSetOperations(t *testing.T) {
	layout, err := NewEspressoCubeLayout([]int{2, 2, 3}, 2)
	if err != nil || layout.Validate() != nil || layout.BitCount() != 7 || layout.VariableCount() != 3 || !slices.Equal(layout.PartSizes(), []int{2, 2, 3}) {
		t.Fatal(layout, err)
	}
	first := espressoSetFromMask(t, 7, 1<<0|1<<2|1<<3|1<<4)
	second := espressoSetFromMask(t, 7, 1<<1|1<<2|1<<4|1<<5)
	third := espressoSetFromMask(t, 7, 1<<1|1<<2|1<<6)
	distance, err := layout.Distance(first, second)
	zero, zeroErr := layout.DistanceZero(first, second)
	atMostOne, oneErr := layout.DistanceAtMostOne(first, second)
	two, twoErr := layout.DistanceAtMostOne(first, third)
	if err != nil || zeroErr != nil || oneErr != nil || twoErr != nil || distance != 1 || zero || atMostOne != 1 || two != 2 {
		t.Fatal(distance, zero, atMostOne, two, err, zeroErr, oneErr, twoErr)
	}
	consensus, err := layout.Consensus(first, second)
	if err != nil || espressoSetMask(t, consensus) != 1<<0|1<<1|1<<2|1<<4 {
		t.Fatal(espressoSetMask(t, consensus), err)
	}
	lower := espressoSetFromMask(t, 7, 1<<6)
	forced, err := layout.ForceLower(lower, first, second)
	if err != nil || espressoSetMask(t, forced) != 1<<0|1<<6 || espressoSetMask(t, lower) != 1<<6 {
		t.Fatal(espressoSetMask(t, forced), espressoSetMask(t, lower), err)
	}
	full, err := layout.FullCube()
	if err != nil || espressoSetMask(t, full) != 0b1111111 {
		t.Fatal(espressoSetMask(t, full), err)
	}
	cofactor := espressoSetFromMask(t, 7, 0b1111111&^espressoSetMask(t, first))
	fullRow, err := layout.FullRow(first, cofactor)
	if err != nil || !fullRow {
		t.Fatal(fullRow, err)
	}
	activeCube := espressoSetFromMask(t, 7, 0b1111111&^(1<<2))
	active, err := layout.ActiveVariable(activeCube)
	none, noneErr := layout.ActiveVariable(full)
	twoActive, twoActiveErr := layout.ActiveVariable(espressoSetFromMask(t, 7, 0b1111111&^(1<<0|1<<2)))
	if err != nil || noneErr != nil || twoActiveErr != nil || active != 1 || none != -1 || twoActive != -1 {
		t.Fatal(active, none, twoActive, err, noneErr, twoActiveErr)
	}
	common, err := layout.CommonActive(first, second, espressoSetFromMask(t, 7, 0))
	commonMasked, maskedErr := layout.CommonActive(first, second, full)
	if err != nil || maskedErr != nil || !common || commonMasked {
		t.Fatal(common, commonMasked, err, maskedErr)
	}
}

func TestEspressoCubeSetDistanceConsensusExhaustive(t *testing.T) {
	layout, err := NewEspressoCubeLayout([]int{2, 2, 3}, 2)
	if err != nil {
		t.Fatal(err)
	}
	for firstMask := uint64(0); firstMask < 1<<7; firstMask++ {
		first := espressoSetFromMask(t, 7, firstMask)
		for secondMask := uint64(0); secondMask < 1<<7; secondMask++ {
			second := espressoSetFromMask(t, 7, secondMask)
			distance, distanceErr := layout.Distance(first, second)
			zero, zeroErr := layout.DistanceZero(first, second)
			atMostOne, oneErr := layout.DistanceAtMostOne(first, second)
			consensus, consensusErr := layout.Consensus(first, second)
			wantDistance := espressoCubeDistanceModel(layout, firstMask, secondMask)
			if distanceErr != nil || zeroErr != nil || oneErr != nil || consensusErr != nil || distance != wantDistance || zero != (wantDistance == 0) || atMostOne != min(wantDistance, 2) || espressoSetMask(t, consensus) != espressoConsensusModel(layout, firstMask, secondMask) {
				t.Fatal(firstMask, secondMask, distance, wantDistance, zero, atMostOne, espressoSetMask(t, consensus), distanceErr, zeroErr, oneErr, consensusErr)
			}
		}
	}
}

func TestEspressoCubeSetComparators(t *testing.T) {
	first := espressoSetFromMask(t, 65, 1<<1)
	second := espressoSetFromMask(t, 65, 1<<2)
	first.Rank, second.Rank = 2, 1
	descending, err := EspressoSetCompareDescending(first, second)
	ascending, ascendingErr := EspressoSetCompareAscending(first, second)
	if err != nil || ascendingErr != nil || descending != -1 || ascending != 1 {
		t.Fatal(descending, ascending, err, ascendingErr)
	}
	second.Rank = 2
	descending, _ = EspressoSetCompareDescending(first, second)
	ascending, _ = EspressoSetCompareAscending(first, second)
	lexical, lexicalErr := EspressoSetCompareLexical(first, second)
	if lexicalErr != nil || descending != 1 || ascending != -1 || lexical != 1 {
		t.Fatal(descending, ascending, lexical, lexicalErr)
	}
	mask := espressoSetFromMask(t, 65, 1<<1|1<<2)
	distanceOne, err := EspressoSetCompareDistanceOne(first, second, mask)
	if err != nil || distanceOne != 0 {
		t.Fatal(distanceOne, err)
	}
	if got, _ := EspressoSetCompareDescendingNullable(nil, nil); got != 0 {
		t.Fatal(got)
	}
	if got, _ := EspressoSetCompareDescendingNullable(nil, first); got != 1 {
		t.Fatal(got)
	}
	if got, _ := EspressoSetCompareDescendingNullable(first, nil); got != -1 {
		t.Fatal(got)
	}
	wideFirst := espressoSetFromMask(t, 65, 0)
	wideSecond := espressoSetFromMask(t, 65, 1)
	_ = wideFirst.Set(64, true)
	wideComparison, err := EspressoSetCompareLexical(wideFirst, wideSecond)
	if err != nil || wideComparison != -1 {
		t.Fatal(wideComparison, err)
	}
}

func TestEspressoCubeLayoutErrors(t *testing.T) {
	bad := []struct {
		parts  []int
		binary int
		want   error
	}{
		{[]int{2}, -1, ErrDimension},
		{[]int{2}, 2, ErrDimension},
		{[]int{3}, 1, ErrDimension},
		{[]int{0}, 0, ErrDimension},
		{[]int{espressoSetMaximumBits, 1}, 0, ErrLimit},
	}
	for index, test := range bad {
		if _, err := NewEspressoCubeLayout(test.parts, test.binary); !errors.Is(err, test.want) {
			t.Fatal(index, err)
		}
	}
	layout, _ := NewEspressoCubeLayout([]int{2}, 1)
	wrong := espressoSetFromMask(t, 3, 1)
	if _, err := layout.Distance(wrong, wrong); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := layout.ForceLower(nil, wrong, wrong); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzEspressoCubeSetOperations(f *testing.F) {
	f.Add([]byte{0x95, 0x03, 0x6a, 0x05})
	f.Add([]byte{0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 4 {
			return
		}
		layout, err := NewEspressoCubeLayout([]int{2, 2, 3, 4}, 2)
		if err != nil {
			t.Fatal(err)
		}
		firstMask := (uint64(data[0]) | uint64(data[1])<<8) & ((1 << 11) - 1)
		secondMask := (uint64(data[2]) | uint64(data[3])<<8) & ((1 << 11) - 1)
		first := espressoSetFromMask(t, 11, firstMask)
		second := espressoSetFromMask(t, 11, secondMask)
		distance, err := layout.Distance(first, second)
		atMostOne, oneErr := layout.DistanceAtMostOne(first, second)
		consensus, consensusErr := layout.Consensus(first, second)
		wantDistance := espressoCubeDistanceModel(layout, firstMask, secondMask)
		if err != nil || oneErr != nil || consensusErr != nil || distance != wantDistance || atMostOne != min(wantDistance, 2) || espressoSetMask(t, consensus) != espressoConsensusModel(layout, firstMask, secondMask) {
			t.Fatal(firstMask, secondMask, distance, wantDistance, atMostOne, espressoSetMask(t, consensus), err, oneErr, consensusErr)
		}
		lower := espressoSetFromMask(t, 11, uint64(data[0])&((1<<11)-1))
		forced, err := layout.ForceLower(lower, first, second)
		if err != nil {
			t.Fatal(err)
		}
		wantForced := espressoSetMask(t, lower)
		for variable, size := range layout.partSizes {
			mask := (uint64(1)<<uint(size) - 1) << uint(layout.firstParts[variable])
			if firstMask&secondMask&mask == 0 {
				wantForced |= firstMask & mask
			}
		}
		if espressoSetMask(t, forced) != wantForced {
			t.Fatal(espressoSetMask(t, forced), wantForced)
		}
	})
}
