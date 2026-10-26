package abc

// Go-only exact-model, property and fuzz tests for src/aig/gia/giaSort.c.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func giaSortFloatBits(values []float32) []uint32 {
	bits := make([]uint32, len(values))
	for i, value := range values {
		bits[i] = math.Float32bits(value)
	}
	return bits
}

func giaSortFloatModel(values []float32, permutation []int) {
	if len(values) <= 15 {
		for i := 0; i+1 < len(values); i++ {
			best := i
			for j := i + 1; j < len(values); j++ {
				if values[j] < values[best] {
					best = j
				}
			}
			values[i], values[best] = values[best], values[i]
			permutation[i], permutation[best] = permutation[best], permutation[i]
		}
		return
	}
	pivot := values[len(values)/2]
	i, j := -1, len(values)
	for {
		for {
			i++
			if !(values[i] < pivot) {
				break
			}
		}
		for {
			j--
			if !(pivot < values[j]) {
				break
			}
		}
		if i >= j {
			break
		}
		values[i], values[j] = values[j], values[i]
		permutation[i], permutation[j] = permutation[j], permutation[i]
	}
	giaSortFloatModel(values[:i], permutation[:i])
	giaSortFloatModel(values[i:], permutation[i:])
}

func giaSortIntModel(values []int) {
	if len(values) <= 15 {
		for i := 0; i+1 < len(values); i++ {
			best := i
			for j := i + 1; j < len(values); j++ {
				if values[j] < values[best] {
					best = j
				}
			}
			values[i], values[best] = values[best], values[i]
		}
		return
	}
	pivot := values[len(values)/2]
	i, j := -1, len(values)
	for {
		for {
			i++
			if values[i] >= pivot {
				break
			}
		}
		for {
			j--
			if pivot >= values[j] {
				break
			}
		}
		if i >= j {
			break
		}
		values[i], values[j] = values[j], values[i]
	}
	giaSortIntModel(values[:i])
	giaSortIntModel(values[i:])
}

func TestGiaSortSelectionAndPermutation(t *testing.T) {
	values := []float32{3, 1, 2, 1}
	permutation := []int{10, 11, 12, 13}
	got, err := GiaSortFloats(values, permutation, GiaSortOptions{})
	if err != nil || &got[0] != &permutation[0] || !slices.Equal(values, []float32{1, 1, 2, 3}) || !slices.Equal(permutation, []int{11, 13, 12, 10}) {
		t.Fatal(values, permutation, err)
	}
	values = []float32{3, 1, 2, 1}
	got, err = GiaSortFloats(values, nil, GiaSortOptions{})
	if err != nil || !slices.Equal(got, []int{1, 3, 2, 0}) {
		t.Fatal(values, got, err)
	}
	ints := []int{3, -1, 2, -1, 0}
	if err = GiaSortInts(ints, GiaSortOptions{}); err != nil || !slices.Equal(ints, []int{-1, -1, 0, 2, 3}) {
		t.Fatal(ints, err)
	}
}

func TestGiaSortPartitionSpecialFloatsAndExactOrder(t *testing.T) {
	nanA := math.Float32frombits(0x7fc00001)
	nanB := math.Float32frombits(0xffc00002)
	negativeZero := math.Float32frombits(1 << 31)
	values := []float32{3, nanA, -2, 0, negativeZero, 7, 7, float32(math.Inf(1)), float32(math.Inf(-1)), 4, nanB, -2, 9, 1, 5, 3, 8, 6, 2, -1, 10, 3, 3}
	original := slices.Clone(values)
	wantValues := slices.Clone(values)
	wantPermutation := make([]int, len(values))
	for i := range wantPermutation {
		wantPermutation[i] = i
	}
	giaSortFloatModel(wantValues, wantPermutation)
	gotPermutation, err := GiaSortFloats(values, nil, GiaSortOptions{})
	if err != nil || !slices.Equal(giaSortFloatBits(values), giaSortFloatBits(wantValues)) || !slices.Equal(gotPermutation, wantPermutation) {
		t.Fatal(giaSortFloatBits(values), giaSortFloatBits(wantValues), gotPermutation, wantPermutation, err)
	}
	for i, source := range gotPermutation {
		if math.Float32bits(values[i]) != math.Float32bits(original[source]) {
			t.Fatal("permutation lost value pairing", i, source)
		}
	}
	ints := []int{7, -1, 4, 4, 9, 0, -8, 3, 3, 3, 2, 1, 11, -5, 6, 7, 4, 20, -9, 8, 2, 2, 14}
	wantInts := slices.Clone(ints)
	giaSortIntModel(wantInts)
	if err = GiaSortInts(ints, GiaSortOptions{}); err != nil || !slices.Equal(ints, wantInts) || !slices.IsSorted(ints) {
		t.Fatal(ints, wantInts, err)
	}
}

func TestGiaSortLimitsCancellationAtomicityAndLargeInput(t *testing.T) {
	for _, options := range []GiaSortOptions{{MaxEntries: -1}, {MaxEntries: 2}} {
		values := []int{3, 2, 1}
		if err := GiaSortInts(values, options); err == nil || !slices.Equal(values, []int{3, 2, 1}) {
			t.Fatal(options, values, err)
		}
	}
	values := []float32{3, 2, 1}
	if permutation, err := GiaSortFloats(values, []int{0, 1}, GiaSortOptions{}); !errors.Is(err, ErrDimension) || permutation != nil || !slices.Equal(values, []float32{3, 2, 1}) {
		t.Fatal(values, permutation, err)
	}

	baseInts := make([]int, 64)
	baseFloats := make([]float32, 64)
	basePermutation := make([]int, 64)
	for i := range baseInts {
		baseInts[i] = (i*37)%19 - 9
		baseFloats[i] = float32(baseInts[i])
		basePermutation[i] = 1000 + i
	}
	for threshold := 1; threshold <= 12; threshold++ {
		ints := slices.Clone(baseInts)
		calls := 0
		err := GiaSortInts(ints, GiaSortOptions{Stop: func() bool { calls++; return calls == threshold }})
		if !errors.Is(err, ErrCancelled) || !slices.Equal(ints, baseInts) {
			t.Fatal("integer cancellation", threshold, calls, err)
		}
		floats := slices.Clone(baseFloats)
		permutation := slices.Clone(basePermutation)
		calls = 0
		got, err := GiaSortFloats(floats, permutation, GiaSortOptions{Stop: func() bool { calls++; return calls == threshold }})
		if !errors.Is(err, ErrCancelled) || got != nil || !slices.Equal(floats, baseFloats) || !slices.Equal(permutation, basePermutation) {
			t.Fatal("float cancellation", threshold, calls, err)
		}
	}

	large := make([]int, 100000)
	for i := range large {
		large[i] = (i*48271)%65537 - i%31
	}
	if err := GiaSortInts(large, GiaSortOptions{MaxEntries: len(large)}); err != nil || !slices.IsSorted(large) {
		t.Fatal("large sort", err)
	}
	emptyPermutation, err := GiaSortFloats([]float32{}, nil, GiaSortOptions{})
	if err != nil || len(emptyPermutation) != 0 {
		t.Fatal(emptyPermutation, err)
	}
}

func FuzzGiaSort(f *testing.F) {
	for _, seed := range [][]byte{{}, {1}, {3, 1, 2, 1}, {0xff, 0, 0x80, 7, 7, 7}, make([]byte, 96)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			data = data[:256]
		}
		ints := make([]int, len(data))
		for i, value := range data {
			ints[i] = int(int8(value))
		}
		wantInts := slices.Clone(ints)
		giaSortIntModel(wantInts)
		if err := GiaSortInts(ints, GiaSortOptions{}); err != nil || !slices.Equal(ints, wantInts) || !slices.IsSorted(ints) {
			t.Fatal(data, ints, wantInts, err)
		}

		count := (len(data) + 3) / 4
		values := make([]float32, count)
		permutation := make([]int, count)
		for i := range values {
			var bits uint32
			for j := 0; j < 4; j++ {
				index := 4*i + j
				if index < len(data) {
					bits |= uint32(data[index]) << (8 * j)
				}
			}
			values[i] = math.Float32frombits(bits)
			permutation[i] = 17*i + 3
		}
		wantValues := slices.Clone(values)
		wantPermutation := slices.Clone(permutation)
		giaSortFloatModel(wantValues, wantPermutation)
		got, err := GiaSortFloats(values, permutation, GiaSortOptions{})
		if err != nil || len(got) != len(permutation) || !slices.Equal(giaSortFloatBits(values), giaSortFloatBits(wantValues)) || !slices.Equal(got, wantPermutation) {
			t.Fatal(data, giaSortFloatBits(values), giaSortFloatBits(wantValues), got, wantPermutation, err)
		}
	})
}
