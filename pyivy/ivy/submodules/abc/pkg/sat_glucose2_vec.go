package abc

import (
	"math"
	"strconv"
)

// Port of src/sat/glucose2/Vec.h (namespace Gluco2).
//
// Conventions for the C++ directories (sat/glucose2, sat/glucose, sat/glucose2):
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
//   - the namespace is a name prefix: Gluco2:: -> Gluco2.
//   - assert() becomes a panic with a message (an assertion failure is a
//     programming error, never recovered); C++ `throw OutOfMemoryException()`
//     becomes panic(Gluco2OutOfMemoryException{}) and, as the C++ callers'
//     catch handlers only print a message and exit, is likewise never recovered.
//   - destructors of elements (data[i].~T()) are replaced by zeroing the slot so
//     that the garbage collector can release what it referenced.

// Gluco2Vec is Gluco2::vec<T>: a growable array (data, sz, cap).
// The private copy constructor and operator= of the C++ class only assert(0)
// and are not ported. data always has length cap.
type Gluco2Vec[T any] struct {
	data []T
	sz   int
	cap  int
}

// gluco2VecImax is vec<T>::imax (private static inline).
// C: static inline int imax(int x, int y)
func gluco2VecImax(x int, y int) int {
	mask := (y - x) >> (strconv.IntSize - 1)
	return (x & mask) + (y & (^mask))
}

// gluco2VecNextCap is vec<T>::nextCap (private static inline).
// C: static inline void nextCap(int& cap)
func gluco2VecNextCap(cap *int) {
	*cap += ((*cap >> 1) + 2) &^ 1
}

// NewGluco2Vec is vec(): the empty vector.
// C: vec()
func NewGluco2Vec[T any]() *Gluco2Vec[T] {
	return &Gluco2Vec[T]{}
}

// NewGluco2VecSize is vec(int size).
// C: explicit vec(int size)
func NewGluco2VecSize[T any](size int) *Gluco2Vec[T] {
	v := &Gluco2Vec[T]{}
	v.GrowTo(size)
	return v
}

// NewGluco2VecPad is vec(int size, const T& pad).
// C: vec(int size, const T& pad)
func NewGluco2VecPad[T any](size int, pad T) *Gluco2Vec[T] {
	v := &Gluco2Vec[T]{}
	v.GrowToPad(size, pad)
	return v
}

// Destroy is ~vec().
// C: ~vec()
func (v *Gluco2Vec[T]) Destroy() {
	v.Clear(true)
}

// Data is operator T*(): the underlying array (length Capacity()).
// C: operator T* (void)
func (v *Gluco2Vec[T]) Data() []T {
	return v.data
}

// Size is vec::size.
// C: int size(void) const
func (v *Gluco2Vec[T]) Size() int {
	return v.sz
}

// Shrink is vec::shrink.
// C: void shrink(int nelems)
func (v *Gluco2Vec[T]) Shrink(nelems int) {
	if nelems > v.sz {
		panic("Gluco2Vec.Shrink: nelems > size")
	}
	var zero T
	for i := 0; i < nelems; i++ {
		v.sz--
		v.data[v.sz] = zero // data[sz].~T()
	}
}

// Shrink_ is vec::shrink_.
// C: void shrink_(int nelems)
func (v *Gluco2Vec[T]) Shrink_(nelems int) {
	if nelems > v.sz {
		panic("Gluco2Vec.Shrink_: nelems > size")
	}
	v.sz -= nelems
}

// Capacity is the getter overload capacity().
// C: int capacity(void) const
func (v *Gluco2Vec[T]) Capacity() int {
	return v.cap
}

// SetCapacity is the overload capacity(int min_cap).
// C: void capacity(int min_cap)
func (v *Gluco2Vec[T]) SetCapacity(minCap int) {
	if v.cap >= minCap {
		return
	}
	add := gluco2VecImax((minCap-v.cap+1)&^1, ((v.cap>>1)+2)&^1) // NOTE: grow by approximately 3/2
	if add > math.MaxInt32-v.cap {
		panic(Gluco2OutOfMemoryException{})
	}
	v.cap += add
	nd := make([]T, v.cap)
	copy(nd, v.data[:v.sz])
	v.data = nd
}

// GrowTo is growTo(int size).
// C: void growTo(int size)
func (v *Gluco2Vec[T]) GrowTo(size int) {
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
func (v *Gluco2Vec[T]) GrowToPad(size int, pad T) {
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
func (v *Gluco2Vec[T]) Clear(dealloc bool) {
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
func (v *Gluco2Vec[T]) Push() {
	if v.sz == v.cap {
		v.SetCapacity(v.sz + 1)
	}
	var zero T
	v.data[v.sz] = zero
	v.sz++
}

// PushElem is the overload push(const T& elem).
// C: void push(const T& elem)
func (v *Gluco2Vec[T]) PushElem(elem T) {
	if v.sz == v.cap {
		v.SetCapacity(v.sz + 1)
	}
	v.data[v.sz] = elem
	v.sz++
}

// Push_ is vec::push_.
// C: void push_(const T& elem)
func (v *Gluco2Vec[T]) Push_(elem T) {
	if v.sz >= v.cap {
		panic("Gluco2Vec.Push_: sz >= cap")
	}
	v.data[v.sz] = elem
	v.sz++
}

// Pop is vec::pop.
// C: void pop(void)
func (v *Gluco2Vec[T]) Pop() {
	if v.sz <= 0 {
		panic("Gluco2Vec.Pop: empty vector")
	}
	v.sz--
	var zero T
	v.data[v.sz] = zero
}

// Last is last() (const and non-const overloads).
// C: T& last(void)
func (v *Gluco2Vec[T]) Last() *T {
	return &v.data[v.sz-1]
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](int index)
func (v *Gluco2Vec[T]) At(index int) *T {
	return &v.data[index]
}

// CopyTo is vec::copyTo.
// C: void copyTo(vec<T>& copy) const
func (v *Gluco2Vec[T]) CopyTo(dst *Gluco2Vec[T]) {
	dst.Clear(false)
	dst.GrowTo(v.sz)
	for i := 0; i < v.sz; i++ {
		dst.data[i] = v.data[i]
	}
}

// CopyTo_ is vec::copyTo_ (drops the destination's elements without destructing them, then copies).
// C: void copyTo_(vec<T>& copy) const
func (v *Gluco2Vec[T]) CopyTo_(dst *Gluco2Vec[T]) {
	dst.Shrink_(dst.Size())
	dst.GrowTo(v.sz)
	for i := 0; i < v.sz; i++ {
		dst.data[i] = v.data[i]
	}
}

// MoveTo is vec::moveTo.
// C: void moveTo(vec<T>& dest)
func (v *Gluco2Vec[T]) MoveTo(dest *Gluco2Vec[T]) {
	dest.Clear(true)
	dest.data = v.data
	dest.sz = v.sz
	dest.cap = v.cap
	v.data = nil
	v.sz = 0
	v.cap = 0
}
