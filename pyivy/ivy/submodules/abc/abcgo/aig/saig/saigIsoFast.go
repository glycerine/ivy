package saig

import (
	"fmt"
	"sort"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

const AIG_ISO_NUM = 16

// Iso_Dat_t stores the compact structural signature fields from the source
// bitfield. Each field is masked to the width used by the C declaration.
type Iso_Dat_t struct {
	nFiNeg uint32
	nFoNeg uint32
	nFoPos uint32
	Fi0Lev uint32
	Fi1Lev uint32
	Level  uint32
	fVisit uint32
}

type Iso_Dat2_t struct{ Data uint32 }

type Iso_Sto_t struct {
	pAig      *aig.Graph
	nObjs     int32
	pData     []Iso_Dat_t
	vVisited  []int
	vRoots    []aig.Edge
	vPlaces   []uint32
	pCounters []uint32
}

func Iso_StoStart(graph *aig.Graph) *Iso_Sto_t {
	if graph == nil {
		return nil
	}
	n := len(graph.Nodes())
	return &Iso_Sto_t{
		pAig: graph, nObjs: int32(n), pData: make([]Iso_Dat_t, n),
		pCounters: make([]uint32, 1<<AIG_ISO_NUM),
	}
}

func Iso_StoStop(manager *Iso_Sto_t) {
	if manager == nil {
		return
	}
	manager.pData = nil
	manager.vVisited = nil
	manager.vRoots = nil
	manager.vPlaces = nil
	manager.pCounters = nil
	manager.pAig = nil
}

// Iso_StoCollectInfo_rec accumulates one regular object and its fanin cone.
func Iso_StoCollectInfo_rec(manager *Iso_Sto_t, object aig.Edge, complemented bool) {
	if manager == nil || manager.pAig == nil {
		return
	}
	object = aig.Aig_Regular(object)
	if !manager.pAig.ValidEdge(object) {
		return
	}
	id := object.NodeID()
	node := manager.pAig.Node(object)
	if node.Kind != aig.Input && node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
		return
	}
	data := &manager.pData[id]
	if data.fVisit != 0 {
		if complemented {
			data.nFoNeg = (data.nFoNeg + 1) & 3
		} else {
			data.nFoPos = (data.nFoPos + 1) & 3
		}
		return
	}
	data.fVisit = 1
	if complemented {
		data.nFoNeg = 1
	} else {
		data.nFoPos = 1
	}
	data.Level = uint32(node.Level) & 7
	data.nFiNeg = 0
	for _, fanin := range node.Fanins {
		if fanin == aig.InvalidEdge {
			continue
		}
		if fanin.Complemented() {
			data.nFiNeg++
		}
	}
	data.nFiNeg &= 7
	switch node.Kind {
	case aig.And, aig.ExclusiveOr:
		first, second := node.Fanins[0], node.Fanins[1]
		if (first.Complemented() && !second.Complemented()) || (first.Complemented() == second.Complemented() && manager.pAig.Node(first).Level >= manager.pAig.Node(second).Level) {
			first, second = second, first
		}
		data.Fi0Lev = uint32(node.Level-manager.pAig.Node(first).Level) & 7
		data.Fi1Lev = uint32(node.Level-manager.pAig.Node(second).Level) & 7
		Iso_StoCollectInfo_rec(manager, node.Fanins[0], node.Fanins[0].Complemented())
		Iso_StoCollectInfo_rec(manager, node.Fanins[1], node.Fanins[1].Complemented())
	case aig.Buffer:
		data.Fi0Lev = uint32(node.Level-manager.pAig.Node(node.Fanins[0]).Level) & 7
		Iso_StoCollectInfo_rec(manager, node.Fanins[0], node.Fanins[0].Complemented())
	case aig.Input:
		if Saig_ObjIsLo(manager.pAig, object) {
			data.Fi0Lev, data.Fi1Lev = 1, 0
			manager.vRoots = append(manager.vRoots, Saig_ObjLoToLi(manager.pAig, object))
		} else if Saig_ObjIsPi(manager.pAig, object) {
			data.Fi0Lev, data.Fi1Lev = 0, 0
		}
	}
	manager.vVisited = append(manager.vVisited, int(id))
}

// Iso_StoCollectInfo returns sorted, frequency-packed structural signatures
// for the cone of one output. The final entry is reserved for the caller's PO ID.
func Iso_StoCollectInfo(manager *Iso_Sto_t, output aig.Edge) *vec.Vec_Int_t {
	if manager == nil || manager.pAig == nil || !manager.pAig.ValidEdge(output) || !aig.Aig_ObjIsCo(manager.pAig, output) {
		return vec.Vec_IntAlloc(0)
	}
	manager.vVisited = manager.vVisited[:0]
	manager.vRoots = manager.vRoots[:0]
	manager.vRoots = append(manager.vRoots, output)
	for i := 0; i < len(manager.vRoots); i++ {
		co := manager.vRoots[i]
		fanin := aig.Aig_ObjChild0(manager.pAig, co)
		if !aig.Aig_ObjIsConst1(manager.pAig, fanin) {
			Iso_StoCollectInfo_rec(manager, fanin, fanin.Complemented())
		}
	}
	counts := make(map[uint32]uint32, len(manager.vVisited))
	for _, id := range manager.vVisited {
		data := &manager.pData[id]
		signature := isoPackData(*data)
		if manager.pCounters[signature] == 0 {
			manager.vPlaces = append(manager.vPlaces, signature)
		}
		manager.pCounters[signature]++
		counts[signature] = manager.pCounters[signature]
		*data = Iso_Dat_t{}
	}
	items := make([]isoFrequency, 0, len(manager.vPlaces))
	for _, signature := range manager.vPlaces {
		items = append(items, isoFrequency{signature: signature, count: counts[signature]})
		manager.pCounters[signature] = 0
	}
	manager.vVisited = manager.vVisited[:0]
	manager.vPlaces = manager.vPlaces[:0]
	sort.SliceStable(items, func(i, j int) bool { return items[i].count < items[j].count })
	info := vec.Vec_IntAlloc(len(items))
	for i := len(items) - 1; i >= 0; i-- {
		packed := (items[i].count << AIG_ISO_NUM) | items[i].signature
		vec.Vec_IntPush(info, int32(packed))
	}
	return info
}

type isoFrequency struct {
	signature uint32
	count     uint32
}

func isoPackData(data Iso_Dat_t) uint32 {
	return (data.nFiNeg & 7) |
		((data.nFoNeg & 3) << 3) |
		((data.nFoPos & 3) << 5) |
		((data.Fi0Lev & 7) << 7) |
		((data.Fi1Lev & 7) << 10) |
		((data.Level & 7) << 13)
}

func Iso_StoCompareVecInt(first, second *vec.Vec_Int_t) int32 {
	limit := min(vec.Vec_IntSize(first), vec.Vec_IntSize(second))
	for i := 0; i < limit; i++ {
		a, b := vec.Vec_IntEntry(first, i), vec.Vec_IntEntry(second, i)
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
	}
	if vec.Vec_IntSize(first) < vec.Vec_IntSize(second) {
		return -1
	}
	if vec.Vec_IntSize(first) > vec.Vec_IntSize(second) {
		return 1
	}
	return 0
}

// Saig_IsoDetectFast groups sequential primary outputs by their structural
// signature. Unlike the C routine's commented-out return, the Go port returns
// the computed classes so callers can use the result directly.
func Saig_IsoDetectFast(graph *aig.Graph) *vec.Vec_Vec_t {
	if graph == nil {
		return vec.Vec_VecAlloc(0)
	}
	manager := Iso_StoStart(graph)
	outputs := graph.Outputs()
	poCount := len(outputs) - int(aig.Aig_ManRegNum(graph))
	infos := make([]*vec.Vec_Int_t, 0, poCount)
	for i := 0; i < poCount; i++ {
		info := Iso_StoCollectInfo(manager, outputs[i])
		vec.Vec_IntPush(info, int32(i))
		infos = append(infos, info)
	}
	Iso_StoStop(manager)
	sort.Slice(infos, func(i, j int) bool { return Iso_StoCompareVecInt(infos[i], infos[j]) < 0 })
	classes := vec.Vec_VecAlloc(len(infos))
	var previous *vec.Vec_Int_t
	for _, info := range infos {
		signature := vec.Vec_IntDup(info)
		po := vec.Vec_IntPop(info)
		vec.Vec_IntPop(signature)
		if previous == nil || Iso_StoCompareVecInt(previous, signature) != 0 {
			vec.Vec_VecExpandInt(classes, vec.Vec_VecSize(classes))
			previous = signature
		}
		vec.Vec_IntPush(vec.Vec_VecEntryInt(classes, vec.Vec_VecSize(classes)-1), po)
	}
	unique := 0
	for i := 0; i < vec.Vec_VecSize(classes); i++ {
		if vec.Vec_IntSize(vec.Vec_VecEntryInt(classes, i)) == 1 {
			unique++
		}
	}
	fmt.Printf("Devided %d outputs into %d cand equiv classes.\n", poCount, vec.Vec_VecSize(classes))
	for i := 0; i < vec.Vec_VecSize(classes); i++ {
		if n := vec.Vec_IntSize(vec.Vec_VecEntryInt(classes, i)); n > 1 {
			fmt.Printf("%d ", n)
		}
	}
	fmt.Printf(" Unique = %d\n", unique)
	return classes
}
