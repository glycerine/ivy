// Sequential cut enumeration and rewriting for Ivy networks.
package ivy

import (
	"fmt"
	"sort"
	"time"

	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/opt/rwt"
)

// Ivy_CutHashValue returns one bit for an encoded cut leaf.
func Ivy_CutHashValue(leaf int32) int32 { return int32(uint32(1) << uint32(leaf%31)) }

// Ivy_ManRewriteSeq rewrites logic cones using cuts that may cross latch
// boundaries. Latches remain explicit edges in the resulting decomposition.
func Ivy_ManRewriteSeq(p *Ivy_Man_t, useZeroCost, verbose int32) int32 {
	if p == nil || p.memoryStopped {
		panic("ivy: sequential rewrite on absent or stopped manager")
	}
	for _, node := range p.vObjs {
		if node != nil && Ivy_ObjIsLatch(node) != 0 {
			node.Init = IVY_INIT_DC
		}
	}
	manager := rwt.Rwt_ManStart(0)
	if manager == nil {
		return 0
	}
	p.pData = manager
	start := time.Now()
	if p.fFanout == 0 {
		Ivy_ManStartFanout(p)
	}
	lastID := Ivy_ManObjIdMax(p)
	for index, node := range p.vObjs {
		if node == nil || Ivy_ObjIsNode(node) == 0 || node.Id > lastID {
			continue
		}
		if index > int(lastID) {
			break
		}
		assert(Ivy_ObjIsBuf(node) == 0 && Ivy_ObjIsBuf(Ivy_ObjFanin0(node)) == 0 && Ivy_ObjIsBuf(Ivy_ObjFanin1(node)) == 0, "sequential rewrite has no buffer fanins")
		gain := ivyNodeRewriteSeq(p, manager, node, useZeroCost)
		if gain < 0 || (gain == 0 && useZeroCost == 0) {
			continue
		}
		graph := rwt.Rwt_ManReadDecs(manager)
		if graph == nil {
			panic("ivy: sequential rewrite selected no decomposition")
		}
		complemented := rwt.Rwt_ManReadCompl(manager) != 0
		if complemented {
			dec.Dec_GraphComplement(graph)
		}
		updateStart := time.Now()
		ivyGraphUpdateNetworkSeq(p, node, graph, gain)
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
	Ivy_ManResetLevels(p)
	if Ivy_ManCheck(p) == 0 {
		fmt.Println("Ivy_ManRewriteSeq(): The check has failed.")
	}
	return 1
}

func ivyNodeRewriteSeq(p *Ivy_Man_t, manager *rwt.Rwt_Man_t, root *Ivy_Obj_t, useZeroCost int32) int32 {
	rwt.Rwt_ManRecordNodeConsidered(manager)
	cutStart := time.Now()
	store := Ivy_CutComputeForNode(p, root, 5)
	rwt.Rwt_ManAddTimeCuts(manager, time.Since(cutStart))
	bestGain := int32(-1)
	var bestGraph *dec.Dec_Graph_t
	var bestFanins []any
	var bestTruth uint16
	var bestCompl bool
	for cutIndex := int32(1); cutIndex < store.nCuts; cutIndex++ {
		cut := &store.pCuts[cutIndex]
		if cut.nSize != 4 {
			continue
		}
		bad := false
		for _, leaf := range cut.pArray[:cut.nSize] {
			if Ivy_ObjIsBuf(Ivy_ManObj(p, Ivy_LeafId(leaf))) != 0 {
				bad = true
				break
			}
		}
		rwt.Rwt_ManRecordCut(manager, !bad)
		if bad {
			continue
		}
		truthStart := time.Now()
		truth := uint16(Ivy_CutGetTruth(p, root, cut.pArray[:cut.nSize], int32(cut.nSize)))
		class, phase, permutation := rwt.Rwt_ManCanonInfo(manager, truth)
		rwt.Rwt_ManAddTimeTruth(manager, time.Since(truthStart))
		if len(permutation) != 4 {
			continue
		}
		fanins := make([]any, 4)
		for i := 0; i < 4; i++ {
			leaf := cut.pArray[permutation[i]]
			fanin := Ivy_ManObj(p, Ivy_LeafId(leaf))
			fanins[i] = Ivy_NotCond(fanin, int32(phase>>i)&1)
		}
		mffcStart := time.Now()
		for _, fanin := range fanins {
			Ivy_ObjRefsInc(Ivy_Regular(fanin.(*Ivy_Obj_t)))
		}
		Ivy_ManIncrementTravId(p)
		nodesSaved := Ivy_ObjMffcLabel(p, root)
		for _, fanin := range fanins {
			Ivy_ObjRefsDec(Ivy_Regular(fanin.(*Ivy_Obj_t)))
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
			for i, fanin := range fanins {
				dec.Dec_GraphSetLeafFunc(graph, int32(i), fanin)
			}
			ivyGraphPrepareSeq(graph, cut, permutation)
			added := ivyGraphToNetworkCountSeq(p, root, graph, nodesSaved)
			if added < 0 || added > nodesSaved {
				continue
			}
			gain := nodesSaved - added
			if gain > bestGain {
				bestGain, bestGraph, bestFanins, bestTruth = gain, graph, append([]any(nil), fanins...), truth
				bestCompl = phase&(1<<4) != 0
				if rwt.Rwt_IsComplement(candidate) != 0 {
					bestCompl = !bestCompl
				}
			}
		}
		rwt.Rwt_ManAddTimeEval(manager, time.Since(evalStart))
	}
	if bestGain < 0 || bestGraph == nil || (bestGain == 0 && useZeroCost == 0) {
		return -1
	}
	rwt.Rwt_ManStoreBest(manager, bestGraph, bestCompl, bestFanins, bestTruth, bestGain, useZeroCost != 0)
	return bestGain
}

var ivySeqTruthMasks = [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}

// Ivy_CutGetTruth evaluates a node over encoded leaves (object ID plus latch
// depth), preserving edge polarity while walking the AIG.
func Ivy_CutGetTruth(p *Ivy_Man_t, root *Ivy_Obj_t, leaves []int32, count int32) uint32 {
	if p == nil || root == nil || count < 0 || count > int32(len(ivySeqTruthMasks)) || int(count) > len(leaves) {
		panic("ivy: invalid sequential cut truth request")
	}
	if Ivy_ObjIsNode(root) == 0 {
		panic("ivy: sequential cut truth root must be a logic node")
	}
	return Ivy_CutGetTruth_rec(p, Ivy_LeafCreate(Ivy_ObjId(root), 0), leaves[:count], count)
}

// Ivy_CutGetTruth_rec evaluates a sequential leaf through latch and logic nodes.
func Ivy_CutGetTruth_rec(p *Ivy_Man_t, leaf int32, leaves []int32, count int32) uint32 {
	if count < 0 || int(count) > len(leaves) || count > int32(len(ivySeqTruthMasks)) {
		panic("ivy: invalid sequential truth leaf count")
	}
	return ivyCutTruthSeq(p, leaf, leaves[:count], make(map[int32]uint32))
}

func ivyCutTruthSeq(p *Ivy_Man_t, leaf int32, leaves []int32, cache map[int32]uint32) uint32 {
	for i, cutLeaf := range leaves {
		if leaf == cutLeaf {
			return ivySeqTruthMasks[i]
		}
	}
	if truth, ok := cache[leaf]; ok {
		return truth
	}
	obj := Ivy_ManObj(p, Ivy_LeafId(leaf))
	if obj == nil {
		panic("ivy: cut references a missing object")
	}
	var result uint32
	switch {
	case Ivy_ObjIsConst1(obj) != 0:
		result = ^uint32(0)
	case Ivy_ObjIsLatch(obj) != 0:
		if Ivy_ObjFaninC0(obj) != 0 {
			panic("ivy: complemented latch input")
		}
		result = ivyCutTruthSeq(p, Ivy_LeafCreate(Ivy_ObjFaninId0(obj), Ivy_LeafLat(leaf)+1), leaves, cache)
	case Ivy_ObjIsBuf(obj) != 0:
		result = ivyCutTruthSeq(p, Ivy_LeafCreate(Ivy_ObjFaninId0(obj), Ivy_LeafLat(leaf)), leaves, cache)
		if Ivy_ObjFaninC0(obj) != 0 {
			result = ^result
		}
	case Ivy_ObjIsNode(obj) != 0:
		left := ivyCutTruthSeq(p, Ivy_LeafCreate(Ivy_ObjFaninId0(obj), Ivy_LeafLat(leaf)), leaves, cache)
		right := ivyCutTruthSeq(p, Ivy_LeafCreate(Ivy_ObjFaninId1(obj), Ivy_LeafLat(leaf)), leaves, cache)
		if Ivy_ObjFaninC0(obj) != 0 {
			left = ^left
		}
		if Ivy_ObjFaninC1(obj) != 0 {
			right = ^right
		}
		if Ivy_ObjIsExor(obj) != 0 {
			result = left ^ right
		} else {
			result = left & right
		}
	default:
		panic("ivy: sequential cut reached an input absent from the cut")
	}
	cache[leaf] = result
	return result
}

// Ivy_CutComputeForNode enumerates non-dominated sequential cuts for a node.
// Latches are folded into leaf-depth tags, while PIs and constants terminate
// expansion. Each call owns its store, avoiding C's shared static workspace.
func Ivy_CutComputeForNode(p *Ivy_Man_t, root *Ivy_Obj_t, maxLeaves int32) *Ivy_Store_t {
	if p == nil || root == nil || maxLeaves < 1 || maxLeaves > IVY_CUT_INPUT || Ivy_ObjIsNode(root) == 0 {
		panic("ivy: invalid sequential cut request")
	}
	store := &Ivy_Store_t{nCutsMax: IVY_CUT_LIMIT}
	initial := Ivy_Cut_t{nSize: 1, nSizeMax: int16(maxLeaves)}
	initial.pArray[0] = Ivy_LeafCreate(Ivy_ObjId(root), 0)
	initial.uHash = uint32(Ivy_CutHashValue(initial.pArray[0]))
	store.pCuts[0], store.nCuts = initial, 1
	var next Ivy_Cut_t
	for ci := int32(0); ci < store.nCuts && store.nCuts < store.nCutsMax; ci++ {
		cut := &store.pCuts[ci]
		if cut.nSize == 0 {
			continue
		}
		for pos := int32(0); pos < int32(cut.nSize); pos++ {
			encoded := cut.pArray[pos]
			obj := Ivy_ManObj(p, Ivy_LeafId(encoded))
			if obj == nil || Ivy_ObjIsCi(obj) != 0 || Ivy_ObjIsConst1(obj) != 0 {
				continue
			}
			if Ivy_ObjIsNode(obj) == 0 && Ivy_ObjIsBuf(obj) == 0 {
				continue
			}
			base0 := ivyCutReadLeaf(p, Ivy_ObjFanin0(obj))
			base1 := base0
			if Ivy_ObjIsBuf(obj) == 0 {
				base1 = ivyCutReadLeaf(p, Ivy_ObjFanin1(obj))
			}
			latency := Ivy_LeafLat(encoded)
			if latency+Ivy_LeafLat(base0) >= IVY_LEAF_MASK || latency+Ivy_LeafLat(base1) >= IVY_LEAF_MASK {
				continue
			}
			leaf0 := Ivy_LeafCreate(Ivy_LeafId(base0), latency+Ivy_LeafLat(base0))
			leaf1 := Ivy_LeafCreate(Ivy_LeafId(base1), latency+Ivy_LeafLat(base1))
			if leaf0 > leaf1 {
				leaf0, leaf1 = leaf1, leaf0
			}
			if !ivyCutPrescreenSeq(cut, leaf0, leaf1) {
				continue
			}
			if !ivyCutDeriveSeq(cut, int(pos), leaf0, leaf1, &next) {
				continue
			}
			ivyCutFindOrAddSeq(store, &next)
			if store.nCuts == store.nCutsMax {
				break
			}
		}
	}
	store.fSatur = ivyBool(store.nCuts == store.nCutsMax)
	kept := int32(0)
	for i := int32(0); i < store.nCuts; i++ {
		cut := store.pCuts[i]
		if cut.nSize == 0 {
			continue
		}
		if cut.nSize < cut.nSizeMax {
			store.nCutsM++
		}
		store.pCuts[kept] = cut
		kept++
	}
	clear(store.pCuts[kept:store.nCuts])
	store.nCuts = kept
	return store
}

func ivyCutReadLeaf(p *Ivy_Man_t, fanin *Ivy_Obj_t) int32 {
	if fanin == nil || Ivy_IsComplement(fanin) != 0 {
		panic("ivy: cut fanin must be a regular object")
	}
	depth := int32(0)
	for Ivy_ObjIsLatch(fanin) != 0 {
		if Ivy_ObjFaninC0(fanin) != 0 {
			panic("ivy: complemented latch input")
		}
		depth++
		fanin = Ivy_ObjFanin0(fanin)
	}
	if depth >= IVY_LEAF_MASK {
		panic("ivy: cut latch depth exceeds leaf encoding")
	}
	return Ivy_LeafCreate(Ivy_ObjId(fanin), depth)
}

func ivyCutPrescreenSeq(cut *Ivy_Cut_t, a, b int32) bool {
	if cut.nSize < cut.nSizeMax {
		return true
	}
	for _, leaf := range cut.pArray[:cut.nSize] {
		if leaf == a || leaf == b {
			return true
		}
	}
	return false
}

func ivyCutDeriveSeq(cut *Ivy_Cut_t, replaced int, a, b int32, out *Ivy_Cut_t) bool {
	leaves := make([]int32, 0, int(cut.nSize)+1)
	for i, leaf := range cut.pArray[:cut.nSize] {
		if i != replaced {
			leaves = append(leaves, leaf)
		}
	}
	leaves = append(leaves, a, b)
	sort.Slice(leaves, func(i, j int) bool { return leaves[i] < leaves[j] })
	unique := leaves[:0]
	for _, leaf := range leaves {
		if len(unique) == 0 || unique[len(unique)-1] != leaf {
			unique = append(unique, leaf)
		}
	}
	if len(unique) > int(cut.nSizeMax) || len(unique) > len(out.pArray) {
		return false
	}
	*out = Ivy_Cut_t{nSize: int16(len(unique)), nSizeMax: cut.nSizeMax}
	for i, leaf := range unique {
		out.pArray[i] = leaf
		out.uHash |= uint32(Ivy_CutHashValue(leaf))
	}
	return true
}

func ivyCutFindOrAddSeq(store *Ivy_Store_t, next *Ivy_Cut_t) bool {
	if next.nSize <= 0 || next.uHash == 0 {
		panic("ivy: invalid empty cut")
	}
	for i := int32(0); i < store.nCuts; i++ {
		cut := &store.pCuts[i]
		if cut.nSize == 0 {
			continue
		}
		if cut.nSize <= next.nSize && cut.uHash&next.uHash == cut.uHash && ivyCutSubsetSeq(cut, next) {
			return false
		}
	}
	for i := int32(0); i < store.nCuts; i++ {
		cut := &store.pCuts[i]
		if cut.nSize > next.nSize && cut.uHash&next.uHash == next.uHash && ivyCutSubsetSeq(next, cut) {
			cut.nSize = 0
		}
	}
	if store.nCuts == store.nCutsMax {
		return false
	}
	store.pCuts[store.nCuts] = *next
	store.nCuts++
	return true
}

func ivyCutSubsetSeq(a, b *Ivy_Cut_t) bool {
	i, j := 0, 0
	for i < int(a.nSize) && j < int(b.nSize) {
		if a.pArray[i] == b.pArray[j] {
			i++
			j++
		} else if a.pArray[i] > b.pArray[j] {
			j++
		} else {
			return false
		}
	}
	return i == int(a.nSize)
}

func ivyGraphPrepareSeq(graph *dec.Dec_Graph_t, cut *Ivy_Cut_t, permutation []byte) {
	if graph == nil || int32(len(permutation)) != int32(cut.nSize) || dec.Dec_GraphLeaveNum(graph) != int32(cut.nSize) {
		panic("ivy: sequential graph/cut leaf mismatch")
	}
	for i := int32(0); i < int32(cut.nSize); i++ {
		leaf := cut.pArray[permutation[i]]
		node := dec.Dec_GraphNode(graph, i)
		node.SetLatches([3]uint32{0, 0, uint32(Ivy_LeafLat(leaf))})
	}
	for i := int32(cut.nSize); i < int32(cut.nSize)+dec.Dec_GraphNodeNum(graph); i++ {
		node := dec.Dec_GraphNode(graph, i)
		e0, e1 := node.Edges()
		lat0 := dec.Dec_GraphNode(graph, int32(e0.Node)).Latches()[2]
		lat1 := dec.Dec_GraphNode(graph, int32(e1.Node)).Latches()[2]
		out := min(lat0, lat1)
		node.SetLatches([3]uint32{lat0 - out, lat1 - out, out})
	}
}

func ivyGraphToNetworkCountSeq(p *Ivy_Man_t, root *Ivy_Obj_t, graph *dec.Dec_Graph_t, nodeMax int32) int32 {
	if dec.Dec_GraphIsConst(graph) != 0 || dec.Dec_GraphIsVar(graph) != 0 {
		return 0
	}
	counter := int32(0)
	first := dec.Dec_GraphLeaveNum(graph)
	for i := first; i < first+dec.Dec_GraphNodeNum(graph); i++ {
		node := dec.Dec_GraphNode(graph, i)
		e0, e1 := node.Edges()
		left := ivyGraphNodeSignal(graph, int32(e0.Node))
		right := ivyGraphNodeSignal(graph, int32(e1.Node))
		latches := node.Latches()
		left = ivyGraphCountLatches(p, left, int32(latches[0]))
		right = ivyGraphCountLatches(p, right, int32(latches[1]))
		if left != nil {
			left = Ivy_NotCond(left, ivyBool(e0.Complemented()))
		}
		if right != nil {
			right = Ivy_NotCond(right, ivyBool(e1.Complemented()))
		}
		var result *Ivy_Obj_t
		if left != nil && right != nil {
			if Ivy_Regular(left) == Ivy_Regular(right) || Ivy_ObjIsConst1(Ivy_Regular(left)) != 0 || Ivy_ObjIsConst1(Ivy_Regular(right)) != 0 {
				result = Ivy_And(p, left, right)
			} else {
				result = Ivy_TableLookup(p, Ivy_ObjCreateGhost(p, left, right, IVY_AND, IVY_INIT_NONE))
			}
			if Ivy_Regular(result) == root {
				return -1
			}
		}
		if result == nil || Ivy_ObjIsTravIdCurrent(p, Ivy_Regular(result)) != 0 {
			counter++
			if counter > nodeMax {
				return -1
			}
		}
		dec.Dec_GraphSetNodeFunc(graph, i, result)
	}
	return counter
}

func ivyGraphNodeSignal(graph *dec.Dec_Graph_t, index int32) *Ivy_Obj_t {
	payload := dec.Dec_GraphNodeFunc(graph, index)
	if payload == nil {
		return nil
	}
	signal, ok := payload.(*Ivy_Obj_t)
	if !ok {
		panic("ivy: malformed sequential graph payload")
	}
	return signal
}

func ivyGraphCountLatches(p *Ivy_Man_t, signal *Ivy_Obj_t, count int32) *Ivy_Obj_t {
	if signal == nil {
		return nil
	}
	for i := int32(0); i < count; i++ {
		complemented := Ivy_IsComplement(signal)
		signal = Ivy_TableLookup(p, Ivy_ObjCreateGhost(p, Ivy_Regular(signal), nil, IVY_LATCH, IVY_INIT_DC))
		if signal != nil {
			signal = Ivy_NotCond(signal, complemented)
		}
	}
	return signal
}

func ivyGraphToNetworkSeq(p *Ivy_Man_t, graph *dec.Dec_Graph_t) *Ivy_Obj_t {
	if dec.Dec_GraphIsConst(graph) != 0 {
		return Ivy_NotCond(Ivy_ManConst1(p), dec.Dec_GraphIsComplement(graph))
	}
	if dec.Dec_GraphIsVar(graph) != 0 {
		root := dec.Dec_GraphRootEdge(graph)
		signal := ivyGraphNodeSignal(graph, int32(root.Node))
		leaf := dec.Dec_GraphNode(graph, int32(root.Node))
		signal = ivyGraphBuildLatches(p, signal, leaf.Latches()[2])
		return Ivy_NotCond(signal, ivyBool(root.Complemented()))
	}
	first := dec.Dec_GraphLeaveNum(graph)
	last := first + dec.Dec_GraphNodeNum(graph)
	for i := first; i < last; i++ {
		node := dec.Dec_GraphNode(graph, i)
		e0, e1 := node.Edges()
		left := Ivy_NotCond(ivyGraphNodeSignal(graph, int32(e0.Node)), ivyBool(e0.Complemented()))
		right := Ivy_NotCond(ivyGraphNodeSignal(graph, int32(e1.Node)), ivyBool(e1.Complemented()))
		latches := node.Latches()
		left = ivyGraphBuildLatches(p, left, latches[0])
		right = ivyGraphBuildLatches(p, right, latches[1])
		dec.Dec_GraphSetNodeFunc(graph, i, Ivy_And(p, left, right))
	}
	root := dec.Dec_GraphRootEdge(graph)
	signal := ivyGraphNodeSignal(graph, int32(root.Node))
	signal = ivyGraphBuildLatches(p, signal, dec.Dec_GraphNode(graph, int32(root.Node)).Latches()[2])
	return Ivy_NotCond(signal, ivyBool(root.Complemented()))
}

func ivyGraphBuildLatches(p *Ivy_Man_t, signal *Ivy_Obj_t, count uint32) *Ivy_Obj_t {
	for i := uint32(0); i < count; i++ {
		signal = Ivy_Latch(p, signal, IVY_INIT_DC)
	}
	return signal
}

func ivyGraphUpdateNetworkSeq(p *Ivy_Man_t, root *Ivy_Obj_t, graph *dec.Dec_Graph_t, gain int32) {
	oldCount := Ivy_ManNodeNum(p)
	newRoot := ivyGraphToNetworkSeq(p, graph)
	Ivy_ObjReplace(p, root, newRoot, 1, 0, 0)
	assert(gain <= oldCount-Ivy_ManNodeNum(p), "sequential rewrite gain does not exceed actual reduction")
	if len(p.vBufs) != 0 {
		if p.fFanout == 0 {
			Ivy_ManStartFanout(p)
		}
		Ivy_ManPropagateBuffers(p, 0)
	}
}

// Ivy_CutPrescreen reports whether replacing a leaf can fit in the cut limit.
func Ivy_CutPrescreen(cut *Ivy_Cut_t, a, b int32) int32 {
	if cut == nil {
		return 0
	}
	return ivyBool(ivyCutPrescreenSeq(cut, a, b))
}

// Ivy_CutDeriveNew derives a sorted cut by replacing old with a and b.
func Ivy_CutDeriveNew(cut, out *Ivy_Cut_t, old, a, b int32) int32 {
	if cut == nil || out == nil {
		return 0
	}
	index := -1
	for i, leaf := range cut.pArray[:cut.nSize] {
		if leaf == old {
			index = i
			break
		}
	}
	if index < 0 {
		return 0
	}
	if a > b {
		a, b = b, a
	}
	return ivyBool(ivyCutDeriveSeq(cut, index, a, b, out))
}

// Ivy_CutDeriveNew2 and Ivy_CutDeriveNew3 preserve the historical entry points.
func Ivy_CutDeriveNew2(cut, out *Ivy_Cut_t, old, a, b int32) int32 {
	return Ivy_CutDeriveNew(cut, out, old, a, b)
}
func Ivy_CutDeriveNew3(cut, out *Ivy_Cut_t, old, a, b int32) int32 {
	return Ivy_CutDeriveNew(cut, out, old, a, b)
}

// Ivy_CutFindOrAddFilter inserts a cut unless an existing subset dominates it.
func Ivy_CutFindOrAddFilter(store *Ivy_Store_t, cut *Ivy_Cut_t) int32 {
	if store == nil || cut == nil {
		return 1
	}
	return ivyBool(!ivyCutFindOrAddSeq(store, cut))
}

// Ivy_CutCompactAll removes cuts invalidated by dominance filtering.
func Ivy_CutCompactAll(store *Ivy_Store_t) {
	if store == nil {
		return
	}
	kept, nontrivial := int32(0), int32(0)
	for i := int32(0); i < store.nCuts; i++ {
		cut := store.pCuts[i]
		if cut.nSize == 0 {
			continue
		}
		store.pCuts[kept] = cut
		kept++
		if cut.nSize < cut.nSizeMax {
			nontrivial++
		}
	}
	clear(store.pCuts[kept:store.nCuts])
	store.nCuts, store.nCutsM = kept, nontrivial
}

// Ivy_CutPrintForNode writes one cut in the compact source format.
func Ivy_CutPrintForNode(cut *Ivy_Cut_t) {
	if cut == nil || cut.nSize <= 0 {
		return
	}
	fmt.Printf("%d : {", cut.nSize)
	for _, leaf := range cut.pArray[:cut.nSize] {
		fmt.Printf(" %d", leaf)
	}
	fmt.Println(" }")
}

// Ivy_CutPrintForNodes prints all cuts in a store.
func Ivy_CutPrintForNodes(store *Ivy_Store_t) {
	if store == nil || store.nCuts == 0 {
		return
	}
	fmt.Printf("Node %d\n", store.pCuts[0].pArray[0])
	for i := int32(0); i < store.nCuts; i++ {
		Ivy_CutPrintForNode(&store.pCuts[i])
	}
}

// Ivy_CutComputeAll reports bounded sequential cut counts for every logic node.
func Ivy_CutComputeAll(p *Ivy_Man_t, maxLeaves int32) {
	if p == nil || p.memoryStopped {
		panic("ivy: cut computation on absent or stopped manager")
	}
	if maxLeaves > IVY_CUT_INPUT {
		fmt.Printf("Cannot compute cuts for more than %d inputs.\n", IVY_CUT_INPUT)
		return
	}
	start := time.Now()
	total, nontrivial, nodes, saturated := -Ivy_ManNodeNum(p), -Ivy_ManNodeNum(p), int32(0), int32(0)
	for _, node := range p.vObjs {
		if node == nil || Ivy_ObjIsNode(node) == 0 {
			continue
		}
		store := Ivy_CutComputeForNode(p, node, maxLeaves)
		total += store.nCuts
		nontrivial += store.nCutsM
		nodes++
		saturated += store.fSatur
	}
	fmt.Printf("All = %6d. Minus = %6d. Triv = %6d.   Node = %6d. Satur = %6d.  Time = %.3fs\n", total, nontrivial, Ivy_ManPiNum(p)+Ivy_ManNodeNum(p), nodes, saturated, time.Since(start).Seconds())
}
