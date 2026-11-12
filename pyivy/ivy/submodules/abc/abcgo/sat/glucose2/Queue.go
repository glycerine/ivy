package glucose2

/*****************************************************************************************[Queue.h]
Copyright (c) 2003-2006, Niklas Een, Niklas Sorensson
Copyright (c) 2007-2010, Niklas Sorensson

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

// Queue is a growable FIFO. The extra slot distinguishes full from empty.
type Queue[T any] struct {
	buf        []T
	first, end int
}

func (q *Queue[T]) Size() int {
	if q.end >= q.first {
		return q.end - q.first
	}
	return q.end - q.first + len(q.buf)
}
func (q *Queue[T]) At(i int) T {
	if i < 0 || i >= q.Size() {
		panic("queue index out of range")
	}
	return q.buf[(q.first+i)%len(q.buf)]
}
func (q *Queue[T]) Peek() T { return q.At(0) }
func (q *Queue[T]) Pop() {
	if q.Size() == 0 {
		panic("empty queue")
	}
	var zero T
	q.buf[q.first] = zero
	q.first = (q.first + 1) % len(q.buf)
}
func (q *Queue[T]) Insert(x T) {
	if len(q.buf) == 0 {
		q.buf = make([]T, 1)
	}
	q.buf[q.end] = x
	q.end = (q.end + 1) % len(q.buf)
	if q.first == q.end {
		next := make([]T, (len(q.buf)*3+1)/2)
		if len(next) <= len(q.buf) {
			next = make([]T, len(q.buf)+1)
		}
		for i := 0; i < len(q.buf); i++ {
			next[i] = q.buf[(q.first+i)%len(q.buf)]
		}
		q.end = len(q.buf)
		q.first = 0
		q.buf = next
	}
}
func (q *Queue[T]) Clear(deallocate ...bool) {
	clear(q.buf)
	q.first = 0
	q.end = 0
	if len(deallocate) > 0 && deallocate[0] {
		q.buf = nil
	}
}
