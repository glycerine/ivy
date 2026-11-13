package pdr

// Flop ordering from pdrMan.c, Alan Mishchenko (UC Berkeley, 2010).
import (
	"sort"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// pdrFlopCosts counts each register output's fanouts, and optionally gives
// MUX controls an additional weight greater than every unweighted count.
func pdrFlopCosts(graph *gia.Gia_Man_t, muxControls bool) []int32 {
	gia.Gia_ManCreateRefs(graph)
	defer func() { graph.Refs = nil }()
	costs := make([]int32, gia.Gia_ManRegNum(graph))
	var maximum int32
	for i := range costs {
		costs[i] = gia.Gia_ObjRefNum(graph, gia.Gia_ManRo(graph, int32(i)))
		if costs[i] > maximum {
			maximum = costs[i]
		}
	}
	if !muxControls {
		return costs
	}
	controls := make([]bool, gia.Gia_ManObjNum(graph))
	for id := int32(1); id < gia.Gia_ManObjNum(graph); id++ {
		obj := gia.Gia_ManObj(graph, id)
		if gia.Gia_ObjIsAnd(obj) == 0 || gia.Gia_ObjIsMuxType(obj) == 0 {
			continue
		}
		var then, otherwise *gia.Gia_Obj_t
		control := gia.Gia_Regular(gia.Gia_ObjRecognizeMux(obj, &then, &otherwise))
		controls[gia.Gia_ObjId(graph, control)] = true
	}
	for i := range costs {
		if controls[gia.Gia_ObjId(graph, gia.Gia_ManRo(graph, int32(i)))] {
			costs[i] += maximum + 1
		}
	}
	return costs
}

// Pdr_ManDeriveFlopPriorities2 assigns each flop its rank in ABC's cost sort.
func Pdr_ManDeriveFlopPriorities2(graph *gia.Gia_Man_t, muxControls int32) *vec.Vec_Int_t {
	costs := pdrFlopCosts(graph, muxControls != 0)
	permutation := util.Abc_MergeSortCost(costs, int32(len(costs)))
	ranks := make([]int32, len(costs))
	for rank, index := range permutation {
		ranks[index] = int32(rank)
	}
	return vec.Vec_IntAllocArrayCopy(ranks, len(ranks))
}

// Pdr_ManDeriveFlopPriorities3 ranks equal-cost flops by their original order,
// matching the source's levelized buckets.
func Pdr_ManDeriveFlopPriorities3(graph *gia.Gia_Man_t, muxControls int32) *vec.Vec_Int_t {
	costs := pdrFlopCosts(graph, muxControls != 0)
	order := make([]int, len(costs))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool { return costs[order[i]] < costs[order[j]] })
	ranks := make([]int32, len(costs))
	for rank, index := range order {
		ranks[index] = int32(rank)
	}
	return vec.Vec_IntAllocArrayCopy(ranks, len(ranks))
}
