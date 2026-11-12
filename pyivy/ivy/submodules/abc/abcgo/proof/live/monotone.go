package live

import "fmt"

/**CFile****************************************************************

  FileName    [kLiveConstraints.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [Constraint analysis module for the k-Liveness algorithm
        invented by Koen Classen, Niklas Sorensson.]

  Author      [Sayak Ray]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 31, 2012.]

  Revision    [$Id: liveness.c,v 1.00 2009/01/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include "base/main/main.h"
//#include "aig/aig/aig.h"
//#include "aig/saig/saig.h"
//#include <string.h>
//#include "base/main/mainInt.h"
//#include "proof/pdr/pdr.h"

//ABC_NAMESPACE_IMPL_START

// extern Aig_Man_t *Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );

//extern Aig_Man_t *createDisjunctiveMonotoneTester(Aig_Man_t *pAig, struct aigPoIndices *aigPoIndicesArg, struct monotoneVectorsStruct *monotoneVectorArg, int *startMonotonePropPo);

type aigPoIndices struct {
    attrPendingSignalIndex int32;
    attrHintSingalBeginningMarker int32;
    attrHintSingalEndMarker int32;
    attrSafetyInvarIndex int32;
}

func allocAigPoIndices() *struct aigPoIndices {
    struct aigPoIndices *newAigPoIndices;

    newAigPoIndices = (struct aigPoIndices *)malloc(sizeof (struct aigPoIndices));
    newAigPoIndices.attrPendingSignalIndex = -1;
    newAigPoIndices.attrHintSingalBeginningMarker = -1;
    newAigPoIndices.attrHintSingalEndMarker = -1;
    newAigPoIndices.attrSafetyInvarIndex = -1;

    assert(newAigPoIndices != nil, "newAigPoIndices != nil");
    return newAigPoIndices;
}

void deallocAigPoIndices(struct aigPoIndices *toBeDeletedAigPoIndices)
{
    assert(toBeDeletedAigPoIndices != nil, "toBeDeletedAigPoIndices != nil");
    free(toBeDeletedAigPoIndices);
}

type monotoneVectorsStruct struct {
    attrKnownMonotone *Vec_Int_t;
    attrCandMonotone *Vec_Int_t;
    attrHintMonotone *Vec_Int_t;
}

func allocPointersToMonotoneVectors() *struct monotoneVectorsStruct {
    struct monotoneVectorsStruct *newPointersToMonotoneVectors;

    newPointersToMonotoneVectors = (struct monotoneVectorsStruct *)malloc(sizeof (struct monotoneVectorsStruct));

    newPointersToMonotoneVectors.attrKnownMonotone = nil;
    newPointersToMonotoneVectors.attrCandMonotone = nil;
    newPointersToMonotoneVectors.attrHintMonotone = nil;
        
    assert(newPointersToMonotoneVectors != nil, "newPointersToMonotoneVectors != nil");
    return newPointersToMonotoneVectors;
}

void deallocPointersToMonotoneVectors(struct monotoneVectorsStruct *toBeDeleted)
{
    assert(toBeDeleted != nil, "toBeDeleted != nil");
    free( toBeDeleted );
}

func findHintOutputs(pNtk *Abc_Ntk_t) *Vec_Int_t {
    var i, numElementPush int32;
    var pNode *Abc_Obj_t;
    var vHintPoIntdex *Vec_Int_t;

    vHintPoIntdex = Vec_IntAlloc(0);
    Abc_NtkForEachPo( pNtk, pNode, i )
    {
        if strstr( Abc_ObjName( pNode ), "hint_" ) != nil {
            Vec_IntPush( vHintPoIntdex, i );
            numElementPush++;
        }
    }    
    
    if numElementPush == 0 {
        return nil;
    } else {
        return vHintPoIntdex; }
}

func findPendingSignal(pNtk *Abc_Ntk_t) int32 {
    var i int32;
    var pendingSignalIndex int32 = -1;
    var pNode *Abc_Obj_t;

    Abc_NtkForEachPo( pNtk, pNode, i )
    {
        if strstr( Abc_ObjName( pNode ), "pendingSignal" ) != nil {
            pendingSignalIndex = i;
            break;
        }
    }    
    
    return pendingSignalIndex;
}

func checkSanityOfKnownMonotone(vKnownMonotone *Vec_Int_t, vCandMonotone *Vec_Int_t, vHintMonotone *Vec_Int_t) int32 {
    var iElem, i int32;

    Vec_IntForEachEntry( vKnownMonotone, iElem, i )
        fmt.Printf("%d ", iElem);
    fmt.Printf("\n");
    Vec_IntForEachEntry( vCandMonotone, iElem, i )
        fmt.Printf("%d ", iElem);
    fmt.Printf("\n");
    Vec_IntForEachEntry( vHintMonotone, iElem, i )
        fmt.Printf("%d ", iElem);
    fmt.Printf("\n");
    return 1;
}

Aig_Man_t *createMonotoneTester(Aig_Man_t *pAig, struct aigPoIndices *aigPoIndicesArg, struct monotoneVectorsStruct *monotoneVectorArg, int *startMonotonePropPo)
{
    var pNewAig *Aig_Man_t;
    var iElem, i, nRegCount, oldPoNum, poSerialNum, iElemHint int32;
    var piCopied, liCopied, liCreated, loCopied, loCreated int32;
    var poCopied, poCreated int32;
    var pObj, pObjPo, pObjDriver, pObjDriverNew, pObjPendingDriverNew, pObjPendingAndNextPending *Aig_Obj_t;
    var pPendingFlop, pHintSignalLo, pHintMonotoneFlop, pObjTemp1, pObjTemp2, pObjKnownMonotoneAnd *Aig_Obj_t;
    var vHintMonotoneLocalDriverNew *Vec_Ptr_t;
    var vHintMonotoneLocalFlopOutput *Vec_Ptr_t;
    var vHintMonotoneLocalProp *Vec_Ptr_t;

    var pendingSignalIndexLocal int32  = aigPoIndicesArg.attrPendingSignalIndex;
    var hintSingalBeginningMarkerLocal int32  = aigPoIndicesArg.attrHintSingalBeginningMarker;
    //int hintSingalEndMarkerLocal = aigPoIndicesArg->attrHintSingalEndMarker;

    var vKnownMonotoneLocal *Vec_Int_t  = monotoneVectorArg.attrKnownMonotone;
    var vCandMonotoneLocal *Vec_Int_t  = monotoneVectorArg.attrCandMonotone;
    var vHintMonotoneLocal *Vec_Int_t  = monotoneVectorArg.attrHintMonotone;
    
    //****************************************************************
    // Step1: create the new manager
    // Note: The new manager is created with "2 * Aig_ManObjNumMax(p)"
    // nodes, but this selection is arbitrary - need to be justified
    //****************************************************************
    pNewAig = Aig_ManStart( Aig_ManObjNumMax(pAig) );
    pNewAig.pName = (char *)malloc( len(pAig.pName) + len("_monotone") + 1 );
    sprintf(pNewAig.pName, "%s_%s", pAig.pName, "_monotone");
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
        piCopied++;
        pObj.pData = Aig_ObjCreateCi(pNewAig);
    }

    //****************************************************************
    // Step 5: create register outputs
    //****************************************************************
        Saig_ManForEachLo( pAig, pObj, i )
        {
        loCopied++;
        pObj.pData = Aig_ObjCreateCi(pNewAig);
        }

    //****************************************************************
    // Step 6: create "D" register output for PENDING flop 
    //****************************************************************
    loCreated++;
    pPendingFlop = Aig_ObjCreateCi( pNewAig );

    //****************************************************************
    // Step 6.a: create "D" register output for HINT_MONOTONE flop 
    //****************************************************************
    vHintMonotoneLocalFlopOutput = Vec_PtrAlloc(Vec_IntSize(vHintMonotoneLocal));
    Vec_IntForEachEntry( vHintMonotoneLocal, iElem, i )
    {
        loCreated++;
        pHintMonotoneFlop = Aig_ObjCreateCi( pNewAig );
        Vec_PtrPush( vHintMonotoneLocalFlopOutput, pHintMonotoneFlop );
    }

    nRegCount = loCreated + loCopied;
    fmt.Printf("\nnRegCount = %d\n", nRegCount);

    //********************************************************************
    // Step 7: create internal nodes
    //********************************************************************
        Aig_ManForEachNode( pAig, pObj, i )
    {
        pObj.pData = Aig_And( pNewAig, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }
    
    //********************************************************************
    // Step 8: mapping appropriate new flop drivers 
    //********************************************************************

    pObjPo = Aig_ManCo( pAig, pendingSignalIndexLocal );
    pObjDriver = Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObjPo), Aig_ObjFaninC0(pObjPo));
    if !Aig_IsComplement(pObjDriver) {
        pObjPendingDriverNew = (Aig_Obj_t *)(Aig_Regular(pObjDriver).pData);
    } else {
        pObjPendingDriverNew = Aig_Not((Aig_Obj_t *)(Aig_Regular(pObjDriver).pData));
    }

    pObjPendingAndNextPending = Aig_And( pNewAig, pObjPendingDriverNew, pPendingFlop );

    oldPoNum = Aig_ManCoNum(pAig) - Aig_ManRegNum(pAig);
    pObjKnownMonotoneAnd = Aig_ManConst1( pNewAig );
         
    if vKnownMonotoneLocal != nil {
        assert((checkSanityOfKnownMonotone( vKnownMonotoneLocal, vCandMonotoneLocal, vHintMonotoneLocal ) != 0), "checkSanityOfKnownMonotone( vKnownMonotoneLocal, vCandMonotoneLocal, vHintMonotoneLocal )");

        Vec_IntForEachEntry( vKnownMonotoneLocal, iElemHint, i )
        {
            iElem = (iElemHint - hintSingalBeginningMarkerLocal) + 1 + pendingSignalIndexLocal;
            fmt.Printf("\nProcessing knownMonotone = %d\n", iElem);
            pObjPo = Aig_ManCo( pAig, iElem );
            pObjDriver = Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObjPo), Aig_ObjFaninC0(pObjPo));
            if !Aig_IsComplement(pObjDriver) {
                pObjDriverNew = (Aig_Obj_t *)(Aig_Regular(pObjDriver).pData);
            } else {
                pObjDriverNew = Aig_Not((Aig_Obj_t *)(Aig_Regular(pObjDriver).pData));
            }
            pHintSignalLo = (Aig_Obj_t *)Vec_PtrEntry(vHintMonotoneLocalFlopOutput, iElem - oldPoNum);
            pObjTemp1 = Aig_Or( pNewAig, Aig_And(pNewAig, pObjDriverNew, pHintSignalLo), 
                        Aig_And(pNewAig, Aig_Not(pObjDriverNew), Aig_Not(pHintSignalLo)) );
            
            pObjKnownMonotoneAnd = Aig_And( pNewAig, pObjKnownMonotoneAnd, pObjTemp1 );
        }
        pObjPendingAndNextPending = Aig_And( pNewAig, pObjPendingAndNextPending, pObjKnownMonotoneAnd );
    }
          

    vHintMonotoneLocalDriverNew = Vec_PtrAlloc(Vec_IntSize(vHintMonotoneLocal));
    vHintMonotoneLocalProp = Vec_PtrAlloc(Vec_IntSize(vHintMonotoneLocal));
    Vec_IntForEachEntry( vHintMonotoneLocal, iElem, i )
    {
        pObjPo = Aig_ManCo( pAig, iElem );
        pObjDriver = Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObjPo), Aig_ObjFaninC0(pObjPo));
        if !Aig_IsComplement(pObjDriver) {
            pObjDriverNew = (Aig_Obj_t *)(Aig_Regular(pObjDriver).pData);
        } else {
            pObjDriverNew = Aig_Not((Aig_Obj_t *)(Aig_Regular(pObjDriver).pData));
        }

        if vKnownMonotoneLocal != nil && Vec_IntFind( vKnownMonotoneLocal, iElem ) != -1 {
            Vec_PtrPush(vHintMonotoneLocalDriverNew, pObjDriverNew);
        
        } else {
            poSerialNum = Vec_IntFind( vHintMonotoneLocal, iElem );
            pHintSignalLo = (Aig_Obj_t *)Vec_PtrEntry(vHintMonotoneLocalFlopOutput, poSerialNum );
            pObjTemp1 = Aig_And( pNewAig, pObjPendingAndNextPending, pHintSignalLo);
            pObjTemp2 = Aig_Or( pNewAig, Aig_Not(pObjTemp1), pObjDriverNew );
            //pObjTemp2 = Aig_Or( pNewAig, Aig_Not(pObjTemp1), Aig_ManConst1( pNewAig ));
            //pObjTemp2 = Aig_ManConst1( pNewAig );
            Vec_PtrPush(vHintMonotoneLocalDriverNew, pObjDriverNew);
            Vec_PtrPush(vHintMonotoneLocalProp, pObjTemp2); 
        }
    }

    //********************************************************************
    // Step 9: create primary outputs 
    //********************************************************************
    Saig_ManForEachPo( pAig, pObj, i )
    {
        poCopied++;
        pObj.pData = Aig_ObjCreateCo( pNewAig, Aig_ObjChild0Copy(pObj) ); 
    }

    *startMonotonePropPo = i;
    Vec_PtrForEachEntry( Aig_Obj_t *, vHintMonotoneLocalProp, pObj, i )
    {
        poCreated++;
        pObjPo = Aig_ObjCreateCo( pNewAig, pObj ); 
    }

    //********************************************************************
    // Step 9: create latch inputs 
    //********************************************************************

    Saig_ManForEachLi( pAig, pObj, i )
    {
        liCopied++;
        Aig_ObjCreateCo( pNewAig, Aig_ObjChild0Copy(pObj) );
    }
    
    //********************************************************************
    // Step 9.a: create latch input for PENDING_FLOP
    //********************************************************************

    liCreated++;
    Aig_ObjCreateCo( pNewAig, pObjPendingDriverNew );

    //********************************************************************
    // Step 9.b: create latch input for MONOTONE_FLOP
    //********************************************************************

    Vec_PtrForEachEntry( Aig_Obj_t *, vHintMonotoneLocalDriverNew, pObj, i )
    {
        liCreated++;
        Aig_ObjCreateCo( pNewAig, pObj );
    }

    fmt.Printf("\npoCopied = %d, poCreated = %d\n", poCopied, poCreated);
    fmt.Printf("\nliCreated++ = %d\n", liCreated);
    Aig_ManSetRegNum( pNewAig, nRegCount );
    Aig_ManCleanup( pNewAig );
    
    assert(Aig_ManCheck( pNewAig ), "Aig_ManCheck( pNewAig )");
    assert(loCopied + loCreated == liCopied + liCreated, "loCopied + loCreated == liCopied + liCreated");

    fmt.Printf("\nSaig_ManPoNum = %d\n", Saig_ManPoNum(pNewAig));
    return pNewAig;
}


Vec_Int_t *findNewMonotone( Aig_Man_t *pAig, struct aigPoIndices *aigPoIndicesArg, struct monotoneVectorsStruct *monotoneVectorArg )
{
    var pAigNew *Aig_Man_t;
    var pObjTargetPo *Aig_Obj_t;
    var poMarker, oldPoNum int32;
    var i, RetValue int32;
    var Pars Pdr_Par_t;
    var pPars *Pdr_Par_t = &Pars;
    var pCex *Abc_Cex_t  = nil;
    var vMonotoneIndex *Vec_Int_t;

    var pendingSignalIndexLocal int32  = aigPoIndicesArg.attrPendingSignalIndex;
    var hintSingalBeginningMarkerLocal int32  = aigPoIndicesArg.attrHintSingalBeginningMarker;
    //int hintSingalEndMarkerLocal = aigPoIndicesArg->attrHintSingalEndMarker;

    //Vec_Int_t *vKnownMonotoneLocal = monotoneVectorArg->attrKnownMonotone;
    //Vec_Int_t *vCandMonotoneLocal = monotoneVectorArg->attrCandMonotone;
    //Vec_Int_t *vHintMonotoneLocal = monotoneVectorArg->attrHintMonotone;

    pAigNew = createMonotoneTester(pAig, aigPoIndicesArg, monotoneVectorArg, &poMarker );
    oldPoNum = Aig_ManCoNum(pAig) - Aig_ManRegNum(pAig);

    vMonotoneIndex = Vec_IntAlloc(0);
    fmt.Printf("\nSaig_ManPoNum(pAigNew) = %d, poMarker = %d\n", Saig_ManPoNum(pAigNew), poMarker);
    for  i=poMarker; i<Saig_ManPoNum(pAigNew); i++  {
        pObjTargetPo = Aig_ManCo( pAigNew, i );
        Aig_ObjChild0Flip( pObjTargetPo );

        Pdr_ManSetDefaultParams( pPars );
        pCex = nil;
        pPars.fVerbose = 0;
        //pPars->iOutput = i;
        //RetValue = Pdr_ManSolve( pAigNew, pPars, &pCex );    
        RetValue = Pdr_ManSolve( pAigNew, pPars );    
        if RetValue == 1 {
            fmt.Printf("\ni = %d, RetValue = %d : %s (Frame %d)\n", i - oldPoNum + hintSingalBeginningMarkerLocal, RetValue, "Property Proved", pCex? (pCex).iFrame : -1);
            Vec_IntPush( vMonotoneIndex, i - (pendingSignalIndexLocal + 1) + hintSingalBeginningMarkerLocal);
        }
        Aig_ObjChild0Flip( pObjTargetPo );
    }
    
    if Vec_IntSize( vMonotoneIndex ) > 0 {
        return vMonotoneIndex;
    } else {
        return nil; }
}

func findRemainingMonotoneCandidates(vKnownMonotone *Vec_Int_t, vHintMonotone *Vec_Int_t) *Vec_Int_t {
    var vCandMonotone *Vec_Int_t;
    var iElem, i int32;

    if vKnownMonotone == nil || Vec_IntSize(vKnownMonotone) <= 0 {
        return vHintMonotone; }
    vCandMonotone = Vec_IntAlloc(0);
    Vec_IntForEachEntry( vHintMonotone, iElem, i )
    {
        if Vec_IntFind( vKnownMonotone, iElem ) == -1 {
            Vec_IntPush( vCandMonotone, iElem ); }
    }
    
    return vCandMonotone;
}

func findMonotoneSignals(pNtk *Abc_Ntk_t) *Vec_Int_t {
    var pAig *Aig_Man_t;
    var vCandidateMonotoneSignals *Vec_Int_t;
    var vKnownMonotoneSignals *Vec_Int_t;
    //Vec_Int_t *vKnownMonotoneSignalsNew;
    //Vec_Int_t *vRemainingCanMonotone;
    var i, iElem int32;
    var pendingSignalIndex int32;
    var pNtkTemp *Abc_Ntk_t;
    var hintSingalBeginningMarker int32;
    var hintSingalEndMarker int32;
    struct aigPoIndices *aigPoIndicesInstance;
    struct monotoneVectorsStruct *monotoneVectorsInstance;
    
    /*******************************************/    
    //Finding the PO index of the pending signal 
    /*******************************************/    
    pendingSignalIndex = findPendingSignal(pNtk);
    if pendingSignalIndex == -1 {
        fmt.Printf("\nNo Pending Signal Found\n");
        return nil;
    
    } else {
        fmt.Printf("Po[%d] = %s\n", pendingSignalIndex, Abc_ObjName( Abc_NtkPo(pNtk, pendingSignalIndex) )); }

    /*******************************************/    
    //Finding the PO indices of all hint signals
    /*******************************************/    
    vCandidateMonotoneSignals = findHintOutputs(pNtk);
    if vCandidateMonotoneSignals == nil {
        return nil;
    } else {
        Vec_IntForEachEntry( vCandidateMonotoneSignals, iElem, i )
            fmt.Printf("Po[%d] = %s\n", iElem, Abc_ObjName( Abc_NtkPo(pNtk, iElem) ));
        hintSingalBeginningMarker = Vec_IntEntry( vCandidateMonotoneSignals, 0 );
        hintSingalEndMarker = Vec_IntEntry( vCandidateMonotoneSignals, Vec_IntSize(vCandidateMonotoneSignals) - 1 );
    }

    /**********************************************/
    //Allocating "struct" with necessary parameters
    /**********************************************/
    aigPoIndicesInstance = allocAigPoIndices();
    aigPoIndicesInstance.attrPendingSignalIndex = pendingSignalIndex;
    aigPoIndicesInstance.attrHintSingalBeginningMarker = hintSingalBeginningMarker;
    aigPoIndicesInstance.attrHintSingalEndMarker = hintSingalEndMarker;

    /****************************************************/
    //Allocating "struct" with necessary monotone vectors
    /****************************************************/
    monotoneVectorsInstance = allocPointersToMonotoneVectors();
    monotoneVectorsInstance.attrCandMonotone = vCandidateMonotoneSignals;
    monotoneVectorsInstance.attrHintMonotone = vCandidateMonotoneSignals;

    /*******************************************/    
    //Generate AIG from Ntk
    /*******************************************/    
    if !Abc_NtkIsStrash( pNtk ) {
        pNtkTemp = Abc_NtkStrash( pNtk, 0, 0, 0 );
        pAig = Abc_NtkToDar( pNtkTemp, 0, 1 );
    
    } else {
        pAig = Abc_NtkToDar( pNtk, 0, 1 );
        pNtkTemp = pNtk;
    }

    /*******************************************/    
    //finding LEVEL 1 monotone signals
    /*******************************************/    
    vKnownMonotoneSignals = findNewMonotone( pAig, aigPoIndicesInstance, monotoneVectorsInstance );
    monotoneVectorsInstance.attrKnownMonotone = vKnownMonotoneSignals;

    /*******************************************/    
    //finding LEVEL >1 monotone signals
    /*******************************************/    
         
                               
     
                                                                     
                                                                                                                  
                                                                           
                                                                                                          
     
           

    deallocAigPoIndices(aigPoIndicesInstance);
    deallocPointersToMonotoneVectors(monotoneVectorsInstance);
    return nil;
}

//ABC_NAMESPACE_IMPL_END
