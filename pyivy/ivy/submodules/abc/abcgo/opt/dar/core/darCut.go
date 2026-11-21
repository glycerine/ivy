package core

import (
	"fmt"
	"iter"
	"math/bits"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bool/kit"
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
	var sorted [4]aig.Edge
	n := 0
	for _, cut := range [2]Cut{a, b} {
		for _, edge := range cut.Leaves {
			i := 0
			for i < n && sorted[i] < edge {
				i++
			}
			if i < n && sorted[i] == edge {
				continue
			}
			if n == len(sorted) {
				return Cut{}, false
			}
			copy(sorted[i+1:], sorted[i:n])
			sorted[i] = edge
			n++
		}
	}
	return Cut{Leaves: append([]aig.Edge(nil), sorted[:n]...)}, true
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
	phase, used := 0, 0
	for i, leaf := range leaves {
		if used < len(c.Leaves) && c.Leaves[used] == leaf {
			phase |= 1 << i
			used++
		}
	}
	if used == len(c.Leaves) {
		return kit.TruthStretch4(c.Truth, uint32(phase))
	}
	// The public truth operation also accepts a reordered boundary.
	return stretchUnorderedCut(c, leaves)
}

func stretchUnorderedCut(c Cut, leaves []aig.Edge) uint16 {
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
		step := 1 << variable
		lowMask := [4]uint16{0x5555, 0x3333, 0x0f0f, 0x00ff}
		if (c.Truth^(c.Truth>>step))&lowMask[variable] != 0 {
			continue
		}
		var cofactor uint16
		mask := uint16(1<<step - 1)
		for dst := 0; dst < 8; dst += step {
			cofactor |= ((c.Truth >> (2 * dst)) & mask) << dst
		}
		c.Truth = cofactor | cofactor<<8
		c.Leaves = append(c.Leaves[:variable], c.Leaves[variable+1:]...)
	}
	return len(c.Leaves) != original
}

// CutSlot preserves DAR's physical slot order and unused holes.
type CutSlot struct {
	Cut   Cut
	Used  bool
	Value int
}

// CutCandidate holds a caller-owned copy of one bounded DAR cut.
type CutCandidate struct {
	Leaves    [4]aig.Edge
	Truth     uint16
	Value     uint16
	NumLeaves uint8
}

// CutEnumerator owns a bounded cut search over an immutable graph snapshot.
type CutEnumerator struct {
	skipTruthMin bool
	graph        *aig.Graph
	limit        int
	sets         [][]cachedCut
	pool         []cachedCut
	nodes        []aig.Node
}

func NewCutEnumerator(g *aig.Graph, limit int) *CutEnumerator {
	if limit < 2 {
		panic("dar: cut limit must be at least two")
	}
	return &CutEnumerator{graph: g, limit: limit}
}

func (m *CutEnumerator) slots(edge aig.Edge) ([]cachedCut, bool) {
	if node := m.graph.Node(edge); node.Kind == aig.Output {
		driver := node.Fanins[0]
		if edge.Complemented() {
			driver = driver.Not()
		}
		edge = driver
	}
	edge = aig.Aig_ObjReal_rec(m.graph, edge)
	if m.sets == nil {
		m.nodes = m.graph.Nodes()
		m.sets = make([][]cachedCut, len(m.nodes))
	}
	return m.build(edge.NodeID()), edge.Complemented()
}

// Candidates yields used cuts in physical slot order. The caller owns each
// four-element leaf array, including when the enumerator serves a cached root.
func (m *CutEnumerator) Candidates(edge aig.Edge) iter.Seq2[int, CutCandidate] {
	return func(yield func(int, CutCandidate) bool) {
		set, invert := m.slots(edge)
		for i := range set {
			if set[i].used && !yield(i, set[i].candidate(invert)) {
				return
			}
		}
	}
}

// Slots returns independently owned physical slots, including holes.
func (m *CutEnumerator) Slots(edge aig.Edge) []CutSlot {
	set, invert := m.slots(edge)
	leaves := 0
	for i := range set {
		leaves += int(set[i].NumLeaves)
	}
	storage := make([]aig.Edge, leaves)
	result := make([]CutSlot, m.limit)
	for i := range set {
		c := set[i].candidate(invert && set[i].used)
		result[i] = CutSlot{Cut: c.copyCut(&storage), Used: set[i].used, Value: int(c.Value)}
	}
	return result
}

// Cuts returns used slots in physical order, without holes.
func (m *CutEnumerator) Cuts(edge aig.Edge) []Cut {
	set, invert := m.slots(edge)
	count, leaves := 0, 0
	for i := range set {
		if set[i].used {
			count++
			leaves += int(set[i].NumLeaves)
		}
	}
	storage := make([]aig.Edge, leaves)
	result := make([]Cut, 0, count)
	for i := range set {
		if set[i].used {
			result = append(result, set[i].candidate(invert).copyCut(&storage))
		}
	}
	return result
}

func (c CutCandidate) copyCut(storage *[]aig.Edge) Cut {
	cut := Cut{Truth: c.Truth}
	if n := c.NumLeaves; n != 0 {
		cut.Leaves = (*storage)[:n:n]
		copy(cut.Leaves, c.Leaves[:n])
		*storage = (*storage)[n:]
	}
	return cut
}

// Dar_ManComputeCutSlots exposes the source physical slot order.
func Dar_ManComputeCutSlots(g *aig.Graph, limit int, skipTruthMin bool) (map[aig.Edge][]CutSlot, error) {
	m, err := NewWholeGraphCutEnumerator(g, limit, skipTruthMin)
	if err != nil {
		return nil, err
	}
	result := make(map[aig.Edge][]CutSlot, g.NodeCount())
	for _, e := range g.LiveEdges() {
		result[e] = m.Slots(e)
	}
	return result, nil
}

// Dar_ManComputeCuts returns independently owned cut boundaries for each live
// node. skipTruthMin preserves syntactically present but functionally unused
// boundary variables, matching the source cut-enumeration switch.
func Dar_ManComputeCuts(g *aig.Graph, limit int, skipTruthMin bool) (map[aig.Edge][]Cut, error) {
	m, err := NewWholeGraphCutEnumerator(g, limit, skipTruthMin)
	if err != nil {
		return nil, err
	}
	result := make(map[aig.Edge][]Cut, g.NodeCount())
	for _, e := range g.LiveEdges() {
		result[e] = m.Cuts(e)
	}
	return result, nil
}

// NewWholeGraphCutEnumerator validates the graph and removes dangling logic
// before evaluating any cut values, as in Dar_ManComputeCuts.
func NewWholeGraphCutEnumerator(g *aig.Graph, limit int, skipTruthMin bool) (*CutEnumerator, error) {
	if g == nil || limit < 2 {
		return nil, fmt.Errorf("dar: invalid cut request")
	}
	if err := g.Check(); err != nil {
		return nil, err
	}
	aig.Aig_ManCleanup(g)
	m := NewCutEnumerator(g, limit)
	m.skipTruthMin = skipTruthMin
	return m, nil
}
