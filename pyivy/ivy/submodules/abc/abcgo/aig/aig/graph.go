package aig

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// Edge identifies a node and its complemented phase. Node zero is constant one,
// matching ABC's AIG convention; toggling bit zero complements the function.
type Edge uint32

const InvalidEdge Edge = ^Edge(0)

func (e Edge) Not() Edge          { return e ^ 1 }
func (e Edge) NodeID() int        { return int(e >> 1) }
func (e Edge) Complemented() bool { return e&1 != 0 }

type NodeKind uint8

const (
	Constant NodeKind = iota
	Input
	And
	Buffer
	ExclusiveOr
	Output
)

type Node struct {
	Kind   NodeKind
	Fanins [2]Edge
	Level  int
	Refs   int
	MarkA  bool
	MarkB  bool
	TravID int32
	CioID  int32 // C stores this in the terminal pNext/CioId union.
	Xsim   int32
	Phase  bool
}
type edgePair struct{ a, b Edge }

// Graph owns an AIG with stable node IDs. Structural hashing shares equivalent
// gates; complemented edges represent inversion directly. Deletion leaves holes
// so existing edges remain valid, and rewriting may change topological ID order.
type Graph struct {
	// Time2Quit is an absolute deadline for source-style AIG optimization.
	// A zero value means no deadline.
	Time2Quit time.Time
	// InputNumbers maps retained CIs to their original interface positions.
	InputNumbers                      []int32
	combinationalCex                  *util.Abc_Cex_t
	seqModelVec                       *vec.Vec_Ptr_t
	PartitionInputs, PartitionOutputs []int32
	ClockDomains                      [][]int32
	// OneHotGroups contains register indices known to be mutually exclusive.
	OneHotGroups [][]int32
	// OneHotInputGroups contains projected partition input indices.
	OneHotInputGroups                           [][]int32
	FlopReprs                                   []int32
	cutManager                                  *Aig_ManCut_t
	reverseLevels                               []int32
	maxRequiredLevel                            int32
	order                                       []orderLinks
	iPrev, iNext, nAndTotal, nAndPrev           int32
	objectData                                  map[int]any
	cexValues                                   [][]bool
	inductionCex                                *util.Abc_Cex_t
	editDepth                                   int
	fanouts                                     *FanoutIndex
	Assertions, BarrierBuffers                  int32
	RetimingChanged                             bool
	FlopNumbers                                 []int32
	Timing, History                             any
	Name, Spec                                  string
	Data                                        any
	nodes                                       []Node
	inputs                                      []Edge
	outputs                                     []Edge
	hash                                        map[edgePair]Edge
	xorHash                                     map[edgePair]Edge
	nRegs, nTruePis, nTruePos, nConstrs, travID int32
	ghost                                       Node
	copies, equivs, reprs, next                 map[int]Edge
	deleted                                     map[int]bool
}

func NewGraph() *Graph {
	return &Graph{nodes: []Node{{Kind: Constant, Phase: true}}, hash: make(map[edgePair]Edge), deleted: make(map[int]bool)}
}
func (g *Graph) Const1() Edge         { return 0 }
func (g *Graph) Const0() Edge         { return 1 }
func (g *Graph) InputCount() int      { return len(g.inputs) }
func (g *Graph) OutputCount() int     { return len(g.outputs) }
func (g *Graph) NodeCount() int       { return len(g.nodes) - len(g.deleted) }
func (g *Graph) RegisterCount() int32 { return g.nRegs }

func (g *Graph) updateTrueTerminalCounts() {
	g.nTruePis = int32(len(g.inputs)) - g.nRegs
	g.nTruePos = int32(len(g.outputs)) - g.nRegs
	Aig_ManSetCioIds(g)
}

// TruePiCount is the explicitly recorded C nTruePis metadata.
func (g *Graph) TruePiCount() int32 { return g.nTruePis }

// TruePoCount is the explicitly recorded C nTruePos metadata.
func (g *Graph) TruePoCount() int32 { return g.nTruePos }

// SetTruePiCount records source-provided true-input metadata.
func (g *Graph) SetTruePiCount(n int32) { g.nTruePis = n }

// SetTruePoCount records source-provided true-output metadata.
func (g *Graph) SetTruePoCount(n int32) { g.nTruePos = n }

// CopyTruePiCount preserves the source metadata in source-style transformations.
func (g *Graph) CopyTruePiCount(src *Graph) { g.nTruePis = src.nTruePis }

// CopyTruePoCount preserves the source output metadata.
func (g *Graph) CopyTruePoCount(src *Graph) { g.nTruePos = src.nTruePos }
func (g *Graph) Nodes() []Node              { return append([]Node(nil), g.nodes...) }
func (g *Graph) Inputs() []Edge             { return append([]Edge(nil), g.inputs...) }
func (g *Graph) Outputs() []Edge            { return append([]Edge(nil), g.outputs...) }
func (g *Graph) valid(e Edge) bool          { return e.NodeID() < len(g.nodes) && !g.deleted[e.NodeID()] }
func (g *Graph) NewInput() Edge {
	e := Edge(len(g.nodes)) << 1
	g.nodes = append(g.nodes, Node{Kind: Input})
	g.inputs = append(g.inputs, e)
	g.refreshFanouts()
	return e
}
func (g *Graph) AND(a, b Edge) Edge {
	if !g.valid(a) || !g.valid(b) {
		panic("aig: edge outside graph")
	}
	if a == 1 || b == 1 {
		return 1
	}
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	if a == b {
		return a
	}
	if a == b.Not() {
		return 1
	}
	if a > b {
		a, b = b, a
	}
	key := edgePair{a, b}
	if e, ok := g.hash[key]; ok {
		return e
	}
	level := g.nodes[a.NodeID()].Level
	if l := g.nodes[b.NodeID()].Level; l > level {
		level = l
	}
	e := Edge(len(g.nodes)) << 1
	g.nodes = append(g.nodes, Node{Kind: And, Fanins: [2]Edge{a, b}, Level: level + 1, Phase: (g.nodes[a.NodeID()].Phase != a.Complemented()) && (g.nodes[b.NodeID()].Phase != b.Complemented())})
	g.nodes[a.NodeID()].Refs++
	g.nodes[b.NodeID()].Refs++
	g.hash[key] = e
	g.refreshFanouts()
	return e
}
func (g *Graph) OR(a, b Edge) Edge  { return g.AND(a.Not(), b.Not()).Not() }
func (g *Graph) XOR(a, b Edge) Edge { return g.OR(g.AND(a, b.Not()), g.AND(a.Not(), b)) }
func (g *Graph) Mux(control, yes, no Edge) Edge {
	return g.OR(g.AND(control, yes), g.AND(control.Not(), no))
}

// AddOutput creates an explicit output object, so every structural reference
// (including output references) has an enumerable consumer node.
func (g *Graph) AddOutput(driver Edge) Edge {
	if !g.valid(driver) {
		panic("aig: output outside graph")
	}
	if g.node(driver).Kind == Output {
		driver = g.outputDriver(driver)
	}
	e := Edge(len(g.nodes)) << 1
	n := Node{Kind: Output, Fanins: [2]Edge{driver, InvalidEdge}, Level: g.node(driver).Level, Phase: Aig_ObjPhaseReal(g, driver), Refs: 1}
	g.nodes = append(g.nodes, n)
	g.outputs = append(g.outputs, e)
	g.node(driver).Refs++
	g.refreshFanouts()
	return e
}

func (g *Graph) Evaluate(inputs []bool) ([]bool, error) {
	if len(inputs) != len(g.inputs) {
		return nil, fmt.Errorf("aig: got %d inputs, need %d", len(inputs), len(g.inputs))
	}
	values := make([]bool, len(g.nodes))
	done := make([]bool, len(g.nodes))
	active := make([]bool, len(g.nodes))
	for i, e := range g.inputs {
		values[e.NodeID()] = inputs[i]
		done[e.NodeID()] = true
	}
	var visit func(Edge) (bool, error)
	visit = func(e Edge) (bool, error) {
		if !g.valid(e) {
			return false, fmt.Errorf("aig: invalid edge")
		}
		id := e.NodeID()
		if done[id] {
			return values[id] != e.Complemented(), nil
		}
		if active[id] {
			return false, fmt.Errorf("aig: cycle at %d", id)
		}
		active[id] = true
		n := g.nodes[id]
		v := true
		switch n.Kind {
		case Constant:
			v = true
		case Input:
			return false, fmt.Errorf("aig: unbound input %d", id)
		case And, ExclusiveOr, Buffer, Output:
			a, err := visit(n.Fanins[0])
			if err != nil {
				return false, err
			}
			v = a
			if n.Kind == And || n.Kind == ExclusiveOr {
				b, err := visit(n.Fanins[1])
				if err != nil {
					return false, err
				}
				if n.Kind == And {
					v = a && b
				} else {
					v = a != b
				}
			}
		default:
			return false, fmt.Errorf("aig: unknown node kind")
		}
		active[id] = false
		done[id] = true
		values[id] = v
		return v != e.Complemented(), nil
	}
	out := make([]bool, len(g.outputs))
	for i, e := range g.outputs {
		v, err := visit(e)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

// Check verifies topology, levels, references and the structural hash invariant.
func (g *Graph) Check() error {
	if g.nRegs < 0 || int(g.nRegs) > len(g.inputs) || int(g.nRegs) > len(g.outputs) {
		return fmt.Errorf("aig: invalid register terminal count")
	}
	if g.nConstrs < 0 || int(g.nConstrs) > len(g.outputs)-int(g.nRegs) {
		return fmt.Errorf("aig: invalid constraint terminal count")
	}
	if len(g.nodes) == 0 || g.nodes[0].Kind != Constant {
		return fmt.Errorf("aig: missing constant node")
	}
	refs := make([]int, len(g.nodes))
	andSeen := map[edgePair]Edge{}
	xorSeen := map[edgePair]Edge{}
	for id, n := range g.nodes {
		if g.deleted[id] {
			continue
		}
		if n.Kind > Output {
			return fmt.Errorf("aig: unknown node kind at %d", id)
		}
		if id > 0 && n.Kind == Constant {
			return fmt.Errorf("aig: extra constant at %d", id)
		}
		if n.Kind == Constant || n.Kind == Input {
			continue
		}
		count := 1
		if n.Kind == And || n.Kind == ExclusiveOr {
			count = 2
		}
		level := 0
		for _, e := range n.Fanins[:count] {
			if !g.valid(e) {
				return fmt.Errorf("aig: invalid fanin at %d", id)
			}
			refs[e.NodeID()]++
			if l := g.node(e).Level; l > level {
				level = l
			}
		}
		if count == 2 {
			level++
		}
		if n.Level != level {
			// Choice classes can use ordinary structural levels until the
			// explicit choice-level pass annotates their maximum depth.
			choiceLevel := level
			if other, ok := g.equivs[id]; ok && g.valid(other) && g.node(other).Level > choiceLevel {
				choiceLevel = g.node(other).Level
			}
			if n.Level != choiceLevel {
				return fmt.Errorf("aig: invalid level at %d", id)
			}
		}
		if count == 2 {
			a, b := n.Fanins[0], n.Fanins[1]
			if a >= b {
				return fmt.Errorf("aig: unordered fanins at %d", id)
			}
			key := edgePair{a, b}
			seen, table := andSeen, g.hash
			if n.Kind == ExclusiveOr {
				seen, table = xorSeen, g.xorHash
			}
			if _, ok := seen[key]; ok {
				return fmt.Errorf("aig: duplicate node at %d", id)
			}
			seen[key] = Edge(id) << 1
			if table[key] != Edge(id)<<1 {
				return fmt.Errorf("aig: missing hash entry at %d", id)
			}
		}
	}
	for _, e := range g.outputs {
		if !g.valid(e) {
			return fmt.Errorf("aig: invalid output")
		}
		refs[e.NodeID()]++
	}
	for id, n := range g.nodes {
		if n.Refs != refs[id] {
			return fmt.Errorf("aig: reference count at %d: %d != %d", id, n.Refs, refs[id])
		}
	}
	if len(andSeen) != len(g.hash) || len(xorSeen) != len(g.xorHash) {
		return fmt.Errorf("aig: stale hash entries")
	}
	// Validate cycles in unreachable nodes too; levels alone do not prove DAGness.
	state := make([]uint8, len(g.nodes))
	var visit func(int) error
	visit = func(id int) error {
		if state[id] == 2 {
			return nil
		}
		if state[id] == 1 {
			return fmt.Errorf("aig: cycle at %d", id)
		}
		state[id] = 1
		n := g.nodes[id]
		if n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer || n.Kind == Output {
			if err := visit(n.Fanins[0].NodeID()); err != nil {
				return err
			}
			if n.Kind == And || n.Kind == ExclusiveOr {
				if err := visit(n.Fanins[1].NodeID()); err != nil {
					return err
				}
			}
		}
		state[id] = 2
		return nil
	}
	for id := range g.nodes {
		if !g.deleted[id] {
			if err := visit(id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Graph) node(e Edge) *Node {
	if !g.valid(e) {
		panic("aig: invalid edge")
	}
	return &g.nodes[e.NodeID()]
}

// NewXor retains an explicit XOR node, as requested by ABC's XOR-capture mode.
func (g *Graph) NewXor(a, b Edge) Edge {
	if !g.valid(a) || !g.valid(b) {
		panic("aig: invalid XOR edge")
	}
	if a == b {
		return g.Const0()
	}
	if a == b.Not() {
		return g.Const1()
	}
	if a.NodeID() == 0 {
		if a == g.Const1() {
			return b.Not()
		}
		return b
	}
	if b.NodeID() == 0 {
		if b == g.Const1() {
			return a.Not()
		}
		return a
	}
	return Aig_ObjCreate(g, &Node{Kind: ExclusiveOr, Fanins: [2]Edge{a, b}})
}

// LiveEdges enumerates regular live objects without exposing deleted slots.
func (g *Graph) LiveEdges() []Edge {
	edges := make([]Edge, 0, g.NodeCount())
	for id := range g.nodes {
		if !g.deleted[id] {
			edges = append(edges, Edge(id)<<1)
		}
	}
	return edges
}

// TopologicalObjects returns every live object, including dangling logic, with
// fanins before users. Rewrites may make object IDs non-topological. This is
// the Go counterpart of an all-object pass that needs dependency order; unlike
// source Aig_ManDfsAll it includes unreachable gates and supports buffers.
// It does not change traversal IDs, object marks or the graph.
func (g *Graph) TopologicalObjects() []Edge {
	return g.dfs(g.LiveEdges(), false, false, false)
}

func (g *Graph) ValidEdge(e Edge) bool { return g.valid(e) }

// Node returns a read-only snapshot of one live object's payload in O(1).
func (g *Graph) Node(e Edge) Node { return *g.node(e) }
