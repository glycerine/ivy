package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// classPartition stores the candidate classes of fraClass.c. Node IDs in each
// class are in topological order, and the first node is the representative.
// The constant class contains nodes represented by AIG constant one (ID 0).
// Simulation supplies the hash and equality predicates; hash collisions are
// resolved by refinement rather than treated as equivalence.
type classPartition struct {
	Graph       *aig.Graph
	Simulation  *fraSimulation
	VImps       *vec.Vec_Int_t
	FRefinement int32
	// BMC replaces the simulation predicates temporarily during refinement.
	NodeIsConst func(int) bool
	NodesEqual  func(int, int) bool
	Classes     [][]int
	Constants   []int
	Reprs       []int // -1 means no representative
}

func newClassPartition(nodeCount int) *classPartition {
	if nodeCount < 0 {
		panic("fra: negative node count")
	}
	p := &classPartition{Reprs: make([]int, nodeCount)}
	for i := range p.Reprs {
		p.Reprs[i] = -1
	}
	return p
}

// prepareGraph selects the same eligible AIG objects as Fra_ClassesPrepare.
// The latch-correspondence mode includes CIs only; otherwise internal nodes
// may be limited by their structural level, while CIs remain eligible.
func (p *classPartition) PrepareGraph(graph *aig.Graph, latchCorrespondence bool, maxLevel int, hash func(int) uint64, isConst func(int) bool, equal func(int, int) bool) {
	if graph == nil {
		panic("fra: nil AIG")
	}
	nodes := graph.Nodes()
	if len(nodes) > len(p.Reprs) {
		panic("fra: class partition does not cover AIG")
	}
	var candidates []int
	for _, edge := range graph.LiveEdges() {
		id := edge.NodeID()
		node := nodes[id]
		isInput := node.Kind == aig.Input
		isInternal := node.Kind == aig.And || node.Kind == aig.ExclusiveOr
		if latchCorrespondence {
			if !isInput {
				continue
			}
		} else {
			if !isInput && !isInternal {
				continue
			}
			if maxLevel != 0 && node.Level > maxLevel {
				continue
			}
		}
		candidates = append(candidates, id)
	}
	p.Prepare(candidates, hash, isConst, equal)
}

// prepare constructs candidates in node order and separates equal simulation
// signatures, matching Fra_ClassesPrepare followed by Fra_ClassesRefine.
func (p *classPartition) Prepare(nodes []int, hash func(int) uint64, isConst func(int) bool, equal func(int, int) bool) {
	previous := -1
	for _, id := range nodes {
		if id <= previous || id <= 0 || id >= len(p.Reprs) {
			panic("fra: candidate IDs must be valid and in topological order")
		}
		previous = id
	}
	p.Classes, p.Constants = nil, nil
	for i := range p.Reprs {
		p.Reprs[i] = -1
	}
	buckets := make(map[uint64]int)
	var groups [][]int
	for _, id := range nodes {
		if isConst(id) {
			p.Constants = append(p.Constants, id)
			p.Reprs[id] = 0
			continue
		}
		key := hash(id)
		index, ok := buckets[key]
		if !ok {
			index = len(groups)
			buckets[key] = index
			groups = append(groups, nil)
		}
		groups[index] = append(groups[index], id)
	}
	for _, group := range groups {
		if len(group) > 1 {
			p.Classes = append(p.Classes, group)
		}
	}
	p.Refine(equal)
}

// refine only splits candidates: later simulation may disprove a candidate
// equivalence but cannot restore pairs that were eliminated in earlier rounds.
func (p *classPartition) Refine(equal func(int, int) bool) int {
	var result [][]int
	splits := 0
	for _, class := range p.Classes {
		remaining := class
		for len(remaining) > 0 {
			repr := remaining[0]
			group := []int{repr}
			var rest []int
			p.Reprs[repr] = -1
			for _, id := range remaining[1:] {
				if equal(repr, id) {
					group = append(group, id)
					p.Reprs[id] = repr
				} else {
					rest = append(rest, id)
				}
			}
			if len(group) > 1 {
				result = append(result, group)
			}
			if len(rest) > 0 {
				splits++
			}
			remaining = rest
		}
	}
	p.Classes = result
	return splits
}

// refineConstants removes nodes no longer constant and optionally partitions
// them into new candidate classes, as Fra_ClassesRefine1 does.
func (p *classPartition) RefineConstants(isConst func(int) bool, equal func(int, int) bool, refineNew bool) int {
	kept := p.Constants[:0]
	var removed []int
	for _, id := range p.Constants {
		if isConst(id) {
			kept = append(kept, id)
		} else {
			p.Reprs[id] = -1
			removed = append(removed, id)
		}
	}
	p.Constants = kept
	if len(removed) == 0 {
		return 0
	}
	if len(removed) > 1 {
		p.Classes = append(p.Classes, removed)
		for _, id := range removed[1:] {
			p.Reprs[id] = removed[0]
		}
		if refineNew {
			// Existing classes were already refined; split only the new group.
			last := len(p.Classes) - 1
			newPart := &classPartition{Classes: p.Classes[last:], Reprs: p.Reprs}
			splits := newPart.Refine(equal)
			p.Classes = append(p.Classes[:last], newPart.Classes...)
			return 1 + splits
		}
	}
	return 1
}

func (p *classPartition) LiteralCount() int {
	count := len(p.Constants)
	for _, class := range p.Classes {
		count += len(class) - 1
	}
	return count
}

func (p *classPartition) PairCount() int {
	count := 0
	for _, class := range p.Classes {
		count += len(class) * (len(class) - 1) / 2
	}
	return count
}

// latchCorrespondence starts with every register output in the constant class.
// Subsequent sequential simulation refines those provisional candidates.
func (p *classPartition) LatchCorrespondence(graph *aig.Graph) {
	if graph == nil || len(graph.Nodes()) > len(p.Reprs) {
		panic("fra: invalid AIG for class partition")
	}
	p.Classes, p.Constants = nil, nil
	for i := range p.Reprs {
		p.Reprs[i] = -1
	}
	for i := int32(0); i < aig.Aig_ManRegNum(graph); i++ {
		id := aig.Aig_ManLo(graph, i).NodeID()
		p.Constants = append(p.Constants, id)
		p.Reprs[id] = 0
	}
}

// pruneByWeight implements the postprocessing half-maximum weight threshold.
// weights are simulation differences from each node's current representative.
func (p *classPartition) PruneByWeight(weights []int) {
	if len(weights) < len(p.Reprs) {
		panic("fra: missing class weights")
	}
	maximum := 0
	for id, repr := range p.Reprs {
		if repr >= 0 && weights[id] > maximum {
			maximum = weights[id]
		}
	}
	threshold := maximum / 2
	constants := p.Constants[:0]
	for _, id := range p.Constants {
		if weights[id] >= threshold {
			constants = append(constants, id)
		} else {
			p.Reprs[id] = -1
		}
	}
	p.Constants = constants
	var classes [][]int
	for _, group := range p.Classes {
		kept := group[:1]
		for _, id := range group[1:] {
			if weights[id] >= threshold {
				kept = append(kept, id)
			} else {
				p.Reprs[id] = -1
			}
		}
		if len(kept) > 1 {
			classes = append(classes, kept)
		}
	}
	p.Classes = classes
}

// selectRepresentatives chooses the least-support member, breaking ties by
// structural level. Input order breaks remaining ties, as in the C code.
func (p *classPartition) SelectRepresentatives(support, level func(int) int) {
	for _, group := range p.Classes {
		best := 0
		bestSupport, bestLevel := support(group[0]), level(group[0])
		for i := 1; i < len(group); i++ {
			candidateSupport, candidateLevel := support(group[i]), level(group[i])
			if candidateSupport < bestSupport || candidateSupport == bestSupport && candidateLevel < bestLevel {
				best = i
				bestSupport, bestLevel = candidateSupport, candidateLevel
			}
		}
		group[0], group[best] = group[best], group[0]
		p.Reprs[group[0]] = -1
		for _, id := range group[1:] {
			p.Reprs[id] = group[0]
		}
	}
}

// copyRepresentatives installs current candidates on the AIG, except for
// explicitly failed nodes. InvalidEdge denotes no representative; edge zero is
// the valid constant-one representative.
func (p *classPartition) CopyRepresentatives(graph *aig.Graph, failed []int) {
	if graph == nil || len(graph.Nodes()) > len(p.Reprs) {
		panic("fra: invalid AIG for class partition")
	}
	failedSet := make(map[int]bool, len(failed))
	for _, id := range failed {
		if id < 0 || id >= len(p.Reprs) {
			panic("fra: invalid failed node")
		}
		failedSet[id] = true
	}
	for _, edge := range graph.LiveEdges() {
		id := edge.NodeID()
		repr := aig.InvalidEdge
		if p.Reprs[id] >= 0 && !failedSet[id] {
			repr = aig.Edge(p.Reprs[id] << 1)
		}
		aig.Aig_ObjSetRepr(graph, edge, repr)
	}
}

// deriveSequentialAIG implements Fra_ClassesDeriveAig: unroll the transition
// relation for frames+1 frames and emit each candidate-equality miter. Earlier
// frame outputs are assumptions; the final frame outputs are checks.
func (p *classPartition) DeriveSequentialAIG(graph *aig.Graph, frames int) *aig.Graph {
	if graph == nil || frames <= 0 || graph.RegisterCount() <= 0 || int(graph.RegisterCount()) >= graph.InputCount() || len(graph.Nodes()) > len(p.Reprs) {
		panic("fra: invalid sequential class derivation")
	}
	out := aig.NewGraph()
	out.Name, out.Spec = graph.Name, graph.Spec
	nodes := graph.Nodes()
	mapping := make([]aig.Edge, len(nodes))
	for i := range mapping {
		mapping[i] = aig.InvalidEdge
	}
	mapping[0] = out.Const1()
	regs := int(graph.RegisterCount())
	inputs := graph.Inputs()
	for _, edge := range inputs[len(inputs)-regs:] {
		mapping[edge.NodeID()] = out.NewInput()
	}
	deriveCandidate := func(id int) {
		repr := p.Reprs[id]
		if repr < 0 {
			return
		}
		if repr >= id || mapping[id] == aig.InvalidEdge || mapping[repr] == aig.InvalidEdge {
			panic("fra: invalid class representative order")
		}
		a, b := mapping[id], mapping[repr]
		if a.NodeID() == b.NodeID() {
			return
		}
		miter := out.XOR(aig.Aig_Regular(a), aig.Aig_Regular(b))
		if out.Node(miter).Phase != miter.Complemented() {
			miter = miter.Not()
		}
		out.AddOutput(miter.Not())
	}
	for frame := 0; frame <= frames; frame++ {
		for _, edge := range inputs[:len(inputs)-regs] {
			mapping[edge.NodeID()] = out.NewInput()
		}
		for _, edge := range inputs[len(inputs)-regs:] {
			deriveCandidate(edge.NodeID())
		}
		for _, edge := range graph.LiveEdges() {
			id := edge.NodeID()
			node := nodes[id]
			if node.Kind == aig.Buffer {
				fanin := node.Fanins[0]
				value := mapping[fanin.NodeID()]
				if value == aig.InvalidEdge {
					panic("fra: missing mapped buffer fanin")
				}
				if fanin.Complemented() {
					value = value.Not()
				}
				mapping[id] = value
				continue
			}
			if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
				continue
			}
			mappedFanin := func(e aig.Edge) aig.Edge {
				value := mapping[e.NodeID()]
				if value == aig.InvalidEdge {
					panic("fra: missing mapped fanin")
				}
				if e.Complemented() {
					value = value.Not()
				}
				return value
			}
			a, b := mappedFanin(node.Fanins[0]), mappedFanin(node.Fanins[1])
			if node.Kind == aig.And {
				mapping[id] = out.AND(a, b)
			} else {
				mapping[id] = out.XOR(a, b)
			}
			deriveCandidate(id)
		}
		if frame == frames {
			break
		}
		if frame == frames-1 {
			out.Assertions = int32(out.OutputCount())
		}
		latches := make([]aig.Edge, regs)
		for i := 0; i < regs; i++ {
			output := aig.Aig_ManLi(graph, int32(i))
			driver := graph.Node(output).Fanins[0]
			latches[i] = mapping[driver.NodeID()]
			if latches[i] == aig.InvalidEdge {
				panic("fra: missing latch input mapping")
			}
			if driver.Complemented() {
				latches[i] = latches[i].Not()
			}
		}
		for i, edge := range inputs[len(inputs)-regs:] {
			mapping[edge.NodeID()] = latches[i]
		}
	}
	aig.Aig_ManCleanup(out)
	return out
}
