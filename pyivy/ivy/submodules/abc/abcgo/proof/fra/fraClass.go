package fra

// Equivalence class operations from ABC's fraClass.c. Simulation belongs to
// the class manager explicitly; in C it was obtained through object pData.
import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Fra_ClassesStart(graph *aig.Graph) *Fra_Cla_t {
	p := newClassPartition(len(graph.Nodes()))
	p.graph = graph
	return p
}

func Fra_ClassesStop(_ *Fra_Cla_t) {}

func Fra_ClassesCopyReprs(p *Fra_Cla_t, failed *vec.Vec_Ptr_t) {
	var ids []int
	if failed != nil {
		for i := 0; i < vec.Vec_PtrSize(failed); i++ {
			edge, ok := vec.Vec_PtrEntry(failed, i).(aig.Edge)
			if !ok {
				panic("fra: failed object is not an AIG edge")
			}
			ids = append(ids, edge.NodeID())
		}
	}
	p.copyRepresentatives(p.graph, ids)
}

func Fra_ClassCount(class []int) int32         { return int32(len(class)) }
func Fra_ClassesCountLits(p *Fra_Cla_t) int32  { return int32(p.literalCount()) }
func Fra_ClassesCountPairs(p *Fra_Cla_t) int32 { return int32(p.pairCount()) }

func Fra_ClassesPrint(p *Fra_Cla_t, veryVerbose int32) {
	fmt.Printf("Const = %5d. Class = %5d. Lit = %5d. ", len(p.constants), len(p.classes), p.literalCount())
	if p.vImps != nil && vec.Vec_IntSize(p.vImps) > 0 {
		fmt.Printf("Imp = %5d. ", vec.Vec_IntSize(p.vImps))
	}
	fmt.Println()
	if veryVerbose != 0 {
		fmt.Print("Constants { ")
		for _, id := range p.constants {
			printClassNode(p.graph, id)
		}
		fmt.Println("}")
		for i, class := range p.classes {
			fmt.Printf("%3d (%3d) : { ", i, len(class))
			for _, id := range class {
				printClassNode(p.graph, id)
			}
			fmt.Println("}")
		}
		fmt.Println()
	}
}

func printClassNode(graph *aig.Graph, id int) {
	edge := aig.Edge(id << 1)
	fmt.Printf("%d(%d,%d) ", id, graph.Node(edge).Level, aig.Aig_SupportSize(graph, edge))
}

func (p *classPartition) requireSimulation() *fraSimulation {
	if p.simulation == nil || p.simulation.Graph != p.graph {
		panic("fra: class simulation is not attached to its AIG")
	}
	return p.simulation
}

func Fra_ClassesPrepare(p *Fra_Cla_t, latchCorr, maxLevels int32) {
	sim := p.requireSimulation()
	p.prepareGraph(p.graph, latchCorr != 0, int(maxLevels), sim.Hash, sim.IsConstant, sim.Equal)
}

func Fra_ClassesRefine(p *Fra_Cla_t) int32 {
	_, equal := p.refinementPredicates()
	return int32(p.refine(equal))
}

func Fra_ClassesRefine1(p *Fra_Cla_t, refineNew int32, skipped *int32) int32 {
	isConst, equal := p.refinementPredicates()
	before := len(p.constants)
	result := p.refineConstants(isConst, equal, refineNew != 0)
	if skipped != nil && refineNew == 0 && before-len(p.constants) > 1 {
		*skipped++
	}
	return int32(result)
}

func (p *classPartition) refinementPredicates() (func(int) bool, func(int, int) bool) {
	if p.nodeIsConst != nil && p.nodesEqual != nil {
		return p.nodeIsConst, p.nodesEqual
	}
	sim := p.requireSimulation()
	return sim.IsConstant, sim.Equal
}

func Fra_ClassesTest(p *Fra_Cla_t, id1, id2 int32) {
	if id1 < 0 || id1 >= id2 || int(id2) >= len(p.reprs) {
		panic("fra: invalid class test node IDs")
	}
	p.classes = append(p.classes, []int{int(id1), int(id2)})
	p.reprs[id2] = int(id1)
}

func Fra_ClassesLatchCorr(p *Fra_Man_t) { p.pCla.latchCorrespondence(p.pManAig) }

func Fra_ClassesPostprocess(p *Fra_Cla_t) {
	sim := newFraSimulation(p.graph, 0, 1, 32)
	sim.Initialize(false)
	sim.SimulateOne()
	weights := make([]int, len(p.reprs))
	for id, repr := range p.reprs {
		if repr >= 0 {
			weights[id] = sim.DifferenceWeight(repr, id)
		}
	}
	fmt.Printf("Before: Const = %6d. Class = %6d.  ", len(p.constants), len(p.classes))
	p.pruneByWeight(weights)
	fmt.Printf("After: Const = %6d. Class = %6d. \n", len(p.constants), len(p.classes))
}

func Fra_ClassesSelectRepr(p *Fra_Cla_t) {
	p.selectRepresentatives(func(id int) int { return int(aig.Aig_SupportSize(p.graph, aig.Edge(id<<1))) },
		func(id int) int { return p.graph.Node(aig.Edge(id << 1)).Level })
}

func Fra_ClassesDeriveAig(p *Fra_Cla_t, frames int32) *aig.Graph {
	return p.deriveSequentialAIG(p.graph, int(frames))
}
