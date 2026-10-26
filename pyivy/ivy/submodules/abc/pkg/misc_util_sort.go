package abc

// Cost permutation from src/misc/util/utilSort.c:Abc_MergeSortCost.

import "sort"

// SortInts sorts in place using the source merge/selection split. Empty and
// singleton slices are already sorted; temporary storage is owned per call.
func SortInts(values []int) {
	if len(values) < 2 {
		return
	}
	scratch := make([]int, len(values))
	var rec func([]int, []int)
	rec = func(input, output []int) {
		switch {
		case len(input) < 2:
			return
		case len(input) < 8:
			for i := 0; i < len(input)-1; i++ {
				best := i
				for j := i + 1; j < len(input); j++ {
					if input[j] < input[best] {
						best = j
					}
				}
				input[i], input[best] = input[best], input[i]
			}
		default:
			middle := len(input) / 2
			rec(input[:middle], output[:middle])
			rec(input[middle:], output[middle:])
			i, j, k := 0, middle, 0
			for i < middle && j < len(input) {
				if input[i] == input[j] {
					output[k], output[k+1] = input[i], input[j]
					i, j, k = i+1, j+1, k+2
				} else if input[i] < input[j] {
					output[k] = input[i]
					i, k = i+1, k+1
				} else {
					output[k] = input[j]
					j, k = j+1, k+1
				}
			}
			k += copy(output[k:], input[i:middle])
			copy(output[k:], input[j:])
			copy(input, output)
		}
	}
	rec(values, scratch)
}

// CostPermutation returns indices in nondecreasing cost order. It preserves
// ABC's selection-sort base cases and alternating equal-cost merge behavior;
// equal costs are therefore not necessarily stable. Costs are not modified.
func CostPermutation(costs []uint64) []int {
	order := make([]int, len(costs))
	for i := range order {
		order[i] = i
	}
	scratch := make([]int, len(order))
	var sort func([]int, []int)
	sort = func(a, out []int) {
		if len(a) < 8 {
			for i := 0; i < len(a)-1; i++ {
				best := i
				for j := i + 1; j < len(a); j++ {
					if costs[a[j]] < costs[a[best]] {
						best = j
					}
				}
				a[i], a[best] = a[best], a[i]
			}
			return
		}
		mid := len(a) / 2
		sort(a[:mid], out[:mid])
		sort(a[mid:], out[mid:])
		i, j, k := 0, mid, 0
		for i < mid && j < len(a) {
			switch {
			case costs[a[i]] == costs[a[j]]:
				out[k] = a[i]
				out[k+1] = a[j]
				i++
				j++
				k += 2
			case costs[a[i]] < costs[a[j]]:
				out[k] = a[i]
				i++
				k++
			default:
				out[k] = a[j]
				j++
				k++
			}
		}
		k += copy(out[k:], a[i:mid])
		copy(out[k:], a[j:])
		copy(a, out)
	}
	sort(order, scratch)
	return order
}

// IntCostPermutation is Abc_MergeSortCost over signed native int costs.
func IntCostPermutation(costs []int) []int {
	order := make([]int, len(costs))
	for i := range order {
		order[i] = i
	}
	scratch := make([]int, len(order))
	var rec func([]int, []int)
	rec = func(input, output []int) {
		if len(input) < 8 {
			for i := 0; i < len(input)-1; i++ {
				best := i
				for j := i + 1; j < len(input); j++ {
					if costs[input[j]] < costs[input[best]] {
						best = j
					}
				}
				input[i], input[best] = input[best], input[i]
			}
			return
		}
		middle := len(input) / 2
		rec(input[:middle], output[:middle])
		rec(input[middle:], output[middle:])
		i, j, k := 0, middle, 0
		for i < middle && j < len(input) {
			if costs[input[i]] == costs[input[j]] {
				output[k], output[k+1] = input[i], input[j]
				i, j, k = i+1, j+1, k+2
			} else if costs[input[i]] < costs[input[j]] {
				output[k] = input[i]
				i, k = i+1, k+1
			} else {
				output[k] = input[j]
				j, k = j+1, k+1
			}
		}
		k += copy(output[k:], input[i:middle])
		copy(output[k:], input[j:])
		copy(input, output)
	}
	rec(order, scratch)
	return order
}

func wordLow(value uint64) uint32 { return uint32(value) }

func selectSortWords(values []uint64, decrease bool) {
	for i := 0; i < len(values)-1; i++ {
		best := i
		for j := i + 1; j < len(values); j++ {
			if !decrease && wordLow(values[j]) < wordLow(values[best]) || decrease && wordLow(values[j]) > wordLow(values[best]) {
				best = j
			}
		}
		values[i], values[best] = values[best], values[i]
	}
}

func quickSortWords2(values []uint64, left, right int, decrease bool) {
	if left >= right {
		return
	}
	if right-left < 10 {
		selectSortWords(values[left:right+1], decrease)
		return
	}
	pivot := wordLow(values[right])
	i, j := left-1, right
	for {
		for {
			i++
			if (!decrease && wordLow(values[i]) >= pivot) || (decrease && wordLow(values[i]) <= pivot) {
				break
			}
		}
		for {
			j--
			if (!decrease && pivot >= wordLow(values[j])) || (decrease && pivot <= wordLow(values[j])) || j == left {
				break
			}
		}
		if i >= j {
			break
		}
		values[i], values[j] = values[j], values[i]
	}
	values[i], values[right] = values[right], values[i]
	quickSortWords2(values, left, i-1, decrease)
	quickSortWords2(values, i+1, right, decrease)
}

func quickSortWords3(values []uint64, left, right int, decrease bool) {
	if left >= right {
		return
	}
	if right-left < 10 {
		selectSortWords(values[left:right+1], decrease)
		return
	}
	pivot := wordLow(values[right])
	i, j, p, q := left-1, right, left-1, right
	for {
		for {
			i++
			if (!decrease && wordLow(values[i]) >= pivot) || (decrease && wordLow(values[i]) <= pivot) {
				break
			}
		}
		for {
			j--
			if (!decrease && pivot >= wordLow(values[j])) || (decrease && pivot <= wordLow(values[j])) || j == left {
				break
			}
		}
		if i >= j {
			break
		}
		values[i], values[j] = values[j], values[i]
		if wordLow(values[i]) == pivot {
			p++
			values[p], values[i] = values[i], values[p]
		}
		if wordLow(values[j]) == pivot {
			q--
			values[j], values[q] = values[q], values[j]
		}
	}
	values[i], values[right] = values[right], values[i]
	j, i = i-1, i+1
	for k := left; k < p; k++ {
		values[k], values[j] = values[j], values[k]
		j--
	}
	for k := right - 1; k > q; k-- {
		values[i], values[k] = values[k], values[i]
		i++
	}
	quickSortWords3(values, left, j, decrease)
	quickSortWords3(values, i, right, decrease)
}

// SortWordsLow32QSort represents Abc_QuickSort1. Equal-key order is
// deliberately unspecified by the source qsort and by this Go sort.
func SortWordsLow32QSort(values []uint64, decrease bool) {
	sort.Slice(values, func(i, j int) bool {
		if decrease {
			return wordLow(values[i]) > wordLow(values[j])
		}
		return wordLow(values[i]) < wordLow(values[j])
	})
}

// SortWordsLow32TwoWay is Abc_QuickSort2 with its selection-sort base cases.
func SortWordsLow32TwoWay(values []uint64, decrease bool) {
	quickSortWords2(values, 0, len(values)-1, decrease)
}

// SortWordsLow32ThreeWay exposes the source's otherwise-unreferenced 3-way
// recursive helpers. Abc_QuickSort3 itself accidentally calls the 2-way helper;
// SortWordsLow32NativeThree preserves that wrapper behavior.
func SortWordsLow32ThreeWay(values []uint64, decrease bool) {
	quickSortWords3(values, 0, len(values)-1, decrease)
}

func SortWordsLow32NativeThree(values []uint64, decrease bool) {
	SortWordsLow32TwoWay(values, decrease)
}

// QuickCostPermutation preserves Abc_QuickSortCost's uint32 comparison of
// signed int cost bit patterns and the native QuickSort3-to-QuickSort2 wrapper.
func QuickCostPermutation(costs []int32, decrease bool) []int {
	data := make([]uint64, len(costs))
	for i, cost := range costs {
		data[i] = uint64(uint32(i))<<32 | uint64(uint32(cost))
	}
	SortWordsLow32NativeThree(data, decrease)
	result := make([]int, len(data))
	for i, value := range data {
		result[i] = int(value >> 32)
	}
	return result
}

const (
	abcRandomZ = uint32(3716960521)
	abcRandomW = uint32(2174103536)
)

// ABCRandom is an owned replacement for utilSort.c's process-global RNG.
// The zero value begins at the native seed on first use.
type ABCRandom struct {
	z, w        uint32
	initialized bool
}

func (r *ABCRandom) Reset() {
	if r != nil {
		r.z, r.w, r.initialized = abcRandomZ, abcRandomW, true
	}
}

// Uint32 resets before advancing when reset is true, matching Abc_Random.
func (r *ABCRandom) Uint32(reset bool) uint32 {
	if r == nil {
		return 0
	}
	if reset || !r.initialized {
		r.Reset()
	}
	r.z = 36969*(r.z&65535) + (r.z >> 16)
	r.w = 18000*(r.w&65535) + (r.w >> 16)
	return (r.z << 16) + r.w
}

// Uint64 calls Uint32 twice with the same reset flag, including the native
// duplicated halves when reset is true.
func (r *ABCRandom) Uint64(reset bool) uint64 {
	return uint64(r.Uint32(reset))<<32 | uint64(r.Uint32(reset))
}
