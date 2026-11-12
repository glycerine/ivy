// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
// Go port of src/sat/satoko/utils/vec/vec_char.h.
package vec

import "fmt"
import "sort"

// vec_char_t stores a contiguous Go-owned buffer. size is the active prefix.
type vec_char_t struct {
	cap, size uint32
	data      []byte
}

func vec_char_alloc(capacity uint32) *vec_char_t {
	if capacity > 0 && capacity < 16 {
		capacity = 16
	}
	return vec_char_alloc_exact(capacity)
}
func vec_char_alloc_exact(capacity uint32) *vec_char_t {
	return &vec_char_t{cap: capacity, data: make([]byte, capacity)}
}
func vec_char_init(size uint32, value byte) *vec_char_t {
	p := vec_char_alloc_exact(size)
	p.size = size
	for i := range p.data {
		p.data[i] = value
	}
	return p
}
func vec_char_free(p *vec_char_t) {
	if p != nil {
		*p = vec_char_t{}
	}
}
func vec_char_size(p *vec_char_t) uint32     { return p.size }
func vec_char_capacity(p *vec_char_t) uint32 { return p.cap }
func vec_char_empty(p *vec_char_t) int32 {
	if p.size == 0 {
		return 1
	}
	return 0
}
func vec_char_reserve(p *vec_char_t, n uint32) {
	if n <= p.cap {
		return
	}
	buf := make([]byte, n)
	copy(buf, p.data)
	p.data = buf
	p.cap = n
}
func vec_char_resize(p *vec_char_t, n uint32) { vec_char_reserve(p, n); p.size = n }
func vec_char_shrink(p *vec_char_t, n uint32) {
	if n > p.cap {
		panic("vector shrink exceeds capacity")
	}
	p.size = n
}
func vec_char_erase(p *vec_char_t) { p.data = nil; p.size = 0; p.cap = 0 }
func vec_char_at(p *vec_char_t, i uint32) byte {
	if i >= p.size {
		panic("vector index out of range")
	}
	return p.data[i]
}
func vec_char_at_ptr(p *vec_char_t, i uint32) *byte {
	if i >= p.size {
		panic("vector index out of range")
	}
	return &p.data[i]
}
func vec_char_data(p *vec_char_t) []byte { return p.data[:p.size] }
func vec_char_duplicate(dst, src *vec_char_t) {
	if dst == src {
		return
	}
	vec_char_resize(dst, src.cap)
	copy(dst.data, src.data)
	dst.size = src.size
}
func vec_char_copy(dst, src *vec_char_t) {
	if dst == src {
		return
	}
	vec_char_resize(dst, src.size)
	copy(dst.data, src.data[:src.size])
}
func vec_char_push_back(p *vec_char_t, value byte) {
	if p.size == p.cap {
		n := 2 * p.cap
		if n < 16 {
			n = 16
		}
		vec_char_reserve(p, n)
	}
	p.data[p.size] = value
	p.size++
}
func vec_char_pop_back(p *vec_char_t) byte {
	if p.size == 0 {
		panic("pop empty vector")
	}
	p.size--
	return p.data[p.size]
}
func vec_char_assign(p *vec_char_t, i uint32, value byte) {
	if i >= p.size {
		panic("vector index out of range")
	}
	p.data[i] = value
}

// Insert shifts the entire old suffix, including its last element.
func vec_char_insert(p *vec_char_t, i uint32, value byte) {
	if i >= p.size {
		panic("vector index out of range")
	}
	vec_char_push_back(p, 0)
	copy(p.data[i+1:p.size], p.data[i:p.size-1])
	p.data[i] = value
}
func vec_char_drop(p *vec_char_t, i uint32) {
	if i >= p.size {
		panic("vector index out of range")
	}
	copy(p.data[i:p.size-1], p.data[i+1:p.size])
	p.size--
}
func vec_char_clear(p *vec_char_t) { p.size = 0 }
func vec_char_asc_compare(a, b *byte) int32 {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
func vec_char_desc_compare(a, b *byte) int32 { return vec_char_asc_compare(b, a) }
func vec_char_sort(p *vec_char_t, ascending int32) {
	sort.Slice(p.data[:p.size], func(i, j int) bool {
		if ascending != 0 {
			return p.data[i] < p.data[j]
		}
		return p.data[i] > p.data[j]
	})
}
func vec_char_memory(p *vec_char_t) int64 {
	if p == nil {
		return 0
	}
	return int64(p.cap)*1 + 32
}
func vec_char_print(p *vec_char_t) {
	fmt.Printf("Vector has %d(%d) entries: {", p.size, p.cap)
	for _, v := range p.data[:p.size] {
		fmt.Printf(" %v", v)
	}
	fmt.Println(" }")
}

// Bytes is a Go-owned Satoko vector. Data exposes the active elements.
type Bytes = vec_char_t

func NewBytes(capacity uint32) *Bytes         { return vec_char_alloc(capacity) }
func (p *vec_char_t) Len() uint32             { return p.size }
func (p *vec_char_t) Cap() uint32             { return p.cap }
func (p *vec_char_t) Data() []byte            { return vec_char_data(p) }
func (p *vec_char_t) At(i uint32) byte        { return vec_char_at(p, i) }
func (p *vec_char_t) Set(i uint32, v byte)    { vec_char_assign(p, i, v) }
func (p *vec_char_t) Push(v byte)             { vec_char_push_back(p, v) }
func (p *vec_char_t) Pop() byte               { return vec_char_pop_back(p) }
func (p *vec_char_t) Resize(n uint32)         { vec_char_resize(p, n) }
func (p *vec_char_t) Reserve(n uint32)        { vec_char_reserve(p, n) }
func (p *vec_char_t) Clear()                  { vec_char_clear(p) }
func (p *vec_char_t) Insert(i uint32, v byte) { vec_char_insert(p, i, v) }
func (p *vec_char_t) Drop(i uint32)           { vec_char_drop(p, i) }
