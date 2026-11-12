package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigDoms.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Computing multi-output dominators.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigDoms.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"
//#include "aig/saig/saig.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Aig_Sto_t = Aig_Sto_t_;
   

type Aig_Dom_t struct {
    uSign             int32;         // signature
    nNodes             int32;        // the number of nodes
    pNodes [0]int32;     // the nodes
}

type Aig_Sto_t_ struct {
    Limit             int32;
    pAig      *Aig_Man_t;          // user's AIG
    pMem  *Aig_MmFixed_t;          // memory manager for dominators
    vDoms      *Vec_Ptr_t;         // dominators
    vFans      *Vec_Int_t;         // temporary fanouts
    vTimes      *Vec_Int_t;        // the number of times each appears
    nDomNodes             int32;     // nodes with dominators
    nDomsTotal             int32;    // total dominators
    nDomsFilter1             int32;  // filtered dominators
    nDomsFilter2             int32;  // filtered dominators
}

// Original macro:
// #define Aig_DomForEachNode( pAig, pDom, pNode, i )                         \
//     for ( i = 0; (i < pDom->nNodes) && ((pNode) = Aig_ManObj(pAig, (pDom)->pNodes[i])); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates dominator manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDomStart(pAig *Aig_Man_t, Limit int32) *Aig_Sto_t {
    
    // inlining of macro ABC_CALLOC
    pSto := &Aig_Sto_t{};
    pSto.pAig     = pAig;
    pSto.Limit    = Limit;
    pSto.pMem     = Aig_MmFixedStart( sizeof(Aig_Dom_t) + sizeof(int) * Limit, 10000 );
    pSto.vDoms    = Vec_PtrStart( Aig_ManObjNumMax(pAig) );
    pSto.vFans    = Vec_IntAlloc( 100 );
    pSto.vTimes   = Vec_IntStart( Aig_ManObjNumMax(pAig) );
    return pSto;
}

/**Function*************************************************************

  Synopsis    [Adds trivial dominator.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjAddTriv(pSto *Aig_Sto_t, Id int32, vDoms *Vec_Ptr_t) {
    var pDom *Aig_Dom_t;
    pDom = (Aig_Dom_t *)Aig_MmFixedEntryFetch( pSto.pMem );
    pDom.uSign     = (1 << (Id % 63));
    pDom.nNodes    = 1;
    pDom.pNodes[0] = Id;
    Vec_PtrPushFirst( vDoms, pDom );
    assert(Vec_PtrEntry( pSto.vDoms, Id ) == nil, "Vec_PtrEntry( pSto.vDoms, Id ) == nil");
    Vec_PtrWriteEntry( pSto.vDoms, Id, vDoms );
}

/**Function*************************************************************

  Synopsis    [Duplicates vector of doms.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomVecDup(pSto *Aig_Sto_t, vDoms *Vec_Ptr_t, fSkip1 int32) *Vec_Ptr_t {
    var vDoms2 *Vec_Ptr_t;
    var pDom, pDom2 *Aig_Dom_t;
    var i int32;
    vDoms2 = Vec_PtrAlloc( 0 );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = fSkip1; (i < Vec_PtrSize(vDoms)) && (((pDom) = (Aig_Dom_t *)Vec_PtrEntry(vDoms, i)), 1); i++  {
        pDom2 = (Aig_Dom_t *)Aig_MmFixedEntryFetch( pSto.pMem );
        memcpy( pDom2, pDom, sizeof(Aig_Dom_t) + sizeof(int) * pSto.Limit );
        Vec_PtrPush( vDoms2, pDom2 );
    }
    return vDoms2;
}

/**Function*************************************************************

  Synopsis    [Recycles vector of doms.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomVecRecycle(pSto *Aig_Sto_t, vDoms *Vec_Ptr_t) {
    var pDom *Aig_Dom_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vDoms)) && (((pDom) = (Aig_Dom_t *)Vec_PtrEntry(vDoms, i)), 1); i++  {
        Aig_MmFixedEntryRecycle( pSto.pMem, (char *)pDom );
    }
    Vec_PtrFree( vDoms );
}

/**Function*************************************************************

  Synopsis    [Duplicates the vector of doms.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomPrint(pSto *Aig_Sto_t, pDom *Aig_Dom_t, Num int32) {
    var k int32;
    fmt.Printf("%4d : {", Num);
    for  k = 0; k < pDom.nNodes; k++  {
        fmt.Printf(" %4d", pDom.pNodes[k]);
    }
    for  ; k < pSto.Limit; k++  {
        fmt.Printf("     ");
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    [Duplicates the vector of doms.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomVecPrint(pSto *Aig_Sto_t, vDoms *Vec_Ptr_t) {
    var pDom *Aig_Dom_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vDoms)) && (((pDom) = (Aig_Dom_t *)Vec_PtrEntry(vDoms, i)), 1); i++  {
        Aig_ObjDomPrint( pSto, pDom, i );
    }
}

/**Function*************************************************************

  Synopsis    [Computes multi-node dominators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDomPrint(pSto *Aig_Sto_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(pSto.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pSto.pAig.vCis, i)), 1); i++  {
        fmt.Printf("*** PI %4d %4d :\n", i, pObj.Id);
        Aig_ObjDomVecPrint( pSto, (Vec_Ptr_t *)Vec_PtrEntry(pSto.vDoms, pObj.Id) );
    }
}

/**Function*************************************************************

  Synopsis    [Divides the circuit into well-balanced parts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDomStop(pSto *Aig_Sto_t) {
    var vDoms *Vec_Ptr_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pSto.vDoms)) && (((vDoms) = (Vec_Ptr_t *)Vec_PtrEntry(pSto.vDoms, i)), 1); i++  {
        if vDoms != nil {
            Aig_ObjDomVecRecycle( pSto, vDoms ); }
    }
    Vec_PtrFree( pSto.vDoms );
    Vec_IntFree( pSto.vFans );
    Vec_IntFree( pSto.vTimes );
    Aig_MmFixedStop( pSto.pMem, 0 );
    // inlining of macro ABC_FREE
    ((pSto) ? (free((char *) (pSto)), (pSto) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Checks correctness of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomCheck(pDom *Aig_Dom_t) int32 {
    var i int32;
    for  i = 1; i < pDom.nNodes; i++  {
        if pDom.pNodes[i-1] >= pDom.pNodes[i] {
            Abc_Print( -1, "Aig_ObjDomCheck(): Cut has wrong ordering of inputs.\n" );
            return 0;
        }
        assert(pDom.pNodes[i-1] < pDom.pNodes[i], "pDom.pNodes[i-1] < pDom.pNodes[i]");
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pDom is contained in pCut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomCheckDominance(pDom *Aig_Dom_t, pCut *Aig_Dom_t) int32 {
    var i, k int32;
    for  i = 0; i < pDom.nNodes; i++  {
        for  k = 0; k < int32(pCut.nNodes); k++  {
            if pDom.pNodes[i] == pCut.pNodes[k] {
                break; }
        }
        if k == int32(pCut.nNodes) { // node i in pDom is not contained in pCut
            return 0; }
    }
    // every node in pDom is contained in pCut
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the cut is contained.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomFilter(pSto *Aig_Sto_t, vDoms *Vec_Ptr_t, pDom *Aig_Dom_t) int32 { 
    var pTemp *Aig_Dom_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vDoms)) && (((pTemp) = (Aig_Dom_t *)Vec_PtrEntry(vDoms, i)), 1); i++  {
        if pTemp.nNodes > pDom.nNodes {
            // do not fiter the first cut
            if i == 0 {
                continue; }
            // skip the non-contained cuts
            if (pTemp.uSign & pDom.uSign) != pDom.uSign {
                continue; }
            // check containment seriously
            if Aig_ObjDomCheckDominance( pDom, pTemp ) != 0 {
                Vec_PtrRemove( vDoms, pTemp );
                Aig_MmFixedEntryRecycle( pSto.pMem, (char *)pTemp );
                i--;
                pSto.nDomsFilter1++;
            }
         
        } else {
            // skip the non-contained cuts
            if (pTemp.uSign & pDom.uSign) != pTemp.uSign {
                continue; }
            // check containment seriously
            if Aig_ObjDomCheckDominance( pTemp, pDom ) != 0 {
                pSto.nDomsFilter2++;
                return 1;
            }
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Merges two cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomMergeOrdered(pD0 *Aig_Dom_t, pD1 *Aig_Dom_t, pD *Aig_Dom_t, Limit int32) int32 { 
    var i, k, c int32;
    assert(pD0.nNodes >= pD1.nNodes, "pD0.nNodes >= pD1.nNodes");
    // the case of the largest cut sizes
    if pD0.nNodes == Limit && pD1.nNodes == Limit {
        for  i = 0; i < pD0.nNodes; i++  {
            if pD0.pNodes[i] != pD1.pNodes[i] {
                return 0; }
        }
        for  i = 0; i < pD0.nNodes; i++  {
            pD.pNodes[i] = pD0.pNodes[i];
        }
        pD.nNodes = pD0.nNodes;
        return 1;
    }
    // the case when one of the cuts is the largest
    if pD0.nNodes == Limit {
        for  i = 0; i < pD1.nNodes; i++  {
            for  k = pD0.nNodes - 1; k >= 0; k--  {
                if pD0.pNodes[k] == pD1.pNodes[i] {
                    break; }
            }
            if k == -1 { // did not find
                return 0; }
        }
        for  i = 0; i < pD0.nNodes; i++  {
            pD.pNodes[i] = pD0.pNodes[i];
        }
        pD.nNodes = pD0.nNodes;
        return 1;
    }

    // compare two cuts with different numbers
    i = k = 0;
    for  c = 0; c < int32(Limit); c++  {
        if k == pD1.nNodes {
            if i == pD0.nNodes {
                pD.nNodes = c;
                return 1;
            }
            pD.pNodes[c] = pD0.pNodes[i]; i++;
            continue;
        }
        if i == pD0.nNodes {
            if k == pD1.nNodes {
                pD.nNodes = c;
                return 1;
            }
            pD.pNodes[c] = pD1.pNodes[k]; k++;
            continue;
        }
        if pD0.pNodes[i] < pD1.pNodes[k] {
            pD.pNodes[c] = pD0.pNodes[i]; i++;
            continue;
        }
        if pD0.pNodes[i] > pD1.pNodes[k] {
            pD.pNodes[c] = pD1.pNodes[k]; k++;
            continue;
        }
        pD.pNodes[c] = pD0.pNodes[i]; i++; 
        k++;
    }
    if i < pD0.nNodes || k < pD1.nNodes {
        return 0; }
    pD.nNodes = c;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Prepares the object for FPGA mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomMergeTwo(pDom0 *Aig_Dom_t, pDom1 *Aig_Dom_t, pDom *Aig_Dom_t, Limit int32) int32 { 
    assert(Limit > 0, "Limit > 0");
    if pDom0.nNodes < pDom1.nNodes {
        if !(Aig_ObjDomMergeOrdered( pDom1, pDom0, pDom, Limit ) != 0) {
            return 0; }
    
    } else {
        if !(Aig_ObjDomMergeOrdered( pDom0, pDom1, pDom, Limit ) != 0) {
            return 0; }
    }
    pDom.uSign = pDom0.uSign | pDom1.uSign;
    assert(pDom.nNodes <= Limit, "pDom.nNodes <= Limit");
    assert((Aig_ObjDomCheck( pDom ) != 0), "Aig_ObjDomCheck( pDom )");
    return 1;
}

/**Function*************************************************************

  Synopsis    [Merge two arrays of dominators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomMerge(pSto *Aig_Sto_t, vDoms0 *Vec_Ptr_t, vDoms1 *Vec_Ptr_t) *Vec_Ptr_t {
    var vDoms *Vec_Ptr_t;
    var pDom0, pDom1, pDom *Aig_Dom_t;
    var i, k int32;
    vDoms = Vec_PtrAlloc( 16 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vDoms0)) && (((pDom0) = (Aig_Dom_t *)Vec_PtrEntry(vDoms0, i)), 1); i++  {
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vDoms1)) && (((pDom1) = (Aig_Dom_t *)Vec_PtrEntry(vDoms1, k)), 1); k++  {
        if Aig_WordCountOnes( pDom0.uSign | pDom1.uSign ) > pSto.Limit {
            continue; }
        // check if the cut exists
        pDom = (Aig_Dom_t *)Aig_MmFixedEntryFetch( pSto.pMem );
        if !(Aig_ObjDomMergeTwo( pDom0, pDom1, pDom, pSto.Limit ) != 0) {
            Aig_MmFixedEntryRecycle( pSto.pMem, (char *)pDom );
            continue;
        }
        // check if this cut is contained in any of the available cuts
        if Aig_ObjDomFilter( pSto, vDoms, pDom ) != 0 {
            Aig_MmFixedEntryRecycle( pSto.pMem, (char *)pDom );
            continue;
        }
        Vec_PtrPush( vDoms, pDom );
    }
    }
    return vDoms;
}

/**Function*************************************************************

  Synopsis    [Union two arrays of dominators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomUnion(pSto *Aig_Sto_t, vDoms2 *Vec_Ptr_t, vDoms1 *Vec_Ptr_t) {
    var pDom1, pDom2 *Aig_Dom_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vDoms1)) && (((pDom1) = (Aig_Dom_t *)Vec_PtrEntry(vDoms1, i)), 1); i++  {
        if i == 0 {
            continue; }
        if Aig_ObjDomFilter( pSto, vDoms2, pDom1 ) != 0 {
            continue; }
        pDom2 = (Aig_Dom_t *)Aig_MmFixedEntryFetch( pSto.pMem );
        memcpy( pDom2, pDom1, sizeof(Aig_Dom_t) + sizeof(int) * pSto.Limit );
        Vec_PtrPush( vDoms2, pDom2 );
    }
}


/**Function*************************************************************

  Synopsis    [Computes multi-node dominators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomCompute(pSto *Aig_Sto_t, pObj *Aig_Obj_t) {
    var vDoms0, vDoms1, vDoms2, vDomsT *Vec_Ptr_t;
    var pFanout *Aig_Obj_t;
    var i, iFanout int32;
    pSto.nDomNodes += Aig_ObjIsNode(pObj);
    Vec_IntClear( pSto.vFans );
    // inlining of macro Aig_ObjForEachFanout
    for  assert(pSto.pAig.pFanData, "pSto.pAig.pFanData"), i = 0; (i < int32((pObj).nRefs)) &&                         (((iFanout) = i? Aig_ObjFanoutNext(pSto.pAig, iFanout) : Aig_ObjFanout0Int(pSto.pAig, pObj.Id)), 1) &&           (((pFanout) = Aig_ManObj(pSto.pAig, iFanout>>1)), 1); i++  {
        if Aig_ObjIsTravIdCurrent(pSto.pAig, pFanout) {
            Vec_IntPush( pSto.vFans, iFanout>>1 ); }
    }
    if Vec_IntSize(pSto.vFans) == 0 {
        return; }
    vDoms0 = (Vec_Ptr_t *)Vec_PtrEntry( pSto.vDoms, Vec_IntEntry(pSto.vFans, 0) );
    vDoms2 = Aig_ObjDomVecDup( pSto, vDoms0, 0 );
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(pSto.vFans)) && (((iFanout) = Vec_IntEntry(pSto.vFans, i)), 1); i++  {
        vDoms1 = (Vec_Ptr_t *)Vec_PtrEntry( pSto.vDoms, iFanout );
        vDoms2 = Aig_ObjDomMerge( pSto, vDomsT = vDoms2, vDoms1 );
        Aig_ObjDomVecRecycle( pSto, vDomsT );
    }
    Aig_ObjAddTriv( pSto, pObj.Id, vDoms2 );
    pSto.nDomsTotal += Vec_PtrSize(vDoms2);
}

/**Function*************************************************************

  Synopsis    [Marks the flop TFI with the current traversal ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManMarkFlopTfi_rec(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    var Count int32;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return 0; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Aig_ObjIsCi(pObj) || Aig_ObjIsConst1(pObj) {
        return 1; }
    Count = Aig_ManMarkFlopTfi_rec( p, Aig_ObjFanin0(pObj) );
    if Aig_ObjIsNode(pObj) {
        Count += Aig_ManMarkFlopTfi_rec( p, Aig_ObjFanin1(pObj) ); }
    return Count;
}

/**Function*************************************************************

  Synopsis    [Marks the flop TFI with the current traversal ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManMarkFlopTfi(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManIncrementTravId( p );
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
        Aig_ManMarkFlopTfi_rec( p, pObj );
    }
}

/**Function*************************************************************

  Synopsis    [Computes multi-node dominators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManComputeDomsFlops(pAig *Aig_Man_t, Limit int32) *Aig_Sto_t {
    var pSto *Aig_Sto_t;
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    var clk abctime  = Abc_Clock();
    pSto = Aig_ManDomStart( pAig, Limit );
    // initialize flop inputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        Aig_ObjAddTriv( pSto, pObj.Id, Vec_PtrAlloc(1) );
    }
    // compute internal nodes
    vNodes = Aig_ManDfsReverse( pAig );
    Aig_ManMarkFlopTfi( pAig );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Aig_ObjIsTravIdCurrent(pSto.pAig, pObj) {
            Aig_ObjDomCompute( pSto, pObj ); }
    }
    Vec_PtrFree( vNodes );
    // compute combinational inputs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        Aig_ObjDomCompute( pSto, pObj );
    }
    // print statistics
    fmt.Printf("Nodes =%4d. Flops =%4d. Doms =%9d. Ave =%8.2f.   ", pSto.nDomNodes, Aig_ManRegNum(pSto.pAig), pSto.nDomsTotal, //        pSto->nDomsFilter1, pSto->nDomsFilter2,
            1.0 * pSto.nDomsTotal / (pSto.nDomNodes + Aig_ManRegNum(pSto.pAig)));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return pSto;
}

/**Function*************************************************************

  Synopsis    [Computes multi-node dominators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManComputeDomsPis(pAig *Aig_Man_t, Limit int32) *Aig_Sto_t {
    var pSto *Aig_Sto_t;
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    var clk abctime  = Abc_Clock();
    pSto = Aig_ManDomStart( pAig, Limit );
    // initialize flop inputs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        Aig_ObjAddTriv( pSto, pObj.Id, Vec_PtrAlloc(1) );
    }
    // compute internal nodes
    vNodes = Aig_ManDfsReverse( pAig );
    Aig_ManMarkFlopTfi( pAig );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Aig_ObjIsTravIdCurrent(pSto.pAig, pObj) {
            Aig_ObjDomCompute( pSto, pObj ); }
    }
    Vec_PtrFree( vNodes );
    // compute combinational inputs
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        Aig_ObjDomCompute( pSto, pObj );
    }
    // print statistics
    fmt.Printf("Nodes =%4d. PIs =%4d. Doms =%9d. Ave =%8.2f.   ", pSto.nDomNodes, Saig_ManPiNum(pSto.pAig), pSto.nDomsTotal, //        pSto->nDomsFilter1, pSto->nDomsFilter2,
            1.0 * pSto.nDomsTotal / (pSto.nDomNodes + Saig_ManPiNum(pSto.pAig)));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return pSto;
}

/**Function*************************************************************

  Synopsis    [Computes multi-node dominators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManComputeDomsNodes(pAig *Aig_Man_t, Limit int32) *Aig_Sto_t {
    var pSto *Aig_Sto_t;
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    var clk abctime  = Abc_Clock();
    pSto = Aig_ManDomStart( pAig, Limit );
    // initialize flop inputs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        Aig_ObjAddTriv( pSto, pObj.Id, Vec_PtrAlloc(1) );
    }
    // compute internal nodes
    vNodes = Aig_ManDfsReverse( pAig );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Aig_ObjDomCompute( pSto, pObj );
    }
    Vec_PtrFree( vNodes );
    // compute combinational inputs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        Aig_ObjDomCompute( pSto, pObj );
    }
    // print statistics
    fmt.Printf("Nodes =%6d. Doms =%9d. Ave =%8.2f.   ", pSto.nDomNodes, pSto.nDomsTotal, //        pSto->nDomsFilter1, pSto->nDomsFilter2,
            1.0 * pSto.nDomsTotal / pSto.nDomNodes);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return pSto;
}

/**Function*************************************************************

  Synopsis    [Collects dominators from the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomCollect(pSto *Aig_Sto_t, vCut *Vec_Int_t) *Vec_Ptr_t {
    var vDoms0, vDoms1, vDoms2 *Vec_Ptr_t;
    var i, ObjId int32;
    vDoms0 = (Vec_Ptr_t *)Vec_PtrEntry( pSto.vDoms, Vec_IntEntry(vCut, 0) );
    vDoms2 = Aig_ObjDomVecDup( pSto, vDoms0, 1 );
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(vCut)) && (((ObjId) = Vec_IntEntry(vCut, i)), 1); i++  {
        vDoms1 = (Vec_Ptr_t *)Vec_PtrEntry( pSto.vDoms, ObjId );
        if vDoms1 == nil {
            continue; }
        Aig_ObjDomUnion( pSto, vDoms2, vDoms1 );
    }
    return vDoms2;
}


/**Function*************************************************************

  Synopsis    [Marks the flop TFI with the current traversal ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomVolume_rec(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    var Count int32;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return 0; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if pObj.fMarkA {
        return 1; }
//    assert( !Aig_ObjIsCi(pObj) && !Aig_ObjIsConst1(pObj) );
    if Aig_ObjIsCi(pObj) || Aig_ObjIsConst1(pObj) {
        return 1; }
    Count = Aig_ObjDomVolume_rec( p, Aig_ObjFanin0(pObj) );
    if Aig_ObjIsNode(pObj) {
        Count += Aig_ObjDomVolume_rec( p, Aig_ObjFanin1(pObj) ); }
    return Count;
}

/**Function*************************************************************

  Synopsis    [Count the number of nodes in the dominator.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomVolume(pSto *Aig_Sto_t, pDom *Aig_Dom_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    Aig_ManIncrementTravId( pSto.pAig );
    // inlining of macro Aig_DomForEachNode
    for  i = 0; (i < pDom.nNodes) && ((pObj) = Aig_ManObj(pSto.pAig, (pDom).pNodes[i])); i++  {
        Counter += Aig_ObjDomVolume_rec( pSto.pAig, pObj );
    }
    return Counter;
}




/**Function*************************************************************

  Synopsis    [Dereferences the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomDeref_rec(pNode *Aig_Obj_t) int32 {
    var Counter int32;
    assert(pNode.nRefs > 0, "pNode.nRefs > 0");
    if --pNode.nRefs > 0 {
        return 0; }
    assert(pNode.nRefs == 0, "pNode.nRefs == 0");
    if pNode.fMarkA {
        return 1; }
    if Aig_ObjIsCi(pNode) {
        return 0; }
    Counter += Aig_ObjDomDeref_rec( Aig_ObjFanin0(pNode) );
    if Aig_ObjIsNode(pNode) {
    Counter += Aig_ObjDomDeref_rec( Aig_ObjFanin1(pNode) ); }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [References the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomRef_rec(pNode *Aig_Obj_t) int32 {
    var Counter int32;
    assert(pNode.nRefs >= 0, "pNode.nRefs >= 0");
    if pNode.nRefs++ > 0 {
        return 0; }
    assert(pNode.nRefs == 1, "pNode.nRefs == 1");
    if pNode.fMarkA {
        return 1; }
    if Aig_ObjIsCi(pNode) {
        return 0; }
    Counter += Aig_ObjDomRef_rec( Aig_ObjFanin0(pNode) );
    if Aig_ObjIsNode(pNode) {
    Counter += Aig_ObjDomRef_rec( Aig_ObjFanin1(pNode) ); }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Count the number of nodes in the dominator.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomDomed(pSto *Aig_Sto_t, pDom *Aig_Dom_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter0, Counter1 int32;
    Counter0 = 0;
    // inlining of macro Aig_DomForEachNode
    for  i = 0; (i < pDom.nNodes) && ((pObj) = Aig_ManObj(pSto.pAig, (pDom).pNodes[i])); i++  {
        assert(!Aig_ObjIsCi(pObj), "!Aig_ObjIsCi(pObj)");
        Counter0 += Aig_ObjDomDeref_rec( Aig_ObjFanin0(pObj) );
        if Aig_ObjIsNode(pObj) {
        Counter0 += Aig_ObjDomDeref_rec( Aig_ObjFanin1(pObj) ); }
    }
    Counter1 = 0;
    // inlining of macro Aig_DomForEachNode
    for  i = 0; (i < pDom.nNodes) && ((pObj) = Aig_ManObj(pSto.pAig, (pDom).pNodes[i])); i++  {
        assert(!Aig_ObjIsCi(pObj), "!Aig_ObjIsCi(pObj)");
        Counter1 += Aig_ObjDomRef_rec( Aig_ObjFanin0(pObj) );
        if Aig_ObjIsNode(pObj) {
        Counter1 += Aig_ObjDomRef_rec( Aig_ObjFanin1(pObj) ); }
    }
    assert(Counter0 == Counter1, "Counter0 == Counter1");
    return Counter0;
}


/**Function*************************************************************

  Synopsis    [Collects dominators from the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomCollectLos(pSto *Aig_Sto_t) *Vec_Int_t {
    var vCut *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vCut = Vec_IntAlloc( Aig_ManRegNum(pSto.pAig) );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pSto.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pSto.pAig.vCis, i+Saig_ManPiNum(pSto.pAig))), 1); i++  {
        Vec_IntPush( vCut, pObj.Id );
        pObj.fMarkA = 1;
    }
    return vCut;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjPoLogicDeref(pSto *Aig_Sto_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pSto.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pSto.pAig.vCos, i)), 1); i++  {
        Aig_ObjDomDeref_rec( Aig_ObjFanin0(pObj) );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjPoLogicRef(pSto *Aig_Sto_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pSto.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pSto.pAig.vCos, i)), 1); i++  {
        Aig_ObjDomRef_rec( Aig_ObjFanin0(pObj) );
    }
}

/**Function*************************************************************

  Synopsis    [Collects dominators from the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomFindGood(pSto *Aig_Sto_t) {
    var pDom *Aig_Dom_t; 
    var vCut *Vec_Int_t;
    var vDoms *Vec_Ptr_t;
    var i int32;
    vCut  = Aig_ObjDomCollectLos( pSto );
    vDoms = Aig_ObjDomCollect( pSto, vCut );
    Vec_IntFree( vCut );
    fmt.Printf("The cut has %d non-trivial %d-dominators.\n", Vec_PtrSize(vDoms), pSto.Limit);

    Aig_ObjPoLogicDeref( pSto );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vDoms)) && (((pDom) = (Aig_Dom_t *)Vec_PtrEntry(vDoms, i)), 1); i++  {
//        if ( Aig_ObjDomDomed(pSto, pDom) <= 1 )
//            continue;
        fmt.Printf("Vol =%3d.  ", Aig_ObjDomVolume(pSto, pDom));
        fmt.Printf("Dom =%3d.  ", Aig_ObjDomDomed(pSto, pDom));
        Aig_ObjDomPrint( pSto, pDom, i );
    }
    Aig_ObjPoLogicRef( pSto );

    Aig_ObjDomVecRecycle( pSto, vDoms );
    Aig_ManCleanMarkA( pSto.pAig );
}


/**Function*************************************************************

  Synopsis    [Computes multi-node dominators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManComputeDomsTest2(pAig *Aig_Man_t, Num int32) {
    var pSto *Aig_Sto_t;
//    int i;
//Aig_ManShow( pAig, 0, NULL );
    Aig_ManFanoutStart( pAig );
//    for ( i = 1; i < 9; i++ )
    {
        fmt.Printf("ITERATION %d:\n", Num);
        pSto = Aig_ManComputeDomsFlops( pAig, Num );
        Aig_ObjDomFindGood( pSto );
//        Aig_ManDomPrint( pSto );
        Aig_ManDomStop( pSto );
    }
    Aig_ManFanoutStop( pAig );
}

/**Function*************************************************************

  Synopsis    [Computes multi-node dominators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManComputeDomsTest(pAig *Aig_Man_t) {
    var pSto *Aig_Sto_t;
    Aig_ManFanoutStart( pAig );
    pSto = Aig_ManComputeDomsPis( pAig, 1 );
    Aig_ManDomPrint( pSto );
    Aig_ManDomStop( pSto );
    Aig_ManFanoutStop( pAig );
}





/**Function*************************************************************

  Synopsis    [Collects dominators from the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDomCount(pSto *Aig_Sto_t, pObj *Aig_Obj_t) {
    var pDom *Aig_Dom_t;
    var pFanout *Aig_Obj_t;
    var vSingles *Vec_Int_t;
    var vDoms *Vec_Ptr_t;
    var i, k, Entry, iFanout, fPrint int32;
    vSingles = Vec_IntAlloc( 100 );
    // for each dominator of a fanout, count how many fanouts have it as a dominator
    // inlining of macro Aig_ObjForEachFanout
    for  assert(pSto.pAig.pFanData, "pSto.pAig.pFanData"), i = 0; (i < int32((pObj).nRefs)) &&                         (((iFanout) = i? Aig_ObjFanoutNext(pSto.pAig, iFanout) : Aig_ObjFanout0Int(pSto.pAig, pObj.Id)), 1) &&           (((pFanout) = Aig_ManObj(pSto.pAig, iFanout>>1)), 1); i++  {
        vDoms = (Vec_Ptr_t *)Vec_PtrEntry( pSto.vDoms, Aig_ObjId(pFanout) );
        // inlining of macro Vec_PtrForEachEntryStart
        for  k = 1; (k < Vec_PtrSize(vDoms)) && (((pDom) = (Aig_Dom_t *)Vec_PtrEntry(vDoms, k)), 1); k++  {
//            printf( "Fanout %d  Dominator %d\n", Aig_ObjId(pFanout), pDom->pNodes[0] );
            Vec_IntAddToEntry( pSto.vTimes, pDom.pNodes[0], 1 );
            Vec_IntPushUnique( vSingles, pDom.pNodes[0] );
        }
    }
    // clear storage
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSingles)) && (((Entry) = Vec_IntEntry(vSingles, i)), 1); i++  {
        if Vec_IntEntry(pSto.vTimes, Entry) > 5 {
            if fPrint == 0 {
                fmt.Printf("%6d : Level =%4d. Fanout =%6d.\n", Aig_ObjId(pObj), Aig_ObjLevel(pObj), Aig_ObjRefs(pObj));
            }
            fPrint = 1;
            fmt.Printf("%d(%d) ", Entry, Vec_IntEntry(pSto.vTimes, Entry));
        }
        Vec_IntWriteEntry( pSto.vTimes, Entry, 0);
    }
    if fPrint != 0 {
        fmt.Printf("\n"); }
    Vec_IntFree( vSingles );
}


/**Function*************************************************************

  Synopsis    [Computes multi-node dominators.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManComputeDomsForCofactoring(pAig *Aig_Man_t) {
    var vDoms *Vec_Ptr_t;
    var pSto *Aig_Sto_t;
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManFanoutStart( pAig );
    pSto = Aig_ManComputeDomsNodes( pAig, 1 );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsCi(pObj) && !Aig_ObjIsNode(pObj) {
            continue; }
        if Aig_ObjRefs(pObj) < 10 {
            continue; }
        vDoms = (Vec_Ptr_t *)Vec_PtrEntry( pSto.vDoms, Aig_ObjId(pObj) );
//        printf( "%6d : Level =%4d. Fanout =%6d.\n", 
//            Aig_ObjId(pObj), Aig_ObjLevel(pObj), Aig_ObjRefs(pObj) );

        Aig_ObjDomCount( pSto, pObj );
    }
    }
    Aig_ManDomStop( pSto );
    Aig_ManFanoutStop( pAig );
}





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

