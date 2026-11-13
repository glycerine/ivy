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
	p.Graph = graph
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
	p.CopyRepresentatives(p.Graph, ids)
}

func Fra_ClassCount(class []int) int32         { return int32(len(class)) }
func Fra_ClassesCountLits(p *Fra_Cla_t) int32  { return int32(p.LiteralCount()) }
func Fra_ClassesCountPairs(p *Fra_Cla_t) int32 { return int32(p.PairCount()) }

func Fra_ClassesPrint(p *Fra_Cla_t, veryVerbose int32) {
	fmt.Printf("Const = %5d. Class = %5d. Lit = %5d. ", len(p.Constants), len(p.Classes), p.LiteralCount())
	if p.VImps != nil && vec.Vec_IntSize(p.VImps) > 0 {
		fmt.Printf("Imp = %5d. ", vec.Vec_IntSize(p.VImps))
	}
	fmt.Println()
	if veryVerbose != 0 {
		fmt.Print("Constants { ")
		for _, id := range p.Constants {
			printClassNode(p.Graph, id)
		}
		fmt.Println("}")
		for i, class := range p.Classes {
			fmt.Printf("%3d (%3d) : { ", i, len(class))
			for _, id := range class {
				printClassNode(p.Graph, id)
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

func Fra_ClassesPrepare(p *Fra_Cla_t, latchCorr, maxLevels int32) {
	sim := p.RequireSimulation()
	p.PrepareGraph(p.Graph, latchCorr != 0, int(maxLevels), sim.Hash, sim.IsConstant, sim.Equal)
}

func Fra_ClassesRefine(p *Fra_Cla_t) int32 {
	_, equal := p.RefinementPredicates()
	return int32(p.Refine(equal))
}

func Fra_ClassesRefine1(p *Fra_Cla_t, refineNew int32, skipped *int32) int32 {
	isConst, equal := p.RefinementPredicates()
	before := len(p.Constants)
	result := p.RefineConstants(isConst, equal, refineNew != 0)
	if skipped != nil && refineNew == 0 && before-len(p.Constants) > 1 {
		*skipped++
	}
	return int32(result)
}

func Fra_ClassesTest(p *Fra_Cla_t, id1, id2 int32) {
	if id1 < 0 || id1 >= id2 || int(id2) >= len(p.Reprs) {
		panic("fra: invalid class test node IDs")
	}
	p.Classes = append(p.Classes, []int{int(id1), int(id2)})
	p.Reprs[id2] = int(id1)
}

func Fra_ClassesLatchCorr(p *Fra_Man_t) { p.PCla.LatchCorrespondence(p.PManAig) }

func Fra_ClassesPostprocess(p *Fra_Cla_t) {
	sim := newFraSimulation(p.Graph, 0, 1, 32)
	sim.Initialize(false)
	sim.SimulateOne()
	weights := make([]int, len(p.Reprs))
	for id, repr := range p.Reprs {
		if repr >= 0 {
			weights[id] = sim.DifferenceWeight(repr, id)
		}
	}
	fmt.Printf("Before: Const = %6d. Class = %6d.  ", len(p.Constants), len(p.Classes))
	p.PruneByWeight(weights)
	fmt.Printf("After: Const = %6d. Class = %6d. \n", len(p.Constants), len(p.Classes))
}

func Fra_ClassesSelectRepr(p *Fra_Cla_t) {
	p.SelectRepresentatives(func(id int) int { return int(aig.Aig_SupportSize(p.Graph, aig.Edge(id<<1))) },
		func(id int) int { return p.Graph.Node(aig.Edge(id << 1)).Level })
}

func Fra_ClassesDeriveAig(p *Fra_Cla_t, frames int32) *aig.Graph {
	return p.DeriveSequentialAIG(p.Graph, int(frames))
}
