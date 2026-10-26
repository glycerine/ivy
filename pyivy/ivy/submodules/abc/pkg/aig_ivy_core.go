package abc

// Derived from ABC src/aig/ivy/{ivy.h,ivyMan.c,ivyMem.c,ivyObj.c,ivyOper.c,ivyCheck.c,ivyDfs.c}.

import (
	"fmt"
	"slices"
)

// IvyType preserves the object categories of Ivy_Type_t. Constant one is
// object zero and has IvyPI type, as in the source representation.
type IvyType uint8

const (
	IvyNone IvyType = iota
	IvyPI
	IvyPO
	IvyAssert
	IvyLatch
	IvyAND
	IvyEXOR
	IvyBuffer
	IvyVoid
)

// IvyInit preserves the four-valued latch initialization convention.
type IvyInit uint8

const (
	IvyInitNone IvyInit = iota
	IvyInit0
	IvyInit1
	IvyInitDC
)

// IvyEdge is an Ivy object ID with a low complement bit. Edge zero is
// constant one and edge one is constant zero.
type IvyEdge uint32

func MakeIvyEdge(id uint32, complemented bool) (IvyEdge, error) {
	if id > MaxLiteralID {
		return 0, ErrInvalidLiteral
	}
	edge := IvyEdge(id << 1)
	if complemented {
		edge |= 1
	}
	return edge, nil
}

func (e IvyEdge) ID() uint32         { return uint32(e >> 1) }
func (e IvyEdge) Complemented() bool { return e&1 != 0 }
func (e IvyEdge) Not() IvyEdge       { return e ^ 1 }
func (e IvyEdge) Regular() IvyEdge   { return e &^ 1 }
func (e IvyEdge) NotIf(c bool) IvyEdge {
	if c {
		return e ^ 1
	}
	return e
}

const (
	IvyConst1 IvyEdge = 0
	IvyConst0 IvyEdge = 1
)

// IvyObject is an owned value snapshot. FaninCount distinguishes constant/PI
// objects from unary objects whose fanin may be constant one.
type IvyObject struct {
	Type           IvyType
	Init           IvyInit
	Fanin0, Fanin1 IvyEdge
	FaninCount     uint8
	Level          uint32
	Refs           uint32
	Phase          bool
	Alive          bool
}

type ivyHashKey struct {
	type_ IvyType
	init  IvyInit
	a, b  IvyEdge
}

// IvyOptions bounds allocation and optional work. MaxObjects zero uses the
// literal representation limit. Cancelled is polled only at transactional
// entry points and in traversals; independent managers share no state.
type IvyOptions struct {
	MaxObjects int
	MaxWork    int
	Cancelled  func() bool
}

// IvyManager is an owned, structurally hashed graph. Stable IDs are never
// reused. Map iteration is never observable; object, fanout and traversal
// order is defined by the slices of stable IDs.
type IvyManager struct {
	objects []IvyObject
	inputs  []uint32
	outputs []uint32
	hash    map[ivyHashKey]uint32
	options IvyOptions
	created int
	deleted int
}

type ivyManagerSnapshot struct {
	objects []IvyObject
	inputs  []uint32
	outputs []uint32
	hash    map[ivyHashKey]uint32
	created int
	deleted int
}

func (m *IvyManager) snapshot() ivyManagerSnapshot {
	s := ivyManagerSnapshot{objects: slices.Clone(m.objects), inputs: slices.Clone(m.inputs), outputs: slices.Clone(m.outputs), hash: make(map[ivyHashKey]uint32, len(m.hash)), created: m.created, deleted: m.deleted}
	for k, v := range m.hash {
		s.hash[k] = v
	}
	return s
}

func (m *IvyManager) restore(s ivyManagerSnapshot) {
	m.objects, m.inputs, m.outputs = s.objects, s.inputs, s.outputs
	m.hash, m.created, m.deleted = s.hash, s.created, s.deleted
}

func (m *IvyManager) atomicEdge(build func() (IvyEdge, error)) (IvyEdge, error) {
	if err := m.ready(); err != nil {
		return 0, err
	}
	s := m.snapshot()
	result, err := build()
	if err != nil {
		m.restore(s)
	}
	return result, err
}

func NewIvyManager(options IvyOptions) (*IvyManager, error) {
	if options.MaxObjects < 0 || options.MaxWork < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = int(MaxLiteralID) + 1
	}
	if options.MaxObjects < 1 {
		return nil, ErrLimit
	}
	m := &IvyManager{options: options, hash: make(map[ivyHashKey]uint32)}
	m.objects = append(m.objects, IvyObject{Type: IvyPI, Alive: true, Phase: true})
	m.created = 1
	return m, nil
}

func (m *IvyManager) ready() error {
	if m == nil || len(m.objects) == 0 || m.hash == nil {
		return fmt.Errorf("%w: uninitialized Ivy manager", ErrInvariant)
	}
	if m.options.Cancelled != nil && m.options.Cancelled() {
		return ErrCancelled
	}
	return nil
}

func (m *IvyManager) room(extra int) error {
	if err := m.ready(); err != nil {
		return err
	}
	if extra < 0 || len(m.objects) > m.options.MaxObjects-extra || uint64(len(m.objects)+extra) > uint64(MaxLiteralID)+1 {
		return ErrLimit
	}
	return nil
}

func (m *IvyManager) validEdge(e IvyEdge) bool {
	id := e.ID()
	return uint64(id) < uint64(len(m.objects)) && m.objects[id].Alive && m.objects[id].Type != IvyPO && m.objects[id].Type != IvyAssert
}

func (m *IvyManager) phase(e IvyEdge) bool { return m.objects[e.ID()].Phase != e.Complemented() }

func (m *IvyManager) Object(id uint32) (IvyObject, error) {
	if m == nil || uint64(id) >= uint64(len(m.objects)) || !m.objects[id].Alive {
		return IvyObject{}, ErrInvalidLiteral
	}
	return m.objects[id], nil
}

func (m *IvyManager) Objects() []uint32 {
	if m == nil {
		return nil
	}
	ids := make([]uint32, 0, len(m.objects)-m.deleted)
	for id := range m.objects {
		if m.objects[id].Alive {
			ids = append(ids, uint32(id))
		}
	}
	return ids
}
func (m *IvyManager) Inputs() []uint32  { return slices.Clone(m.inputs) }
func (m *IvyManager) Outputs() []uint32 { return slices.Clone(m.outputs) }
func (m *IvyManager) ObjectIDMax() int {
	if m == nil {
		return -1
	}
	return len(m.objects) - 1
}
func (m *IvyManager) LiveCount() int {
	if m == nil {
		return 0
	}
	return m.created - m.deleted
}

func (m *IvyManager) Count(t IvyType) int {
	count := 0
	if m != nil {
		for _, obj := range m.objects {
			if obj.Alive && obj.Type == t {
				count++
			}
		}
	}
	if t == IvyPI && count > 0 {
		count--
	} // object zero is the distinguished constant
	return count
}

func (m *IvyManager) NodeCount() int { return m.Count(IvyAND) + m.Count(IvyEXOR) }
func (m *IvyManager) Cost() int      { return m.Count(IvyAND) + 3*m.Count(IvyEXOR) + 8*m.Count(IvyLatch) }

func (m *IvyManager) appendObject(obj IvyObject) uint32 {
	id := uint32(len(m.objects))
	obj.Alive = true
	m.objects = append(m.objects, obj)
	m.created++
	return id
}

func (m *IvyManager) CreatePI() (IvyEdge, error) {
	if err := m.room(1); err != nil {
		return 0, err
	}
	id := m.appendObject(IvyObject{Type: IvyPI})
	m.inputs = append(m.inputs, id)
	return IvyEdge(id << 1), nil
}

func (m *IvyManager) CreatePO(driver IvyEdge) (uint32, error) {
	if !m.validEdge(driver) {
		return 0, ErrInvalidLiteral
	}
	if err := m.room(1); err != nil {
		return 0, err
	}
	if m.objects[driver.ID()].Refs == ^uint32(0) {
		return 0, ErrLimit
	}
	id := m.appendObject(IvyObject{Type: IvyPO, Fanin0: driver, FaninCount: 1, Level: m.objects[driver.ID()].Level, Phase: m.phase(driver)})
	m.objects[driver.ID()].Refs++
	m.outputs = append(m.outputs, id)
	return id, nil
}

func ivyOrdered(a, b IvyEdge) (IvyEdge, IvyEdge) {
	if a.ID() > b.ID() {
		return b, a
	}
	return a, b
}

func (m *IvyManager) createHashed(t IvyType, init IvyInit, a, b IvyEdge, fanins uint8) (IvyEdge, error) {
	if !m.validEdge(a) || fanins == 2 && !m.validEdge(b) {
		return 0, ErrInvalidLiteral
	}
	if fanins == 2 {
		a, b = ivyOrdered(a, b)
	}
	key := ivyHashKey{type_: t, init: init, a: a, b: b}
	if id, ok := m.hash[key]; ok && m.objects[id].Alive {
		return IvyEdge(id << 1), nil
	}
	if err := m.room(1); err != nil {
		return 0, err
	}
	if m.objects[a.ID()].Refs == ^uint32(0) || fanins == 2 && m.objects[b.ID()].Refs == ^uint32(0) {
		return 0, ErrLimit
	}
	level := uint32(0)
	phase := m.phase(a)
	if fanins == 2 {
		level = m.objects[a.ID()].Level
		if m.objects[b.ID()].Level > level {
			level = m.objects[b.ID()].Level
		}
		level++
		if t == IvyEXOR {
			level++
		}
		if t == IvyEXOR {
			phase = m.phase(a) != m.phase(b)
		} else {
			phase = m.phase(a) && m.phase(b)
		}
	}
	id := m.appendObject(IvyObject{Type: t, Init: init, Fanin0: a, Fanin1: b, FaninCount: fanins, Level: level, Phase: phase})
	m.objects[a.ID()].Refs++
	if fanins == 2 {
		m.objects[b.ID()].Refs++
	}
	m.hash[key] = id
	return IvyEdge(id << 1), nil
}

// And implements Ivy_And including its constant and complement reductions.
func (m *IvyManager) And(a, b IvyEdge) (IvyEdge, error) {
	if err := m.ready(); err != nil {
		return 0, err
	}
	if !m.validEdge(a) || !m.validEdge(b) {
		return 0, ErrInvalidLiteral
	}
	if a == b {
		return a, nil
	}
	if a == b.Not() {
		return IvyConst0, nil
	}
	if a.Regular() == IvyConst1 {
		if a == IvyConst1 {
			return b, nil
		}
		return IvyConst0, nil
	}
	if b.Regular() == IvyConst1 {
		if b == IvyConst1 {
			return a, nil
		}
		return IvyConst0, nil
	}
	return m.createHashed(IvyAND, IvyInitNone, a, b, 2)
}

func (m *IvyManager) Or(a, b IvyEdge) (IvyEdge, error) {
	r, err := m.And(a.Not(), b.Not())
	return r.Not(), err
}

func (m *IvyManager) lookupAnd(a, b IvyEdge) (IvyEdge, bool) {
	if !m.validEdge(a) || !m.validEdge(b) || a.Regular() == IvyConst1 || b.Regular() == IvyConst1 || a.ID() == b.ID() {
		return 0, false
	}
	a, b = ivyOrdered(a, b)
	id, ok := m.hash[ivyHashKey{type_: IvyAND, a: a, b: b}]
	return IvyEdge(id << 1), ok && m.objects[id].Alive
}

// Xor follows the active Ivy_Exor source path, which lowers XOR to AND/OR.
func (m *IvyManager) Xor(a, b IvyEdge) (IvyEdge, error) {
	return m.atomicEdge(func() (IvyEdge, error) {
		x, err := m.And(a, b.Not())
		if err != nil {
			return 0, err
		}
		y, err := m.And(a.Not(), b)
		if err != nil {
			return 0, err
		}
		return m.Or(x, y)
	})
}

func (m *IvyManager) Mux(control, whenTrue, whenFalse IvyEdge) (IvyEdge, error) {
	return m.atomicEdge(func() (IvyEdge, error) {
		if whenFalse == whenTrue.Not() {
			return m.Xor(control, whenFalse)
		}
		a1, hasA1 := m.lookupAnd(control, whenTrue)
		a2, hasA2 := m.lookupAnd(control.Not(), whenFalse)
		if hasA1 && hasA2 {
			if top, ok := m.lookupAnd(a1.Not(), a2.Not()); ok {
				return top.Not(), nil
			}
		}
		b1, hasB1 := m.lookupAnd(control, whenTrue.Not())
		b2, hasB2 := m.lookupAnd(control.Not(), whenFalse.Not())
		if hasB1 && hasB2 {
			if top, ok := m.lookupAnd(b1.Not(), b2.Not()); ok {
				return top, nil
			}
		}
		if ivyBoolInt(hasA1)+ivyBoolInt(hasA2) >= ivyBoolInt(hasB1)+ivyBoolInt(hasB2) {
			var err error
			if !hasA1 {
				a1, err = m.And(control, whenTrue)
				if err != nil {
					return 0, err
				}
			}
			if !hasA2 {
				a2, err = m.And(control.Not(), whenFalse)
				if err != nil {
					return 0, err
				}
			}
			return m.Or(a1, a2)
		}
		var err error
		if !hasB1 {
			b1, err = m.And(control, whenTrue.Not())
			if err != nil {
				return 0, err
			}
		}
		if !hasB2 {
			b2, err = m.And(control.Not(), whenFalse.Not())
			if err != nil {
				return 0, err
			}
		}
		result, err := m.Or(b1, b2)
		return result.Not(), err
	})
}

func ivyBoolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (m *IvyManager) Majority(a, b, c IvyEdge) (IvyEdge, error) {
	return m.atomicEdge(func() (IvyEdge, error) {
		ab, err := m.And(a, b)
		if err != nil {
			return 0, err
		}
		ac, err := m.And(a, c)
		if err != nil {
			return 0, err
		}
		bc, err := m.And(b, c)
		if err != nil {
			return 0, err
		}
		x, err := m.Or(ab, ac)
		if err != nil {
			return 0, err
		}
		return m.Or(x, bc)
	})
}

func (m *IvyManager) Multi(args []IvyEdge, t IvyType) (IvyEdge, error) {
	if len(args) == 0 || t != IvyAND && t != IvyEXOR {
		return 0, ErrDimension
	}
	if m.options.MaxWork > 0 && len(args)-1 > m.options.MaxWork {
		return 0, ErrLimit
	}
	for _, e := range args {
		if !m.validEdge(e) {
			return 0, ErrInvalidLiteral
		}
	}
	var build func(int, int) (IvyEdge, error)
	build = func(lo, hi int) (IvyEdge, error) {
		if hi-lo == 1 {
			return args[lo], nil
		}
		mid := lo + (hi-lo)/2
		a, err := build(lo, mid)
		if err != nil {
			return 0, err
		}
		b, err := build(mid, hi)
		if err != nil {
			return 0, err
		}
		if t == IvyAND {
			return m.And(a, b)
		}
		return m.Xor(a, b)
	}
	return m.atomicEdge(func() (IvyEdge, error) { return build(0, len(args)) })
}

func (m *IvyManager) Miter(pairs [][2]IvyEdge) (IvyEdge, error) {
	if len(pairs) == 0 {
		return 0, ErrDimension
	}
	return m.atomicEdge(func() (IvyEdge, error) {
		equal := make([]IvyEdge, len(pairs))
		for i, p := range pairs {
			d, err := m.Xor(p[0], p[1])
			if err != nil {
				return 0, err
			}
			equal[i] = d.Not()
		}
		all, err := m.Multi(equal, IvyAND)
		return all.Not(), err
	})
}

// AnyDifference implements the source Ivy_Miter: OR of all pairwise XORs.
func (m *IvyManager) AnyDifference(pairs [][2]IvyEdge) (IvyEdge, error) {
	return m.Miter(pairs)
}

func (m *IvyManager) Latch(driver IvyEdge, init IvyInit) (IvyEdge, error) {
	if init < IvyInit0 || init > IvyInitDC || driver.Complemented() {
		return 0, ErrDimension
	}
	return m.createHashed(IvyLatch, init, driver, 0, 1)
}

// DFS returns reachable logic and buffer IDs in fanin-before-fanout order.
func (m *IvyManager) DFS() ([]uint32, error) { nodes, _, err := m.SequentialDFS(); return nodes, err }

func (m *IvyManager) SequentialDFS() ([]uint32, []uint32, error) {
	if err := m.ready(); err != nil {
		return nil, nil, err
	}
	state := make([]uint8, len(m.objects))
	nodes := make([]uint32, 0, m.NodeCount())
	latches := make([]uint32, 0, m.Count(IvyLatch))
	work := 0
	type frame struct {
		id   uint32
		exit bool
	}
	visit := func(root uint32) error {
		stack := []frame{{id: root}}
		for len(stack) > 0 {
			last := len(stack) - 1
			f := stack[last]
			stack = stack[:last]
			if uint64(f.id) >= uint64(len(m.objects)) || !m.objects[f.id].Alive {
				return fmt.Errorf("%w: traversal object %d", ErrInvariant, f.id)
			}
			if f.exit {
				state[f.id] = 2
				obj := m.objects[f.id]
				if obj.Type == IvyAND || obj.Type == IvyEXOR || obj.Type == IvyBuffer {
					nodes = append(nodes, f.id)
				}
				work++
				if m.options.MaxWork > 0 && work > m.options.MaxWork {
					return ErrLimit
				}
				if m.options.Cancelled != nil && m.options.Cancelled() {
					return ErrCancelled
				}
				continue
			}
			if state[f.id] == 2 {
				continue
			}
			if state[f.id] == 1 {
				return fmt.Errorf("%w: combinational cycle at %d", ErrInvariant, f.id)
			}
			state[f.id] = 1
			obj := m.objects[f.id]
			if obj.Type == IvyPI || obj.Type == IvyLatch {
				stack = append(stack, frame{id: f.id, exit: true})
				continue
			}
			if obj.Type != IvyAND && obj.Type != IvyEXOR && obj.Type != IvyBuffer {
				return fmt.Errorf("%w: object %d in logic cone", ErrInvariant, f.id)
			}
			stack = append(stack, frame{id: f.id, exit: true})
			if obj.FaninCount == 2 {
				stack = append(stack, frame{id: obj.Fanin1.ID()})
			}
			if obj.FaninCount >= 1 {
				stack = append(stack, frame{id: obj.Fanin0.ID()})
			}
		}
		return nil
	}
	for id, obj := range m.objects {
		if obj.Alive && obj.Type == IvyLatch {
			latches = append(latches, uint32(id))
		}
	}
	for _, id := range m.outputs {
		if err := visit(m.objects[id].Fanin0.ID()); err != nil {
			return nil, nil, err
		}
	}
	for _, id := range latches {
		if err := visit(m.objects[id].Fanin0.ID()); err != nil {
			return nil, nil, err
		}
	}
	return nodes, latches, nil
}

// CollectCone returns root's logic cone in fanin-before-fanout order while
// treating frontier edges as leaves. The returned IDs are owned.
func (m *IvyManager) CollectCone(root IvyEdge, frontier []IvyEdge) ([]uint32, error) {
	if err := m.ready(); err != nil {
		return nil, err
	}
	if !m.validEdge(root) {
		return nil, ErrInvalidLiteral
	}
	stop := make([]bool, len(m.objects))
	for _, e := range frontier {
		if !m.validEdge(e) {
			return nil, ErrInvalidLiteral
		}
		stop[e.ID()] = true
	}
	state := make([]uint8, len(m.objects))
	result := make([]uint32, 0)
	type frame struct {
		id   uint32
		exit bool
	}
	stack := []frame{{id: root.ID()}}
	work := 0
	for len(stack) > 0 {
		last := len(stack) - 1
		f := stack[last]
		stack = stack[:last]
		if f.exit {
			state[f.id] = 2
			result = append(result, f.id)
			continue
		}
		if stop[f.id] || m.objects[f.id].Type == IvyPI || m.objects[f.id].Type == IvyLatch {
			continue
		}
		if state[f.id] == 2 {
			continue
		}
		if state[f.id] == 1 {
			return nil, ErrInvariant
		}
		o := m.objects[f.id]
		if o.Type != IvyAND && o.Type != IvyEXOR && o.Type != IvyBuffer {
			return nil, ErrDimension
		}
		state[f.id] = 1
		stack = append(stack, frame{id: f.id, exit: true})
		if o.FaninCount == 2 {
			stack = append(stack, frame{id: o.Fanin1.ID()})
		}
		stack = append(stack, frame{id: o.Fanin0.ID()})
		work++
		if m.options.MaxWork > 0 && work > m.options.MaxWork {
			return nil, ErrLimit
		}
		if m.options.Cancelled != nil && m.options.Cancelled() {
			return nil, ErrCancelled
		}
	}
	return result, nil
}

func (m *IvyManager) Fanouts(id uint32) ([]uint32, error) {
	if _, err := m.Object(id); err != nil {
		return nil, err
	}
	result := make([]uint32, 0, m.objects[id].Refs)
	for oid, obj := range m.objects {
		if !obj.Alive {
			continue
		}
		if obj.FaninCount >= 1 && obj.Fanin0.ID() == id || obj.FaninCount == 2 && obj.Fanin1.ID() == id {
			result = append(result, uint32(oid))
		}
	}
	return result, nil
}

func (m *IvyManager) Levelize() ([][]uint32, error) {
	if err := m.Check(); err != nil {
		return nil, err
	}
	max := uint32(0)
	for _, o := range m.objects {
		if o.Alive && o.Level > max {
			max = o.Level
		}
	}
	levels := make([][]uint32, int(max)+1)
	for id, o := range m.objects {
		if o.Alive && (o.Type == IvyAND || o.Type == IvyEXOR) {
			levels[o.Level] = append(levels[o.Level], uint32(id))
		}
	}
	return levels, nil
}

func (m *IvyManager) RequiredLevels() ([]uint32, error) {
	if err := m.Check(); err != nil {
		return nil, err
	}
	distance := make([]uint32, len(m.objects))
	max := uint32(0)
	for _, id := range m.outputs {
		l := m.objects[m.objects[id].Fanin0.ID()].Level
		if l > max {
			max = l
		}
	}
	for id := len(m.objects) - 1; id >= 0; id-- {
		o := m.objects[id]
		if !o.Alive || (o.Type != IvyAND && o.Type != IvyEXOR) {
			continue
		}
		step := uint32(1)
		if o.Type == IvyEXOR {
			step = 2
		}
		v := distance[id] + step
		if v > distance[o.Fanin0.ID()] {
			distance[o.Fanin0.ID()] = v
		}
		if o.FaninCount == 2 && v > distance[o.Fanin1.ID()] {
			distance[o.Fanin1.ID()] = v
		}
	}
	for id, o := range m.objects {
		if o.Alive {
			if distance[id] > max {
				return nil, ErrInvariant
			}
			distance[id] = max - distance[id]
		}
	}
	return distance, nil
}

func (m *IvyManager) deleteObject(id uint32) {
	o := m.objects[id]
	if o.FaninCount >= 1 {
		m.objects[o.Fanin0.ID()].Refs--
	}
	if o.FaninCount == 2 {
		m.objects[o.Fanin1.ID()].Refs--
	}
	if o.Type == IvyAND || o.Type == IvyEXOR || o.Type == IvyLatch {
		delete(m.hash, ivyHashKey{type_: o.Type, init: o.Init, a: o.Fanin0, b: o.Fanin1})
	}
	m.objects[id] = IvyObject{}
	m.deleted++
}

// Cleanup removes unreferenced nodes, latches and buffers in reverse ID order.
func (m *IvyManager) Cleanup() (int, error) {
	if err := m.ready(); err != nil {
		return 0, err
	}
	removed := 0
	for id := len(m.objects) - 1; id > 0; id-- {
		o := m.objects[id]
		if o.Alive && o.Refs == 0 && (o.Type == IvyAND || o.Type == IvyEXOR || o.Type == IvyLatch || o.Type == IvyBuffer) {
			m.deleteObject(uint32(id))
			removed++
		}
	}
	return removed, nil
}

func (m *IvyManager) Clone() (*IvyManager, error) {
	if err := m.Check(); err != nil {
		return nil, err
	}
	n := &IvyManager{objects: slices.Clone(m.objects), inputs: slices.Clone(m.inputs), outputs: slices.Clone(m.outputs), hash: make(map[ivyHashKey]uint32, len(m.hash)), options: m.options, created: m.created, deleted: m.deleted}
	for k, v := range m.hash {
		n.hash[k] = v
	}
	return n, nil
}

func (m *IvyManager) Check() error {
	if err := m.ready(); err != nil {
		return err
	}
	found := make([]uint32, len(m.objects))
	liveHash := 0
	for id, o := range m.objects {
		if !o.Alive {
			continue
		}
		if id == 0 {
			if o.Type != IvyPI || o.FaninCount != 0 || !o.Phase {
				return fmt.Errorf("%w: malformed constant", ErrInvariant)
			}
			continue
		}
		if o.Type <= IvyNone || o.Type >= IvyVoid {
			return fmt.Errorf("%w: object %d type", ErrInvariant, id)
		}
		want := uint8(2)
		if o.Type == IvyPI {
			want = 0
		} else if o.Type == IvyPO || o.Type == IvyAssert || o.Type == IvyLatch || o.Type == IvyBuffer {
			want = 1
		}
		if o.FaninCount != want {
			return fmt.Errorf("%w: object %d fanin count", ErrInvariant, id)
		}
		if want >= 1 {
			if !m.validEdge(o.Fanin0) {
				return fmt.Errorf("%w: object %d first fanin", ErrInvariant, id)
			}
			found[o.Fanin0.ID()]++
		}
		if want == 2 {
			if !m.validEdge(o.Fanin1) || o.Fanin0.ID() >= o.Fanin1.ID() {
				return fmt.Errorf("%w: object %d second fanin/order", ErrInvariant, id)
			}
			found[o.Fanin1.ID()]++
			level := m.objects[o.Fanin0.ID()].Level
			if m.objects[o.Fanin1.ID()].Level > level {
				level = m.objects[o.Fanin1.ID()].Level
			}
			level++
			if o.Type == IvyEXOR {
				level++
			}
			if o.Level != level {
				return fmt.Errorf("%w: object %d level", ErrInvariant, id)
			}
		}
		if o.Type == IvyLatch && o.Init == IvyInitNone {
			return fmt.Errorf("%w: latch %d init", ErrInvariant, id)
		}
		if o.Type == IvyAND || o.Type == IvyEXOR || o.Type == IvyLatch {
			k := ivyHashKey{type_: o.Type, init: o.Init, a: o.Fanin0, b: o.Fanin1}
			if m.hash[k] != uint32(id) {
				return fmt.Errorf("%w: object %d hash", ErrInvariant, id)
			}
			liveHash++
		}
	}
	for id, o := range m.objects {
		if o.Alive && o.Refs != found[id] {
			return fmt.Errorf("%w: object %d refs %d want %d", ErrInvariant, id, o.Refs, found[id])
		}
	}
	if liveHash != len(m.hash) {
		return fmt.Errorf("%w: stale hash entries", ErrInvariant)
	}
	for _, id := range m.inputs {
		if uint64(id) >= uint64(len(m.objects)) || !m.objects[id].Alive || m.objects[id].Type != IvyPI {
			return fmt.Errorf("%w: input list", ErrInvariant)
		}
	}
	for _, id := range m.outputs {
		if uint64(id) >= uint64(len(m.objects)) || !m.objects[id].Alive || m.objects[id].Type != IvyPO {
			return fmt.Errorf("%w: output list", ErrInvariant)
		}
	}
	_, _, err := m.SequentialDFS()
	return err
}

// Evaluate computes every primary output for one combinational assignment.
// Latches are accepted through latchValues in stable latch-ID order.
func (m *IvyManager) Evaluate(piValues, latchValues []bool) ([]bool, error) {
	if len(piValues) != len(m.inputs) || len(latchValues) != m.Count(IvyLatch) {
		return nil, ErrDimension
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	values := make([]bool, len(m.objects))
	values[0] = true
	for i, id := range m.inputs {
		values[id] = piValues[i]
	}
	li := 0
	for id, o := range m.objects {
		if !o.Alive || id == 0 {
			continue
		}
		val := func(e IvyEdge) bool {
			v := values[e.ID()]
			if e.Complemented() {
				v = !v
			}
			return v
		}
		switch o.Type {
		case IvyLatch:
			values[id] = latchValues[li]
			li++
		case IvyAND:
			values[id] = val(o.Fanin0) && val(o.Fanin1)
		case IvyEXOR:
			values[id] = val(o.Fanin0) != val(o.Fanin1)
		case IvyBuffer:
			values[id] = val(o.Fanin0)
		case IvyPO, IvyAssert:
			values[id] = val(o.Fanin0)
		}
	}
	result := make([]bool, len(m.outputs))
	for i, id := range m.outputs {
		result[i] = values[id]
	}
	return result, nil
}

// ToAigManager converts a combinational Ivy graph into the package's other
// concrete AIG representation. It is intentionally unavailable for latches
// and assertions because those require an external CI/CO partition contract.
func (m *IvyManager) ToAigManager() (*AigManager, error) {
	if err := m.Check(); err != nil {
		return nil, err
	}
	if m.Count(IvyLatch) != 0 || m.Count(IvyAssert) != 0 {
		return nil, ErrDimension
	}
	result := NewAigManager()
	mapping := make([]AigLit, len(m.objects))
	mapping[0] = AigConst1
	read := func(e IvyEdge) (AigLit, error) {
		if uint64(e.ID()) >= uint64(len(mapping)) || mapping[e.ID()] == InvalidAigLit {
			return InvalidAigLit, ErrInvariant
		}
		return mapping[e.ID()].NotIf(e.Complemented()), nil
	}
	for id, o := range m.objects {
		if id == 0 || !o.Alive {
			continue
		}
		mapping[id] = InvalidAigLit
		switch o.Type {
		case IvyPI:
			v, err := result.CreateCI()
			if err != nil {
				return nil, err
			}
			mapping[id] = v
		case IvyAND:
			a, err := read(o.Fanin0)
			if err != nil {
				return nil, err
			}
			b, err := read(o.Fanin1)
			if err != nil {
				return nil, err
			}
			v, err := result.And(a, b)
			if err != nil {
				return nil, err
			}
			mapping[id] = v
		case IvyEXOR:
			a, err := read(o.Fanin0)
			if err != nil {
				return nil, err
			}
			b, err := read(o.Fanin1)
			if err != nil {
				return nil, err
			}
			v, err := result.Xor(a, b)
			if err != nil {
				return nil, err
			}
			mapping[id] = v
		case IvyBuffer:
			v, err := read(o.Fanin0)
			if err != nil {
				return nil, err
			}
			v, err = result.CreateBuffer(v)
			if err != nil {
				return nil, err
			}
			mapping[id] = v
		case IvyPO:
			v, err := read(o.Fanin0)
			if err != nil {
				return nil, err
			}
			if _, err = result.CreateCO(v); err != nil {
				return nil, err
			}
		default:
			return nil, ErrDimension
		}
	}
	if err := result.Check(); err != nil {
		return nil, err
	}
	return result, nil
}

// IvyInitNotIf, IvyInitAnd and IvyInitXor are the checked four-valued latch
// initialization operations embedded in ivy.h.
func IvyInitNotIf(v IvyInit, complement bool) (IvyInit, error) {
	if v < IvyInit0 || v > IvyInitDC {
		return 0, ErrDimension
	}
	if !complement || v == IvyInitDC {
		return v, nil
	}
	if v == IvyInit0 {
		return IvyInit1, nil
	}
	return IvyInit0, nil
}
func IvyInitAnd(a, b IvyInit) (IvyInit, error) {
	if a < IvyInit0 || a > IvyInitDC || b < IvyInit0 || b > IvyInitDC {
		return 0, ErrDimension
	}
	if a == IvyInit0 || b == IvyInit0 {
		return IvyInit0, nil
	}
	if a == IvyInitDC || b == IvyInitDC {
		return IvyInitDC, nil
	}
	return IvyInit1, nil
}
func IvyInitXor(a, b IvyInit) (IvyInit, error) {
	if a < IvyInit0 || a > IvyInitDC || b < IvyInit0 || b > IvyInitDC {
		return 0, ErrDimension
	}
	if a == IvyInitDC || b == IvyInitDC {
		return IvyInitDC, nil
	}
	if a != b {
		return IvyInit1, nil
	}
	return IvyInit0, nil
}

func IvyLeaf(id uint32, latency uint8) (uint32, error) {
	if id > 0x00ffffff {
		return 0, ErrLimit
	}
	return id<<8 | uint32(latency), nil
}
func IvyLeafID(leaf uint32) uint32     { return leaf >> 8 }
func IvyLeafLatency(leaf uint32) uint8 { return uint8(leaf) }
