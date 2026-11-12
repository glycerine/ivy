package res

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [resStrash.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resynthesis package.]

  Synopsis    [Structural hashing of the nodes in the window.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 15, 2007.]

  Revision    [$Id: resStrash.c,v 1.00 2007/01/15 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "resInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern Abc_Obj_t * Abc_ConvertAigToAig( Abc_Ntk_t * pAig, Abc_Obj_t * pObjOld );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Structurally hashes the given window.]

  Description [The first PO is the observability condition. The second 
  is the node's function. The remaining POs are the candidate divisors.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_WndStrash(p *Res_Win_t) *Abc_Ntk_t {
    var vPairs *Vec_Ptr_t;
    var pAig *Abc_Ntk_t;
    var pObj, pMiter *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkHasAig(p.pNode.pNtk), "Abc_NtkHasAig(p.pNode.pNtk)");
//    Abc_NtkCleanCopy( p->pNode->pNtk );
    // create the network
    pAig = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pAig.pName = Extra_UtilStrsav( "window" );
    // create the inputs
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(p.vLeaves)); i++ { pObj = Vec_PtrEntry(p.vLeaves, i);
        pObj.pCopy = Abc_NtkCreatePi( pAig );
    }
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(p.vBranches)); i++ { pObj = Vec_PtrEntry(p.vBranches, i);
        pObj.pCopy = Abc_NtkCreatePi( pAig );
    }
    // go through the nodes in the topological order
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(p.vNodes)); i++ { pObj = Vec_PtrEntry(p.vNodes, i);
        pObj.pCopy = Abc_ConvertAigToAig( pAig, pObj );
        if pObj == p.pNode {
            pObj.pCopy = Abc_ObjNot( pObj.pCopy ); }
    }
    // collect the POs
    vPairs = Vec_PtrAlloc( 2 * Vec_PtrSize(p.vRoots) );
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(p.vRoots)); i++ { pObj = Vec_PtrEntry(p.vRoots, i);
        Vec_PtrPush( vPairs, pObj.pCopy );
        Vec_PtrPush( vPairs, nil );
    }
    // mark the TFO of the node
    Abc_NtkIncrementTravId( p.pNode.pNtk );
    Res_WinSweepLeafTfo_rec( p.pNode, int32(p.pNode.Level) + p.nWinTfoMax );
    // update strashing of the node
    p.pNode.pCopy = Abc_ObjNot( p.pNode.pCopy );
    Abc_NodeSetTravIdPrevious( p.pNode );
    // redo strashing in the TFO
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(p.vNodes)); i++ { pObj = Vec_PtrEntry(p.vNodes, i);
        if Abc_NodeIsTravIdCurrent(pObj) {
            pObj.pCopy = Abc_ConvertAigToAig( pAig, pObj ); }
    }
    // collect the POs
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(p.vRoots)); i++ { pObj = Vec_PtrEntry(p.vRoots, i);
        Vec_PtrWriteEntry( vPairs, 2 * i + 1, pObj.pCopy );
    }
    // add the miter
    pMiter = Abc_AigMiter( (Abc_Aig_t *)pAig.pManFunc, vPairs, 0 );
    Abc_ObjAddFanin( Abc_NtkCreatePo(pAig), pMiter );
    Vec_PtrFree( vPairs );
    // add the node
    Abc_ObjAddFanin( Abc_NtkCreatePo(pAig), p.pNode.pCopy );
    // add the fanins
    // inlining of macro Abc_ObjForEachFanin
    for i = 0; (i < Abc_ObjFaninNum(p.pNode)); i++ { pObj = Abc_ObjFanin(p.pNode, i);
        Abc_ObjAddFanin( Abc_NtkCreatePo(pAig), pObj.pCopy );
    }
    // add the divisors
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; (i < Vec_PtrSize(p.vDivs)); i++ { pObj = Vec_PtrEntry(p.vDivs, i);
        Abc_ObjAddFanin( Abc_NtkCreatePo(pAig), pObj.pCopy );
    }
    // add the names
    Abc_NtkAddDummyPiNames( pAig );
    Abc_NtkAddDummyPoNames( pAig );
    // check the resulting network
    if !Abc_NtkCheck( pAig ) {
        fmt.Fprintf(os.Stdout, "Res_WndStrash(): Network check has failed.\n"); }
    return pAig;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

