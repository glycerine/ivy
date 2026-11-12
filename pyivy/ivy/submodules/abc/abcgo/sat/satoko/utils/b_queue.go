// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package utils

// BoundedQueue keeps the most recent cap(data) unsigned observations. Its slice
// length is the number of observations; next identifies the oldest full slot.
type BoundedQueue struct {
	data []uint32
	next int
	sum  uint64
}
type b_queue_t = BoundedQueue

func NewBoundedQueue(capacity uint32) *BoundedQueue {
	if capacity == 0 {
		panic("bounded queue capacity must be positive")
	}
	return &BoundedQueue{data: make([]uint32, 0, capacity)}
}
func b_queue_alloc(capacity uint32) *b_queue_t { return NewBoundedQueue(capacity) }
func b_queue_free(p *b_queue_t) {
	if p != nil {
		*p = b_queue_t{}
	}
}
func (p *BoundedQueue) Push(value uint32) {
	if cap(p.data) == 0 {
		panic("push to released bounded queue")
	}
	if len(p.data) < cap(p.data) {
		p.data = append(p.data, value)
	} else {
		p.sum -= uint64(p.data[p.next])
		p.data[p.next] = value
		p.next++
		if p.next == len(p.data) {
			p.next = 0
		}
	}
	p.sum += uint64(value)
}
func b_queue_push(p *b_queue_t, value uint32) { p.Push(value) }
func (p *BoundedQueue) Average() uint32 {
	if len(p.data) == 0 {
		panic("average of empty bounded queue")
	}
	return uint32(p.sum / uint64(len(p.data)))
}
func b_queue_avg(p *b_queue_t) uint32 { return p.Average() }
func (p *BoundedQueue) Full() bool    { return len(p.data) == cap(p.data) }
func b_queue_is_valid(p *b_queue_t) uint32 {
	if p.Full() {
		return 1
	}
	return 0
}
func (p *BoundedQueue) Clear()   { p.data = p.data[:0]; p.next = 0; p.sum = 0 }
func b_queue_clean(p *b_queue_t) { p.Clear() }
