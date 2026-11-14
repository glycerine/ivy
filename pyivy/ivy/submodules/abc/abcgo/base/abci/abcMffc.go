package abci

import (
	"fmt"
	"math"
)

import (
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

// MFFC traversal temporarily adjusts fanout reference counts, as the C
// implementation does, but keeps graph adjacency intact through baseabc's
// reference-count API.
func Abc_MffcDeref_rec(node *Abc_Obj_t, nodes *Vec_Ptr_t) {
	if node == nil || baseabc.Abc_ObjIsCi(node) != 0 {
		return
	}
	for _, edge := range node.Fanins() {
		fanin := edge.Object
		if baseabc.Abc_ObjFanoutNum(fanin) <= 0 {
			panic("MFFC dereference encountered a nonpositive fanout count")
		}
		baseabc.Abc_ObjAdjustFanoutCount(fanin, -1)
		if baseabc.Abc_ObjFanoutNum(fanin) == 0 {
			Abc_MffcDeref_rec(fanin, nodes)
		}
	}
	if nodes != nil {
		Vec_PtrPush(nodes, node)
	}
}

func Abc_MffcRef_rec(node *Abc_Obj_t) {
	if node == nil || baseabc.Abc_ObjIsCi(node) != 0 {
		return
	}
	for _, edge := range node.Fanins() {
		fanin := edge.Object
		before := baseabc.Abc_ObjFanoutNum(fanin)
		baseabc.Abc_ObjAdjustFanoutCount(fanin, 1)
		if before == 0 {
			Abc_MffcRef_rec(fanin)
		}
	}
}

func Abc_MffcCollectNodes(roots []*Abc_Obj_t, nodes *Vec_Ptr_t) {
	Vec_PtrClear(nodes)
	for _, root := range roots {
		Abc_MffcDeref_rec(root, nodes)
	}
	for _, root := range roots {
		Abc_MffcRef_rec(root)
	}
	if len(roots) != 0 {
		baseabc.Abc_NtkFinishFanoutCountAdjustments(roots[0].Network())
	}
}

func Abc_MffcCollectLeaves(nodes, leaves *Vec_Ptr_t) {
	if Vec_PtrSize(nodes) == 0 {
		panic("MFFC leaf collection requires a nonempty node set")
	}
	first := vec.Vec_PtrEntry(nodes, 0).(*Abc_Obj_t)
	network := first.Network()
	baseabc.Abc_NtkIncrementTravId(network)
	for _, item := range vec.Vec_PtrArray(nodes) {
		baseabc.Abc_NodeSetTravIdCurrent(item.(*Abc_Obj_t))
	}
	Vec_PtrClear(leaves)
	for _, item := range vec.Vec_PtrArray(nodes) {
		for _, edge := range item.(*Abc_Obj_t).Fanins() {
			fanin := edge.Object
			if baseabc.Abc_NodeIsTravIdCurrent(fanin) != 0 {
				continue
			}
			baseabc.Abc_NodeSetTravIdCurrent(fanin)
			Vec_PtrPush(leaves, fanin)
		}
	}
}

func Abc_NktMffcMarkRoots(network *Abc_Ntk_t, skipPis int32) *Vec_Ptr_t {
	baseabc.Abc_NtkCleanMarkA(network)
	roots := Vec_PtrAlloc(1000)
	for i := int32(0); i < baseabc.Abc_NtkCoNum(network); i++ {
		co := baseabc.Abc_NtkCo(network, i)
		if baseabc.Abc_ObjFaninNum(co) == 0 {
			continue
		}
		obj := baseabc.Abc_ObjFanin0(co)
		if baseabc.Abc_ObjIsCi(obj) != 0 || obj.MarkA {
			continue
		}
		obj.MarkA = true
		Vec_PtrPush(roots, obj)
	}
	nodes, leaves := Vec_PtrAlloc(100), Vec_PtrAlloc(100)
	for i := int32(0); i < Vec_PtrSize(roots); i++ {
		obj := vec.Vec_PtrEntry(roots, int(i)).(*Abc_Obj_t)
		if baseabc.Abc_ObjIsCi(obj) != 0 {
			continue
		}
		Abc_MffcCollectNodes([]*Abc_Obj_t{obj}, nodes)
		Abc_MffcCollectLeaves(nodes, leaves)
		for _, item := range vec.Vec_PtrArray(leaves) {
			leaf := item.(*Abc_Obj_t)
			if (skipPis != 0 && baseabc.Abc_ObjIsCi(leaf) != 0) || leaf.MarkA {
				continue
			}
			leaf.MarkA = true
			Vec_PtrPush(roots, leaf)
		}
	}
	Vec_PtrFree(nodes)
	Vec_PtrFree(leaves)
	baseabc.Abc_NtkCleanMarkA(network)
	return roots
}

func Abc_NktMffcCollectFanout_rec(obj *Abc_Obj_t, fanouts *Vec_Ptr_t) {
	if obj == nil || baseabc.Abc_ObjIsCo(obj) != 0 || baseabc.Abc_ObjFanoutNum(obj) > 64 || baseabc.Abc_NodeIsTravIdCurrent(obj) != 0 {
		return
	}
	baseabc.Abc_NodeSetTravIdCurrent(obj)
	if obj.MarkA {
		if baseabc.Abc_ObjFanoutNum(obj) > 0 {
			Vec_PtrPush(fanouts, obj)
		}
		return
	}
	for _, fanout := range obj.Fanouts() {
		Abc_NktMffcCollectFanout_rec(fanout, fanouts)
	}
}

func Abc_NktMffcCollectFanout(roots []*Abc_Obj_t, leaves, fanouts *Vec_Ptr_t) {
	for _, root := range roots {
		Abc_MffcDeref_rec(root, nil)
	}
	if len(roots) != 0 {
		baseabc.Abc_NtkFinishFanoutCountAdjustments(roots[0].Network())
	}
	Vec_PtrClear(fanouts)
	if Vec_PtrSize(leaves) != 0 {
		first := vec.Vec_PtrEntry(leaves, 0).(*Abc_Obj_t)
		baseabc.Abc_NtkIncrementTravId(first.Network())
		for _, item := range vec.Vec_PtrArray(leaves) {
			leaf := item.(*Abc_Obj_t)
			for _, fanout := range leaf.Fanouts() {
				Abc_NktMffcCollectFanout_rec(fanout, fanouts)
			}
		}
	}
	for _, root := range roots {
		Abc_MffcRef_rec(root)
	}
	if len(roots) != 0 {
		baseabc.Abc_NtkFinishFanoutCountAdjustments(roots[0].Network())
	}
}

func Abc_NktMffcGrowOne(network *Abc_Ntk_t, roots []*Abc_Obj_t, nodes, leaves, fanouts *Vec_Ptr_t) *Abc_Obj_t {
	if len(roots) == 0 {
		return nil
	}
	Abc_MffcCollectNodes(roots, nodes)
	Abc_MffcCollectLeaves(nodes, leaves)
	Abc_NktMffcCollectFanout(roots, leaves, fanouts)
	var best *Abc_Obj_t
	bestCost := -1.0
	for _, item := range vec.Vec_PtrArray(fanouts) {
		candidate := item.(*Abc_Obj_t)
		found := false
		for _, root := range roots {
			if root == candidate {
				found = true
				break
			}
		}
		if found {
			continue
		}
		trial := append(append([]*Abc_Obj_t(nil), roots...), candidate)
		Abc_MffcCollectNodes(trial, nodes)
		Abc_MffcCollectLeaves(nodes, leaves)
		cost := float64(Vec_PtrSize(nodes)) / float64(max(1, Vec_PtrSize(leaves)))
		if best == nil || cost > bestCost {
			best, bestCost = candidate, cost
		}
	}
	_ = network
	return best
}

func Abc_NktMffcGrowRoots(network *Abc_Ntk_t, roots *Vec_Ptr_t) *Vec_Ptr_t {
	baseabc.Abc_NtkCleanMarkA(network)
	for _, item := range vec.Vec_PtrArray(roots) {
		item.(*Abc_Obj_t).MarkA = true
	}
	result := Vec_PtrAlloc(100)
	nodes, leaves, fanouts := Vec_PtrAlloc(100), Vec_PtrAlloc(100), Vec_PtrAlloc(100)
	for _, item := range vec.Vec_PtrArray(roots) {
		if root := Abc_NktMffcGrowOne(network, []*Abc_Obj_t{item.(*Abc_Obj_t)}, nodes, leaves, fanouts); root != nil {
			Vec_PtrPush(result, root)
		} else {
			Vec_PtrPush(result, nil)
		}
	}
	Vec_PtrFree(nodes)
	Vec_PtrFree(leaves)
	Vec_PtrFree(fanouts)
	baseabc.Abc_NtkCleanMarkA(network)
	return result
}

func Abc_NktMffcGrowRootsAgain(network *Abc_Ntk_t, roots, firstPass *Vec_Ptr_t) *Vec_Ptr_t {
	baseabc.Abc_NtkCleanMarkA(network)
	for _, item := range vec.Vec_PtrArray(roots) {
		item.(*Abc_Obj_t).MarkA = true
	}
	result := Vec_PtrAlloc(Vec_PtrSize(roots))
	nodes, leaves, fanouts := Vec_PtrAlloc(100), Vec_PtrAlloc(100), Vec_PtrAlloc(100)
	for i, item := range vec.Vec_PtrArray(roots) {
		root := item.(*Abc_Obj_t)
		other := vec.Vec_PtrEntry(firstPass, i)
		if other == nil {
			Vec_PtrPush(result, nil)
			continue
		}
		best := Abc_NktMffcGrowOne(network, []*Abc_Obj_t{root, other.(*Abc_Obj_t)}, nodes, leaves, fanouts)
		Vec_PtrPush(result, best)
	}
	Vec_PtrFree(nodes)
	Vec_PtrFree(leaves)
	Vec_PtrFree(fanouts)
	baseabc.Abc_NtkCleanMarkA(network)
	return result
}

// Abc_NktMffCollectLeafRoot marks the requested volume, then collects its
// external fanins and boundary roots. Marks are always cleared before return.
func Abc_NktMffCollectLeafRoot(network *Abc_Ntk_t, nodes, leaves, roots *Vec_Ptr_t) {
	for _, item := range vec.Vec_PtrArray(nodes) {
		item.(*Abc_Obj_t).MarkA = true
	}
	Vec_PtrClear(leaves)
	baseabc.Abc_NtkIncrementTravId(network)
	for _, item := range vec.Vec_PtrArray(nodes) {
		for _, edge := range item.(*Abc_Obj_t).Fanins() {
			fanin := edge.Object
			if fanin.MarkA || baseabc.Abc_NodeIsTravIdCurrent(fanin) != 0 {
				continue
			}
			baseabc.Abc_NodeSetTravIdCurrent(fanin)
			Vec_PtrPush(leaves, fanin)
		}
	}
	if roots != nil {
		Vec_PtrClear(roots)
		for _, item := range vec.Vec_PtrArray(nodes) {
			obj := item.(*Abc_Obj_t)
			for _, fanout := range obj.Fanouts() {
				if !fanout.MarkA {
					Vec_PtrPush(roots, obj)
					break
				}
			}
		}
	}
	for _, item := range vec.Vec_PtrArray(nodes) {
		item.(*Abc_Obj_t).MarkA = false
	}
}

func Abc_NktMffCollectLeafRootInt(network *Abc_Ntk_t, nodes, leaves, roots *Vec_Int_t) {
	ptrNodes, ptrLeaves, ptrRoots := Vec_PtrAlloc(Vec_IntSize(nodes)), Vec_PtrAlloc(100), Vec_PtrAlloc(100)
	for _, id := range vec.Vec_IntArray(nodes) {
		Vec_PtrPush(ptrNodes, baseabc.Abc_NtkObj(network, id))
	}
	Abc_NktMffCollectLeafRoot(network, ptrNodes, ptrLeaves, ptrRoots)
	vec.Vec_IntClear(leaves)
	for _, item := range vec.Vec_PtrArray(ptrLeaves) {
		Vec_IntPush(leaves, baseabc.Abc_ObjId(item.(*Abc_Obj_t)))
	}
	if roots != nil {
		vec.Vec_IntClear(roots)
		for _, item := range vec.Vec_PtrArray(ptrRoots) {
			Vec_IntPush(roots, baseabc.Abc_ObjId(item.(*Abc_Obj_t)))
		}
	}
	Vec_PtrFree(ptrNodes)
	Vec_PtrFree(ptrLeaves)
	Vec_PtrFree(ptrRoots)
}

func Abc_NktMffcDerive(network *Abc_Ntk_t, faninsOut, fanoutsOut, volumesOut **Vec_Ptr_t) *Vec_Ptr_t {
	roots := Abc_NktMffcMarkRoots(network, 0)
	maxID := baseabc.Abc_NtkObjNumMax(network)
	fanins, fanouts, volumes := vec.Vec_PtrStart(int(maxID)), vec.Vec_PtrStart(int(maxID)), vec.Vec_PtrStart(int(maxID))
	for _, item := range vec.Vec_PtrArray(roots) {
		root := item.(*Abc_Obj_t)
		id := int(root.ID)
		vec.Vec_PtrWriteEntry(fanins, id, vec.Vec_IntAlloc(8))
		vec.Vec_PtrWriteEntry(fanouts, id, vec.Vec_IntAlloc(8))
		vec.Vec_PtrWriteEntry(volumes, id, vec.Vec_IntAlloc(8))
	}
	nodes, leaves := Vec_PtrAlloc(100), Vec_PtrAlloc(100)
	for _, item := range vec.Vec_PtrArray(roots) {
		root := item.(*Abc_Obj_t)
		if baseabc.Abc_ObjIsCi(root) != 0 {
			continue
		}
		Abc_MffcCollectNodes([]*Abc_Obj_t{root}, nodes)
		Abc_MffcCollectLeaves(nodes, leaves)
		rootFanins := vec.Vec_PtrEntry(fanins, root.ID).(*Vec_Int_t)
		rootVolumes := vec.Vec_PtrEntry(volumes, root.ID).(*Vec_Int_t)
		for _, leaf := range vec.Vec_PtrArray(leaves) {
			fanin := leaf.(*Abc_Obj_t)
			Vec_IntPush(rootFanins, int32(fanin.ID))
			Vec_IntPush(vec.Vec_PtrEntry(fanouts, fanin.ID).(*Vec_Int_t), int32(root.ID))
		}
		for _, node := range vec.Vec_PtrArray(nodes) {
			Vec_IntPush(rootVolumes, int32(node.(*Abc_Obj_t).ID))
		}
	}
	Vec_PtrFree(nodes)
	Vec_PtrFree(leaves)
	for _, item := range vec.Vec_PtrArray(roots) {
		root := item.(*Abc_Obj_t)
		vec.Vec_IntSort(vec.Vec_PtrEntry(fanins, root.ID).(*Vec_Int_t), 0)
		vec.Vec_IntSort(vec.Vec_PtrEntry(fanouts, root.ID).(*Vec_Int_t), 0)
	}
	*faninsOut, *fanoutsOut, *volumesOut = fanins, fanouts, volumes
	return roots
}

func Abc_NktMffcFree(roots, fanins, fanouts, volumes *Vec_Ptr_t) {
	for _, item := range vec.Vec_PtrArray(roots) {
		root := item.(*Abc_Obj_t)
		for _, table := range []*Vec_Ptr_t{fanins, fanouts, volumes} {
			vec.Vec_IntFree(vec.Vec_PtrEntry(table, root.ID).(*Vec_Int_t))
		}
	}
	Vec_PtrFree(volumes)
	Vec_PtrFree(fanouts)
	Vec_PtrFree(fanins)
	Vec_PtrFree(roots)
}

func Abc_NktMffcCostTwo(supp1, supp2 *Vec_Int_t, volume, limit int32) float64 {
	common := int32(vec.Vec_IntTwoCountCommon(supp1, supp2))
	if Vec_IntSize(supp1)+Vec_IntSize(supp2)-common > limit {
		return -math.MaxInt32
	}
	return 0.6*float64(common) - 1.2*float64(Vec_IntSize(supp2)) + 0.8*float64(volume)
}

func Abc_NktMffcSupport(group, fanins *Vec_Ptr_t) *Vec_Int_t {
	result := Vec_IntAlloc(100)
	for _, item := range vec.Vec_PtrArray(group) {
		root := item.(*Abc_Obj_t)
		inputs := vec.Vec_PtrEntry(fanins, root.ID).(*Vec_Int_t)
		merged := vec.Vec_IntTwoMerge(result, inputs)
		Vec_IntFree(result)
		result = merged
	}
	return result
}

func Abc_NktMffcFindBest(network *Abc_Ntk_t, marks, inputs *Vec_Int_t, fanins, fanouts, volumes *Vec_Ptr_t, limit int32) *Abc_Obj_t {
	candidates := Vec_IntAlloc(100)
	for _, item := range vec.Vec_IntArray(inputs) {
		obj := baseabc.Abc_NtkObj(network, item)
		outs := vec.Vec_PtrEntry(fanouts, obj.ID).(*Vec_Int_t)
		if Vec_IntSize(outs) > 16 {
			continue
		}
		merged := vec.Vec_IntTwoMerge(candidates, outs)
		Vec_IntFree(candidates)
		candidates = merged
	}
	var best *Abc_Obj_t
	bestCost := -float64(math.MaxInt32)
	for _, id := range vec.Vec_IntArray(candidates) {
		if Vec_IntEntry(marks, id) == 0 {
			continue
		}
		candidate := baseabc.Abc_NtkObj(network, id)
		cost := Abc_NktMffcCostTwo(inputs, vec.Vec_PtrEntry(fanins, int(id)).(*Vec_Int_t), Vec_IntSize(vec.Vec_PtrEntry(volumes, int(id)).(*Vec_Int_t)), limit)
		if best == nil || cost > bestCost {
			best, bestCost = candidate, cost
		}
	}
	Vec_IntFree(candidates)
	return best
}

func Abc_NktMffcSaveOne(group, volumes *Vec_Ptr_t) *Vec_Int_t {
	result := Vec_IntAlloc(100)
	for _, item := range vec.Vec_PtrArray(group) {
		root := item.(*Abc_Obj_t)
		volume := vec.Vec_PtrEntry(volumes, root.ID).(*Vec_Int_t)
		for _, id := range vec.Vec_IntArray(volume) {
			Vec_IntPush(result, id)
		}
	}
	return result
}

func Abc_NodeCompareVolumeDecrease(a, b *any) int {
	x, y := (*a).(*Abc_Obj_t), (*b).(*Abc_Obj_t)
	if x.Temp > y.Temp {
		return -1
	}
	if x.Temp < y.Temp {
		return 1
	}
	if x.ID > y.ID {
		return -1
	}
	if x.ID < y.ID {
		return 1
	}
	return 0
}

func Abc_NktMffcServer(network *Abc_Ntk_t, maxInputs, maxOutputs int32) *Vec_Ptr_t {
	if maxOutputs < 1 || maxOutputs > 32 {
		panic("MFFC output limit must be in [1,32]")
	}
	result := Vec_PtrAlloc(100)
	var fanins, fanouts, volumes *Vec_Ptr_t
	pivots := Abc_NktMffcDerive(network, &fanins, &fanouts, &volumes)
	for _, item := range vec.Vec_PtrArray(pivots) {
		root := item.(*Abc_Obj_t)
		root.Temp = int(Vec_IntSize(vec.Vec_PtrEntry(volumes, root.ID).(*Vec_Int_t)))
	}
	vec.Vec_PtrSort(pivots, Abc_NodeCompareVolumeDecrease)
	marks := vec.Vec_IntStart(int(baseabc.Abc_NtkObjNumMax(network)))
	for _, item := range vec.Vec_PtrArray(pivots) {
		root := item.(*Abc_Obj_t)
		if baseabc.Abc_ObjIsNode(root) != 0 && Vec_IntSize(vec.Vec_PtrEntry(volumes, root.ID).(*Vec_Int_t)) > 1 {
			Vec_IntWriteEntry(marks, int32(root.ID), 1)
		}
	}
	group := Vec_PtrAlloc(10)
	for _, item := range vec.Vec_PtrArray(pivots) {
		pivot := item.(*Abc_Obj_t)
		if Vec_IntEntry(marks, int32(pivot.ID)) == 0 {
			continue
		}
		Vec_PtrClear(group)
		Vec_PtrPush(group, pivot)
		Vec_IntWriteEntry(marks, int32(pivot.ID), 0)
		leaves := vec.Vec_PtrEntry(fanins, pivot.ID).(*Vec_Int_t)
		if Vec_IntSize(leaves) > maxInputs {
			Vec_PtrPush(result, Abc_NktMffcSaveOne(group, volumes))
			continue
		}
		for k := int32(1); k < maxOutputs; k++ {
			support := Abc_NktMffcSupport(group, fanins)
			if Vec_IntSize(support) > maxInputs {
				Vec_IntFree(support)
				break
			}
			candidate := Abc_NktMffcFindBest(network, marks, support, fanins, fanouts, volumes, maxInputs)
			Vec_IntFree(support)
			if candidate == nil {
				break
			}
			Vec_PtrPush(group, candidate)
			Vec_IntWriteEntry(marks, int32(candidate.ID), 0)
		}
		Vec_PtrPush(result, Abc_NktMffcSaveOne(group, volumes))
	}
	Vec_PtrFree(group)
	Vec_IntFree(marks)
	Abc_NktMffcFree(pivots, fanins, fanouts, volumes)
	return result
}

func Abc_NktMffcServerTest(network *Abc_Ntk_t) {
	result := Abc_NktMffcServer(network, 18, 3)
	defer Vec_PtrFree(result)
	fmt.Printf("MFFC groups: %d\n", Vec_PtrSize(result))
}
