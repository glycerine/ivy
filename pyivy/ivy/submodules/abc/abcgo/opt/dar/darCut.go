package dar

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"math/bits"
	"sort"
)

// Cut describes a function over up to four sorted, regular graph edges.
// Truth repeats over unused variables, so constants are 0 and 0xffff.
type Cut struct {
	Leaves []aig.Edge
	Truth  uint16
}
type Dar_Cut_t = Cut

func Dar_WordCountOnes(word uint32) int32 { return int32(bits.OnesCount32(word)) }

// Dar_CutMerge returns the ordered union, or false when it exceeds four inputs.
func Dar_CutMerge(a, b Cut) (Cut, bool) {
	leaves := append(append([]aig.Edge(nil), a.Leaves...), b.Leaves...)
	sort.Slice(leaves, func(i, j int) bool { return leaves[i] < leaves[j] })
	unique := leaves[:0]
	for _, e := range leaves {
		if len(unique) == 0 || unique[len(unique)-1] != e {
			unique = append(unique, e)
		}
	}
	if len(unique) > 4 {
		return Cut{}, false
	}
	return Cut{Leaves: unique}, true
}
func Dar_CutCheckDominance(a, b Cut) bool {
	for _, e := range a.Leaves {
		found := false
		for _, f := range b.Leaves {
			if e == f {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func cutStretch(c Cut, leaves []aig.Edge) uint16 {
	var mapping [4]int
	for i, e := range c.Leaves {
		mapping[i] = -1
		for j, f := range leaves {
			if e == f {
				mapping[i] = j
				break
			}
		}
		if mapping[i] < 0 {
			panic("dar: missing cut boundary")
		}
	}
	var truth uint16
	for assignment := 0; assignment < 16; assignment++ {
		index := 0
		for i := range c.Leaves {
			index |= ((assignment >> mapping[i]) & 1) << i
		}
		truth |= ((c.Truth >> index) & 1) << assignment
	}
	return truth
}
func Dar_CutTruth(c, a, b Cut, complementA, complementB bool) uint16 {
	x, y := cutStretch(a, c.Leaves), cutStretch(b, c.Leaves)
	if complementA {
		x = ^x
	}
	if complementB {
		y = ^y
	}
	return x & y
}

// Dar_CutSuppMinimize removes inputs on which Truth does not depend.
func Dar_CutSuppMinimize(c *Cut) bool {
	original := len(c.Leaves)
	for variable := len(c.Leaves) - 1; variable >= 0; variable-- {
		dependent := false
		for index := 0; index < 16; index++ {
			if index&(1<<variable) == 0 && ((c.Truth>>index)^(c.Truth>>(index|(1<<variable))))&1 != 0 {
				dependent = true
				break
			}
		}
		if dependent {
			continue
		}
		var reduced uint16
		for index := 0; index < 16; index++ {
			old := (index & ((1 << variable) - 1)) | ((index >> variable) << (variable + 1))
			reduced |= ((c.Truth >> (old & 15)) & 1) << index
		}
		c.Truth = reduced
		c.Leaves = append(c.Leaves[:variable], c.Leaves[variable+1:]...)
	}
	return len(c.Leaves) != original
}

// CutEnumerator owns a bounded memoized cut search for one immutable graph
// snapshot. Construct a fresh enumerator after graph replacement.
type CutEnumerator struct {
	skipTruthMin bool
	graph        *aig.Graph
	limit        int
	cuts         map[int][]Cut
	visiting     map[int]bool
}

func NewCutEnumerator(g *aig.Graph, limit int) *CutEnumerator {
	if limit < 2 {
		panic("dar: cut limit must be at least two")
	}
	return &CutEnumerator{graph: g, limit: limit, cuts: make(map[int][]Cut), visiting: make(map[int]bool)}
}
func (m *CutEnumerator) Cuts(edge aig.Edge) []Cut {
	result := m.compute(aig.Edge(edge.NodeID() << 1))
	copyCuts := make([]Cut, len(result))
	for i, c := range result {
		copyCuts[i] = Cut{append([]aig.Edge(nil), c.Leaves...), c.Truth}
		if edge.Complemented() {
			copyCuts[i].Truth = ^c.Truth
		}
	}
	return copyCuts
}
func (m *CutEnumerator) compute(edge aig.Edge) []Cut {
	id := edge.NodeID()
	if cached, ok := m.cuts[id]; ok {
		return cached
	}
	if m.visiting[id] {
		panic("dar: cyclic graph")
	}
	m.visiting[id] = true
	defer delete(m.visiting, id)
	node := m.graph.Node(edge)
	if node.Kind == aig.Constant {
		m.cuts[id] = []Cut{{Truth: 0xffff}}
		return m.cuts[id]
	}
	result := []Cut{{Leaves: []aig.Edge{edge}, Truth: 0xaaaa}}
	if node.Kind == aig.Input {
		m.cuts[id] = result
		return result
	}
	left := m.Cuts(node.Fanins[0])
	if node.Kind == aig.Buffer || node.Kind == aig.Output {
		result = append(result, left...)
	} else {
		right := m.Cuts(node.Fanins[1])
		for _, a := range left {
			for _, b := range right {
				merged, ok := Dar_CutMerge(a, b)
				if !ok {
					continue
				}
				x, y := cutStretch(a, merged.Leaves), cutStretch(b, merged.Leaves)
				if node.Kind == aig.ExclusiveOr {
					merged.Truth = x ^ y
				} else {
					merged.Truth = x & y
				}
				if !m.skipTruthMin {
					Dar_CutSuppMinimize(&merged)
				}
				duplicate := false
				for _, old := range result {
					if len(old.Leaves) == len(merged.Leaves) && Dar_CutCheckDominance(old, merged) {
						duplicate = true
						break
					}
				}
				if !duplicate {
					result = append(result, merged)
				}
			}
		}
	}
	// Preserve the unit cut, then prefer compact boundaries deterministically.
	sort.SliceStable(result[1:], func(i, j int) bool { return len(result[i+1].Leaves) < len(result[j+1].Leaves) })
	if len(result) > m.limit {
		result = result[:m.limit]
	}
	m.cuts[id] = result
	return result
}

// Dar_ManComputeCuts returns independently owned cut boundaries for each live
// node. skipTruthMin preserves syntactically present but functionally unused
// boundary variables, matching the source cut-enumeration switch.
func Dar_ManComputeCuts(g *aig.Graph, limit int, skipTruthMin bool) (map[aig.Edge][]Cut, error) {
	if g == nil {
		return nil, fmt.Errorf("dar: nil graph")
	}
	if limit < 2 {
		return nil, fmt.Errorf("dar: cut limit must be at least two")
	}
	if err := g.Check(); err != nil {
		return nil, err
	}
	m := NewCutEnumerator(g, limit)
	m.skipTruthMin = skipTruthMin
	result := make(map[aig.Edge][]Cut)
	for _, e := range g.LiveEdges() {
		result[e] = m.Cuts(e)
	}
	return result, nil
}
