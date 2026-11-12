// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package utils

// BoundedQueue keeps the latest capacity unsigned observations and their sum.
type BoundedQueue struct {
	size, cap, i_first, i_empty uint32
	sum                         uint64
	data                        []uint32
}
type b_queue_t = BoundedQueue

func NewBoundedQueue(capacity uint32) *BoundedQueue {
	if capacity == 0 {
		panic("bounded queue capacity must be positive")
	}
	return &BoundedQueue{cap: capacity, data: make([]uint32, capacity)}
}
func b_queue_alloc(capacity uint32) *b_queue_t { return NewBoundedQueue(capacity) }
func b_queue_free(p *b_queue_t) {
	if p != nil {
		*p = b_queue_t{}
	}
}
func (p *BoundedQueue) Push(value uint32) {
	if p.cap == 0 {
		panic("push to released bounded queue")
	}
	if p.size == p.cap {
		p.sum -= uint64(p.data[p.i_first])
		p.i_first = (p.i_first + 1) % p.cap
	} else {
		p.size++
	}
	p.sum += uint64(value)
	p.data[p.i_empty] = value
	p.i_empty = (p.i_empty + 1) % p.cap
}
func b_queue_push(p *b_queue_t, value uint32) { p.Push(value) }
func (p *BoundedQueue) Average() uint32 {
	if p.size == 0 {
		panic("average of empty bounded queue")
	}
	return uint32(p.sum / uint64(p.size))
}
func b_queue_avg(p *b_queue_t) uint32 { return p.Average() }
func (p *BoundedQueue) Full() bool    { return p.size == p.cap }
func b_queue_is_valid(p *b_queue_t) uint32 {
	if p.Full() {
		return 1
	}
	return 0
}
func (p *BoundedQueue) Clear()   { p.i_first = 0; p.i_empty = 0; p.size = 0; p.sum = 0 }
func b_queue_clean(p *b_queue_t) { p.Clear() }
