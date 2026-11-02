package abc

import "cmp"

// Port of src/sat/glucose2/Sort.h (namespace Gluco2).

// Gluco2LessThan is the requirement on the LessThan template argument of
// sort/selectionSort: a comparator object with bool operator()(T, T).
type Gluco2LessThan[T any] interface {
	Call(x T, y T) bool
}

// Gluco2LessThanDefault is LessThan_default<T> (compares with operator<).
type Gluco2LessThanDefault[T cmp.Ordered] struct{}

// Call is LessThan_default::operator().
// C: bool operator()(T x, T y)
func (l Gluco2LessThanDefault[T]) Call(x T, y T) bool { return x < y }

// Gluco2SelectionSort is selectionSort(T* array, int size, LessThan lt).
// C: template<class T, class LessThan> void selectionSort(T* array, int size, LessThan lt)
func Gluco2SelectionSort[T any, LT Gluco2LessThan[T]](array []T, size int, lt LT) {
	var i, j, bestI int
	for i = 0; i < size-1; i++ {
		bestI = i
		for j = i + 1; j < size; j++ {
			if lt.Call(array[j], array[bestI]) {
				bestI = j
			}
		}
		array[i], array[bestI] = array[bestI], array[i]
	}
}

// gluco2SelectionSortDefault is the overload selectionSort(T* array, int size) (static inline).
// C: template<class T> static inline void selectionSort(T* array, int size)
func gluco2SelectionSortDefault[T cmp.Ordered](array []T, size int) {
	Gluco2SelectionSort(array, size, Gluco2LessThanDefault[T]{})
}

// Gluco2Sort is sort(T* array, int size, LessThan lt) (quicksort with a
// selection-sort cut-off at 15 elements).
// C: template<class T, class LessThan> void sort(T* array, int size, LessThan lt)
func Gluco2Sort[T any, LT Gluco2LessThan[T]](array []T, size int, lt LT) {
	if size <= 15 {
		Gluco2SelectionSort(array, size, lt)
		return
	}
	pivot := array[size/2]
	i := -1
	j := size
	for {
		for {
			i++
			if !lt.Call(array[i], pivot) {
				break
			}
		}
		for {
			j--
			if !lt.Call(pivot, array[j]) {
				break
			}
		}
		if i >= j {
			break
		}
		array[i], array[j] = array[j], array[i]
	}
	Gluco2Sort(array, i, lt)
	Gluco2Sort(array[i:], size-i, lt)
}

// gluco2SortDefault is the overload sort(T* array, int size) (static inline).
// C: template<class T> static inline void sort(T* array, int size)
func gluco2SortDefault[T cmp.Ordered](array []T, size int) {
	Gluco2Sort(array, size, Gluco2LessThanDefault[T]{})
}

// Gluco2SortVec is the overload sort(vec<T>& v, LessThan lt).
// C: template<class T, class LessThan> void sort(vec<T>& v, LessThan lt)
func Gluco2SortVec[T any, LT Gluco2LessThan[T]](v *Gluco2Vec[T], lt LT) {
	Gluco2Sort(v.Data(), v.Size(), lt)
}

// Gluco2SortVecDefault is the overload sort(vec<T>& v).
// C: template<class T> void sort(vec<T>& v)
func Gluco2SortVecDefault[T cmp.Ordered](v *Gluco2Vec[T]) {
	Gluco2SortVec(v, Gluco2LessThanDefault[T]{})
}
