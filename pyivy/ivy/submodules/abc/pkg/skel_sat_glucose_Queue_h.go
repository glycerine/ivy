package abc

// Hand-checked skeleton of src/sat/glucose/Queue.h (namespace Gluco).

// GlucoQueue is Gluco::Queue<T>: a ring buffer (typedef T Key is not ported).
type GlucoQueue[T any] struct {
	buf   GlucoVec[T]
	first int
	end   int
}

// NewGlucoQueue is Queue().
// C: Queue()
func NewGlucoQueue[T any]() *GlucoQueue[T] {
	panic("TODO, stubbed")
}

// Clear is Queue::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (q *GlucoQueue[T]) Clear(dealloc bool) {
	panic("TODO, stubbed")
}

// Size is Queue::size.
// C: int size() const
func (q *GlucoQueue[T]) Size() int {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](int index)
func (q *GlucoQueue[T]) At(index int) *T {
	panic("TODO, stubbed")
}

// Peek is Queue::peek.
// C: T peek() const
func (q *GlucoQueue[T]) Peek() T {
	panic("TODO, stubbed")
}

// Pop is Queue::pop.
// C: void pop()
func (q *GlucoQueue[T]) Pop() {
	panic("TODO, stubbed")
}

// Insert is Queue::insert.
// C: void insert(T elem)
func (q *GlucoQueue[T]) Insert(elem T) {
	panic("TODO, stubbed")
}
