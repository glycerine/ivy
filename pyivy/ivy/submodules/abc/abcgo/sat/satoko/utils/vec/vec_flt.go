// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
// Go port of src/sat/satoko/utils/vec/vec_flt.h.
package vec

import "fmt"
import "sort"
import "strconv"

// vec_flt_t owns its active elements. len(data) and cap(data) are authoritative.
type vec_flt_t struct{ data []float32 }

func vec_flt_alloc(capacity uint32) *vec_flt_t {
	if capacity > 0 && capacity < 16 {
		capacity = 16
	}
	return vec_flt_alloc_exact(capacity)
}
func vec_flt_alloc_exact(capacity uint32) *vec_flt_t {
	return &vec_flt_t{data: make([]float32, 0, capacity)}
}
func vec_flt_init(size uint32, value float32) *vec_flt_t {
	p := &vec_flt_t{data: make([]float32, size)}
	for i := range p.data {
		p.data[i] = value
	}
	return p
}
func vec_flt_free(p *vec_flt_t) {
	if p != nil {
		p.data = nil
	}
}
func vec_flt_size(p *vec_flt_t) uint32     { return uint32(len(p.data)) }
func vec_flt_capacity(p *vec_flt_t) uint32 { return uint32(cap(p.data)) }
func vec_flt_empty(p *vec_flt_t) int32 {
	if len(p.data) == 0 {
		return 1
	}
	return 0
}
func vec_flt_reserve(p *vec_flt_t, n uint32) {
	if n <= uint32(cap(p.data)) {
		return
	}
	buf := make([]float32, len(p.data), n)
	copy(buf, p.data)
	p.data = buf
}
func vec_flt_resize(p *vec_flt_t, n uint32) { vec_flt_reserve(p, n); p.data = p.data[:n] }
func vec_flt_shrink(p *vec_flt_t, n uint32) {
	if n > uint32(cap(p.data)) {
		panic("vector shrink exceeds capacity")
	}
	p.data = p.data[:n]
}
func vec_flt_erase(p *vec_flt_t)                     { p.data = nil }
func vec_flt_at(p *vec_flt_t, i uint32) float32      { return p.data[i] }
func vec_flt_at_ptr(p *vec_flt_t, i uint32) *float32 { return &p.data[i] }
func vec_flt_data(p *vec_flt_t) []float32            { return p.data }
func vec_flt_duplicate(dst, src *vec_flt_t) {
	if dst == src {
		return
	}
	vec_flt_reserve(dst, uint32(cap(src.data)))
	copy(dst.data[:cap(src.data)], src.data[:cap(src.data)])
	dst.data = dst.data[:len(src.data)]
}
func vec_flt_copy(dst, src *vec_flt_t) {
	if dst == src {
		return
	}
	vec_flt_resize(dst, uint32(len(src.data)))
	copy(dst.data, src.data)
}
func vec_flt_push_back(p *vec_flt_t, value float32) {
	if len(p.data) == cap(p.data) {
		n := 2 * cap(p.data)
		if n < 16 {
			n = 16
		}
		vec_flt_reserve(p, uint32(n))
	}
	p.data = append(p.data, value)
}
func vec_flt_pop_back(p *vec_flt_t) float32 {
	if len(p.data) == 0 {
		panic("pop empty vector")
	}
	last := len(p.data) - 1
	value := p.data[last]
	p.data = p.data[:last]
	return value
}
func vec_flt_assign(p *vec_flt_t, i uint32, value float32) { p.data[i] = value }

// Insert shifts the entire old suffix, including its last element.
func vec_flt_insert(p *vec_flt_t, i uint32, value float32) {
	if i >= uint32(len(p.data)) {
		panic("vector index out of range")
	}
	vec_flt_push_back(p, 0)
	copy(p.data[i+1:], p.data[i:len(p.data)-1])
	p.data[i] = value
}
func vec_flt_drop(p *vec_flt_t, i uint32) {
	if i >= uint32(len(p.data)) {
		panic("vector index out of range")
	}
	copy(p.data[i:], p.data[i+1:])
	p.data = p.data[:len(p.data)-1]
}
func vec_flt_clear(p *vec_flt_t) { p.data = p.data[:0] }
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
	sort.Slice(p.data, func(i, j int) bool {
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
	return int64(cap(p.data))*4 + 3*int64(strconv.IntSize/8)
}
func vec_flt_print(p *vec_flt_t) {
	fmt.Printf("Vector has %d(%d) entries: {", len(p.data), cap(p.data))
	for _, v := range p.data {
		fmt.Printf(" %v", v)
	}
	fmt.Println(" }")
}

// Floats is a Go-owned Satoko vector. Data exposes the active elements.
type Floats = vec_flt_t

func NewFloats(capacity uint32) *Floats         { return vec_flt_alloc(capacity) }
func (p *vec_flt_t) Len() uint32                { return uint32(len(p.data)) }
func (p *vec_flt_t) Cap() uint32                { return uint32(cap(p.data)) }
func (p *vec_flt_t) Data() []float32            { return p.data }
func (p *vec_flt_t) At(i uint32) float32        { return vec_flt_at(p, i) }
func (p *vec_flt_t) Set(i uint32, v float32)    { vec_flt_assign(p, i, v) }
func (p *vec_flt_t) Push(v float32)             { vec_flt_push_back(p, v) }
func (p *vec_flt_t) Pop() float32               { return vec_flt_pop_back(p) }
func (p *vec_flt_t) Resize(n uint32)            { vec_flt_resize(p, n) }
func (p *vec_flt_t) Reserve(n uint32)           { vec_flt_reserve(p, n) }
func (p *vec_flt_t) Clear()                     { vec_flt_clear(p) }
func (p *vec_flt_t) Insert(i uint32, v float32) { vec_flt_insert(p, i, v) }
func (p *vec_flt_t) Drop(i uint32)              { vec_flt_drop(p, i) }
