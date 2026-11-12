// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
// Go port of src/sat/satoko/utils/vec/vec_int.h.
package vec

import "fmt"
import "sort"

// vec_int_t stores a contiguous Go-owned buffer. size is the active prefix.
type vec_int_t struct {
	cap, size uint32
	data      []int32
}

func vec_int_alloc(capacity uint32) *vec_int_t {
	if capacity > 0 && capacity < 16 {
		capacity = 16
	}
	return vec_int_alloc_exact(capacity)
}
func vec_int_alloc_exact(capacity uint32) *vec_int_t {
	return &vec_int_t{cap: capacity, data: make([]int32, capacity)}
}
func vec_int_init(size uint32, value int32) *vec_int_t {
	p := vec_int_alloc_exact(size)
	p.size = size
	for i := range p.data {
		p.data[i] = value
	}
	return p
}
func vec_int_free(p *vec_int_t) {
	if p != nil {
		*p = vec_int_t{}
	}
}
func vec_int_size(p *vec_int_t) uint32     { return p.size }
func vec_int_capacity(p *vec_int_t) uint32 { return p.cap }
func vec_int_empty(p *vec_int_t) int32 {
	if p.size == 0 {
		return 1
	}
	return 0
}
func vec_int_reserve(p *vec_int_t, n uint32) {
	if n <= p.cap {
		return
	}
	buf := make([]int32, n)
	copy(buf, p.data)
	p.data = buf
	p.cap = n
}
func vec_int_resize(p *vec_int_t, n uint32) { vec_int_reserve(p, n); p.size = n }
func vec_int_shrink(p *vec_int_t, n uint32) {
	if n > p.cap {
		panic("vector shrink exceeds capacity")
	}
	p.size = n
}
func vec_int_erase(p *vec_int_t) { p.data = nil; p.size = 0; p.cap = 0 }
func vec_int_at(p *vec_int_t, i uint32) int32 {
	if i >= p.size {
		panic("vector index out of range")
	}
	return p.data[i]
}
func vec_int_at_ptr(p *vec_int_t, i uint32) *int32 {
	if i >= p.size {
		panic("vector index out of range")
	}
	return &p.data[i]
}
func vec_int_data(p *vec_int_t) []int32 { return p.data[:p.size] }
func vec_int_duplicate(dst, src *vec_int_t) {
	if dst == src {
		return
	}
	vec_int_resize(dst, src.cap)
	copy(dst.data, src.data)
	dst.size = src.size
}
func vec_int_copy(dst, src *vec_int_t) {
	if dst == src {
		return
	}
	vec_int_resize(dst, src.size)
	copy(dst.data, src.data[:src.size])
}
func vec_int_push_back(p *vec_int_t, value int32) {
	if p.size == p.cap {
		n := 2 * p.cap
		if n < 16 {
			n = 16
		}
		vec_int_reserve(p, n)
	}
	p.data[p.size] = value
	p.size++
}
func vec_int_pop_back(p *vec_int_t) int32 {
	if p.size == 0 {
		panic("pop empty vector")
	}
	p.size--
	return p.data[p.size]
}
func vec_int_assign(p *vec_int_t, i uint32, value int32) {
	if i >= p.size {
		panic("vector index out of range")
	}
	p.data[i] = value
}

// Insert shifts the entire old suffix, including its last element.
func vec_int_insert(p *vec_int_t, i uint32, value int32) {
	if i >= p.size {
		panic("vector index out of range")
	}
	vec_int_push_back(p, 0)
	copy(p.data[i+1:p.size], p.data[i:p.size-1])
	p.data[i] = value
}
func vec_int_drop(p *vec_int_t, i uint32) {
	if i >= p.size {
		panic("vector index out of range")
	}
	copy(p.data[i:p.size-1], p.data[i+1:p.size])
	p.size--
}
func vec_int_clear(p *vec_int_t) { p.size = 0 }
func vec_int_asc_compare(a, b *int32) int32 {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
func vec_int_desc_compare(a, b *int32) int32 { return vec_int_asc_compare(b, a) }
func vec_int_sort(p *vec_int_t, ascending int32) {
	sort.Slice(p.data[:p.size], func(i, j int) bool {
		if ascending != 0 {
			return p.data[i] < p.data[j]
		}
		return p.data[i] > p.data[j]
	})
}
func vec_int_memory(p *vec_int_t) int64 {
	if p == nil {
		return 0
	}
	return int64(p.cap)*4 + 32
}
func vec_int_print(p *vec_int_t) {
	fmt.Printf("Vector has %d(%d) entries: {", p.size, p.cap)
	for _, v := range p.data[:p.size] {
		fmt.Printf(" %v", v)
	}
	fmt.Println(" }")
}

// Ints is a Go-owned Satoko vector. Data exposes the active elements.
type Ints = vec_int_t

func NewInts(capacity uint32) *Ints           { return vec_int_alloc(capacity) }
func (p *vec_int_t) Len() uint32              { return p.size }
func (p *vec_int_t) Cap() uint32              { return p.cap }
func (p *vec_int_t) Data() []int32            { return vec_int_data(p) }
func (p *vec_int_t) At(i uint32) int32        { return vec_int_at(p, i) }
func (p *vec_int_t) Set(i uint32, v int32)    { vec_int_assign(p, i, v) }
func (p *vec_int_t) Push(v int32)             { vec_int_push_back(p, v) }
func (p *vec_int_t) Pop() int32               { return vec_int_pop_back(p) }
func (p *vec_int_t) Resize(n uint32)          { vec_int_resize(p, n) }
func (p *vec_int_t) Reserve(n uint32)         { vec_int_reserve(p, n) }
func (p *vec_int_t) Clear()                   { vec_int_clear(p) }
func (p *vec_int_t) Insert(i uint32, v int32) { vec_int_insert(p, i, v) }
func (p *vec_int_t) Drop(i uint32)            { vec_int_drop(p, i) }
