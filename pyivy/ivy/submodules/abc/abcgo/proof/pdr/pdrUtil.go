package pdr

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [pdrUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 20, 2010.]

  Revision    [$Id: pdrUtil.c,v 1.00 2010/11/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pdrInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetAlloc(nSize int32) *Pdr_Set_t {
    var p *Pdr_Set_t;
    assert(nSize >= 0 && nSize < (1<<30), "nSize >= 0 && nSize < (1<<30)");
    p = (Pdr_Set_t *)ABC_CALLOC( char, sizeof(Pdr_Set_t) + nSize * sizeof(int) );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetCreate(vLits *Vec_Int_t, vPiLits *Vec_Int_t) *Pdr_Set_t {
    var p *Pdr_Set_t;
    var i int32;
    assert(Vec_IntSize(vLits) + Vec_IntSize(vPiLits) < (1<<30), "Vec_IntSize(vLits) + Vec_IntSize(vPiLits) < (1<<30)");
    p = (Pdr_Set_t *)ABC_ALLOC( char, sizeof(Pdr_Set_t) + (Vec_IntSize(vLits) + Vec_IntSize(vPiLits)) * sizeof(int) );
    p.nLits  = Vec_IntSize(vLits);
    p.nTotal = Vec_IntSize(vLits) + Vec_IntSize(vPiLits);
    p.nRefs  = 1;
    p.Sign   = 0;
    for  i = 0; i < p.nLits; i++  {
        p.Lits[i] = Vec_IntEntry(vLits, i);
        p.Sign   |= ((word)1 << (p.Lits[i] % 63));
    }
    Vec_IntSelectSort( p.Lits, p.nLits );
    // remember PI literals 
    for  i = p.nLits; i < p.nTotal; i++  {
        p.Lits[i] = Vec_IntEntry(vPiLits, i-p.nLits);
    }
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetCreateFrom(pSet *Pdr_Set_t, iRemove int32) *Pdr_Set_t {
    var p *Pdr_Set_t;
    var i, k int32;
    assert(iRemove >= 0 && iRemove < pSet.nLits, "iRemove >= 0 && iRemove < pSet.nLits");
    p = (Pdr_Set_t *)ABC_ALLOC( char, sizeof(Pdr_Set_t) + (pSet.nTotal - 1) * sizeof(int) );
    p.nLits  = pSet.nLits - 1;
    p.nTotal = pSet.nTotal - 1;
    p.nRefs  = 1;
    p.Sign   = 0;
    for  i = 0; i < pSet.nTotal; i++  {
        if i == iRemove {
            continue; }
        p.Lits[k++] = pSet.Lits[i];
        if i >= pSet.nLits {
            continue; }
        p.Sign   |= ((word)1 << (pSet.Lits[i] % 63));
    }
    assert(k == p.nTotal, "k == p.nTotal");
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetCreateSubset(pSet *Pdr_Set_t, pLits *int32, nLits int32) *Pdr_Set_t {
    var p *Pdr_Set_t;
    var i, k int32;
    assert(nLits >= 0 && nLits <= pSet.nLits, "nLits >= 0 && nLits <= pSet.nLits");
    p = (Pdr_Set_t *)ABC_ALLOC( char, sizeof(Pdr_Set_t) + (nLits + pSet.nTotal - pSet.nLits) * sizeof(int) );
    p.nLits  = nLits;
    p.nTotal = nLits + pSet.nTotal - pSet.nLits;
    p.nRefs  = 1;
    p.Sign   = 0;
    for  i = 0; i < nLits; i++  {
        assert(pLits[i] >= 0, "pLits[i] >= 0");
        p.Lits[k++] = pLits[i];
        p.Sign   |= ((word)1 << (pLits[i] % 63));
    }
    Vec_IntSelectSort( p.Lits, p.nLits );
    for  i = pSet.nLits; i < pSet.nTotal; i++  {
        p.Lits[k++] = pSet.Lits[i];
    }
    assert(k == p.nTotal, "k == p.nTotal");
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetDup(pSet *Pdr_Set_t) *Pdr_Set_t {
    var p *Pdr_Set_t;
    var i int32;
    p = (Pdr_Set_t *)ABC_ALLOC( char, sizeof(Pdr_Set_t) + pSet.nTotal * sizeof(int) );
    p.nLits  = pSet.nLits;
    p.nTotal = pSet.nTotal;
    p.nRefs  = 1;
    p.Sign   = pSet.Sign;
    for  i = 0; i < pSet.nTotal; i++  {
        p.Lits[i] = pSet.Lits[i];
    }
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetRef(p *Pdr_Set_t) *Pdr_Set_t {
    p.nRefs++;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetDeref(p *Pdr_Set_t) {
    if --p.nRefs == 0 {
        ABC_FREE( p ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetPrint(pFile *os.File, p *Pdr_Set_t, nRegs int32, vFlopCounts *Vec_Int_t) {
    var pBuff string;
    var i, k, Entry int32;
    pBuff = ABC_ALLOC( char, nRegs + 1 );
    for  i = 0; i < nRegs; i++  {
        pBuff[i] = '-';
    }
    pBuff[i] = 0;
    for  i = 0; i < p.nLits; i++  {
        if p.Lits[i] == -1 {
            continue; }
        pBuff[Abc_Lit2Var(p.Lits[i])] = (Abc_LitIsCompl(p.Lits[i])? '0':'1');
    }
    if vFlopCounts != nil {
        // skip some literals
        k = 0;
        Vec_IntForEachEntry( vFlopCounts, Entry, i )
            if ( Entry ) 
                pBuff[k++] = pBuff[i];
        pBuff[k] = 0;
    }
    fmt.Fprintf(pFile, "%s", pBuff);
    ABC_FREE( pBuff );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func ZPdr_SetPrint(p *Pdr_Set_t) {
    var i int32;
    for  i = 0; i < p.nLits; i++ {
        fmt.Printf("%d ", p.Lits[i]);
    }
    fmt.Printf("\n");

}

/**Function*************************************************************

  Synopsis    [Return the intersection of p1 and p2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func ZPdr_SetIntersection(p1 *Pdr_Set_t, p2 *Pdr_Set_t, keep *Hash_Int_t) *Pdr_Set_t {
    var pIntersection *Pdr_Set_t;
    var vCommonLits, vPiLits *Vec_Int_t;
    var i, j, nLits int32;
    nLits = p1.nLits;
    if p2.nLits < nLits {
        nLits = p2.nLits; }
    vCommonLits = Vec_IntAlloc( nLits );
    vPiLits = Vec_IntAlloc( 1 );
    for  i, j = 0, 0; i < p1.nLits && j < p2.nLits;  {
        if p1.Lits[i] > p2.Lits[j] {
            if Hash_IntExists( keep, p2.Lits[j] ) {
                //about to drop a literal that should not be dropped
                Vec_IntFree( vCommonLits );
                Vec_IntFree( vPiLits );
                return nil;
            }
            j++;
        
        } else if p1.Lits[i] < p2.Lits[j] {
            if Hash_IntExists( keep, p1.Lits[i] ) {
                //about to drop a literal that should not be dropped
                Vec_IntFree( vCommonLits );
                Vec_IntFree( vPiLits );
                return nil;
            }
            i++;
        
        } else {
          Vec_IntPush( vCommonLits, p1.Lits[i] );
          i++;
          j++;
        }
    }
    pIntersection = Pdr_SetCreate( vCommonLits, vPiLits );
    Vec_IntFree( vCommonLits );
    Vec_IntFree( vPiLits );
    return pIntersection;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetPrintStr(vStr *Vec_Str_t, p *Pdr_Set_t, nRegs int32, vFlopCounts *Vec_Int_t) {
    var pBuff string;
    var i, k, Entry int32;
    pBuff = ABC_ALLOC( char, nRegs + 1 );
    for  i = 0; i < nRegs; i++  {
        pBuff[i] = '-';
    }
    pBuff[i] = 0;
    for  i = 0; i < p.nLits; i++  {
        if p.Lits[i] == -1 {
            continue; }
        pBuff[Abc_Lit2Var(p.Lits[i])] = (Abc_LitIsCompl(p.Lits[i])? '0':'1');
    }
    if vFlopCounts != nil {
        // skip some literals
        Vec_IntForEachEntry( vFlopCounts, Entry, i )
            if ( Entry ) 
                pBuff[k++] = pBuff[i];
        pBuff[k] = 0;
    }
    Vec_StrPushBuffer( vStr, pBuff, k );
    Vec_StrPush( vStr, ' ' );
    Vec_StrPush( vStr, '1' );
    Vec_StrPush( vStr, '\n' );
    ABC_FREE( pBuff );
}

/**Function*************************************************************

  Synopsis    [Return 1 if pOld set-theoretically contains pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetContains(pOld *Pdr_Set_t, pNew *Pdr_Set_t) int32 {
    var pOldInt, pNewInt *int32;
    assert(pOld.nLits > 0, "pOld.nLits > 0");
    assert(pNew.nLits > 0, "pNew.nLits > 0");
    if pOld.nLits < pNew.nLits {
        return 0; }
    if (pOld.Sign & pNew.Sign) != pNew.Sign {
        return 0; }
    pOldInt = pOld.Lits + pOld.nLits - 1;
    pNewInt = pNew.Lits + pNew.nLits - 1;
    for  pNew.Lits <= pNewInt  {
        if pOld.Lits > pOldInt {
            return 0; }
        assert(*pNewInt != -1, "*pNewInt != -1");
        assert(*pOldInt != -1, "*pOldInt != -1");
        if *pNewInt == *pOldInt {
            pNewInt--, pOldInt--;
        } else if *pNewInt < *pOldInt {
            pOldInt--;
        } else {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Return 1 if pOld set-theoretically contains pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetContainsSimple(pOld *Pdr_Set_t, pNew *Pdr_Set_t) int32 {
    var pOldInt, pNewInt *int32;
    assert(pOld.nLits > 0, "pOld.nLits > 0");
    assert(pNew.nLits > 0, "pNew.nLits > 0");
    pOldInt = pOld.Lits + pOld.nLits - 1;
    pNewInt = pNew.Lits + pNew.nLits - 1;
    for  pNew.Lits <= pNewInt  {
        assert(*pOldInt != -1, "*pOldInt != -1");
        if *pNewInt == -1 {
            pNewInt--;
            continue;
        }
        if pOld.Lits > pOldInt {
            return 0; }
        assert(*pNewInt != -1, "*pNewInt != -1");
        assert(*pOldInt != -1, "*pOldInt != -1");
        if *pNewInt == *pOldInt {
            pNewInt--, pOldInt--;
        } else if *pNewInt < *pOldInt {
            pOldInt--;
        } else {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Return 1 if the state cube contains init state (000...0).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetIsInit(pCube *Pdr_Set_t, iRemove int32) int32 {
    var i int32;
    for  i = 0; i < pCube.nLits; i++  {
        assert(pCube.Lits[i] != -1, "pCube.Lits[i] != -1");
        if i == iRemove {
            continue; }
        if Abc_LitIsCompl( pCube.Lits[i] ) == 0 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_SetCompare(pp1 **Pdr_Set_t, pp2 **Pdr_Set_t) int32 {
    var p1 *Pdr_Set_t  = *pp1;
    var p2 *Pdr_Set_t  = *pp2;
    var i int32;
    for  i = 0; i < p1.nLits && i < p2.nLits; i++  {
        if p1.Lits[i] > p2.Lits[i] {
            return -1; }
        if p1.Lits[i] < p2.Lits[i] {
            return 1; }
    }
    if i == p1.nLits && i < p2.nLits {
        return -1; }
    if i < p1.nLits && i == p2.nLits {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_OblStart(k int32, prio int32, pState *Pdr_Set_t, pNext *Pdr_Obl_t) *Pdr_Obl_t {
    var p *Pdr_Obl_t;
    p = new(Pdr_Obl_t);
    p.iFrame = k;
    p.prio   = prio;
    p.nRefs  = 1;
    p.pState = pState;
    p.pNext  = pNext;
    p.pLink  = nil;
    return p;    
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_OblRef(p *Pdr_Obl_t) *Pdr_Obl_t {
    p.nRefs++;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_OblDeref(p *Pdr_Obl_t) {
    if --p.nRefs == 0 {
        if p.pNext {
            Pdr_OblDeref( p.pNext ); }
        Pdr_SetDeref( p.pState );
        ABC_FREE( p );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_QueueIsEmpty(p *Pdr_Man_t) int32 {
    return p.pQueue == nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_QueueHead(p *Pdr_Man_t) *Pdr_Obl_t {
    return p.pQueue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_QueuePop(p *Pdr_Man_t) *Pdr_Obl_t {
    var pRes *Pdr_Obl_t  = p.pQueue;
    if p.pQueue == nil {
        return nil; }
    p.pQueue = p.pQueue.pLink;
    Pdr_OblDeref( pRes );
    p.nQueCur--;
    return pRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_QueueClean(p *Pdr_Man_t) {
    var pThis *Pdr_Obl_t;
    for  pThis = Pdr_QueuePop(p)  {
        Pdr_OblDeref( pThis );
    }
    pThis = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_QueuePush(p *Pdr_Man_t, pObl *Pdr_Obl_t) {
    var pTemp *Pdr_Obl_t;
    var ppPrev **Pdr_Obl_t;
    p.nObligs++;
    p.nQueCur++;
    p.nQueMax = Abc_MaxInt( p.nQueMax, p.nQueCur );
    Pdr_OblRef( pObl );
    if p.pQueue == nil {
        p.pQueue = pObl;
        return;
    }
    for  ppPrev, pTemp = &p.pQueue, p.pQueue; (pTemp != nil); ppPrev = &pTemp.pLink, pTemp = pTemp.pLink  {
        if pTemp.iFrame > pObl.iFrame || (pTemp.iFrame == pObl.iFrame && pTemp.prio > pObl.prio) {
            break; }
    }
    *ppPrev = pObl;
    pObl.pLink = pTemp;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_QueuePrint(p *Pdr_Man_t) {
    var pObl *Pdr_Obl_t;
    for  pObl = p.pQueue; (pObl != nil); pObl = pObl.pLink  {
        Abc_Print( 1, "Frame = %2d.  Prio = %8d.\n", pObl.iFrame, pObl.prio );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_QueueStop(p *Pdr_Man_t) {
    var pObl *Pdr_Obl_t;
    for  !(Pdr_QueueIsEmpty(p) != 0)  {
        pObl = Pdr_QueuePop(p);
        Pdr_OblDeref( pObl );
    }
    p.pQueue = nil;
    p.nQueCur = 0;
}


const PDR_VAL0 =  1
const PDR_VAL1 =  2
const PDR_VALX =  3

/**Function*************************************************************

  Synopsis    [Returns value (0 or 1) or X if unassigned.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ObjSatValue(pAig *Aig_Man_t, pNode *Aig_Obj_t, fCompl int32) int32 {
    if Aig_ObjIsTravIdCurrent(pAig, pNode) {
        if pNode.fMarkA ^ fCompl {
            return PDR_VAL1;
        } else {
            return PDR_VAL0;
        } }
    return PDR_VALX;
}

/**Function*************************************************************

  Synopsis    [Recursively searched for a satisfying assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_NtkFindSatAssign_rec(pAig *Aig_Man_t, pNode *Aig_Obj_t, Value int32, pCube *Pdr_Set_t, Heur int32) int32 {
    var Value0, Value1 int32;
    if Aig_ObjIsConst1(pNode) {
        return 1; }
    if Aig_ObjIsTravIdCurrent(pAig, pNode) {
        return (int32(pNode.fMarkA) == Value); }
    Aig_ObjSetTravIdCurrent(pAig, pNode);
    pNode.fMarkA = Value;
    if Aig_ObjIsCi(pNode) {
        if Saig_ObjIsLo(pAig, pNode) {
//            pCube->Lits[pCube->nLits++] = Abc_Var2Lit( Aig_ObjCioId(pNode) - Saig_ManPiNum(pAig), !Value );
            pCube.Lits[pCube.nLits++] = Abc_Var2Lit( Aig_ObjCioId(pNode) - Saig_ManPiNum(pAig), Value );
            pCube.Sign |= ((word)1 << (pCube.Lits[pCube.nLits-1] % 63));
        }
        return 1;
    }
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    // propagation
    if Value != 0 {
        if !(Pdr_NtkFindSatAssign_rec(pAig, Aig_ObjFanin0(pNode), !Aig_ObjFaninC0(pNode), pCube, Heur) != 0) {
            return 0; }
        return Pdr_NtkFindSatAssign_rec(pAig, Aig_ObjFanin1(pNode), !Aig_ObjFaninC1(pNode), pCube, Heur);
    }
    // justification
    Value0 = Pdr_ObjSatValue( pAig, Aig_ObjFanin0(pNode), Aig_ObjFaninC0(pNode) );
    if Value0 == PDR_VAL0 {
        return 1; }
    Value1 = Pdr_ObjSatValue( pAig, Aig_ObjFanin1(pNode), Aig_ObjFaninC1(pNode) );
    if Value1 == PDR_VAL0 {
        return 1; }
    if Value0 == PDR_VAL1 && Value1 == PDR_VAL1 {
        return 0; }
    if Value0 == PDR_VAL1 {
        return Pdr_NtkFindSatAssign_rec(pAig, Aig_ObjFanin1(pNode), Aig_ObjFaninC1(pNode), pCube, Heur); }
    if Value1 == PDR_VAL1 {
        return Pdr_NtkFindSatAssign_rec(pAig, Aig_ObjFanin0(pNode), Aig_ObjFaninC0(pNode), pCube, Heur); }
    assert(Value0 == PDR_VALX && Value1 == PDR_VALX, "Value0 == PDR_VALX && Value1 == PDR_VALX");
    // decision making
//    if ( rand() % 10 == Heur )
    if Aig_ObjId(pNode) % 4 == Heur {
        return Pdr_NtkFindSatAssign_rec(pAig, Aig_ObjFanin1(pNode), Aig_ObjFaninC1(pNode), pCube, Heur);
    } else {
        return Pdr_NtkFindSatAssign_rec(pAig, Aig_ObjFanin0(pNode), Aig_ObjFaninC0(pNode), pCube, Heur); }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

