package ssw

import "time"

/**CFile****************************************************************

  FileName    [sswPairs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Calls to the SAT solver.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswPairs.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reports the status of the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_MiterStatus(p *Aig_Man_t, fVerbose int32) int32 {
    var pObj, pChild *Aig_Obj_t;
    var i, CountConst0, CountNonConst0, CountUndecided int32;
//    if ( p->pData )
//        return 0;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pChild = Aig_ObjChild0(pObj);
        // check if the output is constant 0
        if pChild == Aig_ManConst0(p) {
            CountConst0++;
            continue;
        }
        // check if the output is constant 1
        if pChild == Aig_ManConst1(p) {
            CountNonConst0++;
            continue;
        }
        // check if the output is a primary input
        if p.nRegs == 0 && Aig_ObjIsCi(Aig_Regular(pChild)) {
            CountNonConst0++;
            continue;
        }
        // check if the output can be not constant 0
        if Aig_Regular(pChild).fPhase != uint32(Aig_IsComplement(pChild)) {
            CountNonConst0++;
            continue;
        }
        CountUndecided++;
    }

    if fVerbose != 0 {
        Abc_Print( 1, "Miter has %d outputs. ", Saig_ManPoNum(p) );
        Abc_Print( 1, "Const0 = %d.  ", CountConst0 );
        Abc_Print( 1, "NonConst0 = %d.  ", CountNonConst0 );
        Abc_Print( 1, "Undecided = %d.  ", CountUndecided );
        Abc_Print( 1, "\n" );
    }

    if CountNonConst0 {
        return 0; }
    if CountUndecided != 0 {
        return -1; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Transfer equivalent pairs to the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_TransferSignalPairs(pMiter *Aig_Man_t, pAig1 *Aig_Man_t, pAig2 *Aig_Man_t, vIds1 *Vec_Int_t, vIds2 *Vec_Int_t) *Vec_Int_t {
    var vIds *Vec_Int_t;
    var pObj1, pObj2 *Aig_Obj_t;
    var pObj1m, pObj2m *Aig_Obj_t;
    var i int32;
    vIds = Vec_IntAlloc( 2 * Vec_IntSize(vIds1) );
    for  i = 0; i < Vec_IntSize(vIds1); i++  {
        pObj1 = Aig_ManObj( pAig1, Vec_IntEntry(vIds1, i) );
        pObj2 = Aig_ManObj( pAig2, Vec_IntEntry(vIds2, i) );
        pObj1m = Aig_Regular((Aig_Obj_t *)pObj1.pData);
        pObj2m = Aig_Regular((Aig_Obj_t *)pObj2.pData);
        assert(pObj1m && (pObj2m != nil), "pObj1m && pObj2m");
        if pObj1m == pObj2m {
            continue; }
        if pObj1m.Id < pObj2m.Id {
            Vec_IntPush( vIds, pObj1m.Id );
            Vec_IntPush( vIds, pObj2m.Id );
        
        } else {
            Vec_IntPush( vIds, pObj2m.Id );
            Vec_IntPush( vIds, pObj1m.Id );
        }
    }
    return vIds;
}

/**Function*************************************************************

  Synopsis    [Transform pairs into class representation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_TransformPairsIntoTempClasses(vPairs *Vec_Int_t, nObjNumMax int32) **Vec_Int_t {
    var pvClasses **Vec_Int_t;   // vector of classes
    var pReprs *int32;             // mapping nodes into their representatives
    var Entry, idObj, idRepr, idReprObj, idReprRepr, i int32;
    // allocate data-structures
    // inlining of macro ABC_CALLOC
    pvClasses = ((Vec_Int_t * *) calloc((size_t)(nObjNumMax), sizeof(Vec_Int_t *)));
    // inlining of macro ABC_ALLOC
    pReprs    = ((int *) malloc(sizeof(int) * (size_t)(nObjNumMax)));
    for  i = 0; i < nObjNumMax; i++  {
        pReprs[i] = -1;
    }
    // consider pairs
    for  i = 0; i < Vec_IntSize(vPairs); i += 2  {
        // get both objects
        idRepr = Vec_IntEntry( vPairs, i   );
        idObj  = Vec_IntEntry( vPairs, i+1 );
        assert(idObj > 0, "idObj > 0");
        assert((pReprs[idRepr] == -1) || (pvClasses[pReprs[idRepr]] != nil), "(pReprs[idRepr] == -1) || (pvClasses[pReprs[idRepr]] != nil)");
        assert((pReprs[idObj]  == -1) || (pvClasses[pReprs[idObj] ] != nil), "(pReprs[idObj]  == -1) || (pvClasses[pReprs[idObj] ] != nil)");
        // get representatives of both objects
        idReprRepr = pReprs[idRepr];
        idReprObj  = pReprs[idObj];
        // check different situations
        if idReprRepr == -1 && idReprObj == -1 {   // they do not have classes
            // create a class
            pvClasses[idRepr] = Vec_IntAlloc( 4 );
            Vec_IntPush( pvClasses[idRepr], idRepr );
            Vec_IntPush( pvClasses[idRepr], idObj );
            pReprs[ idRepr ] = idRepr;
            pReprs[ idObj  ] = idRepr;
        
        } else if idReprRepr >= 0 && idReprObj == -1 {   // representative has a class
            // add iObj to the same class
            Vec_IntPushUniqueOrder( pvClasses[idReprRepr], idObj );
            pReprs[ idObj ] = idReprRepr;
        
        } else if idReprRepr == -1 && idReprObj >= 0 {   // object has a class
            assert(idReprObj != idRepr, "idReprObj != idRepr");
            if idReprObj < idRepr { // add idRepr to the same class
                Vec_IntPushUniqueOrder( pvClasses[idReprObj], idRepr );
                pReprs[ idRepr ] = idReprObj;
            
            } else { // if ( idReprObj > idRepr )
             // make idRepr new representative
                Vec_IntPushFirst( pvClasses[idReprObj], idRepr );
                pvClasses[idRepr] = pvClasses[idReprObj];
                pvClasses[idReprObj] = nil;
                // set correct representatives of each node
                // inlining of macro Vec_IntForEachEntry
                for  i = 0; (i < Vec_IntSize(pvClasses[idRepr])) && (((Entry) = Vec_IntEntry(pvClasses[idRepr], i)), 1); i++  {
                    pReprs[ Entry ] = idRepr;
                }
            }
        
        } else { // if ( idReprRepr >= 0 && idReprObj >= 0 )
           // both have classes
            if idReprRepr == idReprObj {  // the classes are the same
                // nothing to do
            
            } else {  // the classes are different
                // find the repr of the new class
                if idReprRepr < idReprObj {
                    // inlining of macro Vec_IntForEachEntry
                    for  i = 0; (i < Vec_IntSize(pvClasses[idReprObj])) && (((Entry) = Vec_IntEntry(pvClasses[idReprObj], i)), 1); i++  {
                        Vec_IntPushUniqueOrder( pvClasses[idReprRepr], Entry );
                        pReprs[ Entry ] = idReprRepr;
                    }
                    Vec_IntFree( pvClasses[idReprObj] );
                    pvClasses[idReprObj] = nil;
                
                } else { // if ( idReprRepr > idReprObj )
                
                    // inlining of macro Vec_IntForEachEntry
                    for  i = 0; (i < Vec_IntSize(pvClasses[idReprRepr])) && (((Entry) = Vec_IntEntry(pvClasses[idReprRepr], i)), 1); i++  {
                        Vec_IntPushUniqueOrder( pvClasses[idReprObj], Entry );
                        pReprs[ Entry ] = idReprObj;
                    }
                    Vec_IntFree( pvClasses[idReprRepr] );
                    pvClasses[idReprRepr] = nil;
                }
            }
        }
    }
    pReprs = nil;
    return pvClasses;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_FreeTempClasses(pvClasses **Vec_Int_t, nObjNumMax int32) {
    var i int32;
    for  i = 0; i < nObjNumMax; i++  {
        if pvClasses[i] {
            Vec_IntFree( pvClasses[i] ); }
    }
    pvClasses = nil;
}

/**Function*************************************************************

  Synopsis    [Performs signal correspondence for the miter of two AIGs with node pairs defined.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SignalCorrespondenceWithPairs(pAig1 *Aig_Man_t, pAig2 *Aig_Man_t, vIds1 *Vec_Int_t, vIds2 *Vec_Int_t, pPars *Ssw_Pars_t) *Aig_Man_t {
    var p *Ssw_Man_t;
    var pAigNew, pMiter *Aig_Man_t;
    var Pars Ssw_Pars_t;
    var vPairs *Vec_Int_t;
    var pvClasses **Vec_Int_t;
    assert(Vec_IntSize(vIds1) == Vec_IntSize(vIds2), "Vec_IntSize(vIds1) == Vec_IntSize(vIds2)");
    // create sequential miter
    pMiter = Saig_ManCreateMiter( pAig1, pAig2, 0 );
    Aig_ManCleanup( pMiter );
    // transfer information to the miter
    vPairs = Ssw_TransferSignalPairs( pMiter, pAig1, pAig2, vIds1, vIds2 );
    // create representation of the classes
    pvClasses = Ssw_TransformPairsIntoTempClasses( vPairs, Aig_ManObjNumMax(pMiter) );
    Vec_IntFree( vPairs );
    // if parameters are not given, create them
    if pPars == nil {
        Ssw_ManSetDefaultParams( pPars = &Pars ); }
    // start the induction manager
    p = Ssw_ManCreate( pMiter, pPars );
    // create equivalence classes using these IDs
    p.ppClasses = Ssw_ClassesPreparePairs( pMiter, pvClasses );
    p.pSml = Ssw_SmlStart( pMiter, 0, p.nFrames + p.pPars.nFramesAddSim, 1 );
    Ssw_ClassesSetData( p.ppClasses, p.pSml, (unsigned(*)(void *,Aig_Obj_t *))Ssw_SmlObjHashWord, (int(*)(void *,Aig_Obj_t *))Ssw_SmlObjIsConstWord, (int(*)(void *,Aig_Obj_t *,Aig_Obj_t *))Ssw_SmlObjsAreEqualWord );
    // perform refinement of classes
    pAigNew = Ssw_SignalCorrespondenceRefine( p );
    // cleanup
    Ssw_FreeTempClasses( pvClasses, Aig_ManObjNumMax(pMiter) );
    Ssw_ManStop( p );
    Aig_ManStop( pMiter );
    return pAigNew;
}

/**Function*************************************************************

  Synopsis    [Runs inductive SEC for the miter of two AIGs with node pairs defined.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SignalCorrespondeceTestPairs(pAig *Aig_Man_t) *Aig_Man_t {
    var pAigNew, pAigRes *Aig_Man_t;
    var Pars Ssw_Pars_t;
    var pPars *Ssw_Pars_t = &Pars;
    var vIds1, vIds2 *Vec_Int_t;
    var pObj, pRepr *Aig_Obj_t;
    var RetValue, i int32;
    var clk abctime  = Abc_Clock();
    Ssw_ManSetDefaultParams( pPars );
    pPars.fVerbose = 1;
    pAigNew = Ssw_SignalCorrespondence( pAig, pPars );
    // record pairs of equivalent nodes
    vIds1 = Vec_IntAlloc( Aig_ManObjNumMax(pAig) );
    vIds2 = Vec_IntAlloc( Aig_ManObjNumMax(pAig) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pRepr = Aig_Regular((Aig_Obj_t *)pObj.pData);
        if pRepr == nil {
            continue; }
        if Aig_ManObj(pAigNew, pRepr.Id) == nil {
            continue; }
/*
        if ( Aig_ObjIsNode(pObj) )
            Abc_Print( 1, "n " );
        else if ( Saig_ObjIsPi(pAig, pObj) )
            Abc_Print( 1, "pi " );
        else if ( Saig_ObjIsLo(pAig, pObj) )
            Abc_Print( 1, "lo " );
*/
        Vec_IntPush( vIds1, Aig_ObjId(pObj) );
        Vec_IntPush( vIds2, Aig_ObjId(pRepr) );
    }
    }
    Abc_Print( 1, "Recorded %d pairs (before: %d  after: %d).\n", Vec_IntSize(vIds1), Aig_ManObjNumMax(pAig), Aig_ManObjNumMax(pAigNew) );
    // try the new AIGs
    pAigRes = Ssw_SignalCorrespondenceWithPairs( pAig, pAigNew, vIds1, vIds2, pPars );
    Vec_IntFree( vIds1 );
    Vec_IntFree( vIds2 );
    // report the results
    RetValue = Ssw_MiterStatus( pAigRes, 1 );
    if RetValue == 1 {
        Abc_Print( 1, "Verification successful.  " );
    } else if RetValue == 0 {
        Abc_Print( 1, "Verification failed with the counter-example.  " );
    } else {
        Abc_Print( 1, "Verification UNDECIDED. Remaining registers %d (total %d).  ",
            Aig_ManRegNum(pAigRes), Aig_ManRegNum(pAig) + Aig_ManRegNum(pAigNew) ); }
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    // cleanup
    Aig_ManStop( pAigNew );
    return pAigRes;
}

/**Function*************************************************************

  Synopsis    [Runs inductive SEC for the miter of two AIGs with node pairs defined.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SecWithPairs(pAig1 *Aig_Man_t, pAig2 *Aig_Man_t, vIds1 *Vec_Int_t, vIds2 *Vec_Int_t, pPars *Ssw_Pars_t) int32 {
    var pAigRes *Aig_Man_t;
    var RetValue int32;
    var clk abctime  = Abc_Clock();
    assert(vIds1 != nil && vIds2 != nil, "vIds1 != nil && vIds2 != nil");
    // try the new AIGs
    Abc_Print( 1, "Performing specialized verification with node pairs.\n" );
    pAigRes = Ssw_SignalCorrespondenceWithPairs( pAig1, pAig2, vIds1, vIds2, pPars );
    // report the results
    RetValue = Ssw_MiterStatus( pAigRes, 1 );
    if RetValue == 1 {
        Abc_Print( 1, "Verification successful.  " );
    } else if RetValue == 0 {
        Abc_Print( 1, "Verification failed with a counter-example.  " );
    } else {
        Abc_Print( 1, "Verification UNDECIDED. The number of remaining regs = %d (total = %d).  ",
            Aig_ManRegNum(pAigRes), Aig_ManRegNum(pAig1) + Aig_ManRegNum(pAig2) ); }
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    // cleanup
    Aig_ManStop( pAigRes );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Runs inductive SEC for the miter of two AIGs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SecGeneral(pAig1 *Aig_Man_t, pAig2 *Aig_Man_t, pPars *Ssw_Pars_t) int32 {
    var pAigRes, pMiter *Aig_Man_t;
    var RetValue int32;
    var clk abctime  = Abc_Clock();
    // try the new AIGs
    Abc_Print( 1, "Performing general verification without node pairs.\n" );
    pMiter = Saig_ManCreateMiter( pAig1, pAig2, 0 );
    Aig_ManCleanup( pMiter );
    pAigRes = Ssw_SignalCorrespondence( pMiter, pPars );
    Aig_ManStop( pMiter );
    // report the results
    RetValue = Ssw_MiterStatus( pAigRes, 1 );
    if RetValue == 1 {
        Abc_Print( 1, "Verification successful.  " );
    } else if RetValue == 0 {
        Abc_Print( 1, "Verification failed with a counter-example.  " );
    } else {
        Abc_Print( 1, "Verification UNDECIDED. The number of remaining regs = %d (total = %d).  ",
            Aig_ManRegNum(pAigRes), Aig_ManRegNum(pAig1) + Aig_ManRegNum(pAig2) ); }
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    // cleanup
    Aig_ManStop( pAigRes );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Runs inductive SEC for the miter of two AIGs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SecGeneralMiter(pMiter *Aig_Man_t, pPars *Ssw_Pars_t) int32 {
    var pAigRes *Aig_Man_t;
    var RetValue int32;
    var clk abctime  = Abc_Clock();
    // try the new AIGs
//    Abc_Print( 1, "Performing general verification without node pairs.\n" );
    pAigRes = Ssw_SignalCorrespondence( pMiter, pPars );
    // report the results
    RetValue = Ssw_MiterStatus( pAigRes, 1 );
    if RetValue == 1 {
        Abc_Print( 1, "Verification successful.  " );
    } else if RetValue == 0 {
        Abc_Print( 1, "Verification failed with a counter-example.  " );
    } else {
        Abc_Print( 1, "Verification UNDECIDED. The number of remaining regs = %d (total = %d).  ",
            Aig_ManRegNum(pAigRes), Aig_ManRegNum(pMiter) ); }
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    // cleanup
    Aig_ManStop( pAigRes );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
