package abc

// Dynamic and mapped fanout indexing from aig/gia/giaFanout.c.

import "slices"

type GiaFanoutOptions struct {
	MaxObjects, MaxEdges int // zero: 1M, 3M
	Stop                 func() bool
}

type GiaFanoutEdge struct {
	Fanout GiaID
	Slot   uint8
}

// GiaDynamicFanouts is an owned mutable edge index. Fanout lists retain source
// insertion order; repeated target IDs in different fanin slots remain distinct.
type GiaDynamicFanouts struct {
	objects int
	edges   [][]GiaFanoutEdge
	count   int
	limit   int
}

func normalizeGiaFanoutOptions(o GiaFanoutOptions) (GiaFanoutOptions, error) {
	if o.MaxObjects < 0 || o.MaxEdges < 0 {
		return o, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxEdges == 0 {
		o.MaxEdges = 3 << 20
	}
	if o.Stop != nil && o.Stop() {
		return o, ErrCancelled
	}
	return o, nil
}

// DynamicFanouts implements Gia_ManFanoutStart without attaching storage to
// the graph. Native dynamic fanouts include the two structural slots only;
// extended mux-control edges belong to StaticFanouts.
func (m *GiaManager) DynamicFanouts(o GiaFanoutOptions) (*GiaDynamicFanouts, error) {
	o, err := normalizeGiaFanoutOptions(o)
	if err != nil {
		return nil, err
	}
	if m == nil || len(m.objects) > o.MaxObjects {
		if m == nil {
			return nil, ErrDimension
		}
		return nil, ErrLimit
	}
	if err = m.Check(); err != nil {
		return nil, err
	}
	r := &GiaDynamicFanouts{objects: len(m.objects), edges: make([][]GiaFanoutEdge, len(m.objects)), limit: o.MaxEdges}
	for id, object := range m.objects {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		target := GiaID(id)
		if object.IsAnd() {
			first, err := object.Fanin0(target)
			if err != nil {
				return nil, err
			}
			if err = r.add(first.ID(), target, 0); err != nil {
				return nil, err
			}
			if !object.IsBuffer() {
				second, err := object.Fanin1(target)
				if err != nil {
					return nil, err
				}
				if err = r.add(second.ID(), target, 1); err != nil {
					return nil, err
				}
			}
		} else if object.IsCO() {
			first, err := object.Fanin0(target)
			if err != nil {
				return nil, err
			}
			if err = r.add(first.ID(), target, 0); err != nil {
				return nil, err
			}
		}
	}
	return r, nil
}

func (f *GiaDynamicFanouts) add(source, fanout GiaID, slot uint8) error {
	if f == nil || uint64(source) >= uint64(f.objects) || uint64(fanout) >= uint64(f.objects) || fanout == 0 || slot > 1 {
		return ErrDimension
	}
	for _, edge := range f.edges[source] {
		if edge.Fanout == fanout && edge.Slot == slot {
			return ErrInvariant
		}
	}
	if f.count == f.limit {
		return ErrLimit
	}
	f.edges[source] = append(f.edges[source], GiaFanoutEdge{Fanout: fanout, Slot: slot})
	f.count++
	return nil
}

// Add validates the requested slot against the current graph before appending
// it at the tail of the source's circular-list-equivalent order.
func (f *GiaDynamicFanouts) Add(m *GiaManager, source, fanout GiaID, slot uint8) error {
	if f == nil || m == nil || len(m.objects) != f.objects || uint64(fanout) >= uint64(len(m.objects)) || slot > 1 {
		return ErrDimension
	}
	object := m.objects[fanout]
	var literal GiaLit
	var err error
	if slot == 0 && (object.IsAnd() || object.IsCO()) {
		literal, err = object.Fanin0(fanout)
	} else if slot == 1 && object.IsAnd() && !object.IsBuffer() {
		literal, err = object.Fanin1(fanout)
	} else {
		return ErrDimension
	}
	if err != nil || literal.ID() != source {
		return ErrDimension
	}
	return f.add(source, fanout, slot)
}

func (f *GiaDynamicFanouts) Remove(source, fanout GiaID, slot uint8) error {
	if f == nil || uint64(source) >= uint64(f.objects) || slot > 1 {
		return ErrDimension
	}
	for i, edge := range f.edges[source] {
		if edge.Fanout == fanout && edge.Slot == slot {
			copy(f.edges[source][i:], f.edges[source][i+1:])
			f.edges[source] = f.edges[source][:len(f.edges[source])-1]
			f.count--
			return nil
		}
	}
	return ErrInvalidLiteral
}

func (f *GiaDynamicFanouts) Fanouts(id GiaID) ([]GiaFanoutEdge, error) {
	if f == nil || uint64(id) >= uint64(f.objects) {
		return nil, ErrInvalidLiteral
	}
	return slices.Clone(f.edges[id]), nil
}

func (f *GiaDynamicFanouts) EdgeCount() int {
	if f == nil {
		return 0
	}
	return f.count
}

// StaticMappingFanouts implements Gia_ManStaticMappingFanoutStart. LUT edges
// precede CO-driver edges, roots and COs are visited in increasing native order,
// and duplicate LUT fanins produce duplicate targets.
func (m *GiaManager) StaticMappingFanouts(mapping *GiaLUTMapping, o GiaFanoutOptions) (*GiaFanouts, error) {
	o, err := normalizeGiaFanoutOptions(o)
	if err != nil {
		return nil, err
	}
	if m == nil || mapping == nil {
		return nil, ErrDimension
	}
	if len(m.objects) > o.MaxObjects {
		return nil, ErrLimit
	}
	if err = mapping.checkGraph(m); err != nil {
		return nil, err
	}
	counts := make([]int, len(m.objects))
	edges := 0
	charge := func(source GiaID) error {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		if uint64(source) >= uint64(len(counts)) {
			return ErrInvalidLiteral
		}
		if edges == o.MaxEdges {
			return ErrLimit
		}
		counts[source]++
		edges++
		return nil
	}
	for _, root := range mapping.Roots() {
		record, _ := mapping.Record(root)
		for _, source := range record.Fanins {
			if err = charge(source); err != nil {
				return nil, err
			}
		}
	}
	for _, co := range m.cos {
		literal, err := m.objects[co].Fanin0(co)
		if err != nil {
			return nil, err
		}
		if err = charge(literal.ID()); err != nil {
			return nil, err
		}
	}
	offsets := make([]int, len(counts)+1)
	for i, count := range counts {
		offsets[i+1] = offsets[i] + count
	}
	targets := make([]GiaID, edges)
	positions := slices.Clone(offsets[:len(counts)])
	appendTarget := func(source, target GiaID) {
		targets[positions[source]] = target
		positions[source]++
	}
	for _, root := range mapping.Roots() {
		record, _ := mapping.Record(root)
		for _, source := range record.Fanins {
			appendTarget(source, root)
		}
	}
	for _, co := range m.cos {
		literal, _ := m.objects[co].Fanin0(co)
		appendTarget(literal.ID(), co)
	}
	return &GiaFanouts{offsets: offsets, targets: targets}, nil
}
