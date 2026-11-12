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



/**Function*************************************************************

  Synopsis    [Performs combinational cleanup.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Skip the first outputs during cleanup.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Marks CIs/COs/ANDs unreachable from POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Marks CIs/COs/ANDs unreachable from POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Performs sequential cleanup.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


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
    pMaps = nil;
    if (Counter0 != 0) || (Counter != 0) {
        pNew = Gia_ManDupDfsCiMap( p, (int *)pCi2Lit, nil );
    } else {
        pNew = p; }
    // inlining of macro ABC_FREE
    pCi2Lit = nil;
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

