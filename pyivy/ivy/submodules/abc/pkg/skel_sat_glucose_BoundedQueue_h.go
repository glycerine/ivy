package abc

// Hand-checked skeleton of src/sat/glucose/BoundedQueue.h (namespace Gluco).
// (The Gluco name is a placeholder: the header exists in sat/glucose and sat/glucose2 only.)

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
	queuesize   int // number of current elements (must be < maxsize!)
	expComputed bool
	exp         float64
	value       float64
}

// NewGlucoBqueue is bqueue(void).
// C: bqueue(void)
func NewGlucoBqueue[T GlucoBqueueElem]() *GlucoBqueue[T] {
	panic("TODO, stubbed")
}

// InitSize is bqueue::initSize (init size of the bounded queue).
// C: void initSize(int size)
func (q *GlucoBqueue[T]) InitSize(size int) {
	panic("TODO, stubbed")
}

// Push is bqueue::push.
// C: void push(T x)
func (q *GlucoBqueue[T]) Push(x T) {
	panic("TODO, stubbed")
}

// Peek is bqueue::peek.
// C: T peek()
func (q *GlucoBqueue[T]) Peek() T {
	panic("TODO, stubbed")
}

// Pop is bqueue::pop.
// C: void pop()
func (q *GlucoBqueue[T]) Pop() {
	panic("TODO, stubbed")
}

// Getsum is bqueue::getsum.
// C: uint64_t getsum() const
func (q *GlucoBqueue[T]) Getsum() uint64 {
	panic("TODO, stubbed")
}

// Getavg is bqueue::getavg.
// C: unsigned int getavg() const
func (q *GlucoBqueue[T]) Getavg() uint32 {
	panic("TODO, stubbed")
}

// MaxSize is bqueue::maxSize.
// C: int maxSize() const
func (q *GlucoBqueue[T]) MaxSize() int {
	panic("TODO, stubbed")
}

// GetavgDouble is bqueue::getavgDouble.
// C: double getavgDouble() const
func (q *GlucoBqueue[T]) GetavgDouble() float64 {
	panic("TODO, stubbed")
}

// Isvalid is bqueue::isvalid.
// C: int isvalid() const
func (q *GlucoBqueue[T]) Isvalid() int {
	panic("TODO, stubbed")
}

// GrowTo is bqueue::growTo.
// C: void growTo(int size)
func (q *GlucoBqueue[T]) GrowTo(size int) {
	panic("TODO, stubbed")
}

// GetAvgExp is bqueue::getAvgExp.
// C: double getAvgExp()
func (q *GlucoBqueue[T]) GetAvgExp() float64 {
	panic("TODO, stubbed")
}

// Fastclear is bqueue::fastclear (to be called after restarts: discards the queue).
// C: void fastclear()
func (q *GlucoBqueue[T]) Fastclear() {
	panic("TODO, stubbed")
}

// Size is bqueue::size.
// C: int size(void)
func (q *GlucoBqueue[T]) Size() int {
	panic("TODO, stubbed")
}

// Clear is bqueue::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (q *GlucoBqueue[T]) Clear(dealloc bool) {
	panic("TODO, stubbed")
}
