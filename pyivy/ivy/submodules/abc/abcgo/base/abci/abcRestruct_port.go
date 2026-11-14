package abci

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/dsd"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/cut"
)

// ABC's C implementation uses the process-global glibc rand() stream. Its
// initial state is equivalent to srand(1); each restructure manager consumes
// exactly 60 values (three for each of 20 simulation words).
var restructRandMu sync.Mutex
var restructRand = newGlibcRand(1)

func restructRandomWords(rng *glibcRand) []int32 {
	words := make([]int32, 20)
	for i := range words {
		value := (rng.next() << 24) ^ (rng.next() << 12) ^ rng.next()
		words[i] = int32(value)
	}
	return words
}

// Abc_ManRst_t is the working manager used by ABC's DSD restructuring pass.
type Abc_ManRst_t struct {
	pNtk                                       *abc.Network
	nCutMax, fUpdateLevel, fUseZeros, fVerbose int32
	dd                                         *core.DdManager
	pManDsd                                    *dsd.Dsd_Manager_t
	vVisited, vLeaves, vDecs, vTemp            *vec.Vec_Ptr_t
	vSims, vRands, vOnes, vBinate, vTwos       *vec.Vec_Int_t
	nLastGain, nCutsConsidered, nCutsExplored  int32
	nNodesConsidered, nNodesRestructured       int32
	nNodesGained                               int32
	timeCut, timeBdd, timeDsd, timeEval        time.Duration
	timeRes, timeNtk, timeTotal                time.Duration
}

// Abc_RestructNodeDivisors collects the transitive fanout divisors whose two
// fanins are already in the leaf/divisor set. The scan is deliberately
// insertion ordered, matching ABC's dynamic Vec_Ptr traversal.
func Abc_RestructNodeDivisors(p *Abc_ManRst_t, pRoot *Abc_Obj_t, nNodesSaved int32) {
	if p == nil || p.vLeaves == nil || p.vDecs == nil {
		return
	}
	vec.Vec_PtrClear(p.vDecs)
	for i := 0; i < vec.Vec_PtrSize(p.vLeaves); i++ {
		node, _ := vec.Vec_PtrEntry(p.vLeaves, i).(*abc.Object)
		if node == nil {
			continue
		}
		if node.MarkC {
			panic("abc restructure: leaf already carries temporary MarkC")
		}
		vec.Vec_PtrPush(p.vDecs, node)
		node.MarkC = true
	}
	for i := 0; i < vec.Vec_PtrSize(p.vDecs); i++ {
		node, _ := vec.Vec_PtrEntry(p.vDecs, i).(*abc.Object)
		if node == nil {
			continue
		}
		for _, fanout := range node.Fanouts() {
			if fanout.MarkC || abc.Abc_ObjIsPo(fanout) != 0 {
				continue
			}
			if abc.Abc_ObjFanin0(fanout).MarkC && abc.Abc_ObjFanin1(fanout).MarkC {
				vec.Vec_PtrPush(p.vDecs, fanout)
				fanout.MarkC = true
			}
		}
	}
	for i := 0; i < vec.Vec_PtrSize(p.vDecs); i++ {
		node, _ := vec.Vec_PtrEntry(p.vDecs, i).(*abc.Object)
		if node != nil {
			node.MarkC = false
		}
	}
	fmt.Printf("%d\n", vec.Vec_PtrSize(p.vDecs)-int(nNodesSaved)-vec.Vec_PtrSize(p.vLeaves))
	_ = pRoot // Kept in the source signature; the source routine does not read it.
}

// Abc_NtkManRstStart allocates the BDD/DSD engines and the temporary vectors
// used by the original restructuring manager.
func Abc_NtkManRstStart(nCutMax, fUpdateLevel, fUseZeros, fVerbose int32) *Abc_ManRst_t {
	dd := core.NewBDDManager(int(nCutMax))
	if dd == nil {
		return nil
	}
	core.Cudd_zddVarsFromBddVars(dd, 2)
	m := &Abc_ManRst_t{
		nCutMax: nCutMax, fUpdateLevel: fUpdateLevel, fUseZeros: fUseZeros, fVerbose: fVerbose,
		dd: dd, vVisited: vec.Vec_PtrAlloc(100), vLeaves: vec.Vec_PtrAlloc(100),
		vDecs: vec.Vec_PtrAlloc(100), vTemp: vec.Vec_PtrAlloc(100),
		vSims: vec.Vec_IntAlloc(100), vRands: vec.Vec_IntAlloc(20), vOnes: vec.Vec_IntAlloc(100),
		vBinate: vec.Vec_IntAlloc(100), vTwos: vec.Vec_IntAlloc(100),
	}
	m.pManDsd = dsd.Dsd_ManagerStart(dd, len(dd.Vars), 0)
	restructRandMu.Lock()
	for _, value := range restructRandomWords(restructRand) {
		vec.Vec_IntPush(m.vRands, value)
	}
	restructRandMu.Unlock()
	return m
}

// Abc_NtkManRstStop releases the DSD engine, BDD manager, and temporary arrays.
func Abc_NtkManRstStop(m *Abc_ManRst_t) {
	if m == nil {
		return
	}
	dsd.Dsd_ManagerStop(m.pManDsd)
	core.Cudd_Quit(m.dd)
	vec.Vec_PtrFree(m.vDecs)
	vec.Vec_PtrFree(m.vLeaves)
	vec.Vec_PtrFree(m.vVisited)
	vec.Vec_PtrFree(m.vTemp)
	vec.Vec_IntFree(m.vSims)
	vec.Vec_IntFree(m.vOnes)
	vec.Vec_IntFree(m.vBinate)
	vec.Vec_IntFree(m.vTwos)
	vec.Vec_IntFree(m.vRands)
}

// abcNodeConeBdd evaluates the root cone at the ordered cut leaves and returns
// the same unreferenced BDD result as Abc_NodeConeBdd in abcReconv.c.
func abcNodeConeBdd(dd *core.DdManager, root *abc.Object, leaves, visited *vec.Vec_Ptr_t) *core.DdNode {
	markLeaves := func(mark bool) {
		for i := 0; i < vec.Vec_PtrSize(leaves); i++ {
			leaf, _ := vec.Vec_PtrEntry(leaves, i).(*abc.Object)
			if leaf != nil {
				leaf.MarkA = mark
			}
		}
	}
	markLeaves(true)
	vec.Vec_PtrClear(visited)
	var collect func(*abc.Object)
	collect = func(node *abc.Object) {
		if node == nil || node.MarkA {
			return
		}
		if node.Kind == abc.ObjectNode {
			fanins := node.Fanins()
			if len(fanins) != 2 {
				panic("abc restructure: cone node is not a two-input AIG node")
			}
			collect(fanins[0].Object)
			collect(fanins[1].Object)
		}
		node.MarkA = true
		vec.Vec_PtrPush(visited, node)
	}
	collect(root)
	markLeaves(false)
	defer func() {
		for i := 0; i < vec.Vec_PtrSize(visited); i++ {
			node, _ := vec.Vec_PtrEntry(visited, i).(*abc.Object)
			if node != nil {
				node.MarkA = false
			}
		}
	}()
	if vec.Vec_PtrSize(visited) == 0 {
		return nil
	}
	for i := 0; i < vec.Vec_PtrSize(leaves); i++ {
		leaf, _ := vec.Vec_PtrEntry(leaves, i).(*abc.Object)
		if leaf == nil || i >= len(dd.Vars) {
			return nil
		}
		leaf.CopyData = dd.Vars[i]
	}
	var result *core.DdNode
	for i := 0; i < vec.Vec_PtrSize(visited); i++ {
		node, _ := vec.Vec_PtrEntry(visited, i).(*abc.Object)
		fanins := node.Fanins()
		left, ok0 := fanins[0].Object.CopyData.(*core.DdNode)
		right, ok1 := fanins[1].Object.CopyData.(*core.DdNode)
		if !ok0 || !ok1 || left == nil || right == nil {
			return nil
		}
		left = core.Cudd_NotCond(left, fanins[0].Complement)
		right = core.Cudd_NotCond(right, fanins[1].Complement)
		result = core.Cudd_bddAnd(dd, left, right)
		if result == nil {
			return nil
		}
		core.Cudd_Ref(result)
		node.CopyData = result
	}
	if result == nil {
		return nil
	}
	core.Cudd_Ref(result)
	for i := 0; i < vec.Vec_PtrSize(visited); i++ {
		node, _ := vec.Vec_PtrEntry(visited, i).(*abc.Object)
		if temp, ok := node.CopyData.(*core.DdNode); ok && temp != nil {
			core.Cudd_RecursiveDeref(dd, temp)
		}
	}
	core.Cudd_Deref(result)
	return result
}

func Abc_NtkStartCutManForRestruct(pNtk *Abc_Ntk_t, nCutMax, fDag int32) *Cut_Man_t {
	params := &cut.Cut_Params_t{
		NVarsMax: nCutMax, NKeepMax: 250, NIdsMax: abc.Abc_NtkObjNumMax(pNtk),
		FTruth: 0, FFilter: 1, FSeq: 0, FDrop: 0, FDag: fDag, FTree: 0, FVerbose: 0,
	}
	return cut.Cut_ManStartForNetwork(abcRestructCutNetwork{network: pNtk}, params)
}

type abcRestructCutNetwork struct{ network *abc.Network }

func (n abcRestructCutNetwork) CutNodes() []cut.NetworkNode {
	objects := n.network.Objects()
	nodes := make([]cut.NetworkNode, 0, len(objects))
	for _, object := range objects {
		nodes = append(nodes, abcCutNetworkNode(object))
	}
	return nodes
}

const restructInfeasibleNode = uint32(2006)

func restructNodeSignal(graph *dec.Dec_Graph_t, edge dec.Dec_Edge_t) (abc.Edge, bool) {
	value := dec.Dec_GraphNodeFunc(graph, int32(edge.Node))
	signal, ok := value.(abc.Edge)
	if !ok || signal.Object == nil {
		return abc.Edge{}, false
	}
	if dec.Dec_EdgeIsComplement(edge) {
		signal = signal.Not()
	}
	return signal, true
}

func restructSetNodeSignal(graph *dec.Dec_Graph_t, edge dec.Dec_Edge_t, signal abc.Edge) {
	if dec.Dec_EdgeIsComplement(edge) {
		signal = signal.Not()
	}
	dec.Dec_GraphSetNodeFunc(graph, int32(edge.Node), signal)
}

func restructPushOrdered(graph *dec.Dec_Graph_t, edges *vec.Vec_Int_t, value uint32) {
	newEdge := dec.Dec_IntToEdge(value)
	level := dec.Dec_GraphNode(graph, int32(newEdge.Node)).Level
	vec.Vec_IntPush(edges, 0)
	for i := vec.Vec_IntSize(edges) - 2; i >= 0; i-- {
		old := dec.Dec_IntToEdge(uint32(vec.Vec_IntEntry(edges, i)))
		if dec.Dec_GraphNode(graph, int32(old.Node)).Level <= level {
			vec.Vec_IntWriteEntry(edges, i+1, int32(dec.Dec_EdgeToInt(old)))
		} else {
			vec.Vec_IntWriteEntry(edges, i+1, int32(value))
			return
		}
	}
	vec.Vec_IntWriteEntry(edges, 0, int32(value))
}

func restructPermute(graph *dec.Dec_Graph_t, manager *abc.Abc_Aig_t, edges *vec.Vec_Int_t, xor bool) {
	left, right := 0, vec.Vec_IntSize(edges)-2
	if left == right {
		return
	}
	e1 := dec.Dec_IntToEdge(uint32(vec.Vec_IntEntry(edges, right+1)))
	e2 := dec.Dec_IntToEdge(uint32(vec.Vec_IntEntry(edges, right)))
	p1, ok1 := restructNodeSignal(graph, e1)
	p2, _ := restructNodeSignal(graph, e2)
	if !ok1 {
		return
	}
	for i := right; i >= left; i-- {
		e3 := dec.Dec_IntToEdge(uint32(vec.Vec_IntEntry(edges, i)))
		p3, ok3 := restructNodeSignal(graph, e3)
		if !ok3 {
			continue
		}
		var candidate abc.Edge
		if xor {
			candidate = abc.Abc_AigXorLookup(manager, p1, p3, nil)
		} else {
			candidate = abc.Abc_AigAndLookup(manager, p1.Not(), p3.Not())
			if candidate.Object != nil {
				candidate = candidate.Not()
			}
		}
		if candidate.Object == nil || abc.Abc_NodeIsTravIdCurrent(candidate.Object) != 0 {
			continue
		}
		if xor && p3 == p2 || !xor && e3.Node == e2.Node {
			return
		}
		vec.Vec_IntWriteEntry(edges, i, int32(dec.Dec_EdgeToInt(e2)))
		vec.Vec_IntWriteEntry(edges, right, int32(dec.Dec_EdgeToInt(e3)))
		return
	}
}

func restructEvalDsdRec(graph *dec.Dec_Graph_t, m *Abc_ManRst_t, node *dsd.Dsd_Node_t, required, nodesSaved int32, nodesAdded *int32) dec.Dec_Edge_t {
	quit := dec.Dec_EdgeCreate(int32(restructInfeasibleNode), 0)
	compl := dsd.Dsd_IsComplement(node)
	node = dsd.Dsd_Regular(node)
	typ := dsd.Dsd_NodeReadType(node)
	if typ == dsd.DSD_NODE_BUF {
		index := int32(dsd.Dsd_NodeReadFunc(node).Index)
		if index >= dec.Dec_GraphLeaveNum(graph) {
			return quit
		}
		return dec.Dec_EdgeCreate(index, boolToDec(compl))
	}
	if typ != dsd.DSD_NODE_OR && typ != dsd.DSD_NODE_EXOR && typ != dsd.DSD_NODE_PRIME {
		return quit
	}
	edges := vec.Vec_IntAlloc(dsd.Dsd_NodeReadDecsNum(node))
	defer vec.Vec_IntFree(edges)
	for _, child := range dsd.Dsd_NodeReadDecs(node) {
		result := restructEvalDsdRec(graph, m, child, required, nodesSaved, nodesAdded)
		if result.Node == restructInfeasibleNode {
			return quit
		}
		encoded := int32(dec.Dec_EdgeToInt(result))
		if typ == dsd.DSD_NODE_PRIME {
			vec.Vec_IntPush(edges, encoded)
		} else {
			restructPushOrdered(graph, edges, uint32(encoded))
		}
	}
	if typ == dsd.DSD_NODE_OR || typ == dsd.DSD_NODE_EXOR {
		if vec.Vec_IntSize(edges) < 2 {
			return quit
		}
		for vec.Vec_IntSize(edges) > 1 {
			if vec.Vec_IntSize(edges) > 2 {
				restructPermute(graph, m.aig(), edges, typ == dsd.DSD_NODE_EXOR)
			}
			e1 := dec.Dec_IntToEdge(uint32(vec.Vec_IntPop(edges)))
			e2 := dec.Dec_IntToEdge(uint32(vec.Vec_IntPop(edges)))
			p1, ok1 := restructNodeSignal(graph, e1)
			p2, ok2 := restructNodeSignal(graph, e2)
			kind := int32(0)
			var found abc.Edge
			var result dec.Dec_Edge_t
			if typ == dsd.DSD_NODE_OR {
				if ok1 && ok2 {
					found = abc.Abc_AigAndLookup(m.aig(), p1.Not(), p2.Not())
					if found.Object != nil {
						found = found.Not()
					}
				}
				result = dec.Dec_GraphAddNodeOr(graph, e1, e2)
			} else {
				if ok1 && ok2 {
					found = abc.Abc_AigXorLookup(m.aig(), p1, p2, &kind)
				}
				result = dec.Dec_GraphAddNodeXor(graph, e1, e2, kind)
			}
			level1 := dec.Dec_GraphNode(graph, int32(e1.Node)).Level
			level2 := dec.Dec_GraphNode(graph, int32(e2.Node)).Level
			level := max(level1, level2)
			if typ == dsd.DSD_NODE_EXOR {
				level += 2
			} else {
				level++
			}
			dec.Dec_GraphNode(graph, int32(result.Node)).Level = level
			if found.Object != nil {
				restructSetNodeSignal(graph, result, abc.Abc_ObjNotCond(found, boolToDec(dec.Dec_EdgeIsComplement(result))))
				if required != 1<<31-1 && level != uint32(found.Object.Level) {
					panic("abc restructure: operator lookup level differs from required level")
				}
			}
			if found.Object == nil || abc.Abc_NodeIsTravIdCurrent(found.Object) != 0 {
				*nodesAdded++
				if typ == dsd.DSD_NODE_EXOR {
					if !ok1 || !ok2 {
						*nodesAdded += 2
					} else if kind == 0 {
						for _, pair := range [][2]abc.Edge{{p1, p2.Not()}, {p1.Not(), p2}} {
							temp := abc.Abc_AigAndLookup(m.aig(), pair[0], pair[1])
							if temp.Object == nil || abc.Abc_NodeIsTravIdCurrent(temp.Object) != 0 {
								*nodesAdded++
							}
						}
					} else {
						for _, pair := range [][2]abc.Edge{{p1.Not(), p2.Not()}, {p1, p2}} {
							temp := abc.Abc_AigAndLookup(m.aig(), pair[0], pair[1])
							if temp.Object == nil || abc.Abc_NodeIsTravIdCurrent(temp.Object) != 0 {
								*nodesAdded++
							}
						}
					}
				}
				if *nodesAdded > nodesSaved {
					return quit
				}
			}
			vec.Vec_IntPush(edges, int32(dec.Dec_EdgeToInt(result)))
		}
		result := dec.Dec_IntToEdge(uint32(vec.Vec_IntPop(edges)))
		if compl {
			result = dec.Dec_EdgeNot(result)
		}
		return result
	}
	return restructEvalPrime(graph, m, node, edges, compl, required, nodesSaved, nodesAdded, quit)
}

func restructPrimeVariable(dd *core.DdManager, f *core.DdNode) bool {
	if f == nil || core.Cudd_IsConstant(f) {
		return false
	}
	regular := core.Cudd_Regular(f)
	if int(regular.Index) >= len(dd.Vars) {
		return false
	}
	return core.Cudd_NotCond(dd.Vars[regular.Index], core.Cudd_IsComplement(f)) == f
}

func restructEvalPrime(graph *dec.Dec_Graph_t, m *Abc_ManRst_t, node *dsd.Dsd_Node_t, edges *vec.Vec_Int_t, compl bool, required, nodesSaved int32, nodesAdded *int32, quit dec.Dec_Edge_t) dec.Dec_Edge_t {
	local := dsd.Dsd_TreeGetPrimeFunction(m.dd, node)
	if local == nil {
		return quit
	}
	core.Cudd_Ref(local)
	defer core.Cudd_RecursiveDeref(m.dd, local)
	var control, cofT, cofE *core.DdNode
	for variable := 0; variable < 3 && variable < len(m.dd.Vars); variable++ {
		control = m.dd.Vars[variable]
		cofE = core.Cudd_Cofactor(m.dd, local, core.Cudd_Not(control))
		cofT = core.Cudd_Cofactor(m.dd, local, control)
		if cofE == nil || cofT == nil {
			return quit
		}
		core.Cudd_Ref(cofE)
		core.Cudd_Ref(cofT)
		if restructPrimeVariable(m.dd, cofE) && restructPrimeVariable(m.dd, cofT) {
			break
		}
		core.Cudd_RecursiveDeref(m.dd, cofE)
		core.Cudd_RecursiveDeref(m.dd, cofT)
		cofE, cofT = nil, nil
	}
	if cofE == nil || cofT == nil || !restructPrimeVariable(m.dd, cofE) || !restructPrimeVariable(m.dd, cofT) || control == nil {
		if cofE != nil {
			core.Cudd_RecursiveDeref(m.dd, cofE)
		}
		if cofT != nil {
			core.Cudd_RecursiveDeref(m.dd, cofT)
		}
		return quit
	}
	defer core.Cudd_RecursiveDeref(m.dd, cofE)
	defer core.Cudd_RecursiveDeref(m.dd, cofT)
	selectEdge := func(index int32, complement bool) (dec.Dec_Edge_t, bool) {
		if index < 0 || int(index) >= vec.Vec_IntSize(edges) {
			return dec.Dec_Edge_t{}, false
		}
		e := dec.Dec_IntToEdge(uint32(vec.Vec_IntEntry(edges, int(index))))
		if complement {
			e = dec.Dec_EdgeNot(e)
		}
		return e, true
	}
	e1, ok1 := selectEdge(int32(control.Index), false)
	e2, ok2 := selectEdge(int32(core.Cudd_Regular(cofT).Index), core.Cudd_IsComplement(cofT))
	e3, ok3 := selectEdge(int32(core.Cudd_Regular(cofE).Index), core.Cudd_IsComplement(cofE))
	if !ok1 || !ok2 || !ok3 {
		return quit
	}
	p1, has1 := restructNodeSignal(graph, e1)
	p2, has2 := restructNodeSignal(graph, e2)
	p3, has3 := restructNodeSignal(graph, e3)
	kind := int32(0)
	var found abc.Edge
	if has1 && has2 && has3 {
		found = abc.Abc_AigMuxLookup(m.aig(), p1, p2, p3, &kind)
	}
	result := dec.Dec_GraphAddNodeMux(graph, e1, e2, e3, kind)
	level := max(dec.Dec_GraphNode(graph, int32(e1.Node)).Level, dec.Dec_GraphNode(graph, int32(e2.Node)).Level)
	level = max(level, dec.Dec_GraphNode(graph, int32(e3.Node)).Level) + 2
	dec.Dec_GraphNode(graph, int32(result.Node)).Level = level
	if found.Object != nil {
		restructSetNodeSignal(graph, result, abc.Abc_ObjNotCond(found, boolToDec(dec.Dec_EdgeIsComplement(result))))
		if required != 1<<31-1 && level != uint32(found.Object.Level) {
			panic("abc restructure: mux lookup level differs from required level")
		}
	}
	if found.Object == nil || abc.Abc_NodeIsTravIdCurrent(found.Object) != 0 {
		*nodesAdded++
		if kind == 0 {
			for _, pair := range [][2]abc.Edge{{p1, p2}, {p1.Not(), p3}} {
				if pair[0].Object == nil || pair[1].Object == nil {
					*nodesAdded++
					continue
				}
				temp := abc.Abc_AigAndLookup(m.aig(), pair[0], pair[1])
				if temp.Object == nil || abc.Abc_NodeIsTravIdCurrent(temp.Object) != 0 {
					*nodesAdded++
				}
			}
		} else {
			for _, pair := range [][2]abc.Edge{{p1, p2.Not()}, {p1.Not(), p3.Not()}} {
				if pair[0].Object == nil || pair[1].Object == nil {
					*nodesAdded++
					continue
				}
				temp := abc.Abc_AigAndLookup(m.aig(), pair[0], pair[1])
				if temp.Object == nil || abc.Abc_NodeIsTravIdCurrent(temp.Object) != 0 {
					*nodesAdded++
				}
			}
		}
		if *nodesAdded > nodesSaved {
			return quit
		}
	}
	if compl {
		result = dec.Dec_EdgeNot(result)
	}
	return result
}

func Abc_NodeEvaluateDsd(m *Abc_ManRst_t, dsdNode *dsd.Dsd_Node_t, root *Abc_Obj_t, required, nodesSaved int32, nodesAdded *int32) *dec.Dec_Graph_t {
	graph := dec.Dec_GraphCreate(int32(vec.Vec_PtrSize(m.vLeaves)))
	for i := 0; i < vec.Vec_PtrSize(m.vLeaves); i++ {
		leaf, _ := vec.Vec_PtrEntry(m.vLeaves, i).(*abc.Object)
		dec.Dec_GraphSetLeafFunc(graph, int32(i), leaf.Signal())
		dec.Dec_GraphNode(graph, int32(i)).Level = uint32(leaf.Level)
	}
	*nodesAdded = 0
	edge := restructEvalDsdRec(graph, m, dsdNode, required, nodesSaved, nodesAdded)
	if edge.Node > 1000 {
		*nodesAdded = -1
		dec.Dec_GraphFree(graph)
		return nil
	}
	if signal, ok := restructNodeSignal(graph, edge); ok && signal.Object == root {
		*nodesAdded = -1
		dec.Dec_GraphFree(graph)
		return nil
	}
	dec.Dec_GraphSetRoot(graph, edge)
	return graph
}

func Abc_NodeRestructureCut(m *Abc_ManRst_t, root *Abc_Obj_t, c *Cut_Cut_t) *dec.Dec_Graph_t {
	m.nCutsConsidered++
	required := int32(1<<31 - 1)
	if m.fUpdateLevel != 0 {
		required = abc.Abc_ObjRequiredLevel(root)
	}
	vec.Vec_PtrClear(m.vLeaves)
	for _, id := range c.Leaves() {
		leaf := abc.Abc_NtkObj(root.Network(), id)
		if leaf == nil {
			return nil
		}
		vec.Vec_PtrPush(m.vLeaves, leaf)
	}
	started := time.Now()
	function := abcNodeConeBdd(m.dd, root, m.vLeaves, m.vVisited)
	m.timeBdd += time.Since(started)
	if function == nil {
		return nil
	}
	core.Cudd_Ref(function)
	if core.Cudd_IsConstant(function) {
		m.nLastGain = abc.Abc_NodeMffcSize(root)
		m.nNodesGained += m.nLastGain
		m.nNodesRestructured++
		zero := core.Cudd_IsComplement(function)
		core.Cudd_RecursiveDeref(m.dd, function)
		if zero {
			return dec.Dec_GraphCreateConst0()
		}
		return dec.Dec_GraphCreateConst1()
	}
	started = time.Now()
	dsdNode := dsd.Dsd_DecomposeOne(m.pManDsd, function)
	m.timeDsd += time.Since(started)
	maxBlock := 0
	dsd.Dsd_TreeNodeGetInfoOne(dsdNode, nil, &maxBlock)
	if maxBlock > 3 {
		core.Cudd_RecursiveDeref(m.dd, function)
		return nil
	}
	m.nCutsExplored++
	for i := 0; i < vec.Vec_PtrSize(m.vLeaves); i++ {
		leaf, _ := vec.Vec_PtrEntry(m.vLeaves, i).(*abc.Object)
		abc.Abc_ObjAdjustFanoutCount(leaf, 1)
	}
	abc.Abc_NtkIncrementTravId(root.Network())
	nodesSaved := abc.Abc_NodeMffcLabelAig(root)
	for i := 0; i < vec.Vec_PtrSize(m.vLeaves); i++ {
		leaf, _ := vec.Vec_PtrEntry(m.vLeaves, i).(*abc.Object)
		abc.Abc_ObjAdjustFanoutCount(leaf, -1)
	}
	abc.Abc_NtkFinishFanoutCountAdjustments(root.Network())
	started = time.Now()
	var nodesAdded int32
	graph := Abc_NodeEvaluateDsd(m, dsdNode, root, required, nodesSaved, &nodesAdded)
	m.timeEval += time.Since(started)
	if graph == nil || nodesAdded == -1 || nodesAdded == nodesSaved && m.fUseZeros == 0 {
		core.Cudd_RecursiveDeref(m.dd, function)
		if graph != nil {
			dec.Dec_GraphFree(graph)
		}
		return nil
	}
	m.nLastGain = nodesSaved - nodesAdded
	m.nNodesGained += m.nLastGain
	m.nNodesRestructured++
	core.Cudd_RecursiveDeref(m.dd, function)
	return graph
}

func Abc_NodeRestructure(m *Abc_ManRst_t, node *Abc_Obj_t, cuts *Cut_Cut_t) *dec.Dec_Graph_t {
	m.nNodesConsidered++
	for c := cuts; c != nil; c = c.Next {
		if cut.Cut_CutReadLeaveNum(c) < 4 {
			continue
		}
		if graph := Abc_NodeRestructureCut(m, node, c); graph != nil {
			return graph
		}
	}
	return nil
}

func Abc_NtkManRstPrintStats(m *Abc_ManRst_t) {
	fmt.Printf("Refactoring statistics:\n")
	fmt.Printf("Nodes considered   = %8d.\n", m.nNodesConsidered)
	fmt.Printf("Cuts considered    = %8d.\n", m.nCutsConsidered)
	fmt.Printf("Cuts explored      = %8d.\n", m.nCutsExplored)
	fmt.Printf("Nodes restructured = %8d.\n", m.nNodesRestructured)
	fmt.Printf("Calculated gain    = %8d.\n", m.nNodesGained)
	fmt.Printf("Cuts       %s\n", m.timeCut)
	fmt.Printf("Resynthesis%s\n", m.timeRes)
	fmt.Printf("    BDD    %s\n", m.timeBdd)
	fmt.Printf("    DSD    %s\n", m.timeDsd)
	fmt.Printf("    Eval   %s\n", m.timeEval)
	fmt.Printf("AIG update %s\n", m.timeNtk)
	fmt.Printf("TOTAL      %s\n", m.timeTotal)
}

// Abc_NtkRestructure performs one source-order pass over the AIG nodes and
// applies the first improving DSD reconstruction for each node.
func Abc_NtkRestructure(pNtk *Abc_Ntk_t, nCutMax, fUpdateLevel, fUseZeros, fVerbose int32) int32 {
	if abc.Abc_NtkIsStrash(pNtk) == 0 {
		panic("Abc_NtkRestructure requires a strashed AIG")
	}
	aig, _ := pNtk.FunctionManager.(*abc.Abc_Aig_t)
	if aig == nil {
		return 0
	}
	clkStart := time.Now()
	abc.Abc_AigCleanup(aig)
	abc.Abc_NtkCleanCopy(pNtk)
	if fUpdateLevel != 0 {
		abc.Abc_NtkStartReverseLevels(pNtk, 0)
	}
	m := Abc_NtkManRstStart(nCutMax, fUpdateLevel, fUseZeros, fVerbose)
	if m == nil {
		return 0
	}
	m.pNtk = pNtk
	cutsStart := time.Now()
	cutMan := Abc_NtkStartCutManForRestruct(pNtk, nCutMax, 1)
	m.timeCut += time.Since(cutsStart)
	nNodes := abc.Abc_NtkObjNumMax(pNtk)
	progress := extra.Extra_ProgressBarStart(os.Stdout, nNodes)
	for i := int32(0); i < nNodes; i++ {
		node := abc.Abc_NtkObj(pNtk, i)
		if node == nil || node.Kind != abc.ObjectNode {
			continue
		}
		extra.Extra_ProgressBarUpdate(progress, i, "")
		if node.Persist || abc.Abc_ObjFanoutNum(node) > 1000 {
			continue
		}
		started := time.Now()
		cutList := cut.Abc_NodeGetCutsRecursive(cutMan, abcCutNetworkNode(node), 1, 0)
		m.timeCut += time.Since(started)
		started = time.Now()
		graph := Abc_NodeRestructure(m, node, cutList)
		m.timeRes += time.Since(started)
		if graph == nil {
			continue
		}
		started = time.Now()
		Dec_GraphUpdateNetwork(node, graph, fUpdateLevel, m.nLastGain)
		m.timeNtk += time.Since(started)
		dec.Dec_GraphFree(graph)
	}
	extra.Extra_ProgressBarStop(progress)
	m.timeTotal = time.Since(clkStart)
	Abc_NtkManRstPrintStats(m)
	cut.Cut_ManStop(cutMan)
	Abc_NtkManRstStop(m)
	abc.Abc_NtkReassignIds(pNtk)
	if fUpdateLevel != 0 {
		abc.Abc_NtkStopReverseLevels(pNtk)
	} else {
		abc.Abc_NtkLevel(pNtk)
	}
	if abc.Abc_NtkCheck(pNtk) == 0 {
		fmt.Printf("Abc_NtkRefactor: The network check has failed.\n")
		return 0
	}
	return 1
}

func Abc_Abc_NodeResubCollectDivs(m *Abc_ManRst_t, root *Abc_Obj_t, c *Cut_Cut_t) int32 {
	vec.Vec_PtrClear(m.vDecs)
	abc.Abc_NtkIncrementTravId(root.Network())
	for _, id := range c.Leaves() {
		node := abc.Abc_NtkObj(root.Network(), id)
		if node == nil {
			return 0
		}
		vec.Vec_PtrPush(m.vDecs, node)
		abc.Abc_NodeSetTravIdCurrent(node)
	}
	for i := 0; i < vec.Vec_PtrSize(m.vDecs); i++ {
		node, _ := vec.Vec_PtrEntry(m.vDecs, i).(*abc.Object)
		for _, fanout := range node.Fanouts() {
			if abc.Abc_NodeIsTravIdCurrent(fanout) != 0 || abc.Abc_ObjIsPo(fanout) != 0 {
				continue
			}
			if abc.Abc_NodeIsTravIdCurrent(abc.Abc_ObjFanin0(fanout)) != 0 && abc.Abc_NodeIsTravIdCurrent(abc.Abc_ObjFanin1(fanout)) != 0 {
				vec.Vec_PtrPush(m.vDecs, fanout)
				abc.Abc_NodeSetTravIdCurrent(fanout)
			}
		}
	}
	return 1
}

func Abc_NodeResubMffc_rec(node *abc.Object) int32 {
	if abc.Abc_NodeIsTravIdCurrent(node) != 0 {
		return 0
	}
	abc.Abc_NodeSetTravIdCurrent(node)
	return 1 + Abc_NodeResubMffc_rec(abc.Abc_ObjFanin0(node)) + Abc_NodeResubMffc_rec(abc.Abc_ObjFanin1(node))
}

func Abc_NodeResubMffc(m *Abc_ManRst_t, divisors *vec.Vec_Ptr_t, leaves int32, root *abc.Object) int32 {
	abc.Abc_NtkIncrementTravId(root.Network())
	for i := int32(0); i < leaves; i++ {
		node, _ := vec.Vec_PtrEntry(divisors, int(i)).(*abc.Object)
		abc.Abc_NodeSetTravIdCurrent(node)
	}
	if abc.Abc_NodeIsTravIdPrevious(root) == 0 {
		panic("abc restructure: MFFC root is outside the collected cone")
	}
	counter := Abc_NodeResubMffc_rec(root)
	vec.Vec_PtrClear(m.vTemp)
	k := 0
	for i := int(leaves); i < vec.Vec_PtrSize(divisors); i++ {
		node, _ := vec.Vec_PtrEntry(divisors, i).(*abc.Object)
		if abc.Abc_NodeIsTravIdCurrent(node) != 0 {
			vec.Vec_PtrPush(m.vTemp, node)
		} else {
			vec.Vec_PtrWriteEntry(divisors, k, node)
			k++
		}
	}
	for i := 0; i < vec.Vec_PtrSize(m.vTemp); i++ {
		vec.Vec_PtrWriteEntry(divisors, k, vec.Vec_PtrEntry(m.vTemp, i))
		k++
	}
	if k != vec.Vec_PtrSize(divisors) {
		panic("abc restructure: divisor ordering lost nodes")
	}
	last, _ := vec.Vec_PtrEntryLast(divisors).(*abc.Object)
	if root != last {
		panic("abc restructure: MFFC root is not the final divisor")
	}
	return counter
}

func Abc_NodeMffcSimulate(divisors *vec.Vec_Ptr_t, leaves int32, randoms, sims *vec.Vec_Int_t) {
	vec.Vec_IntClear(sims)
	values := make(map[*abc.Object]uint32, vec.Vec_PtrSize(divisors))
	for i := int32(0); i < leaves; i++ {
		node, _ := vec.Vec_PtrEntry(divisors, int(i)).(*abc.Object)
		value := uint32(vec.Vec_IntEntry(randoms, int(i)))
		values[node] = value
		vec.Vec_IntPush(sims, int32(value))
	}
	for i := int(leaves); i < vec.Vec_PtrSize(divisors); i++ {
		node, _ := vec.Vec_PtrEntry(divisors, i).(*abc.Object)
		fanins := node.Fanins()
		if len(fanins) != 2 {
			panic("abc restructure: simulation requires two-input AIG nodes")
		}
		data0, ok0 := values[fanins[0].Object]
		data1, ok1 := values[fanins[1].Object]
		if !ok0 || !ok1 {
			panic("abc restructure: simulation fanin has no random word")
		}
		if fanins[0].Complement {
			data0 = ^data0
		}
		if fanins[1].Complement {
			data1 = ^data1
		}
		value := data0 & data1
		values[node] = value
		vec.Vec_IntPush(sims, int32(value))
	}
}

// Abc_NodeCheckFull matches the source routine, which currently accepts all
// candidates after its placeholder full-equivalence check.
func Abc_NodeCheckFull(_ *Abc_ManRst_t, _ *dec.Dec_Graph_t) int32 { return 1 }

func Abc_NodeMffcConstants(m *Abc_ManRst_t, sims *vec.Vec_Int_t) *dec.Dec_Graph_t {
	root := uint32(vec.Vec_IntEntryLast(sims))
	var graph *dec.Dec_Graph_t
	if root == 0 {
		graph = dec.Dec_GraphCreateConst0()
	} else if root == ^uint32(0) {
		graph = dec.Dec_GraphCreateConst1()
	}
	assert(graph != nil, "pGraph")
	if Abc_NodeCheckFull(m, graph) != 0 {
		return graph
	}
	dec.Dec_GraphFree(graph)
	return nil
}

func Abc_NodeMffcSingleVar(m *Abc_ManRst_t, sims *vec.Vec_Int_t, nodes int32, ones *vec.Vec_Int_t) *dec.Dec_Graph_t {
	vec.Vec_IntClear(ones)
	vec.Vec_IntClear(m.vBinate)
	root := uint32(vec.Vec_IntEntryLast(sims))
	for i := int32(0); i < nodes; i++ {
		value := uint32(vec.Vec_IntEntry(sims, int(i)))
		if root == value || root == ^value {
			graph := dec.Dec_GraphCreate(1)
			leaf, _ := vec.Vec_PtrEntry(m.vDecs, int(i)).(*abc.Object)
			dec.Dec_GraphSetLeafFunc(graph, 0, leaf.Signal())
			dec.Dec_GraphSetRoot(graph, dec.Dec_IntToEdge(uint32(boolToDec(root == ^value))))
			if Abc_NodeCheckFull(m, graph) != 0 {
				return graph
			}
			dec.Dec_GraphFree(graph)
		}
		if root&value == 0 {
			vec.Vec_IntPush(ones, int32(i<<1))
		} else if root & ^value == 0 {
			vec.Vec_IntPush(ones, int32((i<<1)+1))
		} else {
			vec.Vec_IntPush(m.vBinate, i)
		}
	}
	return nil
}

func Abc_NodeMffcSingleNode(m *Abc_ManRst_t, sims *vec.Vec_Int_t, _ int32, ones *vec.Vec_Int_t) *dec.Dec_Graph_t {
	root := uint32(vec.Vec_IntEntryLast(sims))
	for i := 0; i < vec.Vec_IntSize(ones); i++ {
		for k := i + 1; k < vec.Vec_IntSize(ones); k++ {
			a, b := uint32(vec.Vec_IntEntry(ones, i)), uint32(vec.Vec_IntEntry(ones, k))
			if ^root != a|b {
				continue
			}
			e0 := dec.Dec_IntToEdge(a ^ 1)
			e1 := dec.Dec_IntToEdge(b ^ 1)
			graph := dec.Dec_GraphCreate(2)
			leaf0, _ := vec.Vec_PtrEntry(m.vDecs, int(e0.Node)).(*abc.Object)
			leaf1, _ := vec.Vec_PtrEntry(m.vDecs, int(e1.Node)).(*abc.Object)
			dec.Dec_GraphSetLeafFunc(graph, int32(e0.Node), leaf0.Signal())
			dec.Dec_GraphSetLeafFunc(graph, int32(e1.Node), leaf1.Signal())
			dec.Dec_GraphSetRoot(graph, dec.Dec_GraphAddNodeAnd(graph, e0, e1))
			if Abc_NodeCheckFull(m, graph) != 0 {
				return graph
			}
			dec.Dec_GraphFree(graph)
		}
	}
	return nil
}

func Abc_NodeMffcDoubleNode(_ *Abc_ManRst_t, _ *vec.Vec_Int_t, _ int32, _ *vec.Vec_Int_t) *dec.Dec_Graph_t {
	return nil
}

func Abc_NodeResubEval(m *Abc_ManRst_t, root *Abc_Obj_t, c *Cut_Cut_t) *dec.Dec_Graph_t {
	if Abc_Abc_NodeResubCollectDivs(m, root, c) == 0 {
		return nil
	}
	nodesSaved := Abc_NodeResubMffc(m, m.vDecs, int32(len(c.Leaves())), root)
	assert(nodesSaved > 0, "nNodesSaved > 0")
	Abc_NodeMffcSimulate(m.vDecs, int32(len(c.Leaves())), m.vRands, m.vSims)
	if graph := Abc_NodeMffcConstants(m, m.vSims); graph != nil {
		m.nNodesGained += nodesSaved
		m.nNodesRestructured++
		return graph
	}
	nodes := int32(vec.Vec_IntSize(m.vSims)) - nodesSaved
	if graph := Abc_NodeMffcSingleVar(m, m.vSims, nodes, m.vOnes); graph != nil {
		m.nNodesGained += nodesSaved
		m.nNodesRestructured++
		return graph
	}
	if nodesSaved == 1 {
		return nil
	}
	if graph := Abc_NodeMffcSingleNode(m, m.vSims, nodes, m.vOnes); graph != nil {
		m.nNodesGained += nodesSaved - 1
		m.nNodesRestructured++
		return graph
	}
	if nodesSaved == 2 {
		return nil
	}
	if graph := Abc_NodeMffcDoubleNode(m, m.vSims, nodes, m.vOnes); graph != nil {
		m.nNodesGained += nodesSaved - 2
		m.nNodesRestructured++
		return graph
	}
	return nil
}

func Abc_NodeResubstitute(m *Abc_ManRst_t, node *Abc_Obj_t, cuts *Cut_Cut_t) *dec.Dec_Graph_t {
	m.nNodesConsidered++
	nCuts := int32(0)
	for c := cuts; c != nil; c = c.Next {
		if cut.Cut_CutReadLeaveNum(c) > 3 {
			nCuts++
		}
	}
	fmt.Printf("-----------------------------------\nNode %6d : Factor-cuts = %5d.\n", node.ID, nCuts)
	var best *dec.Dec_Graph_t
	for c := cuts; c != nil; c = c.Next {
		if cut.Cut_CutReadLeaveNum(c) < 4 {
			continue
		}
		graph := Abc_NodeResubEval(m, node, c)
		if graph == nil {
			continue
		}
		if best == nil || dec.Dec_GraphNodeNum(graph) < dec.Dec_GraphNodeNum(best) {
			dec.Dec_GraphFree(best)
			best = graph
		} else {
			dec.Dec_GraphFree(graph)
		}
	}
	return best
}

func boolToDec(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

func (m *Abc_ManRst_t) aig() *abc.Abc_Aig_t {
	aig, _ := m.pNtk.FunctionManager.(*abc.Abc_Aig_t)
	return aig
}
