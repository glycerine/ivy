// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
// Go port of src/sat/satoko/utils/vec/vec_char.h.
package vec

import (
	"fmt"
	"sort"
	"strconv"
)

// vec_char_t owns its active elements. len(data) and cap(data) are authoritative.
type vec_char_t struct{ data []byte }

func vec_char_alloc(capacity uint32) *vec_char_t {
	if capacity > 0 && capacity < 16 {
		capacity = 16
	}
	return vec_char_alloc_exact(capacity)
}
func vec_char_alloc_exact(capacity uint32) *vec_char_t {
	return &vec_char_t{data: make([]byte, 0, capacity)}
}
func vec_char_init(size uint32, value byte) *vec_char_t {
	p := &vec_char_t{data: make([]byte, size)}
	for i := range p.data {
		p.data[i] = value
	}
	return p
}
func vec_char_free(p *vec_char_t) {
	if p != nil {
		p.data = nil
	}
}
func vec_char_size(p *vec_char_t) uint32     { return uint32(len(p.data)) }
func vec_char_capacity(p *vec_char_t) uint32 { return uint32(cap(p.data)) }
func vec_char_empty(p *vec_char_t) int32 {
	if len(p.data) == 0 {
		return 1
	}
	return 0
}
func vec_char_reserve(p *vec_char_t, n uint32) {
	if n <= uint32(cap(p.data)) {
		return
	}
	buf := make([]byte, len(p.data), n)
	copy(buf, p.data)
	p.data = buf
}
func vec_char_resize(p *vec_char_t, n uint32) { vec_char_reserve(p, n); p.data = p.data[:n] }
func vec_char_shrink(p *vec_char_t, n uint32) {
	if n > uint32(cap(p.data)) {
		panic("vector shrink exceeds capacity")
	}
	p.data = p.data[:n]
}
func vec_char_erase(p *vec_char_t)                  { p.data = nil }
func vec_char_at(p *vec_char_t, i uint32) byte      { return p.data[i] }
func vec_char_at_ptr(p *vec_char_t, i uint32) *byte { return &p.data[i] }
func vec_char_data(p *vec_char_t) []byte            { return p.data }
func vec_char_duplicate(dst, src *vec_char_t) {
	if dst == src {
		return
	}
	vec_char_reserve(dst, uint32(cap(src.data)))
	copy(dst.data[:cap(src.data)], src.data[:cap(src.data)])
	dst.data = dst.data[:len(src.data)]
}
func vec_char_copy(dst, src *vec_char_t) {
	if dst == src {
		return
	}
	vec_char_resize(dst, uint32(len(src.data)))
	copy(dst.data, src.data)
}
func vec_char_push_back(p *vec_char_t, value byte) {
	if len(p.data) == cap(p.data) {
		n := 2 * cap(p.data)
		if n < 16 {
			n = 16
		}
		vec_char_reserve(p, uint32(n))
	}
	p.data = append(p.data, value)
}
func vec_char_pop_back(p *vec_char_t) byte {
	if len(p.data) == 0 {
		panic("pop empty vector")
	}
	last := len(p.data) - 1
	value := p.data[last]
	p.data = p.data[:last]
	return value
}
func vec_char_assign(p *vec_char_t, i uint32, value byte) { p.data[i] = value }

// Insert shifts the entire old suffix, including its last element.
func vec_char_insert(p *vec_char_t, i uint32, value byte) {
	if i >= uint32(len(p.data)) {
		panic("vector index out of range")
	}
	vec_char_push_back(p, 0)
	copy(p.data[i+1:], p.data[i:len(p.data)-1])
	p.data[i] = value
}
func vec_char_drop(p *vec_char_t, i uint32) {
	if i >= uint32(len(p.data)) {
		panic("vector index out of range")
	}
	copy(p.data[i:], p.data[i+1:])
	p.data = p.data[:len(p.data)-1]
}
func vec_char_clear(p *vec_char_t) { p.data = p.data[:0] }
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
	sort.Slice(p.data, func(i, j int) bool {
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
	return int64(cap(p.data))*1 + 3*int64(strconv.IntSize/8)
}
func vec_char_print(p *vec_char_t) {
	fmt.Printf("Vector has %d(%d) entries: {", len(p.data), cap(p.data))
	for _, v := range p.data {
		fmt.Printf(" %v", v)
	}
	fmt.Println(" }")
}

// Bytes is a Go-owned Satoko vector. Data exposes the active elements.
type Bytes = vec_char_t

func NewBytes(capacity uint32) *Bytes         { return vec_char_alloc(capacity) }
func (p *vec_char_t) Len() uint32             { return uint32(len(p.data)) }
func (p *vec_char_t) Cap() uint32             { return uint32(cap(p.data)) }
func (p *vec_char_t) Data() []byte            { return p.data }
func (p *vec_char_t) At(i uint32) byte        { return vec_char_at(p, i) }
func (p *vec_char_t) Set(i uint32, v byte)    { vec_char_assign(p, i, v) }
func (p *vec_char_t) Push(v byte)             { vec_char_push_back(p, v) }
func (p *vec_char_t) Pop() byte               { return vec_char_pop_back(p) }
func (p *vec_char_t) Resize(n uint32)         { vec_char_resize(p, n) }
func (p *vec_char_t) Reserve(n uint32)        { vec_char_reserve(p, n) }
func (p *vec_char_t) Clear()                  { vec_char_clear(p) }
func (p *vec_char_t) Insert(i uint32, v byte) { vec_char_insert(p, i, v) }
func (p *vec_char_t) Drop(i uint32)           { vec_char_drop(p, i) }
