package util

import (
	"math/rand"
	"slices"
	"testing"
)

func TestSortIntegersAndCostPermutation(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	for n := 0; n < 160; n++ {
		values := make([]int32, n)
		for i := range values {
			values[i] = int32(rng.Intn(31) - 15)
		}
		want := slices.Clone(values)
		slices.Sort(want)
		got := slices.Clone(values)
		Abc_MergeSort(got, int32(n))
		if !slices.Equal(got, want) {
			t.Fatal(n, got, want)
		}
		perm := Abc_MergeSortCost(values, int32(n))
		seen := map[int32]bool{}
		for i, p := range perm {
			if p < 0 || int(p) >= n || seen[p] {
				t.Fatal("bad permutation", perm)
			}
			seen[p] = true
			if i > 0 && values[perm[i-1]] > values[p] {
				t.Fatal("unsorted cost", perm)
			}
		}
	}
}
func TestPackedSortLowBits(t *testing.T) {
	for _, sorter := range []func([]uint64, int32, int32){Abc_QuickSort1, Abc_QuickSort2, Abc_QuickSort3} {
		for _, decrease := range []int32{0, 1} {
			data := []uint64{100<<32 | 2, 3<<32 | 9, 30<<32 | 1, 8<<32 | 3, 10<<32 | 2}
			sorter(data, int32(len(data)), decrease)
			for i := 1; i < len(data); i++ {
				a, b := uint32(data[i-1]), uint32(data[i])
				if decrease == 0 && a > b || decrease != 0 && a < b {
					t.Fatal(data)
				}
			}
		}
	}
}
func TestRandomResetSequence(t *testing.T) {
	a := Abc_Random(1)
	b := Abc_Random(0)
	if a == b || Abc_Random(1) != a || Abc_Random(0) != b {
		t.Fatal("nonreproducible sequence")
	}
	z, w := NUMBER1, NUMBER2
	z = 36969*(z&65535) + (z >> 16)
	w = 18000*(w&65535) + (w >> 16)
	if a != z<<16+w {
		t.Fatal("C sequence mismatch")
	}
}
