package fra

// Partitioned latch correspondence from fraLcr.c. Per-operation maps replace
// the original temporary AIG object pData and pNext fields.
import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
)

type Fra_Lcr_t struct {
	graph                  *aig.Graph
	classes                *classPartition
	parts                  [][]int32
	inputPart, inputOutput []int
	inputIndices           map[int]int
	fraigs                 []*aig.Graph
	marked                 map[int]bool
	nIters                 int32
}

func Lcr_ManAlloc(g *aig.Graph) *Fra_Lcr_t {
	return &Fra_Lcr_t{graph: g, inputPart: make([]int, g.InputCount()), inputOutput: make([]int, g.InputCount()), inputIndices: lcrInputIndices(g), marked: make(map[int]bool)}
}
func Lcr_ManFree(p *Fra_Lcr_t) {
	if p != nil {
		*p = Fra_Lcr_t{}
	}
}
func Lcr_ManPrint(p *Fra_Lcr_t) {
	fmt.Printf("Iterations = %d. Classes = %d. Constants = %d.\n", p.nIters, len(p.classes.Classes), len(p.classes.Constants))
}

func lcrInputIndices(g *aig.Graph) map[int]int {
	result := make(map[int]int, g.InputCount())
	for i, e := range g.Inputs() {
		result[e.NodeID()] = i
	}
	return result
}
func lcrCopy(src, dst *aig.Graph, copies map[int]aig.Edge, edge aig.Edge) aig.Edge {
	id := edge.NodeID()
	result, ok := copies[id]
	if !ok {
		node := src.Node(edge)
		switch node.Kind {
		case aig.And:
			result = dst.AND(lcrCopy(src, dst, copies, node.Fanins[0]), lcrCopy(src, dst, copies, node.Fanins[1]))
		case aig.ExclusiveOr:
			result = dst.XOR(lcrCopy(src, dst, copies, node.Fanins[0]), lcrCopy(src, dst, copies, node.Fanins[1]))
		case aig.Output, aig.Buffer:
			result = lcrCopy(src, dst, copies, node.Fanins[0])
		default:
			panic("unmapped latch correspondence input")
		}
		copies[id] = result
	}
	if edge.Complemented() {
		return result.Not()
	}
	return result
}
func lcrStartCopies(src *aig.Graph) (*aig.Graph, map[int]aig.Edge) {
	dst := aig.NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	copies := map[int]aig.Edge{0: dst.Const1()}
	for _, in := range src.Inputs() {
		image := dst.NewInput()
		aig.Aig_ObjSetLevel(dst, image, int32(src.Node(in).Level))
		copies[in.NodeID()] = image
	}
	return dst, copies
}
func (p *Fra_Lcr_t) outputForInput(id int) int {
	index, ok := p.inputIndices[id]
	if !ok {
		panic("latch class contains non-input")
	}
	out := p.graph.OutputCount() - p.graph.InputCount() + index
	if out < 0 || out >= p.graph.OutputCount() {
		panic("latch class includes primary input")
	}
	return out
}

// Fra_LcrDeriveAigForPartitioning produces one structural output per class,
// followed by one for each constant candidate.
func Fra_LcrDeriveAigForPartitioning(p *Fra_Lcr_t) *aig.Graph {
	dst, copies := lcrStartCopies(p.graph)
	outputs := p.graph.Outputs()
	for _, class := range p.classes.Classes {
		result := dst.Const0()
		for _, id := range class {
			result = dst.XOR(result, lcrCopy(p.graph, dst, copies, p.graph.Node(outputs[p.outputForInput(id)]).Fanins[0]))
		}
		dst.AddOutput(result)
	}
	for _, id := range p.classes.Constants {
		dst.AddOutput(lcrCopy(p.graph, dst, copies, p.graph.Node(outputs[p.outputForInput(id)]).Fanins[0]))
	}
	return dst
}
func Fra_LcrRemapPartitions(p *Fra_Lcr_t, parts [][]int32) [][]int32 {
	indices := p.inputIndices
	out := make([][]int32, len(parts))
	for part, group := range parts {
		for _, index := range group {
			var ids []int
			if int(index) < len(p.classes.Classes) {
				ids = p.classes.Classes[index]
			} else {
				offset := int(index) - len(p.classes.Classes)
				if offset < 0 || offset >= len(p.classes.Constants) {
					panic("invalid latch partition")
				}
				ids = []int{p.classes.Constants[offset]}
			}
			for _, id := range ids {
				in := indices[id]
				p.inputPart[in] = part
				p.inputOutput[in] = len(out[part])
				out[part] = append(out[part], int32(p.outputForInput(id)))
			}
		}
	}
	return out
}
func Fra_ClassNodesMark(p *Fra_Lcr_t) {
	clear(p.marked)
	for _, class := range p.classes.Classes {
		for _, id := range class {
			p.marked[p.outputForInput(id)] = true
		}
	}
	for _, id := range p.classes.Constants {
		p.marked[p.outputForInput(id)] = true
	}
}
func Fra_ClassNodesUnmark(p *Fra_Lcr_t) { clear(p.marked) }
func Fra_LcrCreatePart(p *Fra_Lcr_t, outputs []int32) *aig.Graph {
	src := p.graph
	dst, _ := lcrStartCopies(src)
	copies := map[int]aig.Edge{0: dst.Const1()}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(edge aig.Edge) aig.Edge {
		id := edge.NodeID()
		result, ok := copies[id]
		if !ok {
			n := src.Node(edge)
			switch n.Kind {
			case aig.Input:
				repr := p.classes.Reprs[id]
				if repr < 0 {
					result = dst.NewInput()
				} else {
					result = copyEdge(aig.Edge(repr << 1))
					if src.Node(aig.Edge(repr<<1)).Phase != n.Phase {
						result = result.Not()
					}
				}
			case aig.And, aig.ExclusiveOr:
				// Fra_LcrCreatePart_rec uses Aig_And for every source node.
				result = dst.AND(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
			case aig.Buffer, aig.Output:
				result = copyEdge(n.Fanins[0])
			default:
				panic("invalid latch partition node")
			}
			copies[id] = result
		}
		if edge.Complemented() {
			return result.Not()
		}
		return result
	}
	for _, index := range outputs {
		if index < 0 || int(index) >= src.OutputCount() {
			panic("invalid latch output")
		}
		edge := dst.Const1()
		if p.marked[int(index)] {
			edge = copyEdge(src.Node(src.Outputs()[index]).Fanins[0])
		}
		dst.AddOutput(edge)
	}
	return dst
}
func (p *Fra_Lcr_t) outputForNode(id int) (aig.Edge, *aig.Graph) {
	index, ok := p.inputIndices[id]
	if !ok {
		panic("invalid latch input")
	}
	part := p.inputPart[index]
	g := p.fraigs[part]
	return g.Node(g.Outputs()[p.inputOutput[index]]).Fanins[0], g
}
func (p *Fra_Lcr_t) nodesEqual(a, b int) bool {
	ia, ib := p.inputIndices[a], p.inputIndices[b]
	if p.inputPart[ia] != p.inputPart[ib] {
		panic("equivalent latch candidates in different partitions")
	}
	x, _ := p.outputForNode(a)
	y, _ := p.outputForNode(b)
	return x.NodeID() == y.NodeID()
}
func (p *Fra_Lcr_t) nodeIsConst(id int) bool { e, _ := p.outputForNode(id); return e.NodeID() == 0 }
func (p *Fra_Lcr_t) repartition() {
	graph := Fra_LcrDeriveAigForPartitioning(p)
	p.parts = Fra_LcrRemapPartitions(p, aig.Aig_ManPartitionSmart(graph, 200, 0, nil))
}

// Fra_FraigLatchCorrespondence performs simulation, partitioned FRAIG, and
// class refinement as in the original latch-correspondence procedure.
func Fra_FraigLatchCorrespondence(graph *aig.Graph, prefix, conflicts, prove, verbose int32, iterations *int32, timeLimit float32) *aig.Graph {
	if graph == nil {
		panic("nil latch correspondence graph")
	}
	if aig.Aig_ManNodeNum(graph) == 0 {
		if iterations != nil {
			*iterations = 0
		}
		aig.Aig_ManReprStart(graph, int32(len(graph.Nodes())))
		return aig.Aig_ManDupOrdered(graph)
	}
	if graph.RegisterCount() <= 0 {
		panic("latch correspondence requires registers")
	}
	var deadline int64
	if timeLimit != 0 {
		defer fraLcrPin()()
		deadline = fraLcrClock() + int64(float64(timeLimit)*fraLcrTicksPerSecond)
	}
	sim := Fra_SmlSimulateSeq(graph, prefix, 32, 1, 1)
	if prove != 0 && sim.NonconstantOutput {
		if cex := Fra_SmlGetCounterExample(sim); cex != nil {
			_ = graph.SetSequentialCounterexample(cex)
		}
		return nil
	}
	p := Lcr_ManAlloc(graph)
	p.classes = Fra_ClassesStart(graph)
	p.classes.Simulation = sim
	Fra_ClassesPrepare(p.classes, 1, 0)
	p.classes.NodeIsConst = p.nodeIsConst
	p.classes.NodesEqual = p.nodesEqual
	p.repartition()
	for {
		Fra_ClassNodesMark(p)
		p.fraigs = nil
		for _, part := range p.parts {
			if timeLimit != 0 && fraLcrClock() > deadline {
				if iterations != nil {
					*iterations = p.nIters
				}
				fmt.Println("Fra_FraigLatchCorrespondence: runtime limit exceeded.")
				return nil
			}
			fraig := Fra_FraigEquivence(Fra_LcrCreatePart(p, part), conflicts, 0)
			if fraig == nil {
				if iterations != nil {
					*iterations = p.nIters
				}
				return nil
			}
			p.fraigs = append(p.fraigs, fraig)
		}
		Fra_ClassNodesUnmark(p)
		changed := Fra_ClassesRefine(p.classes) + Fra_ClassesRefine1(p.classes, 0, nil)
		p.nIters++
		if verbose != 0 {
			fmt.Printf("%3d : Const = %d. Class = %d. Part = %d.\n", p.nIters-1, len(p.classes.Constants), len(p.classes.Classes), len(p.parts))
		}
		p.repartition()
		if changed == 0 {
			break
		}
	}
	Fra_ClassesCopyReprs(p.classes, nil)
	result := aig.Aig_ManDupRepr(graph, 0)
	aig.Aig_ManSeqCleanup(result)
	if iterations != nil {
		*iterations = p.nIters
	}
	return result
}
