package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaScl.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Sequential cleanup.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaScl.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Marks unreachable internal nodes and returns their number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCombMarkUsed_rec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if pObj == nil {
        return 0; }
    if !pObj.fMark0 {
        return 0; }
    pObj.fMark0 = 0;
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
    return 1 + Gia_ManCombMarkUsed_rec( p, Gia_ObjFanin0(pObj) )
             + Gia_ManCombMarkUsed_rec( p, Gia_ObjFanin1(pObj) )
             + (p.pNexts ? Gia_ManCombMarkUsed_rec( p, Gia_ObjNextObj(p, Gia_ObjId(p, pObj)) ) : 0)
             + (p.pSibls ? Gia_ManCombMarkUsed_rec( p, Gia_ObjSiblObj(p, Gia_ObjId(p, pObj)) ) : 0)
             + (p.pMuxes ? Gia_ManCombMarkUsed_rec( p, Gia_ObjFanin2(p, pObj) ) : 0);
}
func Gia_ManCombMarkUsed(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, nNodes int32;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        pObj.fMark0 = Gia_ObjIsAnd(pObj) && !Gia_ObjIsBuf(pObj);
    }
    // inlining of macro Gia_ManForEachBuf
    for  i = Gia_ManBufNum(p) ? 0 : p.nObjs; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsBuf(pObj) {} else {
        nNodes += Gia_ManCombMarkUsed_rec( p, Gia_ObjFanin0(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        nNodes += Gia_ManCombMarkUsed_rec( p, Gia_ObjFanin0(pObj) );
    }
    return nNodes;
}

/**Function*************************************************************

  Synopsis    [Performs combinational cleanup.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCleanup(p *Gia_Man_t) *Gia_Man_t {
    Gia_ManCombMarkUsed( p );
    return Gia_ManDupMarked( p );
}

/**Function*************************************************************

  Synopsis    [Skip the first outputs during cleanup.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCleanupOutputs(p *Gia_Man_t, nOutputs int32) *Gia_Man_t {
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(nOutputs < Gia_ManCoNum(p), "nOutputs < Gia_ManCoNum(p)");
    Gia_ManCombMarkUsed( p );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if i < nOutputs {
            pObj.fMark0 = 1;
        } else {
            break; }
    }
    return Gia_ManDupMarked( p );
}


/**Function*************************************************************

  Synopsis    [Marks CIs/COs/ANDs unreachable from POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSeqMarkUsed_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vRoots *Vec_Int_t) int32 {
    if !pObj.fMark0 {
        return 0; }
    pObj.fMark0 = 0;
    if Gia_ObjIsCo(pObj) {
        return Gia_ManSeqMarkUsed_rec( p, Gia_ObjFanin0(pObj), vRoots ); }
    if Gia_ObjIsRo(p, pObj) {
        Vec_IntPush( vRoots, Gia_ObjId(p, Gia_ObjRoToRi(p, pObj)) );
        return 0;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    return 1 + Gia_ManSeqMarkUsed_rec( p, Gia_ObjFanin0(pObj), vRoots )
             + Gia_ManSeqMarkUsed_rec( p, Gia_ObjFanin1(pObj), vRoots );
}

/**Function*************************************************************

  Synopsis    [Marks CIs/COs/ANDs unreachable from POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSeqMarkUsed(p *Gia_Man_t) int32 {
    var vRoots *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, nNodes int32;
    Gia_ManSetMark0( p );
    Gia_ManConst0(p).fMark0 = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.fMark0 = 0;
    }
    vRoots = Gia_ManCollectPoIds( p );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vRoots)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vRoots,i))); i++  {
        nNodes += Gia_ManSeqMarkUsed_rec( p, pObj, vRoots );
    }
    Vec_IntFree( vRoots );
    return nNodes;
}

/**Function*************************************************************

  Synopsis    [Performs sequential cleanup.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSeqCleanup(p *Gia_Man_t) *Gia_Man_t {
    Gia_ManSeqMarkUsed( p );
    return Gia_ManDupMarked( p );
}

/**Function*************************************************************

  Synopsis    [Find representatives due to identical fanins.]

  Description [Returns the old manager if there is no changes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManReduceEquiv(p *Gia_Man_t, fVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObjRi, pObjRo *Gia_Obj_t;
    var pCi2Lit, pMaps *uint32;
    var i int32;
    var iLit int32;
    var nFanins int32 = 1;
    var Counter0 int32;
    var Counter int32;
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        Gia_ObjFanin0(pObjRi).Value = 0;
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        if Gia_ObjFanin0(pObjRi).Value == 0 {
            Gia_ObjFanin0(pObjRi).Value = 2*nFanins++; }
    }
    // inlining of macro ABC_FALLOC
    pCi2Lit = ((unsigned *) memset(malloc(sizeof(unsigned) * (size_t)(Gia_ManCiNum(p))), 0xff, sizeof(unsigned) * (size_t)(Gia_ManCiNum(p))));
    // inlining of macro ABC_FALLOC
    pMaps   = ((unsigned *) memset(malloc(sizeof(unsigned) * (size_t)(2 * nFanins)), 0xff, sizeof(unsigned) * (size_t)(2 * nFanins)));
    // inlining of macro Gia_ManForEachRiRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        iLit = Gia_ObjFanin0Copy( pObjRi );
        if Gia_ObjFaninId0p(p, pObjRi) == 0 && Gia_ObjFaninC0(pObjRi) == 0 { // const 0 
            pCi2Lit[Gia_ManPiNum(p)+i] = 0, Counter0++;
        } else if ~pMaps[iLit] { // in this case, ID(pObj) > ID(pRepr) 
            pCi2Lit[Gia_ManPiNum(p)+i] = pMaps[iLit], Counter++; 
        } else {
            pMaps[iLit] = Abc_Var2Lit( Gia_ObjId(p, pObjRo), 0 ); }
    }
/*
    Gia_ManForEachCi( p, pObjRo, i )
    {
        if ( ~pCi2Lit[i] )
        {
            Gia_Obj_t * pObj0 = Gia_ObjRoToRi(p, pObjRo);
            Gia_Obj_t * pObj1 = Gia_ObjRoToRi(p, Gia_ManObj(p, pCi2Lit[i]));
            Gia_Obj_t * pFan0 = Gia_ObjChild0( p, Gia_ObjRoToRi(p, pObjRo) );
            Gia_Obj_t * pFan1 = Gia_ObjChild0( p, Gia_ObjRoToRi(p, Gia_ManObj(p, pCi2Lit[i])) );
            assert( pFan0 == pFan1 );
        }
    }
*/
//    if ( fVerbose )
//        printf( "ReduceEquiv detected %d constant regs and %d equivalent regs.\n", Counter0, Counter );
    // inlining of macro ABC_FREE
    ((pMaps) ? (free((char *) (pMaps)), (pMaps) = 0) : 0);
    if (Counter0 != 0) || (Counter != 0) {
        pNew = Gia_ManDupDfsCiMap( p, (int *)pCi2Lit, nil );
    } else {
        pNew = p; }
    // inlining of macro ABC_FREE
    ((pCi2Lit) ? (free((char *) (pCi2Lit)), (pCi2Lit) = 0) : 0);
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Performs sequential cleanup.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSeqStructSweep(p *Gia_Man_t, fConst int32, fEquiv int32, fVerbose int32) *Gia_Man_t {
    var pTemp *Gia_Man_t;
    if Gia_ManRegNum(p) == 0 {
        return Gia_ManCleanup( p ); }
    if fVerbose != 0 {
        fmt.Printf("Performing sequential cleanup.\n"); }
    p = Gia_ManSeqCleanup( pTemp = p );
    if fVerbose != 0 {
        Gia_ManReportImprovement( pTemp, p ); }
    if (fConst != 0) && Gia_ManRegNum(p) {
        p = Gia_ManReduceConst( pTemp = p, fVerbose );
        if fVerbose != 0 {
            Gia_ManReportImprovement( pTemp, p ); }
        Gia_ManStop( pTemp );
    }
    if (fVerbose != 0) && (fEquiv != 0) {
        fmt.Printf("Merging combinationally equivalent flops.\n"); }
    if fEquiv != 0 {
    for {
        p = Gia_ManSeqCleanup( pTemp = p );
        if fVerbose != 0 {
            Gia_ManReportImprovement( pTemp, p ); }
        Gia_ManStop( pTemp );
        if Gia_ManRegNum(p) == 0 {
            break; }
        p = Gia_ManReduceEquiv( pTemp = p, fVerbose );
        if p == pTemp {
            break; }
        Gia_ManStop( pTemp );
    } }
    return p;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

