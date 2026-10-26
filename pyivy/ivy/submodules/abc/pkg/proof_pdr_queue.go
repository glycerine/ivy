package abc

// Port of src/proof/pdr/pdrUtil.c: Pdr_Obl* and Pdr_Queue* operations.

// PdrObligation carries a cube and a successor chain toward a failed property.
// Scheduling fields can change only while the obligation is outside a queue.
// State/Next are immutable; Go lifetimes replace allocation reference counts.
type PdrObligation struct {
	frame, priority int
	state           *PdrCube
	next            *PdrObligation
	link            *PdrObligation
	owner           *PdrQueue
}

func NewPdrObligation(frame, priority int, state *PdrCube, next *PdrObligation) (*PdrObligation, error) {
	if frame < 0 || state == nil {
		return nil, ErrDimension
	}
	if err := state.Check(); err != nil {
		return nil, err
	}
	return &PdrObligation{frame: frame, priority: priority, state: state, next: next}, nil
}
func (o *PdrObligation) Frame() int           { return o.frame }
func (o *PdrObligation) Priority() int        { return o.priority }
func (o *PdrObligation) State() *PdrCube      { return o.state }
func (o *PdrObligation) Next() *PdrObligation { return o.next }
func (o *PdrObligation) Reschedule(frame, priority int) error {
	if o.owner != nil {
		return ErrInvariant
	}
	if frame < 0 {
		return ErrDimension
	}
	o.frame = frame
	o.priority = priority
	return nil
}

// PdrQueue is the stable ascending frame/priority intrusive queue from C.
// Its zero value is usable. It must not be copied while populated.
type PdrQueue struct {
	head            *PdrObligation
	length, maximum int
	pushes          uint64
}

func (q *PdrQueue) Len() int             { return q.length }
func (q *PdrQueue) Empty() bool          { return q.head == nil }
func (q *PdrQueue) Peek() *PdrObligation { return q.head }
func (q *PdrQueue) MaximumLen() int      { return q.maximum }
func (q *PdrQueue) PushCount() uint64    { return q.pushes }
func (q *PdrQueue) Push(o *PdrObligation) error {
	if o == nil || o.state == nil || o.frame < 0 {
		return ErrDimension
	}
	if o.owner != nil || o.link != nil {
		return ErrInvariant
	}
	previous := &q.head
	for *previous != nil {
		current := *previous
		if current.frame > o.frame || (current.frame == o.frame && current.priority > o.priority) {
			break
		}
		previous = &current.link
	}
	o.link = *previous
	*previous = o
	o.owner = q
	q.length++
	q.pushes++
	q.maximum = max(q.maximum, q.length)
	return nil
}
func (q *PdrQueue) Pop() *PdrObligation {
	if q.head == nil {
		return nil
	}
	o := q.head
	q.head = o.link
	o.link = nil
	o.owner = nil
	q.length--
	return o
}

// Clear releases queue ownership while preserving externally retained witness
// chains and cumulative queue statistics. It never recursively frees chains.
func (q *PdrQueue) Clear() {
	for q.Pop() != nil {
	}
}
func (q *PdrQueue) Check() error {
	slow, fast := q.head, q.head
	for fast != nil && fast.link != nil {
		slow = slow.link
		fast = fast.link.link
		if slow == fast {
			return ErrInvariant
		}
	}
	count := 0
	var previous *PdrObligation
	for o := q.head; o != nil; o = o.link {
		if o.owner != q || o.state == nil || o.frame < 0 {
			return ErrInvariant
		}
		if previous != nil && (previous.frame > o.frame || (previous.frame == o.frame && previous.priority > o.priority)) {
			return ErrInvariant
		}
		previous = o
		count++
	}
	if count != q.length || q.maximum < q.length || q.pushes < uint64(count) {
		return ErrInvariant
	}
	return nil
}
