package bsat2

/*******************************************************************************************[Vec.h]
Copyright (c) 2003-2007, Niklas Een, Niklas Sorensson
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

// Vec owns its storage. Slice exposes the live range; CopyTo creates an
// independent vector, while MoveTo transfers the backing array.
type Vec[T any] struct{ data []T }

func (v *Vec[T]) Len() int       { return len(v.data) }
func (v *Vec[T]) Cap() int       { return cap(v.data) }
func (v *Vec[T]) Slice() []T     { return v.data }
func (v *Vec[T]) At(i int) T     { return v.data[i] }
func (v *Vec[T]) Set(i int, x T) { v.data[i] = x }
func (v *Vec[T]) Reserve(minimum int) {
	if minimum < 0 {
		panic("negative vector capacity")
	}
	if cap(v.data) >= minimum {
		return
	}
	next := cap(v.data) + (cap(v.data) >> 1) + 2
	if next < minimum {
		next = minimum
	}
	next = (next + 1) &^ 1
	data := make([]T, len(v.data), next)
	copy(data, v.data)
	v.data = data
}
func (v *Vec[T]) Push(x T) { v.Reserve(len(v.data) + 1); v.data = append(v.data, x) }
func (v *Vec[T]) Pop() T {
	if len(v.data) == 0 {
		panic("pop from empty vector")
	}
	i := len(v.data) - 1
	x := v.data[i]
	var zero T
	v.data[i] = zero
	v.data = v.data[:i]
	return x
}
func (v *Vec[T]) Last() T { return v.data[len(v.data)-1] }
func (v *Vec[T]) Shrink(count int) {
	if count < 0 || count > len(v.data) {
		panic("invalid vector shrink")
	}
	size := len(v.data) - count
	clear(v.data[size:])
	v.data = v.data[:size]
}
func (v *Vec[T]) GrowTo(size int, pad T) {
	if size < 0 {
		panic("negative vector size")
	}
	v.Reserve(size)
	for len(v.data) < size {
		v.data = append(v.data, pad)
	}
}
func (v *Vec[T]) Clear(deallocate bool) {
	clear(v.data)
	if deallocate {
		v.data = nil
	} else {
		v.data = v.data[:0]
	}
}
func (v *Vec[T]) CopyTo(to *Vec[T]) {
	if v == to {
		return
	}
	to.Clear(false)
	to.Reserve(len(v.data))
	to.data = append(to.data, v.data...)
}
func (v *Vec[T]) MoveTo(to *Vec[T]) {
	if v == to {
		return
	}
	to.data = v.data
	v.data = nil
}
