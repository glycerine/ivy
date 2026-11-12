package llb

/**CFile****************************************************************

  FileName    [llb2Bad.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Computing bad states.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb2Bad.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Computes bad in working manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_BddComputeBad(pInit *Aig_Man_t, dd *DdManager, TimeOut abctime) *DdNode {
    var vNodes *Vec_Ptr_t;
    var bBdd0, bBdd1, bTemp, bResult *DdNode;
    var pObj *Aig_Obj_t;
    var i, k int32;
    assert(Cudd_ReadSize(dd) == Aig_ManCiNum(pInit), "Cudd_ReadSize(dd) == Aig_ManCiNum(pInit)");
    // initialize elementary variables
    Aig_ManConst1(pInit).pData = Cudd_ReadOne( dd );
    Saig_ManForEachLo( pInit, pObj, i )
        pObj.pData = Cudd_bddIthVar( dd, i );
    Saig_ManForEachPi( pInit, pObj, i )
        pObj.pData = Cudd_bddIthVar( dd, Aig_ManRegNum(pInit) + i );
    // compute internal nodes
    vNodes = Aig_ManDfsNodes( pInit, (Aig_Obj_t **)Vec_PtrArray(pInit.vCos), Saig_ManPoNum(pInit) );
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
    {
        if !Aig_ObjIsNode(pObj) {
            continue; }
        // inlining of macro Cudd_NotCond
        bBdd0 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin0(pObj).pData) ^ (Aig_ObjFaninC0(pObj))));
        // inlining of macro Cudd_NotCond
        bBdd1 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin1(pObj).pData) ^ (Aig_ObjFaninC1(pObj))));
//        pObj->pData = Extra_bddAndTime( dd, bBdd0, bBdd1, TimeOut );
        pObj.pData = Cudd_bddAnd( dd, bBdd0, bBdd1 );
        if pObj.pData == nil {
            Vec_PtrForEachEntryStop( Aig_Obj_t *, vNodes, pObj, k, i )
                if ( pObj.pData )
                    Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData );
            Vec_PtrFree( vNodes );
            return nil;
        }
        Cudd_Ref( (DdNode *)pObj.pData );
    }
    // quantify PIs of each PO
    bResult = Cudd_ReadLogicZero( dd );  Cudd_Ref( bResult );
    Saig_ManForEachPo( pInit, pObj, i )
    {
        // inlining of macro Cudd_NotCond
        bBdd0   = ((DdNode *)((ptrint)(Aig_ObjFanin0(pObj).pData) ^ (Aig_ObjFaninC0(pObj))));
        bResult = Cudd_bddOr( dd, bTemp = bResult, bBdd0 );     Cudd_Ref( bResult );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    // deref
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
    {
        if !Aig_ObjIsNode(pObj) {
            continue; }
        Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData );
    }
    Vec_PtrFree( vNodes );
    Cudd_Deref( bResult );
    return bResult;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_BddQuantifyPis(pInit *Aig_Man_t, dd *DdManager, bFunc *DdNode) *DdNode {
    var bVar, bCube, bTemp *DdNode;
    var pObj *Aig_Obj_t;
    var i int32;
    var TimeStop abctime;
    assert(Cudd_ReadSize(dd) == Aig_ManCiNum(pInit), "Cudd_ReadSize(dd) == Aig_ManCiNum(pInit)");
    TimeStop = dd.TimeStop; dd.TimeStop = 0;
    // create PI cube
    bCube = Cudd_ReadOne( dd );  Cudd_Ref( bCube );
    Saig_ManForEachPi( pInit, pObj, i )    {
        bVar  = Cudd_bddIthVar( dd, Aig_ManRegNum(pInit) + i );
        bCube = Cudd_bddAnd( dd, bTemp = bCube, bVar );  Cudd_Ref( bCube );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    // quantify PI cube
    bFunc = Cudd_bddExistAbstract( dd, bFunc, bCube );  Cudd_Ref( bFunc );
    Cudd_RecursiveDeref( dd, bCube );
    Cudd_Deref( bFunc );
    dd.TimeStop = TimeStop;
    return bFunc;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

