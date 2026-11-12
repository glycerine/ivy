// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
// Go port of src/sat/satoko/utils/vec/vec_sdbl.h.
package vec

import "fmt"
import "sort"

// vec_sdbl_t stores a contiguous Go-owned buffer. size is the active prefix.
type vec_sdbl_t struct {
	cap, size uint32
	data      []uint64
}

func vec_sdbl_alloc(capacity uint32) *vec_sdbl_t {
	if capacity > 0 && capacity < 16 {
		capacity = 16
	}
	return vec_sdbl_alloc_exact(capacity)
}
func vec_sdbl_alloc_exact(capacity uint32) *vec_sdbl_t {
	return &vec_sdbl_t{cap: capacity, data: make([]uint64, capacity)}
}
func vec_sdbl_init(size uint32, value uint64) *vec_sdbl_t {
	p := vec_sdbl_alloc_exact(size)
	p.size = size
	for i := range p.data {
		p.data[i] = value
	}
	return p
}
func vec_sdbl_free(p *vec_sdbl_t) {
	if p != nil {
		*p = vec_sdbl_t{}
	}
}
func vec_sdbl_size(p *vec_sdbl_t) uint32     { return p.size }
func vec_sdbl_capacity(p *vec_sdbl_t) uint32 { return p.cap }
func vec_sdbl_empty(p *vec_sdbl_t) int32 {
	if p.size == 0 {
		return 1
	}
	return 0
}
func vec_sdbl_reserve(p *vec_sdbl_t, n uint32) {
	if n <= p.cap {
		return
	}
	buf := make([]uint64, n)
	copy(buf, p.data)
	p.data = buf
	p.cap = n
}
func vec_sdbl_resize(p *vec_sdbl_t, n uint32) { vec_sdbl_reserve(p, n); p.size = n }
func vec_sdbl_shrink(p *vec_sdbl_t, n uint32) {
	if n > p.cap {
		panic("vector shrink exceeds capacity")
	}
	p.size = n
}
func vec_sdbl_erase(p *vec_sdbl_t) { p.data = nil; p.size = 0; p.cap = 0 }
func vec_sdbl_at(p *vec_sdbl_t, i uint32) uint64 {
	if i >= p.size {
		panic("vector index out of range")
	}
	return p.data[i]
}
func vec_sdbl_at_ptr(p *vec_sdbl_t, i uint32) *uint64 {
	if i >= p.size {
		panic("vector index out of range")
	}
	return &p.data[i]
}
func vec_sdbl_data(p *vec_sdbl_t) []uint64 { return p.data[:p.size] }
func vec_sdbl_duplicate(dst, src *vec_sdbl_t) {
	if dst == src {
		return
	}
	vec_sdbl_resize(dst, src.cap)
	copy(dst.data, src.data)
	dst.size = src.size
}
func vec_sdbl_copy(dst, src *vec_sdbl_t) {
	if dst == src {
		return
	}
	vec_sdbl_resize(dst, src.size)
	copy(dst.data, src.data[:src.size])
}
func vec_sdbl_push_back(p *vec_sdbl_t, value uint64) {
	if p.size == p.cap {
		n := 2 * p.cap
		if n < 16 {
			n = 16
		}
		vec_sdbl_reserve(p, n)
	}
	p.data[p.size] = value
	p.size++
}
func vec_sdbl_pop_back(p *vec_sdbl_t) uint64 {
	if p.size == 0 {
		panic("pop empty vector")
	}
	p.size--
	return p.data[p.size]
}
func vec_sdbl_assign(p *vec_sdbl_t, i uint32, value uint64) {
	if i >= p.size {
		panic("vector index out of range")
	}
	p.data[i] = value
}

// Insert shifts the entire old suffix, including its last element.
func vec_sdbl_insert(p *vec_sdbl_t, i uint32, value uint64) {
	if i >= p.size {
		panic("vector index out of range")
	}
	vec_sdbl_push_back(p, 0)
	copy(p.data[i+1:p.size], p.data[i:p.size-1])
	p.data[i] = value
}
func vec_sdbl_drop(p *vec_sdbl_t, i uint32) {
	if i >= p.size {
		panic("vector index out of range")
	}
	copy(p.data[i:p.size-1], p.data[i+1:p.size])
	p.size--
}
func vec_sdbl_clear(p *vec_sdbl_t) { p.size = 0 }
func vec_sdbl_asc_compare(a, b *uint64) int32 {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
func vec_sdbl_desc_compare(a, b *uint64) int32 { return vec_sdbl_asc_compare(b, a) }
func vec_sdbl_sort(p *vec_sdbl_t, ascending int32) {
	sort.Slice(p.data[:p.size], func(i, j int) bool {
		if ascending != 0 {
			return p.data[i] < p.data[j]
		}
		return p.data[i] > p.data[j]
	})
}
func vec_sdbl_memory(p *vec_sdbl_t) int64 {
	if p == nil {
		return 0
	}
	return int64(p.cap)*8 + 32
}
func vec_sdbl_print(p *vec_sdbl_t) {
	fmt.Printf("Vector has %d(%d) entries: {", p.size, p.cap)
	for _, v := range p.data[:p.size] {
		fmt.Printf(" %v", v)
	}
	fmt.Println(" }")
}

// Activities is a Go-owned Satoko vector. Data exposes the active elements.
type Activities = vec_sdbl_t

func NewActivities(capacity uint32) *Activities { return vec_sdbl_alloc(capacity) }
func (p *vec_sdbl_t) Len() uint32               { return p.size }
func (p *vec_sdbl_t) Cap() uint32               { return p.cap }
func (p *vec_sdbl_t) Data() []uint64            { return vec_sdbl_data(p) }
func (p *vec_sdbl_t) At(i uint32) uint64        { return vec_sdbl_at(p, i) }
func (p *vec_sdbl_t) Set(i uint32, v uint64)    { vec_sdbl_assign(p, i, v) }
func (p *vec_sdbl_t) Push(v uint64)             { vec_sdbl_push_back(p, v) }
func (p *vec_sdbl_t) Pop() uint64               { return vec_sdbl_pop_back(p) }
func (p *vec_sdbl_t) Resize(n uint32)           { vec_sdbl_resize(p, n) }
func (p *vec_sdbl_t) Reserve(n uint32)          { vec_sdbl_reserve(p, n) }
func (p *vec_sdbl_t) Clear()                    { vec_sdbl_clear(p) }
func (p *vec_sdbl_t) Insert(i uint32, v uint64) { vec_sdbl_insert(p, i, v) }
func (p *vec_sdbl_t) Drop(i uint32)             { vec_sdbl_drop(p, i) }
