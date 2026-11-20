package core

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"math/bits"
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

// A proper subset has at most three inputs and 256 distinct functions.
// Each row expands a compact function onto the selected four-variable positions.
var stretchedCutTruth = func() (table [15][256]uint16) {
	for phase := range table {
		for assignment := 0; assignment < 16; assignment++ {
			src, bit := 0, 0
			for i := 0; i < 4; i++ {
				if phase&(1<<i) != 0 {
					src |= ((assignment >> i) & 1) << bit
					bit++
				}
			}
			for value := range table[phase] {
				table[phase][value] |= uint16((value>>src)&1) << assignment
			}
		}
	}
	return table
}()

func cutStretch(c Cut, leaves []aig.Edge) uint16 {
	phase, used := 0, 0
	for i, leaf := range leaves {
		if used < len(c.Leaves) && c.Leaves[used] == leaf {
			phase |= 1 << i
			used++
		}
	}
	if used == len(c.Leaves) {
		if phase == 15 {
			return c.Truth
		}
		return stretchedCutTruth[phase][byte(c.Truth)]
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

// CutEnumerator owns a bounded cut search over an immutable graph snapshot.
type CutEnumerator struct {
	skipTruthMin bool
	graph        *aig.Graph
	limit        int
	cuts         map[int][]CutSlot
	nodes        []aig.Node
}

func NewCutEnumerator(g *aig.Graph, limit int) *CutEnumerator {
	if limit < 2 {
		panic("dar: cut limit must be at least two")
	}
	return &CutEnumerator{graph: g, limit: limit}
}

// Slots returns independently owned physical slots, including holes.
func (m *CutEnumerator) Slots(edge aig.Edge) []CutSlot {
	if node := m.graph.Node(edge); node.Kind == aig.Output {
		if edge.Complemented() {
			edge = node.Fanins[0].Not()
		} else {
			edge = node.Fanins[0]
		}
	}
	edge = aig.Aig_ObjReal_rec(m.graph, edge)
	if m.cuts == nil {
		m.cuts = make(map[int][]CutSlot)
		m.nodes = m.graph.Nodes()
	}
	if _, ok := m.cuts[edge.NodeID()]; !ok {
		if err := darComputeSlots(m.graph, m.nodes, m.limit, m.skipTruthMin, edge, m.cuts); err != nil {
			panic(err)
		}
	}
	result := append([]CutSlot(nil), m.cuts[edge.NodeID()]...)
	for i := range result {
		result[i].Cut.Leaves = append([]aig.Edge(nil), result[i].Cut.Leaves...)
		if result[i].Used && edge.Complemented() {
			result[i].Cut.Truth = ^result[i].Cut.Truth
		}
	}
	return result
}

// Cuts returns used slots in physical order, without holes.
func (m *CutEnumerator) Cuts(edge aig.Edge) []Cut {
	var result []Cut
	for _, slot := range m.Slots(edge) {
		if slot.Used {
			result = append(result, slot.Cut)
		}
	}
	return result
}

// Dar_ManComputeCutSlots exposes the source physical slot order.
func Dar_ManComputeCutSlots(g *aig.Graph, limit int, skipTruthMin bool) (map[aig.Edge][]CutSlot, error) {
	m, err := wholeGraphCutEnumerator(g, limit, skipTruthMin)
	if err != nil {
		return nil, err
	}
	result := make(map[aig.Edge][]CutSlot)
	for _, e := range g.LiveEdges() {
		result[e] = m.Slots(e)
	}
	return result, nil
}

func darComputeSlots(g *aig.Graph, nodes []aig.Node, limit int, skipTruthMin bool, root aig.Edge, slots map[int][]CutSlot) error {
	if g == nil || limit < 2 {
		return fmt.Errorf("dar: invalid cut request")
	}
	active := make(map[int]bool)
	findValue := func(c Cut) int {
		if len(c.Leaves) < 2 {
			return 1001
		}
		sum, ones := 0, 0
		for _, e := range c.Leaves {
			r := nodes[e.NodeID()].Refs
			sum += r
			if r == 1 {
				ones++
			}
		}
		if sum > 1000 {
			sum = 1000
		}
		if ones > 3 {
			sum = 5 - ones
		}
		return sum
	}
	signature := func(c Cut) uint32 {
		var sig uint32
		for _, e := range c.Leaves {
			sig |= 1 << uint(e.NodeID()&31)
		}
		return sig
	}
	filter := func(set []CutSlot, index int) bool {
		c := set[index].Cut
		cs := signature(c)
		for j := range set {
			if j == index || !set[j].Used {
				continue
			}
			old := set[j].Cut
			os := signature(old)
			if len(old.Leaves) > len(c.Leaves) {
				if os&cs == cs && Dar_CutCheckDominance(c, old) {
					set[j].Used = false
				}
			} else if os&cs == os && Dar_CutCheckDominance(old, c) {
				set[index].Used = false
				return true
			}
		}
		return false
	}
	freeSlot := func(set []CutSlot) int {
		for i := range set {
			if !set[i].Used {
				return i
			}
		}
		best := -1
		for pass := 0; pass < 3; pass++ {
			for i, s := range set {
				if pass == 0 && len(s.Cut.Leaves) < 3 || pass == 1 && len(s.Cut.Leaves) < 2 {
					continue
				}
				if best < 0 || set[best].Value > s.Value {
					best = i
				}
			}
			if best >= 0 {
				return best
			}
		}
		panic("dar: cut slots exhausted")
	}
	var build func(aig.Edge) error
	build = func(edge aig.Edge) error {
		id := edge.NodeID()
		if _, ok := slots[id]; ok {
			return nil
		}
		if active[id] {
			return fmt.Errorf("dar: cyclic AIG")
		}
		active[id] = true
		defer delete(active, id)
		node := nodes[id]
		set := make([]CutSlot, limit)
		unit := Cut{Leaves: []aig.Edge{aig.Edge(id << 1)}, Truth: 0xaaaa}
		if id == 0 {
			unit = Cut{Truth: 0xffff}
		}
		set[0] = CutSlot{Cut: unit, Used: true, Value: findValue(unit)}
		if node.Kind == aig.Constant || node.Kind == aig.Input {
			slots[id] = set
			return nil
		}
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
			return fmt.Errorf("dar: unsupported node kind %d", node.Kind)
		}
		a, b := aig.Aig_ObjReal_rec(g, node.Fanins[0]), aig.Aig_ObjReal_rec(g, node.Fanins[1])
		if err := build(a); err != nil {
			return err
		}
		if err := build(b); err != nil {
			return err
		}
		for _, left := range slots[a.NodeID()] {
			if !left.Used {
				continue
			}
			for _, right := range slots[b.NodeID()] {
				if !right.Used {
					continue
				}
				if bits.OnesCount32(signature(left.Cut)|signature(right.Cut)) > 4 {
					continue
				}
				index := freeSlot(set)
				set[index].Used = false
				merged, ok := Dar_CutMerge(left.Cut, right.Cut)
				if !ok {
					continue
				}
				set[index] = CutSlot{Cut: merged, Used: true}
				if filter(set, index) {
					continue
				}
				x, y := cutStretch(left.Cut, merged.Leaves), cutStretch(right.Cut, merged.Leaves)
				if a.Complemented() {
					x = ^x
				}
				if b.Complemented() {
					y = ^y
				}
				if node.Kind == aig.ExclusiveOr {
					set[index].Cut.Truth = x ^ y
				} else {
					set[index].Cut.Truth = x & y
				}
				if !skipTruthMin && Dar_CutSuppMinimize(&set[index].Cut) {
					if filter(set, index) {
						return fmt.Errorf("dar: invalid minimized cut")
					}
				}
				set[index].Value = findValue(set[index].Cut)
				if set[index].Value == 0 {
					set[index].Used = false
				} else if len(set[index].Cut.Leaves) < 2 {
					slots[id] = set
					return nil
				}
			}
		}
		slots[id] = set
		return nil
	}
	return build(root)
}

// Dar_ManComputeCuts returns independently owned cut boundaries for each live
// node. skipTruthMin preserves syntactically present but functionally unused
// boundary variables, matching the source cut-enumeration switch.
func Dar_ManComputeCuts(g *aig.Graph, limit int, skipTruthMin bool) (map[aig.Edge][]Cut, error) {
	m, err := wholeGraphCutEnumerator(g, limit, skipTruthMin)
	if err != nil {
		return nil, err
	}
	result := make(map[aig.Edge][]Cut)
	for _, e := range g.LiveEdges() {
		result[e] = m.Cuts(e)
	}
	return result, nil
}

func wholeGraphCutEnumerator(g *aig.Graph, limit int, skipTruthMin bool) (*CutEnumerator, error) {
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
