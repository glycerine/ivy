package llb

import (
	"fmt"
	"sort"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Llb_ManMaxFanoutCi returns the CI index with the greatest reference count.
func Llb_ManMaxFanoutCi(graph *Aig_Man_t) int32 {
	var bestIndex int32 = -1
	var bestRefs int32 = -1
	for i := int32(0); i < aig.Aig_ManCiNum(graph); i++ {
		edge := aig.Aig_ManCi(graph, i)
		refs := aig.Aig_ObjRefs(graph, edge)
		if refs > bestRefs {
			bestIndex, bestRefs = i, refs
		}
	}
	assert(bestIndex >= 0, "AIG has at least one CI")
	return bestIndex
}

func llbIsLo(graph *aig.Graph, edge aig.Edge) bool {
	if !aig.Aig_ObjIsCi(graph, edge) {
		return false
	}
	piCount := aig.Aig_ManCiNum(graph) - aig.Aig_ManRegNum(graph)
	return aig.Aig_ObjCioId(graph, edge) >= piCount
}

// Llb_ManPerformHints successively fixes the highest-fanout CI to one.
func Llb_ManPerformHints(graph *Aig_Man_t, hintDepth int32) *Aig_Man_t {
	if graph == nil || hintDepth <= 0 {
		return aig.Aig_ManDupDfs(graph)
	}
	result := aig.Aig_ManDupDfs(graph)
	for i := int32(0); i < hintDepth && aig.Aig_ManCiNum(result) > 0; i++ {
		input := Llb_ManMaxFanoutCi(result)
		fmt.Printf("%d %3d\n", i, input)
		next := aig.Aig_ManDupCof(result, input, 1)
		aig.Aig_ManStop(result)
		result = next
	}
	return result
}

// Llb_ManCollectHighFanoutObjects selects high-reference LOs and, unless
// fCisOnly is set, internal logic nodes. Results are sorted by descending refs.
func Llb_ManCollectHighFanoutObjects(graph *Aig_Man_t, maxCandidates, cisOnly int32) *Vec_Int_t {
	if graph == nil || maxCandidates <= 0 {
		return vec.Vec_IntAlloc(0)
	}
	eligible := make([]aig.Edge, 0)
	refs := make([]int32, 0)
	for _, edge := range graph.LiveEdges() {
		isLatchOutput := llbIsLo(graph, edge)
		if !isLatchOutput && (cisOnly != 0 || !aig.Aig_ObjIsNode(graph, edge)) {
			continue
		}
		eligible = append(eligible, edge)
		refs = append(refs, aig.Aig_ObjRefs(graph, edge))
	}
	if len(eligible) == 0 {
		return vec.Vec_IntAlloc(0)
	}
	cut := min(int(maxCandidates), len(eligible)-1)
	orderedRefs := append([]int32(nil), refs...)
	sort.Slice(orderedRefs, func(i, j int) bool { return orderedRefs[i] < orderedRefs[j] })
	threshold := orderedRefs[cut]
	order := make([]int, 0, len(eligible))
	for i, count := range refs {
		if count >= threshold {
			order = append(order, i)
		}
	}
	sort.SliceStable(order, func(i, j int) bool { return refs[order[i]] > refs[order[j]] })
	result := vec.Vec_IntAlloc(len(order))
	for _, index := range order {
		vec.Vec_IntPush(result, int32(eligible[index].NodeID()))
	}
	assert(int32(vec.Vec_IntSize(result)) >= min(maxCandidates, int32(len(eligible))-1), "selected at least requested candidate count")
	return result
}

// Llb_ManModelCheckAigWithHints progressively adds then relaxes candidate
// literals, retaining the least restrictive hint set that proves the property.
func Llb_ManModelCheckAigWithHints(graph *Aig_Man_t, params *Gia_ParLlb_t) int32 {
	if graph == nil || params == nil || params.nHintDepth <= 0 {
		return -1
	}
	started := util.Abc_Clock()
	candidates := Llb_ManCollectHighFanoutObjects(graph, params.nHintDepth+params.HintFirst, 1)
	hints := vec.Vec_IntStartFull(int(aig.Aig_ManObjNumMax(graph)))
	result := int32(-1)
	lastTried := -1
	var globalBDD *DdManager
	for index, objectID := range vec.Vec_IntArray(candidates) {
		if index < int(params.HintFirst) {
			continue
		}
		lastTried = index
		vec.Vec_IntWriteEntry(hints, int(objectID), 1)
		result = Llb_ManModelCheckAig(graph, params, hints, &globalBDD)
		if result == 0 || result == 1 {
			if result == 0 {
				break
			}
			break
		}
	}
	if result == 1 {
		for index := lastTried; index >= int(params.HintFirst); index-- {
			objectID := vec.Vec_IntEntry(candidates, index)
			vec.Vec_IntWriteEntry(hints, int(objectID), -1)
			result = Llb_ManModelCheckAig(graph, params, hints, &globalBDD)
			if result == 0 {
				break
			}
			if result == 1 {
				continue
			}
			break
		}
	}
	if globalBDD != nil {
		core.Cudd_Quit(globalBDD)
	}
	vec.Vec_IntFree(candidates)
	vec.Vec_IntFree(hints)
	if params.fVerbose != 0 {
		util.Abc_PrintTime(1, "Total runtime", util.Abc_Clock()-started)
	}
	return result
}
