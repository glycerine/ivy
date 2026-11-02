package abc

// Port of src/sat/glucose/Queue.h (namespace Gluco).

// GlucoQueue is Gluco::Queue<T>: a ring buffer (typedef T Key is not ported).
// INVARIANT: buf[end] is always unused.
type GlucoQueue[T any] struct {
	buf   GlucoVec[T]
	first int
	end   int
}

// NewGlucoQueue is Queue(): buf(1), first(0), end(0).
// C: Queue()
func NewGlucoQueue[T any]() *GlucoQueue[T] {
	q := &GlucoQueue[T]{}
	q.buf.GrowTo(1)
	return q
}

// Clear is Queue::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (q *GlucoQueue[T]) Clear(dealloc bool) {
	q.buf.Clear(dealloc)
	q.buf.GrowTo(1)
	q.first = 0
	q.end = 0
}

// Size is Queue::size.
// C: int size() const
func (q *GlucoQueue[T]) Size() int {
	if q.end >= q.first {
		return q.end - q.first
	}
	return q.end - q.first + q.buf.Size()
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](int index)
func (q *GlucoQueue[T]) At(index int) *T {
	if index < 0 || index >= q.Size() {
		panic("GlucoQueue.At: index out of range")
	}
	return q.buf.At((q.first + index) % q.buf.Size())
}

// Peek is Queue::peek.
// C: T peek() const
func (q *GlucoQueue[T]) Peek() T {
	if q.first == q.end {
		panic("GlucoQueue.Peek: empty queue")
	}
	return *q.buf.At(q.first)
}

// Pop is Queue::pop.
// C: void pop()
func (q *GlucoQueue[T]) Pop() {
	if q.first == q.end {
		panic("GlucoQueue.Pop: empty queue")
	}
	q.first++
	if q.first == q.buf.Size() {
		q.first = 0
	}
}

// Insert is Queue::insert.
// C: void insert(T elem)
func (q *GlucoQueue[T]) Insert(elem T) {
	*q.buf.At(q.end) = elem
	q.end++
	if q.end == q.buf.Size() {
		q.end = 0
	}
	if q.first == q.end { // Resize:
		tmp := NewGlucoVecSize[T]((q.buf.Size()*3 + 1) >> 1)
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
