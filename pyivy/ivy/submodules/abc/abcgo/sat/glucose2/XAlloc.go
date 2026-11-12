package glucose2

/****************************************************************************************[XAlloc.h]
Copyright (c) 2009-2010, Niklas Sorensson

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

// OutOfMemoryException represents explicit checked allocator-limit failures.
// Go's runtime remains responsible for process-wide allocation exhaustion.
type OutOfMemoryException struct{}

func (OutOfMemoryException) Error() string { return "glucose2: allocation limit exceeded" }
func XRealloc[T any](p []T, n int) []T {
	if n < 0 {
		panic(OutOfMemoryException{})
	}
	if n <= cap(p) {
		old := len(p)
		p = p[:n]
		if n > old {
			clear(p[old:])
		}
		return p
	}
	next := make([]T, n)
	copy(next, p)
	return next
}
