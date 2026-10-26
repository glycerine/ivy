package abc

// Bounded SAT-window structural cuts from aig/gia/giaSatLut.c.

import "math/bits"

type GiaWindowCutOptions struct {
	MaxCuts, MaxVisits int // zero: one million each; includes leaf and unit records
	Stop               func() bool
}
type GiaWindowCuts struct {
	Cuts   []GiaWindowCut
	Ranges []GiaWindowCutRange // internal nodes only; each range ends in a unit cut
}

// GenerateWindowCuts enumerates native ordered, subset-minimal structural cuts
// for an explicit topological two-fanin window. Leaves and nodes are disjoint,
// unique ordered object IDs, each limited to 128. All node fanins must already
// appear in the window. Native feasibility uses width four for lutSize<=4 and
// width six for lutSize 5 or 6; it does not enforce the requested smaller width.
// Native mux nodes are rejected because the C two-fanin routine omits control.
func (m *GiaManager) GenerateWindowCuts(leaves, nodes []GiaID, lutSize int, o GiaWindowCutOptions) (GiaWindowCuts, error) {
	fail := func(e error) (GiaWindowCuts, error) { return GiaWindowCuts{}, e }
	if m == nil || len(leaves) > 128 || len(nodes) > 128 || lutSize < 1 || lutSize > 6 || o.MaxCuts < 0 || o.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if o.MaxCuts == 0 {
		o.MaxCuts = 1 << 20
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 1 << 20
	}
	work := 0
	charge := func() error {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		if work == o.MaxVisits {
			return ErrLimit
		}
		work++
		return nil
	}
	if e := charge(); e != nil {
		return fail(e)
	}
	r := GiaWindowCuts{Ranges: make([]GiaWindowCutRange, 0, len(nodes))}
	positions := make(map[GiaID]int, len(leaves)+len(nodes))
	spans := make([]GiaWindowCutRange, 0, len(leaves)+len(nodes))
	appendCut := func(c GiaWindowCut) error {
		if len(r.Cuts) == o.MaxCuts {
			return ErrLimit
		}
		r.Cuts = append(r.Cuts, c)
		return nil
	}
	for i, id := range leaves {
		if e := charge(); e != nil {
			return fail(e)
		}
		if int(id) >= len(m.objects) || m.objects[id].IsCO() {
			return fail(ErrDimension)
		}
		if _, ok := positions[id]; ok {
			return fail(ErrDimension)
		}
		positions[id] = i
		spans = append(spans, GiaWindowCutRange{Start: len(r.Cuts), Count: 1})
		c := GiaWindowCut{Node: -1}
		c.Leaves[i/64] = 1 << uint(i%64)
		if e := appendCut(c); e != nil {
			return fail(e)
		}
	}
	subset := func(a, b GiaWindowCut) bool {
		return a.Leaves[0]&b.Leaves[0] == a.Leaves[0] && a.Leaves[1]&b.Leaves[1] == a.Leaves[1] && a.Nodes[0]&b.Nodes[0] == a.Nodes[0] && a.Nodes[1]&b.Nodes[1] == a.Nodes[1]
	}
	width := 4
	if lutSize > 4 {
		width = 6
	}
	for node, id := range nodes {
		if e := charge(); e != nil {
			return fail(e)
		}
		if int(id) >= len(m.objects) || !m.objects[id].IsAnd() {
			return fail(ErrDimension)
		}
		if _, ok := positions[id]; ok {
			return fail(ErrDimension)
		}
		if len(m.muxes) > int(id) && m.muxes[id] != 0 {
			return fail(ErrDimension)
		}
		obj := m.objects[id]
		a, e := obj.Fanin0(id)
		if e != nil {
			return fail(e)
		}
		b, e := obj.Fanin1(id)
		if e != nil {
			return fail(e)
		}
		ai, ok := positions[a.ID()]
		if !ok {
			return fail(ErrDimension)
		}
		bi, ok := positions[b.ID()]
		if !ok {
			return fail(ErrDimension)
		}
		sa, sb := spans[ai], spans[bi]
		var temp []GiaWindowCut
		for i := sa.Start; i < sa.Start+sa.Count; i++ {
			for j := sb.Start; j < sb.Start+sb.Count; j++ {
				if e := charge(); e != nil {
					return fail(e)
				}
				x, y := r.Cuts[i], r.Cuts[j]
				c := GiaWindowCut{Node: node, Leaves: [2]uint64{x.Leaves[0] | y.Leaves[0], x.Leaves[1] | y.Leaves[1]}, Nodes: [2]uint64{x.Nodes[0] | y.Nodes[0], x.Nodes[1] | y.Nodes[1]}}
				if bits.OnesCount64(c.Leaves[0])+bits.OnesCount64(c.Leaves[1])+bits.OnesCount64(c.Nodes[0])+bits.OnesCount64(c.Nodes[1]) > width {
					continue
				}
				contained := false
				for _, old := range temp {
					if e := charge(); e != nil {
						return fail(e)
					}
					if subset(old, c) {
						contained = true
						break
					}
				}
				if contained {
					continue
				}
				k := 0
				for _, old := range temp {
					if e := charge(); e != nil {
						return fail(e)
					}
					if !subset(c, old) {
						temp[k] = old
						k++
					}
				}
				temp = temp[:k]
				if len(temp) >= o.MaxCuts-len(r.Cuts)-1 {
					return fail(ErrLimit)
				}
				temp = append(temp, c)
			}
		}
		span := GiaWindowCutRange{Start: len(r.Cuts), Count: len(temp) + 1}
		for _, c := range temp {
			if e := appendCut(c); e != nil {
				return fail(e)
			}
		}
		c := GiaWindowCut{Node: node}
		c.Nodes[node/64] = 1 << uint(node%64)
		if e := appendCut(c); e != nil {
			return fail(e)
		}
		r.Ranges = append(r.Ranges, span)
		spans = append(spans, span)
		positions[id] = len(leaves) + node
	}
	return r, nil
}
