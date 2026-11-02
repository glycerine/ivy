package abc

import "cmp"

// Port of src/sat/glucose/Sort.h (namespace Gluco).

// GlucoLessThan is the requirement on the LessThan template argument of
// sort/selectionSort: a comparator object with bool operator()(T, T).
type GlucoLessThan[T any] interface {
	Call(x T, y T) bool
}

// GlucoLessThanDefault is LessThan_default<T> (compares with operator<).
type GlucoLessThanDefault[T cmp.Ordered] struct{}

// Call is LessThan_default::operator().
// C: bool operator()(T x, T y)
func (l GlucoLessThanDefault[T]) Call(x T, y T) bool { return x < y }

// GlucoSelectionSort is selectionSort(T* array, int size, LessThan lt).
// C: template<class T, class LessThan> void selectionSort(T* array, int size, LessThan lt)
func GlucoSelectionSort[T any, LT GlucoLessThan[T]](array []T, size int, lt LT) {
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

// glucoSelectionSortDefault is the overload selectionSort(T* array, int size) (static inline).
// C: template<class T> static inline void selectionSort(T* array, int size)
func glucoSelectionSortDefault[T cmp.Ordered](array []T, size int) {
	GlucoSelectionSort(array, size, GlucoLessThanDefault[T]{})
}

// GlucoSort is sort(T* array, int size, LessThan lt) (quicksort with a
// selection-sort cut-off at 15 elements).
// C: template<class T, class LessThan> void sort(T* array, int size, LessThan lt)
func GlucoSort[T any, LT GlucoLessThan[T]](array []T, size int, lt LT) {
	if size <= 15 {
		GlucoSelectionSort(array, size, lt)
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
	GlucoSort(array, i, lt)
	GlucoSort(array[i:], size-i, lt)
}

// glucoSortDefault is the overload sort(T* array, int size) (static inline).
// C: template<class T> static inline void sort(T* array, int size)
func glucoSortDefault[T cmp.Ordered](array []T, size int) {
	GlucoSort(array, size, GlucoLessThanDefault[T]{})
}

// GlucoSortVec is the overload sort(vec<T>& v, LessThan lt).
// C: template<class T, class LessThan> void sort(vec<T>& v, LessThan lt)
func GlucoSortVec[T any, LT GlucoLessThan[T]](v *GlucoVec[T], lt LT) {
	GlucoSort(v.Data(), v.Size(), lt)
}

// GlucoSortVecDefault is the overload sort(vec<T>& v).
// C: template<class T> void sort(vec<T>& v)
func GlucoSortVecDefault[T cmp.Ordered](v *GlucoVec[T]) {
	GlucoSortVec(v, GlucoLessThanDefault[T]{})
}
