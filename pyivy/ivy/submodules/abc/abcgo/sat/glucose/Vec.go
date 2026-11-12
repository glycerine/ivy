package glucose

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

// Vec provides the original vector operations over ordinary Go slice storage.
type Vec[T any] struct{ data []T }

func NewVec[T any](size int, pad T) *Vec[T] { v := new(Vec[T]); v.GrowTo(size, pad); return v }
func (v *Vec[T]) Size() int                 { return len(v.data) }
func (v *Vec[T]) Capacity() int             { return cap(v.data) }
func (v *Vec[T]) Data() []T                 { return v.data }
func (v *Vec[T]) At(i int) T                { return v.data[i] }
func (v *Vec[T]) Set(i int, x T)            { v.data[i] = x }
func (v *Vec[T]) Last() T                   { return v.data[len(v.data)-1] }
func (v *Vec[T]) Reserve(n int) {
	if n < 0 {
		panic("negative capacity")
	}
	if cap(v.data) < n {
		next := cap(v.data) + ((cap(v.data)/2 + 2) &^ 1)
		if next < n {
			next = n
		}
		p := make([]T, len(v.data), next)
		copy(p, v.data)
		v.data = p
	}
}
func (v *Vec[T]) GrowTo(n int, pad T) {
	if n < 0 {
		panic("negative vector size")
	}
	if n <= len(v.data) {
		return
	}
	old := len(v.data)
	v.Reserve(n)
	v.data = v.data[:n]
	for i := old; i < n; i++ {
		v.data[i] = pad
	}
}
func (v *Vec[T]) Push(x T) { v.data = append(v.data, x) }
func (v *Vec[T]) PushReserved(x T) {
	if len(v.data) == cap(v.data) {
		panic("vector capacity exceeded")
	}
	v.data = v.data[:len(v.data)+1]
	v.data[len(v.data)-1] = x
}
func (v *Vec[T]) Shrink(n int) {
	if n < 0 || n > len(v.data) {
		panic("invalid vector shrink")
	}
	clear(v.data[len(v.data)-n:])
	v.data = v.data[:len(v.data)-n]
}
func (v *Vec[T]) Pop() { v.Shrink(1) }
func (v *Vec[T]) Clear(deallocate ...bool) {
	clear(v.data)
	v.data = v.data[:0]
	if len(deallocate) > 0 && deallocate[0] {
		v.data = nil
	}
}
func (v *Vec[T]) CopyTo(to *Vec[T]) {
	if v == to {
		return
	}
	to.data = append(to.data[:0], v.data...)
}
func (v *Vec[T]) MoveTo(to *Vec[T]) {
	if v == to {
		return
	}
	to.data = v.data
	v.data = nil
}
