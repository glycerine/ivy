package abc

// Hand-checked skeleton of src/sat/glucose2/Queue.h (namespace Gluco2).

// Gluco2Queue is Gluco2::Queue<T>: a ring buffer (typedef T Key is not ported).
type Gluco2Queue[T any] struct {
	buf   Gluco2Vec[T]
	first int
	end   int
}

// NewGluco2Queue is Queue().
// C: Queue()
func NewGluco2Queue[T any]() *Gluco2Queue[T] {
	panic("TODO, stubbed")
}

// Clear is Queue::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (q *Gluco2Queue[T]) Clear(dealloc bool) {
	panic("TODO, stubbed")
}

// Size is Queue::size.
// C: int size() const
func (q *Gluco2Queue[T]) Size() int {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](int index)
func (q *Gluco2Queue[T]) At(index int) *T {
	panic("TODO, stubbed")
}

// Peek is Queue::peek.
// C: T peek() const
func (q *Gluco2Queue[T]) Peek() T {
	panic("TODO, stubbed")
}

// Pop is Queue::pop.
// C: void pop()
func (q *Gluco2Queue[T]) Pop() {
	panic("TODO, stubbed")
}

// Insert is Queue::insert.
// C: void insert(T elem)
func (q *Gluco2Queue[T]) Insert(elem T) {
	panic("TODO, stubbed")
}
