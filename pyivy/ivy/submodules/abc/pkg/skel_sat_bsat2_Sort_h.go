package abc

import "cmp"

// Hand-checked skeleton of src/sat/bsat2/Sort.h (namespace Minisat).

// MinisatLessThan is the requirement on the LessThan template argument of
// sort/selectionSort: a comparator object with bool operator()(T, T).
type MinisatLessThan[T any] interface {
	Call(x T, y T) bool
}

// MinisatLessThanDefault is LessThan_default<T> (compares with operator<).
type MinisatLessThanDefault[T cmp.Ordered] struct{}

// Call is LessThan_default::operator().
// C: bool operator()(T x, T y)
func (l MinisatLessThanDefault[T]) Call(x T, y T) bool {
	panic("TODO, stubbed")
}

// MinisatSelectionSort is selectionSort(T* array, int size, LessThan lt).
// C: template<class T, class LessThan> void selectionSort(T* array, int size, LessThan lt)
func MinisatSelectionSort[T any, LT MinisatLessThan[T]](array []T, size int, lt LT) {
	panic("TODO, stubbed")
}

// minisatSelectionSortDefault is the overload selectionSort(T* array, int size) (static inline).
// C: template<class T> static inline void selectionSort(T* array, int size)
func minisatSelectionSortDefault[T cmp.Ordered](array []T, size int) {
	panic("TODO, stubbed")
}

// MinisatSort is sort(T* array, int size, LessThan lt).
// C: template<class T, class LessThan> void sort(T* array, int size, LessThan lt)
func MinisatSort[T any, LT MinisatLessThan[T]](array []T, size int, lt LT) {
	panic("TODO, stubbed")
}

// minisatSortDefault is the overload sort(T* array, int size) (static inline).
// C: template<class T> static inline void sort(T* array, int size)
func minisatSortDefault[T cmp.Ordered](array []T, size int) {
	panic("TODO, stubbed")
}

// MinisatSortVec is the overload sort(vec<T>& v, LessThan lt).
// C: template<class T, class LessThan> void sort(vec<T>& v, LessThan lt)
func MinisatSortVec[T any, LT MinisatLessThan[T]](v *MinisatVec[T], lt LT) {
	panic("TODO, stubbed")
}

// MinisatSortVecDefault is the overload sort(vec<T>& v).
// C: template<class T> void sort(vec<T>& v)
func MinisatSortVecDefault[T cmp.Ordered](v *MinisatVec[T]) {
	panic("TODO, stubbed")
}
