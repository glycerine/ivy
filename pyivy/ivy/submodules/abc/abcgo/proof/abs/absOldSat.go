package abs

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [saigRefSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [SAT based refinement of a counter-example.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigRefSat.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abs.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// local manager
   
type Saig_RefMan_t struct {
    // user data
    pAig  *Aig_Man_t;       // user's AIG
    pCex  *Abc_Cex_t;       // user's CEX
    nInputs         int32;    // the number of first inputs to skip
    fVerbose         int32;   // verbose flag
    // unrolling
    pFrames  *Aig_Man_t;    // unrolled timeframes
    vMapPiF2A  *Vec_Int_t;  // mapping of frame PIs into real PIs
}

// performs ternary simulation
// extern int Saig_ManSimDataInit( Aig_Man_t * p, Abc_Cex_t * pCex, Vec_Ptr_t * vSimInfo, Vec_Int_t * vRes );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Maps array of frame PI IDs into array of original PI IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_RefManReason2Inputs(p *Saig_RefMan_t, vReasons *Vec_Int_t) *Vec_Int_t {
    var vOriginal, vVisited *Vec_Int_t;
    var i, Entry int32;
    vOriginal = Vec_IntAlloc( Saig_ManPiNum(p.pAig) ); 
    vVisited = Vec_IntStart( Saig_ManPiNum(p.pAig) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vReasons)) && (((Entry) = Vec_IntEntry(vReasons, i)), 1); i++  {
        var iInput int32  = Vec_IntEntry( p.vMapPiF2A, 2*Entry );
        assert(iInput >= 0 && iInput < Aig_ManCiNum(p.pAig), "iInput >= 0 && iInput < Aig_ManCiNum(p.pAig)");
        if Vec_IntEntry(vVisited, iInput) == 0 {
            Vec_IntPush( vOriginal, iInput ); }
        Vec_IntAddToEntry( vVisited, iInput, 1 );
    }
    Vec_IntFree( vVisited );
    return vOriginal;
}

/**Function*************************************************************

  Synopsis    [Creates the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_RefManReason2Cex(p *Saig_RefMan_t, vReasons *Vec_Int_t) *Abc_Cex_t {
    var pCare *Abc_Cex_t;
    var i, Entry, iInput, iFrame int32;
    pCare = Abc_CexDup( p.pCex, p.pCex.nRegs );
    memset( pCare.pData, 0, sizeof(unsigned) * Abc_BitWordNum(pCare.nBits) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vReasons)) && (((Entry) = Vec_IntEntry(vReasons, i)), 1); i++  {
        assert(Entry >= 0 && Entry < Aig_ManCiNum(p.pFrames), "Entry >= 0 && Entry < Aig_ManCiNum(p.pFrames)");
        iInput = Vec_IntEntry( p.vMapPiF2A, 2*Entry );
        iFrame = Vec_IntEntry( p.vMapPiF2A, 2*Entry+1 );
        Abc_InfoSetBit( pCare.pData, pCare.nRegs + pCare.nPis * iFrame + iInput );
    }
    return pCare;
}

/**Function*************************************************************

  Synopsis    [Returns reasons for the property to fail.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_RefManFindReason_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vPrios *Vec_Int_t, vReasons *Vec_Int_t) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Aig_ObjIsCi(pObj) {
        Vec_IntPush( vReasons, Aig_ObjCioId(pObj) );
        return;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    if pObj.fPhase {
        Saig_RefManFindReason_rec( p, Aig_ObjFanin0(pObj), vPrios, vReasons );
        Saig_RefManFindReason_rec( p, Aig_ObjFanin1(pObj), vPrios, vReasons );
    
    } else {
        var fPhase0 int32  = Aig_ObjFaninC0(pObj) ^ Aig_ObjFanin0(pObj).fPhase;
        var fPhase1 int32  = Aig_ObjFaninC1(pObj) ^ Aig_ObjFanin1(pObj).fPhase;
        assert(!(fPhase0 != 0) || !(fPhase1 != 0), "!fPhase0 || !fPhase1");
        if !(fPhase0 != 0) && (fPhase1 != 0) {
            Saig_RefManFindReason_rec( p, Aig_ObjFanin0(pObj), vPrios, vReasons );
        } else if (fPhase0 != 0) && !(fPhase1 != 0) {
            Saig_RefManFindReason_rec( p, Aig_ObjFanin1(pObj), vPrios, vReasons );
        } else {
            var iPrio0 int32  = Vec_IntEntry( vPrios, Aig_ObjFaninId0(pObj) );
            var iPrio1 int32  = Vec_IntEntry( vPrios, Aig_ObjFaninId1(pObj) );
            if iPrio0 <= iPrio1 {
                Saig_RefManFindReason_rec( p, Aig_ObjFanin0(pObj), vPrios, vReasons );
            } else {
                Saig_RefManFindReason_rec( p, Aig_ObjFanin1(pObj), vPrios, vReasons ); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Returns reasons for the property to fail.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_RefManFindReason(p *Saig_RefMan_t) *Vec_Int_t {
    var pObj *Aig_Obj_t;
    var vPrios, vPi2Prio, vReasons *Vec_Int_t;
    var i, CountPrios int32;

    vPi2Prio = Vec_IntStartFull( Saig_ManPiNum(p.pAig) );
    vPrios   = Vec_IntStartFull( Aig_ManObjNumMax(p.pFrames) );

    // set PI values according to CEX
    CountPrios = 0;
    Aig_ManConst1(p.pFrames).fPhase = 1;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pFrames.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pFrames.vCis, i)), 1); i++  {
        var iInput int32  = Vec_IntEntry( p.vMapPiF2A, 2*i );
        var iFrame int32  = Vec_IntEntry( p.vMapPiF2A, 2*i+1 );
        pObj.fPhase = Abc_InfoHasBit( p.pCex.pData, p.pCex.nRegs + p.pCex.nPis * iFrame + iInput );
        // assign priority
        if Vec_IntEntry(vPi2Prio, iInput) == ~0 {
            Vec_IntWriteEntry( vPi2Prio, iInput, CountPrios++ ); }
//        Vec_IntWriteEntry( vPrios, Aig_ObjId(pObj), Vec_IntEntry(vPi2Prio, iInput) );
        Vec_IntWriteEntry( vPrios, Aig_ObjId(pObj), i );
    }
//    printf( "Priority numbers = %d.\n", CountPrios );
    Vec_IntFree( vPi2Prio );

    // traverse and set the priority
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pFrames.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pFrames.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        var fPhase0 int32  = Aig_ObjFaninC0(pObj) ^ Aig_ObjFanin0(pObj).fPhase;
        var fPhase1 int32  = Aig_ObjFaninC1(pObj) ^ Aig_ObjFanin1(pObj).fPhase;
        var iPrio0 int32   = Vec_IntEntry( vPrios, Aig_ObjFaninId0(pObj) );
        var iPrio1 int32   = Vec_IntEntry( vPrios, Aig_ObjFaninId1(pObj) );
        pObj.fPhase = fPhase0 && fPhase1;
        if (fPhase0 != 0) && (fPhase1 != 0) { // both are one
            Vec_IntWriteEntry( vPrios, Aig_ObjId(pObj), Abc_MaxInt(iPrio0, iPrio1) );
        } else if !(fPhase0 != 0) && (fPhase1 != 0) { 
            Vec_IntWriteEntry( vPrios, Aig_ObjId(pObj), iPrio0 );
        } else if (fPhase0 != 0) && !(fPhase1 != 0) {
            Vec_IntWriteEntry( vPrios, Aig_ObjId(pObj), iPrio1 );
        } else { // both are zero
            Vec_IntWriteEntry( vPrios, Aig_ObjId(pObj), Abc_MinInt(iPrio0, iPrio1) ); }
    }
    }
    // check the property output
    pObj = Aig_ManCo( p.pFrames, 0 );
    assert(int32(Aig_ObjFanin0(pObj).fPhase) == Aig_ObjFaninC0(pObj), "int32(Aig_ObjFanin0(pObj).fPhase) == Aig_ObjFaninC0(pObj)");

    // select the reason
    vReasons = Vec_IntAlloc( 100 );
    Aig_ManIncrementTravId( p.pFrames );
    if !Aig_ObjIsConst1(Aig_ObjFanin0(pObj)) {
        Saig_RefManFindReason_rec( p.pFrames, Aig_ObjFanin0(pObj), vPrios, vReasons ); }
    Vec_IntFree( vPrios );
    return vReasons;
}


/**Function*************************************************************

  Synopsis    [Collect nodes in the unrolled timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManUnrollCollect_rec(pAig *Aig_Man_t, pObj *Aig_Obj_t, vObjs *Vec_Int_t, vRoots *Vec_Int_t) {
    if Aig_ObjIsTravIdCurrent(pAig, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(pAig, pObj);
    if Aig_ObjIsCo(pObj) {
        Saig_ManUnrollCollect_rec( pAig, Aig_ObjFanin0(pObj), vObjs, vRoots );
    } else if Aig_ObjIsNode(pObj) {
        Saig_ManUnrollCollect_rec( pAig, Aig_ObjFanin0(pObj), vObjs, vRoots );
        Saig_ManUnrollCollect_rec( pAig, Aig_ObjFanin1(pObj), vObjs, vRoots );
    }
    if (vRoots != nil) && Saig_ObjIsLo( pAig, pObj ) {
        Vec_IntPush( vRoots, Aig_ObjId( Saig_ObjLoToLi(pAig, pObj) ) ); }
    Vec_IntPush( vObjs, Aig_ObjId(pObj) );
}

/**Function*************************************************************

  Synopsis    [Derive unrolled timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManUnrollWithCex(pAig *Aig_Man_t, pCex *Abc_Cex_t, nInputs int32, pvMapPiF2A **Vec_Int_t) *Aig_Man_t {
    var pFrames *Aig_Man_t;     // unrolled timeframes
    var vFrameCos *Vec_Vec_t;   // the list of COs per frame
    var vFrameObjs *Vec_Vec_t;  // the list of objects per frame
    var vRoots, vObjs *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i, f int32;
    // sanity checks
    assert(Saig_ManPiNum(pAig) == pCex.nPis, "Saig_ManPiNum(pAig) == pCex.nPis");
    assert(Saig_ManRegNum(pAig) == pCex.nRegs, "Saig_ManRegNum(pAig) == pCex.nRegs");
    assert(pCex.iPo >= 0 && pCex.iPo < Saig_ManPoNum(pAig), "pCex.iPo >= 0 && pCex.iPo < Saig_ManPoNum(pAig)");

    // map PIs of the unrolled frames into PIs of the original design
    *pvMapPiF2A = Vec_IntAlloc( 1000 );

    // collect COs and Objs visited in each frame
    vFrameCos  = Vec_VecStart( pCex.iFrame+1 );
    vFrameObjs = Vec_VecStart( pCex.iFrame+1 );
    // initialized the topmost frame
    pObj = Aig_ManCo( pAig, pCex.iPo );
    Vec_VecPushInt( vFrameCos, pCex.iFrame, Aig_ObjId(pObj) );
    for  f = pCex.iFrame; f >= 0; f--  {
        // collect nodes starting from the roots
        Aig_ManIncrementTravId( pAig );
        vRoots = Vec_VecEntryInt( vFrameCos, f );
        // inlining of macro Aig_ManForEachObjVec
        for  i = 0; i < Vec_IntSize(vRoots) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vRoots,i))), 1); i++  {
            Saig_ManUnrollCollect_rec( pAig, pObj, 
                Vec_VecEntryInt(vFrameObjs, f),
                (Vec_Int_t *)(f ? Vec_VecEntry(vFrameCos, f-1) : nil) );
        }
    }

    // derive unrolled timeframes
    pFrames = Aig_ManStart( 10000 );
    pFrames.pName = Abc_UtilStrsav( pAig.pName );
    pFrames.pSpec = Abc_UtilStrsav( pAig.pSpec );
    // initialize the flops 
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.pData = Aig_NotCond( Aig_ManConst1(pFrames), !Abc_InfoHasBit(pCex.pData, i) );
    }
    // iterate through the frames
    for  f = 0; f <= pCex.iFrame; f++  {
        // construct
        vObjs = Vec_VecEntryInt( vFrameObjs, f );
        // inlining of macro Aig_ManForEachObjVec
        for  i = 0; i < Vec_IntSize(vObjs) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vObjs,i))), 1); i++  {
            if Aig_ObjIsNode(pObj) {
                pObj.pData = Aig_And( pFrames, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
            } else if Aig_ObjIsCo(pObj) {
                pObj.pData = Aig_ObjChild0Copy(pObj);
            } else if Aig_ObjIsConst1(pObj) {
                pObj.pData = Aig_ManConst1(pFrames);
            } else if Saig_ObjIsPi(pAig, pObj) {
                if Aig_ObjCioId(pObj) < nInputs {
                    var iBit int32  = pCex.nRegs + f * pCex.nPis + Aig_ObjCioId(pObj);
                    pObj.pData = Aig_NotCond( Aig_ManConst1(pFrames), !Abc_InfoHasBit(pCex.pData, iBit) );
                
                } else {
                    pObj.pData = Aig_ObjCreateCi( pFrames );
                    Vec_IntPush( *pvMapPiF2A, Aig_ObjCioId(pObj) );
                    Vec_IntPush( *pvMapPiF2A, f );
                }
            }
        }
        if f == pCex.iFrame {
            break; }
        // transfer
        vRoots = Vec_VecEntryInt( vFrameCos, f );
        // inlining of macro Aig_ManForEachObjVec
        for  i = 0; i < Vec_IntSize(vRoots) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vRoots,i))), 1); i++  {
            Saig_ObjLiToLo( pAig, pObj ).pData = pObj.pData;
        }
    }
    // create output
    pObj = Aig_ManCo( pAig, pCex.iPo );
    Aig_ObjCreateCo( pFrames, Aig_Not((Aig_Obj_t *)pObj.pData) );
    Aig_ManSetRegNum( pFrames, 0 );
    // cleanup
    Vec_VecFree( vFrameCos );
    Vec_VecFree( vFrameObjs );
    // finallize
    Aig_ManCleanup( pFrames );
    // return
    return pFrames;
}

/**Function*************************************************************

  Synopsis    [Creates refinement manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_RefManStart(pAig *Aig_Man_t, pCex *Abc_Cex_t, nInputs int32, fVerbose int32) *Saig_RefMan_t {
    
    // inlining of macro ABC_CALLOC
    p := &Saig_RefMan_t{};
    p.pAig = pAig;
    p.pCex = pCex;
    p.nInputs = nInputs;
    p.fVerbose = fVerbose;
    p.pFrames = Saig_ManUnrollWithCex( pAig, pCex, nInputs, &p.vMapPiF2A );
    return p;
}

/**Function*************************************************************

  Synopsis    [Destroys refinement manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_RefManStop(p *Saig_RefMan_t) {
    Aig_ManStopP( &p.pFrames );
    Vec_IntFreeP( &p.vMapPiF2A );
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Sets phase bits in the timeframe AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_RefManSetPhases(p *Saig_RefMan_t, pCare *Abc_Cex_t, fValue1 int32) int32 {
    var pObj *Aig_Obj_t;
    var i, iFrame, iInput int32;
    Aig_ManConst1( p.pFrames ).fPhase = 1;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pFrames.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pFrames.vCis, i)), 1); i++  {
        iInput = Vec_IntEntry( p.vMapPiF2A, 2*i );
        iFrame = Vec_IntEntry( p.vMapPiF2A, 2*i+1 );
        pObj.fPhase = Abc_InfoHasBit( p.pCex.pData, p.pCex.nRegs + p.pCex.nPis * iFrame + iInput );
        // update value if it is a don't-care
        if (pCare != nil) && !Abc_InfoHasBit( pCare.pData, p.pCex.nRegs + p.pCex.nPis * iFrame + iInput ) {
            pObj.fPhase = fValue1; }
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pFrames.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pFrames.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.fPhase = ( Aig_ObjFanin0(pObj).fPhase ^ Aig_ObjFaninC0(pObj) )
                     & ( Aig_ObjFanin1(pObj).fPhase ^ Aig_ObjFaninC1(pObj) ); }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pFrames.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pFrames.vCos, i)), 1); i++  {
        pObj.fPhase = ( Aig_ObjFanin0(pObj).fPhase ^ Aig_ObjFaninC0(pObj) );
    }
    pObj = Aig_ManCo( p.pFrames, 0 );
    return pObj.fPhase;
}

/**Function*************************************************************

  Synopsis    [Tries to remove literals from abstraction.]

  Description [The literals are sorted more desirable first.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_RefManOrderLiterals(p *Saig_RefMan_t, vVar2PiId *Vec_Int_t, vAssumps *Vec_Int_t) *Vec_Vec_t {
    var vLits *Vec_Vec_t;
    var vVar2New *Vec_Int_t;
    var i, Entry, iInput, iFrame int32;
    // collect literals
    vLits = Vec_VecAlloc( 100 );
    vVar2New = Vec_IntStartFull( Saig_ManPiNum(p.pAig) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vAssumps)) && (((Entry) = Vec_IntEntry(vAssumps, i)), 1); i++  {
        var iPiNum int32  = Vec_IntEntry( vVar2PiId, lit_var(Entry) );
        assert(iPiNum >= 0 && iPiNum < Aig_ManCiNum(p.pFrames), "iPiNum >= 0 && iPiNum < Aig_ManCiNum(p.pFrames)");
        iInput = Vec_IntEntry( p.vMapPiF2A, 2*iPiNum );
        iFrame = Vec_IntEntry( p.vMapPiF2A, 2*iPiNum+1 );
//        Abc_InfoSetBit( pCare->pData, pCare->nRegs + pCare->nPis * iFrame + iInput );
        if Vec_IntEntry( vVar2New, iInput ) == ~0 {
            Vec_IntWriteEntry( vVar2New, iInput, Vec_VecSize(vLits) ); }
        Vec_VecPushInt( vLits, Vec_IntEntry( vVar2New, iInput ), Entry );
    }
    Vec_IntFree( vVar2New );
    return vLits;
}


/**Function*************************************************************

  Synopsis    [Generate the care set using SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_RefManCreateCex(p *Saig_RefMan_t, vVar2PiId *Vec_Int_t, vAssumps *Vec_Int_t) *Abc_Cex_t {
    var pCare *Abc_Cex_t;
    var i, Entry, iInput, iFrame int32;
    // create counter-example
    pCare = Abc_CexDup( p.pCex, p.pCex.nRegs );
    memset( pCare.pData, 0, sizeof(unsigned) * Abc_BitWordNum(pCare.nBits) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vAssumps)) && (((Entry) = Vec_IntEntry(vAssumps, i)), 1); i++  {
        var iPiNum int32  = Vec_IntEntry( vVar2PiId, lit_var(Entry) );
        assert(iPiNum >= 0 && iPiNum < Aig_ManCiNum(p.pFrames), "iPiNum >= 0 && iPiNum < Aig_ManCiNum(p.pFrames)");
        iInput = Vec_IntEntry( p.vMapPiF2A, 2*iPiNum );
        iFrame = Vec_IntEntry( p.vMapPiF2A, 2*iPiNum+1 );
        Abc_InfoSetBit( pCare.pData, pCare.nRegs + pCare.nPis * iFrame + iInput );
    }
    return pCare;
}

/**Function*************************************************************

  Synopsis    [Generate the care set using SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_RefManRunSat(p *Saig_RefMan_t, fNewOrder int32) *Abc_Cex_t {
    var nConfLimit int32  = 1000000;
    var pCare *Abc_Cex_t;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var pObj *Aig_Obj_t;
    var vLits *Vec_Vec_t  = nil;
    var vAssumps, vVar2PiId *Vec_Int_t;
    var i, k, Entry, RetValue int32;//, f = 0, Counter = 0;
    var nCoreLits int32;
    var pCoreLits *int32;
    var clk abctime  = Abc_Clock();
    // create CNF
    assert(Aig_ManRegNum(p.pFrames) == 0, "Aig_ManRegNum(p.pFrames) == 0");
//    pCnf = Cnf_Derive( p->pFrames, 0 ); // too slow
    pCnf = Cnf_DeriveSimple( p.pFrames, 0 );
    RetValue = Saig_RefManSetPhases( p, nil, 0 );
    if RetValue != 0 {
        fmt.Printf("Constructed frames are incorrect.\n");
        Cnf_DataFree( pCnf );
        return nil;
    }
    Cnf_DataTranformPolarity( pCnf, 0 );
    // create SAT solver
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    if pSat == nil {
        Cnf_DataFree( pCnf );
        return nil;
    }
//Abc_PrintTime( 1, "Preparing", Abc_Clock() - clk );
    // look for a true counter-example
    if p.nInputs > 0 {
        RetValue = sat_solver_solve( pSat, nil, nil, 
            (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if RetValue == l_False {
            fmt.Printf("The problem is trivially UNSAT. The CEX is real.\n");
            // create counter-example
            pCare = Abc_CexDup( p.pCex, p.pCex.nRegs );
            memset( pCare.pData, 0, sizeof(unsigned) * Abc_BitWordNum(pCare.nBits) );
            return pCare;
        }
        // the problem is SAT - it is expected
    }
    // create assumptions
    vVar2PiId = Vec_IntStartFull( pCnf.nVars );
    vAssumps = Vec_IntAlloc( Aig_ManCiNum(p.pFrames) );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pFrames.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pFrames.vCis, i)), 1); i++  {
//        RetValue = Abc_InfoHasBit( p->pCex->pData, p->pCex->nRegs + p->pCex->nPis * iFrame + iInput );
//        Vec_IntPush( vAssumps, toLitCond( pCnf->pVarNums[Aig_ObjId(pObj)], !RetValue ) );
        Vec_IntPush( vAssumps, toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], 1 ) );
        Vec_IntWriteEntry( vVar2PiId, pCnf.pVarNums[Aig_ObjId(pObj)], i );
    }

    // reverse the order of assumptions
//    if ( fNewOrder )
//    Vec_IntReverseOrder( vAssumps );

    if fNewOrder != 0 {
        // create literals
        vLits = Saig_RefManOrderLiterals( p, vVar2PiId, vAssumps );
        // sort literals
        Vec_VecSort( vLits, 1 );
        // save literals
        Vec_IntClear( vAssumps );
        // inlining of macro Vec_VecForEachEntryInt
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; i < Vec_VecSize(vLits); i++  {                                                        for  k = 0; (k < Vec_IntSize(Vec_VecEntryInt(vLits, i))) && (((Entry) = Vec_IntEntry(Vec_VecEntryInt(vLits, i), k)), 1); k++  { 
            Vec_IntPush( vAssumps, Entry );
        }
        }

        for  i = 0; i < Vec_VecSize(vLits); i++  {
            fmt.Printf("%d ", Vec_IntSize( Vec_VecEntryInt(vLits, i) ));
        }
        fmt.Printf("\n");

        if p.fVerbose {
            fmt.Printf("Total PIs = %d. Essential PIs = %d.\n", Saig_ManPiNum(p.pAig) - p.nInputs, Vec_VecSize(vLits)); }
    }

    // solve
clk = Abc_Clock();
    RetValue = sat_solver_solve( pSat, Vec_IntArray(vAssumps), Vec_IntArray(vAssumps) + Vec_IntSize(vAssumps), 
        (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
//Abc_PrintTime( 1, "Solving", Abc_Clock() - clk );
    if RetValue != l_False {
        if RetValue == l_True {
            fmt.Printf("Internal Error!!! The resulting problem is SAT.\n");
        } else {
            fmt.Printf("Internal Error!!! SAT solver timed out.\n"); }
        Cnf_DataFree( pCnf );
        sat_solver_delete( pSat );
        Vec_IntFree( vAssumps );
        Vec_IntFree( vVar2PiId );
        return nil;
    }
    assert(RetValue == l_False, "RetValue == l_False"); // UNSAT

    // get relevant SAT literals
    nCoreLits = sat_solver_final( pSat, &pCoreLits );
    assert(nCoreLits > 0, "nCoreLits > 0");
    if p.fVerbose {
    fmt.Printf("AnalizeFinal selected %d assumptions (out of %d). Conflicts = %d.\n", nCoreLits, Vec_IntSize(vAssumps), int32(pSat.stats.conflicts)); }

    // save literals
    Vec_IntClear( vAssumps );
    for  i = 0; i < nCoreLits; i++  {
        Vec_IntPush( vAssumps, pCoreLits[i] );
    }


    // create literals
    vLits = Saig_RefManOrderLiterals( p, vVar2PiId, vAssumps );
    // sort literals
//    Vec_VecSort( vLits, 0 );
    // save literals
    Vec_IntClear( vAssumps );
    // inlining of macro Vec_VecForEachEntryInt
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; i < Vec_VecSize(vLits); i++  {                                                        for  k = 0; (k < Vec_IntSize(Vec_VecEntryInt(vLits, i))) && (((Entry) = Vec_IntEntry(Vec_VecEntryInt(vLits, i), k)), 1); k++  { 
        Vec_IntPush( vAssumps, Entry );
    }
    }

//    for ( i = 0; i < Vec_VecSize(vLits); i++ )
//        printf( "%d ", Vec_IntSize( Vec_VecEntryInt(vLits, i) ) );
//    printf( "\n" );

    if p.fVerbose {
        fmt.Printf("Total PIs = %d. Essential PIs = %d.\n", Saig_ManPiNum(p.pAig) - p.nInputs, Vec_VecSize(vLits)); }
/*
    // try assumptions in different order
    RetValue = sat_solver_solve( pSat, Vec_IntArray(vAssumps), Vec_IntArray(vAssumps) + Vec_IntSize(vAssumps), 
        (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    printf( "Assumpts = %2d. Intermediate instance is %5s. Conflicts = %2d.\n", 
        Vec_IntSize(vAssumps), (RetValue == l_False ? "UNSAT" : "SAT"), (int)pSat->stats.conflicts );

    // create different sets of assumptions
    Counter = Vec_VecSize(vLits);
    for ( f = 0; f < Vec_VecSize(vLits); f++ )
    {
        Vec_IntClear( vAssumps );
        Vec_VecForEachEntryInt( vLits, Entry, i, k )
            if ( i != f )
                Vec_IntPush( vAssumps, Entry );

        // try the new assumptions
        RetValue = sat_solver_solve( pSat, Vec_IntArray(vAssumps), Vec_IntArray(vAssumps) + Vec_IntSize(vAssumps), 
            (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        printf( "Assumpts = %2d. Intermediate instance is %5s. Conflicts = %2d.\n", 
            Vec_IntSize(vAssumps), RetValue == l_False ? "UNSAT" : "SAT", (int)pSat->stats.conflicts );
        if ( RetValue != l_False )
            continue;

        // UNSAT - remove literals
        Vec_IntClear( Vec_VecEntryInt(vLits, f) );
        Counter--;
    }

    for ( i = 0; i < Vec_VecSize(vLits); i++ )
        printf( "%d ", Vec_IntSize( Vec_VecEntryInt(vLits, i) ) );
    printf( "\n" );

    if ( p->fVerbose )
        printf( "Total PIs = %d. Essential PIs = %d.\n", 
            Saig_ManPiNum(p->pAig) - p->nInputs, Counter );

    // save literals
    Vec_IntClear( vAssumps );
    Vec_VecForEachEntryInt( vLits, Entry, i, k )
        Vec_IntPush( vAssumps, Entry );
*/
    // create counter-example
    pCare = Saig_RefManCreateCex( p, vVar2PiId, vAssumps );

    // cleanup
    Cnf_DataFree( pCnf );
    sat_solver_delete( pSat );
    Vec_IntFree( vAssumps );
    Vec_IntFree( vVar2PiId );
    Vec_VecFreeP( &vLits );

    // verify counter-example
    RetValue = Saig_RefManSetPhases( p, pCare, 0 );
    if RetValue != 0 {
        fmt.Printf("Reduced CEX verification has failed.\n"); }
    RetValue = Saig_RefManSetPhases( p, pCare, 1 );
    if RetValue != 0 {
        fmt.Printf("Reduced CEX verification has failed.\n"); }
    return pCare;
}


/**Function*************************************************************

  Synopsis    []

  Description []               

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_RefManRefineWithSat(p *Saig_RefMan_t, vAigPis *Vec_Int_t) *Vec_Int_t {
    var nConfLimit int32  = 1000000;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var pObj *Aig_Obj_t;
    var vLits *Vec_Vec_t;
    var vReasons, vAssumps, vVisited, vVar2PiId *Vec_Int_t;
    var i, k, f, Entry, RetValue, Counter int32;

    // create CNF and SAT solver
    pCnf = Cnf_DeriveSimple( p.pFrames, 0 );
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    if pSat == nil {
        Cnf_DataFree( pCnf );
        return nil;
    }

    // mark used AIG inputs
    vVisited = Vec_IntStart( Saig_ManPiNum(p.pAig) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vAigPis)) && (((Entry) = Vec_IntEntry(vAigPis, i)), 1); i++  {
        assert(Entry >= 0 && Entry < Aig_ManCiNum(p.pAig), "Entry >= 0 && Entry < Aig_ManCiNum(p.pAig)");
        Vec_IntWriteEntry( vVisited, Entry, 1 );
    }

    // create assumptions
    vVar2PiId = Vec_IntStartFull( pCnf.nVars );
    vAssumps = Vec_IntAlloc( Aig_ManCiNum(p.pFrames) );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pFrames.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pFrames.vCis, i)), 1); i++  {
        var iInput int32  = Vec_IntEntry( p.vMapPiF2A, 2*i );
        var iFrame int32  = Vec_IntEntry( p.vMapPiF2A, 2*i+1 );
        if Vec_IntEntry(vVisited, iInput) == 0 {
            continue; }
        RetValue = Abc_InfoHasBit( p.pCex.pData, p.pCex.nRegs + p.pCex.nPis * iFrame + iInput );
        Vec_IntPush( vAssumps, toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], !RetValue ) );
//        Vec_IntPush( vAssumps, toLitCond( pCnf->pVarNums[Aig_ObjId(pObj)], 1 ) );
        Vec_IntWriteEntry( vVar2PiId, pCnf.pVarNums[Aig_ObjId(pObj)], i );
    }
    Vec_IntFree( vVisited );

    // try assumptions in different order
    RetValue = sat_solver_solve( pSat, Vec_IntArray(vAssumps), Vec_IntArray(vAssumps) + Vec_IntSize(vAssumps), 
        (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    fmt.Printf("Assumpts = %2d. Intermediate instance is %5s. Conflicts = %2d.\n", Vec_IntSize(vAssumps), (RetValue == l_False ? "UNSAT" : "SAT"), int32(pSat.stats.conflicts));

/*
    // AnalizeFinal does not work because it implications propagate directly
    // and SAT solver does not kick in (the number of conflicts in 0).

    // count the number of lits in the unsat core
    {
        int nCoreLits, * pCoreLits;
        nCoreLits = sat_solver_final( pSat, &pCoreLits );
        assert( nCoreLits > 0 );

        // count the number of flops
        vVisited = Vec_IntStart( Saig_ManPiNum(p->pAig) );
        for ( i = 0; i < nCoreLits; i++ )
        {
            int iPiNum = Vec_IntEntry( vVar2PiId, lit_var(pCoreLits[i]) );
            int iInput = Vec_IntEntry( p->vMapPiF2A, 2*iPiNum );
            int iFrame = Vec_IntEntry( p->vMapPiF2A, 2*iPiNum+1 );
            Vec_IntWriteEntry( vVisited, iInput, 1 );
        }
        // count the number of entries
        Counter = 0;
        Vec_IntForEachEntry( vVisited, Entry, i )
            Counter += Entry;
        Vec_IntFree( vVisited );

//        if ( p->fVerbose )
        printf( "AnalizeFinal: Assumptions %d (out of %d). Essential PIs = %d. Conflicts = %d.\n", 
            nCoreLits, Vec_IntSize(vAssumps), Counter, (int)pSat->stats.conflicts );
    }
*/

    // derive literals
    vLits = Saig_RefManOrderLiterals( p, vVar2PiId, vAssumps );
    for  i = 0; i < Vec_VecSize(vLits); i++  {
        fmt.Printf("%d ", Vec_IntSize( Vec_VecEntryInt(vLits, i) ));
    }
    fmt.Printf("\n");

    // create different sets of assumptions
    Counter = Vec_VecSize(vLits);
    for  f = 0; f < Vec_VecSize(vLits); f++  {
        Vec_IntClear( vAssumps );
        // inlining of macro Vec_VecForEachEntryInt
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; i < Vec_VecSize(vLits); i++  {                                                        for  k = 0; (k < Vec_IntSize(Vec_VecEntryInt(vLits, i))) && (((Entry) = Vec_IntEntry(Vec_VecEntryInt(vLits, i), k)), 1); k++  { 
            if i != f {
                Vec_IntPush( vAssumps, Entry ); }
        }
        }

        // try the new assumptions
        RetValue = sat_solver_solve( pSat, Vec_IntArray(vAssumps), Vec_IntArray(vAssumps) + Vec_IntSize(vAssumps), 
            (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        fmt.Printf("Assumpts = %2d. Intermediate instance is %5s. Conflicts = %2d.\n", Vec_IntSize(vAssumps), RetValue == l_False ? "UNSAT" : "SAT", int32(pSat.stats.conflicts));
        if RetValue != l_False {
            continue; }

        // UNSAT - remove literals
        Vec_IntClear( Vec_VecEntryInt(vLits, f) );
        Counter--;
    }

    for  i = 0; i < Vec_VecSize(vLits); i++  {
        fmt.Printf("%d ", Vec_IntSize( Vec_VecEntryInt(vLits, i) ));
    }
    fmt.Printf("\n");

    // create assumptions
    Vec_IntClear( vAssumps );
    // inlining of macro Vec_VecForEachEntryInt
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; i < Vec_VecSize(vLits); i++  {                                                        for  k = 0; (k < Vec_IntSize(Vec_VecEntryInt(vLits, i))) && (((Entry) = Vec_IntEntry(Vec_VecEntryInt(vLits, i), k)), 1); k++  { 
        Vec_IntPush( vAssumps, Entry );
    }
    }

    // try assumptions in different order
    RetValue = sat_solver_solve( pSat, Vec_IntArray(vAssumps), Vec_IntArray(vAssumps) + Vec_IntSize(vAssumps), 
        (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    fmt.Printf("Assumpts = %2d. Intermediate instance is %5s. Conflicts = %2d.\n", Vec_IntSize(vAssumps), (RetValue == l_False ? "UNSAT" : "SAT"), int32(pSat.stats.conflicts));

//    if ( p->fVerbose )
//        printf( "Total PIs = %d. Essential PIs = %d.\n", 
//            Saig_ManPiNum(p->pAig) - p->nInputs, Counter );


    // transform assumptions into reasons
    vReasons = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vAssumps)) && (((Entry) = Vec_IntEntry(vAssumps, i)), 1); i++  {
        var iPiNum int32  = Vec_IntEntry( vVar2PiId, lit_var(Entry) );
        assert(iPiNum >= 0 && iPiNum < Aig_ManCiNum(p.pFrames), "iPiNum >= 0 && iPiNum < Aig_ManCiNum(p.pFrames)");
        Vec_IntPush( vReasons, iPiNum );
    }

    // cleanup
    Cnf_DataFree( pCnf );
    sat_solver_delete( pSat );
    Vec_IntFree( vAssumps );
    Vec_IntFree( vVar2PiId );
    Vec_VecFreeP( &vLits );

    return vReasons;
}

/**Function*************************************************************

  Synopsis    [SAT-based refinement of the counter-example.]

  Description [The first parameter (nInputs) indicates how many first 
  primary inputs to skip without considering as care candidates.]
               

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManFindCexCareBits(pAig *Aig_Man_t, pCex *Abc_Cex_t, nInputs int32, fNewOrder int32, fVerbose int32) *Abc_Cex_t {
    var p *Saig_RefMan_t;
    var vReasons *Vec_Int_t;
    var pCare *Abc_Cex_t;
    var clk abctime  = Abc_Clock();

    clk = Abc_Clock();
    p = Saig_RefManStart( pAig, pCex, nInputs, fVerbose );
    vReasons = Saig_RefManFindReason( p );

if fVerbose != 0 {
Aig_ManPrintStats( p.pFrames ); }

//    if ( fVerbose )
    {
        var vRes *Vec_Int_t  = Saig_RefManReason2Inputs( p, vReasons );
        fmt.Printf("Frame PIs = %4d (essential = %4d)   AIG PIs = %4d (essential = %4d)   ", Aig_ManCiNum(p.pFrames), Vec_IntSize(vReasons), Saig_ManPiNum(p.pAig) - p.nInputs, Vec_IntSize(vRes));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;

        Vec_IntFree( vRes );

/*
        ////////////////////////////////////
        Vec_IntFree( vReasons );
        vReasons = Saig_RefManRefineWithSat( p, vRes );
        ////////////////////////////////////

        Vec_IntFree( vRes );
        vRes = Saig_RefManReason2Inputs( p, vReasons );
        printf( "Frame PIs = %4d (essential = %4d)   AIG PIs = %4d (essential = %4d)   ",
            Aig_ManCiNum(p->pFrames), Vec_IntSize(vReasons), 
            Saig_ManPiNum(p->pAig) - p->nInputs, Vec_IntSize(vRes) );

        Vec_IntFree( vRes );
ABC_PRT( "Time", Abc_Clock() - clk );
*/
    }

    pCare = Saig_RefManReason2Cex( p, vReasons );
    Vec_IntFree( vReasons );
    Saig_RefManStop( p );

if fVerbose != 0 {
Abc_CexPrintStats( pCex ); }
if fVerbose != 0 {
Abc_CexPrintStats( pCare ); }

    return pCare;
}

/**Function*************************************************************

  Synopsis    [Returns the array of PIs for flops that should not be absracted.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManExtendCounterExampleTest3(pAig *Aig_Man_t, iFirstFlopPi int32, pCex *Abc_Cex_t, fVerbose int32) *Vec_Int_t {
    var p *Saig_RefMan_t;
    var vRes, vReasons *Vec_Int_t;
    var clk abctime;
    if Saig_ManPiNum(pAig) != pCex.nPis {
        fmt.Printf("Saig_ManExtendCounterExampleTest3(): The PI count of AIG (%d) does not match that of cex (%d).\n", Aig_ManCiNum(pAig), pCex.nPis);
        return nil;
    }

clk = Abc_Clock();

    p = Saig_RefManStart( pAig, pCex, iFirstFlopPi, fVerbose );
    vReasons = Saig_RefManFindReason( p );
    vRes = Saig_RefManReason2Inputs( p, vReasons );

//    if ( fVerbose )
    {
        fmt.Printf("Frame PIs = %4d (essential = %4d)   AIG PIs = %4d (essential = %4d)   ", Aig_ManCiNum(p.pFrames), Vec_IntSize(vReasons), Saig_ManPiNum(p.pAig) - p.nInputs, Vec_IntSize(vRes));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }

/*
    ////////////////////////////////////
    Vec_IntFree( vReasons );
    vReasons = Saig_RefManRefineWithSat( p, vRes );
    ////////////////////////////////////

    // derive new result
    Vec_IntFree( vRes );
    vRes = Saig_RefManReason2Inputs( p, vReasons );
//    if ( fVerbose )
    {
        printf( "Frame PIs = %4d (essential = %4d)   AIG PIs = %4d (essential = %4d)   ",
            Aig_ManCiNum(p->pFrames), Vec_IntSize(vReasons), 
            Saig_ManPiNum(p->pAig) - p->nInputs, Vec_IntSize(vRes) );
ABC_PRT( "Time", Abc_Clock() - clk );
    }
*/

    Vec_IntFree( vReasons );
    Saig_RefManStop( p );
    return vRes;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

