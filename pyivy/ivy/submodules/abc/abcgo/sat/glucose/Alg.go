package glucose

/*******************************************************************************************[Alg.h]
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

func Find[T comparable](values []T, target T) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
func Remove[T comparable](values []T, target T) []T {
	for i, v := range values {
		if v == target {
			copy(values[i:], values[i+1:])
			clear(values[len(values)-1:])
			return values[:len(values)-1]
		}
	}
	panic("element not found")
}

// CopyValues clones nested vectors through an explicit element copier, avoiding
// reflection and making ownership of nested slices visible at the call site.
func CopyValues[T any](source []T, clone func(T) T) []T {
	out := make([]T, len(source))
	for i, v := range source {
		if clone == nil {
			out[i] = v
		} else {
			out[i] = clone(v)
		}
	}
	return out
}
