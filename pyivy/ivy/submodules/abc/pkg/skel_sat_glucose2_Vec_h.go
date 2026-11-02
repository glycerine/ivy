package abc

// Hand-checked skeleton of src/sat/glucose2/Vec.h (namespace Gluco2).
//
// See skel_sat_bsat2_Vec_h.go for the C++ -> Go conventions.

// Gluco2Vec is Gluco2::vec<T>: a growable array (data, sz, cap).
// The private copy constructor and operator= of the C++ class only assert(0)
// and are not ported.
type Gluco2Vec[T any] struct {
	data []T
	sz   int
	cap  int
}

// gluco2VecImax is vec<T>::imax (private static inline).
// C: static inline int imax(int x, int y)
func gluco2VecImax(x int, y int) int {
	panic("TODO, stubbed")
}

// gluco2VecNextCap is vec<T>::nextCap (private static inline).
// C: static inline void nextCap(int& cap)
func gluco2VecNextCap(cap *int) {
	panic("TODO, stubbed")
}

// NewGluco2Vec is vec(): the empty vector.
// C: vec()
func NewGluco2Vec[T any]() *Gluco2Vec[T] {
	panic("TODO, stubbed")
}

// NewGluco2VecSize is vec(int size).
// C: explicit vec(int size)
func NewGluco2VecSize[T any](size int) *Gluco2Vec[T] {
	panic("TODO, stubbed")
}

// NewGluco2VecPad is vec(int size, const T& pad).
// C: vec(int size, const T& pad)
func NewGluco2VecPad[T any](size int, pad T) *Gluco2Vec[T] {
	panic("TODO, stubbed")
}

// Destroy is ~vec().
// C: ~vec()
func (v *Gluco2Vec[T]) Destroy() {
	panic("TODO, stubbed")
}

// Data is operator T*().
// C: operator T* (void)
func (v *Gluco2Vec[T]) Data() []T {
	panic("TODO, stubbed")
}

// Size is vec::size.
// C: int size(void) const
func (v *Gluco2Vec[T]) Size() int {
	panic("TODO, stubbed")
}

// Shrink is vec::shrink.
// C: void shrink(int nelems)
func (v *Gluco2Vec[T]) Shrink(nelems int) {
	panic("TODO, stubbed")
}

// Shrink_ is vec::shrink_.
// C: void shrink_(int nelems)
func (v *Gluco2Vec[T]) Shrink_(nelems int) {
	panic("TODO, stubbed")
}

// Capacity is the getter overload capacity().
// C: int capacity(void) const
func (v *Gluco2Vec[T]) Capacity() int {
	panic("TODO, stubbed")
}

// SetCapacity is the overload capacity(int min_cap).
// C: void capacity(int min_cap)
func (v *Gluco2Vec[T]) SetCapacity(minCap int) {
	panic("TODO, stubbed")
}

// GrowTo is growTo(int size).
// C: void growTo(int size)
func (v *Gluco2Vec[T]) GrowTo(size int) {
	panic("TODO, stubbed")
}

// GrowToPad is growTo(int size, const T& pad).
// C: void growTo(int size, const T& pad)
func (v *Gluco2Vec[T]) GrowToPad(size int, pad T) {
	panic("TODO, stubbed")
}

// Clear is vec::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (v *Gluco2Vec[T]) Clear(dealloc bool) {
	panic("TODO, stubbed")
}

// Push is the overload push(void): appends a default-constructed element.
// C: void push(void)
func (v *Gluco2Vec[T]) Push() {
	panic("TODO, stubbed")
}

// PushElem is the overload push(const T& elem).
// C: void push(const T& elem)
func (v *Gluco2Vec[T]) PushElem(elem T) {
	panic("TODO, stubbed")
}

// Push_ is vec::push_.
// C: void push_(const T& elem)
func (v *Gluco2Vec[T]) Push_(elem T) {
	panic("TODO, stubbed")
}

// Pop is vec::pop.
// C: void pop(void)
func (v *Gluco2Vec[T]) Pop() {
	panic("TODO, stubbed")
}

// Last is last() (const and non-const overloads).
// C: T& last(void)
func (v *Gluco2Vec[T]) Last() *T {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](int index)
func (v *Gluco2Vec[T]) At(index int) *T {
	panic("TODO, stubbed")
}

// CopyTo is vec::copyTo.
// C: void copyTo(vec<T>& copy) const
func (v *Gluco2Vec[T]) CopyTo(copy *Gluco2Vec[T]) {
	panic("TODO, stubbed")
}

// CopyTo_ is vec::copyTo_ (drops the destination's elements without destructing them, then copies).
// C: void copyTo_(vec<T>& copy) const
func (v *Gluco2Vec[T]) CopyTo_(copy *Gluco2Vec[T]) {
	panic("TODO, stubbed")
}

// MoveTo is vec::moveTo.
// C: void moveTo(vec<T>& dest)
func (v *Gluco2Vec[T]) MoveTo(dest *Gluco2Vec[T]) {
	panic("TODO, stubbed")
}
