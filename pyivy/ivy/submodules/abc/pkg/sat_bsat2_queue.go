package abc

// Port of src/sat/bsat2/Queue.h (namespace Minisat).

// MinisatQueue is Minisat::Queue<T>: a ring buffer (typedef T Key is not ported).
// INVARIANT: buf[end] is always unused.
type MinisatQueue[T any] struct {
	buf   MinisatVec[T]
	first int
	end   int
}

// NewMinisatQueue is Queue(): buf(1), first(0), end(0).
// C: Queue()
func NewMinisatQueue[T any]() *MinisatQueue[T] {
	q := &MinisatQueue[T]{}
	q.buf.GrowTo(1)
	return q
}

// Clear is Queue::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (q *MinisatQueue[T]) Clear(dealloc bool) {
	q.buf.Clear(dealloc)
	q.buf.GrowTo(1)
	q.first = 0
	q.end = 0
}

// Size is Queue::size.
// C: int size() const
func (q *MinisatQueue[T]) Size() int {
	if q.end >= q.first {
		return q.end - q.first
	}
	return q.end - q.first + q.buf.Size()
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](int index)
func (q *MinisatQueue[T]) At(index int) *T {
	if index < 0 || index >= q.Size() {
		panic("MinisatQueue.At: index out of range")
	}
	return q.buf.At((q.first + index) % q.buf.Size())
}

// Peek is Queue::peek.
// C: T peek() const
func (q *MinisatQueue[T]) Peek() T {
	if q.first == q.end {
		panic("MinisatQueue.Peek: empty queue")
	}
	return *q.buf.At(q.first)
}

// Pop is Queue::pop.
// C: void pop()
func (q *MinisatQueue[T]) Pop() {
	if q.first == q.end {
		panic("MinisatQueue.Pop: empty queue")
	}
	q.first++
	if q.first == q.buf.Size() {
		q.first = 0
	}
}

// Insert is Queue::insert.
// C: void insert(T elem)
func (q *MinisatQueue[T]) Insert(elem T) {
	*q.buf.At(q.end) = elem
	q.end++
	if q.end == q.buf.Size() {
		q.end = 0
	}
	if q.first == q.end { // Resize:
		tmp := NewMinisatVecSize[T]((q.buf.Size()*3 + 1) >> 1)
		i := 0
		var j int
		for j = q.first; j < q.buf.Size(); j++ {
			*tmp.At(i) = *q.buf.At(j)
			i++
		}
		for j = 0; j < q.end; j++ {
			*tmp.At(i) = *q.buf.At(j)
			i++
		}
		q.first = 0
		q.end = q.buf.Size()
		tmp.MoveTo(&q.buf)
	}
}
