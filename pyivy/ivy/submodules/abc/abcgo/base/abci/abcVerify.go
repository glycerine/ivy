package abci

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcVerify.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Combinational and sequential verification for two networks.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcVerify.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "base/cmd/cmd.h"
//#include "proof/fraig/fraig.h"
//#include "opt/sim/sim.h"
//#include "aig/aig/aig.h"
//#include "aig/saig/saig.h"
//#include "aig/gia/gia.h"
//#include "proof/ssw/ssw.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
// static void  Abc_NtkVerifyReportError( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel );

// extern void  Abc_NtkVerifyReportErrorSeq( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel, int nFrames );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Verifies combinational equivalence by brute-force SAT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCecSat(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nConfLimit int32, nInsLimit int32) {
    extern Abc_Ntk_t * Abc_NtkMulti( Abc_Ntk_t * pNtk, int nThresh, int nFaninMax, int fCnf, int fMulti, int fSimple, int fFactor );
    var pMiter *Abc_Ntk_t;
    var pCnf *Abc_Ntk_t;
    var RetValue int32;

    // get the miter of the two networks
    pMiter = Abc_NtkMiter( pNtk1, pNtk2, 1, 0, 0, 0 );
    if pMiter == nil {
        fmt.Printf("Miter computation has failed.\n");
        return;
    }
    RetValue = Abc_NtkMiterIsConstant( pMiter );
    if RetValue == 0 {
        fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.\n");
        // report the error
        pMiter.pModel = Abc_NtkVerifyGetCleanModel( pMiter, 1 );
        Abc_NtkVerifyReportError( pNtk1, pNtk2, pMiter.pModel );
        pMiter.pModel = nil;
        Abc_NtkDelete( pMiter );
        return;
    }
    if RetValue == 1 {
        Abc_NtkDelete( pMiter );
        fmt.Printf("Networks are equivalent after structural hashing.\n");
        return;
    }

    // convert the miter into a CNF
    pCnf = Abc_NtkMulti( pMiter, 0, 100, 1, 0, 0, 0 );
    Abc_NtkDelete( pMiter );
    if pCnf == nil {
        fmt.Printf("Renoding for CNF has failed.\n");
        return;
    }

    // solve the CNF using the SAT solver
    RetValue = Abc_NtkMiterSat( pCnf, (ABC_INT64_T)nConfLimit, (ABC_INT64_T)nInsLimit, 0, nil, nil );
    if RetValue == -1 {
        fmt.Printf("Networks are undecided (SAT solver timed out).\n");
    } else if RetValue == 0 {
        fmt.Printf("Networks are NOT EQUIVALENT after SAT.\n");
    } else {
        fmt.Printf("Networks are equivalent after SAT.\n"); }
    if pCnf.pModel {
        Abc_NtkVerifyReportError( pNtk1, pNtk2, pCnf.pModel ); }
    pCnf.pModel = nil;
    Abc_NtkDelete( pCnf );
}


/**Function*************************************************************

  Synopsis    [Verifies sequential equivalence by fraiging followed by SAT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCecFraig(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nSeconds int32, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var Params Prove_Params_t;
    var pParams *Prove_Params_t = &Params;
//    Fraig_Params_t Params;
//    Fraig_Man_t * pMan;
    var pMiter, pTemp *Abc_Ntk_t;
    var pExdc *Abc_Ntk_t  = nil;
    var RetValue int32;

    if pNtk1.pExdc != nil || pNtk2.pExdc != nil {
        if pNtk1.pExdc != nil && pNtk2.pExdc != nil {
            fmt.Printf("Comparing EXDC of the two networks:\n");
            Abc_NtkCecFraig( pNtk1.pExdc, pNtk2.pExdc, nSeconds, fVerbose );
            fmt.Printf("Comparing networks under EXDC of the first network.\n");
            pExdc = pNtk1.pExdc;
        
        } else if pNtk1.pExdc != nil {
            fmt.Printf("Second network has no EXDC. Comparing main networks under EXDC of the first network.\n");
            pExdc = pNtk1.pExdc;
        
        } else if pNtk2.pExdc != nil {
            fmt.Printf("First network has no EXDC. Comparing main networks under EXDC of the second network.\n");
            pExdc = pNtk2.pExdc;
        
        } else { assert(false, "0"); }
    }

    // get the miter of the two networks
    pMiter = Abc_NtkMiter( pNtk1, pNtk2, 1, 0, 0, 0 );
    if pMiter == nil {
        fmt.Printf("Miter computation has failed.\n");
        return;
    }
    // add EXDC to the miter
    if pExdc != nil {
        assert(Abc_NtkPoNum(pMiter) == 1, "Abc_NtkPoNum(pMiter) == 1");
        assert(Abc_NtkPoNum(pExdc) == 1, "Abc_NtkPoNum(pExdc) == 1");
        pMiter = Abc_NtkMiter( pTemp = pMiter, pExdc, 1, 0, 1, 0 );
        Abc_NtkDelete( pTemp );
    }
    // handle trivial case
    RetValue = Abc_NtkMiterIsConstant( pMiter );
    if RetValue == 0 {
        fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.  ");
        // report the error
        pMiter.pModel = Abc_NtkVerifyGetCleanModel( pMiter, 1 );
        Abc_NtkVerifyReportError( pNtk1, pNtk2, pMiter.pModel );
        pMiter.pModel = nil;
        Abc_NtkDelete( pMiter );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        return;
    }
    if RetValue == 1 {
        fmt.Printf("Networks are equivalent after structural hashing.  ");
        Abc_NtkDelete( pMiter );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        return;
    }
/*
    // convert the miter into a FRAIG
    Fraig_ParamsSetDefault( &Params );
    Params.fVerbose = fVerbose;
    Params.nSeconds = nSeconds;
//    Params.fFuncRed = 0;
//    Params.nPatsRand = 0;
//    Params.nPatsDyna = 0;
    pMan = (Fraig_Man_t *)Abc_NtkToFraig( pMiter, &Params, 0, 0 ); 
    Fraig_ManProveMiter( pMan );

    // analyze the result
    RetValue = Fraig_ManCheckMiter( pMan );
    // report the result
    if ( RetValue == -1 )
        printf( "Networks are undecided (SAT solver timed out on the final miter).\n" );
    else if ( RetValue == 1 )
        printf( "Networks are equivalent after fraiging.\n" );
    else if ( RetValue == 0 )
    {
        printf( "Networks are NOT EQUIVALENT after fraiging.\n" );
        Abc_NtkVerifyReportError( pNtk1, pNtk2, Fraig_ManReadModel(pMan) );
    }
    else assert( 0 );
    // delete the fraig manager
    Fraig_ManFree( pMan );
    // delete the miter
    Abc_NtkDelete( pMiter );
*/
    // solve the CNF using the SAT solver
    Prove_ParamsSetDefault( pParams );
    pParams.nItersMax = 5;
//    RetValue = Abc_NtkMiterProve( &pMiter, pParams );
//    pParams->fVerbose = 1;
    RetValue = Abc_NtkIvyProve( &pMiter, pParams );
    if RetValue == -1 {
        fmt.Printf("Networks are undecided (resource limits is reached).  ");
    } else if RetValue == 0 {
        var pSimInfo *int32  = Abc_NtkVerifySimulatePattern( pMiter, pMiter.pModel );
        if pSimInfo[0] != 1 {
            fmt.Printf("ERROR in Abc_NtkMiterProve(): Generated counter-example is invalid.\n");
        } else {
            fmt.Printf("Networks are NOT EQUIVALENT.  "); }
        pSimInfo = nil;
    
    } else {
        fmt.Printf("Networks are equivalent.  "); }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if pMiter.pModel {
        Abc_NtkVerifyReportError( pNtk1, pNtk2, pMiter.pModel ); }
    Abc_NtkDelete( pMiter );
}

/**Function*************************************************************

  Synopsis    [Verifies sequential equivalence by fraiging followed by SAT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCecFraigPart(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nSeconds int32, nPartSize int32, fVerbose int32) {
    var Params Prove_Params_t;
    var pParams *Prove_Params_t = &Params;
    var pMiter, pMiterPart *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i, RetValue, Status, nOutputs int32;

    // solve the CNF using the SAT solver
    Prove_ParamsSetDefault( pParams );
    pParams.nItersMax = 5;
    //    pParams->fVerbose = 1;

    assert(nPartSize > 0, "nPartSize > 0");

    // get the miter of the two networks
    pMiter = Abc_NtkMiter( pNtk1, pNtk2, 1, nPartSize, 0, 0 );
    if pMiter == nil {
        fmt.Printf("Miter computation has failed.\n");
        return;
    }
    RetValue = Abc_NtkMiterIsConstant( pMiter );
    if RetValue == 0 {
        fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.\n");
        // report the error
        pMiter.pModel = Abc_NtkVerifyGetCleanModel( pMiter, 1 );
        Abc_NtkVerifyReportError( pNtk1, pNtk2, pMiter.pModel );
        pMiter.pModel = nil;
        Abc_NtkDelete( pMiter );
        return;
    }
    if RetValue == 1 {
        fmt.Printf("Networks are equivalent after structural hashing.\n");
        Abc_NtkDelete( pMiter );
        return;
    }

    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "unset progressbar" );

    // solve the problem iteratively for each output of the miter
    Status = 1;
    nOutputs = 0;
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pMiter)) && (((pObj) = Abc_NtkPo(pMiter, i)), 1); i++  {
        if Abc_ObjFanin0(pObj) == Abc_AigConst1(pMiter) {
            if Abc_ObjFaninC0(pObj) { // complemented -> const 0
                RetValue = 1;
            } else {
                RetValue = 0; }
            pMiterPart = nil;
        
        } else {
            // get the cone of this output
            pMiterPart = Abc_NtkCreateCone( pMiter, Abc_ObjFanin0(pObj), Abc_ObjName(pObj), 0 );
            if Abc_ObjFaninC0(pObj) {
                Abc_ObjXorFaninC( Abc_NtkPo(pMiterPart,0), 0 ); }
            // solve the cone
        //    RetValue = Abc_NtkMiterProve( &pMiterPart, pParams );
            RetValue = Abc_NtkIvyProve( &pMiterPart, pParams );
        }

        if RetValue == -1 {
            fmt.Printf("Networks are undecided (resource limits is reached).\r");
            Status = -1;
        
        } else if RetValue == 0 {
            var pSimInfo *int32  = Abc_NtkVerifySimulatePattern( pMiterPart, pMiterPart.pModel );
            if pSimInfo[0] != 1 {
                fmt.Printf("ERROR in Abc_NtkMiterProve(): Generated counter-example is invalid.\n");
            } else {
                fmt.Printf("Networks are NOT EQUIVALENT.                 \n"); }
            pSimInfo = nil;
            Status = 0;
            break;
        
        } else {
            fmt.Printf("Finished part %5d (out of %5d)\r", i+1, Abc_NtkPoNum(pMiter));
            nOutputs += nPartSize;
        }
//        if ( pMiter->pModel )
//            Abc_NtkVerifyReportError( pNtk1, pNtk2, pMiter->pModel );
        if pMiterPart != nil {
            Abc_NtkDelete( pMiterPart ); }
    }
  
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "set progressbar" );

    if Status == 1 {
        fmt.Printf("Networks are equivalent.                         \n");
    } else if Status == -1 {
        fmt.Printf("Timed out after verifying %d outputs (out of %d).\n", nOutputs, Abc_NtkCoNum(pNtk1)); }
    Abc_NtkDelete( pMiter );
}

/**Function*************************************************************

  Synopsis    [Verifies sequential equivalence by fraiging followed by SAT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCecFraigPartAuto(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nSeconds int32, fVerbose int32) {
    extern Vec_Ptr_t * Abc_NtkPartitionSmart( Abc_Ntk_t * pNtk, int nPartSizeLimit, int fVerbose );
    extern void Abc_NtkConvertCos( Abc_Ntk_t * pNtk, Vec_Int_t * vOuts, Vec_Ptr_t * vOnePtr );

    var vParts, vOnePtr *Vec_Ptr_t;
    var vOne *Vec_Int_t;
    var Params Prove_Params_t;
    var pParams *Prove_Params_t = &Params;
    var pMiter, pMiterPart *Abc_Ntk_t;
    var i, RetValue, Status, nOutputs int32;

    // solve the CNF using the SAT solver
    Prove_ParamsSetDefault( pParams );
    pParams.nItersMax = 5;
    //    pParams->fVerbose = 1;

    // get the miter of the two networks
    pMiter = Abc_NtkMiter( pNtk1, pNtk2, 1, 1, 0, 0 );
    if pMiter == nil {
        fmt.Printf("Miter computation has failed.\n");
        return;
    }
    RetValue = Abc_NtkMiterIsConstant( pMiter );
    if RetValue == 0 {
        fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.\n");
        // report the error
        pMiter.pModel = Abc_NtkVerifyGetCleanModel( pMiter, 1 );
        Abc_NtkVerifyReportError( pNtk1, pNtk2, pMiter.pModel );
        pMiter.pModel = nil;
        Abc_NtkDelete( pMiter );
        return;
    }
    if RetValue == 1 {
        fmt.Printf("Networks are equivalent after structural hashing.\n");
        Abc_NtkDelete( pMiter );
        return;
    }

    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "unset progressbar" );

    // partition the outputs
    vParts = Abc_NtkPartitionSmart( pMiter, 300, 0 );

    // fraig each partition
    Status = 1;
    nOutputs = 0;
    vOnePtr = Vec_PtrAlloc( 1000 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        // get this part of the miter
        Abc_NtkConvertCos( pMiter, vOne, vOnePtr );
        pMiterPart = Abc_NtkCreateConeArray( pMiter, vOnePtr, 0 );
        Abc_NtkCombinePos( pMiterPart, 0, 0 );
        // check the miter for being constant
        RetValue = Abc_NtkMiterIsConstant( pMiterPart );
        if RetValue == 0 {
            fmt.Printf("Networks are NOT EQUIVALENT after partitioning.\n");
            Abc_NtkDelete( pMiterPart );
            break;
        }
        if RetValue == 1 {
            Abc_NtkDelete( pMiterPart );
            continue;
        }
        fmt.Printf("Verifying part %4d  (out of %4d)  PI = %5d. PO = %5d. And = %6d. Lev = %4d.\r", i+1, Vec_PtrSize(vParts), Abc_NtkPiNum(pMiterPart), Abc_NtkPoNum(pMiterPart), Abc_NtkNodeNum(pMiterPart), Abc_AigLevel(pMiterPart));
        fflush( os.Stdout );
        // solve the problem
        RetValue = Abc_NtkIvyProve( &pMiterPart, pParams );
        if RetValue == -1 {
            fmt.Printf("Networks are undecided (resource limits is reached).\r");
            Status = -1;
        
        } else if RetValue == 0 {
            var pSimInfo *int32  = Abc_NtkVerifySimulatePattern( pMiterPart, pMiterPart.pModel );
            if pSimInfo[0] != 1 {
                fmt.Printf("ERROR in Abc_NtkMiterProve(): Generated counter-example is invalid.\n");
            } else {
                fmt.Printf("Networks are NOT EQUIVALENT.                 \n"); }
            pSimInfo = nil;
            Status = 0;
            Abc_NtkDelete( pMiterPart );
            break;
        
        } else {
//            printf( "Finished part %5d (out of %5d)\r", i+1, Vec_PtrSize(vParts) );
            nOutputs += Vec_IntSize(vOne);
        }
        Abc_NtkDelete( pMiterPart );
    }
    fmt.Printf("                                                                                          \r");
    Vec_VecFree( (Vec_Vec_t *)vParts );
    Vec_PtrFree( vOnePtr );

    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "set progressbar" );

    if Status == 1 {
        fmt.Printf("Networks are equivalent.                         \n");
    } else if Status == -1 {
        fmt.Printf("Timed out after verifying %d outputs (out of %d).\n", nOutputs, Abc_NtkCoNum(pNtk1)); }
    Abc_NtkDelete( pMiter );
}

/**Function*************************************************************

  Synopsis    [Verifies sequential equivalence by brute-force SAT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSecSat(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nConfLimit int32, nInsLimit int32, nFrames int32) {
    extern Abc_Ntk_t * Abc_NtkMulti( Abc_Ntk_t * pNtk, int nThresh, int nFaninMax, int fCnf, int fMulti, int fSimple, int fFactor );
    var pMiter *Abc_Ntk_t;
    var pFrames *Abc_Ntk_t;
    var pCnf *Abc_Ntk_t;
    var RetValue int32;

    // get the miter of the two networks
    pMiter = Abc_NtkMiter( pNtk1, pNtk2, 0, 0, 0, 0 );
    if pMiter == nil {
        fmt.Printf("Miter computation has failed.\n");
        return;
    }
    RetValue = Abc_NtkMiterIsConstant( pMiter );
    if RetValue == 0 {
        Abc_NtkDelete( pMiter );
        fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.\n");
        return;
    }
    if RetValue == 1 {
        Abc_NtkDelete( pMiter );
        fmt.Printf("Networks are equivalent after structural hashing.\n");
        return;
    }

    // create the timeframes
    pFrames = Abc_NtkFrames( pMiter, nFrames, 1, 0 );
    Abc_NtkDelete( pMiter );
    if pFrames == nil {
        fmt.Printf("Frames computation has failed.\n");
        return;
    }
    RetValue = Abc_NtkMiterIsConstant( pFrames );
    if RetValue == 0 {
        Abc_NtkDelete( pFrames );
        fmt.Printf("Networks are NOT EQUIVALENT after framing.\n");
        return;
    }
    if RetValue == 1 {
        Abc_NtkDelete( pFrames );
        fmt.Printf("Networks are equivalent after framing.\n");
        return;
    }

    // convert the miter into a CNF
    pCnf = Abc_NtkMulti( pFrames, 0, 100, 1, 0, 0, 0 );
    Abc_NtkDelete( pFrames );
    if pCnf == nil {
        fmt.Printf("Renoding for CNF has failed.\n");
        return;
    }

    // solve the CNF using the SAT solver
    RetValue = Abc_NtkMiterSat( pCnf, (ABC_INT64_T)nConfLimit, (ABC_INT64_T)nInsLimit, 0, nil, nil );
    if RetValue == -1 {
        fmt.Printf("Networks are undecided (SAT solver timed out).\n");
    } else if RetValue == 0 {
        fmt.Printf("Networks are NOT EQUIVALENT after SAT.\n");
    } else {
        fmt.Printf("Networks are equivalent after SAT.\n"); }
    Abc_NtkDelete( pCnf );
}

/**Function*************************************************************

  Synopsis    [Verifies combinational equivalence by fraiging followed by SAT]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSecFraig(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nSeconds int32, nFrames int32, fVerbose int32) int32 {
    var Params Fraig_Params_t;
    var pMan *Fraig_Man_t;
    var pMiter *Abc_Ntk_t;
    var pFrames *Abc_Ntk_t;
    var RetValue int32;

    // get the miter of the two networks
    pMiter = Abc_NtkMiter( pNtk1, pNtk2, 0, 0, 0, 0 );
    if pMiter == nil {
        fmt.Printf("Miter computation has failed.\n");
        return 0;
    }
    RetValue = Abc_NtkMiterIsConstant( pMiter );
    if RetValue == 0 {
        fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.\n");
        // report the error
        pMiter.pModel = Abc_NtkVerifyGetCleanModel( pMiter, nFrames );
        Abc_NtkVerifyReportErrorSeq( pNtk1, pNtk2, pMiter.pModel, nFrames );
        pMiter.pModel = nil;
        Abc_NtkDelete( pMiter );
        return 0;
    }
    if RetValue == 1 {
        Abc_NtkDelete( pMiter );
        fmt.Printf("Networks are equivalent after structural hashing.\n");
        return 1;
    }

    // create the timeframes
    pFrames = Abc_NtkFrames( pMiter, nFrames, 1, 0 );
    Abc_NtkDelete( pMiter );
    if pFrames == nil {
        fmt.Printf("Frames computation has failed.\n");
        return 0;
    }
    RetValue = Abc_NtkMiterIsConstant( pFrames );
    if RetValue == 0 {
        fmt.Printf("Networks are NOT EQUIVALENT after framing.\n");
        // report the error
        pFrames.pModel = Abc_NtkVerifyGetCleanModel( pFrames, 1 );
//        Abc_NtkVerifyReportErrorSeq( pNtk1, pNtk2, pFrames->pModel, nFrames );
        pFrames.pModel = nil;
        Abc_NtkDelete( pFrames );
        return 0;
    }
    if RetValue == 1 {
        Abc_NtkDelete( pFrames );
        fmt.Printf("Networks are equivalent after framing.\n");
        return 1;
    }

    // convert the miter into a FRAIG
    Fraig_ParamsSetDefault( &Params );
    Params.fVerbose = fVerbose;
    Params.nSeconds = nSeconds;
//    Params.fFuncRed = 0;
//    Params.nPatsRand = 0;
//    Params.nPatsDyna = 0;
    pMan = (Fraig_Man_t *)Abc_NtkToFraig( pFrames, &Params, 0, 0 ); 
    Fraig_ManProveMiter( pMan );

    // analyze the result
    RetValue = Fraig_ManCheckMiter( pMan );
    // report the result
    if RetValue == -1 {
        fmt.Printf("Networks are undecided (SAT solver timed out on the final miter).\n");
    } else if RetValue == 1 {
        fmt.Printf("Networks are equivalent after fraiging.\n");
    } else if RetValue == 0 {
        fmt.Printf("Networks are NOT EQUIVALENT after fraiging.\n");
//        Abc_NtkVerifyReportErrorSeq( pNtk1, pNtk2, Fraig_ManReadModel(pMan), nFrames );
    
    } else { assert(false, "0"); }
    // delete the fraig manager
    Fraig_ManFree( pMan );
    // delete the miter
    Abc_NtkDelete( pFrames );
    return RetValue == 1;
}

/**Function*************************************************************

  Synopsis    [Returns a dummy pattern full of zeros.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkVerifyGetCleanModel(pNtk *Abc_Ntk_t, nFrames int32) *int32 {
    // inlining of macro ABC_ALLOC
    var pModel *int32  = ((int *) malloc(sizeof(int) * (size_t)(Abc_NtkCiNum(pNtk) * nFrames)));
    memset( pModel, 0, sizeof(int) * Abc_NtkCiNum(pNtk) * nFrames );
    return pModel;
}

/**Function*************************************************************

  Synopsis    [Returns the PO values under the given input pattern.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Abc_NtkVerifySimulatePattern(pNtk *Abc_Ntk_t, pModel *int32) *int32 {
    var pNode *Abc_Obj_t;
    var pValues *int32;
    var Value0 int32;
    var Value1 int32;
    var i int32;
    var fStrashed int32;
    if !Abc_NtkIsStrash(pNtk) {
        pNtk = Abc_NtkStrash(pNtk, 0, 0, 0);
        fStrashed = 1;
    }
/*
    printf( "Counter example: " );
    Abc_NtkForEachCi( pNtk, pNode, i )
        printf( " %d", pModel[i] );
    printf( "\n" );
*/
    // increment the trav ID
    Abc_NtkIncrementTravId( pNtk );
    // set the CI values
    Abc_AigConst1(pNtk).pCopy = (Abc_Obj_t *)1;
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pNode.pCopy = (Abc_Obj_t *)(ABC_PTRINT_T)pModel[i];
    }
    // simulate in the topological order
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Value0 = (int32((ABC_PTRINT_T))Abc_ObjFanin0(pNode).pCopy) ^ int32(Abc_ObjFaninC0(pNode));
        Value1 = (int32((ABC_PTRINT_T))Abc_ObjFanin1(pNode).pCopy) ^ int32(Abc_ObjFaninC1(pNode));
        pNode.pCopy = (Abc_Obj_t *)(ABC_PTRINT_T)(Value0 & Value1);
    }
    }
    // fill the output values
    // inlining of macro ABC_ALLOC
    pValues = ((int *) malloc(sizeof(int) * (size_t)(Abc_NtkCoNum(pNtk))));
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pValues[i] = (int32((ABC_PTRINT_T))Abc_ObjFanin0(pNode).pCopy) ^ int32(Abc_ObjFaninC0(pNode));
    }
    if fStrashed != 0 {
        Abc_NtkDelete( pNtk ); }
    return pValues;
}


/**Function*************************************************************

  Synopsis    [Reports mismatch between the two networks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkVerifyReportError(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, pModel *int32) {
    var vNodes *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var pValues1, pValues2 *int32;
    var nErrors int32;
    var nPrinted int32;
    var i int32;
    var iNode int32 = -1;

    assert(Abc_NtkCiNum(pNtk1) == Abc_NtkCiNum(pNtk2), "Abc_NtkCiNum(pNtk1) == Abc_NtkCiNum(pNtk2)");
    assert(Abc_NtkCoNum(pNtk1) == Abc_NtkCoNum(pNtk2), "Abc_NtkCoNum(pNtk1) == Abc_NtkCoNum(pNtk2)");
    // get the CO values under this model
    pValues1 = Abc_NtkVerifySimulatePattern( pNtk1, pModel );
    pValues2 = Abc_NtkVerifySimulatePattern( pNtk2, pModel );
    // count the mismatches
    nErrors = 0;
    for  i = 0; i < Abc_NtkCoNum(pNtk1); i++  {
        nErrors += int32(( pValues1[i] != pValues2[i] ));
    }
    fmt.Printf("Verification failed for at least %d outputs: ", nErrors);
    // print the first 3 outputs
    nPrinted = 0;
    for  i = 0; i < Abc_NtkCoNum(pNtk1); i++  {
        if pValues1[i] != pValues2[i] {
            if iNode == -1 {
                iNode = i; }
            fmt.Printf(" %s", Abc_ObjName(Abc_NtkCo(pNtk1,i)));
            if nPrinted++; nPrinted == 3 {
                break; }
        }
    }
    if nPrinted != nErrors {
        fmt.Printf(" ..."); }
    fmt.Printf("\n");
    // report mismatch for the first output
    if iNode >= 0 {
        fmt.Printf("Output %s: Value in Network1 = %d. Value in Network2 = %d.\n", Abc_ObjName(Abc_NtkCo(pNtk1,iNode)), pValues1[iNode], pValues2[iNode]);
        fmt.Printf("Input pattern: ");
        // collect PIs in the cone
        pNode = Abc_NtkCo(pNtk1,iNode);
        vNodes = Abc_NtkNodeSupport( pNtk1, &pNode, 1 );
        // set the PI numbers
        // inlining of macro Abc_NtkForEachCi
        for  i = 0; (i < Abc_NtkCiNum(pNtk1)) && (((pNode) = Abc_NtkCi(pNtk1, i)), 1); i++  {
            pNode.pCopy = (Abc_Obj_t *)(ABC_PTRINT_T)i;
        }
        // print the model
        pNode = (Abc_Obj_t *)Vec_PtrEntry( vNodes, 0 );
        if Abc_ObjIsCi(pNode) {
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
                assert(Abc_ObjIsCi(pNode), "Abc_ObjIsCi(pNode)");
                fmt.Printf(" %s=%d", Abc_ObjName(pNode), pModel[int32((ABC_PTRINT_T))pNode.pCopy]);
            }
        }
        fmt.Printf("\n");
        Vec_PtrFree( vNodes );
    }
    pValues1 = nil;
    pValues2 = nil;
}


/**Function*************************************************************

  Synopsis    [Computes the COs in the support of the PO in the given frame.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkGetSeqPoSupp(pNtk *Abc_Ntk_t, iFrame int32, iNumPo int32) {
    var pFrames *Abc_Ntk_t;
    var pObj, pNodePo *Abc_Obj_t;
    var vSupp *Vec_Ptr_t;
    var i, k int32;
    // get the timeframes of the network
    pFrames = Abc_NtkFrames( pNtk, iFrame + 1, 0, 0 );
//Abc_NtkShowAig( pFrames );

    // get the PO of the timeframes
    pNodePo = Abc_NtkPo( pFrames, iFrame * Abc_NtkPoNum(pNtk) + iNumPo );
    // set the support
    vSupp   = Abc_NtkNodeSupport( pFrames, &pNodePo, 1 );
    // mark the support of the frames
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pFrames)) && (((pObj) = Abc_NtkCi(pFrames, i)), 1); i++  {
        pObj.pCopy = nil;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupp)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vSupp, i)), 1); i++  {
        pObj.pCopy = (Abc_Obj_t *)1;
    }
    // mark the support of the network if the support of the timeframes is marked
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.pCopy = nil;
    }
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else
        if Abc_NtkBox(pFrames, i).pCopy {
            pObj.pCopy = (Abc_Obj_t *)1; }
    }
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        for  k = 0; k <= iFrame; k++  {
            if Abc_NtkPi(pFrames, k*Abc_NtkPiNum(pNtk) + i).pCopy {
                pObj.pCopy = (Abc_Obj_t *)1; }
        }
    }
    // free stuff
    Vec_PtrFree( vSupp );
    Abc_NtkDelete( pFrames );
}

/**Function*************************************************************

  Synopsis    [Reports mismatch between the two sequential networks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkVerifyReportErrorSeq(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, pModel *int32, nFrames int32) {
    var vInfo1, vInfo2 *Vec_Ptr_t;
    var pObj, pObjError, pObj1, pObj2 *Abc_Obj_t;
    var ValueError1 int32 = -1;
    var ValueError2 int32 = -1;
    var pPats1, pPats2 *uint32;
    var i int32;
    var o int32;
    var k int32;
    var nErrors int32;
    var iFrameError int32 = -1;
    var iNodePo int32 = -1;
    var nPrinted int32;
    var fRemove1, fRemove2 int32;

    if !Abc_NtkIsStrash(pNtk1) {
        fRemove1 = 1;
pNtk1 = Abc_NtkStrash( pNtk1, 0, 0, 0 ); }
    if !Abc_NtkIsStrash(pNtk2) {
        fRemove2 = 1;
pNtk2 = Abc_NtkStrash( pNtk2, 0, 0, 0 ); }

    // simulate sequential circuits
    vInfo1 = Sim_SimulateSeqModel( pNtk1, nFrames, pModel );
    vInfo2 = Sim_SimulateSeqModel( pNtk2, nFrames, pModel );

    // look for a discrepancy in the PO values
    nErrors = 0;
    pObjError = nil;
    for  i = 0; i < nFrames; i++  {
        if pObjError {
            break; }
        // inlining of macro Abc_NtkForEachPo
        for  o = 0; (o < Abc_NtkPoNum(pNtk1)) && (((pObj1) = Abc_NtkPo(pNtk1, o)), 1); o++  {
            pObj2  = Abc_NtkPo( pNtk2, o );
            // inlining of macro Sim_SimInfoGet
            pPats1 = ((unsigned *)((vInfo1).pArray[(pObj1).Id]));
            // inlining of macro Sim_SimInfoGet
            pPats2 = ((unsigned *)((vInfo2).pArray[(pObj2).Id]));
            if pPats1[i] == pPats2[i] {
                continue; }
            nErrors++;
            if pObjError == nil {
                pObjError   = pObj1;
                iFrameError = i;
                iNodePo     = o;
                ValueError1 = (pPats1[i] > 0);
                ValueError2 = (pPats2[i] > 0);
            }
        }
    }

    if pObjError == nil {
        fmt.Printf("No output mismatches detected.\n");
        Sim_UtilInfoFree( vInfo1 );
        Sim_UtilInfoFree( vInfo2 );
        if fRemove1 { Abc_NtkDelete( pNtk1 ); }
        if fRemove2 != 0 { Abc_NtkDelete( pNtk2 ); }
        return;
    }

    fmt.Printf("Verification failed for at least %d output%s of frame %d: ", nErrors, (nErrors>1? "s":""), iFrameError+1);
    // print the first 3 outputs
    nPrinted = 0;
    // inlining of macro Abc_NtkForEachPo
    for  o = 0; (o < Abc_NtkPoNum(pNtk1)) && (((pObj1) = Abc_NtkPo(pNtk1, o)), 1); o++  {
        pObj2 = Abc_NtkPo( pNtk2, o );
        // inlining of macro Sim_SimInfoGet
        pPats1 = ((unsigned *)((vInfo1).pArray[(pObj1).Id]));
        // inlining of macro Sim_SimInfoGet
        pPats2 = ((unsigned *)((vInfo2).pArray[(pObj2).Id]));
        if pPats1[iFrameError] == pPats2[iFrameError] {
            continue; }
        fmt.Printf(" %s", Abc_ObjName(pObj1));
        if nPrinted++; nPrinted == 3 {
            break; }
    }
    if nPrinted != nErrors {
        fmt.Printf(" ..."); }
    fmt.Printf("\n");

    // mark CIs of the networks in the cone of influence of this output
    Abc_NtkGetSeqPoSupp( pNtk1, iFrameError, iNodePo );
    Abc_NtkGetSeqPoSupp( pNtk2, iFrameError, iNodePo );

    // report mismatch for the first output
    fmt.Printf("Output %s: Value in Network1 = %d. Value in Network2 = %d.\n", Abc_ObjName(pObjError), ValueError1, ValueError2);

    fmt.Printf("The cone of influence of output %s in Network1:\n", Abc_ObjName(pObjError));
    fmt.Printf("PIs: ");
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk1)) && (((pObj) = Abc_NtkPi(pNtk1, i)), 1); i++  {
        if pObj.pCopy {
            fmt.Printf("%s ", Abc_ObjName(pObj)); }
    }
    fmt.Printf("\n");
    fmt.Printf("Latches: ");
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk1).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk1, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else
        if pObj.pCopy {
            fmt.Printf("%s ", Abc_ObjName(pObj)); }
    }
    fmt.Printf("\n");

    fmt.Printf("The cone of influence of output %s in Network2:\n", Abc_ObjName(pObjError));
    fmt.Printf("PIs: ");
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk2)) && (((pObj) = Abc_NtkPi(pNtk2, i)), 1); i++  {
        if pObj.pCopy {
            fmt.Printf("%s ", Abc_ObjName(pObj)); }
    }
    fmt.Printf("\n");
    fmt.Printf("Latches: ");
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk2).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk2, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else
        if pObj.pCopy {
            fmt.Printf("%s ", Abc_ObjName(pObj)); }
    }
    fmt.Printf("\n");

    // print the patterns
    for  i = 0; i <= iFrameError; i++  {
        fmt.Printf("Frame %d:  ", i+1);

        fmt.Printf("PI(1):");
        // inlining of macro Abc_NtkForEachPi
        for  k = 0; (k < Abc_NtkPiNum(pNtk1)) && (((pObj) = Abc_NtkPi(pNtk1, k)), 1); k++  {
            if pObj.pCopy {
                // inlining of macro Sim_SimInfoGet
                fmt.Printf("%d", ((unsigned *)((vInfo1).pArray[(pObj).Id]))[i] > 0); }
        }
        fmt.Printf(" ");
        fmt.Printf("L(1):");
        // inlining of macro Abc_NtkForEachLatch
        for  k = 0; (k < Vec_PtrSize((pNtk1).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk1, k)), 1); k++  {           if !Abc_ObjIsLatch(pObj) {} else
            if pObj.pCopy {
                // inlining of macro Sim_SimInfoGet
                fmt.Printf("%d", ((unsigned *)((vInfo1).pArray[(pObj).Id]))[i] > 0); }
        }
        fmt.Printf(" ");
        fmt.Printf("%s(1):", Abc_ObjName(pObjError));
        // inlining of macro Sim_SimInfoGet
        fmt.Printf("%d", ((unsigned *)((vInfo1).pArray[(pObjError).Id]))[i] > 0);

        fmt.Printf("  ");

        fmt.Printf("PI(2):");
        // inlining of macro Abc_NtkForEachPi
        for  k = 0; (k < Abc_NtkPiNum(pNtk2)) && (((pObj) = Abc_NtkPi(pNtk2, k)), 1); k++  {
            if pObj.pCopy {
                // inlining of macro Sim_SimInfoGet
                fmt.Printf("%d", ((unsigned *)((vInfo2).pArray[(pObj).Id]))[i] > 0); }
        }
        fmt.Printf(" ");
        fmt.Printf("L(2):");
        // inlining of macro Abc_NtkForEachLatch
        for  k = 0; (k < Vec_PtrSize((pNtk2).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk2, k)), 1); k++  {           if !Abc_ObjIsLatch(pObj) {} else
            if pObj.pCopy {
                // inlining of macro Sim_SimInfoGet
                fmt.Printf("%d", ((unsigned *)((vInfo2).pArray[(pObj).Id]))[i] > 0); }
        }
        fmt.Printf(" ");
        fmt.Printf("%s(2):", Abc_ObjName(pObjError));
        // inlining of macro Sim_SimInfoGet
        fmt.Printf("%d", ((unsigned *)((vInfo2).pArray[(pObjError).Id]))[i] > 0);

        fmt.Printf("\n");
    }
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk1)) && (((pObj) = Abc_NtkCi(pNtk1, i)), 1); i++  {
        pObj.pCopy = nil;
    }
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk2)) && (((pObj) = Abc_NtkCi(pNtk2, i)), 1); i++  {
        pObj.pCopy = nil;
    }

    Sim_UtilInfoFree( vInfo1 );
    Sim_UtilInfoFree( vInfo2 );
    if fRemove1 { Abc_NtkDelete( pNtk1 ); }
    if fRemove2 != 0 { Abc_NtkDelete( pNtk2 ); }
}

/**Function*************************************************************

  Synopsis    [Simulates buggy miter emailed by Mike.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSimulteBuggyMiter(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    var pModel1, pModel2, pResult1, pResult2 *int32;
    var vPiValues1 string  = "01001011100000000011010110101000000";
    var vPiValues2 string  = "11001101011101011111110100100010001";

    assert(len(vPiValues1) == uint32(Abc_NtkPiNum(pNtk)), "strlen(vPiValues1) == uint32(Abc_NtkPiNum(pNtk))");
    assert(1 == Abc_NtkPoNum(pNtk), "1 == Abc_NtkPoNum(pNtk)");

    // inlining of macro ABC_ALLOC
    pModel1 = ((int *) malloc(sizeof(int) * (size_t)(Abc_NtkCiNum(pNtk))));
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pModel1[i] = vPiValues1[i] - '0';
    }
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        pModel1[Abc_NtkPiNum(pNtk)+i] = (int32((ABC_PTRINT_T))pObj.pData) - 1; }
    }

    pResult1 = Abc_NtkVerifySimulatePattern( pNtk, pModel1 );
    fmt.Printf("Value = %d\n", pResult1[0]);

    // inlining of macro ABC_ALLOC
    pModel2 = ((int *) malloc(sizeof(int) * (size_t)(Abc_NtkCiNum(pNtk))));
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pModel2[i] = vPiValues2[i] - '0';
    }
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        pModel2[Abc_NtkPiNum(pNtk)+i] = pResult1[Abc_NtkPoNum(pNtk)+i]; }
    }

    pResult2 = Abc_NtkVerifySimulatePattern( pNtk, pModel2 );
    fmt.Printf("Value = %d\n", pResult2[0]);

    pModel1 = nil;
    pModel2 = nil;
    pResult1 = nil;
    pResult2 = nil;
}


/**Function*************************************************************

  Synopsis    [Returns the PO values under the given input pattern.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Abc_NtkIsTrueCex(pNtk *Abc_Ntk_t, pCex *Abc_Cex_t) int32 {
    extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
    var pMan *Aig_Man_t;
    var status, fStrashed int32;
    if !Abc_NtkIsStrash(pNtk) {
        pNtk = Abc_NtkStrash(pNtk, 0, 0, 0);
        fStrashed = 1;
    }
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan {
        status = Saig_ManVerifyCex( pMan, pCex );
        Aig_ManStop( pMan );
    }
    if fStrashed != 0 {
        Abc_NtkDelete( pNtk ); }
    return status;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the number of PIs matches.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Abc_NtkIsValidCex(pNtk *Abc_Ntk_t, pCex *Abc_Cex_t) int32 {
    return Abc_NtkPiNum(pNtk) == pCex.nPis;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

