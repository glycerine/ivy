package llb

/**CFile****************************************************************

  FileName    [llb2Map.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Non-linear quantification scheduling.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb2Map.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"
//#include "base/abc/abc.h"
//#include "map/if/if.h"

//ABC_NAMESPACE_IMPL_START
 

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns internal nodes used in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_AigMap(pAig *Aig_Man_t, nLutSize int32, nLutMin int32) *Vec_Int_t {
    extern Abc_Ntk_t * Abc_NtkFromAigPhase( Aig_Man_t * pMan );
    extern If_Man_t *  Abc_NtkToIf( Abc_Ntk_t * pNtk, If_Par_t * pPars );
    extern void        Gia_ManSetIfParsDefault( void * pPars );
    var Pars If_Par_t;
    var pPars *If_Par_t = &Pars;
    var pIfMan *If_Man_t;
    var pAnd *If_Obj_t;
    var pNtk *Abc_Ntk_t;
    var pNode *Abc_Obj_t;
    var vNodes *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;

    // create ABC network
    pNtk = Abc_NtkFromAigPhase( pAig );
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");

    // derive mapping parameters
    Gia_ManSetIfParsDefault( pPars );
    pPars.nLutSize = nLutSize;

    // get timing information
    pPars.pTimesArr = Abc_NtkGetCiArrivalFloats(pNtk);
    pPars.pTimesReq = Abc_NtkGetCoRequiredFloats(pNtk);

    // perform LUT mapping
    pIfMan = Abc_NtkToIf( pNtk, pPars );    
    if pIfMan == nil {
        Abc_NtkDelete( pNtk );
        return nil;
    }
    if !If_ManPerformMapping( pIfMan ) {
        Abc_NtkDelete( pNtk );
        If_ManStop( pIfMan );
        return nil;
    }

    // mark nodes in the AIG used in the mapping
    Aig_ManCleanMarkA( pAig );
    Aig_ManForEachNode( pAig, pObj, i )
    {
        pNode = (Abc_Obj_t *)pObj.pData;
        if pNode == nil {
            continue; }
        pAnd = (If_Obj_t *)pNode.pCopy;
        if pAnd == nil {
            continue; }
        if pAnd.nRefs > 0 && int32(If_ObjCutBest(pAnd).nLeaves) >= nLutMin {
            pObj.fMarkA = 1; }
    }
    Abc_NtkDelete( pNtk );
    If_ManStop( pIfMan );

    // unmark flop drivers
    Saig_ManForEachLi( pAig, pObj, i )
        Aig_ObjFanin0(pObj).fMarkA = 0;

    // collect mapping
    vNodes = Vec_IntAlloc( 100 );
    Aig_ManForEachNode( pAig, pObj, i )
        if ( pObj.fMarkA )
            Vec_IntPush( vNodes, Aig_ObjId(pObj) );
    Aig_ManCleanMarkA( pAig );
    return vNodes;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

