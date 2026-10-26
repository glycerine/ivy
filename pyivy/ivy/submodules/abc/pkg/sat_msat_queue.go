package abc

// Fixed propagation queue from src/sat/msat/msatQueue.c.

import "slices"

// MsatQueue owns one fixed-capacity propagation epoch. Popped slots are not
// reused until Clear, matching the source queue's first/last cursors.
type MsatQueue struct {
	values      []int
	first, last int
	closed      bool
}

func NewMsatQueue(capacity int) (*MsatQueue, error) {
	if capacity < 0 {
		return nil, ErrDimension
	}
	return &MsatQueue{values: make([]int, capacity)}, nil
}

func (q *MsatQueue) checkLive() error {
	if q == nil {
		return ErrDimension
	}
	if q.closed || q.first < 0 || q.first > q.last || q.last > len(q.values) {
		return ErrInvariant
	}
	return nil
}

func (q *MsatQueue) Len() int {
	if q == nil || q.closed {
		return 0
	}
	return q.last - q.first
}

func (q *MsatQueue) Capacity() int {
	if q == nil || q.closed {
		return 0
	}
	return len(q.values)
}

// Insert appends a literal to the current epoch. Capacity is measured by the
// append cursor, not Len: callers must Clear before reusing extracted slots.
func (q *MsatQueue) Insert(literal int) error {
	if err := q.checkLive(); err != nil {
		return err
	}
	if q.last == len(q.values) {
		return ErrLimit
	}
	q.values[q.last] = literal
	q.last++
	return nil
}

// Extract removes the oldest pending literal. found is false for an empty
// queue, replacing the source's ambiguous -1 sentinel.
func (q *MsatQueue) Extract() (literal int, found bool, err error) {
	if err = q.checkLive(); err != nil {
		return 0, false, err
	}
	if q.first == q.last {
		return 0, false, nil
	}
	literal = q.values[q.first]
	q.first++
	return literal, true, nil
}

// Pending returns a detached FIFO-order snapshot.
func (q *MsatQueue) Pending() []int {
	if q == nil || q.closed || q.first > q.last || q.last > len(q.values) {
		return nil
	}
	return slices.Clone(q.values[q.first:q.last])
}

func (q *MsatQueue) Clear() error {
	if err := q.checkLive(); err != nil {
		return err
	}
	clear(q.values[:q.last])
	q.first, q.last = 0, 0
	return nil
}

// Close releases storage. It is idempotent; all other operations then reject.
func (q *MsatQueue) Close() {
	if q == nil || q.closed {
		return
	}
	clear(q.values)
	q.values = nil
	q.first, q.last = 0, 0
	q.closed = true
}
