package abc

// Go-only brute-force models for src/misc/espresso/unate.c.

import (
	"errors"
	"slices"
	"sort"
	"testing"
)

func espressoMaskFamily(t testing.TB, bits int, masks []uint64) *EspressoSetFamily {
	t.Helper()
	return espressoMapFamily(t, bits, masks...)
}

func espressoSortedMasks(t testing.TB, family *EspressoSetFamily) []uint64 {
	t.Helper()
	result := espressoFamilyMasks(t, family)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func espressoBruteMinimalHittingSets(columns int, rows []uint64) []uint64 {
	result := make([]uint64, 0)
	for candidate := uint64(0); candidate < uint64(1)<<uint(columns); candidate++ {
		covers := true
		for _, row := range rows {
			if candidate&row == 0 {
				covers = false
				break
			}
		}
		if !covers {
			continue
		}
		minimal := true
		for part := 0; part < columns; part++ {
			if candidate>>uint(part)&1 == 0 {
				continue
			}
			smaller := candidate &^ (uint64(1) << uint(part))
			coversWithout := true
			for _, row := range rows {
				if smaller&row == 0 {
					coversWithout = false
					break
				}
			}
			if coversWithout {
				minimal = false
				break
			}
		}
		if minimal {
			result = append(result, candidate)
		}
	}
	return result
}

func espressoBruteUnateIntersection(first, second []uint64, largest bool) []uint64 {
	values := make(map[uint64]struct{})
	maximum := -1
	for _, left := range first {
		for _, right := range second {
			value := left & right
			if value == 0 {
				continue
			}
			cardinality := bitsCount64(value)
			if largest && cardinality < maximum {
				continue
			}
			if largest && cardinality > maximum {
				clear(values)
				maximum = cardinality
			}
			values[value] = struct{}{}
		}
	}
	result := make([]uint64, 0)
	for value := range values {
		maximal := true
		for other := range values {
			if value != other && value&^other == 0 {
				maximal = false
				break
			}
		}
		if maximal {
			result = append(result, value)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func TestEspressoUnateComplementAndExactMinimumCovers(t *testing.T) {
	rows := []uint64{0b0011, 0b0110, 0b1100}
	family := espressoMaskFamily(t, 4, rows)
	want := espressoBruteMinimalHittingSets(4, rows)
	complement, err := UnateComplement(family, EspressoUnateOptions{})
	exact, exactErr := ExactMinimumCovers(family, EspressoUnateOptions{})
	if err != nil || exactErr != nil || !slices.Equal(espressoSortedMasks(t, complement), want) || !slices.Equal(espressoSortedMasks(t, exact), want) {
		t.Fatal(espressoSortedMasks(t, complement), espressoSortedMasks(t, exact), want, err, exactErr)
	}
	empty := espressoMaskFamily(t, 3, nil)
	emptyResult, _ := UnateComplement(empty, EspressoUnateOptions{})
	impossible := espressoMaskFamily(t, 3, []uint64{0})
	impossibleResult, _ := UnateComplement(impossible, EspressoUnateOptions{})
	if !slices.Equal(espressoFamilyMasks(t, emptyResult), []uint64{0}) || len(espressoFamilyMasks(t, impossibleResult)) != 0 {
		t.Fatal(espressoFamilyMasks(t, emptyResult), espressoFamilyMasks(t, impossibleResult))
	}
}

func TestEspressoUnateIntersectLargestAndOwnership(t *testing.T) {
	firstMasks := []uint64{0b1110, 0b1011}
	secondMasks := []uint64{0b1101, 0b0111}
	first := espressoMaskFamily(t, 4, firstMasks)
	second := espressoMaskFamily(t, 4, secondMasks)
	for _, largest := range []bool{false, true} {
		got, err := IntersectUnateFamilies(first, second, largest, EspressoUnateOptions{})
		want := espressoBruteUnateIntersection(firstMasks, secondMasks, largest)
		if err != nil || !slices.Equal(espressoSortedMasks(t, got), want) {
			t.Fatal(largest, espressoSortedMasks(t, got), want, err)
		}
	}
	if !slices.Equal(espressoFamilyMasks(t, first), firstMasks) || !slices.Equal(espressoFamilyMasks(t, second), secondMasks) {
		t.Fatal("source family mutated")
	}
}

func TestEspressoUnateMappingComplementRoundTrip(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	cover := espressoMapFamily(t, 4, 0b10|0b11<<2, 0b11|0b10<<2)
	mapping, err := layout.MapCoverToUnate(cover, EspressoUnateOptions{})
	if err != nil || !slices.Equal(mapping.ActiveParts, []int{0, 2}) || !slices.Equal(espressoFamilyMasks(t, mapping.Family), []uint64{1, 2}) {
		t.Fatal(mapping.ActiveParts, espressoFamilyMasks(t, mapping.Family), err)
	}
	unateComplement, err := UnateComplement(mapping.Family, EspressoUnateOptions{})
	positional, mapErr := mapping.ToCover(unateComplement, EspressoUnateOptions{})
	want := espressoTruthUniverse(layout) &^ espressoFamilyTruth(t, layout, cover)
	if err != nil || mapErr != nil || espressoFamilyTruth(t, layout, positional) != want {
		t.Fatal(espressoFamilyMasks(t, positional), want, err, mapErr)
	}
	positional.sets[0].words[0] = 0
	if espressoFamilyTruth(t, layout, cover) == 0 || mapping.Layout == layout {
		t.Fatal("mapping ownership violated")
	}
}

func TestEspressoUnateErrorsLimitsAndCancellation(t *testing.T) {
	family := espressoMaskFamily(t, 4, []uint64{0b0011, 0b1100})
	if _, err := UnateComplement(family, EspressoUnateOptions{MaxSets: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := UnateComplement(family, EspressoUnateOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := UnateComplement(family, EspressoUnateOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	wrong := espressoMaskFamily(t, 3, []uint64{0b111})
	if _, err := IntersectUnateFamilies(family, wrong, false, EspressoUnateOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	binate := espressoMapFamily(t, 4, 0b01|0b11<<2, 0b10|0b11<<2)
	if _, err := layout.MapCoverToUnate(binate, EspressoUnateOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzEspressoUnateFamilies(f *testing.F) {
	f.Add([]byte{4, 3, 2, 3, 6, 12, 14})
	f.Add([]byte{1, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		columns := int(data[0]%7) + 1
		firstCount, secondCount := int(data[1]%7), int(data[2]%7)
		mask := uint64(1)<<uint(columns) - 1
		makeRows := func(count, offset int) []uint64 {
			rows := make([]uint64, count)
			for index := range rows {
				rows[index] = uint64(data[(offset+index)%len(data)]) & mask
			}
			return rows
		}
		firstRows := makeRows(firstCount, 3)
		secondRows := makeRows(secondCount, 3+firstCount)
		first := espressoMaskFamily(t, columns, firstRows)
		second := espressoMaskFamily(t, columns, secondRows)
		complement, err := UnateComplement(first, EspressoUnateOptions{MaxSets: 1 << 16, MaxOperations: 1 << 20})
		if err != nil || !slices.Equal(espressoSortedMasks(t, complement), espressoBruteMinimalHittingSets(columns, firstRows)) {
			t.Fatal(columns, firstRows, espressoSortedMasks(t, complement), err)
		}
		for _, largest := range []bool{false, true} {
			intersection, err := IntersectUnateFamilies(first, second, largest, EspressoUnateOptions{MaxSets: 1 << 16, MaxOperations: 1 << 20})
			if err != nil || !slices.Equal(espressoSortedMasks(t, intersection), espressoBruteUnateIntersection(firstRows, secondRows, largest)) {
				t.Fatal(columns, firstRows, secondRows, largest, espressoSortedMasks(t, intersection), err)
			}
		}
	})
}

func FuzzEspressoUnateMapping(f *testing.F) {
	f.Add([]byte{3, 3, 0b101, 0x55, 0xaa, 0x33})
	f.Add([]byte{1, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		variables := int(data[0]%4) + 1
		count := int(data[1] % 7)
		polarity := data[2]
		masks := make([]uint64, count)
		for cube := range masks {
			for variable := 0; variable < variables; variable++ {
				part := uint64(3)
				if data[(3+cube*variables+variable)%len(data)]&1 != 0 {
					part = uint64(1) << uint((polarity>>uint(variable))&1)
				}
				masks[cube] |= part << uint(2*variable)
			}
		}
		layout, _ := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), variables)
		cover := espressoMapFamily(t, layout.BitCount(), masks...)
		mapping, err := layout.MapCoverToUnate(cover, EspressoUnateOptions{})
		if err != nil {
			t.Fatal(err)
		}
		unateComplement, err := UnateComplement(mapping.Family, EspressoUnateOptions{MaxSets: 1 << 16, MaxOperations: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		positional, err := mapping.ToCover(unateComplement, EspressoUnateOptions{})
		if err != nil {
			t.Fatal(err)
		}
		truth := espressoFamilyTruth(t, layout, cover)
		want := espressoTruthUniverse(layout) &^ truth
		if got := espressoFamilyTruth(t, layout, positional); got != want || !slices.Equal(espressoFamilyMasks(t, cover), masks) {
			t.Fatal(variables, masks, truth, espressoFamilyMasks(t, mapping.Family), espressoFamilyMasks(t, positional), got, want)
		}
	})
}
