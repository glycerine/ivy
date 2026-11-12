package abci

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcSense.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abc_.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "proof/fraig/fraig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Copies the topmost levels of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSensitivityMiter_rec(pNtkNew *Abc_Ntk_t, pNode *Abc_Obj_t) *Abc_Obj_t {
    assert(!Abc_ObjIsComplement(pNode), "!Abc_ObjIsComplement(pNode)");
    if pNode.pCopy {
        return pNode.pCopy; }
    Abc_NtkSensitivityMiter_rec( pNtkNew, Abc_ObjFanin0(pNode) );
    Abc_NtkSensitivityMiter_rec( pNtkNew, Abc_ObjFanin1(pNode) );
    return pNode.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, Abc_ObjChild0Copy(pNode), Abc_ObjChild1Copy(pNode) );
}

/**Function*************************************************************

  Synopsis    [Creates miter for the sensitivity analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSensitivityMiter(pNtk *Abc_Ntk_t, iVar int32) *Abc_Ntk_t {
    var pMiter *Abc_Ntk_t;
    var vNodes *Vec_Ptr_t;
    var pObj, pNext, pFanin, pOutput, pObjNew *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    assert(iVar < Abc_NtkCiNum(pNtk), "iVar < Abc_NtkCiNum(pNtk)");

    // duplicate the network
    pMiter = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pMiter.pName = Extra_UtilStrsav(pNtk.pName);
    pMiter.pSpec = Extra_UtilStrsav(pNtk.pSpec);

    // assign the PIs
    Abc_NtkCleanCopy( pNtk );
    Abc_AigConst1(pNtk).pCopy = Abc_AigConst1(pMiter);
    Abc_AigConst1(pNtk).pData = Abc_AigConst1(pMiter);
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.pCopy = Abc_NtkCreatePi( pMiter );
        pObj.pData = pObj.pCopy;
    }
    Abc_NtkAddDummyPiNames( pMiter );

    // assign the cofactors of the CI node to be constants
    pObj = Abc_NtkCi( pNtk, iVar );
    pObj.pCopy = Abc_ObjNot( Abc_AigConst1(pMiter) ); 
    pObj.pData = Abc_AigConst1(pMiter); 

    // collect the internal nodes
    vNodes = Abc_NtkDfsReverseNodes( pNtk, &pObj, 1 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        for  pNext = pObj? pObj.pCopy : pObj; (pObj != nil); pObj = pNext, pNext = pObj? pObj.pCopy : pObj  {
            pFanin = Abc_ObjFanin0(pObj);
            if !Abc_NodeIsTravIdCurrent(pFanin) {
                pFanin.pData = Abc_NtkSensitivityMiter_rec( pMiter, pFanin ); }
            pFanin = Abc_ObjFanin1(pObj);
            if !Abc_NodeIsTravIdCurrent(pFanin) {
                pFanin.pData = Abc_NtkSensitivityMiter_rec( pMiter, pFanin ); }
            pObj.pCopy = Abc_AigAnd( (Abc_Aig_t *)pMiter.pManFunc, Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj) );
            pObj.pData = Abc_AigAnd( (Abc_Aig_t *)pMiter.pManFunc, Abc_ObjChild0Data(pObj), Abc_ObjChild1Data(pObj) );
        }
    }
    Vec_PtrFree( vNodes );

    // update the affected COs
    pOutput = Abc_ObjNot( Abc_AigConst1(pMiter) ); 
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if !Abc_NodeIsTravIdCurrent(pObj) {
            continue; }
        // get the result of quantification
        if i == Abc_NtkCoNum(pNtk) - 1 {
            pOutput = Abc_AigAnd( (Abc_Aig_t *)pMiter.pManFunc, pOutput, Abc_ObjChild0Data(pObj) );
            pOutput = Abc_AigAnd( (Abc_Aig_t *)pMiter.pManFunc, pOutput, Abc_ObjChild0Copy(pObj) );
        
        } else {
            pNext   = Abc_AigXor( (Abc_Aig_t *)pMiter.pManFunc, Abc_ObjChild0Copy(pObj), Abc_ObjChild0Data(pObj) );
            pOutput = Abc_AigOr( (Abc_Aig_t *)pMiter.pManFunc, pOutput, pNext );
        }
    }
    // add the PO node and name
    pObjNew = Abc_NtkCreatePo(pMiter);
    Abc_ObjAddFanin( pObjNew, pOutput );
    Abc_ObjAssignName( pObjNew, "miter", nil );
    // make sure everything is okay
    if !Abc_NtkCheck( pMiter ) {
        fmt.Printf("Abc_NtkSensitivityMiter: The network check has failed.\n");
        Abc_NtkDelete( pMiter );
        return nil;
    }
    return pMiter;
}

/**Function*************************************************************

  Synopsis    [Computing sensitivity of POs to POs under constraints.]

  Description [The input network is a combinatonal AIG. The last output
  is a constraint. The procedure returns the list of number of PIs, 
  such that at least one PO depends on this PI, under the constraint.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSensitivity(pNtk *Abc_Ntk_t, nConfLim int32, fVerbose int32) *Vec_Int_t {
    var pProgress *ProgressBar;
    var Params Prove_Params_t;
    var pParams *Prove_Params_t = &Params;
    var vResult *Vec_Int_t  = nil;
    var pMiter *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var RetValue, i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    assert(Abc_NtkLatchNum(pNtk) == 0, "Abc_NtkLatchNum(pNtk) == 0");
    // set up solving parameters
    Prove_ParamsSetDefault( pParams );
    pParams.nItersMax = 3;
    pParams.nMiteringLimitLast = nConfLim;
    // iterate through the PIs
    vResult = Vec_IntAlloc( 100 );
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkCiNum(pNtk) );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        // generate the sensitivity miter
        pMiter = Abc_NtkSensitivityMiter( pNtk, i );
        // solve the miter using CEC engine
        RetValue = Abc_NtkIvyProve( &pMiter, pParams );
        if RetValue == -1 { // undecided
            Vec_IntPush( vResult, i );
        } else if RetValue == 0 {
            var pSimInfo *int32  = Abc_NtkVerifySimulatePattern( pMiter, pMiter.pModel );
            if pSimInfo[0] != 1 {
                fmt.Printf("ERROR in Abc_NtkMiterProve(): Generated counter-example is invalid.\n"); }
//            else
//                printf( "Networks are NOT EQUIVALENT.\n" );
            // inlining of macro ABC_FREE
            ((pSimInfo) ? (free((char *) (pSimInfo)), (pSimInfo) = 0) : 0);
            Vec_IntPush( vResult, i );
        }
        Abc_NtkDelete( pMiter );
    }
    Extra_ProgressBarStop( pProgress );
    if fVerbose != 0 {
        fmt.Printf("The outputs are sensitive to %d (out of %d) inputs:\n", Vec_IntSize(vResult), Abc_NtkCiNum(pNtk));
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vResult)) && (((RetValue) = Vec_IntEntry(vResult, i)), 1); i++  {
            fmt.Printf("%d ", RetValue);
        }
        fmt.Printf("\n");
    }
    return vResult;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

