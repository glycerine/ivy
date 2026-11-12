package dch

/**CFile****************************************************************

  FileName    [dchAig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Choice computation for tech-mapping.]

  Synopsis    [AIG manipulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: dchAig.c,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dchInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derives the cumulative AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_DeriveTotalAig_rec(p *Aig_Man_t, pObj *Aig_Obj_t) {
    if pObj.pData {
        return; }
    Dch_DeriveTotalAig_rec( p, Aig_ObjFanin0(pObj) );
    Dch_DeriveTotalAig_rec( p, Aig_ObjFanin1(pObj) );
    pObj.pData = Aig_And( p, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Derives the cumulative AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_DeriveTotalAig(vAigs *Vec_Ptr_t) *Aig_Man_t {
    var pAig, pAig2, pAigTotal *Aig_Man_t;
    var pObj, pObjPi, pObjPo *Aig_Obj_t;
    var i, k, nNodes int32;
    assert(Vec_PtrSize(vAigs) > 0, "Vec_PtrSize(vAigs) > 0");
    // make sure they have the same number of PIs/POs
    nNodes = 0;
    pAig = (Aig_Man_t *)Vec_PtrEntry( vAigs, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vAigs)) && (((pAig2) = (Aig_Man_t *)Vec_PtrEntry(vAigs, i)), 1); i++  {
        assert(Aig_ManCiNum(pAig) == Aig_ManCiNum(pAig2), "Aig_ManCiNum(pAig) == Aig_ManCiNum(pAig2)");
        assert(Aig_ManCoNum(pAig) == Aig_ManCoNum(pAig2), "Aig_ManCoNum(pAig) == Aig_ManCoNum(pAig2)");
        nNodes += Aig_ManNodeNum(pAig2);
        Aig_ManCleanData( pAig2 );
    }
    // map constant nodes
    pAigTotal = Aig_ManStart( nNodes );
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vAigs)) && (((pAig2) = (Aig_Man_t *)Vec_PtrEntry(vAigs, k)), 1); k++  {
        Aig_ManConst1(pAig2).pData = Aig_ManConst1(pAigTotal);
    }
    // map primary inputs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObjPi = Aig_ObjCreateCi( pAigTotal );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vAigs)) && (((pAig2) = (Aig_Man_t *)Vec_PtrEntry(vAigs, k)), 1); k++  {
            Aig_ManCi( pAig2, i ).pData = pObjPi;
        }
    }
    // construct the AIG in the order of POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vAigs)) && (((pAig2) = (Aig_Man_t *)Vec_PtrEntry(vAigs, k)), 1); k++  {
            pObjPo = Aig_ManCo( pAig2, i );
            Dch_DeriveTotalAig_rec( pAigTotal, Aig_ObjFanin0(pObjPo) );
        }
        Aig_ObjCreateCo( pAigTotal, Aig_ObjChild0Copy(pObj) );
    }
/*
    // mark the cone of the first AIG
    Aig_ManIncrementTravId( pAigTotal );
    Aig_ManForEachObj( pAig, pObj, i )
        if ( pObj->pData ) 
            Aig_ObjSetTravIdCurrent( pAigTotal, pObj->pData );
*/
    // cleanup should not be done
    return pAigTotal;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

