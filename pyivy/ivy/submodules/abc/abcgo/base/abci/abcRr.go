package abci

import (
	"fmt"
	"os"
	"time"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/sim"
	"github.com/glycerine/abcgo/proof/fraig"
)

// Abc_RRMan_t owns the temporary cones, window, and proof miter used by abcRr.c.
type Abc_RRMan_t struct {
	pNtk                                                    *Abc_Ntk_t
	nFaninLevels, nFanoutLevels                             int32
	pNode, pFanin, pFanout                                  *Abc_Obj_t
	vFaninLeaves, vFanoutRoots                              *Vec_Ptr_t
	vLeaves, vCone, vRoots                                  *Vec_Ptr_t
	pWnd, pMiter                                            *Abc_Ntk_t
	pParams                                                 *fraig.Prove_Params_t
	nNodesOld, nLevelsOld, nEdgesTried, nEdgesRemoved       int32
	timeWindow, timeMiter, timeProve, timeUpdate, timeTotal int64
}

func Abc_NtkRR(pNtk *Abc_Ntk_t, nFaninLevels, nFanoutLevels, fUseFanouts, fVerbose int32) int32 {
	clkTotal := util.Abc_Clock()
	p := Abc_RRManStart()
	p.pNtk, p.nFaninLevels, p.nFanoutLevels = pNtk, nFaninLevels, nFanoutLevels
	p.nNodesOld, p.nLevelsOld = baseabc.Abc_NtkNodeNum(pNtk), baseabc.Abc_AigLevel(pNtk)
	baseabc.Abc_NtkCleanCopy(pNtk)
	nNodes := baseabc.Abc_NtkObjNumMax(pNtk)
	Abc_NtkRRSimulateStart(pNtk)
	progress := extra.Extra_ProgressBarStart(os.Stdout, nNodes)
	for i := int32(0); i < nNodes; i++ {
		pNode := baseabc.Abc_NtkObj(pNtk, i)
		if pNode == nil || pNode.Kind != baseabc.ObjectNode {
			continue
		}
		extra.Extra_ProgressBarUpdate(progress, i, "")
		if i >= nNodes {
			break
		}
		if pNode.Persist || baseabc.Abc_ObjFanoutNum(pNode) > 1000 {
			continue
		}
		if fUseFanouts == 0 {
			for k := int32(0); k < baseabc.Abc_ObjFaninNum(pNode); k++ {
				pFanin := baseabc.Abc_ObjFanin(pNode, k)
				if baseabc.Abc_ObjFanoutNum(pFanin) == 1 {
					continue
				}
				p.nEdgesTried++
				Abc_RRManClean(p)
				p.pNode, p.pFanin, p.pFanout = pNode, pFanin, nil
				clk := util.Abc_Clock()
				ret := Abc_NtkRRWindow(p)
				p.timeWindow += util.Abc_Clock() - clk
				if ret == 0 {
					continue
				}
				clk = util.Abc_Clock()
				ret = Abc_NtkRRProve(p)
				p.timeMiter += util.Abc_Clock() - clk
				if ret == 0 {
					continue
				}
				clk = util.Abc_Clock()
				Abc_NtkRRUpdate(pNtk, p.pNode, p.pFanin, p.pFanout)
				p.timeUpdate += util.Abc_Clock() - clk
				p.nEdgesRemoved++
				break
			}
			continue
		}
		for k := int32(0); k < baseabc.Abc_ObjFaninNum(pNode); k++ {
			pFanin := baseabc.Abc_ObjFanin(pNode, k)
			for m := int32(0); m < baseabc.Abc_ObjFanoutNum(pNode); m++ {
				pFanout := baseabc.Abc_ObjFanout(pNode, m)
				p.nEdgesTried++
				Abc_RRManClean(p)
				p.pNode, p.pFanin, p.pFanout = pNode, pFanin, pFanout
				clk := util.Abc_Clock()
				ret := Abc_NtkRRWindow(p)
				p.timeWindow += util.Abc_Clock() - clk
				if ret == 0 {
					continue
				}
				clk = util.Abc_Clock()
				ret = Abc_NtkRRProve(p)
				p.timeMiter += util.Abc_Clock() - clk
				if ret == 0 {
					continue
				}
				clk = util.Abc_Clock()
				Abc_NtkRRUpdate(pNtk, p.pNode, p.pFanin, p.pFanout)
				p.timeUpdate += util.Abc_Clock() - clk
				p.nEdgesRemoved++
				break
			}
		}
	}
	Abc_NtkRRSimulateStop(pNtk)
	extra.Extra_ProgressBarStop(progress)
	p.timeTotal = util.Abc_Clock() - clkTotal
	if fVerbose != 0 {
		Abc_RRManPrintStats(p)
	}
	Abc_RRManStop(p)
	baseabc.Abc_NtkReassignIds(pNtk)
	baseabc.Abc_NtkLevel(pNtk)
	if baseabc.Abc_NtkCheck(pNtk) == 0 {
		fmt.Println("Abc_NtkRR: The network check has failed.")
		return 0
	}
	return 1
}

func Abc_RRManStart() *Abc_RRMan_t {
	p := &Abc_RRMan_t{
		vFaninLeaves: vec.Vec_PtrAlloc(100), vFanoutRoots: vec.Vec_PtrAlloc(100),
		vLeaves: vec.Vec_PtrAlloc(100), vCone: vec.Vec_PtrAlloc(100), vRoots: vec.Vec_PtrAlloc(100),
		pParams: &fraig.Prove_Params_t{},
	}
	fraig.Prove_ParamsSetDefault(p.pParams)
	return p
}
func Abc_RRManStop(p *Abc_RRMan_t) {
	Abc_RRManClean(p)
	vec.Vec_PtrFree(p.vFaninLeaves)
	vec.Vec_PtrFree(p.vFanoutRoots)
	vec.Vec_PtrFree(p.vLeaves)
	vec.Vec_PtrFree(p.vCone)
	vec.Vec_PtrFree(p.vRoots)
	*p = Abc_RRMan_t{}
}
func Abc_RRManPrintStats(p *Abc_RRMan_t) {
	nodesNow := baseabc.Abc_NtkNodeNum(p.pNtk)
	levelsNow := baseabc.Abc_AigLevel(p.pNtk)
	ratio := 100.0 * float64(p.nNodesOld-nodesNow) / float64(p.nNodesOld)
	fmt.Println("Redundancy removal statistics:")
	fmt.Printf("Edges tried     = %6d.\n", p.nEdgesTried)
	fmt.Printf("Edges removed   = %6d. (%5.2f %%)\n", p.nEdgesRemoved, 100.0*float64(p.nEdgesRemoved)/float64(p.nEdgesTried))
	fmt.Printf("Node gain       = %6d. (%5.2f %%)\n", p.nNodesOld-nodesNow, ratio)
	fmt.Printf("Level gain      = %6d.\n", p.nLevelsOld-levelsNow)
	for _, item := range []struct {
		label    string
		duration int64
	}{
		{"Windowing      ", p.timeWindow}, {"Miter          ", p.timeMiter},
		{"    Construct  ", p.timeMiter - p.timeProve}, {"    Prove      ", p.timeProve},
		{"Update         ", p.timeUpdate}, {"TOTAL          ", p.timeTotal},
	} {
		fmt.Printf("%s =%9.2f sec\n", item.label, time.Duration(item.duration).Seconds())
	}
}
func Abc_RRManClean(p *Abc_RRMan_t) {
	p.pNode, p.pFanin, p.pFanout = nil, nil, nil
	vec.Vec_PtrClear(p.vFaninLeaves)
	vec.Vec_PtrClear(p.vFanoutRoots)
	vec.Vec_PtrClear(p.vLeaves)
	vec.Vec_PtrClear(p.vCone)
	vec.Vec_PtrClear(p.vRoots)
	if p.pWnd != nil {
		baseabc.Abc_NtkDelete(p.pWnd)
	}
	if p.pMiter != nil {
		baseabc.Abc_NtkDelete(p.pMiter)
	}
	p.pWnd, p.pMiter = nil, nil
}
func Abc_NtkRRProve(p *Abc_RRMan_t) int32 {
	windowCopy := baseabc.Abc_NtkDup(p.pWnd)
	mapCopy := func(original *Abc_Obj_t) *Abc_Obj_t {
		windowObject := baseabc.Abc_ObjCopyEdge(original).Object
		return baseabc.Abc_ObjCopyEdge(windowObject).Object
	}
	var fanout *Abc_Obj_t
	if p.pFanout != nil {
		fanout = mapCopy(p.pFanout)
	}
	Abc_NtkRRUpdate(windowCopy, mapCopy(p.pNode), mapCopy(p.pFanin), fanout)
	if baseabc.Abc_NtkIsDfsOrdered(windowCopy) == 0 {
		baseabc.Abc_NtkReassignIds(windowCopy)
	}
	p.pMiter = Abc_NtkMiter(p.pWnd, windowCopy, 1, 0, 0, 0)
	baseabc.Abc_NtkDelete(windowCopy)
	clk := util.Abc_Clock()
	ret := Abc_NtkMiterProve(&p.pMiter, p.pParams)
	p.timeProve += util.Abc_Clock() - clk
	if ret == 1 {
		return 1
	}
	return 0
}
func Abc_NtkRRUpdate(ntk *Abc_Ntk_t, node, fanin, fanout *Abc_Obj_t) int32 {
	assert(fanout == nil, "pFanout == NULL")
	assert(baseabc.Abc_ObjIsComplement(node) == 0, "!Abc_ObjIsComplement(pNode)")
	assert(baseabc.Abc_ObjIsComplement(fanin) == 0, "!Abc_ObjIsComplement(pFanin)")
	if fanout != nil {
		assert(baseabc.Abc_ObjIsComplement(fanout) == 0, "!Abc_ObjIsComplement(pFanout)")
	}
	var nodeNew any
	if fanin == baseabc.Abc_ObjFanin0(node) {
		nodeNew = baseabc.Abc_ObjChild1(node)
	} else if fanin == baseabc.Abc_ObjFanin1(node) {
		nodeNew = baseabc.Abc_ObjChild0(node)
	} else {
		assert(false, "fanin belongs to node")
	}
	manager := ntk.FunctionManager.(*baseabc.Abc_Aig_t)
	if fanout == nil {
		baseabc.Abc_AigReplace(manager, node, nodeNew, 1)
		return 1
	}
	var fanoutNew any
	if node == baseabc.Abc_ObjFanin0(fanout) {
		fanoutNew = baseabc.Abc_AigAnd(manager, baseabc.Abc_ObjNotCond(nodeNew, baseabc.Abc_ObjFaninC0(fanout)), baseabc.Abc_ObjChild1(fanout))
	} else if node == baseabc.Abc_ObjFanin1(fanout) {
		fanoutNew = baseabc.Abc_AigAnd(manager, baseabc.Abc_ObjNotCond(nodeNew, baseabc.Abc_ObjFaninC1(fanout)), baseabc.Abc_ObjChild0(fanout))
	} else {
		assert(false, "node is a fanin of fanout")
	}
	baseabc.Abc_AigReplace(manager, fanout, fanoutNew, 1)
	return 1
}

func Abc_NtkRRWindow(p *Abc_RRMan_t) int32 {
	pEdgeFanout, pEdgeFanin := p.pNode, p.pFanin
	if p.pFanout != nil {
		pEdgeFanout, pEdgeFanin = p.pFanout, p.pNode
	}
	levelMin := max(0, int32(p.pFanin.Level)-p.nFaninLevels)
	levelMax := int32(pEdgeFanout.Level) + p.nFanoutLevels
	baseabc.Abc_NtkIncrementTravId(p.pNtk)
	baseabc.Abc_NodeSetTravIdCurrent(p.pFanin)
	vec.Vec_PtrPush(p.vFaninLeaves, p.pFanin)
	for Abc_NtkRRTfi_int(p.vFaninLeaves, levelMin) != 0 {
	}
	baseabc.Abc_NtkIncrementTravId(p.pNtk)
	for _, value := range vec.Vec_PtrArray(p.vFaninLeaves) {
		baseabc.Abc_NodeSetTravIdCurrent(value.(*Abc_Obj_t))
	}
	for Abc_NtkRRTfo_int(p.vFaninLeaves, p.vFanoutRoots, levelMax, pEdgeFanin, pEdgeFanout) != 0 {
	}
	for _, value := range vec.Vec_PtrArray(p.vFanoutRoots) {
		value.(*Abc_Obj_t).MarkA = true
	}
	ret := Abc_NtkRRTfo_rec(pEdgeFanout, p.vRoots, levelMax+1)
	for _, value := range vec.Vec_PtrArray(p.vFanoutRoots) {
		value.(*Abc_Obj_t).MarkA = false
	}
	if ret == 0 {
		return 0
	}
	baseabc.Abc_NtkIncrementTravId(p.pNtk)
	for _, value := range vec.Vec_PtrArray(p.vRoots) {
		Abc_NtkRRTfi_rec(value.(*Abc_Obj_t), p.vLeaves, p.vCone, levelMin)
	}
	p.pWnd = Abc_NtkWindow(p.pNtk, p.vLeaves, p.vCone, p.vRoots)
	if p.pWnd == nil {
		return 0
	}
	return 1
}
func Abc_NtkRRTfi_int(leaves *Vec_Ptr_t, levelLimit int32) int32 {
	assert(levelLimit >= 0, "LevelLimit >= 0")
	levelMax := int32(0)
	for _, value := range vec.Vec_PtrArray(leaves) {
		if level := int32(value.(*Abc_Obj_t).Level); levelMax < level {
			levelMax = level
		}
	}
	if levelMax <= levelLimit {
		return 0
	}
	size := vec.Vec_PtrSize(leaves)
	for i := 0; i < size; i++ {
		obj := vec.Vec_PtrEntry(leaves, i).(*Abc_Obj_t)
		if int32(obj.Level) != levelMax {
			continue
		}
		for _, edge := range obj.Fanins() {
			if baseabc.Abc_NodeIsTravIdCurrent(edge.Object) == 0 {
				baseabc.Abc_NodeSetTravIdCurrent(edge.Object)
				vec.Vec_PtrPush(leaves, edge.Object)
			}
		}
	}
	kept := 0
	for _, value := range vec.Vec_PtrArray(leaves) {
		obj := value.(*Abc_Obj_t)
		if int32(obj.Level) != levelMax {
			vec.Vec_PtrWriteEntry(leaves, kept, obj)
			kept++
		}
	}
	vec.Vec_PtrShrink(leaves, kept)
	if vec.Vec_PtrSize(leaves) > 2000 {
		return 0
	}
	return 1
}
func Abc_NtkRRTfo_int(leaves, roots *Vec_Ptr_t, levelLimit int32, edgeFanin, edgeFanout *Abc_Obj_t) int32 {
	levelMin := int32(abcInfinity)
	for _, value := range vec.Vec_PtrArray(leaves) {
		if level := int32(value.(*Abc_Obj_t).Level); levelMin > level {
			levelMin = level
		}
	}
	if levelMin > levelLimit {
		return 0
	}
	size := vec.Vec_PtrSize(leaves)
	for i := 0; i < size; i++ {
		obj := vec.Vec_PtrEntry(leaves, i).(*Abc_Obj_t)
		if int32(obj.Level) != levelMin {
			continue
		}
		isRoot := false
		for _, next := range obj.Fanouts() {
			if next.IsCO() || int32(next.Level) > levelLimit {
				isRoot = true
				continue
			}
			if obj == edgeFanin && next == edgeFanout {
				continue
			}
			if baseabc.Abc_NodeIsTravIdCurrent(next) != 0 {
				continue
			}
			baseabc.Abc_NodeSetTravIdCurrent(next)
			vec.Vec_PtrPush(leaves, next)
		}
		if isRoot {
			vec.Vec_PtrPush(roots, obj)
		}
	}
	kept := 0
	for _, value := range vec.Vec_PtrArray(leaves) {
		obj := value.(*Abc_Obj_t)
		if int32(obj.Level) != levelMin {
			vec.Vec_PtrWriteEntry(leaves, kept, obj)
			kept++
		}
	}
	vec.Vec_PtrShrink(leaves, kept)
	if vec.Vec_PtrSize(leaves) > 2000 {
		return 0
	}
	return 1
}
func Abc_NtkRRTfo_rec(node *Abc_Obj_t, roots *Vec_Ptr_t, levelLimit int32) int32 {
	if node.IsCO() || int32(node.Level) > levelLimit {
		return 0
	}
	if node.MarkA {
		vec.Vec_PtrPushUnique(roots, node)
		return 1
	}
	baseabc.Abc_NodeSetTravIdCurrent(node)
	for _, fanout := range node.Fanouts() {
		if Abc_NtkRRTfo_rec(fanout, roots, levelLimit) == 0 {
			return 0
		}
	}
	return 1
}
func Abc_NtkRRTfi_rec(node *Abc_Obj_t, leaves, cone *Vec_Ptr_t, levelLimit int32) {
	if baseabc.Abc_NodeIsTravIdCurrent(node) != 0 {
		return
	}
	if baseabc.Abc_NodeIsTravIdPrevious(node) == 0 || int32(node.Level) <= levelLimit {
		baseabc.Abc_NodeSetTravIdCurrent(node)
		vec.Vec_PtrPush(leaves, node)
		return
	}
	baseabc.Abc_NodeSetTravIdCurrent(node)
	for _, edge := range node.Fanins() {
		Abc_NtkRRTfi_rec(edge.Object, leaves, cone, levelLimit)
	}
	vec.Vec_PtrPush(cone, node)
}
func Abc_NtkWindow(ntk *Abc_Ntk_t, leaves, cone, roots *Vec_Ptr_t) *Abc_Ntk_t {
	assert(baseabc.Abc_NtkIsStrash(ntk) != 0, "Abc_NtkIsStrash(pNtk)")
	window := baseabc.Abc_NtkAlloc(baseabc.Abc_NtkType_t(ntk.Type), baseabc.Abc_NtkFunc_t(ntk.Function), 1)
	window.Name = "temp"
	baseabc.Abc_AigConst1(ntk).Object.SetCopyEdge(baseabc.Abc_AigConst1(window))
	for _, value := range vec.Vec_PtrArray(leaves) {
		obj := value.(*Abc_Obj_t)
		obj.SetCopyEdge(baseabc.Abc_NtkCreatePi(window).Signal())
	}
	for _, value := range vec.Vec_PtrArray(cone) {
		obj := value.(*Abc_Obj_t)
		a := baseabc.Abc_ObjCopyEdge(baseabc.Abc_ObjFanin0(obj))
		b := baseabc.Abc_ObjCopyEdge(baseabc.Abc_ObjFanin1(obj))
		if baseabc.Abc_ObjFaninC0(obj) != 0 {
			a = a.Not()
		}
		if baseabc.Abc_ObjFaninC1(obj) != 0 {
			b = b.Not()
		}
		copy := baseabc.Abc_AigAnd(window.FunctionManager.(*baseabc.Abc_Aig_t), a, b)
		obj.SetCopyEdge(copy)
	}
	if vec.Vec_PtrSize(cone) != int(baseabc.Abc_NtkNodeNum(window)) {
		fmt.Printf("Warning: Structural hashing during windowing reduced %d nodes (this is a bug).\n", vec.Vec_PtrSize(cone)-int(baseabc.Abc_NtkNodeNum(window)))
	}
	for _, value := range vec.Vec_PtrArray(roots) {
		obj := value.(*Abc_Obj_t)
		copy := baseabc.Abc_ObjCopyEdge(obj)
		assert(!copy.Complement, "!Abc_ObjIsComplement(pObj->pCopy)")
		po := baseabc.Abc_NtkCreatePo(window)
		_ = window.AddFanin(po, copy)
	}
	baseabc.Abc_NtkAddDummyPiNames(window)
	baseabc.Abc_NtkAddDummyPoNames(window)
	if baseabc.Abc_NtkCheck(window) == 0 {
		fmt.Println("Abc_NtkWindow: The network check has failed.")
		baseabc.Abc_NtkDelete(window)
		return nil
	}
	return window
}

func Abc_NtkRRSimulateStart(ntk *Abc_Ntk_t) {
	baseabc.Abc_ObjSetData(baseabc.Abc_AigConst1(ntk).Object, ^uint32(0))
	for i := int32(0); i < baseabc.Abc_NtkCiNum(ntk); i++ {
		baseabc.Abc_ObjSetData(baseabc.Abc_NtkCi(ntk, i), abcRrRandomWord())
	}
	for i := int32(0); i < baseabc.Abc_NtkObjNumMax(ntk); i++ {
		obj := baseabc.Abc_NtkObj(ntk, i)
		if obj == nil || obj.Kind != baseabc.ObjectNode || i == 0 {
			continue
		}
		a, b := rrWord(baseabc.Abc_ObjData(baseabc.Abc_ObjFanin0(obj))), rrWord(baseabc.Abc_ObjData(baseabc.Abc_ObjFanin1(obj)))
		if baseabc.Abc_ObjFaninC0(obj) != 0 {
			a = ^a
		}
		if baseabc.Abc_ObjFaninC1(obj) != 0 {
			b = ^b
		}
		assert(obj.Data == nil, "pObj->pData == NULL")
		baseabc.Abc_ObjSetData(obj, a&b)
	}
}
func Abc_NtkRRSimulateStop(ntk *Abc_Ntk_t) {
	for _, obj := range ntk.Objects() {
		baseabc.Abc_ObjSetData(obj, nil)
	}
}

func Abc_NtkRRSimulate(ntk *Abc_Ntk_t) *Vec_Str_t {
	targets := vec.Vec_StrStart(int(baseabc.Abc_NtkObjNumMax(ntk)) + 1)
	for i := int32(0); i < baseabc.Abc_NtkObjNumMax(ntk); i++ {
		obj := baseabc.Abc_NtkObj(ntk, i)
		if obj == nil || obj.Kind != baseabc.ObjectNode {
			continue
		}
		phase := byte(0)
		if baseabc.Abc_ObjFanoutNum(baseabc.Abc_ObjFanin1(obj)) > 1 {
			phase |= 2
		}
		if baseabc.Abc_ObjFanoutNum(baseabc.Abc_ObjFanin0(obj)) > 1 {
			phase |= 1
		}
		vec.Vec_StrWriteEntry(targets, int(obj.ID), phase)
	}
	baseabc.Abc_ObjSetCopyData(baseabc.Abc_AigConst1(ntk).Object, ^uint32(0))
	for i := int32(0); i < baseabc.Abc_NtkCiNum(ntk); i++ {
		baseabc.Abc_ObjSetCopyData(baseabc.Abc_NtkCi(ntk, i), abcRrRandomWord())
	}
	for i := int32(0); i < baseabc.Abc_NtkObjNumMax(ntk); i++ {
		obj := baseabc.Abc_NtkObj(ntk, i)
		if obj == nil || obj.Kind != baseabc.ObjectNode || i == 0 {
			continue
		}
		a, b := rrWord(baseabc.Abc_ObjData(baseabc.Abc_ObjFanin0(obj))), rrWord(baseabc.Abc_ObjData(baseabc.Abc_ObjFanin1(obj)))
		if baseabc.Abc_ObjFaninC0(obj) != 0 {
			a = ^a
		}
		if baseabc.Abc_ObjFaninC1(obj) != 0 {
			b = ^b
		}
		baseabc.Abc_ObjSetCopyData(obj, a&b)
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(ntk); i++ {
		co := baseabc.Abc_NtkCo(ntk, i)
		value := rrWord(baseabc.Abc_ObjData(baseabc.Abc_ObjFanin0(co)))
		if baseabc.Abc_ObjFaninC0(co) != 0 {
			value = ^value
		}
		baseabc.Abc_ObjSetData(co, value)
	}
	for prev := int32(0); prev < baseabc.Abc_NtkCiNum(ntk); {
		vNodes := vec.Vec_PtrAlloc(10)
		baseabc.Abc_NtkIncrementTravId(ntk)
		i := prev
		for ; i < baseabc.Abc_NtkCiNum(ntk); i++ {
			Sim_TraverseNodes_rec(baseabc.Abc_NtkCi(ntk, i), targets, vNodes)
			if vec.Vec_PtrSize(vNodes) > 128 {
				break
			}
		}
		vField := vec.Vec_PtrAlloc(10)
		baseabc.Abc_NtkIncrementTravId(ntk)
		for k := int32(0); k < baseabc.Abc_NtkCoNum(ntk); k++ {
			Sim_CollectNodes_rec(baseabc.Abc_NtkCo(ntk, k), vField)
		}
		Sim_SimulateCollected(targets, vNodes, vField)
		vec.Vec_PtrFree(vNodes)
		prev = i
	}
	for _, obj := range ntk.Objects() {
		baseabc.Abc_ObjSetData(obj, nil)
	}
	return targets
}
func Sim_TraverseNodes_rec(root *Abc_Obj_t, targets *Vec_Str_t, nodes *Vec_Ptr_t) {
	if baseabc.Abc_NodeIsTravIdCurrent(root) != 0 {
		return
	}
	baseabc.Abc_NodeSetTravIdCurrent(root)
	entry := vec.Vec_StrEntry(targets, root.ID)
	if entry&1 != 0 {
		vec.Vec_PtrPush(nodes, baseabc.Abc_ObjNot(root))
	}
	if entry&2 != 0 {
		vec.Vec_PtrPush(nodes, root)
	}
	for _, fanout := range root.Fanouts() {
		Sim_TraverseNodes_rec(fanout, targets, nodes)
	}
}
func Sim_CollectNodes_rec(root *Abc_Obj_t, field *Vec_Ptr_t) {
	if baseabc.Abc_NodeIsTravIdCurrent(root) != 0 || baseabc.Abc_NodeIsTravIdPrevious(root) == 0 {
		return
	}
	baseabc.Abc_NodeSetTravIdCurrent(root)
	for _, edge := range root.Fanins() {
		Sim_CollectNodes_rec(edge.Object, field)
	}
	if !root.IsCO() {
		baseabc.Abc_ObjSetData(root, uint32(vec.Vec_PtrSize(field)))
	}
	vec.Vec_PtrPush(field, root)
}
func Sim_SimulateCollected(targets *Vec_Str_t, nodes, field *Vec_Ptr_t) {
	// The built abcRr.c ends this routine after fetching the node fanins. Keep
	// that source behavior; the pending AND-node simulation block is commented out.
	sims := sim.Sim_UtilInfoAlloc(vec.Vec_PtrSize(field), vec.Vec_PtrSize(nodes), false)
	for i := 0; i < vec.Vec_PtrSize(field); i++ {
		obj := vec.Vec_PtrEntry(field, i).(*Abc_Obj_t)
		if obj.IsCI() {
			for k := 0; k < vec.Vec_PtrSize(nodes); k++ {
				sims[i][k] = rrWord(obj.CopyData)
			}
			continue
		}
		if !obj.IsCO() {
			_, _ = baseabc.Abc_ObjFanin0(obj), baseabc.Abc_ObjFanin1(obj)
			continue
		}
		faninData := baseabc.Abc_ObjData(baseabc.Abc_ObjFanin0(obj))
		faninIndex := int(rrWord(faninData))
		out := sims[i]
		src := sims[faninIndex]
		for k := range out {
			out[k] = src[k]
			if baseabc.Abc_ObjFaninC0(obj) != 0 {
				out[k] = ^out[k]
			}
		}
		for k := 0; k < vec.Vec_PtrSize(nodes); k++ {
			if out[k] == rrWord(obj.Data) {
				continue
			}
			disproved := vec.Vec_PtrEntry(nodes, k).(baseabc.Edge)
			phase := vec.Vec_StrEntry(targets, disproved.Object.ID)
			if disproved.Complement {
				phase &= 2
			} else {
				phase &= 1
			}
			vec.Vec_StrWriteEntry(targets, disproved.Object.ID, phase)
		}
	}
}

func rrWord(value any) uint32 {
	if word, ok := value.(uint32); ok {
		return word
	}
	return 0
}
func abcRrRandomWord() uint32 {
	return (abcFpgaSimulationRNG.next() << 24) ^ (abcFpgaSimulationRNG.next() << 12) ^ abcFpgaSimulationRNG.next()
}

const abcInfinity = int32(100000000)
