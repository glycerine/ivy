package gia




/**CFile****************************************************************

  FileName    [giaMap.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Manipulation of mapping associated with the AIG.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaMap.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "aig/aig/aig.h"
//#include "map/if/if.h"
//#include "bool/kit/kit.h"
//#include "base/main/main.h"
//#include "sat/bsat/satSolver.h"

// Windows console headers are unnecessary in Go.

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );

// extern int Abc_RecToGia3( Gia_Man_t * pMan, If_Man_t * pIfMan, If_Cut_t * pCut, Vec_Int_t * vLeaves, int fHash );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Load the network into FPGA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Calculate mapping overlap.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Calculate mapping overlap.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Computes levels for AIG with choices and white boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

 


/**Function*************************************************************

  Synopsis    [Checks integrity of choice nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Converts GIA into IF manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Rebuilds GIA from mini AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Converts IF into GIA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Write mapping for LUT with given fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Write mapping for LUT with given fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Write the node into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Recursively derives the local AIG for the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Converts IF into GIA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Converts IF into GIA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFromIfLogicFindCell(pIfMan *If_Man_t, pNew *Gia_Man_t, pTemp *Gia_Man_t, pCutBest *If_Cut_t, pNtkCell *Ifn_Ntk_t, nLutMax int32, vLeaves *Vec_Int_t, vLits *Vec_Int_t, vCover *Vec_Int_t, vMapping *Vec_Int_t, vMapping2 *Vec_Int_t, vConfigs *Vec_Int_t) int32 {
    var iLit int32;
    assert(false, "0");
    if Vec_IntSize(vLeaves) <= nLutMax {
        iLit = Gia_ManFromIfLogicCreateLut( pNew, If_CutTruthW(pIfMan, pCutBest), vLeaves, vCover, vMapping, vMapping2 );
    } else {
        var pObj *Gia_Obj_t;
        var i, Id, iLitTemp int32;
        // extract variable permutation
        //char * pCutPerm = If_CutDsdPerm( pIfMan, pCutBest ); // DSD input -> cut input
        var pPerm *word  = If_DsdManGetFuncConfig( pIfMan.pIfDsdMan, If_CutDsdLit(pIfMan, pCutBest) ); // cell input -> DSD input
        //int nBits = If_DsdManTtBitNum( pIfMan->pIfDsdMan );
        // use config bits to generate the network
        iLit = If_ManSatDeriveGiaFromBits( pTemp, pNtkCell, pPerm + 1, vLeaves, vCover );
        // copy GIA back into the manager
        Vec_IntFillExtra( &pTemp.vCopies, Gia_ManObjNum(pTemp), -1 );
        Gia_ObjSetCopyArray( pTemp, 0, 0 );
        Vec_IntForEachEntry( vLeaves, iLitTemp, i )
            Gia_ObjSetCopyArray( pTemp, Gia_ManCiIdToId(pTemp, i), iLitTemp );
        // collect nodes
        Gia_ManIncrementTravId( pTemp );
        Id = Abc_Lit2Var( iLit );
        Gia_ManCollectAnds( pTemp, &Id, 1, vCover, nil );
        Vec_IntPrint( vCover );
        Gia_ManForEachObjVec( vCover, pTemp, pObj, i )
            Gia_ObjPrint( pTemp, pObj );
        // copy GIA
        Gia_ManForEachObjVec( vCover, pTemp, pObj, i )
        {
            iLit = Gia_ManAppendAnd( pNew, Gia_ObjFanin0CopyArray(pTemp, pObj), Gia_ObjFanin1CopyArray(pTemp, pObj) );
            Gia_ObjSetCopyArray( pTemp, Gia_ObjId(pTemp, pObj), iLit );
        }
        iLit = Abc_LitNotCond( Gia_ObjCopyArray(pTemp, Id), Abc_LitIsCompl(iLit) );
    }
    // write packing
//    if ( vConfigs && !Gia_ObjIsCi(Gia_ManObj(pNew, Abc_Lit2Var(iLit))) && iLit > 1 )
//        Gia_ManFromIfGetConfig( vConfigs, pIfMan, pCutBest );
    return iLit;
}


/**Function*************************************************************

  Synopsis    [Converts IF into GIA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Converts IF into GIA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Verifies mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Transfers mapping from hie GIA to normalized GIA.]

  Description [Hie GIA (pGia) points to normalized GIA (p).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Interface of LUT mapping package.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Interface of other mapping-based procedures.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Tests decomposition structures.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs hashing for a mapped AIG.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

