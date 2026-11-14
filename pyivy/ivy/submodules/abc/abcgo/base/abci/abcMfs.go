package abci

import (
	"fmt"
	"os"

	baseabc "github.com/glycerine/abcgo/base/abc"
	baseio "github.com/glycerine/abcgo/base/io"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/sfm"
)

type Vec_Wec_t = vec.Vec_Wec_t
type Sfm_Ntk_t = sfm.Sfm_Ntk_t
type Sfm_Par_t = sfm.Sfm_Par_t

/**CFile****************************************************************

  FileName    [abcMfs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Optimization with don't-cares.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcMfs.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "bool/kit/kit.h"
//#include "opt/sfm/sfm.h"
//#include "base/io/ioAbc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkAssignIDs(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
	nodes := baseabc.Abc_NtkDfs(pNtk, 0)
	vNodes := vec.Vec_PtrAlloc(len(nodes))
	baseabc.Abc_NtkCleanCopy(pNtk)
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		baseabc.Abc_NtkCi(pNtk, i).Temp = int(i)
	}
	for i, pObj := range nodes {
		pObj.Temp = int(baseabc.Abc_NtkCiNum(pNtk)) + i
		vec.Vec_PtrPush(vNodes, pObj)
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		baseabc.Abc_NtkCo(pNtk, i).Temp = int(baseabc.Abc_NtkCiNum(pNtk) + int32(len(nodes)) + i)
	}
	return vNodes
}
func Abc_NtkAssignIDs2(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
	baseabc.Abc_NtkCleanCopy(pNtk)
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		baseabc.Abc_NtkCi(pNtk, i).Temp = int(i)
	}
	vNodes := vec.Vec_PtrAlloc(len(pNtk.ObjectsOfKind(baseabc.ObjectNode)))
	for _, pObj := range pNtk.ObjectsOfKind(baseabc.ObjectNode) {
		pObj.Temp = int(baseabc.Abc_NtkCiNum(pNtk)) + vec.Vec_PtrSize(vNodes)
		vec.Vec_PtrPush(vNodes, pObj)
	}
	assert(int32(vec.Vec_PtrSize(vNodes)) == baseabc.Abc_NtkNodeNum(pNtk), "Vec_PtrSize(vNodes) == Abc_NtkNodeNum(pNtk)")
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		baseabc.Abc_NtkCo(pNtk, i).Temp = int(baseabc.Abc_NtkCiNum(pNtk) + int32(vec.Vec_PtrSize(vNodes)) + i)
	}
	return vNodes
}

/*
*Function*************************************************************

	Synopsis    [Assign truth table sizes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkAssignStarts(pNtk *Abc_Ntk_t, vNodes *Vec_Ptr_t, pnTotal *int32) *Vec_Int_t {
	var Counter int32
	vStarts := vec.Vec_IntStart(int(baseabc.Abc_NtkObjNum(pNtk)))
	for i := int32(0); i < Vec_PtrSize(vNodes); i++ {
		pObj := Vec_PtrEntry(vNodes, i).(*Abc_Obj_t)
		vec.Vec_IntWriteEntry(vStarts, pObj.Temp, Counter)
		Counter += util.Abc_Truth6WordNum(baseabc.Abc_ObjFaninNum(pObj))
	}
	*pnTotal = Counter
	return vStarts
}

var spStatic_Abc_NtkFillTruthStore_Truth6 [6]word = [6]word{
	0xAAAAAAAAAAAAAAAA,
	0xCCCCCCCCCCCCCCCC,
	0xF0F0F0F0F0F0F0F0,
	0xFF00FF00FF00FF00,
	0xFFFF0000FFFF0000,
	0xFFFFFFFF00000000,
}

func Abc_NtkFillTruthStore(TruthStore [][1 << 10]word) {
	if TruthStore[0][0] == 0 {

		var nVarsMax int32 = 16
		var nWordsMax int32 = (1 << 10)
		var i, k int32
		assert(nVarsMax <= 16, "nVarsMax <= 16")
		for i = 0; i < 6; i++ {
			for k = 0; k < nWordsMax; k++ {
				TruthStore[i][k] = spStatic_Abc_NtkFillTruthStore_Truth6[i]
			}
		}
		for i = 6; i < nVarsMax; i++ {
			for k = 0; k < nWordsMax; k++ {
				if ((k >> (i - 6)) & 1) != 0 {
					TruthStore[i][k] = ^word(0)
				} else {
					TruthStore[i][k] = 0
				}
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Extracts information about the network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkExtractMfs(pNtk *Abc_Ntk_t, nFirstFixed int32) *Sfm_Ntk_t {
	assert(nFirstFixed >= 0 && nFirstFixed < baseabc.Abc_NtkNodeNum(pNtk), "nFirstFixed >= 0 && nFirstFixed < Abc_NtkNodeNum(pNtk)")
	var nodes *Vec_Ptr_t
	if nFirstFixed == 0 {
		nodes = Abc_NtkAssignIDs(pNtk)
	} else {
		nodes = Abc_NtkAssignIDs2(pNtk)
	}
	return mfsExtract(pNtk, nodes, nFirstFixed, false)
}
func Abc_NtkExtractMfs2(pNtk *Abc_Ntk_t, iPivot int32) *Sfm_Ntk_t {
	return mfsExtract(pNtk, Abc_NtkAssignIDs2(pNtk), iPivot, true)
}

func mfsExtract(pNtk *Abc_Ntk_t, nodes *Vec_Ptr_t, fixedFrom int32, fixedByObjectID bool) *Sfm_Ntk_t {
	nObjects := int(Abc_NtkCiNum(pNtk) + Vec_PtrSize(nodes) + Abc_NtkCoNum(pNtk))
	fanins := make([][]int, nObjects)
	fixed := make([]bool, nObjects)
	truths := make([]uint64, nObjects)
	starts := make([]int, nObjects)

	var total int
	for i := int32(0); i < Vec_PtrSize(nodes); i++ {
		object := Vec_PtrEntry(nodes, i).(*Abc_Obj_t)
		starts[object.Temp] = total
		total += int(util.Abc_Truth6WordNum(baseabc.Abc_ObjFaninNum(object)))
	}
	largeTruths := make([]uint64, total)
	truthStore := [16][1 << 10]uint64{}
	Abc_NtkFillTruthStore(truthStore[:])
	truthVars := make([][]uint64, len(truthStore))
	for i := range truthStore {
		truthVars[i] = truthStore[i][:]
	}
	var cube [1 << 10]uint64

	for i := int32(0); i < Vec_PtrSize(nodes); i++ {
		object := Vec_PtrEntry(nodes, i).(*Abc_Obj_t)
		faninCount := baseabc.Abc_ObjFaninNum(object)
		sop, ok := object.Data.(string)
		if !ok {
			panic("abcMfs: node data is not an SOP string")
		}
		if faninCount <= 6 {
			truth := baseabc.Abc_SopToTruth(sop, faninCount)
			truths[object.Temp] = truth
			if truth == 0 || ^truth == 0 {
				continue
			}
		} else {
			nWords := int(util.Abc_Truth6WordNum(faninCount))
			truth := largeTruths[starts[object.Temp] : starts[object.Temp]+nWords]
			baseabc.Abc_SopToTruthBig(sop, faninCount, truthVars, cube[:], truth)
			allZero, allOne := true, true
			for _, word := range truth {
				allZero = allZero && word == 0
				allOne = allOne && word == ^uint64(0)
			}
			if allZero || allOne {
				continue
			}
		}
		faninIDs := make([]int, faninCount)
		for k := range faninIDs {
			faninIDs[k] = int(baseabc.Abc_ObjFanin(object, int32(k)).Temp)
		}
		fanins[object.Temp] = faninIDs
	}
	for i := int32(0); i < Abc_NtkCoNum(pNtk); i++ {
		object := baseabc.Abc_NtkCo(pNtk, i)
		faninIDs := make([]int, baseabc.Abc_ObjFaninNum(object))
		for k := range faninIDs {
			faninIDs[k] = int(baseabc.Abc_ObjFanin(object, int32(k)).Temp)
		}
		fanins[object.Temp] = faninIDs
	}
	if fixedByObjectID {
		for objectID := int32(0); objectID < baseabc.Abc_NtkObjNumMax(pNtk); objectID++ {
			object := baseabc.Abc_NtkObj(pNtk, objectID)
			if object != nil && object.Kind == baseabc.ObjectNode && objectID >= fixedFrom {
				fixed[object.Temp] = true
			}
		}
	} else {
		for i := Abc_NtkCiNum(pNtk); i < Abc_NtkCiNum(pNtk)+fixedFrom; i++ {
			fixed[i] = true
		}
	}
	model, err := sfm.Sfm_NtkConstruct(fanins, int(baseabc.Abc_NtkCiNum(pNtk)), int(baseabc.Abc_NtkCoNum(pNtk)), fixed, nil, truths, starts, largeTruths)
	if err != nil {
		panic(fmt.Sprintf("abcMfs: construct SFM network: %v", err))
	}
	Vec_PtrFree(nodes)
	return model
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkInsertMfs(pNtk *Abc_Ntk_t, p *Sfm_Ntk_t) {
	vMap := make([]int32, int(baseabc.Abc_NtkObjNumMax(pNtk)))
	for i := int32(0); i < Abc_NtkCiNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCi(pNtk, i)
		vMap[pNode.Temp] = int32(pNode.ID)
	}
	nodes := pNtk.ObjectsOfKind(baseabc.ObjectNode)
	for _, pNode := range nodes {
		if pNode.Temp > 0 {
			vMap[pNode.Temp] = int32(pNode.ID)
		}
	}
	for _, pNode := range nodes {
		if !sfm.Sfm_NodeReadFixed(p, int(pNode.Temp)) {
			baseabc.Abc_ObjRemoveFanins(pNode)
		}
	}
	cover := vec.Vec_IntAlloc(1 << 16)
	defer vec.Vec_IntFree(cover)
	for _, pNode := range nodes {
		id := int(pNode.Temp)
		if id == 0 || sfm.Sfm_NodeReadFixed(p, id) {
			continue
		}
		if !sfm.Sfm_NodeReadUsed(p, id) {
			baseabc.Abc_NtkDeleteObj(pNode)
			continue
		}
		faninIDs := sfm.Sfm_NodeReadFanins(p, id)
		for _, fanin := range faninIDs {
			baseabc.Abc_ObjAddFanin(pNode, baseabc.Abc_NtkObj(pNtk, vMap[fanin]))
		}
		truth := sfm.Sfm_NodeReadTruth(p, id)
		faninCount := int32(len(faninIDs))
		pNode.Data = baseabc.Abc_SopCreateFromTruthIsop(pNtk.FunctionManager.(*mem.Mem_Flex_t), faninCount, truth, cover)
		assert(baseabc.Abc_SopGetVarNum(pNode.Data.(string)) == faninCount, "Abc_SopGetVarNum(pNode.Data) == faninCount")
	}
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPerformMfs(pNtk *Abc_Ntk_t, pars *Sfm_Par_t) int32 {
	if baseabc.Abc_NtkIsLogic(pNtk) == 0 {
		panic("Abc_NtkPerformMfs requires a logic network")
	}
	baseabc.Abc_NtkSweep(pNtk, 0)
	if faninMax := baseabc.Abc_NtkGetFaninMax(pNtk); faninMax > 15 {
		fmt.Println("Currently \"mfs\" cannot process the network containing nodes with more than 15 fanins.")
		return 1
	}
	if baseabc.Abc_NtkHasSop(pNtk) == 0 && baseabc.Abc_NtkToSop(pNtk, -1, util.ABC_INFINITY) == 0 {
		fmt.Println("Conversion to SOP has failed due to low resource limit.")
		return 0
	}
	model := Abc_NtkExtractMfs(pNtk, int32(pars.FirstFixed))
	nodesChanged := sfm.Sfm_NtkPerform(model, pars)
	if nodesChanged != 0 {
		Abc_NtkInsertMfs(pNtk, model)
		if pars.Verbose {
			fmt.Printf("The network has %d nodes changed by \"mfs\".\n", nodesChanged)
		}
	}
	sfm.Sfm_NtkFree(model)
	return 1
}

// Abc_NtkUnrollAndDrop unrolls the logic network while dropping selected
// next-state functions, retaining ABC's frame and CO ordering.
func Abc_NtkUnrollAndDrop(p *Abc_Ntk_t, nFrames, nFramesAdd int32, vFlops *Vec_Int_t, piPivot *int32) *Abc_Ntk_t {
	assert(nFramesAdd >= 0, "nFramesAdd >= 0")
	assert(baseabc.Abc_NtkIsLogic(p) != 0, "Abc_NtkIsLogic(p)")
	assert(Vec_IntSize(vFlops) == baseabc.Abc_NtkLatchNum(p), "Vec_IntSize(vFlops) == Abc_NtkLatchNum(p)")
	*piPivot = -1
	pNtk := baseabc.Abc_NtkAlloc(baseabc.Abc_NtkType_t(p.Type), baseabc.Abc_NtkFunc_t(p.Function), 1)
	pNtk.Name = extra.Extra_UtilStrsav(baseabc.Abc_NtkName(p))
	for i := int32(0); i < baseabc.Abc_NtkCiNum(p); i++ {
		baseabc.Abc_NtkCi(p, i).Copy = baseabc.Abc_NtkCreatePi(pNtk)
	}
	nodes := baseabc.Abc_NtkDfs(p, 0)
	for frame := int32(0); frame <= nFrames+nFramesAdd; frame++ {
		if frame > 0 {
			for i := int32(0); i < baseabc.Abc_NtkPiNum(p); i++ {
				baseabc.Abc_NtkPi(p, i).Copy = baseabc.Abc_NtkCreatePi(pNtk)
			}
		}
		for _, node := range nodes {
			baseabc.Abc_NtkDupObj(pNtk, node, 0)
			for _, fanin := range node.Fanins() {
				baseabc.Abc_ObjAddFanin(node.Copy, fanin.Object.Copy)
			}
		}
		for i := int32(0); i < baseabc.Abc_NtkCoNum(p); i++ {
			co := baseabc.Abc_NtkCo(p, i)
			co.Copy = baseabc.Abc_ObjFanin0(co).Copy
		}
		for i := int32(0); i < baseabc.Abc_NtkPoNum(p); i++ {
			po := baseabc.Abc_NtkPo(p, i)
			baseabc.Abc_ObjAddFanin(baseabc.Abc_NtkCreatePo(pNtk), po.Copy)
		}
		if frame == 0 {
			*piPivot = baseabc.Abc_NtkObjNum(pNtk)
		}
		for _, latch := range p.ObjectsOfKind(baseabc.ObjectLatch) {
			baseabc.Abc_ObjFanout0(latch).Copy = baseabc.Abc_ObjFanin0(latch).Copy
		}
		if frame > nFramesAdd {
			for i := int32(0); i < Vec_IntSize(vFlops); i++ {
				if Vec_IntEntry(vFlops, i) == 0 {
					continue
				}
				co := baseabc.Abc_NtkCo(p, baseabc.Abc_NtkPoNum(p)+i)
				baseabc.Abc_ObjAddFanin(baseabc.Abc_NtkCreatePo(pNtk), co.Copy)
			}
		}
	}
	baseabc.Abc_NtkAddDummyPiNames(pNtk)
	baseabc.Abc_NtkAddDummyPoNames(pNtk)
	baseabc.Abc_NtkCleanup(pNtk, 0)
	if baseabc.Abc_NtkCheck(pNtk) == 0 {
		fmt.Fprintln(os.Stdout, "Abc_NtkCreateFromNode(): Network check has failed.")
	}
	return pNtk
}

// Abc_NtkReinsertNodes maps optimized unrolled nodes back to their originals.
func Abc_NtkReinsertNodes(p, pNtk *Abc_Ntk_t, iPivot int32) {
	assert(baseabc.Abc_NtkIsLogic(p) != 0, "Abc_NtkIsLogic(p)")
	assert(baseabc.Abc_NtkCiNum(p) <= baseabc.Abc_NtkCiNum(pNtk), "Abc_NtkCiNum(p) <= Abc_NtkCiNum(pNtk)")
	nodes := baseabc.Abc_NtkDfs(p, 0)
	baseabc.Abc_NtkCleanCopy(p)
	for _, node := range p.ObjectsOfKind(baseabc.ObjectNode) {
		baseabc.Abc_ObjRemoveFanins(node)
		node.Data = baseabc.Abc_SopRegister(p.FunctionManager.(*mem.Mem_Flex_t), " 0\n")
	}
	for i := int32(0); i < baseabc.Abc_NtkCiNum(p); i++ {
		baseabc.Abc_NtkCi(pNtk, i).Copy = baseabc.Abc_NtkCi(p, i)
	}
	assert(int32(len(nodes))+baseabc.Abc_NtkCiNum(p)+baseabc.Abc_NtkPoNum(p) == iPivot, "DFS node count + CIs + POs == pivot")
	for i, node := range nodes {
		newNode := baseabc.Abc_NtkObj(pNtk, baseabc.Abc_NtkCiNum(p)+int32(i)+1)
		if newNode != nil {
			newNode.Copy = node
		}
	}
	for i, node := range nodes {
		newNode := baseabc.Abc_NtkObj(pNtk, baseabc.Abc_NtkCiNum(p)+int32(i)+1)
		if newNode == nil {
			continue
		}
		assert(newNode.Copy == node, "newNode.Copy == node")
		for _, fanin := range newNode.Fanins() {
			baseabc.Abc_ObjAddFanin(newNode.Copy, fanin.Object.Copy)
		}
		node.Data = baseabc.Abc_SopRegister(p.FunctionManager.(*mem.Mem_Flex_t), newNode.Data.(string))
	}
}

// Abc_NtkMfsAfterICheck performs MFS on a dropped-flop unrolling and copies
// the optimized logic back to the original network.
func Abc_NtkMfsAfterICheck(p *Abc_Ntk_t, nFrames, nFramesAdd int32, vFlops *Vec_Int_t, pars *Sfm_Par_t) int32 {
	assert(baseabc.Abc_NtkIsLogic(p) != 0, "Abc_NtkIsLogic(p)")
	if baseabc.Abc_NtkGetFaninMax(p) > 15 {
		fmt.Println("Currently \"mfs\" cannot process the network containing nodes with more than 15 fanins.")
		return 0
	}
	if baseabc.Abc_NtkHasSop(p) == 0 {
		baseabc.Abc_NtkToSop(p, -1, util.ABC_INFINITY)
	}
	var pivot int32
	unrolled := Abc_NtkUnrollAndDrop(p, nFrames, nFramesAdd, vFlops, &pivot)
	_ = baseio.Io_WriteBlifLogic(unrolled, "unroll_dump.blif", 0)
	model := Abc_NtkExtractMfs2(unrolled, pivot)
	nodesChanged := sfm.Sfm_NtkPerform(model, pars)
	if nodesChanged != 0 {
		Abc_NtkInsertMfs(unrolled, model)
		if pars.Verbose {
			fmt.Printf("The network has %d nodes changed by \"mfs\".\n", nodesChanged)
		}
		Abc_NtkReinsertNodes(p, unrolled, pivot)
	}
	baseabc.Abc_NtkDelete(unrolled)
	sfm.Sfm_NtkFree(model)
	baseabc.Abc_NtkSweep(p, 0)
	if baseabc.Abc_NtkHasSop(p) == 0 {
		baseabc.Abc_NtkToSop(p, -1, util.ABC_INFINITY)
	}
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
