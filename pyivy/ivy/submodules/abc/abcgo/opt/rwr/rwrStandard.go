package rwr

import (
	"slices"
	"sync"

	"github.com/glycerine/abcgo/bool/dec"
)

type libraryGate struct {
	input0, input1 uint32
	flags          [3]uint8
}

type libraryGraph struct {
	forestID int32
	leaves   int32
	root     uint32
	constant bool
	gates    []libraryGate
}

type libraryTraversal struct {
	id     int32
	volume uint32
}

// Only the built-in forest always starts from the same table and traversal
// counter. Graph payloads, scores and class members belong to each manager.
var standardLibrary struct {
	once      sync.Once
	classes   [][]libraryGraph
	inverse   []uint16
	traversal []libraryTraversal
	lastID    int32
}

func preprocessStandardLibrary(p *Rwr_Man_t) {
	initialized := false
	standardLibrary.once.Do(func() {
		Rwr_ManPreprocess(p)
		standardLibrary.inverse = slices.Clone(p.pMapInv)
		standardLibrary.lastID = p.nTravIds
		standardLibrary.traversal = make([]libraryTraversal, len(p.vForest))
		for i, node := range p.vForest {
			standardLibrary.traversal[i] = libraryTraversal{node.TravId, node.Volume}
		}
		standardLibrary.classes = make([][]libraryGraph, len(p.vClasses))
		for class, members := range p.vClasses {
			graphs := make([]libraryGraph, len(members))
			for i, node := range members {
				graphs[i] = describeLibraryGraph(node)
			}
			standardLibrary.classes[class] = graphs
		}
		initialized = true
	})
	if initialized {
		return
	}
	p.nTravIds = standardLibrary.lastID
	for i, state := range standardLibrary.traversal {
		p.vForest[i].TravId, p.vForest[i].Volume = state.id, state.volume
	}
	p.pMapInv = slices.Clone(standardLibrary.inverse)
	p.vClasses = make([][]*Rwr_Node_t, len(standardLibrary.classes))
	for class, graphs := range standardLibrary.classes {
		if len(graphs) == 0 {
			continue
		}
		members := make([]*Rwr_Node_t, len(graphs))
		for i := range graphs {
			g := &graphs[i]
			node := p.vForest[g.forestID]
			node.graph = g.instantiate()
			members[i] = node
		}
		p.vClasses[class] = members
	}
}

func describeLibraryGraph(node *Rwr_Node_t) libraryGraph {
	g := node.graph
	description := libraryGraph{
		forestID: node.Id, leaves: dec.Dec_GraphLeaveNum(g),
		root: dec.Dec_EdgeToInt(dec.Dec_GraphRootEdge(g)), constant: dec.Dec_GraphIsConst(g) != 0,
		gates: make([]libraryGate, dec.Dec_GraphNodeNum(g)),
	}
	for i := range description.gates {
		n := dec.Dec_GraphNode(g, description.leaves+int32(i))
		a, b := n.Edges()
		f := n.PrintingFlags()
		description.gates[i] = libraryGate{
			dec.Dec_EdgeToInt(a), dec.Dec_EdgeToInt(b), [3]uint8{uint8(f[0]), uint8(f[1]), uint8(f[2])},
		}
	}
	return description
}

func (d *libraryGraph) instantiate() *dec.Dec_Graph_t {
	var g *dec.Dec_Graph_t
	if d.constant {
		g = dec.Dec_GraphCreateConst1()
	} else {
		g = dec.Dec_GraphCreateSized(d.leaves, int32(len(d.gates)))
	}
	for i, gate := range d.gates {
		n := dec.Dec_GraphNode(g, d.leaves+int32(i))
		n.SetEdges(dec.Dec_IntToEdge(gate.input0), dec.Dec_IntToEdge(gate.input1))
		n.SetPrintingFlags([3]uint32{uint32(gate.flags[0]), uint32(gate.flags[1]), uint32(gate.flags[2])})
	}
	dec.Dec_GraphSetRoot(g, dec.Dec_IntToEdge(d.root))
	return g
}
