package llb

import (
	"fmt"
	"math"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [llb2Nonlin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Non-linear quantification scheduling.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb2Nonlin.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"

//ABC_NAMESPACE_IMPL_START
 

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Llb_Mnn_t struct {
    pInit      *Aig_Man_t;          // AIG manager
    pAig      *Aig_Man_t;           // AIG manager
    pPars   *Gia_ParLlb_t;          // parameters

    dd      *DdManager;             // BDD manager
    ddG      *DdManager;            // BDD manager
    ddR      *DdManager;            // BDD manager
    vRings      *Vec_Ptr_t;         // onion rings in ddR

    vLeaves      *Vec_Ptr_t;        
    vRoots      *Vec_Ptr_t;
    pVars2Q            *int32;
    pOrderL            *int32;
    pOrderL2            *int32;
    pOrderG            *int32;

    vCs2Glo      *Vec_Int_t;        // cur state variables into global variables
    vNs2Glo      *Vec_Int_t;        // next state variables into global variables
    vGlo2Cs      *Vec_Int_t;        // global variables into cur state variables
    vGlo2Ns      *Vec_Int_t;        // global variables into next state variables

    ddLocReos             int32;
    ddLocGrbs             int32;

    timeImage         abctime;
    timeTran1         abctime;
    timeTran2         abctime;
    timeGloba         abctime;
    timeOther         abctime;
    timeTotal         abctime;
    timeReo         abctime;
    timeReoG         abctime;

}

var timeBuild, timeAndEx, timeOther abctime;
var nSuppMax int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Finds variable whose 0-cofactor is the smallest.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinFindBestVar(dd *DdManager, bFunc *DdNode, pAig *Aig_Man_t) int32 {
    var fVerbose int32;
    var pObj *Aig_Obj_t;
    var bCof, bVar *DdNode;
    var i int32;
    var iVar int32;
    var iVarBest int32 = -1;
    var iValue int32;
    var iValueBest int32 = ABC_INFINITY;
    var Size0Best int32 = -1;
    var Size, Size0, Size1 int32;
    var clk abctime  = Abc_Clock();
    Size = Cudd_DagSize(bFunc);
//    printf( "Original = %6d.  SuppSize = %3d. Vars = %3d.\n", 
//        Size = Cudd_DagSize(bFunc), Cudd_SupportSize(dd, bFunc), Aig_ManRegNum(pAig) );
    Saig_ManForEachLo( pAig, pObj, i )
    {
        iVar = Aig_ObjId(pObj);

if fVerbose != 0 {
fmt.Printf("Var =%3d : ", iVar); }
        bVar = Cudd_bddIthVar(dd, iVar);

        // inlining of macro Cudd_Not
        bCof = Cudd_bddAnd( dd, bFunc, ((DdNode *)((ptrint)(bVar) ^ 01)) );          Cudd_Ref( bCof );
        Size0 = Cudd_DagSize(bCof);
if fVerbose != 0 {
fmt.Printf("Supp0 =%3d  ", Cudd_SupportSize(dd, bCof)); }
if fVerbose != 0 {
fmt.Printf("Size0 =%6d   ", Size0); }
        Cudd_RecursiveDeref( dd, bCof );

        bCof = Cudd_bddAnd( dd, bFunc, bVar );                    Cudd_Ref( bCof );
        Size1 = Cudd_DagSize(bCof);
if fVerbose != 0 {
fmt.Printf("Supp1 =%3d  ", Cudd_SupportSize(dd, bCof)); }
if fVerbose != 0 {
fmt.Printf("Size1 =%6d   ", Size1); }
        Cudd_RecursiveDeref( dd, bCof );

        iValue = Abc_MaxInt(Size0, Size1) - Abc_MinInt(Size0, Size1) + Size0 + Size1 - Size;
if fVerbose != 0 {
fmt.Printf("D =%6d  ", Size0 + Size1 - Size); }
if fVerbose != 0 {
fmt.Printf("B =%6d  ", Abc_MaxInt(Size0, Size1) - Abc_MinInt(Size0, Size1)); }
if fVerbose != 0 {
fmt.Printf("S =%6d\n", iValue); }
        if Size0 > 1 && Size1 > 1 && iValueBest > iValue {
            iValueBest = iValue;
            iVarBest   = i;
            Size0Best  = Size0;
        }
    }
    fmt.Printf("BestVar = %4d/%4d.  Value =%6d.  Orig =%6d. Size0 =%6d. ", iVarBest, Aig_ObjId(Saig_ManLo(pAig,iVarBest)), iValueBest, Size, Size0Best);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return iVarBest;
}


/**Function*************************************************************

  Synopsis    [Finds variable whose 0-cofactor is the smallest.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinTrySubsetting(dd *DdManager, bFunc *DdNode) {
    var bNew *DdNode;
    fmt.Printf("Original = %6d.  SuppSize = %3d.    ", Cudd_DagSize(bFunc), Cudd_SupportSize(dd, bFunc));
    bNew = Cudd_SubsetHeavyBranch( dd, bFunc, Cudd_SupportSize(dd, bFunc), 1000 );  Cudd_Ref( bNew );
    fmt.Printf("Result   = %6d.  SuppSize = %3d.\n", Cudd_DagSize(bNew), Cudd_SupportSize(dd, bNew));
    Cudd_RecursiveDeref( dd, bNew );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinPrepareVarMap(p *Llb_Mnn_t) {
    var pObjLi, pObjLo, pObj *Aig_Obj_t;
    var i, iVarLi, iVarLo int32;
    p.vCs2Glo = Vec_IntStartFull( Aig_ManObjNumMax(p.pAig) );
    p.vNs2Glo = Vec_IntStartFull( Aig_ManObjNumMax(p.pAig) );
    p.vGlo2Cs = Vec_IntStartFull( Aig_ManRegNum(p.pAig) );
    p.vGlo2Ns = Vec_IntStartFull( Aig_ManRegNum(p.pAig) );
    Saig_ManForEachLiLo( p.pAig, pObjLi, pObjLo, i )
    {
        iVarLi = Aig_ObjId(pObjLi);
        iVarLo = Aig_ObjId(pObjLo);
        assert(iVarLi >= 0 && iVarLi < Aig_ManObjNumMax(p.pAig), "iVarLi >= 0 && iVarLi < Aig_ManObjNumMax(p.pAig)");
        assert(iVarLo >= 0 && iVarLo < Aig_ManObjNumMax(p.pAig), "iVarLo >= 0 && iVarLo < Aig_ManObjNumMax(p.pAig)");
        Vec_IntWriteEntry( p.vCs2Glo, iVarLo, i );
        Vec_IntWriteEntry( p.vNs2Glo, iVarLi, i );
        Vec_IntWriteEntry( p.vGlo2Cs, i, iVarLo );
        Vec_IntWriteEntry( p.vGlo2Ns, i, iVarLi );
    }
    // add mapping of the PIs
    Saig_ManForEachPi( p.pAig, pObj, i )
    {
        Vec_IntWriteEntry( p.vCs2Glo, Aig_ObjId(pObj), Aig_ManRegNum(p.pAig)+i );
        Vec_IntWriteEntry( p.vNs2Glo, Aig_ObjId(pObj), Aig_ManRegNum(p.pAig)+i );
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinComputeInitState(pAig *Aig_Man_t, dd *DdManager) *DdNode {
    var pObj *Aig_Obj_t;
    var bRes, bVar, bTemp *DdNode;
    var i, iVar int32;
    var TimeStop abctime;
    TimeStop = dd.TimeStop;  dd.TimeStop = 0;
    bRes = Cudd_ReadOne( dd );   Cudd_Ref( bRes );
    Saig_ManForEachLo( pAig, pObj, i )
    {
        if Cudd_ReadSize(dd) == Aig_ManRegNum(pAig) {
            iVar = i;
        } else {
            iVar = Aig_ObjId(pObj);
        }
        bVar = Cudd_bddIthVar( dd, iVar );
        // inlining of macro Cudd_Not
        bRes = Cudd_bddAnd( dd, bTemp = bRes, ((DdNode *)((ptrint)(bVar) ^ 01)) );  Cudd_Ref( bRes );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    Cudd_Deref( bRes );
    dd.TimeStop = TimeStop;
    return bRes;
}


/**Function*************************************************************

  Synopsis    [Derives counter-example by backward reachability.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinDeriveCex(p *Llb_Mnn_t) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var pObj *Aig_Obj_t;
    var vVarsNs *Vec_Int_t;
    var bState *DdNode = nil;
    var bImage *DdNode;
    var bOneCube *DdNode;
    var bTemp *DdNode;
    var bRing *DdNode;
    var i, v, RetValue, nPiOffset int32;
    var pValues string  = ABC_ALLOC( char, Cudd_ReadSize(p.ddR) );
    assert(Vec_PtrSize(p.vRings) > 0, "Vec_PtrSize(p.vRings) > 0");

    p.dd.TimeStop  = 0;
    p.ddR.TimeStop = 0;

    // update quantifiable vars
    memset( p.pVars2Q, 0, sizeof(int) * Cudd_ReadSize(p.dd) );
    vVarsNs = Vec_IntAlloc( Aig_ManRegNum(p.pAig) );
    Saig_ManForEachLi( p.pAig, pObj, i )
    {
        p.pVars2Q[Aig_ObjId(pObj)] = 1;
        Vec_IntPush( vVarsNs, Aig_ObjId(pObj) );
    }
/*
    Saig_ManForEachLo( p->pAig, pObj, i )
        printf( "%d ", pObj->Id );
    printf( "\n" );
    Saig_ManForEachLi( p->pAig, pObj, i )
        printf( "%d(%d) ", pObj->Id, Aig_ObjFaninId0(pObj) );
    printf( "\n" );
*/
    // allocate room for the counter-example
    pCex = Abc_CexAlloc( Saig_ManRegNum(p.pAig), Saig_ManPiNum(p.pAig), Vec_PtrSize(p.vRings) );
    pCex.iFrame = Vec_PtrSize(p.vRings) - 1;
    pCex.iPo = -1;

    // get the last cube
    bOneCube = Cudd_bddIntersect( p.ddR, (DdNode *)Vec_PtrEntryLast(p.vRings), p.ddR.bFunc );  Cudd_Ref( bOneCube );
    RetValue = Cudd_bddPickOneCube( p.ddR, bOneCube, pValues );
    Cudd_RecursiveDeref( p.ddR, bOneCube );
    assert((RetValue != 0), "RetValue");

    // write PIs of counter-example
    nPiOffset = Saig_ManRegNum(p.pAig) + Saig_ManPiNum(p.pAig) * (Vec_PtrSize(p.vRings) - 1);
    Saig_ManForEachPi( p.pAig, pObj, i )
        if ( pValues[Saig_ManRegNum(p.pAig)+i] == 1 )
            Abc_InfoSetBit( pCex.pData, nPiOffset + i );

    // write state in terms of NS variables
    if Vec_PtrSize(p.vRings) > 1 {
        bState = Llb_CoreComputeCube( p.dd, vVarsNs, 1, pValues );   Cudd_Ref( bState );
    }
    // perform backward analysis
    Vec_PtrForEachEntryReverse( DdNode *, p.vRings, bRing, v )
    { 
        if v == Vec_PtrSize(p.vRings) - 1 {
            continue; }
//Extra_bddPrintSupport( p->dd, bState );  printf( "\n" );
//Extra_bddPrintSupport( p->dd, bRing );   printf( "\n" );
        // compute the next states
        bImage = Llb_NonlinImage( p.pAig, p.vLeaves, p.vRoots, p.pVars2Q, p.dd, bState, 
            p.pPars.fReorder, p.pPars.fVeryVerbose, nil ); // consumed reference
        assert(bImage != nil, "bImage != nil");
        Cudd_Ref( bImage );
//Extra_bddPrintSupport( p->dd, bImage );  printf( "\n" );

        // move reached states into ring manager
        bImage = Extra_TransferPermute( p.dd, p.ddR, bTemp = bImage, Vec_IntArray(p.vCs2Glo) );    Cudd_Ref( bImage );
        Cudd_RecursiveDeref( p.dd, bTemp );

        // intersect with the previous set
        bOneCube = Cudd_bddIntersect( p.ddR, bImage, bRing );                Cudd_Ref( bOneCube );
        Cudd_RecursiveDeref( p.ddR, bImage );

        // find any assignment of the BDD
        RetValue = Cudd_bddPickOneCube( p.ddR, bOneCube, pValues );
        Cudd_RecursiveDeref( p.ddR, bOneCube );
        assert((RetValue != 0), "RetValue");

        // write PIs of counter-example
        nPiOffset -= Saig_ManPiNum(p.pAig);
        Saig_ManForEachPi( p.pAig, pObj, i )
            if ( pValues[Saig_ManRegNum(p.pAig)+i] == 1 )
                Abc_InfoSetBit( pCex.pData, nPiOffset + i );

        // check that we get the init state
        if v == 0 {
            Saig_ManForEachLo( p.pAig, pObj, i )
                assert(pValues[i] == 0, "pValues[i] == 0");
            break;
        }

        // write state in terms of NS variables
        bState = Llb_CoreComputeCube( p.dd, vVarsNs, 1, pValues );   Cudd_Ref( bState );
    }
    assert(nPiOffset == Saig_ManRegNum(p.pAig), "nPiOffset == Saig_ManRegNum(p.pAig)");
    // update the output number
//Abc_CexPrint( pCex );
    RetValue = Saig_ManFindFailedPoCex( p.pInit, pCex );
    assert(RetValue >= 0 && RetValue < Saig_ManPoNum(p.pInit), "RetValue >= 0 && RetValue < Saig_ManPoNum(p.pInit)"); // invalid CEX!!!
    pCex.iPo = RetValue;
    // cleanup
    ABC_FREE( pValues );
    Vec_IntFree( vVarsNs );
    return pCex;
}


/**Function*************************************************************

  Synopsis    [Perform reachability with hints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinReoHook(dd *DdManager, Type string, Method any) int32 {
    var pAig *Aig_Man_t  = (Aig_Man_t *)dd.bFunc;
    var pObj *Aig_Obj_t;
    var i int32;
    fmt.Printf("Order: ");
    for  i = 0; i < Cudd_ReadSize(dd); i++  {
        pObj = Aig_ManObj( pAig, i );
        if pObj == nil {
            continue; }
        if Saig_ObjIsPi(pAig, pObj) {
            fmt.Printf("pi");
        } else if Saig_ObjIsLo(pAig, pObj) {
            fmt.Printf("lo");
        } else if Saig_ObjIsPo(pAig, pObj) {
            fmt.Printf("po");
        } else if Saig_ObjIsLi(pAig, pObj) {
            fmt.Printf("li");
        } else { continue; }
        fmt.Printf("%d=%d ", i, dd.perm[i]);
    }
    fmt.Printf("\n");
    return 1;
}

/**Function*************************************************************

  Synopsis    [Perform reachability with hints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinCompPerms(dd *DdManager, pVar2Lev *int32) int32 {
    var pSubt *DdSubtable;
    var i, Sum, Entry int32;
    for  i = 0; i < dd.size; i++  {
        pSubt = &(dd.subtables[dd.perm[i]]);
        if pSubt.keys == pSubt.dead + 1 {
            continue; }
        Entry = Abc_MaxInt(dd.perm[i], pVar2Lev[i]) - Abc_MinInt(dd.perm[i], pVar2Lev[i]);
        Sum += Entry;
//printf( "%d-%d(%d) ", dd->perm[i], pV2L[i], Entry );
    }
    return Sum;
}

/**Function*************************************************************

  Synopsis    [Perform reachability with hints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinReachability(p *Llb_Mnn_t) int32 { 
    var bTemp, bNext *DdNode;
    var nIters int32;
    var nBddSize0 int32;
    var nBddSize int32 = -1;
    var NumCmp int32;//, Limit = p->pPars->nBddMax;
    var clk2 abctime;
    var clk3 abctime;
    var clk abctime = Abc_Clock();
    assert(Aig_ManRegNum(p.pAig) > 0, "Aig_ManRegNum(p.pAig) > 0");

    // compute time to stop
    if p.pPars.TimeLimit {
        p.pPars.TimeTarget = p.pPars.TimeLimit * int64(time.Second) + Abc_Clock();
    } else {
        p.pPars.TimeTarget = 0;
    }

    // set the stop time parameter
    p.dd.TimeStop  = p.pPars.TimeTarget;
    p.ddG.TimeStop = p.pPars.TimeTarget;
    p.ddR.TimeStop = p.pPars.TimeTarget;

    // set reordering hooks
    assert(p.dd.bFunc == nil, "p.dd.bFunc == nil");
//    p->dd->bFunc = (DdNode *)p->pAig;
//    Cudd_AddHook( p->dd, Llb_NonlinReoHook, CUDD_POST_REORDERING_HOOK );

    // create bad state in the ring manager
    p.ddR.bFunc  = Llb_BddComputeBad( p.pInit, p.ddR, p.pPars.TimeTarget );          
    if p.ddR.bFunc == nil {
        if !p.pPars.fSilent {
            fmt.Printf("Reached timeout (%d seconds) during constructing the bad states.\n", p.pPars.TimeLimit); }
        p.pPars.iFrame = -1;
        return -1;
    }
    Cudd_Ref( p.ddR.bFunc );
    // compute the starting set of states
    Cudd_Quit( p.dd );
    p.dd = Llb_NonlinImageStart( p.pAig, p.vLeaves, p.vRoots, p.pVars2Q, p.pOrderL, 1, p.pPars.TimeTarget );
    if p.dd == nil {
        if !p.pPars.fSilent {
            fmt.Printf("Reached timeout (%d seconds) during constructing the bad states.\n", p.pPars.TimeLimit); }
        p.pPars.iFrame = -1;
        return -1;
    }
    p.dd.bFunc   = Llb_NonlinComputeInitState( p.pAig, p.dd );   Cudd_Ref( p.dd.bFunc );   // current
    p.ddG.bFunc  = Llb_NonlinComputeInitState( p.pAig, p.ddG );  Cudd_Ref( p.ddG.bFunc );  // reached
    p.ddG.bFunc2 = Llb_NonlinComputeInitState( p.pAig, p.ddG );  Cudd_Ref( p.ddG.bFunc2 ); // frontier 
    for  nIters = 0; nIters < p.pPars.nIterMax; nIters++  { 
        // check the runtime limit
        clk2 = Abc_Clock();
        if p.pPars.TimeLimit && Abc_Clock() > p.pPars.TimeTarget {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Llb_NonlinImageQuit();
            return -1;
        }

        // save the onion ring
        bTemp = Extra_TransferPermute( p.dd, p.ddR, p.dd.bFunc, Vec_IntArray(p.vCs2Glo) );
        if bTemp == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during ring transfer.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Llb_NonlinImageQuit();
            return -1;
        }
        Cudd_Ref( bTemp );
        Vec_PtrPush( p.vRings, bTemp );

        // check it for bad states
        // inlining of macro Cudd_Not
        if !p.pPars.fSkipOutCheck && !Cudd_bddLeq( p.ddR, bTemp, ((DdNode *)((ptrint)(p.ddR.bFunc) ^ 01)) ) {
            assert(p.pInit.pSeqModel == nil, "p.pInit.pSeqModel == nil");
            if !p.pPars.fBackward {
                p.pInit.pSeqModel = Llb_NonlinDeriveCex( p ); } 
            if !p.pPars.fSilent {
                if !p.pPars.fBackward {
                    Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.  ", p.pInit.pSeqModel.iPo, nIters );
                } else {
                    Abc_Print( 1, "Output ??? was asserted in frame %d (counter-example is not produced).  ", nIters ); }
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            p.pPars.iFrame = nIters - 1;
            Llb_NonlinImageQuit();
            return 0;
        }

        // compute the next states
        clk3 = Abc_Clock();
        nBddSize0 = Cudd_DagSize( p.dd.bFunc );
        bNext = Llb_NonlinImageCompute( p.dd.bFunc, p.pPars.fReorder, 0, 1, p.pOrderL ); // consumes ref   
//        bNext = Llb_NonlinImage( p->pAig, p->vLeaves, p->vRoots, p->pVars2Q, p->dd, bCurrent, 
//            p->pPars->fReorder, p->pPars->fVeryVerbose, NULL, ABC_INFINITY, p->pPars->TimeTarget );
        if bNext == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation in quantification.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Llb_NonlinImageQuit();
            return -1;
        }
        Cudd_Ref( bNext );
        nBddSize = Cudd_DagSize( bNext );
        p.timeImage += Abc_Clock() - clk3;


        // transfer to the state manager
        clk3 = Abc_Clock();
        Cudd_RecursiveDeref( p.ddG, p.ddG.bFunc2 );
        p.ddG.bFunc2 = Extra_TransferPermute( p.dd, p.ddG, bNext, Vec_IntArray(p.vNs2Glo) );    
//        p->ddG->bFunc2 = Extra_bddAndPermute( p->ddG, Cudd_Not(p->ddG->bFunc), p->dd, bNext, Vec_IntArray(p->vNs2Glo) );    
        if p.ddG.bFunc2 == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation in transfer 1.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Cudd_RecursiveDeref( p.dd,  bNext );  
            Llb_NonlinImageQuit();
            return -1;
        }
        Cudd_Ref( p.ddG.bFunc2 );
        Cudd_RecursiveDeref( p.dd, bNext );
        p.timeTran1 += Abc_Clock() - clk3;

        // save permutation
        NumCmp = Llb_NonlinCompPerms( p.dd, p.pOrderL2 );
        // save order before image computation
        memcpy( p.pOrderL2, p.dd.perm, sizeof(int) * p.dd.size );
        // update the image computation manager
        p.timeReo   += Cudd_ReadReorderingTime(p.dd);
        p.ddLocReos += Cudd_ReadReorderings(p.dd);
        p.ddLocGrbs += Cudd_ReadGarbageCollections(p.dd);
        Llb_NonlinImageQuit();
        p.dd = Llb_NonlinImageStart( p.pAig, p.vLeaves, p.vRoots, p.pVars2Q, p.pOrderL, 0, p.pPars.TimeTarget );
        if p.dd == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during constructing the bad states.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            return -1;
        }
        //Extra_TestAndPerm( p->ddG, Cudd_Not(p->ddG->bFunc), p->ddG->bFunc2 );    

        // derive new states
        clk3 = Abc_Clock();
        // inlining of macro Cudd_Not
        p.ddG.bFunc2 = Cudd_bddAnd( p.ddG, bTemp = p.ddG.bFunc2, ((DdNode *)((ptrint)(p.ddG.bFunc) ^ 01)) );     
        if p.ddG.bFunc2 == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation in transfer 1.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Cudd_RecursiveDeref( p.ddG, bTemp );  
            Llb_NonlinImageQuit();
            return -1;
        }
        Cudd_Ref( p.ddG.bFunc2 );
        Cudd_RecursiveDeref( p.ddG, bTemp );
        p.timeGloba += Abc_Clock() - clk3;

        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        if  (((DdNode *)((ptruint)(p.ddG.bFunc2) & ~01))).index == CUDD_CONST_INDEX {
            break; }
        // add to the reached set
        clk3 = Abc_Clock();
        p.ddG.bFunc = Cudd_bddOr( p.ddG, bTemp = p.ddG.bFunc, p.ddG.bFunc2 );                 
        if p.ddG.bFunc == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation in transfer 1.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Cudd_RecursiveDeref( p.ddG, bTemp );  
            Llb_NonlinImageQuit();
            return -1;
        }
        Cudd_Ref( p.ddG.bFunc );
        Cudd_RecursiveDeref( p.ddG, bTemp );
        p.timeGloba += Abc_Clock() - clk3;

        // reset permutation
//        RetValue = Cudd_CheckZeroRef( dd );
//        assert( RetValue == 0 );
//        Cudd_ShuffleHeap( dd, pOrderG );

        // move new states to the working manager
        clk3 = Abc_Clock();
        p.dd.bFunc = Extra_TransferPermute( p.ddG, p.dd, p.ddG.bFunc2, Vec_IntArray(p.vGlo2Cs) ); 
        if p.dd.bFunc == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation in transfer 2.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Llb_NonlinImageQuit();
            return -1;
        }
        Cudd_Ref( p.dd.bFunc );
        p.timeTran2 += Abc_Clock() - clk3;

        // report the results
        if p.pPars.fVerbose {
            fmt.Printf("I =%3d : ", nIters);
            fmt.Printf("Fr =%7d ", nBddSize0);
            fmt.Printf("Im =%7d  ", nBddSize);
            fmt.Printf("(%4d %4d)  ", p.ddLocReos, p.ddLocGrbs);
            fmt.Printf("Rea =%6d  ", Cudd_DagSize(p.ddG.bFunc));
            fmt.Printf("(%4d %4d)  ", Cudd_ReadReorderings(p.ddG), Cudd_ReadGarbageCollections(p.ddG));
            fmt.Printf("S =%4d ", nSuppMax);
            fmt.Printf("cL =%5d ", NumCmp);
            fmt.Printf("cG =%5d ", Llb_NonlinCompPerms( p.ddG, p.pOrderG ));
            Abc_PrintTime( 1, "T", Abc_Clock() - clk2 );
            memcpy( p.pOrderG, p.ddG.perm, sizeof(int) * p.ddG.size );
        }
/*
        if ( pPars->fVerbose )
        {
            double nMints = Cudd_CountMinterm(ddG, bReached, Saig_ManRegNum(pAig) );
//            Extra_bddPrint( ddG, bReached );printf( "\n" );
            printf( "Reachable states = %.0f. (Ratio = %.4f %%)\n", nMints, 100.0*nMints/pow(2.0, Saig_ManRegNum(pAig)) );
            fflush( stdout ); 
        }
*/
        if nIters == p.pPars.nIterMax - 1 {
            if !p.pPars.fSilent {
                fmt.Printf("Reached limit on the number of timeframes (%d).\n", p.pPars.nIterMax); }
            p.pPars.iFrame = nIters;
            Llb_NonlinImageQuit();
            return -1;
        }
    }
    Llb_NonlinImageQuit();
    
    // report the stats
    if p.pPars.fVerbose {
        var nMints float64  = Cudd_CountMinterm(p.ddG, p.ddG.bFunc, Saig_ManRegNum(p.pAig) );
        if nIters >= p.pPars.nIterMax || nBddSize > p.pPars.nBddMax {
            fmt.Printf("Reachability analysis is stopped after %d frames.\n", nIters);
        } else {
            fmt.Printf("Reachability analysis completed after %d frames.\n", nIters); }
        fmt.Printf("Reachable states = %.0f. (Ratio = %.4f %%)\n", nMints, 100.0*nMints/math.Pow(float64(2.0), float64(Saig_ManRegNum(p.pAig))));
        fflush( os.Stdout ); 
    }
    if nIters >= p.pPars.nIterMax || nBddSize > p.pPars.nBddMax {
        if !p.pPars.fSilent {
            fmt.Printf("Verified only for states reachable in %d frames.  ", nIters); }
        p.pPars.iFrame = p.pPars.nIterMax;
        return -1; // undecided
    }
    // report
    if !p.pPars.fSilent {
        fmt.Printf("The miter is proved unreachable after %d iterations.  ", nIters); }
    p.pPars.iFrame = nIters - 1;
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return 1; // unreachable
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_MnnStart(pInit *Aig_Man_t, pAig *Aig_Man_t, pPars *Gia_ParLlb_t) *Llb_Mnn_t {
    var p *Llb_Mnn_t;
    var pObj *Aig_Obj_t;
    var i int32;
    p = new(Llb_Mnn_t);
    p.pInit = pInit;
    p.pAig  = pAig;
    p.pPars = pPars;
    p.dd    = Cudd_Init( Aig_ManObjNumMax(pAig), 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    p.ddG   = Cudd_Init( Aig_ManRegNum(pAig),    0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    p.ddR   = Cudd_Init( Aig_ManCiNum(pAig),     0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    Cudd_AutodynEnable( p.dd,  CUDD_REORDER_SYMM_SIFT );
    Cudd_AutodynEnable( p.ddG, CUDD_REORDER_SYMM_SIFT );
    Cudd_AutodynEnable( p.ddR, CUDD_REORDER_SYMM_SIFT );
    p.vRings = Vec_PtrAlloc( 100 );
    // create leaves
    p.vLeaves = Vec_PtrAlloc( Aig_ManCiNum(pAig) );
    Aig_ManForEachCi( pAig, pObj, i )
        Vec_PtrPush( p.vLeaves, pObj );
    // create roots
    p.vRoots = Vec_PtrAlloc( Aig_ManCoNum(pAig) );
    Saig_ManForEachLi( pAig, pObj, i )
        Vec_PtrPush( p.vRoots, pObj );
    // variables to quantify
    p.pOrderL = ABC_CALLOC( int, Aig_ManObjNumMax(pAig) );
    p.pOrderL2= ABC_CALLOC( int, Aig_ManObjNumMax(pAig) );
    p.pOrderG = ABC_CALLOC( int, Aig_ManObjNumMax(pAig) );
    p.pVars2Q = ABC_CALLOC( int, Aig_ManObjNumMax(pAig) );
    Aig_ManForEachCi( pAig, pObj, i )
        p.pVars2Q[Aig_ObjId(pObj)] = 1;
    for  i = 0; i < Aig_ManObjNumMax(pAig); i++  {
        p.pOrderL[i] = p.pOrderL2[i] = p.pOrderG[i] = i;
    }
    Llb_NonlinPrepareVarMap( p ); 
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_MnnStop(p *Llb_Mnn_t) {
    var bTemp *DdNode;
    var i int32;
    if p.pPars.fVerbose {
        p.timeOther = p.timeTotal - p.timeImage - p.timeTran1 - p.timeTran2 - p.timeGloba;
        p.timeReoG  = Cudd_ReadReorderingTime(p.ddG);
        ABC_PRTP( "Image    ", p.timeImage, p.timeTotal );
        ABC_PRTP( "  build  ",    timeBuild, p.timeTotal );
        ABC_PRTP( "  and-ex ",    timeAndEx, p.timeTotal );
        ABC_PRTP( "  other  ",    timeOther, p.timeTotal );
        ABC_PRTP( "Transfer1", p.timeTran1, p.timeTotal );
        ABC_PRTP( "Transfer2", p.timeTran2, p.timeTotal );
        ABC_PRTP( "Global   ", p.timeGloba, p.timeTotal );
        ABC_PRTP( "Other    ", p.timeOther, p.timeTotal );
        ABC_PRTP( "TOTAL    ", p.timeTotal, p.timeTotal );
        ABC_PRTP( "  reo    ", p.timeReo,   p.timeTotal );
        ABC_PRTP( "  reoG   ", p.timeReoG,  p.timeTotal );
    }
    if p.ddR.bFunc {
        Cudd_RecursiveDeref( p.ddR, p.ddR.bFunc ); }
    Vec_PtrForEachEntry( DdNode *, p.vRings, bTemp, i )
        Cudd_RecursiveDeref( p.ddR, bTemp );
    Vec_PtrFree( p.vRings );
    if p.ddG.bFunc {
        Cudd_RecursiveDeref( p.ddG, p.ddG.bFunc ); }
    if p.ddG.bFunc2 {
        Cudd_RecursiveDeref( p.ddG, p.ddG.bFunc2 ); }
//    printf( "manager1\n" );
//    Extra_StopManager( p->dd );
//    printf( "manager2\n" );
    Extra_StopManager( p.ddG );
//    printf( "manager3\n" );
    Extra_StopManager( p.ddR );
    Vec_IntFreeP( &p.vCs2Glo );
    Vec_IntFreeP( &p.vNs2Glo );
    Vec_IntFreeP( &p.vGlo2Cs );
    Vec_IntFreeP( &p.vGlo2Ns );
    Vec_PtrFree( p.vLeaves );
    Vec_PtrFree( p.vRoots );
    ABC_FREE( p.pVars2Q );
    ABC_FREE( p.pOrderL );
    ABC_FREE( p.pOrderL2 );
    ABC_FREE( p.pOrderG );
    ABC_FREE( p );
}


/**Function*************************************************************

  Synopsis    [Finds balanced cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinExperiment(pAig *Aig_Man_t, Num int32) {
    var pMnn *Llb_Mnn_t;
    var Pars Gia_ParLlb_t;
    var pPars *Gia_ParLlb_t = &Pars;
    var p *Aig_Man_t;
    var clk abctime  = Abc_Clock();

    Llb_ManSetDefaultParams( pPars );
    pPars.fVerbose = 1;

    p = Aig_ManDupFlopsOnly( pAig );
//Aig_ManShow( p, 0, NULL );
    Aig_ManPrintStats( pAig );
    Aig_ManPrintStats( p );

    pMnn = Llb_MnnStart( pAig, p, pPars );
    Llb_NonlinReachability( pMnn );
    pMnn.timeTotal = Abc_Clock() - clk;
    Llb_MnnStop( pMnn );

    Aig_ManStop( p );
}

/**Function*************************************************************

  Synopsis    [Finds balanced cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinCoreReach(pAig *Aig_Man_t, pPars *Gia_ParLlb_t) int32 {
    var pMnn *Llb_Mnn_t;
    var p *Aig_Man_t;
    var RetValue int32  = -1;

    p = Aig_ManDupFlopsOnly( pAig );
//Aig_ManShow( p, 0, NULL );
    if pPars.fVerbose {
    Aig_ManPrintStats( pAig ); }
    if pPars.fVerbose {
    Aig_ManPrintStats( p ); }

    if !pPars.fSkipReach {
        var clk abctime  = Abc_Clock();
        pMnn = Llb_MnnStart( pAig, p, pPars );
        RetValue = Llb_NonlinReachability( pMnn );
        pMnn.timeTotal = Abc_Clock() - clk;
        Llb_MnnStop( pMnn );
    }

    Aig_ManStop( p );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

