package bsat2

/******************************************************************************************[Sort.h]
Copyright (c) 2003-2007, Niklas Een, Niklas Sorensson
Copyright (c) 2007-2010, Niklas Sorensson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

// Sort preserves MiniSAT's middle-pivot quicksort and selection-sort cutoff.
func Sort[T any](items []T, less func(T, T) bool) {
	if len(items) <= 15 {
		for i := 0; i < len(items)-1; i++ {
			best := i
			for j := i + 1; j < len(items); j++ {
				if less(items[j], items[best]) {
					best = j
				}
			}
			items[i], items[best] = items[best], items[i]
		}
		return
	}
	pivot := items[len(items)/2]
	i, j := -1, len(items)
	for {
		i++
		for less(items[i], pivot) {
			i++
		}
		j--
		for less(pivot, items[j]) {
			j--
		}
		if i >= j {
			break
		}
		items[i], items[j] = items[j], items[i]
	}
	Sort(items[:i], less)
	Sort(items[i:], less)
}
