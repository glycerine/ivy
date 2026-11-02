package abc

// Hand-checked skeleton of src/sat/glucose2/BoundedQueue.h (namespace Gluco2).
// (The Gluco2 name is a placeholder: the header exists in sat/glucose and sat/glucose2 only.)

// Gluco2BqueueElem is what bqueue<T> requires of T: arithmetic (it sums its elements).
type Gluco2BqueueElem interface {
	~int | ~int32 | ~uint32 | ~int64 | ~uint64 | ~float32 | ~float64
}

// Gluco2Bqueue is bqueue<T>: a bounded queue keeping a running sum and averages.
type Gluco2Bqueue[T Gluco2BqueueElem] struct {
	elems       Gluco2Vec[T]
	first       int
	last        int
	sumofqueue  uint64
	maxsize     int
	queuesize   int // number of current elements (must be < maxsize!)
	expComputed bool
	exp         float64
	value       float64
}

// NewGluco2Bqueue is bqueue(void).
// C: bqueue(void)
func NewGluco2Bqueue[T Gluco2BqueueElem]() *Gluco2Bqueue[T] {
	panic("TODO, stubbed")
}

// InitSize is bqueue::initSize (init size of the bounded queue).
// C: void initSize(int size)
func (q *Gluco2Bqueue[T]) InitSize(size int) {
	panic("TODO, stubbed")
}

// Push is bqueue::push.
// C: void push(T x)
func (q *Gluco2Bqueue[T]) Push(x T) {
	panic("TODO, stubbed")
}

// Peek is bqueue::peek.
// C: T peek()
func (q *Gluco2Bqueue[T]) Peek() T {
	panic("TODO, stubbed")
}

// Pop is bqueue::pop.
// C: void pop()
func (q *Gluco2Bqueue[T]) Pop() {
	panic("TODO, stubbed")
}

// Getsum is bqueue::getsum.
// C: uint64_t getsum() const
func (q *Gluco2Bqueue[T]) Getsum() uint64 {
	panic("TODO, stubbed")
}

// Getavg is bqueue::getavg.
// C: unsigned int getavg() const
func (q *Gluco2Bqueue[T]) Getavg() uint32 {
	panic("TODO, stubbed")
}

// MaxSize is bqueue::maxSize.
// C: int maxSize() const
func (q *Gluco2Bqueue[T]) MaxSize() int {
	panic("TODO, stubbed")
}

// GetavgDouble is bqueue::getavgDouble.
// C: double getavgDouble() const
func (q *Gluco2Bqueue[T]) GetavgDouble() float64 {
	panic("TODO, stubbed")
}

// Isvalid is bqueue::isvalid.
// C: int isvalid() const
func (q *Gluco2Bqueue[T]) Isvalid() int {
	panic("TODO, stubbed")
}

// GrowTo is bqueue::growTo.
// C: void growTo(int size)
func (q *Gluco2Bqueue[T]) GrowTo(size int) {
	panic("TODO, stubbed")
}

// GetAvgExp is bqueue::getAvgExp.
// C: double getAvgExp()
func (q *Gluco2Bqueue[T]) GetAvgExp() float64 {
	panic("TODO, stubbed")
}

// Fastclear is bqueue::fastclear (to be called after restarts: discards the queue).
// C: void fastclear()
func (q *Gluco2Bqueue[T]) Fastclear() {
	panic("TODO, stubbed")
}

// Size is bqueue::size.
// C: int size(void)
func (q *Gluco2Bqueue[T]) Size() int {
	panic("TODO, stubbed")
}

// Clear is bqueue::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (q *Gluco2Bqueue[T]) Clear(dealloc bool) {
	panic("TODO, stubbed")
}
