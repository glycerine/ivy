package core

import (
	"fmt"
	"math/bits"

	"github.com/glycerine/abcgo/aig/aig"
)

type cachedCut struct {
	CutCandidate
	sign uint32
	used bool
}

func (s *cachedCut) candidate(invert bool) CutCandidate {
	c := s.CutCandidate
	if invert {
		c.Truth = ^c.Truth
	}
	return c
}

func (s *cachedCut) cut() Cut { return Cut{s.Leaves[:s.NumLeaves], s.Truth} }

func cutSign(edge aig.Edge) uint32 { return 1 << (edge.NodeID() & 31) }

// Reserve enough room for the search. build commits the populated prefix
// after both children have returned, so the unused tail serves the next node.
func (m *CutEnumerator) reserveSlots(n int) []cachedCut {
	if len(m.pool) < n {
		count := min(len(m.nodes), max(1, 1024/n))
		m.pool = make([]cachedCut, count*n)
	}
	return m.pool[:n:n]
}

func (m *CutEnumerator) build(id int) []cachedCut {
	if set := m.sets[id]; set != nil {
		if len(set) == 0 {
			panic("dar: cyclic AIG")
		}
		return set
	}
	m.sets[id] = []cachedCut{} // active recursion
	node := &m.nodes[id]
	var left, right []cachedCut
	var a, b aig.Edge
	capacity := 1
	switch node.Kind {
	case aig.And, aig.ExclusiveOr:
		a = aig.Aig_ObjReal_rec(m.graph, node.Fanins[0])
		b = aig.Aig_ObjReal_rec(m.graph, node.Fanins[1])
		left, right = m.build(a.NodeID()), m.build(b.NodeID())
		capacity = m.limit
	case aig.Constant, aig.Input:
	default:
		panic(fmt.Sprintf("dar: unsupported node kind %d", node.Kind))
	}
	set := m.reserveSlots(capacity)
	unit := &set[0]
	unit.used, unit.Value = true, 1001
	if id == 0 {
		unit.Truth = 0xffff
	} else {
		unit.NumLeaves, unit.Truth = 1, 0xaaaa
		unit.Leaves[0] = aig.Edge(id) << 1
		unit.sign = cutSign(unit.Leaves[0])
	}
	length := 1
	if left != nil {
		length = m.combine(set, left, right, a.Complemented(), b.Complemented(), node.Kind == aig.ExclusiveOr)
	}
	set = set[:length:length]
	m.pool = m.pool[length:]
	m.sets[id] = set
	return set
}

func (m *CutEnumerator) combine(set, left, right []cachedCut, invertLeft, invertRight, xor bool) int {
	extent := 1
	for i := range left {
		a := &left[i]
		if !a.used {
			continue
		}
		for j := range right {
			b := &right[j]
			if !b.used || bits.OnesCount32(a.sign|b.sign) > 4 {
				continue
			}
			// ABC evicts before the exact union test, even when the merge fails.
			pos := freeCutSlot(set[:min(extent+1, len(set))])
			extent = max(extent, pos+1)
			set[pos].used = false
			merged, leftPhase, rightPhase, ok := mergeCutSlots(a, b)
			if !ok {
				continue
			}
			set[pos] = merged
			if filterCutSlot(set[:extent], pos) {
				continue
			}
			s := &set[pos]
			x := expandCutTruth(a.Truth, leftPhase)
			y := expandCutTruth(b.Truth, rightPhase)
			if invertLeft {
				x = ^x
			}
			if invertRight {
				y = ^y
			}
			s.Truth = x & y
			if xor {
				s.Truth = x ^ y
			}
			if !m.skipTruthMin {
				c := s.cut()
				if Dar_CutSuppMinimize(&c) {
					s.NumLeaves, s.Truth = uint8(len(c.Leaves)), c.Truth
					s.sign = 0
					for _, leaf := range c.Leaves {
						s.sign |= cutSign(leaf)
					}
					if filterCutSlot(set[:extent], pos) {
						panic("dar: invalid minimized cut")
					}
				}
			}
			s.Value = m.value(s)
			if s.Value == 0 {
				s.used = false
			} else if s.NumLeaves < 2 {
				return extent
			}
		}
	}
	return extent
}

// Merge two sorted boundaries and record the positions of each input in their
// union. Their truth tables can then be expanded without searching the leaves.
func mergeCutSlots(a, b *cachedCut) (merged cachedCut, phaseA, phaseB int, ok bool) {
	left, right := a.Leaves[:a.NumLeaves], b.Leaves[:b.NumLeaves]
	for len(left) != 0 || len(right) != 0 {
		pos := merged.NumLeaves
		if pos == 4 {
			return cachedCut{}, 0, 0, false
		}
		leaf := aig.InvalidEdge
		if len(left) != 0 {
			leaf = left[0]
		}
		if len(right) != 0 && right[0] < leaf {
			leaf = right[0]
		}
		if len(left) != 0 && left[0] == leaf {
			phaseA |= 1 << pos
			left = left[1:]
		}
		if len(right) != 0 && right[0] == leaf {
			phaseB |= 1 << pos
			right = right[1:]
		}
		merged.Leaves[pos] = leaf
		merged.NumLeaves++
	}
	merged.sign, merged.used = a.sign|b.sign, true
	return merged, phaseA, phaseB, true
}

func filterCutSlot(set []cachedCut, pos int) bool {
	cur := &set[pos]
	for i := range set {
		old := &set[i]
		if i == pos || !old.used {
			continue
		}
		if old.NumLeaves > cur.NumLeaves {
			if old.sign&cur.sign == cur.sign && Dar_CutCheckDominance(cur.cut(), old.cut()) {
				old.used = false
			}
		} else if old.sign&cur.sign == old.sign && Dar_CutCheckDominance(old.cut(), cur.cut()) {
			cur.used = false
			return true
		}
	}
	return false
}

func freeCutSlot(set []cachedCut) int {
	best, bestClass := 0, 0
	for i := range set {
		s := &set[i]
		if !s.used {
			return i
		}
		// Evict from 3+ leaves first, then two, then trivial cuts. Within
		// the same class, the first cut with minimum value loses its slot.
		class := min(3, max(1, int(s.NumLeaves)))
		if class > bestClass || class == bestClass && s.Value < set[best].Value {
			best, bestClass = i, class
		}
	}
	return best
}

func (m *CutEnumerator) value(c *cachedCut) uint16 {
	if c.NumLeaves < 2 {
		return 1001
	}
	sum, ones := 0, 0
	for _, edge := range c.Leaves[:c.NumLeaves] {
		refs := m.nodes[edge.NodeID()].Refs
		sum += refs
		if refs == 1 {
			ones++
		}
	}
	if ones > 3 {
		return uint16(5 - ones)
	}
	return uint16(min(sum, 1000))
}
