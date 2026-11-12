package gia

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [giaSatLE.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Mapping with edges.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSatLE.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/extra/extra.h"
//#include "misc/tim/tim.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Sle_CutSize(pCut *int32) int32 { return pCut[0] & 0xF;  }  //  4 bits
func Sle_CutSign(pCut *int32) int32 { return (uint32(pCut[0])) >> 4;   }  // 28 bits
func Sle_CutSetSizeSign(s int32, S int32) int32 { return (S << 4) | s;   }
func Sle_CutLeaves(pCut *int32) *int32 { return pCut + 1;       } 

func Sle_CutIsUsed(pCut *int32) int32 { return pCut[1] != 0;   }
func Sle_CutSetUnused(pCut *int32) { pCut[1] = 0;           }

func Sle_ListCutNum(pList *int32) int32 { return pList[0];       }

// Original macro:
// #define Sle_ForEachCut( pList, pCut, i )   for ( i = 0, pCut = pList + 1; i <  pList[0]; i++, pCut += Sle_CutSize(pCut) + 1 ) // cuts with unit-cut

// Original macro:
// #define Sle_ForEachCut1( pList, pCut, i )  for ( i = 0, pCut = pList + 1; i <= pList[0]; i++, pCut += Sle_CutSize(pCut) + 1 ) // only non-unit cuts


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Cut computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_CutMergeOrder(pCut0 *int32, pCut1 *int32, pCut *int32, nLutSize int32) int32 { 
    var nSize0 int32    = Sle_CutSize(pCut0);
    var nSize1 int32    = Sle_CutSize(pCut1);
    var i int32;
    var pC0 *int32 = Sle_CutLeaves(pCut0);
    var k int32;
    var pC1 *int32 = Sle_CutLeaves(pCut1);
    var c int32;
    var pC *int32 = Sle_CutLeaves(pCut);
    // the case of the largest cut sizes
    if nSize0 == nLutSize && nSize1 == nLutSize {
        for  i = 0; i < nSize0; i++  {
            if pC0[i] != pC1[i] {  return 0; }
            pC[i] = pC0[i];
        }
        pCut[0] = Sle_CutSetSizeSign( nLutSize, Sle_CutSign(pCut0) | Sle_CutSign(pCut1) );
        return 1;
    }
    // compare two cuts with different numbers
    i = k = c = 0;
    if nSize0 == 0 { goto FlushCut1; }
    if nSize1 == 0 { goto FlushCut0; }
    for {
        if c == nLutSize { return 0; }
        if pC0[i] < pC1[k] {
            pC[c++] = pC0[i++];
            if i >= nSize0 { goto FlushCut1; }
        
        } else if pC0[i] > pC1[k] {
            pC[c++] = pC1[k++];
            if k >= nSize1 { goto FlushCut0; }
        
        } else {
            pC[c++] = pC0[i++]; k++;
            if i >= nSize0 { goto FlushCut1; }
            if k >= nSize1 { goto FlushCut0; }
        }
    }

FlushCut0:
    if c + nSize0 > nLutSize + i { return 0; }
    for  i < nSize0  {
        pC[c++] = pC0[i++];
    }
    pCut[0] = Sle_CutSetSizeSign( c, Sle_CutSign(pCut0) | Sle_CutSign(pCut1) );
    return 1;

FlushCut1:
    if c + nSize1 > nLutSize + k { return 0; }
    for  k < nSize1  {
        pC[c++] = pC1[k++];
    }
    pCut[0] = Sle_CutSetSizeSign( c, Sle_CutSign(pCut0) | Sle_CutSign(pCut1) );
    return 1;
}
static inline int Sle_SetCutIsContainedOrder( int * pBase, int * pCut ) // check if pCut is contained in pBase
{
    var i int32;
    var nSizeB int32 = Sle_CutSize(pBase);
    var k int32;
    var nSizeC int32 = Sle_CutSize(pCut);
    var pLeaveB *int32  = Sle_CutLeaves(pBase);
    var pLeaveC *int32  = Sle_CutLeaves(pCut);
    if nSizeB == nSizeC {
        for  i = 0; i < nSizeB; i++  {
            if pLeaveB[i] != pLeaveC[i] {
                return 0; }
        }
        return 1;
    }
    assert(nSizeB > nSizeC, "nSizeB > nSizeC"); 
    if nSizeC == 0 {
        return 1; }
    for  i = k = 0; i < nSizeB; i++  {
        if pLeaveB[i] > pLeaveC[k] {
            return 0; }
        if pLeaveB[i] == pLeaveC[k] {
            if k++; k == nSizeC {
                return 1; }
        }
    }
    return 0;
}
func Sle_CutCountBits(i uint32) int32 {
    i = i - ((i >> 1) & 0x55555555);
    i = (i & 0x33333333) + ((i >> 2) & 0x33333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F);
    return (i*(0x01010101))>>24;
}
func Sle_SetLastCutIsContained(vTemp *Vec_Int_t, pBase *int32) int32 {
    var i int32;
    var pCut *int32;
    var pList *int32 = Vec_IntArray(vTemp);
    // inlining of macro Sle_ForEachCut
    for  i, pCut = 0, pList + 1; i <  pList[0]; i++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

        if (Sle_CutIsUsed(pCut) != 0) && Sle_CutSize(pCut) <= Sle_CutSize(pBase) && (Sle_CutSign(pCut) & Sle_CutSign(pBase)) == Sle_CutSign(pCut) && Sle_SetCutIsContainedOrder(pBase, pCut) {
            return 1; }
    }
    return 0;
}
func Sle_SetAddCut(vTemp *Vec_Int_t, pCut *int32) {
    var i int32;
    var pBase *int32;
    var pList *int32 = Vec_IntArray(vTemp);
    // inlining of macro Sle_ForEachCut
    for  i, pBase = 0, pList + 1; i <  pList[0]; i++, pBase += Sle_CutSize(pBase) + 1  { // cuts with unit-cut

        if (Sle_CutIsUsed(pBase) != 0) && Sle_CutSize(pCut) < Sle_CutSize(pBase) && (Sle_CutSign(pCut) & Sle_CutSign(pBase)) == Sle_CutSign(pCut) && Sle_SetCutIsContainedOrder(pBase, pCut) {
            Sle_CutSetUnused( pBase ); }
    }
    Vec_IntPushArray( vTemp, pCut, Sle_CutSize(pCut)+1 );
    Vec_IntAddToEntry( vTemp, 0, 1 );
}
func Sle_ManCutMerge(p *Gia_Man_t, iObj int32, vCuts *Vec_Int_t, vTemp *Vec_Int_t, nLutSize int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pList0 *int32  = Vec_IntEntryP( vCuts, Vec_IntEntry(vCuts, Gia_ObjFaninId0(pObj, iObj)) );
    var pList1 *int32  = Vec_IntEntryP( vCuts, Vec_IntEntry(vCuts, Gia_ObjFaninId1(pObj, iObj)) );
    var pCut0 *int32;
    var pCut1 *int32;
    var i int32;
    var k int32;
    var Cut [8]int32;
    var nCuts int32;
    Vec_IntFill( vTemp, 1, 0 );
    // inlining of macro Sle_ForEachCut1
    for  i, pCut0 = 0, pList0 + 1; i <= pList0[0]; i++, pCut0 += Sle_CutSize(pCut0) + 1  { // only non-unit cuts

    // inlining of macro Sle_ForEachCut1
    for  k, pCut1 = 0, pList1 + 1; k <= pList1[0]; k++, pCut1 += Sle_CutSize(pCut1) + 1  { // only non-unit cuts

    
        if Sle_CutSize(pCut0) + Sle_CutSize(pCut1) > nLutSize && Sle_CutCountBits(Sle_CutSign(pCut0) | Sle_CutSign(pCut1)) > nLutSize {
            continue; }
        if !(Sle_CutMergeOrder(pCut0, pCut1, Cut, nLutSize) != 0) {
            continue; }
        if Sle_SetLastCutIsContained(vTemp, Cut) != 0 {
            continue; }
        Sle_SetAddCut( vTemp, Cut );
    }
    }
    // reload
    Vec_IntWriteEntry( vCuts, iObj, Vec_IntSize(vCuts) );
    Vec_IntPush( vCuts, -1 );
    pList0 = Vec_IntArray(vTemp);
    // inlining of macro Sle_ForEachCut
    for  i, pCut0 = 0, pList0 + 1; i <  pList0[0]; i++, pCut0 += Sle_CutSize(pCut0) + 1  { // cuts with unit-cut

    
        if !(Sle_CutIsUsed(pCut0) != 0) {
            continue; }
        Vec_IntPushArray( vCuts, pCut0, Sle_CutSize(pCut0)+1 );
        nCuts++;
    }
    // add unit cut
    Vec_IntPush( vCuts, Sle_CutSetSizeSign(1, 1<<(iObj % 28)) );
    Vec_IntPush( vCuts, iObj );
    Vec_IntWriteEntry( vCuts, Vec_IntEntry(vCuts, iObj), nCuts );
    return nCuts;
}
func Sle_ManComputeCuts(p *Gia_Man_t, nLutSize int32, fVerbose int32) *Vec_Int_t {
    var i, iObj, nCuts int32;
    var vTemp *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vCuts *Vec_Int_t  = Vec_IntAlloc( 30 * Gia_ManAndNum(p) );
    assert(nLutSize <= 6, "nLutSize <= 6");
    Vec_IntFill( vCuts, Gia_ManObjNum(p), 0 );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((iObj) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        Vec_IntWriteEntry( vCuts, iObj, Vec_IntSize(vCuts) );
        Vec_IntPush( vCuts, 0 );
        Vec_IntPush( vCuts, Sle_CutSetSizeSign(1, 1<<(iObj % 28)) );
        Vec_IntPush( vCuts, iObj );
    }
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p, iObj)) {} else {
        nCuts += Sle_ManCutMerge( p, iObj, vCuts, vTemp, nLutSize ); }
    }
    if fVerbose != 0 {
        fmt.Printf("Nodes = %d.  Cuts = %d.  Cuts/Node = %.2f.  Ints/Node = %.2f.  Mem = %.2f MB.\n", Gia_ManAndNum(p), nCuts, 1.0*nCuts/Gia_ManAndNum(p), 1.0*(Vec_IntSize(vCuts)-Gia_ManObjNum(p))/Gia_ManAndNum(p), 1.0*Vec_IntMemory(vCuts) / (1<<20)); }
    Vec_IntFree( vTemp );
    return vCuts;
}

/**Function*************************************************************

  Synopsis    [Cut delay computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManComputeDelayCut(p *Gia_Man_t, pCut *int32, vTime *Vec_Int_t) int32 {
    var nSize int32    = Sle_CutSize(pCut);
    var k int32;
    var pC *int32 = Sle_CutLeaves(pCut);
    var DelayMax int32;
    for  k = 0; k < nSize; k++  {
        DelayMax = Abc_MaxInt( DelayMax, Vec_IntEntry(vTime, pC[k]) );
    }
    return DelayMax + 1;
}
func Sle_ManComputeDelayOne(p *Gia_Man_t, iObj int32, vCuts *Vec_Int_t, vTime *Vec_Int_t) int32 {
    var i int32;
    var pCut *int32;
    var Delay int32;
    var DelayMin int32 = ABC_INFINITY;
    var pList *int32  = Vec_IntEntryP( vCuts, Vec_IntEntry(vCuts, iObj) );
    // inlining of macro Sle_ForEachCut
    for  i, pCut = 0, pList + 1; i <  pList[0]; i++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

    
        Delay = Sle_ManComputeDelayCut( p, pCut, vTime );
        DelayMin = Abc_MinInt( DelayMin, Delay );
    }
    Vec_IntWriteEntry( vTime, iObj, DelayMin );
    return DelayMin;
}
func Sle_ManComputeDelay(p *Gia_Man_t, vCuts *Vec_Int_t) int32 {
    var iObj, Delay, DelayMax int32;
    var vTime *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p, iObj)) {} else {
        Delay = Sle_ManComputeDelayOne( p, iObj, vCuts, vTime );
        DelayMax = Abc_MaxInt( DelayMax, Delay );
    }
    }
    Vec_IntFree( vTime );
    //printf( "Delay = %d.\n", DelayMax );
    return DelayMax;
}

/**Function*************************************************************

  Synopsis    [Cut printing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManPrintCut(pCut *int32) {
    var nSize int32    = Sle_CutSize(pCut);
    var k int32;
    var pC *int32 = Sle_CutLeaves(pCut);
    fmt.Printf("{");
    for  k = 0; k < nSize; k++  {
        fmt.Printf(" %d", pC[k]);
    }
    fmt.Printf(" }\n");
}
func Sle_ManPrintCuts(p *Gia_Man_t, vCuts *Vec_Int_t, iObj int32) {
    var i int32;
    var pCut *int32;
    var pList *int32  = Vec_IntEntryP( vCuts, Vec_IntEntry(vCuts, iObj) );
    fmt.Printf("Obj %3d\n", iObj);
    // inlining of macro Sle_ForEachCut
    for  i, pCut = 0, pList + 1; i <  pList[0]; i++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

        Sle_ManPrintCut( pCut );
    }
    fmt.Printf("\n");
}
func Sle_ManPrintCutsAll(p *Gia_Man_t, vCuts *Vec_Int_t) {
    var iObj int32;
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p, iObj)) {} else {
        Sle_ManPrintCuts( p, vCuts, iObj ); }
    }
}
func Sle_ManComputeCutsTest(p *Gia_Man_t) {
    var vCuts *Vec_Int_t  = Sle_ManComputeCuts( p, 4, 1 );
    //Sle_ManPrintCutsAll( p, vCuts );
    Vec_IntFree( vCuts );
}



/**Function*************************************************************

  Synopsis    [Derive mask representing internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManInternalNodeMask(pGia *Gia_Man_t) *Vec_Bit_t {
    var iObj int32;
    var vMask *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(pGia) );  
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(pGia, iObj)) {} else {
        Vec_BitWriteEntry( vMask, iObj, 1 ); }
    }
    return vMask;
}

/**Function*************************************************************

  Synopsis    [Check if the cut contains only primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManCutHasPisOnly(pCut *int32, vMask *Vec_Bit_t) int32 {
    var k int32;
    var pC *int32 = Sle_CutLeaves(pCut);
    for  k = 0; k < Sle_CutSize(pCut); k++  {
        if Vec_BitEntry(vMask, pC[k]) { // internal node
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Derive cut fanins of each node.]

  Description [These are nodes that are fanins of some cut of this node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManCollectCutFaninsOne(pGia *Gia_Man_t, iObj int32, vCuts *Vec_Int_t, vMask *Vec_Bit_t, vCutFanins *Vec_Int_t, vMap *Vec_Bit_t) {
    var i int32;
    var iFanin int32;
    var pCut *int32;
    var pList *int32 = Vec_IntEntryP( vCuts, Vec_IntEntry(vCuts, iObj) );
    // inlining of macro Sle_ForEachCut
    for  i, pCut = 0, pList + 1; i <  pList[0]; i++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

    
        var nSize int32    = Sle_CutSize(pCut);
        var k int32;
        var pC *int32 = Sle_CutLeaves(pCut);
        assert(nSize > 1, "nSize > 1");
        for  k = 0; k < nSize; k++  {
            if Vec_BitEntry(vMask, pC[k]) && !Vec_BitEntry(vMap, pC[k]) {
                Vec_BitWriteEntry( vMap, pC[k], 1 );
                Vec_IntPush( vCutFanins, pC[k] );
            }
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCutFanins)) && (((iFanin) = Vec_IntEntry(vCutFanins, i)), 1); i++  {
        Vec_BitWriteEntry( vMap, iFanin, 0 );
    }
}
func Sle_ManCollectCutFanins(pGia *Gia_Man_t, vCuts *Vec_Int_t, vMask *Vec_Bit_t) *Vec_Wec_t {
    var iObj int32;
    var vMap *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(pGia) );
    var vCutFanins *Vec_Wec_t  = Vec_WecStart( Gia_ManObjNum(pGia) );
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(pGia, iObj)) {} else {
        Sle_ManCollectCutFaninsOne( pGia, iObj, vCuts, vMask, Vec_WecEntry(vCutFanins, iObj), vMap ); }
    }
    Vec_BitFree( vMap );
    return vCutFanins;
}


   
type Sle_Man_t struct {
    // user's data
    pGia     *Gia_Man_t;         // user's manager (with mapping and edges)
    nLevels            int32;      // total number of levels
    fVerbose            int32;     // verbose flag
    nSatCalls            int32;    // the number of SAT calls
    // SAT variables
    nNodeVars            int32;    // node variables (Gia_ManAndNum(pGia))
    nCutVars            int32;     // cut variables (total number of non-trivial cuts)
    nEdgeVars            int32;    // edge variables (total number of internal edges)
    nDelayVars            int32;   // delay variables (nNodeVars * nLevelsMax)
    nVarsTotal            int32;   // total number of variables
    // SAT clauses
    nCutClas            int32;     // cut clauses
    nEdgeClas            int32;    // edge clauses
    nEdgeClas2            int32;   // edge clauses exclusivity
    nDelayClas            int32;   // delay clauses
    // internal data
    pSat    *sat_solver;         // SAT solver
    vMask     *Vec_Bit_t;        // internal node mask
    vCuts     *Vec_Int_t;        // cuts for each node
    vCutFanins     *Vec_Wec_t;   // internal cut fanins of each node
    vFanoutEdges     *Vec_Wec_t; // internal cut fanins of each node
    vEdgeCuts     *Vec_Wec_t;    // cuts of each edge for one node
    vObjMap     *Vec_Int_t;      // temporary object map
    vCutFirst     *Vec_Int_t;    // first cut of each node
    vEdgeFirst     *Vec_Int_t;   // first edge of each node
    vDelayFirst     *Vec_Int_t;  // first edge of each node
    vPolars     *Vec_Int_t;      // initial 
    vLits     *Vec_Int_t;        // literals 
    // statistics
    timeStart        abctime;
}

func Sle_ManList(p *Sle_Man_t, i int32) *int32 { return Vec_IntEntryP(p.vCuts, Vec_IntEntry(p.vCuts, i)); }

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManAlloc(pGia *Gia_Man_t, nLevels int32, fVerbose int32) *Sle_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Sle_Man_t    = &Sle_Man_t{};
    p.pGia         = pGia;
    p.nLevels      = nLevels;
    p.fVerbose     = fVerbose;
    p.vMask        = Sle_ManInternalNodeMask( pGia );
    p.vCuts        = Sle_ManComputeCuts( pGia, 4, fVerbose );
    p.vCutFanins   = Sle_ManCollectCutFanins( pGia, p.vCuts, p.vMask );
    p.vFanoutEdges = Vec_WecStart( Gia_ManObjNum(pGia) );
    p.vEdgeCuts    = Vec_WecAlloc( 100 );
    p.vObjMap      = Vec_IntStartFull( Gia_ManObjNum(pGia) );
    p.vCutFirst    = Vec_IntStartFull( Gia_ManObjNum(pGia) );
    p.vEdgeFirst   = Vec_IntStartFull( Gia_ManObjNum(pGia) );
    p.vDelayFirst  = Vec_IntStartFull( Gia_ManObjNum(pGia) );
    p.vPolars      = Vec_IntAlloc( 100 );
    p.vLits        = Vec_IntAlloc( 100 );
    p.nLevels      = Sle_ManComputeDelay( pGia, p.vCuts );
    return p;
}
func Sle_ManStop(p *Sle_Man_t) {
    sat_solver_delete( p.pSat );
    Vec_BitFree( p.vMask );
    Vec_IntFree( p.vCuts );
    Vec_WecFree( p.vCutFanins );
    Vec_WecFree( p.vFanoutEdges );
    Vec_WecFree( p.vEdgeCuts );
    Vec_IntFree( p.vObjMap );
    Vec_IntFree( p.vCutFirst );
    Vec_IntFree( p.vEdgeFirst );
    Vec_IntFree( p.vDelayFirst );
    Vec_IntFree( p.vPolars );
    Vec_IntFree( p.vLits );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManMarkupVariables(p *Sle_Man_t) {
    var iObj int32;
    var Counter int32 = Gia_ManObjNum(p.pGia);
    // node variables
    p.nNodeVars = Counter;
    // cut variables
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        Vec_IntWriteEntry( p.vCutFirst, iObj, Counter );
        Counter += Sle_ListCutNum( Sle_ManList(p, iObj) );
    }
    }
    p.nCutVars = Counter - p.nNodeVars;
    // edge variables
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        Vec_IntWriteEntry( p.vEdgeFirst, iObj, Counter );
        Counter += Vec_IntSize( Vec_WecEntry(p.vCutFanins, iObj) );
    }
    }
    p.nEdgeVars = Counter - p.nCutVars - p.nNodeVars;
    // delay variables
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        Vec_IntWriteEntry( p.vDelayFirst, iObj, Counter );
        Counter += p.nLevels;
    }
    }
    p.nDelayVars = Counter - p.nEdgeVars - p.nCutVars - p.nNodeVars;
    p.nVarsTotal = Counter;
}


/**Function*************************************************************

  Synopsis    [Derive initial variable assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// returns 1 if Cut can represent LUT (Cut is equal or is contained in LUT)
func Sle_ManCheckContained(pCutLeaves *int32, nCutLeaves int32, pLutFanins *int32, nLutFanins int32) int32 {
    var i, k int32;
    if nCutLeaves > nLutFanins {
        return 0; }
    for  i = 0; i < nCutLeaves; i++  {
        for  k = 0; k < nLutFanins; k++  {
            if pCutLeaves[i] == pLutFanins[k] {
                break; }
        }
        if k == nLutFanins { // not found
            return 0; }
    }
    return 1;
}
func Sle_ManDeriveInit(p *Sle_Man_t) {
    var vEdges *Vec_Int_t;
    var i, iObj, iFanin, iEdge int32;
    if !Gia_ManHasMapping(p.pGia) {
        return; }
    // derive initial state
    Vec_IntClear( p.vPolars );
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        var nFanins int32;
        var pFanins *int32;
        var pCut *int32;
        var pList *int32;
        var iFound int32 = -1;
        if !Gia_ObjIsLut(p.pGia, iObj) {
            continue; }
        Vec_IntPush( p.vPolars, iObj ); // node var
        nFanins = Gia_ObjLutSize( p.pGia, iObj );
        pFanins = Gia_ObjLutFanins( p.pGia, iObj );
        // find cut
        pList = Sle_ManList( p, iObj );
        // inlining of macro Sle_ForEachCut
        for  i, pCut = 0, pList + 1; i <  pList[0]; i++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

            if Sle_ManCheckContained( Sle_CutLeaves(pCut), Sle_CutSize(pCut), pFanins, nFanins ) != 0 {
                iFound = i;
                break;
            }
        }
        if iFound == -1 {
            fmt.Printf("Cannot find the following cut at node %d: {", iObj);
            for  i = 0; i < nFanins; i++  {
                fmt.Printf(" %d", pFanins[i]);
            }
            fmt.Printf(" }\n");
            Sle_ManPrintCuts( p.pGia, p.vCuts, iObj );
            fflush( os.Stdout );
        }
        assert(iFound >= 0, "iFound >= 0");
        Vec_IntPush( p.vPolars, Vec_IntEntry(p.vCutFirst, iObj) + iFound ); // cut var
        // check if the cut contains only primary inputs - if so, its delay is equal to 1
        if Sle_ManCutHasPisOnly(pCut, p.vMask) != 0 {
            Vec_IntPush( p.vPolars, Vec_IntEntry(p.vDelayFirst, iObj) ); } // delay var
    }
    }
    Vec_IntSort( p.vPolars, 0 );
    // find zero-delay edges
    if !p.pGia.vEdge1 {
        return; }
    vEdges = Gia_ManEdgeToArray( p.pGia );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vEdges)) && (((iFanin) = Vec_IntEntry(vEdges, i)), 1) && (((iObj) = Vec_IntEntry(vEdges, i+1)), 1); i += 2  {
        assert(iFanin < iObj, "iFanin < iObj");
        assert(Gia_ObjIsLut(p.pGia, iFanin), "Gia_ObjIsLut(p.pGia, iFanin)");
        assert(Gia_ObjIsLut(p.pGia, iObj), "Gia_ObjIsLut(p.pGia, iObj)");
        assert(Gia_ObjIsAnd(Gia_ManObj(p.pGia, iFanin)), "Gia_ObjIsAnd(Gia_ManObj(p.pGia, iFanin))");
        assert(Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)), "Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj))");
        // find edge
        iEdge = Vec_IntFind( Vec_WecEntry(p.vCutFanins, iObj), iFanin );
        if iEdge < 0 {
            continue; }
        assert(iEdge >= 0, "iEdge >= 0");
        Vec_IntPush( p.vPolars, Vec_IntEntry(p.vEdgeFirst, iObj) + iEdge ); // edge
    }
    Vec_IntFree( vEdges );
}

/**Function*************************************************************

  Synopsis    [Derive CNF.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManDeriveCnf(p *Sle_Man_t, nBTLimit int32, fDynamic int32) {
    var nTimeOut int32;
    var i, iObj, value int32;
    var vArray *Vec_Int_t;

    // start the SAT solver
    p.pSat = sat_solver_new();
    sat_solver_setnvars( p.pSat, p.nVarsTotal );
    sat_solver_set_resource_limits( p.pSat, nBTLimit, 0, 0, 0 );
    sat_solver_set_runtime_limit( p.pSat, nTimeOut ? nTimeOut * int64(time.Second) + Abc_Clock(): 0 );
    sat_solver_set_random( p.pSat, 1 );
    sat_solver_set_polarity( p.pSat, Vec_IntArray(p.vPolars), Vec_IntSize(p.vPolars) );
    //sat_solver_set_var_activity( p->pSat, NULL, p->nVarsTotal );

    // set drivers to be mapped
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && (((iObj) = Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))), 1); i++  {
        if Vec_BitEntry(p.vMask, iObj) { // internal node
        
            Vec_IntFill( p.vLits, 1, Abc_Var2Lit(iObj, 0) ); // pos lit
            value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
            assert((value != 0), "value");
        }
    }

    // add cover clauses and edge-to-cut clauses
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        var e, iEdge, nEdges, Entry int32;
        var iCutVar0 int32   = Vec_IntEntry( p.vCutFirst, iObj );
        var iEdgeVar0 int32  = Vec_IntEntry( p.vEdgeFirst, iObj );
        var pCut *int32;
        var pList *int32 = Sle_ManList( p, iObj );
        var vCutFans *Vec_Int_t  = Vec_WecEntry( p.vCutFanins, iObj );
        assert(iCutVar0 > 0 && iEdgeVar0 > 0, "iCutVar0 > 0 && iEdgeVar0 > 0");
        // node requires one of the cuts
        Vec_IntFill( p.vLits, 1, Abc_Var2Lit(iObj, 1) ); // neg lit
        for  i = 0; i < Sle_ListCutNum(pList); i++  {
            Vec_IntPush( p.vLits, Abc_Var2Lit(iCutVar0 + i, 0) );
        }
        value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
        assert((value != 0), "value");
        // cuts are mutually exclusive
        for  i = 0; i < Sle_ListCutNum(pList); i++  {
            for  e = i+1; e < Sle_ListCutNum(pList); e++  {
                Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(iCutVar0 + i, 1), Abc_Var2Lit(iCutVar0 + e, 1) );
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");
            }
        }
        // cut requires fanin nodes
        Vec_WecInit( p.vEdgeCuts, Vec_IntSize(vCutFans) );
        // inlining of macro Sle_ForEachCut
        for  i, pCut = 0, pList + 1; i <  pList[0]; i++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

        
            var nSize int32    = Sle_CutSize(pCut);
            var k int32;
            var pC *int32 = Sle_CutLeaves(pCut);
            assert(nSize > 1, "nSize > 1");
            for  k = 0; k < nSize; k++  {
                if !Vec_BitEntry(p.vMask, pC[k]) {
                    continue; }
                Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(iCutVar0 + i, 1), Abc_Var2Lit(pC[k], 0) );
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");
                // find the edge ID between pC[k] and iObj
                iEdge = Vec_IntEntry(p.vObjMap, pC[k]);
                if iEdge == -1 {
                    Vec_IntWriteEntry( p.vObjMap, pC[k], (iEdge = nEdges++) );
                    Vec_WecPush( p.vFanoutEdges, pC[k], iEdgeVar0 + iEdge );
                }
                Vec_WecPush( p.vEdgeCuts, iEdge, iCutVar0 + i );
                p.nCutClas++;
            }
            // cut requires the node
            Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(iCutVar0 + i, 1), Abc_Var2Lit(iObj, 0) );
            value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
            assert((value != 0), "value");
        }
        assert(nEdges == Vec_IntSize(vCutFans), "nEdges == Vec_IntSize(vCutFans)");

        // edge requires one of the fanout cuts
        // inlining of macro Vec_WecForEachLevel
        for  e = 0; (e < Vec_WecSize(p.vEdgeCuts)) && (((vArray) = Vec_WecEntry(p.vEdgeCuts, e)), 1); e++  {
            assert(Vec_IntSize(vArray) > 0, "Vec_IntSize(vArray) > 0");
            Vec_IntFill( p.vLits, 1, Abc_Var2Lit(iEdgeVar0 + e, 1) ); // neg lit (edge)
            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vArray)) && (((Entry) = Vec_IntEntry(vArray, i)), 1); i++  {
                Vec_IntPush( p.vLits, Abc_Var2Lit(Entry, 0) );
            } // pos lit (cut)
            value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
            assert((value != 0), "value");
            p.nEdgeClas++;
        }

        // clean object map
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vCutFans)) && (((Entry) = Vec_IntEntry(vCutFans, i)), 1); i++  {
            Vec_IntWriteEntry( p.vObjMap, Entry, -1 );
        }
    }
    }

    // mutual exclusivity of edges
    // inlining of macro Vec_WecForEachLevel
    for  iObj = 0; (iObj < Vec_WecSize(p.vFanoutEdges)) && (((vArray) = Vec_WecEntry(p.vFanoutEdges, iObj)), 1); iObj++  {
        var j, k, EdgeJ, EdgeK int32;
        var iEdgeVar0 int32  = Vec_IntEntry( p.vEdgeFirst, iObj );
        var vCutFans *Vec_Int_t  = Vec_WecEntry( p.vCutFanins, iObj );
        // add fanin edges
        for  i = 0; i < Vec_IntSize(vCutFans); i++  {
            Vec_IntPush( vArray, iEdgeVar0 + i );
        }
        // generate pairs
        if fDynamic != 0 {
            continue; }
        // inlining of macro Vec_IntForEachEntry
        for  j = 0; (j < Vec_IntSize(vArray)) && (((EdgeJ) = Vec_IntEntry(vArray, j)), 1); j++  {
            // inlining of macro Vec_IntForEachEntryStart
            for  k = j+1; (k < Vec_IntSize(vArray)) && (((EdgeK) = Vec_IntEntry(vArray, k)), 1); k++  {
                Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(EdgeJ, 1), Abc_Var2Lit(EdgeK, 1) );
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");
            }
        }
        p.nEdgeClas2 += Vec_IntSize(vArray) * (Vec_IntSize(vArray) - 1) / 2;
    }

    // add delay clauses
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        var e, iFanin int32;
        var iEdgeVar0 int32  = Vec_IntEntry( p.vEdgeFirst, iObj );
        var iDelayVar0 int32  = Vec_IntEntry( p.vDelayFirst, iObj );
        var vCutFans *Vec_Int_t  = Vec_WecEntry( p.vCutFanins, iObj );
        // check if the node has cuts containing only primary inputs
        var pCut *int32;
        var pList *int32 = Sle_ManList( p, iObj );
        // inlining of macro Sle_ForEachCut
        for  i, pCut = 0, pList + 1; i <  pList[0]; i++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

            if Sle_ManCutHasPisOnly(pCut, p.vMask) != 0 {
                Vec_IntFill( p.vLits, 1, Abc_Var2Lit(iDelayVar0, 0) ); // pos lit
//                Vec_IntFillTwo( p->vLits, 2, Abc_Var2Lit(iObj, 1), Abc_Var2Lit(iDelayVar0, 0) );
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");
                break;
            }
        }
//        if ( i < Sle_ListCutNum(pList) )
//            continue;
        // create delay requirements for each cut fanin of this node
        // inlining of macro Vec_IntForEachEntry
        for  e = 0; (e < Vec_IntSize(vCutFans)) && (((iFanin) = Vec_IntEntry(vCutFans, e)), 1); e++  {
            var d int32;
            var iDelayVarIn int32 = Vec_IntEntry( p.vDelayFirst, iFanin );
            for  d = 0; d < p.nLevels; d++  {
                Vec_IntClear( p.vLits );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iObj, 1) );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iFanin, 1) );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iDelayVarIn + d, 1) );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iEdgeVar0 + e, 0) );
                if d < p.nLevels-1 {
                    Vec_IntPush( p.vLits, Abc_Var2Lit(iDelayVar0 + d+1, 0) ); }
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");

                Vec_IntClear( p.vLits );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iObj, 1) );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iFanin, 1) );
                Vec_IntPush( p.vLits, Abc_Var2Lit(iDelayVarIn + d, 1) );
                if d < p.nLevels-1 {
                    Vec_IntPush( p.vLits, Abc_Var2Lit(iEdgeVar0 + e, 1) ); }
                Vec_IntPush( p.vLits, Abc_Var2Lit(iDelayVar0 + d, 0) );
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                assert((value != 0), "value");
            }
            p.nDelayClas += 2*p.nLevels;
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Add edge compatibility constraints.]

  Description [Returns 1 if constraints have been added.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManAddEdgeConstraints(p *Sle_Man_t, nEdges int32) int32 {
    var vArray *Vec_Int_t;
    var vTemp *Vec_Int_t  = Vec_IntAlloc( 100 );
    var value, iObj, nAdded int32;
    assert(nEdges == 1 || nEdges == 2, "nEdges == 1 || nEdges == 2");
    // inlining of macro Vec_WecForEachLevel
    for  iObj = 0; (iObj < Vec_WecSize(p.vFanoutEdges)) && (((vArray) = Vec_WecEntry(p.vFanoutEdges, iObj)), 1); iObj++  {
        var j, k, EdgeJ, EdgeK int32;
        // check if they are incompatible
        Vec_IntClear( vTemp );
        // inlining of macro Vec_IntForEachEntry
        for  j = 0; (j < Vec_IntSize(vArray)) && (((EdgeJ) = Vec_IntEntry(vArray, j)), 1); j++  {
            if sat_solver_var_value(p.pSat, EdgeJ) {
                Vec_IntPush( vTemp, EdgeJ ); }
        }
        if Vec_IntSize(vTemp) <= nEdges {
            continue; }
        nAdded++;
        if nEdges == 1 {
            // generate pairs
            // inlining of macro Vec_IntForEachEntry
            for  j = 0; (j < Vec_IntSize(vTemp)) && (((EdgeJ) = Vec_IntEntry(vTemp, j)), 1); j++  {
                // inlining of macro Vec_IntForEachEntryStart
                for  k = j+1; (k < Vec_IntSize(vTemp)) && (((EdgeK) = Vec_IntEntry(vTemp, k)), 1); k++  {
                    Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(EdgeJ, 1), Abc_Var2Lit(EdgeK, 1) );
                    value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                    assert((value != 0), "value");
                }
            }
            p.nEdgeClas2 += Vec_IntSize(vTemp) * (Vec_IntSize(vTemp) - 1) / 2;
        
        } else if nEdges == 2 {
            var m, EdgeM int32;
            // generate triples
            // inlining of macro Vec_IntForEachEntry
            for  j = 0; (j < Vec_IntSize(vTemp)) && (((EdgeJ) = Vec_IntEntry(vTemp, j)), 1); j++  {
                // inlining of macro Vec_IntForEachEntryStart
                for  k = j+1; (k < Vec_IntSize(vTemp)) && (((EdgeK) = Vec_IntEntry(vTemp, k)), 1); k++  {
                    // inlining of macro Vec_IntForEachEntryStart
                    for  m = k+1; (m < Vec_IntSize(vTemp)) && (((EdgeM) = Vec_IntEntry(vTemp, m)), 1); m++  {
                        Vec_IntFillTwo( p.vLits, 2, Abc_Var2Lit(EdgeJ, 1), Abc_Var2Lit(EdgeK, 1) );
                        Vec_IntPush( p.vLits, Abc_Var2Lit(EdgeM, 1) );
                        value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vLits), Vec_IntLimit(p.vLits)  );
                        assert((value != 0), "value");
                    }
                }
            }
            p.nEdgeClas2 += Vec_IntSize(vTemp) * (Vec_IntSize(vTemp) - 1) * (Vec_IntSize(vTemp) - 2) / 6;
        
        } else { assert(false, "0"); }
    }
    Vec_IntFree( vTemp );
    //printf( "Added clauses to %d nodes.\n", nAdded );
    return nAdded;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManDeriveResult(p *Sle_Man_t, vEdge2 *Vec_Int_t, vMapping *Vec_Int_t) {
    var vMapTemp *Vec_Int_t;
    var iObj int32;
    // create mapping
    Vec_IntFill( vMapping, Gia_ManObjNum(p.pGia), 0 );
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        var i int32;
        var iCut int32;
        var iCutVar0 int32 = Vec_IntEntry( p.vCutFirst, iObj );
        var pCut *int32;
        var pCutThis *int32 = nil;
        var pList *int32 = Sle_ManList( p, iObj );
        if !sat_solver_var_value(p.pSat, iObj) {
            continue; }
        // inlining of macro Sle_ForEachCut
        for  iCut, pCut = 0, pList + 1; iCut <  pList[0]; iCut++, pCut += Sle_CutSize(pCut) + 1  { // cuts with unit-cut

            if sat_solver_var_value(p.pSat, iCutVar0 + iCut) {
                assert(pCutThis == nil, "pCutThis == nil");
                pCutThis = pCut;
            }
        }
        assert(pCutThis != nil, "pCutThis != nil");
        Vec_IntWriteEntry( vMapping, iObj, Vec_IntSize(vMapping) );
        Vec_IntPush( vMapping, Sle_CutSize(pCutThis) );
        for  i = 0; i < Sle_CutSize(pCutThis); i++  {
            Vec_IntPush( vMapping, Sle_CutLeaves(pCutThis)[i] );
        }
        Vec_IntPush( vMapping, iObj );
    }
    }
    vMapTemp = p.pGia.vMapping;
    p.pGia.vMapping = vMapping;
    // collect edges
    Vec_IntClear( vEdge2 );
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {} else {
        var i int32;
        var iFanin int32;
        var iEdgeVar0 int32 = Vec_IntEntry( p.vEdgeFirst, iObj );
        var vCutFans *Vec_Int_t  = Vec_WecEntry( p.vCutFanins, iObj );
        //int * pCut, * pList  = Sle_ManList( p, iObj );
        // int iCutVar0  = Vec_IntEntry( p->vCutFirst, iObj );
        if !sat_solver_var_value(p.pSat, iObj) {
            continue; }
        //for ( i = 0; i < Sle_ListCutNum(pList); i++ )
        //    printf( "%d", sat_solver_var_value(p->pSat, iCutVar0 + i) );
        //printf( "\n" );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vCutFans)) && (((iFanin) = Vec_IntEntry(vCutFans, i)), 1); i++  {
            if sat_solver_var_value(p.pSat, iFanin) && sat_solver_var_value(p.pSat, iEdgeVar0 + i) {
                // verify edge
                var pFanins *int32  = Gia_ObjLutFanins( p.pGia, iObj );
                var k int32;
                var nFanins int32 = Gia_ObjLutSize( p.pGia, iObj );
                for  k = 0; k < nFanins; k++  {
                    //printf( "%d ", pFanins[k] );
                    if pFanins[k] == iFanin {
                        break; }
                }
                //printf( "\n" );
                if k == nFanins {
//                    printf( "Cannot find LUT with input %d at node %d.\n", iFanin, iObj );
                    continue; }
                Vec_IntPushTwo( vEdge2, iFanin, iObj );
            }
        }
    }
    }
    p.pGia.vMapping = vMapTemp;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManExplore(pGia *Gia_Man_t, nBTLimit int32, DelayInit int32, fDynamic int32, fTwoEdges int32, fVerbose int32) {
    var fVeryVerbose int32;
    var clk abctime  = Abc_Clock();
    var vEdges2 *Vec_Int_t   = Vec_IntAlloc(1000);
    var vMapping *Vec_Int_t  = Vec_IntAlloc(1000);
    var i, iLut, nConfs, status, Delay, iFirstVar int32;
    var DelayStart int32;
    if (DelayInit != 0) || !Gia_ManHasMapping(pGia) {
        DelayStart = DelayInit;
    } else {
        DelayStart = Gia_ManLutLevel(pGia, nil);
    }
    var p *Sle_Man_t  = Sle_ManAlloc( pGia, DelayStart, fVerbose );
    if fVerbose != 0 {
        fmt.Printf("Running solver with %d conflicts, %d initial delay, and %d edges. Dynamic constraints = %s.\n", nBTLimit, DelayInit, 1+fTwoEdges, fDynamic?"yes":"no"); }
    Sle_ManMarkupVariables( p );
    if fVerbose != 0 {
        fmt.Printf("Vars:  Total = %d.  Node = %d. Cut = %d. Edge = %d. Delay = %d.\n", p.nVarsTotal, p.nNodeVars, p.nCutVars, p.nEdgeVars, p.nDelayVars); }
    Sle_ManDeriveInit( p );
    Sle_ManDeriveCnf( p, nBTLimit, fDynamic || fTwoEdges );
    if fVerbose != 0 {
        fmt.Printf("Clas:  Total = %d.  Cut = %d. Edge = %d. EdgeEx = %d. Delay = %d.\n", sat_solver_nclauses(p.pSat), p.nCutClas, p.nEdgeClas, p.nEdgeClas2, p.nDelayClas); }
    //Sat_SolverWriteDimacs( p->pSat, "temp.cnf", NULL, NULL, 0 );
    for  Delay = p.nLevels; Delay >= 0; Delay--  {
        // we constrain COs, although it would be fine to constrain only POs
        if Delay < p.nLevels {
            // inlining of macro Gia_ManForEachCoDriverId
            for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && (((iLut) = Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))), 1); i++  {
                if Vec_BitEntry(p.vMask, iLut) { // internal node
                
                    iFirstVar = Vec_IntEntry( p.vDelayFirst, iLut );
                    if !sat_solver_push(p.pSat, Abc_Var2Lit(iFirstVar + Delay, 1)) {
                        break; }
                }
            }
            if i < Gia_ManCoNum(p.pGia) {
                if fVerbose != 0 {
                    fmt.Printf("Proved UNSAT for delay %d.  ", Delay);
                    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
                }
                break;
            }
        }
        // solve with assumptions
        //clk = Abc_Clock();
        nConfs = sat_solver_nconflicts( p.pSat );
        for {
            p.nSatCalls++;
            status = sat_solver_solve_internal( p.pSat );
            if status != l_True {
                break; }
            if !(Sle_ManAddEdgeConstraints(p, 1+fTwoEdges) != 0) {
                break; }
        }
        nConfs = sat_solver_nconflicts( p.pSat ) - nConfs;
        if status == l_True {
            if fVerbose != 0 {
                var nNodes, nEdges int32;
                for  i = 0; i < p.nNodeVars; i++  {
                    nNodes += sat_solver_var_value(p.pSat, i);
                }
                for  i = 0; i < p.nEdgeVars; i++  {
                    nEdges += sat_solver_var_value(p.pSat, p.nNodeVars + p.nCutVars + i);
                }
                fmt.Printf("Solution with delay %2d, node count %5d, and edge count %5d exists. Conf = %8d.  ", Delay, nNodes, nEdges, nConfs);
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            // save the result
            Sle_ManDeriveResult( p, vEdges2, vMapping );
            if fVeryVerbose != 0 {
                fmt.Printf("Nodes:  ");
                for  i = 0; i < p.nNodeVars; i++  {
                    if sat_solver_var_value(p.pSat, i) {
                        fmt.Printf("%d ", i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");

                Vec_IntPrint( p.vCutFirst );
                fmt.Printf("Cuts:   ");
                for  i = 0; i < p.nCutVars; i++  {
                    if sat_solver_var_value(p.pSat, p.nNodeVars + i) {
                        fmt.Printf("%d ", p.nNodeVars + i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");

                Vec_IntPrint( p.vEdgeFirst );
                fmt.Printf("Edges:  ");
                for  i = 0; i < p.nEdgeVars; i++  {
                    if sat_solver_var_value(p.pSat, p.nNodeVars + p.nCutVars + i) {
                        fmt.Printf("%d ", p.nNodeVars + p.nCutVars + i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");

                Vec_IntPrint( p.vDelayFirst );
                fmt.Printf("Delays: ");
                for  i = 0; i < p.nDelayVars; i++  {
                    if sat_solver_var_value(p.pSat, p.nNodeVars + p.nCutVars + p.nEdgeVars + i) {
                        fmt.Printf("%d ", p.nNodeVars + p.nCutVars + p.nEdgeVars + i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");
            }
        
        } else {
            if fVerbose != 0 {
                if status == l_False {
                    fmt.Printf("Proved UNSAT for delay %d. Conf = %8d.  ", Delay, nConfs);
                } else {
                    fmt.Printf("Resource limit reached for delay %d. Conf = %8d.  ", Delay, nConfs); }
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            break;
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Clas:  Total = %d.  Cut = %d. Edge = %d. EdgeEx = %d. Delay = %d.  Calls = %d.\n", sat_solver_nclauses(p.pSat), p.nCutClas, p.nEdgeClas, p.nEdgeClas2, p.nDelayClas, p.nSatCalls); }
    if Vec_IntSize(vMapping) > 0 {
        Gia_ManEdgeFromArray( p.pGia, vEdges2 );
        Vec_IntFree( vEdges2 );
        Vec_IntFreeP( &p.pGia.vMapping );
        p.pGia.vMapping = vMapping;
    
    } else {
        Vec_IntFree( vEdges2 );
        Vec_IntFree( vMapping );
    }
    Vec_IntFreeP( &p.pGia.vPacking );
    Sle_ManStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

