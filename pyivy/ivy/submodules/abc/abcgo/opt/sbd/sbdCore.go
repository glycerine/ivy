package sbd

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [sbdCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sbdCore.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sbdInt.h"
//#include "opt/dau/dau.h"
//#include "misc/tim/tim.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const SBD_MAX_LUTSIZE = 6

   
type Sbd_Man_t struct {
    pPars      *Sbd_Par_t;       // user's parameters
    pGia      *Gia_Man_t;        // user's AIG manager (will be modified by adding nodes)
    vTfos      *Vec_Wec_t;       // TFO for each node (roots are marked) (windowing)
    vLutLevs      *Vec_Int_t;    // LUT level for each node after resynthesis
    vLutCuts      *Vec_Int_t;    // LUT cut for each nodes after resynthesis
    vLutCuts2      *Vec_Int_t;   // LUT cut for each nodes after resynthesis
    vMirrors      *Vec_Int_t;    // alternative node
    Vec_Wrd_t *     vSims[4];    // simulation information (main, backup, controlability)
    vCover      *Vec_Int_t;      // temporary
    vLits      *Vec_Int_t;       // temporary
    vLits2      *Vec_Int_t;      // temporary
    nLuts [6]int32;    // 0=const, 1=1lut, 2=2lut, 3=3lut
    nTried             int32;  
    nUsed             int32;  
    timeWin         abctime;
    timeCut         abctime;
    timeCov         abctime;
    timeCnf         abctime;
    timeSat         abctime;
    timeQbf         abctime;
    timeNew         abctime;
    timeOther         abctime;
    timeTotal         abctime;
    pSto      *Sbd_Sto_t;
    pSrv      *Sbd_Srv_t;
    // target node
    Pivot             int32;       // target node
    DivCutoff             int32;   // the place where D-2 divisors begin
    vTfo      *Vec_Int_t;        // TFO (excludes node, includes roots) - precomputed
    vRoots      *Vec_Int_t;      // TFO root nodes
    vWinObjs      *Vec_Int_t;    // TFI + Pivot + sideTFI + TFO (including roots)
    vObj2Var      *Vec_Int_t;    // SAT variables for the window (indexes of objects in vWinObjs)
    vDivSet      *Vec_Int_t;     // divisor variables
    vDivVars      *Vec_Int_t;    // divisor variables
    vDivValues      *Vec_Int_t;  // SAT variables values for the divisor variables
    vDivLevels      *Vec_Wec_t;  // divisors collected by levels
    Vec_Int_t *     vCounts[2];  // counters of zeros and ones
    vMatrix      *Vec_Wrd_t;     // covering matrix
    pSat     *sat_solver;        // SAT solver
}

func Sbd_ObjCut(p *Sbd_Man_t, i int32) *int32 { return Vec_IntEntryP( p.vLutCuts,  (p.pPars.nLutSize + 1) * i ); }
func Sbd_ObjCut2(p *Sbd_Man_t, i int32) *int32 { return Vec_IntEntryP( p.vLutCuts2, (p.pPars.nLutSize + 1) * i ); }

func Sbd_ObjSim0(p *Sbd_Man_t, i int32) *word { return Vec_WrdEntryP( p.vSims[0], p.pPars.nWords * i );         }
func Sbd_ObjSim1(p *Sbd_Man_t, i int32) *word { return Vec_WrdEntryP( p.vSims[1], p.pPars.nWords * i );         }
func Sbd_ObjSim2(p *Sbd_Man_t, i int32) *word { return Vec_WrdEntryP( p.vSims[2], p.pPars.nWords * i );         }
func Sbd_ObjSim3(p *Sbd_Man_t, i int32) *word { return Vec_WrdEntryP( p.vSims[3], p.pPars.nWords * i );         }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ParSetDefault(pPars *Sbd_Par_t) {
    memset( pPars, 0, sizeof(Sbd_Par_t) );
    pPars.nLutSize     = 4;    // target LUT size
    pPars.nLutNum      = 3;    // target LUT count
    pPars.nCutSize     = (pPars.nLutSize - 1) * pPars.nLutNum + 1;  // target cut size
    pPars.nCutNum      = 128;  // target cut count
    pPars.nTfoLevels   = 5;    // the number of TFO levels (windowing)
    pPars.nTfoFanMax   = 4;    // the max number of fanouts (windowing)
    pPars.nWinSizeMax  = 2000; // maximum window size (windowing)
    pPars.nBTLimit     = 0;    // maximum number of SAT conflicts 
    pPars.nWords       = 1;    // simulation word count
    pPars.fMapping     = 1;    // generate mapping
    pPars.fMoreCuts    = 0;    // use several cuts
    pPars.fFindDivs    = 0;    // perform divisor search
    pPars.fUsePath     = 0;    // optimize only critical path
    pPars.fArea        = 0;    // area-oriented optimization
    pPars.fCover       = 0;    // use complete cover procedure
    pPars.fVerbose     = 0;    // verbose flag
    pPars.fVeryVerbose = 0;    // verbose flag
}

/**Function*************************************************************

  Synopsis    [Computes TFO and window roots for all nodes.]

  Description [TFO does not include the node itself. If TFO is empty,
  it means that the node itself is its own root, which may happen if
  the node is pointed by a PO or if it has too many fanouts.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ManWindowRoots(p *Gia_Man_t, nTfoLevels int32, nTfoFanMax int32) *Vec_Wec_t {
    var vTfos *Vec_Wec_t  = Vec_WecStart( Gia_ManObjNum(p) ); // TFO nodes with roots marked
    var vTemp *Vec_Wec_t  = Vec_WecStart( Gia_ManObjNum(p) ); // storage
    var vNodes, vNodes0, vNodes1 *Vec_Int_t;
    var vPoDrivers *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var i, k, k2, Id, Fan int32;
    Gia_ManLevelNum( p );
    Gia_ManCreateRefs( p );
    Gia_ManCleanMark0( p );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        vNodes = Vec_WecEntry( vTemp, Id );
        Vec_IntGrow( vNodes, 1 );
        Vec_IntPush( vNodes, Id );
    }
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && (((Id) = Gia_ObjFaninId0p(p, Gia_ManCo(p, i))), 1); i++  {
        Vec_BitWriteEntry( vPoDrivers, Id, 1 );
    }
    // inlining of macro Gia_ManForEachAndId
    for  Id = 0; (Id < p.nObjs); Id++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p, Id)) {} else {
        var fAlwaysRoot int32  = Vec_BitEntry(vPoDrivers, Id) || (Gia_ObjRefNumId(p, Id) >= nTfoFanMax);
        vNodes0 = Vec_WecEntry( vTemp, Gia_ObjFaninId0(Gia_ManObj(p, Id), Id) );
        vNodes1 = Vec_WecEntry( vTemp, Gia_ObjFaninId1(Gia_ManObj(p, Id), Id) );
        vNodes = Vec_WecEntry( vTemp, Id );
        Vec_IntTwoMerge2( vNodes0, vNodes1, vNodes );
        k2 = 0;
        Vec_IntForEachEntry( vNodes, Fan, k )
        {
            var fRoot int32  = fAlwaysRoot || (Gia_ObjLevelId(p, Id) - Gia_ObjLevelId(p, Fan) >= nTfoLevels);
            Vec_WecPush( vTfos, Fan, Abc_Var2Lit(Id, fRoot) );
            if !(fRoot != 0) { Vec_IntWriteEntry( vNodes, k2++, Fan ); }
        }
        Vec_IntShrink( vNodes, k2 );
        if !(fAlwaysRoot != 0) {
            Vec_IntPush( vNodes, Id ); }
    }
    }
    Vec_WecFree( vTemp );
    Vec_BitFree( vPoDrivers );

    // print the results
    if ( 0 )
    Vec_WecForEachLevel( vTfos, vNodes, i )
    {
        if !Gia_ObjIsAnd(Gia_ManObj(p, i)) {
            continue; }
        fmt.Printf("Node %3d : ", i);
        Vec_IntForEachEntry( vNodes, Fan, k )
            fmt.Printf("%d%s ", Abc_Lit2Var(Fan), Abc_LitIsCompl(Fan)? "*":"");
        fmt.Printf("\n");
    }

    return vTfos;
}

/**Function*************************************************************

  Synopsis    [Manager manipulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ManStart(pGia *Gia_Man_t, pPars *Sbd_Par_t) *Sbd_Man_t {
    var i, w, Id int32;
    var p *Sbd_Man_t  = new(Sbd_Man_t);  
    p.timeTotal  = Abc_Clock();
    p.pPars      = pPars;
    p.pGia       = pGia;
    p.vTfos      = Sbd_ManWindowRoots( pGia, pPars.nTfoLevels, pPars.nTfoFanMax );
    p.vLutLevs   = Vec_IntStart( Gia_ManObjNum(pGia) );
    p.vLutCuts   = Vec_IntStart( Gia_ManObjNum(pGia) * (p.pPars.nLutSize + 1) );
    p.vMirrors   = Vec_IntStartFull( Gia_ManObjNum(pGia) );
    for  i = 0; i < 4; i++  {
        p.vSims[i] = Vec_WrdStart( Gia_ManObjNum(pGia) * p.pPars.nWords );
    }
    // target node
    p.vCover     = Vec_IntAlloc( 100 );
    p.vLits      = Vec_IntAlloc( 100 );
    p.vLits2     = Vec_IntAlloc( 100 );
    p.vRoots     = Vec_IntAlloc( 100 );
    p.vWinObjs   = Vec_IntAlloc( Gia_ManObjNum(pGia) );
    p.vObj2Var   = Vec_IntStart( Gia_ManObjNum(pGia) );
    p.vDivSet    = Vec_IntAlloc( 100 );
    p.vDivVars   = Vec_IntAlloc( 100 );
    p.vDivValues = Vec_IntAlloc( 100 );
    p.vDivLevels = Vec_WecAlloc( 100 );
    p.vCounts[0] = Vec_IntAlloc( 100 );
    p.vCounts[1] = Vec_IntAlloc( 100 );
    p.vMatrix    = Vec_WrdAlloc( 100 );
    // start input cuts
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((Id) = Gia_ObjId(pGia, Gia_ManCi(pGia, i))); i++  {
        var pCut *int32  = Sbd_ObjCut( p, Id );
        pCut[0] = 1;
        pCut[1] = Id;
    }
    // generate random input
    Gia_ManRandom( 1 );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((Id) = Gia_ObjId(pGia, Gia_ManCi(pGia, i))); i++  {
        for  w = 0; w < p.pPars.nWords; w++  {
            Sbd_ObjSim0(p, Id)[w] = Gia_ManRandomW( 0 );
        }
    }     
    // cut enumeration
    if pPars.fMoreCuts {
        p.pSto = Sbd_StoAlloc( pGia, p.vMirrors, pPars.nLutSize, pPars.nCutSize, pPars.nCutNum, !pPars.fMapping, 1 );
    } else {
        p.pSto = Sbd_StoAlloc( pGia, p.vMirrors, pPars.nLutSize, pPars.nLutSize, pPars.nCutNum, !pPars.fMapping, 1 );
        p.pSrv = Sbd_ManCutServerStart( pGia, p.vMirrors, p.vLutLevs, nil, nil, pPars.nLutSize, pPars.nCutSize, pPars.nCutNum, 0 );
    }
    return p;
}
func Sbd_ManStop(p *Sbd_Man_t) {
    var i int32;
    Vec_WecFree( p.vTfos );
    Vec_IntFree( p.vLutLevs );
    Vec_IntFree( p.vLutCuts );
    Vec_IntFree( p.vMirrors );
    for  i = 0; i < 4; i++  {
        Vec_WrdFree( p.vSims[i] );
    }
    Vec_IntFree( p.vCover );
    Vec_IntFree( p.vLits );
    Vec_IntFree( p.vLits2 );
    Vec_IntFree( p.vRoots );
    Vec_IntFree( p.vWinObjs );
    Vec_IntFree( p.vObj2Var );
    Vec_IntFree( p.vDivSet );
    Vec_IntFree( p.vDivVars );
    Vec_IntFree( p.vDivValues );
    Vec_WecFree( p.vDivLevels );
    Vec_IntFree( p.vCounts[0] );
    Vec_IntFree( p.vCounts[1] );
    Vec_WrdFree( p.vMatrix );
    sat_solver_delete_p( &p.pSat );
    if p.pSto { Sbd_StoFree( p.pSto ); }
    if p.pSrv { Sbd_ManCutServerStop( p.pSrv ); }
    ABC_FREE( p );
}


/**Function*************************************************************

  Synopsis    [Constructing window.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ManPropagateControlOne(p *Sbd_Man_t, Node int32) {
    var pNode *Gia_Obj_t  = Gia_ManObj(p.pGia, Node);  var w int32;
    var iObj0 int32  = Gia_ObjFaninId0(pNode, Node);
    var iObj1 int32  = Gia_ObjFaninId1(pNode, Node);

//    word * pSims  = Sbd_ObjSim0(p, Node);
//    word * pSims0 = Sbd_ObjSim0(p, iObj0);
//    word * pSims1 = Sbd_ObjSim0(p, iObj1);

    var pCtrl *word   = Sbd_ObjSim2(p, Node);
    var pCtrl0 *word  = Sbd_ObjSim2(p, iObj0);
    var pCtrl1 *word  = Sbd_ObjSim2(p, iObj1);

    var pDtrl *word   = Sbd_ObjSim3(p, Node);
    var pDtrl0 *word  = Sbd_ObjSim3(p, iObj0);
    var pDtrl1 *word  = Sbd_ObjSim3(p, iObj1);

//    Gia_ObjPrint( p->pGia, pNode );
//    printf( "Node %2d : %d %d\n\n", Node, (int)(pSims[0] & 1), (int)(pCtrl[0] & 1) );

    for  w = 0; w < p.pPars.nWords; w++  {
//        word Sim0 = Gia_ObjFaninC0(pNode) ? ~pSims0[w] : pSims0[w];
//        word Sim1 = Gia_ObjFaninC1(pNode) ? ~pSims1[w] : pSims1[w];

        pCtrl0[w] |= pCtrl[w];// & (pSims[w] | Sim1 | (~Sim0 & ~Sim1));
        pCtrl1[w] |= pCtrl[w];// & (pSims[w] | Sim0 | (~Sim0 & ~Sim1));

        pDtrl0[w] |= pDtrl[w];// & (pSims[w] | Sim1 | (~Sim0 & ~Sim1));
        pDtrl1[w] |= pDtrl[w];// & (pSims[w] | Sim0 | (~Sim0 & ~Sim1));
    }
}
func Sbd_ManPropagateControl(p *Sbd_Man_t, Pivot int32) {
    var clk abctime  = Abc_Clock();
    var i, Node int32;
    Abc_TtCopy( Sbd_ObjSim3(p, Pivot), Sbd_ObjSim2(p, Pivot), p.pPars.nWords, 0 );
    // clean controlability
    for  i = 0; i < Vec_IntEntry(p.vObj2Var, Pivot) && ((Node = Vec_IntEntry(p.vWinObjs, i)), 1); i++  {
        Abc_TtClear( Sbd_ObjSim2(p, Node), p.pPars.nWords );
        Abc_TtClear( Sbd_ObjSim3(p, Node), p.pPars.nWords );
        //printf( "Clearing node %d.\n", Node );
    }
    // propagate controlability to fanins for the TFI nodes starting from the pivot
    for  i = Vec_IntEntry(p.vObj2Var, Pivot); i >= 0 && ((Node = Vec_IntEntry(p.vWinObjs, i)), 1); i--  {
        if Gia_ObjIsAnd(Gia_ManObj(p.pGia, Node)) {
            Sbd_ManPropagateControlOne( p, Node ); }
    }
    p.timeWin += Abc_Clock() - clk;
}
func Sbd_ManUpdateOrder(p *Sbd_Man_t, Pivot int32) {
    var i, k, Node int32;
    var vLevel *Vec_Int_t;
    var nTimeValidDivs int32;
    // collect divisors by logic level
    var LevelMax int32  = Vec_IntEntry(p.vLutLevs, Pivot);
    Vec_WecClear( p.vDivLevels );
    Vec_WecInit( p.vDivLevels, LevelMax + 1 );
    Vec_IntForEachEntry( p.vWinObjs, Node, i )
        Vec_WecPush( p.vDivLevels, Vec_IntEntry(p.vLutLevs, Node), Node );
    // reload divisors
    Vec_IntClear( p.vWinObjs );
    Vec_WecForEachLevel( p.vDivLevels, vLevel, i )
    {
        Vec_IntSort( vLevel, 0 );
        Vec_IntForEachEntry( vLevel, Node, k )
        {
            Vec_IntWriteEntry( p.vObj2Var, Node, Vec_IntSize(p.vWinObjs) );
            Vec_IntPush( p.vWinObjs, Node );
        }
        // remember divisor cutoff
        if i == LevelMax - 2 {
            nTimeValidDivs = Vec_IntSize(p.vWinObjs); }
    }
    assert(nTimeValidDivs > 0, "nTimeValidDivs > 0");
    Vec_IntClear( p.vDivVars );
    p.DivCutoff = -1;
    Vec_IntForEachEntryStartStop( p.vWinObjs, Node, i, Abc_MaxInt(0, nTimeValidDivs-63), nTimeValidDivs )
    {
        if p.DivCutoff == -1 && Vec_IntEntry(p.vLutLevs, Node) == LevelMax - 2 {
            p.DivCutoff = Vec_IntSize(p.vDivVars); }
        Vec_IntPush( p.vDivVars, i );
    }
    if p.DivCutoff == -1 {
        p.DivCutoff = 0; }
    // verify
/*
    assert( Vec_IntSize(p->vDivVars) < 64 );
    Vec_IntForEachEntryStart( p->vDivVars, Node, i, p->DivCutoff )
        assert( Vec_IntEntry(p->vLutLevs, Vec_IntEntry(p->vWinObjs, Node)) == LevelMax - 2 );
    Vec_IntForEachEntryStop( p->vDivVars, Node, i, p->DivCutoff )
        assert( Vec_IntEntry(p->vLutLevs, Vec_IntEntry(p->vWinObjs, Node)) < LevelMax - 2 );
*/
    Vec_IntFill( p.vDivValues, Vec_IntSize(p.vDivVars), 0 );
    //printf( "%d ", Vec_IntSize(p->vDivVars) );
//    printf( "Node %4d :  Win = %5d.   Divs = %5d.    D1 = %5d.  D2 = %5d.\n",  
//        Pivot, Vec_IntSize(p->vWinObjs), Vec_IntSize(p->vDivVars), Vec_IntSize(p->vDivVars)-p->DivCutoff, p->DivCutoff );
}
func Sbd_ManWindowSim_rec(p *Sbd_Man_t, NodeInit int32) {
    var pObj *Gia_Obj_t;
    var Node int32  = NodeInit;
    if Vec_IntEntry(p.vMirrors, Node) >= 0 {
        Node = Abc_Lit2Var( Vec_IntEntry(p.vMirrors, Node) ); }
    if Gia_ObjIsTravIdCurrentId(p.pGia, Node) {
        return; }
    Gia_ObjSetTravIdCurrentId(p.pGia, Node);
    pObj = Gia_ManObj( p.pGia, Node );
    if Gia_ObjIsAnd(pObj) {
        Sbd_ManWindowSim_rec( p, Gia_ObjFaninId0(pObj, Node) );
        Sbd_ManWindowSim_rec( p, Gia_ObjFaninId1(pObj, Node) );
    }
    if !pObj.fMark0 {
        Vec_IntWriteEntry( p.vObj2Var, Node, Vec_IntSize(p.vWinObjs) );
        Vec_IntPush( p.vWinObjs, Node );
    }
    if Gia_ObjIsCi(pObj) {
        return; }
    // simulate
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if Gia_ObjIsXor(pObj) {
        Abc_TtXor( Sbd_ObjSim0(p, Node), 
            Sbd_ObjSim0(p, Gia_ObjFaninId0(pObj, Node)), 
            Sbd_ObjSim0(p, Gia_ObjFaninId1(pObj, Node)), 
            p.pPars.nWords, 
            Gia_ObjFaninC0(pObj) ^ Gia_ObjFaninC1(pObj) );

        if pObj.fMark0 {
            Abc_TtXor( Sbd_ObjSim1(p, Node), 
                Gia_ObjFanin0(pObj).fMark0 ? Sbd_ObjSim1(p, Gia_ObjFaninId0(pObj, Node)) : Sbd_ObjSim0(p, Gia_ObjFaninId0(pObj, Node)), 
                Gia_ObjFanin1(pObj).fMark0 ? Sbd_ObjSim1(p, Gia_ObjFaninId1(pObj, Node)) : Sbd_ObjSim0(p, Gia_ObjFaninId1(pObj, Node)), 
                p.pPars.nWords, 
                Gia_ObjFaninC0(pObj) ^ Gia_ObjFaninC1(pObj) ); }
    
    } else {
        Abc_TtAndCompl( Sbd_ObjSim0(p, Node), 
            Sbd_ObjSim0(p, Gia_ObjFaninId0(pObj, Node)), Gia_ObjFaninC0(pObj), 
            Sbd_ObjSim0(p, Gia_ObjFaninId1(pObj, Node)), Gia_ObjFaninC1(pObj), 
            p.pPars.nWords );

        if pObj.fMark0 {
            Abc_TtAndCompl( Sbd_ObjSim1(p, Node), 
                Gia_ObjFanin0(pObj).fMark0 ? Sbd_ObjSim1(p, Gia_ObjFaninId0(pObj, Node)) : Sbd_ObjSim0(p, Gia_ObjFaninId0(pObj, Node)), Gia_ObjFaninC0(pObj), 
                Gia_ObjFanin1(pObj).fMark0 ? Sbd_ObjSim1(p, Gia_ObjFaninId1(pObj, Node)) : Sbd_ObjSim0(p, Gia_ObjFaninId1(pObj, Node)), Gia_ObjFaninC1(pObj), 
                p.pPars.nWords ); }
    }
    if Node != NodeInit {
        Abc_TtCopy( Sbd_ObjSim0(p, NodeInit), Sbd_ObjSim0(p, Node), p.pPars.nWords, Abc_LitIsCompl(Vec_IntEntry(p.vMirrors, NodeInit)) ); }
}
func Sbd_ManWindow(p *Sbd_Man_t, Pivot int32) int32 {
    var clk abctime  = Abc_Clock();
    var i, Node int32;
    // assign pivot and TFO (assume siminfo is assigned at the PIs)
    p.Pivot = Pivot;
    p.vTfo = Vec_WecEntry( p.vTfos, Pivot );
    // add constant node
    Vec_IntClear( p.vWinObjs );
    Vec_IntWriteEntry( p.vObj2Var, 0, Vec_IntSize(p.vWinObjs) );
    Vec_IntPush( p.vWinObjs, 0 );
    // simulate TFI cone
    Gia_ManIncrementTravId( p.pGia );
    Gia_ObjSetTravIdCurrentId(p.pGia, 0);
    Sbd_ManWindowSim_rec( p, Pivot );
    if p.pPars.nWinSizeMax && Vec_IntSize(p.vWinObjs) > p.pPars.nWinSizeMax {
        p.timeWin += Abc_Clock() - clk;
        return 0;
    }
    Sbd_ManUpdateOrder( p, Pivot );
    assert(Vec_IntSize(p.vDivVars) == Vec_IntSize(p.vDivValues), "Vec_IntSize(p.vDivVars) == Vec_IntSize(p.vDivValues)");
    assert(Vec_IntSize(p.vDivVars) < Vec_IntSize(p.vWinObjs), "Vec_IntSize(p.vDivVars) < Vec_IntSize(p.vWinObjs)");
    // simulate node
    Gia_ManObj(p.pGia, Pivot).fMark0 = 1;
    Abc_TtCopy( Sbd_ObjSim1(p, Pivot), Sbd_ObjSim0(p, Pivot), p.pPars.nWords, 1 );
    // mark TFO and simulate extended TFI without adding TFO nodes
    Vec_IntClear( p.vRoots );
    Vec_IntForEachEntry( p.vTfo, Node, i )
    {
        Gia_ManObj(p.pGia, Abc_Lit2Var(Node)).fMark0 = 1;
        if !Abc_LitIsCompl(Node) { 
            continue; }
        Sbd_ManWindowSim_rec( p, Abc_Lit2Var(Node) );
        Vec_IntPush( p.vRoots, Abc_Lit2Var(Node) );
    }
    // add TFO nodes and remove marks
    Gia_ManObj(p.pGia, Pivot).fMark0 = 0;
    Vec_IntForEachEntry( p.vTfo, Node, i )
    {
        Gia_ManObj(p.pGia, Abc_Lit2Var(Node)).fMark0 = 0;
        Vec_IntWriteEntry( p.vObj2Var, Abc_Lit2Var(Node), Vec_IntSize(p.vWinObjs) );
        Vec_IntPush( p.vWinObjs, Abc_Lit2Var(Node) );
    }
    if p.pPars.nWinSizeMax && Vec_IntSize(p.vWinObjs) > p.pPars.nWinSizeMax {
        p.timeWin += Abc_Clock() - clk;
        return 0;
    }
    // compute controlability for node
    if Vec_IntSize(p.vTfo) == 0 {
        Abc_TtFill( Sbd_ObjSim2(p, Pivot), p.pPars.nWords );
    } else {
        Abc_TtClear( Sbd_ObjSim2(p, Pivot), p.pPars.nWords ); }
    Vec_IntForEachEntry( p.vTfo, Node, i )
        if ( Abc_LitIsCompl(Node) ) // root
            Abc_TtOrXor( Sbd_ObjSim2(p, Pivot), Sbd_ObjSim0(p, Abc_Lit2Var(Node)), Sbd_ObjSim1(p, Abc_Lit2Var(Node)), p.pPars.nWords );
    p.timeWin += Abc_Clock() - clk;
    // propagate controlability to fanins for the TFI nodes starting from the pivot
    Sbd_ManPropagateControl( p, Pivot );
    assert(Vec_IntSize(p.vDivValues) <= 64, "Vec_IntSize(p.vDivValues) <= 64");
    return int32((Vec_IntSize(p.vDivValues) <= 64));
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ManCheckConst(p *Sbd_Man_t, Pivot int32) int32 {
    var nMintCount int32  = 1;
    var vSims *Vec_Ptr_t;
    var pSims *word  = Sbd_ObjSim0( p, Pivot );
    var pCtrl *word  = Sbd_ObjSim2( p, Pivot );
    var PivotVar int32  = Vec_IntEntry(p.vObj2Var, Pivot);
    var RetValue int32;
    var i int32;
    var iObj int32;
    var Ind int32;
    var fFindOnset int32;
    var nCares [2]int32 ;

    var clk abctime  = Abc_Clock();
    p.pSat = Sbd_ManSatSolver( p.pSat, p.pGia, p.vMirrors, Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, 0 );
    p.timeCnf += Abc_Clock() - clk;
    if p.pSat == nil {
        //if ( p->pPars->fVerbose )
        //    printf( "Found stuck-at-%d node %d.\n", 0, Pivot );
        Vec_IntWriteEntry( p.vLutLevs, Pivot, 0 );
        p.nLuts[0]++;
        return 0;
    }
    //return -1;
    //Sbd_ManPrintObj( p, Pivot );

    // count the number of on-set and off-set care-set minterms
    Vec_IntClear( p.vLits );
    for  i = 0; i < 64; i++  {
        if Abc_TtGetBit(pCtrl, i) {
            nCares[Abc_TtGetBit(pSims, i)]++;
        } else {
            Vec_IntPush( p.vLits, i ); }
    }
    fFindOnset = int32((nCares[0] < nCares[1]));
    if nCares[0] >= nMintCount && nCares[1] >= nMintCount {
        return -1; }
    // find how many do we need
    nCares[0] = nCares[0] < nMintCount ? nMintCount - nCares[0] : 0;
    nCares[1] = nCares[1] < nMintCount ? nMintCount - nCares[1] : 0;

    if p.pPars.fVeryVerbose {
        fmt.Printf("Computing %d offset and %d onset minterms for node %d.\n", nCares[0], nCares[1], Pivot); }

    if Vec_IntSize(p.vLits) >= nCares[0] + nCares[1] {
        Vec_IntShrink( p.vLits, nCares[0] + nCares[1] );
    } else {
        // collect places to insert new minterms
        for  i = 0; i < 64 && Vec_IntSize(p.vLits) < nCares[0] + nCares[1]; i++  {
            if fFindOnset == Abc_TtGetBit(pSims, i) {
                Vec_IntPush( p.vLits, i ); }
        }
    }
    // collect simulation pointers
    vSims = Vec_PtrAlloc( PivotVar + 1 );
    Vec_IntForEachEntry( p.vWinObjs, iObj, i )
    {
        Vec_PtrPush( vSims, Sbd_ObjSim0(p, iObj) );
        if iObj == Pivot {
            break; }
    }
    assert(i == PivotVar, "i == PivotVar");
    // compute patterns
    RetValue = Sbd_ManCollectConstants( p.pSat, nCares, PivotVar, (word **)Vec_PtrArray(vSims), p.vLits );
    // print computed miterms
    if false && RetValue < 0 {
        var vPis *Vec_Int_t  = Vec_WecEntry(p.vDivLevels, 0);
        var i, k, Ind int32;
        fmt.Printf("Additional minterms:\n");
        Vec_IntForEachEntry( p.vLits, Ind, k )
        {
            for  i = 0; i < Vec_IntSize(vPis); i++  {
                fmt.Printf("%d", Abc_TtGetBit( (word *)Vec_PtrEntry(vSims, Vec_IntEntry(p.vWinObjs, i)), Ind ));
            }
            fmt.Printf("\n");
        }
    }
    Vec_PtrFree( vSims );
    if RetValue >= 0 {
        if p.pPars.fVeryVerbose {
            fmt.Printf("Found stuck-at-%d node %d.\n", RetValue, Pivot); }
        Vec_IntWriteEntry( p.vLutLevs, Pivot, 0 );
        p.nLuts[0]++;
        return RetValue;
    }
    // set controlability of these minterms
    Vec_IntForEachEntry( p.vLits, Ind, i )
        Abc_TtSetBit( pCtrl, Ind );
    // propagate controlability to fanins for the TFI nodes starting from the pivot
    Sbd_ManPropagateControl( p, Pivot );
    // double check that we now have enough minterms
    for  i = 0; i < 64; i++  {
        if Abc_TtGetBit(pCtrl, i) {
            nCares[Abc_TtGetBit(pSims, i)]++; }
    }
    assert(nCares[0] >= nMintCount && nCares[1] >= nMintCount, "nCares[0] >= nMintCount && nCares[1] >= nMintCount");
    return -1;
}

/**Function*************************************************************

  Synopsis    [Transposing 64-bit matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_TransposeMatrix64(A []word) {
    var j, k int32;
    var t word;
    var m word = 0x00000000FFFFFFFF;
    for  j = 32; j != 0; j = j >> 1, m = m ^ (m << j)  {
        for  k = 0; k < 64; k = (k + j + 1) & ~j  {
            t = (A[k] ^ (A[k+j] >> j)) & m;
            A[k] = A[k] ^ t;
            A[k+j] = A[k+j] ^ (t << j);
        }
    }
}
func Sbd_PrintMatrix64(A []word) {
    var j, k int32;
    for  j = 0; j < 64; j++, printf("\n")  {
    for  k = 0; k < 64; k++  {
        fmt.Printf("%d", int32(((A[j] >> k) & 1)));
    }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Profiling divisor candidates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ManPrintObj(p *Sbd_Man_t, Pivot int32) {
    var nDivs int32  = Vec_IntEntry(p.vObj2Var, Pivot) + 1;
    var i, k, k0, k1, Id, Bit0, Bit1 int32;

    Vec_IntForEachEntryStop( p.vWinObjs, Id, i, nDivs )
        fmt.Printf("%3d : ", Id);
Extra_PrintBinary( os.Stdout, (unsigned *)Sbd_ObjSim0(p, Id), 64 );
fmt.Printf("\n");

    assert(p.Pivot == Pivot, "p.Pivot == Pivot");
    Vec_IntClear( p.vCounts[0] );
    Vec_IntClear( p.vCounts[1] );

    fmt.Printf("Node %d.  Useful divisors = %d.\n", Pivot, Vec_IntSize(p.vDivValues));
    fmt.Printf("Lev : ");
    Vec_IntForEachEntryStop( p.vWinObjs, Id, i, nDivs )
    {
        if i == nDivs-1 {
            fmt.Printf(" "); }
        fmt.Printf("%d", Vec_IntEntry(p.vLutLevs, Id));
    }
    fmt.Printf("\n");
    fmt.Printf("\n");

    if nDivs > 99 {
        fmt.Printf("    : ");
        Vec_IntForEachEntryStop( p.vWinObjs, Id, i, nDivs )
        {
            if i == nDivs-1 {
                fmt.Printf(" "); }
            fmt.Printf("%d", Id / 100);
        }
        fmt.Printf("\n");
    }
    if nDivs > 9 {
        fmt.Printf("    : ");
        Vec_IntForEachEntryStop( p.vWinObjs, Id, i, nDivs )
        {
            if i == nDivs-1 {
                fmt.Printf(" "); }
            fmt.Printf("%d", (Id % 100) / 10);
        }
        fmt.Printf("\n");
    }
    if nDivs > 0 {
        fmt.Printf("    : ");
        Vec_IntForEachEntryStop( p.vWinObjs, Id, i, nDivs )
        {
            if i == nDivs-1 {
                fmt.Printf(" "); }
            fmt.Printf("%d", Id % 10);
        }
        fmt.Printf("\n");
        fmt.Printf("\n");
    }

    // sampling matrix
    for  k = 0; k < p.pPars.nWords * 64; k++  {
        if !Abc_TtGetBit(Sbd_ObjSim2(p, Pivot), k) {
            continue; }

        fmt.Printf("%3d : ", k);
        Vec_IntForEachEntryStop( p.vWinObjs, Id, i, nDivs )
        {
            var pSims *word  = Sbd_ObjSim0( p, Id );
            var pCtrl *word  = Sbd_ObjSim2( p, Id );
            if i == nDivs-1 {
                if Abc_TtGetBit(pCtrl, k) {
                    Vec_IntPush( p.vCounts[Abc_TtGetBit(pSims, k)], k ); }
                fmt.Printf(" ");
            }
            fmt.Printf("%c", Abc_TtGetBit(pCtrl, k) ? '0' + Abc_TtGetBit(pSims, k) : '.');
        }
        fmt.Printf("\n");

        fmt.Printf("%3d : ", k);
        Vec_IntForEachEntryStop( p.vWinObjs, Id, i, nDivs )
        {
            var pSims *word  = Sbd_ObjSim0( p, Id );
            var pCtrl *word  = Sbd_ObjSim3( p, Id );
            if i == nDivs-1 {
                if Abc_TtGetBit(pCtrl, k) {
                    Vec_IntPush( p.vCounts[Abc_TtGetBit(pSims, k)], k ); }
                fmt.Printf(" ");
            }
            fmt.Printf("%c", Abc_TtGetBit(pCtrl, k) ? '0' + Abc_TtGetBit(pSims, k) : '.');
        }
        fmt.Printf("\n");

        fmt.Printf("Sims: ");
        Vec_IntForEachEntryStop( p.vWinObjs, Id, i, nDivs )
        {
            var pSims *word  = Sbd_ObjSim0( p, Id );
            //word * pCtrl = Sbd_ObjSim2( p, Id );
            if i == nDivs-1 {
                fmt.Printf(" "); }
            fmt.Printf("%c", '0' + Abc_TtGetBit(pSims, k));
        }
        fmt.Printf("\n");

        fmt.Printf("Ctrl: ");
        Vec_IntForEachEntryStop( p.vWinObjs, Id, i, nDivs )
        {
            //word * pSims = Sbd_ObjSim0( p, Id );
            var pCtrl *word  = Sbd_ObjSim2( p, Id );
            if i == nDivs-1 {
                fmt.Printf(" "); }
            fmt.Printf("%c", '0' + Abc_TtGetBit(pCtrl, k));
        }
        fmt.Printf("\n");


        fmt.Printf("\n");
    }
    // covering table
    fmt.Printf("Exploring %d x %d covering table.\n", Vec_IntSize(p.vCounts[0]), Vec_IntSize(p.vCounts[1]));
/*
    Vec_IntForEachEntryStop( p->vCounts[0], Bit0, k0, Abc_MinInt(Vec_IntSize(p->vCounts[0]), 8) )
    Vec_IntForEachEntryStop( p->vCounts[1], Bit1, k1, Abc_MinInt(Vec_IntSize(p->vCounts[1]), 8) )
    {
        printf( "%3d %3d : ", Bit0, Bit1 );
        Vec_IntForEachEntryStop( p->vWinObjs, Id, i, nDivs )
        {
            word * pSims = Sbd_ObjSim0( p, Id );
            word * pCtrl = Sbd_ObjSim2( p, Id );
            if ( i == nDivs-1 )
                printf( " " );
            printf( "%c", (Abc_TtGetBit(pCtrl, Bit0) && Abc_TtGetBit(pCtrl, Bit1) && Abc_TtGetBit(pSims, Bit0) != Abc_TtGetBit(pSims, Bit1)) ? '1' : '.' );
        }
        printf( "\n" );
    }
*/
    Vec_WrdClear( p.vMatrix );
    Vec_IntForEachEntryStop( p.vCounts[0], Bit0, k0, Abc_MinInt(Vec_IntSize(p.vCounts[0]), 64) )
    Vec_IntForEachEntryStop( p.vCounts[1], Bit1, k1, Abc_MinInt(Vec_IntSize(p.vCounts[1]), 64) )
    {
        var Row word;
        Vec_IntForEachEntryStop( p.vWinObjs, Id, i, nDivs )
        {
            var pSims *word  = Sbd_ObjSim0( p, Id );
            var pCtrl *word  = Sbd_ObjSim2( p, Id );
            if Abc_TtGetBit(pCtrl, Bit0) && Abc_TtGetBit(pCtrl, Bit1) && Abc_TtGetBit(pSims, Bit0) != Abc_TtGetBit(pSims, Bit1) {
                Abc_TtXorBit( &Row, i ); }
        }
        if Vec_WrdPushUnique( p.vMatrix, Row ) {
            continue; }
        for  i = 0; i < nDivs; i++  {
            fmt.Printf("%d", int32(((Row >> i) & 1)));
        }
        fmt.Printf("\n");
    }
}

func Sbd_ManMatrPrint(p *Sbd_Man_t, Cover []word, nCol int32, nRows int32) {
    var i, k int32;
    for  i = 0; i <= nCol; i++  {
        fmt.Printf("%2d : ", i);
        fmt.Printf("%d ", i == nCol ? Vec_IntEntry(p.vLutLevs, p.Pivot) : Vec_IntEntry(p.vLutLevs, Vec_IntEntry(p.vWinObjs, Vec_IntEntry(p.vDivVars, i))));
        for  k = 0; k < nRows; k++  {
            fmt.Printf("%d", int32(((Cover[i] >> k) & 1)));
        }
        fmt.Printf("\n"); 
    }
    fmt.Printf("\n");
}
func Sbd_ManCoverReverseOrder(Cover []word) {
    var i int32;
    for  i = 0; i < 32; i++  {
        var Cube word  = Cover[i];
        Cover[i] = Cover[63-i]; 
        Cover[63-i] = Cube;
    }
}

func Sbd_ManAddCube1(nRowLimit int32, Cover []word, nRows int32, Cube word) int32 {
    var n, m int32;
    if false {
        fmt.Printf("Adding cube: ");
        for  n = 0; n < nRowLimit; n++  {
            fmt.Printf("%d", int32(((Cube >> n) & 1)));
        }
        fmt.Printf("\n");
    }
    // do not add contained Cube
    assert(nRows <= nRowLimit, "nRows <= nRowLimit");
    for  n = 0; n < nRows; n++  {
        if (Cover[n] & Cube) == Cover[n] { // Cube is contained
            return nRows; }
    }
    // remove rows contained by Cube
    for  n = m = 0; n < nRows; n++  {
        if (Cover[n] & Cube) != Cube { // Cover[n] is not contained
            Cover[m++] = Cover[n]; }
    }
    if m < nRowLimit {
        Cover[m++] = Cube; }
    for  n = m; n < nRows; n++  {
        Cover[n] = 0;
    }
    nRows = m;
    return nRows;
}
func Sbd_ManAddCube2(Cover [][64]word, nRows int32, Cube []word) int32 {
    var n, m int32;
    // do not add contained Cube
    assert(nRows <= 64, "nRows <= 64");
    for  n = 0; n < nRows; n++  {
        if (Cover[0][n] & Cube[0]) == Cover[0][n] && (Cover[1][n] & Cube[1]) == Cover[1][n] { // Cube is contained
            return nRows; }
    }
    // remove rows contained by Cube
    for  n = m = 0; n < nRows; n++  {
        if (Cover[0][n] & Cube[0]) != Cube[0] || (Cover[1][n] & Cube[1]) != Cube[1] { // Cover[n] is not contained
        
            Cover[0][m] = Cover[0][n];
            Cover[1][m] = Cover[1][n];
            m++;
        }
    }
    if m < 64 {
        Cover[0][m] = Cube[0];
        Cover[1][m] = Cube[1];
        m++;
    }
    for  n = m; n < nRows; n++  {
        Cover[0][n] = Cover[1][n] = 0;
    }
    nRows = m;
    return nRows;
}

func Sbd_ManFindCandsSimple(p *Sbd_Man_t, Cover []word, nDivs int32) int32 {
    var c0, c1, c2, c3 int32;
    var Target word  = Cover[nDivs];
    Vec_IntClear( p.vDivSet );
    for  c0 = 0;    c0 < nDivs; c0++  {
        if Cover[c0] == Target {
            Vec_IntPush( p.vDivSet, c0 );
            return 1;
        }
    }

    for  c0 = 0;    c0 < nDivs; c0++  {
    for  c1 = c0+1; c1 < nDivs; c1++  {
        if (Cover[c0] | Cover[c1]) == Target {
            Vec_IntPush( p.vDivSet, c0 );
            Vec_IntPush( p.vDivSet, c1 );
            return 1;
        }
    }
    }

    for  c0 = 0;    c0 < nDivs; c0++  {
    for  c1 = c0+1; c1 < nDivs; c1++  {
    for  c2 = c1+1; c2 < nDivs; c2++  {
        if (Cover[c0] | Cover[c1] | Cover[c2]) == Target {
            Vec_IntPush( p.vDivSet, c0 );
            Vec_IntPush( p.vDivSet, c1 );
            Vec_IntPush( p.vDivSet, c2 );
            return 1;
        }
    }
    }
    }

    for  c0 = 0;    c0 < nDivs; c0++  {
    for  c1 = c0+1; c1 < nDivs; c1++  {
    for  c2 = c1+1; c2 < nDivs; c2++  {
    for  c3 = c2+1; c3 < nDivs; c3++  {
        if (Cover[c0] | Cover[c1] | Cover[c2] | Cover[c3]) == Target {
            Vec_IntPush( p.vDivSet, c0 );
            Vec_IntPush( p.vDivSet, c1 );
            Vec_IntPush( p.vDivSet, c2 );
            Vec_IntPush( p.vDivSet, c3 );
            return 1;
        }
    }
    }
    }
    }
    return 0;
}

func Sbd_ManFindCands(p *Sbd_Man_t, Cover []word, nDivs int32) int32 {
    var Ones, Order [64]int32;
    var Limits [4]int32 = [4]int32{ nDivs/4+1, nDivs/3+2, nDivs/2+3, nDivs };
    var c0, c1, c2, c3 int32;
    var Target word  = Cover[nDivs];

    if nDivs < 8 || p.pPars.fCover {
        return Sbd_ManFindCandsSimple( p, Cover, nDivs ); }

    Vec_IntClear( p.vDivSet );
    for  c0 = 0; c0 < nDivs; c0++  {
        if Cover[c0] == Target {
            Vec_IntPush( p.vDivSet, c0 );
            return 1;
        }
    }

    for  c0 = 0;    c0 < nDivs; c0++  {
    for  c1 = c0+1; c1 < nDivs; c1++  {
        if (Cover[c0] | Cover[c1]) == Target {
            Vec_IntPush( p.vDivSet, c0 );
            Vec_IntPush( p.vDivSet, c1 );
            return 1;
        }
    }
    }

    // count ones
    for  c0 = 0; c0 < nDivs; c0++  {
        Ones[c0] = Abc_TtCountOnes( Cover[c0] );
    }

    // sort by the number of ones
    for  c0 = 0; c0 < nDivs; c0++  {
        Order[c0] = c0;
    }
    Vec_IntSelectSortCost2Reverse( Order, nDivs, Ones );

    // sort with limits
    for  c0 = 0;    c0 < Limits[0]; c0++  {
    for  c1 = c0+1; c1 < Limits[1]; c1++  {
    for  c2 = c1+1; c2 < Limits[2]; c2++  {
        if (Cover[Order[c0]] | Cover[Order[c1]] | Cover[Order[c2]]) == Target {
            Vec_IntPush( p.vDivSet, Order[c0] );
            Vec_IntPush( p.vDivSet, Order[c1] );
            Vec_IntPush( p.vDivSet, Order[c2] );
            return 1;
        }
    }
    }
    }

    for  c0 = 0;    c0 < Limits[0]; c0++  {
    for  c1 = c0+1; c1 < Limits[1]; c1++  {
    for  c2 = c1+1; c2 < Limits[2]; c2++  {
    for  c3 = c2+1; c3 < Limits[3]; c3++  {
        if (Cover[Order[c0]] | Cover[Order[c1]] | Cover[Order[c2]] | Cover[Order[c3]]) == Target {
            Vec_IntPush( p.vDivSet, Order[c0] );
            Vec_IntPush( p.vDivSet, Order[c1] );
            Vec_IntPush( p.vDivSet, Order[c2] );
            Vec_IntPush( p.vDivSet, Order[c3] );
            return 1;
        }
    }
    }
    }
    }
    return 0;
}


func Sbd_ManExplore(p *Sbd_Man_t, Pivot int32, pTruth *word) int32 {
    var fVerbose int32;
    var clk abctime;
    var nIters int32;
    var nItersMax int32 = 32;

    var MatrS [64]word ;
    var MatrC [2][64]word = [2][64]word{{0}};
    var Cubes [2][2][64]word = [2][2][64]word{{{0}}};
    var Cover [64]word ;
    var Cube word;
    var CubeNew [2]word;
    var i int32;
    var k int32;
    var n int32;
    var Node int32;
    var Index int32;
    var nCubes [2]int32 ;
    var nRows int32;
    var nRowsOld int32;

    var nDivs int32  = Vec_IntSize(p.vDivValues);
    var PivotVar int32  = Vec_IntEntry(p.vObj2Var, Pivot);
    var FreeVar int32  = Vec_IntSize(p.vWinObjs) + Vec_IntSize(p.vTfo) + Vec_IntSize(p.vRoots);
    var RetValue int32;

    if p.pPars.fVerbose {
        fmt.Printf("Node %d.  Useful divisors = %d.\n", Pivot, nDivs); }

    if fVerbose != 0 {
        Sbd_ManPrintObj( p, Pivot ); }

    // collect bit-matrices
    Vec_IntForEachEntry( p.vDivVars, Node, i )
    {
        MatrS[63-i]    = *Sbd_ObjSim0( p, Vec_IntEntry(p.vWinObjs, Node) );
        MatrC[0][63-i] = *Sbd_ObjSim2( p, Vec_IntEntry(p.vWinObjs, Node) );
        MatrC[1][63-i] = *Sbd_ObjSim3( p, Vec_IntEntry(p.vWinObjs, Node) );
    }
    MatrS[63-i]    = *Sbd_ObjSim0( p, Pivot );
    MatrC[0][63-i] = *Sbd_ObjSim2( p, Pivot );
    MatrC[1][63-i] = *Sbd_ObjSim3( p, Pivot );

    //Sbd_PrintMatrix64( MatrS );
    Sbd_TransposeMatrix64( MatrS );
    Sbd_TransposeMatrix64( MatrC[0] );
    Sbd_TransposeMatrix64( MatrC[1] );
    //Sbd_PrintMatrix64( MatrS );

    // collect cubes
    for  i = 0; i < 64; i++  {
        assert(Abc_TtGetBit(&MatrC[0][i], nDivs) == Abc_TtGetBit(&MatrC[1][i], nDivs), "Abc_TtGetBit(&MatrC[0][i], nDivs) == Abc_TtGetBit(&MatrC[1][i], nDivs)");
        if !Abc_TtGetBit(&MatrC[0][i], nDivs) {
            continue; }
        Index = Abc_TtGetBit(&MatrS[i], nDivs); // Index==0 offset; Index==1 onset
        for  n = 0; n < 2; n++  {
            if (n != 0) && MatrC[0][i] == MatrC[1][i] {
                continue; }
            assert(MatrC[n][i], "MatrC[n][i]");
            CubeNew[0] = ~MatrS[i] & MatrC[n][i];
            CubeNew[1] =  MatrS[i] & MatrC[n][i];
            assert(CubeNew[0] || CubeNew[1], "CubeNew[0] || CubeNew[1]");
            nCubes[Index] = Sbd_ManAddCube2( Cubes[Index], nCubes[Index], CubeNew );
        }
    }

    if p.pPars.fVerbose {
        fmt.Printf("Generated matrix with %d x %d entries.\n", nCubes[0], nCubes[1]); }

    if p.pPars.fVerbose {
    for  n = 0; n < 2; n++  {
        fmt.Printf("%s:\n", n ? "Onset" : "Offset");
        for  i = 0; i < nCubes[n]; i++, printf( "\n" )  {
            for  k = 0; k < 64; k++  {
                if Abc_TtGetBit(&Cubes[n][0][i], k) {
                    fmt.Printf("0");
                } else if Abc_TtGetBit(&Cubes[n][1][i], k) {
                    fmt.Printf("1");
                } else { 
                    fmt.Printf("."); }
            }
        }
        fmt.Printf("\n");
    } }

    // create covering table
    nRows = 0;
    for  i = 0; i < nCubes[0] && nRows < 32; i++  {
    for  k = 0; k < nCubes[1] && nRows < 32; k++  {
        Cube = (Cubes[0][1][i] & Cubes[1][0][k]) | (Cubes[0][0][i] & Cubes[1][1][k]);
        assert(Cube, "Cube");
        nRows = Sbd_ManAddCube1( 64, Cover, nRows, Cube );
    }
    }

    Sbd_ManCoverReverseOrder( Cover );

    if p.pPars.fVerbose {
        fmt.Printf("Generated cover with %d entries.\n", nRows); }

    //if ( p->pPars->fVerbose )
    //Sbd_PrintMatrix64( Cover );
    Sbd_TransposeMatrix64( Cover );
    //if ( p->pPars->fVerbose )
    //Sbd_PrintMatrix64( Cover );

    Sbd_ManCoverReverseOrder( Cover );

    nRowsOld = nRows;
    for  nIters = 0; nIters < nItersMax && nRows < 64; nIters++  {
        if p.pPars.fVerbose {
            Sbd_ManMatrPrint( p, Cover, nDivs, nRows ); }

        clk = Abc_Clock();
        if !(Sbd_ManFindCands( p, Cover, nDivs ) != 0) {
            if p.pPars.fVerbose {
                fmt.Printf("Cannot find a feasible cover.\n"); }
            p.timeCov += Abc_Clock() - clk;
            return RetValue;
        }
        p.timeCov += Abc_Clock() - clk;
    
        if p.pPars.fVerbose {
            fmt.Printf("Candidate support:  ");
Vec_IntPrint( p.vDivSet ); }

        clk = Abc_Clock();
        *pTruth = Sbd_ManSolve( p.pSat, PivotVar, FreeVar+nIters, p.vDivSet, p.vDivVars, p.vDivValues, p.vLits );
        p.timeSat += Abc_Clock() - clk;

        if *pTruth == SBD_SAT_UNDEC {
            fmt.Printf("Node %d:  Undecided.\n", Pivot);
        } else if *pTruth == SBD_SAT_SAT {
            if p.pPars.fVerbose {
                var i int32;
                fmt.Printf("Node %d:  SAT.\n", Pivot);
                for  i = 0; i < nDivs; i++  {
                    fmt.Printf("%d", i % 10);
                }
                fmt.Printf("\n");
                for  i = 0; i < nDivs; i++  {
                    fmt.Printf("%c", (Vec_IntEntry(p.vDivValues, i) & 0x4) ? '0' + (Vec_IntEntry(p.vDivValues, i) & 1) : 'x');
                }
                fmt.Printf("\n");
                for  i = 0; i < nDivs; i++  {
                    fmt.Printf("%c", (Vec_IntEntry(p.vDivValues, i) & 0x8) ? '0' + ((Vec_IntEntry(p.vDivValues, i) >> 1) & 1) : 'x');
                }
                fmt.Printf("\n");
            }
            // add row to the covering table
            for  i = 0; i < nDivs; i++  {
                if Vec_IntEntry(p.vDivValues, i) == 0xE || Vec_IntEntry(p.vDivValues, i) == 0xD {
                    Cover[i] |= ((word)1 << nRows); }
            }
            Cover[nDivs] |= ((word)1 << nRows);
            nRows++;
        
        } else {
            if p.pPars.fVerbose {
                fmt.Printf("Node %d:  UNSAT.\n", Pivot);
                Extra_PrintBinary( os.Stdout, (unsigned *)pTruth, 1 << Vec_IntSize(p.vDivSet) );
fmt.Printf("\n");
            }
            RetValue = 1;
            break;
        }
        //break;
    }
    return RetValue;
}

func Sbd_ManExplore2(p *Sbd_Man_t, Pivot int32, pTruth *word) int32 {
    var clk abctime;
    var Onset [64]word ;
    var Offset [64]word ;
    var Cube word;
    var CoverRows [64]word ;
    var CoverCols [64]word ;
    var nIters int32;
    var nItersMax int32 = 32;
    var i, k, nRows int32;

    var PivotVar int32  = Vec_IntEntry(p.vObj2Var, Pivot);
    var FreeVar int32  = Vec_IntSize(p.vWinObjs) + Vec_IntSize(p.vTfo) + Vec_IntSize(p.vRoots);
    var nDivs int32  = Vec_IntSize( p.vDivVars );
    var nConsts int32  = 4;
    var RetValue int32;

    clk = Abc_Clock();
    //sat_solver_delete_p( &p->pSat );
    p.pSat = Sbd_ManSatSolver( p.pSat, p.pGia, p.vMirrors, Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, 0 );
    p.timeCnf += Abc_Clock() - clk;

    assert(nConsts <= 8, "nConsts <= 8");
    clk = Abc_Clock();
    RetValue = Sbd_ManCollectConstantsNew( p.pSat, p.vDivVars, nConsts, PivotVar, Onset, Offset );
    p.timeSat += Abc_Clock() - clk;
    if RetValue >= 0 {
        if p.pPars.fVeryVerbose {
            fmt.Printf("Found stuck-at-%d node %d.\n", RetValue, Pivot); }
        Vec_IntWriteEntry( p.vLutLevs, Pivot, 0 );
        p.nLuts[0]++;
        return RetValue;
    }
    RetValue = 0;

    // create rows of the table
    nRows = 0;
    for  i = 0; i < nConsts; i++  {
    for  k = 0; k < nConsts; k++  {
        Cube = Onset[i] ^ Offset[k];
        assert(Cube, "Cube");
        nRows = Sbd_ManAddCube1( 256, CoverRows, nRows, Cube );
    }
    }
    assert(nRows <= 64, "nRows <= 64");
    
    // create columns of the table
    for  i = 0; i < nRows; i++  {
        for  k = 0; k <= nDivs; k++  {
            if (CoverRows[i] >> k) & 1 {
                Abc_TtXorBit(&CoverCols[k], i); }
        }
    }

    // solve the covering problem
    for  nIters = 0; nIters < nItersMax && nRows < 64; nIters++  {
        if p.pPars.fVeryVerbose {
            Sbd_ManMatrPrint( p, CoverCols, nDivs, nRows ); }

        clk = Abc_Clock();
        if !(Sbd_ManFindCands( p, CoverCols, nDivs ) != 0) {
            if p.pPars.fVeryVerbose {
                fmt.Printf("Cannot find a feasible cover.\n"); }
            p.timeCov += Abc_Clock() - clk;
            return 0;
        }
        p.timeCov += Abc_Clock() - clk;

        if p.pPars.fVeryVerbose {
            fmt.Printf("Candidate support:  ");
Vec_IntPrint( p.vDivSet ); }

        clk = Abc_Clock();
        *pTruth = Sbd_ManSolve( p.pSat, PivotVar, FreeVar+nIters, p.vDivSet, p.vDivVars, p.vDivValues, p.vLits );
        p.timeSat += Abc_Clock() - clk;

        if *pTruth == SBD_SAT_UNDEC {
            fmt.Printf("Node %d:  Undecided.\n", Pivot);
        } else if *pTruth == SBD_SAT_SAT {
            if p.pPars.fVeryVerbose {
                var i int32;
                fmt.Printf("Node %d:  SAT.\n", Pivot);
                for  i = 0; i < nDivs; i++  {
                    fmt.Printf("%d", Vec_IntEntry(p.vLutLevs, Vec_IntEntry(p.vWinObjs, Vec_IntEntry(p.vDivVars, i))));
                }
                fmt.Printf("\n");
                for  i = 0; i < nDivs; i++  {
                    fmt.Printf("%d", i % 10);
                }
                fmt.Printf("\n");
                for  i = 0; i < nDivs; i++  {
                    fmt.Printf("%c", (Vec_IntEntry(p.vDivValues, i) & 0x4) ? '0' + (Vec_IntEntry(p.vDivValues, i) & 1) : 'x');
                }
                fmt.Printf("\n");
                for  i = 0; i < nDivs; i++  {
                    fmt.Printf("%c", (Vec_IntEntry(p.vDivValues, i) & 0x8) ? '0' + ((Vec_IntEntry(p.vDivValues, i) >> 1) & 1) : 'x');
                }
                fmt.Printf("\n");
            }
            // add row to the covering table
            for  i = 0; i < nDivs; i++  {
                if Vec_IntEntry(p.vDivValues, i) == 0xE || Vec_IntEntry(p.vDivValues, i) == 0xD {
                    CoverCols[i] |= ((word)1 << nRows); }
            }
            CoverCols[nDivs] |= ((word)1 << nRows);
            nRows++;
        
        } else {
            if p.pPars.fVeryVerbose {
                fmt.Printf("Node %d:  UNSAT.   ", Pivot);
                Extra_PrintBinary( os.Stdout, (unsigned *)pTruth, 1 << Vec_IntSize(p.vDivSet) );
fmt.Printf("\n");
            }
            p.nLuts[1]++;
            RetValue = 1;
            break;
        }
    }
    return RetValue;
}

func Sbd_ManExploreCut(p *Sbd_Man_t, Pivot int32, nLeaves int32, pLeaves *int32, pnStrs *int32, Strs *Sbd_Str_t, pFreeVar *int32) int32 {
    var clk abctime  = Abc_Clock();
    var PivotVar int32  = Vec_IntEntry(p.vObj2Var, Pivot);
    var Delay int32  = Vec_IntEntry( p.vLutLevs, Pivot );
    var pNodesTop, pNodesBot, pNodesBot1, pNodesBot2 [SBD_DIV_MAX]int32;
    var nNodesTop, nNodesBot, nNodesBot1, nNodesBot2, nNodesDiff, nNodesDiff1, nNodesDiff2 int32;
    var i, k, iObj, nIters, RetValue int32;

    // try to remove fanins
    for  nIters = 0; nIters < nLeaves; nIters++  {
        var Truth word;
        // try to remove one variable from divisors
        Vec_IntClear( p.vDivSet );
        for  i = 0; i < nLeaves; i++  {
            if i != nLeaves-1-nIters && pLeaves[i] != -1 {
                Vec_IntPush( p.vDivSet, Vec_IntEntry(p.vObj2Var, pLeaves[i]) ); }
        }
        assert(Vec_IntSize(p.vDivSet) < nLeaves, "Vec_IntSize(p.vDivSet) < nLeaves");
        // compute truth table
        clk = Abc_Clock();
        Truth = Sbd_ManSolve( p.pSat, PivotVar, (*pFreeVar)++, p.vDivSet, p.vDivVars, p.vDivValues, p.vLits );
        p.timeSat += Abc_Clock() - clk;
        if Truth == SBD_SAT_UNDEC {
            fmt.Printf("Node %d:  Undecided.\n", Pivot);
        } else if Truth == SBD_SAT_SAT {
            var DelayDiff int32  = Vec_IntEntry(p.vLutLevs, pLeaves[nLeaves-1-nIters]) - Delay;
            if DelayDiff > -2 {
                return 0; }
        
        } else {
            pLeaves[nLeaves-1-nIters] = -1; }
    }
    Vec_IntClear( p.vDivSet );
    for  i = 0; i < nLeaves; i++  {
        if pLeaves[i] != -1 {
            Vec_IntPush( p.vDivSet, pLeaves[i] ); }
    }
    //printf( "Reduced %d -> %d\n", nLeaves, Vec_IntSize(p->vDivSet) );
    if Vec_IntSize(p.vDivSet) <= p.pPars.nLutSize {
        var Truth word;
        *pnStrs = 1;
        // remap divisors
        Vec_IntForEachEntry( p.vDivSet, iObj, i )
            Vec_IntWriteEntry( p.vDivSet, i, Vec_IntEntry(p.vObj2Var, iObj) );
        // compute truth table
        clk = Abc_Clock();
        Truth = Sbd_ManSolve( p.pSat, PivotVar, (*pFreeVar)++, p.vDivSet, p.vDivVars, p.vDivValues, p.vLits );
        p.timeSat += Abc_Clock() - clk;
        if Truth == SBD_SAT_SAT {
            fmt.Printf("The cut at node %d is not topological.\n", p.Pivot);
            return 0;
        }
        assert(Truth != SBD_SAT_UNDEC && Truth != SBD_SAT_SAT, "Truth != SBD_SAT_UNDEC && Truth != SBD_SAT_SAT");
        // create structure
        Strs.fLut = 1;
        Strs.nVarIns = Vec_IntSize( p.vDivSet );
        for  i = 0; i < Strs.nVarIns; i++  {
            Strs.VarIns[i] = i;
        }
        Strs.Res = Truth;
        p.nLuts[1]++;
        //Extra_PrintBinary( stdout, (unsigned *)&Truth, 1 << Strs->nVarIns ), printf( "\n" );
        return 1;
    }
    assert(Vec_IntSize(p.vDivSet) > p.pPars.nLutSize, "Vec_IntSize(p.vDivSet) > p.pPars.nLutSize");

    // count number of nodes on each level
    nNodesTop = nNodesBot = nNodesBot1 = nNodesBot2 = 0;
    Vec_IntForEachEntry( p.vDivSet, iObj, i )
    {
        var DelayDiff int32  = Vec_IntEntry(p.vLutLevs, iObj) - Delay;
        if DelayDiff > -2 {
            break; }
        if DelayDiff == -2 {
            pNodesTop[nNodesTop++] = i;
        } else { // if ( DelayDiff < -2 )
        
            pNodesBot[nNodesBot++] = i;
            if DelayDiff == -3 {
                pNodesBot1[nNodesBot1++] = i;
            } else { // if ( DelayDiff < -3 )
                pNodesBot2[nNodesBot2++] = i; }
        }
        Vec_IntWriteEntry( p.vDivSet, i, Vec_IntEntry(p.vObj2Var, iObj) );
    }
    assert(nNodesBot == nNodesBot1 + nNodesBot2, "nNodesBot == nNodesBot1 + nNodesBot2");
    if i < Vec_IntSize(p.vDivSet) {
        return 0; }
    if nNodesTop > p.pPars.nLutSize-1 {
        return 0; }

    // try 44
    if Vec_IntSize(p.vDivSet) <= 2*p.pPars.nLutSize-1 {
        var nMoved int32;
        if nNodesBot > p.pPars.nLutSize { // need to move bottom left-over to the top
        
            for  nNodesBot > p.pPars.nLutSize  {
                pNodesTop[nNodesTop++] = pNodesBot[--nNodesBot], nMoved++;
            }
            assert(nNodesBot == p.pPars.nLutSize, "nNodesBot == p.pPars.nLutSize");
        }
        assert(nNodesBot <= p.pPars.nLutSize, "nNodesBot <= p.pPars.nLutSize");
        assert(nNodesTop <= p.pPars.nLutSize-1, "nNodesTop <= p.pPars.nLutSize-1");

        Strs[0].fLut = 1;
        Strs[0].nVarIns = p.pPars.nLutSize;
        for  i = 0; i < nNodesTop; i++  {
            Strs[0].VarIns[i] = pNodesTop[i];
        }
        for  ; i < p.pPars.nLutSize; i++  {
            Strs[0].VarIns[i] = Vec_IntSize(p.vDivSet)+1 + i-nNodesTop;
        }
        Strs[0].Res = 0;

        Strs[1].fLut = 1;
        Strs[1].nVarIns = nNodesBot;
        for  i = 0; i < nNodesBot; i++  {
            Strs[1].VarIns[i] = pNodesBot[i];
        }
        Strs[1].Res = 0;

        nNodesDiff = p.pPars.nLutSize-1 - nNodesTop;
        assert(nNodesDiff >= 0 && nNodesDiff <= 3, "nNodesDiff >= 0 && nNodesDiff <= 3");
        for  k = 0; k < nNodesDiff; k++  {
            Strs[2+k].fLut = 0;
            Strs[2+k].nVarIns = nNodesBot;
            for  i = 0; i < nNodesBot; i++  {
                Strs[2+k].VarIns[i] = pNodesBot[i];
            }
            Strs[2+k].Res = 0;
        }

        *pnStrs = 2 + nNodesDiff;
        clk = Abc_Clock();
        RetValue = Sbd_ProblemSolve( p.pGia, p.vMirrors,   Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots,   p.vDivSet, *pnStrs, Strs );
        p.timeQbf += Abc_Clock() - clk;
        if RetValue != 0 { 
            p.nLuts[2]++; }

        for  nMoved--  {
            pNodesBot[nNodesBot++] = pNodesTop[--nNodesTop];
        }
    }

    if RetValue != 0 {
        return RetValue; }
    if p.pPars.nLutNum < 3 {
        return 0; }
    if Vec_IntSize(p.vDivSet) < 2*p.pPars.nLutSize-1 {
        return 0; }

    // try 444 -- LUT(LUT, LUT)
    if nNodesTop <= p.pPars.nLutSize-2 {
        var nMoved int32;
        if nNodesBot > 2*p.pPars.nLutSize { // need to move bottom left-over to the top
        
            for  nNodesBot > 2*p.pPars.nLutSize  {
                pNodesTop[nNodesTop++] = pNodesBot[--nNodesBot], nMoved++;
            }
            assert(nNodesBot == 2*p.pPars.nLutSize, "nNodesBot == 2*p.pPars.nLutSize");
        }
        assert(nNodesBot > p.pPars.nLutSize, "nNodesBot > p.pPars.nLutSize");
        assert(nNodesBot <= 2*p.pPars.nLutSize, "nNodesBot <= 2*p.pPars.nLutSize");
        assert(nNodesTop <= p.pPars.nLutSize-2, "nNodesTop <= p.pPars.nLutSize-2");

        Strs[0].fLut = 1;
        Strs[0].nVarIns = p.pPars.nLutSize;
        for  i = 0; i < nNodesTop; i++  {
            Strs[0].VarIns[i] = pNodesTop[i];
        }
        for  ; i < p.pPars.nLutSize; i++  {
            Strs[0].VarIns[i] = Vec_IntSize(p.vDivSet)+1 + i-nNodesTop;
        }
        Strs[0].Res = 0;

        Strs[1].fLut = 1;
        Strs[1].nVarIns = p.pPars.nLutSize;
        for  i = 0; i < Strs[1].nVarIns; i++  {
            Strs[1].VarIns[i] = pNodesBot[i];
        }
        Strs[1].Res = 0;

        Strs[2].fLut = 1;
        Strs[2].nVarIns = p.pPars.nLutSize;
        for  i = 0; i < Strs[2].nVarIns; i++  {
            Strs[2].VarIns[i] = pNodesBot[nNodesBot-p.pPars.nLutSize+i];
        }
        Strs[2].Res = 0;

        nNodesDiff = p.pPars.nLutSize-2 - nNodesTop;
        assert(nNodesDiff >= 0 && nNodesDiff <= 2, "nNodesDiff >= 0 && nNodesDiff <= 2");
        for  k = 0; k < nNodesDiff; k++  {
            Strs[3+k].fLut = 0;
            Strs[3+k].nVarIns = nNodesBot;
            for  i = 0; i < nNodesBot; i++  {
                Strs[3+k].VarIns[i] = pNodesBot[i];
            }
            Strs[3+k].Res = 0;
        }

        *pnStrs = 3 + nNodesDiff;
        clk = Abc_Clock();
        RetValue = Sbd_ProblemSolve( p.pGia, p.vMirrors,   Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots,   p.vDivSet, *pnStrs, Strs );
        p.timeQbf += Abc_Clock() - clk;
        if RetValue != 0 { 
            p.nLuts[3]++; }

        for  nMoved--  {
            pNodesBot[nNodesBot++] = pNodesTop[--nNodesTop];
        }
    }
    if RetValue != 0 {
        return RetValue; }

    // try 444 -- LUT(LUT(LUT))
    if nNodesBot1 + nNodesTop <= 2*p.pPars.nLutSize-2 {
        if nNodesBot2 > p.pPars.nLutSize { // need to move bottom left-over to the top
        
            for  nNodesBot2 > p.pPars.nLutSize  {
                pNodesBot1[nNodesBot1++] = pNodesBot2[--nNodesBot2];
            }
            assert(nNodesBot2 == p.pPars.nLutSize, "nNodesBot2 == p.pPars.nLutSize");
        }
        if nNodesBot1 > p.pPars.nLutSize-1 { // need to move bottom left-over to the top
        
            for  nNodesBot1 > p.pPars.nLutSize-1  {
                pNodesTop[nNodesTop++] = pNodesBot1[--nNodesBot1];
            }
            assert(nNodesBot1 == p.pPars.nLutSize-1, "nNodesBot1 == p.pPars.nLutSize-1");
        }
        assert(nNodesBot2 <= p.pPars.nLutSize, "nNodesBot2 <= p.pPars.nLutSize");
        assert(nNodesBot1 <= p.pPars.nLutSize-1, "nNodesBot1 <= p.pPars.nLutSize-1");
        assert(nNodesTop <= p.pPars.nLutSize-1, "nNodesTop <= p.pPars.nLutSize-1");

        Strs[0].fLut = 1;
        Strs[0].nVarIns = p.pPars.nLutSize;
        for  i = 0; i < nNodesTop; i++  {
            Strs[0].VarIns[i] = pNodesTop[i];
        }
        Strs[0].VarIns[i++] = Vec_IntSize(p.vDivSet)+1;
        for  ; i < p.pPars.nLutSize; i++  {
            Strs[0].VarIns[i] = Vec_IntSize(p.vDivSet)+2 + i-nNodesTop;
        }
        Strs[0].Res = 0;
        nNodesDiff1 = p.pPars.nLutSize-1 - nNodesTop;

        Strs[1].fLut = 1;
        Strs[1].nVarIns = p.pPars.nLutSize;
        for  i = 0; i < nNodesBot1; i++  {
            Strs[1].VarIns[i] = pNodesBot1[i];
        }
        Strs[1].VarIns[i++] = Vec_IntSize(p.vDivSet)+2;
        for  ; i < p.pPars.nLutSize; i++  {
            Strs[1].VarIns[i] = Vec_IntSize(p.vDivSet)+2+nNodesDiff1 + i-nNodesBot1;
        }
        Strs[1].Res = 0;
        nNodesDiff2 = p.pPars.nLutSize-1 - nNodesBot1;

        Strs[2].fLut = 1;
        Strs[2].nVarIns = nNodesBot2;
        for  i = 0; i < Strs[2].nVarIns; i++  {
            Strs[2].VarIns[i] = pNodesBot2[i];
        }
        Strs[2].Res = 0;

        nNodesDiff = nNodesDiff1 + nNodesDiff2;
        assert(nNodesDiff >= 0 && nNodesDiff <= 3, "nNodesDiff >= 0 && nNodesDiff <= 3");
        for  k = 0; k < nNodesDiff; k++  {
            Strs[3+k].fLut = 0;
            Strs[3+k].nVarIns = nNodesBot2;
            for  i = 0; i < nNodesBot2; i++  {
                Strs[3+k].VarIns[i] = pNodesBot2[i];
            }
            Strs[3+k].Res = 0;
            if k >= nNodesDiff1 {
                continue; }
            Strs[3+k].nVarIns += nNodesBot1;
            for  i = 0; i < nNodesBot1; i++  {
                Strs[3+k].VarIns[nNodesBot2 + i] = pNodesBot1[i];
            }
        }

        *pnStrs = 3 + nNodesDiff;
        clk = Abc_Clock();
        RetValue = Sbd_ProblemSolve( p.pGia, p.vMirrors,   Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots,   p.vDivSet, *pnStrs, Strs );
        p.timeQbf += Abc_Clock() - clk;
        if RetValue != 0 { 
            p.nLuts[4]++; }
    }
    return RetValue;
}
func Sbd_ManExplore3(p *Sbd_Man_t, Pivot int32, pnStrs *int32, Strs *Sbd_Str_t) int32 {
    var FreeVar int32  = Vec_IntSize(p.vWinObjs) + Vec_IntSize(p.vTfo) + Vec_IntSize(p.vRoots);
    var FreeVarStart int32  = FreeVar;
    var nSize int32;
    var nLeaves int32;
    var pLeaves [SBD_DIV_MAX]int32;
    //sat_solver_delete_p( &p->pSat );
    var clk abctime  = Abc_Clock();
    p.pSat = Sbd_ManSatSolver( p.pSat, p.pGia, p.vMirrors, Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, 0 );
    p.timeCnf += Abc_Clock() - clk;
    // extract one cut
    if p.pSrv {
        nLeaves = Sbd_ManCutServerFirst( p.pSrv, Pivot, pLeaves );
        if nLeaves == -1 {
            return 0; }
        assert(nLeaves <= p.pPars.nCutSize, "nLeaves <= p.pPars.nCutSize");
        if Sbd_ManExploreCut( p, Pivot, nLeaves, pLeaves, pnStrs, Strs, &FreeVar ) != 0 {
            return 1; }
        return 0;
    }
    // extract one cut
    for  nSize = p.pPars.nLutSize + 1; nSize <= p.pPars.nCutSize; nSize++  {
        nLeaves = Sbd_StoObjBestCut( p.pSto, Pivot, nSize, pLeaves );
        if nLeaves == -1 {
            continue; }
        assert(nLeaves == nSize, "nLeaves == nSize");
        if Sbd_ManExploreCut( p, Pivot, nLeaves, pLeaves, pnStrs, Strs, &FreeVar ) != 0 {
            return 1; }
    }
    assert(FreeVar - FreeVarStart <= SBD_FVAR_MAX, "FreeVar - FreeVarStart <= SBD_FVAR_MAX"); 
    return 0;
}


/**Function*************************************************************

  Synopsis    [Computes delay-oriented k-feasible cut at the node.]

  Description [Return 1 if node's LUT level does not exceed those of the fanins.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_CutMergeSimple(p *Sbd_Man_t, pCut1 *int32, pCut2 *int32, pCut *int32) int32 {
    var pBeg *int32   = pCut + 1;
    var pBeg1 *int32  = pCut1 + 1;
    var pBeg2 *int32  = pCut2 + 1;
    var pEnd1 *int32  = pCut1 + 1 + pCut1[0];
    var pEnd2 *int32  = pCut2 + 1 + pCut2[0];
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            *pBeg++ = *pBeg1++, pBeg2++;
        } else if *pBeg1 < *pBeg2 {
            *pBeg++ = *pBeg1++;
        } else { 
            *pBeg++ = *pBeg2++; }
    }
    for  pBeg1 < pEnd1  {
        *pBeg++ = *pBeg1++;
    }
    for  pBeg2 < pEnd2  {
        *pBeg++ = *pBeg2++;
    }
    return (pCut[0] = pBeg - pCut - 1);
}
/*
int Sbd_ManMergeCuts( Sbd_Man_t * p, int Node )
{
    int Result  = 1; // no need to resynthesize
    int pCut[2*SBD_MAX_LUTSIZE+1];     
    int iFan0   = Gia_ObjFaninId0( Gia_ManObj(p->pGia, Node), Node );
    int iFan1   = Gia_ObjFaninId1( Gia_ManObj(p->pGia, Node), Node );
    int Level0  = Vec_IntEntry( p->vLutLevs, iFan0 );
    int Level1  = Vec_IntEntry( p->vLutLevs, iFan1 );
    int LevMax  = (Level0 || Level1) ? Abc_MaxInt(Level0, Level1) : 1;
    int * pCut0 = Sbd_ObjCut( p, iFan0 );
    int * pCut1 = Sbd_ObjCut( p, iFan1 );
    int nSize   = Sbd_CutMergeSimple( p, pCut0, pCut1, pCut );
    if ( nSize > p->pPars->nLutSize )
    {
        if ( Level0 != Level1 )
        {
            int Cut0[2] = {1, iFan0}, * pCut0Temp = Level0 < LevMax ? Cut0 : pCut0; 
            int Cut1[2] = {1, iFan1}, * pCut1Temp = Level1 < LevMax ? Cut1 : pCut1; 
            nSize = Sbd_CutMergeSimple( p, pCut0Temp, pCut1Temp, pCut );
        }
        if ( nSize > p->pPars->nLutSize )
        {
            pCut[0] = 2;
            pCut[1] = iFan0 < iFan1 ? iFan0 : iFan1;
            pCut[2] = iFan0 < iFan1 ? iFan1 : iFan0;
            Result  = LevMax ? 0 : 1;
            LevMax++;
        }
    }
    assert( iFan0 != iFan1 );
    assert( Vec_IntEntry(p->vLutLevs, Node) == 0 );
    Vec_IntWriteEntry( p->vLutLevs, Node, LevMax );
    memcpy( Sbd_ObjCut(p, Node), pCut, sizeof(int) * (pCut[0] + 1) );
    //printf( "Setting node %d with delay %d (result = %d).\n", Node, LevMax, Result );
    return Result;
}
*/
func Sbd_ManMergeCuts(p *Sbd_Man_t, Node int32) int32 {
    var pCut11 [2*SBD_MAX_LUTSIZE+1]int32;     
    var pCut01 [2*SBD_MAX_LUTSIZE+1]int32;     
    var pCut10 [2*SBD_MAX_LUTSIZE+1]int32;     
    var pCut00 [2*SBD_MAX_LUTSIZE+1]int32;     
    var iFan0 int32    = Gia_ObjFaninId0( Gia_ManObj(p.pGia, Node), Node );
    var iFan1 int32    = Gia_ObjFaninId1( Gia_ManObj(p.pGia, Node), Node );
    var Level0 int32;
    if Vec_IntEntry( p.vLutLevs, iFan0 ) {
        Level0 = Vec_IntEntry( p.vLutLevs, iFan0 );
    } else {
        Level0 = 1;
    }
    var Level1 int32;
    if Vec_IntEntry( p.vLutLevs, iFan1 ) {
        Level1 = Vec_IntEntry( p.vLutLevs, iFan1 );
    } else {
        Level1 = 1;
    }
    var pCut0 *int32  = Sbd_ObjCut( p, iFan0 );
    var pCut1 *int32  = Sbd_ObjCut( p, iFan1 );
    var Cut0 [2]int32 = [2]int32{1, iFan0}; 
    var Cut1 [2]int32 = [2]int32{1, iFan1}; 
    var nSize11 int32  = Sbd_CutMergeSimple( p, pCut0, pCut1, pCut11 );
    var nSize01 int32  = Sbd_CutMergeSimple( p,  Cut0, pCut1, pCut01 );
    var nSize10 int32  = Sbd_CutMergeSimple( p, pCut0,  Cut1, pCut10 );
    var nSize00 int32  = Sbd_CutMergeSimple( p,  Cut0,  Cut1, pCut00 );
    var Lev11 int32;
    if nSize11 <= p.pPars.nLutSize {
        Lev11 = Abc_MaxInt(Level0,   Level1);
    } else {
        Lev11 = ABC_INFINITY;
    }
    var Lev01 int32;
    if nSize01 <= p.pPars.nLutSize {
        Lev01 = Abc_MaxInt(Level0+1, Level1);
    } else {
        Lev01 = ABC_INFINITY;
    }
    var Lev10 int32;
    if nSize10 <= p.pPars.nLutSize {
        Lev10 = Abc_MaxInt(Level0,   Level1+1);
    } else {
        Lev10 = ABC_INFINITY;
    }
    var Lev00 int32;
    if nSize00 <= p.pPars.nLutSize {
        Lev00 = Abc_MaxInt(Level0+1, Level1+1);
    } else {
        Lev00 = ABC_INFINITY;
    }
    var pCutRes *int32  = pCut11;
    var LevCur int32     = Lev11;
    if Lev01 < LevCur || (Lev01 == LevCur && pCut01[0] < pCutRes[0]) {
        pCutRes = pCut01;
        LevCur  = Lev01;
    }
    if Lev10 < LevCur || (Lev10 == LevCur && pCut10[0] < pCutRes[0]) {
        pCutRes = pCut10;
        LevCur  = Lev10;
    }
    if Lev00 < LevCur || (Lev00 == LevCur && pCut00[0] < pCutRes[0]) {
        pCutRes = pCut00;
        LevCur  = Lev00;
    }
    assert(iFan0 != iFan1, "iFan0 != iFan1");
    assert(Vec_IntEntry(p.vLutLevs, Node) == 0, "Vec_IntEntry(p.vLutLevs, Node) == 0");
    Vec_IntWriteEntry( p.vLutLevs, Node, LevCur );
    //Vec_IntWriteEntry( p->vLevs, Node, 1+Abc_MaxInt(Vec_IntEntry(p->vLevs, iFan0), Vec_IntEntry(p->vLevs, iFan1)) );
    assert(pCutRes[0] <= p.pPars.nLutSize, "pCutRes[0] <= p.pPars.nLutSize");
    memcpy( Sbd_ObjCut(p, Node), pCutRes, sizeof(int) * (pCutRes[0] + 1) );
//printf( "Setting node %d with delay %d.\n", Node, LevCur );
    return LevCur == 1; // LevCur == Abc_MaxInt(Level0, Level1);
}
func Sbd_ManDelay(p *Sbd_Man_t) int32 {
    var i, Id, Delay int32;
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && (((Id) = Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))), 1); i++  {
        Delay = Abc_MaxInt( Delay, Vec_IntEntry(p.vLutLevs, Id) );
    }
    return Delay;
}
func Sbd_ManMergeTest(p *Sbd_Man_t) {
    var Node int32;
    // inlining of macro Gia_ManForEachAndId
    for  Node = 0; (Node < p.pGia.nObjs); Node++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, Node)) {} else {
        Sbd_ManMergeCuts( p, Node ); }
    }
    fmt.Printf("Delay %d.\n", Sbd_ManDelay(p));
}

func Sbd_ManFindCut_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if pObj.fMark1 {
        return; }
    pObj.fMark1 = 1;
    if pObj.fMark0 {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Sbd_ManFindCut_rec( p, Gia_ObjFanin0(pObj) );
    Sbd_ManFindCut_rec( p, Gia_ObjFanin1(pObj) );
}
func Sbd_ManFindCutUnmark_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if !pObj.fMark1 {
        return; }
    pObj.fMark1 = 0;
    if pObj.fMark0 {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Sbd_ManFindCutUnmark_rec( p, Gia_ObjFanin0(pObj) );
    Sbd_ManFindCutUnmark_rec( p, Gia_ObjFanin1(pObj) );
}
func Sbd_ManFindCut(p *Sbd_Man_t, Node int32, vCutLits *Vec_Int_t) {
    var pCut [SBD_MAX_LUTSIZE+1]int32;     
    var i, LevelMax int32;
    // label reachable nodes 
    var pTemp *Gia_Obj_t;
    var pObj *Gia_Obj_t = Gia_ManObj(p.pGia, Node);
    Sbd_ManFindCut_rec( p.pGia, pObj );
    // collect 
    pCut[0] = 0;
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCutLits)) && ((pTemp) = Gia_ManObj(p.pGia, Vec_IntEntry(vCutLits,i))); i++  {
        if pTemp.fMark1 {
            LevelMax = Abc_MaxInt( LevelMax, Vec_IntEntry(p.vLutLevs, Gia_ObjId(p.pGia, pTemp)) );
            pCut[1+pCut[0]++] = Gia_ObjId(p.pGia, pTemp);
        }
    }
    assert(pCut[0] <= p.pPars.nLutSize, "pCut[0] <= p.pPars.nLutSize");
    // unlabel reachable nodes
    Sbd_ManFindCutUnmark_rec( p.pGia, pObj );
    // create cut
    assert(Vec_IntEntry(p.vLutLevs, Node) == 0, "Vec_IntEntry(p.vLutLevs, Node) == 0");
    Vec_IntWriteEntry( p.vLutLevs, Node, LevelMax+1 );
    //Vec_IntWriteEntry( p->vLevs, Node, 1+Abc_MaxInt(Vec_IntEntry(p->vLevs, Gia_ObjFaninId0(pObj, Node)), Vec_IntEntry(p->vLevs, Gia_ObjFaninId1(pObj, Node))) );
    memcpy( Sbd_ObjCut(p, Node), pCut, sizeof(int) * (pCut[0] + 1) );
}

func Sbd_ManImplement(p *Sbd_Man_t, Pivot int32, Truth word) int32 {
    var pObj *Gia_Obj_t;
    var i, k, w, iLit, Entry, Node int32;
    var iObjLast int32  = Gia_ManObjNum(p.pGia);
    var iCurLev int32  = Vec_IntEntry(p.vLutLevs, Pivot);
    var iNewLev int32;
    // collect leaf literals
    Vec_IntClear( p.vLits );
    Vec_IntForEachEntry( p.vDivSet, Node, i )
    {
        Node = Vec_IntEntry( p.vWinObjs, Node );
        if Vec_IntEntry(p.vMirrors, Node) >= 0 {
            Vec_IntPush( p.vLits, Vec_IntEntry(p.vMirrors, Node) );
        } else {
            Vec_IntPush( p.vLits, Abc_Var2Lit(Node, 0) ); }
    }
    // pretend to have MUXes
//    assert( p->pGia->pMuxes == NULL );
    if p.pGia.nXors && p.pGia.pMuxes == nil {
        p.pGia.pMuxes = (unsigned *)p; }
    // derive new function of the node
    iLit = Dsm_ManTruthToGia( p.pGia, &Truth, p.vLits, p.vCover );
    if p.pGia.pMuxes == (unsigned *)p {
        p.pGia.pMuxes = nil; }
    // remember this function
    assert(Vec_IntEntry(p.vMirrors, Pivot) == -1, "Vec_IntEntry(p.vMirrors, Pivot) == -1");
    Vec_IntWriteEntry( p.vMirrors, Pivot, iLit );
    if p.pPars.fVerbose {
        fmt.Printf("Replacing node %d by literal %d.\n", Pivot, iLit); }
    // translate literals into variables
    Vec_IntForEachEntry( p.vLits, Entry, i )
        Vec_IntWriteEntry( p.vLits, i, Abc_Lit2Var(Entry) );
    // label inputs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vLits)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vLits,i))); i++  {
        pObj.fMark0 = 1;
    }
    // extend data-structure to accommodate new nodes
    assert(Vec_IntSize(p.vLutLevs) == iObjLast, "Vec_IntSize(p.vLutLevs) == iObjLast");
    for  i = iObjLast; i < Gia_ManObjNum(p.pGia); i++  {
        Vec_IntPush( p.vLutLevs, 0 );
        Vec_IntPush( p.vObj2Var, 0 );
        Vec_IntPush( p.vMirrors, -1 );
        Vec_IntFillExtra( p.vLutCuts, Vec_IntSize(p.vLutCuts) + p.pPars.nLutSize + 1, 0 );
        Sbd_ManFindCut( p, i, p.vLits );
        for  k = 0; k < 4; k++  {
            for  w = 0; w < p.pPars.nWords; w++  {
                Vec_WrdPush( p.vSims[k], 0 );
            }
        }
    }
    // unlabel inputs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vLits)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vLits,i))); i++  {
        pObj.fMark0 = 0;
    }
    // make sure delay reduction is achieved
    iNewLev = Vec_IntEntry( p.vLutLevs, Abc_Lit2Var(iLit) );
    assert(iNewLev < iCurLev, "iNewLev < iCurLev");
    // update delay of the initial node
    assert(Vec_IntEntry(p.vLutLevs, Pivot) == iCurLev, "Vec_IntEntry(p.vLutLevs, Pivot) == iCurLev");
    Vec_IntWriteEntry( p.vLutLevs, Pivot, iNewLev );
    //Vec_IntWriteEntry( p->vLevs, Pivot, 1+Abc_MaxInt(Vec_IntEntry(p->vLevs, Gia_ObjFaninId0(pObj, Pivot)), Vec_IntEntry(p->vLevs, Gia_ObjFaninId1(pObj, Pivot))) );
    return 0;
}

func Sbd_ManImplement2(p *Sbd_Man_t, Pivot int32, nStrs int32, pStrs *Sbd_Str_t) int32 {
    //Gia_Obj_t * pObj = NULL;
    var i, k, w, iLit, Node int32;
    var iObjLast int32  = Gia_ManObjNum(p.pGia);
    var iCurLev int32  = Vec_IntEntry(p.vLutLevs, Pivot);
    var iNewLev int32;
    // collect leaf literals
    Vec_IntClear( p.vLits );
    Vec_IntForEachEntry( p.vDivSet, Node, i )
    {
        Node = Vec_IntEntry( p.vWinObjs, Node );
        if Vec_IntEntry(p.vMirrors, Node) >= 0 {
            Vec_IntPush( p.vLits, Vec_IntEntry(p.vMirrors, Node) );
        } else {
            Vec_IntPush( p.vLits, Abc_Var2Lit(Node, 0) ); }
    }
    // collect structure nodes
    for  i = 0; i < nStrs; i++  {
        Vec_IntPush( p.vLits, -1 );
    }
    // implement structures
    for  i = nStrs-1; i >= 0; i--  {
        if pStrs[i].fLut {
            // collect local literals
            Vec_IntClear( p.vLits2 );
            for  k = 0; k < int32(pStrs[i].nVarIns); k++  {
                Vec_IntPush( p.vLits2, Vec_IntEntry(p.vLits, pStrs[i].VarIns[k]) );
            }
            // pretend to have MUXes
        //    assert( p->pGia->pMuxes == NULL );
            if p.pGia.nXors && p.pGia.pMuxes == nil {
                p.pGia.pMuxes = (unsigned *)p; }
            // derive new function of the node
            iLit = Dsm_ManTruthToGia( p.pGia, &pStrs[i].Res, p.vLits2, p.vCover );
            if p.pGia.pMuxes == (unsigned *)p {
                p.pGia.pMuxes = nil; }
        
        } else {
            iLit = Vec_IntEntry( p.vLits, int32(pStrs[i].Res) );
            assert(iLit > 0, "iLit > 0");
        }
        // update literal
        assert(Vec_IntEntry(p.vLits, Vec_IntSize(p.vLits)-nStrs+i) == -1, "Vec_IntEntry(p.vLits, Vec_IntSize(p.vLits)-nStrs+i) == -1");
        Vec_IntWriteEntry( p.vLits, Vec_IntSize(p.vLits)-nStrs+i, iLit );
    }
    iLit = Vec_IntEntry( p.vLits, Vec_IntSize(p.vDivSet) );
    //assert( iObjLast == Gia_ManObjNum(p->pGia) || Abc_Lit2Var(iLit) == Gia_ManObjNum(p->pGia)-1 );
    // remember this function
    assert(Vec_IntEntry(p.vMirrors, Pivot) == -1, "Vec_IntEntry(p.vMirrors, Pivot) == -1");
    Vec_IntWriteEntry( p.vMirrors, Pivot, iLit );
    if p.pPars.fVeryVerbose {
        fmt.Printf("Replacing node %d by literal %d.\n", Pivot, iLit); }

    // extend data-structure to accommodate new nodes
    assert(Vec_IntSize(p.vLutLevs) == iObjLast, "Vec_IntSize(p.vLutLevs) == iObjLast");
    for  i = iObjLast; i < Gia_ManObjNum(p.pGia); i++  {
        assert(i == Vec_IntSize(p.vMirrors), "i == Vec_IntSize(p.vMirrors)");
        Vec_IntPush( p.vMirrors, -1 );
        Sbd_StoRefObj( p.pSto, i, i == Gia_ManObjNum(p.pGia)-1 ? Pivot : -1 );
    }
    Sbd_StoDerefObj( p.pSto, Pivot );
    for  i = iObjLast; i < Gia_ManObjNum(p.pGia); i++  {
        //Gia_Obj_t * pObjI = Gia_ManObj( p->pGia, i );
        var clk abctime  = Abc_Clock();
        var Delay int32  = Sbd_StoComputeCutsNode( p.pSto, i );
        p.timeCut += Abc_Clock() - clk;
        assert(i == Vec_IntSize(p.vLutLevs), "i == Vec_IntSize(p.vLutLevs)");
        Vec_IntPush( p.vLutLevs, Delay );
        //Vec_IntPush( p->vLevs, 1+Abc_MaxInt(Vec_IntEntry(p->vLevs, Gia_ObjFaninId0(pObjI, i)), Vec_IntEntry(p->vLevs, Gia_ObjFaninId1(pObjI, i))) );
        Vec_IntPush( p.vObj2Var, 0 );
        Vec_IntFillExtra( p.vLutCuts, Vec_IntSize(p.vLutCuts) + p.pPars.nLutSize + 1, 0 );
        Sbd_StoSaveBestDelayCut( p.pSto, i, Sbd_ObjCut(p, i) );
        //Sbd_ManFindCut( p, i, p->vLits );
        for  k = 0; k < 4; k++  {
            for  w = 0; w < p.pPars.nWords; w++  {
                Vec_WrdPush( p.vSims[k], 0 );
            }
        }
    }
    // make sure delay reduction is achieved
    iNewLev = Vec_IntEntry( p.vLutLevs, Abc_Lit2Var(iLit) );
    assert(!(iNewLev != 0) || iNewLev < iCurLev, "!iNewLev || iNewLev < iCurLev");
    // update delay of the initial node
    //pObj = Gia_ManObj( p->pGia, Pivot );
    assert(Vec_IntEntry(p.vLutLevs, Pivot) == iCurLev, "Vec_IntEntry(p.vLutLevs, Pivot) == iCurLev");
    Vec_IntWriteEntry( p.vLutLevs, Pivot, iNewLev );
    //Vec_IntWriteEntry( p->vLevs, Pivot, Pivot ? 1+Abc_MaxInt(Vec_IntEntry(p->vLevs, Gia_ObjFaninId0(pObj, Pivot)), Vec_IntEntry(p->vLevs, Gia_ObjFaninId1(pObj, Pivot))) : 0 );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Derives new AIG after resynthesis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ManDeriveMapping_rec(p *Sbd_Man_t, pNew *Gia_Man_t, iObj int32) {
    var pObj *Gia_Obj_t; var k int32;
var pCut *int32;
    if !(iObj != 0) || Gia_ObjIsTravIdCurrentId(pNew, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(pNew, iObj);
    pObj = Gia_ManObj( pNew, iObj );
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    pCut = Sbd_ObjCut2( p, iObj );
    for  k = 1; k <= pCut[0]; k++  {
        Sbd_ManDeriveMapping_rec( p, pNew, pCut[k] );
    }
    // add mapping
    Vec_IntWriteEntry( pNew.vMapping, iObj, Vec_IntSize(pNew.vMapping) );
    for  k = 0; k <= pCut[0]; k++  {
        Vec_IntPush( pNew.vMapping, pCut[k] );
    }
    Vec_IntPush( pNew.vMapping, iObj );
}
func Sbd_ManDeriveMapping(p *Sbd_Man_t, pNew *Gia_Man_t) {
    var pObj, pFan *Gia_Obj_t; 
    var i int32;
    var k int32;
    var iFan int32;
    var iObjNew int32;
    var iFanNew int32;
    var pCut *int32;
    var pCutNew *int32;
    var vLeaves *Vec_Int_t  = Vec_IntAlloc( 100 );
    // derive cuts for the new manager
    p.vLutCuts2 = Vec_IntStart( Gia_ManObjNum(pNew) * (p.pPars.nLutSize + 1) );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Vec_IntEntry(p.vMirrors, i) >= 0 {
            continue; }
        if pObj.Value == ~0 {
            continue; }
        iObjNew = Abc_Lit2Var( pObj.Value );
        if !Gia_ObjIsAnd(Gia_ManObj(pNew, iObjNew)) {
            continue; }
        pCutNew = Sbd_ObjCut2( p, iObjNew );
        pCut    = Sbd_ObjCut( p, i );
        Vec_IntClear( vLeaves );
        for  k = 1; k <= pCut[0]; k++  {
            if Vec_IntEntry(p.vMirrors, pCut[k]) >= 0 {
                iFan = Abc_Lit2Var(Vec_IntEntry(p.vMirrors, pCut[k]));
            } else {
                iFan = pCut[k];
            }
            pFan = Gia_ManObj( p.pGia, iFan );
            if pFan.Value == ~0 {
                continue; }
            iFanNew = Abc_Lit2Var( pFan.Value );
            if iFanNew == 0 || iFanNew == iObjNew {
                continue; }
            Vec_IntPushUniqueOrder( vLeaves, iFanNew );
        }
        assert(Vec_IntSize(vLeaves) <= p.pPars.nLutSize, "Vec_IntSize(vLeaves) <= p.pPars.nLutSize");
        //assert( Vec_IntSize(vLeaves) > 1 );
        pCutNew[0] = Vec_IntSize(vLeaves);
        memcpy( pCutNew+1, Vec_IntArray(vLeaves), sizeof(int) * Vec_IntSize(vLeaves) );
    }
    }
    Vec_IntFree( vLeaves );
    // create new mapping
    Vec_IntFreeP( &pNew.vMapping );
    pNew.vMapping = Vec_IntAlloc( (p.pPars.nLutSize + 2) * Gia_ManObjNum(pNew) );
    Vec_IntFill( pNew.vMapping, Gia_ManObjNum(pNew), 0 );
    Gia_ManIncrementTravId( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pNew.vCos)) && ((pObj) = Gia_ManCo(pNew, i)); i++  {
        Sbd_ManDeriveMapping_rec( p, pNew, Gia_ObjFaninId0p(pNew, pObj) );
    }
    Vec_IntFreeP( &p.vLutCuts2 );
}
func Sbd_ManDerive_rec(pNew *Gia_Man_t, p *Gia_Man_t, Node int32, vMirrors *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var Obj int32  = Node;
    if Vec_IntEntry(vMirrors, Node) >= 0 {
        Obj = Abc_Lit2Var( Vec_IntEntry(vMirrors, Node) ); }
    pObj = Gia_ManObj( p, Obj );
    if !~pObj.Value {
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        Sbd_ManDerive_rec( pNew, p, Gia_ObjFaninId0(pObj, Obj), vMirrors );
        Sbd_ManDerive_rec( pNew, p, Gia_ObjFaninId1(pObj, Obj), vMirrors );
        if Gia_ObjIsXor(pObj) {
            pObj.Value = Gia_ManHashXorReal( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // set the original node as well
    if Obj != Node {
        Gia_ManObj(p, Node).Value = Abc_LitNotCond( pObj.Value, Abc_LitIsCompl(Vec_IntEntry(vMirrors, Node)) ); }
} 
func Sbd_ManDerive(pMan *Sbd_Man_t, p *Gia_Man_t, vMirrors *Vec_Int_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    if p.pMuxes {
        pNew.pMuxes = ABC_CALLOC( unsigned, Gia_ManObjNum(p) ); }
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Sbd_ManDerive_rec( pNew, p, Gia_ObjFaninId0p(p, pObj), vMirrors );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    Gia_ManTransferTiming( pNew, p );
    if pMan.pPars.fMapping {
        Sbd_ManDeriveMapping( pMan, pNew ); }
    // remove dangling nodes
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManTransferTiming( pNew, pTemp );
    Gia_ManTransferMapping( pNew, pTemp );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Performs delay optimization for the given LUT size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_NtkPerformOne(p *Sbd_Man_t, Pivot int32) {
    var Strs [SBD_DIV_MAX]Sbd_Str_t; var Truth word;
    var RetValue, nStrs int32;
    if !p.pSto && (Sbd_ManMergeCuts( p, Pivot ) != 0) {
        return; }
    if !(Sbd_ManWindow( p, Pivot ) != 0) {
        return; }
    //if ( Vec_IntSize(p->vWinObjs) > 100 )
    //    printf( "Obj %d : Win = %d   TFO = %d.  Roots = %d.\n", Pivot, Vec_IntSize(p->vWinObjs), Vec_IntSize(p->vTfo), Vec_IntSize(p->vRoots) );
    p.nTried++;
    p.nUsed++;
    RetValue = Sbd_ManCheckConst( p, Pivot );
    if RetValue >= 0 {
        Vec_IntWriteEntry( p.vMirrors, Pivot, RetValue );
        //if ( p->pPars->fVerbose ) printf( "Node %5d:  Detected constant %d.\n", Pivot, RetValue );
    
    } else if p.pPars.fFindDivs && p.pPars.nLutNum >= 1 && (Sbd_ManExplore2( p, Pivot, &Truth ) != 0) {
        var i int32;
        Strs.fLut = 1;
        Strs.nVarIns = Vec_IntSize( p.vDivSet );
        for  i = 0; i < Strs.nVarIns; i++  {
            Strs.VarIns[i] = i;
        }
        Strs.Res = Truth;
        Sbd_ManImplement2( p, Pivot, 1, Strs );
        //if ( p->pPars->fVerbose ) printf( "Node %5d:  Detected LUT%d\n", Pivot, p->pPars->nLutSize );
    
    } else if p.pPars.nLutNum >= 2 && (Sbd_ManExplore3( p, Pivot, &nStrs, Strs ) != 0) {
        Sbd_ManImplement2( p, Pivot, nStrs, Strs );
        if !p.pPars.fVerbose { 
            return; }
        //if ( Vec_IntSize(p->vDivSet) <= 4 )
        //    printf( "Node %5d:  Detected %d\n", Pivot, p->pPars->nLutSize );
        //else if ( Vec_IntSize(p->vDivSet) <= 6 || (Vec_IntSize(p->vDivSet) == 7 && nStrs == 2) )
        //    printf( "Node %5d:  Detected %d%d\n", Pivot, p->pPars->nLutSize, p->pPars->nLutSize );
        //else 
        //    printf( "Node %5d:  Detected %d%d%d\n", Pivot, p->pPars->nLutSize, p->pPars->nLutSize, p->pPars->nLutSize );
    
    } else {
        p.nUsed--; }
}
func Sbd_NtkPerform(pGia *Gia_Man_t, pPars *Sbd_Par_t) *Gia_Man_t {
    var pNew *Gia_Man_t;  
    var pObj *Gia_Obj_t;
    var vPath *Vec_Bit_t;
    var p *Sbd_Man_t  = Sbd_ManStart( pGia, pPars );
    var nNodesOld int32  = Gia_ManObjNum(pGia);
    var k, Pivot int32;
    assert(pPars.nLutSize <= 6, "pPars.nLutSize <= 6");
    // prepare references
    // inlining of macro Gia_ManForEachObj
    for  Pivot = 0; (Pivot < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, Pivot)); Pivot++  {
        Sbd_StoRefObj( p.pSto, Pivot, -1 );
    }
    //return NULL;
    if pPars.fUsePath && Gia_ManHasMapping(pGia) {
        vPath = Sbc_ManCriticalPath(pGia);
    } else {
        vPath = nil;
    }
    if pGia.pManTime != nil && Tim_ManBoxNum((Tim_Man_t*)pGia.pManTime) {
        var vNodes *Vec_Int_t  = Gia_ManOrderWithBoxes( pGia );
        var pTimOld *Tim_Man_t  = (Tim_Man_t *)pGia.pManTime;
        pGia.pManTime = Tim_ManDup( pTimOld, 1 );
        //Tim_ManPrint( pGia->pManTime );
        Tim_ManIncrementTravId( (Tim_Man_t *)pGia.pManTime );
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(pGia, Vec_IntEntry(vNodes,k))); k++  {
            Pivot = Gia_ObjId( pGia, pObj );
            if Pivot >= nNodesOld {
                break; }
            if Gia_ObjIsAnd(pObj) {
                var clk abctime  = Abc_Clock();
                var Delay int32  = Sbd_StoComputeCutsNode( p.pSto, Pivot );
                Sbd_StoSaveBestDelayCut( p.pSto, Pivot, Sbd_ObjCut(p, Pivot) );
                p.timeCut += Abc_Clock() - clk;
                Vec_IntWriteEntry( p.vLutLevs, Pivot, Delay );
                if Delay > 1 && (!(vPath != nil) || Vec_BitEntry(vPath, Pivot)) {
                    Sbd_NtkPerformOne( p, Pivot ); }
            
            } else if Gia_ObjIsCi(pObj) {
                var arrTime int32  = Tim_ManGetCiArrival( (Tim_Man_t*)pGia.pManTime, Gia_ObjCioId(pObj) );
                Vec_IntWriteEntry( p.vLutLevs, Pivot, arrTime );
                Sbd_StoComputeCutsCi( p.pSto, Pivot, arrTime, arrTime );
            
            } else if Gia_ObjIsCo(pObj) {
                var arrTime int32  = Vec_IntEntry( p.vLutLevs, Gia_ObjFaninId0(pObj, Pivot) );
                Tim_ManSetCoArrival( (Tim_Man_t*)pGia.pManTime, Gia_ObjCioId(pObj), arrTime );
            
            } else if Gia_ObjIsConst0(pObj) {
                Sbd_StoComputeCutsConst0( p.pSto, 0 );
            } else { assert(false, "0"); }
        }
        Tim_ManStop( (Tim_Man_t *)pGia.pManTime );
        pGia.pManTime = pTimOld;
        Vec_IntFree( vNodes );
    
    } else {
        Sbd_StoComputeCutsConst0( p.pSto, 0 );
        // inlining of macro Gia_ManForEachObj
        for  Pivot = 0; (Pivot < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, Pivot)); Pivot++  {
            if Pivot >= nNodesOld {
                break; }
            if Gia_ObjIsCi(pObj) {
                Sbd_StoComputeCutsCi( p.pSto, Pivot, 0, 0 );
            } else if Gia_ObjIsAnd(pObj) {
                var clk abctime  = Abc_Clock();
                var Delay int32  = Sbd_StoComputeCutsNode( p.pSto, Pivot );
                Sbd_StoSaveBestDelayCut( p.pSto, Pivot, Sbd_ObjCut(p, Pivot) );
                p.timeCut += Abc_Clock() - clk;
                Vec_IntWriteEntry( p.vLutLevs, Pivot, Delay );
                if Delay > 1 && (!(vPath != nil) || Vec_BitEntry(vPath, Pivot)) {
                    Sbd_NtkPerformOne( p, Pivot ); }
            }
            //if ( nNodesOld != Gia_ManObjNum(pGia) )
            //    break;
        }
    }
    Vec_BitFreeP( &vPath );
    p.timeTotal = Abc_Clock() - p.timeTotal;
    if p.pPars.fVerbose {
        fmt.Printf("K = %d. S = %d. N = %d. P = %d.  ", p.pPars.nLutSize, p.pPars.nLutNum, p.pPars.nCutSize, p.pPars.nCutNum);
        fmt.Printf("Try = %d. Use = %d.  C = %d. 1 = %d. 2 = %d. 3a = %d. 3b = %d.  Lev = %d.  ", p.nTried, p.nUsed, p.nLuts[0], p.nLuts[1], p.nLuts[2], p.nLuts[3], p.nLuts[4], Sbd_ManDelay(p));
        Abc_PrintTime( 1, "Time", p.timeTotal );
    }
    pNew = Sbd_ManDerive( p, pGia, p.vMirrors );
    // print runtime statistics
    p.timeOther = p.timeTotal - p.timeWin - p.timeCut - p.timeCov - p.timeCnf - p.timeSat - p.timeQbf;
    if p.pPars.fVerbose {
        ABC_PRTP( "Win", p.timeWin  ,  p.timeTotal );
        ABC_PRTP( "Cut", p.timeCut  ,  p.timeTotal );
        ABC_PRTP( "Cov", p.timeCov  ,  p.timeTotal );
        ABC_PRTP( "Cnf", p.timeCnf  ,  p.timeTotal );
        ABC_PRTP( "Sat", p.timeSat  ,  p.timeTotal );
        ABC_PRTP( "Qbf", p.timeQbf  ,  p.timeTotal );
        ABC_PRTP( "Oth", p.timeOther,  p.timeTotal );
        ABC_PRTP( "ALL", p.timeTotal,  p.timeTotal );
    }
    Sbd_ManStop( p );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

