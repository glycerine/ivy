package ivy

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/opt/rwt"
)

// Ivy_ManRewritePre runs ABC's precomputed four-input cut rewriter. Candidate
// truth/decomposition data is owned by the RWT manager; Ivy supplies graph
// counting and update operations through ivyDecNetwork.
func Ivy_ManRewritePre(p *Ivy_Man_t, updateLevels, useZeroCost, verbose int32) int32 {
	if p == nil || p.memoryStopped {
		panic("ivy: rewrite on absent or stopped manager")
	}
	manager := rwt.Rwt_ManStart(0)
	if manager == nil {
		return 0
	}
	p.pData = manager
	start := time.Now()
	if updateLevels != 0 {
		if p.fFanout == 0 {
			Ivy_ManStartFanout(p)
		}
		Ivy_ManRequiredLevels(p)
	}
	maxNodeID := Ivy_ManObjIdMax(p)
	for _, node := range p.vObjs {
		if node == nil || Ivy_ObjIsNode(node) == 0 {
			continue
		}
		if node.Id > maxNodeID {
			break
		}
		Ivy_NodeFixBufferFanins(p, node, 1)
		if Ivy_ObjIsBuf(node) != 0 {
			continue
		}
		gain := Ivy_NodeRewrite(p, manager, node, updateLevels, useZeroCost)
		if gain <= 0 && !(gain == 0 && useZeroCost != 0) {
			continue
		}
		graph := rwt.Rwt_ManReadDecs(manager)
		if graph == nil {
			panic("ivy: RWT selected no decomposition graph")
		}
		complemented := rwt.Rwt_ManReadCompl(manager) != 0
		if complemented {
			dec.Dec_GraphComplement(graph)
		}
		updateStart := time.Now()
		Ivy_GraphUpdateNetwork(p, node, graph, updateLevels, gain)
		rwt.Rwt_ManAddTimeUpdate(manager, time.Since(updateStart))
		if complemented {
			dec.Dec_GraphComplement(graph)
		}
	}
	rwt.Rwt_ManAddTimeTotal(manager, time.Since(start))
	if verbose != 0 {
		rwt.Rwt_ManPrintStats(manager)
	}
	rwt.Rwt_ManStop(manager)
	p.pData = nil
	if updateLevels != 0 {
		p.vRequired = nil
	} else {
		Ivy_ManResetLevels(p)
	}
	if removed := Ivy_ManCleanup(p); removed != 0 {
		fmt.Printf("Cleanup after rewriting removed %d dangling nodes.\n", removed)
	}
	if Ivy_ManCheck(p) == 0 {
		fmt.Println("Ivy_ManRewritePre(): The check has failed.")
	}
	return 1
}

// Ivy_NodeRewrite finds the best precomputed decomposition for any 4-input
// cut of node. It returns the node-count gain, or -1 when no candidate fits.
func Ivy_NodeRewrite(p *Ivy_Man_t, manager *rwt.Rwt_Man_t, node *Ivy_Obj_t, updateLevels, useZeroCost int32) int32 {
	if p == nil || manager == nil || node == nil || Ivy_ObjIsNode(node) == 0 {
		panic("ivy: invalid rewrite candidate")
	}
	rwt.Rwt_ManRecordNodeConsidered(manager)
	levelLimit := int32(1 << 30)
	if updateLevels != 0 {
		if int(node.Id) >= len(p.vRequired) {
			panic("ivy: missing required level for rewrite root")
		}
		levelLimit = p.vRequired[node.Id]
	}
	cutStart := time.Now()
	store := Ivy_NodeFindCutsAll(p, node, 5)
	rwt.Rwt_ManAddTimeCuts(manager, time.Since(cutStart))
	if store == nil {
		return -1
	}
	var bestGraph *dec.Dec_Graph_t
	var bestFanins []any
	var bestTruth uint16
	var bestCompl bool
	var bestGain int32 = -1
	for ci := int32(1); ci < store.nCuts; ci++ {
		cut := &store.pCuts[ci]
		if cut.nSize != 4 {
			continue
		}
		bufferFound := false
		for i := 0; i < int(cut.nSize); i++ {
			if Ivy_ObjIsBuf(Ivy_ManObj(p, cut.pArray[i])) != 0 {
				bufferFound = true
				break
			}
		}
		if bufferFound {
			rwt.Rwt_ManRecordCut(manager, false)
			continue
		}
		rwt.Rwt_ManRecordCut(manager, true)
		truthStart := time.Now()
		truth := uint16(Ivy_NodeGetTruth(node, cut.pArray[:int(cut.nSize)]))
		class, phase, permutation := rwt.Rwt_ManCanonInfo(manager, truth)
		rwt.Rwt_ManAddTimeTruth(manager, time.Since(truthStart))
		if len(permutation) != 4 {
			continue
		}

		current := rwt.Rwt_ManReadCurrentLeaves(manager)
		Vec_PtrClear(current)
		Vec_PtrFill(current, 4, nil)
		for i := int32(0); i < 4; i++ {
			fanin := Ivy_ManObj(p, cut.pArray[permutation[i]])
			fanin = Ivy_NotCond(fanin, int32(phase>>i)&1)
			Vec_PtrWriteEntry(current, i, fanin)
		}

		mffcStart := time.Now()
		for i := int32(0); i < Vec_PtrSize(current); i++ {
			Ivy_ObjRefsInc(Ivy_Regular(ivyMultiVecObj(current, i)))
		}
		Ivy_ManIncrementTravId(p)
		nodesSaved := Ivy_ObjMffcLabel(p, node)
		for i := int32(0); i < Vec_PtrSize(current); i++ {
			Ivy_ObjRefsDec(Ivy_Regular(ivyMultiVecObj(current, i)))
		}
		rwt.Rwt_ManAddTimeMffc(manager, time.Since(mffcStart))

		evalStart := time.Now()
		graphs := rwt.Rwt_ManClassSubgraphs(manager, int32(class))
		rwt.Rwt_ManAddSubgraphs(manager, int32(len(graphs)))
		for _, candidate := range graphs {
			graph := rwt.Rwt_ManNodeDecomposition(manager, candidate)
			if graph == nil {
				continue
			}
			fanins := make([]any, 4)
			for i := range fanins {
				fanin, ok := Vec_PtrEntry(current, int32(i)).(*Ivy_Obj_t)
				if !ok || fanin == nil {
					panic("ivy: malformed RWT current fanin")
				}
				fanins[i] = fanin
				dec.Dec_GraphSetLeafFunc(graph, int32(i), fanin)
			}
			added := Ivy_GraphToNetworkCount(p, node, graph, nodesSaved, levelLimit)
			if added < 0 || nodesSaved < added {
				continue
			}
			gain := nodesSaved - added
			if gain > bestGain {
				bestGain = gain
				bestGraph = graph
				bestFanins = fanins
				bestTruth = truth
				bestCompl = phase&(1<<4) != 0
				if rwt.Rwt_IsComplement(candidate) != 0 {
					bestCompl = !bestCompl
				}
			}
		}
		rwt.Rwt_ManAddTimeEval(manager, time.Since(evalStart))
	}
	if bestGain < 0 || bestGraph == nil {
		return -1
	}
	if bestGain == 0 && useZeroCost == 0 {
		return 0
	}
	rwt.Rwt_ManStoreBest(manager, bestGraph, bestCompl, bestFanins, bestTruth, bestGain, useZeroCost != 0)
	return bestGain
}

var ivyTruthMasks = [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}

func Ivy_NodeGetTruth(node *Ivy_Obj_t, leaves []int32) uint32 {
	if node == nil || len(leaves) >= 6 {
		panic("ivy: truth cut needs fewer than six leaves")
	}
	leafTruth := make(map[int32]uint32, len(leaves))
	for i, id := range leaves {
		leafTruth[id] = ivyTruthMasks[i]
	}
	cache := make(map[*Ivy_Obj_t]uint32)
	var eval func(*Ivy_Obj_t) uint32
	eval = func(edge *Ivy_Obj_t) uint32 {
		n := Ivy_Regular(edge)
		if Ivy_ObjIsConst1(n) != 0 {
			return ^uint32(0)
		}
		if truth, found := leafTruth[n.Id]; found {
			if Ivy_IsComplement(edge) != 0 {
				return ^truth
			}
			return truth
		}
		truth, found := cache[n]
		if !found {
			switch {
			case Ivy_ObjIsBuf(n) != 0:
				truth = eval(Ivy_ObjChild0(n))
			case Ivy_ObjIsAnd(n) != 0:
				truth = eval(Ivy_ObjChild0(n)) & eval(Ivy_ObjChild1(n))
			case Ivy_ObjIsExor(n) != 0:
				truth = eval(Ivy_ObjChild0(n)) ^ eval(Ivy_ObjChild1(n))
			default:
				panic("ivy: truth cut reached a leaf absent from the cut")
			}
			cache[n] = truth
		}
		if Ivy_IsComplement(edge) != 0 {
			return ^truth
		}
		return truth
	}
	return eval(node)
}

type ivyDecNetwork struct{ manager *Ivy_Man_t }

func (n ivyDecNetwork) ConstantOne() any { return Ivy_ManConst1(n.manager) }
func (n ivyDecNetwork) Complement(value any) any {
	return Ivy_Not(ivyNetworkNode(value))
}
func (n ivyDecNetwork) And(a, b any) any {
	return Ivy_And(n.manager, ivyNetworkNode(a), ivyNetworkNode(b))
}
func (n ivyDecNetwork) CreateAnd(a, b any) any {
	return Ivy_ObjCreate(n.manager, Ivy_ObjCreateGhost(n.manager, ivyNetworkNode(a), ivyNetworkNode(b), IVY_AND, IVY_INIT_NONE))
}
func (n ivyDecNetwork) Variable(index int) any {
	if index < 0 || index >= len(n.manager.vPis) {
		panic("ivy: decomposition leaf variable out of bounds")
	}
	return n.manager.vPis[index]
}
func (n ivyDecNetwork) Replace(old, replacement any, update bool) {
	levelUpdate := int32(0)
	if update {
		levelUpdate = 1
	}
	Ivy_ObjReplace(n.manager, ivyNetworkNode(old), ivyNetworkNode(replacement), 1, 0, levelUpdate)
}
func (n ivyDecNetwork) LookupAnd(a, b any) any {
	return ivyLookupAnd(n.manager, ivyNetworkNode(a), ivyNetworkNode(b))
}
func (n ivyDecNetwork) SameRegular(a, b any) bool {
	return Ivy_Regular(ivyNetworkNode(a)) == Ivy_Regular(ivyNetworkNode(b))
}
func (n ivyDecNetwork) Level(value any) int { return int(Ivy_ObjLevel(ivyNetworkNode(value))) }
func (n ivyDecNetwork) InDeletionCone(value any) bool {
	return Ivy_ObjIsTravIdCurrent(n.manager, ivyNetworkNode(value)) != 0
}

func ivyNetworkNode(value any) *Ivy_Obj_t {
	node, ok := value.(*Ivy_Obj_t)
	if !ok || node == nil {
		panic("ivy: decomposition payload is not an Ivy object")
	}
	return node
}

func Ivy_GraphToNetworkCount(p *Ivy_Man_t, root *Ivy_Obj_t, graph *dec.Dec_Graph_t, nodeMax, levelMax int32) int32 {
	return dec.Dec_GraphToNetworkCount(ivyDecNetwork{p}, root, graph, nodeMax, levelMax)
}

func Ivy_GraphToNetwork(p *Ivy_Man_t, graph *dec.Dec_Graph_t) *Ivy_Obj_t {
	result, ok := dec.Dec_GraphToNetwork(ivyDecNetwork{p}, graph).(*Ivy_Obj_t)
	if !ok || result == nil {
		panic("ivy: decomposition did not produce an Ivy node")
	}
	return result
}

func Ivy_GraphUpdateNetwork(p *Ivy_Man_t, root *Ivy_Obj_t, graph *dec.Dec_Graph_t, updateLevels, gain int32) {
	required := int32(1 << 30)
	if updateLevels != 0 {
		required = p.vRequired[root.Id]
	}
	oldCount := Ivy_ManNodeNum(p)
	newRoot := Ivy_GraphToNetwork(p, graph)
	assert(Ivy_ObjLevel(newRoot) <= required, "new root respects required level")
	Ivy_ObjReplace(p, root, newRoot, 1, 0, 1)
	newCount := Ivy_ManNodeNum(p)
	assert(gain <= oldCount-newCount, "node gain is no greater than actual reduction")
	if len(p.vBufs) > 0 {
		if p.fFanout == 0 {
			Ivy_ManStartFanout(p)
		}
		Ivy_ManPropagateBuffers(p, updateLevels)
	}
}

func Ivy_GraphUpdateNetwork3(p *Ivy_Man_t, root *Ivy_Obj_t, graph *dec.Dec_Graph_t, updateLevels, gain int32) {
	oldCount := Ivy_ManNodeNum(p)
	var fanins []*Ivy_Obj_t
	if p.pData != nil {
		if manager, ok := p.pData.(*rwt.Rwt_Man_t); ok {
			entries := rwt.Rwt_ManReadLeaves(manager)
			for i := int32(0); i < Vec_PtrSize(entries); i++ {
				fanin, ok := Vec_PtrEntry(entries, i).(*Ivy_Obj_t)
				if ok && fanin != nil {
					fanins = append(fanins, Ivy_Regular(fanin))
					Ivy_ObjRefsInc(Ivy_Regular(fanin))
				}
			}
		}
	}
	oldRefs := root.nRefs
	root.nRefs = 0
	Ivy_ObjDelete_rec(p, root, 0)
	root.nRefs = oldRefs
	for _, fanin := range fanins {
		Ivy_ObjRefsDec(fanin)
	}
	newRoot := Ivy_GraphToNetwork(p, graph)
	Ivy_ObjReplace(p, root, newRoot, 0, 0, updateLevels)
	for _, fanin := range fanins {
		if fanin.Type != IVY_NONE && fanin.nRefs == 0 {
			Ivy_ObjDelete_rec(p, fanin, 1)
		}
	}
	assert(gain <= oldCount-Ivy_ManNodeNum(p), "node gain is no greater than actual reduction")
}
