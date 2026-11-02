package abc

// Hand-checked skeleton of src/sat/glucose/Vec.h (namespace Gluco).
//
// See skel_sat_bsat2_Vec_h.go for the C++ -> Go conventions.

// GlucoVec is Gluco::vec<T>: a growable array (data, sz, cap).
// The private copy constructor and operator= of the C++ class only assert(0)
// and are not ported.
type GlucoVec[T any] struct {
	data []T
	sz   int
	cap  int
}

// glucoVecImax is vec<T>::imax (private static inline).
// C: static inline int imax(int x, int y)
func glucoVecImax(x int, y int) int {
	panic("TODO, stubbed")
}

// glucoVecNextCap is vec<T>::nextCap (private static inline).
// C: static inline void nextCap(int& cap)
func glucoVecNextCap(cap *int) {
	panic("TODO, stubbed")
}

// NewGlucoVec is vec(): the empty vector.
// C: vec()
func NewGlucoVec[T any]() *GlucoVec[T] {
	panic("TODO, stubbed")
}

// NewGlucoVecSize is vec(int size).
// C: explicit vec(int size)
func NewGlucoVecSize[T any](size int) *GlucoVec[T] {
	panic("TODO, stubbed")
}

// NewGlucoVecPad is vec(int size, const T& pad).
// C: vec(int size, const T& pad)
func NewGlucoVecPad[T any](size int, pad T) *GlucoVec[T] {
	panic("TODO, stubbed")
}

// Destroy is ~vec().
// C: ~vec()
func (v *GlucoVec[T]) Destroy() {
	panic("TODO, stubbed")
}

// Data is operator T*().
// C: operator T* (void)
func (v *GlucoVec[T]) Data() []T {
	panic("TODO, stubbed")
}

// Size is vec::size.
// C: int size(void) const
func (v *GlucoVec[T]) Size() int {
	panic("TODO, stubbed")
}

// Shrink is vec::shrink.
// C: void shrink(int nelems)
func (v *GlucoVec[T]) Shrink(nelems int) {
	panic("TODO, stubbed")
}

// Shrink_ is vec::shrink_.
// C: void shrink_(int nelems)
func (v *GlucoVec[T]) Shrink_(nelems int) {
	panic("TODO, stubbed")
}

// Capacity is the getter overload capacity().
// C: int capacity(void) const
func (v *GlucoVec[T]) Capacity() int {
	panic("TODO, stubbed")
}

// SetCapacity is the overload capacity(int min_cap).
// C: void capacity(int min_cap)
func (v *GlucoVec[T]) SetCapacity(minCap int) {
	panic("TODO, stubbed")
}

// GrowTo is growTo(int size).
// C: void growTo(int size)
func (v *GlucoVec[T]) GrowTo(size int) {
	panic("TODO, stubbed")
}

// GrowToPad is growTo(int size, const T& pad).
// C: void growTo(int size, const T& pad)
func (v *GlucoVec[T]) GrowToPad(size int, pad T) {
	panic("TODO, stubbed")
}

// Clear is vec::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (v *GlucoVec[T]) Clear(dealloc bool) {
	panic("TODO, stubbed")
}

// Push is the overload push(void): appends a default-constructed element.
// C: void push(void)
func (v *GlucoVec[T]) Push() {
	panic("TODO, stubbed")
}

// PushElem is the overload push(const T& elem).
// C: void push(const T& elem)
func (v *GlucoVec[T]) PushElem(elem T) {
	panic("TODO, stubbed")
}

// Push_ is vec::push_.
// C: void push_(const T& elem)
func (v *GlucoVec[T]) Push_(elem T) {
	panic("TODO, stubbed")
}

// Pop is vec::pop.
// C: void pop(void)
func (v *GlucoVec[T]) Pop() {
	panic("TODO, stubbed")
}

// Last is last() (const and non-const overloads).
// C: T& last(void)
func (v *GlucoVec[T]) Last() *T {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](int index)
func (v *GlucoVec[T]) At(index int) *T {
	panic("TODO, stubbed")
}

// CopyTo is vec::copyTo.
// C: void copyTo(vec<T>& copy) const
func (v *GlucoVec[T]) CopyTo(copy *GlucoVec[T]) {
	panic("TODO, stubbed")
}

// CopyTo_ is vec::copyTo_ (drops the destination's elements without destructing them, then copies).
// C: void copyTo_(vec<T>& copy) const
func (v *GlucoVec[T]) CopyTo_(copy *GlucoVec[T]) {
	panic("TODO, stubbed")
}

// MoveTo is vec::moveTo.
// C: void moveTo(vec<T>& dest)
func (v *GlucoVec[T]) MoveTo(dest *GlucoVec[T]) {
	panic("TODO, stubbed")
}
