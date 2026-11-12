package live

/**CFile****************************************************************

  FileName    [arenaViolation.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [module for addition of arena violator detector
        induced by stabilizing constraints]

  Author      [Sayak Ray]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 31, 2012.]

***********************************************************************/

//#include <stdio.h>
//#include "base/main/main.h"
//#include "aig/aig/aig.h"
//#include "aig/saig/saig.h"
//#include <string.h>
//#include "base/main/mainInt.h"
//#include "proof/pdr/pdr.h"

//#define DISJUNCTIVE_CONSTRAINT_ENABLE_MODE
#define BARRIER_MONOTONE_TEST

//ABC_NAMESPACE_IMPL_START

func createArenaLO(pAigNew *Aig_Man_t, vBarriers *Vec_Ptr_t) *Vec_Ptr_t {
    var vArenaLO *Vec_Ptr_t;
    var barrierCount int32;
    var pObj *Aig_Obj_t;
    var i int32;

    if vBarriers == nil {
        return nil; }

    barrierCount = Vec_PtrSize(vBarriers);
    if barrierCount <= 0 {
        return nil; }

    vArenaLO = Vec_PtrAlloc(barrierCount);    
    for  i=0; i<barrierCount; i++  {
        pObj = Aig_ObjCreateCi( pAigNew );
        Vec_PtrPush( vArenaLO, pObj );
    }

    return vArenaLO;
}

func createArenaLi(pAigNew *Aig_Man_t, vBarriers *Vec_Ptr_t, vArenaSignal *Vec_Ptr_t) *Vec_Ptr_t {
    var vArenaLi *Vec_Ptr_t;
    var barrierCount int32;
    var i int32;
    var pObj, pObjDriver *Aig_Obj_t;

    if vBarriers == nil {
        return nil; }

    barrierCount = Vec_PtrSize(vBarriers);
    if barrierCount <= 0 {
        return nil; }

    vArenaLi = Vec_PtrAlloc(barrierCount);    
    for  i=0; i<barrierCount; i++  {
        pObjDriver = (Aig_Obj_t *)Vec_PtrEntry( vArenaSignal, i );
        pObj = Aig_ObjCreateCo( pAigNew, pObjDriver );
        Vec_PtrPush( vArenaLi, pObj );
    }

    return vArenaLi;
}

func createMonotoneBarrierLO(pAigNew *Aig_Man_t, vBarriers *Vec_Ptr_t) *Vec_Ptr_t {
    var vMonotoneLO *Vec_Ptr_t;
    var barrierCount int32;
    var pObj *Aig_Obj_t;
    var i int32;

    if vBarriers == nil {
        return nil; }

    barrierCount = Vec_PtrSize(vBarriers);
    if barrierCount <= 0 {
        return nil; }

    vMonotoneLO = Vec_PtrAlloc(barrierCount);    
    for  i=0; i<barrierCount; i++  {
        pObj = Aig_ObjCreateCi( pAigNew );
        Vec_PtrPush( vMonotoneLO, pObj );
    }

    return vMonotoneLO;
}

func driverToPoNew(pAig *Aig_Man_t, pObjPo *Aig_Obj_t) *Aig_Obj_t {
    var poDriverOld *Aig_Obj_t;
    var poDriverNew *Aig_Obj_t;    
    
    //Aig_ObjPrint( pAig, pObjPo );
    //printf("\n");

    assert(Aig_ObjIsCo(pObjPo), "Aig_ObjIsCo(pObjPo)");
    poDriverOld = Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObjPo), Aig_ObjFaninC0(pObjPo));
    assert(!Aig_ObjIsCo(poDriverOld), "!Aig_ObjIsCo(poDriverOld)");
    if !Aig_IsComplement(poDriverOld) {
        poDriverNew = (Aig_Obj_t *)(Aig_Regular(poDriverOld).pData);
    } else {
        poDriverNew = Aig_Not((Aig_Obj_t *)(Aig_Regular(poDriverOld).pData));
    }
    //assert( !Aig_ObjIsCo(poDriverNew) );
    return poDriverNew;
}

func collectBarrierDisjunctions(pAigOld *Aig_Man_t, pAigNew *Aig_Man_t, vBarriers *Vec_Ptr_t) *Vec_Ptr_t {
    var barrierCount, i, j, jElem int32;
    var vIthBarrier *Vec_Int_t;
    var pObjBarrier, pObjCurr, pObjTargetPoOld *Aig_Obj_t;
    var vNewBarrierSignals *Vec_Ptr_t;

    if vBarriers == nil {
        return nil; }
    barrierCount = Vec_PtrSize( vBarriers );
    if barrierCount <= 0 {
        return nil; }

    vNewBarrierSignals = Vec_PtrAlloc( barrierCount );

    for  i=0; i<barrierCount; i++  {
        vIthBarrier = (Vec_Int_t *)Vec_PtrEntry( vBarriers, i );
        assert(Vec_IntSize( vIthBarrier ) >= 1, "Vec_IntSize( vIthBarrier ) >= 1");
        pObjBarrier = Aig_Not(Aig_ManConst1(pAigNew));
        Vec_IntForEachEntry( vIthBarrier, jElem, j )
        {
            pObjTargetPoOld = Aig_ManCo( pAigOld, jElem );
            //Aig_ObjPrint( pAigOld, pObjTargetPoOld );
            //printf("\n");
            pObjCurr = driverToPoNew( pAigOld, pObjTargetPoOld );
            pObjBarrier = Aig_Or( pAigNew, pObjCurr, pObjBarrier );
        }
        assert(pObjBarrier, "pObjBarrier");
        Vec_PtrPush(vNewBarrierSignals, pObjBarrier);
    }
    assert(Vec_PtrSize( vNewBarrierSignals ) == barrierCount, "Vec_PtrSize( vNewBarrierSignals ) == barrierCount");

    return vNewBarrierSignals;
}

func Aig_Xor(pAig *Aig_Man_t, pObj1 *Aig_Obj_t, pObj2 *Aig_Obj_t) *Aig_Obj_t {
    return Aig_Or( pAig, Aig_And( pAig, pObj1, Aig_Not(pObj2) ), Aig_And( pAig, Aig_Not(pObj1), pObj2 ) );
}

func createArenaViolation(pAigOld *Aig_Man_t, pAigNew *Aig_Man_t, pWindowBegins *Aig_Obj_t, pWithinWindow *Aig_Obj_t, vMasterBarriers *Vec_Ptr_t, vBarrierLo *Vec_Ptr_t, vBarrierLiDriver *Vec_Ptr_t, vMonotoneDisjunctionNodes *Vec_Ptr_t) *Aig_Obj_t {
    var pWindowBeginsLocal *Aig_Obj_t  = pWindowBegins;
    var pWithinWindowLocal *Aig_Obj_t  = pWithinWindow;
    var i int32;
    var pObj, pObjAnd1, pObjOr1, pObjAnd2, pObjBarrierLo, pObjBarrierSwitch, pObjArenaViolation *Aig_Obj_t;
    var vBarrierSignals *Vec_Ptr_t;

    assert(vBarrierLiDriver != nil, "vBarrierLiDriver != nil");
    assert(vMonotoneDisjunctionNodes != nil, "vMonotoneDisjunctionNodes != nil");

    pObjArenaViolation = Aig_Not(Aig_ManConst1( pAigNew ));

    vBarrierSignals = collectBarrierDisjunctions(pAigOld, pAigNew, vMasterBarriers);
    assert(vBarrierSignals != nil, "vBarrierSignals != nil");

    assert(Vec_PtrSize( vMonotoneDisjunctionNodes ) == 0, "Vec_PtrSize( vMonotoneDisjunctionNodes ) == 0");
    Vec_PtrForEachEntry( Aig_Obj_t *, vBarrierSignals, pObj, i )
        Vec_PtrPush( vMonotoneDisjunctionNodes, pObj );
    assert(Vec_PtrSize( vMonotoneDisjunctionNodes ) == Vec_PtrSize( vMasterBarriers ), "Vec_PtrSize( vMonotoneDisjunctionNodes ) == Vec_PtrSize( vMasterBarriers )");

    Vec_PtrForEachEntry( Aig_Obj_t *, vBarrierSignals, pObj, i )
    {
        //pObjNew = driverToPoNew( pAigOld, pObj );
        pObjAnd1 = Aig_And(pAigNew, pObj, pWindowBeginsLocal);
        pObjBarrierLo = (Aig_Obj_t *)Vec_PtrEntry( vBarrierLo, i );
        pObjOr1 = Aig_Or(pAigNew, pObjAnd1, pObjBarrierLo);
        Vec_PtrPush( vBarrierLiDriver, pObjOr1 );

        pObjBarrierSwitch = Aig_Xor( pAigNew, pObj, pObjBarrierLo );
        pObjAnd2 = Aig_And( pAigNew, pObjBarrierSwitch, pWithinWindowLocal );
        pObjArenaViolation = Aig_Or( pAigNew, pObjAnd2, pObjArenaViolation );
    }

    Vec_PtrFree(vBarrierSignals);
    return pObjArenaViolation;
}

func createConstrained0LiveConeWithDSC(pNewAig *Aig_Man_t, signalList *Vec_Ptr_t) *Aig_Obj_t {
    var pConsequent, pConsequentCopy, pAntecedent, p0LiveCone, pObj *Aig_Obj_t;
    var i, numSigAntecedent int32;
    
    numSigAntecedent = Vec_PtrSize( signalList ) - 1;

    pAntecedent = Aig_ManConst1( pNewAig );
    pConsequent = (Aig_Obj_t *)Vec_PtrEntry( signalList, numSigAntecedent );
    pConsequentCopy = Aig_NotCond( (Aig_Obj_t *)(Aig_Regular(pConsequent).pData), Aig_IsComplement( pConsequent ) );

    for i=0; i<numSigAntecedent; i++  {
        pObj = (Aig_Obj_t *)Vec_PtrEntry( signalList, i );
        assert(Aig_Regular(pObj).pData, "Aig_Regular(pObj).pData");
        pAntecedent = Aig_And( pNewAig, pAntecedent, Aig_NotCond((Aig_Obj_t *)(Aig_Regular(pObj).pData), Aig_IsComplement(pObj)) );
    }

    p0LiveCone = Aig_Or( pNewAig, Aig_Not(pAntecedent), pConsequentCopy );

    return p0LiveCone;
}

func collectCSSignalsWithDSC(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) *Vec_Ptr_t {
    var i int32;
    var pObj *Aig_Obj_t;
    var pConsequent *Aig_Obj_t = nil;
    var vNodeArray *Vec_Ptr_t;

    vNodeArray = Vec_PtrAlloc(1);

    Saig_ManForEachPo( pAig, pObj, i )
    {
        if strstr( Abc_ObjName(Abc_NtkPo( pNtk, i )), "csLiveConst_" ) != nil {
            Vec_PtrPush( vNodeArray, Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObj), Aig_ObjFaninC0(pObj)) );    
        } else if strstr( Abc_ObjName(Abc_NtkPo( pNtk, i )), "csLiveTarget_" ) != nil {
            pConsequent = Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObj), Aig_ObjFaninC0(pObj)); }
    }
    assert((pConsequent != nil), "pConsequent");
    Vec_PtrPush( vNodeArray, pConsequent );    
    return vNodeArray;
}

func collectWindowBeginSignalWithDSC(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) int32 {
    var i int32;
    var pObj *Aig_Obj_t;

    Saig_ManForEachPo( pAig, pObj, i )
    {
        if strstr( Abc_ObjName(Abc_NtkPo( pNtk, i )), "windowBegins_" ) != nil {
            return i;            
        }
    }
    
    return -1;
}

func collectWithinWindowSignalWithDSC(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) int32 {
    var i int32;
    var pObj *Aig_Obj_t;

    Saig_ManForEachPo( pAig, pObj, i )
    {
        if strstr( Abc_ObjName(Abc_NtkPo( pNtk, i )), "withinWindow_" ) != nil {
            return i; }            
    }
    
    return -1;
}

func collectPendingSignalWithDSC(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) int32 {
    var i int32;
    var pObj *Aig_Obj_t;

    Saig_ManForEachPo( pAig, pObj, i )
    {
        if strstr( Abc_ObjName(Abc_NtkPo( pNtk, i )), "pendingSignal" ) != nil {
            return i; }            
    }
    
    return -1;
}

func createAndGateForMonotonicityVerification(pNewAig *Aig_Man_t, vDisjunctionSignals *Vec_Ptr_t, vDisjunctionLo *Vec_Ptr_t, pendingLo *Aig_Obj_t, pendingSignal *Aig_Obj_t) *Aig_Obj_t {
    var pObjBigAnd, pObj, pObjLo, pObjImply *Aig_Obj_t;
    var pObjPendingAndPendingLo *Aig_Obj_t;
    var i int32;

    pObjBigAnd = Aig_ManConst1( pNewAig );
    pObjPendingAndPendingLo = Aig_And( pNewAig, pendingLo, pendingSignal );
    Vec_PtrForEachEntry( Aig_Obj_t *, vDisjunctionSignals, pObj, i )
    {
        pObjLo = (Aig_Obj_t *)Vec_PtrEntry( vDisjunctionLo, i );    
        pObjImply = Aig_Or( pNewAig, Aig_Not(Aig_And( pNewAig, pObjPendingAndPendingLo, pObjLo)),
                    pObj );
        pObjBigAnd = Aig_And( pNewAig, pObjBigAnd, pObjImply );
    }
    return pObjBigAnd;
}

func createNewAigWith0LivePoWithDSC(pAig *Aig_Man_t, signalList *Vec_Ptr_t, index0Live *int32, windowBeginIndex int32, withinWindowIndex int32, pendingSignalIndex int32, vBarriers *Vec_Ptr_t) *Aig_Man_t {
    var pNewAig *Aig_Man_t;
    var pObj, pObjNewPoDriver *Aig_Obj_t;
    var i int32;
    var loCopied, loCreated, liCopied, liCreated int32;
    var pObjWindowBeginsNew, pObjWithinWindowNew, pObjArenaViolation, pObjTarget, pObjArenaViolationLiDriver *Aig_Obj_t;
    var pObjNewPoDriverArenaViolated, pObjArenaViolationLo *Aig_Obj_t;
    var vBarrierLo, vBarrierLiDriver, vBarrierLi *Vec_Ptr_t;
    var vMonotoneNodes *Vec_Ptr_t;

                            
    var pObjPendingSignal *Aig_Obj_t;
    var pObjPendingFlopLo *Aig_Obj_t;
    var vMonotoneBarrierLo *Vec_Ptr_t;
    var pObjPendingAndPendingSignal, pObjMonotoneAnd, pObjCurrMonotoneLo *Aig_Obj_t;
      

    //assert( Vec_PtrSize( signalList ) > 1 );

    //****************************************************************
    // Step1: create the new manager
    // Note: The new manager is created with "2 * Aig_ManObjNumMax(p)"
    // nodes, but this selection is arbitrary - need to be justified
    //****************************************************************
    pNewAig = Aig_ManStart( Aig_ManObjNumMax(pAig) );
    pNewAig.pName = (char *)malloc( len(pAig.pName) + len("_0Live") + 1 );
    sprintf(pNewAig.pName, "%s_%s", pAig.pName, "0Live");
        pNewAig.pSpec = nil;

    //****************************************************************
    // Step 2: map constant nodes
    //****************************************************************
        pObj = Aig_ManConst1( pAig );
        pObj.pData = Aig_ManConst1( pNewAig );

    //****************************************************************
        // Step 3: create true PIs
    //****************************************************************
        Saig_ManForEachPi( pAig, pObj, i )
    {
        pObj.pData = Aig_ObjCreateCi( pNewAig );
    }

    //****************************************************************
    // Step 4: create register outputs
    //****************************************************************
        Saig_ManForEachLo( pAig, pObj, i )
        {
        loCopied++;
        pObj.pData = Aig_ObjCreateCi( pNewAig );
        }

    //****************************************************************
    // Step 4.a: create register outputs for the barrier flops
    //****************************************************************
    vBarrierLo = createArenaLO( pNewAig, vBarriers );
    loCreated = Vec_PtrSize(vBarrierLo);

    //****************************************************************
    // Step 4.b: create register output for arenaViolationFlop
    //****************************************************************
    pObjArenaViolationLo = Aig_ObjCreateCi( pNewAig );
    loCreated++;

                            
    //****************************************************************
    // Step 4.c: create register output for pendingFlop
    //****************************************************************
    pObjPendingFlopLo = Aig_ObjCreateCi( pNewAig );
    loCreated++;

    //****************************************************************
    // Step 4.d: create register outputs for the barrier flops
    // for asserting monotonicity
    //****************************************************************
    vMonotoneBarrierLo = createMonotoneBarrierLO( pNewAig, vBarriers );
    loCreated = loCreated + Vec_PtrSize(vMonotoneBarrierLo);
      

    //********************************************************************
    // Step 5: create internal nodes
    //********************************************************************
        Aig_ManForEachNode( pAig, pObj, i )
    {
        pObj.pData = Aig_And( pNewAig, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }

    //********************************************************************
    // Step 5.a: create internal nodes corresponding to arenaViolation
    //********************************************************************
    pObjTarget = Aig_ManCo( pAig, windowBeginIndex );
    pObjWindowBeginsNew = driverToPoNew( pAig, pObjTarget );

    pObjTarget = Aig_ManCo( pAig, withinWindowIndex );
    pObjWithinWindowNew = driverToPoNew( pAig, pObjTarget );

    vBarrierLiDriver = Vec_PtrAlloc( Vec_PtrSize(vBarriers) );
    vMonotoneNodes = Vec_PtrAlloc( Vec_PtrSize(vBarriers) );

    pObjArenaViolation = createArenaViolation( pAig, pNewAig, 
                pObjWindowBeginsNew, pObjWithinWindowNew, 
                vBarriers, vBarrierLo, vBarrierLiDriver, vMonotoneNodes ); 
    assert(Vec_PtrSize(vMonotoneNodes) == Vec_PtrSize(vBarriers), "Vec_PtrSize(vMonotoneNodes) == Vec_PtrSize(vBarriers)");

#ifdef ARENA_VIOLATION_CONSTRAINT

#endif

    pObjArenaViolationLiDriver = Aig_Or( pNewAig, pObjArenaViolation, pObjArenaViolationLo );

                            
    //********************************************************************
    // Step 5.b: Create internal nodes for monotone testing
    //********************************************************************

    pObjTarget = Aig_ManCo( pAig, pendingSignalIndex );
    pObjPendingSignal = driverToPoNew( pAig, pObjTarget );
    
    pObjPendingAndPendingSignal = Aig_And( pNewAig, pObjPendingSignal, pObjPendingFlopLo );
    pObjMonotoneAnd = Aig_ManConst1( pNewAig );
    Vec_PtrForEachEntry( Aig_Obj_t *, vMonotoneNodes, pObj, i )
    {
        pObjCurrMonotoneLo = (Aig_Obj_t *)Vec_PtrEntry(vMonotoneBarrierLo, i);
        pObjMonotoneAnd = Aig_And( pNewAig, pObjMonotoneAnd,
            Aig_Or( pNewAig, 
            Aig_Not(Aig_And(pNewAig, pObjPendingAndPendingSignal, pObjCurrMonotoneLo)),
            pObj ) );
    }
      

    //********************************************************************
    // Step 6: create primary outputs
    //********************************************************************

    Saig_ManForEachPo( pAig, pObj, i )
    {
        pObj.pData = Aig_ObjCreateCo( pNewAig, Aig_ObjChild0Copy(pObj) ); 
    }

    pObjNewPoDriver = createConstrained0LiveConeWithDSC( pNewAig, signalList );
    pObjNewPoDriverArenaViolated = Aig_Or( pNewAig, pObjNewPoDriver, pObjArenaViolationLo );
                            
    pObjNewPoDriverArenaViolated = Aig_And( pNewAig, pObjNewPoDriverArenaViolated, pObjMonotoneAnd );
      
    Aig_ObjCreateCo( pNewAig, pObjNewPoDriverArenaViolated );

    *index0Live = i;

    //********************************************************************
    // Step 7: create register inputs
    //********************************************************************

    Saig_ManForEachLi( pAig, pObj, i )
    {
        liCopied++;
        pObj.pData = Aig_ObjCreateCo( pNewAig, Aig_ObjChild0Copy(pObj) );
    }

    //********************************************************************
    // Step 7.a: create register inputs for barrier flops
    //********************************************************************
    assert(Vec_PtrSize(vBarrierLiDriver) == Vec_PtrSize(vBarriers), "Vec_PtrSize(vBarrierLiDriver) == Vec_PtrSize(vBarriers)");
    vBarrierLi = createArenaLi( pNewAig, vBarriers, vBarrierLiDriver );
    liCreated = Vec_PtrSize( vBarrierLi );

    //********************************************************************
    // Step 7.b: create register inputs for arenaViolation flop
    //********************************************************************
    Aig_ObjCreateCo( pNewAig, pObjArenaViolationLiDriver );
    liCreated++;

                            
    //****************************************************************
    // Step 7.c: create register input for pendingFlop
    //****************************************************************
    Aig_ObjCreateCo( pNewAig, pObjPendingSignal);
    liCreated++;

    //********************************************************************
    // Step 7.d: create register inputs for the barrier flops
    // for asserting monotonicity
    //********************************************************************
    Vec_PtrForEachEntry( Aig_Obj_t *, vMonotoneNodes, pObj, i )
    {
        Aig_ObjCreateCo( pNewAig, pObj );
        liCreated++;
    }
      

    assert(loCopied + loCreated == liCopied + liCreated, "loCopied + loCreated == liCopied + liCreated");
    //next step should be changed
    Aig_ManSetRegNum( pNewAig, loCopied + loCreated );
    Aig_ManCleanup( pNewAig );

    assert(Aig_ManCheck( pNewAig ), "Aig_ManCheck( pNewAig )");

    Vec_PtrFree(vBarrierLo);
    Vec_PtrFree(vMonotoneBarrierLo);
    Vec_PtrFree(vBarrierLiDriver);
    Vec_PtrFree(vBarrierLi);
    Vec_PtrFree(vMonotoneNodes);

    return pNewAig;
}

func generateWorkingAigWithDSC(pAig *Aig_Man_t, pNtk *Abc_Ntk_t, pIndex0Live *int32, vMasterBarriers *Vec_Ptr_t) *Aig_Man_t {
    var vSignalVector *Vec_Ptr_t;
    var pAigNew *Aig_Man_t;
    var pObjWithinWindow int32;
    var pObjWindowBegin int32;
    var pObjPendingSignal int32;

    vSignalVector = collectCSSignalsWithDSC( pNtk, pAig );    

    pObjWindowBegin = collectWindowBeginSignalWithDSC( pNtk, pAig );
    pObjWithinWindow = collectWithinWindowSignalWithDSC( pNtk, pAig );
    pObjPendingSignal = collectPendingSignalWithDSC( pNtk, pAig );

    pAigNew = createNewAigWith0LivePoWithDSC( pAig, vSignalVector, pIndex0Live, pObjWindowBegin, pObjWithinWindow, pObjPendingSignal, vMasterBarriers );
    Vec_PtrFree(vSignalVector);

    return pAigNew;
}

//ABC_NAMESPACE_IMPL_END
