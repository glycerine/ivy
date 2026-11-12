package nwk

/**CFile****************************************************************

  FileName    [nwkStrash.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Logic network representation.]

  Synopsis    [Performs structural hashing for the network.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkStrash.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "nwk.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derives AIG from the local functions of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManStrashNode_rec(p *Aig_Man_t, pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return; }
    Nwk_ManStrashNode_rec( p, Hop_ObjFanin0(pObj) ); 
    Nwk_ManStrashNode_rec( p, Hop_ObjFanin1(pObj) );
    pObj.pData = Aig_And( p, (Aig_Obj_t *)Hop_ObjChild0Copy(pObj), (Aig_Obj_t *)Hop_ObjChild1Copy(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
}

/**Function*************************************************************

  Synopsis    [Derives AIG from the local functions of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManStrashNode(p *Aig_Man_t, pObj *Nwk_Obj_t) *Aig_Obj_t {
    var pMan *Hop_Man_t  = pObj.pMan.pManHop;
    var pRoot *Hop_Obj_t  = pObj.pFunc;
    var pFanin *Nwk_Obj_t;
    var i int32;
    assert(Nwk_ObjIsNode(pObj), "Nwk_ObjIsNode(pObj)");
    // check the constant case
    if Hop_Regular(pRoot) == Hop_ManConst1(pMan) {
        return Aig_NotCond( Aig_ManConst1(p), Hop_IsComplement(pRoot) ); }
    // set elementary variables
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pObj).nFanins)) && ((pFanin) = (pObj).pFanio[i]); i++  {
        Hop_IthVar(pMan, i).pData = pFanin.pCopy;
    }
    // strash the AIG of this node
    Nwk_ManStrashNode_rec( p, Hop_Regular(pRoot) );
    Hop_ConeUnmark_rec( Hop_Regular(pRoot) );
    // return the final node
    return Aig_NotCond( (Aig_Obj_t *)Hop_Regular(pRoot).pData, Hop_IsComplement(pRoot) );
}

/**Function*************************************************************

  Synopsis    [Derives AIG from the logic network.]

  Description [Assumes topological ordering of nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManStrash(pNtk *Nwk_Man_t) *Aig_Man_t {
    var vObjs *Vec_Ptr_t;
    var pMan *Aig_Man_t;
    var pObjNew *Aig_Obj_t  = nil;
    var pObj *Nwk_Obj_t;
    var i, Level int32;
    pMan = Aig_ManStart( Nwk_ManGetAigNodeNum(pNtk) );
    pMan.pName = Abc_UtilStrsav( pNtk.pName );
    pMan.pSpec = Abc_UtilStrsav( pNtk.pSpec );
    pMan.pManTime = Tim_ManDup( (Tim_Man_t *)pNtk.pManTime, 1 );
    Tim_ManIncrementTravId( (Tim_Man_t *)pMan.pManTime );
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        pObj.pCopy = nil; }
    }
//    Nwk_ManForEachObj( pNtk, pObj, i )
    vObjs = Nwk_ManDfs( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vObjs, i)), 1); i++  {
        if Nwk_ObjIsCi(pObj) {
            pObjNew = Aig_ObjCreateCi(pMan);
            Level = Tim_ManGetCiArrival( (Tim_Man_t *)pMan.pManTime, pObj.PioId );
            Aig_ObjSetLevel( pObjNew, Level );
        
        } else if Nwk_ObjIsCo(pObj) {
            pObjNew = Aig_ObjCreateCo( pMan, Aig_NotCond((Aig_Obj_t *)Nwk_ObjFanin0(pObj).pCopy, pObj.fInvert) );
            Level = Aig_ObjLevel( pObjNew );
            Tim_ManSetCoArrival( (Tim_Man_t *)pMan.pManTime, pObj.PioId, float32(Level) );
        
        } else if Nwk_ObjIsNode(pObj) {
            pObjNew = Nwk_ManStrashNode( pMan, pObj );
        
        } else {
            assert(false, "0"); }
        pObj.pCopy = pObjNew;
    }
    Vec_PtrFree( vObjs );
    Aig_ManCleanup( pMan );
    Aig_ManSetRegNum( pMan, 0 );
    return pMan;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

