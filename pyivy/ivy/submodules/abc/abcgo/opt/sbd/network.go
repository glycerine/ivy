package sbd

import (
	"fmt"
	"slices"

	"github.com/glycerine/abcgo/aig/aig"
)

// LUTCut is a mapped root's ordered object-ID support and normalized truth.
// Truth bit i evaluates the support assignment whose kth input is bit k of i.
type LUTCut struct {
	Leaves []int32
	Truth  []uint64
}

// Network keeps the authoritative native graph together with SBD's local mapping
// metadata. BoxInputs maps a box-output CI object ID to its box-input CO IDs.
// Mirrors use ABC/Gia literals (0=false) and -1 for no replacement; only boundary
// conversion changes constant polarity because aig.Graph uses 0=true.
type Network struct {
	Graph     *aig.Graph
	LUTs      map[int32]LUTCut
	Levels    []int32
	Mirrors   []int32
	LUTRefs   []int32
	BoxInputs map[int32][]int32
	// BoxLevelInputs optionally restricts level propagation to enabled timing arcs.
	// Critical-path backtracing still uses all BoxInputs, matching the source.
	BoxLevelInputs map[int32][]int32
}

func sourceLiteral(edge aig.Edge) int32 {
	if edge.NodeID() == 0 {
		return int32(edge) ^ 1
	}
	return int32(edge)
}
func graphLiteral(lit int32) aig.Edge {
	if lit < 0 {
		panic("negative SBD literal")
	}
	if lit>>1 == 0 {
		return aig.Edge(lit ^ 1)
	}
	return aig.Edge(lit)
}
func NewNetwork(graph *aig.Graph) *Network {
	if graph == nil {
		panic("nil SBD graph")
	}
	if err := graph.Check(); err != nil {
		panic(err)
	}
	p := &Network{Graph: graph, LUTs: map[int32]LUTCut{}, BoxInputs: map[int32][]int32{}}
	p.resize()
	for _, edge := range graph.LiveEdges() {
		node := graph.Node(edge)
		if node.Kind == aig.And || node.Kind == aig.ExclusiveOr {
			a, b := uint64(0xaaaaaaaaaaaaaaaa), uint64(0xcccccccccccccccc)
			if node.Fanins[0].Complemented() {
				a = ^a
			}
			if node.Fanins[1].Complemented() {
				b = ^b
			}
			truth := a & b
			if node.Kind == aig.ExclusiveOr {
				truth = a ^ b
			}
			p.LUTs[int32(edge.NodeID())] = LUTCut{Leaves: []int32{int32(node.Fanins[0].NodeID()), int32(node.Fanins[1].NodeID())}, Truth: []uint64{truth}}
		}
	}
	return p
}
func (p *Network) resize() {
	count := len(p.Graph.Nodes())
	for len(p.Mirrors) < count {
		p.Mirrors = append(p.Mirrors, -1)
	}
	p.Levels = append(p.Levels, make([]int32, max(0, count-len(p.Levels)))...)
	p.LUTRefs = append(p.LUTRefs, make([]int32, max(0, count-len(p.LUTRefs)))...)
}
func (p *Network) object(id int32) aig.Node {
	if id < 0 || !p.Graph.ValidEdge(aig.Edge(id)<<1) {
		panic("invalid SBD object")
	}
	return p.Graph.Node(aig.Edge(id) << 1)
}
func (p *Network) cut(id int32) []int32 {
	cut, ok := p.LUTs[id]
	if !ok {
		panic(fmt.Sprintf("missing LUT cut for object %d", id))
	}
	return cut.Leaves
}
func (p *Network) outputDriver(co int32) int32 {
	node := p.object(co)
	if node.Kind != aig.Output {
		panic("box input is not a CO")
	}
	return int32(node.Fanins[0].NodeID())
}
func (p *Network) deriveLevels() int32 {
	p.resize()
	state := make([]byte, len(p.Levels))
	var visit func(int32) int32
	visit = func(id int32) int32 {
		if state[id] == 2 {
			return p.Levels[id]
		}
		if state[id] == 1 {
			panic("cyclic mapped timing graph")
		}
		state[id] = 1
		node := p.object(id)
		level := int32(0)
		switch node.Kind {
		case aig.Input:
			boxInputs := p.BoxInputs[id]
			if enabled, ok := p.BoxLevelInputs[id]; ok {
				boxInputs = enabled
			}
			for _, co := range boxInputs {
				driver := p.outputDriver(co)
				if driver != 0 {
					level = max(level, visit(driver)+1)
				}
			}
		case aig.And, aig.ExclusiveOr:
			for _, fanin := range p.cut(id) {
				level = max(level, visit(fanin)+1)
			}
		case aig.Buffer, aig.Output:
			level = visit(int32(node.Fanins[0].NodeID()))
		}
		p.Levels[id] = level
		state[id] = 2
		return level
	}
	maximum := int32(0)
	for _, co := range p.Graph.Outputs() {
		maximum = max(maximum, visit(int32(co.NodeID())))
	}
	return maximum
}
func (p *Network) SetLUT(root int32, cut LUTCut) {
	p.object(root)
	for _, leaf := range cut.Leaves {
		p.object(leaf)
	}
	p.LUTs[root] = LUTCut{Leaves: slices.Clone(cut.Leaves), Truth: slices.Clone(cut.Truth)}
}

// SetBoxOutput records all box-input COs and a separate enabled-arc mask for
// level propagation. A nil enabled mask means every arc is enabled.
func (p *Network) SetBoxOutput(ci int32, inputs []int32, enabled []bool) {
	if p.object(ci).Kind != aig.Input {
		panic("box output must be a CI")
	}
	if enabled != nil && len(enabled) != len(inputs) {
		panic("box timing mask size")
	}
	all, active := slices.Clone(inputs), make([]int32, 0, len(inputs))
	for i, co := range all {
		p.outputDriver(co)
		if enabled == nil || enabled[i] {
			active = append(active, co)
		}
	}
	if p.BoxInputs == nil {
		p.BoxInputs = map[int32][]int32{}
	}
	if p.BoxLevelInputs == nil {
		p.BoxLevelInputs = map[int32][]int32{}
	}
	p.BoxInputs[ci] = all
	p.BoxLevelInputs[ci] = active
}
