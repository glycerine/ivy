package abci

import (
	"fmt"
	"os"
	"time"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

const ABC_RS_DIV1_MAX = 150
const ABC_RS_DIV2_MAX = 500

type Abc_ManRes_t struct {
	nLeavesMax, nDivsMax                                          int32
	pRoot                                                         *Abc_Obj_t
	nLeaves, nDivs, nMffc, nLastGain                              int32
	vDivs                                                         *Vec_Ptr_t
	nBits, nWords                                                 int32
	vSims                                                         *Vec_Ptr_t
	pInfo                                                         []uint32
	pCareSet                                                      []uint32
	vDivs1UP, vDivs1UN, vDivs1B                                   *Vec_Ptr_t
	vDivs2UP0, vDivs2UP1, vDivs2UN0, vDivs2UN1                    *Vec_Ptr_t
	vTemp                                                         *Vec_Ptr_t
	timeCut, timeTruth, timeRes, timeDiv, timeMffc, timeSim       int64
	timeRes1, timeResD, timeRes2, timeRes3, timeNtk, timeTotal    int64
	nUsedNodeC, nUsedNode0, nUsedNode1Or, nUsedNode1And           int32
	nUsedNode2Or, nUsedNode2And, nUsedNode2OrAnd, nUsedNode2AndOr int32
	nUsedNode3OrAnd, nUsedNode3AndOr, nUsedNodeTotal              int32
	nTotalDivs, nTotalLeaves, nTotalGain, nNodesBeg, nNodesEnd    int32
}

func Abc_ManResubStart(nLeavesMax, nDivsMax int32) *Abc_ManRes_t {
	p := &Abc_ManRes_t{nLeavesMax: nLeavesMax, nDivsMax: nDivsMax}
	p.vDivs = vec.Vec_PtrAlloc(int(nDivsMax))
	p.nBits = 1 << uint(nLeavesMax)
	p.nWords = max(int32(1), p.nBits/32)
	p.pInfo = make([]uint32, int(p.nWords*(nDivsMax+1)))
	p.vSims = vec.Vec_PtrAlloc(int(nDivsMax))
	for i := int32(0); i < nDivsMax; i++ {
		lo := int(i * p.nWords)
		Vec_PtrPush(p.vSims, p.pInfo[lo:lo+int(p.nWords)])
	}
	lo := int(nDivsMax * p.nWords)
	p.pCareSet = p.pInfo[lo : lo+int(p.nWords)]
	for i := range p.pCareSet {
		p.pCareSet[i] = ^uint32(0)
	}
	for k := int32(0); k < nLeavesMax; k++ {
		sim := vec.Vec_PtrEntry(p.vSims, int(k)).([]uint32)
		for bit := int32(0); bit < p.nBits; bit++ {
			if bit&(1<<uint(k)) != 0 {
				sim[bit>>5] |= uint32(1) << uint(bit&31)
			}
		}
	}
	p.vDivs1UP = vec.Vec_PtrAlloc(int(nDivsMax))
	p.vDivs1UN = vec.Vec_PtrAlloc(int(nDivsMax))
	p.vDivs1B = vec.Vec_PtrAlloc(int(nDivsMax))
	p.vDivs2UP0 = vec.Vec_PtrAlloc(int(nDivsMax))
	p.vDivs2UP1 = vec.Vec_PtrAlloc(int(nDivsMax))
	p.vDivs2UN0 = vec.Vec_PtrAlloc(int(nDivsMax))
	p.vDivs2UN1 = vec.Vec_PtrAlloc(int(nDivsMax))
	p.vTemp = vec.Vec_PtrAlloc(int(nDivsMax))
	return p
}
func Abc_ManResubStop(p *Abc_ManRes_t) {
	if p == nil {
		return
	}
	for _, v := range []*Vec_Ptr_t{p.vDivs, p.vSims, p.vDivs1UP, p.vDivs1UN, p.vDivs1B, p.vDivs2UP0, p.vDivs2UP1, p.vDivs2UN0, p.vDivs2UN1, p.vTemp} {
		Vec_PtrFree(v)
	}
	p.vDivs, p.vSims = nil, nil
	p.vDivs1UP, p.vDivs1UN, p.vDivs1B = nil, nil, nil
	p.vDivs2UP0, p.vDivs2UP1, p.vDivs2UN0, p.vDivs2UN1 = nil, nil, nil, nil
	p.vTemp = nil
	p.pInfo = nil
	p.pCareSet = nil
}
func Abc_ManResubPrint(p *Abc_ManRes_t) {
	total := p.nUsedNodeC + p.nUsedNode0 + p.nUsedNode1Or + p.nUsedNode1And + p.nUsedNode2Or + p.nUsedNode2And + p.nUsedNode2OrAnd + p.nUsedNode2AndOr + p.nUsedNode3OrAnd + p.nUsedNode3AndOr
	print := func(name string, elapsed int64) {
		fmt.Printf("%s = %9.2f sec\n", name, time.Duration(elapsed).Seconds())
	}
	fmt.Printf("Used constants    = %6d.             ", p.nUsedNodeC)
	print("Cuts  ", p.timeCut)
	fmt.Printf("Used replacements = %6d.             ", p.nUsedNode0)
	print("Resub ", p.timeRes)
	fmt.Printf("Used single ORs   = %6d.             ", p.nUsedNode1Or)
	print(" Div  ", p.timeDiv)
	fmt.Printf("Used single ANDs  = %6d.             ", p.nUsedNode1And)
	print(" Mffc ", p.timeMffc)
	fmt.Printf("Used double ORs   = %6d.             ", p.nUsedNode2Or)
	print(" Sim  ", p.timeSim)
	fmt.Printf("Used double ANDs  = %6d.             ", p.nUsedNode2And)
	print(" 1    ", p.timeRes1)
	fmt.Printf("Used OR-AND       = %6d.             ", p.nUsedNode2OrAnd)
	print(" D    ", p.timeResD)
	fmt.Printf("Used AND-OR       = %6d.             ", p.nUsedNode2AndOr)
	print(" 2    ", p.timeRes2)
	fmt.Printf("Used OR-2ANDs     = %6d.             ", p.nUsedNode3OrAnd)
	print("Truth ", p.timeTruth)
	fmt.Printf("Used AND-2ORs     = %6d.             ", p.nUsedNode3AndOr)
	print("AIG   ", p.timeNtk)
	fmt.Printf("TOTAL             = %6d.             ", total)
	print("TOTAL ", p.timeTotal)
	fmt.Printf("Total leaves   = %8d.\nTotal divisors = %8d.\n", p.nTotalLeaves, p.nTotalDivs)
	gain := p.nNodesBeg - p.nNodesEnd
	pct := 0.0
	if p.nNodesBeg != 0 {
		pct = 100 * float64(gain) / float64(p.nNodesBeg)
	}
	fmt.Printf("Gain           = %8d. (%6.2f %%).\n", gain, pct)
}

func Abc_ManResubCollectDivs_rec(node *Abc_Obj_t, internal *Vec_Ptr_t) {
	if abc.Abc_NodeIsTravIdCurrent(node) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(node)
	Abc_ManResubCollectDivs_rec(abc.Abc_ObjFanin0(node), internal)
	Abc_ManResubCollectDivs_rec(abc.Abc_ObjFanin1(node), internal)
	if !node.MarkA {
		Vec_PtrPush(internal, node)
	}
}
func Abc_ManResubCollectDivs(p *Abc_ManRes_t, root *Abc_Obj_t, leaves *Vec_Ptr_t, required int32) int32 {
	Vec_PtrClear(p.vDivs1UP)
	Vec_PtrClear(p.vDivs1UN)
	Vec_PtrClear(p.vDivs1B)
	Vec_PtrClear(p.vDivs)
	abc.Abc_NtkIncrementTravId(root.Network())
	for i := 0; i < vec.Vec_PtrSize(leaves); i++ {
		n := vec.Vec_PtrEntry(leaves, i)
		Vec_PtrPush(p.vDivs, n)
		abc.Abc_NodeSetTravIdCurrent(n.(*Abc_Obj_t))
	}
	for i := 0; i < vec.Vec_PtrSize(p.vTemp); i++ {
		vec.Vec_PtrEntry(p.vTemp, i).(*Abc_Obj_t).MarkA = true
	}
	Abc_ManResubCollectDivs_rec(root, p.vDivs)
	for i := 0; i < vec.Vec_PtrSize(p.vTemp); i++ {
		vec.Vec_PtrEntry(p.vTemp, i).(*Abc_Obj_t).MarkA = false
	}
	capacity := vec.Vec_PtrSize(p.vSims) - int(p.nLeavesMax)
	used := vec.Vec_PtrSize(p.vDivs) - vec.Vec_PtrSize(leaves) + vec.Vec_PtrSize(p.vTemp)
	if used >= capacity {
		return 0
	}
	limit := capacity - used
	counter := 0
	for i := 0; i < vec.Vec_PtrSize(p.vDivs); i++ {
		node := vec.Vec_PtrEntry(p.vDivs, i).(*Abc_Obj_t)
		if abc.Abc_ObjFanoutNum(node) > 100 {
			continue
		}
		for _, fanout := range node.Fanouts() {
			if abc.Abc_NodeIsTravIdCurrent(fanout) != 0 || abc.Abc_ObjIsCo(fanout) != 0 || int32(fanout.Level) > required {
				continue
			}
			f0, f1 := abc.Abc_ObjFanin0(fanout), abc.Abc_ObjFanin1(fanout)
			if abc.Abc_NodeIsTravIdCurrent(f0) != 0 && abc.Abc_NodeIsTravIdCurrent(f1) != 0 {
				if f0 == root || f1 == root {
					continue
				}
				Vec_PtrPush(p.vDivs, fanout)
				abc.Abc_NodeSetTravIdCurrent(fanout)
				counter++
				if counter == limit {
					goto done
				}
			}
		}
	}
done:
	p.nDivs = int32(vec.Vec_PtrSize(p.vDivs))
	for i := 0; i < vec.Vec_PtrSize(p.vTemp); i++ {
		Vec_PtrPush(p.vDivs, vec.Vec_PtrEntry(p.vTemp, i))
	}
	if vec.Vec_PtrEntryLast(p.vDivs) != root {
		panic("resub divisor order lost root")
	}
	if vec.Vec_PtrSize(p.vDivs)-vec.Vec_PtrSize(leaves) > capacity {
		panic("resub divisor simulation capacity")
	}
	return 1
}
func Abc_ManResubPrintDivs(p *Abc_ManRes_t, root *Abc_Obj_t, leaves *Vec_Ptr_t) {
	for i := 0; i < vec.Vec_PtrSize(p.vDivs); i++ {
		n := vec.Vec_PtrEntry(p.vDivs, i).(*Abc_Obj_t)
		if i < vec.Vec_PtrSize(leaves) {
			fmt.Printf("%6d : %c\n", n.ID, 'a'+i)
			continue
		}
		fmt.Printf("%6d : %2d = ", n.ID, i)
		for j := 0; j < 2; j++ {
			fanin := abc.Abc_ObjFanin(n, int32(j))
			k := 0
			for k < vec.Vec_PtrSize(p.vDivs) && vec.Vec_PtrEntry(p.vDivs, k) != fanin {
				k++
			}
			if k < vec.Vec_PtrSize(leaves) {
				fmt.Printf("%c", 'a'+k)
			} else {
				fmt.Printf("%d", k)
			}
			if abc.Abc_ObjFaninC(n, int32(j)) != 0 {
				fmt.Print("'")
			}
			if j == 0 {
				fmt.Print(" ")
			}
		}
		if n == root {
			fmt.Print(" root")
		}
		fmt.Println()
	}
	fmt.Println()
}
func Abc_ManResubCleanup(p *Abc_ManRes_t) {
	if p == nil {
		return
	}
	for i := 0; i < vec.Vec_PtrSize(p.vDivs); i++ {
		if n, ok := vec.Vec_PtrEntry(p.vDivs, i).(*Abc_Obj_t); ok {
			n.Data = nil
		}
	}
	Vec_PtrClear(p.vDivs)
	p.pRoot = nil
}

func Abc_ManResubEval(p *Abc_ManRes_t, root *Abc_Obj_t, leaves *Vec_Ptr_t, steps, updateLevel, verbose int32) *dec.Dec_Graph_t {
	required := int32(util.ABC_INFINITY)
	if updateLevel != 0 {
		required = abc.Abc_ObjRequiredLevel(root)
	}
	if steps < 0 || steps > 3 {
		panic("resub steps must be in [0,3]")
	}
	p.pRoot = root
	p.nLeaves = int32(vec.Vec_PtrSize(leaves))
	p.nLastGain = -1
	started := time.Now()
	cutLeaves := make([]*abc.Object, vec.Vec_PtrSize(leaves))
	for i := range cutLeaves {
		cutLeaves[i] = vec.Vec_PtrEntry(leaves, i).(*Abc_Obj_t)
	}
	var inside []*abc.Object
	p.nMffc = abc.Abc_NodeMffcInside(root, cutLeaves, &inside)
	Vec_PtrClear(p.vTemp)
	for _, node := range inside {
		Vec_PtrPush(p.vTemp, node)
	}
	p.timeMffc += int64(time.Since(started))
	if p.nMffc <= 0 {
		panic("empty resub MFFC")
	}
	started = time.Now()
	if Abc_ManResubCollectDivs(p, root, leaves, required) == 0 {
		return nil
	}
	p.timeDiv += int64(time.Since(started))
	p.nTotalDivs += p.nDivs
	p.nTotalLeaves += p.nLeaves
	started = time.Now()
	Abc_ManResubSimulate(p.vDivs, p.nLeaves, p.vSims, p.nLeavesMax, p.nWords)
	p.timeSim += int64(time.Since(started))
	clk := time.Now()
	graph := Abc_ManResubQuit(p)
	if graph != nil {
		p.nUsedNodeC++
		p.nLastGain = p.nMffc
		return graph
	}
	graph = Abc_ManResubDivs0(p)
	if graph != nil {
		p.timeRes1 += int64(time.Since(clk))
		p.nUsedNode0++
		p.nLastGain = p.nMffc
		return graph
	}
	if steps == 0 || p.nMffc == 1 {
		p.timeRes1 += int64(time.Since(clk))
		return nil
	}
	Abc_ManResubDivsS(p, required)
	graph = Abc_ManResubDivs1(p, required)
	p.timeRes1 += int64(time.Since(clk))
	if graph != nil {
		p.nLastGain = p.nMffc - 1
		return graph
	}
	if steps == 1 || p.nMffc == 2 {
		return nil
	}
	started = time.Now()
	graph = Abc_ManResubDivs12(p, required)
	p.timeRes2 += int64(time.Since(started))
	if graph != nil {
		p.nLastGain = p.nMffc - 2
		return graph
	}
	started = time.Now()
	Abc_ManResubDivsD(p, required)
	p.timeResD += int64(time.Since(started))
	started = time.Now()
	graph = Abc_ManResubDivs2(p, required)
	p.timeRes2 += int64(time.Since(started))
	if graph != nil {
		p.nLastGain = p.nMffc - 2
		return graph
	}
	if steps == 2 || p.nMffc == 3 {
		return nil
	}
	started = time.Now()
	graph = Abc_ManResubDivs3(p, required)
	p.timeRes3 += int64(time.Since(started))
	if graph != nil {
		p.nLastGain = p.nMffc - 3
		return graph
	}
	return nil
}
func Abc_NtkResubstitute(n *Abc_Ntk_t, cutMax, stepsMax, levelsOdc, updateLevel, verbose, veryVerbose int32) int32 {
	clkStart := time.Now()
	if abc.Abc_NtkIsStrash(n) == 0 {
		panic("Abc_NtkResubstitute requires a strashed network")
	}
	aig, _ := n.FunctionManager.(*abc.Abc_Aig_t)
	if aig == nil {
		panic("resub requires AIG manager")
	}
	abc.Abc_AigCleanup(aig)
	cut := Abc_NtkManCutStart(cutMax, 100000, 100000, 100000)
	res := Abc_ManResubStart(cutMax, ABC_RS_DIV1_MAX)
	if res == nil {
		Abc_NtkManCutStop(cut)
		return 0
	}
	var odc *Odc_Man_t
	if levelsOdc > 0 {
		odc = Abc_NtkDontCareAlloc(cutMax, levelsOdc, verbose, veryVerbose)
	}
	if updateLevel != 0 {
		abc.Abc_NtkStartReverseLevels(n, 0)
	}
	savedLatchData := make(map[*abc.Object]any)
	if abc.Abc_NtkLatchNum(n) != 0 {
		for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
			node := n.Object(int(i))
			if node != nil && abc.Abc_ObjIsLatch(node) != 0 {
				savedLatchData[node] = node.Data
				if prior, ok := node.Data.(*abc.Object); ok {
					node.Next = prior
				}
			}
		}
	}
	res.nNodesBeg = abc.Abc_NtkNodeNum(n)
	count := abc.Abc_NtkObjNumMax(n)
	progress := extra.Extra_ProgressBarStart(os.Stdout, count)
	started := clkStart
	for i := int32(0); i < count; i++ {
		node := n.Object(int(i))
		if node == nil || abc.Abc_ObjIsNode(node) == 0 {
			continue
		}
		extra.Extra_ProgressBarUpdate(progress, i, "")
		if node.Persist || abc.Abc_ObjFanoutNum(node) > 1000 {
			continue
		}
		clk := time.Now()
		leaves := Abc_NodeFindCut(cut, node, 0)
		res.timeCut += int64(time.Since(clk))
		if odc != nil {
			clk = time.Now()
			Abc_NtkDontCareClear(odc)
			if len(res.pCareSet) != 0 {
				Abc_NtkDontCareCompute(odc, node, leaves, &res.pCareSet[0])
			}
			res.timeTruth += int64(time.Since(clk))
		}
		clk = time.Now()
		graph := Abc_ManResubEval(res, node, leaves, stepsMax, updateLevel, verbose)
		res.timeRes += int64(time.Since(clk))
		if graph == nil {
			continue
		}
		res.nTotalGain += res.nLastGain
		clk = time.Now()
		Dec_GraphUpdateNetwork(node, graph, updateLevel, res.nLastGain)
		res.timeNtk += int64(time.Since(clk))
		dec.Dec_GraphFree(graph)
	}
	extra.Extra_ProgressBarStop(progress)
	res.timeTotal = int64(time.Since(started))
	res.nNodesEnd = abc.Abc_NtkNodeNum(n)
	if verbose != 0 {
		Abc_ManResubPrint(res)
	}
	Abc_ManResubStop(res)
	Abc_NtkManCutStop(cut)
	if odc != nil {
		Abc_NtkDontCareFree(odc)
	}
	for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
		if node := n.Object(int(i)); node != nil {
			node.Data = nil
		}
	}
	if abc.Abc_NtkLatchNum(n) != 0 {
		for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
			node := n.Object(int(i))
			if node != nil && abc.Abc_ObjIsLatch(node) != 0 {
				node.Data = savedLatchData[node]
				node.Next = nil
			}
		}
	}
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
	s_ResubTime = int64(time.Since(clkStart))
	return 1
}

func Abc_CutVolumeCheckRec(node *Abc_Obj_t) int32 {
	if abc.Abc_NodeIsTravIdCurrent(node) != 0 {
		return 0
	}
	abc.Abc_NodeSetTravIdCurrent(node)
	if abc.Abc_ObjIsCi(node) != 0 {
		fmt.Println("Abc_CutVolumeCheck() ERROR: The set of nodes is not a cut!")
	}
	return 1 + Abc_CutVolumeCheckRec(abc.Abc_ObjFanin0(node)) + Abc_CutVolumeCheckRec(abc.Abc_ObjFanin1(node))
}
func Abc_CutVolumeCheck_rec(node *Abc_Obj_t) int32 { return Abc_CutVolumeCheckRec(node) }
func Abc_CutVolumeCheck(node *Abc_Obj_t, leaves *Vec_Ptr_t) int32 {
	abc.Abc_NtkIncrementTravId(node.Network())
	for i := 0; i < vec.Vec_PtrSize(leaves); i++ {
		abc.Abc_NodeSetTravIdCurrent(vec.Vec_PtrEntry(leaves, i).(*Abc_Obj_t))
	}
	return Abc_CutVolumeCheckRec(node)
}
func Abc_CutFactorRec(node *Abc_Obj_t, leaves *Vec_Ptr_t) {
	if node.MarkA {
		return
	}
	if abc.Abc_ObjIsCi(node) != 0 || (abc.Abc_ObjFanoutNum(node) > 1 && abc.Abc_NodeIsMuxControlType(node) == 0) {
		Vec_PtrPush(leaves, node)
		node.MarkA = true
		return
	}
	Abc_CutFactorRec(abc.Abc_ObjFanin0(node), leaves)
	Abc_CutFactorRec(abc.Abc_ObjFanin1(node), leaves)
}
func Abc_CutFactor_rec(node *Abc_Obj_t, leaves *Vec_Ptr_t) { Abc_CutFactorRec(node, leaves) }
func Abc_CutFactor(node *Abc_Obj_t) *Vec_Ptr_t {
	if abc.Abc_ObjIsCi(node) != 0 {
		panic("factor cut root must be non-CI")
	}
	leaves := vec.Vec_PtrAlloc(10)
	Abc_CutFactorRec(abc.Abc_ObjFanin0(node), leaves)
	Abc_CutFactorRec(abc.Abc_ObjFanin1(node), leaves)
	for i := 0; i < vec.Vec_PtrSize(leaves); i++ {
		vec.Vec_PtrEntry(leaves, i).(*Abc_Obj_t).MarkA = false
	}
	return leaves
}
func Abc_CutFactorLarge(node *Abc_Obj_t, maxLeaves int32) *Vec_Ptr_t {
	if abc.Abc_ObjIsNode(node) == 0 {
		panic("factor cut requires node")
	}
	leaves := Abc_CutFactor(node)
	if int32(vec.Vec_PtrSize(leaves)) > maxLeaves {
		Vec_PtrFree(leaves)
		return nil
	}
	if int32(vec.Vec_PtrSize(leaves)) == maxLeaves {
		return leaves
	}
	factors := vec.Vec_PtrAlloc(int(maxLeaves))
	abc.Abc_NtkIncrementTravId(node.Network())
	for i := 0; i < vec.Vec_PtrSize(leaves); i++ {
		leaf := vec.Vec_PtrEntry(leaves, i).(*Abc_Obj_t)
		abc.Abc_NodeSetTravIdCurrent(leaf)
		if abc.Abc_ObjIsCi(leaf) != 0 {
			Vec_PtrPush(factors, nil)
		} else {
			Vec_PtrPush(factors, Abc_CutFactor(leaf))
		}
	}
	for int32(vec.Vec_PtrSize(leaves)) < maxLeaves {
		best, bestShare := -1, -1
		for i := 0; i < vec.Vec_PtrSize(factors); i++ {
			fact, ok := vec.Vec_PtrEntry(factors, i).(*Vec_Ptr_t)
			if !ok || fact == nil {
				continue
			}
			newCount := 0
			for j := 0; j < vec.Vec_PtrSize(fact); j++ {
				if abc.Abc_NodeIsTravIdCurrent(vec.Vec_PtrEntry(fact, j).(*Abc_Obj_t)) == 0 {
					newCount++
				}
			}
			if int32(newCount) <= maxLeaves-int32(vec.Vec_PtrSize(leaves))+1 {
				share := vec.Vec_PtrSize(fact) - newCount
				if best < 0 || share > bestShare {
					best, bestShare = i, share
				}
			}
		}
		if best < 0 {
			break
		}
		oldLeaf := vec.Vec_PtrEntry(leaves, best).(*Abc_Obj_t)
		next := vec.Vec_PtrEntry(factors, best).(*Vec_Ptr_t)
		abc.Abc_NodeSetTravIdPrevious(oldLeaf)
		for i := best; i < vec.Vec_PtrSize(leaves)-1; i++ {
			vec.Vec_PtrWriteEntry(leaves, i, vec.Vec_PtrEntry(leaves, i+1))
			vec.Vec_PtrWriteEntry(factors, i, vec.Vec_PtrEntry(factors, i+1))
		}
		vec.Vec_PtrShrink(leaves, vec.Vec_PtrSize(leaves)-1)
		vec.Vec_PtrShrink(factors, vec.Vec_PtrSize(factors)-1)
		for i := 0; i < vec.Vec_PtrSize(next); i++ {
			leaf := vec.Vec_PtrEntry(next, i).(*Abc_Obj_t)
			if abc.Abc_NodeIsTravIdCurrent(leaf) != 0 {
				continue
			}
			abc.Abc_NodeSetTravIdCurrent(leaf)
			Vec_PtrPush(leaves, leaf)
			if abc.Abc_ObjIsCi(leaf) != 0 {
				Vec_PtrPush(factors, nil)
			} else {
				Vec_PtrPush(factors, Abc_CutFactor(leaf))
			}
		}
		Vec_PtrFree(next)
		if int32(vec.Vec_PtrSize(leaves)) > maxLeaves {
			panic("factor cut exceeded bound")
		}
	}
	for i := 0; i < vec.Vec_PtrSize(factors); i++ {
		if fact, ok := vec.Vec_PtrEntry(factors, i).(*Vec_Ptr_t); ok && fact != nil {
			Vec_PtrFree(fact)
		}
	}
	Vec_PtrFree(factors)
	return leaves
}
