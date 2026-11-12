package util

// Sorting and multiply-with-carry random sequences from ABC utilSort.c.
// Packed-word sorts compare only their low 32 bits, leaving the upper bits
// available for payloads (notably the original index in cost sorting).
import (
	"sort"
	"sync"
)

func Abc_SortMerge(a, b, out []int32) {
	i, j, k := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i] <= b[j] {
			out[k] = a[i]
			i++
		} else {
			out[k] = b[j]
			j++
		}
		k++
	}
	k += copy(out[k:], a[i:])
	copy(out[k:], b[j:])
}
func Abc_Sort_rec(input, scratch []int32) {
	if len(input) < 2 {
		return
	}
	mid := len(input) / 2
	Abc_Sort_rec(input[:mid], scratch[:mid])
	Abc_Sort_rec(input[mid:], scratch[mid:])
	Abc_SortMerge(input[:mid], input[mid:], scratch[:len(input)])
	copy(input, scratch)
}
func Abc_MergeSort(input []int32, n int32) { Abc_Sort_rec(input[:n], make([]int32, n)) }

// Cost merge records alternate original index and signed cost.
func Abc_MergeSortCostMerge(a, b, out []int32) {
	i, j, k := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i+1] == b[j+1] {
			copy(out[k:k+2], a[i:i+2])
			copy(out[k+2:k+4], b[j:j+2])
			i += 2
			j += 2
			k += 4
		} else if a[i+1] < b[j+1] {
			copy(out[k:k+2], a[i:i+2])
			i += 2
			k += 2
		} else {
			copy(out[k:k+2], b[j:j+2])
			j += 2
			k += 2
		}
	}
	k += copy(out[k:], a[i:])
	copy(out[k:], b[j:])
}
func Abc_MergeSortCost_rec(input, scratch []int32) {
	n := len(input) / 2
	if n < 2 {
		return
	}
	if n < 8 {
		for i := 0; i < n-1; i++ {
			best := i
			for j := i + 1; j < n; j++ {
				if input[2*j+1] < input[2*best+1] {
					best = j
				}
			}
			input[2*i], input[2*best] = input[2*best], input[2*i]
			input[2*i+1], input[2*best+1] = input[2*best+1], input[2*i+1]
		}
		return
	}
	mid := (n / 2) * 2
	Abc_MergeSortCost_rec(input[:mid], scratch[:mid])
	Abc_MergeSortCost_rec(input[mid:], scratch[mid:])
	Abc_MergeSortCostMerge(input[:mid], input[mid:], scratch[:len(input)])
	copy(input, scratch)
}
func Abc_MergeSortCost(costs []int32, n int32) []int32 {
	records := make([]int32, 2*n)
	for i := int32(0); i < n; i++ {
		records[2*i] = i
		records[2*i+1] = costs[i]
	}
	Abc_MergeSortCost_rec(records, make([]int32, len(records)))
	result := make([]int32, n)
	for i := range result {
		result[i] = records[2*i]
	}
	return result
}
func Abc_SortNumCompare(a, b int32) int32 {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
func Abc_QuickSort1CompareInc(a, b uint64) int32 {
	if uint32(a) < uint32(b) {
		return -1
	}
	if uint32(a) > uint32(b) {
		return 1
	}
	return 0
}
func Abc_QuickSort1CompareDec(a, b uint64) int32 { return -Abc_QuickSort1CompareInc(a, b) }
func Abc_QuickSort1(data []uint64, n, decrease int32) {
	data = data[:n]
	sort.Slice(data, func(i, j int) bool {
		if decrease != 0 {
			return uint32(data[i]) > uint32(data[j])
		}
		return uint32(data[i]) < uint32(data[j])
	})
}
func Abc_SelectSortInc(data []uint64, n int32) { selectionSortWords(data[:n], false) }
func Abc_SelectSortDec(data []uint64, n int32) { selectionSortWords(data[:n], true) }
func selectionSortWords(data []uint64, decrease bool) {
	for i := 0; i < len(data)-1; i++ {
		best := i
		for j := i + 1; j < len(data); j++ {
			if (!decrease && uint32(data[j]) < uint32(data[best])) || (decrease && uint32(data[j]) > uint32(data[best])) {
				best = j
			}
		}
		data[i], data[best] = data[best], data[i]
	}
}
func Abc_QuickSort2Inc_rec(data []uint64, l, r int32) { quickSortWords(data, l, r, false) }
func Abc_QuickSort2Dec_rec(data []uint64, l, r int32) { quickSortWords(data, l, r, true) }
func Abc_QuickSort3Inc_rec(data []uint64, l, r int32) { quickSortWords(data, l, r, false) }
func Abc_QuickSort3Dec_rec(data []uint64, l, r int32) { quickSortWords(data, l, r, true) }
func quickSortWords(data []uint64, l, r int32, decrease bool) {
	if l >= r {
		return
	}
	if r-l < 10 {
		selectionSortWords(data[l:r+1], decrease)
		return
	}
	pivot := uint32(data[(l+r)/2])
	i, j := l, r
	less := func(a, b uint32) bool {
		if decrease {
			return a > b
		}
		return a < b
	}
	for i <= j {
		for less(uint32(data[i]), pivot) {
			i++
		}
		for less(pivot, uint32(data[j])) {
			j--
		}
		if i <= j {
			data[i], data[j] = data[j], data[i]
			i++
			j--
		}
	}
	if l < j {
		quickSortWords(data, l, j, decrease)
	}
	if i < r {
		quickSortWords(data, i, r, decrease)
	}
}
func Abc_QuickSort2(data []uint64, n, decrease int32) {
	quickSortWords(data[:n], 0, n-1, decrease != 0)
}
func Abc_QuickSort3(data []uint64, n, decrease int32) {
	quickSortWords(data[:n], 0, n-1, decrease != 0)
}
func Abc_QuickSortCostData(costs []int32, n, decrease int32, data []uint64, result []int32) {
	for i := int32(0); i < n; i++ {
		data[i] = uint64(i)<<32 | uint64(uint32(costs[i]))
	}
	Abc_QuickSort3(data, n, decrease)
	for i := int32(0); i < n; i++ {
		result[i] = int32(data[i] >> 32)
	}
}
func Abc_QuickSortCost(costs []int32, n, decrease int32) []int32 {
	result := make([]int32, n)
	Abc_QuickSortCostData(costs, n, decrease, make([]uint64, n), result)
	return result
}

const NUMBER1 uint32 = 3716960521
const NUMBER2 uint32 = 2174103536

var randomState = struct {
	sync.Mutex
	z, w uint32
}{z: NUMBER1, w: NUMBER2}

func abcRandomLocked(reset bool) uint32 {
	if reset {
		randomState.z = NUMBER1
		randomState.w = NUMBER2
	}
	randomState.z = 36969*(randomState.z&65535) + (randomState.z >> 16)
	randomState.w = 18000*(randomState.w&65535) + (randomState.w >> 16)
	return randomState.z<<16 + randomState.w
}
func Abc_Random(reset int32) uint32 {
	randomState.Lock()
	defer randomState.Unlock()
	return abcRandomLocked(reset != 0)
}
func Abc_RandomW(reset int32) uint64 {
	randomState.Lock()
	defer randomState.Unlock()
	return uint64(abcRandomLocked(reset != 0))<<32 | uint64(abcRandomLocked(reset != 0))
}

// Retain diagnostic entry points as real sortedness checks.
func Abc_SortTest() {
	a := []int32{7, 3, 9, 1}
	Abc_MergeSort(a, int32(len(a)))
	if !sort.SliceIsSorted(a, func(i, j int) bool { return a[i] < a[j] }) {
		panic("merge sort verification")
	}
}
func Abc_QuickSortTest() {
	a := []uint64{7, 3, 9, 1}
	Abc_QuickSort3(a, int32(len(a)), 0)
	if !sort.SliceIsSorted(a, func(i, j int) bool { return uint32(a[i]) < uint32(a[j]) }) {
		panic("quick sort verification")
	}
}
