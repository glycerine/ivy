package bsat

/**************************************************************************************************
MiniSat -- Copyright (c) 2005, Niklas Sorensson
http://www.cs.chalmers.se/Cs/Research/FormalMethods/MiniSat/

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
// SAT vectors own Go slices. begin returns a live view; growing may relocate it.
type veci_t struct {
	cap, size int32
	ptr       []int32
}
type veci = veci_t

func veci_new(v *veci)           { v.cap = 4; v.size = 0; v.ptr = make([]int32, 4) }
func veci_delete(v *veci)        { *v = veci{} }
func veci_begin(v *veci) []int32 { return v.ptr[:v.size] }
func veci_size(v *veci) int32    { return v.size }
func veci_resize(v *veci, k int32) {
	if k < 0 || k > v.size {
		panic("veci_resize: only shrinking is allowed")
	}
	v.size = k
}
func veci_pop(v *veci) int32 {
	if v.size == 0 {
		panic("veci_pop: empty vector")
	}
	v.size--
	return v.ptr[v.size]
}
func veci_push(v *veci, e int32) {
	if v.size == v.cap {
		n := satVectorCapacity(v.cap)
		p := make([]int32, n)
		copy(p, v.ptr)
		v.ptr = p
		v.cap = n
	}
	v.ptr[v.size] = e
	v.size++
}
func veci_remove(v *veci, e int32) {
	for j := int32(0); j < v.size; j++ {
		if v.ptr[j] == e {
			copy(v.ptr[j:], v.ptr[j+1:v.size])
			v.size--
			return
		}
	}
	panic("veci_remove: missing element")
}
func satVectorCapacity(old int32) int32 {
	if old == 0 {
		return 4
	}
	n := old * 2
	if old >= 4 {
		n = (old / 2) * 3
	}
	if n <= old {
		panic("SAT vector capacity overflow")
	}
	return n
}

type vecp_t struct {
	cap, size int32
	ptr       []any
}
type vecp = vecp_t

func vecp_new(v *vecp)         { v.cap = 4; v.size = 0; v.ptr = make([]any, 4) }
func vecp_delete(v *vecp)      { *v = vecp{} }
func vecp_begin(v *vecp) []any { return v.ptr[:v.size] }
func vecp_size(v *vecp) int32  { return v.size }
func vecp_resize(v *vecp, k int32) {
	if k < 0 || k > v.size {
		panic("vecp_resize: only shrinking is allowed")
	}
	clear(v.ptr[k:v.size])
	v.size = k
}
func vecp_push(v *vecp, e any) {
	if v.size == v.cap {
		n := satVectorCapacity(v.cap)
		p := make([]any, n)
		copy(p, v.ptr)
		v.ptr = p
		v.cap = n
	}
	v.ptr[v.size] = e
	v.size++
}

// Pointer-vector values must be comparable (the original elements were pointers).
func vecp_remove(v *vecp, e any) {
	for j := int32(0); j < v.size; j++ {
		if v.ptr[j] == e {
			copy(v.ptr[j:], v.ptr[j+1:v.size])
			v.size--
			v.ptr[v.size] = nil
			return
		}
	}
	panic("vecp_remove: missing element")
}

type lit = int32
type cla = int32
type lbool = int8

const (
	var_Undef int32 = -1
	lit_Undef lit   = -2
	l_Undef         = 0
	l_True          = 1
	l_False         = -1
)

func toLit(v int32) lit {
	if v < 0 || v > 0x3fffffff {
		panic("SAT variable exceeds literal range")
	}
	return v + v
}
func toLitCond(v, c int32) lit {
	l := toLit(v)
	if c != 0 {
		l++
	}
	return l
}
func lit_neg(l lit) lit    { return l ^ 1 }
func lit_var(l lit) int32  { return l >> 1 }
func lit_sign(l lit) int32 { return l & 1 }
func lit_print(l lit) int32 {
	if lit_sign(l) != 0 {
		return -lit_var(l) - 1
	}
	return lit_var(l) + 1
}
func lit_read(s int32) lit {
	if s < -0x40000000 || s > 0x40000000 {
		panic("DIMACS literal exceeds SAT literal range")
	}
	if s == 0 {
		panic("lit_read: DIMACS zero is a clause terminator")
	}
	if s > 0 {
		return toLit(s - 1)
	}
	return lit_neg(toLit(-s - 1))
}
func lit_check(l lit, n int32) int32 {
	if l >= 0 && lit_var(l) < n {
		return 1
	}
	return 0
}

type stats_t struct {
	starts, clauses, learnts                         uint32
	decisions, propagations, inspects, conflicts     int64
	clauses_literals, learnts_literals, tot_literals int64
}
