package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaSatLut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSatLut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/tim/tim.h"
//#include "sat/bsat/satStore.h"
//#include "misc/util/utilNam.h"
//#include "map/scl/sclCon.h"
//#include "misc/vec/vecHsh.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Sbl_Man_t struct {
    pSat    *sat_solver;         // SAT solver
    vCardVars     *Vec_Int_t;    // candinality variables
    nVars            int32;        // max vars
    LogN            int32;         // base-2 log of max vars
    Power2            int32;       // power-2 of LogN
    FirstVar            int32;     // first variable to be used
    // statistics
    nTried            int32;       // nodes tried
    nImproved            int32;    // nodes improved
    nRuns            int32;        // the number of runs
    nHashWins            int32;    // the number of hashed windows
    nSmallWins            int32;   // the number of small windows
    nLargeWins            int32;   // the number of large windows
    nIterOuts            int32;    // the number of iters exceeded
    // parameters
    LutSize            int32;      // LUT size
    nBTLimit            int32;     // conflicts
    DelayMax            int32;     // external delay
    nEdges            int32;       // the number of edges
    fDelay            int32;       // delay mode
    fReverse            int32;     // reverse windowing
    fVerbose            int32;     // verbose
    fVeryVerbose            int32; // verbose
    fVeryVeryVerbose            int32; // verbose
    // window
    pGia     *Gia_Man_t;
    vLeaves     *Vec_Int_t;      // leaf nodes
    vAnds     *Vec_Int_t;        // AND-gates
    vNodes     *Vec_Int_t;       // internal LUTs
    vRoots     *Vec_Int_t;       // driver nodes (a subset of vAnds)
    vRootVars     *Vec_Int_t;    // driver nodes (as SAT variables)
    pHash  *Hsh_VecMan_t;        // hash table for windows
    // timing 
    vArrs     *Vec_Int_t;        // arrival times  
    vReqs     *Vec_Int_t;        // required times  
    vWindow     *Vec_Wec_t;      // fanins of each node in the window
    vPath     *Vec_Int_t;        // critical path (as SAT variables)
    vEdges     *Vec_Int_t;       // fanin edges
    // cuts
    vCutsI1     *Vec_Wrd_t;      // input bit patterns
    vCutsI2     *Vec_Wrd_t;      // input bit patterns
    vCutsN1     *Vec_Wrd_t;      // node bit patterns
    vCutsN2     *Vec_Wrd_t;      // node bit patterns
    vCutsNum     *Vec_Int_t;     // cut counts for each obj
    vCutsStart     *Vec_Int_t;   // starting cuts for each obj
    vCutsObj     *Vec_Int_t;     // object to which this cut belongs
    vTempI1     *Vec_Wrd_t;      // temporary cuts
    vTempI2     *Vec_Wrd_t;      // temporary cuts
    vTempN1     *Vec_Wrd_t;      // temporary cuts
    vTempN2     *Vec_Wrd_t;      // temporary cuts
    vSolInit     *Vec_Int_t;     // initial solution
    vSolCur     *Vec_Int_t;      // current solution
    vSolBest     *Vec_Int_t;     // best solution
    // temporary
    vLits     *Vec_Int_t;        // literals
    vAssump     *Vec_Int_t;      // literals
    vPolar     *Vec_Int_t;       // variables polarity
    // statistics
    timeWin        abctime;      // windowing
    timeCut        abctime;      // cut computation
    timeSat        abctime;      // SAT runtime
    timeSatSat        abctime;   // satisfiable time
    timeSatUns        abctime;   // unsatisfiable time
    timeSatUnd        abctime;   // undecided time
    timeTime        abctime;     // timing time
    timeStart        abctime;    // starting time
    timeTotal        abctime;    // all runtime
    timeOther        abctime;    // other time
}

// extern sat_solver * Sbm_AddCardinSolver( int LogN, Vec_Int_t ** pvVars );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbl_ManAlloc(pGia *Gia_Man_t, Number int32) *Sbl_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Sbl_Man_t  = &Sbl_Man_t{};
    p.nVars      = Number;
    p.LogN       = Abc_Base2Log(Number);
    p.Power2     = 1 << p.LogN;
    p.pSat       = Sbm_AddCardinSolver( p.LogN, &p.vCardVars );
    p.FirstVar   = sat_solver_nvars( p.pSat );
    sat_solver_bookmark( p.pSat );
    // window
    p.pGia       = pGia;
    p.vLeaves    = Vec_IntAlloc( p.nVars );
    p.vAnds      = Vec_IntAlloc( p.nVars );
    p.vNodes     = Vec_IntAlloc( p.nVars );
    p.vRoots     = Vec_IntAlloc( p.nVars );
    p.vRootVars  = Vec_IntAlloc( p.nVars );
    p.pHash      = Hsh_VecManStart( 1000 );
    // timing
    p.vArrs      = Vec_IntAlloc( 0 );
    p.vReqs      = Vec_IntAlloc( 0 );
    p.vWindow    = Vec_WecAlloc( 128 );
    p.vPath      = Vec_IntAlloc( 32 );
    p.vEdges     = Vec_IntAlloc( 32 );
    // cuts
    p.vCutsI1    = Vec_WrdAlloc( 1000 );     // input bit patterns
    p.vCutsI2    = Vec_WrdAlloc( 1000 );     // input bit patterns
    p.vCutsN1    = Vec_WrdAlloc( 1000 );     // node bit patterns
    p.vCutsN2    = Vec_WrdAlloc( 1000 );     // node bit patterns
    p.vCutsNum   = Vec_IntAlloc( 64 );       // cut counts for each obj
    p.vCutsStart = Vec_IntAlloc( 64 );       // starting cuts for each obj
    p.vCutsObj   = Vec_IntAlloc( 1000 );
    p.vSolInit   = Vec_IntAlloc( 64 );
    p.vSolCur    = Vec_IntAlloc( 64 );
    p.vSolBest   = Vec_IntAlloc( 64 );
    p.vTempI1    = Vec_WrdAlloc( 32 ); 
    p.vTempI2    = Vec_WrdAlloc( 32 ); 
    p.vTempN1    = Vec_WrdAlloc( 32 ); 
    p.vTempN2    = Vec_WrdAlloc( 32 ); 
    // internal
    p.vLits      = Vec_IntAlloc( 64 );
    p.vAssump    = Vec_IntAlloc( 64 );
    p.vPolar     = Vec_IntAlloc( 1000 );
    // other
    Gia_ManFillValue( pGia );
    return p;
}
func Sbl_ManClean(p *Sbl_Man_t) {
    p.timeStart = Abc_Clock();
    sat_solver_rollback( p.pSat );
    sat_solver_bookmark( p.pSat );
    // internal
    Vec_IntClear( p.vLeaves );
    Vec_IntClear( p.vAnds );
    Vec_IntClear( p.vNodes );
    Vec_IntClear( p.vRoots );
    Vec_IntClear( p.vRootVars );
    // timing
    Vec_IntClear( p.vArrs );
    Vec_IntClear( p.vReqs );
    Vec_WecClear( p.vWindow );
    Vec_IntClear( p.vPath );
    Vec_IntClear( p.vEdges );
    // cuts
    Vec_WrdClear( p.vCutsI1 );
    Vec_WrdClear( p.vCutsI2 );
    Vec_WrdClear( p.vCutsN1 );
    Vec_WrdClear( p.vCutsN2 );
    Vec_IntClear( p.vCutsNum );
    Vec_IntClear( p.vCutsStart );
    Vec_IntClear( p.vCutsObj );
    Vec_IntClear( p.vSolInit );
    Vec_IntClear( p.vSolCur );
    Vec_IntClear( p.vSolBest );
    Vec_WrdClear( p.vTempI1 );
    Vec_WrdClear( p.vTempI2 );
    Vec_WrdClear( p.vTempN1 );
    Vec_WrdClear( p.vTempN2 );
    // temporary
    Vec_IntClear( p.vLits );
    Vec_IntClear( p.vAssump );
    Vec_IntClear( p.vPolar );
    // other
    Gia_ManFillValue( p.pGia );
}
func Sbl_ManStop(p *Sbl_Man_t) {
    sat_solver_delete( p.pSat );
    Vec_IntFree( p.vCardVars );
    // internal
    Vec_IntFree( p.vLeaves );
    Vec_IntFree( p.vAnds );
    Vec_IntFree( p.vNodes );
    Vec_IntFree( p.vRoots );
    Vec_IntFree( p.vRootVars );
    Hsh_VecManStop( p.pHash );
    // timing
    Vec_IntFree( p.vArrs );
    Vec_IntFree( p.vReqs );
    Vec_WecFree( p.vWindow );
    Vec_IntFree( p.vPath );
    Vec_IntFree( p.vEdges );
    // cuts
    Vec_WrdFree( p.vCutsI1 );
    Vec_WrdFree( p.vCutsI2 );
    Vec_WrdFree( p.vCutsN1 );
    Vec_WrdFree( p.vCutsN2 );
    Vec_IntFree( p.vCutsNum );
    Vec_IntFree( p.vCutsStart );
    Vec_IntFree( p.vCutsObj );
    Vec_IntFree( p.vSolInit );
    Vec_IntFree( p.vSolCur );
    Vec_IntFree( p.vSolBest );
    Vec_WrdFree( p.vTempI1 );
    Vec_WrdFree( p.vTempI2 );
    Vec_WrdFree( p.vTempN1 );
    Vec_WrdFree( p.vTempN2 );
    // temporary
    Vec_IntFree( p.vLits );
    Vec_IntFree( p.vAssump );
    Vec_IntFree( p.vPolar );
    // other
    // inlining of macro ABC_FREE
    p = nil;
}

/**Function*************************************************************

  Synopsis    [For each node in the window, create fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbl_ManGetCurrentMapping(p *Sbl_Man_t) {
    var vObj *Vec_Int_t;
    var CutI1, CutI2, CutN1, CutN2 word;
    var i, c, b, iObj int32;
    Vec_WecClear( p.vWindow );
    Vec_WecInit( p.vWindow, Vec_IntSize(p.vAnds) );
    assert(Vec_IntSize(p.vSolCur) > 0, "Vec_IntSize(p.vSolCur) > 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vSolCur)) && (((c) = Vec_IntEntry(p.vSolCur, i)), 1); i++  {
        CutI1 = Vec_WrdEntry( p.vCutsI1, c );
        CutI2 = Vec_WrdEntry( p.vCutsI2, c );
        CutN1 = Vec_WrdEntry( p.vCutsN1, c );
        CutN2 = Vec_WrdEntry( p.vCutsN2, c );
        iObj  = Vec_IntEntry( p.vCutsObj, c );
        //iObj  = Vec_IntEntry( p->vAnds, iObj );
        vObj  = Vec_WecEntry( p.vWindow, iObj );
        Vec_IntClear( vObj );
        assert(Vec_IntSize(vObj) == 0, "Vec_IntSize(vObj) == 0");
        for  b = 0; b < 64; b++  {
            if (CutI1 >> b) & 1 {
                Vec_IntPush( vObj, Vec_IntEntry(p.vLeaves, b) ); }
        }
        for  b = 0; b < 64; b++  {
            if (CutI2 >> b) & 1 {
                Vec_IntPush( vObj, Vec_IntEntry(p.vLeaves, 64+b) ); }
        }
        for  b = 0; b < 64; b++  {
            if (CutN1 >> b) & 1 {
                Vec_IntPush( vObj, Vec_IntEntry(p.vAnds, b) ); }
        }
        for  b = 0; b < 64; b++  {
            if (CutN2 >> b) & 1 {
                Vec_IntPush( vObj, Vec_IntEntry(p.vAnds, 64+b) ); }
        }
    }
}


/**Function*************************************************************

  Synopsis    [Timing computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbl_ManComputeDelay(p *Sbl_Man_t, iLut int32, vFanins *Vec_Int_t) int32 {
    var k, iFan, Delay int32;
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vFanins)) && (((iFan) = Vec_IntEntry(vFanins, k)), 1); k++  {
        Delay = Abc_MaxInt( Delay, Vec_IntEntry(p.vArrs, iFan) + 1 );
    }
    return Delay;
}
func Sbl_ManCreateTiming(p *Sbl_Man_t, DelayStart int32) int32 {
    var vFanins *Vec_Int_t;
    var DelayMax int32 = DelayStart;
    var Delay int32;
    var iLut int32;
    var iFan int32;
    var k int32;
    // compute arrival times
    Vec_IntFill( p.vArrs,  Gia_ManObjNum(p.pGia), 0 );
    if p.pGia.pManTime != nil && Tim_ManBoxNum((Tim_Man_t*)p.pGia.pManTime) {
        var pObj *Gia_Obj_t; 
        var vNodes *Vec_Int_t  = Gia_ManOrderWithBoxes( p.pGia );
        Tim_ManIncrementTravId( (Tim_Man_t*)p.pGia.pManTime );
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vNodes,k))); k++  {
            iLut = Gia_ObjId( p.pGia, pObj );
            if Gia_ObjIsAnd(pObj) {
                if Gia_ObjIsLut2(p.pGia, iLut) {
                    vFanins = Gia_ObjLutFanins2(p.pGia, iLut);
                    Delay = Sbl_ManComputeDelay( p, iLut, vFanins );
                    Vec_IntWriteEntry( p.vArrs,  iLut, Delay );
                    DelayMax = Abc_MaxInt( DelayMax, Delay );
                }
            
            } else if Gia_ObjIsCi(pObj) {
                var arrTime int32  = Tim_ManGetCiArrival( (Tim_Man_t*)p.pGia.pManTime, Gia_ObjCioId(pObj) );
                Vec_IntWriteEntry( p.vArrs,  iLut, arrTime );
            
            } else if Gia_ObjIsCo(pObj) {
                var arrTime int32  = Vec_IntEntry( p.vArrs, Gia_ObjFaninId0(pObj, iLut) );
                Tim_ManSetCoArrival( (Tim_Man_t*)p.pGia.pManTime, Gia_ObjCioId(pObj), arrTime );
            
            } else if !Gia_ObjIsConst0(pObj) { 
                assert(false, "0"); }
        }
        Vec_IntFree( vNodes );
    
    } else {
        // inlining of macro Gia_ManForEachLut2
        for  iLut = 1; iLut < Gia_ManObjNum(p.pGia); iLut++  { if !Gia_ObjIsLut2(p.pGia, iLut) {} else {
            vFanins = Gia_ObjLutFanins2(p.pGia, iLut);
            Delay = Sbl_ManComputeDelay( p, iLut, vFanins );
            Vec_IntWriteEntry( p.vArrs,  iLut, Delay );
            DelayMax = Abc_MaxInt( DelayMax, Delay );
        }
        }
    }
    // compute required times
    Vec_IntFill( p.vReqs, Gia_ManObjNum(p.pGia), ABC_INFINITY );
    // inlining of macro Gia_ManForEachCoDriverId
    for  k = 0; (k < Vec_IntSize(p.pGia.vCos)) && (((iLut) = Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, k))), 1); k++  {
        Vec_IntDowndateEntry( p.vReqs, iLut, DelayMax );
    }
    if p.pGia.pManTime != nil && Tim_ManBoxNum((Tim_Man_t*)p.pGia.pManTime) {
        var pObj *Gia_Obj_t; 
        var vNodes *Vec_Int_t  = Gia_ManOrderWithBoxes( p.pGia );
        Tim_ManIncrementTravId( (Tim_Man_t*)p.pGia.pManTime );
        Tim_ManInitPoRequiredAll( (Tim_Man_t*)p.pGia.pManTime, DelayMax );
        // inlining of macro Gia_ManForEachObjVecReverse
        for  k = Vec_IntSize(vNodes) - 1; (k >= 0) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vNodes,k))); k--  {
            iLut = Gia_ObjId( p.pGia, pObj );
            if Gia_ObjIsAnd(pObj) {
                if Gia_ObjIsLut2(p.pGia, iLut) {
                    Delay = Vec_IntEntry(p.vReqs, iLut) - 1;
                    vFanins = Gia_ObjLutFanins2(p.pGia, iLut);
                    // inlining of macro Vec_IntForEachEntry
                    for  k = 0; (k < Vec_IntSize(vFanins)) && (((iFan) = Vec_IntEntry(vFanins, k)), 1); k++  {
                        Vec_IntDowndateEntry( p.vReqs, iFan, Delay );
                    }
                }
            
            } else if Gia_ObjIsCi(pObj) {
                var reqTime int32  = Vec_IntEntry( p.vReqs, iLut );
                Tim_ManSetCiRequired( (Tim_Man_t*)p.pGia.pManTime, Gia_ObjCioId(pObj), reqTime );
            
            } else if Gia_ObjIsCo(pObj) {
                var reqTime int32  = Tim_ManGetCoRequired( (Tim_Man_t*)p.pGia.pManTime, Gia_ObjCioId(pObj) );
                Vec_IntWriteEntry( p.vReqs, Gia_ObjFaninId0(pObj, iLut), reqTime );
            
            } else if !Gia_ObjIsConst0(pObj) { 
                assert(false, "0"); }
        }
        Vec_IntFree( vNodes );
    
    } else {
        // inlining of macro Gia_ManForEachLut2Reverse
        for  iLut = Gia_ManObjNum(p.pGia) - 1; iLut > 0; iLut--  { if !Gia_ObjIsLut2(p.pGia, iLut) {} else {
            Delay = Vec_IntEntry(p.vReqs, iLut) - 1;
            vFanins = Gia_ObjLutFanins2(p.pGia, iLut);
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vFanins)) && (((iFan) = Vec_IntEntry(vFanins, k)), 1); k++  {
                Vec_IntDowndateEntry( p.vReqs, iFan, Delay );
            }
        }
        }
    }
    return DelayMax;
}


/**Function*************************************************************

  Synopsis    [Given mapping in p->vSolCur, check if mapping meets delay.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbl_ManEvaluateMappingEdge(p *Sbl_Man_t, DelayGlo int32) int32 {
    var clk abctime  = Abc_Clock();
    var vArray *Vec_Int_t; 
    var i, DelayMax int32;
    Vec_IntClear( p.vPath );
    // update new timing
    Sbl_ManGetCurrentMapping( p );
    // derive new timing
    DelayMax = Gia_ManEvalWindow( p.pGia, p.vLeaves, p.vAnds, p.vWindow, p.vPolar, 1 );
    p.timeTime += Abc_Clock() - clk;
    if DelayMax <= DelayGlo {
        return 1; }
    // create critical path composed of all nodes
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p.vWindow)) && (((vArray) = Vec_WecEntry(p.vWindow, i)), 1); i++  {
        if Vec_IntSize(vArray) > 0 {
            Vec_IntPush( p.vPath, Abc_Var2Lit(i, 1) ); }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Given mapping in p->vSolCur, check the critical path.]

  Description [Returns 1 if the mapping satisfies the timing. Returns 0, 
  if the critical path is detected.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbl_ManCriticalFanin(p *Sbl_Man_t, iLut int32, vFanins *Vec_Int_t) int32 {
    var k int32;
    var iFan int32;
    var Delay int32 = Vec_IntEntry(p.vArrs, iLut);
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vFanins)) && (((iFan) = Vec_IntEntry(vFanins, k)), 1); k++  {
        if Vec_IntEntry(p.vArrs, iFan) + 1 == Delay {
            return iFan; }
    }
    return -1;
}
func Sbl_ManEvaluateMapping(p *Sbl_Man_t, DelayGlo int32) int32 {
    var clk abctime  = Abc_Clock();
    var vFanins *Vec_Int_t;
    var i int32;
    var iLut int32 = -1;
    var iAnd int32;
    var Delay int32;
    var Required int32;
    if p.pGia.vEdge1 {
        return Sbl_ManEvaluateMappingEdge( p, DelayGlo ); }
    Vec_IntClear( p.vPath );
    // derive timing
    Sbl_ManCreateTiming( p, DelayGlo );
    // update new timing
    Sbl_ManGetCurrentMapping( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vAnds)) && (((iLut) = Vec_IntEntry(p.vAnds, i)), 1); i++  {
        vFanins = Vec_WecEntry( p.vWindow, i );
        Delay   = Sbl_ManComputeDelay( p, iLut, vFanins );
        Vec_IntWriteEntry( p.vArrs, iLut, Delay );
    }
    // compare timing at the root nodes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vRoots)) && (((iLut) = Vec_IntEntry(p.vRoots, i)), 1); i++  {
        Delay    = Vec_IntEntry( p.vArrs, iLut );
        Required = Vec_IntEntry( p.vReqs, iLut );
        if Delay > Required { // updated timing exceeded original timing
            break; }
    }
    p.timeTime += Abc_Clock() - clk;
    if i == Vec_IntSize(p.vRoots) {
        return 1; }
    // derive the critical path

    // find SAT variable of the node whose GIA ID is "iLut"
    iAnd = Vec_IntFind( p.vAnds, iLut );
    assert(iAnd >= 0, "iAnd >= 0");
    // critical path begins in node "iLut", which is i-th root of the window
    assert(iAnd == Vec_IntEntry(p.vRootVars, i), "iAnd == Vec_IntEntry(p.vRootVars, i)");
    for {
        Vec_IntPush( p.vPath, Abc_Var2Lit(iAnd, 1) );
        // find fanins of this node
        vFanins = Vec_WecEntry( p.vWindow, iAnd );
        // find critical fanin
        iLut = Sbl_ManCriticalFanin( p, iLut, vFanins );
        assert(iLut > 0, "iLut > 0");
        // find SAT variable of the node whose GIA ID is "iLut"
        iAnd = Vec_IntFind( p.vAnds, iLut );
        if iAnd == -1 {
            break; }
    }
    return 0;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbl_ManUpdateMapping(p *Sbl_Man_t) {
//    Gia_Obj_t * pObj;
    var vObj *Vec_Int_t;
    var CutI1, CutI2, CutN1, CutN2 word;
    var i, c, b, iObj, iTemp int32; 
    assert(Vec_IntSize(p.vSolBest) < Vec_IntSize(p.vSolInit), "Vec_IntSize(p.vSolBest) < Vec_IntSize(p.vSolInit)");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vAnds)) && (((iObj) = Vec_IntEntry(p.vAnds, i)), 1); i++  {
        vObj = Vec_WecEntry(p.pGia.vMapping2, iObj);
        // inlining of macro Vec_IntForEachEntry
        for  b = 0; (b < Vec_IntSize(vObj)) && (((iTemp) = Vec_IntEntry(vObj, b)), 1); b++  {
            Gia_ObjLutRefDecId( p.pGia, iTemp );
        }
        Vec_IntClear( vObj );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vSolBest)) && (((c) = Vec_IntEntry(p.vSolBest, i)), 1); i++  {
        CutI1 = Vec_WrdEntry( p.vCutsI1, c );
        CutI2 = Vec_WrdEntry( p.vCutsI2, c );
        CutN1 = Vec_WrdEntry( p.vCutsN1, c );
        CutN2 = Vec_WrdEntry( p.vCutsN2, c );
        iObj = Vec_IntEntry( p.vCutsObj, c );
        iObj = Vec_IntEntry( p.vAnds, iObj );
        vObj = Vec_WecEntry( p.pGia.vMapping2, iObj );
        Vec_IntClear( vObj );
        assert(Vec_IntSize(vObj) == 0, "Vec_IntSize(vObj) == 0");
        for  b = 0; b < 64; b++  {
            if (CutI1 >> b) & 1 {
                Vec_IntPush( vObj, Vec_IntEntry(p.vLeaves, b) ); }
        }
        for  b = 0; b < 64; b++  {
            if (CutI2 >> b) & 1 {
                Vec_IntPush( vObj, Vec_IntEntry(p.vLeaves, 64+b) ); }
        }
        for  b = 0; b < 64; b++  {
            if (CutN1 >> b) & 1 {
                Vec_IntPush( vObj, Vec_IntEntry(p.vAnds, b) ); }
        }
        for  b = 0; b < 64; b++  {
            if (CutN2 >> b) & 1 {
                Vec_IntPush( vObj, Vec_IntEntry(p.vAnds, 64+b) ); }
        }
        // inlining of macro Vec_IntForEachEntry
        for  b = 0; (b < Vec_IntSize(vObj)) && (((iTemp) = Vec_IntEntry(vObj, b)), 1); b++  {
            Gia_ObjLutRefIncId( p.pGia, iTemp );
        }
    }
/*
    // verify
    Gia_ManForEachLut2Vec( p->pGia, vObj, i )
        Vec_IntForEachEntry( vObj, iTemp, b )
            Gia_ObjLutRefDecId( p->pGia, iTemp );
    Gia_ManForEachCo( p->pGia, pObj, i )
        Gia_ObjLutRefDecId( p->pGia, Gia_ObjFaninId0p(p->pGia, pObj) );

    for ( i = 0; i < Gia_ManObjNum(p->pGia); i++ )
        if ( p->pGia->pLutRefs[i] )
            printf( "Object %d has %d refs\n", i, p->pGia->pLutRefs[i] );

    Gia_ManForEachCo( p->pGia, pObj, i )
        Gia_ObjLutRefIncId( p->pGia, Gia_ObjFaninId0p(p->pGia, pObj) );
    Gia_ManForEachLut2Vec( p->pGia, vObj, i )
        Vec_IntForEachEntry( vObj, iTemp, b )
            Gia_ObjLutRefIncId( p->pGia, iTemp );
*/
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbl_ManFindAndPrintCut(p *Sbl_Man_t, c int32) int32 {
    return Sbl_ManPrintCut( Vec_WrdEntry(p.vCutsI1, c), Vec_WrdEntry(p.vCutsI2, c), Vec_WrdEntry(p.vCutsN1, c), Vec_WrdEntry(p.vCutsN2, c) );
}
func Sbl_ManComputeCutsOne(p *Sbl_Man_t, Fan0 int32, Fan1 int32, Obj int32) {
    var pCutsI1 *word  = Vec_WrdArray(p.vCutsI1);
    var pCutsI2 *word  = Vec_WrdArray(p.vCutsI2);
    var pCutsN1 *word  = Vec_WrdArray(p.vCutsN1);
    var pCutsN2 *word  = Vec_WrdArray(p.vCutsN2);
    var Start0 int32  = Vec_IntEntry( p.vCutsStart, Fan0 );
    var Start1 int32  = Vec_IntEntry( p.vCutsStart, Fan1 );
    var Limit0 int32  = Start0 + Vec_IntEntry( p.vCutsNum, Fan0 );
    var Limit1 int32  = Start1 + Vec_IntEntry( p.vCutsNum, Fan1 );
    var i, k int32;
    assert(Obj >= 0 && Obj < 128, "Obj >= 0 && Obj < 128");
    Vec_WrdClear( p.vTempI1 );
    Vec_WrdClear( p.vTempI2 );
    Vec_WrdClear( p.vTempN1 );
    Vec_WrdClear( p.vTempN2 );
    for  i = Start0; i < Limit0; i++  {
        for  k = Start1; k < Limit1; k++  {
            if Sbl_CutIsFeasible(pCutsI1[i] | pCutsI1[k], pCutsI2[i] | pCutsI2[k], pCutsN1[i] | pCutsN1[k], pCutsN2[i] | pCutsN2[k], p.LutSize) != 0 {
                Sbl_CutPushUncontained( p.vTempI1, p.vTempI2, p.vTempN1, p.vTempN2, pCutsI1[i] | pCutsI1[k], pCutsI2[i] | pCutsI2[k], pCutsN1[i] | pCutsN1[k], pCutsN2[i] | pCutsN2[k] ); }
        }
    }
    Vec_IntPush( p.vCutsStart, Vec_WrdSize(p.vCutsI1) );
    Vec_IntPush( p.vCutsNum, Vec_WrdSize(p.vTempI1) + 1 );
    Vec_WrdAppend( p.vCutsI1, p.vTempI1 );
    Vec_WrdAppend( p.vCutsI2, p.vTempI2 );
    Vec_WrdAppend( p.vCutsN1, p.vTempN1 );
    Vec_WrdAppend( p.vCutsN2, p.vTempN2 );
    Vec_WrdPush( p.vCutsI1, 0 );
    Vec_WrdPush( p.vCutsI2, 0 );
    if Obj < 64 {
        Vec_WrdPush( p.vCutsN1, (((word)1) << Obj) );
        Vec_WrdPush( p.vCutsN2, 0 );
    
    } else {
        Vec_WrdPush( p.vCutsN1, 0 );
        Vec_WrdPush( p.vCutsN2, (((word)1) << (Obj-64)) );
    }
    for  i = 0; i <= Vec_WrdSize(p.vTempI1); i++  {
        Vec_IntPush( p.vCutsObj, Obj );
    }
}
func Sbl_ManFindCut(p *Sbl_Man_t, Obj int32, CutI1 word, CutI2 word, CutN1 word, CutN2 word) int32 {
    var pCutsI1 *word  = Vec_WrdArray(p.vCutsI1);
    var pCutsI2 *word  = Vec_WrdArray(p.vCutsI2);
    var pCutsN1 *word  = Vec_WrdArray(p.vCutsN1);
    var pCutsN2 *word  = Vec_WrdArray(p.vCutsN2);
    var Start0 int32  = Vec_IntEntry( p.vCutsStart, Obj );
    var Limit0 int32  = Start0 + Vec_IntEntry( p.vCutsNum, Obj );
    var i int32;
    //printf( "\nLooking for:\n" );
    //Sbl_ManPrintCut( CutI, CutN );
    //printf( "\n" );
    for  i = Start0; i < Limit0; i++  {
        //Sbl_ManPrintCut( pCutsI[i], pCutsN[i] );
        if pCutsI1[i] == CutI1 && pCutsI2[i] == CutI2 && pCutsN1[i] == CutN1 && pCutsN2[i] == CutN2 {
            return i; }
    }
    return -1;
}
func Sbl_ManComputeCuts(p *Sbl_Man_t) int32 {
    var clk abctime  = Abc_Clock();
    var pObj *Gia_Obj_t; var vFanins *Vec_Int_t;
    var i int32;
    var k int32;
    var Index int32;
    var Fanin int32;
    var nObjs int32 = Vec_IntSize(p.vLeaves) + Vec_IntSize(p.vAnds);
    assert(Vec_IntSize(p.vLeaves) <= 128 && Vec_IntSize(p.vAnds) <= p.nVars, "Vec_IntSize(p.vLeaves) <= 128 && Vec_IntSize(p.vAnds) <= p.nVars");
    // assign leaf cuts
    Vec_IntClear( p.vCutsStart );
    Vec_IntClear( p.vCutsObj );
    Vec_IntClear( p.vCutsNum );
    Vec_WrdClear( p.vCutsI1 );
    Vec_WrdClear( p.vCutsI2 );
    Vec_WrdClear( p.vCutsN1 );
    Vec_WrdClear( p.vCutsN2 );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vLeaves)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vLeaves,i))); i++  {
        Vec_IntPush( p.vCutsStart, Vec_WrdSize(p.vCutsI1) );
        Vec_IntPush( p.vCutsObj, -1 );
        Vec_IntPush( p.vCutsNum, 1 );
        if i < 64 {
            Vec_WrdPush( p.vCutsI1, (((word)1) << i) );
            Vec_WrdPush( p.vCutsI2, 0 );
        
        } else {
            Vec_WrdPush( p.vCutsI1, 0 );
            Vec_WrdPush( p.vCutsI2, (((word)1) << (i-64)) );
        }
        Vec_WrdPush( p.vCutsN1, 0 );
        Vec_WrdPush( p.vCutsN2, 0 );
        pObj.Value = i;
    }
    // assign internal cuts
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vAnds)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vAnds,i))); i++  {
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        assert(~Gia_ObjFanin0(pObj).Value, "~Gia_ObjFanin0(pObj).Value");
        assert(~Gia_ObjFanin1(pObj).Value, "~Gia_ObjFanin1(pObj).Value");
        Sbl_ManComputeCutsOne( p, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value, i );
        pObj.Value = Vec_IntSize(p.vLeaves) + i;
    }
    assert(Vec_IntSize(p.vCutsStart) == nObjs, "Vec_IntSize(p.vCutsStart) == nObjs");
    assert(Vec_IntSize(p.vCutsNum)   == nObjs, "Vec_IntSize(p.vCutsNum)   == nObjs");
    assert(Vec_WrdSize(p.vCutsI1)    == Vec_WrdSize(p.vCutsN1), "Vec_WrdSize(p.vCutsI1)    == Vec_WrdSize(p.vCutsN1)");
    assert(Vec_WrdSize(p.vCutsI2)    == Vec_WrdSize(p.vCutsN2), "Vec_WrdSize(p.vCutsI2)    == Vec_WrdSize(p.vCutsN2)");
    assert(Vec_WrdSize(p.vCutsI1)    == Vec_IntSize(p.vCutsObj), "Vec_WrdSize(p.vCutsI1)    == Vec_IntSize(p.vCutsObj)");
    // check that roots are mapped nodes
    Vec_IntClear( p.vRootVars );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vRoots)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vRoots,i))); i++  {
        var Obj int32  = Gia_ObjId(p.pGia, pObj);
        if Gia_ObjIsCi(pObj) {
            continue; }
        assert(Gia_ObjIsLut2(p.pGia, Obj), "Gia_ObjIsLut2(p.pGia, Obj)");
        assert(~pObj.Value, "~pObj.Value");
        Vec_IntPush( p.vRootVars, pObj.Value - Vec_IntSize(p.vLeaves) );
    }
    // create current solution
    Vec_IntClear( p.vPolar );
    Vec_IntClear( p.vSolInit );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vAnds)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vAnds,i))); i++  {
        var CutI1, CutI2, CutN1, CutN2 word;
        var Obj int32  = Gia_ObjId(p.pGia, pObj);
        if !Gia_ObjIsLut2(p.pGia, Obj) {
            continue; }
        assert(int32(pObj.Value) == Vec_IntSize(p.vLeaves) + i, "int32(pObj.Value) == Vec_IntSize(p.vLeaves) + i");
        // add node
        Vec_IntPush( p.vPolar, i );
        Vec_IntPush( p.vSolInit, i );
        // add its cut
        //Gia_LutForEachFaninObj( p->pGia, Obj, pFanin, k )
        vFanins = Gia_ObjLutFanins2( p.pGia, Obj );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vFanins)) && (((Fanin) = Vec_IntEntry(vFanins, k)), 1); k++  {
            var pFanin *Gia_Obj_t  = Gia_ManObj( p.pGia, Fanin );
            assert(int32(pFanin.Value) < Vec_IntSize(p.vLeaves) || Gia_ObjIsLut2(p.pGia, Fanin), "int32(pFanin.Value) < Vec_IntSize(p.vLeaves) || Gia_ObjIsLut2(p.pGia, Fanin)");
//            if ( ~pFanin->Value == 0 )
//                Gia_ManPrintConeMulti( p->pGia, p->vAnds, p->vLeaves, p->vPath );
            if ~pFanin.Value == 0 { 
                continue; }
            assert(~pFanin.Value, "~pFanin.Value");
            if int32(pFanin.Value) < Vec_IntSize(p.vLeaves) {
                if int32(pFanin.Value) < 64 {
                    CutI1 |= ((word)1 << pFanin.Value);
                } else {
                    CutI2 |= ((word)1 << (pFanin.Value - 64)); }
            
            } else {
                if pFanin.Value - Vec_IntSize(p.vLeaves) < 64 {
                    CutN1 |= ((word)1 << (pFanin.Value - Vec_IntSize(p.vLeaves)));
                } else {
                    CutN2 |= ((word)1 << (pFanin.Value - Vec_IntSize(p.vLeaves) - 64)); }
            }
        }
        // find the new cut
        Index = Sbl_ManFindCut( p, Vec_IntSize(p.vLeaves) + i, CutI1, CutI2, CutN1, CutN2 );
        if Index < 0 {
            //printf( "Cannot find the available cut.\n" );
            continue;
        }
        assert(Index >= 0, "Index >= 0");
        Vec_IntPush( p.vPolar, p.FirstVar+Index );
    }
    // clean value
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vLeaves)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vLeaves,i))); i++  {
        pObj.Value = ~0;
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vAnds)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vAnds,i))); i++  {
        pObj.Value = ~0;
    }
    p.timeCut += Abc_Clock() - clk;
    return Vec_WrdSize(p.vCutsI1);
}
func Sbl_ManCreateCnf(p *Sbl_Man_t) int32 {
    var i int32;
    var k int32;
    var c int32;
    var pLits [2]int32;
    var value int32;
    var pCutsN1 *word  = Vec_WrdArray(p.vCutsN1);
    var pCutsN2 *word  = Vec_WrdArray(p.vCutsN2);
    assert(p.FirstVar == sat_solver_nvars(p.pSat), "p.FirstVar == sat_solver_nvars(p.pSat)");
    sat_solver_setnvars( p.pSat, sat_solver_nvars(p.pSat) + Vec_WrdSize(p.vCutsI1) );
    //printf( "\n" );
    for  i = 0; i < Vec_IntSize(p.vAnds); i++  {
        var Start0 int32  = Vec_IntEntry( p.vCutsStart, Vec_IntSize(p.vLeaves) + i );
        var Limit0 int32  = Start0 + Vec_IntEntry( p.vCutsNum, Vec_IntSize(p.vLeaves) + i ) - 1;
        // add main clause
        Vec_IntClear( p.vLits );
        Vec_IntPush( p.vLits, Abc_Var2Lit(i, 1) );
        //printf( "Node %d implies cuts: ", i );
        for  k = Start0; k < Limit0; k++  {
            Vec_IntPush( p.vLits, Abc_Var2Lit(p.FirstVar+k, 0) );
            //printf( "%d ", p->FirstVar+k );
        }
        //printf( "\n" );
        value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
        assert((value != 0), "value");
        // binary clauses
        for  k = Start0; k < Limit0; k++  {
            var Cut1 word  = pCutsN1[k];
            var Cut2 word  = pCutsN2[k];
            //printf( "Cut %d implies root node %d.\n", p->FirstVar+k, i );
            // root clause
            pLits[0] = Abc_Var2Lit( p.FirstVar+k, 1 );
            pLits[1] = Abc_Var2Lit( i, 0 );
            value = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
            assert((value != 0), "value");
            // fanin clauses
            for  c = 0; c < 64 && Cut1; c++, Cut1 >>= 1  {
                if (Cut1 & 1) == 0 {
                    continue; }
                //printf( "Cut %d implies fanin %d.\n", p->FirstVar+k, c );
                pLits[1] = Abc_Var2Lit( c, 0 );
                value = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
                assert((value != 0), "value");
            }
            for  c = 0; c < 64 && Cut2; c++, Cut2 >>= 1  {
                if (Cut2 & 1) == 0 {
                    continue; }
                //printf( "Cut %d implies fanin %d.\n", p->FirstVar+k, c );
                pLits[1] = Abc_Var2Lit( c+64, 0 );
                value = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
                assert((value != 0), "value");
            }
        }
    }
    sat_solver_set_polarity( p.pSat, Vec_IntArray(p.vPolar), Vec_IntSize(p.vPolar) );
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func Sbl_ManWindow(p *Sbl_Man_t) {
    var i, ObjId int32;
    // collect leaves
    Vec_IntClear( p.vLeaves );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((ObjId) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i))); i++  {
        Vec_IntPush( p.vLeaves, ObjId );
    }
    // collect internal
    Vec_IntClear( p.vAnds );
    // inlining of macro Gia_ManForEachAndId
    for  ObjId = 0; (ObjId < p.pGia.nObjs); ObjId++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, ObjId)) {} else {
        Vec_IntPush( p.vAnds, ObjId ); }
    }
    // collect roots
    Vec_IntClear( p.vRoots );
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && (((ObjId) = Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))), 1); i++  {
        Vec_IntPush( p.vRoots, ObjId );
    }
}

func Sbl_ManWindow2(p *Sbl_Man_t, iPivot int32) int32 {
    var clk abctime  = Abc_Clock();
    var vRoots, vNodes, vLeaves, vAnds *Vec_Int_t;
    var Count int32  = Gia_ManComputeOneWin( p.pGia, iPivot, &vRoots, &vNodes, &vLeaves, &vAnds );
    p.timeWin += Abc_Clock() - clk;
    if Count == 0 {
        return 0; }
    Vec_IntClear( p.vRoots );   Vec_IntAppend( p.vRoots,  vRoots );
    Vec_IntClear( p.vNodes );   Vec_IntAppend( p.vNodes,  vNodes );
    Vec_IntClear( p.vLeaves );  Vec_IntAppend( p.vLeaves, vLeaves );
    Vec_IntClear( p.vAnds );    Vec_IntAppend( p.vAnds,   vAnds );
//Vec_IntPrint( vRoots );
//Vec_IntPrint( vAnds );
//Vec_IntPrint( vLeaves );
    // recompute internal nodes
//    Gia_ManIncrementTravId( p->pGia );
//    Gia_ManCollectAnds( p->pGia, Vec_IntArray(p->vRoots), Vec_IntSize(p->vRoots), p->vAnds, p->vLeaves );
    return Count;
}

func Sbl_ManTestSat(p *Sbl_Man_t, iPivot int32) int32 {
    var fKeepTrying int32  = 1;
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    var i, status, Root, Count, StartSol, nConfTotal, nIters int32;
    var nEntries int32  = Hsh_VecSize( p.pHash );
    p.nTried++;

    Sbl_ManClean( p );

    // compute one window
    Count = Sbl_ManWindow2( p, iPivot );
    if Count == 0 {
        if p.fVeryVerbose {
        fmt.Printf("Obj %d: Window with less than %d nodes does not exist.\n", iPivot, p.nVars); }
        p.nSmallWins++;
        return 0;
    }
    Hsh_VecManAdd( p.pHash, p.vAnds );
    if nEntries == Hsh_VecSize(p.pHash) {
        if p.fVeryVerbose {
        fmt.Printf("Obj %d: This window was already tried.\n", iPivot); }
        p.nHashWins++;
        return 0;
    }
    if p.fVeryVerbose {
    fmt.Printf("\nObj = %6d : Leaf = %2d.  AND = %2d.  Root = %2d.    LUT = %2d.\n", iPivot, Vec_IntSize(p.vLeaves), Vec_IntSize(p.vAnds), Vec_IntSize(p.vRoots), Vec_IntSize(p.vNodes)); } 

    if Vec_IntSize(p.vLeaves) > 128 || Vec_IntSize(p.vAnds) > p.nVars {
        if p.fVeryVerbose {
        fmt.Printf("Obj %d: Encountered window with %d inputs and %d internal nodes.\n", iPivot, Vec_IntSize(p.vLeaves), Vec_IntSize(p.vAnds)); }
        p.nLargeWins++;
        return 0;
    }
    if Vec_IntSize(p.vAnds) < 10 {
        if p.fVeryVerbose {
        fmt.Printf("Skipping.\n"); }
        return 0;
    }

    // derive cuts
    Sbl_ManComputeCuts( p );
    // derive SAT instance
    Sbl_ManCreateCnf( p );

    if p.fVeryVeryVerbose {
    fmt.Printf("All clauses = %d.  Multi clauses = %d.  Binary clauses = %d.  Other clauses = %d.\n\n", sat_solver_nclauses(p.pSat), Vec_IntSize(p.vAnds), Vec_WrdSize(p.vCutsI1) - Vec_IntSize(p.vAnds), sat_solver_nclauses(p.pSat) - Vec_WrdSize(p.vCutsI1)); }

    // create assumptions
    // cardinality
    Vec_IntClear( p.vAssump );
    Vec_IntPush( p.vAssump, -1 );
    // unused variables
    for  i = Vec_IntSize(p.vAnds); i < p.Power2; i++  {
        Vec_IntPush( p.vAssump, Abc_Var2Lit(i, 1) );
    }
    // root variables
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vRootVars)) && (((Root) = Vec_IntEntry(p.vRootVars, i)), 1); i++  {
        Vec_IntPush( p.vAssump, Abc_Var2Lit(Root, 0) );
    }
//    Vec_IntPrint( p->vAssump );

    StartSol = Vec_IntSize(p.vSolInit) + 1;
//    StartSol = 30;
    for  (fKeepTrying != 0) && StartSol-fKeepTrying > 0  {
        var Count, LitCount int32;
        var nConfBef, nConfAft int32;
        if p.fVeryVerbose {
            fmt.Printf("Trying to find mapping with %d LUTs.\n", StartSol-fKeepTrying); }
    //    for ( i = Vec_IntSize(p->vSolInit)-5; i < nVars; i++ )
    //        Vec_IntPush( p->vAssump, Abc_Var2Lit(Vec_IntEntry(p->vCardVars, i), 1) );
        Vec_IntWriteEntry( p.vAssump, 0, Abc_Var2Lit(Vec_IntEntry(p.vCardVars, StartSol-fKeepTrying), 1) );
        // solve the problem
        clk2 = Abc_Clock();
        nConfBef = int32(p.pSat.stats.conflicts);
        status = sat_solver_solve( p.pSat, Vec_IntArray(p.vAssump), Vec_IntLimit(p.vAssump), p.nBTLimit, 0, 0, 0 );
        p.timeSat += Abc_Clock() - clk2;
        nConfAft = int32(p.pSat.stats.conflicts);
        nConfTotal += nConfAft - nConfBef;
        nIters++;
        p.nRuns++;
        if status == l_True {
            p.timeSatSat += Abc_Clock() - clk2;
        } else if status == l_False {
            p.timeSatUns += Abc_Clock() - clk2;
        } else { 
            p.timeSatUnd += Abc_Clock() - clk2; }
        if status == l_Undef {
            break; }
        if status == l_True {
            if p.fVeryVeryVerbose {
                for  i = 0; i < Vec_IntSize(p.vAnds); i++  {
                    fmt.Printf("%d", sat_solver_var_value(p.pSat, i));
                }
                fmt.Printf("\n");
                for  i = 0; i < Vec_IntSize(p.vAnds); i++  {
                    if sat_solver_var_value(p.pSat, i) {
                        fmt.Printf("%d=%d ", i, sat_solver_var_value(p.pSat, i));
                        Count++;
                    }
                }
                fmt.Printf("Count = %d\n", Count);
            }
//            for ( i = p->FirstVar; i < sat_solver_nvars(p->pSat); i++ )
//                printf( "%d", sat_solver_var_value(p->pSat, i) );
//            printf( "\n" );
            Count = 1;
            Vec_IntClear( p.vSolCur );
            for  i = p.FirstVar; i < sat_solver_nvars(p.pSat); i++  {
                if sat_solver_var_value(p.pSat, i) {
                    if p.fVeryVeryVerbose {
                        fmt.Printf("Cut %3d : Node = %3d %6d  ", Count++, Vec_IntEntry(p.vCutsObj, i-p.FirstVar), Vec_IntEntry(p.vAnds, Vec_IntEntry(p.vCutsObj, i-p.FirstVar))); }
                    if p.fVeryVeryVerbose {
                        LitCount += Sbl_ManFindAndPrintCut( p, i-p.FirstVar ); }
                    Vec_IntPush( p.vSolCur, i-p.FirstVar );
                }
            }
            //Vec_IntPrint( p->vRootVars );
            //Vec_IntPrint( p->vRoots );
            //Vec_IntPrint( p->vAnds );
            //Vec_IntPrint( p->vLeaves );
        }

//        fKeepTrying = status == l_True ? fKeepTrying + 1 : 0;
        // check the timing
        if status == l_True {
            if p.fDelay && !(Sbl_ManEvaluateMapping(p, p.DelayMax) != 0) {
                var iLit, value int32;
                if p.fVeryVerbose {
                    fmt.Printf("Critical path of length (%d) is detected:   ", Vec_IntSize(p.vPath));
                    // inlining of macro Vec_IntForEachEntry
                    for  i = 0; (i < Vec_IntSize(p.vPath)) && (((iLit) = Vec_IntEntry(p.vPath, i)), 1); i++  {
                        fmt.Printf("%d=%d ", i, Vec_IntEntry(p.vAnds, Abc_Lit2Var(iLit)));
                    }
                    fmt.Printf("\n");
                }
                // add the clause
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vPath), Vec_IntLimit(p.vPath)  );
                assert((value != 0), "value");
            
            } else {
                Vec_IntClear( p.vSolBest );
                Vec_IntAppend( p.vSolBest, p.vSolCur );
                fKeepTrying++;
            }
        
        } else {
            fKeepTrying = 0; }
        if p.fVeryVerbose {
            if status == l_False {
                fmt.Printf("UNSAT ");
            } else if status == l_True {
                fmt.Printf("SAT   ");
            } else { 
                fmt.Printf("UNDEC "); }
            fmt.Printf("confl =%8d.    ", nConfAft - nConfBef);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk2 );

            fmt.Printf("Total ");
            fmt.Printf("confl =%8d.    ", nConfTotal);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            if p.fVeryVeryVerbose && status == l_True {
                fmt.Printf("LitCount = %d.\n", LitCount); }
            fmt.Printf("\n");
        }
        if nIters == 10 {
            p.nIterOuts++;
            //printf( "Obj %d : Quitting after %d iterations.\n", iPivot, nIters );
            break;
        }
    }

    // update solution
    if Vec_IntSize(p.vSolBest) > 0 && Vec_IntSize(p.vSolBest) < Vec_IntSize(p.vSolInit) {
        var nDelayCur, nEdgesCur int32;
        Sbl_ManUpdateMapping( p );
        if p.pGia.vEdge1 {
            nDelayCur = Gia_ManEvalEdgeDelay( p.pGia );
            nEdgesCur = Gia_ManEvalEdgeCount( p.pGia );
        
        } else {
            nDelayCur = Sbl_ManCreateTiming( p, p.DelayMax ); }
        if p.fVerbose {
        fmt.Printf("Object %5d : Saved %2d nodes  (Conf =%8d)  Iter =%3d  Delay = %d  Edges = %4d\n", iPivot, Vec_IntSize(p.vSolInit)-Vec_IntSize(p.vSolBest), nConfTotal, nIters, nDelayCur, nEdgesCur); }
        p.timeTotal += Abc_Clock() - p.timeStart;
        p.nImproved++;
        return 2;
    
    } else {
//        printf( "Object %5d : Saved %2d nodes  (Conf =%8d)  Iter =%3d\n", iPivot, 0, nConfTotal, nIters );
    }
    p.timeTotal += Abc_Clock() - p.timeStart;
    return 1;
}
func Sbl_ManPrintRuntime(p *Sbl_Man_t) {
    fmt.Printf("Runtime breakdown:\n");
    p.timeOther = p.timeTotal - p.timeWin - p.timeCut - p.timeSat - p.timeTime;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Win   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeWin)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeWin)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Cut   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCut)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeCut)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Sat   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Sat  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Unsat"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUns)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUns)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Undec"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnd)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUnd)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Timing"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTime)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTime)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Other "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeOther)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("ALL   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;
}
func Gia_ManLutSat(pGia *Gia_Man_t, LutSize int32, nNumber int32, nImproves int32, nBTLimit int32, DelayMax int32, nEdges int32, fDelay int32, fReverse int32, fVerbose int32, fVeryVerbose int32) {
    var iLut, nImproveCount int32;
    var p *Sbl_Man_t    = Sbl_ManAlloc( pGia, nNumber );
    p.LutSize      = LutSize;      // LUT size
    p.nBTLimit     = nBTLimit;     // conflicts
    p.DelayMax     = DelayMax;     // external delay
    p.nEdges       = nEdges;       // the number of edges
    p.fDelay       = fDelay;       // delay mode
    p.fReverse     = fReverse;     // reverse windowing
    p.fVerbose     = fVerbose | fVeryVerbose;
    p.fVeryVerbose = fVeryVerbose;
    if p.fVerbose {
    fmt.Printf("Parameters: WinSize = %d AIG nodes.  Conf = %d.  DelayMax = %d.\n", p.nVars, p.nBTLimit, p.DelayMax); }
    // determine delay limit
    if (fDelay != 0) && pGia.vEdge1 && p.DelayMax == 0 {
        p.DelayMax = Gia_ManEvalEdgeDelay( pGia ); }
    // iterate through the internal nodes
    Gia_ManComputeOneWinStart( pGia, nNumber, fReverse );
    // inlining of macro Gia_ManForEachLut2
    for  iLut = 1; iLut < Gia_ManObjNum(pGia); iLut++  { if !Gia_ObjIsLut2(pGia, iLut) {} else {
        if Sbl_ManTestSat( p, iLut ) != 2 {
            continue; }
        if nImproveCount++; nImproveCount == nImproves {
            break; }
    }
    }
    Gia_ManComputeOneWin( pGia, -1, nil, nil, nil, nil );
    if p.fVerbose {
    fmt.Printf("Tried = %d. Used = %d. HashWin = %d. SmallWin = %d. LargeWin = %d. IterOut = %d.  SAT runs = %d.\n", p.nTried, p.nImproved, p.nHashWins, p.nSmallWins, p.nLargeWins, p.nIterOuts, p.nRuns); }
    if p.fVerbose {
    Sbl_ManPrintRuntime( p ); }
    Sbl_ManStop( p );
    Vec_IntFreeP( &pGia.vPacking );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

