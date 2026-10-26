package abc

// Dynamically maintained AIG order from aig/aig/aigOrder.c.

import "slices"

type AigOrderState struct {
	Previous, Next AigID
	Before, Total  int
}

// AigOrder is an owned circular order with object zero as its sentinel. It is
// independent of AigManager so graph-editing passes can stage and discard an
// ordering transaction without installing auxiliary manager state.
type AigOrder struct {
	predecessor, successor []AigID
	present                []bool
	previous, next         AigID
	before, total          int
}

// BuildAigOrder implements Aig_ManOrderStart. Native order storage excludes
// buffers and initially contains all ANDs in ascending object order.
func BuildAigOrder(g *AigManager, o AigTraversalOptions) (*AigOrder, error) {
	checked, err := checkAigTraversal(g, o)
	if err != nil {
		return nil, err
	}
	order := &AigOrder{
		predecessor: make([]AigID, len(g.objects)),
		successor:   make([]AigID, len(g.objects)),
		present:     make([]bool, len(g.objects)),
	}
	order.present[0] = true
	visit := aigTraversalVisit(checked)
	for id, object := range g.objects {
		if err = visit(); err != nil {
			return nil, err
		}
		if object.Type == AigBuffer {
			return nil, ErrInvariant
		}
		if object.Type != AigAND {
			continue
		}
		if err = order.Insert(g, AigID(id)); err != nil {
			return nil, err
		}
	}
	return order, nil
}

func (o *AigOrder) Clone() *AigOrder {
	if o == nil {
		return nil
	}
	return &AigOrder{
		predecessor: slices.Clone(o.predecessor),
		successor:   slices.Clone(o.successor),
		present:     slices.Clone(o.present),
		previous:    o.previous,
		next:        o.next,
		before:      o.before,
		total:       o.total,
	}
}

func (o *AigOrder) State() AigOrderState {
	if o == nil {
		return AigOrderState{}
	}
	return AigOrderState{o.previous, o.next, o.before, o.total}
}

// Sequence returns all AND IDs from the sentinel in maintained order.
func (o *AigOrder) Sequence() ([]AigID, error) {
	if err := o.check(); err != nil {
		return nil, err
	}
	result := make([]AigID, 0, o.total)
	for id := o.successor[0]; id != 0; id = o.successor[id] {
		result = append(result, id)
		if len(result) > o.total {
			return nil, ErrInvariant
		}
	}
	if len(result) != o.total {
		return nil, ErrInvariant
	}
	return result, nil
}

func (o *AigOrder) check() error {
	if o == nil || len(o.predecessor) == 0 || len(o.predecessor) != len(o.successor) || len(o.present) != len(o.successor) || !o.present[0] || o.total < 0 || o.before < 0 || o.before > o.total || uint64(o.previous) >= uint64(len(o.present)) || uint64(o.next) >= uint64(len(o.present)) || !o.present[o.previous] || !o.present[o.next] {
		return ErrInvariant
	}
	count := 0
	for id := o.successor[0]; id != 0; id = o.successor[id] {
		if uint64(id) >= uint64(len(o.present)) || !o.present[id] || o.successor[o.predecessor[id]] != id || o.predecessor[o.successor[id]] != id {
			return ErrInvariant
		}
		count++
		if count > o.total {
			return ErrInvariant
		}
	}
	if count != o.total || o.predecessor[o.successor[0]] != 0 || o.successor[o.predecessor[0]] != 0 {
		return ErrInvariant
	}
	return nil
}

func (o *AigOrder) grow(size int) error {
	if size <= len(o.present) {
		return nil
	}
	if uint64(size) > uint64(MaxLiteralID)+1 {
		return ErrLimit
	}
	o.predecessor = append(o.predecessor, make([]AigID, size-len(o.predecessor))...)
	o.successor = append(o.successor, make([]AigID, size-len(o.successor))...)
	o.present = append(o.present, make([]bool, size-len(o.present))...)
	return nil
}

// Insert implements Aig_ObjOrderInsert by inserting id immediately before the
// current Next cursor. A newly appended graph object is accepted after growing
// private order storage.
func (o *AigOrder) Insert(g *AigManager, id AigID) error {
	if err := o.check(); err != nil {
		return err
	}
	if g == nil || id == 0 || uint64(id) >= uint64(len(g.objects)) || g.objects[id].Type != AigAND {
		return ErrInvalidLiteral
	}
	if err := o.grow(int(id) + 1); err != nil {
		return err
	}
	if o.present[id] {
		return ErrInvariant
	}
	prior := o.predecessor[o.next]
	o.predecessor[id] = prior
	o.successor[id] = o.next
	o.successor[prior] = id
	o.predecessor[o.next] = id
	o.present[id] = true
	o.total++
	return nil
}

// Remove implements Aig_ObjOrderRemove, including native cursor repair.
func (o *AigOrder) Remove(g *AigManager, id AigID) error {
	if err := o.check(); err != nil {
		return err
	}
	if g == nil || id == 0 || uint64(id) >= uint64(len(g.objects)) || g.objects[id].Type != AigAND {
		return ErrInvalidLiteral
	}
	if uint64(id) >= uint64(len(o.present)) || !o.present[id] {
		return ErrInvariant
	}
	prior, following := o.predecessor[id], o.successor[id]
	if o.previous == id {
		if o.before == 0 {
			return ErrInvariant
		}
		o.before--
		o.previous = prior
	}
	if o.next == id {
		o.next = following
	}
	o.successor[prior] = following
	o.predecessor[following] = prior
	o.predecessor[id], o.successor[id] = 0, 0
	o.present[id] = false
	o.total--
	return nil
}

// Begin initializes the cursor used by the native order iterator.
func (o *AigOrder) Begin() error {
	if err := o.check(); err != nil {
		return err
	}
	o.previous = 0
	o.next = o.successor[0]
	o.before = 0
	return nil
}

// Current returns the next node to process, or false at the sentinel.
func (o *AigOrder) Current() (AigID, bool) {
	if o == nil || o.next == 0 {
		return 0, false
	}
	return o.next, true
}

// Advance implements Aig_ObjOrderAdvance exactly: Next becomes Previous, while
// Next itself is left unchanged until the iterator increment calls Refresh.
func (o *AigOrder) Advance() error {
	if err := o.check(); err != nil {
		return err
	}
	if o.next == 0 || o.successor[o.previous] != o.next {
		return ErrInvariant
	}
	o.previous = o.next
	o.before++
	return nil
}

// Refresh implements the increment expression of Aig_ManForEachNodeInOrder.
// Keeping it distinct from Advance preserves insertion/removal behavior inside
// an iteration body.
func (o *AigOrder) Refresh() error {
	if err := o.check(); err != nil {
		return err
	}
	o.next = o.successor[o.previous]
	return nil
}

// Step performs the common no-edit Advance/Refresh pair.
func (o *AigOrder) Step() error {
	if err := o.Advance(); err != nil {
		return err
	}
	return o.Refresh()
}
