package abc

import (
	"slices"
	"testing"
)

func checkLowWordOrder(t testing.TB, before, after []uint64, decrease bool) {
	t.Helper()
	want := slices.Clone(before)
	slices.Sort(want)
	got := slices.Clone(after)
	slices.Sort(got)
	if !slices.Equal(want, got) {
		t.Fatal("permutation", before, after)
	}
	for i := 1; i < len(after); i++ {
		if !decrease && wordLow(after[i-1]) > wordLow(after[i]) || decrease && wordLow(after[i-1]) < wordLow(after[i]) {
			t.Fatal("order", after)
		}
	}
}

func TestUtilitySortAlgorithms(t *testing.T) {
	values := []int{9, -1, 7, 7, 0, -8, 2, 1, 5, 4, 3, 6, 8}
	SortInts(values)
	if !slices.Equal(values, []int{-8, -1, 0, 1, 2, 3, 4, 5, 6, 7, 7, 8, 9}) {
		t.Fatal(values)
	}
	if got := IntCostPermutation([]int{0, 0, 0, 0, 0, 0, 0, 0}); !slices.Equal(got, []int{0, 4, 1, 5, 2, 6, 3, 7}) {
		t.Fatal(got)
	}
	words := []uint64{4<<32 | 3, 1<<32 | 2, 2<<32 | 2, 9<<32 | 0xffffffff, 7<<32 | 0, 8<<32 | 3, 6<<32 | 1, 5<<32 | 8, 3<<32 | 7, 10<<32 | 6, 11<<32 | 5, 12<<32 | 4}
	for _, decrease := range []bool{false, true} {
		for _, sorter := range []func([]uint64, bool){SortWordsLow32QSort, SortWordsLow32TwoWay, SortWordsLow32ThreeWay, SortWordsLow32NativeThree} {
			copy := slices.Clone(words)
			sorter(copy, decrease)
			checkLowWordOrder(t, words, copy, decrease)
		}
		two, three := slices.Clone(words), slices.Clone(words)
		SortWordsLow32TwoWay(two, decrease)
		SortWordsLow32NativeThree(three, decrease)
		if !slices.Equal(two, three) {
			t.Fatal("native QuickSort3 wrapper")
		}
	}
	for _, decrease := range []bool{false, true} {
		costs := []int32{-1, 0, 2, -3, 2, 1, 0, 5, -1, 4, 3, 2}
		order := QuickCostPermutation(costs, decrease)
		seen := make([]bool, len(costs))
		for i, index := range order {
			if index < 0 || index >= len(costs) || seen[index] {
				t.Fatal(order)
			}
			seen[index] = true
			if i > 0 {
				previous, current := uint32(costs[order[i-1]]), uint32(costs[index])
				if !decrease && previous > current || decrease && previous < current {
					t.Fatal(costs, order)
				}
			}
		}
	}
}

func TestABCRandomOwnedState(t *testing.T) {
	var first, second ABCRandom
	want := []uint32{929569942, 3633709067, 1962020170, 3748422435, 3781595286, 2881179162}
	for i, value := range want {
		if got := first.Uint32(false); got != value {
			t.Fatal(i, got, value)
		}
	}
	if second.Uint32(false) != want[0] || first.Uint32(true) != want[0] {
		t.Fatal("owned/reset")
	}
	if got := second.Uint64(true); got != uint64(want[0])<<32|uint64(want[0]) {
		t.Fatal(got)
	}
	if (*ABCRandom)(nil).Uint32(false) != 0 {
		t.Fatal("nil")
	}
}

func FuzzUtilitySort(f *testing.F) {
	f.Add([]byte{9, 1, 7, 7, 0, 8, 2, 1}, false)
	f.Add([]byte{0, 255, 0, 255, 1, 1, 1, 2}, true)
	f.Fuzz(func(t *testing.T, data []byte, decrease bool) {
		if len(data) > 256 {
			data = data[:256]
		}
		ints := make([]int, len(data))
		words := make([]uint64, len(data))
		costs := make([]int32, len(data))
		for i, value := range data {
			ints[i] = int(int8(value))
			words[i] = uint64(uint32(i))<<32 | uint64(uint32(value)%17)
			costs[i] = int32(int8(value))
		}
		wantInts := slices.Clone(ints)
		slices.Sort(wantInts)
		SortInts(ints)
		if !slices.Equal(ints, wantInts) {
			t.Fatal(ints, wantInts)
		}
		for _, sorter := range []func([]uint64, bool){SortWordsLow32QSort, SortWordsLow32TwoWay, SortWordsLow32ThreeWay, SortWordsLow32NativeThree} {
			copy := slices.Clone(words)
			sorter(copy, decrease)
			checkLowWordOrder(t, words, copy, decrease)
		}
		order := IntCostPermutation(ints)
		for i := 1; i < len(order); i++ {
			if ints[order[i-1]] > ints[order[i]] {
				t.Fatal(order)
			}
		}
		quick := QuickCostPermutation(costs, decrease)
		seen := make([]bool, len(quick))
		for i, index := range quick {
			if index < 0 || index >= len(quick) || seen[index] {
				t.Fatal(quick)
			}
			seen[index] = true
			if i > 0 {
				previous, current := uint32(costs[quick[i-1]]), uint32(costs[index])
				if !decrease && previous > current || decrease && previous < current {
					t.Fatal(quick)
				}
			}
		}
	})
}
