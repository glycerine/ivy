// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package utils

import "sort"

func select_sort(data []any, size uint32, less func(any, any) bool) {
	data = data[:size]
	for i := 0; i < len(data); i++ {
		best := i
		for j := i + 1; j < len(data); j++ {
			if less(data[j], data[best]) {
				best = j
			}
		}
		data[i], data[best] = data[best], data[i]
	}
}
func satoko_sort(data []any, size uint32, less func(any, any) bool) {
	sort.Slice(data[:size], func(i, j int) bool { return less(data[i], data[j]) })
}
