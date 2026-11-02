package abc

import "cmp"

// Port of src/sat/bsat2/Sort.h (namespace Minisat).

// MinisatLessThan is the requirement on the LessThan template argument of
// sort/selectionSort: a comparator object with bool operator()(T, T).
type MinisatLessThan[T any] interface {
	Call(x T, y T) bool
}

// MinisatLessThanDefault is LessThan_default<T> (compares with operator<).
type MinisatLessThanDefault[T cmp.Ordered] struct{}

// Call is LessThan_default::operator().
// C: bool operator()(T x, T y)
func (l MinisatLessThanDefault[T]) Call(x T, y T) bool { return x < y }

// MinisatSelectionSort is selectionSort(T* array, int size, LessThan lt).
// C: template<class T, class LessThan> void selectionSort(T* array, int size, LessThan lt)
func MinisatSelectionSort[T any, LT MinisatLessThan[T]](array []T, size int, lt LT) {
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

// minisatSelectionSortDefault is the overload selectionSort(T* array, int size) (static inline).
// C: template<class T> static inline void selectionSort(T* array, int size)
func minisatSelectionSortDefault[T cmp.Ordered](array []T, size int) {
	MinisatSelectionSort(array, size, MinisatLessThanDefault[T]{})
}

// MinisatSort is sort(T* array, int size, LessThan lt) (quicksort with a
// selection-sort cut-off at 15 elements).
// C: template<class T, class LessThan> void sort(T* array, int size, LessThan lt)
func MinisatSort[T any, LT MinisatLessThan[T]](array []T, size int, lt LT) {
	if size <= 15 {
		MinisatSelectionSort(array, size, lt)
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
	MinisatSort(array, i, lt)
	MinisatSort(array[i:], size-i, lt)
}

// minisatSortDefault is the overload sort(T* array, int size) (static inline).
// C: template<class T> static inline void sort(T* array, int size)
func minisatSortDefault[T cmp.Ordered](array []T, size int) {
	MinisatSort(array, size, MinisatLessThanDefault[T]{})
}

// MinisatSortVec is the overload sort(vec<T>& v, LessThan lt).
// C: template<class T, class LessThan> void sort(vec<T>& v, LessThan lt)
func MinisatSortVec[T any, LT MinisatLessThan[T]](v *MinisatVec[T], lt LT) {
	MinisatSort(v.Data(), v.Size(), lt)
}

// MinisatSortVecDefault is the overload sort(vec<T>& v).
// C: template<class T> void sort(vec<T>& v)
func MinisatSortVecDefault[T cmp.Ordered](v *MinisatVec[T]) {
	MinisatSortVec(v, MinisatLessThanDefault[T]{})
}
