package abc

// Go-only exhaustive, model, and fuzz tests for src/misc/espresso/set.c.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func espressoSetFromMask(t testing.TB, bitCount int, mask uint64) *EspressoSet {
	t.Helper()
	set, err := NewEspressoSet(bitCount)
	if err != nil {
		t.Fatal(err)
	}
	for bit := 0; bit < bitCount && bit < 64; bit++ {
		if mask>>uint(bit)&1 != 0 {
			if err = set.Set(bit, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return set
}

func espressoSetMask(t testing.TB, set *EspressoSet) uint64 {
	t.Helper()
	if set.BitCount() > 64 {
		t.Fatal("test mask too wide")
	}
	elements, err := set.Elements()
	if err != nil {
		t.Fatal(err)
	}
	var result uint64
	for _, element := range elements {
		result |= uint64(1) << uint(element)
	}
	return result
}

func espressoFamilyMasks(t testing.TB, family *EspressoSetFamily) []uint64 {
	t.Helper()
	count, _, err := family.Counts()
	if err != nil {
		t.Fatal(err)
	}
	result := make([]uint64, count)
	for index := range result {
		set, setErr := family.Set(index)
		if setErr != nil {
			t.Fatal(setErr)
		}
		result[index] = espressoSetMask(t, set)
	}
	return result
}

func TestEspressoSetPackedOperationsAndOwnership(t *testing.T) {
	set, err := NewEspressoSet(65)
	if err != nil {
		t.Fatal(err)
	}
	set.Rank = 123
	set.Flags = EspressoSetPrime | EspressoSetActive
	for _, element := range []int{0, 31, 32, 64} {
		if err = set.Set(element, true); err != nil {
			t.Fatal(err)
		}
	}
	cardinality, err := set.Cardinality()
	elements, elementsErr := set.Elements()
	if err != nil || elementsErr != nil || cardinality != 4 || !slices.Equal(elements, []int{0, 31, 32, 64}) || set.String() != "[0,31,32,64]" || len(set.BitString()) != 65 {
		t.Fatal(cardinality, elements, set.String(), len(set.BitString()), err, elementsErr)
	}
	clone, err := set.Clone()
	if err != nil || clone.Rank != 123 || clone.Flags != EspressoSetPrime|EspressoSetActive {
		t.Fatal(clone, err)
	}
	_ = clone.Set(1, true)
	present, _ := set.Has(1)
	if present {
		t.Fatal("clone aliases set")
	}
	full, err := NewFullEspressoSet(65)
	fullCount, _ := full.Cardinality()
	isFull, _ := full.Full()
	if err != nil || fullCount != 65 || !isFull {
		t.Fatal(fullCount, isFull, err)
	}
	if err = full.Clear(); err != nil {
		t.Fatal(err)
	}
	empty, _ := full.Empty()
	if !empty {
		t.Fatal("clear did not empty set")
	}
	zero, err := NewFullEspressoSet(0)
	zeroFull, _ := zero.Full()
	zeroEmpty, _ := zero.Empty()
	if err != nil || !zeroFull || !zeroEmpty {
		t.Fatal(zeroFull, zeroEmpty, err)
	}
	if EspressoBitIndex(0) != -1 || EspressoBitIndex(0x40) != 6 {
		t.Fatal(EspressoBitIndex(0), EspressoBitIndex(0x40))
	}
}

func TestEspressoSetOperationsExhaustive(t *testing.T) {
	for bitCount := 0; bitCount <= 8; bitCount++ {
		limit := uint64(1) << uint(bitCount)
		universe := limit - 1
		for firstMask := uint64(0); firstMask < limit; firstMask++ {
			first := espressoSetFromMask(t, bitCount, firstMask)
			cardinality, err := first.Cardinality()
			if err != nil || cardinality != bits.OnesCount64(firstMask) {
				t.Fatal(bitCount, firstMask, cardinality, err)
			}
			for secondMask := uint64(0); secondMask < limit; secondMask++ {
				second := espressoSetFromMask(t, bitCount, secondMask)
				intersection, err := first.Intersection(second)
				intersectionP, intersectionNonempty, intersectionPErr := first.IntersectionNonempty(second)
				union, unionErr := first.Union(second)
				unionP, unionNonempty, unionPErr := first.UnionNonempty(second)
				difference, differenceErr := first.Difference(second)
				xor, xorErr := first.SymmetricDifference(second)
				common, commonErr := first.CommonCount(second)
				equal, equalErr := first.Equal(second)
				disjoint, disjointErr := first.Disjoint(second)
				implies, impliesErr := first.Implies(second)
				if err != nil || intersectionPErr != nil || unionErr != nil || unionPErr != nil || differenceErr != nil || xorErr != nil || commonErr != nil || equalErr != nil || disjointErr != nil || impliesErr != nil {
					t.Fatal(err, intersectionPErr, unionErr, unionPErr, differenceErr, xorErr, commonErr, equalErr, disjointErr, impliesErr)
				}
				if espressoSetMask(t, intersection) != firstMask&secondMask || espressoSetMask(t, intersectionP) != firstMask&secondMask || intersectionNonempty != (firstMask&secondMask != 0) || espressoSetMask(t, union) != firstMask|secondMask || espressoSetMask(t, unionP) != firstMask|secondMask || unionNonempty != (firstMask|secondMask != 0) || espressoSetMask(t, difference) != firstMask&^secondMask&universe || espressoSetMask(t, xor) != firstMask^secondMask || common != bits.OnesCount64(firstMask&secondMask) || equal != (firstMask == secondMask) || disjoint != (firstMask&secondMask == 0) || implies != (firstMask&^secondMask == 0) {
					t.Fatal(bitCount, firstMask, secondMask)
				}
				mask := espressoSetFromMask(t, bitCount, 0x55&universe)
				merged, mergeErr := first.Merge(second, mask)
				if mergeErr != nil || espressoSetMask(t, merged) != firstMask&(0x55&universe)|secondMask&^(0x55&universe)&universe {
					t.Fatal(bitCount, firstMask, secondMask, espressoSetMask(t, merged), mergeErr)
				}
			}
		}
	}
}

func TestEspressoSetFamilyLifecycleOrderAndCounts(t *testing.T) {
	family, err := NewEspressoSetFamily(3, 5)
	if err != nil {
		t.Fatal(err)
	}
	for index, mask := range []uint64{0b00101, 0b10100, 0b11111} {
		set := espressoSetFromMask(t, 5, mask)
		set.Rank = index + 10
		if index != 1 {
			set.Flags |= EspressoSetActive
		}
		if err = family.Add(set); err != nil {
			t.Fatal(err)
		}
	}
	count, active, err := family.Counts()
	if err != nil || count != 3 || active != 2 {
		t.Fatal(count, active, err)
	}
	union, err := family.UnionAll()
	intersection, intersectionErr := family.IntersectionAll()
	columnCounts, countsErr := family.ColumnCounts()
	if err != nil || intersectionErr != nil || countsErr != nil || espressoSetMask(t, union) != 0b11111 || espressoSetMask(t, intersection) != 0b00100 || !slices.Equal(columnCounts, []int{2, 1, 3, 1, 2}) {
		t.Fatal(espressoSetMask(t, union), espressoSetMask(t, intersection), columnCounts, err, intersectionErr, countsErr)
	}
	restrict := espressoSetFromMask(t, 5, 0b10101)
	restricted, err := family.RestrictedColumnCounts(restrict)
	if err != nil || !slices.Equal(restricted, []int{1280, 0, 2304, 0, 1280}) {
		t.Fatal(restricted, err)
	}
	adjusted := make([]int, 5)
	set, _ := family.Set(0)
	if err = set.AdjustCounts(adjusted, 3); err != nil || !slices.Equal(adjusted, []int{3, 0, 3, 0, 0}) {
		t.Fatal(adjusted, err)
	}
	clone, err := family.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if err = clone.ActivateAll(); err != nil {
		t.Fatal(err)
	}
	if err = clone.SetActive(1, false); err != nil {
		t.Fatal(err)
	}
	if err = clone.CompactActive(); err != nil || !slices.Equal(espressoFamilyMasks(t, clone), []uint64{0b00101, 0b11111}) {
		t.Fatal(espressoFamilyMasks(t, clone), err)
	}
	deleteFamily, _ := NewEspressoSetFamily(3, 3)
	for _, mask := range []uint64{1, 2, 4} {
		_ = deleteFamily.Add(espressoSetFromMask(t, 3, mask))
	}
	if err = deleteFamily.Delete(0); err != nil || !slices.Equal(espressoFamilyMasks(t, deleteFamily), []uint64{4, 2}) {
		t.Fatal(espressoFamilyMasks(t, deleteFamily), err)
	}
	joined, err := JoinEspressoSetFamilies(family, family)
	if err != nil {
		t.Fatal(err)
	}
	joinedCount, joinedActive, _ := joined.Counts()
	if joinedCount != 6 || joinedActive != 4 {
		t.Fatal(joinedCount, joinedActive)
	}
	appendTarget, _ := family.Clone()
	appendSource, _ := family.Clone()
	if err = appendTarget.AppendAndClose(appendSource); err != nil {
		t.Fatal(err)
	}
	if _, _, err = appendSource.Counts(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	for _, owned := range []*EspressoSetFamily{family, clone, deleteFamily, joined, appendTarget} {
		if err = owned.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEspressoSetFamilyTransformsAndFormats(t *testing.T) {
	family, _ := NewEspressoSetFamily(2, 5)
	first := espressoSetFromMask(t, 5, 0b10101)
	first.Rank, first.Flags = 7, EspressoSetPrime|EspressoSetActive
	second := espressoSetFromMask(t, 5, 0b01110)
	second.Rank = 8
	_ = family.Add(first)
	_ = family.Add(second)
	deleted, err := family.DeleteColumns(1, 2)
	added, addedErr := family.AddColumns(2, 2)
	permuted, permutedErr := family.Permute([]int{4, 0, 2})
	selector := espressoSetFromMask(t, 5, 0b10101)
	compressed, compressedErr := family.Compress(selector)
	transposed, transposeErr := family.Transpose()
	if err != nil || addedErr != nil || permutedErr != nil || compressedErr != nil || transposeErr != nil {
		t.Fatal(err, addedErr, permutedErr, compressedErr, transposeErr)
	}
	if !slices.Equal(espressoFamilyMasks(t, deleted), []uint64{0b101, 0b010}) || !slices.Equal(espressoFamilyMasks(t, added), []uint64{0b1010001, 0b0110010}) || !slices.Equal(espressoFamilyMasks(t, permuted), []uint64{0b111, 0b100}) || !slices.Equal(espressoFamilyMasks(t, compressed), []uint64{0b111, 0b010}) || !slices.Equal(espressoFamilyMasks(t, transposed), []uint64{1, 2, 3, 2, 1}) {
		t.Fatal(espressoFamilyMasks(t, deleted), espressoFamilyMasks(t, added), espressoFamilyMasks(t, permuted), espressoFamilyMasks(t, compressed), espressoFamilyMasks(t, transposed))
	}
	for _, rebuilt := range []*EspressoSetFamily{deleted, added, permuted, compressed, transposed} {
		rebuiltFirst, _ := rebuilt.Set(0)
		if rebuiltFirst.Rank != 0 || rebuiltFirst.Flags != 0 {
			t.Fatal("rebuilt family retained packed metadata", rebuiltFirst)
		}
	}
	quickAdded, err := family.AddColumns(5, 2)
	if err != nil {
		t.Fatal(err)
	}
	quickFirst, _ := quickAdded.Set(0)
	if quickFirst.Rank != 7 || quickFirst.Flags != EspressoSetPrime|EspressoSetActive || !slices.Equal(espressoFamilyMasks(t, quickAdded), []uint64{0b10101, 0b01110}) {
		t.Fatal(quickFirst, espressoFamilyMasks(t, quickAdded))
	}
	destination, _ := NewEspressoSetFamily(2, 2)
	_ = destination.Add(espressoSetFromMask(t, 2, 0))
	_ = destination.Add(espressoSetFromMask(t, 2, 0))
	if err = destination.CopyColumnFrom(1, family, 2); err != nil || !slices.Equal(espressoFamilyMasks(t, destination), []uint64{2, 2}) {
		t.Fatal(espressoFamilyMasks(t, destination), err)
	}
	native, err := family.MarshalText()
	parsed, parseErr := ParseEspressoSetFamily(native, EspressoSetFamilyLimits{})
	if err != nil || parseErr != nil || !slices.Equal(espressoFamilyMasks(t, parsed), espressoFamilyMasks(t, family)) {
		t.Fatal(string(native), err, parseErr)
	}
	parsedFirst, _ := parsed.Set(0)
	if parsedFirst.Rank != 7 || parsedFirst.Flags != EspressoSetPrime|EspressoSetActive {
		t.Fatal(parsedFirst)
	}
	bitMatrix, err := family.MarshalBitMatrix()
	parsedBits, bitErr := ParseEspressoBitMatrix(bitMatrix, EspressoSetFamilyLimits{})
	if err != nil || bitErr != nil || !slices.Equal(espressoFamilyMasks(t, parsedBits), espressoFamilyMasks(t, family)) {
		t.Fatal(string(bitMatrix), err, bitErr)
	}
	formatted, err := family.FormatSets()
	bitFormatted, bitFormatErr := family.FormatBitMatrix()
	if err != nil || bitFormatErr != nil || formatted != "A[0] = [0,2,4]\nA[1] = [1,2,3]\n" || bitFormatted != "[   0] 10101\n[   1] 01110\n" {
		t.Fatal(formatted, bitFormatted, err, bitFormatErr)
	}
	for _, owned := range []*EspressoSetFamily{family, deleted, added, quickAdded, permuted, compressed, transposed, destination, parsed, parsedBits} {
		if err = owned.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEspressoSetAndFamilyErrorsLimits(t *testing.T) {
	if _, err := NewEspressoSet(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewEspressoSet(espressoSetMaximumBits + 1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	set := espressoSetFromMask(t, 3, 1)
	other := espressoSetFromMask(t, 4, 1)
	if _, err := set.Union(other); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := set.Set(3, true); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	family, _ := NewEspressoSetFamily(1, 3)
	_ = family.Add(set)
	if err := family.Add(other); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := family.DeleteColumns(2, 2); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := family.Permute([]int{3}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	singletonFamily, _ := NewEspressoSetFamily(1, 3)
	_ = singletonFamily.Add(set)
	if _, err := singletonFamily.RestrictedColumnCounts(set); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	badNative := [][]byte{[]byte(""), []byte("1 x"), []byte("1 3"), []byte("1 3 2 1"), []byte("1 3 1 8")}
	for _, data := range badNative {
		if _, err := ParseEspressoSetFamily(data, EspressoSetFamilyLimits{}); !errors.Is(err, ErrFormat) {
			t.Fatal(string(data), err)
		}
	}
	if _, err := ParseEspressoSetFamily([]byte("1 3 1 1"), EspressoSetFamilyLimits{MaxSets: 0, MaxBits: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	badBits := [][]byte{[]byte(""), []byte("1 x"), []byte("1 3 01"), []byte("1 3 0x1")}
	for _, data := range badBits {
		if _, err := ParseEspressoBitMatrix(data, EspressoSetFamilyLimits{}); !errors.Is(err, ErrFormat) {
			t.Fatal(string(data), err)
		}
	}
	_ = family.Close()
	_ = singletonFamily.Close()
}

func FuzzEspressoSetOperations(f *testing.F) {
	f.Add(byte(65), uint64(0x8000000100000001), uint64(0x100000003))
	f.Add(byte(8), uint64(0x55), uint64(0xaa))
	f.Fuzz(func(t *testing.T, widthByte byte, firstMask, secondMask uint64) {
		bitCount := int(widthByte % 65)
		mask := ^uint64(0)
		if bitCount < 64 {
			mask = uint64(1)<<uint(bitCount) - 1
		}
		firstMask &= mask
		secondMask &= mask
		first := espressoSetFromMask(t, bitCount, firstMask)
		second := espressoSetFromMask(t, bitCount, secondMask)
		intersection, err := first.Intersection(second)
		union, unionErr := first.Union(second)
		difference, differenceErr := first.Difference(second)
		xor, xorErr := first.SymmetricDifference(second)
		if err != nil || unionErr != nil || differenceErr != nil || xorErr != nil || espressoSetMask(t, intersection) != firstMask&secondMask || espressoSetMask(t, union) != firstMask|secondMask || espressoSetMask(t, difference) != firstMask&^secondMask&mask || espressoSetMask(t, xor) != firstMask^secondMask {
			t.Fatal(bitCount, firstMask, secondMask, err, unionErr, differenceErr, xorErr)
		}
	})
}

func FuzzEspressoSetFamilyRoundTrip(f *testing.F) {
	f.Add([]byte{5, 3, 0x15, 0x0e, 0x1f, 4, 0, 2})
	f.Add([]byte{0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		bitCount := int(data[0] % 33)
		setCount := int(data[1] % 17)
		family, err := NewEspressoSetFamily(setCount, bitCount)
		if err != nil {
			t.Fatal(err)
		}
		mask := uint64(1)<<uint(bitCount) - 1
		for row := 0; row < setCount; row++ {
			value := uint64(row + 1)
			if 2+row < len(data) {
				value = uint64(data[2+row])
			}
			set := espressoSetFromMask(t, bitCount, value&mask)
			set.Rank = row
			if row&1 == 0 {
				set.Flags = EspressoSetActive
			}
			_ = family.Add(set)
		}
		native, err := family.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseEspressoSetFamily(native, EspressoSetFamilyLimits{})
		if err != nil || !slices.Equal(espressoFamilyMasks(t, parsed), espressoFamilyMasks(t, family)) {
			t.Fatal(err, string(native))
		}
		permutation := make([]int, 0)
		for index := 2 + setCount; index < len(data) && len(permutation) < 32 && bitCount != 0; index++ {
			permutation = append(permutation, int(data[index])%bitCount)
		}
		permuted, err := family.Permute(permutation)
		if err != nil {
			t.Fatal(err)
		}
		for row := 0; row < setCount; row++ {
			source, _ := family.Set(row)
			destination, _ := permuted.Set(row)
			for column, sourceColumn := range permutation {
				want, _ := source.Has(sourceColumn)
				got, _ := destination.Has(column)
				if got != want {
					t.Fatal(row, column, sourceColumn, got, want)
				}
			}
		}
		_ = permuted.Close()
		_ = parsed.Close()
		_ = family.Close()
	})
}
