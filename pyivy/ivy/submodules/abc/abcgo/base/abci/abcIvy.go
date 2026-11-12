package abci

import (
	"fmt"
	"os"
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

#ifdef ABC_USE_CUDD
#include "bdd/extrab/extraBdd.h"
#endif

//ABC_NAMESPACE_IMPL_START

// extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );

// extern void   Aig_ManStop( Aig_Man_t * pMan );

//extern int Fra_FraigSat( Aig_Man_t * pMan, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, int fFlipBits, int fAndOuts, int fNewSolver, int fVerbose );
// extern Ivy_Obj_t * Dec_GraphToNetworkIvy( Ivy_Man_t * pMan, Dec_Graph_t * pGraph );

// extern void Ivy_CutComputeAll( Ivy_Man_t * p, int nInputs );


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Abc_Ntk_t *  Abc_NtkFromIvy( Abc_Ntk_t * pNtkOld, Ivy_Man_t * pMan );

// static Abc_Ntk_t *  Abc_NtkFromIvySeq( Abc_Ntk_t * pNtkOld, Ivy_Man_t * pMan, int fHaig );

// static Ivy_Man_t *  Abc_NtkToIvy( Abc_Ntk_t * pNtkOld );


// static void         Abc_NtkStrashPerformAig( Abc_Ntk_t * pNtk, Ivy_Man_t * pMan );

// static Ivy_Obj_t *  Abc_NodeStrashAig( Ivy_Man_t * pMan, Abc_Obj_t * pNode );

// static Ivy_Obj_t *  Abc_NodeStrashAigSopAig( Ivy_Man_t * pMan, Abc_Obj_t * pNode, char * pSop );

// static Ivy_Obj_t *  Abc_NodeStrashAigExorAig( Ivy_Man_t * pMan, Abc_Obj_t * pNode, char * pSop );

// static Ivy_Obj_t *  Abc_NodeStrashAigFactorAig( Ivy_Man_t * pMan, Abc_Obj_t * pNode, char * pSop );


typedef int   Abc_Edge_t;
func Abc_EdgeCreate(Id int32, fCompl int32) Abc_Edge_t { return (Id << 1) | fCompl;             }
func Abc_EdgeId(Edge Abc_Edge_t) int32 { return Edge >> 1;                      }
func Abc_EdgeIsComplement(Edge Abc_Edge_t) int32 { return Edge & 1;                       }
func Abc_EdgeRegular(Edge Abc_Edge_t) Abc_Edge_t { return (Edge >> 1) << 1;               }
func Abc_EdgeNot(Edge Abc_Edge_t) Abc_Edge_t { return Edge ^ 1;                       }
func Abc_EdgeNotCond(Edge Abc_Edge_t, fCond int32) Abc_Edge_t { return Edge ^ fCond;                   }
func Abc_EdgeFromNode(pNode *Abc_Obj_t) Abc_Edge_t { return Abc_EdgeCreate( Abc_ObjRegular(pNode).Id, Abc_ObjIsComplement(pNode) );       }
func Abc_EdgeToNode(p *Abc_Ntk_t, Edge Abc_Edge_t) *Abc_Obj_t { return Abc_ObjNotCond( Abc_NtkObj(p, Abc_EdgeId(Edge)), Abc_EdgeIsComplement(Edge) ); }

func Abc_ObjFanin0Ivy(p *Abc_Ntk_t, pObj *Ivy_Obj_t) *Abc_Obj_t { return Abc_ObjNotCond( Abc_EdgeToNode(p, Ivy_ObjFanin0(pObj).TravId), Ivy_ObjFaninC0(pObj) ); }
func Abc_ObjFanin1Ivy(p *Abc_Ntk_t, pObj *Ivy_Obj_t) *Abc_Obj_t { return Abc_ObjNotCond( Abc_EdgeToNode(p, Ivy_ObjFanin1(pObj).TravId), Ivy_ObjFaninC1(pObj) ); }

// static Vec_Int_t * Abc_NtkCollectLatchValuesIvy( Abc_Ntk_t * pNtk, int fUseDcs );


var timeRetime int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Prepares the IVY package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvyBefore(pNtk *Abc_Ntk_t, fSeq int32, fUseDc int32) *Ivy_Man_t {
    var pMan *Ivy_Man_t;
//timeRetime = Abc_Clock();
    assert(!Abc_NtkIsNetlist(pNtk), "!Abc_NtkIsNetlist(pNtk)");
    if Abc_NtkIsBddLogic(pNtk) {
        if !Abc_NtkBddToSop(pNtk, -1, ABC_INFINITY) {
            fmt.Printf("Abc_NtkIvyBefore(): Converting to SOPs has failed.\n");
            return nil;
        }
    }
    if (fSeq != 0) && Abc_NtkCountSelfFeedLatches(pNtk) {
        fmt.Printf("Warning: The network has %d self-feeding latches.\n", Abc_NtkCountSelfFeedLatches(pNtk));
//        return NULL;
    }
    // print warning about choice nodes
    if Abc_NtkGetChoiceNum( pNtk ) {
        fmt.Printf("Warning: The choice nodes in the initial AIG are removed by strashing.\n"); }
    // convert to the AIG manager
    pMan = Abc_NtkToIvy( pNtk );
    if !Ivy_ManCheck( pMan ) {
        fmt.Printf("AIG check has failed.\n");
        Ivy_ManStop( pMan );
        return nil;
    }
//    Ivy_ManPrintStats( pMan );
    if fSeq != 0 {
        var nLatches int32  = Abc_NtkLatchNum(pNtk);
        var vInit *Vec_Int_t  = Abc_NtkCollectLatchValuesIvy( pNtk, fUseDc );
        Ivy_ManMakeSeq( pMan, nLatches, vInit.pArray );
        Vec_IntFree( vInit );
//        Ivy_ManPrintStats( pMan );
    }
//timeRetime = Abc_Clock() - timeRetime;
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Prepares the IVY package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvyAfter(pNtk *Abc_Ntk_t, pMan *Ivy_Man_t, fSeq int32, fHaig int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var nNodes int32;
    var fCleanup int32 = 1;
    // convert from the AIG manager
    if fSeq != 0 {
        pNtkAig = Abc_NtkFromIvySeq( pNtk, pMan, fHaig );
    } else {
        pNtkAig = Abc_NtkFromIvy( pNtk, pMan ); }
    // report the cleanup results
    if !fHaig && fCleanup && (nNodes = Abc_AigCleanup((Abc_Aig_t *)pNtkAig.pManFunc)) {
        fmt.Printf("Warning: AIG cleanup removed %d nodes (this is not a bug).\n", nNodes); }
    // duplicate EXDC 
    if pNtk.pExdc {
        pNtkAig.pExdc = Abc_NtkDup( pNtk.pExdc ); }
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkAig ) {
        fmt.Printf("Abc_NtkStrash: The network check has failed.\n");
        Abc_NtkDelete( pNtkAig );
        return nil;
    }
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvyStrash(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan *Ivy_Man_t;
    pMan = Abc_NtkIvyBefore( pNtk, 1, 0 );
    if pMan == nil {
        return nil; }
    pNtkAig = Abc_NtkIvyAfter( pNtk, pMan, 1, 0 );
    Ivy_ManStop( pMan );
    return pNtkAig;
}
 
/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvyHaig(pNtk *Abc_Ntk_t, nIters int32, fUseZeroCost int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan *Ivy_Man_t;
    var clk abctime;
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
    pMan = Abc_NtkIvyBefore( pNtk, 1, 1 );
    if pMan == nil {
        return nil; }
//timeRetime = Abc_Clock();

clk = Abc_Clock();
    Ivy_ManHaigStart( pMan, fVerbose );
//    Ivy_ManRewriteSeq( pMan, 0, 0 );
//    for ( i = 0; i < nIters; i++ )
//        Ivy_ManRewriteSeq( pMan, fUseZeroCost, 0 );

//printf( "%d ", Ivy_ManNodeNum(pMan) );
    Ivy_ManRewriteSeq( pMan, 0, 0 );
    Ivy_ManRewriteSeq( pMan, 0, 0 );
    Ivy_ManRewriteSeq( pMan, 1, 0 );
//printf( "%d ", Ivy_ManNodeNum(pMan) );
//printf( "%d ", Ivy_ManNodeNum(pMan->pHaig) );
//ABC_PRT( " ", Abc_Clock() - clk );
//printf( "\n" );
/*
    printf( "Moves = %d.  ", nMoves );
    printf( "MovesS = %d.  ", nMovesS );
    printf( "Clauses = %d.  ", nClauses );
    ABC_PRT( "Time", timeInv );
*/
//    Ivy_ManRewriteSeq( pMan, 1, 0 );
//printf( "Haig size = %d.\n", Ivy_ManNodeNum(pMan->pHaig) );
//    Ivy_ManHaigPostprocess( pMan, fVerbose );
//timeRetime = Abc_Clock() - timeRetime;

    // write working AIG into the current network
//    pNtkAig = Abc_NtkIvyAfter( pNtk, pMan, 1, 0 ); 
    // write HAIG into the current network
    pNtkAig = Abc_NtkIvyAfter( pNtk, pMan.pHaig, 1, 1 );

    Ivy_ManHaigStop( pMan );
    Ivy_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvyCuts(pNtk *Abc_Ntk_t, nInputs int32) {
    var pMan *Ivy_Man_t;
    pMan = Abc_NtkIvyBefore( pNtk, 1, 0 );
    if pMan == nil {
        return; }
    Ivy_CutComputeAll( pMan, nInputs );
    Ivy_ManStop( pMan );
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvyRewrite(pNtk *Abc_Ntk_t, fUpdateLevel int32, fUseZeroCost int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan *Ivy_Man_t;
    pMan = Abc_NtkIvyBefore( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
//timeRetime = Abc_Clock();
    Ivy_ManRewritePre( pMan, fUpdateLevel, fUseZeroCost, fVerbose );
//timeRetime = Abc_Clock() - timeRetime;
    pNtkAig = Abc_NtkIvyAfter( pNtk, pMan, 0, 0 );
    Ivy_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvyRewriteSeq(pNtk *Abc_Ntk_t, fUseZeroCost int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan *Ivy_Man_t;
    pMan = Abc_NtkIvyBefore( pNtk, 1, 1 );
    if pMan == nil {
        return nil; }
//timeRetime = Abc_Clock();
    Ivy_ManRewriteSeq( pMan, fUseZeroCost, fVerbose );
//timeRetime = Abc_Clock() - timeRetime;
//    Ivy_ManRewriteSeq( pMan, 1, 0 );
//    Ivy_ManRewriteSeq( pMan, 1, 0 );
    pNtkAig = Abc_NtkIvyAfter( pNtk, pMan, 1, 0 );
    Ivy_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvyResyn0(pNtk *Abc_Ntk_t, fUpdateLevel int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Ivy_Man_t;
    pMan = Abc_NtkIvyBefore( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
    pMan = Ivy_ManResyn0( pTemp = pMan, fUpdateLevel, fVerbose );
    Ivy_ManStop( pTemp );
    pNtkAig = Abc_NtkIvyAfter( pNtk, pMan, 0, 0 );
    Ivy_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvyResyn(pNtk *Abc_Ntk_t, fUpdateLevel int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Ivy_Man_t;
    pMan = Abc_NtkIvyBefore( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
    pMan = Ivy_ManResyn( pTemp = pMan, fUpdateLevel, fVerbose );
    Ivy_ManStop( pTemp );
    pNtkAig = Abc_NtkIvyAfter( pNtk, pMan, 0, 0 );
    Ivy_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvySat(pNtk *Abc_Ntk_t, nConfLimit int32, fVerbose int32) *Abc_Ntk_t {
    var Params Ivy_FraigParams_t;
    var pParams *Ivy_FraigParams_t = &Params; 
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Ivy_Man_t;
    pMan = Abc_NtkIvyBefore( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
    Ivy_FraigParamsDefault( pParams );
    pParams.nBTLimitMiter = nConfLimit;
    pParams.fVerbose = fVerbose;
//    pMan = Ivy_FraigPerform( pTemp = pMan, pParams );
    pMan = Ivy_FraigMiter( pTemp = pMan, pParams );
    Ivy_ManStop( pTemp );
    pNtkAig = Abc_NtkIvyAfter( pNtk, pMan, 0, 0 );
    Ivy_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Sets the final nodes to point to the original nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTransferPointers(pNtk *Abc_Ntk_t, pNtkAig *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var pObjIvy, pObjFraig *Ivy_Obj_t;
    var i int32;
    pObj = Abc_AigConst1(pNtk);
    pObj.pCopy = Abc_AigConst1(pNtkAig);
    Abc_NtkForEachCi( pNtk, pObj, i )
        pObj.pCopy = Abc_NtkCi(pNtkAig, i);
    Abc_NtkForEachCo( pNtk, pObj, i )
        pObj.pCopy = Abc_NtkCo(pNtkAig, i);
    Abc_NtkForEachLatch( pNtk, pObj, i )
        pObj.pCopy = Abc_NtkBox(pNtkAig, i);
    Abc_NtkForEachNode( pNtk, pObj, i )
    {
        pObjIvy = (Ivy_Obj_t *)pObj.pCopy;
        if pObjIvy == nil {
            continue; }
        pObjFraig = Ivy_ObjEquiv( pObjIvy );
        if pObjFraig == nil {
            continue; }
        pObj.pCopy = Abc_EdgeToNode( pNtkAig, Ivy_Regular(pObjFraig).TravId );
        pObj.pCopy = Abc_ObjNotCond( pObj.pCopy, Ivy_IsComplement(pObjFraig) );
    }
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvyFraig(pNtk *Abc_Ntk_t, nConfLimit int32, fDoSparse int32, fProve int32, fTransfer int32, fVerbose int32) *Abc_Ntk_t {
    var Params Ivy_FraigParams_t;
    var pParams *Ivy_FraigParams_t = &Params; 
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Ivy_Man_t;
    pMan = Abc_NtkIvyBefore( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
    Ivy_FraigParamsDefault( pParams );
    pParams.nBTLimitNode = nConfLimit;
    pParams.fVerbose     = fVerbose;
    pParams.fProve       = fProve;
    pParams.fDoSparse    = fDoSparse;
    pMan = Ivy_FraigPerform( pTemp = pMan, pParams );
    // transfer the pointers
    if fTransfer == 1 {
        var vCopies *Vec_Ptr_t;
        vCopies = Abc_NtkSaveCopy( pNtk );
        pNtkAig = Abc_NtkIvyAfter( pNtk, pMan, 0, 0 );
        Abc_NtkLoadCopy( pNtk, vCopies );
        Vec_PtrFree( vCopies );
        Abc_NtkTransferPointers( pNtk, pNtkAig );
    
    } else {
        pNtkAig = Abc_NtkIvyAfter( pNtk, pMan, 0, 0 ); }
    pNtkAig.pModel = (int *)pMan.pData; pMan.pData = nil;
    Ivy_ManStop( pTemp );
    Ivy_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Abc_NtkIvyProve(ppNtk **Abc_Ntk_t, pPars any) int32 {
    var pParams *Prove_Params_t  = (Prove_Params_t *)pPars;
    var pNtk *Abc_Ntk_t = *ppNtk;
    var pNtkTemp *Abc_Ntk_t;
    var pObj, pFanin *Abc_Obj_t;
    var pMan *Ivy_Man_t;
    var pMan2 *Aig_Man_t;
    var RetValue int32;
    assert(Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk), "Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk)");
    // experiment with various parameters settings
//    pParams->fUseBdds = 1;
//    pParams->fBddReorder = 1;
//    pParams->nTotalBacktrackLimit = 10000;
 
    // strash the network if it is not strashed already
    if !Abc_NtkIsStrash(pNtk) {
        pNtk = Abc_NtkStrash( pNtkTemp = pNtk, 0, 1, 0 );
        Abc_NtkDelete( pNtkTemp );
    }
 
    // check the case when the 0000 simulation pattern detect the bug
    pObj = Abc_NtkPo(pNtk,0);
    pFanin = Abc_ObjFanin0(pObj);
    if Abc_ObjFanin0(pObj).fPhase != uint32(Abc_ObjFaninC0(pObj)) {
        pNtk.pModel = ABC_CALLOC( int, Abc_NtkCiNum(pNtk) );
        return 0;
    }

    // changed in "src\sat\fraig\fraigMan.c"
    //    pParams->nMiteringLimitStart  = 300;    // starting mitering limit
    // to be
    //    pParams->nMiteringLimitStart  = 5000;    // starting mitering limit

    // if SAT only, solve without iteration
//    RetValue = Abc_NtkMiterSat( pNtk, 2*(ABC_INT64_T)pParams->nMiteringLimitStart, (ABC_INT64_T)0, 0, NULL, NULL );
    pMan2 = Abc_NtkToDar( pNtk, 0, 0 );
    RetValue = Fra_FraigSat( pMan2, (ABC_INT64_T)pParams.nMiteringLimitStart, (ABC_INT64_T)0, 0, 0, 0, 1, 0, 0, 0 ); 
    pNtk.pModel = (int *)pMan2.pData, pMan2.pData = nil;
    Aig_ManStop( pMan2 );
//    pNtk->pModel = Aig_ManReleaseData( pMan2 );
    if RetValue >= 0 {
        return RetValue; }

    // apply AIG rewriting
    if pParams.fUseRewriting && Abc_NtkNodeNum(pNtk) > 500 {
//        abctime clk = Abc_Clock();
//printf( "Before rwsat = %d. ", Abc_NtkNodeNum(pNtk) );
        pParams.fUseRewriting = 0;
        pNtk = Abc_NtkBalance( pNtkTemp = pNtk, 0, 0, 0 );          
        Abc_NtkDelete( pNtkTemp );
        Abc_NtkRewrite( pNtk, 0, 0, 0, 0, 0 );
        pNtk = Abc_NtkBalance( pNtkTemp = pNtk, 0, 0, 0 );          
        Abc_NtkDelete( pNtkTemp );
        Abc_NtkRewrite( pNtk, 0, 0, 0, 0, 0 );
        Abc_NtkRefactor( pNtk, 10, 16, 0, 0, 0, 0 );
//printf( "After rwsat = %d. ", Abc_NtkNodeNum(pNtk) );
//ABC_PRT( "Time", Abc_Clock() - clk );
    }

    // convert ABC network into IVY network
    pMan = Abc_NtkIvyBefore( pNtk, 0, 0 );

    // solve the CEC problem
    RetValue = Ivy_FraigProve( &pMan, pParams );
//    RetValue = -1;

    // convert IVY network into ABC network    
    pNtk = Abc_NtkIvyAfter( pNtkTemp = pNtk, pMan, 0, 0 );
    Abc_NtkDelete( pNtkTemp );
    // transfer model if given
    pNtk.pModel = (int *)pMan.pData; pMan.pData = nil;
    Ivy_ManStop( pMan );

    // try to prove it using brute force SAT with good CNF encoding
    if RetValue < 0 {
        pMan2 = Abc_NtkToDar( pNtk, 0, 0 );
        // dump the miter before entering high-effort solving
        if pParams.fVerbose {
            var pFileName string // was [100]byte in C

            pFileName = fmt.Sprintf("cecmiter.aig");
            Ioa_WriteAiger( pMan2, pFileName, 0, 0 );
            printf( "Intermediate reduced miter is written into file \"%s\".\n", pFileName );
        }
        RetValue = Fra_FraigSat( pMan2, pParams.nMiteringLimitLast, 0, 0, 0, 0, 0, 0, 0, pParams.fVerbose ); 
        pNtk.pModel = (int *)pMan2.pData, pMan2.pData = nil;
        Aig_ManStop( pMan2 );
    }

    // try to prove it using brute force BDDs
#ifdef ABC_USE_CUDD
    if ( RetValue < 0 && pParams->fUseBdds )
    {
        if ( pParams->fVerbose )
        {
            printf( "Attempting BDDs with node limit %d ...\n", pParams->nBddSizeLimit );
            fflush( stdout );
        }
        pNtk = Abc_NtkCollapse( pNtkTemp = pNtk, pParams->nBddSizeLimit, 0, pParams->fBddReorder, 0, 0, 0 );
        if ( pNtk )   
        {
            Abc_NtkDelete( pNtkTemp );
            RetValue = ( (Abc_NtkNodeNum(pNtk) == 1) && (Abc_ObjFanin0(Abc_NtkPo(pNtk,0))->pData == Cudd_ReadLogicZero((DdManager *)pNtk->pManFunc)) );
        }
        else 
            pNtk = pNtkTemp;
    }
#endif

    // return the result
    *ppNtk = pNtk;
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIvy(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
//    Abc_Ntk_t * pNtkAig;
    var pMan *Ivy_Man_t;//, * pTemp;
//    int fCleanup = 1;
//    int nNodes;
//    int nLatches = Abc_NtkLatchNum(pNtk);
    var vInit *Vec_Int_t  = Abc_NtkCollectLatchValuesIvy( pNtk, 0 );

    assert(!Abc_NtkIsNetlist(pNtk), "!Abc_NtkIsNetlist(pNtk)");
    if Abc_NtkIsBddLogic(pNtk) {
        if !Abc_NtkBddToSop(pNtk, -1, ABC_INFINITY) {
            Vec_IntFree( vInit );
            fmt.Printf("Abc_NtkIvy(): Converting to SOPs has failed.\n");
            return nil;
        }
    }
    if Abc_NtkCountSelfFeedLatches(pNtk) {
        fmt.Printf("Warning: The network has %d self-feeding latches. Quitting.\n", Abc_NtkCountSelfFeedLatches(pNtk));
        return nil;
    }

    // print warning about choice nodes
    if Abc_NtkGetChoiceNum( pNtk ) {
        fmt.Printf("Warning: The choice nodes in the initial AIG are removed by strashing.\n"); }

    // convert to the AIG manager
    pMan = Abc_NtkToIvy( pNtk );
    if !Ivy_ManCheck( pMan ) {
        Vec_IntFree( vInit );
        fmt.Printf("AIG check has failed.\n");
        Ivy_ManStop( pMan );
        return nil;
    }

//    Ivy_MffcTest( pMan );
//    Ivy_ManPrintStats( pMan );

//    pMan = Ivy_ManBalance( pTemp = pMan, 1 );
//    Ivy_ManStop( pTemp );

//    Ivy_ManSeqRewrite( pMan, 0, 0 );
//    Ivy_ManTestCutsAlg( pMan );
//    Ivy_ManTestCutsBool( pMan );
//    Ivy_ManRewriteAlg( pMan, 1, 1 );

//    pMan = Ivy_ManResyn( pTemp = pMan, 1, 0 );
//    Ivy_ManStop( pTemp );

//    Ivy_ManTestCutsAll( pMan );
//    Ivy_ManTestCutsTravAll( pMan );

//    Ivy_ManPrintStats( pMan );

//    Ivy_ManPrintStats( pMan );
//    Ivy_ManRewritePre( pMan, 1, 0, 0 );
//    Ivy_ManPrintStats( pMan );
//    printf( "\n" );

//    Ivy_ManPrintStats( pMan );
//    Ivy_ManMakeSeq( pMan, nLatches, pInit );
//    Ivy_ManPrintStats( pMan );

//    Ivy_ManRequiredLevels( pMan );

//    Ivy_FastMapPerform( pMan, 8 );
    Ivy_ManStop( pMan );
    return nil;


/*
    // convert from the AIG manager
    pNtkAig = Abc_NtkFromIvy( pNtk, pMan );
//    pNtkAig = Abc_NtkFromIvySeq( pNtk, pMan );
    Ivy_ManStop( pMan );

    // report the cleanup results
    if ( fCleanup && (nNodes = Abc_AigCleanup(pNtkAig->pManFunc)) )
        printf( "Warning: AIG cleanup removed %d nodes (this is not a bug).\n", nNodes );
    // duplicate EXDC 
    if ( pNtk->pExdc )
        pNtkAig->pExdc = Abc_NtkDup( pNtk->pExdc );
    // make sure everything is okay
    if ( !Abc_NtkCheck( pNtkAig ) )
    {
        ABC_FREE( pInit );
        printf( "Abc_NtkStrash: The network check has failed.\n" );
        Abc_NtkDelete( pNtkAig );
        return NULL;
    }

    ABC_FREE( pInit );
    return pNtkAig;
*/
}



/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromIvy(pNtkOld *Abc_Ntk_t, pMan *Ivy_Man_t) *Abc_Ntk_t {
    var vNodes *Vec_Int_t;
    var pNtk *Abc_Ntk_t;
    var pObj, pObjNew, pFaninNew, pFaninNew0, pFaninNew1 *Abc_Obj_t;
    var pNode *Ivy_Obj_t;
    var i int32;
    // perform strashing
    pNtk = Abc_NtkStartFrom( pNtkOld, ABC_NTK_STRASH, ABC_FUNC_AIG );
    // transfer the pointers to the basic nodes
    Ivy_ManConst1(pMan).TravId = Abc_EdgeFromNode( Abc_AigConst1(pNtk) );
    Abc_NtkForEachCi( pNtkOld, pObj, i )
        Ivy_ManPi(pMan, i).TravId = Abc_EdgeFromNode( pObj.pCopy );
    // rebuild the AIG
    vNodes = Ivy_ManDfs( pMan );
    Ivy_ManForEachNodeVec( pMan, vNodes, pNode, i )
    {
        // add the first fanin
        pFaninNew0 = Abc_ObjFanin0Ivy( pNtk, pNode );
        if Ivy_ObjIsBuf(pNode) {
            pNode.TravId = Abc_EdgeFromNode( pFaninNew0 );
            continue;
        }
        // add the second fanin
        pFaninNew1 = Abc_ObjFanin1Ivy( pNtk, pNode );
        // create the new node
        if Ivy_ObjIsExor(pNode) {
            pObjNew = Abc_AigXor( (Abc_Aig_t *)pNtk.pManFunc, pFaninNew0, pFaninNew1 );
        } else {
            pObjNew = Abc_AigAnd( (Abc_Aig_t *)pNtk.pManFunc, pFaninNew0, pFaninNew1 ); }
        pNode.TravId = Abc_EdgeFromNode( pObjNew );
    }
    // connect the PO nodes
    Abc_NtkForEachCo( pNtkOld, pObj, i )
    {
        pFaninNew = Abc_ObjFanin0Ivy( pNtk, Ivy_ManPo(pMan, i) );
        Abc_ObjAddFanin( pObj.pCopy, pFaninNew );
    }
    Vec_IntFree( vNodes );
    if !Abc_NtkCheck( pNtk ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkFromIvy(): Network check has failed.\n"); }
    return pNtk;
}

/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromIvySeq(pNtkOld *Abc_Ntk_t, pMan *Ivy_Man_t, fHaig int32) *Abc_Ntk_t {
    var vNodes, vLatches *Vec_Int_t;
    var pNtk *Abc_Ntk_t;
    var pObj, pObjNew, pFaninNew, pFaninNew0, pFaninNew1 *Abc_Obj_t;
    var pNode, pTemp *Ivy_Obj_t;
    var i int32;
//    assert( Ivy_ManLatchNum(pMan) > 0 );
    // perform strashing
    pNtk = Abc_NtkStartFromNoLatches( pNtkOld, ABC_NTK_STRASH, ABC_FUNC_AIG );
    // transfer the pointers to the basic nodes
    Ivy_ManConst1(pMan).TravId = Abc_EdgeFromNode( Abc_AigConst1(pNtk) );
    Abc_NtkForEachPi( pNtkOld, pObj, i )
        Ivy_ManPi(pMan, i).TravId = Abc_EdgeFromNode( pObj.pCopy );
    // create latches of the new network
    vNodes = Ivy_ManDfsSeq( pMan, &vLatches );
    Ivy_ManForEachNodeVec( pMan, vLatches, pNode, i )
    {
        pObjNew = Abc_NtkCreateLatch( pNtk );
        pFaninNew0 = Abc_NtkCreateBi( pNtk );
        pFaninNew1 = Abc_NtkCreateBo( pNtk );
        Abc_ObjAddFanin( pObjNew, pFaninNew0 );
        Abc_ObjAddFanin( pFaninNew1, pObjNew );
        if (fHaig != 0) || Ivy_ObjInit(pNode) == IVY_INIT_DC {
            Abc_LatchSetInitDc( pObjNew );
        } else if Ivy_ObjInit(pNode) == IVY_INIT_1 {
            Abc_LatchSetInit1( pObjNew );
        } else if Ivy_ObjInit(pNode) == IVY_INIT_0 {
            Abc_LatchSetInit0( pObjNew );
        } else { assert(false, "0"); }
        pNode.TravId = Abc_EdgeFromNode( pFaninNew1 );
    }
    Abc_NtkAddDummyBoxNames( pNtk );
    // rebuild the AIG
    Ivy_ManForEachNodeVec( pMan, vNodes, pNode, i )
    {
        // add the first fanin
        pFaninNew0 = Abc_ObjFanin0Ivy( pNtk, pNode );
        if Ivy_ObjIsBuf(pNode) {
            pNode.TravId = Abc_EdgeFromNode( pFaninNew0 );
            continue;
        }
        // add the second fanin
        pFaninNew1 = Abc_ObjFanin1Ivy( pNtk, pNode );
        // create the new node
        if Ivy_ObjIsExor(pNode) {
            pObjNew = Abc_AigXor( (Abc_Aig_t *)pNtk.pManFunc, pFaninNew0, pFaninNew1 );
        } else {
            pObjNew = Abc_AigAnd( (Abc_Aig_t *)pNtk.pManFunc, pFaninNew0, pFaninNew1 ); }
        pNode.TravId = Abc_EdgeFromNode( pObjNew );
        // process the choice nodes
        if (fHaig != 0) && pNode.pEquiv && Ivy_ObjRefs(pNode) > 0 {
            pFaninNew = Abc_EdgeToNode( pNtk, pNode.TravId );
//            pFaninNew->fPhase = 0;
            assert(!Ivy_IsComplement(pNode.pEquiv), "!Ivy_IsComplement(pNode.pEquiv)");
            for  pTemp = pNode.pEquiv; pTemp != pNode; pTemp = Ivy_Regular(pTemp.pEquiv)  {
                pFaninNew1 = Abc_EdgeToNode( pNtk, pTemp.TravId );
//                pFaninNew1->fPhase = Ivy_IsComplement( pTemp->pEquiv );
                pFaninNew.pData = pFaninNew1;
                pFaninNew = pFaninNew1;
            }
            pFaninNew.pData = nil;
//            printf( "Writing choice node %d.\n", pNode->Id );
        }
    }
    // connect the PO nodes
    Abc_NtkForEachPo( pNtkOld, pObj, i )
    {
        pFaninNew = Abc_ObjFanin0Ivy( pNtk, Ivy_ManPo(pMan, i) );
        Abc_ObjAddFanin( pObj.pCopy, pFaninNew );
    }
    // connect the latches
    Ivy_ManForEachNodeVec( pMan, vLatches, pNode, i )
    {
        pFaninNew = Abc_ObjFanin0Ivy( pNtk, pNode );
        Abc_ObjAddFanin( Abc_ObjFanin0(Abc_NtkBox(pNtk, i)), pFaninNew );
    }
    Vec_IntFree( vLatches );
    Vec_IntFree( vNodes );
    if !Abc_NtkCheck( pNtk ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkFromIvySeq(): Network check has failed.\n"); }
    return pNtk;
}

/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkToIvy(pNtkOld *Abc_Ntk_t) *Ivy_Man_t {
    var pMan *Ivy_Man_t;
    var pObj *Abc_Obj_t;
    var pFanin *Ivy_Obj_t;
    var i int32;
    // create the manager
    assert(Abc_NtkHasSop(pNtkOld) || Abc_NtkIsStrash(pNtkOld), "Abc_NtkHasSop(pNtkOld) || Abc_NtkIsStrash(pNtkOld)");
    pMan = Ivy_ManStart();
    // create the PIs
    if Abc_NtkIsStrash(pNtkOld) {
        Abc_AigConst1(pNtkOld).pCopy = (Abc_Obj_t *)Ivy_ManConst1(pMan); }
    Abc_NtkForEachCi( pNtkOld, pObj, i )
        pObj.pCopy = (Abc_Obj_t *)Ivy_ObjCreatePi(pMan);
    // perform the conversion of the internal nodes
    Abc_NtkStrashPerformAig( pNtkOld, pMan );
    // create the POs
    Abc_NtkForEachCo( pNtkOld, pObj, i )
    {
        pFanin = (Ivy_Obj_t *)Abc_ObjFanin0(pObj).pCopy;
        pFanin = Ivy_NotCond( pFanin, Abc_ObjFaninC0(pObj) );
        Ivy_ObjCreatePo( pMan, pFanin );
    }
    Ivy_ManCleanup( pMan );
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Prepares the network for strashing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkStrashPerformAig(pNtk *Abc_Ntk_t, pMan *Ivy_Man_t) {
//    ProgressBar * pProgress;
    var vNodes *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var i int32;
    vNodes = Abc_NtkDfs( pNtk, 0 );
//    pProgress = Extra_ProgressBarStart( stdout, vNodes->nSize );
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pNode, i )
    {
//        Extra_ProgressBarUpdate( pProgress, i, NULL );
        pNode.pCopy = (Abc_Obj_t *)Abc_NodeStrashAig( pMan, pNode );
    }
//    Extra_ProgressBarStop( pProgress );
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Strashes one logic node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeStrashAig(pMan *Ivy_Man_t, pNode *Abc_Obj_t) *Ivy_Obj_t {
    var fUseFactor int32  = 1;
    var pSop string;
    var pFanin0, pFanin1 *Ivy_Obj_t;

    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");

    // consider the case when the graph is an AIG
    if Abc_NtkIsStrash(pNode.pNtk) {
        if Abc_AigNodeIsConst(pNode) {
            return Ivy_ManConst1(pMan); }
        pFanin0 = (Ivy_Obj_t *)Abc_ObjFanin0(pNode).pCopy;
        pFanin0 = Ivy_NotCond( pFanin0, Abc_ObjFaninC0(pNode) );
        pFanin1 = (Ivy_Obj_t *)Abc_ObjFanin1(pNode).pCopy;
        pFanin1 = Ivy_NotCond( pFanin1, Abc_ObjFaninC1(pNode) );
        return Ivy_And( pMan, pFanin0, pFanin1 );
    }

    // get the SOP of the node
    if Abc_NtkHasMapping(pNode.pNtk) {
        pSop = Mio_GateReadSop((Mio_Gate_t *)pNode.pData);
    } else {
        pSop = (char *)pNode.pData; }

    // consider the constant node
    if Abc_NodeIsConst(pNode) {
        return Ivy_NotCond( Ivy_ManConst1(pMan), Abc_SopIsConst0(pSop) ); }

    // decide when to use factoring
    if (fUseFactor != 0) && Abc_ObjFaninNum(pNode) > 2 && Abc_SopGetCubeNum(pSop) > 1 && !Abc_SopIsExorType(pSop) {
        return Abc_NodeStrashAigFactorAig( pMan, pNode, pSop ); }
    return Abc_NodeStrashAigSopAig( pMan, pNode, pSop );
}

/**Function*************************************************************

  Synopsis    [Strashes one logic node using its SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeStrashAigSopAig(pMan *Ivy_Man_t, pNode *Abc_Obj_t, pSop string) *Ivy_Obj_t {
    var pFanin *Abc_Obj_t;
    var pAnd, pSum *Ivy_Obj_t;
    var pCube string;
    var i, nFanins int32;
    var fExor int32  = Abc_SopIsExorType(pSop);

    // get the number of node's fanins
    nFanins = Abc_ObjFaninNum( pNode );
    assert(nFanins == Abc_SopGetVarNum(pSop), "nFanins == Abc_SopGetVarNum(pSop)");
    // go through the cubes of the node's SOP
    pSum = Ivy_Not( Ivy_ManConst1(pMan) );
    Abc_SopForEachCube( pSop, nFanins, pCube )
    {
        // create the AND of literals
        pAnd = Ivy_ManConst1(pMan);
        Abc_ObjForEachFanin( pNode, pFanin, i ) // pFanin can be a net
        {
            if pCube[i] == '1' {
                pAnd = Ivy_And( pMan, pAnd, (Ivy_Obj_t *)pFanin.pCopy );
            } else if pCube[i] == '0' {
                pAnd = Ivy_And( pMan, pAnd, Ivy_Not((Ivy_Obj_t *)pFanin.pCopy) ); }
        }
        // add to the sum of cubes
        if fExor != 0 {
            pSum = Ivy_Exor( pMan, pSum, pAnd );
        } else {
            pSum = Ivy_Or( pMan, pSum, pAnd ); }
    }
    // decide whether to complement the result
    if Abc_SopIsComplement(pSop) {
        pSum = Ivy_Not(pSum); }
    return pSum;
}

/**Function*************************************************************

  Synopsis    [Strashed n-input XOR function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeStrashAigExorAig(pMan *Ivy_Man_t, pNode *Abc_Obj_t, pSop string) *Ivy_Obj_t {
    var pFanin *Abc_Obj_t;
    var pSum *Ivy_Obj_t;
    var i, nFanins int32;
    // get the number of node's fanins
    nFanins = Abc_ObjFaninNum( pNode );
    assert(nFanins == Abc_SopGetVarNum(pSop), "nFanins == Abc_SopGetVarNum(pSop)");
    // go through the cubes of the node's SOP
    pSum = Ivy_Not( Ivy_ManConst1(pMan) );
    for  i = 0; i < nFanins; i++  {
        pFanin = Abc_ObjFanin( pNode, i );
        pSum = Ivy_Exor( pMan, pSum, (Ivy_Obj_t *)pFanin.pCopy );
    }
    if Abc_SopIsComplement(pSop) {
        pSum = Ivy_Not(pSum); }
    return pSum;
}

/**Function*************************************************************

  Synopsis    [Strashes one logic node using its SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeStrashAigFactorAig(pMan *Ivy_Man_t, pRoot *Abc_Obj_t, pSop string) *Ivy_Obj_t {
    var pFForm *Dec_Graph_t;
    var pNode *Dec_Node_t;
    var pAnd *Ivy_Obj_t;
    var i int32;

//    extern Ivy_Obj_t * Dec_GraphToNetworkAig( Ivy_Man_t * pMan, Dec_Graph_t * pGraph );

//    assert( 0 );

    // perform factoring
    pFForm = Dec_Factor( pSop );
    // collect the fanins
    Dec_GraphForEachLeaf( pFForm, pNode, i )
        pNode.pFunc = Abc_ObjFanin(pRoot,i).pCopy;
    // perform strashing
//    pAnd = Dec_GraphToNetworkAig( pMan, pFForm );
    pAnd = Dec_GraphToNetworkIvy( pMan, pFForm );
//    pAnd = NULL;

    Dec_GraphFree( pFForm );
    return pAnd;
}

/**Function*************************************************************

  Synopsis    [Strashes one logic node using its SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectLatchValuesIvy(pNtk *Abc_Ntk_t, fUseDcs int32) *Vec_Int_t {
    var pLatch *Abc_Obj_t;
    var vArray *Vec_Int_t;
    var i int32;
    vArray = Vec_IntAlloc( Abc_NtkLatchNum(pNtk) );
    Abc_NtkForEachLatch( pNtk, pLatch, i )
    {
        if (fUseDcs != 0) || Abc_LatchIsInitDc(pLatch) {
            Vec_IntPush( vArray, IVY_INIT_DC );
        } else if Abc_LatchIsInit1(pLatch) {
            Vec_IntPush( vArray, IVY_INIT_1 );
        } else if Abc_LatchIsInit0(pLatch) {
            Vec_IntPush( vArray, IVY_INIT_0 );
        } else { assert(false, "0"); }
    }
    return vArray;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

