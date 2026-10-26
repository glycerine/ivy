package abc

// Go-only cover transformation models for src/misc/espresso/cvrm.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestEspressoUnravelRangeCartesianSemanticsAndOwnership(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 3, 2}, 1)
	mask := uint64(0b11 | 0b111<<2 | 0b11<<5)
	family := espressoMapFamily(t, 7, mask)
	family.sets[0].Flags = EspressoSetPrime
	_ = family.SetActive(0, true)
	family.sets[0].Rank = 17
	result, err := layout.UnravelRange(family, 1, 2, EspressoCoverManipulationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.sets) != 6 || espressoFamilyTruth(t, layout, result) != espressoFamilyTruth(t, layout, family) || espressoSetMask(t, family.sets[0]) != mask {
		t.Fatal(len(result.sets), espressoFamilyMasks(t, result))
	}
	for _, cube := range result.sets {
		if cube.Rank != 17 || cube.Flags != EspressoSetPrime|EspressoSetActive {
			t.Fatal(cube.Rank, cube.Flags)
		}
		for variable := 1; variable <= 2; variable++ {
			count := 0
			for part := layout.firstParts[variable]; part < layout.firstParts[variable]+layout.partSizes[variable]; part++ {
				present, _ := cube.Has(part)
				if present {
					count++
				}
			}
			if count != 1 {
				t.Fatal(variable, espressoSetMask(t, cube))
			}
		}
	}
}

func TestEspressoCoverSortsAndSeededShuffle(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	family := espressoMapFamily(t, 4, 0b0101, 0b1111, 0b0011, 0b1001)
	lexical, lexErr := family.SortLexical(EspressoCoverManipulationOptions{})
	bySize, sizeErr := family.SortBySize(EspressoCoverManipulationOptions{})
	ascending, ascErr := family.SortByColumnWeight(false, EspressoCoverManipulationOptions{})
	descending, descErr := family.SortByColumnWeight(true, EspressoCoverManipulationOptions{})
	reduction, reduceErr := layout.SortForReduction(family, EspressoCoverManipulationOptions{})
	first, firstErr := family.ShuffleCover(123, EspressoCoverManipulationOptions{})
	second, secondErr := family.ShuffleCover(123, EspressoCoverManipulationOptions{})
	if lexErr != nil || sizeErr != nil || ascErr != nil || descErr != nil || reduceErr != nil || firstErr != nil || secondErr != nil {
		t.Fatal(lexErr, sizeErr, ascErr, descErr, reduceErr, firstErr, secondErr)
	}
	for _, sorted := range []*EspressoSetFamily{lexical, bySize, ascending, descending, reduction, first} {
		got := espressoFamilyMasks(t, sorted)
		slices.Sort(got)
		want := espressoFamilyMasks(t, family)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Fatal(got, want)
		}
	}
	if !slices.Equal(espressoFamilyMasks(t, first), espressoFamilyMasks(t, second)) || !slices.IsSortedFunc(bySize.sets, func(a, b *EspressoSet) int {
		comparison, _ := EspressoSetCompareDescending(a, b)
		return comparison
	}) {
		t.Fatal(espressoFamilyMasks(t, first), espressoFamilyMasks(t, second), espressoFamilyMasks(t, bySize))
	}
}

func TestEspressoCubeListPartitionAndOutputCofactors(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	family := espressoMapFamily(t, 4, 0b01|0b11<<2, 0b10|0b11<<2, 0b11|0b01<<2)
	list, _ := NewEspressoCubeList(layout, family)
	partition, err := list.PartitionComponents(EspressoCoverManipulationOptions{})
	if err != nil || !partition.Partitioned || !slices.Equal(partition.LeftIndices, []int{0, 1}) || !slices.Equal(partition.RightIndices, []int{2}) {
		t.Fatal(partition.LeftIndices, partition.RightIndices, err)
	}
	leftCount, _ := partition.Left.Count()
	rightCount, _ := partition.Right.Count()
	if leftCount != 2 || rightCount != 1 {
		t.Fatal(leftCount, rightCount)
	}

	outputLayout, _ := NewEspressoCubeLayout([]int{2, 3}, 1)
	outputs := espressoMapFamily(t, 5, 0b01|0b011<<2, 0b10|0b110<<2)
	cofactored, err := outputLayout.CofactorOutput(outputs, 0, EspressoCoverManipulationOptions{})
	if err != nil || !slices.Equal(espressoFamilyMasks(t, cofactored), []uint64{0b01 | 0b111<<2}) || cofactored.sets[0].Flags&EspressoSetPrime != 0 {
		t.Fatal(espressoFamilyMasks(t, cofactored), err)
	}
	restricted, err := outputLayout.UncofactorOutput(cofactored, 2, EspressoCoverManipulationOptions{})
	if err != nil || !slices.Equal(espressoFamilyMasks(t, restricted), []uint64{0b01 | 0b100<<2}) {
		t.Fatal(espressoFamilyMasks(t, restricted), err)
	}
}

func TestEspressoCoverManipulationErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 1)
	family := espressoMapFamily(t, 4, 0b1111, 0b0101)
	if result, err := layout.UnravelRange(family, 1, 0, EspressoCoverManipulationOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := layout.UnravelRange(family, 0, 1, EspressoCoverManipulationOptions{MaxCubes: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := layout.UnravelRange(family, 0, 1, EspressoCoverManipulationOptions{MaxOperations: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := layout.CofactorOutput(family, 0, EspressoCoverManipulationOptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	if result, err := family.SortByColumnWeight(false, EspressoCoverManipulationOptions{MaxOperations: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := family.SortLexical(EspressoCoverManipulationOptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	list, _ := NewEspressoCubeList(layout, family)
	if result, err := list.PartitionComponents(EspressoCoverManipulationOptions{MaxOperations: 1}); result.Partitioned || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	pureBinary, _ := NewEspressoCubeLayout([]int{2}, 1)
	if result, err := pureBinary.CofactorOutput(espressoMapFamily(t, 2), 0, EspressoCoverManipulationOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
}

func FuzzEspressoCoverManipulation(f *testing.F) {
	f.Add([]byte{3, 0x7f, 1})
	f.Add([]byte{2, 0x33, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		variables := int(data[0]%3) + 1
		layout, _ := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), max(0, variables-1))
		mask := uint64(data[1]) & (uint64(1)<<uint(2*variables) - 1)
		family := espressoMapFamily(t, 2*variables, mask)
		first := int(data[2] % byte(variables))
		result, err := layout.Unravel(family, first, EspressoCoverManipulationOptions{MaxCubes: 1 << 12, MaxOperations: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		if espressoFamilyTruth(t, layout, result) != espressoFamilyTruth(t, layout, family) {
			t.Fatal(variables, first, mask, espressoFamilyMasks(t, result))
		}
		if layout.OutputVariable() >= 0 {
			part := int(data[2]) % layout.partSizes[layout.OutputVariable()]
			cofactored, err := layout.CofactorOutput(family, part, EspressoCoverManipulationOptions{})
			if err != nil {
				t.Fatal(err)
			}
			restricted, err := layout.UncofactorOutput(cofactored, part, EspressoCoverManipulationOptions{})
			if err != nil {
				t.Fatal(err)
			}
			for _, cube := range restricted.sets {
				present, _ := cube.Has(layout.firstParts[layout.OutputVariable()] + part)
				if !present {
					t.Fatal(part, espressoSetMask(t, cube))
				}
			}
		}
	})
}
