package abc

import "cmp"

// Hand-checked skeleton of src/sat/glucose/Sort.h (namespace Gluco).

// GlucoLessThan is the requirement on the LessThan template argument of
// sort/selectionSort: a comparator object with bool operator()(T, T).
type GlucoLessThan[T any] interface {
	Call(x T, y T) bool
}

// GlucoLessThanDefault is LessThan_default<T> (compares with operator<).
type GlucoLessThanDefault[T cmp.Ordered] struct{}

// Call is LessThan_default::operator().
// C: bool operator()(T x, T y)
func (l GlucoLessThanDefault[T]) Call(x T, y T) bool {
	panic("TODO, stubbed")
}

// GlucoSelectionSort is selectionSort(T* array, int size, LessThan lt).
// C: template<class T, class LessThan> void selectionSort(T* array, int size, LessThan lt)
func GlucoSelectionSort[T any, LT GlucoLessThan[T]](array []T, size int, lt LT) {
	panic("TODO, stubbed")
}

// glucoSelectionSortDefault is the overload selectionSort(T* array, int size) (static inline).
// C: template<class T> static inline void selectionSort(T* array, int size)
func glucoSelectionSortDefault[T cmp.Ordered](array []T, size int) {
	panic("TODO, stubbed")
}

// GlucoSort is sort(T* array, int size, LessThan lt).
// C: template<class T, class LessThan> void sort(T* array, int size, LessThan lt)
func GlucoSort[T any, LT GlucoLessThan[T]](array []T, size int, lt LT) {
	panic("TODO, stubbed")
}

// glucoSortDefault is the overload sort(T* array, int size) (static inline).
// C: template<class T> static inline void sort(T* array, int size)
func glucoSortDefault[T cmp.Ordered](array []T, size int) {
	panic("TODO, stubbed")
}

// GlucoSortVec is the overload sort(vec<T>& v, LessThan lt).
// C: template<class T, class LessThan> void sort(vec<T>& v, LessThan lt)
func GlucoSortVec[T any, LT GlucoLessThan[T]](v *GlucoVec[T], lt LT) {
	panic("TODO, stubbed")
}

// GlucoSortVecDefault is the overload sort(vec<T>& v).
// C: template<class T> void sort(vec<T>& v)
func GlucoSortVecDefault[T cmp.Ordered](v *GlucoVec[T]) {
	panic("TODO, stubbed")
}
