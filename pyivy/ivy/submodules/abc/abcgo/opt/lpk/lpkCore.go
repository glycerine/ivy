package lpk

import (
	"fmt"
	hop "github.com/glycerine/abcgo/aig/hop"
	abc "github.com/glycerine/abcgo/base/abc"
	abci "github.com/glycerine/abcgo/base/abci"
	abcmain "github.com/glycerine/abcgo/base/main"
	kit "github.com/glycerine/abcgo/bool/kit"
	ifpkg "github.com/glycerine/abcgo/map/if"
	extra "github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/mem"
	util "github.com/glycerine/abcgo/misc/util"
	vec "github.com/glycerine/abcgo/misc/vec"
	"os"
)

/**CFile****************************************************************

  FileName    [lpkCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkCore.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "lpkInt.h"
//#include "bool/kit/cloud.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Returns decomposed network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDecFromTruth(truth []uint32, nVars, nLutSize int32) *Abc_Ntk_t {
	if nVars < 0 || nVars >= 16 || int32(len(truth)) < kit.Kit_TruthWordNum(nVars) || nLutSize < 1 {
		return nil
	}
	cover := vec.Vec_IntAlloc(1 << 16)
	defer vec.Vec_IntFree(cover)
	manager := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	functionManager, ok := manager.FunctionManager.(*mem.Mem_Flex_t)
	if !ok || functionManager == nil {
		abc.Abc_NtkDelete(manager)
		return nil
	}
	// ABC's C API accepts word* (uint64); LPK stores truth tables as
	// uint32 words, so pack adjacent halves into the API's word layout.
	truthWords := lpkPackTruthWords(truth)
	sop := abc.Abc_SopCreateFromTruthIsop(functionManager, nVars, truthWords, cover)
	abc.Abc_NtkDelete(manager)
	initial := abc.Abc_NtkCreateWithNode(sop)
	network := abci.Abc_NtkLutmin(initial, nLutSize, 0)
	abc.Abc_NtkDelete(initial)
	if network == nil {
		return nil
	}
	if abc.Abc_NtkToAig(network) == 0 {
		abc.Abc_NtkDelete(network)
		return nil
	}
	return network
}

func lpkPackTruthWords(truth []uint32) []uint64 {
	words := make([]uint64, (len(truth)+1)/2)
	for i, value := range truth {
		words[i/2] |= uint64(value) << uint((i%2)*32)
	}
	return words
}

func Abc_NtkLutMinDecompose(network *Abc_Ntk_t, leaves *Vec_Ptr_t, truth []uint32, nLutSize, required int32) *Abc_Obj_t {
	if network == nil || leaves == nil {
		return nil
	}
	newNetwork := Abc_NtkDecFromTruth(truth, int32(vec.Vec_PtrSize(leaves)), nLutSize)
	if newNetwork == nil {
		return nil
	}
	defer abc.Abc_NtkDelete(newNetwork)
	nodes := abc.Abc_NtkDfs(newNetwork, 0)
	if len(nodes) == 0 {
		return nil
	}
	for i := 0; i < vec.Vec_PtrSize(leaves); i++ {
		newNetworkCI := abc.Abc_NtkCi(newNetwork, int32(i))
		leaf := vec.Vec_PtrEntry(leaves, i).(*abc.Object)
		newNetworkCI.Level = leaf.Level
		newNetworkCI.Copy = leaf
	}
	for _, node := range nodes {
		var level int32
		for _, edge := range node.Fanins() {
			if edge.Object != nil && int32(edge.Object.Level)+1 > level {
				level = int32(edge.Object.Level) + 1
			}
		}
		node.Level = int(level)
	}
	if int32(nodes[len(nodes)-1].Level) > required {
		return nil
	}
	var result *abc.Object
	for _, node := range nodes {
		copyNode := abc.Abc_NtkDupObj(network, node, 0)
		copyNode.Level = 0
		for _, edge := range node.Fanins() {
			fanin := edge.Object.Copy
			faninEdge := fanin.Signal()
			if edge.Complement {
				faninEdge = faninEdge.Not()
			}
			abc.Abc_ObjAddFanin(copyNode, faninEdge)
			if copyNode.Level < fanin.Level {
				copyNode.Level = fanin.Level
			}
		}
		copyNode.Level++
		node.Copy = copyNode
		result = copyNode
	}
	return result
}

/*
*Function*************************************************************

	Synopsis    [Prepares the mapping manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_IfManStart(p *Lpk_Man_t) {
	assert(p.pIfMan == nil, "p.pIfMan == nil")
	pPars := new(If_Par_t)
	ifpkg.If_ManSetLpkPars(pPars, p.pPars.nLutSize)
	// start the mapping manager and set its parameters
	p.pIfMan = ifpkg.If_ManStart(pPars)
	ifpkg.If_ManSetupSetAll(p.pIfMan, 1000)
	pPars.SetArrivalTimes(make([]float32, 32))
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if at least one entry has changed.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_NodeHasChanged(p *Lpk_Man_t, iNode int32) int32 {
	if p == nil || p.vVisited == nil {
		return 1
	}
	nodes := vec.Vec_VecEntry(p.vVisited, int(iNode))
	if nodes == nil || vec.Vec_PtrSize(nodes) == 0 {
		return 1
	}
	for i := 0; i < vec.Vec_PtrSize(nodes); i += 2 {
		var id int32
		switch value := vec.Vec_PtrEntry(nodes, i).(type) {
		case int:
			id = int32(value)
		case int32:
			id = value
		default:
			return 1
		}
		if abc.Abc_NtkObj(p.pNtk, id) == nil {
			return 1
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Prepares the mapping manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_ExploreCut(p *Lpk_Man_t, pCut *Lpk_Cut_t, pNtk *Kit_DsdNtk_t) int32 {

	var pRoot *Kit_DsdObj_t
	var pDriver *If_Obj_t
	var ppLeaves [16]*If_Obj_t
	var pLeaf, pObjNew *Abc_Obj_t
	var nGain, i int32
	var clk abctime
	var nNodesBef int32
	//    int nOldShared;

	// check special cases
	pRoot = kit.Kit_DsdNtkRoot(pNtk)
	if pRoot.Type == kit.KIT_DSD_CONST1 {
		if util.Abc_LitIsCompl(int32(pNtk.Root)) != 0 {
			pObjNew = abc.Abc_NtkCreateNodeConst0(p.pNtk)
		} else {
			pObjNew = abc.Abc_NtkCreateNodeConst1(p.pNtk)
		}
		abc.Abc_NtkUpdate(p.pObj, pObjNew, p.vLevels)
		p.nGainTotal += int32(pCut.nNodes) - int32(pCut.nNodesDup)
		return 1
	}
	if pRoot.Type == kit.KIT_DSD_VAR {
		pObjNew = abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[util.Abc_Lit2Var(int32(kit.Kit_DsdObjFans(pRoot)[0]))])
		if util.Abc_LitIsCompl(int32(pNtk.Root))^util.Abc_LitIsCompl(int32(kit.Kit_DsdObjFans(pRoot)[0])) != 0 {
			pObjNew = abc.Abc_NtkCreateNodeInv(p.pNtk, pObjNew)
		}
		abc.Abc_NtkUpdate(p.pObj, pObjNew, p.vLevels)
		p.nGainTotal += int32(pCut.nNodes) - int32(pCut.nNodesDup)
		return 1
	}
	assert(pRoot.Type == kit.KIT_DSD_AND || pRoot.Type == kit.KIT_DSD_XOR || pRoot.Type == kit.KIT_DSD_PRIME, "pRoot.Type == kit.KIT_DSD_AND || pRoot.Type == kit.KIT_DSD_XOR || pRoot.Type == kit.KIT_DSD_PRIME")

	// start the mapping manager
	if p.pIfMan == nil {
		Lpk_IfManStart(p)
	}

	// prepare the mapping manager
	ifpkg.If_ManRestart(p.pIfMan)
	// create the PI variables
	for i = 0; i < p.pPars.nVarsMax; i++ {
		ppLeaves[i] = ifpkg.If_ManCreateCi(p.pIfMan)
	}
	// set the arrival times
	// inlining of macro Lpk_CutForEachLeaf
	for i = 0; i < int32(pCut.nLeaves); i++ {
		pLeaf = abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[i])
		p.pIfMan.Parameters().SetArrivalTime(int(i), float32(pLeaf.Level))
	}
	// prepare the PI cuts
	ifpkg.If_ManSetupCiCutSets(p.pIfMan)
	// create the internal nodes
	p.fCalledOnce = 0
	p.nCalledSRed = 0
	pDriver = Lpk_MapTree_rec(p, pNtk, ppLeaves[:], int32(pNtk.Root), nil)
	if pDriver == nil {
		return 0
	}
	// create the PO node
	ifpkg.If_ManCreateCo(p.pIfMan, ifpkg.If_Regular(pDriver))

	// perform mapping
	p.pIfMan.Parameters().SetAreaOnly(true)
	clk = abctime(util.Abc_Clock())
	ifpkg.If_ManPerformMappingComb(p.pIfMan)
	p.timeMap += abctime(util.Abc_Clock()) - clk

	// compute the gain in area
	nGain = int32(pCut.nNodes) - int32(pCut.nNodesDup) - int32(p.pIfMan.AreaGlo)
	if p.pPars.fVeryVerbose != 0 {
		fmt.Printf("       Mffc = %2d. Mapped = %2d. Gain = %3d. Depth increase = %d. SReds = %d.\n", pCut.nNodes-pCut.nNodesDup, int32(p.pIfMan.AreaGlo), nGain, int32(p.pIfMan.RequiredGlo)-int32(p.pObj.Level), p.nCalledSRed)
	}

	// quit if there is no gain
	if !(nGain > 0 || (p.pPars.fZeroCost != 0 && nGain == 0)) {
		return 0
	}

	// quit if depth increases too much
	if int32(p.pIfMan.RequiredGlo) > abc.Abc_ObjRequiredLevel(p.pObj) {
		return 0
	}

	// perform replacement
	p.nGainTotal += nGain
	p.nChanges++
	if p.nCalledSRed != 0 {
		p.nBenefited++
	}

	nNodesBef = abc.Abc_NtkNodeNum(p.pNtk)
	_ = nNodesBef
	// prepare the mapping manager
	ifpkg.If_ManCleanNodeCopy(p.pIfMan)
	ifpkg.If_ManCleanCutData(p.pIfMan)
	// set the PIs of the cut
	// inlining of macro Lpk_CutForEachLeaf
	for i = 0; i < int32(pCut.nLeaves); i++ {
		pLeaf = abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[i])
		ifpkg.If_ObjSetCopy(ifpkg.If_ManCi(p.pIfMan, i), pLeaf)
	}
	// get the area of mapping
	pObjNew = abci.Abc_NodeFromIfRec(p.pNtk, p.pIfMan, ifpkg.If_Regular(pDriver), p.vCover)
	pObjNew.Data = hop.Hop_NotCond(pObjNew.Data.(*hop.Hop_Obj_t), ifpkg.If_IsComplement(pDriver))
	// perform replacement
	abc.Abc_NtkUpdate(p.pObj, pObjNew, p.vLevels)
	//printf( "%3d : %d-%d=%d(%d) \n", p->nChanges, nNodesBef, Abc_NtkNodeNum(p->pNtk), nNodesBef-Abc_NtkNodeNum(p->pNtk), nGain );
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Performs resynthesis for one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_ResynthesizeNode(p *Lpk_Man_t) int32 {
	//    static int Count = 0;
	var pDsdNtk *Kit_DsdNtk_t
	var pCut *Lpk_Cut_t
	var pTruth []uint32
	var i, k, nSuppSize, nCutNodes, RetValue int32
	var clk abctime

	// compute the cuts
	clk = abctime(util.Abc_Clock())
	if Lpk_NodeCuts(p) == 0 {
		p.timeCuts += abctime(util.Abc_Clock()) - clk
		return 0
	}
	p.timeCuts += abctime(util.Abc_Clock()) - clk

	//return 0;

	if p.pPars.fVeryVerbose != 0 {
		fmt.Printf("Node %5d : Mffc size = %5d. Cuts = %5d.\n", p.pObj.ID, p.nMffc, p.nEvals)
	}
	// try the good cuts
	p.nCutsTotal += p.nCuts
	p.nCutsUseful += p.nEvals
	for i = 0; i < p.nEvals; i++ {
		// get the cut
		pCut = &p.pCuts[p.pEvals[i]]
		if p.pPars.fFirst != 0 && i == 1 {
			break
		}

		// skip bad cuts
		//        printf( "Mffc size = %d.  ", Abc_NodeMffcLabel(p->pObj) );
		leaves := make([]*Abc_Obj_t, pCut.nLeaves)
		for k = 0; k < int32(pCut.nLeaves); k++ {
			leaves[k] = abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[k])
		}
		nCutNodes = abc.Abc_NodeMffcLabelWithLeaves(p.pObj, leaves, nil)
		//        printf( "Mffc cut = %d.  ", pCut->nNodes - pCut->nNodesDup );
		//        printf( "\n" );
		if nCutNodes != int32(pCut.nNodes)-int32(pCut.nNodesDup) {
			continue
		}

		// compute the truth table
		clk = abctime(util.Abc_Clock())
		pTruth = Lpk_CutTruth(p, pCut, 0)
		nSuppSize = kit.Kit_TruthSupportSize(pTruth, int32(pCut.nLeaves))
		p.timeTruth += abctime(util.Abc_Clock()) - clk

		pDsdNtk = kit.Kit_DsdDecompose(pTruth, int32(pCut.nLeaves))
		//        Kit_DsdVerify( pDsdNtk, pTruth, pCut->nLeaves );
		// skip 16-input non-DSD because ISOP will not work
		if kit.Kit_DsdObjFanNum(kit.Kit_DsdNtkRoot(pDsdNtk)) == 16 {
			kit.Kit_DsdNtkFree(pDsdNtk)
			continue
		}

		// if DSD has nodes that require splitting to fit them into LUTs
		// we can skip those cuts that cannot lead to improvement
		// (a full DSD network requires  V = Nmin * (K-1) + 1 for improvement)
		if kit.Kit_DsdNonDsdSizeMax(pDsdNtk) > p.pPars.nLutSize &&
			nSuppSize >= (int32(pCut.nNodes)-int32(pCut.nNodesDup)-1)*(p.pPars.nLutSize-1)+1 {
			kit.Kit_DsdNtkFree(pDsdNtk)
			continue
		}

		if p.pPars.fVeryVerbose != 0 {
			//            char * pFileName;
			fmt.Printf("  C%02d: L= %2d/%2d  V= %2d/%d  N= %d  W= %4.2f  ", i, pCut.nLeaves, nSuppSize, pCut.nNodes, pCut.nNodesDup, pCut.nLuts, pCut.Weight)
			kit.Kit_DsdPrint(os.Stdout, pDsdNtk)
			kit.Kit_DsdPrintFromTruth(pTruth, int32(pCut.nLeaves))
			//            pFileName = Kit_TruthDumpToFile( pTruth, pCut->nLeaves, Count++ );
			//            printf( "Saved truth table in file \"%s\".\n", pFileName );
		}

		// update the network
		clk = abctime(util.Abc_Clock())
		RetValue = Lpk_ExploreCut(p, pCut, pDsdNtk)
		p.timeEval += abctime(util.Abc_Clock()) - clk
		kit.Kit_DsdNtkFree(pDsdNtk)
		if RetValue != 0 {
			break
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Computes supports of the cofactors of the function.]

	Description [This procedure should be called after Lpk_CutTruth(p,pCut,0)]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_ComputeSupports(p *Lpk_Man_t, pCut *Lpk_Cut_t, pTruth []uint32) {
	var pTruthInv []uint32
	var RetValue1, RetValue2 int32
	pTruthInv = Lpk_CutTruth(p, pCut, 1)
	RetValue1 = kit.Kit_CreateCloudFromTruth(kit.Kit_DsdCloud(p.pDsdMan), pTruth, int32(pCut.nLeaves), vec.Vec_IntArrayP(p.vBddDir))
	RetValue2 = kit.Kit_CreateCloudFromTruth(kit.Kit_DsdCloud(p.pDsdMan), pTruthInv, int32(pCut.nLeaves), vec.Vec_IntArrayP(p.vBddInv))
	if RetValue1 != 0 && (RetValue2 != 0) && vec.Vec_IntSize(p.vBddDir) > 1 && vec.Vec_IntSize(p.vBddInv) > 1 {
		var scratch []uint32
		kit.Kit_TruthCofSupports(vec.Vec_IntArray(p.vBddDir), vec.Vec_IntArray(p.vBddInv), int32(pCut.nLeaves), &scratch, p.puSupps[:])
	} else {
		p.puSupps[0], p.puSupps[1] = 0, 0
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs resynthesis for one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_ResynthesizeNodeNew(p *Lpk_Man_t) int32 {
	//    static int Count = 0;
	var NodeCounts [16]int32 = [16]int32{0, 0, 0, 0, 1, 3, 6, 14, 26, 57, 106, 230, 425, 1000000, 1000000, 1000000}
	var pObjNew, pLeaf *Abc_Obj_t
	var pCut *Lpk_Cut_t
	var pTruth []uint32
	var nNodesBef, nNodesAft, nCutNodes int32
	var i, k int32
	var clk abctime
	var Required int32 = abc.Abc_ObjRequiredLevel(p.pObj)
	//    CloudNode * pFun2;//, * pFun1;

	// compute the cuts
	clk = abctime(util.Abc_Clock())
	if Lpk_NodeCuts(p) == 0 {
		p.timeCuts += abctime(util.Abc_Clock()) - clk
		return 0
	}
	p.timeCuts += abctime(util.Abc_Clock()) - clk

	if p.pPars.fVeryVerbose != 0 {
		fmt.Printf("Node %5d : Mffc size = %5d. Cuts = %5d.  Level = %2d. Req = %2d.\n", p.pObj.ID, p.nMffc, p.nEvals, p.pObj.Level, Required)
	}
	// try the good cuts
	p.nCutsTotal += p.nCuts
	p.nCutsUseful += p.nEvals
	for i = 0; i < p.nEvals; i++ {
		// get the cut
		pCut = &p.pCuts[p.pEvals[i]]
		if p.pPars.fFirst != 0 && i == 1 {
			break
		}
		//        if ( pCut->Weight < 1.05 )
		//            continue;

		// skip bad cuts
		//        printf( "Mffc size = %d.  ", Abc_NodeMffcLabel(p->pObj) );
		leaves := make([]*Abc_Obj_t, pCut.nLeaves)
		for k = 0; k < int32(pCut.nLeaves); k++ {
			leaves[k] = abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[k])
		}
		nCutNodes = abc.Abc_NodeMffcLabelWithLeaves(p.pObj, leaves, nil)
		//        printf( "Mffc cut = %d.  ", pCut->nNodes - pCut->nNodesDup );
		//        printf( "\n" );
		if nCutNodes != int32(pCut.nNodes)-int32(pCut.nNodesDup) {
			continue
		}

		// collect nodes into the array
		vec.Vec_PtrClear(p.vLeaves)
		for k = 0; k < int32(pCut.nLeaves); k++ {
			vec.Vec_PtrPush(p.vLeaves, abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[k]))
		}

		// compute the truth table
		clk = abctime(util.Abc_Clock())
		pTruth = Lpk_CutTruth(p, pCut, 0)
		p.timeTruth += abctime(util.Abc_Clock()) - clk
		clk = abctime(util.Abc_Clock())
		Lpk_ComputeSupports(p, pCut, pTruth)
		p.timeSupps += abctime(util.Abc_Clock()) - clk
		//clk = Abc_Clock();
		//        pFun1 = Lpk_CutTruthBdd( p, pCut );
		//p->timeTruth2 += Abc_Clock() - clk;
		/*
		   clk = abctime(util.Abc_Clock());
		           kit.Cloud_Restart( p->pDsdMan->dd );
		           pFun2 = kit.Kit_TruthToCloud( p->pDsdMan->dd, pTruth, pCut->nLeaves );
		           RetValue = kit.Kit_CreateCloud( p->pDsdMan->dd, pFun2, p->vBddNodes );
		   p->timeTruth3 += abctime(util.Abc_Clock()) - clk;
		*/
		//        if ( pFun1 != pFun2 )
		//            printf( "Truth tables do not agree!\n" );
		//        else
		//            printf( "Fine!\n" );

		if p.pPars.fVeryVerbose != 0 {
			//            char * pFileName;
			var nSuppSize int32 = kit.Kit_TruthSupportSize(pTruth, int32(pCut.nLeaves))
			fmt.Printf("  C%02d: L= %2d/%2d  V= %2d/%d  N= %d  W= %4.2f  ", i, pCut.nLeaves, nSuppSize, pCut.nNodes, pCut.nNodesDup, pCut.nLuts, pCut.Weight)
			for k = 0; k < int32(vec.Vec_PtrSize(p.vLeaves)); k++ {
				pLeaf = vec.Vec_PtrEntry(p.vLeaves, int(k)).(*abc.Object)
				fmt.Printf("%c=%d ", 'a'+k, abc.Abc_ObjLevel(pLeaf))
			}
			fmt.Printf("\n")
			kit.Kit_DsdPrintFromTruth(pTruth, int32(pCut.nLeaves))
			fmt.Printf("\n")
			//            pFileName = Kit_TruthDumpToFile( pTruth, pCut->nLeaves, Count++ );
			//            printf( "Saved truth table in file \"%s\".\n", pFileName );
		}

		// update the network
		nNodesBef = abc.Abc_NtkNodeNum(p.pNtk)
		clk = abctime(util.Abc_Clock())
		pObjNew = Lpk_Decompose(p, p.pNtk, p.vLeaves, pTruth, p.puSupps[:], p.pPars.nLutSize,
			int32(pCut.nNodes)-int32(pCut.nNodesDup)-1+lpkBoolInt(p.pPars.fZeroCost > 0), Required)
		if pObjNew == nil && p.pPars.nLutSize == 4 && int32(pCut.nNodes) > NodeCounts[vec.Vec_PtrSize(p.vLeaves)]+lpkBoolInt(p.pPars.fZeroCost == 0) {
			pObjNew = Abc_NtkLutMinDecompose(p.pNtk, p.vLeaves, pTruth, p.pPars.nLutSize, Required)
		}
		p.timeEval += abctime(util.Abc_Clock()) - clk
		nNodesAft = abc.Abc_NtkNodeNum(p.pNtk)

		// perform replacement
		if pObjNew != nil {
			var nGain int32 = int32(pCut.nNodes) - int32(pCut.nNodesDup) - (nNodesAft - nNodesBef)
			//assert( nGain >= 1 - p->pPars->fZeroCost );
			assert(abc.Abc_ObjLevel(pObjNew) <= Required, "abc.Abc_ObjLevel(pObjNew) <= Required")
			/*
			   if ( nGain <= 0 )
			   {
			       int x = 0;
			   }
			   if ( abc.Abc_ObjLevel(pObjNew) > Required )
			   {
			       int x = 0;
			   }
			*/
			p.nGainTotal += nGain
			p.nChanges++
			if p.pPars.fVeryVerbose != 0 {
				fmt.Printf("Performed resynthesis: Gain = %2d. Level = %2d. Req = %2d.\n", nGain, abc.Abc_ObjLevel(pObjNew), Required)
			}
			abc.Abc_NtkUpdate(p.pObj, pObjNew, p.vLevels)
			//printf( "%3d : %d-%d=%d(%d) \n", p->nChanges, nNodesBef, Abc_NtkNodeNum(p->pNtk), nNodesBef-Abc_NtkNodeNum(p->pNtk), nGain );
			break
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Performs resynthesis for one network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_Resynthesize(pNtk *Abc_Ntk_t, pPars *Lpk_Par_t) int32 {
	var pProgress *extra.ProgressBar = nil // Suppress "might be used uninitialized"
	var p *Lpk_Man_t
	var pObj *Abc_Obj_t
	var Delta float64
	//    int * pnFanouts, nObjMax;
	var i, Iter, nNodes, nNodesPrev int32
	var clk abctime = abctime(util.Abc_Clock())
	assert(abc.Abc_NtkIsLogic(pNtk), "abc.Abc_NtkIsLogic(pNtk)")

	// sweep dangling nodes as a preprocessing step
	abc.Abc_NtkSweep(pNtk, 0)

	// get the number of inputs
	if lib := abcmain.Abc_FrameReadLibLut(); lib != nil {
		pPars.nLutSize = lib.(*If_LibLut_t).LutMax
	} else {
		pPars.nLutSize = abc.Abc_NtkGetFaninMax(pNtk)
	}
	if pPars.nLutSize > 6 {
		pPars.nLutSize = 6
	}
	if pPars.nLutSize < 3 {
		pPars.nLutSize = 3
	}
	// adjust the number of crossbars based on LUT size
	if pPars.nVarsShared > pPars.nLutSize-2 {
		pPars.nVarsShared = pPars.nLutSize - 2
	}
	// get the max number of LUTs tried
	pPars.nVarsMax = pPars.nLutsMax*(pPars.nLutSize-1) + 1 // V = N * (K-1) + 1
	for pPars.nVarsMax > 16 {
		pPars.nLutsMax--
		pPars.nVarsMax = pPars.nLutsMax*(pPars.nLutSize-1) + 1

	}
	if pPars.fVerbose != 0 {
		fmt.Printf("Resynthesis for %d %d-LUTs with %d non-MFFC LUTs, %d crossbars, and %d-input cuts.\n", pPars.nLutsMax, pPars.nLutSize, pPars.nLutsOver, pPars.nVarsShared, pPars.nVarsMax)
	}

	// convert into the AIG
	if abc.Abc_NtkToAig(pNtk) == 0 {
		fmt.Fprintf(os.Stdout, "Converting to BDD has failed.\n")
		return 0
	}
	assert(abc.Abc_NtkHasAig(pNtk), "abc.Abc_NtkHasAig(pNtk)")

	// set the number of levels
	abc.Abc_NtkLevel(pNtk)
	abc.Abc_NtkStartReverseLevels(pNtk, pPars.nGrowthLevel)

	// start the manager
	p = Lpk_ManStart(pPars)
	p.pNtk = pNtk
	p.nNodesTotal = abc.Abc_NtkNodeNum(pNtk)
	p.vLevels = vec.Vec_VecStart(abc.Abc_NtkReverseLevelMax(pNtk))
	if p.pPars.fSatur != 0 {
		p.vVisited = vec.Vec_VecStart(0)
	}
	if pPars.fVerbose != 0 {
		p.nTotalNets = abc.Abc_NtkGetTotalFanins(pNtk)
		p.nTotalNodes = abc.Abc_NtkNodeNum(pNtk)
	}
	/*
	   // save the number of fanouts of all objects
	   nObjMax = abc.Abc_NtkObjNumMax( pNtk );
	   pnFanouts = ABC_ALLOC( int, nObjMax );
	   memset( pnFanouts, 0, sizeof * nObjMax );
	   abc.Abc_NtkForEachObj( pNtk, pObj, i )
	       pnFanouts[pObj->Id] = abc.Abc_ObjFanoutNum(pObj);
	*/

	// iterate over the network
	nNodesPrev = p.nNodesTotal
	for Iter = 1; ; Iter++ {
		// expand storage for changed nodes
		if p.pPars.fSatur != 0 {
			vec.Vec_VecExpand(p.vVisited, int(abc.Abc_NtkObjNumMax(pNtk)+1))
		}

		// consider all nodes
		nNodes = abc.Abc_NtkObjNumMax(pNtk)
		if !(pPars.fVeryVerbose != 0) {
			pProgress = extra.Extra_ProgressBarStart(os.Stdout, nNodes)
		}
		for i = 0; i < abc.Abc_NtkObjNumMax(pNtk); i++ {
			pObj = abc.Abc_NtkObj(pNtk, i)
			if pObj == nil || abc.Abc_ObjIsNode(pObj) == 0 {
				continue
			}
			// skip all except the final node
			if pPars.fFirst != 0 {
				if abc.Abc_ObjIsCo(abc.Abc_ObjFanout0(pObj)) == 0 {
					continue
				}
			}
			if i >= nNodes {
				break
			}
			if !(pPars.fVeryVerbose != 0) {
				extra.Extra_ProgressBarUpdate(pProgress, i, "")
			}
			// skip the nodes that did not change
			if p.pPars.fSatur != 0 && !(Lpk_NodeHasChanged(p, int32(pObj.ID)) != 0) {
				continue
			}
			// resynthesize
			p.pObj = pObj
			if p.pPars.fOldAlgo != 0 {
				Lpk_ResynthesizeNode(p)
			} else {
				Lpk_ResynthesizeNodeNew(p)
			}
		}
		if !(pPars.fVeryVerbose != 0) {
			extra.Extra_ProgressBarStop(pProgress)
		}

		// check the increase
		Delta = 100.00 * float64(nNodesPrev-abc.Abc_NtkNodeNum(pNtk)) / float64(p.nNodesTotal)
		if Delta < 0.05 {
			break
		}
		nNodesPrev = abc.Abc_NtkNodeNum(pNtk)
		if !(p.pPars.fSatur != 0) {
			break
		}

		if pPars.fFirst != 0 {
			break
		}
	}
	abc.Abc_NtkStopReverseLevels(pNtk)
	/*
	   // report the fanout changes
	   abc.Abc_NtkForEachObj( pNtk, pObj, i )
	   {
	       if ( i >= nObjMax )
	           continue;
	       if ( abc.Abc_ObjFanoutNum(pObj) - pnFanouts[pObj->Id] == 0 )
	           continue;
	       printf( "%d ", abc.Abc_ObjFanoutNum(pObj) - pnFanouts[pObj->Id] );
	   }
	   printf( "\n" );
	*/

	if pPars.fVerbose != 0 {
		//        Cloud_PrintInfo( p->pDsdMan->dd );
		p.nTotalNets2 = abc.Abc_NtkGetTotalFanins(pNtk)
		p.nTotalNodes2 = abc.Abc_NtkNodeNum(pNtk)
		fmt.Printf("Node gain = %5d. (%.2f %%)  ", p.nTotalNodes-p.nTotalNodes2, 100.0*float64(p.nTotalNodes-p.nTotalNodes2)/float64(p.nTotalNodes))
		fmt.Printf("Edge gain = %5d. (%.2f %%)  ", p.nTotalNets-p.nTotalNets2, 100.0*float64(p.nTotalNets-p.nTotalNets2)/float64(p.nTotalNets))
		fmt.Printf("Muxes = %4d. Dsds = %4d.", p.nMuxes, p.nDsds)
		fmt.Printf("\n")
		fmt.Printf("Nodes = %5d (%3d)  Cuts = %5d (%4d)  Changes = %5d  Iter = %2d  Benefit = %d.\n", p.nNodesTotal, p.nNodesOver, p.nCutsTotal, p.nCutsUseful, p.nChanges, Iter, p.nBenefited)

		fmt.Printf("Non-DSD:")
		for i = 3; i <= pPars.nVarsMax; i++ {
			if p.nBlocks[i] != 0 {
				fmt.Printf(" %d=%d", i, p.nBlocks[i])
			}
		}
		fmt.Printf("\n")

		p.timeTotal = abctime(util.Abc_Clock()) - clk
		p.timeEval = p.timeEval - p.timeMap
		p.timeOther = p.timeTotal - p.timeCuts - p.timeTruth - p.timeEval - p.timeMap
		lpkPrintTimePercent("Cuts  ", p.timeCuts, p.timeTotal)
		lpkPrintTimePercent("Truth ", p.timeTruth, p.timeTotal)
		lpkPrintTimePercent("CSupps", p.timeSupps, p.timeTotal)
		lpkPrintTimePercent("Eval  ", p.timeEval, p.timeTotal)
		lpkPrintTimePercent(" MuxAn", p.timeEvalMuxAn, p.timeEval)
		lpkPrintTimePercent(" MuxSp", p.timeEvalMuxSp, p.timeEval)
		lpkPrintTimePercent(" DsdAn", p.timeEvalDsdAn, p.timeEval)
		lpkPrintTimePercent(" DsdSp", p.timeEvalDsdSp, p.timeEval)
		lpkPrintTimePercent(" Other", p.timeEval-p.timeEvalMuxAn-p.timeEvalMuxSp-p.timeEvalDsdAn-p.timeEvalDsdSp, p.timeEval)
		lpkPrintTimePercent("Map   ", p.timeMap, p.timeTotal)
		lpkPrintTimePercent("Other ", p.timeOther, p.timeTotal)
		lpkPrintTimePercent("TOTAL ", p.timeTotal, p.timeTotal)
	}

	Lpk_ManStop(p)
	// check the resulting network
	if abc.Abc_NtkCheck(pNtk) == 0 {
		fmt.Printf("Lpk_Resynthesize: The network check has failed.\n")
		return 0
	}
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

func lpkBoolInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
func lpkPrintTimePercent(label string, elapsed, total abctime) {
	util.Abc_PrintTimeP(1, label, int64(elapsed), int64(total))
}
