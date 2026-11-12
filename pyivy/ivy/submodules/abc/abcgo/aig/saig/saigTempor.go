package saig

import "fmt"

/**CFile****************************************************************

  FileName    [saigTempor.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Temporal decomposition.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigTempor.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates initialized timeframes for temporal decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManTemporFrames(pAig *Aig_Man_t, nFrames int32) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i, f int32;
    // start the frames package
    Aig_ManCleanData( pAig );
    pFrames = Aig_ManStart( Aig_ManObjNumMax(pAig) * nFrames );
    pFrames.pName = Abc_UtilStrsav( pAig.pName );
    // initiliaze the flops
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.pData = Aig_ManConst0(pFrames);
    }
    // for each timeframe
    for  f = 0; f < nFrames; f++  {
        Aig_ManConst1(pAig).pData = Aig_ManConst1(pFrames);
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            pObj.pData = Aig_ObjCreateCi(pFrames);
        }
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.pData = Aig_And( pFrames, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            pObj.pData = Aig_ObjChild0Copy(pObj);
        }
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
            pObjLo.pData = pObjLi.pData;
        }
    }
    // create POs for the flop inputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        Aig_ObjCreateCo( pFrames, (Aig_Obj_t *)pObj.pData );
    }
    Aig_ManCleanup( pFrames );
    return pFrames;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManTemporDecompose(pAig *Aig_Man_t, nFrames int32) *Aig_Man_t {
    var pAigNew, pFrames *Aig_Man_t;
    var pObj, pReset *Aig_Obj_t;
    var i int32;
    if pAig.nConstrs > 0 {
        fmt.Printf("The AIG manager should have no constraints.\n");
        return nil;
    }
    // create initialized timeframes
    pFrames = Saig_ManTemporFrames( pAig, nFrames );
    assert(Aig_ManCoNum(pFrames) == Aig_ManRegNum(pAig), "Aig_ManCoNum(pFrames) == Aig_ManRegNum(pAig)");

    // start the new manager
    Aig_ManCleanData( pAig );
    pAigNew = Aig_ManStart( Aig_ManNodeNum(pAig) );
    pAigNew.pName = Abc_UtilStrsav( pAig.pName );
    // map the constant node and primary inputs
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pAigNew );
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pAigNew );
    }

    // insert initialization logic
    Aig_ManConst1(pFrames).pData = Aig_ManConst1( pAigNew );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pFrames.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pFrames.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pAigNew );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pFrames.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pFrames.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pAigNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pFrames.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pFrames.vCos, i)), 1); i++  {
        pObj.pData = Aig_ObjChild0Copy(pObj);
    }

    // create reset latch (the first one among the latches)
    pReset = Aig_ObjCreateCi( pAigNew );

    // create flop output values
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.pData = Aig_Mux( pAigNew, pReset, Aig_ObjCreateCi(pAigNew), (Aig_Obj_t *)Aig_ManCo(pFrames, i).pData );
    }
    Aig_ManStop( pFrames );

    // add internal nodes of this frame
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pAigNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // create primary outputs
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        Aig_ObjCreateCo( pAigNew, Aig_ObjChild0Copy(pObj) );
    }

    // create reset latch (the first one among the latches)
    Aig_ObjCreateCo( pAigNew, Aig_ManConst1(pAigNew) );
    // create latch inputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        Aig_ObjCreateCo( pAigNew, Aig_ObjChild0Copy(pObj) );
    }

    // finalize
    Aig_ManCleanup( pAigNew );
    Aig_ManSetRegNum( pAigNew, Aig_ManRegNum(pAig)+1 ); // + reset latch (011111...)
    return pAigNew;
}

/**Function*************************************************************

  Synopsis    [Find index of first non-zero entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntLastNonZeroBeforeLimit(vTemp *Vec_Int_t, Limit int32) int32 {
    var Entry, i int32;
    if vTemp == nil {
        return -1; }
    // inlining of macro Vec_IntForEachEntryReverse
    for  i = Vec_IntSize(vTemp) - 1; (i >= 0) && (((Entry) = Vec_IntEntry(vTemp, i)), 1); i--  {
        if i >= Limit {
            continue; }
        if Entry != 0 {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManTempor(pAig *Aig_Man_t, nFrames int32, TimeOut int32, nConfLimit int32, fUseBmc int32, fUseTransSigs int32, fVerbose int32, fVeryVerbose int32) *Aig_Man_t { 
    extern int Saig_ManPhasePrefixLength( Aig_Man_t * p, int fVerbose, int fVeryVerbose, Vec_Int_t ** pvTrans );

    var vTransSigs *Vec_Int_t  = nil;
    var RetValue int32;
    var nFramesFinished int32 = -1;
    assert(nFrames >= 0, "nFrames >= 0"); 
    if nFrames == 0 {
        nFrames = Saig_ManPhasePrefixLength( pAig, fVerbose, fVeryVerbose, &vTransSigs );
        if nFrames == 0 {
            Vec_IntFreeP( &vTransSigs );
            fmt.Printf("The leading sequence has length 0. Temporal decomposition is not performed.\n");
            return nil;
        }
        if nFrames == 1 {
            Vec_IntFreeP( &vTransSigs );
            fmt.Printf("The leading sequence has length 1. Temporal decomposition is not performed.\n");
            return nil;
        }
        if fUseTransSigs != 0 {
            var Entry int32;
            var i int32;
            var iLast int32 = -1;
            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vTransSigs)) && (((Entry) = Vec_IntEntry(vTransSigs, i)), 1); i++  {
                if Entry != 0 {
                    iLast = i;
                } else {
                    iLast = iLast;
                }
            }
            if iLast > 0 && iLast < nFrames {
                Abc_Print( 1, "Reducing frame count from %d to %d to fit the last transient.\n", nFrames, iLast );
                nFrames = iLast;
            }
        }
        Abc_Print( 1, "Using computed frame number (%d).\n", nFrames );
    
    } else {
        Abc_Print( 1, "Using user-given frame number (%d).\n", nFrames ); }
    // run BMC2
    if fUseBmc != 0 {
        RetValue = Saig_BmcPerform( pAig, 0, nFrames, 2000, TimeOut, nConfLimit, 0, fVerbose, 0, &nFramesFinished, 0, 0 );
        if RetValue == 0 {
            Vec_IntFreeP( &vTransSigs );
            fmt.Printf("A cex found in the first %d frames.\n", nFrames);
            return nil;
        }
        if nFramesFinished + 1 < nFrames {
            var iLastBefore int32  = Vec_IntLastNonZeroBeforeLimit( vTransSigs, nFramesFinished );
            if iLastBefore < 1 || !(fUseTransSigs != 0) {
                Vec_IntFreeP( &vTransSigs );
                fmt.Printf("BMC for %d frames could not be completed. A cex may exist!\n", nFrames);
                return nil;
            }
            assert(iLastBefore < nFramesFinished, "iLastBefore < nFramesFinished");
            fmt.Printf("BMC succeeded to frame %d. Adjusting frame count to be (%d) based on the last transient signal.\n", nFramesFinished, iLastBefore);
            nFrames = iLastBefore;
        }
    }
    Vec_IntFreeP( &vTransSigs );
    return Saig_ManTemporDecompose( pAig, nFrames );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

