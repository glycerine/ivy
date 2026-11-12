// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
// Go port of src/sat/satoko/utils/vec/vec_flt.h.
package vec

import "fmt"
import "sort"

// vec_flt_t stores a contiguous Go-owned buffer. size is the active prefix.
type vec_flt_t struct {
	cap, size uint32
	data      []float32
}

func vec_flt_alloc(capacity uint32) *vec_flt_t {
	if capacity > 0 && capacity < 16 {
		capacity = 16
	}
	return vec_flt_alloc_exact(capacity)
}
func vec_flt_alloc_exact(capacity uint32) *vec_flt_t {
	return &vec_flt_t{cap: capacity, data: make([]float32, capacity)}
}
func vec_flt_init(size uint32, value float32) *vec_flt_t {
	p := vec_flt_alloc_exact(size)
	p.size = size
	for i := range p.data {
		p.data[i] = value
	}
	return p
}
func vec_flt_free(p *vec_flt_t) {
	if p != nil {
		*p = vec_flt_t{}
	}
}
func vec_flt_size(p *vec_flt_t) uint32     { return p.size }
func vec_flt_capacity(p *vec_flt_t) uint32 { return p.cap }
func vec_flt_empty(p *vec_flt_t) int32 {
	if p.size == 0 {
		return 1
	}
	return 0
}
func vec_flt_reserve(p *vec_flt_t, n uint32) {
	if n <= p.cap {
		return
	}
	buf := make([]float32, n)
	copy(buf, p.data)
	p.data = buf
	p.cap = n
}
func vec_flt_resize(p *vec_flt_t, n uint32) { vec_flt_reserve(p, n); p.size = n }
func vec_flt_shrink(p *vec_flt_t, n uint32) {
	if n > p.cap {
		panic("vector shrink exceeds capacity")
	}
	p.size = n
}
func vec_flt_erase(p *vec_flt_t) { p.data = nil; p.size = 0; p.cap = 0 }
func vec_flt_at(p *vec_flt_t, i uint32) float32 {
	if i >= p.size {
		panic("vector index out of range")
	}
	return p.data[i]
}
func vec_flt_at_ptr(p *vec_flt_t, i uint32) *float32 {
	if i >= p.size {
		panic("vector index out of range")
	}
	return &p.data[i]
}
func vec_flt_data(p *vec_flt_t) []float32 { return p.data[:p.size] }
func vec_flt_duplicate(dst, src *vec_flt_t) {
	if dst == src {
		return
	}
	vec_flt_resize(dst, src.cap)
	copy(dst.data, src.data)
	dst.size = src.size
}
func vec_flt_copy(dst, src *vec_flt_t) {
	if dst == src {
		return
	}
	vec_flt_resize(dst, src.size)
	copy(dst.data, src.data[:src.size])
}
func vec_flt_push_back(p *vec_flt_t, value float32) {
	if p.size == p.cap {
		n := 2 * p.cap
		if n < 16 {
			n = 16
		}
		vec_flt_reserve(p, n)
	}
	p.data[p.size] = value
	p.size++
}
func vec_flt_pop_back(p *vec_flt_t) float32 {
	if p.size == 0 {
		panic("pop empty vector")
	}
	p.size--
	return p.data[p.size]
}
func vec_flt_assign(p *vec_flt_t, i uint32, value float32) {
	if i >= p.size {
		panic("vector index out of range")
	}
	p.data[i] = value
}

// Insert shifts the entire old suffix, including its last element.
func vec_flt_insert(p *vec_flt_t, i uint32, value float32) {
	if i >= p.size {
		panic("vector index out of range")
	}
	vec_flt_push_back(p, 0)
	copy(p.data[i+1:p.size], p.data[i:p.size-1])
	p.data[i] = value
}
func vec_flt_drop(p *vec_flt_t, i uint32) {
	if i >= p.size {
		panic("vector index out of range")
	}
	copy(p.data[i:p.size-1], p.data[i+1:p.size])
	p.size--
}
func vec_flt_clear(p *vec_flt_t) { p.size = 0 }
func vec_flt_asc_compare(a, b *float32) int32 {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
func vec_flt_desc_compare(a, b *float32) int32 { return vec_flt_asc_compare(b, a) }
func vec_flt_sort(p *vec_flt_t, ascending int32) {
	sort.Slice(p.data[:p.size], func(i, j int) bool {
		if ascending != 0 {
			return p.data[i] < p.data[j]
		}
		return p.data[i] > p.data[j]
	})
}
func vec_flt_memory(p *vec_flt_t) int64 {
	if p == nil {
		return 0
	}
	return int64(p.cap)*4 + 32
}
func vec_flt_print(p *vec_flt_t) {
	fmt.Printf("Vector has %d(%d) entries: {", p.size, p.cap)
	for _, v := range p.data[:p.size] {
		fmt.Printf(" %v", v)
	}
	fmt.Println(" }")
}

// Floats is a Go-owned Satoko vector. Data exposes the active elements.
type Floats = vec_flt_t

func NewFloats(capacity uint32) *Floats         { return vec_flt_alloc(capacity) }
func (p *vec_flt_t) Len() uint32                { return p.size }
func (p *vec_flt_t) Cap() uint32                { return p.cap }
func (p *vec_flt_t) Data() []float32            { return vec_flt_data(p) }
func (p *vec_flt_t) At(i uint32) float32        { return vec_flt_at(p, i) }
func (p *vec_flt_t) Set(i uint32, v float32)    { vec_flt_assign(p, i, v) }
func (p *vec_flt_t) Push(v float32)             { vec_flt_push_back(p, v) }
func (p *vec_flt_t) Pop() float32               { return vec_flt_pop_back(p) }
func (p *vec_flt_t) Resize(n uint32)            { vec_flt_resize(p, n) }
func (p *vec_flt_t) Reserve(n uint32)           { vec_flt_reserve(p, n) }
func (p *vec_flt_t) Clear()                     { vec_flt_clear(p) }
func (p *vec_flt_t) Insert(i uint32, v float32) { vec_flt_insert(p, i, v) }
func (p *vec_flt_t) Drop(i uint32)              { vec_flt_drop(p, i) }
