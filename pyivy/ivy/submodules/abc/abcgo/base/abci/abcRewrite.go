package abci

import (
	"fmt"
	"os"
	"time"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/cut"
	"github.com/glycerine/abcgo/opt/rwr"
)

type Cut_Man_t = cut.Cut_Man_t
type Cut_Cut_t = cut.Cut_Cut_t
type Cut_Params_t = cut.Cut_Params_t
type Rwr_Man_t = rwr.Rwr_Man_t

// Dec_GraphUpdateNetwork is the Go binding for the C extern used by abcRewrite.c.
func Dec_GraphUpdateNetwork(pRoot *Abc_Obj_t, pGraph *dec.Dec_Graph_t, fUpdateLevel, nGain int32) {
	dec.Dec_GraphUpdateNetwork(abc.DecNetwork{Network: pRoot.Network()}, pRoot.Signal(), pGraph, fUpdateLevel, nGain)
}

// Abc_NtkRewrite performs one incremental DAG-aware rewriting pass over a strashed AIG.
func Abc_NtkRewrite(pNtk *Abc_Ntk_t, fUpdateLevel, fUseZeros, fVerbose, fVeryVerbose, fPlaceEnable int32) int32 {
	if abc.Abc_NtkIsStrash(pNtk) == 0 {
		panic("Abc_NtkRewrite requires a strashed network")
	}
	clkStart := time.Now()
	aig, _ := pNtk.FunctionManager.(*abc.Abc_Aig_t)
	if aig == nil {
		panic("Abc_NtkRewrite requires an AIG function manager")
	}
	abc.Abc_AigCleanup(aig)
	pManRwr := rwr.Rwr_ManStart(0)
	if pManRwr == nil {
		return 0
	}
	if fUpdateLevel != 0 {
		abc.Abc_NtkStartReverseLevels(pNtk, 0)
	}
	clk := time.Now()
	pManCut := Abc_NtkStartCutManForRewrite(pNtk)
	rwr.Rwr_ManAddTimeCuts(pManRwr, int64(time.Since(clk)))
	if fVeryVerbose != 0 {
		rwr.Rwr_ScoresClean(pManRwr)
	}
	rwr.Rwr_ManSetNodesBeg(pManRwr, abc.Abc_NtkNodeNum(pNtk))
	nNodes := abc.Abc_NtkObjNumMax(pNtk)
	progress := extra.Extra_ProgressBarStart(os.Stdout, nNodes)
	if fPlaceEnable != 0 {
		// The Go AIG requires update tracking before reset; the C placement setup
		// that starts tracking is disabled, so bracket this pass locally.
		abc.Abc_AigUpdateStart(aig, nil)
	}
	// C's Abc_NtkForEachNode walks object IDs in ascending order, including holes.
	for i := int32(0); i < nNodes; i++ {
		pNode := pNtk.Object(int(i))
		if pNode == nil || abc.Abc_ObjIsNode(pNode) == 0 {
			continue
		}
		extra.Extra_ProgressBarUpdate(progress, i, "")
		if i >= nNodes {
			break
		}
		if pNode.Persist || abc.Abc_ObjFanoutNum(pNode) > 1000 {
			continue
		}
		nGain := rwr.Rwr_NodeRewrite(pManRwr, pManCut, pNode, fUpdateLevel, fUseZeros, fPlaceEnable)
		if !(nGain > 0 || (nGain == 0 && fUseZeros != 0)) {
			continue
		}
		pGraph := rwr.Rwr_ManReadDecs(pManRwr)
		fCompl := rwr.Rwr_ManReadCompl(pManRwr)
		if fPlaceEnable != 0 {
			abc.Abc_AigUpdateReset(aig)
		}
		if fCompl != 0 {
			dec.Dec_GraphComplement(pGraph)
		}
		clk = time.Now()
		// C's extern Dec_GraphUpdateNetwork is implemented by bool/dec and the ABC DecNetwork adapter.
		Dec_GraphUpdateNetwork(pNode, pGraph, fUpdateLevel, nGain)
		rwr.Rwr_ManAddTimeUpdate(pManRwr, int64(time.Since(clk)))
		if fCompl != 0 {
			dec.Dec_GraphComplement(pGraph)
		}
	}
	extra.Extra_ProgressBarStop(progress)
	if fPlaceEnable != 0 {
		abc.Abc_AigUpdateStop(aig)
	}
	rwr.Rwr_ManAddTimeTotal(pManRwr, int64(time.Since(clkStart)))
	rwr.Rwr_ManSetNodesEnd(pManRwr, abc.Abc_NtkNodeNum(pNtk))
	if fVerbose != 0 {
		rwr.Rwr_ManPrintStats(pManRwr)
	}
	if fVeryVerbose != 0 {
		rwr.Rwr_ScoresReport(pManRwr)
	}
	rwr.Rwr_ManStop(pManRwr)
	cut.Cut_ManStop(pManCut)
	abc.Abc_NtkReassignIds(pNtk)
	if fUpdateLevel != 0 {
		abc.Abc_NtkStopReverseLevels(pNtk)
	} else {
		abc.Abc_NtkLevel(pNtk)
	}
	if abc.Abc_NtkCheck(pNtk) == 0 {
		fmt.Printf("Abc_NtkRewrite: The network check has failed.\n")
		return 0
	}
	return 1
}

// Abc_NtkStartCutManForRewrite starts the four-input cut manager used by rewriting.
func Abc_NtkStartCutManForRewrite(pNtk *Abc_Ntk_t) *Cut_Man_t {
	params := cut.DefaultParams(4)
	params.NKeepMax, params.FTruth, params.FFilter = 250, 1, 1
	params.NIdsMax = abc.Abc_NtkObjNumMax(pNtk)
	return cut.Abc_NtkCuts(abcRewriteCutNetwork{network: pNtk}, params)
}

// Abc_NodePrintCuts prints a node's stored cut list.
func Abc_NodePrintCuts(pNode *Abc_Obj_t) {
	fmt.Printf("\nNode %s\n", abc.Abc_ObjName(pNode))
	vCuts, _ := pNode.CopyData.(*vec.Vec_Ptr_t)
	if vCuts == nil {
		return
	}
	for k := 0; k < vec.Vec_PtrSize(vCuts); k++ {
		pCut, _ := vec.Vec_PtrEntry(vCuts, k).(*Cut_Cut_t)
		if pCut == nil {
			continue
		}
		fmt.Printf("%016b   ", pCut.USign&0xffff)
		cut.Cut_CutPrint(pCut, 0)
		fmt.Println()
	}
}

// Abc_ManRewritePrintDivs prints the divisor AIG in cut-relative form.
func Abc_ManRewritePrintDivs(vDivs *Vec_Ptr_t, nLeaves int32) {
	pRoot, _ := vec.Vec_PtrEntryLast(vDivs).(*Abc_Obj_t)
	for i := 0; i < vec.Vec_PtrSize(vDivs); i++ {
		pNode, _ := vec.Vec_PtrEntry(vDivs, i).(*Abc_Obj_t)
		if int32(i) < nLeaves {
			fmt.Printf("%6d : %c\n", pNode.ID, 'a'+i)
			continue
		}
		fmt.Printf("%6d : %2d = ", pNode.ID, i)
		printRewriteFanin(vDivs, pNode, 0, nLeaves)
		fmt.Printf("%s ", rewriteComplMark(pNode, 0))
		printRewriteFanin(vDivs, pNode, 1, nLeaves)
		fmt.Printf("%s ", rewriteComplMark(pNode, 1))
		if pNode == pRoot {
			fmt.Print(" root")
		}
		fmt.Println()
	}
	fmt.Println()
}
func printRewriteFanin(v *Vec_Ptr_t, node *Abc_Obj_t, index int, leaves int32) {
	fanin := abc.Abc_ObjFanin(node, int32(index))
	k := vec.Vec_PtrSize(v)
	for i := 0; i < vec.Vec_PtrSize(v); i++ {
		if x, _ := vec.Vec_PtrEntry(v, i).(*Abc_Obj_t); x == fanin {
			k = i
			break
		}
	}
	if int32(k) < leaves {
		fmt.Printf("%c", 'a'+k)
	} else {
		fmt.Printf("%d", k)
	}
}
func rewriteComplMark(node *Abc_Obj_t, index int) string {
	if abc.Abc_ObjFaninC(node, int32(index)) != 0 {
		return "'"
	}
	return ""
}

// Abc_ManShowCutCone_rec traverses the cone in fanin-first order, once per node.
func Abc_ManShowCutCone_rec(pNode *Abc_Obj_t, vDivs *Vec_Ptr_t) {
	if abc.Abc_NodeIsTravIdCurrent(pNode) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(pNode)
	Abc_ManShowCutCone_rec(abc.Abc_ObjFanin0(pNode), vDivs)
	Abc_ManShowCutCone_rec(abc.Abc_ObjFanin1(pNode), vDivs)
	vec.Vec_PtrPush(vDivs, pNode)
}

// Abc_ManShowCutCone prints the cone rooted at pNode after marking cut leaves.
func Abc_ManShowCutCone(pNode *Abc_Obj_t, vLeaves *Vec_Ptr_t) {
	vDivs := vec.Vec_PtrAlloc(100)
	abc.Abc_NtkIncrementTravId(pNode.Network())
	for i := 0; i < vec.Vec_PtrSize(vLeaves); i++ {
		leaf, _ := vec.Vec_PtrEntry(vLeaves, i).(*Abc_Obj_t)
		leaf = abc.Abc_ObjRegular(leaf)
		abc.Abc_NodeSetTravIdCurrent(leaf)
		vec.Vec_PtrPush(vDivs, leaf)
	}
	Abc_ManShowCutCone_rec(pNode, vDivs)
	Abc_ManRewritePrintDivs(vDivs, int32(vec.Vec_PtrSize(vLeaves)))
	vec.Vec_PtrFree(vDivs)
}

// Abc_RwrExpWithCut_rec marks leaves in one root's expansion.
func rewriteEdge(value any) abc.Edge {
	switch edge := value.(type) {
	case abc.Edge:
		return edge
	case *Abc_Obj_t:
		return edge.Signal()
	default:
		panic("rewrite: expected object signal")
	}
}

func rewriteCutContains(vLeaves *Vec_Ptr_t, signal abc.Edge) bool {
	for i := 0; i < vec.Vec_PtrSize(vLeaves); i++ {
		leaf := vec.Vec_PtrEntry(vLeaves, i)
		if rewriteEdge(leaf) == signal {
			return true
		}
	}
	return false
}

func Abc_RwrExpWithCut_rec(pNode any, vLeaves *Vec_Ptr_t, fUseA int32) {
	edge := rewriteEdge(pNode)
	if rewriteCutContains(vLeaves, edge) || rewriteCutContains(vLeaves, edge.Not()) {
		pNode := edge.Object
		if fUseA != 0 {
			pNode.MarkA = true
		} else {
			pNode.MarkB = true
		}
		return
	}
	if abc.Abc_ObjIsNode(edge.Object) == 0 {
		panic("Abc_RwrExpWithCut_rec reached a non-node before a cut leaf")
	}
	Abc_RwrExpWithCut_rec(abc.Abc_ObjChild0(edge.Object), vLeaves, fUseA)
	Abc_RwrExpWithCut_rec(abc.Abc_ObjChild1(edge.Object), vLeaves, fUseA)
}

// Abc_RwrExpWithCut counts leaves used by each fanin expansion and prints the overlap.
func Abc_RwrExpWithCut(pNode *Abc_Obj_t, vLeaves *Vec_Ptr_t) {
	Abc_RwrExpWithCut_rec(abc.Abc_ObjChild0(pNode), vLeaves, 1)
	Abc_RwrExpWithCut_rec(abc.Abc_ObjChild1(pNode), vLeaves, 0)
	var countA, countB int32
	for i := 0; i < vec.Vec_PtrSize(vLeaves); i++ {
		leaf, _ := vec.Vec_PtrEntry(vLeaves, i).(*Abc_Obj_t)
		leaf = abc.Abc_ObjRegular(leaf)
		if leaf.MarkA {
			countA++
		}
		if leaf.MarkB {
			countB++
		}
		leaf.MarkA = false
		leaf.MarkB = false
	}
	fmt.Printf("(%d,%d:%d) ", countA, countB, countA+countB-int32(vec.Vec_PtrSize(vLeaves)))
}
