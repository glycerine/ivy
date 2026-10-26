package abc

// Go-only essential-point models for src/misc/espresso/reduce.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoBinarySupercube(variables int, truth uint64) uint64 {
	if truth == 0 {
		return 0
	}
	var result uint64
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		if truth>>uint(assignment)&1 == 0 {
			continue
		}
		for variable := 0; variable < variables; variable++ {
			result |= uint64(1) << uint(2*variable+((assignment>>uint(variable))&1))
		}
	}
	return result
}

func espressoReduceModel(layout *EspressoCubeLayout, onMasks, dcMasks []uint64, order EspressoReduceOrder) []uint64 {
	ordered := slices.Clone(onMasks)
	scores := make([]int, len(ordered))
	if len(ordered) > 1 {
		if order == EspressoReduceDistanceOrder {
			largest := ordered[0]
			for _, cube := range ordered[1:] {
				if bitsCount64(cube) > bitsCount64(largest) {
					largest = cube
				}
			}
			for index, cube := range ordered {
				scores[index] = (layout.VariableCount()-espressoCubeDistanceModel(layout, largest, cube))<<7 + min(bitsCount64(cube), 127)
			}
		} else {
			counts := make([]int, layout.BitCount())
			for _, cube := range ordered {
				for part := range counts {
					if cube>>uint(part)&1 != 0 {
						counts[part]++
					}
				}
			}
			for index, cube := range ordered {
				for part, count := range counts {
					if cube>>uint(part)&1 != 0 {
						scores[index] += count
					}
				}
			}
		}
		indices := make([]int, len(ordered))
		for index := range indices {
			indices[index] = index
		}
		slices.SortStableFunc(indices, func(first, second int) int { return scores[second] - scores[first] })
		sorted := make([]uint64, len(ordered))
		for index, source := range indices {
			sorted[index] = ordered[source]
		}
		ordered = sorted
	}
	dcTruth := uint64(0)
	for _, cube := range dcMasks {
		dcTruth |= espressoCubeTruth(layout, cube)
	}
	for index, cube := range ordered {
		allowed := dcTruth
		for current, other := range ordered {
			if current != index {
				allowed |= espressoCubeTruth(layout, other)
			}
		}
		essential := espressoCubeTruth(layout, cube) &^ allowed
		ordered[index] = espressoBinarySupercube(layout.VariableCount(), essential)
	}
	result := ordered[:0]
	for _, cube := range ordered {
		if cube != 0 {
			result = append(result, cube)
		}
	}
	return result
}

func bitsCount64(value uint64) int {
	count := 0
	for value != 0 {
		value &= value - 1
		count++
	}
	return count
}

func TestEspressoSCCCAndReduceCube(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	cover := espressoMapFamily(t, 4, 0b01|0b11<<2, 0b10|0b01<<2)
	sccc, err := layout.SmallestCubeContainingComplement(cover, EspressoReduceOptions{})
	complementTruth := espressoTruthUniverse(layout) &^ espressoFamilyTruth(t, layout, cover)
	if err != nil || espressoSetMask(t, sccc) != espressoBinarySupercube(2, complementTruth) {
		t.Fatal(espressoSetMask(t, sccc), complementTruth, err)
	}
	dc := espressoMapFamily(t, 4, 0b01|0b10<<2)
	reduced, err := layout.ReduceCube(cover, dc, 0, EspressoReduceOptions{})
	otherTruth := espressoCubeTruth(layout, 0b10|0b01<<2) | espressoFamilyTruth(t, layout, dc)
	want := espressoBinarySupercube(2, espressoCubeTruth(layout, 0b01|0b11<<2)&^otherTruth)
	if err != nil || espressoSetMask(t, reduced) != want {
		t.Fatal(espressoSetMask(t, reduced), want, err)
	}
}

func TestEspressoReduceCoverOrdersFlagsOwnershipAndSemantics(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 3)
	masks := []uint64{0b01 | 0b11<<2 | 0b11<<4, 0b11 | 0b01<<2 | 0b01<<4, 0b10 | 0b11<<2 | 0b10<<4}
	on := espressoMapFamily(t, 6, masks...)
	on.sets[0].Flags = EspressoSetPrime
	dcMasks := []uint64{0b01 | 0b10<<2 | 0b10<<4}
	dc := espressoMapFamily(t, 6, dcMasks...)
	before := espressoFamilyMasks(t, on)
	for _, order := range []EspressoReduceOrder{EspressoReduceDistanceOrder, EspressoReduceColumnOrder} {
		got, err := layout.ReduceCover(on, dc, EspressoReduceOptions{Order: order})
		want := espressoReduceModel(layout, masks, dcMasks, order)
		if err != nil || !slices.Equal(espressoFamilyMasks(t, got), want) {
			t.Fatal(order, espressoFamilyMasks(t, got), want, err)
		}
		oldTruth := espressoFamilyTruth(t, layout, on) | espressoFamilyTruth(t, layout, dc)
		newTruth := espressoFamilyTruth(t, layout, got) | espressoFamilyTruth(t, layout, dc)
		if oldTruth != newTruth || !slices.Equal(espressoFamilyMasks(t, on), before) {
			t.Fatal(order, oldTruth, newTruth)
		}
		for _, cube := range got.sets {
			if cube.Flags&EspressoSetActive == 0 {
				t.Fatal("reduced cube is not active")
			}
		}
	}
}

func TestEspressoReduceErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	on := espressoMapFamily(t, 4, 0b0111, 0b1101)
	empty := espressoMapFamily(t, 4)
	if _, err := layout.ReduceCover(on, empty, EspressoReduceOptions{Order: 2}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := layout.ReduceCover(on, empty, EspressoReduceOptions{MaxCubes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := layout.ReduceCover(on, empty, EspressoReduceOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := layout.ReduceCover(on, empty, EspressoReduceOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err := layout.ReduceCube(on, empty, -1, EspressoReduceOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	wrong := espressoMapFamily(t, 3)
	if _, err := layout.ReduceCover(on, wrong, EspressoReduceOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzEspressoReduce(f *testing.F) {
	f.Add([]byte{3, 3, 1, 0x55, 0xaa, 0x33})
	f.Add([]byte{1, 1, 0, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		variables := int(data[0]%4) + 1
		layout, _ := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), variables)
		onCount, dcCount := int(data[1]%6), int(data[2]%4)
		makeMasks := func(count, offset int) []uint64 {
			masks := make([]uint64, count)
			for cube := range masks {
				for variable := 0; variable < variables; variable++ {
					masks[cube] |= uint64(data[(offset+cube*variables+variable)%len(data)]%3+1) << uint(2*variable)
				}
			}
			return masks
		}
		onMasks := makeMasks(onCount, 3)
		dcMasks := makeMasks(dcCount, 3+onCount*variables)
		on := espressoMapFamily(t, layout.BitCount(), onMasks...)
		dc := espressoMapFamily(t, layout.BitCount(), dcMasks...)
		for _, order := range []EspressoReduceOrder{EspressoReduceDistanceOrder, EspressoReduceColumnOrder} {
			got, err := layout.ReduceCover(on, dc, EspressoReduceOptions{Order: order})
			if err != nil {
				t.Fatal(err)
			}
			want := espressoReduceModel(layout, onMasks, dcMasks, order)
			if !slices.Equal(espressoFamilyMasks(t, got), want) {
				t.Fatal(order, onMasks, dcMasks, espressoFamilyMasks(t, got), want)
			}
			if espressoFamilyTruth(t, layout, got)|espressoFamilyTruth(t, layout, dc) != espressoFamilyTruth(t, layout, on)|espressoFamilyTruth(t, layout, dc) {
				t.Fatal("ON union DC changed")
			}
		}
	})
}
