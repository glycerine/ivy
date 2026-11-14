package abci

import (
	"fmt"
	"os"
	"time"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Abc_ManRef_t struct {
	nNodeSizeMax, nConeSizeMax, fVerbose                                                int32
	vVars, vFuncs, vVisited, vLeaves                                                    *Vec_Ptr_t
	vMemory                                                                             []uint32
	vCube                                                                               *vec.Vec_Str_t
	nLastGain, nNodesConsidered, nNodesRefactored, nNodesGained                         int32
	nNodesBeg, nNodesEnd                                                                int32
	timeCut, timeTru, timeDcs, timeSop, timeFact, timeEval, timeRes, timeNtk, timeTotal int64
	vMffc                                                                               []*abc.Object
	vMffcMarks                                                                          map[*abc.Object]bool
}

func refactorClearMffc(p *Abc_ManRef_t) {
	for _, node := range p.vMffc {
		node.MarkA = p.vMffcMarks[node]
	}
	p.vMffc = nil
	p.vMffcMarks = nil
}
func refactorLabelMffc(p *Abc_ManRef_t, root *Abc_Obj_t, leaves *Vec_Ptr_t) int32 {
	refactorClearMffc(p)
	cutLeaves := make([]*abc.Object, Vec_PtrSize(leaves))
	for i := range cutLeaves {
		cutLeaves[i], _ = Vec_PtrEntry(leaves, int32(i)).(*Abc_Obj_t)
	}
	var cone []*abc.Object
	size := abc.Abc_NodeMffcInside(root, cutLeaves, &cone)
	p.vMffcMarks = make(map[*abc.Object]bool, len(cone))
	for _, node := range cone {
		p.vMffcMarks[node] = node.MarkA
		node.MarkA = true
	}
	p.vMffc = cone
	return size
}

func refTruthWords(vars int32) int { return int(util.Abc_Truth6WordNum(vars)) }
func refVariableTruth(vars, index int32) []uint64 {
	words := refTruthWords(vars)
	truth := make([]uint64, words)
	for bit := 0; bit < words*64; bit++ {
		if (bit>>index)&1 != 0 {
			truth[bit>>6] |= uint64(1) << uint(bit&63)
		}
	}
	return truth
}
func refTruthFromKit(truth []uint64, vars int32) []uint32 {
	words := 1
	if vars > 5 {
		words = 1 << uint(vars-5)
	}
	result := make([]uint32, words)
	for i := range result {
		w := truth[i>>1]
		if i&1 != 0 {
			w >>= 32
		}
		result[i] = uint32(w)
	}
	return result
}
func refactorCollectCone(root *Abc_Obj_t, leaves, visited *Vec_Ptr_t) {
	network := root.Network()
	marks := make([]bool, int(abc.Abc_NtkObjNumMax(network)))
	for i := range marks {
		if node := network.Object(i); node != nil {
			marks[i] = node.MarkA
			node.MarkA = false
		}
	}
	Abc_NodeConeCollect([]*Abc_Obj_t{root}, leaves, visited, 0)
	for i, marked := range marks {
		if node := network.Object(i); node != nil {
			node.MarkA = marked
		}
	}
}
func refKitToDec(g *kit.Kit_Graph_t) *dec.Dec_Graph_t {
	if kit.Kit_GraphIsConst(g) != 0 {
		if kit.Kit_GraphIsConst0(g) != 0 {
			return dec.Dec_GraphCreateConst0()
		}
		return dec.Dec_GraphCreateConst1()
	}
	root := kit.Kit_GraphRootEdge(g)
	rootBits := kit.Kit_EdgeToInt(root)
	if kit.Kit_GraphIsVar(g) != 0 {
		return dec.Dec_GraphCreateLeaf(int32(rootBits>>1), kit.Kit_GraphLeaveNum(g), int32(rootBits&1))
	}
	out := dec.Dec_GraphCreate(kit.Kit_GraphLeaveNum(g))
	for i := kit.Kit_GraphLeaveNum(g); i < kit.Kit_GraphLeaveNum(g)+kit.Kit_GraphNodeNum(g); i++ {
		a, b := kit.Kit_NodeEdges(kit.Kit_GraphNode(g, i))
		x, y := kit.Kit_EdgeToInt(a), kit.Kit_EdgeToInt(b)
		dec.Dec_GraphAddNodeAnd(out, dec.Dec_EdgeCreate(int32(x>>1), int32(x&1)), dec.Dec_EdgeCreate(int32(y>>1), int32(y&1)))
	}
	dec.Dec_GraphSetRoot(out, dec.Dec_EdgeCreate(int32(rootBits>>1), int32(rootBits&1)))
	return out
}

// Abc_NodeConeTruth returns the exact truth table of root relative to leaves.
func Abc_NodeConeTruth(vVars, vFuncs *Vec_Ptr_t, nWordsMax int32, root *Abc_Obj_t, leaves, visited *Vec_Ptr_t) []uint64 {
	refactorCollectCone(root, leaves, visited)
	functions := make(map[*Abc_Obj_t][]uint64, Vec_PtrSize(leaves)+Vec_PtrSize(visited))
	for i := int32(0); i < Vec_PtrSize(leaves); i++ {
		leaf, _ := Vec_PtrEntry(leaves, i).(*Abc_Obj_t)
		variable, _ := Vec_PtrEntry(vVars, i).([]uint64)
		functions[leaf] = variable
	}
	for Vec_PtrSize(vFuncs) < Vec_PtrSize(visited) {
		Vec_PtrPush(vFuncs, make([]uint64, nWordsMax))
	}
	words := refTruthWords(Vec_PtrSize(leaves))
	var result []uint64
	for i := int32(0); i < Vec_PtrSize(visited); i++ {
		node, _ := Vec_PtrEntry(visited, i).(*Abc_Obj_t)
		if node == nil || Abc_ObjIsNode(node) == 0 {
			continue
		}
		left, right := functions[Abc_ObjFanin0(node)], functions[Abc_ObjFanin1(node)]
		truth, _ := Vec_PtrEntry(vFuncs, i).([]uint64)
		if len(left) == 0 || len(right) == 0 || len(truth) < words {
			return nil
		}
		for k := 0; k < words; k++ {
			a, b := left[k], right[k]
			if abc.Abc_ObjFaninC0(node) != 0 {
				a = ^a
			}
			if abc.Abc_ObjFaninC1(node) != 0 {
				b = ^b
			}
			truth[k] = a & b
		}
		functions[node] = truth
		result = truth
	}
	return result
}
func Abc_NodeConeIsConst0(truth []uint64, vars int32) int32 {
	for _, word := range truth[:refTruthWords(vars)] {
		if word != 0 {
			return 0
		}
	}
	return 1
}
func Abc_NodeConeIsConst1(truth []uint64, vars int32) int32 {
	for _, word := range truth[:refTruthWords(vars)] {
		if word != ^uint64(0) {
			return 0
		}
	}
	return 1
}

func Abc_NodeRefactor(p *Abc_ManRef_t, node *Abc_Obj_t, fanins *Vec_Ptr_t, updateLevel, useZeros, useDcs, verbose int32) *dec.Dec_Graph_t {
	p.nNodesConsidered++
	required := int32(util.ABC_INFINITY)
	if updateLevel != 0 {
		required = abc.Abc_ObjRequiredLevel(node)
	}
	started := time.Now()
	truth := Abc_NodeConeTruth(p.vVars, p.vFuncs, int32(refTruthWords(p.nNodeSizeMax)), node, fanins, p.vVisited)
	p.timeTru += int64(time.Since(started))
	if truth == nil {
		return nil
	}
	vars := Vec_PtrSize(fanins)
	if Abc_NodeConeIsConst0(truth, vars) != 0 || Abc_NodeConeIsConst1(truth, vars) != 0 {
		p.nLastGain = refactorLabelMffc(p, node, fanins)
		p.nNodesGained += p.nLastGain
		p.nNodesRefactored++
		if Abc_NodeConeIsConst0(truth, vars) != 0 {
			return dec.Dec_GraphCreateConst0()
		}
		return dec.Dec_GraphCreateConst1()
	}
	started = time.Now()
	kitGraph := kit.Kit_TruthToGraph(refTruthFromKit(truth, vars), vars, &p.vMemory)
	p.timeFact += int64(time.Since(started))
	if kitGraph == nil {
		return nil
	}
	graph := refKitToDec(kitGraph)
	kit.Kit_GraphFree(kitGraph)
	for i := int32(0); i < vars; i++ {
		fanin, _ := Vec_PtrEntry(fanins, i).(*Abc_Obj_t)
		dec.Dec_GraphSetLeafFunc(graph, i, fanin.Signal())
	}
	abc.Abc_NtkIncrementTravId(node.Network())
	saved := refactorLabelMffc(p, node, fanins)
	started = time.Now()
	added := dec.Dec_GraphToNetworkCount(abc.DecNetwork{Network: node.Network()}, node.Signal(), graph, saved, required)
	p.timeEval += int64(time.Since(started))
	if added == -1 || (added == saved && useZeros == 0) {
		dec.Dec_GraphFree(graph)
		refactorClearMffc(p)
		return nil
	}
	p.nLastGain = saved - added
	p.nNodesGained += p.nLastGain
	p.nNodesRefactored++
	return graph
}

func Abc_NtkManRefStart(nodeSizeMax, coneSizeMax, useDcs, verbose int32) *Abc_ManRef_t {
	p := &Abc_ManRef_t{nNodeSizeMax: nodeSizeMax, nConeSizeMax: coneSizeMax, fVerbose: verbose}
	p.vCube = vec.Vec_StrAlloc(100)
	p.vVisited = Vec_PtrAlloc(100)
	p.vFuncs = Vec_PtrAlloc(100)
	p.vVars = Vec_PtrAlloc(0)
	varCount := nodeSizeMax
	if varCount < 6 {
		varCount = 6
	}
	for i := int32(0); i < varCount; i++ {
		Vec_PtrPush(p.vVars, refVariableTruth(varCount, i))
	}
	p.vMemory = make([]uint32, 0, 1<<16)
	return p
}
func Abc_NtkManRefStop(p *Abc_ManRef_t) {
	if p == nil {
		return
	}
	refactorClearMffc(p)
	for i := 0; i < int(Vec_PtrSize(p.vFuncs)); i++ {
		vec.Vec_PtrWriteEntry(p.vFuncs, i, nil)
	}
	Vec_PtrFree(p.vFuncs)
	Vec_PtrFree(p.vVars)
	Vec_PtrFree(p.vVisited)
	vec.Vec_StrFree(p.vCube)
	p.vFuncs, p.vVars, p.vVisited = nil, nil, nil
	p.vCube = nil
	p.vMemory = nil
}
func Abc_NtkManRefPrintStats(p *Abc_ManRef_t) {
	fmt.Println("Refactoring statistics:")
	fmt.Printf("Nodes considered  = %8d.\n", p.nNodesConsidered)
	fmt.Printf("Nodes refactored  = %8d.\n", p.nNodesRefactored)
	gain := p.nNodesBeg - p.nNodesEnd
	pct := 0.0
	if p.nNodesBeg != 0 {
		pct = 100 * float64(gain) / float64(p.nNodesBeg)
	}
	fmt.Printf("Gain              = %8d. (%6.2f %%).\n", gain, pct)
	for _, stat := range []struct {
		name    string
		elapsed int64
	}{{"Cuts", p.timeCut}, {"Resynthesis", p.timeRes}, {"BDD", p.timeTru}, {"DCs", p.timeDcs}, {"SOP", p.timeSop}, {"FF", p.timeFact}, {"Eval", p.timeEval}, {"AIG update", p.timeNtk}, {"TOTAL", p.timeTotal}} {
		fmt.Printf("%-12s %9.2f sec\n", stat.name, time.Duration(stat.elapsed).Seconds())
	}
}

func Abc_NtkRefactor(n *Abc_Ntk_t, nodeSizeMax, coneSizeMax, updateLevel, useZeros, useDcs, verbose int32) int32 {
	if abc.Abc_NtkIsStrash(n) == 0 {
		panic("Abc_NtkRefactor requires a strashed network")
	}
	started := time.Now()
	manager, ok := n.FunctionManager.(*abc.Abc_Aig_t)
	if !ok || manager == nil {
		panic("Abc_NtkRefactor requires AIG manager")
	}
	abc.Abc_AigCleanup(manager)
	cut := Abc_NtkManCutStart(nodeSizeMax, coneSizeMax, 2, 1000)
	ref := Abc_NtkManRefStart(nodeSizeMax, coneSizeMax, useDcs, verbose)
	ref.vLeaves = Abc_NtkManCutReadCutLarge(cut)
	if updateLevel != 0 {
		abc.Abc_NtkStartReverseLevels(n, 0)
	}
	ref.nNodesBeg = abc.Abc_NtkNodeNum(n)
	limit := abc.Abc_NtkObjNumMax(n)
	progress := extra.Extra_ProgressBarStart(os.Stdout, limit)
	for i := int32(0); i < limit; i++ {
		node := n.Object(int(i))
		if node == nil || abc.Abc_ObjIsNode(node) == 0 {
			continue
		}
		extra.Extra_ProgressBarUpdate(progress, i, "")
		if node.Persist || abc.Abc_ObjFanoutNum(node) > 1000 {
			continue
		}
		clk := time.Now()
		fanins := Abc_NodeFindCut(cut, node, useDcs)
		ref.timeCut += int64(time.Since(clk))
		clk = time.Now()
		graph := Abc_NodeRefactor(ref, node, fanins, updateLevel, useZeros, useDcs, verbose)
		ref.timeRes += int64(time.Since(clk))
		if graph == nil {
			continue
		}
		clk = time.Now()
		Dec_GraphUpdateNetwork(node, graph, updateLevel, ref.nLastGain)
		ref.timeNtk += int64(time.Since(clk))
		refactorClearMffc(ref)
		dec.Dec_GraphFree(graph)
	}
	extra.Extra_ProgressBarStop(progress)
	ref.timeTotal = int64(time.Since(started))
	ref.nNodesEnd = abc.Abc_NtkNodeNum(n)
	if verbose != 0 {
		Abc_NtkManRefPrintStats(ref)
	}
	Abc_NtkManCutStop(cut)
	Abc_NtkManRefStop(ref)
	abc.Abc_NtkReassignIds(n)
	if updateLevel != 0 {
		abc.Abc_NtkStopReverseLevels(n)
	} else {
		abc.Abc_NtkLevel(n)
	}
	if abc.Abc_NtkCheck(n) == 0 {
		fmt.Println("Abc_NtkRefactor: The network check has failed.")
		return 0
	}
	return 1
}
