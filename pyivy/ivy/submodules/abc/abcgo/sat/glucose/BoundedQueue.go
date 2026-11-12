package glucose

/***********************************************************************************[BoundedQueue.h]
 Glucose -- Copyright (c) 2009, Gilles Audemard, Laurent Simon
                CRIL - Univ. Artois, France
                LRI  - Univ. Paris Sud, France

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

// BoundedQueue retains the most recent samples for Glucose's restart heuristic.
// The sum is an algorithmic cache, updated whenever the logical queue changes.
type BoundedQueue struct {
	elems        []uint64
	first, count int
	sum          uint64
}

func NewBoundedQueue(size int) *BoundedQueue { q := new(BoundedQueue); q.InitSize(size); return q }
func (q *BoundedQueue) InitSize(size int) {
	if size <= 0 {
		panic("nonpositive queue capacity")
	}
	q.elems = make([]uint64, size)
	q.first = 0
	q.count = 0
	q.sum = 0
}
func (q *BoundedQueue) GrowTo(size int) { q.InitSize(size) }
func (q *BoundedQueue) Size() int       { return q.count }
func (q *BoundedQueue) MaxSize() int    { return len(q.elems) }
func (q *BoundedQueue) Valid() bool     { return len(q.elems) > 0 && q.count == len(q.elems) }
func (q *BoundedQueue) Push(x uint64) {
	if len(q.elems) == 0 {
		panic("uninitialized bounded queue")
	}
	if q.count == len(q.elems) {
		q.sum -= q.elems[q.first]
	} else {
		q.count++
	}
	q.elems[q.first] = x
	q.sum += x
	q.first = (q.first + 1) % len(q.elems)
}
func (q *BoundedQueue) oldest() int {
	if q.count == 0 {
		panic("empty bounded queue")
	}
	return (q.first + len(q.elems) - q.count) % len(q.elems)
}
func (q *BoundedQueue) Peek() uint64 { return q.elems[q.oldest()] }
func (q *BoundedQueue) Pop()         { q.sum -= q.elems[q.oldest()]; q.count-- }
func (q *BoundedQueue) Sum() uint64  { return q.sum }

// Avg deliberately truncates, matching the original restart heuristic.
func (q *BoundedQueue) Avg() uint64 {
	if q.count == 0 {
		panic("empty bounded queue")
	}
	return q.sum / uint64(q.count)
}
func (q *BoundedQueue) AvgDouble() float64 {
	if len(q.elems) == 0 {
		return 0
	}
	var s uint64
	for _, v := range q.elems {
		s += v
	}
	return float64(s) / float64(len(q.elems))
}

// AvgExp preserves the original physical-buffer weighting used by diagnostics.
func (q *BoundedQueue) AvgExp() float64 {
	if len(q.elems) == 0 {
		return 0
	}
	exp := 2 / float64(len(q.elems)+1)
	a := exp
	value := float64(q.elems[q.first])
	last := (q.first + len(q.elems) - q.count) % len(q.elems)
	for i := q.first; i < len(q.elems); i++ {
		value += a * float64(q.elems[i])
		a *= exp
	}
	for i := 0; i < last; i++ {
		value += a * float64(q.elems[i])
		a *= exp
	}
	if a == 1 {
		return value
	}
	return value * (1 - exp) / (1 - a)
}
func (q *BoundedQueue) FastClear() { q.first = 0; q.count = 0; q.sum = 0 }
func (q *BoundedQueue) Clear()     { *q = BoundedQueue{} }
