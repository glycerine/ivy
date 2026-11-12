package bsat2

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

type Queue[T any] struct {
	buffer     []T
	first, end int
}

func (q *Queue[T]) Len() int {
	if q.end >= q.first {
		return q.end - q.first
	}
	return q.end - q.first + len(q.buffer)
}
func (q *Queue[T]) At(index int) T {
	if index < 0 || index >= q.Len() {
		panic("queue index outside live range")
	}
	return q.buffer[(q.first+index)%len(q.buffer)]
}
func (q *Queue[T]) Peek() T { return q.At(0) }
func (q *Queue[T]) Pop() T {
	x := q.Peek()
	var zero T
	q.buffer[q.first] = zero
	q.first++
	if q.first == len(q.buffer) {
		q.first = 0
	}
	return x
}
func (q *Queue[T]) Insert(x T) {
	if len(q.buffer) == 0 {
		q.buffer = make([]T, 1)
	}
	q.buffer[q.end] = x
	q.end++
	if q.end == len(q.buffer) {
		q.end = 0
	}
	if q.first == q.end {
		next := make([]T, (len(q.buffer)*3+1)/2)
		n := copy(next, q.buffer[q.first:])
		copy(next[n:], q.buffer[:q.end])
		q.first = 0
		q.end = len(q.buffer)
		q.buffer = next
	}
}
func (q *Queue[T]) Clear(deallocate bool) {
	clear(q.buffer)
	if deallocate {
		q.buffer = nil
	}
	q.first, q.end = 0, 0
}
