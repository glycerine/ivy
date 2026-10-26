package abc

// Integer-ID float32 max heap from src/misc/vec/vecQue.h.

import (
	"fmt"
	"math"
	"strings"
)

// FloatPriorityQueue owns heap, inverse positions and optional priorities.
// Nil priorities mean float32(ID), matching C's default. Equal priorities do
// not move upward and prefer the left child downward, matching C exactly.
// The zero value is empty/usable. IDs and allocations default to a million.
// NaNs reject; infinities and signed zeros are supported. Not concurrent-safe.
type FloatPriorityQueue struct {
	heap, order []int
	costs       []float32
	limit       int
}

func NewFloatPriorityQueue(initialCapacity, maxIDs int) (*FloatPriorityQueue, error) {
	if initialCapacity < 0 || maxIDs < 0 {
		return nil, ErrDimension
	}
	q := &FloatPriorityQueue{limit: maxIDs}
	if e := q.grow(initialCapacity); e != nil {
		return nil, e
	}
	return q, nil
}
func (q *FloatPriorityQueue) maxIDs() int {
	if q.limit == 0 {
		return 1 << 20
	}
	return q.limit
}
func (q *FloatPriorityQueue) Len() int {
	if q == nil || len(q.heap) == 0 {
		return 0
	}
	return len(q.heap) - 1
}
func (q *FloatPriorityQueue) Contains(id int) bool {
	return q != nil && id >= 0 && id < len(q.order) && q.order[id] >= 1
}
func (q *FloatPriorityQueue) grow(required int) error {
	if required > q.maxIDs() {
		return ErrLimit
	}
	if required <= len(q.order) {
		return nil
	}
	size := max(required, 17)
	if len(q.order) <= q.maxIDs()/2 {
		size = max(size, 2*len(q.order))
	}
	size = min(size, q.maxIDs())
	// Heap needs one sentinel slot in addition to all permissible IDs.
	slots, e := checkedSize(size, 1, 1, int(^uint(0)>>1))
	if e != nil {
		return e
	}
	order := make([]int, size)
	for i := range order {
		order[i] = -1
	}
	copy(order, q.order)
	heap := make([]int, max(1, len(q.heap)), slots)
	if len(q.heap) == 0 {
		heap[0] = -1
	} else {
		copy(heap, q.heap)
	}
	q.order, q.heap = order, heap
	return nil
}
func (q *FloatPriorityQueue) priority(id int) float32 {
	if q.costs != nil {
		return q.costs[id]
	}
	return float32(id)
}

// Priority reads a configured ID's cost, even if it is not currently a member.
func (q *FloatPriorityQueue) Priority(id int) (float32, error) {
	if q == nil || id < 0 {
		return 0, ErrDimension
	}
	if id >= q.maxIDs() {
		return 0, ErrLimit
	}
	if q.costs != nil && id >= len(q.costs) {
		return 0, ErrDimension
	}
	return q.priority(id), nil
}

// SetPriorities owns a copy and rebuilds the heap, so it can safely replace C's
// pointer-to-caller-array setup. Nil selects ID priorities. All current members
// must have a cost. Failure leaves priorities and ordering unchanged.
func (q *FloatPriorityQueue) SetPriorities(costs []float32) error {
	if q == nil {
		return ErrDimension
	}
	if len(costs) > q.maxIDs() {
		return ErrLimit
	}
	for _, v := range costs {
		if math.IsNaN(float64(v)) {
			return ErrDimension
		}
	}
	if costs != nil {
		for _, id := range q.IDs() {
			if id >= len(costs) {
				return ErrDimension
			}
		}
	}
	if costs == nil {
		q.costs = nil
	} else {
		q.costs = append(make([]float32, 0, len(costs)), costs...)
	}
	for i := q.Len() / 2; i >= 1; i-- {
		q.moveDown(q.heap[i])
	}
	return nil
}
func (q *FloatPriorityQueue) moveUp(id int) bool {
	cost := q.priority(id)
	i := q.order[id]
	moved := false
	for i > 1 && cost > q.priority(q.heap[i/2]) {
		q.heap[i] = q.heap[i/2]
		q.order[q.heap[i]] = i
		i /= 2
		moved = true
	}
	q.heap[i] = id
	q.order[id] = i
	return moved
}
func (q *FloatPriorityQueue) moveDown(id int) {
	cost := q.priority(id)
	i := q.order[id]
	for child := 2 * i; child < len(q.heap); child = 2 * i {
		if child+1 < len(q.heap) && q.priority(q.heap[child]) < q.priority(q.heap[child+1]) {
			child++
		}
		if cost >= q.priority(q.heap[child]) {
			break
		}
		q.heap[i] = q.heap[child]
		q.order[q.heap[i]] = i
		i = child
	}
	q.heap[i] = id
	q.order[id] = i
}

// SetCost updates an owned priority and repairs a member upward or downward.
// Nonmembers may be updated for a later Push. ID-priority mode has no costs.
func (q *FloatPriorityQueue) SetCost(id int, value float32) error {
	if q == nil || id < 0 || q.costs == nil || id >= len(q.costs) || math.IsNaN(float64(value)) {
		return ErrDimension
	}
	q.costs[id] = value
	if q.Contains(id) && !q.moveUp(id) {
		q.moveDown(id)
	}
	return nil
}
func (q *FloatPriorityQueue) Push(id int) error {
	if _, e := q.Priority(id); e != nil {
		return e
	}
	if q.Contains(id) {
		return ErrDimension
	}
	required, e := checkedSize(id, 1, 1, q.maxIDs())
	if e != nil {
		return e
	}
	if e = q.grow(required); e != nil {
		return e
	}
	q.order[id] = len(q.heap)
	q.heap = append(q.heap, id)
	q.moveUp(id)
	return nil
}

// Top returns (-1, 0, false) for an empty queue instead of C's numeric sentinel.
func (q *FloatPriorityQueue) Top() (int, float32, bool) {
	if q.Len() == 0 {
		return -1, 0, false
	}
	id := q.heap[1]
	return id, q.priority(id), true
}
func (q *FloatPriorityQueue) Pop() (int, error) {
	if q.Len() == 0 {
		return -1, ErrDimension
	}
	id := q.heap[1]
	q.order[id] = -1
	last := q.heap[len(q.heap)-1]
	q.heap = q.heap[:len(q.heap)-1]
	if len(q.heap) > 1 {
		q.heap[1] = last
		q.order[last] = 1
		q.moveDown(last)
	}
	return id, nil
}
func (q *FloatPriorityQueue) Clear() {
	if q == nil {
		return
	}
	for _, id := range q.IDs() {
		q.order[id] = -1
	}
	if len(q.heap) > 0 {
		q.heap = q.heap[:1]
	}
}

// IDs returns an owned snapshot in heap order, suitable for deterministic diagnostics.
func (q *FloatPriorityQueue) IDs() []int {
	if q.Len() == 0 {
		return nil
	}
	return append([]int(nil), q.heap[1:]...)
}

// Format replaces Vec_QuePrint's process-global stdout side effect with the
// same level-by-level heap IDs and rounded priorities in an owned string.
func (q *FloatPriorityQueue) Format() string {
	if q.Len() == 0 {
		return ""
	}
	var output strings.Builder
	for start, width := 1, 1; start < len(q.heap); start, width = start+width, width*2 {
		end := min(start+width, len(q.heap))
		for _, id := range q.heap[start:end] {
			fmt.Fprintf(&output, "%-5d", id)
		}
		output.WriteByte('\n')
		for _, id := range q.heap[start:end] {
			fmt.Fprintf(&output, "%-5.0f", q.priority(id))
		}
		output.WriteString("\n\n")
	}
	return output.String()
}
func (q *FloatPriorityQueue) Clone() *FloatPriorityQueue {
	if q == nil {
		return nil
	}
	clone := &FloatPriorityQueue{heap: append([]int(nil), q.heap...), order: append([]int(nil), q.order...), limit: q.limit}
	if q.costs != nil {
		clone.costs = append(make([]float32, 0, len(q.costs)), q.costs...)
	}
	return clone
}

// Check validates arbitrary sparse membership, unlike C's dense-prefix checker.
func (q *FloatPriorityQueue) Check() error {
	if q == nil {
		return ErrDimension
	}
	if len(q.order) > q.maxIDs() || q.Len() > len(q.order) {
		return ErrInvariant
	}
	if len(q.heap) > 0 && q.heap[0] != -1 {
		return ErrInvariant
	}
	for id, pos := range q.order {
		if pos == -1 {
			continue
		}
		if pos < 1 || pos >= len(q.heap) || q.heap[pos] != id {
			return ErrInvariant
		}
	}
	for i := 1; i < len(q.heap); i++ {
		id := q.heap[i]
		if !q.Contains(id) || q.order[id] != i {
			return ErrInvariant
		}
		cost, e := q.Priority(id)
		if e != nil || math.IsNaN(float64(cost)) {
			return ErrInvariant
		}
		if i > 1 && cost > q.priority(q.heap[i/2]) {
			return ErrInvariant
		}
	}
	return nil
}
