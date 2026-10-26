package abc

// Persistent private discovery state from giaSplit.c:Spl_ManAlloc/ComputeOne.

import "slices"

type giaWindowDiscoverySession struct {
	graph                      *GiaManager
	objects                    []GiaObject
	network                    *GiaLUTNetwork
	io, in, no, used, leafSeen []bool
	refs, baseRefs, seen       []int
	fanouts                    [][]GiaID
	builds, refreshes, queries int
}

// prepare stages initial state and mapped reference refreshes before commit.
// Scratch is reset on every query, including recovery after cancelled queries.
// Network values are immutable; a changed network pointer triggers validation
// and refreshed references. Graph changes invalidate this fixed-graph session.
func (s *giaWindowDiscoverySession) prepare(n *GiaLUTNetwork, m *GiaManager, o GiaLUTAnalysisOptions, charge func() error) error {
	size := len(m.objects)
	if size > o.MaxObjects {
		return ErrLimit
	}
	if s.graph == nil {
		if _, e := n.ValidateFunctions(m, o); e != nil {
			return e
		}
		candidate := &giaWindowDiscoverySession{graph: m, objects: slices.Clone(m.objects), io: make([]bool, size), in: make([]bool, size), no: make([]bool, size), used: make([]bool, size), leafSeen: make([]bool, size), refs: make([]int, size), baseRefs: make([]int, size), seen: make([]int, size), fanouts: make([][]GiaID, size)}
		candidate.io[0] = true
		for id, obj := range m.objects {
			if e := charge(); e != nil {
				return e
			}
			if obj.IsCI() || obj.IsCO() {
				candidate.io[id] = true
			}
			if _, mux := m.MuxControl(GiaID(id)); mux {
				return ErrUnsupported
			}
			if obj.IsAnd() || obj.IsCO() {
				a, _ := obj.Fanin0(GiaID(id))
				candidate.fanouts[a.ID()] = append(candidate.fanouts[a.ID()], GiaID(id))
			}
			if obj.IsAnd() && !obj.IsBuffer() {
				b, _ := obj.Fanin1(GiaID(id))
				candidate.fanouts[b.ID()] = append(candidate.fanouts[b.ID()], GiaID(id))
			}
		}
		// Refresh recursively on the private candidate; no partial cache escapes.
		if e := candidate.prepare(n, m, o, charge); e != nil {
			return e
		}
		candidate.builds = 1
		*s = *candidate
		return nil
	}
	if s.graph != m || len(s.objects) != size {
		return ErrInvariant
	}
	for i, obj := range m.objects {
		if e := charge(); e != nil {
			return e
		}
		if obj != s.objects[i] {
			return ErrInvariant
		}
	}
	if s.network != n {
		if _, e := n.ValidateFunctions(m, o); e != nil {
			return e
		}
		clear(s.refs)
		for _, fanins := range n.fanins {
			if e := charge(); e != nil {
				return e
			}
			for _, id := range fanins {
				if e := charge(); e != nil {
					return e
				}
				s.refs[id]++
			}
		}
		for _, id := range m.cos {
			if e := charge(); e != nil {
				return e
			}
			a, _ := m.objects[id].Fanin0(id)
			s.refs[a.ID()]++
		}
		copy(s.baseRefs, s.refs)
		s.network = n
		s.refreshes++
	}
	for i := 0; i < size; i++ {
		if e := charge(); e != nil {
			return e
		}
		s.in[i] = false
		s.no[i] = false
		s.used[i] = false
		s.leafSeen[i] = false
		s.seen[i] = 0
		s.refs[i] = s.baseRefs[i]
	}
	return nil
}
