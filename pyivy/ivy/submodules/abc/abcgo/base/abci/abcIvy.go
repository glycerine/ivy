package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	ivypkg "github.com/glycerine/abcgo/aig/ivy"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/proof/fra"
)

/**CFile****************************************************************

  FileName    [abcIvy.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Strashing of the current network.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcIvy.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "bool/dec/dec.h"
//#include "proof/fra/fra.h"
//#include "aig/ivy/ivy.h"
//#include "proof/fraig/fraig.h"
//#include "map/mio/mio.h"
//#include "aig/aig/aig.h"

// The optional CUDD include is unnecessary; BDD routines are provided by base/abc.

//ABC_NAMESPACE_IMPL_START

// extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );

// extern void   aig.Aig_ManStop( Aig_Man_t * pMan );

//extern int fra.Fra_FraigSat( Aig_Man_t * pMan, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, int fFlipBits, int fAndOuts, int fNewSolver, int fVerbose );
// extern ivypkg.Ivy_Obj_t * Dec_GraphToNetworkIvy( ivypkg.Ivy_Man_t * pMan, Dec_Graph_t * pGraph );

// extern void ivypkg.Ivy_CutComputeAll( ivypkg.Ivy_Man_t * p, int nInputs );

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Abc_Ntk_t *  Abc_NtkFromIvy( Abc_Ntk_t * pNtkOld, ivypkg.Ivy_Man_t * pMan );

// static Abc_Ntk_t *  Abc_NtkFromIvySeq( Abc_Ntk_t * pNtkOld, ivypkg.Ivy_Man_t * pMan, int fHaig );

// static ivypkg.Ivy_Man_t *  Abc_NtkToIvy( Abc_Ntk_t * pNtkOld );

// static void         Abc_NtkStrashPerformAig( Abc_Ntk_t * pNtk, ivypkg.Ivy_Man_t * pMan );

// static ivypkg.Ivy_Obj_t *  Abc_NodeStrashAig( ivypkg.Ivy_Man_t * pMan, Abc_Obj_t * pNode );

// static ivypkg.Ivy_Obj_t *  Abc_NodeStrashAigSopAig( ivypkg.Ivy_Man_t * pMan, Abc_Obj_t * pNode, char * pSop );

// static ivypkg.Ivy_Obj_t *  Abc_NodeStrashAigExorAig( ivypkg.Ivy_Man_t * pMan, Abc_Obj_t * pNode, char * pSop );

// static ivypkg.Ivy_Obj_t *  Abc_NodeStrashAigFactorAig( ivypkg.Ivy_Man_t * pMan, Abc_Obj_t * pNode, char * pSop );

type Abc_Edge_t int32

func Abc_EdgeCreate(Id int32, fCompl int32) Abc_Edge_t        { return (Id << 1) | fCompl }
func Abc_EdgeId(Edge Abc_Edge_t) int32                        { return Edge >> 1 }
func Abc_EdgeIsComplement(Edge Abc_Edge_t) int32              { return Edge & 1 }
func Abc_EdgeRegular(Edge Abc_Edge_t) Abc_Edge_t              { return (Edge >> 1) << 1 }
func Abc_EdgeNot(Edge Abc_Edge_t) Abc_Edge_t                  { return Edge ^ 1 }
func Abc_EdgeNotCond(Edge Abc_Edge_t, fCond int32) Abc_Edge_t { return Edge ^ fCond }
func Abc_EdgeFromNode(pNode *Abc_Obj_t) Abc_Edge_t {
	return Abc_EdgeCreate(baseabc.Abc_ObjRegular(pNode).Id, baseabc.Abc_ObjIsComplement(pNode))
}
func Abc_EdgeToNode(p *Abc_Ntk_t, Edge Abc_Edge_t) baseabc.Edge {
	return baseabc.Abc_ObjNotCond(baseabc.Abc_NtkObj(p, Abc_EdgeId(Edge)), Abc_EdgeIsComplement(Edge))
}

func Abc_ObjFanin0Ivy(p *Abc_Ntk_t, pObj *ivypkg.Ivy_Obj_t) baseabc.Edge {
	return baseabc.Abc_ObjNotCond(Abc_EdgeToNode(p, ivypkg.Ivy_ObjFanin0(pObj).TravId), ivypkg.Ivy_ObjFaninC0(pObj))
}
func Abc_ObjFanin1Ivy(p *Abc_Ntk_t, pObj *ivypkg.Ivy_Obj_t) baseabc.Edge {
	return baseabc.Abc_ObjNotCond(Abc_EdgeToNode(p, ivypkg.Ivy_ObjFanin1(pObj).TravId), ivypkg.Ivy_ObjFaninC1(pObj))
}

// static Vec_Int_t * Abc_NtkCollectLatchValuesIvy( Abc_Ntk_t * pNtk, int fUseDcs );

var timeRetime int32

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Prepares the IVY package.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvyBefore(pNtk *Abc_Ntk_t, fSeq int32, fUseDc int32) *ivypkg.Ivy_Man_t {
	var pMan *ivypkg.Ivy_Man_t
	//timeRetime = Abc_Clock();
	assert(baseabc.Abc_NtkIsNetlist(pNtk) == 0, "!baseabc.Abc_NtkIsNetlist(pNtk)")
	if baseabc.Abc_NtkIsBddLogic(pNtk) != 0 {
		if baseabc.Abc_NtkBddToSop(pNtk, -1, util.ABC_INFINITY) == 0 {
			fmt.Printf("Abc_NtkIvyBefore(): Converting to SOPs has failed.\n")
			return nil
		}
	}
	if fSeq != 0 && baseabc.Abc_NtkCountSelfFeedLatches(pNtk) != 0 {
		fmt.Printf("Warning: The network has %d self-feeding latches.\n", baseabc.Abc_NtkCountSelfFeedLatches(pNtk))
		//        return NULL;
	}
	// print warning about choice nodes
	if baseabc.Abc_NtkGetChoiceNum(pNtk) != 0 {
		fmt.Printf("Warning: The choice nodes in the initial AIG are removed by strashing.\n")
	}
	// convert to the AIG manager
	pMan = Abc_NtkToIvy(pNtk)
	if ivypkg.Ivy_ManCheck(pMan) == 0 {
		fmt.Printf("AIG check has failed.\n")
		ivypkg.Ivy_ManStop(pMan)
		return nil
	}
	//    ivypkg.Ivy_ManPrintStats( pMan );
	if fSeq != 0 {
		var nLatches int32 = baseabc.Abc_NtkLatchNum(pNtk)
		var vInit *Vec_Int_t = Abc_NtkCollectLatchValuesIvy(pNtk, fUseDc)
		ivypkg.Ivy_ManMakeSeq(pMan, nLatches, vInit.pArray)
		Vec_IntFree(vInit)
		//        ivypkg.Ivy_ManPrintStats( pMan );
	}
	//timeRetime = Abc_Clock() - timeRetime;
	return pMan
}

/*
*Function*************************************************************

	Synopsis    [Prepares the IVY package.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvyAfter(pNtk *Abc_Ntk_t, pMan *ivypkg.Ivy_Man_t, fSeq int32, fHaig int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var nNodes int32
	var fCleanup int32 = 1
	// convert from the AIG manager
	if fSeq != 0 {
		pNtkAig = Abc_NtkFromIvySeq(pNtk, pMan, fHaig)
	} else {
		pNtkAig = Abc_NtkFromIvy(pNtk, pMan)
	}
	// report the cleanup results
	if fHaig == 0 && fCleanup != 0 && pNtkAig != nil {
		nNodes = baseabc.Abc_AigCleanup(pNtkAig.FunctionManager.(*baseabc.Abc_Aig_t))
		if nNodes != 0 {
			fmt.Printf("Warning: AIG cleanup removed %d nodes (this is not a bug).\n", nNodes)
		}
	}
	// duplicate EXDC
	if pNtk.Exdc != nil {
		pNtkAig.Exdc = baseabc.Abc_NtkDup(pNtk.Exdc)
	}
	// make sure everything is okay
	if baseabc.Abc_NtkCheck(pNtkAig) == 0 {
		fmt.Printf("Abc_NtkStrash: The network check has failed.\n")
		Abc_NtkDelete(pNtkAig)
		return nil
	}
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvyStrash(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan *ivypkg.Ivy_Man_t
	pMan = Abc_NtkIvyBefore(pNtk, 1, 0)
	if pMan == nil {
		return nil
	}
	pNtkAig = Abc_NtkIvyAfter(pNtk, pMan, 1, 0)
	ivypkg.Ivy_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvyHaig(pNtk *Abc_Ntk_t, nIters int32, fUseZeroCost int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan *ivypkg.Ivy_Man_t
	var clk abctime
	//    int i;
	/*
	   extern int nMoves;
	   extern int nMovesS;
	   extern int nClauses;
	   extern int timeInv;

	   nMoves = 0;
	   nMovesS = 0;
	   nClauses = 0;
	   timeInv = 0;
	*/
	pMan = Abc_NtkIvyBefore(pNtk, 1, 1)
	if pMan == nil {
		return nil
	}
	//timeRetime = Abc_Clock();

	ivypkg.Ivy_ManHaigStart(pMan, fVerbose)
	//    ivypkg.Ivy_ManRewriteSeq( pMan, 0, 0 );
	//    for ( i = 0; i < nIters; i++ )
	//        ivypkg.Ivy_ManRewriteSeq( pMan, fUseZeroCost, 0 );

	//printf( "%d ", ivypkg.Ivy_ManNodeNum(pMan) );
	ivypkg.Ivy_ManRewriteSeq(pMan, 0, 0)
	ivypkg.Ivy_ManRewriteSeq(pMan, 0, 0)
	ivypkg.Ivy_ManRewriteSeq(pMan, 1, 0)
	//printf( "%d ", ivypkg.Ivy_ManNodeNum(pMan) );
	//printf( "%d ", ivypkg.Ivy_ManNodeNum(pMan->pHaig) );
	//ABC_PRT( " ", Abc_Clock() - clk );
	//printf( "\n" );
	/*
	   printf( "Moves = %d.  ", nMoves );
	   printf( "MovesS = %d.  ", nMovesS );
	   printf( "Clauses = %d.  ", nClauses );
	   ABC_PRT( "Time", timeInv );
	*/
	//    ivypkg.Ivy_ManRewriteSeq( pMan, 1, 0 );
	//printf( "Haig size = %d.\n", ivypkg.Ivy_ManNodeNum(pMan->pHaig) );
	//    ivypkg.Ivy_ManHaigPostprocess( pMan, fVerbose );
	//timeRetime = Abc_Clock() - timeRetime;

	// write working AIG into the current network
	//    pNtkAig = Abc_NtkIvyAfter( pNtk, pMan, 1, 0 );
	// write HAIG into the current network
	pNtkAig = Abc_NtkIvyAfter(pNtk, ivypkg.Ivy_ManReadHaig(pMan), 1, 1)

	ivypkg.Ivy_ManHaigStop(pMan)
	ivypkg.Ivy_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvyCuts(pNtk *Abc_Ntk_t, nInputs int32) {
	var pMan *ivypkg.Ivy_Man_t
	pMan = Abc_NtkIvyBefore(pNtk, 1, 0)
	if pMan == nil {
		return
	}
	ivypkg.Ivy_CutComputeAll(pMan, nInputs)
	ivypkg.Ivy_ManStop(pMan)
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvyRewrite(pNtk *Abc_Ntk_t, fUpdateLevel int32, fUseZeroCost int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan *ivypkg.Ivy_Man_t
	pMan = Abc_NtkIvyBefore(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	//timeRetime = Abc_Clock();
	ivypkg.Ivy_ManRewritePre(pMan, fUpdateLevel, fUseZeroCost, fVerbose)
	//timeRetime = Abc_Clock() - timeRetime;
	pNtkAig = Abc_NtkIvyAfter(pNtk, pMan, 0, 0)
	ivypkg.Ivy_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvyRewriteSeq(pNtk *Abc_Ntk_t, fUseZeroCost int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan *ivypkg.Ivy_Man_t
	pMan = Abc_NtkIvyBefore(pNtk, 1, 1)
	if pMan == nil {
		return nil
	}
	//timeRetime = Abc_Clock();
	ivypkg.Ivy_ManRewriteSeq(pMan, fUseZeroCost, fVerbose)
	//timeRetime = Abc_Clock() - timeRetime;
	//    ivypkg.Ivy_ManRewriteSeq( pMan, 1, 0 );
	//    ivypkg.Ivy_ManRewriteSeq( pMan, 1, 0 );
	pNtkAig = Abc_NtkIvyAfter(pNtk, pMan, 1, 0)
	ivypkg.Ivy_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvyResyn0(pNtk *Abc_Ntk_t, fUpdateLevel int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *ivypkg.Ivy_Man_t
	pMan = Abc_NtkIvyBefore(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	pMan = ivypkg.Ivy_ManResyn0(pTemp, fUpdateLevel, fVerbose)
	ivypkg.Ivy_ManStop(pTemp)
	pNtkAig = Abc_NtkIvyAfter(pNtk, pMan, 0, 0)
	ivypkg.Ivy_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvyResyn(pNtk *Abc_Ntk_t, fUpdateLevel int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *ivypkg.Ivy_Man_t
	pMan = Abc_NtkIvyBefore(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	pMan = ivypkg.Ivy_ManResyn(pTemp, fUpdateLevel, fVerbose)
	ivypkg.Ivy_ManStop(pTemp)
	pNtkAig = Abc_NtkIvyAfter(pNtk, pMan, 0, 0)
	ivypkg.Ivy_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvySat(pNtk *Abc_Ntk_t, nConfLimit int32, fVerbose int32) *Abc_Ntk_t {
	var Params ivypkg.Ivy_FraigParams_t
	var pParams *ivypkg.Ivy_FraigParams_t = &Params
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *ivypkg.Ivy_Man_t
	pMan = Abc_NtkIvyBefore(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	ivypkg.Ivy_FraigParamsDefault(pParams)
	pParams.NBTLimitMiter = nConfLimit
	pParams.FVerbose = fVerbose
	//    pTemp = pMan
	pTemp = pMan
	pMan = ivypkg.Ivy_FraigMiter(pTemp, pParams)
	ivypkg.Ivy_ManStop(pTemp)
	pNtkAig = Abc_NtkIvyAfter(pNtk, pMan, 0, 0)
	ivypkg.Ivy_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Sets the final nodes to point to the original nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkTransferPointers(pNtk *Abc_Ntk_t, pNtkAig *Abc_Ntk_t) {
	baseabc.Abc_AigConst1(pNtk).Copy = baseabc.Abc_AigConst1(pNtkAig)
	for i, obj := range pNtk.Inputs() {
		obj.Copy = pNtkAig.Inputs()[i]
	}
	for i, obj := range pNtk.Outputs() {
		obj.Copy = pNtkAig.Outputs()[i]
	}
	for i, obj := range pNtk.Boxes() {
		obj.Copy = pNtkAig.Boxes()[i]
	}
	for _, obj := range pNtk.ObjectsOfKind(baseabc.ObjectNode) {
		ivyNode, _ := obj.Copy.(*ivypkg.Ivy_Obj_t)
		if ivyNode == nil {
			continue
		}
		equiv := ivypkg.Ivy_ObjEquiv(ivyNode)
		if equiv == nil {
			continue
		}
		mapped := Abc_EdgeToNode(pNtkAig, ivypkg.Ivy_Regular(equiv).TravId)
		obj.Copy = mapped.Object
		obj.CopyEdge = baseabc.Edge{Object: mapped.Object, Complement: mapped.Complement != (ivypkg.Ivy_IsComplement(equiv) != 0)}
	}
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvyFraig(pNtk *Abc_Ntk_t, nConfLimit int32, fDoSparse int32, fProve int32, fTransfer int32, fVerbose int32) *Abc_Ntk_t {
	var Params ivypkg.Ivy_FraigParams_t
	var pParams *ivypkg.Ivy_FraigParams_t = &Params
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *ivypkg.Ivy_Man_t
	pMan = Abc_NtkIvyBefore(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	ivypkg.Ivy_FraigParamsDefault(pParams)
	pParams.NBTLimitNode = nConfLimit
	pParams.FVerbose = fVerbose
	pParams.FProve = fProve
	pParams.FDoSparse = fDoSparse
	pTemp = pMan
	pMan = ivypkg.Ivy_FraigPerform(pTemp, pParams)
	// transfer the pointers
	if fTransfer == 1 {
		var vCopies *Vec_Ptr_t
		vCopies = baseabc.Abc_NtkSaveCopy(pNtk)
		pNtkAig = Abc_NtkIvyAfter(pNtk, pMan, 0, 0)
		baseabc.Abc_NtkLoadCopy(pNtk, vCopies)
		Vec_PtrFree(vCopies)
		Abc_NtkTransferPointers(pNtk, pNtkAig)

	} else {
		pNtkAig = Abc_NtkIvyAfter(pNtk, pMan, 0, 0)
	}
	model, _ := ivypkg.Ivy_ManTakeData(pMan).([]int32)
	baseabc.Abc_NtkSetModel(pNtkAig, model)
	ivypkg.Ivy_ManStop(pTemp)
	ivypkg.Ivy_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvyProve(ppNtk **Abc_Ntk_t, pPars any) int32 {
	pParams := pPars.(*Prove_Params_t)
	var pNtk *Abc_Ntk_t = *ppNtk
	var pNtkTemp *Abc_Ntk_t
	var pObj, pFanin *Abc_Obj_t
	var pMan *ivypkg.Ivy_Man_t
	var pMan2 *Aig_Man_t
	var RetValue int32
	assert(baseabc.Abc_NtkIsStrash(pNtk) != 0 || baseabc.Abc_NtkIsLogic(pNtk) != 0, "Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk)")
	// experiment with various parameters settings
	//    pParams->fUseBdds = 1;
	//    pParams->fBddReorder = 1;
	//    pParams->nTotalBacktrackLimit = 10000;

	// strash the network if it is not strashed already
	if baseabc.Abc_NtkIsStrash(pNtk) == 0 {
		pNtkTemp = pNtk
		pNtk = Abc_NtkStrash(pNtkTemp, 0, 1, 0)
		Abc_NtkDelete(pNtkTemp)
	}

	// check the case when the 0000 simulation pattern detect the bug
	pObj = baseabc.Abc_NtkPo(pNtk, 0)
	pFanin = baseabc.Abc_ObjFanin0(pObj)
	if pFanin.Phase != (baseabc.Abc_ObjFaninC0(pObj) != 0) {
		baseabc.Abc_NtkSetModel(pNtk, make([]int32, baseabc.Abc_NtkCiNum(pNtk)))
		return 0
	}

	// changed in "src\sat\fraig\fraigMan.c"
	//    pParams->nMiteringLimitStart  = 300;    // starting mitering limit
	// to be
	//    pParams->nMiteringLimitStart  = 5000;    // starting mitering limit

	// if SAT only, solve without iteration
	//    RetValue = Abc_NtkMiterSat( pNtk, 2*(ABC_INT64_T)pParams->nMiteringLimitStart, 0, 0, NULL, NULL );
	pMan2 = Abc_NtkToDar(pNtk, 0, 0)
	RetValue = fra.Fra_FraigSat(pMan2, int64(pParams.nMiteringLimitStart), 0, 0, 0, 0, 1, 0, 0, 0)
	model, _ := aig.Aig_ManReleaseData(pMan2).([]int32)
	baseabc.Abc_NtkSetModel(pNtk, model)
	aig.Aig_ManStop(pMan2)
	//    pNtk->pModel = Aig_ManReleaseData( pMan2 );
	if RetValue >= 0 {
		return RetValue
	}

	// apply AIG rewriting
	if pParams.fUseRewriting != 0 && baseabc.Abc_NtkNodeNum(pNtk) > 500 {
		//        abctime 		//printf( "Before rwsat = %d. ", baseabc.Abc_NtkNodeNum(pNtk) );
		pParams.fUseRewriting = 0
		pNtkTemp = pNtk
		pNtk = Abc_NtkBalance(pNtkTemp, 0, 0, 0)
		Abc_NtkDelete(pNtkTemp)
		Abc_NtkRewrite(pNtk, 0, 0, 0, 0, 0)
		pNtkTemp = pNtk
		pNtk = Abc_NtkBalance(pNtkTemp, 0, 0, 0)
		Abc_NtkDelete(pNtkTemp)
		Abc_NtkRewrite(pNtk, 0, 0, 0, 0, 0)
		Abc_NtkRefactor(pNtk, 10, 16, 0, 0, 0, 0)
		//printf( "After rwsat = %d. ", baseabc.Abc_NtkNodeNum(pNtk) );
		//ABC_PRT( "Time", Abc_Clock() - clk );
	}

	// convert ABC network into IVY network
	pMan = Abc_NtkIvyBefore(pNtk, 0, 0)

	// solve the CEC problem
	ivyParams := &ivypkg.Ivy_ProveParams{
		FUseFraiging: pParams.fUseFraiging, FUseRewriting: pParams.fUseRewriting,
		FUseBdds: pParams.fUseBdds, FVerbose: pParams.fVerbose,
		NItersMax: pParams.nItersMax, NMiteringLimitStart: pParams.nMiteringLimitStart,
		NMiteringLimitMulti: pParams.nMiteringLimitMulti, NRewritingLimitStart: pParams.nRewritingLimitStart,
		NRewritingLimitMulti: pParams.nRewritingLimitMulti, NFraigingLimitStart: pParams.nFraigingLimitStart,
		NFraigingLimitMulti: pParams.nFraigingLimitMulti, NBddSizeLimit: pParams.nBddSizeLimit,
		FBddReorder: pParams.fBddReorder, NMiteringLimitLast: pParams.nMiteringLimitLast,
		NTotalBacktrackLimit: pParams.nTotalBacktrackLimit, NTotalInspectLimit: pParams.nTotalInspectLimit,
		NTotalBacktracksMade: pParams.nTotalBacktracksMade, NTotalInspectsMade: pParams.nTotalInspectsMade,
	}
	RetValue = ivypkg.Ivy_FraigProveWithParams(&pMan, ivyParams)
	pParams.fUseRewriting = ivyParams.FUseRewriting
	pParams.nTotalBacktracksMade = ivyParams.NTotalBacktracksMade
	pParams.nTotalInspectsMade = ivyParams.NTotalInspectsMade
	//    RetValue = -1;

	// convert IVY network into ABC network
	pNtkTemp = pNtk
	pNtk = Abc_NtkIvyAfter(pNtkTemp, pMan, 0, 0)
	Abc_NtkDelete(pNtkTemp)
	// transfer model if given
	model, _ := ivypkg.Ivy_ManTakeData(pMan).([]int32)
	baseabc.Abc_NtkSetModel(pNtk, model)
	ivypkg.Ivy_ManStop(pMan)

	// try to prove it using brute force SAT with good CNF encoding
	if RetValue < 0 {
		pMan2 = Abc_NtkToDar(pNtk, 0, 0)
		// dump the miter before entering high-effort solving
		if pParams.fVerbose != 0 {
			var pFileName string // was [100]byte in C

			pFileName = fmt.Sprintf("cecmiter.aig")
			_ = ioa.Ioa_WriteAiger(pMan2, pFileName, 0, 0)
			fmt.Printf("Intermediate reduced miter is written into file \"%s\".\n", pFileName)
		}
		RetValue = fra.Fra_FraigSat(pMan2, int64(pParams.nMiteringLimitLast), 0, 0, 0, 0, 0, 0, 0, pParams.fVerbose)
		if model, _ := aig.Aig_ManReleaseData(pMan2).([]int32); model != nil {
			baseabc.Abc_NtkSetModel(pNtk, model)
		}
		aig.Aig_ManStop(pMan2)
	}

	// try to prove it using brute force BDDs
	if RetValue < 0 && pParams.fUseBdds != 0 {
		if pParams.fVerbose != 0 {
			fmt.Printf("Attempting BDDs with node limit %d ...\n", pParams.nBddSizeLimit)
		}
		pNtkTemp = pNtk
		pNtk = Abc_NtkCollapse(pNtkTemp, pParams.nBddSizeLimit, 0, pParams.fBddReorder, 0, 0, 0)
		if pNtk != nil {
			Abc_NtkDelete(pNtkTemp)
			dd := pNtk.FunctionManager.(*core.DdManager)
			RetValue = 0
			if baseabc.Abc_NtkNodeNum(pNtk) == 1 && baseabc.Abc_ObjFanin0(baseabc.Abc_NtkPo(pNtk, 0)).Data == core.Cudd_ReadLogicZero(dd) {
				RetValue = 1
			}
		} else {
			pNtk = pNtkTemp
		}
	}

	// return the result
	*ppNtk = pNtk
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIvy(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	vInit := Abc_NtkCollectLatchValuesIvy(pNtk, 0)
	assert(baseabc.Abc_NtkIsNetlist(pNtk) == 0, "!baseabc.Abc_NtkIsNetlist(pNtk)")
	if baseabc.Abc_NtkIsBddLogic(pNtk) != 0 {
		if baseabc.Abc_NtkBddToSop(pNtk, -1, util.ABC_INFINITY) == 0 {
			Vec_IntFree(vInit)
			fmt.Printf("Abc_NtkIvy(): Converting to SOPs has failed.\n")
			return nil
		}
	}
	if baseabc.Abc_NtkCountSelfFeedLatches(pNtk) != 0 {
		fmt.Printf("Warning: The network has %d self-feeding latches. Quitting.\n", baseabc.Abc_NtkCountSelfFeedLatches(pNtk))
		return nil
	}
	if baseabc.Abc_NtkGetChoiceNum(pNtk) != 0 {
		fmt.Printf("Warning: The choice nodes in the initial AIG are removed by strashing.\n")
	}
	pMan := Abc_NtkToIvy(pNtk)
	if ivypkg.Ivy_ManCheck(pMan) == 0 {
		Vec_IntFree(vInit)
		fmt.Printf("AIG check has failed.\n")
		ivypkg.Ivy_ManStop(pMan)
		return nil
	}
	// C leaves the experimental rewrites and conversion back to ABC commented
	// out. The active routine stops the checked manager and returns NULL.
	ivypkg.Ivy_ManStop(pMan)
	return nil
}

/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
