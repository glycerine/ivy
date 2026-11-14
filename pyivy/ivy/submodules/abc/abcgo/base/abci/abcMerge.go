package abci

import (
	"fmt"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/nwk"
)

// Nwk_LMPars_t retains the C field names used by the ABC command wrapper.
// The graph implementation owns a Go-shaped parameter structure, converted at
// the boundary below.
type Nwk_LMPars_t struct {
	nMaxLutSize, nMaxSuppSize, nMaxDistance, nMaxLevelDiff, nMaxFanout int32
	fUseDiffSupp, fUseTfiTfo, fVeryVerbose, fVerbose                   int32
}

func Abc_NtkMarkFanins_rec(lut *Abc_Obj_t, minLevel int32) {
	if lut == nil || baseabc.Abc_ObjIsNode(lut) == 0 || baseabc.Abc_NodeIsTravIdCurrent(lut) != 0 {
		return
	}
	baseabc.Abc_NodeSetTravIdCurrent(lut)
	if baseabc.Abc_ObjLevel(lut) < minLevel {
		return
	}
	for i := int32(0); i < baseabc.Abc_ObjFaninNum(lut); i++ {
		Abc_NtkMarkFanins_rec(baseabc.Abc_ObjFanin(lut, i), minLevel)
	}
}

func Abc_NtkMarkFanouts_rec(lut *Abc_Obj_t, maxLevel, maxFanout int32) {
	if lut == nil || baseabc.Abc_ObjIsNode(lut) == 0 || baseabc.Abc_NodeIsTravIdCurrent(lut) != 0 {
		return
	}
	baseabc.Abc_NodeSetTravIdCurrent(lut)
	if baseabc.Abc_ObjLevel(lut) > maxLevel || baseabc.Abc_ObjFanoutNum(lut) > maxFanout {
		return
	}
	for i := int32(0); i < baseabc.Abc_ObjFanoutNum(lut); i++ {
		Abc_NtkMarkFanouts_rec(baseabc.Abc_ObjFanout(lut, i), maxLevel, maxFanout)
	}
}

func Abc_NtkCollectCircle(start, next *Vec_Ptr_t, maxFanout int32) {
	vec.Vec_PtrClear(next)
	for i := 0; i < vec.Vec_PtrSize(start); i++ {
		obj, _ := vec.Vec_PtrEntry(start, i).(*Abc_Obj_t)
		for k := int32(0); k < baseabc.Abc_ObjFaninNum(obj); k++ {
			n := baseabc.Abc_ObjFanin(obj, k)
			if baseabc.Abc_ObjIsNode(n) == 0 || baseabc.Abc_NodeIsTravIdCurrent(n) != 0 {
				continue
			}
			baseabc.Abc_NodeSetTravIdCurrent(n)
			vec.Vec_PtrPush(next, n)
		}
		for k := int32(0); k < baseabc.Abc_ObjFanoutNum(obj); k++ {
			n := baseabc.Abc_ObjFanout(obj, k)
			if baseabc.Abc_ObjIsNode(n) == 0 || baseabc.Abc_NodeIsTravIdCurrent(n) != 0 {
				continue
			}
			baseabc.Abc_NodeSetTravIdCurrent(n)
			if baseabc.Abc_ObjFanoutNum(n) <= maxFanout {
				vec.Vec_PtrPush(next, n)
			}
		}
	}
}

func Abc_NtkCollectNonOverlapCands(lut *Abc_Obj_t, start, next, cands *Vec_Ptr_t, pars *Nwk_LMPars_t) {
	vec.Vec_PtrClear(cands)
	if pars.nMaxSuppSize-baseabc.Abc_ObjFaninNum(lut) <= 1 {
		return
	}
	assert(pars.nMaxDistance > 0, "pPars.nMaxDistance > 0")
	vec.Vec_PtrClear(start)
	vec.Vec_PtrPush(start, lut)
	baseabc.Abc_NtkIncrementTravId(baseabc.Abc_ObjNtk(lut))
	baseabc.Abc_NodeSetTravIdCurrent(lut)
	for depth := int32(1); depth <= pars.nMaxDistance; depth++ {
		Abc_NtkCollectCircle(start, next, pars.nMaxFanout)
		start, next = next, start
		for k := 0; k < vec.Vec_PtrSize(start); k++ {
			vec.Vec_PtrPush(cands, vec.Vec_PtrEntry(start, k))
		}
	}
	baseabc.Abc_NtkIncrementTravId(baseabc.Abc_ObjNtk(lut))
	if pars.fUseTfiTfo != 0 {
		baseabc.Abc_NodeSetTravIdCurrent(lut)
	} else {
		baseabc.Abc_NodeSetTravIdPrevious(lut)
		Abc_NtkMarkFanins_rec(lut, baseabc.Abc_ObjLevel(lut)-pars.nMaxDistance)
		baseabc.Abc_NodeSetTravIdPrevious(lut)
		Abc_NtkMarkFanouts_rec(lut, baseabc.Abc_ObjLevel(lut)+pars.nMaxDistance, pars.nMaxFanout)
	}
	k := 0
	for i := 0; i < vec.Vec_PtrSize(cands); i++ {
		obj, _ := vec.Vec_PtrEntry(cands, i).(*Abc_Obj_t)
		if baseabc.Abc_NodeIsTravIdCurrent(obj) != 0 || baseabc.Abc_ObjFaninNum(lut)+baseabc.Abc_ObjFaninNum(obj) > pars.nMaxSuppSize {
			continue
		}
		delta := baseabc.Abc_ObjLevel(lut) - baseabc.Abc_ObjLevel(obj)
		if delta > pars.nMaxLevelDiff || -delta > pars.nMaxLevelDiff {
			continue
		}
		vec.Vec_PtrWriteEntry(cands, k, obj)
		k++
	}
	vec.Vec_PtrShrink(cands, k)
}

func Abc_NtkCountTotalFanins(lut, cand *Abc_Obj_t) int32 {
	n := baseabc.Abc_ObjFaninNum(lut)
	for i := int32(0); i < baseabc.Abc_ObjFaninNum(cand); i++ {
		if !baseabc.Abc_ObjFanin(cand, i).MarkC {
			n++
		}
	}
	return n
}

func Abc_NtkCollectOverlapCands(lut *Abc_Obj_t, cands *Vec_Ptr_t, pars *Nwk_LMPars_t) {
	vec.Vec_PtrClear(cands)
	baseabc.Abc_NtkIncrementTravId(baseabc.Abc_ObjNtk(lut))
	baseabc.Abc_NodeSetTravIdCurrent(lut)
	for i := int32(0); i < baseabc.Abc_ObjFaninNum(lut); i++ {
		baseabc.Abc_ObjFanin(lut, i).MarkC = true
	}
	for i := int32(0); i < baseabc.Abc_ObjFaninNum(lut); i++ {
		fanin := baseabc.Abc_ObjFanin(lut, i)
		if baseabc.Abc_ObjIsNode(fanin) == 0 || baseabc.Abc_ObjFanoutNum(fanin) > pars.nMaxFanout {
			continue
		}
		for k := int32(0); k < baseabc.Abc_ObjFanoutNum(fanin); k++ {
			obj := baseabc.Abc_ObjFanout(fanin, k)
			if baseabc.Abc_ObjIsNode(obj) == 0 || baseabc.Abc_NodeIsTravIdCurrent(obj) != 0 {
				continue
			}
			baseabc.Abc_NodeSetTravIdCurrent(obj)
			delta := baseabc.Abc_ObjLevel(lut) - baseabc.Abc_ObjLevel(obj)
			if delta > pars.nMaxLevelDiff || -delta > pars.nMaxLevelDiff || Abc_NtkCountTotalFanins(lut, obj) > pars.nMaxSuppSize {
				continue
			}
			vec.Vec_PtrPush(cands, obj)
		}
	}
	for i := int32(0); i < baseabc.Abc_ObjFaninNum(lut); i++ {
		baseabc.Abc_ObjFanin(lut, i).MarkC = false
	}
}

func Abc_NtkLutMerge(ntk *Abc_Ntk_t, pars *Nwk_LMPars_t) *Vec_Int_t {
	started := util.Abc_Clock()
	nVerts := int32(0)
	for _, lut := range ntk.ObjectsOfKind(baseabc.ObjectNode) {
		if baseabc.Abc_ObjFaninNum(lut) <= pars.nMaxLutSize {
			nVerts++
		}
	}
	g := nwk.Nwk_ManGraphAlloc(nVerts)
	start, next := vec.Vec_PtrAlloc(1000), vec.Vec_PtrAlloc(1000)
	cands1, cands2 := vec.Vec_PtrAlloc(1000), vec.Vec_PtrAlloc(1000)
	nCands := 0
	for _, lut := range ntk.ObjectsOfKind(baseabc.ObjectNode) {
		if baseabc.Abc_ObjFaninNum(lut) > pars.nMaxLutSize {
			continue
		}
		Abc_NtkCollectOverlapCands(lut, cands1, pars)
		if pars.fUseDiffSupp != 0 {
			Abc_NtkCollectNonOverlapCands(lut, start, next, cands2, pars)
		} else {
			vec.Vec_PtrClear(cands2)
		}
		if vec.Vec_PtrSize(cands1)+vec.Vec_PtrSize(cands2) == 0 {
			continue
		}
		nCands += vec.Vec_PtrSize(cands1) + vec.Vec_PtrSize(cands2)
		for _, list := range []*Vec_Ptr_t{cands1, cands2} {
			for k := 0; k < vec.Vec_PtrSize(list); k++ {
				cand, _ := vec.Vec_PtrEntry(list, k).(*Abc_Obj_t)
				nwk.Nwk_ManGraphHashEdge(g, baseabc.Abc_ObjId(lut), baseabc.Abc_ObjId(cand))
			}
		}
		if pars.fVeryVerbose != 0 {
			fmt.Printf("Node %6d : Fanins = %d. Fanouts = %3d.  Cand1 = %3d. Cand2 = %3d.\n", baseabc.Abc_ObjId(lut), baseabc.Abc_ObjFaninNum(lut), baseabc.Abc_ObjFaninNum(lut), vec.Vec_PtrSize(cands1), vec.Vec_PtrSize(cands2))
		}
	}
	if pars.fVerbose != 0 {
		fmt.Printf("Mergable LUTs = %6d. Total cands = %6d. ", nVerts, nCands)
		util.Abc_PrintTime(1, "Deriving graph", util.Abc_Clock()-started)
	}
	started = util.Abc_Clock()
	nwk.Nwk_ManGraphSolve(g)
	result := g.Pairs
	g.Pairs = nil
	if pars.fVerbose != 0 {
		fmt.Printf("GRAPH: Nodes = %6d. Edges = %6d. Pairs = %6d. ", len(g.Vertices)-1, len(g.Edges), vec.Vec_IntSize(result)/2)
		util.Abc_PrintTime(1, "Solving", util.Abc_Clock()-started)
		nwk.Nwk_ManGraphReportMemoryUsage(g)
	}
	nwk.Nwk_ManGraphFree(g)
	return result
}
