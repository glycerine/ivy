package abc

// Port of src/sat/glucose/BoundedQueue.h (namespace Gluco).

// GlucoBqueueElem is what bqueue<T> requires of T: arithmetic (it sums its elements).
type GlucoBqueueElem interface {
	~int | ~int32 | ~uint32 | ~int64 | ~uint64 | ~float32 | ~float64
}

// GlucoBqueue is bqueue<T>: a bounded queue keeping a running sum and averages.
type GlucoBqueue[T GlucoBqueueElem] struct {
	elems       GlucoVec[T]
	first       int
	last        int
	sumofqueue  uint64
	maxsize     int
	queuesize   int // number of current elements (must be < maxsize !)
	expComputed bool
	exp         float64
	value       float64
}

// NewGlucoBqueue is bqueue(void).
// C: bqueue(void)
func NewGlucoBqueue[T GlucoBqueueElem]() *GlucoBqueue[T] {
	return &GlucoBqueue[T]{}
}

// InitSize is bqueue::initSize (init size of the bounded queue).
// C: void initSize(int size)
func (q *GlucoBqueue[T]) InitSize(size int) {
	q.GrowTo(size)
	q.exp = 2.0 / float64(size+1)
}

// Push is bqueue::push.
// C: void push(T x)
func (q *GlucoBqueue[T]) Push(x T) {
	q.expComputed = false
	if q.queuesize == q.maxsize {
		if q.last != q.first { // The queue is full, next value to enter will replace oldest one
			panic("GlucoBqueue.Push: full queue with last != first")
		}
		q.sumofqueue -= uint64(*q.elems.At(q.last))
		q.last++
		if q.last == q.maxsize {
			q.last = 0
		}
	} else {
		q.queuesize++
	}
	q.sumofqueue += uint64(x)
	*q.elems.At(q.first) = x
	q.first++
	if q.first == q.maxsize {
		q.first = 0
		q.last = 0
	}
}

// Peek is bqueue::peek.
// C: T peek()
func (q *GlucoBqueue[T]) Peek() T {
	if q.queuesize <= 0 {
		panic("GlucoBqueue.Peek: empty queue")
	}
	return *q.elems.At(q.last)
}

// Pop is bqueue::pop.
// C: void pop()
func (q *GlucoBqueue[T]) Pop() {
	q.sumofqueue -= uint64(*q.elems.At(q.last))
	q.queuesize--
	q.last++
	if q.last == q.maxsize {
		q.last = 0
	}
}

// Getsum is bqueue::getsum.
// C: uint64_t getsum() const
func (q *GlucoBqueue[T]) Getsum() uint64 { return q.sumofqueue }

// Getavg is bqueue::getavg.
// C: unsigned int getavg() const
func (q *GlucoBqueue[T]) Getavg() uint32 {
	return uint32(q.sumofqueue / uint64(q.queuesize))
}

// MaxSize is bqueue::maxSize.
// C: int maxSize() const
func (q *GlucoBqueue[T]) MaxSize() int { return q.maxsize }

// GetavgDouble is bqueue::getavgDouble.
// C: double getavgDouble() const
func (q *GlucoBqueue[T]) GetavgDouble() float64 {
	tmp := 0.0
	for i := 0; i < q.elems.Size(); i++ {
		tmp += float64(*q.elems.At(i))
	}
	return tmp / float64(q.elems.Size())
}

// Isvalid is bqueue::isvalid.
// C: int isvalid() const
func (q *GlucoBqueue[T]) Isvalid() int {
	if q.queuesize == q.maxsize {
		return 1
	}
	return 0
}

// GrowTo is bqueue::growTo.
// C: void growTo(int size)
func (q *GlucoBqueue[T]) GrowTo(size int) {
	q.elems.GrowTo(size)
	q.first = 0
	q.maxsize = size
	q.queuesize = 0
	q.last = 0
	for i := 0; i < size; i++ {
		var zero T
		*q.elems.At(i) = zero
	}
}

// GetAvgExp is bqueue::getAvgExp.
// C: double getAvgExp()
func (q *GlucoBqueue[T]) GetAvgExp() float64 {
	if q.expComputed {
		return q.value
	}
	a := q.exp
	q.value = float64(*q.elems.At(q.first))
	for i := q.first; i < q.maxsize; i++ {
		q.value += a * float64(*q.elems.At(i))
		a = a * q.exp
	}
	for i := 0; i < q.last; i++ {
		q.value += a * float64(*q.elems.At(i))
		a = a * q.exp
	}
	q.value = q.value * (1 - q.exp) / (1 - a)
	q.expComputed = true
	return q.value
}

// Fastclear is bqueue::fastclear (to be called after restarts: discards the queue).
// C: void fastclear()
func (q *GlucoBqueue[T]) Fastclear() {
	q.first = 0
	q.last = 0
	q.queuesize = 0
	q.sumofqueue = 0
}

// Size is bqueue::size.
// C: int size(void)
func (q *GlucoBqueue[T]) Size() int { return q.queuesize }

// Clear is bqueue::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (q *GlucoBqueue[T]) Clear(dealloc bool) {
	q.elems.Clear(dealloc)
	q.first = 0
	q.maxsize = 0
	q.queuesize = 0
	q.sumofqueue = 0
}
