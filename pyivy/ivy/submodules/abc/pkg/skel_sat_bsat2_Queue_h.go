package abc

// Hand-checked skeleton of src/sat/bsat2/Queue.h (namespace Minisat).

// MinisatQueue is Minisat::Queue<T>: a ring buffer (typedef T Key is not ported).
type MinisatQueue[T any] struct {
	buf   MinisatVec[T]
	first int
	end   int
}

// NewMinisatQueue is Queue().
// C: Queue()
func NewMinisatQueue[T any]() *MinisatQueue[T] {
	panic("TODO, stubbed")
}

// Clear is Queue::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (q *MinisatQueue[T]) Clear(dealloc bool) {
	panic("TODO, stubbed")
}

// Size is Queue::size.
// C: int size() const
func (q *MinisatQueue[T]) Size() int {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](int index)
func (q *MinisatQueue[T]) At(index int) *T {
	panic("TODO, stubbed")
}

// Peek is Queue::peek.
// C: T peek() const
func (q *MinisatQueue[T]) Peek() T {
	panic("TODO, stubbed")
}

// Pop is Queue::pop.
// C: void pop()
func (q *MinisatQueue[T]) Pop() {
	panic("TODO, stubbed")
}

// Insert is Queue::insert.
// C: void insert(T elem)
func (q *MinisatQueue[T]) Insert(elem T) {
	panic("TODO, stubbed")
}
