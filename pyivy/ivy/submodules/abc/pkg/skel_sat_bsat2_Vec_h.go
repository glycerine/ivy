package abc

// Hand-checked skeleton of src/sat/bsat2/Vec.h (namespace Minisat).
//
// Conventions for the C++ directories (sat/bsat2, sat/glucose, sat/glucose2):
//   - templates become Go generics; a template functor argument (an object with
//     operator()) becomes a type parameter constrained by an interface whose
//     method Call is the operator().
//   - classes become structs; public members are exported (first letter
//     capitalised, the rest of the C++ name verbatim), protected/private members
//     keep the C++ spelling; C++ overloads get a descriptive suffix and the
//     original C++ signature is given in the "C:" comment.
//   - constructors are NewXxx functions, destructors are Destroy methods,
//     operator[] is At (returns a pointer when C++ returns T&).
//   - default arguments are explicit Go parameters (the default is quoted in the
//     "C:" comment); a subclass embeds its base class; virtual dispatch is not
//     modelled by the skeleton.
//   - the namespace is a name prefix: Minisat:: -> Minisat.

// MinisatVec is Minisat::vec<T>: a growable array (data, sz, cap).
// The private copy constructor and operator= of the C++ class only assert(0)
// and are not ported.
type MinisatVec[T any] struct {
	data []T
	sz   int
	cap  int
}

// minisatVecImax is vec<T>::imax (private static inline).
// C: static inline int imax(int x, int y)
func minisatVecImax(x int, y int) int {
	panic("TODO, stubbed")
}

// minisatVecNextCap is vec<T>::nextCap (private static inline).
// C: static inline void nextCap(int& cap)
func minisatVecNextCap(cap *int) {
	panic("TODO, stubbed")
}

// NewMinisatVec is vec(): the empty vector.
// C: vec()
func NewMinisatVec[T any]() *MinisatVec[T] {
	panic("TODO, stubbed")
}

// NewMinisatVecSize is vec(int size).
// C: explicit vec(int size)
func NewMinisatVecSize[T any](size int) *MinisatVec[T] {
	panic("TODO, stubbed")
}

// NewMinisatVecPad is vec(int size, const T& pad).
// C: vec(int size, const T& pad)
func NewMinisatVecPad[T any](size int, pad T) *MinisatVec[T] {
	panic("TODO, stubbed")
}

// Destroy is ~vec().
// C: ~vec()
func (v *MinisatVec[T]) Destroy() {
	panic("TODO, stubbed")
}

// Data is operator T*().
// C: operator T* (void)
func (v *MinisatVec[T]) Data() []T {
	panic("TODO, stubbed")
}

// Size is vec::size.
// C: int size(void) const
func (v *MinisatVec[T]) Size() int {
	panic("TODO, stubbed")
}

// Shrink is vec::shrink.
// C: void shrink(int nelems)
func (v *MinisatVec[T]) Shrink(nelems int) {
	panic("TODO, stubbed")
}

// Shrink_ is vec::shrink_.
// C: void shrink_(int nelems)
func (v *MinisatVec[T]) Shrink_(nelems int) {
	panic("TODO, stubbed")
}

// Capacity is the getter overload capacity().
// C: int capacity(void) const
func (v *MinisatVec[T]) Capacity() int {
	panic("TODO, stubbed")
}

// SetCapacity is the overload capacity(int min_cap).
// C: void capacity(int min_cap)
func (v *MinisatVec[T]) SetCapacity(minCap int) {
	panic("TODO, stubbed")
}

// GrowTo is growTo(int size).
// C: void growTo(int size)
func (v *MinisatVec[T]) GrowTo(size int) {
	panic("TODO, stubbed")
}

// GrowToPad is growTo(int size, const T& pad).
// C: void growTo(int size, const T& pad)
func (v *MinisatVec[T]) GrowToPad(size int, pad T) {
	panic("TODO, stubbed")
}

// Clear is vec::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (v *MinisatVec[T]) Clear(dealloc bool) {
	panic("TODO, stubbed")
}

// Push is the overload push(void): appends a default-constructed element.
// C: void push(void)
func (v *MinisatVec[T]) Push() {
	panic("TODO, stubbed")
}

// PushElem is the overload push(const T& elem).
// C: void push(const T& elem)
func (v *MinisatVec[T]) PushElem(elem T) {
	panic("TODO, stubbed")
}

// Push_ is vec::push_.
// C: void push_(const T& elem)
func (v *MinisatVec[T]) Push_(elem T) {
	panic("TODO, stubbed")
}

// Pop is vec::pop.
// C: void pop(void)
func (v *MinisatVec[T]) Pop() {
	panic("TODO, stubbed")
}

// Last is last() (const and non-const overloads).
// C: T& last(void)
func (v *MinisatVec[T]) Last() *T {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](int index)
func (v *MinisatVec[T]) At(index int) *T {
	panic("TODO, stubbed")
}

// CopyTo is vec::copyTo.
// C: void copyTo(vec<T>& copy) const
func (v *MinisatVec[T]) CopyTo(copy *MinisatVec[T]) {
	panic("TODO, stubbed")
}

// MoveTo is vec::moveTo.
// C: void moveTo(vec<T>& dest)
func (v *MinisatVec[T]) MoveTo(dest *MinisatVec[T]) {
	panic("TODO, stubbed")
}
