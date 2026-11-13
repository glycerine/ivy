package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/gia/mf"
	"sort"
)

type cecMFNode struct {
	a, b, c int32
	kind    int
	sibling int32
}
type cecMFNetwork struct {
	nodes           []cecMFNode
	inputs, outputs []int32
	registers       int
}

func (n *cecMFNetwork) ObjectCount() int        { return len(n.nodes) }
func (n *cecMFNetwork) InputLiterals() []int32  { return n.inputs }
func (n *cecMFNetwork) OutputLiterals() []int32 { return n.outputs }
func (n *cecMFNetwork) RegisterCount() int      { return n.registers }
func (n *cecMFNetwork) Sibling(id int32) int32  { return n.nodes[id].sibling }
func (n *cecMFNetwork) Gate(id int32) (int32, int32, int32, int) {
	x := n.nodes[id]
	return x.a, x.b, x.c, x.kind
}

// cecNaiveMappedCNF mirrors MF's source CI/CO variable remapping. MF returns
// IDs for its coarsened network; the source maps those IDs back to the input
// manager's objects. ABC coarsens COs in physical order, then maps variables back in CO-vector
// order; reordered CO vectors therefore retain the source behavior.
func cecNaiveMappedCNF(graph *Gia_Man_t, coIDs []int32) (*mf.CNF, []int32, error) {
	n := &cecMFNetwork{nodes: []cecMFNode{{}}, registers: int(gia.Gia_ManRegNum(graph))}
	copies := make([]int32, graph.ObjectCount())
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	hasChoices := false
	for id := int32(1); int(id) < graph.ObjectCount(); id++ {
		a, b, c, kind := graph.Gate(id)
		if kind == 6 {
			continue
		}
		if kind < 1 || kind > 5 {
			return nil, nil, fmt.Errorf("cec: invalid MF object %d", id)
		}
		node := cecMFNode{kind: kind}
		remap := func(lit int32) int32 {
			if lit < 0 || int(lit>>1) >= len(copies) || copies[lit>>1] < 0 {
				return -1
			}
			return copies[lit>>1] ^ (lit & 1)
		}
		switch kind {
		case 2, 3:
			node.a, node.b = remap(a), remap(b)
		case 4:
			node.a, node.b, node.c = remap(a), remap(b), remap(c)
		case 5:
			node.a = remap(a)
		}
		if (kind == 2 || kind == 3) && (node.a < 0 || node.b < 0) || kind == 4 && (node.a < 0 || node.b < 0 || node.c < 0) || kind == 5 && node.a < 0 {
			return nil, nil, fmt.Errorf("cec: invalid MF fanin at %d", id)
		}
		if sibling := gia.Gia_ObjSibl(graph, id); sibling != 0 {
			if sibling < 0 || int(sibling) >= len(copies) || copies[sibling] < 0 {
				return nil, nil, fmt.Errorf("cec: invalid MF sibling")
			}
			node.sibling = copies[sibling] >> 1
			hasChoices = true
		}
		copies[id] = int32(len(n.nodes)) << 1
		n.nodes = append(n.nodes, node)
		if kind == 1 {
			n.inputs = append(n.inputs, copies[id])
		}
	}
	physicalCOs := append([]int32(nil), coIDs...)
	sort.Slice(physicalCOs, func(i, j int) bool { return physicalCOs[i] < physicalCOs[j] })
	for _, id := range physicalCOs {
		if id < 0 || int(id) >= graph.ObjectCount() {
			return nil, nil, fmt.Errorf("cec: invalid output ID")
		}
		a, _, _, kind := graph.Gate(id)
		if kind != 6 || a < 0 || copies[a>>1] < 0 {
			return nil, nil, fmt.Errorf("cec: invalid output object")
		}
		lit := copies[a>>1] ^ (a & 1)
		n.outputs = append(n.outputs, lit)
		n.nodes = append(n.nodes, cecMFNode{a: lit, kind: 6})
	}
	var network mf.Network = n
	if !hasChoices {
		coarse, err := mf.Coarsen(n, 3)
		if err != nil {
			return nil, nil, err
		}
		network = coarse
	}
	options := mf.DefaultOptions()
	options.LUTSize = 8
	mapping, err := mf.GenerateMapping(network, options)
	if err != nil {
		return nil, nil, err
	}
	data, err := mf.DeriveCNF(mapping, false, false, false)
	if err != nil {
		return nil, nil, err
	}
	vars := make([]int32, 0, len(coIDs))
	for id := 0; id < network.ObjectCount(); id++ {
		_, _, _, kind := network.Gate(int32(id))
		if kind == 6 {
			vars = append(vars, data.VarNums[id])
		}
	}
	if len(vars) != len(coIDs) {
		return nil, nil, fmt.Errorf("cec: MF output mapping mismatch")
	}
	mapped := make([]int32, graph.ObjectCount())
	for i := range mapped {
		mapped[i] = -1
	}
	for i, id := range coIDs {
		mapped[id] = vars[i]
	}
	return data, mapped, nil
}
