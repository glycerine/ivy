// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
// Go port of src/sat/satoko/utils/vec/vec_uint.h.
package vec

import "fmt"
import "sort"

// vec_uint_t stores a contiguous Go-owned buffer. size is the active prefix.
type vec_uint_t struct {
	cap, size uint32
	data      []uint32
}

func vec_uint_alloc(capacity uint32) *vec_uint_t {
	if capacity > 0 && capacity < 16 {
		capacity = 16
	}
	return vec_uint_alloc_exact(capacity)
}
func vec_uint_alloc_exact(capacity uint32) *vec_uint_t {
	return &vec_uint_t{cap: capacity, data: make([]uint32, capacity)}
}
func vec_uint_init(size uint32, value uint32) *vec_uint_t {
	p := vec_uint_alloc_exact(size)
	p.size = size
	for i := range p.data {
		p.data[i] = value
	}
	return p
}
func vec_uint_free(p *vec_uint_t) {
	if p != nil {
		*p = vec_uint_t{}
	}
}
func vec_uint_size(p *vec_uint_t) uint32     { return p.size }
func vec_uint_capacity(p *vec_uint_t) uint32 { return p.cap }
func vec_uint_empty(p *vec_uint_t) int32 {
	if p.size == 0 {
		return 1
	}
	return 0
}
func vec_uint_reserve(p *vec_uint_t, n uint32) {
	if n <= p.cap {
		return
	}
	buf := make([]uint32, n)
	copy(buf, p.data)
	p.data = buf
	p.cap = n
}
func vec_uint_resize(p *vec_uint_t, n uint32) { vec_uint_reserve(p, n); p.size = n }
func vec_uint_shrink(p *vec_uint_t, n uint32) {
	if n > p.cap {
		panic("vector shrink exceeds capacity")
	}
	p.size = n
}
func vec_uint_erase(p *vec_uint_t) { p.data = nil; p.size = 0; p.cap = 0 }
func vec_uint_at(p *vec_uint_t, i uint32) uint32 {
	if i >= p.size {
		panic("vector index out of range")
	}
	return p.data[i]
}
func vec_uint_at_ptr(p *vec_uint_t, i uint32) *uint32 {
	if i >= p.size {
		panic("vector index out of range")
	}
	return &p.data[i]
}
func vec_uint_data(p *vec_uint_t) []uint32 { return p.data[:p.size] }
func vec_uint_duplicate(dst, src *vec_uint_t) {
	if dst == src {
		return
	}
	vec_uint_resize(dst, src.cap)
	copy(dst.data, src.data)
	dst.size = src.size
}
func vec_uint_copy(dst, src *vec_uint_t) {
	if dst == src {
		return
	}
	vec_uint_resize(dst, src.size)
	copy(dst.data, src.data[:src.size])
}
func vec_uint_push_back(p *vec_uint_t, value uint32) {
	if p.size == p.cap {
		n := 2 * p.cap
		if n < 16 {
			n = 16
		}
		vec_uint_reserve(p, n)
	}
	p.data[p.size] = value
	p.size++
}
func vec_uint_pop_back(p *vec_uint_t) uint32 {
	if p.size == 0 {
		panic("pop empty vector")
	}
	p.size--
	return p.data[p.size]
}
func vec_uint_assign(p *vec_uint_t, i uint32, value uint32) {
	if i >= p.size {
		panic("vector index out of range")
	}
	p.data[i] = value
}

// Insert shifts the entire old suffix, including its last element.
func vec_uint_insert(p *vec_uint_t, i uint32, value uint32) {
	if i >= p.size {
		panic("vector index out of range")
	}
	vec_uint_push_back(p, 0)
	copy(p.data[i+1:p.size], p.data[i:p.size-1])
	p.data[i] = value
}
func vec_uint_drop(p *vec_uint_t, i uint32) {
	if i >= p.size {
		panic("vector index out of range")
	}
	copy(p.data[i:p.size-1], p.data[i+1:p.size])
	p.size--
}
func vec_uint_clear(p *vec_uint_t) { p.size = 0 }
func vec_uint_asc_compare(a, b *uint32) int32 {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
func vec_uint_desc_compare(a, b *uint32) int32 { return vec_uint_asc_compare(b, a) }
func vec_uint_sort(p *vec_uint_t, ascending int32) {
	sort.Slice(p.data[:p.size], func(i, j int) bool {
		if ascending != 0 {
			return p.data[i] < p.data[j]
		}
		return p.data[i] > p.data[j]
	})
}
func vec_uint_memory(p *vec_uint_t) int64 {
	if p == nil {
		return 0
	}
	return int64(p.cap)*4 + 32
}
func vec_uint_print(p *vec_uint_t) {
	fmt.Printf("Vector has %d(%d) entries: {", p.size, p.cap)
	for _, v := range p.data[:p.size] {
		fmt.Printf(" %v", v)
	}
	fmt.Println(" }")
}

func vec_uint_find(p *vec_uint_t, value uint32) uint32 {
	for _, v := range p.data[:p.size] {
		if v == value {
			return 1
		}
	}
	return 0
}

// Uints is a Go-owned Satoko vector. Data exposes the active elements.
type Uints = vec_uint_t

func NewUints(capacity uint32) *Uints           { return vec_uint_alloc(capacity) }
func (p *vec_uint_t) Len() uint32               { return p.size }
func (p *vec_uint_t) Cap() uint32               { return p.cap }
func (p *vec_uint_t) Data() []uint32            { return vec_uint_data(p) }
func (p *vec_uint_t) At(i uint32) uint32        { return vec_uint_at(p, i) }
func (p *vec_uint_t) Set(i uint32, v uint32)    { vec_uint_assign(p, i, v) }
func (p *vec_uint_t) Push(v uint32)             { vec_uint_push_back(p, v) }
func (p *vec_uint_t) Pop() uint32               { return vec_uint_pop_back(p) }
func (p *vec_uint_t) Resize(n uint32)           { vec_uint_resize(p, n) }
func (p *vec_uint_t) Reserve(n uint32)          { vec_uint_reserve(p, n) }
func (p *vec_uint_t) Clear()                    { vec_uint_clear(p) }
func (p *vec_uint_t) Insert(i uint32, v uint32) { vec_uint_insert(p, i, v) }
func (p *vec_uint_t) Drop(i uint32)             { vec_uint_drop(p, i) }
