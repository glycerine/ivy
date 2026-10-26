package abc

// Go-only model and fuzz tests for src/misc/espresso/contain.c.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func espressoContainFamily(t testing.TB, bitCount int, masks []uint64) *EspressoSetFamily {
	t.Helper()
	family, err := NewEspressoSetFamily(len(masks), bitCount)
	if err != nil {
		t.Fatal(err)
	}
	for index, mask := range masks {
		set := espressoSetFromMask(t, bitCount, mask)
		set.Rank = 500 + index
		if index&1 == 0 {
			set.Flags = EspressoSetActive
		}
		if err = family.Add(set); err != nil {
			t.Fatal(err)
		}
	}
	return family
}

func espressoContainModel(masks []uint64, reverse bool) ([]uint64, []int) {
	type item struct {
		mask     uint64
		original int
	}
	items := make([]item, len(masks))
	for index, mask := range masks {
		items[index] = item{mask: mask, original: index}
	}
	slices.SortStableFunc(items, func(first, second item) int {
		firstCount, secondCount := bits.OnesCount64(first.mask), bits.OnesCount64(second.mask)
		if firstCount != secondCount {
			if reverse {
				return firstCount - secondCount
			}
			return secondCount - firstCount
		}
		if first.mask < second.mask {
			if reverse {
				return -1
			}
			return 1
		}
		if first.mask > second.mask {
			if reverse {
				return 1
			}
			return -1
		}
		return 0
	})
	unique := items[:0]
	for _, candidate := range items {
		if len(unique) == 0 || unique[len(unique)-1].mask != candidate.mask {
			unique = append(unique, candidate)
		}
	}
	result := make([]item, 0, len(unique))
	for _, candidate := range unique {
		contained := false
		for _, retained := range result {
			if bits.OnesCount64(candidate.mask) == bits.OnesCount64(retained.mask) {
				continue
			}
			if reverse {
				contained = retained.mask&^candidate.mask == 0
			} else {
				contained = candidate.mask&^retained.mask == 0
			}
			if contained {
				break
			}
		}
		if !contained {
			result = append(result, candidate)
		}
	}
	resultMasks, indices := make([]uint64, len(result)), make([]int, len(result))
	for index, candidate := range result {
		resultMasks[index], indices[index] = candidate.mask, candidate.original
	}
	return resultMasks, indices
}

func TestEspressoSetFamilyContainmentOrderIndicesAndOwnership(t *testing.T) {
	masks := []uint64{0b0011, 0b0111, 0b0011, 0b1000, 0b1100}
	family := espressoContainFamily(t, 4, masks)
	contained, err := family.Contain(EspressoContainOptions{})
	reverse, reverseErr := family.ReverseContain(EspressoContainOptions{})
	deduplicated, duplicateErr := family.Deduplicate(EspressoContainOptions{})
	indexed, indices, indexedErr := family.IndexedContain([]int{10, 11, 12, 13, 14}, EspressoContainOptions{})
	if err != nil || reverseErr != nil || duplicateErr != nil || indexedErr != nil {
		t.Fatal(err, reverseErr, duplicateErr, indexedErr)
	}
	if !slices.Equal(espressoFamilyMasks(t, contained), []uint64{0b0111, 0b1100}) || !slices.Equal(espressoFamilyMasks(t, reverse), []uint64{0b1000, 0b0011}) || !slices.Equal(espressoFamilyMasks(t, deduplicated), []uint64{0b0111, 0b1100, 0b0011, 0b1000}) || !slices.Equal(espressoFamilyMasks(t, indexed), []uint64{0b0111, 0b1100}) || !slices.Equal(indices, []int{11, 14}) {
		t.Fatal(espressoFamilyMasks(t, contained), espressoFamilyMasks(t, reverse), espressoFamilyMasks(t, deduplicated), espressoFamilyMasks(t, indexed), indices)
	}
	if !slices.Equal(espressoFamilyMasks(t, family), masks) {
		t.Fatal("input family mutated", espressoFamilyMasks(t, family))
	}
	first, _ := contained.Set(0)
	second, _ := contained.Set(1)
	if first.Rank != 3 || second.Rank != 2 || first.Flags != 0 || second.Flags != EspressoSetActive {
		t.Fatal(first, second)
	}
	for _, owned := range []*EspressoSetFamily{family, contained, reverse, deduplicated, indexed} {
		if err = owned.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEspressoSetFamilyContainedUnionAndDistanceMerge(t *testing.T) {
	first := espressoContainFamily(t, 4, []uint64{0b0011, 0b1000})
	second := espressoContainFamily(t, 4, []uint64{0b0111, 0b1100})
	union, err := ContainedUnionEspressoSetFamilies(first, second, EspressoContainOptions{})
	if err != nil || !slices.Equal(espressoFamilyMasks(t, union), []uint64{0b0111, 0b1100}) {
		t.Fatal(espressoFamilyMasks(t, union), err)
	}
	mergeFamily := espressoContainFamily(t, 4, []uint64{0b0001, 0b0010, 0b0100})
	mask := espressoSetFromMask(t, 4, 0b0011)
	merged, err := mergeFamily.DistanceMerge(mask, EspressoContainOptions{})
	if err != nil || !slices.Equal(espressoFamilyMasks(t, merged), []uint64{0b0100, 0b0011}) {
		t.Fatal(espressoFamilyMasks(t, merged), err)
	}
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	distanceOne, err := mergeFamily.DistanceOneMerge(layout, 0, EspressoContainOptions{})
	if err != nil || !slices.Equal(espressoFamilyMasks(t, distanceOne), []uint64{0b0100, 0b0011}) {
		t.Fatal(espressoFamilyMasks(t, distanceOne), err)
	}
	if !slices.Equal(espressoFamilyMasks(t, first), []uint64{0b0011, 0b1000}) || !slices.Equal(espressoFamilyMasks(t, second), []uint64{0b0111, 0b1100}) || !slices.Equal(espressoFamilyMasks(t, mergeFamily), []uint64{1, 2, 4}) {
		t.Fatal("operation mutated input")
	}
	for _, owned := range []*EspressoSetFamily{first, second, union, mergeFamily, merged, distanceOne} {
		if err = owned.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEspressoSetFamilyContainmentErrorsAndLimits(t *testing.T) {
	family := espressoContainFamily(t, 4, []uint64{1, 2, 4})
	if _, err := family.Contain(EspressoContainOptions{MaxSets: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := family.Contain(EspressoContainOptions{MaxSets: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := family.Contain(EspressoContainOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	checks := espressoContainFamily(t, 4, []uint64{0b0111, 0b1100, 0b0001})
	if _, err := checks.Contain(EspressoContainOptions{MaxChecks: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err := family.IndexedContain([]int{1}, EspressoContainOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	wrongMask := espressoSetFromMask(t, 3, 1)
	if _, err := family.DistanceMerge(wrongMask, EspressoContainOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	if _, err := family.DistanceOneMerge(layout, 2, EspressoContainOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	other, _ := NewEspressoSetFamily(0, 3)
	if _, err := ContainedUnionEspressoSetFamilies(family, other, EspressoContainOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	_ = other.Close()
	_ = checks.Close()
	_ = family.Close()
}

func FuzzEspressoSetFamilyContainment(f *testing.F) {
	f.Add([]byte{4, 5, 3, 7, 3, 8, 12})
	f.Add([]byte{8, 8, 1, 2, 4, 8, 3, 5, 9, 15})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		bitCount := int(data[0]%8) + 1
		rowCount := int(data[1] % 17)
		mask := uint64(1)<<uint(bitCount) - 1
		masks := make([]uint64, rowCount)
		for row := range masks {
			value := byte(row)
			if row+2 < len(data) {
				value = data[row+2]
			}
			masks[row] = uint64(value) & mask
		}
		family := espressoContainFamily(t, bitCount, masks)
		contained, err := family.Contain(EspressoContainOptions{})
		reverse, reverseErr := family.ReverseContain(EspressoContainOptions{})
		indexed, indices, indexedErr := family.IndexedContain(func() []int {
			result := make([]int, rowCount)
			for index := range result {
				result[index] = 1000 + index
			}
			return result
		}(), EspressoContainOptions{})
		if err != nil || reverseErr != nil || indexedErr != nil {
			t.Fatal(err, reverseErr, indexedErr)
		}
		wantContained, wantIndices := espressoContainModel(masks, false)
		wantReverse, _ := espressoContainModel(masks, true)
		for index := range wantIndices {
			wantIndices[index] += 1000
		}
		if !slices.Equal(espressoFamilyMasks(t, contained), wantContained) || !slices.Equal(espressoFamilyMasks(t, reverse), wantReverse) || !slices.Equal(espressoFamilyMasks(t, indexed), wantContained) || !slices.Equal(indices, wantIndices) || !slices.Equal(espressoFamilyMasks(t, family), masks) {
			t.Fatal(masks, espressoFamilyMasks(t, contained), wantContained, espressoFamilyMasks(t, reverse), wantReverse, indices, wantIndices)
		}
		_ = contained.Close()
		_ = reverse.Close()
		_ = indexed.Close()
		_ = family.Close()
	})
}

func FuzzEspressoSetFamilyDistanceMerge(f *testing.F) {
	f.Add([]byte{4, 3, 1, 2, 4, 3})
	f.Add([]byte{8, 6, 0x55, 0xaa, 0x0f, 0xf0, 3, 5, 0x33})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		bitCount := int(data[0]%8) + 1
		rowCount := int(data[1] % 17)
		universe := uint64(1)<<uint(bitCount) - 1
		masks := make([]uint64, rowCount)
		for row := range masks {
			value := byte(row)
			if 2+row < len(data) {
				value = data[2+row]
			}
			masks[row] = uint64(value) & universe
		}
		mergeByte := data[len(data)-1]
		mergeMask := uint64(mergeByte) & universe
		family := espressoContainFamily(t, bitCount, masks)
		merged, err := family.DistanceMerge(espressoSetFromMask(t, bitCount, mergeMask), EspressoContainOptions{})
		if err != nil {
			t.Fatal(err)
		}
		type item struct{ mask uint64 }
		items := make([]item, len(masks))
		for index, mask := range masks {
			items[index].mask = mask
		}
		slices.SortStableFunc(items, func(first, second item) int {
			firstKey, secondKey := first.mask|mergeMask, second.mask|mergeMask
			if firstKey > secondKey {
				return -1
			}
			if firstKey < secondKey {
				return 1
			}
			return 0
		})
		want := make([]uint64, 0, len(items))
		for index := 0; index < len(items); {
			key, union := items[index].mask|mergeMask, items[index].mask
			index++
			for index < len(items) && items[index].mask|mergeMask == key {
				union |= items[index].mask
				index++
			}
			want = append(want, union)
		}
		if !slices.Equal(espressoFamilyMasks(t, merged), want) || !slices.Equal(espressoFamilyMasks(t, family), masks) {
			t.Fatal(masks, mergeMask, espressoFamilyMasks(t, merged), want)
		}
		_ = merged.Close()
		_ = family.Close()
	})
}
