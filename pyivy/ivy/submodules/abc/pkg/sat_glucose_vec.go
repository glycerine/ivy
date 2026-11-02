package abc

import (
	"math"
	"strconv"
)

// Port of src/sat/glucose/Vec.h (namespace Gluco).
//
// Conventions for the C++ directories (sat/glucose, sat/glucose, sat/glucose2):
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
//     modelled.
//   - the namespace is a name prefix: Gluco:: -> Gluco.
//   - assert() becomes a panic with a message (an assertion failure is a
//     programming error, never recovered); C++ `throw OutOfMemoryException()`
//     becomes panic(GlucoOutOfMemoryException{}) and, as the C++ callers'
//     catch handlers only print a message and exit, is likewise never recovered.
//   - destructors of elements (data[i].~T()) are replaced by zeroing the slot so
//     that the garbage collector can release what it referenced.

// GlucoVec is Gluco::vec<T>: a growable array (data, sz, cap).
// The private copy constructor and operator= of the C++ class only assert(0)
// and are not ported. data always has length cap.
type GlucoVec[T any] struct {
	data []T
	sz   int
	cap  int
}

// glucoVecImax is vec<T>::imax (private static inline).
// C: static inline int imax(int x, int y)
func glucoVecImax(x int, y int) int {
	mask := (y - x) >> (strconv.IntSize - 1)
	return (x & mask) + (y & (^mask))
}

// glucoVecNextCap is vec<T>::nextCap (private static inline).
// C: static inline void nextCap(int& cap)
func glucoVecNextCap(cap *int) {
	*cap += ((*cap >> 1) + 2) &^ 1
}

// NewGlucoVec is vec(): the empty vector.
// C: vec()
func NewGlucoVec[T any]() *GlucoVec[T] {
	return &GlucoVec[T]{}
}

// NewGlucoVecSize is vec(int size).
// C: explicit vec(int size)
func NewGlucoVecSize[T any](size int) *GlucoVec[T] {
	v := &GlucoVec[T]{}
	v.GrowTo(size)
	return v
}

// NewGlucoVecPad is vec(int size, const T& pad).
// C: vec(int size, const T& pad)
func NewGlucoVecPad[T any](size int, pad T) *GlucoVec[T] {
	v := &GlucoVec[T]{}
	v.GrowToPad(size, pad)
	return v
}

// Destroy is ~vec().
// C: ~vec()
func (v *GlucoVec[T]) Destroy() {
	v.Clear(true)
}

// Data is operator T*(): the underlying array (length Capacity()).
// C: operator T* (void)
func (v *GlucoVec[T]) Data() []T {
	return v.data
}

// Size is vec::size.
// C: int size(void) const
func (v *GlucoVec[T]) Size() int {
	return v.sz
}

// Shrink is vec::shrink.
// C: void shrink(int nelems)
func (v *GlucoVec[T]) Shrink(nelems int) {
	if nelems > v.sz {
		panic("GlucoVec.Shrink: nelems > size")
	}
	var zero T
	for i := 0; i < nelems; i++ {
		v.sz--
		v.data[v.sz] = zero // data[sz].~T()
	}
}

// Shrink_ is vec::shrink_.
// C: void shrink_(int nelems)
func (v *GlucoVec[T]) Shrink_(nelems int) {
	if nelems > v.sz {
		panic("GlucoVec.Shrink_: nelems > size")
	}
	v.sz -= nelems
}

// Capacity is the getter overload capacity().
// C: int capacity(void) const
func (v *GlucoVec[T]) Capacity() int {
	return v.cap
}

// SetCapacity is the overload capacity(int min_cap).
// C: void capacity(int min_cap)
func (v *GlucoVec[T]) SetCapacity(minCap int) {
	if v.cap >= minCap {
		return
	}
	add := glucoVecImax((minCap-v.cap+1)&^1, ((v.cap>>1)+2)&^1) // NOTE: grow by approximately 3/2
	if add > math.MaxInt32-v.cap {
		panic(GlucoOutOfMemoryException{})
	}
	v.cap += add
	nd := make([]T, v.cap)
	copy(nd, v.data[:v.sz])
	v.data = nd
}

// GrowTo is growTo(int size).
// C: void growTo(int size)
func (v *GlucoVec[T]) GrowTo(size int) {
	if v.sz >= size {
		return
	}
	v.SetCapacity(size)
	var zero T
	for i := v.sz; i < size; i++ {
		v.data[i] = zero
	}
	v.sz = size
}

// GrowToPad is growTo(int size, const T& pad).
// C: void growTo(int size, const T& pad)
func (v *GlucoVec[T]) GrowToPad(size int, pad T) {
	if v.sz >= size {
		return
	}
	v.SetCapacity(size)
	for i := v.sz; i < size; i++ {
		v.data[i] = pad
	}
	v.sz = size
}

// Clear is vec::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (v *GlucoVec[T]) Clear(dealloc bool) {
	if v.data == nil {
		return
	}
	var zero T
	for i := 0; i < v.sz; i++ {
		v.data[i] = zero
	}
	v.sz = 0
	if dealloc {
		v.data = nil
		v.cap = 0
	}
}

// Push is the overload push(void): appends a default-constructed element.
// C: void push(void)
func (v *GlucoVec[T]) Push() {
	if v.sz == v.cap {
		v.SetCapacity(v.sz + 1)
	}
	var zero T
	v.data[v.sz] = zero
	v.sz++
}

// PushElem is the overload push(const T& elem).
// C: void push(const T& elem)
func (v *GlucoVec[T]) PushElem(elem T) {
	if v.sz == v.cap {
		v.SetCapacity(v.sz + 1)
	}
	v.data[v.sz] = elem
	v.sz++
}

// Push_ is vec::push_.
// C: void push_(const T& elem)
func (v *GlucoVec[T]) Push_(elem T) {
	if v.sz >= v.cap {
		panic("GlucoVec.Push_: sz >= cap")
	}
	v.data[v.sz] = elem
	v.sz++
}

// Pop is vec::pop.
// C: void pop(void)
func (v *GlucoVec[T]) Pop() {
	if v.sz <= 0 {
		panic("GlucoVec.Pop: empty vector")
	}
	v.sz--
	var zero T
	v.data[v.sz] = zero
}

// Last is last() (const and non-const overloads).
// C: T& last(void)
func (v *GlucoVec[T]) Last() *T {
	return &v.data[v.sz-1]
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](int index)
func (v *GlucoVec[T]) At(index int) *T {
	return &v.data[index]
}

// CopyTo is vec::copyTo.
// C: void copyTo(vec<T>& copy) const
func (v *GlucoVec[T]) CopyTo(dst *GlucoVec[T]) {
	dst.Clear(false)
	dst.GrowTo(v.sz)
	for i := 0; i < v.sz; i++ {
		dst.data[i] = v.data[i]
	}
}

// CopyTo_ is vec::copyTo_ (drops the destination's elements without destructing them, then copies).
// C: void copyTo_(vec<T>& copy) const
func (v *GlucoVec[T]) CopyTo_(dst *GlucoVec[T]) {
	dst.Shrink_(dst.Size())
	dst.GrowTo(v.sz)
	for i := 0; i < v.sz; i++ {
		dst.data[i] = v.data[i]
	}
}

// MoveTo is vec::moveTo.
// C: void moveTo(vec<T>& dest)
func (v *GlucoVec[T]) MoveTo(dest *GlucoVec[T]) {
	dest.Clear(true)
	dest.data = v.data
	dest.sz = v.sz
	dest.cap = v.cap
	v.data = nil
	v.sz = 0
	v.cap = 0
}
